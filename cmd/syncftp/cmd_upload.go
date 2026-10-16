package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"syncftp/internal/config"
	ftpclient "syncftp/internal/ftp"
	"syncftp/internal/ignore"
	"syncftp/internal/lang"
	"syncftp/internal/release"
	"syncftp/internal/scanner"
	"syncftp/internal/state"
)

// checkpointEvery — kaç başarılı upload'da bir state diske yazılır. Sync yarıda
// kalırsa (elektrik, Ctrl+C, kopma) bir sonraki sync yalnızca kalanları yükler.
const checkpointEvery = 20

// uploadOutcome — runUploadSession sonucu.
type uploadOutcome struct {
	results      []ftpclient.UploadResult
	successFiles map[string]string // relPath → hash
	releaseDir   string            // boş → release yazılmadı
	releaseID    string
	poolSize     int
	backed       int
	newFiles     int
	bytes        int64
}

// runUploadSession tüm sync akışlarının (CLI sync, shell sync, milestone, serve/GUI)
// ortak upload çekirdeği:
//
//   - Yedek (rollback): sunucuda BackupEnabled ise her dosyanın eski uzak hali
//     .syncftp/releases/<srv>/<id>/files/ altına indirilir; sunucuda olmayanlar "new" kaydedilir.
//   - Pool: host başına paylaşılan bağlantı limiti + paylaşılan bant genişliği sınırı.
//   - Checkpoint: başarılı dosyalar anında st.Files'a yazılır, her checkpointEvery dosyada state kaydedilir.
//   - Release manifest: başarılı dosyalar + rollback bilgisi (Backed/New/PrevHashes).
//
// onResult her sonuç geldiğinde (canlı) çağrılır. State'in son temizliği (silinen
// dosyalar, FirstSyncDone) ve final kaydı çağıranın işidir.
func runUploadSession(w io.Writer, configDir string, cfg *config.Config, srv config.Server, tasks []ftpclient.UploadTask, st *state.State, onResult func(ftpclient.UploadResult)) (uploadOutcome, error) {
	out := uploadOutcome{successFiles: make(map[string]string)}

	backup := srv.BackupEnabled(cfg.Sync)
	out.releaseID = release.NewID(configDir, srv.Name)
	if backup {
		dir := release.FilesDir(configDir, srv.Name, out.releaseID)
		for i := range tasks {
			tasks[i].BackupDir = dir
		}
	}

	if kbps := srv.EffectiveBandwidth(cfg.Sync); kbps > 0 {
		fmt.Fprintf(w, lang.L.SyncBandwidthFmt, kbps)
	}

	pool, err := ftpclient.NewPoolLimited(srv, srv.EffectiveBandwidth(cfg.Sync))
	if err != nil {
		return out, err
	}
	defer pool.Close()
	out.poolSize = pool.Size()
	if want := max(srv.MaxConnections, 1); out.poolSize < want {
		fmt.Fprintf(w, lang.L.SyncPoolLimitedFmt, out.poolSize, want)
	}

	prev := make(map[string]string, len(tasks))
	for _, t := range tasks {
		if h, ok := st.Files[t.RelPath]; ok {
			prev[t.RelPath] = h
		}
	}

	m := &release.Manifest{Files: map[string]string{}, PrevHashes: map[string]string{}}
	sinceCheckpoint := 0
	for r := range pool.UploadCh(tasks) {
		out.results = append(out.results, r)
		if onResult != nil {
			onResult(r)
		}
		if r.Err != nil {
			continue
		}
		out.successFiles[r.RelPath] = r.Hash
		out.bytes += r.Size
		st.Files[r.RelPath] = r.Hash
		m.Files[r.RelPath] = r.Hash
		if h, ok := prev[r.RelPath]; ok {
			m.PrevHashes[r.RelPath] = h
		}
		switch {
		case r.BackedUp:
			m.Backed = append(m.Backed, r.RelPath)
			out.backed++
		case r.IsNew:
			m.New = append(m.New, r.RelPath)
			out.newFiles++
		case r.BackupErr != nil:
			m.NoBackup = append(m.NoBackup, r.RelPath)
			fmt.Fprintf(w, lang.L.SyncBackupErrFmt, r.RelPath, r.BackupErr)
		}
		if sinceCheckpoint++; sinceCheckpoint >= checkpointEvery {
			_ = state.Save(configDir, st)
			sinceCheckpoint = 0
		}
	}

	if len(out.successFiles) == 0 {
		release.Discard(configDir, srv.Name, out.releaseID)
		return out, nil
	}
	if !backup {
		m.Version = 1 // yedeksiz — rollback edilemez
	}
	dir, werr := release.Write(configDir, srv.Name, out.releaseID, m)
	if werr != nil {
		fmt.Fprintf(w, lang.L.SyncReleaseErr, werr)
		return out, nil
	}
	out.releaseDir = dir
	if backup {
		fmt.Fprintf(w, lang.L.SyncBackupSummaryFmt, out.backed, out.newFiles, srv.Name)
		release.Prune(configDir, srv.Name, cfg.Sync.EffectiveBackupKeep())
	}
	return out, nil
}

// printUploadResult tek bir upload sonucunu standart formatta yazar.
func printUploadResult(w io.Writer, r ftpclient.UploadResult) {
	switch {
	case r.Err != nil && r.Attempts > 1:
		fmt.Fprintf(w, lang.L.SyncAttemptsFmt, r.RelPath, r.Attempts, r.Err)
	case r.Err != nil:
		fmt.Fprintf(w, lang.L.SyncUploadErrFmt, r.RelPath, r.Err)
	case r.Attempts > 1:
		fmt.Fprintf(w, lang.L.SyncAttemptOkFmt, r.RelPath, r.Attempts)
	default:
		fmt.Fprintf(w, lang.L.SyncUploadOkFmt, r.RelPath)
	}
}

// scanForServer sunucunun kendi yerel dizinini (local_path / default_path) tarar.
func scanForServer(configDir string, cfg *config.Config, srv config.Server) (map[string]string, map[string]scanner.File, error) {
	localDir := filepath.Join(configDir, srv.EffectiveLocalPath(cfg.Project))
	if st, err := os.Stat(localDir); err != nil || !st.IsDir() {
		return nil, nil, fmt.Errorf(lang.L.LocalDirMissingFmt, srv.EffectiveLocalPath(cfg.Project), localDir)
	}
	matcher, err := ignore.Load(localDir, cfg.Sync.IgnoreFiles)
	if err != nil {
		return nil, nil, err
	}
	files, err := scanner.Scan(localDir, matcher)
	if err != nil {
		return nil, nil, err
	}
	current := make(map[string]string, len(files))
	byKey := make(map[string]scanner.File, len(files))
	for _, f := range files {
		current[f.RelPath] = f.Hash
		byKey[f.RelPath] = f
	}
	return current, byKey, nil
}

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
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
	canceled     bool     // ctx iptal edildi — kalan görevler başlatılmadı
	deleted      []string // sunucudan silinenler (delete_remote)
	deleteFailed []string
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
func runUploadSession(ctx context.Context, w io.Writer, configDir string, cfg *config.Config, srv config.Server, tasks []ftpclient.UploadTask, st *state.State, onResult func(ftpclient.UploadResult), deletes ...string) (uploadOutcome, error) {
	out := uploadOutcome{successFiles: make(map[string]string)}
	if ctx == nil {
		ctx = context.Background()
	}

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

	pool, err := ftpclient.NewPoolCtx(ctx, srv, srv.EffectiveBandwidth(cfg.Sync))
	if err != nil {
		return out, err
	}
	poolOpen := true
	defer func() {
		if poolOpen {
			pool.Close()
		}
	}()
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

	// Yüklenen sürümün kopyası release'e (yedekle birlikte) → Geçmiş'te dosya diff'i
	srcOf := make(map[string]string, len(tasks))
	for _, t := range tasks {
		srcOf[t.RelPath] = t.LocalPath
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
		if backup {
			_ = release.SaveAfter(configDir, srv.Name, out.releaseID, r.RelPath, srcOf[r.RelPath])
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

	out.canceled = ctx.Err() != nil

	// Sunucudan silme (delete_remote): yüklemelerden SONRA (taşınan dosyanın yenisi önce çıksın),
	// her dosya önce yedeklenir — yedek alınamazsa SİLİNMEZ (geri alınamaz kayıp olmasın).
	if len(deletes) > 0 && !out.canceled {
		pool.Close() // bağlantı hakkını bırak; silme tek bağlantıyla
		poolOpen = false
		if derr := deleteRemoteFiles(ctx, w, configDir, srv, backup, out.releaseID, deletes, st, m, &out); derr != nil {
			fmt.Fprintf(w, lang.L.SyncServerErrFmt, derr)
			out.deleteFailed = append(out.deleteFailed, deletes...)
		}
		fmt.Fprintf(w, lang.L.SyncDeleteDoneFmt, len(out.deleted), len(out.deleteFailed))
		out.canceled = ctx.Err() != nil
	}
	if len(out.successFiles) == 0 && len(out.deleted) == 0 {
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

// deleteRemoteFiles yerelde silinen dosyaları sunucudan siler (yedekli) ve release'e / state'e işler.
func deleteRemoteFiles(ctx context.Context, w io.Writer, configDir string, srv config.Server, backup bool, releaseID string, rels []string, st *state.State, m *release.Manifest, out *uploadOutcome) error {
	client, err := ftpclient.Connect(srv)
	if err != nil {
		return err
	}
	defer client.Close()
	stop := context.AfterFunc(ctx, client.Abort)
	defer stop()
	filesDir := release.FilesDir(configDir, srv.Name, releaseID)
	for _, rel := range rels {
		if ctx.Err() != nil {
			break
		}
		backed := false
		if backup {
			existed, berr := client.BackupTo(rel, filepath.Join(filesDir, filepath.FromSlash(rel)))
			if berr != nil {
				fmt.Fprintf(w, lang.L.SyncDeleteErrFmt, rel, berr)
				out.deleteFailed = append(out.deleteFailed, rel)
				continue
			}
			if !existed { // sunucuda zaten yok — yalnızca state'ten düşer
				delete(st.Files, rel)
				continue
			}
			backed = true
		}
		if derr := client.DeleteFile(path.Join(srv.RemotePath, rel)); derr != nil && !ftpclient.IsNotFound(derr) {
			fmt.Fprintf(w, lang.L.SyncDeleteErrFmt, rel, derr)
			out.deleteFailed = append(out.deleteFailed, rel)
			continue
		}
		fmt.Fprintf(w, lang.L.SyncDeleteOkFmt, rel)
		out.deleted = append(out.deleted, rel)
		if h, ok := st.Files[rel]; ok {
			m.PrevHashes[rel] = h
		}
		delete(st.Files, rel)
		m.Files[rel] = release.DeletedHash
		m.Deleted = append(m.Deleted, rel)
		if backed {
			out.backed++
			m.Backed = append(m.Backed, rel)
		} else {
			m.NoBackup = append(m.NoBackup, rel)
		}
	}
	return nil
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

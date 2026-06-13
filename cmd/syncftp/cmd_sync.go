package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"syncftp/internal/config"
	"syncftp/internal/failed"
	ftpclient "syncftp/internal/ftp"
	"syncftp/internal/frozen"
	"syncftp/internal/ignore"
	"syncftp/internal/lang"
	"syncftp/internal/release"
	"syncftp/internal/scanner"
	"syncftp/internal/state"
)

var (
	flagFull        bool
	flagAll         bool
	flagServer      string
	flagDryRun      bool
	flagInclude     []string
	flagExclude     []string
	flagRetryFailed bool
)

func init() {
	syncCmd.Flags().BoolVar(&flagFull, "full", false, "State'i yoksay, tüm dosyaları yükle")
	syncCmd.Flags().BoolVar(&flagAll, "all", false, "Tüm aktif sunuculara sync et (seçici açılmaz)")
	syncCmd.Flags().StringVar(&flagServer, "server", "", "Sadece belirtilen sunucuya yükle")
	syncCmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "Ne yükleneceğini göster, gerçekten yükleme")
	syncCmd.Flags().StringArrayVar(&flagInclude, "include", nil, "Sadece bu yol/klasörleri sync et (whitelist; JSON include'u geçersiz kılar)")
	syncCmd.Flags().StringArrayVar(&flagExclude, "exclude", nil, "Bu yol/klasörleri bu sync'ten hariç tut (tek seferlik)")
	syncCmd.Flags().BoolVar(&flagRetryFailed, "retry-failed", false, "Önceki sync'te başarısız olan dosyaları tekrar yükle")
	rootCmd.AddCommand(syncCmd)
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Değişen dosyaları FTP sunucularına yükler",
	RunE:  runSync,
}

func runSync(cmd *cobra.Command, args []string) error {
	dir, _ := os.Getwd()

	cfg, err := config.Load(dir)
	if err != nil {
		return err
	}

	servers := cfg.EnabledServers()
	if flagServer != "" {
		var found []config.Server
		for _, s := range servers {
			if s.Name == flagServer {
				found = append(found, s)
			}
		}
		if len(found) == 0 {
			return fmt.Errorf("sunucu bulunamadı: %q (mevcut: %v)", flagServer, serverNames(servers))
		}
		servers = found
	} else if !flagAll && len(servers) > 1 {
		selected, err := pickServerMultiTUI(servers, "")
		if err != nil {
			return err
		}
		if selected == nil {
			fmt.Println(lang.L.SyncCancelled)
			return nil
		}
		servers = selected
	}

	if flagDryRun {
		fmt.Println(lang.L.SyncDryRunNote)
	}

	if len(flagInclude) > 0 {
		fmt.Printf(lang.L.SyncWhitelistFmt, len(flagInclude))
		for _, p := range flagInclude {
			fmt.Printf("  + %s\n", p)
		}
		fmt.Println()
	}
	if len(flagExclude) > 0 {
		fmt.Printf(lang.L.SyncExcludeFmt, len(flagExclude))
		for _, p := range flagExclude {
			fmt.Printf("  - %s\n", p)
		}
		fmt.Println()
	}

	for _, srv := range servers {
		fmt.Printf("══ %s (%s) ══\n", srv.Name, srv.Host)

		localPath := srv.EffectiveLocalPath(cfg.Project)
		localDir := filepath.Join(dir, localPath)
		fmt.Printf("  Taranıyor: %s\n", localPath)

		matcher, err := ignore.Load(localDir, cfg.Sync.IgnoreFiles)
		if err != nil {
			fmt.Printf("  Tarama hatası: %v\n\n", err)
			continue
		}
		files, err := scanner.Scan(localDir, matcher)
		if err != nil {
			fmt.Printf("  Tarama hatası: %v\n\n", err)
			continue
		}

		current := make(map[string]string)
		byKey := make(map[string]scanner.File)
		for _, f := range files {
			current[f.RelPath] = f.Hash
			byKey[f.RelPath] = f
		}
		fmt.Printf("  %d dosya\n\n", len(current))

		if err := syncToServer(dir, cfg, srv, current, byKey, flagInclude, flagExclude); err != nil {
			fmt.Printf(lang.L.SyncServerErrFmt, err)
		}
		fmt.Println()
	}

	return nil
}

func syncToServer(configDir string, cfg *config.Config, srv config.Server, current map[string]string, byKey map[string]scanner.File, cliInclude, cliExclude []string) error {
	st, err := state.Load(configDir, srv.Name)
	if err != nil {
		return err
	}

	// Include önceliği: CLI > sunucu > global
	effectiveInclude := cfg.Sync.Include
	if len(srv.Include) > 0 {
		effectiveInclude = srv.Include
	}
	if len(cliInclude) > 0 {
		effectiveInclude = cliInclude
	}

	// Exclude: global + sunucu + CLI (hepsi toplanır)
	effectiveExclude := append(append([]string{}, cfg.Sync.Exclude...), srv.Exclude...)
	effectiveExclude = append(effectiveExclude, cliExclude...)

	// Eski browser'la kaydedilen "local_path/dosya" prefix'li pattern'leri normalize et
	localPath := srv.EffectiveLocalPath(cfg.Project)
	effectiveInclude = normalizePatterns(effectiveInclude, localPath)
	effectiveExclude = normalizePatterns(effectiveExclude, localPath)

	// Protect normalize (global + sunucu)
	effectiveProtect := append(append([]string{}, cfg.Sync.Protect...), srv.Protect...)
	effectiveProtect = normalizePatterns(effectiveProtect, localPath)

	var toUpload []string

	if flagRetryFailed {
		fl, err := failed.Load(configDir, srv.Name)
		if err != nil {
			return fmt.Errorf("failed listesi okunamadı: %w", err)
		}
		if fl == nil || len(fl.Files) == 0 {
			fmt.Println(lang.L.SyncRetryNoFiles)
			return nil
		}
		fmt.Printf(lang.L.SyncRetryModeFmt, len(fl.Files), fl.CreatedAt.Format("2006-01-02 15:04:05"))
		// Sadece hala local'de var olan dosyaları al
		for _, rel := range fl.Files {
			if _, exists := current[rel]; exists {
				toUpload = append(toUpload, rel)
			} else {
				fmt.Printf(lang.L.SyncRetrySkipFmt, rel)
			}
		}
	} else {
		isFirst := !st.FirstSyncDone || flagFull

		if isFirst {
			if flagFull && st.FirstSyncDone {
				fmt.Println(lang.L.SyncFullFlag)
				toUpload = mapKeys(current)
			} else {
				// Akıllı ilk sync: sunucudaki dosyaları boyutla karşılaştır
				toUpload = smartFirstSync(srv, current, byKey, st, cfg)
			}

			if len(effectiveInclude) > 0 {
				toUpload = filterByInclude(toUpload, effectiveInclude)
			}
		} else {
			diff := state.Diff(st, current)

			if len(diff.Deleted) > 0 {
				sort.Strings(diff.Deleted)
				fmt.Print(lang.L.SyncDeletedHeader)
				for _, p := range diff.Deleted {
					fmt.Printf("      - %s\n", p)
				}
			}

			toUpload = append(diff.New, diff.Changed...)
			if len(effectiveInclude) > 0 {
				toUpload = filterByInclude(toUpload, effectiveInclude)
			}
		}

		toUpload = filterByExclude(toUpload, effectiveExclude)
	}

	// Frozen dosyaları atla
	fl, _ := frozen.Load(configDir, srv.Name)
	if fl != nil && len(fl.Files) > 0 {
		var unfrozen []string
		frozenSkipped := 0
		for _, rel := range toUpload {
			if frozen.IsFrozen(fl, rel) {
				frozenSkipped++
			} else {
				unfrozen = append(unfrozen, rel)
			}
		}
		if frozenSkipped > 0 {
			fmt.Printf("  ❄ %d frozen (skipped)\n", frozenSkipped)
		}
		toUpload = unfrozen
	}

	sort.Strings(toUpload)

	if len(toUpload) == 0 {
		fmt.Println(lang.L.SyncNoChange2)
		return nil
	}

	// Korunan dosyaları ayır
	// protect kalıpları hem stateKey'e hem kaynak-içi RelPath'e karşı kontrol edilir
	var tasks []ftpclient.UploadTask
	skipped := 0
	for _, rel := range toUpload {
		f := byKey[rel]
		if ftpclient.IsProtected(rel, effectiveProtect) {
			skipped++
			continue
		}
		tasks = append(tasks, ftpclient.UploadTask{LocalPath: f.AbsPath, RelPath: rel, Hash: current[rel]})
	}

	total := len(tasks) + skipped
	fmt.Printf(lang.L.SyncProcessingFmt, total)
	if skipped > 0 {
		fmt.Printf(lang.L.SyncProtectedFmt, skipped)
	}
	fmt.Println()

	if flagDryRun {
		for _, rel := range toUpload {
			f := byKey[rel]
			if ftpclient.IsProtected(rel, effectiveProtect) || ftpclient.IsProtected(f.RelPath, effectiveProtect) {
				fmt.Printf(lang.L.SyncProtectedLabel, rel)
			} else {
				fmt.Printf(lang.L.SyncUploadLabel, rel)
			}
		}
		return nil
	}

	maxConn := srv.MaxConnections
	if maxConn <= 0 {
		maxConn = 1
	}
	maxRetry := srv.MaxRetries
	if maxRetry == 0 {
		maxRetry = 2
	}
	fmt.Printf(lang.L.SyncPoolFmt, maxConn, maxRetry)

	pool, err := ftpclient.NewPool(srv)
	if err != nil {
		return err
	}
	defer pool.Close()

	results := pool.Upload(tasks)

	// Korunan dosyaları logla
	for _, rel := range toUpload {
		if ftpclient.IsProtected(rel, effectiveProtect) {
			fmt.Printf(lang.L.SyncProtectedLabel, rel)
		}
	}

	successFiles := make(map[string]string)
	uploaded, failedCount := 0, 0
	for _, r := range results {
		if r.Err != nil {
			if r.Attempts > 1 {
				fmt.Printf(lang.L.SyncAttemptsFmt, r.RelPath, r.Attempts, r.Err)
			} else {
				fmt.Printf(lang.L.SyncUploadErrFmt, r.RelPath, r.Err)
			}
			failedCount++
		} else {
			if r.Attempts > 1 {
				fmt.Printf(lang.L.SyncAttemptOkFmt, r.RelPath, r.Attempts)
			} else {
				fmt.Printf(lang.L.SyncUploadOkFmt, r.RelPath)
			}
			successFiles[r.RelPath] = r.Hash
			uploaded++
		}
	}

	fmt.Printf(lang.L.SyncDoneFullFmt, uploaded, skipped, failedCount)

	// Başarısız dosyaları kaydet veya temizle
	var failedPaths []string
	for _, r := range results {
		if r.Err != nil {
			failedPaths = append(failedPaths, r.RelPath)
		}
	}
	if len(failedPaths) > 0 {
		if err := failed.Save(configDir, srv.Name, failedPaths); err != nil {
			fmt.Printf(lang.L.SyncFailedSaveErr, err)
		} else {
			fmt.Printf(lang.L.SyncFailedSavedFmt, len(failedPaths), srv.Name)
		}
	} else {
		failed.Clear(configDir, srv.Name)
	}

	// Update state with successful uploads
	for rel, hash := range successFiles {
		st.Files[rel] = hash
	}
	// Remove locally deleted files from state
	for rel := range st.Files {
		if _, exists := current[rel]; !exists {
			delete(st.Files, rel)
		}
	}
	st.FirstSyncDone = true

	if err := state.Save(configDir, st); err != nil {
		fmt.Printf(lang.L.SyncStateErr, err)
	}

	if len(successFiles) > 0 {
		if relDir, err := release.Create(configDir, srv.Name, successFiles); err != nil {
			fmt.Printf(lang.L.SyncReleaseErr, err)
		} else {
			fmt.Printf(lang.L.SyncReleaseFmt, relDir)
		}
	}

	return nil
}

// filterByInclude returns only paths that match any of the include patterns.
// A pattern matches if it's an exact path or a directory prefix.
// If include is empty, all paths are returned unchanged.
func filterByInclude(paths, include []string) []string {
	if len(include) == 0 {
		return paths
	}
	var out []string
	for _, p := range paths {
		for _, inc := range include {
			inc = filepath.ToSlash(inc)
			if p == inc || strings.HasPrefix(p, inc+"/") {
				out = append(out, p)
				break
			}
		}
	}
	return out
}

// filterByExclude removes paths that match any of the exclude patterns.
// A pattern matches if it's an exact path or a directory prefix.
func filterByExclude(paths, exclude []string) []string {
	if len(exclude) == 0 {
		return paths
	}
	var out []string
	for _, p := range paths {
		excluded := false
		for _, exc := range exclude {
			exc = filepath.ToSlash(exc)
			if p == exc || strings.HasPrefix(p, exc+"/") {
				excluded = true
				break
			}
		}
		if !excluded {
			out = append(out, p)
		}
	}
	return out
}

// normalizePatterns eski tarayıcıyla kaydedilen local_path prefix'li pattern'leri
// temizler. Örn: "app.aracservistakip.com/composer.json" → "composer.json"
// localPath boş veya "." ise dokunulmaz.
func normalizePatterns(patterns []string, localPath string) []string {
	if localPath == "" || localPath == "." {
		return patterns
	}
	prefix := filepath.ToSlash(strings.TrimSuffix(localPath, "/")) + "/"
	out := make([]string, len(patterns))
	for i, p := range patterns {
		p = filepath.ToSlash(p)
		if strings.HasPrefix(p, prefix) {
			out[i] = p[len(prefix):]
		} else {
			out[i] = p
		}
	}
	return out
}

func mapKeys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// smartFirstSync sunucudaki dosyaları local boyutlarla karşılaştırır.
// Eşleşen dosyalar state'e "zaten senkron" olarak kaydedilir, yüklenmez.
// Farklı boyuttaki veya sunucuda olmayan dosyalar yükleme listesine girer.
// Sunucuya bağlanılamazsa güvenli taraf: tüm dosyalar yüklenir.
// smartFirstSync stateKey'leri doğrudan FTP listeleme sonuçlarıyla karşılaştırır.
// stateKey = srcKey(source.Remote, relPath) olduğundan FTP göreli yoluyla birebir eşleşir.
func smartFirstSync(srv config.Server, current map[string]string, byKey map[string]scanner.File, st *state.State, cfg *config.Config) []string {
	fmt.Println(lang.L.SmartSyncScanning)

	client, err := ftpclient.Connect(srv)
	if err != nil {
		fmt.Printf(lang.L.SmartSyncConnErr, err)
		return mapKeys(current)
	}
	defer client.Close()

	remoteFiles, err := client.ListRecursive(srv.RemotePath)
	if err != nil {
		fmt.Printf(lang.L.SmartSyncListErr, err)
		return mapKeys(current)
	}

	fmt.Printf(lang.L.SmartSyncFoundFmt, len(remoteFiles))

	var toUpload []string
	alreadySynced := 0

	for key, hash := range current {
		f := byKey[key]

		localInfo, err := os.Stat(f.AbsPath)
		if err != nil {
			toUpload = append(toUpload, key)
			continue
		}
		localSize := uint64(localInfo.Size())

		remoteSize, exists := remoteFiles[key]
		if !exists || remoteSize != localSize {
			toUpload = append(toUpload, key)
		} else {
			// Boyut eşleşiyor → zaten güncel, state'e kaydet
			st.Files[key] = hash
			alreadySynced++
		}
	}

	fmt.Printf(lang.L.SmartSyncResultFmt, alreadySynced, len(toUpload))

	return toUpload
}

func serverNames(servers []config.Server) []string {
	names := make([]string, len(servers))
	for i, s := range servers {
		names[i] = s.Name
	}
	return names
}

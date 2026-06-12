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
	"syncftp/internal/ignore"
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

	projectDir := filepath.Join(dir, cfg.Project.LocalPath)

	matcher, err := ignore.Load(projectDir)
	if err != nil {
		return fmt.Errorf("ignore dosyası yüklenemedi: %w", err)
	}

	fmt.Printf("Taranıyor: %s\n", projectDir)
	files, err := scanner.Scan(projectDir, matcher)
	if err != nil {
		return fmt.Errorf("tarama başarısız: %w", err)
	}
	fmt.Printf("%d dosya bulundu\n\n", len(files))

	current := make(map[string]string, len(files))
	byRel := make(map[string]scanner.File, len(files))
	for _, f := range files {
		current[f.RelPath] = f.Hash
		byRel[f.RelPath] = f
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
		selected, err := pickServerMultiTUI(servers)
		if err != nil {
			return err
		}
		if selected == nil {
			fmt.Println("İptal.")
			return nil
		}
		servers = selected
	}

	if flagDryRun {
		fmt.Println("[DRY RUN] Hiçbir şey yüklenmeyecek")
	}

	if len(flagInclude) > 0 {
		fmt.Printf("Whitelist (%d yol): yalnızca bu yollar sync edilecek\n", len(flagInclude))
		for _, p := range flagInclude {
			fmt.Printf("  + %s\n", p)
		}
		fmt.Println()
	}
	if len(flagExclude) > 0 {
		fmt.Printf("Exclude (%d yol): bu yollar bu sync'ten hariç tutulacak\n", len(flagExclude))
		for _, p := range flagExclude {
			fmt.Printf("  - %s\n", p)
		}
		fmt.Println()
	}

	for _, srv := range servers {
		fmt.Printf("══ %s (%s) ══\n", srv.Name, srv.Host)
		if err := syncToServer(dir, cfg, srv, current, byRel, flagInclude, flagExclude); err != nil {
			fmt.Printf("  HATA: %v\n", err)
		}
		fmt.Println()
	}

	return nil
}

func syncToServer(configDir string, cfg *config.Config, srv config.Server, current map[string]string, byRel map[string]scanner.File, cliInclude, cliExclude []string) error {
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

	var toUpload []string

	if flagRetryFailed {
		fl, err := failed.Load(configDir, srv.Name)
		if err != nil {
			return fmt.Errorf("failed listesi okunamadı: %w", err)
		}
		if fl == nil || len(fl.Files) == 0 {
			fmt.Println("  Yeniden denenecek başarısız dosya yok")
			return nil
		}
		fmt.Printf("  Retry modu: %d başarısız dosya (%s)\n", len(fl.Files), fl.CreatedAt.Format("2006-01-02 15:04:05"))
		// Sadece hala local'de var olan dosyaları al
		for _, rel := range fl.Files {
			if _, exists := current[rel]; exists {
				toUpload = append(toUpload, rel)
			} else {
				fmt.Printf("  ! %s artık local'de yok, atlanıyor\n", rel)
			}
		}
	} else {
		isFirst := !st.FirstSyncDone || flagFull

		if isFirst {
			if flagFull && st.FirstSyncDone {
				fmt.Println("  Tam sync (--full): tüm dosyalar yüklenecek")
				toUpload = mapKeys(current)
			} else {
				// Akıllı ilk sync: sunucudaki dosyaları boyutla karşılaştır
				toUpload = smartFirstSync(srv, current, byRel, st, cfg)
			}

			if len(effectiveInclude) > 0 {
				toUpload = filterByInclude(toUpload, effectiveInclude)
			}
		} else {
			diff := state.Diff(st, current)

			if len(diff.Deleted) > 0 {
				sort.Strings(diff.Deleted)
				fmt.Printf("  ! SİLİNEN dosyalar (FTP'de bırakıldı):\n")
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

	sort.Strings(toUpload)

	if len(toUpload) == 0 {
		fmt.Println("  Değişiklik yok")
		return nil
	}

	// Korunan dosyaları ayır
	var tasks []ftpclient.UploadTask
	skipped := 0
	for _, rel := range toUpload {
		if ftpclient.IsProtected(rel, cfg.Sync.Protect) {
			skipped++
			continue
		}
		f := byRel[rel]
		tasks = append(tasks, ftpclient.UploadTask{LocalPath: f.AbsPath, RelPath: rel, Hash: current[rel]})
	}

	total := len(tasks) + skipped
	fmt.Printf("  %d dosya işlenecek", total)
	if skipped > 0 {
		fmt.Printf(" (%d korunuyor)", skipped)
	}
	fmt.Println()

	if flagDryRun {
		for _, rel := range toUpload {
			if ftpclient.IsProtected(rel, cfg.Sync.Protect) {
				fmt.Printf("    KORUNUYOR  %s\n", rel)
			} else {
				fmt.Printf("    YÜKLENECEK %s\n", rel)
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
	fmt.Printf("  Bağlantı havuzu: %d / Retry: %d\n", maxConn, maxRetry)

	pool, err := ftpclient.NewPool(srv)
	if err != nil {
		return err
	}
	defer pool.Close()

	results := pool.Upload(tasks)

	// Korunan dosyaları logla
	for _, rel := range toUpload {
		if ftpclient.IsProtected(rel, cfg.Sync.Protect) {
			fmt.Printf("    KORUNUYOR  %s\n", rel)
		}
	}

	successFiles := make(map[string]string)
	uploaded, failedCount := 0, 0
	for _, r := range results {
		if r.Err != nil {
			if r.Attempts > 1 {
				fmt.Printf("    ✗ %s (%d deneme): %v\n", r.RelPath, r.Attempts, r.Err)
			} else {
				fmt.Printf("    ✗ %s: %v\n", r.RelPath, r.Err)
			}
			failedCount++
		} else {
			if r.Attempts > 1 {
				fmt.Printf("    ✓ %s (%d. denemede başarılı)\n", r.RelPath, r.Attempts)
			} else {
				fmt.Printf("    ✓ %s\n", r.RelPath)
			}
			successFiles[r.RelPath] = r.Hash
			uploaded++
		}
	}

	fmt.Printf("  Tamamlandı: %d yüklendi, %d korundu, %d hata\n", uploaded, skipped, failedCount)

	// Başarısız dosyaları kaydet veya temizle
	var failedPaths []string
	for _, r := range results {
		if r.Err != nil {
			failedPaths = append(failedPaths, r.RelPath)
		}
	}
	if len(failedPaths) > 0 {
		if err := failed.Save(configDir, srv.Name, failedPaths); err != nil {
			fmt.Printf("  ! Failed listesi kaydedilemedi: %v\n", err)
		} else {
			fmt.Printf("  ! %d başarısız dosya .syncftp/failed/%s.json'a kaydedildi — tekrar denemek için: syncftp sync --retry-failed\n", len(failedPaths), srv.Name)
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
		fmt.Printf("  ! State kaydedilemedi: %v\n", err)
	}

	if len(successFiles) > 0 {
		if relDir, err := release.Create(configDir, srv.Name, successFiles); err != nil {
			fmt.Printf("  ! Release oluşturulamadı: %v\n", err)
		} else {
			fmt.Printf("  Release: %s\n", relDir)
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
func smartFirstSync(srv config.Server, current map[string]string, byRel map[string]scanner.File, st *state.State, cfg *config.Config) []string {
	fmt.Println("  İlk sync: sunucu taranıyor, mevcut dosyalar karşılaştırılıyor...")

	client, err := ftpclient.Connect(srv)
	if err != nil {
		fmt.Printf("  ! Sunucuya bağlanılamadı (%v) — tüm dosyalar yüklenecek\n", err)
		return mapKeys(current)
	}
	defer client.Close()

	remoteFiles, err := client.ListRecursive(srv.RemotePath)
	if err != nil {
		fmt.Printf("  ! Uzak dizin okunamadı (%v) — tüm dosyalar yüklenecek\n", err)
		return mapKeys(current)
	}

	fmt.Printf("  Sunucuda %d dosya bulundu\n", len(remoteFiles))

	var toUpload []string
	alreadySynced := 0

	for rel, hash := range current {
		f := byRel[rel]

		localInfo, err := os.Stat(f.AbsPath)
		if err != nil {
			toUpload = append(toUpload, rel)
			continue
		}
		localSize := uint64(localInfo.Size())

		remoteSize, exists := remoteFiles[rel]
		if !exists || remoteSize != localSize {
			toUpload = append(toUpload, rel)
		} else {
			// Boyut eşleşiyor → zaten güncel, state'e kaydet
			st.Files[rel] = hash
			alreadySynced++
		}
	}

	new := len(toUpload)
	fmt.Printf("  Sonuç: %d güncel (yüklenmeyecek)  |  %d farklı/eksik (yüklenecek)\n", alreadySynced, new)

	return toUpload
}

func serverNames(servers []config.Server) []string {
	names := make([]string, len(servers))
	for i, s := range servers {
		names[i] = s.Name
	}
	return names
}

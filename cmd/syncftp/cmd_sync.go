package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"syncftp/internal/config"
	ftpclient "syncftp/internal/ftp"
	"syncftp/internal/ignore"
	"syncftp/internal/release"
	"syncftp/internal/scanner"
	"syncftp/internal/state"
)

var (
	flagFull    bool
	flagServer  string
	flagDryRun  bool
	flagInclude []string
	flagExclude []string
)

func init() {
	syncCmd.Flags().BoolVar(&flagFull, "full", false, "State'i yoksay, tüm dosyaları yükle")
	syncCmd.Flags().StringVar(&flagServer, "server", "", "Sadece belirtilen sunucuya yükle")
	syncCmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "Ne yükleneceğini göster, gerçekten yükleme")
	syncCmd.Flags().StringArrayVar(&flagInclude, "include", nil, "Sadece bu yol/klasörleri sync et (whitelist; TOML include'u geçersiz kılar)")
	syncCmd.Flags().StringArrayVar(&flagExclude, "exclude", nil, "Bu yol/klasörleri bu sync'ten hariç tut (tek seferlik)")
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
	isFirst := !st.FirstSyncDone || flagFull

	if isFirst {
		label := "İlk sync"
		if flagFull && st.FirstSyncDone {
			label = "Tam sync (--full)"
		}
		fmt.Printf("  %s\n", label)

		if cfg.FirstSync.Full || flagFull {
			toUpload = mapKeys(current)
		} else {
			toUpload = filterByInclude(mapKeys(current), effectiveInclude)
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

	sort.Strings(toUpload)

	if len(toUpload) == 0 {
		fmt.Println("  Değişiklik yok")
		return nil
	}

	fmt.Printf("  %d dosya işlenecek\n", len(toUpload))

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

	client, err := ftpclient.Connect(srv)
	if err != nil {
		return err
	}
	defer client.Close()

	successFiles := make(map[string]string)
	uploaded, skipped, failed := 0, 0, 0

	for _, rel := range toUpload {
		f := byRel[rel]
		if ftpclient.IsProtected(rel, cfg.Sync.Protect) {
			fmt.Printf("    KORUNUYOR  %s\n", rel)
			skipped++
			continue
		}
		if err := client.Upload(f.AbsPath, rel); err != nil {
			fmt.Printf("    ✗ %s: %v\n", rel, err)
			failed++
		} else {
			fmt.Printf("    ✓ %s\n", rel)
			successFiles[rel] = current[rel]
			uploaded++
		}
	}

	fmt.Printf("  Tamamlandı: %d yüklendi, %d korundu, %d hata\n", uploaded, skipped, failed)

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

func serverNames(servers []config.Server) []string {
	names := make([]string, len(servers))
	for i, s := range servers {
		names[i] = s.Name
	}
	return names
}

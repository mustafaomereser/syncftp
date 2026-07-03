package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"

	"syncftp/internal/config"
	"syncftp/internal/ignore"
	"syncftp/internal/lang"
	"syncftp/internal/scanner"
)

const watchDebounce = 800 * time.Millisecond

func init() {
	watchCmd.Flags().Bool("all", false, "Tüm aktif sunucuları izle")
	watchCmd.Flags().String("server", "", "Sadece belirtilen sunucuyu izle")
	rootCmd.AddCommand(watchCmd)
}

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Dosya değişikliklerini izle, değişince otomatik sync yap",
	RunE:  runWatchCmd,
}

func runWatchCmd(cmd *cobra.Command, args []string) error {
	dir, _ := os.Getwd()
	all, _ := cmd.Flags().GetBool("all")
	srvFilter, _ := cmd.Flags().GetString("server")
	return runWatch(dir, all, srvFilter)
}

func runWatch(dir string, all bool, srvFilter string) error {
	cfg, err := config.Load(dir)
	if err != nil {
		return err
	}

	// Sunucu listesi
	var servers []config.Server
	for _, s := range cfg.Servers {
		if !s.Enabled {
			continue
		}
		if srvFilter != "" && s.Name != srvFilter {
			continue
		}
		servers = append(servers, s)
	}
	if len(servers) == 0 {
		fmt.Print(lang.L.WatchNoServers)
		return nil
	}

	// Tek sunucu varsa direkt izle; birden fazlaysa --all yoksa multi-select picker aç
	if !all && srvFilter == "" && len(servers) > 1 {
		items := make([]PickerItem, len(servers))
		for i, s := range servers {
			localDir := filepath.Join(dir, s.EffectiveLocalPath(cfg.Project))
			items[i] = PickerItem{
				Icon:  "🖥",
				Label: s.Name,
				Desc:  localDir,
				Value: s.Name,
			}
		}
		selected, err := RunMultiPicker("Watch — sunucu seç", "Space = seç/bırak  |  Enter = başlat  |  q = iptal", items)
		if err != nil || len(selected) == 0 {
			return err
		}
		selSet := make(map[string]bool, len(selected))
		for _, v := range selected {
			selSet[v] = true
		}
		filtered := servers[:0]
		for _, s := range servers {
			if selSet[s.Name] {
				filtered = append(filtered, s)
			}
		}
		servers = filtered
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	// dizin → o dizini izleyen sunucular (aynı local_path birden fazla sunucuda olabilir)
	dirToServers := make(map[string][]string)
	serverMap := make(map[string]config.Server)

	for _, srv := range servers {
		localDir := filepath.Clean(filepath.Join(dir, srv.EffectiveLocalPath(cfg.Project)))
		alreadyWatched := len(dirToServers[localDir]) > 0
		dirToServers[localDir] = append(dirToServers[localDir], srv.Name)
		serverMap[srv.Name] = srv
		if !alreadyWatched {
			if err := watchAddRecursive(watcher, localDir); err != nil {
				fmt.Printf(lang.L.WatchErrFmt, err)
				continue
			}
		}
		fmt.Printf(lang.L.WatchStartFmt, srv.Name, localDir)
	}
	fmt.Print(lang.L.WatchReady)

	// SIGINT/SIGTERM yakalanınca watch durur, shell'e döner (terminal kapanmaz)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	var (
		timersMu sync.Mutex
		timers   = make(map[string]*time.Timer)
		syncMu   sync.Mutex
	)

	for {
		select {
		case <-sigCh:
			fmt.Println("\nWatch durduruldu.")
			return nil

		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if watchIsSyncftpPath(event.Name) {
				continue
			}

			srvNames := watchResolveServers(event.Name, dirToServers)
			if len(srvNames) == 0 {
				continue
			}

			// Yeni klasör oluşturulduysa watcher'a ekle
			if event.Has(fsnotify.Create) {
				if fi, statErr := os.Stat(event.Name); statErr == nil && fi.IsDir() {
					_ = watchAddRecursive(watcher, event.Name)
				}
			}

			// Debounce: dizini izleyen HER sunucu için timer'ı sıfırla ya da yenisini oluştur
			timersMu.Lock()
			for _, srvName := range srvNames {
				if t, exists := timers[srvName]; exists {
					t.Reset(watchDebounce)
					continue
				}
				sn := srvName
				srv := serverMap[sn]
				timers[sn] = time.AfterFunc(watchDebounce, func() {
					syncMu.Lock()
					defer syncMu.Unlock()
					timersMu.Lock()
					delete(timers, sn)
					timersMu.Unlock()
					watchDoSync(dir, cfg, srv)
				})
			}
			timersMu.Unlock()

		case watchErr, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Printf(lang.L.WatchErrFmt, watchErr)
		}
	}
}

// watchAddRecursive — dizini ve tüm alt dizinlerini watcher'a ekler.
func watchAddRecursive(w *fsnotify.Watcher, root string) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		if watchIsSyncftpPath(path) {
			return filepath.SkipDir
		}
		_ = w.Add(path)
		return nil
	})
}

// watchIsSyncftpPath — .syncftp ve .git dizinlerini/dosyalarını elendirir (segment bazlı).
func watchIsSyncftpPath(p string) bool {
	s := filepath.ToSlash(p)
	for _, seg := range []string{".syncftp", ".git"} {
		if s == seg || strings.HasSuffix(s, "/"+seg) || strings.Contains(s, "/"+seg+"/") || strings.HasPrefix(s, seg+"/") {
			return true
		}
	}
	return false
}

// watchResolveServers — dosya yolunu izleyen TÜM sunucuları döner.
// Aynı dizini paylaşan ve iç içe dizinleri izleyen sunucuların hepsi eşleşir.
// Prefix eşleşmesi ayraç sınırında yapılır ("frontend" ≠ "frontend2").
func watchResolveServers(filePath string, dirToServers map[string][]string) []string {
	clean := filepath.Clean(filePath)
	var out []string
	for d, names := range dirToServers {
		if clean == d || strings.HasPrefix(clean, d+string(filepath.Separator)) {
			out = append(out, names...)
		}
	}
	return out
}

// watchDoSync — debounce süresi dolunca çağrılır; scan + sync.
func watchDoSync(configDir string, cfg *config.Config, srv config.Server) {
	fmt.Printf(lang.L.WatchChangeFmt, srv.Name)

	localDir := filepath.Join(configDir, srv.EffectiveLocalPath(cfg.Project))
	matcher, _ := ignore.Load(localDir, cfg.Sync.IgnoreFiles)
	files, err := scanner.Scan(localDir, matcher)
	if err != nil {
		fmt.Printf(lang.L.WatchErrFmt, err)
		return
	}
	current := make(map[string]string, len(files))
	byKey := make(map[string]scanner.File, len(files))
	for _, f := range files {
		current[f.RelPath] = f.Hash
		byKey[f.RelPath] = f
	}
	syncToServer(configDir, cfg, srv, current, byKey, nil, nil)
}

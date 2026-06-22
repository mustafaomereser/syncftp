package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

	// Tek sunucu varsa direkt izle; birden fazlaysa --all yoksa picker aç
	if !all && srvFilter == "" && len(servers) > 1 {
		items := make([]PickerItem, len(servers)+1)
		for i, s := range servers {
			localDir := filepath.Join(dir, s.EffectiveLocalPath(cfg.Project))
			items[i] = PickerItem{
				Icon:  "🖥",
				Label: s.Name,
				Desc:  localDir,
				Value: s.Name,
			}
		}
		items[len(servers)] = PickerItem{Icon: "⊕", Label: "Tümünü izle", Desc: "all", Value: "__all__"}
		val, err := RunPicker("Watch — sunucu seç", "birden fazla seçmek için 'Tümünü izle'", items)
		if err != nil {
			return err
		}
		if val == "" {
			return nil
		}
		if val != "__all__" {
			filtered := servers[:0]
			for _, s := range servers {
				if s.Name == val {
					filtered = append(filtered, s)
					break
				}
			}
			servers = filtered
		}
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	// dizin → server adı eşlemesi
	dirToServer := make(map[string]string)
	serverMap := make(map[string]config.Server)

	for _, srv := range servers {
		localDir := filepath.Clean(filepath.Join(dir, srv.EffectiveLocalPath(cfg.Project)))
		dirToServer[localDir] = srv.Name
		serverMap[srv.Name] = srv
		if err := watchAddRecursive(watcher, localDir); err != nil {
			fmt.Printf(lang.L.WatchErrFmt, err)
			continue
		}
		fmt.Printf(lang.L.WatchStartFmt, srv.Name, localDir)
	}
	fmt.Print(lang.L.WatchReady)

	var (
		timersMu sync.Mutex
		timers   = make(map[string]*time.Timer)
		syncMu   sync.Mutex // eş zamanlı sync'i önle (aynı server)
	)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if watchIsSyncftpPath(event.Name) {
				continue
			}

			srvName := watchResolveServer(event.Name, dirToServer)
			if srvName == "" {
				continue
			}

			// Yeni klasör oluşturulduysa watcher'a ekle
			if event.Has(fsnotify.Create) {
				if fi, statErr := os.Stat(event.Name); statErr == nil && fi.IsDir() {
					_ = watchAddRecursive(watcher, event.Name)
				}
			}

			// Debounce: timer'ı sıfırla ya da yenisini oluştur
			timersMu.Lock()
			if t, exists := timers[srvName]; exists {
				t.Reset(watchDebounce)
			} else {
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

// watchIsSyncftpPath — .syncftp dizini/dosyalarını elendirir.
func watchIsSyncftpPath(p string) bool {
	return strings.Contains(p, ".syncftp")
}

// watchResolveServer — dosya yolunun hangi server'a ait olduğunu bulur.
func watchResolveServer(filePath string, dirToServer map[string]string) string {
	clean := filepath.Clean(filePath)
	for d, name := range dirToServer {
		if strings.HasPrefix(clean, d) {
			return name
		}
	}
	return ""
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

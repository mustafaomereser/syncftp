package main

import (
	"fmt"
	"io"
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

	// SIGINT/SIGTERM yakalanınca watch durur, shell'e döner (terminal kapanmaz)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)
	stop := make(chan struct{})
	go func() {
		<-sigCh
		close(stop)
	}()
	err = watchLoop(dir, cfg, servers, os.Stdout, stop, nil)
	fmt.Print(lang.L.WatchStopped)
	return err
}

// watchLoop izleme çekirdeği — CLI/shell watch ve GUI'nin arka plan watch'ları ortak kullanır.
// stop kapanınca döner. onSync != nil ise her sync sonrası sonucu ile çağrılır (GUI durum).
func watchLoop(dir string, cfg *config.Config, servers []config.Server, out io.Writer, stop <-chan struct{}, onSync func(serverSyncResult)) error {
	release, err := acquireWatchLock(dir)
	if err != nil {
		return err
	}
	defer release()
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	// Her izlenen dizin bir kez eklenir; o dizini izleyen tüm sunucular aynı
	// ignore matcher'ını kullanır (ignore dosyaları local dizinden okunur).
	var roots []*watchRoot
	rootByDir := make(map[string]*watchRoot)
	workers := make(map[string]*watchWorker)
	var syncMu sync.Mutex // sync'ler serialize — çıktı karışmasın

	defer func() {
		for _, wk := range workers {
			close(wk.trigger)
		}
	}()

	for _, srv := range servers {
		localDir := filepath.Clean(filepath.Join(dir, srv.EffectiveLocalPath(cfg.Project)))
		wr, ok := rootByDir[localDir]
		if !ok {
			matcher, _ := ignore.Load(localDir, cfg.Sync.IgnoreFiles)
			wr = &watchRoot{dir: localDir, matcher: matcher}
			if err := watchAddRecursive(watcher, localDir, matcher); err != nil {
				fmt.Fprintf(out, lang.L.WatchErrFmt, err)
				continue
			}
			rootByDir[localDir] = wr
			roots = append(roots, wr)
		}
		wr.servers = append(wr.servers, srv.Name)
		if _, exists := workers[srv.Name]; !exists {
			wk := &watchWorker{trigger: make(chan struct{}, 1)}
			workers[srv.Name] = wk
			go wk.run(func() {
				syncMu.Lock()
				defer syncMu.Unlock()
				r := watchDoSync(out, dir, cfg, srv)
				if onSync != nil {
					onSync(r)
				}
			})
		}
		fmt.Fprintf(out, lang.L.WatchStartFmt, srv.Name, localDir)
	}
	fmt.Fprint(out, lang.L.WatchReady)

	for {
		select {
		case <-stop:
			return nil

		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if watchIsSyncftpPath(event.Name) || watchIsEditorTemp(event.Name) {
				continue
			}
			for _, wr := range watchResolveRoots(event.Name, roots) {
				rel, relErr := filepath.Rel(wr.dir, event.Name)
				if relErr != nil {
					continue
				}
				rel = filepath.ToSlash(rel)

				// Yeni klasör: ignore edilmiyorsa arka planda ekle (büyük ağaç ana döngüyü bloklamasın)
				if event.Has(fsnotify.Create) {
					if fi, statErr := os.Stat(event.Name); statErr == nil && fi.IsDir() {
						if wr.matcher.MatchDir(rel) {
							continue
						}
						go func(p string, m *ignore.Matcher) { _ = watchAddRecursive(watcher, p, m) }(event.Name, wr.matcher)
					}
				}
				// Ignore edilen dosya/dizin (vendor, node_modules, storage/logs...) sync tetiklemez
				if wr.matcher.Match(rel) || watchUnderIgnoredDir(rel, wr.matcher) {
					continue
				}
				for _, name := range wr.servers {
					workers[name].notify()
				}
			}

		case watchErr, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Fprintf(out, lang.L.WatchErrFmt, watchErr)
		}
	}
}

// watchRoot — izlenen bir yerel dizin ve onu izleyen sunucular.
type watchRoot struct {
	dir     string
	matcher *ignore.Matcher
	servers []string
}

// watchWorker — sunucu başına tek goroutine. trigger 1 kapasiteli: sync sürerken
// gelen olaylar tek bir bekleyen tetikte birleşir → sync bitince bir tur daha
// çalışır, hiçbir değişiklik kaçmaz ve aynı sync iki kez kuyruğa girmez.
type watchWorker struct {
	trigger chan struct{}
}

func (w *watchWorker) notify() {
	select {
	case w.trigger <- struct{}{}:
	default: // zaten bekleyen tetik var
	}
}

func (w *watchWorker) run(doSync func()) {
	for range w.trigger {
		// Debounce: son olaydan sonra watchDebounce kadar sessizlik bekle
		t := time.NewTimer(watchDebounce)
	quiet:
		for {
			select {
			case <-w.trigger:
				if !t.Stop() {
					<-t.C
				}
				t.Reset(watchDebounce)
			case <-t.C:
				break quiet
			}
		}
		doSync()
	}
}

// watchAddRecursive — dizini ve ignore edilmeyen tüm alt dizinlerini watcher'a ekler.
func watchAddRecursive(w *fsnotify.Watcher, root string, matcher *ignore.Matcher) error {
	base := root
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}
		if watchIsSyncftpPath(path) {
			return filepath.SkipDir
		}
		if matcher != nil && path != base {
			if rel, relErr := filepath.Rel(base, path); relErr == nil && matcher.MatchDir(filepath.ToSlash(rel)) {
				return filepath.SkipDir
			}
		}
		_ = w.Add(path)
		return nil
	})
}

// watchUnderIgnoredDir — rel'in üst dizinlerinden biri ignore ediliyor mu
// (alt klasörü sonradan eklenen ignore'lu dizinler için).
func watchUnderIgnoredDir(rel string, m *ignore.Matcher) bool {
	parts := strings.Split(rel, "/")
	for i := 1; i < len(parts); i++ {
		if m.MatchDir(strings.Join(parts[:i], "/")) {
			return true
		}
	}
	return false
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

// watchIsEditorTemp — editörlerin geçici dosyaları (vim swap, ~ yedekleri, JetBrains ___jb_)
// sync tetiklemesin.
func watchIsEditorTemp(p string) bool {
	b := filepath.Base(p)
	return strings.HasSuffix(b, "~") || strings.HasSuffix(b, ".swp") || strings.HasSuffix(b, ".swx") ||
		b == "4913" || strings.Contains(b, "___jb_") || strings.HasPrefix(b, ".#")
}

// watchResolveRoots — dosya yolunu kapsayan TÜM izlenen dizinleri döner
// (iç içe dizinleri izleyen sunucuların hepsi tetiklenir).
// Prefix eşleşmesi ayraç sınırında yapılır ("frontend" ≠ "frontend2").
func watchResolveRoots(filePath string, roots []*watchRoot) []*watchRoot {
	clean := filepath.Clean(filePath)
	var out []*watchRoot
	for _, r := range roots {
		if clean == r.dir || strings.HasPrefix(clean, r.dir+string(filepath.Separator)) {
			out = append(out, r)
		}
	}
	return out
}

// watchDoSync — debounce süresi dolunca çağrılır; scan + sync.
// Modlar sabit (syncOpts{}): CLI flag'lerinden etkilenmez.
func watchDoSync(out io.Writer, configDir string, cfg *config.Config, srv config.Server) serverSyncResult {
	fmt.Fprintf(out, lang.L.WatchChangeFmt, srv.Name)
	return syncServer(out, configDir, cfg, srv, nil, nil, nil, syncOpts{})
}

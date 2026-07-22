package main

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"syncftp/internal/config"
	ftpclient "syncftp/internal/ftp"
	"syncftp/internal/frozen"
	"syncftp/internal/ignore"
	"syncftp/internal/lang"
	"syncftp/internal/scanner"
	"syncftp/internal/state"
	"syncftp/internal/synclog"
)

var (
	calibrateFlagServer string
	calibrateFlagAll    bool
)

func init() {
	calibrateCmd.Flags().StringVarP(&calibrateFlagServer, "server", "s", "", "Sadece bu sunucuyu kalibre et")
	calibrateCmd.Flags().BoolVar(&calibrateFlagAll, "all", false, "Tüm aktif sunucuları kalibre et")
	rootCmd.AddCommand(calibrateCmd)
}

var calibrateCmd = &cobra.Command{
	Use:   "calibrate",
	Short: "Yerel dosyaları FTP boyutlarıyla karşılaştırıp state'i günceller (yükleme yapmaz)",
	RunE:  runCalibrateCmd,
}

func runCalibrateCmd(cmd *cobra.Command, args []string) error {
	dir, _ := os.Getwd()
	cfg, err := config.Load(dir)
	if err != nil {
		return err
	}

	servers := cfg.Servers
	if calibrateFlagServer != "" {
		var found []config.Server
		for _, s := range servers {
			if s.Name == calibrateFlagServer {
				found = append(found, s)
				break
			}
		}
		if len(found) == 0 {
			fmt.Print(lang.L.ResyncNoServers)
			return nil
		}
		servers = found
	} else if !calibrateFlagAll {
		// varsayılan: sadece enabled sunucular
		var enabled []config.Server
		for _, s := range servers {
			if s.Enabled {
				enabled = append(enabled, s)
			}
		}
		servers = enabled
	}

	if len(servers) == 0 {
		fmt.Print(lang.L.ResyncNoServers)
		return nil
	}

	for _, srv := range servers {
		runCalibrateOpts(dir, srv, cfg, calibrateOpts{interactive: true, saveLog: true})
	}
	return nil
}

// countCRLF dosyadaki \r\n çiftlerini sayar (32 KB chunk'larla, bellek verimli).
func countCRLF(path string) uint64 {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()

	buf := make([]byte, 32*1024)
	var count uint64
	var prev byte
	for {
		n, err := f.Read(buf)
		for i := 0; i < n; i++ {
			if prev == '\r' && buf[i] == '\n' {
				count++
			}
			prev = buf[i]
		}
		if err != nil {
			break
		}
	}
	return count
}

// passesFilter sync ile aynı include/exclude mantığını uygular.
func passesFilter(relPath string, include, exclude []string) bool {
	if len(include) > 0 {
		matched := false
		for _, inc := range include {
			inc = strings.TrimSuffix(filepath.ToSlash(inc), "/")
			if relPath == inc || strings.HasPrefix(relPath, inc+"/") {
				matched = true
				break
			}
			if ok, _ := path.Match(inc, relPath); ok {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	for _, exc := range exclude {
		exc = strings.TrimSuffix(filepath.ToSlash(exc), "/")
		if relPath == exc || strings.HasPrefix(relPath, exc+"/") {
			return false
		}
		if ok, _ := path.Match(exc, relPath); ok {
			return false
		}
	}
	return true
}

// runCalibrate yerel dosyaları FTP boyutlarıyla karşılaştırıp eşleşenleri state'e yazar.
// Yükleme yapmaz. Hata olursa sessizce devam eder (state değişmez).
// Sync içindeki otomatik tetikleme bu sürümü kullanır (TUI ve log yok — paralel güvenli).
func runCalibrate(dir string, srv config.Server, cfg *config.Config) {
	runCalibrateOpts(dir, srv, cfg, calibrateOpts{})
}

type calibrateOpts struct {
	interactive bool // listeleme fazında inline TUI: → taranan dosyaları canlı gösterir
	saveLog     bool // özet + taranan dosya listesi synclog'a kaydedilir
}

func runCalibrateOpts(dir string, srv config.Server, cfg *config.Config, opts calibrateOpts) {
	// p hem terminale yazar hem (saveLog açıksa) log tamponuna toplar;
	// \r'li canlı sayaç satırları p'den geçmez, log'a girmez.
	var logBuf strings.Builder
	p := func(format string, a ...any) {
		fmt.Printf(format, a...)
		if opts.saveLog {
			fmt.Fprintf(&logBuf, format, a...)
		}
	}

	localPath := srv.EffectiveLocalPath(cfg.Project)
	localDir := filepath.Join(dir, localPath)

	matcher, _ := ignore.Load(localDir, cfg.Sync.IgnoreFiles)
	files, err := scanner.Scan(localDir, matcher)
	if err != nil {
		return
	}

	// Sync ile aynı include/exclude filtreleri (server > global, exclude additive)
	effectiveInclude := cfg.Sync.Include
	if len(srv.Include) > 0 {
		effectiveInclude = srv.Include
	}
	effectiveExclude := append(append([]string{}, cfg.Sync.Exclude...), srv.Exclude...)

	current := make(map[string]string)
	byKey := make(map[string]scanner.File)
	excluded := 0
	for _, f := range files {
		if !passesFilter(f.RelPath, effectiveInclude, effectiveExclude) {
			excluded++
			continue
		}
		current[f.RelPath] = f.Hash
		byKey[f.RelPath] = f
	}

	// Ignore edilen dosya ve dizinleri say (matcher'ı kullanarak — hız için ignored dizinlere girme)
	var ignoredFiles int
	var ignoredDirs []string
	_ = filepath.WalkDir(localDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(localDir, p)
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil
		}
		if rel == ".syncftp" || strings.HasPrefix(rel, ".syncftp/") ||
			rel == ".git" || strings.HasPrefix(rel, ".git/") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if matcher.Match(rel) {
			if d.IsDir() {
				ignoredDirs = append(ignoredDirs, rel+"/")
				return filepath.SkipDir
			}
			ignoredFiles++
			return nil
		}
		return nil
	})
	ignoredCount := ignoredFiles

	st, _ := state.Load(dir, srv.Name)

	p("  [%s]\n", srv.Name)
	p(lang.L.ResyncLocalFmt, len(current))
	if len(ignoredDirs) > 0 {
		p(lang.L.ResyncIgnoreDirsFmt, len(ignoredDirs), strings.Join(ignoredDirs, ", "))
	}
	if ignoredCount > 0 {
		p(lang.L.ResyncIgnoreFilesFmt, ignoredCount)
	}
	if excluded > 0 {
		p(lang.L.ResyncFilteredFmt, excluded)
	}
	fmt.Print(lang.L.ResyncScanning)

	client, err := ftpclient.Connect(srv)
	if err != nil {
		fmt.Printf(lang.L.ResyncConnErr, err)
		return
	}
	defer client.Close()
	fmt.Print(lang.L.ResyncConnected)

	fmt.Printf("  %s\n", lang.L.ResyncListing)
	var remoteFiles map[string]uint64
	var err2 error
	if opts.interactive {
		var aborted bool
		remoteFiles, aborted, err2 = runListingTUI(client, srv.RemotePath)
		if aborted {
			fmt.Print(lang.L.ResyncListCancelled)
			return
		}
	} else {
		lastList := time.Now()
		remoteFiles, err2 = client.ListRecursiveProgress(srv.RemotePath, func(n int) {
			if time.Since(lastList) >= 80*time.Millisecond {
				fmt.Printf(lang.L.ResyncListProgressFmt, n)
				lastList = time.Now()
			}
		})
		fmt.Print("\r                                        \r")
	}
	if err2 != nil {
		fmt.Printf(lang.L.ResyncListErr, err2)
		return
	}
	p(lang.L.ResyncFoundFmt, len(remoteFiles))

	frozenList, _ := frozen.Load(dir, srv.Name)

	total := len(current)
	done := 0
	matched := 0
	different := 0
	frozenDiff := 0
	lastPrint := time.Now()
	for key, hash := range current {
		done++
		if done == total || time.Since(lastPrint) >= 80*time.Millisecond {
			fmt.Printf(lang.L.ResyncComparingFmt, done, total)
			lastPrint = time.Now()
		}

		f := byKey[key]
		localInfo, err := os.Stat(f.AbsPath)
		if err != nil {
			different++
			continue
		}
		localSize := uint64(localInfo.Size())
		remoteSize, exists := remoteFiles[key]
		if !exists {
			delete(st.Files, key)
			different++
			if frozen.IsFrozen(frozenList, key) {
				frozenDiff++
			}
			continue
		}
		// Boyut doğrudan eşleşiyorsa veya CRLF→LF farkı açıklıyorsa eşleşti say
		crlfCount := countCRLF(f.AbsPath)
		if remoteSize == localSize || (crlfCount > 0 && remoteSize == localSize-crlfCount) {
			st.Files[key] = hash
			matched++
		} else {
			// Boyut uyuşmadı — bir sonraki sync'in yeniden yüklemesi için state'ten sil
			delete(st.Files, key)
			different++
			if frozen.IsFrozen(frozenList, key) {
				frozenDiff++
			}
		}
	}

	p(lang.L.ResyncMatchedFmt, matched, different)
	if frozenDiff > 0 {
		p(lang.L.ResyncFrozenDiffFmt, frozenDiff)
	}

	st.FirstSyncDone = true
	_ = state.Save(dir, st)

	p(lang.L.ResyncDoneFmt, srv.Name)

	if opts.saveLog {
		fmt.Fprintf(&logBuf, lang.L.ResyncLogFilesHeader, len(remoteFiles))
		keys := make([]string, 0, len(remoteFiles))
		for k := range remoteFiles {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(&logBuf, "  %s (%s)\n", k, humanBytes(int64(remoteFiles[k])))
		}
		if lp, lerr := synclog.Save(dir, srv.Name, logBuf.String()); lerr == nil {
			fmt.Printf(lang.L.SyncLogSavedFmt, lp)
		}
	}
}

// ── listeleme TUI (interaktif calibrate) ──────────────────────────────────────

type calListEntryMsg struct {
	count int
	path  string
}

type calListDoneMsg struct {
	files map[string]uint64
	err   error
}

const calListRecentMax = 10

type calListModel struct {
	ch      chan tea.Msg
	count   int
	recent  []string // son bulunan dosyalar (verbose görünümde akar)
	verbose bool
	done    bool
	aborted bool
	files   map[string]uint64
	err     error
}

func (m calListModel) Init() tea.Cmd { return m.wait() }

func (m calListModel) wait() tea.Cmd {
	ch := m.ch
	return func() tea.Msg { return <-ch }
}

func (m calListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case calListEntryMsg:
		m.count = msg.count
		m.recent = append(m.recent, msg.path)
		if len(m.recent) > calListRecentMax {
			m.recent = m.recent[len(m.recent)-calListRecentMax:]
		}
		return m, m.wait()
	case calListDoneMsg:
		m.done = true
		m.files = msg.files
		m.err = msg.err
		return m, tea.Quit
	case tea.KeyMsg:
		s := msg.String()
		for strings.HasPrefix(s, "alt+") {
			s = strings.TrimPrefix(s, "alt+")
		}
		switch s {
		case "right", "l":
			m.verbose = true
		case "left", "h":
			m.verbose = false
		case "ctrl+c", "q":
			m.aborted = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m calListModel) View() string {
	if m.done || m.aborted {
		return "" // inline render temizlenir, akış normal printf'lerle sürer
	}
	var b strings.Builder
	if m.verbose {
		for _, pth := range m.recent {
			b.WriteString(treeGrey.Render("    "+pth) + "\n")
		}
	}
	b.WriteString(fmt.Sprintf(lang.L.ResyncListTUIFmt, m.count))
	return b.String()
}

// runListingTUI listeleme fazını inline (alt-screen'siz) TUI ile çalıştırır:
// canlı sayaç, → ile taranan dosyaların akışı. ctrl+c/q → aborted (calibrate iptal).
func runListingTUI(client *ftpclient.Client, remotePath string) (map[string]uint64, bool, error) {
	ch := make(chan tea.Msg, 1024)
	go func() {
		files, err := client.ListRecursiveProgressPath(remotePath, func(n int, latest string) {
			select {
			case ch <- calListEntryMsg{count: n, path: latest}:
			default: // TUI yetişemezse ara mesaj düşebilir — done asla düşmez
			}
		})
		ch <- calListDoneMsg{files: files, err: err}
	}()

	final, err := tea.NewProgram(calListModel{ch: ch}).Run()
	if err != nil {
		// TUI açılamadı — sayaçsız bekle, done gelince dön
		for {
			if d, ok := (<-ch).(calListDoneMsg); ok {
				return d.files, false, d.err
			}
		}
	}
	fm := final.(calListModel)
	if fm.aborted {
		// Arka plandaki listeleme done gönderene kadar kanalı boşalt (goroutine sızmasın);
		// caller'ın client.Close() çağrısı yürüyen List'i de sonlandırır.
		go func() {
			for {
				if _, ok := (<-ch).(calListDoneMsg); ok {
					return
				}
			}
		}()
		return nil, true, nil
	}
	return fm.files, false, fm.err
}

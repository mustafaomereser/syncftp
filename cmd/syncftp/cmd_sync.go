package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	"syncftp/internal/synclog"
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

	type scanResult struct {
		srv     config.Server
		current map[string]string
		byKey   map[string]scanner.File
		err     error
	}

	// Tarama fazı — sequential (local disk, hızlı)
	scans := make([]scanResult, 0, len(servers))
	for _, srv := range servers {
		localPath := srv.EffectiveLocalPath(cfg.Project)
		localDir := filepath.Join(dir, localPath)
		matcher, merr := ignore.Load(localDir, cfg.Sync.IgnoreFiles)
		if merr != nil {
			scans = append(scans, scanResult{srv: srv, err: merr})
			continue
		}
		files, ferr := scanner.Scan(localDir, matcher)
		if ferr != nil {
			scans = append(scans, scanResult{srv: srv, err: ferr})
			continue
		}
		current := make(map[string]string, len(files))
		byKey := make(map[string]scanner.File, len(files))
		for _, f := range files {
			current[f.RelPath] = f.Hash
			byKey[f.RelPath] = f
		}
		scans = append(scans, scanResult{srv: srv, current: current, byKey: byKey})
	}

	// Upload fazı — parallel (her sunucu kendi FTP pool'unu açar)
	type runResult struct {
		idx    int
		output string
		result serverSyncResult
	}
	runResults := make([]runResult, len(scans))
	var wg sync.WaitGroup

	for i, sc := range scans {
		wg.Add(1)
		go func(idx int, sc scanResult) {
			defer wg.Done()
			var buf strings.Builder

			if sc.err != nil {
				buf.WriteString(fmt.Sprintf("══ %s ══\n", sc.srv.Name))
				buf.WriteString(fmt.Sprintf("  Tarama hatası: %v\n\n", sc.err))
				runResults[idx] = runResult{idx: idx, output: buf.String(), result: serverSyncResult{name: sc.srv.Name, err: sc.err}}
				return
			}

			buf.WriteString(fmt.Sprintf("══ %s (%s) ══\n", sc.srv.Name, sc.srv.Host))
			buf.WriteString(fmt.Sprintf("  Taranıyor: %s → %d dosya\n\n", sc.srv.EffectiveLocalPath(cfg.Project), len(sc.current)))

			result := syncToServerResult(&buf, dir, cfg, sc.srv, sc.current, sc.byKey, flagInclude, flagExclude)
			runResults[idx] = runResult{idx: idx, output: buf.String(), result: result}
		}(i, sc)
	}
	wg.Wait()

	// Çıktıları sıralı yazdır + değişiklik olan sunucular için log kaydet
	allResults := make([]serverSyncResult, len(runResults))
	for _, rr := range runResults {
		fmt.Print(rr.output)
		allResults[rr.idx] = rr.result
		if !flagDryRun && (rr.result.uploaded > 0 || rr.result.failed > 0) {
			if p, err := synclog.Save(dir, rr.result.name, rr.output); err == nil {
				fmt.Printf(lang.L.SyncLogSavedFmt, p)
			}
		}
	}

	// Birden fazla sunucu varsa özet TUI göster
	showSyncSummary(allResults)

	return nil
}

// syncToServer eski signature — tek sunucu veya shell akışı için.
func syncToServer(configDir string, cfg *config.Config, srv config.Server, current map[string]string, byKey map[string]scanner.File, cliInclude, cliExclude []string) error {
	r := syncToServerResult(os.Stdout, configDir, cfg, srv, current, byKey, cliInclude, cliExclude)
	return r.err
}

// syncToServerResult çıktıyı w'ye yazar, sayım bilgilerini döner (parallel sync için).
func syncToServerResult(w io.Writer, configDir string, cfg *config.Config, srv config.Server, current map[string]string, byKey map[string]scanner.File, cliInclude, cliExclude []string) serverSyncResult {
	result := serverSyncResult{name: srv.Name}
	startTime := time.Now()

	// Failsafe: artık diskte olmayan tam dosya yollarını include/exclude/protect listelerinden sil
	{
		lp := srv.EffectiveLocalPath(cfg.Project)
		ld := filepath.Join(configDir, lp)
		dirty := false
		if newL, ch := cleanPatternList(cfg.Sync.Include, ld); ch {
			cfg.Sync.Include = newL
			dirty = true
		}
		if newL, ch := cleanPatternList(cfg.Sync.Exclude, ld); ch {
			cfg.Sync.Exclude = newL
			dirty = true
		}
		if newL, ch := cleanPatternList(cfg.Sync.Protect, ld); ch {
			cfg.Sync.Protect = newL
			dirty = true
		}
		for i := range cfg.Servers {
			if cfg.Servers[i].Name == srv.Name {
				if newL, ch := cleanPatternList(cfg.Servers[i].Include, ld); ch {
					cfg.Servers[i].Include = newL
					dirty = true
				}
				if newL, ch := cleanPatternList(cfg.Servers[i].Exclude, ld); ch {
					cfg.Servers[i].Exclude = newL
					dirty = true
				}
				if newL, ch := cleanPatternList(cfg.Servers[i].Protect, ld); ch {
					cfg.Servers[i].Protect = newL
					dirty = true
				}
				break
			}
		}
		if dirty {
			if saveErr := config.Save(configDir, cfg); saveErr != nil {
				fmt.Fprintf(w, lang.L.SyncConfigSaveErr, saveErr)
			}
		}
	}

	st, err := state.Load(configDir, srv.Name)
	if err != nil {
		result.err = err
		return result
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

	localPath := srv.EffectiveLocalPath(cfg.Project)
	effectiveInclude = normalizePatterns(effectiveInclude, localPath)
	effectiveExclude = normalizePatterns(effectiveExclude, localPath)

	effectiveProtect := append(append([]string{}, cfg.Sync.Protect...), srv.Protect...)
	effectiveProtect = normalizePatterns(effectiveProtect, localPath)

	effectiveBlockExts := append(append([]string{}, cfg.Sync.BlockExtensions...), srv.BlockExtensions...)

	var toUpload []string

	if flagRetryFailed {
		fl, err := failed.Load(configDir, srv.Name)
		if err != nil {
			result.err = fmt.Errorf("failed listesi okunamadı: %w", err)
			return result
		}
		if fl == nil || len(fl.Files) == 0 {
			fmt.Fprintln(w, lang.L.SyncRetryNoFiles)
			return result
		}
		fmt.Fprintf(w, lang.L.SyncRetryModeFmt, len(fl.Files), fl.CreatedAt.Format("2006-01-02 15:04:05"))
		for _, rel := range fl.Files {
			if _, exists := current[rel]; exists {
				toUpload = append(toUpload, rel)
			} else {
				fmt.Fprintf(w, lang.L.SyncRetrySkipFmt, rel)
			}
		}
	} else {
		isFirst := !st.FirstSyncDone || flagFull

		if isFirst {
			if flagFull && st.FirstSyncDone {
				fmt.Fprintln(w, lang.L.SyncFullFlag)
				toUpload = mapKeys(current)
			} else {
				fmt.Fprint(w, lang.L.ResyncAutoMsg)
				runCalibrate(configDir, srv, cfg)
				if reloaded, err := state.Load(configDir, srv.Name); err == nil {
					st = reloaded
				}
				diff := state.Diff(st, current)
				toUpload = append(diff.New, diff.Changed...)
			}
			if len(effectiveInclude) > 0 {
				toUpload = filterByInclude(toUpload, effectiveInclude)
			}
		} else {
			diff := state.Diff(st, current)
			if len(diff.Deleted) > 0 {
				sort.Strings(diff.Deleted)
				fmt.Fprint(w, lang.L.SyncDeletedHeader)
				for _, p := range diff.Deleted {
					fmt.Fprintf(w, "      - %s\n", p)
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
			fmt.Fprintf(w, "  ❄ %d frozen (skipped)\n", frozenSkipped)
			result.frozenSkipped = frozenSkipped
		}
		toUpload = unfrozen
	}

	// Block filter
	if len(effectiveBlockExts) > 0 {
		var blocked int
		toUpload, blocked = filterByBlock(toUpload, effectiveBlockExts)
		if blocked > 0 {
			fmt.Fprintf(w, lang.L.SyncBlockedFmt, blocked)
			result.blocked = blocked
		}
	}

	sort.Strings(toUpload)

	if len(toUpload) == 0 {
		fmt.Fprintln(w, lang.L.SyncNoChange2)
		return result
	}

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
	result.protected = skipped

	total := len(tasks) + skipped
	fmt.Fprintf(w, lang.L.SyncProcessingFmt, total)
	if skipped > 0 {
		fmt.Fprintf(w, lang.L.SyncProtectedFmt, skipped)
	}
	fmt.Fprintln(w, "")

	if flagDryRun {
		for _, rel := range toUpload {
			f := byKey[rel]
			if ftpclient.IsProtected(rel, effectiveProtect) || ftpclient.IsProtected(f.RelPath, effectiveProtect) {
				fmt.Fprintf(w, lang.L.SyncProtectedLabel, rel)
			} else {
				fmt.Fprintf(w, lang.L.SyncUploadLabel, rel)
			}
		}
		return result
	}

	maxConn := srv.MaxConnections
	if maxConn <= 0 {
		maxConn = 1
	}
	maxRetry := srv.MaxRetries
	if maxRetry == 0 {
		maxRetry = 2
	}
	fmt.Fprintf(w, lang.L.SyncPoolFmt, maxConn, maxRetry)

	// Minify / Obfuscate
	if srv.Minify || srv.Obfuscate {
		if tmpDir, tdErr := os.MkdirTemp("", "syncftp-minify-*"); tdErr == nil {
			var minCount, obfCount int
			var tmpFiles []string
			tasks, tmpFiles, minCount, obfCount = ProcessMinify(tasks, srv.Minify, srv.Obfuscate, tmpDir)
			defer func() {
				for _, p := range tmpFiles {
					os.Remove(p)
				}
				os.Remove(tmpDir)
			}()
			if minCount > 0 {
				fmt.Fprintf(w, lang.L.SyncMinifiedFmt, minCount)
			}
			if obfCount > 0 {
				fmt.Fprintf(w, lang.L.SyncObfuscatedFmt, obfCount)
			}
		}
	}

	pool, poolErr := ftpclient.NewPool(srv)
	if poolErr != nil {
		result.err = poolErr
		return result
	}
	defer pool.Close()

	uploadResults := pool.Upload(tasks)

	for _, rel := range toUpload {
		if ftpclient.IsProtected(rel, effectiveProtect) {
			fmt.Fprintf(w, lang.L.SyncProtectedLabel, rel)
		}
	}

	successFiles := make(map[string]string)
	uploaded, failedCount := 0, 0
	for _, r := range uploadResults {
		if r.Err != nil {
			if r.Attempts > 1 {
				fmt.Fprintf(w, lang.L.SyncAttemptsFmt, r.RelPath, r.Attempts, r.Err)
			} else {
				fmt.Fprintf(w, lang.L.SyncUploadErrFmt, r.RelPath, r.Err)
			}
			failedCount++
			result.failedFiles = append(result.failedFiles, r.RelPath)
		} else {
			if r.Attempts > 1 {
				fmt.Fprintf(w, lang.L.SyncAttemptOkFmt, r.RelPath, r.Attempts)
			} else {
				fmt.Fprintf(w, lang.L.SyncUploadOkFmt, r.RelPath)
			}
			successFiles[r.RelPath] = r.Hash
			uploaded++
			result.bytes += r.Size
			result.uploadedFiles = append(result.uploadedFiles, r.RelPath)
		}
	}
	result.uploaded = uploaded
	result.failed = failedCount
	result.duration = time.Since(startTime)

	fmt.Fprintf(w, lang.L.SyncDoneFullFmt, uploaded, skipped, failedCount)
	fmt.Fprintf(w, lang.L.SyncReportFmt+"\n", humanBytes(result.bytes), result.duration.Round(time.Second))

	var failedPaths []string
	for _, r := range uploadResults {
		if r.Err != nil {
			failedPaths = append(failedPaths, r.RelPath)
		}
	}
	if len(failedPaths) > 0 {
		if err := failed.Save(configDir, srv.Name, failedPaths); err != nil {
			fmt.Fprintf(w, lang.L.SyncFailedSaveErr, err)
		} else {
			fmt.Fprintf(w, lang.L.SyncFailedSavedFmt, len(failedPaths), srv.Name)
		}
	} else {
		failed.Clear(configDir, srv.Name)
	}

	for rel, hash := range successFiles {
		st.Files[rel] = hash
	}
	for rel := range st.Files {
		if _, exists := current[rel]; !exists {
			delete(st.Files, rel)
		}
	}
	st.FirstSyncDone = true

	if err := state.Save(configDir, st); err != nil {
		fmt.Fprintf(w, lang.L.SyncStateErr, err)
	}

	if len(successFiles) > 0 {
		if relDir, err := release.Create(configDir, srv.Name, successFiles); err != nil {
			fmt.Fprintf(w, lang.L.SyncReleaseErr, err)
		} else {
			fmt.Fprintf(w, lang.L.SyncReleaseFmt, relDir)
		}
	}

	// Webhook
	webhookURL := srv.Webhook
	if webhookURL == "" {
		webhookURL = cfg.Sync.Webhook
	}
	if webhookURL != "" {
		sendWebhook(webhookURL, srv.Name, uploaded, failedCount, result.uploadedFiles)
	}

	return result
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

// ── Webhook ───────────────────────────────────────────────────────────────────

// sendWebhook senkron HTTP POST atar. Hata terminal'e yazılır, sync kesilmez.
func sendWebhook(webhookURL, serverName string, uploaded, failed int, files []string) {
	if webhookURL == "" {
		return
	}
	payload := fmt.Sprintf(
		`{"server":%q,"uploaded":%d,"failed":%d,"files":%s}`,
		serverName, uploaded, failed, jsonStringArray(files),
	)
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBufferString(payload))
	if err != nil {
		fmt.Printf(lang.L.SyncWebhookErrFmt, err)
		return
	}
	resp.Body.Close()
	fmt.Printf(lang.L.SyncWebhookSentFmt, webhookURL, resp.StatusCode)
}

func jsonStringArray(ss []string) string {
	if len(ss) == 0 {
		return "[]"
	}
	var b strings.Builder
	b.WriteByte('[')
	for i, s := range ss {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b, "%q", s)
	}
	b.WriteByte(']')
	return b.String()
}

// ── Block filter ──────────────────────────────────────────────────────────────

// filterByBlock uzantıya göre engeller, kalan yolları ve engellenen sayısını döner.
// Uzantılar nokta olmadan da girilebilir: "jpg" → ".jpg"
func filterByBlock(paths []string, blockExts []string) (filtered []string, blocked int) {
	if len(blockExts) == 0 {
		return paths, 0
	}
	extSet := make(map[string]bool, len(blockExts))
	for _, e := range blockExts {
		ext := strings.ToLower(e)
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		extSet[ext] = true
	}
	for _, p := range paths {
		ext := strings.ToLower(filepath.Ext(p))
		if extSet[ext] {
			blocked++
		} else {
			filtered = append(filtered, p)
		}
	}
	return filtered, blocked
}

// ── Sync result + summary TUI ─────────────────────────────────────────────────

type serverSyncResult struct {
	name          string
	uploaded      int
	failed        int
	protected     int
	blocked       int
	frozenSkipped int
	bytes         int64
	duration      time.Duration
	err           error
	uploadedFiles []string
	failedFiles   []string
}

// humanBytes byte sayısını okunur biçimde döner: "1.4 MB", "230 KB", "87 B".
func humanBytes(n int64) string {
	switch {
	case n >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(n)/float64(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(n)/float64(1<<10))
	default:
		return fmt.Sprintf("%d B", n)
	}
}

// syncSummaryTUI birden fazla sunucu sync'i sonrası özet gösterir.
type syncSummaryTUI struct {
	results  []serverSyncResult
	cursor   int
	expanded int // -1 = liste, >=0 = detay
	scroll   int
	width    int
	height   int
}

var (
	ssTitle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5")).Padding(0, 1)
	ssCursor = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	ssName   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	ssOk     = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	ssErr    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	ssCount  = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	ssHint   = lipgloss.NewStyle().Faint(true)
	ssDiv    = lipgloss.NewStyle().Faint(true)
	ssUp     = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	ssFail   = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	ssBlock  = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
)

func (m syncSummaryTUI) Init() tea.Cmd { return nil }

func (m syncSummaryTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		if m.expanded >= 0 {
			switch msg.String() {
			case "q", "Q", "esc", "left", "h":
				m.expanded = -1
				m.scroll = 0
			case "up", "k":
				if m.scroll > 0 {
					m.scroll--
				}
			case "down", "j":
				m.scroll++
			case "g":
				m.scroll = 0
			}
		} else {
			switch msg.String() {
			case "q", "Q", "esc", "ctrl+c":
				return m, tea.Quit
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.results)-1 {
					m.cursor++
				}
			case "right", "l", "enter":
				m.expanded = m.cursor
				m.scroll = 0
			}
		}
	}
	return m, nil
}

func (m syncSummaryTUI) View() string {
	w := m.width
	if w < 60 {
		w = 80
	}
	h := m.height
	if h < 10 {
		h = 24
	}
	div := ssDiv.Render(strings.Repeat("─", w))

	var b strings.Builder
	b.WriteString("\n")

	if m.expanded >= 0 && m.expanded < len(m.results) {
		r := m.results[m.expanded]
		b.WriteString(ssTitle.Render("  "+r.name) + "\n")
		b.WriteString(ssHint.Render("  ←/Esc = geri  |  ↑↓/g = kaydır  |  q = çık") + "\n")
		b.WriteString(div + "\n")

		var lines []string
		for _, f := range r.uploadedFiles {
			lines = append(lines, ssUp.Render("  ✓ ")+f)
		}
		for _, f := range r.failedFiles {
			lines = append(lines, ssFail.Render("  ✗ ")+f)
		}
		if len(lines) == 0 {
			lines = append(lines, ssOk.Render("  ✓ up to date / no changes"))
		}

		visRows := h - 6
		if visRows < 1 {
			visRows = 1
		}
		scroll := m.scroll
		if scroll > len(lines)-visRows {
			scroll = len(lines) - visRows
		}
		if scroll < 0 {
			scroll = 0
		}
		end := scroll + visRows
		if end > len(lines) {
			end = len(lines)
		}
		for _, l := range lines[scroll:end] {
			b.WriteString(l + "\n")
		}
		b.WriteString(div + "\n")
		if len(lines) > visRows {
			b.WriteString(ssHint.Render(fmt.Sprintf("  %d/%d", scroll+1, len(lines))) + "\n")
		}
	} else {
		b.WriteString(ssTitle.Render(lang.L.SyncSummaryTitle) + "\n")
		b.WriteString(ssHint.Render(lang.L.SyncSummaryNavHint) + "\n")
		b.WriteString(div + "\n\n")

		for i, r := range m.results {
			var info string
			if r.err != nil {
				info = ssErr.Render(fmt.Sprintf(lang.L.SyncSummaryErrFmt, r.err))
			} else if r.uploaded == 0 && r.failed == 0 {
				info = ssOk.Render(lang.L.SyncSummaryOk)
			} else {
				info = ssCount.Render(fmt.Sprintf("↑%d", r.uploaded))
				if r.failed > 0 {
					info += "  " + ssFail.Render(fmt.Sprintf("✗%d", r.failed))
				}
				if r.blocked > 0 {
					info += "  " + ssBlock.Render(fmt.Sprintf("⊘%d", r.blocked))
				}
				if r.protected > 0 {
					info += "  " + ssHint.Render(fmt.Sprintf("🔒%d", r.protected))
				}
				if r.bytes > 0 {
					info += "  " + ssHint.Render(humanBytes(r.bytes)+" · "+r.duration.Round(time.Second).String())
				}
				info += ssHint.Render("  →")
			}
			if i == m.cursor {
				b.WriteString(ssCursor.Render("▶ ") + ssName.Render(r.name) + "  " + info + "\n")
			} else {
				b.WriteString("  " + ssName.Render(r.name) + "  " + info + "\n")
			}
		}
		b.WriteString("\n" + div + "\n")
	}
	return b.String()
}

func showSyncSummary(results []serverSyncResult) {
	if len(results) <= 1 {
		return
	}
	m := syncSummaryTUI{results: results, expanded: -1}
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, _ = p.Run()
}

func serverNames(servers []config.Server) []string {
	names := make([]string, len(servers))
	for i, s := range servers {
		names[i] = s.Name
	}
	return names
}

// isExactPath returns true if p looks like an exact file path (no wildcards, no trailing slash).
func isExactPath(p string) bool {
	return !strings.ContainsAny(p, "*?{") && !strings.HasSuffix(p, "/")
}

// cleanPatternList removes exact file paths from list that no longer exist on disk.
// Directory patterns (trailing /) and glob patterns (* ? {) are preserved unchanged.
func cleanPatternList(list []string, localDir string) ([]string, bool) {
	if len(list) == 0 {
		return list, false
	}
	var out []string
	changed := false
	for _, p := range list {
		if isExactPath(p) {
			absPath := filepath.Join(localDir, filepath.FromSlash(p))
			if _, err := os.Stat(absPath); os.IsNotExist(err) {
				fmt.Printf(lang.L.SyncCleanupRemovedFmt, p)
				changed = true
				continue
			}
		}
		out = append(out, p)
	}
	return out, changed
}

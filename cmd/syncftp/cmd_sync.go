package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
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
	"syncftp/internal/lang"
	"syncftp/internal/pattern"
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
			return fmt.Errorf(lang.L.ServerNotFoundFmt+" (%v)", flagServer, serverNames(servers))
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

	// Her sunucu paralel (kendi taraması + FTP havuzu); çıktılar sırayla yazılır
	outputs := make([]string, len(servers))
	allResults := make([]serverSyncResult, len(servers))
	var wg sync.WaitGroup
	for i, srv := range servers {
		wg.Add(1)
		go func(idx int, srv config.Server) {
			defer wg.Done()
			var buf strings.Builder
			allResults[idx] = syncServer(&buf, dir, cfg, srv, flagInclude, flagExclude, nil, cliSyncOpts())
			outputs[idx] = buf.String()
		}(i, srv)
	}
	wg.Wait()
	for i, out := range outputs {
		fmt.Print(out)
		if p := allResults[i].logPath; p != "" {
			fmt.Printf(lang.L.SyncLogSavedFmt, p)
		}
	}

	// Birden fazla sunucu varsa özet TUI göster
	showSyncSummary(allResults)

	// Hata alan sunucu varsa çıkış kodu 1 (CI/script'ler başarısızlığı görebilsin)
	failedSrv := 0
	for _, r := range allResults {
		if r.err != nil {
			failedSrv++
		}
	}
	if failedSrv > 0 {
		cmd.SilenceUsage = true
		return fmt.Errorf(lang.L.SyncServersFailedFmt, failedSrv, len(allResults))
	}
	return nil
}

// syncOpts — tek bir sync çağrısının modları. Paket seviyesindeki CLI flag'lerinden
// bağımsızdır: GUI/serve/watch eş zamanlı farklı modlarla sync çalıştırabilir.
type syncOpts struct {
	full        bool
	dryRun      bool
	retryFailed bool
	// Shell'in canlı ilerleme ekranı için: yükleme başlamadan hemen önce toplam görev sayısı,
	// sonra her dosyanın sonucu (onResult verilirse sonuçlar w'ye ayrıca yazılmaz).
	onUploadStart func(total int)
	onResult      func(ftpclient.UploadResult)
}

// syncServer tarama + sync + log — TÜM sync girişlerinin ortak sarmalayıcısı (CLI sync / milestone
// sync, shell, panel işleri, watch, serve API, webhook). Değişiklik olan (dry-run olmayan) turun
// çıktısı .syncftp/logs/<sunucu>/ altına kaydedilir; yol result.logPath'te döner.
func syncServer(w io.Writer, configDir string, cfg *config.Config, srv config.Server, cliInclude, cliExclude []string, since *time.Time, opts syncOpts) serverSyncResult {
	// Aynı proje+sunucuya eş zamanlı iki sync (watch + panel işi + serve + deploy…) state
	// yükle→yükle→kaydet adımlarını birbirine karıştırmasın
	defer lockServer(configDir, srv.Name)()
	var buf strings.Builder
	out := io.MultiWriter(w, &buf)
	fmt.Fprintf(out, "══ %s (%s://%s) ══\n", srv.Name, srv.ProtocolName(), srv.Host)
	current, byKey, err := scanForServer(configDir, cfg, srv)
	if err != nil {
		fmt.Fprintf(out, "  ✗ %v\n\n", err)
		return serverSyncResult{name: srv.Name, err: err}
	}
	fmt.Fprintf(out, lang.L.GuiScannedFmt, srv.EffectiveLocalPath(cfg.Project), len(current))
	r := syncToServerOpts(out, configDir, cfg, srv, current, byKey, cliInclude, cliExclude, since, opts)
	if !opts.dryRun && (r.uploaded > 0 || r.failed > 0) {
		if p, lerr := synclog.Save(configDir, srv.Name, buf.String()); lerr == nil {
			r.logPath = p
		}
	}
	return r
}

// cliSyncOpts CLI flag'lerinden syncOpts üretir.
func cliSyncOpts() syncOpts {
	return syncOpts{full: flagFull, dryRun: flagDryRun, retryFailed: flagRetryFailed}
}

// syncToServerOpts sync çekirdeği — modlar opts'tan gelir.
// since != nil → milestone modu: state diff yerine mtime >= since olan dosyalar yüklenir.
func syncToServerOpts(w io.Writer, configDir string, cfg *config.Config, srv config.Server, current map[string]string, byKey map[string]scanner.File, cliInclude, cliExclude []string, since *time.Time, opts syncOpts) serverSyncResult {
	result := serverSyncResult{name: srv.Name}
	startTime := time.Now()

	// Failsafe: artık diskte olmayan tam dosya yollarını include/exclude listelerinden sil
	// (taze config üzerinde, kilitli — failsafeCleanup). Dry-run config'e dokunmaz.
	if !opts.dryRun {
		failsafeCleanup(w, configDir, cfg, srv.Name)
	}

	st, err := state.Load(configDir, srv.Name)
	if err != nil {
		fmt.Fprintf(w, lang.L.SyncServerErrFmt, err)
		result.err = err
		return result
	}

	// Etkin filtreler (cmd_plan.go — status ekranlarıyla ortak)
	flt := effectiveFilters(cfg, srv, cliInclude, cliExclude)

	var toUpload []string

	if since != nil {
		fmt.Fprintf(w, lang.L.MilestoneSinceFmt, since.Format("2006-01-02 15:04:05"))
		for rel := range current {
			// mtime tarama anında zaten alındı (scanner.File.ModTime) — dosya başına ek Stat yok
			if !byKey[rel].ModTime.Before(*since) {
				toUpload = append(toUpload, rel)
			}
		}
		toUpload = flt.pick(toUpload)
	} else if opts.retryFailed {
		fl, err := failed.Load(configDir, srv.Name)
		if err != nil {
			result.err = fmt.Errorf(lang.L.SyncFailedListReadFmt, err)
			fmt.Fprintf(w, lang.L.SyncServerErrFmt, result.err)
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
		isFirst := !st.FirstSyncDone || opts.full

		if isFirst {
			if opts.full && st.FirstSyncDone {
				fmt.Fprintln(w, lang.L.SyncFullFlag)
				toUpload = mapKeys(current)
			} else {
				fmt.Fprint(w, lang.L.ResyncAutoMsg)
				// Calibrate başarısızsa (bağlantı / listeleme hatası) devam ETME: FirstSyncDone=false
				// ile her dosya yüklenir (ve yedeklenir) — bu sunucunun sync'i iptal
				if calErr := runCalibrateOpts(configDir, srv, cfg, calibrateOpts{out: w}); calErr != nil {
					fmt.Fprintf(w, lang.L.SyncCalibrateAbortFmt, calErr)
					result.err = calErr
					return result
				}
				if reloaded, err := state.Load(configDir, srv.Name); err == nil {
					st = reloaded
				}
				diff := state.Diff(st, current)
				toUpload = append(diff.New, diff.Changed...)
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
		}
		toUpload = flt.pick(toUpload)
	}

	// Milestone / frozen / block / protect — ortak plan (milestone sync ve retry-failed kalıcı filtreden muaf)
	plan := planUploads(configDir, cfg, srv, toUpload, flt, since == nil && !opts.retryFailed)
	if plan.msDate != nil {
		fmt.Fprintf(w, lang.L.MilestoneFilterFmt, plan.msDate.Format("2006-01-02 15:04"), plan.msDropped)
	}
	if n := len(plan.frozen); n > 0 {
		fmt.Fprintf(w, lang.L.SyncFrozenSkippedFmt, n)
		result.frozenSkipped = n
	}
	if n := len(plan.blocked); n > 0 {
		fmt.Fprintf(w, lang.L.SyncBlockedFmt, n)
		result.blocked = n
	}
	if len(plan.upload)+len(plan.protected) == 0 {
		fmt.Fprintln(w, lang.L.SyncNoChange2)
		return result
	}

	tasks := make([]ftpclient.UploadTask, 0, len(plan.upload))
	for _, rel := range plan.upload {
		tasks = append(tasks, ftpclient.UploadTask{LocalPath: byKey[rel].AbsPath, RelPath: rel, Hash: current[rel]})
	}
	skipped := len(plan.protected)
	result.protected = skipped
	// Aşağıdaki dry-run / protect listeleri: yüklenecekler + korunanlar
	toUpload = append(append([]string{}, plan.upload...), plan.protected...)
	sort.Strings(toUpload)

	total := len(tasks) + skipped
	fmt.Fprintf(w, lang.L.SyncProcessingFmt, total)
	if skipped > 0 {
		fmt.Fprintf(w, lang.L.SyncProtectedFmt, skipped)
	}
	fmt.Fprintln(w, "")

	if opts.dryRun {
		for _, rel := range toUpload {
			f := byKey[rel]
			if ftpclient.IsProtected(rel, flt.protect) || ftpclient.IsProtected(f.RelPath, flt.protect) {
				fmt.Fprintf(w, lang.L.SyncProtectedLabel, rel)
			} else {
				fmt.Fprintf(w, lang.L.SyncUploadLabel, rel)
			}
		}
		return result
	}

	// Pre-sync hook'ları (global + sunucu); hata → bu sunucunun sync'i iptal
	preHooks := append(append([]string{}, cfg.Sync.PreSync...), srv.PreSync...)
	if len(preHooks) > 0 {
		if hookErr := runHooks(w, configDir, preHooks, hookEnv(srv.Name, 0, 0)); hookErr != nil {
			fmt.Fprintf(w, lang.L.SyncHookAbortFmt, hookErr)
			result.err = hookErr
			return result
		}
	}
	// Uzak pre-sync (SFTP/SSH): hata → sync iptal
	if hookErr := runRemoteHooks(w, srv, srv.RemotePreSync, 0, 0); hookErr != nil {
		fmt.Fprintf(w, lang.L.SyncHookAbortFmt, hookErr)
		result.err = hookErr
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

	for _, rel := range toUpload {
		if ftpclient.IsProtected(rel, flt.protect) {
			fmt.Fprintf(w, lang.L.SyncProtectedLabel, rel)
		}
	}

	if opts.onUploadStart != nil {
		opts.onUploadStart(len(tasks))
	}
	outcome, sessErr := runUploadSession(w, configDir, cfg, srv, tasks, st, func(r ftpclient.UploadResult) {
		if opts.onResult != nil {
			opts.onResult(r)
			return
		}
		printUploadResult(w, r)
	})
	if sessErr != nil { // bağlantı kurulamadı vb. — eskiden tek sunucuda hiçbir yere yazılmıyordu
		fmt.Fprintf(w, lang.L.SyncServerErrFmt, sessErr)
		result.err = sessErr
		return result
	}

	uploaded, failedCount := 0, 0
	var failedPaths []string
	for _, r := range outcome.results {
		if r.Err != nil {
			failedCount++
			result.failedFiles = append(result.failedFiles, r.RelPath)
			failedPaths = append(failedPaths, r.RelPath)
		} else {
			uploaded++
			result.uploadedFiles = append(result.uploadedFiles, r.RelPath)
		}
	}
	result.bytes = outcome.bytes
	result.uploaded = uploaded
	result.failed = failedCount
	result.duration = time.Since(startTime)
	result.releaseID = outcome.releaseID

	fmt.Fprintf(w, lang.L.SyncDoneFullFmt, uploaded, skipped, failedCount)
	fmt.Fprintf(w, lang.L.SyncReportFmt+"\n", humanBytes(result.bytes), result.duration.Round(time.Second))

	if len(failedPaths) > 0 {
		if err := failed.Save(configDir, srv.Name, failedPaths); err != nil {
			fmt.Fprintf(w, lang.L.SyncFailedSaveErr, err)
		} else {
			fmt.Fprintf(w, lang.L.SyncFailedSavedFmt, len(failedPaths), srv.Name)
		}
	} else {
		failed.Clear(configDir, srv.Name)
	}

	for rel := range st.Files {
		if _, exists := current[rel]; !exists {
			delete(st.Files, rel)
		}
	}
	// Milestone modu ilk sync sayılmaz — auto-calibrate hakkını tüketme
	if since == nil {
		st.FirstSyncDone = true
	}

	if err := state.Save(configDir, st); err != nil {
		fmt.Fprintf(w, lang.L.SyncStateErr, err)
	}

	if outcome.releaseDir != "" {
		fmt.Fprintf(w, lang.L.SyncReleaseFmt, outcome.releaseDir)
	}

	// Uzak post-sync (SFTP/SSH): migrate, cache clear, restart… — hata sync sonucunu değiştirmez
	if hookErr := runRemoteHooks(w, srv, srv.RemotePostSync, uploaded, failedCount); hookErr != nil {
		fmt.Fprintf(w, lang.L.SyncHookErrFmt, hookErr)
	}

	// Webhook
	webhookURL := srv.Webhook
	if webhookURL == "" {
		webhookURL = cfg.Sync.Webhook
	}
	if webhookURL != "" {
		sendWebhook(w, webhookURL, srv.Name, uploaded, failedCount, result.uploadedFiles)
	}

	// Post-sync hook'ları (global + sunucu); hata sync sonucunu etkilemez
	postHooks := append(append([]string{}, cfg.Sync.PostSync...), srv.PostSync...)
	if len(postHooks) > 0 {
		if hookErr := runHooks(w, configDir, postHooks, hookEnv(srv.Name, uploaded, failedCount)); hookErr != nil {
			fmt.Fprintf(w, lang.L.SyncHookErrFmt, hookErr)
		}
	}

	return result
}

// filterByInclude include kalıplarından birine uyan yollar (internal/pattern; boş liste = hepsi).
func filterByInclude(paths, include []string) []string {
	if len(include) == 0 {
		return paths
	}
	var out []string
	for _, p := range paths {
		if pattern.Any(p, include) {
			out = append(out, p)
		}
	}
	return out
}

// filterByExclude exclude kalıplarına uyan yolları çıkarır (internal/pattern).
func filterByExclude(paths, exclude []string) []string {
	if len(exclude) == 0 {
		return paths
	}
	var out []string
	for _, p := range paths {
		if !pattern.Any(p, exclude) {
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

// webhookClient — timeout'lu: cevap vermeyen webhook sunucusu sync'i (ve watch'ı) kilitlemesin.
var webhookClient = &http.Client{Timeout: 15 * time.Second}

// sendWebhook senkron HTTP POST atar. Hata w'ye yazılır, sync kesilmez.
func sendWebhook(w io.Writer, webhookURL, serverName string, uploaded, failed int, files []string) {
	if webhookURL == "" {
		return
	}
	payload := webhookPayload(webhookURL, serverName, uploaded, failed, files)
	resp, err := webhookClient.Post(webhookURL, "application/json", bytes.NewBufferString(payload))
	if err != nil {
		fmt.Fprintf(w, lang.L.SyncWebhookErrFmt, err)
		return
	}
	resp.Body.Close()
	fmt.Fprintf(w, lang.L.SyncWebhookSentFmt, webhookURL, resp.StatusCode)
}

// webhookPayload URL'e göre biçim seçer: Slack ({"text"}), Discord ({"content"}),
// diğerleri syncFTP JSON'u ({"server","uploaded","failed","files"}).
func webhookPayload(webhookURL, serverName string, uploaded, failed int, files []string) string {
	summary := fmt.Sprintf("syncFTP · %s — ↑ %d, ✗ %d", serverName, uploaded, failed)
	if len(files) > 0 {
		shown := files
		if len(shown) > 10 {
			shown = shown[:10]
		}
		summary += "\n" + strings.Join(shown, "\n")
		if len(files) > 10 {
			summary += fmt.Sprintf("\n… +%d", len(files)-10)
		}
	}
	u := strings.ToLower(webhookURL)
	switch {
	case strings.Contains(u, "hooks.slack.com"):
		return fmt.Sprintf(`{"text":%s}`, jsonQuote(summary))
	case strings.Contains(u, "discord.com/api/webhooks") || strings.Contains(u, "discordapp.com/api/webhooks"):
		return fmt.Sprintf(`{"content":%s}`, jsonQuote(summary))
	}
	return fmt.Sprintf(`{"server":%q,"uploaded":%d,"failed":%d,"files":%s}`, serverName, uploaded, failed, jsonStringArray(files))
}

func jsonQuote(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
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
	releaseID     string
	uploadedFiles []string
	failedFiles   []string
	logPath       string // syncServer kaydettiyse
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

// serverLocks proje dizini + sunucu adı → *sync.Mutex (süreç geneli).
var serverLocks sync.Map

// lockServer proje+sunucu kilidini alır, bırakma fonksiyonunu döner. Anahtar:
// filepath.Clean(mutlak dizin) + "|" + sunucu (Windows'ta yol büyük/küçük harf duyarsız).
// Kilit içinde aynı kilidi tekrar isteyen çağrı (syncServer → saveRemoteEdit vb.) YOKTUR — öyle kalmalı.
func lockServer(dir, server string) func() {
	abs, err := filepath.Abs(dir)
	if err != nil {
		abs = dir
	}
	key := filepath.Clean(abs)
	if runtime.GOOS == "windows" {
		key = strings.ToLower(key)
	}
	m, _ := serverLocks.LoadOrStore(key+"|"+server, &sync.Mutex{})
	mu := m.(*sync.Mutex)
	mu.Lock()
	return mu.Unlock
}

// isExactPath returns true if p looks like an exact file path (no wildcards, no trailing slash).
func isExactPath(p string) bool { return pattern.Exact(p) }

// failsafeMu failsafe temizliğinin config oku→yaz adımını süreç genelinde serileştirir.
var failsafeMu sync.Mutex

// failsafeCleanup diskte artık olmayan tam yolları include/exclude listelerinden çıkarır.
// Güvenlik kuralları:
//   - protect listelerine ASLA dokunulmaz (.env bir sunucunun dizininde yok diye korumasız kalmasın)
//   - include listesi temizlikle boşaltılmaz (boş include = her şey yüklenir; build sırasında dosya
//     geçici olarak yok olabilir)
//   - sunucu listeleri o sunucunun yerel dizinine göre; global listeler yalnızca yol HİÇBİR
//     sunucunun yerel dizininde yoksa temizlenir
//   - değişiklik çağıranın cfg'sine değil, kilit altında diskten taze okunan config'e yazılır:
//     paralel sync goroutine'leri aynı *Config'i paylaşır, uzun yaşayan watch/shell kopyası da
//     daha yeni düzenlemelerin üzerine eski config yazmamalı. Çıkarılan yollar zaten hiçbir dosyaya
//     uymadığından bu turun filtre sonucu değişmez → bellekteki kopyayı güncellemeye gerek yok.
func failsafeCleanup(w io.Writer, configDir string, cfg *config.Config, serverName string) {
	// Ön kontrol çağıranın config'inin kopyası üzerinde (salt okunur): çoğu sync'te temizlenecek
	// bir şey yoktur → config.Load (anahtarlık okuması) hiç yapılmaz
	cp := *cfg
	cp.Servers = append([]config.Server(nil), cfg.Servers...)
	if !applyFailsafeCleanup(io.Discard, configDir, &cp, serverName) {
		return
	}

	failsafeMu.Lock()
	defer failsafeMu.Unlock()
	fresh, err := config.Load(configDir)
	if err != nil {
		return
	}
	if !applyFailsafeCleanup(w, configDir, fresh, serverName) {
		return
	}
	if saveErr := config.Save(configDir, fresh); saveErr != nil {
		fmt.Fprintf(w, lang.L.SyncConfigSaveErr, saveErr)
	}
}

// applyFailsafeCleanup c üzerinde temizliği yapar (yalnızca alanlara yeni dilim atar, eski
// dilimleri değiştirmez); değişiklik olduysa true.
func applyFailsafeCleanup(w io.Writer, configDir string, c *config.Config, serverName string) bool {
	idx := -1
	for i := range c.Servers {
		if c.Servers[i].Name == serverName {
			idx = i
			break
		}
	}
	if idx < 0 {
		return false
	}
	self := c.Servers[idx]
	own := func(p string) bool { return patternExistsFor(configDir, c.Project, self, p) }
	anyServer := func(p string) bool {
		for _, s := range c.Servers {
			if patternExistsFor(configDir, c.Project, s, p) {
				return true
			}
		}
		return false
	}
	dirty := false
	if l, ch := cleanPatternList(w, c.Sync.Include, anyServer, true); ch {
		c.Sync.Include, dirty = l, true
	}
	if l, ch := cleanPatternList(w, c.Sync.Exclude, anyServer, false); ch {
		c.Sync.Exclude, dirty = l, true
	}
	if l, ch := cleanPatternList(w, self.Include, own, true); ch {
		c.Servers[idx].Include, dirty = l, true
	}
	if l, ch := cleanPatternList(w, self.Exclude, own, false); ch {
		c.Servers[idx].Exclude, dirty = l, true
	}
	return dirty
}

// patternExistsFor kalıp, sunucunun yerel dizininde (local_path önekli eski kayıtlar da
// normalize edilerek) var mı. Stat'in "yok" dışındaki hataları var sayılır (temkinli taraf).
func patternExistsFor(configDir string, proj config.Project, s config.Server, p string) bool {
	lp := s.EffectiveLocalPath(proj)
	np := normalizePatterns([]string{p}, lp)[0]
	_, err := os.Stat(filepath.Join(configDir, lp, filepath.FromSlash(strings.TrimPrefix(np, "/"))))
	return !os.IsNotExist(err)
}

// cleanPatternList listeden exists'e göre artık olmayan tam yolları çıkarır; klasör (sonda /) ve
// jokerli kalıplar olduğu gibi kalır. keepNonEmpty → temizlik listeyi boşaltacaksa hiç dokunulmaz.
// Girdi dilimi değiştirilmez (yeni dilim döner).
func cleanPatternList(w io.Writer, list []string, exists func(string) bool, keepNonEmpty bool) ([]string, bool) {
	if len(list) == 0 {
		return list, false
	}
	var out, removed []string
	for _, p := range list {
		if isExactPath(p) && strings.TrimSpace(p) != "" && !exists(p) {
			removed = append(removed, p)
			continue
		}
		out = append(out, p)
	}
	if len(removed) == 0 {
		return list, false
	}
	if keepNonEmpty && len(out) == 0 {
		fmt.Fprint(w, lang.L.SyncCleanupIncludeKept)
		return list, false
	}
	for _, p := range removed {
		fmt.Fprintf(w, lang.L.SyncCleanupRemovedFmt, p)
	}
	return out, true
}

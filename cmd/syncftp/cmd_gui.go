package main

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"crypto/subtle"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spf13/cobra"

	"syncftp/internal/lang"
	"syncftp/internal/projects"
	"syncftp/internal/synclog"
)

// ══════════════════════════════════════════════════════════════════════════════
// syncftp gui — masaüstü yönetim paneli
//
// Mimari (launcher örneğiyle aynı desen): Go süreci 127.0.0.1'de rastgele portta
// gömülü web arayüzünü sunar; Edge/Chrome "--app" modu adres çubuksuz ayrı bir
// pencere açar. Açılış yolları:
//   - exe'ye çift tık (Explorer)  → konsol gizlenir, GUI bu süreçte çalışır
//   - `syncftp gui`               → GUI ayrık süreçte başlar, terminal hemen döner
//   - `syncftp` (terminalden)     → her zamanki shell
// Tek örnek: <UserConfigDir>/syncftp/gui.json'daki port canlıysa yeni sunucu açılmaz,
// mevcut pencere öne getirilir (yoksa yeni pencere açılır).
// Yaşam süresi: pencere kapanınca heartbeat kesilir; çalışan job/watch yoksa süreç kapanır.
// ══════════════════════════════════════════════════════════════════════════════

//go:embed web
var guiWeb embed.FS

// Panel penceresinin başlığı "syncFTP — <klasör>" (app.js meta'dan kurar) → her klasörün paneli ayrı bulunur.
func guiWindowTitle() string { return "syncFTP — " + guiStartDir() }

var (
	guiFlagForeground bool
	guiFlagNoWindow   bool
	guiFlagBackground bool
)

func init() {
	guiCmd.Flags().BoolVar(&guiFlagForeground, "foreground", false, "Paneli bu süreçte çalıştır (terminal bloklanır)")
	guiCmd.Flags().BoolVar(&guiFlagNoWindow, "no-window", false, "Pencere açma, giriş adresini yazdır (uzak/headless kullanım; --foreground ile)")
	guiCmd.Flags().BoolVar(&guiFlagBackground, "background", false, "Yalnızca tray'de başla, pencere açma (Windows açılışında başlat bunu kullanır)")
	rootCmd.AddCommand(guiCmd)
}

var guiCmd = &cobra.Command{
	Use:   "gui",
	Short: "Masaüstü yönetim panelini açar (projeler, sync, rollback, watch, ayarlar)",
	Long: `Tüm kayıtlı projeleri tek pencereden yönetmek için grafik arayüz.

Exe'ye çift tıklamak da paneli açar. Terminalden argümansız 'syncftp' her zamanki
shell'i başlatır; 'syncftp gui' paneli ayrı süreçte açar ve terminali serbest bırakır.

Panel yalnızca 127.0.0.1'den, oturuma özel token ile erişilebilir.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if guiFlagForeground {
			return runGUIForeground()
		}
		if err := launchGUIDetached(); err != nil {
			return err
		}
		fmt.Print(lang.L.GuiLaunched)
		return nil
	},
}

// ── Tek örnek ─────────────────────────────────────────────────────────────────

type guiInstance struct {
	Port  int    `json:"port"`
	Token string `json:"token"`
	PID   int    `json:"pid"`
}

func guiStateDir() string {
	base, err := os.UserConfigDir()
	if err != nil {
		base = os.TempDir()
	}
	return filepath.Join(base, "syncftp")
}

// Tek örnek klasör başına: panel CLI gibi açıldığı klasörü esas alır, başka klasörden açılan exe kendi panelini açar.
func guiInstancePath() string {
	sum := sha1.Sum([]byte(strings.ToLower(filepath.Clean(guiStartDir()))))
	return filepath.Join(guiStateDir(), "gui-"+hex.EncodeToString(sum[:])[:10]+".json")
}

func guiProfileDir() string {
	base, err := os.UserCacheDir()
	if err != nil {
		base = os.TempDir()
	}
	return filepath.Join(base, "syncftp", "panel-profile")
}

func readGUIInstance() (guiInstance, bool) {
	var gi guiInstance
	data, err := os.ReadFile(guiInstancePath())
	if err != nil || json.Unmarshal(data, &gi) != nil || gi.Port == 0 {
		return gi, false
	}
	c := &http.Client{Timeout: 1500 * time.Millisecond}
	resp, err := c.Get(fmt.Sprintf("http://127.0.0.1:%d/api/alive", gi.Port))
	if err != nil {
		return gi, false
	}
	resp.Body.Close()
	return gi, resp.StatusCode == http.StatusOK
}

func (gi guiInstance) url() string {
	return fmt.Sprintf("http://127.0.0.1:%d/login?t=%s", gi.Port, gi.Token)
}

// launchGUIDetached panel çalışıyorsa penceresini öne getirir/açar; değilse ayrık süreçte başlatır.
func launchGUIDetached() error {
	if gi, ok := readGUIInstance(); ok {
		if !bringWindowToFront(guiWindowTitle()) {
			return openAppWindow(gi.url(), guiProfileDir())
		}
		return nil
	}
	return spawnDetached("gui", "--foreground")
}

// runGUIForeground paneli bu süreçte çalıştırır (çift tık + `gui --foreground`).
func runGUIForeground() error {
	if gi, ok := readGUIInstance(); ok {
		if !bringWindowToFront(guiWindowTitle()) {
			return openAppWindow(gi.url(), guiProfileDir())
		}
		return nil
	}
	loadGUILang()

	tokBuf := make([]byte, 24)
	if _, err := rand.Read(tokBuf); err != nil {
		return err
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	sv := &guiServer{
		token:    hex.EncodeToString(tokBuf),
		port:     ln.Addr().(*net.TCPAddr).Port,
		started:  time.Now(),
		jobs:     newJobManager(),
		watches:  newWatchManager(),
		quit:     make(chan struct{}),
		hasTray:  trayAvailable() && !guiFlagNoWindow,
		startDir: guiStartDir(),
	}
	projects.SetBase(sv.startDir)
	sv.hooks = newDeployManager(sv.jobs, nil)
	sv.ai = newAIManager()
	gi := guiInstance{Port: sv.port, Token: sv.token, PID: os.Getpid()}
	_ = os.MkdirAll(guiStateDir(), 0700)
	if data, err := json.Marshal(gi); err == nil {
		_ = os.WriteFile(guiInstancePath(), data, 0600)
	}
	defer os.Remove(guiInstancePath())

	srv := &http.Server{Handler: sv.handler(), ReadHeaderTimeout: 10 * time.Second}
	go func() { _ = srv.Serve(ln) }()
	fmt.Printf(lang.L.GuiStartFmt, fmt.Sprintf("http://127.0.0.1:%d", sv.port))

	switch {
	case guiFlagNoWindow:
		fmt.Println(gi.url())
	case guiFlagBackground && sv.hasTray:
		// yalnızca tray — pencere tray'den açılır
	default:
		if err := openAppWindow(gi.url(), guiProfileDir()); err != nil {
			fmt.Print(lang.L.GuiNoBrowser)
			_ = openDefaultBrowser(gi.url())
		}
	}

	// CI/CD webhook dinleyicisi (yalnızca 127.0.0.1; dışarıya tünel ile açılır)
	if err := sv.hooks.listen(""); err == nil {
		if loadGUIPersist().HookTunnel {
			go func() { _, _ = sv.hooks.startTunnel() }()
		}
	}
	defer sv.hooks.close()
	// Önceki oturumdan kalan watch'lar
	sv.restoreWatches()

	go sv.idleWatchdog()
	if sv.hasTray {
		sv.runTray(gi.url()) // tray "Çıkış" ya da panelden "Kapat" → döner
	} else {
		<-sv.quit
	}
	// Panel penceresi (Edge/Chrome --app) süreçten bağımsız yaşar → tepsi "Çıkış" / panel "Kapat"
	// sonrası açık kalmasın. Sayfa da presence'tan "bye" alınca kendini kapatmayı dener.
	closeAppWindows(guiWindowTitle())
	sv.watches.stopAll()
	closeAllTerms()
	sv.ai.closeAll() // `claude -p` / `opencode run` çocuk süreçleri panel kapanınca yaşamasın
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return srv.Shutdown(ctx)
}

// GUI dil tercihi (proje bağımsız): <UserConfigDir>/syncftp/gui-lang
func loadGUILang() {
	if data, err := os.ReadFile(lang.UserLangPath()); err == nil {
		if l := strings.TrimSpace(string(data)); l == "en" || l == "tr" {
			lang.Set(l)
		}
	}
}

func saveGUILang(l string) {
	_ = os.MkdirAll(guiStateDir(), 0700)
	_ = os.WriteFile(lang.UserLangPath(), []byte(l), 0600)
}

// ── Sunucu ────────────────────────────────────────────────────────────────────

type guiServer struct {
	token    string
	port     int
	started  time.Time
	lastSeen atomic.Int64 // unix nano — son heartbeat / son açık pencere
	clients  atomic.Int32 // açık /api/presence bağlantısı (= açık panel penceresi) sayısı
	jobs     *jobManager
	watches  *watchManager
	quit     chan struct{}
	quitOnce sync.Once
	hasTray  bool           // tray varken pencere kapanınca süreç tray'de yaşamaya devam eder
	hooks    *deployManager // CI/CD webhook dinleyicisi + cloudflared tüneli
	ai       *aiManager     // yapay zekâ asistanı oturumları (yalnızca panel)
	startDir string         // dosya seçicilerin başlangıç dizini (guiStartDir)

	navMu   sync.Mutex
	navSubs map[chan string]struct{} // açık panel pencereleri (presence) — tray'den sayfa değiştirme
}

// navigate açık panel pencerelerine sayfa değiştirme olayı yollar (örn. "#/p/shop").
func (sv *guiServer) navigate(hash string) {
	sv.navMu.Lock()
	defer sv.navMu.Unlock()
	for ch := range sv.navSubs {
		select {
		case ch <- hash:
		default:
		}
	}
}

// guiStartDir panelin açıldığı dizin — CLI'daki gibi seçiciler buradan başlar: çift tıkta exe'nin
// klasörü (Explorer çalışma dizinini oraya kurar), `syncftp gui` ile terminalin bulunduğu dizin.
// Windows açılışında (Run anahtarı) çalışma dizini System32 olur → exe'nin klasörüne düşülür.
func guiStartDir() string {
	exeDir := ""
	if exe, err := os.Executable(); err == nil {
		exeDir = filepath.Dir(exe)
	}
	wd, err := os.Getwd()
	if err != nil || wd == "" {
		return exeDir
	}
	if root := os.Getenv("SystemRoot"); root != "" && exeDir != "" {
		if rel, err := filepath.Rel(root, wd); err == nil && !strings.HasPrefix(rel, "..") {
			return exeDir
		}
	}
	return wd
}

func (sv *guiServer) shutdown() { sv.quitOnce.Do(func() { close(sv.quit) }) }

// idleWatchdog son panel penceresi kapanınca süreci kapatır. Varlık, sayfanın açık
// tuttuğu /api/presence SSE bağlantısıyla izlenir — timer tabanlı heartbeat'in aksine
// küçültülmüş/arka plandaki pencerede tarayıcı tarafından yavaşlatılmaz.
// Çalışan job ya da watch varsa açık kalır — panel yeniden açılınca kaldığı yerden görünür.
func (sv *guiServer) idleWatchdog() {
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	for range t.C {
		if sv.hasTray {
			return // tray'den "Çıkış" ile kapanır
		}
		if sv.clients.Load() > 0 {
			sv.lastSeen.Store(time.Now().UnixNano())
			continue
		}
		if sv.jobs.running() > 0 || sv.watches.count() > 0 {
			continue
		}
		last := sv.lastSeen.Load()
		if last == 0 {
			if time.Since(sv.started) > 2*time.Minute {
				sv.shutdown()
				return
			}
			continue
		}
		// Sayfa yenileme / gezinme sırasında bağlantı kısa süre kopabilir → 20 sn tolerans
		if time.Since(time.Unix(0, last)) > 20*time.Second {
			sv.shutdown()
			return
		}
	}
}

func (sv *guiServer) hostOK(r *http.Request) bool {
	h := r.Host
	p := strconv.Itoa(sv.port)
	return h == "127.0.0.1:"+p || h == "localhost:"+p
}

func (sv *guiServer) authed(r *http.Request) bool {
	tok := r.Header.Get("X-Token")
	if c, err := r.Cookie("syncftp_t"); err == nil && tok == "" {
		tok = c.Value
	}
	return subtle.ConstantTimeCompare([]byte(tok), []byte(sv.token)) == 1
}

func (sv *guiServer) handler() http.Handler {
	mux := http.NewServeMux()
	static, _ := fs.Sub(guiWeb, "web")
	var fileSrv http.Handler = http.FileServer(http.FS(static))
	// Geliştirme: SYNCFTP_WEB_DIR=<.../cmd/syncftp/web> → arayüz diskten okunur, build gerekmez (F5 yeter)
	if d := os.Getenv("SYNCFTP_WEB_DIR"); d != "" {
		disk := http.FileServer(http.Dir(d))
		fileSrv = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")
			disk.ServeHTTP(w, r)
		})
	}

	mux.HandleFunc("GET /api/alive", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("GET /login", func(w http.ResponseWriter, r *http.Request) {
		if subtle.ConstantTimeCompare([]byte(r.URL.Query().Get("t")), []byte(sv.token)) != 1 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.SetCookie(w, &http.Cookie{Name: "syncftp_t", Value: sv.token, Path: "/", HttpOnly: true, SameSite: http.SameSiteStrictMode})
		dest := "/"
		if next := r.URL.Query().Get("next"); strings.HasPrefix(next, "#/") {
			dest = "/" + next // tray'den belirli bir sayfa (proje) açma
		}
		http.Redirect(w, r, dest, http.StatusFound)
	})

	api := http.NewServeMux()
	sv.routes(api)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if !sv.hostOK(r) {
			http.Error(w, "bad host", http.StatusForbidden)
			return
		}
		if !sv.authed(r) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				http.Error(w, `{"ok":false,"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, `<!doctype html><meta charset="utf-8"><title>syncFTP</title><body style="background:#111113;color:#edeef0;font:14px system-ui;display:grid;place-items:center;height:100vh;margin:0"><div>%s</div>`, lang.L.GuiSessionInvalidHTML)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/") {
			sv.lastSeen.Store(time.Now().UnixNano())
			api.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		fileSrv.ServeHTTP(w, r)
	})
	return mux
}

// ── JSON yardımcıları ─────────────────────────────────────────────────────────

func guiJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "data": v})
}

func guiErr(w http.ResponseWriter, code int, err error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": err.Error()})
}

func guiDecode(r *http.Request, v any) error {
	defer r.Body.Close()
	return json.NewDecoder(io.LimitReader(r.Body, 8<<20)).Decode(v)
}

// ══════════════════════════════════════════════════════════════════════════════
// Job yöneticisi — uzun süren işler (sync, calibrate, verify, rollback, detect…)
// arka planda çalışır; çıktıları satır satır tutulur ve SSE ile canlı akar.
// ══════════════════════════════════════════════════════════════════════════════

type job struct {
	ID      string     `json:"id"`
	Kind    string     `json:"kind"`
	Project string     `json:"project"`
	Server  string     `json:"server"`
	Title   string     `json:"title"`
	Status  string     `json:"status"` // running | done | error
	Error   string     `json:"error,omitempty"`
	Started time.Time  `json:"started"`
	Ended   *time.Time `json:"ended,omitempty"`
	Result  any        `json:"result,omitempty"`
	Lines   int        `json:"lines"`

	mu      sync.Mutex
	lines   []string
	partial string
	subs    map[chan string]struct{}
	done    chan struct{}
	lockKey string
	cancel  func() // watch gibi durdurulabilir job'lar için
}

const jobMaxLines = 20000

// Write io.Writer: satırlara böler; \r ile yazılan canlı sayaç satırlarının yalnızca sonu tutulur.
func (j *job) Write(p []byte) (int, error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	text := j.partial + strings.ReplaceAll(string(p), "\r\n", "\n")
	parts := strings.Split(text, "\n")
	j.partial = parts[len(parts)-1]
	for _, line := range parts[:len(parts)-1] {
		if i := strings.LastIndex(line, "\r"); i >= 0 {
			line = line[i+1:]
		}
		j.emitLocked(synclog.StripANSI(line))
	}
	return len(p), nil
}

func (j *job) emitLocked(line string) {
	if len(j.lines) >= jobMaxLines {
		j.lines = j.lines[len(j.lines)/4:]
	}
	j.lines = append(j.lines, line)
	j.Lines++
	for ch := range j.subs {
		select {
		case ch <- line:
		default: // yavaş istemci — satır düşer, tam log /lines ile alınabilir
		}
	}
}

func (j *job) finish(result any, err error) {
	j.mu.Lock()
	if j.partial != "" {
		j.emitLocked(synclog.StripANSI(j.partial))
		j.partial = ""
	}
	now := time.Now()
	j.Ended = &now
	j.Result = result
	if err != nil {
		j.Status = "error"
		j.Error = err.Error()
	} else {
		j.Status = "done"
	}
	j.mu.Unlock()
	close(j.done)
}

// isRunning Status'u j.mu altında okur (finish aynı kilitle yazar).
func (j *job) isRunning() bool {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.Status == "running"
}

// cancelFunc durdurulabilir işin durdurma fonksiyonu (yoksa nil).
func (j *job) cancelFunc() func() {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.cancel
}

func (j *job) snapshot() map[string]any {
	j.mu.Lock()
	defer j.mu.Unlock()
	return map[string]any{
		"id": j.ID, "kind": j.Kind, "project": j.Project, "server": j.Server, "title": j.Title,
		"status": j.Status, "error": j.Error, "started": j.Started, "ended": j.Ended,
		"result": j.Result, "lines": j.Lines, "cancelable": j.cancel != nil,
	}
}

type jobManager struct {
	mu    sync.Mutex
	seq   int
	jobs  []*job
	locks map[string]string // project|server → job id
}

func newJobManager() *jobManager { return &jobManager{locks: map[string]string{}} }

// start işi arka planda çalıştırır. lockKey doluysa aynı anahtarda çalışan iş varken
// reddedilir (aynı sunucunun state'ine iki iş birden yazmasın).
func (m *jobManager) start(kind, project, server, title, lockKey string, fn func(j *job) (any, error)) (*job, error) {
	return m.startCancelable(kind, project, server, title, lockKey, nil, fn)
}

// startCancelable start ile aynı; cancel iş goroutine'i başlamadan önce atanır (yarışsız durdurma).
func (m *jobManager) startCancelable(kind, project, server, title, lockKey string, cancel func(), fn func(j *job) (any, error)) (*job, error) {
	m.mu.Lock()
	if lockKey != "" {
		if id, busy := m.locks[lockKey]; busy {
			m.mu.Unlock()
			return nil, fmt.Errorf(lang.L.GuiBusyFmt, server, id)
		}
	}
	m.seq++
	j := &job{
		ID: fmt.Sprintf("j%d", m.seq), Kind: kind, Project: project, Server: server, Title: title,
		Status: "running", Started: time.Now(), subs: map[chan string]struct{}{}, done: make(chan struct{}), lockKey: lockKey, cancel: cancel,
	}
	if lockKey != "" {
		m.locks[lockKey] = j.ID
	}
	m.jobs = append(m.jobs, j)
	if len(m.jobs) > 200 { // eski bitmiş işleri unut
		var keep []*job
		for i, x := range m.jobs {
			if x.isRunning() || i >= len(m.jobs)-100 {
				keep = append(keep, x)
			}
		}
		m.jobs = keep
	}
	m.mu.Unlock()

	go func() {
		var res any
		var err error
		func() {
			defer func() {
				if p := recover(); p != nil {
					err = fmt.Errorf("panic: %v", p)
				}
			}()
			res, err = fn(j)
		}()
		j.finish(res, err)
		if lockKey != "" {
			m.mu.Lock()
			if m.locks[lockKey] == j.ID {
				delete(m.locks, lockKey)
			}
			m.mu.Unlock()
		}
	}()
	return j, nil
}

func (m *jobManager) get(id string) *job {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, j := range m.jobs {
		if j.ID == id {
			return j
		}
	}
	return nil
}

func (m *jobManager) list(project string) []map[string]any {
	m.mu.Lock()
	jobs := append([]*job(nil), m.jobs...)
	m.mu.Unlock()
	out := []map[string]any{}
	for i := len(jobs) - 1; i >= 0; i-- {
		if project == "" || jobs[i].Project == project {
			out = append(out, jobs[i].snapshot())
		}
	}
	return out
}

func (m *jobManager) running() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := 0
	for _, j := range m.jobs {
		if j.Kind != "watch" && j.isRunning() {
			n++
		}
	}
	return n
}

// watchJobsRunning projenin hâlâ bitmemiş watch işi sayısı (durdurma sonrası bekleme için).
func (m *jobManager) watchJobsRunning(project string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := 0
	for _, j := range m.jobs {
		if j.Kind == "watch" && j.Project == project && j.isRunning() {
			n++
		}
	}
	return n
}

// busyServers projede şu an işi olan sunucular.
func (m *jobManager) busyServers(project string) map[string]string {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := map[string]string{}
	for k, id := range m.locks {
		if p, s, ok := strings.Cut(k, "|"); ok && p == project {
			out[s] = id
		}
	}
	return out
}

// ── Watch yöneticisi (proje başına bir arka plan watch) ──────────────────────

// guiWatch tek sunucunun watch'ı. Eskiden proje başına tek watch vardı (tüm sunucular birlikte açılıp
// kapanıyordu); CLI'daki gibi her sunucu ayrı izlenir.
type guiWatch struct {
	Project  string    `json:"project"`
	Server   string    `json:"server"`
	Servers  []string  `json:"servers"` // [Server] — eski istemciler için
	Started  time.Time `json:"started"`
	JobID    string    `json:"job_id"`
	LastSync string    `json:"last_sync,omitempty"`
	Syncs    int       `json:"syncs"`
	stop     chan struct{}
	path     string // proje dizini — kalıcı kayıt anahtarı
}

type watchManager struct {
	mu sync.Mutex
	m  map[string]*guiWatch
}

func newWatchManager() *watchManager { return &watchManager{m: map[string]*guiWatch{}} }

func (wm *watchManager) count() int {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	return len(wm.m)
}

func watchKey(project, server string) string { return project + "\x00" + server }

// reserve gw'nin anahtarını tek kilitte kontrol edip ayırır; zaten izleniyorsa false.
func (wm *watchManager) reserve(gw *guiWatch) bool {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	k := watchKey(gw.Project, gw.Server)
	if _, ok := wm.m[k]; ok {
		return false
	}
	wm.m[k] = gw
	return true
}

// release başlatılamayan watch'ın ayrılmış anahtarını bırakır (yerine başkası konduysa dokunmaz).
func (wm *watchManager) release(gw *guiWatch) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	k := watchKey(gw.Project, gw.Server)
	if wm.m[k] == gw {
		delete(wm.m, k)
	}
}

// forProject projenin açık watch'ları (sunucu adına göre).
func (wm *watchManager) forProject(project string) []guiWatch {
	out := []guiWatch{}
	for _, w := range wm.all() {
		if w.Project == project {
			out = append(out, w)
		}
	}
	return out
}

func (wm *watchManager) all() []guiWatch {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	out := []guiWatch{}
	for _, w := range wm.m {
		out = append(out, *w)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Project != out[j].Project {
			return out[i].Project < out[j].Project
		}
		return out[i].Server < out[j].Server
	})
	return out
}

// stop sunucunun watch'ını durdurur; server boşsa projenin tüm watch'ları. Durdurulan sayıyı döner.
func (wm *watchManager) stop(project, server string) int {
	wm.mu.Lock()
	var stopped []*guiWatch
	for k, w := range wm.m {
		if w.Project == project && (server == "" || w.Server == server) {
			stopped = append(stopped, w)
			delete(wm.m, k)
		}
	}
	wm.mu.Unlock()
	for _, w := range stopped {
		close(w.stop)
	}
	return len(stopped)
}

func (wm *watchManager) stopAll() {
	wm.mu.Lock()
	all := wm.m
	wm.m = map[string]*guiWatch{}
	wm.mu.Unlock()
	for _, w := range all {
		close(w.stop)
	}
}

func fileExistsPlain(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}

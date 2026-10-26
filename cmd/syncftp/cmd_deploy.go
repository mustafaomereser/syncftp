package main

import (
	"bufio"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"syncftp/internal/config"
	"syncftp/internal/lang"
	"syncftp/internal/projects"
)

// ══════════════════════════════════════════════════════════════════════════════
// CI/CD — gelen webhook ile deploy
//
//   GitHub/GitLab/Gitea/Bitbucket push ─▶ https://<tünel>/hook/<proje>/<ad>
//        └─ cloudflared quick tunnel (trycloudflare) ya da kendi reverse proxy'n
//             └─ 127.0.0.1:<hook port>  (panelden AYRI dinleyici — panel asla dışarı açılmaz)
//                  └─ imza doğrula → branch filtresi → [git pull] → sunucu başına sync işi
//
// Dinleyici GUI sürecinde (tray ile sürekli açık) ya da headless `syncftp hook serve` ile çalışır.
// ══════════════════════════════════════════════════════════════════════════════

const defaultHookPort = 47821

type delivery struct {
	Time    time.Time `json:"time"`
	Project string    `json:"project"`
	Trigger string    `json:"trigger"`
	Event   string    `json:"event"`
	Branch  string    `json:"branch,omitempty"`
	Commit  string    `json:"commit,omitempty"`
	Status  string    `json:"status"` // accepted | ignored | rejected | error
	Detail  string    `json:"detail,omitempty"`
	JobID   string    `json:"job_id,omitempty"`
}

type deployManager struct {
	jobs *jobManager
	tee  io.Writer // headless modda iş çıktıları buraya da akar

	mu         sync.Mutex
	ln         net.Listener
	port       int
	srv        *http.Server
	deliveries []delivery
	running    map[string]bool // proje|tetikleyici → deploy sürüyor
	pending    map[string]bool // sürerken yeni push geldi → bitince bir tur daha

	tunnel tunnelState
}

type tunnelState struct {
	cmd    *exec.Cmd
	URL    string `json:"url"`
	Error  string `json:"error,omitempty"`
	Active bool   `json:"active"`
}

func newDeployManager(jobs *jobManager, tee io.Writer) *deployManager {
	return &deployManager{jobs: jobs, tee: tee, running: map[string]bool{}, pending: map[string]bool{}}
}

// listen webhook dinleyicisini başlatır. addr boşsa 127.0.0.1:<kayıtlı port>; port doluysa rastgele.
func (dm *deployManager) listen(addr string) error {
	if addr == "" {
		port := loadGUIPersist().HookPort
		if port == 0 {
			port = defaultHookPort
		}
		addr = fmt.Sprintf("127.0.0.1:%d", port)
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil && strings.HasPrefix(addr, "127.0.0.1:") {
		ln, err = net.Listen("tcp", "127.0.0.1:0")
	}
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /hook/{p}/{t}", dm.handleHook)
	mux.HandleFunc("GET /hook/{p}/{t}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "syncFTP deploy hook — POST bekleniyor")
	})
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "syncFTP")
	})
	dm.mu.Lock()
	dm.ln = ln
	dm.port = ln.Addr().(*net.TCPAddr).Port
	dm.srv = &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	dm.mu.Unlock()
	go func() { _ = dm.srv.Serve(ln) }()
	return nil
}

func (dm *deployManager) close() {
	dm.stopTunnel()
	dm.mu.Lock()
	srv := dm.srv
	dm.mu.Unlock()
	if srv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}
}

func (dm *deployManager) localURL() string {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	if dm.port == 0 {
		return ""
	}
	return fmt.Sprintf("http://127.0.0.1:%d", dm.port)
}

// publicURL öncelik: kullanıcının sabit adresi > çalışan tünel > yerel adres.
func (dm *deployManager) publicURL() string {
	if u := strings.TrimRight(loadGUIPersist().HookPublicURL, "/"); u != "" {
		return u
	}
	dm.mu.Lock()
	tu := dm.tunnel.URL
	dm.mu.Unlock()
	if tu != "" {
		return tu
	}
	return dm.localURL()
}

func hookURL(base, project, trigger string) string {
	return strings.TrimRight(base, "/") + "/hook/" + url.PathEscape(project) + "/" + url.PathEscape(trigger)
}

func (dm *deployManager) record(d delivery) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.deliveries = append(dm.deliveries, d)
	if len(dm.deliveries) > 200 {
		dm.deliveries = dm.deliveries[len(dm.deliveries)-200:]
	}
	if dm.tee != nil {
		fmt.Fprintf(dm.tee, "[%s] %s/%s %s %s %s %s\n", d.Time.Format("15:04:05"), d.Project, d.Trigger, d.Event, d.Branch, d.Status, d.Detail)
	}
}

func (dm *deployManager) recent(project string) []delivery {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	out := []delivery{}
	for i := len(dm.deliveries) - 1; i >= 0; i-- {
		if project == "" || dm.deliveries[i].Project == project {
			out = append(out, dm.deliveries[i])
		}
	}
	return out
}

// statusLine tray için kısa durum ("webhook: tünel açık" vb.)
func (dm *deployManager) statusLine() string {
	if dm == nil {
		return ""
	}
	dm.mu.Lock()
	defer dm.mu.Unlock()
	if dm.tunnel.Active && dm.tunnel.URL != "" {
		return lang.L.TrayHookTunnel
	}
	return ""
}

// ── İmza doğrulama ────────────────────────────────────────────────────────────

func hmacSHA256Hex(secret string, body []byte) string {
	m := hmac.New(sha256.New, []byte(secret))
	m.Write(body)
	return hex.EncodeToString(m.Sum(nil))
}

func eqConst(a, b string) bool { return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1 }

// verifyHook sağlayıcıya göre isteği doğrular; olay adını döner.
func verifyHook(d config.Deploy, r *http.Request, body []byte) (event string, err error) {
	if d.Secret == "" {
		return "", errors.New("secret tanımlı değil")
	}
	switch d.Provider {
	case "github", "bitbucket":
		sig := r.Header.Get("X-Hub-Signature-256")
		if sig == "" {
			sig = r.Header.Get("X-Hub-Signature") // Bitbucket Cloud: "sha256=…"
		}
		if !eqConst(sig, "sha256="+hmacSHA256Hex(d.Secret, body)) {
			return "", errors.New("imza geçersiz")
		}
		event = r.Header.Get("X-GitHub-Event")
		if event == "" {
			event = r.Header.Get("X-Event-Key") // bitbucket: repo:push
		}
	case "gitea":
		sig := r.Header.Get("X-Gitea-Signature")
		if sig == "" {
			sig = strings.TrimPrefix(r.Header.Get("X-Hub-Signature-256"), "sha256=")
		}
		if !eqConst(sig, hmacSHA256Hex(d.Secret, body)) {
			return "", errors.New("imza geçersiz")
		}
		event = r.Header.Get("X-Gitea-Event")
	case "gitlab":
		if !eqConst(r.Header.Get("X-Gitlab-Token"), d.Secret) {
			return "", errors.New("token geçersiz")
		}
		event = r.Header.Get("X-Gitlab-Event")
	default: // generic
		tok := r.Header.Get("X-Syncftp-Token")
		if tok == "" {
			tok = r.URL.Query().Get("token")
		}
		if tok == "" && strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			tok = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		}
		if !eqConst(tok, d.Secret) {
			return "", errors.New("token geçersiz")
		}
		event = "push"
	}
	return event, nil
}

// pushInfo payload'dan branch ve commit özetini çıkarır (GitHub/GitLab/Gitea/Bitbucket).
func pushInfo(body []byte) (branch, commit string) {
	var p struct {
		Ref        string `json:"ref"`
		After      string `json:"after"`
		HeadCommit *struct {
			ID      string `json:"id"`
			Message string `json:"message"`
		} `json:"head_commit"`
		Push *struct {
			Changes []struct {
				New *struct {
					Name   string `json:"name"`
					Target struct {
						Hash    string `json:"hash"`
						Message string `json:"message"`
					} `json:"target"`
				} `json:"new"`
			} `json:"changes"`
		} `json:"push"`
	}
	if json.Unmarshal(body, &p) != nil {
		return "", ""
	}
	branch = strings.TrimPrefix(p.Ref, "refs/heads/")
	if p.HeadCommit != nil {
		commit = shortSHA(p.HeadCommit.ID) + " " + firstLine(p.HeadCommit.Message)
	} else if p.After != "" {
		commit = shortSHA(p.After)
	}
	if branch == "" && p.Push != nil && len(p.Push.Changes) > 0 && p.Push.Changes[0].New != nil {
		n := p.Push.Changes[0].New
		branch = n.Name
		commit = shortSHA(n.Target.Hash) + " " + firstLine(n.Target.Message)
	}
	return branch, strings.TrimSpace(commit)
}

func shortSHA(s string) string {
	if len(s) > 7 {
		return s[:7]
	}
	return s
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	if len(s) > 80 {
		s = s[:80] + "…"
	}
	return s
}

func isPushEvent(provider, event string) bool {
	e := strings.ToLower(event)
	switch provider {
	case "gitlab":
		return e == "push hook" || e == "tag push hook"
	case "bitbucket":
		return e == "repo:push" || e == "push"
	case "generic":
		return true
	}
	return e == "push"
}

// handleHook POST /hook/{proje}/{tetikleyici}
func (dm *deployManager) handleHook(w http.ResponseWriter, r *http.Request) {
	d := delivery{Time: time.Now(), Project: r.PathValue("p"), Trigger: r.PathValue("t")}
	body, _ := io.ReadAll(io.LimitReader(r.Body, 5<<20))
	reply := func(code int, status, detail string) {
		d.Status, d.Detail = status, detail
		dm.record(d)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": status, "detail": detail, "job": d.JobID})
	}

	p, ok := projects.Find(d.Project)
	if !ok {
		reply(404, "rejected", "project not found")
		return
	}
	cfg, err := config.Load(p.Path)
	if err != nil {
		reply(500, "error", err.Error())
		return
	}
	var trig *config.Deploy
	for i := range cfg.Deploy {
		if cfg.Deploy[i].Name == d.Trigger {
			trig = &cfg.Deploy[i]
		}
	}
	if trig == nil || !trig.Enabled {
		reply(404, "rejected", "trigger not found or disabled")
		return
	}
	event, err := verifyHook(*trig, r, body)
	d.Event = event
	if err != nil {
		reply(401, "rejected", err.Error())
		return
	}
	if strings.EqualFold(event, "ping") || strings.EqualFold(event, "System Hook") {
		reply(200, "ignored", "pong")
		return
	}
	if !isPushEvent(trig.Provider, event) {
		reply(202, "ignored", "event: "+event)
		return
	}
	d.Branch, d.Commit = pushInfo(body)
	if trig.Branch != "" && d.Branch != "" && d.Branch != trig.Branch {
		reply(202, "ignored", "branch "+d.Branch+" ≠ "+trig.Branch)
		return
	}
	jobID, err := dm.runDeploy(p, *trig, fmt.Sprintf("%s %s", d.Branch, d.Commit))
	if err != nil {
		reply(500, "error", err.Error())
		return
	}
	d.JobID = jobID
	reply(202, "accepted", d.Commit)
}

// runDeploy deploy işini başlatır. Aynı tetikleyici için deploy sürüyorsa yeni istek
// bekleyen olarak işaretlenir ve bitince bir tur daha çalışır (push'lar kaybolmaz).
func (dm *deployManager) runDeploy(p projects.Project, trig config.Deploy, info string) (string, error) {
	key := p.Name + "|" + trig.Name
	dm.mu.Lock()
	if dm.running[key] {
		dm.pending[key] = true
		dm.mu.Unlock()
		return "", nil
	}
	dm.running[key] = true
	dm.mu.Unlock()

	j, err := dm.jobs.start("deploy", p.Name, trig.Name, lang.L.GuiJobDeploy, "", func(j *job) (any, error) {
		var out io.Writer = j
		if dm.tee != nil {
			out = io.MultiWriter(j, dm.tee)
		}
		res, err := dm.deployOnce(out, p, trig, info)
		for {
			dm.mu.Lock()
			again := dm.pending[key]
			dm.pending[key] = false
			if !again {
				dm.running[key] = false
			}
			dm.mu.Unlock()
			if !again {
				break
			}
			fmt.Fprint(out, lang.L.DeployQueuedRun)
			res, err = dm.deployOnce(out, p, trig, "")
		}
		return res, err
	})
	if err != nil {
		dm.mu.Lock()
		dm.running[key] = false
		dm.mu.Unlock()
		return "", err
	}
	return j.ID, nil
}

// deployOnce: [git pull] → her sunucu için ayrı sync işi (sunucu kilidine uyar) → özet.
func (dm *deployManager) deployOnce(out io.Writer, p projects.Project, trig config.Deploy, info string) (any, error) {
	fmt.Fprintf(out, lang.L.DeployStartFmt, p.Name, trig.Name, info)
	if trig.GitPull {
		fmt.Fprint(out, lang.L.DeployGitPull)
		args := []string{"-C", p.Path, "pull", "--ff-only"}
		cmd := exec.Command("git", args...)
		hideCmdWindow(cmd)
		cmd.Stdout, cmd.Stderr = indentWriter(out), indentWriter(out)
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(out, lang.L.DeployGitPullErrFmt, err)
			return nil, fmt.Errorf("git pull: %w", err)
		}
	}
	cfg, err := config.Load(p.Path)
	if err != nil {
		return nil, err
	}
	servers, err := selectServers(cfg, trig.Servers)
	if err != nil {
		return nil, err
	}
	type srvRes struct {
		Server   string `json:"server"`
		Uploaded int    `json:"uploaded"`
		Failed   int    `json:"failed"`
		Error    string `json:"error,omitempty"`
	}
	var results []srvRes
	var firstErr error
	for _, srv := range servers {
		srv := srv
		req := guiJobReq{Kind: "sync", DryRun: trig.DryRun}
		var sj *job
		// Sunucuda başka iş varsa (elle sync, watch sync'i değil) bitmesini bekle — en fazla 10 dk
		deadline := time.Now().Add(10 * time.Minute)
		for {
			ctx, cancel := context.WithCancel(context.Background())
			sj, err = dm.jobs.startCancelable("sync", p.Name, srv.Name, lang.L.GuiJobSync+" (deploy)", p.Name+"|"+srv.Name, cancel, func(j *job) (any, error) {
				defer cancel()
				return guiRunSync(ctx, j, p.Path, srv, req)
			})
			if err != nil {
				cancel()
			}
			if err == nil || time.Now().After(deadline) {
				break
			}
			time.Sleep(2 * time.Second)
		}
		if err != nil {
			fmt.Fprintf(out, lang.L.DeploySrvErrFmt, srv.Name, err)
			results = append(results, srvRes{Server: srv.Name, Error: err.Error()})
			firstErr = err
			continue
		}
		<-sj.done
		r := srvRes{Server: srv.Name}
		if m, ok := sj.Result.(map[string]any); ok {
			r.Uploaded, _ = m["uploaded"].(int)
			r.Failed, _ = m["failed"].(int)
		}
		if sj.Error != "" {
			r.Error = sj.Error
			firstErr = errors.New(sj.Error)
		}
		sj.mu.Lock()
		for _, l := range sj.lines {
			fmt.Fprintln(out, "  │ "+l)
		}
		sj.mu.Unlock()
		results = append(results, r)
	}
	fmt.Fprint(out, lang.L.DeployDone)
	return results, firstErr
}

// indentWriter alt süreç çıktısını girintili satırlara çevirir.
func indentWriter(w io.Writer) io.Writer {
	pr, pw := io.Pipe()
	go func() {
		sc := bufio.NewScanner(pr)
		for sc.Scan() {
			fmt.Fprintln(w, "    "+sc.Text())
		}
	}()
	return pw
}

// ── cloudflared quick tunnel (trycloudflare) ─────────────────────────────────

var tryCloudflareRe = regexp.MustCompile(`https://[a-z0-9-]+\.trycloudflare\.com`)

func cloudflaredLocalPath() string {
	name := "cloudflared"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join(guiStateDir(), "bin", name)
}

// findCloudflared PATH'te ya da syncFTP'nin bin dizininde cloudflared arar.
func findCloudflared() string {
	if p, err := exec.LookPath("cloudflared"); err == nil {
		return p
	}
	if p := cloudflaredLocalPath(); fileExistsPlain(p) {
		return p
	}
	return ""
}

// startTunnel cloudflared quick tunnel'ı açar ve public URL gelene kadar (en fazla 30 sn) bekler.
func (dm *deployManager) startTunnel() (string, error) {
	dm.mu.Lock()
	if dm.tunnel.Active {
		u := dm.tunnel.URL
		dm.mu.Unlock()
		return u, nil
	}
	port := dm.port
	dm.mu.Unlock()
	if port == 0 {
		return "", errors.New("hook listener not running")
	}
	bin := findCloudflared()
	if bin == "" {
		return "", errors.New(lang.L.DeployNoCloudflared)
	}
	cmd := exec.Command(bin, "tunnel", "--no-autoupdate", "--url", fmt.Sprintf("http://127.0.0.1:%d", port))
	hideCmdWindow(cmd)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}
	cmd.Stdout = io.Discard
	if err := cmd.Start(); err != nil {
		return "", err
	}
	urlCh := make(chan string, 1)
	go func() {
		sc := bufio.NewScanner(stderr)
		sent := false
		for sc.Scan() {
			if !sent {
				if m := tryCloudflareRe.FindString(sc.Text()); m != "" {
					urlCh <- m
					sent = true
				}
			}
		}
		// süreç bitti
		dm.mu.Lock()
		if dm.tunnel.cmd == cmd {
			dm.tunnel.Active = false
			dm.tunnel.URL = ""
			if dm.tunnel.Error == "" {
				dm.tunnel.Error = "cloudflared exited"
			}
			dm.tunnel.cmd = nil
		}
		dm.mu.Unlock()
	}()
	dm.mu.Lock()
	dm.tunnel = tunnelState{cmd: cmd, Active: true}
	dm.mu.Unlock()
	go func() { _ = cmd.Wait() }()

	select {
	case u := <-urlCh:
		dm.mu.Lock()
		dm.tunnel.URL = u
		dm.mu.Unlock()
		return u, nil
	case <-time.After(30 * time.Second):
		dm.stopTunnel()
		dm.mu.Lock()
		dm.tunnel.Error = "timeout waiting for trycloudflare URL"
		dm.mu.Unlock()
		return "", errors.New(dm.tunnel.Error)
	}
}

func (dm *deployManager) stopTunnel() {
	dm.mu.Lock()
	cmd := dm.tunnel.cmd
	dm.tunnel = tunnelState{}
	dm.mu.Unlock()
	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
}

func (dm *deployManager) tunnelInfo() tunnelState {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	return dm.tunnel
}

// installCloudflared resmi GitHub sürümünü indirir (Windows/Linux amd64/arm64).
func installCloudflared(out io.Writer) (string, error) {
	var asset string
	switch runtime.GOOS + "/" + runtime.GOARCH {
	case "windows/amd64":
		asset = "cloudflared-windows-amd64.exe"
	case "windows/386":
		asset = "cloudflared-windows-386.exe"
	case "linux/amd64":
		asset = "cloudflared-linux-amd64"
	case "linux/arm64":
		asset = "cloudflared-linux-arm64"
	default:
		return "", fmt.Errorf("unsupported platform %s/%s — install cloudflared manually", runtime.GOOS, runtime.GOARCH)
	}
	src := "https://github.com/cloudflare/cloudflared/releases/latest/download/" + asset
	fmt.Fprintf(out, lang.L.DeployDownloadingFmt, src)
	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Get(src)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	dst := cloudflaredLocalPath()
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return "", err
	}
	tmp := dst + ".part"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return "", err
	}
	var written int64
	buf := make([]byte, 256*1024)
	lastLog := time.Now()
	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := f.Write(buf[:n]); werr != nil {
				f.Close()
				return "", werr
			}
			written += int64(n)
			if time.Since(lastLog) > time.Second {
				fmt.Fprintf(out, "    %s\n", humanBytes(written))
				lastLog = time.Now()
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			f.Close()
			return "", rerr
		}
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	if err := os.Rename(tmp, dst); err != nil {
		return "", err
	}
	fmt.Fprintf(out, lang.L.DeployInstalledFmt, dst, humanBytes(written))
	return dst, nil
}

// ── headless: syncftp hook serve ──────────────────────────────────────────────

var (
	hookFlagAddr   string
	hookFlagTunnel bool
)

func init() {
	hookServeCmd.Flags().StringVar(&hookFlagAddr, "addr", "", "Dinlenecek adres (varsayılan 127.0.0.1:47821; VPS'te doğrudan açmak için 0.0.0.0:47821)")
	hookServeCmd.Flags().BoolVar(&hookFlagTunnel, "tunnel", false, "cloudflared quick tunnel aç (trycloudflare public URL)")
	hookCmd.AddCommand(hookServeCmd, hookListCmd)
	rootCmd.AddCommand(hookCmd)
}

var hookCmd = &cobra.Command{
	Use:   "hook",
	Short: "CI/CD: GitHub/GitLab/Gitea/Bitbucket push webhook'u ile otomatik deploy",
	Long: `Projelerin config'indeki "deploy" tetikleyicileri için webhook dinleyicisi.

Push → https://<public>/hook/<proje>/<tetikleyici> → imza doğrulama → [git pull] → sync.
Tetikleyiciler panelden (CI/CD sekmesi) ya da syncftp.json'daki "deploy" dizisinden tanımlanır.
Panel (tray) açıkken dinleyici zaten çalışır; GUI'siz makinelerde 'syncftp hook serve' kullanın.`,
}

var hookListCmd = &cobra.Command{
	Use:   "list",
	Short: "Tüm projelerdeki deploy tetikleyicilerini ve URL yollarını listeler",
	RunE: func(cmd *cobra.Command, args []string) error {
		any := false
		for _, p := range projects.List() {
			cfg, err := config.Load(p.Path)
			if err != nil {
				continue
			}
			for _, d := range cfg.Deploy {
				any = true
				state := "on"
				if !d.Enabled {
					state = "off"
				}
				fmt.Printf("  %-3s %-10s %-24s %s\n", state, d.Provider, p.Name+"/"+d.Name, hookURL("<public>", p.Name, d.Name))
			}
		}
		if !any {
			fmt.Print(lang.L.DeployNoTriggers)
		}
		return nil
	},
}

var hookServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Webhook dinleyicisini başlatır (GUI'siz sunucular için)",
	RunE: func(cmd *cobra.Command, args []string) error {
		dm := newDeployManager(newJobManager(), os.Stdout)
		if err := dm.listen(hookFlagAddr); err != nil {
			return err
		}
		defer dm.close()
		fmt.Printf(lang.L.DeployListeningFmt, dm.localURL())
		if hookFlagTunnel {
			u, err := dm.startTunnel()
			if err != nil {
				return err
			}
			fmt.Printf(lang.L.DeployTunnelFmt, u)
		}
		base := dm.publicURL()
		for _, p := range projects.List() {
			if cfg, err := config.Load(p.Path); err == nil {
				for _, d := range cfg.Deploy {
					if d.Enabled {
						fmt.Printf("  %s/%s → %s\n", p.Name, d.Name, hookURL(base, p.Name, d.Name))
					}
				}
			}
		}
		fmt.Print(lang.L.DeployCtrlC)
		// Ctrl+C → dinleyici ve cloudflared düzgün kapansın (defer dm.close)
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig
		return nil
	},
}

// activePort dinleyicinin şu an bağlı olduğu port (0 = çalışmıyor).
func (dm *deployManager) activePort() int {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	return dm.port
}

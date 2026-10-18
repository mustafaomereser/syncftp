package main

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"syncftp/internal/config"
	ftpclient "syncftp/internal/ftp"
	"syncftp/internal/ignore"
	"syncftp/internal/lang"
	"syncftp/internal/projects"
	"syncftp/internal/scanner"
)

// GUI 2.1 uçları: panel içi dosya seçici, proje yerel ağacı, uzak dosya işlemleri
// (yükle / klasör / taşı / ara), CI/CD (deploy tetikleyicileri + tünel), webhook testi, ignore editörü.
func (sv *guiServer) routesMore(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/fs/list", sv.apiFSList)
	mux.HandleFunc("GET /api/fs/tree", sv.apiFSTree)
	mux.HandleFunc("GET /api/help", sv.apiHelp)
	sv.registerSettingsAPI(mux)
	sv.registerLibraryAPI(mux)
	sv.registerTermAPI(mux)
	mux.HandleFunc("GET /api/p/{p}/local/ls", sv.apiLocalLs)
	mux.HandleFunc("GET /api/p/{p}/local/files", sv.apiLocalFiles)

	mux.HandleFunc("POST /api/p/{p}/remote/upload", sv.apiRemoteUpload)
	mux.HandleFunc("POST /api/p/{p}/remote/mkdir", sv.apiRemoteMkdir)
	mux.HandleFunc("POST /api/p/{p}/remote/rename", sv.apiRemoteRename)
	mux.HandleFunc("GET /api/p/{p}/remote/search", sv.apiRemoteSearch)

	mux.HandleFunc("GET /api/hooks", sv.apiHooks)
	mux.HandleFunc("PUT /api/hooks/settings", sv.apiHooksSettings)
	mux.HandleFunc("POST /api/hooks/tunnel", sv.apiHooksTunnel)
	mux.HandleFunc("GET /api/p/{p}/deploy", sv.apiDeployList)
	mux.HandleFunc("POST /api/p/{p}/deploy/{t}/run", sv.apiDeployRun)

	mux.HandleFunc("POST /api/p/{p}/webhook-test", sv.apiWebhookTest)
	mux.HandleFunc("GET /api/p/{p}/ignore", sv.apiIgnoreGet)
	mux.HandleFunc("PUT /api/p/{p}/ignore", sv.apiIgnorePut)
	mux.HandleFunc("GET /api/ignore-templates", func(w http.ResponseWriter, r *http.Request) { guiJSON(w, ignoreTemplates) })
}

// ── Panel içi dosya / klasör seçici ──────────────────────────────────────────

type fsEntry struct {
	Name string `json:"name"`
	Dir  bool   `json:"dir"`
	Size int64  `json:"size,omitempty"`
}

// apiFSList yerel dosya sistemini listeler. path boşsa kök (Windows: sürücüler, diğer: / ve ev dizini).
func (sv *guiServer) apiFSList(w http.ResponseWriter, r *http.Request) {
	p := strings.TrimSpace(r.URL.Query().Get("path"))
	files := r.URL.Query().Get("files") == "1"
	hidden := r.URL.Query().Get("hidden") == "1"
	home, _ := os.UserHomeDir()
	if p == "" {
		var roots []fsEntry
		if runtime.GOOS == "windows" {
			for c := 'C'; c <= 'Z'; c++ {
				d := string(c) + `:\`
				if st, err := os.Stat(d); err == nil && st.IsDir() {
					roots = append(roots, fsEntry{Name: d, Dir: true})
				}
			}
			for c := 'A'; c <= 'B'; c++ {
				d := string(c) + `:\`
				if st, err := os.Stat(d); err == nil && st.IsDir() {
					roots = append([]fsEntry{{Name: d, Dir: true}}, roots...)
				}
			}
		} else {
			roots = append(roots, fsEntry{Name: "/", Dir: true})
		}
		guiJSON(w, map[string]any{"path": "", "parent": "", "entries": roots, "home": home, "sep": string(filepath.Separator)})
		return
	}
	p = filepath.Clean(p)
	if runtime.GOOS == "windows" && len(p) == 2 && p[1] == ':' {
		p += `\`
	}
	ents, err := os.ReadDir(p)
	if err != nil {
		guiErr(w, 400, err)
		return
	}
	out := []fsEntry{}
	for _, e := range ents {
		name := e.Name()
		if !hidden && strings.HasPrefix(name, ".") && name != ".syncftp" {
			continue
		}
		isDir := e.IsDir()
		if !isDir && e.Type()&os.ModeSymlink != 0 {
			if st, err := os.Stat(filepath.Join(p, name)); err == nil && st.IsDir() {
				isDir = true
			}
		}
		if !isDir && !files {
			continue
		}
		fe := fsEntry{Name: name, Dir: isDir}
		if !isDir {
			if info, err := e.Info(); err == nil {
				fe.Size = info.Size()
			}
		}
		out = append(out, fe)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Dir != out[j].Dir {
			return out[i].Dir
		}
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	parent := filepath.Dir(p)
	if parent == p {
		parent = "" // kökteyiz → sürücü listesine dön
	}
	guiJSON(w, map[string]any{
		"path": p, "parent": parent, "entries": out, "home": home, "sep": string(filepath.Separator),
		"has_config": projects.HasConfig(p),
	})
}

// apiLocalLs proje yerel ağacını (sunucunun yerel dizinine göreceli) ignore bilgisiyle listeler.
// server boşsa proje varsayılan dizini kullanılır (global ayarlar için).
func (sv *guiServer) apiLocalLs(w http.ResponseWriter, r *http.Request) {
	p, cfg, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	localPath := cfg.Project.EffectiveLocalPath()
	if srv, ok := findServer(cfg, r.URL.Query().Get("server")); ok {
		localPath = srv.EffectiveLocalPath(cfg.Project)
	}
	root := filepath.Join(p.Path, localPath)
	rel := strings.Trim(path.Clean("/"+r.URL.Query().Get("path")), "/")
	full := filepath.Join(root, filepath.FromSlash(rel))
	if rc, err := filepath.Rel(root, full); err != nil || strings.HasPrefix(rc, "..") {
		guiErr(w, 400, errors.New("path"))
		return
	}
	matcher, _ := ignore.Load(root, cfg.Sync.IgnoreFiles)
	listLocalTree(w, root, rel, full, matcher)
}

// apiFSTree proje gerektirmeyen yerel ağaç (hızlı yükleme / varsayılanlar için): root mutlak klasör,
// path ona göreceli. Ignore bilgisi klasörün kendi .gitignore + .syncftp/syncftp.ignore'undan.
func (sv *guiServer) apiFSTree(w http.ResponseWriter, r *http.Request) {
	root := filepath.Clean(strings.TrimSpace(r.URL.Query().Get("root")))
	if st, err := os.Stat(root); err != nil || !st.IsDir() || !filepath.IsAbs(root) {
		guiErr(w, 400, fmt.Errorf(lang.L.GuiDirNotFoundFmt, root))
		return
	}
	rel := strings.Trim(path.Clean("/"+r.URL.Query().Get("path")), "/")
	full := filepath.Join(root, filepath.FromSlash(rel))
	if rc, err := filepath.Rel(root, full); err != nil || strings.HasPrefix(rc, "..") {
		guiErr(w, 400, errors.New("path"))
		return
	}
	matcher, _ := ignore.Load(root, nil)
	listLocalTree(w, root, rel, full, matcher)
}

// listLocalTree bir dizinin girdilerini (klasörler önce, ignore işaretli) döndürür.
func listLocalTree(w http.ResponseWriter, root, rel, full string, matcher *ignore.Matcher) {
	ents, err := localDirEntries(full) // CLI yerel tarayıcısıyla aynı sıralama/eleme
	if err != nil {
		guiErr(w, 400, err)
		return
	}
	type le struct {
		Name    string `json:"name"`
		Dir     bool   `json:"dir"`
		Size    int64  `json:"size,omitempty"`
		Ignored bool   `json:"ignored,omitempty"`
	}
	out := []le{}
	for _, e := range ents {
		name := e.Name()
		er := name
		if rel != "" {
			er = rel + "/" + name
		}
		x := le{Name: name, Dir: e.IsDir()}
		if x.Dir {
			x.Ignored = matcher.MatchDir(er)
		} else {
			x.Ignored = matcher.Match(er)
			if info, err := e.Info(); err == nil {
				x.Size = info.Size()
			}
		}
		out = append(out, x)
	}
	guiJSON(w, map[string]any{"root": root, "path": rel, "entries": out})
}

// ── Uzak dosya işlemleri ──────────────────────────────────────────────────────

func (sv *guiServer) apiRemoteUpload(w http.ResponseWriter, r *http.Request) {
	pname, srv, ok := sv.remoteServer(w, r)
	if !ok {
		return
	}
	dir := guiRemotePath(srv, r.URL.Query().Get("path"))
	r.Body = http.MaxBytesReader(w, r.Body, guiUploadMax)
	if err := r.ParseMultipartForm(64 << 20); err != nil {
		guiErr(w, 400, err)
		return
	}
	defer r.MultipartForm.RemoveAll() // diske taşan parçalar (multipart-*) temp'te kalmasın
	var done []string
	for _, fh := range r.MultipartForm.File["files"] {
		local, cleanup, err := multipartLocalFile(fh)
		if err != nil {
			guiErr(w, 500, err)
			return
		}
		name := path.Base(strings.ReplaceAll(fh.Filename, `\`, "/"))
		target := path.Join(dir, name)
		err = withRemote(pname, srv, func(c *ftpclient.Client) error { return c.UploadTo(local, target) })
		cleanup()
		if err != nil {
			guiErr(w, 502, fmt.Errorf("%s: %w", name, err))
			return
		}
		done = append(done, name)
	}
	guiJSON(w, done)
}

// guiUploadMax panelden tek istekte yüklenebilecek toplam boyut.
const guiUploadMax = 4 << 30

// multipartLocalFile parçanın diskteki yolunu döner: büyük parçalar zaten temp dosyadadır (ek kopya
// yok); bellekte tutulan küçük parçalar geçici dosyaya yazılır. cleanup'ı her durumda çağır.
func multipartLocalFile(fh *multipart.FileHeader) (string, func(), error) {
	src, err := fh.Open()
	if err != nil {
		return "", func() {}, err
	}
	if f, ok := src.(*os.File); ok {
		return f.Name(), func() { f.Close() }, nil // RemoveAll siler
	}
	defer src.Close()
	tmp, err := os.CreateTemp("", "syncftp-up-*")
	if err != nil {
		return "", func() {}, err
	}
	_, cerr := io.Copy(tmp, src)
	tmp.Close()
	if cerr != nil {
		os.Remove(tmp.Name())
		return "", func() {}, cerr
	}
	return tmp.Name(), func() { os.Remove(tmp.Name()) }, nil
}

func (sv *guiServer) apiRemoteMkdir(w http.ResponseWriter, r *http.Request) {
	pname, srv, ok := sv.remoteServer(w, r)
	if !ok {
		return
	}
	full := guiRemotePath(srv, r.URL.Query().Get("path"))
	if err := withRemote(pname, srv, func(c *ftpclient.Client) error { return c.MakeDirAll(full) }); err != nil {
		guiErr(w, 502, err)
		return
	}
	guiJSON(w, true)
}

func (sv *guiServer) apiRemoteRename(w http.ResponseWriter, r *http.Request) {
	pname, srv, ok := sv.remoteServer(w, r)
	if !ok {
		return
	}
	from := guiRemotePath(srv, r.URL.Query().Get("from"))
	to := guiRemotePath(srv, r.URL.Query().Get("to"))
	if path.Clean(from) == path.Clean(srv.RemotePath) {
		guiErr(w, 400, errors.New("root"))
		return
	}
	err := withRemote(pname, srv, func(c *ftpclient.Client) error { return remoteMove(c, from, to) })
	if err != nil {
		guiErr(w, 502, err)
		return
	}
	guiJSON(w, true)
}

// ── CI/CD ─────────────────────────────────────────────────────────────────────

func (sv *guiServer) apiHooks(w http.ResponseWriter, r *http.Request) {
	st := loadGUIPersist()
	tun := sv.hooks.tunnelInfo()
	bin := findCloudflared()
	guiJSON(w, map[string]any{
		"local_url":   sv.hooks.localURL(),
		"public_url":  sv.hooks.publicURL(),
		"custom_url":  st.HookPublicURL,
		"auto_tunnel": st.HookTunnel,
		"tunnel":      tun,
		"cloudflared": bin,
		"can_install": runtime.GOOS == "windows" || runtime.GOOS == "linux",
		"deliveries":  sv.hooks.recent(""),
		"providers":   config.DeployProviders,
	})
}

func (sv *guiServer) apiHooksSettings(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CustomURL  string `json:"custom_url"`
		AutoTunnel bool   `json:"auto_tunnel"`
	}
	if err := guiDecode(r, &req); err != nil {
		guiErr(w, 400, err)
		return
	}
	updateGUIPersist(func(st *guiPersist) {
		st.HookPublicURL = strings.TrimRight(strings.TrimSpace(req.CustomURL), "/")
		st.HookTunnel = req.AutoTunnel
	})
	guiJSON(w, true)
}

// apiHooksTunnel {on:true} tüneli açar (cloudflared yoksa install:true ile önce indirir), {on:false} kapatır.
func (sv *guiServer) apiHooksTunnel(w http.ResponseWriter, r *http.Request) {
	var req struct {
		On      bool `json:"on"`
		Install bool `json:"install"`
	}
	if err := guiDecode(r, &req); err != nil {
		guiErr(w, 400, err)
		return
	}
	if !req.On {
		sv.hooks.stopTunnel()
		guiJSON(w, map[string]any{"active": false})
		return
	}
	if req.Install && findCloudflared() == "" {
		j, err := sv.jobs.start("install", "", "cloudflared", lang.L.GuiJobInstall, "install-cloudflared", func(j *job) (any, error) {
			if _, err := installCloudflared(j); err != nil {
				return nil, err
			}
			u, err := sv.hooks.startTunnel()
			if err == nil {
				fmt.Fprintf(j, lang.L.DeployTunnelFmt, u)
			}
			return map[string]string{"url": u}, err
		})
		if err != nil {
			guiErr(w, 409, err)
			return
		}
		guiJSON(w, map[string]any{"job": j.ID})
		return
	}
	u, err := sv.hooks.startTunnel()
	if err != nil {
		guiErr(w, 400, err)
		return
	}
	guiJSON(w, map[string]any{"active": true, "url": u})
}

func (sv *guiServer) apiDeployList(w http.ResponseWriter, r *http.Request) {
	p, cfg, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	base := sv.hooks.publicURL()
	type item struct {
		config.Deploy
		URL string `json:"url"`
	}
	out := []item{}
	for _, d := range cfg.Deploy {
		out = append(out, item{Deploy: d, URL: hookURL(base, p.Name, d.Name)})
	}
	guiJSON(w, map[string]any{"triggers": out, "deliveries": sv.hooks.recent(p.Name), "public_url": base, "local_only": strings.HasPrefix(base, "http://127.0.0.1")})
}

// apiDeployRun tetikleyiciyi webhook beklemeden elle çalıştırır ("şimdi deploy et").
func (sv *guiServer) apiDeployRun(w http.ResponseWriter, r *http.Request) {
	p, cfg, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	for _, d := range cfg.Deploy {
		if d.Name == r.PathValue("t") {
			id, err := sv.hooks.runDeploy(p, d, "manual")
			if err != nil {
				guiErr(w, 409, err)
				return
			}
			sv.hooks.record(delivery{Time: time.Now(), Project: p.Name, Trigger: d.Name, Event: "manual", Status: "accepted", JobID: id})
			guiJSON(w, map[string]any{"job": id})
			return
		}
	}
	guiErr(w, 404, errors.New("trigger"))
}

// ── Giden webhook testi ───────────────────────────────────────────────────────

func (sv *guiServer) apiWebhookTest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL string `json:"url"`
	}
	if err := guiDecode(r, &req); err != nil || strings.TrimSpace(req.URL) == "" {
		guiErr(w, 400, errors.New("url"))
		return
	}
	// sendWebhook ile aynı istemci + gövde; başarı HTTP durum koduyla belirlenir (metinden tahmin yok)
	u := strings.TrimSpace(req.URL)
	payload := webhookPayload(u, fmt.Sprintf(lang.L.WebhookTestFmt, r.PathValue("p")), 1, 0, []string{"index.php"})
	resp, err := webhookClient.Post(u, "application/json", strings.NewReader(payload))
	if err != nil {
		guiErr(w, 502, errors.New(strings.TrimSpace(fmt.Sprintf(lang.L.SyncWebhookErrFmt, err))))
		return
	}
	resp.Body.Close()
	out := strings.TrimSpace(fmt.Sprintf(lang.L.SyncWebhookSentFmt, u, resp.StatusCode))
	if resp.StatusCode >= 400 {
		guiErr(w, 502, errors.New(out))
		return
	}
	guiJSON(w, out)
}

// ── .syncftp/syncftp.ignore editörü ───────────────────────────────────────────

// ignoreTargets sync'in gerçekten okuduğu ignore dosyaları: her farklı sunucu yerel dizini için
// <yerel dizin>/<ignore.EditableFile>. Eskiden hep <proje>/.syncftp/syncftp.ignore düzenleniyordu —
// local_path "." değilse ya da ignore_files özelse sync o dosyayı hiç okumuyordu.
type ignoreTarget struct {
	Dir       string   `json:"dir"`  // proje köküne göreceli yerel dizin
	File      string   `json:"file"` // dizine göreceli ignore dosyası ("" → ignore_files'ta syncFTP dosyası yok)
	Content   string   `json:"content"`
	Gitignore string   `json:"gitignore"`
	Servers   []string `json:"servers"`
}

func ignoreTargets(projectDir string, cfg *config.Config) []ignoreTarget {
	file := ignore.EditableFile(cfg.Sync.IgnoreFiles)
	var out []ignoreTarget
	idx := map[string]int{}
	add := func(dir, srvName string) {
		key := filepath.ToSlash(filepath.Clean(dir))
		if i, ok := idx[key]; ok {
			if srvName != "" {
				out[i].Servers = append(out[i].Servers, srvName)
			}
			return
		}
		t := ignoreTarget{Dir: key, File: file}
		base := filepath.Join(projectDir, dir)
		if file != "" {
			data, _ := os.ReadFile(filepath.Join(base, filepath.FromSlash(file)))
			t.Content = strings.ReplaceAll(string(data), "\r\n", "\n")
		}
		gi, _ := os.ReadFile(filepath.Join(base, ".gitignore"))
		t.Gitignore = strings.ReplaceAll(string(gi), "\r\n", "\n")
		if srvName != "" {
			t.Servers = []string{srvName}
		}
		idx[key] = len(out)
		out = append(out, t)
	}
	for _, s := range cfg.Servers {
		add(s.EffectiveLocalPath(cfg.Project), s.Name)
	}
	if len(out) == 0 {
		add(cfg.Project.EffectiveLocalPath(), "")
	}
	return out
}

func (sv *guiServer) apiIgnoreGet(w http.ResponseWriter, r *http.Request) {
	p, cfg, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	guiJSON(w, map[string]any{"targets": ignoreTargets(p.Path, cfg)})
}

func (sv *guiServer) apiIgnorePut(w http.ResponseWriter, r *http.Request) {
	p, cfg, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	var req struct {
		Dir     string `json:"dir"`
		Content string `json:"content"`
	}
	if err := guiDecode(r, &req); err != nil {
		guiErr(w, 400, err)
		return
	}
	for _, t := range ignoreTargets(p.Path, cfg) {
		if t.Dir != req.Dir {
			continue
		}
		if t.File == "" {
			guiErr(w, 400, errors.New(lang.L.GuiIgnoreNoFile))
			return
		}
		full := filepath.Join(p.Path, filepath.FromSlash(t.Dir), filepath.FromSlash(t.File))
		werr := os.MkdirAll(filepath.Dir(full), 0755)
		if werr == nil {
			werr = os.WriteFile(full, []byte(req.Content), 0644)
		}
		if werr != nil {
			err = werr
			guiErr(w, 500, err)
			return
		}
		guiJSON(w, true)
		return
	}
	guiErr(w, 400, errors.New("dir"))
}

// apiLocalFiles bir yerel klasörün altındaki (ignore edilmeyen) tüm dosyaların göreceli yollarını
// döner — freeze listesine klasör eklemek için (freeze dosya yolu bekler). En fazla 20000.
func (sv *guiServer) apiLocalFiles(w http.ResponseWriter, r *http.Request) {
	p, cfg, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	srv, ok := findServer(cfg, r.URL.Query().Get("server"))
	if !ok {
		guiErr(w, 404, fmt.Errorf(lang.L.ServerNotFoundFmt, r.URL.Query().Get("server")))
		return
	}
	root := filepath.Join(p.Path, srv.EffectiveLocalPath(cfg.Project))
	rel := strings.Trim(path.Clean("/"+r.URL.Query().Get("path")), "/")
	matcher, _ := ignore.Load(root, cfg.Sync.IgnoreFiles)
	files, _ := scanner.List(root, rel, matcher, 20000) // sync taramasıyla aynı kurallar (hash'siz)
	out := make([]string, len(files))
	for i, f := range files {
		out[i] = f.RelPath
	}
	guiJSON(w, out)
}

// apiHelp CLI komut referansını doğrudan cobra tanımlarından üretir
// (panel yardımı ile `syncftp help` her zaman aynı kalır).
func (sv *guiServer) apiHelp(w http.ResponseWriter, r *http.Request) {
	type cmdHelp struct {
		Path  string `json:"path"`
		Use   string `json:"use"`
		Short string `json:"short"`
		Long  string `json:"long,omitempty"`
		Flags string `json:"flags,omitempty"`
	}
	var out []cmdHelp
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		if c.Hidden || c.Name() == "help" || c.Name() == "completion" {
			return
		}
		if c != rootCmd {
			out = append(out, cmdHelp{
				Path: c.CommandPath(), Use: c.UseLine(), Short: c.Short, Long: strings.TrimSpace(c.Long),
				Flags: strings.TrimRight(c.LocalFlags().FlagUsages(), "\n"),
			})
		}
		for _, sub := range c.Commands() {
			walk(sub)
		}
	}
	walk(rootCmd)
	guiJSON(w, map[string]any{
		"commands":     out,
		"global_flags": strings.TrimRight(rootCmd.PersistentFlags().FlagUsages(), "\n"),
		"shell":        lang.L.ShellHelp,
		"version":      appVersion,
	})
}

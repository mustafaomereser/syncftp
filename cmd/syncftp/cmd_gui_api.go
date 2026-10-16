package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"syncftp/internal/config"
	"syncftp/internal/failed"
	ftpclient "syncftp/internal/ftp"
	"syncftp/internal/frozen"
	"syncftp/internal/lang"
	"syncftp/internal/milestone"
	"syncftp/internal/projects"
	"syncftp/internal/release"
	"syncftp/internal/state"
	"syncftp/internal/synclog"
)

// appVersion — GUI ve `syncftp --version`.
const appVersion = "2.0.0"

// keepSecret — GUI'ye şifre yerine gönderilen işaret; kayıtta aynen dönerse mevcut şifre korunur.
const keepSecret = "__keep__"

func (sv *guiServer) routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/ping", func(w http.ResponseWriter, r *http.Request) {
		guiJSON(w, map[string]any{"jobs": sv.jobs.running(), "watches": sv.watches.count()})
	})
	mux.HandleFunc("GET /api/presence", sv.apiPresence)
	mux.HandleFunc("GET /api/meta", sv.apiMeta)
	mux.HandleFunc("PUT /api/lang", sv.apiLang)
	mux.HandleFunc("POST /api/quit", func(w http.ResponseWriter, r *http.Request) {
		guiJSON(w, true)
		go func() { time.Sleep(300 * time.Millisecond); sv.shutdown() }()
	})

	mux.HandleFunc("GET /api/projects", sv.apiProjects)
	mux.HandleFunc("POST /api/projects", sv.apiProjectAdd)
	mux.HandleFunc("PATCH /api/projects/{p}", sv.apiProjectRename)
	mux.HandleFunc("DELETE /api/projects/{p}", sv.apiProjectRemove)

	mux.HandleFunc("GET /api/p/{p}/overview", sv.apiOverview)
	mux.HandleFunc("GET /api/p/{p}/status", sv.apiStatus)
	mux.HandleFunc("GET /api/p/{p}/config", sv.apiConfigGet)
	mux.HandleFunc("PUT /api/p/{p}/config", sv.apiConfigPut)
	mux.HandleFunc("POST /api/p/{p}/test", sv.apiTestConn)
	mux.HandleFunc("POST /api/p/{p}/keyring", sv.apiKeyring)
	mux.HandleFunc("POST /api/p/{p}/jobs", sv.apiJobStart)

	mux.HandleFunc("GET /api/p/{p}/releases", sv.apiReleases)
	mux.HandleFunc("GET /api/p/{p}/releases/{srv}/{id}", sv.apiReleaseDetail)
	mux.HandleFunc("GET /api/p/{p}/logs", sv.apiLogs)
	mux.HandleFunc("GET /api/p/{p}/logs/{srv}/{file}", sv.apiLogFile)
	mux.HandleFunc("GET /api/p/{p}/milestones", sv.apiMilestones)
	mux.HandleFunc("PUT /api/p/{p}/milestones", sv.apiMilestoneSet)
	mux.HandleFunc("DELETE /api/p/{p}/milestones", sv.apiMilestoneClear)
	mux.HandleFunc("GET /api/p/{p}/frozen", sv.apiFrozenGet)
	mux.HandleFunc("PUT /api/p/{p}/frozen", sv.apiFrozenPut)
	mux.HandleFunc("GET /api/p/{p}/failed", sv.apiFailed)

	// Uzak dosyalar — bağlantı önbellekli (cmd_gui_remote.go)
	mux.HandleFunc("GET /api/p/{p}/remote/ls", sv.apiRemoteLs)
	mux.HandleFunc("GET /api/p/{p}/remote/cat", sv.apiRemoteCat)
	mux.HandleFunc("GET /api/p/{p}/remote/get", sv.apiRemoteGet)
	mux.HandleFunc("DELETE /api/p/{p}/remote/rm", sv.apiRemoteRm)
	mux.HandleFunc("GET /api/p/{p}/local/cat", sv.apiLocalCat)

	mux.HandleFunc("GET /api/p/{p}/watch", sv.apiWatchGet)
	mux.HandleFunc("POST /api/p/{p}/watch", sv.apiWatchStart)
	mux.HandleFunc("DELETE /api/p/{p}/watch", sv.apiWatchStop)
	mux.HandleFunc("GET /api/watches", func(w http.ResponseWriter, r *http.Request) { guiJSON(w, sv.watches.all()) })

	mux.HandleFunc("GET /api/jobs", func(w http.ResponseWriter, r *http.Request) {
		guiJSON(w, sv.jobs.list(r.URL.Query().Get("project")))
	})
	sv.routesMore(mux)
	sv.routesAI(mux)
	mux.HandleFunc("GET /api/jobs/{id}", sv.apiJobGet)
	mux.HandleFunc("GET /api/jobs/{id}/stream", sv.apiJobStream)
	mux.HandleFunc("DELETE /api/jobs/{id}", sv.apiJobCancel)
}

// ── proje çözümleme ───────────────────────────────────────────────────────────

func (sv *guiServer) project(r *http.Request) (projects.Project, *config.Config, error) {
	p, ok := projects.Find(r.PathValue("p"))
	if !ok {
		return p, nil, fmt.Errorf(lang.L.ProjectNotFoundFmt, r.PathValue("p"))
	}
	cfg, err := config.Load(p.Path)
	return p, cfg, err
}

func findServer(cfg *config.Config, name string) (config.Server, bool) {
	for _, s := range cfg.Servers {
		if s.Name == name {
			return s, true
		}
	}
	return config.Server{}, false
}

// apiPresence panel penceresi açık olduğu sürece açık kalan SSE bağlantısı.
func (sv *guiServer) apiPresence(w http.ResponseWriter, r *http.Request) {
	fl, ok := w.(http.Flusher)
	if !ok {
		guiErr(w, 500, errors.New("sse"))
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, ": hello\n\n")
	fl.Flush()
	sv.clients.Add(1)
	nav := make(chan string, 4)
	sv.navMu.Lock()
	if sv.navSubs == nil {
		sv.navSubs = map[chan string]struct{}{}
	}
	sv.navSubs[nav] = struct{}{}
	sv.navMu.Unlock()
	defer func() {
		sv.clients.Add(-1)
		sv.lastSeen.Store(time.Now().UnixNano())
		sv.navMu.Lock()
		delete(sv.navSubs, nav)
		sv.navMu.Unlock()
	}()
	t := time.NewTicker(20 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-sv.quit:
			fmt.Fprint(w, "event: bye\ndata: quit\n\n")
			fl.Flush()
			return
		case h := <-nav:
			fmt.Fprintf(w, "event: navigate\ndata: %s\n\n", h)
			fl.Flush()
		case <-t.C:
			fmt.Fprint(w, ": ping\n\n")
			fl.Flush()
		}
	}
}

// ── meta / dil ────────────────────────────────────────────────────────────────

func (sv *guiServer) apiMeta(w http.ResponseWriter, r *http.Request) {
	home, _ := os.UserHomeDir()
	guiJSON(w, map[string]any{
		"version":     appVersion,
		"lang":        lang.Current(),
		"protocols":   config.Protocols,
		"home":        home,
		"known_hosts": ftpclient.KnownHostsPath(),
		"registry":    projects.Path(),
		"start_dir":   sv.startDir,
		"base":        projects.Base(),
	})
}

func (sv *guiServer) apiLang(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Lang string `json:"lang"`
	}
	if err := guiDecode(r, &req); err != nil || (req.Lang != "en" && req.Lang != "tr") {
		guiErr(w, 400, errors.New("lang: en | tr"))
		return
	}
	lang.Set(req.Lang)
	saveGUILang(req.Lang)
	guiJSON(w, req.Lang)
}

// ── projeler ──────────────────────────────────────────────────────────────────

type guiProject struct {
	Name     string     `json:"name"`
	Path     string     `json:"path"`
	OK       bool       `json:"ok"`
	Error    string     `json:"error,omitempty"`
	Servers  int        `json:"servers"`
	Enabled  int        `json:"enabled"`
	Watching bool       `json:"watching"`
	LastUsed *time.Time `json:"last_used,omitempty"`
	Keyring  bool       `json:"keyring"`
}

func (sv *guiServer) apiProjects(w http.ResponseWriter, r *http.Request) {
	out := []guiProject{}
	for _, p := range projects.List() {
		gp := guiProject{Name: p.Name, Path: p.Path, Watching: sv.watches.get(p.Name) != nil}
		if !p.LastUsed.IsZero() {
			t := p.LastUsed
			gp.LastUsed = &t
		}
		if cfg, err := config.Load(p.Path); err != nil {
			gp.Error = err.Error()
		} else {
			gp.OK = true
			gp.Servers = len(cfg.Servers)
			gp.Enabled = len(cfg.EnabledServers())
			gp.Keyring = cfg.Project.Keyring
		}
		out = append(out, gp)
	}
	sort.SliceStable(out, func(i, j int) bool { return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name) })
	guiJSON(w, out)
}

// apiProjectAdd mevcut projeyi kaydeder; create=true ve config yoksa boş bir config oluşturur.
func (sv *guiServer) apiProjectAdd(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path       string `json:"path"`
		Name       string `json:"name"`
		Create     bool   `json:"create"`
		Template   string `json:"template"`   // "" → global varsayılan, "none" → yok
		Connection string `json:"connection"` // kütüphaneden ilk sunucu (boş → sunucusuz)
		Remote     string `json:"remote"`
		ServerName string `json:"server_name"`
	}
	if err := guiDecode(r, &req); err != nil || strings.TrimSpace(req.Path) == "" {
		guiErr(w, 400, errors.New(lang.L.GuiPathRequired))
		return
	}
	dir := filepath.Clean(strings.TrimSpace(req.Path))
	if st, err := os.Stat(dir); err != nil || !st.IsDir() {
		guiErr(w, 400, fmt.Errorf(lang.L.GuiDirNotFoundFmt, dir))
		return
	}
	if !projects.InBase(dir) {
		guiErr(w, 400, fmt.Errorf(lang.L.GuiOutsideBaseFmt, dir, projects.Base()))
		return
	}
	if !projects.HasConfig(dir) {
		if !req.Create {
			guiJSON(w, map[string]any{"needs_create": true, "path": dir})
			return
		}
		var srv *config.Server
		if req.Connection != "" {
			lib, _ := config.LoadLibrary()
			c, ok := lib.FindConnection(req.Connection)
			if !ok {
				guiErr(w, 400, fmt.Errorf(lang.L.GlobalConnNotFoundFmt, req.Connection))
				return
			}
			s := lib.Defaults.NewServer(orDefault(strings.TrimSpace(req.ServerName), "production"))
			s.ApplyConnection(c)
			s.RemotePath = orDefault(strings.TrimSpace(req.Remote), "/")
			srv = &s
		}
		if _, _, err := createProject(io.Discard, dir, newProjectOpts{Name: req.Name, Template: req.Template, Server: srv}); err != nil {
			guiErr(w, 500, err)
			return
		}
	}
	p, err := projects.Add(dir, req.Name)
	if err != nil {
		guiErr(w, 400, err)
		return
	}
	guiJSON(w, p)
}

func (sv *guiServer) apiProjectRename(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := guiDecode(r, &req); err != nil || strings.TrimSpace(req.Name) == "" {
		guiErr(w, 400, errors.New("name"))
		return
	}
	// Klasör modu: görünen ad config'teki project.name
	p, ok := projects.Find(r.PathValue("p"))
	if !ok {
		guiErr(w, 404, errors.New("project"))
		return
	}
	cfg, err := config.Load(p.Path)
	if err != nil {
		guiErr(w, 400, err)
		return
	}
	cfg.Project.Name = strings.TrimSpace(req.Name)
	if err := config.Save(p.Path, cfg); err != nil {
		guiErr(w, 500, err)
		return
	}
	projects.Invalidate()
	guiJSON(w, true)
}

func (sv *guiServer) apiProjectRemove(w http.ResponseWriter, r *http.Request) {
	sv.watches.stop(r.PathValue("p"))
	if err := projects.Remove(r.PathValue("p")); err != nil {
		guiErr(w, 404, err)
		return
	}
	guiJSON(w, true)
}

// ── genel bakış / durum ───────────────────────────────────────────────────────

type guiServerInfo struct {
	Name           string     `json:"name"`
	Protocol       string     `json:"protocol"`
	Host           string     `json:"host"`
	Port           int        `json:"port"`
	User           string     `json:"user"`
	RemotePath     string     `json:"remote_path"`
	LocalPath      string     `json:"local_path"`
	Connection     string     `json:"connection,omitempty"`
	Enabled        bool       `json:"enabled"`
	MaxConnections int        `json:"max_connections"`
	Bandwidth      int        `json:"bandwidth"`
	Tunnel         string     `json:"tunnel,omitempty"`
	Backup         bool       `json:"backup"`
	Atomic         bool       `json:"atomic"`
	FirstSyncDone  bool       `json:"first_sync_done"`
	LastSync       *time.Time `json:"last_sync,omitempty"`
	Files          int        `json:"files"`
	New            int        `json:"new"`
	Changed        int        `json:"changed"`
	Deleted        int        `json:"deleted"`
	Milestone      string     `json:"milestone,omitempty"`
	Frozen         int        `json:"frozen"`
	Failed         int        `json:"failed"`
	Releases       int        `json:"releases"`
	Rollbackable   int        `json:"rollbackable"`
	LastRelease    string     `json:"last_release,omitempty"`
	Busy           string     `json:"busy,omitempty"`
	ScanError      string     `json:"scan_error,omitempty"`
}

// pendingFor sync'in gerçekten yükleyeceği dosyalar: diff → include → exclude →
// milestone → frozen → block (sync akışıyla aynı sıra). Silinenler ham döner.
func pendingFor(configDir string, cfg *config.Config, srv config.Server, current map[string]string, st *state.State) (newL, changedL, deleted []string) {
	diff := state.Diff(st, current)
	localPath := srv.EffectiveLocalPath(cfg.Project)
	include := cfg.Sync.Include
	if len(srv.Include) > 0 {
		include = srv.Include
	}
	include = normalizePatterns(include, localPath)
	exclude := normalizePatterns(append(append([]string{}, cfg.Sync.Exclude...), srv.Exclude...), localPath)
	blocks := append(append([]string{}, cfg.Sync.BlockExtensions...), srv.BlockExtensions...)
	fl, _ := frozen.Load(configDir, srv.Name)
	ms, _ := milestone.Load(configDir, srv.Name)
	localDir := filepath.Join(configDir, localPath)
	filter := func(in []string) []string {
		out := filterByExclude(filterByInclude(in, include), exclude)
		if ms != nil {
			out, _ = milestoneFilterMtime(out, localDir, ms.Date)
		}
		if fl != nil && len(fl.Files) > 0 {
			var kept []string
			for _, p := range out {
				if !frozen.IsFrozen(fl, p) {
					kept = append(kept, p)
				}
			}
			out = kept
		}
		out, _ = filterByBlock(out, blocks)
		sort.Strings(out)
		return out
	}
	sort.Strings(diff.Deleted)
	return filter(diff.New), filter(diff.Changed), diff.Deleted
}

func (sv *guiServer) serverInfo(p projects.Project, cfg *config.Config, srv config.Server, busy map[string]string, scan bool) guiServerInfo {
	gi := guiServerInfo{
		Name: srv.Name, Protocol: srv.ProtocolName(), Host: srv.Host, Port: srv.Port, User: srv.User,
		RemotePath: srv.RemotePath, LocalPath: srv.EffectiveLocalPath(cfg.Project), Connection: srv.Connection,
		Enabled: srv.Enabled, MaxConnections: max(srv.MaxConnections, 1), Bandwidth: srv.EffectiveBandwidth(cfg.Sync),
		Backup: srv.BackupEnabled(cfg.Sync), Atomic: !srv.DirectUpload, Busy: busy[srv.Name],
	}
	if srv.HasTunnel() {
		gi.Tunnel = fmt.Sprintf("%s:%d", srv.TunnelHost, srv.TunnelPort)
	}
	st, err := state.Load(p.Path, srv.Name)
	if err == nil {
		gi.FirstSyncDone = st.FirstSyncDone
		if !st.LastSync.IsZero() {
			t := st.LastSync
			gi.LastSync = &t
		}
	}
	if ms, _ := milestone.Load(p.Path, srv.Name); ms != nil {
		gi.Milestone = ms.Date.Format("2006-01-02 15:04")
	}
	if fl, _ := frozen.Load(p.Path, srv.Name); fl != nil {
		gi.Frozen = len(fl.Files)
	}
	if f, _ := failed.Load(p.Path, srv.Name); f != nil {
		gi.Failed = len(f.Files)
	}
	if list, _ := release.List(p.Path, srv.Name); len(list) > 0 {
		gi.Releases = len(list)
		gi.LastRelease = list[0].CreatedAt.Format("2006-01-02 15:04")
		for _, m := range list {
			if m.HasBackup() && !m.RolledBack {
				gi.Rollbackable++
			}
		}
	}
	if scan && err == nil {
		current, _, serr := scanForServer(p.Path, cfg, srv)
		if serr != nil {
			gi.ScanError = serr.Error()
		} else {
			gi.Files = len(current)
			n, c, d := pendingFor(p.Path, cfg, srv, current, st)
			gi.New, gi.Changed, gi.Deleted = len(n), len(c), len(d)
		}
	}
	return gi
}

func (sv *guiServer) apiOverview(w http.ResponseWriter, r *http.Request) {
	p, cfg, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	scan := r.URL.Query().Get("scan") != "0"
	busy := sv.jobs.busyServers(p.Name)
	infos := make([]guiServerInfo, len(cfg.Servers))
	var wg sync.WaitGroup
	for i, srv := range cfg.Servers {
		wg.Add(1)
		go func(i int, srv config.Server) {
			defer wg.Done()
			infos[i] = sv.serverInfo(p, cfg, srv, busy, scan && srv.Enabled)
		}(i, srv)
	}
	wg.Wait()
	guiJSON(w, map[string]any{
		"project":       p,
		"name":          cfg.Project.Name,
		"keyring":       cfg.Project.Keyring,
		"secret_errors": cfg.SecretErrors,
		"servers":       infos,
		"watch":         sv.watches.get(p.Name),
	})
}

func (sv *guiServer) apiStatus(w http.ResponseWriter, r *http.Request) {
	p, cfg, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	srv, ok := findServer(cfg, r.URL.Query().Get("server"))
	if !ok {
		guiErr(w, 404, errors.New("server"))
		return
	}
	st, err := state.Load(p.Path, srv.Name)
	if err != nil {
		guiErr(w, 500, err)
		return
	}
	current, _, err := scanForServer(p.Path, cfg, srv)
	if err != nil {
		guiErr(w, 500, err)
		return
	}
	n, c, d := pendingFor(p.Path, cfg, srv, current, st)
	capL := func(l []string) []string {
		if l == nil {
			return []string{}
		}
		if len(l) > 3000 {
			return l[:3000]
		}
		return l
	}
	guiJSON(w, map[string]any{"new": capL(n), "changed": capL(c), "deleted": capL(d),
		"counts": map[string]int{"new": len(n), "changed": len(c), "deleted": len(d)}, "first_sync_done": st.FirstSyncDone})
}

// ── config ────────────────────────────────────────────────────────────────────

func maskSecret(s *string) {
	if *s != "" {
		*s = keepSecret
	}
}

func (sv *guiServer) apiConfigGet(w http.ResponseWriter, r *http.Request) {
	_, cfg, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	out := *cfg
	out.Servers = append([]config.Server(nil), cfg.Servers...)
	out.Connections = append([]config.Connection(nil), cfg.Connections...)
	for i := range out.Servers {
		maskSecret(&out.Servers[i].Password)
		maskSecret(&out.Servers[i].TunnelPassword)
	}
	for i := range out.Connections {
		maskSecret(&out.Connections[i].Password)
		maskSecret(&out.Connections[i].TunnelPassword)
	}
	if out.Servers == nil {
		out.Servers = []config.Server{}
	}
	if out.Connections == nil {
		out.Connections = []config.Connection{}
	}
	guiJSON(w, out)
}

// restoreSecrets "__keep__" olarak dönen şifreleri mevcut config'ten geri koyar.
// Yeniden adlandırma için ad bulunamazsa aynı sıradaki kayıt denenir.
func restoreSecrets(in, old *config.Config) {
	oldSrv := map[string]config.Server{}
	for _, s := range old.Servers {
		oldSrv[s.Name] = s
	}
	oldConn := map[string]config.Connection{}
	for _, c := range old.Connections {
		oldConn[c.Name] = c
	}
	for i := range in.Servers {
		s := &in.Servers[i]
		o, ok := oldSrv[s.Name]
		if !ok && i < len(old.Servers) {
			o, ok = old.Servers[i], true
		}
		if s.Password == keepSecret {
			s.Password = map[bool]string{true: o.Password}[ok]
		}
		if s.TunnelPassword == keepSecret {
			s.TunnelPassword = map[bool]string{true: o.TunnelPassword}[ok]
		}
	}
	for i := range in.Connections {
		c := &in.Connections[i]
		o, ok := oldConn[c.Name]
		if !ok && i < len(old.Connections) {
			o, ok = old.Connections[i], true
		}
		if c.Password == keepSecret {
			c.Password = map[bool]string{true: o.Password}[ok]
		}
		if c.TunnelPassword == keepSecret {
			c.TunnelPassword = map[bool]string{true: o.TunnelPassword}[ok]
		}
	}
}

func validateConfig(cfg *config.Config) error {
	seen := map[string]bool{}
	for _, c := range cfg.Connections {
		if strings.TrimSpace(c.Name) == "" {
			return errors.New(lang.L.GuiValConnName)
		}
		if seen["c:"+c.Name] {
			return fmt.Errorf(lang.L.GuiValDupFmt, c.Name)
		}
		seen["c:"+c.Name] = true
	}
	for _, s := range cfg.Servers {
		if strings.TrimSpace(s.Name) == "" {
			return errors.New(lang.L.GuiValServerName)
		}
		if seen["s:"+s.Name] {
			return fmt.Errorf(lang.L.GuiValDupFmt, s.Name)
		}
		seen["s:"+s.Name] = true
		if s.Connection != "" && !seen["c:"+s.Connection] {
			return fmt.Errorf(lang.L.GuiValConnMissingFmt, s.Name, s.Connection)
		}
		if s.Connection == "" && strings.TrimSpace(s.Host) == "" {
			return fmt.Errorf(lang.L.GuiValHostFmt, s.Name)
		}
		if strings.TrimSpace(s.RemotePath) == "" {
			return fmt.Errorf(lang.L.GuiValRemoteFmt, s.Name)
		}
	}
	return nil
}

func (sv *guiServer) apiConfigPut(w http.ResponseWriter, r *http.Request) {
	p, old, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	var in config.Config
	if err := guiDecode(r, &in); err != nil {
		guiErr(w, 400, err)
		return
	}
	restoreSecrets(&in, old)
	in.Project.Keyring = old.Project.Keyring // anahtarlık modu ayrı uçtan değişir
	if err := validateConfig(&in); err != nil {
		guiErr(w, 400, err)
		return
	}
	// Connection profilini kullanan sunucularda credential'lar profilden gelir (Save sıfırlar)
	if err := config.Save(p.Path, &in); err != nil {
		guiErr(w, 500, err)
		return
	}
	if _, err := config.Load(p.Path); err != nil {
		guiErr(w, 500, err)
		return
	}
	guiJSON(w, true)
}

// apiTestConn düzenlenmekte olan sunucu ayarlarıyla bağlantıyı dener (kaydetmeden).
func (sv *guiServer) apiTestConn(w http.ResponseWriter, r *http.Request) {
	_, old, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	var req struct {
		Server      config.Server       `json:"server"`
		Connections []config.Connection `json:"connections"`
	}
	if err := guiDecode(r, &req); err != nil {
		guiErr(w, 400, err)
		return
	}
	tmp := config.Config{Servers: []config.Server{req.Server}, Connections: req.Connections}
	restoreSecrets(&tmp, old)
	srv := tmp.Servers[0]
	if srv.Connection != "" {
		for _, c := range tmp.Connections {
			if c.Name == srv.Connection {
				srv.Host, srv.Port, srv.User, srv.Password = c.Host, c.Port, c.User, c.Password
				srv.Passive, srv.DisableEPSV, srv.NATWorkaround = c.Passive, c.DisableEPSV, c.NATWorkaround
				srv.Protocol, srv.KeyFile, srv.TLSSkipVerify, srv.SSHTunnel = c.Protocol, c.KeyFile, c.TLSSkipVerify, c.SSHTunnel
			}
		}
	}
	if srv.Port == 0 {
		srv.Port = map[bool]int{true: 22, false: 21}[srv.IsSFTP()]
	}
	if srv.HasTunnel() && srv.TunnelPort == 0 {
		srv.TunnelPort = 22
	}
	start := time.Now()
	c, err := ftpclient.Connect(srv)
	if err != nil {
		guiErr(w, 400, err)
		return
	}
	defer c.Close()
	entries, lerr := c.List(srv.RemotePath)
	res := map[string]any{"ms": time.Since(start).Milliseconds(), "entries": len(entries)}
	if lerr != nil {
		res["list_error"] = lerr.Error()
	}
	guiJSON(w, res)
}

func (sv *guiServer) apiKeyring(w http.ResponseWriter, r *http.Request) {
	p, cfg, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	var req struct {
		On bool `json:"on"`
	}
	if err := guiDecode(r, &req); err != nil {
		guiErr(w, 400, err)
		return
	}
	var buf strings.Builder
	if err := setKeyringMode(&buf, p.Path, cfg, req.On); err != nil {
		guiErr(w, 500, err)
		return
	}
	guiJSON(w, strings.TrimSpace(buf.String()))
}

// ── job'lar ───────────────────────────────────────────────────────────────────

type guiJobReq struct {
	Kind        string   `json:"kind"` // sync | calibrate | verify | rollback | detect | milestone-sync
	Servers     []string `json:"servers"`
	Full        bool     `json:"full"`
	DryRun      bool     `json:"dry_run"`
	RetryFailed bool     `json:"retry_failed"`
	Release     string   `json:"release"`
	Fix         bool     `json:"fix"`
	Max         int      `json:"max"`
	Command     string   `json:"command"`
	Include     []string `json:"include"` // tek seferlik (CLI sync --include gibi; config'teki include'u geçersiz kılar)
	Exclude     []string `json:"exclude"` // tek seferlik (CLI sync --exclude gibi; ek)
}

func (sv *guiServer) apiJobStart(w http.ResponseWriter, r *http.Request) {
	p, cfg, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	var req guiJobReq
	if err := guiDecode(r, &req); err != nil {
		guiErr(w, 400, err)
		return
	}
	var servers []config.Server
	if len(req.Servers) == 0 {
		servers = cfg.EnabledServers()
	} else {
		for _, n := range req.Servers {
			s, ok := findServer(cfg, n)
			if !ok {
				guiErr(w, 404, fmt.Errorf("server: %s", n))
				return
			}
			servers = append(servers, s)
		}
	}
	if len(servers) == 0 {
		guiErr(w, 400, errors.New(lang.L.ShellNoServers))
		return
	}

	var ids []string
	var errs []string
	add := func(j *job, err error) {
		if err != nil {
			errs = append(errs, err.Error())
			return
		}
		ids = append(ids, j.ID)
	}
	dir := p.Path

	switch req.Kind {
	case "sync", "milestone-sync":
		for _, srv := range servers {
			srv := srv
			title := lang.L.GuiJobSync
			if req.DryRun {
				title = lang.L.GuiJobDryRun
			}
			if req.Kind == "milestone-sync" {
				title = lang.L.GuiJobMilestoneSync
			}
			add(sv.jobs.start(req.Kind, p.Name, srv.Name, title, p.Name+"|"+srv.Name, func(j *job) (any, error) {
				return guiRunSync(j, dir, srv, req)
			}))
		}
	case "calibrate":
		for _, srv := range servers {
			srv := srv
			add(sv.jobs.start("calibrate", p.Name, srv.Name, lang.L.GuiJobCalibrate, p.Name+"|"+srv.Name, func(j *job) (any, error) {
				c, err := config.Load(dir)
				if err != nil {
					return nil, err
				}
				runCalibrateOpts(dir, srv, c, calibrateOpts{out: j, saveLog: true})
				return nil, nil
			}))
		}
	case "verify":
		for _, srv := range servers {
			srv := srv
			add(sv.jobs.start("verify", p.Name, srv.Name, lang.L.GuiJobVerify, p.Name+"|"+srv.Name, func(j *job) (any, error) {
				c, err := config.Load(dir)
				if err != nil {
					return nil, err
				}
				return verifyServer(j, dir, c, srv, req.Fix)
			}))
		}
	case "rollback":
		srv := servers[0]
		add(sv.jobs.start("rollback", p.Name, srv.Name, lang.L.GuiJobRollback+" "+req.Release, p.Name+"|"+srv.Name, func(j *job) (any, error) {
			list, err := release.List(dir, srv.Name)
			if err != nil {
				return nil, err
			}
			plan, err := rollbackPlan(list, req.Release)
			if err != nil {
				return nil, err
			}
			var buf strings.Builder
			rep, err := performRollback(io.MultiWriter(j, &buf), dir, srv, plan)
			if err == nil {
				_, _ = synclog.Save(dir, srv.Name, buf.String())
			}
			return rep, err
		}))
	case "detect":
		names := make([]string, len(servers))
		for i, s := range servers {
			names[i] = s.Name
		}
		add(sv.jobs.start("detect", p.Name, strings.Join(names, ", "), lang.L.GuiJobDetect, "", func(j *job) (any, error) {
			c, err := config.Load(dir)
			if err != nil {
				return nil, err
			}
			var fresh []config.Server
			for _, n := range names {
				if s, ok := findServer(c, n); ok {
					fresh = append(fresh, s)
				}
			}
			maxProbe := req.Max
			if maxProbe <= 0 {
				maxProbe = 32
			}
			if detectLimits(j, c, fresh, maxProbe, nil) {
				if err := config.Save(dir, c); err != nil {
					return nil, err
				}
				fmt.Fprint(j, lang.L.DetectSaved)
			}
			out := map[string]int{}
			for _, s := range c.Servers {
				out[s.Name] = s.MaxConnections
			}
			return out, nil
		}))
	case "exec":
		srv := servers[0]
		command := strings.TrimSpace(req.Command)
		if command == "" {
			guiErr(w, 400, errors.New("command"))
			return
		}
		add(sv.jobs.start("exec", p.Name, srv.Name, lang.L.GuiJobExec, "", func(j *job) (any, error) {
			c, err := ftpclient.Connect(srv)
			if err != nil {
				return nil, err
			}
			defer c.Close()
			fmt.Fprintf(j, "$ %s\n", command)
			return nil, c.Exec(command, nil, j)
		}))
	case "diff":
		if len(servers) != 2 {
			guiErr(w, 400, errors.New(strings.TrimSpace(lang.L.DiffNeedTwo)))
			return
		}
		a, b := servers[0], servers[1]
		add(sv.jobs.start("diff", p.Name, a.Name+" ↔ "+b.Name, lang.L.GuiJobDiff, "", func(j *job) (any, error) {
			return diffServers(j, a, b)
		}))
	default:
		guiErr(w, 400, fmt.Errorf("kind: %s", req.Kind))
		return
	}
	if len(ids) == 0 && len(errs) > 0 {
		guiErr(w, 409, errors.New(strings.Join(errs, "; ")))
		return
	}
	guiJSON(w, map[string]any{"jobs": ids, "errors": errs})
}

// guiRunSync tek sunucu sync job'u: tarama + ana sync çekirdeği + log kaydı.
func guiRunSync(j *job, dir string, srv config.Server, req guiJobReq) (any, error) {
	cfg, err := config.Load(dir)
	if err != nil {
		return nil, err
	}
	if fresh, ok := findServer(cfg, srv.Name); ok {
		srv = fresh
	}
	var buf strings.Builder
	out := io.MultiWriter(j, &buf)
	fmt.Fprintf(out, "══ %s (%s://%s) ══\n", srv.Name, srv.ProtocolName(), srv.Host)
	current, byKey, err := scanForServer(dir, cfg, srv)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(out, lang.L.GuiScannedFmt, srv.EffectiveLocalPath(cfg.Project), len(current))

	var since *time.Time
	if req.Kind == "milestone-sync" {
		ms, _ := milestone.Load(dir, srv.Name)
		if ms == nil {
			return nil, fmt.Errorf(lang.L.GuiNoMilestoneFmt, srv.Name)
		}
		since = &ms.Date
	}
	r := syncToServerOpts(out, dir, cfg, srv, current, byKey, req.Include, req.Exclude, since,
		syncOpts{full: req.Full, dryRun: req.DryRun, retryFailed: req.RetryFailed})
	if !req.DryRun && (r.uploaded > 0 || r.failed > 0) {
		if p, err := synclog.Save(dir, srv.Name, buf.String()); err == nil {
			fmt.Fprintf(j, lang.L.SyncLogSavedFmt, p)
		}
	}
	res := map[string]any{
		"uploaded": r.uploaded, "failed": r.failed, "protected": r.protected, "blocked": r.blocked,
		"frozen": r.frozenSkipped, "bytes": r.bytes, "duration_ms": r.duration.Milliseconds(), "release": r.releaseID,
	}
	return res, r.err
}

func (sv *guiServer) apiJobGet(w http.ResponseWriter, r *http.Request) {
	j := sv.jobs.get(r.PathValue("id"))
	if j == nil {
		guiErr(w, 404, errors.New("job"))
		return
	}
	snap := j.snapshot()
	j.mu.Lock()
	snap["log"] = append([]string(nil), j.lines...)
	j.mu.Unlock()
	guiJSON(w, snap)
}

// apiJobStream SSE: önce mevcut satırlar, sonra canlı satırlar; bitince "done" olayı.
func (sv *guiServer) apiJobStream(w http.ResponseWriter, r *http.Request) {
	j := sv.jobs.get(r.PathValue("id"))
	if j == nil {
		guiErr(w, 404, errors.New("job"))
		return
	}
	fl, ok := w.(http.Flusher)
	if !ok {
		guiErr(w, 500, errors.New("sse"))
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)

	ch := make(chan string, 512)
	j.mu.Lock()
	backlog := append([]string(nil), j.lines...)
	j.subs[ch] = struct{}{}
	j.mu.Unlock()
	defer func() {
		j.mu.Lock()
		delete(j.subs, ch)
		j.mu.Unlock()
	}()

	send := func(event, data string) {
		for _, line := range strings.Split(data, "\n") {
			if event != "" {
				fmt.Fprintf(w, "event: %s\n", event)
			}
			fmt.Fprintf(w, "data: %s\n\n", line)
		}
	}
	for _, l := range backlog {
		send("", l)
	}
	fl.Flush()

	ping := time.NewTicker(15 * time.Second)
	defer ping.Stop()
	for {
		select {
		case l := <-ch:
			send("", l)
			// Biriken satırları tek flush ile gönder
			for drained := false; !drained; {
				select {
				case l2 := <-ch:
					send("", l2)
				default:
					drained = true
				}
			}
			fl.Flush()
		case <-j.done:
			for drained := false; !drained; {
				select {
				case l2 := <-ch:
					send("", l2)
				default:
					drained = true
				}
			}
			snap := j.snapshot()
			send("done", mustJSON(snap))
			fl.Flush()
			return
		case <-ping.C:
			fmt.Fprint(w, ": ping\n\n")
			fl.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func (sv *guiServer) apiJobCancel(w http.ResponseWriter, r *http.Request) {
	j := sv.jobs.get(r.PathValue("id"))
	if j == nil {
		guiErr(w, 404, errors.New("job"))
		return
	}
	if j.cancel != nil {
		j.cancel()
		guiJSON(w, true)
		return
	}
	guiErr(w, 400, errors.New(lang.L.GuiJobNotCancelable))
}

// ── release / log ─────────────────────────────────────────────────────────────

func (sv *guiServer) apiReleases(w http.ResponseWriter, r *http.Request) {
	p, cfg, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	filter := r.URL.Query().Get("server")
	type relInfo struct {
		Server     string    `json:"server"`
		ID         string    `json:"id"`
		CreatedAt  time.Time `json:"created_at"`
		Files      int       `json:"files"`
		Backed     int       `json:"backed"`
		New        int       `json:"new"`
		NoBackup   int       `json:"no_backup"`
		HasBackup  bool      `json:"has_backup"`
		RolledBack bool      `json:"rolled_back"`
	}
	out := []relInfo{}
	for _, srv := range cfg.Servers {
		if filter != "" && srv.Name != filter {
			continue
		}
		list, _ := release.List(p.Path, srv.Name)
		for _, m := range list {
			out = append(out, relInfo{Server: srv.Name, ID: m.ID, CreatedAt: m.CreatedAt, Files: m.FileCount,
				Backed: len(m.Backed), New: len(m.New), NoBackup: len(m.NoBackup), HasBackup: m.HasBackup(), RolledBack: m.RolledBack})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	if len(out) > 300 {
		out = out[:300]
	}
	guiJSON(w, out)
}

func (sv *guiServer) apiReleaseDetail(w http.ResponseWriter, r *http.Request) {
	p, ok := projects.Find(r.PathValue("p"))
	if !ok {
		guiErr(w, 404, errors.New("project"))
		return
	}
	m, err := release.Load(p.Path, filepath.Base(r.PathValue("srv")), filepath.Base(r.PathValue("id")))
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	files := make([]string, 0, len(m.Files))
	for f := range m.Files {
		files = append(files, f)
	}
	sort.Strings(files)
	guiJSON(w, map[string]any{"manifest": m, "files": files, "has_backup": m.HasBackup()})
}

func (sv *guiServer) apiLogs(w http.ResponseWriter, r *http.Request) {
	p, cfg, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	filter := r.URL.Query().Get("server")
	type logInfo struct {
		Server string    `json:"server"`
		File   string    `json:"file"`
		Size   int64     `json:"size"`
		Time   time.Time `json:"time"`
	}
	out := []logInfo{}
	for _, srv := range cfg.Servers {
		if filter != "" && srv.Name != filter {
			continue
		}
		dir := filepath.Join(p.Path, ".syncftp", "logs", srv.Name)
		entries, _ := os.ReadDir(dir)
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			if info, err := e.Info(); err == nil {
				out = append(out, logInfo{Server: srv.Name, File: e.Name(), Size: info.Size(), Time: info.ModTime()})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Time.After(out[j].Time) })
	if len(out) > 300 {
		out = out[:300]
	}
	guiJSON(w, out)
}

func (sv *guiServer) apiLogFile(w http.ResponseWriter, r *http.Request) {
	p, ok := projects.Find(r.PathValue("p"))
	if !ok {
		guiErr(w, 404, errors.New("project"))
		return
	}
	path := filepath.Join(p.Path, ".syncftp", "logs", filepath.Base(r.PathValue("srv")), filepath.Base(r.PathValue("file")))
	data, err := os.ReadFile(path)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	if len(data) > 4<<20 {
		data = data[len(data)-4<<20:]
	}
	guiJSON(w, string(data))
}

// ── milestone / freeze / failed ───────────────────────────────────────────────

func (sv *guiServer) apiMilestones(w http.ResponseWriter, r *http.Request) {
	p, cfg, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	out := map[string]string{}
	for _, s := range cfg.Servers {
		if ms, _ := milestone.Load(p.Path, s.Name); ms != nil {
			out[s.Name] = ms.Date.Format("2006-01-02 15:04:05")
		}
	}
	guiJSON(w, out)
}

func (sv *guiServer) apiMilestoneSet(w http.ResponseWriter, r *http.Request) {
	p, cfg, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	var req struct {
		Servers []string `json:"servers"`
		Date    string   `json:"date"`
	}
	if err := guiDecode(r, &req); err != nil {
		guiErr(w, 400, err)
		return
	}
	date := time.Now()
	if strings.TrimSpace(req.Date) != "" {
		d, err := parseMilestoneDate(req.Date)
		if err != nil {
			guiErr(w, 400, fmt.Errorf(strings.TrimSpace(lang.L.MilestoneDateErrFmt), req.Date))
			return
		}
		date = d
	}
	for _, n := range req.Servers {
		if _, ok := findServer(cfg, n); !ok {
			continue
		}
		if err := milestone.Save(p.Path, n, date); err != nil {
			guiErr(w, 500, err)
			return
		}
	}
	guiJSON(w, date.Format("2006-01-02 15:04:05"))
}

func (sv *guiServer) apiMilestoneClear(w http.ResponseWriter, r *http.Request) {
	p, _, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	_ = milestone.Clear(p.Path, r.URL.Query().Get("server"))
	guiJSON(w, true)
}

func (sv *guiServer) apiFrozenGet(w http.ResponseWriter, r *http.Request) {
	p, _, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	out := []string{}
	if fl, _ := frozen.Load(p.Path, r.URL.Query().Get("server")); fl != nil {
		for f := range fl.Files {
			out = append(out, f)
		}
	}
	sort.Strings(out)
	guiJSON(w, out)
}

func (sv *guiServer) apiFrozenPut(w http.ResponseWriter, r *http.Request) {
	p, _, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	var req struct {
		Server string   `json:"server"`
		Files  []string `json:"files"`
	}
	if err := guiDecode(r, &req); err != nil || req.Server == "" {
		guiErr(w, 400, errors.New("server"))
		return
	}
	var clean []string
	for _, f := range req.Files {
		if f = strings.Trim(filepath.ToSlash(strings.TrimSpace(f)), "/"); f != "" {
			clean = append(clean, f)
		}
	}
	if err := frozen.Save(p.Path, req.Server, clean); err != nil {
		guiErr(w, 500, err)
		return
	}
	guiJSON(w, len(clean))
}

func (sv *guiServer) apiFailed(w http.ResponseWriter, r *http.Request) {
	p, _, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	fl, _ := failed.Load(p.Path, r.URL.Query().Get("server"))
	if fl == nil {
		guiJSON(w, []string{})
		return
	}
	guiJSON(w, fl.Files)
}

// ── watch ─────────────────────────────────────────────────────────────────────

func (sv *guiServer) apiWatchGet(w http.ResponseWriter, r *http.Request) {
	guiJSON(w, sv.watches.get(r.PathValue("p")))
}

func (sv *guiServer) apiWatchStart(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Servers []string `json:"servers"`
	}
	_ = guiDecode(r, &req)
	gw, err := sv.startWatch(r.PathValue("p"), req.Servers)
	if err != nil {
		guiErr(w, 400, err)
		return
	}
	sv.persistWatches()
	guiJSON(w, gw)
}

func (sv *guiServer) apiWatchStop(w http.ResponseWriter, r *http.Request) {
	ok := sv.watches.stop(r.PathValue("p"))
	sv.persistWatches()
	guiJSON(w, ok)
}

// startWatch projede arka plan watch'ı başlatır (panel, tray ve açılışta geri yükleme ortak).
// servers boşsa tüm aktif sunucular izlenir.
func (sv *guiServer) startWatch(project string, serverNames []string) (*guiWatch, error) {
	p, ok := projects.Find(project)
	if !ok {
		return nil, fmt.Errorf(lang.L.ProjectNotFoundFmt, project)
	}
	cfg, err := config.Load(p.Path)
	if err != nil {
		return nil, err
	}
	if sv.watches.get(p.Name) != nil {
		return nil, errors.New(lang.L.GuiWatchRunning)
	}
	var servers []config.Server
	if len(serverNames) == 0 {
		servers = cfg.EnabledServers()
	} else {
		for _, n := range serverNames {
			if s, ok := findServer(cfg, n); ok {
				servers = append(servers, s)
			}
		}
	}
	if len(servers) == 0 {
		return nil, errors.New(lang.L.ShellNoServers)
	}
	names := make([]string, len(servers))
	for i, s := range servers {
		names[i] = s.Name
	}
	gw := &guiWatch{Project: p.Name, Servers: names, Started: time.Now(), stop: make(chan struct{}), explicit: len(serverNames) > 0}
	j, err := sv.jobs.start("watch", p.Name, strings.Join(names, ", "), lang.L.GuiJobWatch, "", func(j *job) (any, error) {
		return nil, watchLoop(p.Path, cfg, servers, j, gw.stop, func(res serverSyncResult) {
			sv.watches.mu.Lock()
			gw.Syncs++
			gw.LastSync = time.Now().Format("15:04:05")
			sv.watches.mu.Unlock()
		})
	})
	if err != nil {
		return nil, err
	}
	j.cancel = func() { sv.watches.stop(p.Name); sv.persistWatches() }
	gw.JobID = j.ID
	sv.watches.mu.Lock()
	sv.watches.m[p.Name] = gw
	sv.watches.mu.Unlock()
	return gw, nil
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

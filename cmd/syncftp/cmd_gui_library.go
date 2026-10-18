package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"syncftp/internal/config"
	"syncftp/internal/lang"
	"syncftp/internal/projects"
)

// ══════════════════════════════════════════════════════════════════════════════
// Proje bağımsız panel uçları: bağlantı kütüphanesi + yeni proje varsayılanları
// (config.Library — CLI `syncftp global` ile aynı veri), proje config dışa/içe aktarma, kütüphaneden projeye bağlantı ve proje başına HTTP API (serve).
// ══════════════════════════════════════════════════════════════════════════════

func (sv *guiServer) registerLibraryAPI(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/library", sv.apiLibraryGet)
	mux.HandleFunc("PUT /api/library", sv.apiLibraryPut)
	mux.HandleFunc("POST /api/library/test", sv.apiLibraryTest)
	mux.HandleFunc("POST /api/library/import/{p}", sv.apiLibraryImport)
	mux.HandleFunc("POST /api/library/defaults-from/{p}", sv.apiLibraryDefaultsFrom)

	mux.HandleFunc("GET /api/p/{p}/config/export", sv.apiConfigExport)
	mux.HandleFunc("POST /api/p/{p}/config/import", sv.apiConfigImport)
	mux.HandleFunc("POST /api/p/{p}/connections/from-library", sv.apiConnFromLibrary)
	mux.HandleFunc("POST /api/p/{p}/connections/to-library", sv.apiConnToLibrary)
	mux.HandleFunc("GET /api/p/{p}/connections/library", sv.apiConnLibraryStatus)
	mux.HandleFunc("POST /api/p/{p}/connections/extract", sv.apiConnExtract)

	mux.HandleFunc("GET /api/p/{p}/serve", sv.apiServeGet)
	mux.HandleFunc("POST /api/p/{p}/serve", sv.apiServeSet)
}

// ── kütüphane ─────────────────────────────────────────────────────────────────

func maskLibrary(lib *config.Library) map[string]any {
	conns := make([]config.Connection, len(lib.Connections))
	copy(conns, lib.Connections)
	for i := range conns {
		maskSecret(&conns[i].Password)
		maskSecret(&conns[i].TunnelPassword)
	}
	tmpls := make([]string, 0, len(ignoreTemplates))
	for k := range ignoreTemplates {
		tmpls = append(tmpls, k)
	}
	sort.Strings(tmpls)
	d := lib.Defaults
	if d.Sync.Include == nil {
		d.Sync.Include = []string{}
	}
	if d.Sync.Exclude == nil {
		d.Sync.Exclude = []string{}
	}
	if d.Sync.Protect == nil {
		d.Sync.Protect = []string{}
	}
	return map[string]any{
		"path": config.LibraryPath(), "connections": conns, "defaults": d,
		"secret_errors": lib.SecretErrors, "templates": tmpls, "builtin": config.BuiltinDefaults(),
	}
}

func (sv *guiServer) apiLibraryGet(w http.ResponseWriter, r *http.Request) {
	lib, err := config.LoadLibrary()
	if err != nil {
		guiErr(w, 500, err)
		return
	}
	guiJSON(w, maskLibrary(lib))
}

// restoreLibSecrets "__keep__" şifreleri mevcut kütüphaneden (ad, yoksa sıra) geri koyar.
func restoreLibSecrets(in []config.Connection, old []config.Connection) {
	restoreKept(in, old, connSecrets)
}

func (sv *guiServer) apiLibraryPut(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Connections *[]config.Connection `json:"connections"`
		Defaults    *config.Defaults     `json:"defaults"`
	}
	if err := guiDecode(r, &in); err != nil {
		guiErr(w, 400, err)
		return
	}
	lib, err := config.LoadLibrary()
	if err != nil {
		guiErr(w, 500, err)
		return
	}
	if in.Connections != nil {
		conns := *in.Connections
		restoreLibSecrets(conns, lib.Connections)
		for i := range conns {
			if conns[i].Port == 0 {
				conns[i].Port = config.DefaultPort(conns[i].Protocol)
			}
		}
		lib.Connections = conns
	}
	if in.Defaults != nil {
		if t := in.Defaults.IgnoreTemplate; t != "" {
			if _, ok := ignoreTemplates[t]; !ok {
				guiErr(w, 400, fmt.Errorf("template: %s", t))
				return
			}
		}
		lib.Defaults = *in.Defaults
	}
	if err := lib.ValidateConnections(); err != nil {
		guiErr(w, 400, err)
		return
	}
	if err := config.SaveLibrary(lib); err != nil {
		guiErr(w, 500, err)
		return
	}
	lib, _ = config.LoadLibrary()
	guiJSON(w, maskLibrary(lib))
}

// apiLibraryTest düzenlenen bağlantıyı (kaydetmeden) dener; remote verilirse onu listeler.
func (sv *guiServer) apiLibraryTest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Connection config.Connection `json:"connection"`
		Remote     string            `json:"remote"`
	}
	if err := guiDecode(r, &req); err != nil {
		guiErr(w, 400, err)
		return
	}
	lib, _ := config.LoadLibrary()
	one := []config.Connection{req.Connection}
	restoreLibSecrets(one, lib.Connections)
	srv := config.Server{RemotePath: orDefault(req.Remote, "/")}
	srv.ApplyConnection(one[0])
	srv.Normalize()
	testConnection(w, srv)
}

func (sv *guiServer) apiLibraryImport(w http.ResponseWriter, r *http.Request) {
	_, cfg, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	lib, err := config.LoadLibrary()
	if err != nil {
		guiErr(w, 500, err)
		return
	}
	n := importConnections(lib, cfg)
	if n > 0 {
		if err := config.SaveLibrary(lib); err != nil {
			guiErr(w, 500, err)
			return
		}
	}
	guiJSON(w, map[string]any{"added": n})
}

func (sv *guiServer) apiLibraryDefaultsFrom(w http.ResponseWriter, r *http.Request) {
	_, cfg, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	lib, err := config.LoadLibrary()
	if err != nil {
		guiErr(w, 500, err)
		return
	}
	defaultsFromProject(&lib.Defaults, cfg)
	if err := config.SaveLibrary(lib); err != nil {
		guiErr(w, 500, err)
		return
	}
	guiJSON(w, maskLibrary(lib))
}

// ── proje config dışa / içe aktarma ───────────────────────────────────────────

// sanitizedConfig şifreleri boşaltılmış kopya (paylaşılabilir dışa aktarma).
func (sv *guiServer) apiConfigExport(w http.ResponseWriter, r *http.Request) {
	p, cfg, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	out := cfg.Redacted(config.MaskEmpty, true)
	data, _ := json.MarshalIndent(out, "", "  ")
	name := strings.NewReplacer(" ", "-", "/", "-", "\\", "-").Replace(p.Name) + ".syncftp.json"
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+name+`"`)
	_, _ = w.Write(data)
}

// apiConfigImport başka bir projeden ya da yapıştırılan JSON'dan seçili bölümleri alır.
// parts: sync | servers | connections | deploy. mode: merge (aynı adlıyı değiştir, yeniyi ekle) | replace.
func (sv *guiServer) apiConfigImport(w http.ResponseWriter, r *http.Request) {
	p, cfg, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	var req struct {
		FromProject string          `json:"from_project"`
		JSON        json.RawMessage `json:"json"`
		Parts       []string        `json:"parts"`
		Mode        string          `json:"mode"`
	}
	if err := guiDecode(r, &req); err != nil {
		guiErr(w, 400, err)
		return
	}
	var src *config.Config
	switch {
	case req.FromProject != "":
		pp, ok := projects.Find(req.FromProject)
		if !ok {
			guiErr(w, 404, fmt.Errorf(lang.L.ProjectNotFoundFmt, req.FromProject))
			return
		}
		if src, err = config.Load(pp.Path); err != nil {
			guiErr(w, 400, err)
			return
		}
	case len(req.JSON) > 0:
		src = &config.Config{}
		raw := req.JSON
		var s string
		if json.Unmarshal(raw, &s) == nil { // metin olarak yapıştırılmış
			raw = json.RawMessage(s)
		}
		if err := json.Unmarshal(raw, src); err != nil {
			guiErr(w, 400, fmt.Errorf("JSON: %w", err))
			return
		}
	default:
		guiErr(w, 400, errors.New("from_project | json"))
		return
	}
	replace := req.Mode == "replace"
	// Dışa aktarma şifreleri boşaltır (MaskEmpty) → aynı adlı kayıtta boş gelen sır mevcut olanı
	// silmesin (anahtarlıktaki kayıt da Save'de boşla ezilirdi).
	keepEmptySecrets(src.Servers, cfg.Servers, func(s *config.Server) (string, []*string) {
		return s.Name, []*string{&s.Password, &s.TunnelPassword}
	})
	keepEmptySecrets(src.Connections, cfg.Connections, connSecrets)
	keepEmptySecrets(src.Deploy, cfg.Deploy, func(d *config.Deploy) (string, []*string) {
		return d.Name, []*string{&d.Secret}
	})
	counts := map[string]int{}
	for _, part := range req.Parts {
		switch part {
		case "sync":
			cfg.Sync = src.Sync.Clone()
			cfg.FirstSync = src.FirstSync
			counts["sync"] = 1
		case "connections":
			cfg.Connections = mergeByName(cfg.Connections, src.Connections, replace, func(c config.Connection) string { return c.Name })
			counts["connections"] = len(src.Connections)
		case "servers":
			cfg.Servers = mergeByName(cfg.Servers, src.Servers, replace, func(s config.Server) string { return s.Name })
			counts["servers"] = len(src.Servers)
		case "deploy":
			cfg.Deploy = mergeByName(cfg.Deploy, src.Deploy, replace, func(d config.Deploy) string { return d.Name })
			counts["deploy"] = len(src.Deploy)
		}
	}
	if err := cfg.Validate(); err != nil {
		guiErr(w, 400, err)
		return
	}
	if err := config.Save(p.Path, cfg); err != nil {
		guiErr(w, 500, err)
		return
	}
	guiJSON(w, counts)
}

// keepEmptySecrets içe aktarılan kayıtlarda boş sırları aynı adlı (mergeByName gibi tam ad) mevcut
// kayıttan doldurur. restoreKept'ten farkı: işaret "__keep__" değil boş değer, sıra eşleşmesi yok.
func keepEmptySecrets[T any](in, old []T, fields func(*T) (string, []*string)) {
	byName := map[string]int{}
	for i := range old {
		n, _ := fields(&old[i])
		byName[n] = i
	}
	for i := range in {
		n, dst := fields(&in[i])
		oi, ok := byName[n]
		if !ok {
			continue
		}
		_, src := fields(&old[oi])
		for k, d := range dst {
			if *d == "" {
				*d = *src[k]
			}
		}
	}
}

func mergeByName[T any](dst, src []T, replace bool, name func(T) string) []T {
	if replace {
		return append([]T(nil), src...)
	}
	out := append([]T(nil), dst...)
	for _, s := range src {
		found := false
		for i := range out {
			if name(out[i]) == name(s) {
				out[i] = s
				found = true
				break
			}
		}
		if !found {
			out = append(out, s)
		}
	}
	return out
}

// apiConnFromLibrary kütüphanedeki bağlantıyı projenin bağlantı profillerine kopyalar
// (şifre gerçek değeriyle; proje keyring modundaysa Save anahtarlığa yazar). Dönüş: profil adı.
func (sv *guiServer) apiConnFromLibrary(w http.ResponseWriter, r *http.Request) {
	p, cfg, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	var req struct {
		Name string `json:"name"`
	}
	if err := guiDecode(r, &req); err != nil {
		guiErr(w, 400, err)
		return
	}
	lib, _ := config.LoadLibrary()
	c, ok := lib.FindConnection(req.Name)
	if !ok {
		guiErr(w, 404, fmt.Errorf(lang.L.GlobalConnNotFoundFmt, req.Name))
		return
	}
	for _, e := range cfg.Connections {
		if e.Name == c.Name {
			if e.Host == c.Host && e.User == c.User && e.Port == c.Port {
				guiJSON(w, map[string]any{"name": e.Name, "existing": true})
				return
			}
			c.Name = c.Name + "-lib"
		}
	}
	cfg.Connections = append(cfg.Connections, c)
	if err := config.Save(p.Path, cfg); err != nil {
		guiErr(w, 500, err)
		return
	}
	guiJSON(w, map[string]any{"name": c.Name})
}

// apiConnToLibrary projedeki bağlantı profilini (ya da profilsiz sunucunun bilgilerini) kütüphaneye kaydeder.
func (sv *guiServer) apiConnToLibrary(w http.ResponseWriter, r *http.Request) {
	_, cfg, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	var req struct {
		Connection string `json:"connection"`
		Server     string `json:"server"`
		SaveAs     string `json:"save_as"`
	}
	if err := guiDecode(r, &req); err != nil {
		guiErr(w, 400, err)
		return
	}
	var c config.Connection
	found := false
	if req.Connection != "" {
		for _, x := range cfg.Connections {
			if x.Name == req.Connection {
				c, found = x, true
			}
		}
	} else if s, ok := findServer(cfg, req.Server); ok {
		c = config.ConnectionFromServer(s.Name, s)
		found = true
	}
	if !found {
		guiErr(w, 404, errors.New("not found"))
		return
	}
	if strings.TrimSpace(req.SaveAs) != "" {
		c.Name = strings.TrimSpace(req.SaveAs)
	}
	lib, _ := config.LoadLibrary()
	replaced := upsertLibConn(lib, c)
	if err := lib.ValidateConnections(); err != nil {
		guiErr(w, 400, err)
		return
	}
	if err := config.SaveLibrary(lib); err != nil {
		guiErr(w, 500, err)
		return
	}
	guiJSON(w, map[string]any{"name": c.Name, "replaced": replaced})
}

// apiConnLibraryStatus projedeki her bağlantı profilinin kütüphanedeki karşılığı: aynı hesap
// (AccountKey) kayıtlıysa o kaydın adı, yoksa "". name_taken: aynı adla FARKLI hesap var (kaydetmek ezer).
func (sv *guiServer) apiConnLibraryStatus(w http.ResponseWriter, r *http.Request) {
	_, cfg, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	lib, _ := config.LoadLibrary()
	type st struct {
		Library   string `json:"library"`
		NameTaken bool   `json:"name_taken"`
	}
	out := map[string]st{}
	for _, c := range cfg.Connections {
		var s st
		for _, l := range lib.Connections {
			if l.AccountKey() == c.AccountKey() {
				s.Library = l.Name
				break
			}
		}
		if s.Library == "" {
			if l, ok := lib.FindConnection(c.Name); ok && l.AccountKey() != c.AccountKey() {
				s.NameTaken = true
			}
		}
		out[c.Name] = s
	}
	guiJSON(w, out)
}

// apiConnExtract profilsiz sunucunun bilgilerini bağlantı profiline taşır (CLI `config extract-conn` ile aynı çekirdek).
func (sv *guiServer) apiConnExtract(w http.ResponseWriter, r *http.Request) {
	p, cfg, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	var req struct {
		Server string `json:"server"`
		Name   string `json:"name"`
	}
	if err := guiDecode(r, &req); err != nil {
		guiErr(w, 400, err)
		return
	}
	name, created, err := cfg.ExtractConnection(req.Server, req.Name)
	if err != nil {
		guiErr(w, 400, err)
		return
	}
	if err := cfg.Validate(); err != nil {
		guiErr(w, 400, err)
		return
	}
	if err := config.Save(p.Path, cfg); err != nil {
		guiErr(w, 500, err)
		return
	}
	guiJSON(w, map[string]any{"name": name, "created": created})
}

// ── proje başına HTTP API (syncftp serve) ─────────────────────────────────────

type guiServe struct {
	srv  *http.Server
	Port int    `json:"port"`
	URL  string `json:"url"`
}

var (
	guiServesMu sync.Mutex
	guiServes   = map[string]*guiServe{} // proje adı → çalışan API
)

func (sv *guiServer) apiServeGet(w http.ResponseWriter, r *http.Request) {
	p, _, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	guiServesMu.Lock()
	s := guiServes[p.Name]
	guiServesMu.Unlock()
	if s == nil {
		guiJSON(w, map[string]any{"active": false})
		return
	}
	guiJSON(w, map[string]any{"active": true, "port": s.Port, "url": s.URL})
}

func (sv *guiServer) apiServeSet(w http.ResponseWriter, r *http.Request) {
	p, _, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	var req struct {
		On   bool `json:"on"`
		Port int  `json:"port"`
	}
	if err := guiDecode(r, &req); err != nil {
		guiErr(w, 400, err)
		return
	}
	guiServesMu.Lock()
	defer guiServesMu.Unlock()
	if cur := guiServes[p.Name]; cur != nil {
		_ = cur.srv.Close()
		delete(guiServes, p.Name)
	}
	if !req.On {
		guiJSON(w, map[string]any{"active": false})
		return
	}
	port := req.Port
	if port == 0 {
		port = 8080
	}
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		guiErr(w, 400, err)
		return
	}
	hs := &http.Server{Handler: newServeHandler(p.Path), ReadHeaderTimeout: 10 * time.Second}
	go func() { _ = hs.Serve(ln) }()
	s := &guiServe{srv: hs, Port: port, URL: fmt.Sprintf("http://127.0.0.1:%d", port)}
	guiServes[p.Name] = s
	guiJSON(w, map[string]any{"active": true, "port": s.Port, "url": s.URL})
}

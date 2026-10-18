package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	goftp "github.com/jlaffaye/ftp"

	"syncftp/internal/config"
	ftpclient "syncftp/internal/ftp"
	"syncftp/internal/ignore"
	"syncftp/internal/lang"
	"syncftp/internal/projects"
)

// GUI dosya tarayıcısı için bağlantı önbelleği: her klasör tıklamasında yeniden
// bağlanmak yerine proje+sunucu başına bir bağlantı açık tutulur; 90 sn boşta
// kalırsa kapatılır. Hata alınırsa bağlantı atılıp bir kez yeniden denenir.

type cachedClient struct {
	client *ftpclient.Client
	used   time.Time
	mu     sync.Mutex // aynı bağlantıda tek istek
}

var (
	remoteCacheMu sync.Mutex
	remoteCache   = map[string]*cachedClient{}
	remoteJanitor sync.Once
)

func remoteKey(project string, srv config.Server) string {
	return project + "|" + srv.Name + "|" + ftpclient.HostKey(srv) + "|" + srv.RemotePath
}

func withRemote(project string, srv config.Server, fn func(c *ftpclient.Client) error) error {
	remoteJanitor.Do(func() {
		go func() {
			for range time.Tick(30 * time.Second) {
				// Kapatma (QUIT) ağ işlemi — 90 sn'ye kadar sürebilir; önbellek kilidi tutulurken yapılmaz
				var stale []*ftpclient.Client
				remoteCacheMu.Lock()
				for k, cc := range remoteCache {
					if cc.mu.TryLock() {
						if time.Since(cc.used) > 90*time.Second {
							stale = append(stale, cc.client)
							delete(remoteCache, k)
						}
						cc.mu.Unlock()
					}
				}
				remoteCacheMu.Unlock()
				for _, c := range stale {
					c.Close()
				}
			}
		}()
	})

	key := remoteKey(project, srv)
	for attempt := 0; attempt < 2; attempt++ {
		remoteCacheMu.Lock()
		cc, ok := remoteCache[key]
		if !ok {
			remoteCacheMu.Unlock()
			c, err := ftpclient.Connect(srv)
			if err != nil {
				return err
			}
			cc = &cachedClient{client: c}
			remoteCacheMu.Lock()
			if existing, raced := remoteCache[key]; raced {
				c.Close()
				cc = existing
			} else {
				remoteCache[key] = cc
			}
		}
		remoteCacheMu.Unlock()

		cc.mu.Lock()
		err := fn(cc.client)
		cc.used = time.Now()
		cc.mu.Unlock()
		if err == nil || ftpclient.IsNotFound(err) {
			return err
		}
		// Bağlantı bozulmuş olabilir → at. Yalnızca taşıma hatasında (ölü önbellek bağlantısı) bir kez
		// daha denenir: sunucunun reddettiği (izin, kota…) ya da yarıda kalmış idempotent olmayan
		// işlem (yükleme, taşıma, klasör oluşturma) tekrarlanmaz.
		remoteCacheMu.Lock()
		if remoteCache[key] == cc {
			delete(remoteCache, key)
		}
		remoteCacheMu.Unlock()
		cc.client.Close()
		if attempt == 1 || !isTransportErr(err) {
			return err
		}
	}
	return nil
}

// isTransportErr bağlantının kendisinin koptuğunu gösteren hatalar (boşta kapanmış önbellek bağlantısı).
func isTransportErr(err error) bool {
	var ne net.Error
	if errors.As(err, &ne) || errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, net.ErrClosed) {
		return true
	}
	s := strings.ToLower(err.Error())
	for _, k := range []string{"use of closed network connection", "connection reset", "broken pipe",
		"connection lost", "connection aborted", "eof", "421 "} {
		if strings.Contains(s, k) {
			return true
		}
	}
	return false
}

// guiRemotePath GUI'nin RemotePath'e göreceli yolunu mutlak uzak yola çevirir;
// ".." ile kökün dışına çıkılamaz.
func guiRemotePath(srv config.Server, rel string) string {
	clean := path.Clean("/" + strings.TrimPrefix(rel, "/"))
	return path.Join(srv.RemotePath, clean)
}

func (sv *guiServer) remoteServer(w http.ResponseWriter, r *http.Request) (string, config.Server, bool) {
	p, _, srv, ok := sv.remoteServerCfg(w, r)
	return p.Name, srv, ok
}

// remoteServerCfg remoteServer + proje ve config (config'i ikinci kez yüklememek için).
func (sv *guiServer) remoteServerCfg(w http.ResponseWriter, r *http.Request) (projects.Project, *config.Config, config.Server, bool) {
	p, cfg, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return p, nil, config.Server{}, false
	}
	name := r.URL.Query().Get("server")
	srv, ok := findServer(cfg, name)
	if !ok {
		guiErr(w, 404, fmt.Errorf(lang.L.ServerNotFoundFmt, name))
		return p, nil, config.Server{}, false
	}
	return p, cfg, srv, true
}

type guiEntry struct {
	Name string    `json:"name"`
	Type string    `json:"type"`
	Size uint64    `json:"size"`
	Time time.Time `json:"time"`
	Skip string    `json:"skip,omitempty"` // syncSkipReason
}

func (sv *guiServer) apiRemoteLs(w http.ResponseWriter, r *http.Request) {
	p, cfg, srv, ok := sv.remoteServerCfg(w, r)
	if !ok {
		return
	}
	full := guiRemotePath(srv, r.URL.Query().Get("path"))
	var entries []*goftp.Entry
	err := withRemote(p.Name, srv, func(c *ftpclient.Client) error {
		var e error
		entries, e = remoteEntries(c, full)
		return e
	})
	if err != nil {
		guiErr(w, 502, err)
		return
	}
	// Sync'e hiç girmeyenler (ignore / exclude / include dışı) satırda rozetle gösterilir
	m, _ := ignore.Load(filepath.Join(p.Path, srv.EffectiveLocalPath(cfg.Project)), cfg.Sync.IgnoreFiles)
	f := effectiveFilters(cfg, srv, nil, nil)
	skip := func(rel string, dir bool) string { return syncSkipReason(rel, dir, m, f) }
	base := editRel(r.URL.Query().Get("path"))
	out := make([]guiEntry, 0, len(entries))
	for _, e := range entries {
		kind := entryKind(e)
		out = append(out, guiEntry{Name: e.Name, Type: kind, Size: e.Size, Time: e.Time, Skip: skip(path.Join(base, e.Name), kind == "dir")})
	}
	guiJSON(w, map[string]any{"path": full, "root": srv.RemotePath, "entries": out})
}

func (sv *guiServer) apiRemoteCat(w http.ResponseWriter, r *http.Request) {
	pname, srv, ok := sv.remoteServer(w, r)
	if !ok {
		return
	}
	full := guiRemotePath(srv, r.URL.Query().Get("path"))
	var pv textPreview
	err := withRemote(pname, srv, func(c *ftpclient.Client) error {
		var e error
		pv, e = remotePreview(c, full, 256*1024)
		return e
	})
	if err != nil {
		guiErr(w, 502, err)
		return
	}
	guiJSON(w, map[string]any{"path": full, "content": pv.Content, "binary": pv.Binary, "truncated": pv.Truncated, "size": pv.Size})
}

func (sv *guiServer) apiRemoteGet(w http.ResponseWriter, r *http.Request) {
	pname, srv, ok := sv.remoteServer(w, r)
	if !ok {
		return
	}
	full := guiRemotePath(srv, r.URL.Query().Get("path"))
	if err := withRemote(pname, srv, func(c *ftpclient.Client) error { return serveRemoteDownload(w, r, c, full) }); err != nil {
		guiErr(w, 502, err)
	}
}

func (sv *guiServer) apiRemoteRm(w http.ResponseWriter, r *http.Request) {
	pname, srv, ok := sv.remoteServer(w, r)
	if !ok {
		return
	}
	rel := r.URL.Query().Get("path")
	full := guiRemotePath(srv, rel)
	if path.Clean(full) == path.Clean(srv.RemotePath) {
		guiErr(w, 400, errors.New("root"))
		return
	}
	// CLI `remote rm` ile aynı: recursive=false iken dolu klasör silinmez (varsayılan true — panel onay sorar)
	recursive := r.URL.Query().Get("recursive") != "false"
	err := withRemote(pname, srv, func(c *ftpclient.Client) error {
		_, e := remoteDelete(c, full, recursive)
		return e
	})
	if err != nil {
		guiErr(w, 502, err)
		return
	}
	guiJSON(w, true)
}

// apiLocalCat sunucunun yerel dizinindeki dosyayı okur (diff görünümü için).
// Yol yerel dizinin dışına çıkamaz.
func (sv *guiServer) apiLocalCat(w http.ResponseWriter, r *http.Request) {
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
	rel := path.Clean("/" + strings.TrimPrefix(r.URL.Query().Get("path"), "/"))
	full := filepath.Join(root, filepath.FromSlash(rel))
	if relCheck, err := filepath.Rel(root, full); err != nil || strings.HasPrefix(relCheck, "..") {
		guiErr(w, 400, errors.New("path"))
		return
	}
	pv, ok := localPreview(full, 256*1024)
	if !ok {
		guiJSON(w, map[string]any{"exists": false})
		return
	}
	guiJSON(w, map[string]any{"exists": true, "content": pv.Content, "binary": pv.Binary, "truncated": pv.Truncated, "size": pv.Size})
}

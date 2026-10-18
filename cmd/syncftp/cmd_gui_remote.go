package main

import (
	"errors"
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
				remoteCacheMu.Lock()
				for k, cc := range remoteCache {
					if cc.mu.TryLock() {
						if time.Since(cc.used) > 90*time.Second {
							cc.client.Close()
							delete(remoteCache, k)
						}
						cc.mu.Unlock()
					}
				}
				remoteCacheMu.Unlock()
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
		// Bağlantı ölmüş olabilir → at, bir kez daha dene
		remoteCacheMu.Lock()
		if remoteCache[key] == cc {
			delete(remoteCache, key)
		}
		remoteCacheMu.Unlock()
		cc.client.Close()
		if attempt == 1 {
			return err
		}
	}
	return nil
}

// guiRemotePath GUI'nin RemotePath'e göreceli yolunu mutlak uzak yola çevirir;
// ".." ile kökün dışına çıkılamaz.
func guiRemotePath(srv config.Server, rel string) string {
	clean := path.Clean("/" + strings.TrimPrefix(rel, "/"))
	return path.Join(srv.RemotePath, clean)
}

func (sv *guiServer) remoteServer(w http.ResponseWriter, r *http.Request) (string, config.Server, bool) {
	p, cfg, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return "", config.Server{}, false
	}
	srv, ok := findServer(cfg, r.URL.Query().Get("server"))
	if !ok {
		guiErr(w, 404, errors.New("server"))
		return "", config.Server{}, false
	}
	return p.Name, srv, true
}

type guiEntry struct {
	Name string    `json:"name"`
	Type string    `json:"type"`
	Size uint64    `json:"size"`
	Time time.Time `json:"time"`
	Skip string    `json:"skip,omitempty"` // syncSkipReason
}

func (sv *guiServer) apiRemoteLs(w http.ResponseWriter, r *http.Request) {
	pname, srv, ok := sv.remoteServer(w, r)
	if !ok {
		return
	}
	full := guiRemotePath(srv, r.URL.Query().Get("path"))
	var entries []*goftp.Entry
	err := withRemote(pname, srv, func(c *ftpclient.Client) error {
		var e error
		entries, e = remoteEntries(c, full)
		return e
	})
	if err != nil {
		guiErr(w, 502, err)
		return
	}
	// Sync'e hiç girmeyenler (ignore / exclude / include dışı) satırda rozetle gösterilir
	skip := func(string, bool) string { return "" }
	if p, cfg, err := sv.project(r); err == nil {
		m, _ := ignore.Load(filepath.Join(p.Path, srv.EffectiveLocalPath(cfg.Project)), cfg.Sync.IgnoreFiles)
		f := effectiveFilters(cfg, srv, nil, nil)
		skip = func(rel string, dir bool) string { return syncSkipReason(rel, dir, m, f) }
	}
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
		guiErr(w, 404, errors.New("server"))
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

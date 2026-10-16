package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	goftp "github.com/jlaffaye/ftp"

	"syncftp/internal/config"
	ftpclient "syncftp/internal/ftp"
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
		entries, e = c.List(full)
		return e
	})
	if err != nil {
		guiErr(w, 502, err)
		return
	}
	out := []guiEntry{}
	for _, e := range entries {
		if e.Name == "." || e.Name == ".." {
			continue
		}
		t := "file"
		switch e.Type {
		case goftp.EntryTypeFolder:
			t = "dir"
		case goftp.EntryTypeLink:
			t = "link"
		}
		out = append(out, guiEntry{Name: e.Name, Type: t, Size: e.Size, Time: e.Time})
	}
	sort.Slice(out, func(i, j int) bool {
		if (out[i].Type == "dir") != (out[j].Type == "dir") {
			return out[i].Type == "dir"
		}
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	guiJSON(w, map[string]any{"path": full, "root": srv.RemotePath, "entries": out})
}

func (sv *guiServer) apiRemoteCat(w http.ResponseWriter, r *http.Request) {
	pname, srv, ok := sv.remoteServer(w, r)
	if !ok {
		return
	}
	full := guiRemotePath(srv, r.URL.Query().Get("path"))
	const max = 256 * 1024
	var data []byte
	err := withRemote(pname, srv, func(c *ftpclient.Client) error {
		var e error
		data, e = c.Preview(full, max)
		return e
	})
	if err != nil {
		guiErr(w, 502, err)
		return
	}
	binary := false
	for _, b := range data[:min(len(data), 8000)] {
		if b == 0 {
			binary = true
			break
		}
	}
	content := ""
	if !binary {
		content = strings.ReplaceAll(string(data), "\r\n", "\n")
	}
	guiJSON(w, map[string]any{"path": full, "content": content, "binary": binary, "truncated": len(data) >= max, "size": len(data)})
}

func (sv *guiServer) apiRemoteGet(w http.ResponseWriter, r *http.Request) {
	pname, srv, ok := sv.remoteServer(w, r)
	if !ok {
		return
	}
	full := guiRemotePath(srv, r.URL.Query().Get("path"))
	tmp, err := os.CreateTemp("", "syncftp-gui-*")
	if err != nil {
		guiErr(w, 500, err)
		return
	}
	tmp.Close()
	defer os.Remove(tmp.Name())
	if err := withRemote(pname, srv, func(c *ftpclient.Client) error { return c.Download(full, tmp.Name()) }); err != nil {
		guiErr(w, 502, err)
		return
	}
	f, err := os.Open(tmp.Name())
	if err != nil {
		guiErr(w, 500, err)
		return
	}
	defer f.Close()
	st, _ := f.Stat()
	name := path.Base(full)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, strings.ReplaceAll(name, `"`, "")))
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeContent(w, r, name, st.ModTime(), f)
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
	isDir := r.URL.Query().Get("dir") == "true"
	err := withRemote(pname, srv, func(c *ftpclient.Client) error {
		if isDir {
			return c.DeleteDir(full, true)
		}
		return c.DeleteFile(full)
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
	f, err := os.Open(full)
	if err != nil {
		guiJSON(w, map[string]any{"exists": false})
		return
	}
	defer f.Close()
	const max = 256 * 1024
	data, _ := io.ReadAll(io.LimitReader(f, max))
	binary := false
	for _, b := range data[:min(len(data), 8000)] {
		if b == 0 {
			binary = true
			break
		}
	}
	content := ""
	if !binary {
		content = strings.ReplaceAll(string(data), "\r\n", "\n")
	}
	guiJSON(w, map[string]any{"exists": true, "content": content, "binary": binary, "truncated": len(data) >= max, "size": len(data)})
}

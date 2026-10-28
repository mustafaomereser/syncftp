package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"syncftp/internal/fsutil"
	ftpclient "syncftp/internal/ftp"
	"syncftp/internal/lang"
)

// Panel seçicilerinde (sunucuda klasör seç, yerel klasör/dosya seç, proje içinden seç, taşı)
// "Yeni klasör" / "Yeni dosya". Ad tek bir öğe adı olmalı (yol değil); dosya boş oluşturulur,
// var olan öğenin üzerine yazılmaz.

// newItemReq seçicinin oluşturma isteği; Kind "dir" | "file".
type newItemReq struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

// checkNewName adın tek parça ve her iki tarafta (Windows dahil) geçerli olduğunu doğrular.
func checkNewName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == ".." || name != fsutil.SafeName(name) {
		return "", fmt.Errorf(lang.L.GuiBadNameFmt, name)
	}
	return name, nil
}

// createLocalItem dir altında klasör ya da boş dosya oluşturur.
func createLocalItem(dir string, q newItemReq) (string, error) {
	name, err := checkNewName(q.Name)
	if err != nil {
		return "", err
	}
	full := filepath.Join(dir, name)
	if _, err := os.Lstat(full); err == nil {
		return "", fmt.Errorf(lang.L.GuiItemExistsFmt, name)
	}
	if q.Kind == "dir" {
		return full, os.Mkdir(full, 0o755)
	}
	f, err := os.OpenFile(full, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return "", err
	}
	return full, f.Close()
}

// remoteCreateItem sunucuda dir altında klasör ya da boş dosya oluşturur (CLI/panel ortak çekirdek).
func remoteCreateItem(c *ftpclient.Client, dir string, q newItemReq) (string, error) {
	name, err := checkNewName(q.Name)
	if err != nil {
		return "", err
	}
	full := path.Join(dir, name)
	if entries, err := remoteEntries(c, dir); err == nil {
		for _, e := range entries {
			if e.Name == name {
				return "", fmt.Errorf(lang.L.GuiItemExistsFmt, name)
			}
		}
	}
	if q.Kind == "dir" {
		return full, c.MakeDirAll(full)
	}
	tmp, err := os.CreateTemp("", "syncftp-empty-*")
	if err != nil {
		return "", err
	}
	tmp.Close()
	defer os.Remove(tmp.Name())
	return full, c.UploadTo(tmp.Name(), full)
}

// underRoot rel'i root'a bağlar; dışarı çıkılamaz.
func underRoot(root, rel string) (string, error) {
	rel = strings.Trim(path.Clean("/"+filepath.ToSlash(rel)), "/")
	full := filepath.Join(root, filepath.FromSlash(rel))
	if rc, err := filepath.Rel(root, full); err != nil || strings.HasPrefix(rc, "..") {
		return "", errors.New("path")
	}
	return full, nil
}

// POST /api/fs/create {root, path, kind, name} — pickPath (root = açık klasör) ve proje dışı
// pickLocalTree (root = gezilen kök, path = ona göreceli).
func (sv *guiServer) apiFSCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		newItemReq
		Root string `json:"root"`
		Path string `json:"path"`
	}
	if err := guiDecode(r, &req); err != nil {
		guiErr(w, 400, err)
		return
	}
	root := filepath.Clean(strings.TrimSpace(req.Root))
	if st, err := os.Stat(root); err != nil || !st.IsDir() || !filepath.IsAbs(root) {
		guiErr(w, 400, fmt.Errorf(lang.L.GuiDirNotFoundFmt, req.Root))
		return
	}
	dir, err := underRoot(root, req.Path)
	if err != nil {
		guiErr(w, 400, err)
		return
	}
	full, err := createLocalItem(dir, req.newItemReq)
	if err != nil {
		guiErr(w, 400, err)
		return
	}
	guiJSON(w, map[string]any{"path": full, "name": filepath.Base(full)})
}

// POST /api/p/{p}/local/create?server=&path= {kind, name} — proje içi pickLocalTree (sunucunun yerel dizini).
func (sv *guiServer) apiLocalCreate(w http.ResponseWriter, r *http.Request) {
	p, cfg, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	localPath := cfg.Project.EffectiveLocalPath()
	if srv, ok := findServer(cfg, r.URL.Query().Get("server")); ok {
		localPath = srv.EffectiveLocalPath(cfg.Project)
	}
	var req newItemReq
	if err := guiDecode(r, &req); err != nil {
		guiErr(w, 400, err)
		return
	}
	dir, err := underRoot(filepath.Join(p.Path, localPath), r.URL.Query().Get("path"))
	if err != nil {
		guiErr(w, 400, err)
		return
	}
	full, err := createLocalItem(dir, req)
	if err != nil {
		guiErr(w, 400, err)
		return
	}
	guiJSON(w, map[string]any{"name": filepath.Base(full)})
}

// POST /api/p/{p}/remote/create?server=&path= {kind, name} — dosya tarayıcısı ve "taşı" seçicisi
// (yollar RemotePath'e göreceli; önbellekli tarayıcı bağlantısı).
func (sv *guiServer) apiRemoteCreate(w http.ResponseWriter, r *http.Request) {
	pname, srv, ok := sv.remoteServer(w, r)
	if !ok {
		return
	}
	var req newItemReq
	if err := guiDecode(r, &req); err != nil {
		guiErr(w, 400, err)
		return
	}
	dir := guiRemotePath(srv, r.URL.Query().Get("path"))
	var full string
	err := withRemote(pname, srv, func(c *ftpclient.Client) error {
		var e error
		full, e = remoteCreateItem(c, dir, req)
		return e
	})
	if err != nil {
		guiErr(w, 400, err)
		return
	}
	guiJSON(w, map[string]any{"name": path.Base(full)})
}

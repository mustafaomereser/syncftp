package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"

	"syncftp/internal/lang"
	"syncftp/internal/synclog"

	ftpclient "syncftp/internal/ftp"
)

// GET /api/p/{p}/remote/edit?server=&path= — düzenleyici için dosyanın tamamı + yerel karşılaştırması
func (sv *guiServer) apiRemoteEditOpen(w http.ResponseWriter, r *http.Request) {
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
	rel := editRel(r.URL.Query().Get("path"))
	localFile, err := editLocalPath(p.Path, cfg, srv, rel)
	if err != nil {
		guiErr(w, 400, err)
		return
	}
	var data []byte
	if err := withRemote(p.Name, srv, func(c *ftpclient.Client) error {
		var e error
		data, e = readRemoteFull(c, guiRemotePath(srv, rel))
		return e
	}); err != nil {
		guiErr(w, 502, err)
		return
	}
	local, lerr := os.ReadFile(localFile)
	guiJSON(w, openEditDoc(data, local, lerr == nil))
}

// PUT /api/p/{p}/remote/edit — kaydet. Sunucudaki dosya açıldıktan sonra değiştiyse (base_hash
// tutmuyor) ve force yoksa 409 + güncel içerik döner. Sync ile aynı sunucu kilidini kullanan bir iş
// olarak çalışır (İşler'de görünür, aynı anda o sunucuya sync yazamaz).
func (sv *guiServer) apiRemoteEditSave(w http.ResponseWriter, r *http.Request) {
	p, cfg, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	var req struct {
		Server   string `json:"server"`
		Path     string `json:"path"`
		Content  string `json:"content"`
		BaseHash string `json:"base_hash"`
		Local    bool   `json:"local"`
		Force    bool   `json:"force"`
	}
	defer r.Body.Close()
	if err := json.NewDecoder(io.LimitReader(r.Body, 3*editMaxBytes)).Decode(&req); err != nil {
		guiErr(w, 400, err)
		return
	}
	srv, ok := findServer(cfg, req.Server)
	if !ok {
		guiErr(w, 404, errors.New("server"))
		return
	}
	rel := editRel(req.Path)
	if rel == "" {
		guiErr(w, 400, errors.New("path"))
		return
	}
	var conflict *editDoc
	var releaseID string
	j, err := sv.jobs.start("edit", p.Name, srv.Name, lang.L.GuiJobEdit+": "+rel, p.Name+"|"+srv.Name, func(j *job) (any, error) {
		if !req.Force && req.BaseHash != "" {
			var cur []byte
			if err := withRemote(p.Name, srv, func(c *ftpclient.Client) error {
				var e error
				cur, e = readRemoteFull(c, guiRemotePath(srv, rel))
				if e != nil && ftpclient.IsNotFound(e) {
					cur, e = nil, nil
				}
				return e
			}); err != nil {
				return nil, err
			}
			if contentHash(cur) != req.BaseHash {
				d := openEditDoc(cur, nil, false)
				conflict = &d
				return nil, errors.New(lang.L.EditConflict)
			}
		}
		var buf bytes.Buffer
		out := io.MultiWriter(j, &buf)
		id, err := saveRemoteEdit(out, p.Path, cfg, srv, rel, req.Content, editSaveOpts{Local: req.Local})
		if err == nil {
			_, _ = synclog.Save(p.Path, srv.Name, buf.String())
		}
		releaseID = id
		return map[string]any{"release": id}, err
	})
	if err != nil {
		guiErr(w, 409, err)
		return
	}
	<-j.done
	switch {
	case conflict != nil:
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(409)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": lang.L.EditConflict, "conflict": conflict})
	case j.Error != "":
		guiErr(w, 502, errors.New(j.Error))
	default:
		guiJSON(w, map[string]any{"release": releaseID, "hash": contentHash([]byte(req.Content)), "job_id": j.ID})
	}
}

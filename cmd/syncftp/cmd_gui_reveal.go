package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"syncftp/internal/lang"
)

// apiLocalReveal projenin (server verilirse o profilin) yerel klasörünü dosya yöneticisinde açar.
// Yol istemciden alınmaz: config'ten çözülür (scanForServer ile aynı hesap).
func (sv *guiServer) apiLocalReveal(w http.ResponseWriter, r *http.Request) {
	p, cfg, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	var req struct {
		Server string `json:"server"`
	}
	if err := guiDecode(r, &req); err != nil {
		guiErr(w, 400, err)
		return
	}
	dir := p.Path
	if req.Server != "" {
		srv, ok := findServer(cfg, req.Server)
		if !ok {
			guiErr(w, 404, fmt.Errorf(lang.L.ServerNotFoundFmt, req.Server))
			return
		}
		dir = filepath.Join(p.Path, srv.EffectiveLocalPath(cfg.Project))
	}
	if st, err := os.Stat(dir); err != nil || !st.IsDir() {
		guiErr(w, 404, fmt.Errorf(lang.L.LocalDirMissingFmt, dir, dir))
		return
	}
	if err := revealInFileManager(dir, false); err != nil {
		guiErr(w, 500, err)
		return
	}
	guiJSON(w, map[string]any{"ok": true})
}

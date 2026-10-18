package main

import (
	"bufio"
	"bytes"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"

	"syncftp/internal/config"
	ftpclient "syncftp/internal/ftp"
	"syncftp/internal/lang"
	"syncftp/internal/projects"
	"syncftp/internal/scanner"
)

// ══════════════════════════════════════════════════════════════════════════════
// Uygulama ayarları (#/settings) — proje bağımsız, makine geneli ayarlar:
// Windows ile başlat, webhook dinleyici portu, hash cache, known_hosts, veri klasörleri.
// Proje ayarları (sync/global config) projenin kendi "Ayarlar" sekmesinde kalır.
// ══════════════════════════════════════════════════════════════════════════════

func (sv *guiServer) registerSettingsAPI(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/settings", sv.apiSettings)
	mux.HandleFunc("PUT /api/settings", sv.apiSettingsPut)
	mux.HandleFunc("POST /api/settings/cache-clear", sv.apiSettingsCacheClear)
	mux.HandleFunc("DELETE /api/settings/known-hosts/{line}", sv.apiSettingsKnownHostDel)
	mux.HandleFunc("POST /api/settings/open", sv.apiSettingsOpen)
}

type knownHostEntry struct {
	Line        int    `json:"line"`
	Hosts       string `json:"hosts"`
	Type        string `json:"type"`
	Fingerprint string `json:"fingerprint"`
}

// readKnownHosts syncFTP'nin kendi known_hosts dosyasını satır numaralarıyla okur.
// knownHostsMu panelden eş zamanlı silmeleri sıralar.
var knownHostsMu sync.Mutex

func readKnownHosts() []knownHostEntry {
	out := []knownHostEntry{}
	data, err := os.ReadFile(ftpclient.KnownHostsPath())
	if err != nil {
		return out
	}
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	n := 0
	for sc.Scan() {
		n++
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		_, hosts, key, _, _, err := ssh.ParseKnownHosts([]byte(line))
		if err != nil {
			continue
		}
		out = append(out, knownHostEntry{Line: n, Hosts: strings.Join(hosts, ", "), Type: key.Type(), Fingerprint: ssh.FingerprintSHA256(key)})
	}
	return out
}

func dirSize(dir string) (int64, int) {
	var size int64
	files := 0
	_ = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if fi, err := d.Info(); err == nil {
			size += fi.Size()
			files++
		}
		return nil
	})
	return size, files
}

func (sv *guiServer) apiSettings(w http.ResponseWriter, r *http.Request) {
	exe, _ := os.Executable()
	st := loadGUIPersist()
	port := st.HookPort
	if port == 0 {
		port = defaultHookPort
	}
	cacheBytes, cacheFiles := dirSize(scanner.CacheDir())
	guiJSON(w, map[string]any{
		"version":             appVersion,
		"exe":                 exe,
		"os":                  runtime.GOOS + "/" + runtime.GOARCH,
		"go":                  runtime.Version(),
		"config_dir":          guiStateDir(),
		"cache_dir":           scanner.CacheDir(),
		"cache_bytes":         cacheBytes,
		"cache_files":         cacheFiles,
		"registry":            projects.Path(),
		"known_hosts_path":    ftpclient.KnownHostsPath(),
		"known_hosts":         readKnownHosts(),
		"autostart":           autostartEnabled(),
		"autostart_supported": runtime.GOOS == "windows",
		"tray":                sv.hasTray,
		"hook_port":           port,
		"hook_port_active":    sv.hooks.activePort(),
		"hook_default_port":   defaultHookPort,
		"edit_save_target":    st.EditSaveTarget,
		"locations":           dataLocations(),
	})
}

func (sv *guiServer) apiSettingsPut(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Autostart *bool `json:"autostart"`
		HookPort  *int  `json:"hook_port"`
		// EditSaveTarget "" | "both" | "remote"
		EditSaveTarget *string `json:"edit_save_target"`
	}
	if err := guiDecode(r, &req); err != nil {
		guiErr(w, 400, err)
		return
	}
	if req.Autostart != nil {
		if err := setAutostart(*req.Autostart); err != nil {
			guiErr(w, 500, err)
			return
		}
	}
	if req.HookPort != nil {
		p := *req.HookPort
		if p != 0 && (p < 1024 || p > 65535) {
			guiErr(w, 400, errors.New("port: 1024-65535 (0 = default)"))
			return
		}
		updateGUIPersist(func(st *guiPersist) { st.HookPort = p })
	}
	if req.EditSaveTarget != nil {
		v := *req.EditSaveTarget
		if v != "" && v != "both" && v != "remote" {
			guiErr(w, 400, errors.New("edit_save_target: \"\" | both | remote"))
			return
		}
		updateGUIPersist(func(st *guiPersist) { st.EditSaveTarget = v })
	}
	guiJSON(w, true)
}

func (sv *guiServer) apiSettingsCacheClear(w http.ResponseWriter, r *http.Request) {
	if err := scanner.ClearAllCache(); err != nil {
		guiErr(w, 500, err)
		return
	}
	guiJSON(w, true)
}

// apiSettingsKnownHostDel syncFTP known_hosts dosyasından bir satırı siler (sunucu anahtarı
// değiştiyse bağlantı reddedilir; kayıt silinince bir sonraki bağlantıda yeniden öğrenilir).
// Satır numarası liste çizildiğinden beri kaymış olabilir (yeni host öğrenildi, başka sekmede silindi)
// → istemci satırın parmak izini de yollar (?fp=), satır hâlâ o anahtarsa silinir, değilse 409.
func (sv *guiServer) apiSettingsKnownHostDel(w http.ResponseWriter, r *http.Request) {
	n, err := strconv.Atoi(r.PathValue("line"))
	fp := r.URL.Query().Get("fp")
	if err != nil || n < 1 || fp == "" {
		guiErr(w, 400, errors.New("line, fp"))
		return
	}
	knownHostsMu.Lock()
	defer knownHostsMu.Unlock()
	p := ftpclient.KnownHostsPath()
	data, err := os.ReadFile(p)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	lines := strings.Split(string(data), "\n")
	if n > len(lines) {
		guiErr(w, 404, errors.New("line"))
		return
	}
	if _, _, key, _, _, err := ssh.ParseKnownHosts([]byte(strings.TrimSpace(lines[n-1]))); err != nil || ssh.FingerprintSHA256(key) != fp {
		guiErr(w, 409, errors.New(lang.L.GuiKnownHostChanged))
		return
	}
	lines = append(lines[:n-1], lines[n:]...)
	if err := os.WriteFile(p, []byte(strings.Join(lines, "\n")), 0600); err != nil {
		guiErr(w, 500, err)
		return
	}
	guiJSON(w, readKnownHosts())
}

// dataLocation syncFTP'nin diske yazdığı yerlerden biri (Ayarlar → Sistem → Veri konumları).
type dataLocation struct {
	Key     string `json:"key"`
	Group   string `json:"group"` // app | cache | temp | project
	Path    string `json:"path"`
	File    bool   `json:"file"`
	Exists  bool   `json:"exists"`
	Project string `json:"project,omitempty"`
}

// dataLocations uygulama geneli dosyalar + bu panelin projelerinin .syncftp klasörleri.
// "Klasörü aç" da YALNIZCA bu listedeki yolları açar (beyaz liste).
func dataLocations() []dataLocation {
	cacheRoot := filepath.Dir(scanner.CacheDir())
	list := []dataLocation{
		{Key: "config_dir", Group: "app", Path: guiStateDir()},
		{Key: "library", Group: "app", Path: config.LibraryPath(), File: true},
		{Key: "gui_state", Group: "app", Path: guiPersistPath(), File: true},
		{Key: "gui_lang", Group: "app", Path: lang.UserLangPath(), File: true},
		{Key: "registry", Group: "app", Path: projects.Path(), File: true},
		{Key: "known_hosts", Group: "app", Path: ftpclient.KnownHostsPath(), File: true},
		{Key: "cloudflared", Group: "app", Path: cloudflaredLocalPath(), File: true},
		{Key: "cache_dir", Group: "cache", Path: cacheRoot},
		{Key: "hashcache", Group: "cache", Path: scanner.CacheDir()},
		{Key: "panel_profile", Group: "cache", Path: filepath.Join(cacheRoot, "panel-profile")},
		{Key: "temp", Group: "temp", Path: os.TempDir()},
	}
	for _, p := range projects.List() {
		list = append(list, dataLocation{Key: "project:" + p.Name, Group: "project", Project: p.Name, Path: filepath.Join(p.Path, ".syncftp")})
	}
	for i := range list {
		_, err := os.Stat(list[i].Path)
		list[i].Exists = err == nil
	}
	return list
}

// apiSettingsOpen veri konumlarından birini dosya yöneticisinde açar (dosyaysa seçili olarak)
// ve pencereyi öne getirir.
func (sv *guiServer) apiSettingsOpen(w http.ResponseWriter, r *http.Request) {
	var req struct {
		What string `json:"what"`
	}
	if err := guiDecode(r, &req); err != nil {
		guiErr(w, 400, err)
		return
	}
	for _, loc := range dataLocations() {
		if loc.Key != req.What {
			continue
		}
		path, sel := loc.Path, loc.File
		if !loc.Exists {
			if loc.File {
				path, sel = filepath.Dir(loc.Path), false // dosya henüz yok → klasörü
			}
			if loc.Group != "project" {
				_ = os.MkdirAll(path, 0700)
			}
		}
		if err := revealInFileManager(path, sel); err != nil {
			guiErr(w, 500, err)
			return
		}
		guiJSON(w, true)
		return
	}
	guiErr(w, 400, errors.New("what"))
}

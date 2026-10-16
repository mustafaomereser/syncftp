package main

import (
	"bufio"
	"bytes"
	"errors"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"golang.org/x/crypto/ssh"

	ftpclient "syncftp/internal/ftp"
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
	})
}

func (sv *guiServer) apiSettingsPut(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Autostart *bool `json:"autostart"`
		HookPort  *int  `json:"hook_port"`
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
func (sv *guiServer) apiSettingsKnownHostDel(w http.ResponseWriter, r *http.Request) {
	n, err := strconv.Atoi(r.PathValue("line"))
	if err != nil || n < 1 {
		guiErr(w, 400, errors.New("line"))
		return
	}
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
	lines = append(lines[:n-1], lines[n:]...)
	if err := os.WriteFile(p, []byte(strings.Join(lines, "\n")), 0600); err != nil {
		guiErr(w, 500, err)
		return
	}
	guiJSON(w, readKnownHosts())
}

// apiSettingsOpen bilinen veri klasörlerinden birini dosya yöneticisinde açar (yalnızca beyaz liste).
func (sv *guiServer) apiSettingsOpen(w http.ResponseWriter, r *http.Request) {
	var req struct {
		What string `json:"what"`
	}
	if err := guiDecode(r, &req); err != nil {
		guiErr(w, 400, err)
		return
	}
	dirs := map[string]string{
		"config":      guiStateDir(),
		"cache":       filepath.Dir(scanner.CacheDir()),
		"known_hosts": filepath.Dir(ftpclient.KnownHostsPath()),
	}
	dir, ok := dirs[req.What]
	if !ok {
		guiErr(w, 400, errors.New("what"))
		return
	}
	_ = os.MkdirAll(dir, 0700)
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", dir)
	case "darwin":
		cmd = exec.Command("open", dir)
	default:
		cmd = exec.Command("xdg-open", dir)
	}
	_ = cmd.Start()
	go func() { _ = cmd.Wait() }()
	guiJSON(w, true)
}

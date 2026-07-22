package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	goftp "github.com/jlaffaye/ftp"
	"github.com/spf13/cobra"

	"syncftp/internal/config"
	"syncftp/internal/failed"
	ftpclient "syncftp/internal/ftp"
	"syncftp/internal/ignore"
	"syncftp/internal/milestone"
	"syncftp/internal/release"
	"syncftp/internal/scanner"
	"syncftp/internal/state"
)

var servePort int

func init() {
	serveCmd.Flags().IntVar(&servePort, "port", 8080, "Dinlenecek port")
	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Yerel HTTP API sunucusu başlatır (PHP/web arayüzleri için)",
	RunE:  runServe,
}

// ── sunucu ────────────────────────────────────────────────────────────────────

func runServe(cmd *cobra.Command, args []string) error {
	dir, _ := os.Getwd()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/servers", wrap(dir, handleServers))
	mux.HandleFunc("/api/status", wrap(dir, handleStatus))
	mux.HandleFunc("/api/sync", wrap(dir, handleSync))
	mux.HandleFunc("/api/sync/stream", wrap(dir, handleSyncStream))
	mux.HandleFunc("/api/failed", wrap(dir, handleFailed))
	mux.HandleFunc("/api/releases", wrap(dir, handleReleases))
	mux.HandleFunc("/api/reload", wrap(dir, handleReload))
	mux.HandleFunc("/api/trigger/github", wrap(dir, handleGitHubTrigger))
	mux.HandleFunc("/api/remote/ls", wrap(dir, handleRemoteLs))
	mux.HandleFunc("/api/remote/cat", wrap(dir, handleRemoteCat))
	mux.HandleFunc("/api/remote/get", wrap(dir, handleRemoteGet))
	mux.HandleFunc("/api/remote/rm", wrap(dir, handleRemoteRm))

	addr := fmt.Sprintf("127.0.0.1:%d", servePort)
	fmt.Printf("syncFTP API → http://%s\n\n", addr)
	fmt.Println("Endpoint'ler:")
	fmt.Printf("  GET    /api/servers\n")
	fmt.Printf("  GET    /api/status[?server=<ad>]\n")
	fmt.Printf("  POST   /api/sync\n")
	fmt.Printf("  GET    /api/sync/stream?server=<ad>[&full=true&dry_run=true]   (SSE canlı log)\n")
	fmt.Printf("  GET    /api/failed[?server=<ad>]\n")
	fmt.Printf("  GET    /api/releases[?server=<ad>&limit=10]\n")
	fmt.Printf("  POST   /api/reload\n")
	fmt.Printf("  POST   /api/trigger/github[?server=<ad>&branch=main&secret=xxx]\n")
	fmt.Printf("  GET    /api/remote/ls?server=<ad>&path=<yol>[&recursive=true]\n")
	fmt.Printf("  GET    /api/remote/cat?server=<ad>&path=<yol>[&max_kb=10]\n")
	fmt.Printf("  GET    /api/remote/get?server=<ad>&path=<yol>\n")
	fmt.Printf("  DELETE /api/remote/rm?server=<ad>&path=<yol>[&recursive=true]\n")
	fmt.Println("\nDurdurmak için Ctrl+C")

	return http.ListenAndServe(addr, corsMiddleware(mux))
}

// ── middleware ────────────────────────────────────────────────────────────────

type handler func(w http.ResponseWriter, r *http.Request, configDir string)

func wrap(dir string, h handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { h(w, r, dir) }
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ── JSON yardımcıları ─────────────────────────────────────────────────────────

type apiResp struct {
	OK    bool   `json:"ok"`
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

func jsonOK(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(apiResp{OK: true, Data: data})
}

func jsonErr(w http.ResponseWriter, code int, err error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(apiResp{OK: false, Error: err.Error()})
}

func loadCfg(configDir string) (*config.Config, error) {
	return config.Load(configDir)
}

func getServer(cfg *config.Config, name string) (config.Server, error) {
	servers := cfg.EnabledServers()
	if name == "" {
		if len(servers) == 1 {
			return servers[0], nil
		}
		names := make([]string, len(servers))
		for i, s := range servers {
			names[i] = s.Name
		}
		return config.Server{}, fmt.Errorf("server parametresi zorunlu (mevcut: %s)", strings.Join(names, ", "))
	}
	for _, s := range servers {
		if s.Name == name {
			return s, nil
		}
	}
	return config.Server{}, fmt.Errorf("sunucu bulunamadı: %q", name)
}

// ── GET /api/servers ──────────────────────────────────────────────────────────

type serverInfo struct {
	Name           string   `json:"name"`
	Host           string   `json:"host"`
	Port           int      `json:"port"`
	User           string   `json:"user"`
	RemotePath     string   `json:"remote_path"`
	Enabled        bool     `json:"enabled"`
	MaxConnections int      `json:"max_connections"`
	MaxRetries     int      `json:"max_retries"`
	Include        []string `json:"include"`
	Exclude        []string `json:"exclude"`
}

func handleServers(w http.ResponseWriter, r *http.Request, configDir string) {
	if r.Method != http.MethodGet {
		jsonErr(w, 405, fmt.Errorf("GET bekleniyor"))
		return
	}
	cfg, err := loadCfg(configDir)
	if err != nil {
		jsonErr(w, 500, err)
		return
	}
	var list []serverInfo
	for _, s := range cfg.Servers {
		list = append(list, serverInfo{
			Name: s.Name, Host: s.Host, Port: s.Port, User: s.User,
			RemotePath: s.RemotePath, Enabled: s.Enabled,
			MaxConnections: s.MaxConnections, MaxRetries: s.MaxRetries,
			Include: s.Include, Exclude: s.Exclude,
		})
	}
	jsonOK(w, list)
}

// ── GET /api/status ───────────────────────────────────────────────────────────

type statusData struct {
	Server        string   `json:"server"`
	Host          string   `json:"host"`
	FirstSyncDone bool     `json:"first_sync_done"`
	Milestone     string   `json:"milestone,omitempty"` // set ise new/changed mtime filtreli
	New           []string `json:"new"`
	Changed       []string `json:"changed"`
	Deleted       []string `json:"deleted"`
}

func handleStatus(w http.ResponseWriter, r *http.Request, configDir string) {
	if r.Method != http.MethodGet {
		jsonErr(w, 405, fmt.Errorf("GET bekleniyor"))
		return
	}
	cfg, err := loadCfg(configDir)
	if err != nil {
		jsonErr(w, 500, err)
		return
	}

	projectDir := filepath.Join(configDir, cfg.Project.LocalPath)
	matcher, err := ignore.Load(projectDir, cfg.Sync.IgnoreFiles)
	if err != nil {
		jsonErr(w, 500, err)
		return
	}
	files, err := scanner.Scan(projectDir, matcher)
	if err != nil {
		jsonErr(w, 500, err)
		return
	}
	current := make(map[string]string, len(files))
	for _, f := range files {
		current[f.RelPath] = f.Hash
	}

	filterName := r.URL.Query().Get("server")
	servers := cfg.EnabledServers()
	if filterName != "" {
		var found []config.Server
		for _, s := range servers {
			if s.Name == filterName {
				found = append(found, s)
			}
		}
		if len(found) == 0 {
			jsonErr(w, 404, fmt.Errorf("sunucu bulunamadı: %q", filterName))
			return
		}
		servers = found
	}

	var results []statusData
	for _, srv := range servers {
		st, err := state.Load(configDir, srv.Name)
		if err != nil {
			continue
		}
		diff := state.Diff(st, current)
		msDate := ""
		if ms, _ := milestone.Load(configDir, srv.Name); ms != nil {
			diff.New, _ = milestoneFilterMtime(diff.New, projectDir, ms.Date)
			diff.Changed, _ = milestoneFilterMtime(diff.Changed, projectDir, ms.Date)
			msDate = ms.Date.Format("2006-01-02 15:04:05")
		}
		results = append(results, statusData{
			Server:        srv.Name,
			Host:          srv.Host,
			FirstSyncDone: st.FirstSyncDone,
			Milestone:     msDate,
			New:           nullSlice(diff.New),
			Changed:       nullSlice(diff.Changed),
			Deleted:       nullSlice(diff.Deleted),
		})
	}
	jsonOK(w, results)
}

// ── POST /api/sync ────────────────────────────────────────────────────────────

type syncRequest struct {
	Server      string   `json:"server"`
	Full        bool     `json:"full"`
	DryRun      bool     `json:"dry_run"`
	Include     []string `json:"include"`
	Exclude     []string `json:"exclude"`
	RetryFailed bool     `json:"retry_failed"`
}

type syncFileResult struct {
	Path     string `json:"path"`
	Attempts int    `json:"attempts,omitempty"`
	Error    string `json:"error,omitempty"`
}

type syncData struct {
	Server   string           `json:"server"`
	DryRun   bool             `json:"dry_run"`
	Uploaded []syncFileResult `json:"uploaded"`
	Failed   []syncFileResult `json:"failed"`
	Skipped  []string         `json:"skipped"`
	Deleted  []string         `json:"deleted_locally"`
}

func handleSync(w http.ResponseWriter, r *http.Request, configDir string) {
	if r.Method != http.MethodPost {
		jsonErr(w, 405, fmt.Errorf("POST bekleniyor"))
		return
	}

	var req syncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err.Error() != "EOF" {
		jsonErr(w, 400, fmt.Errorf("geçersiz JSON: %w", err))
		return
	}

	cfg, err := loadCfg(configDir)
	if err != nil {
		jsonErr(w, 500, err)
		return
	}

	projectDir := filepath.Join(configDir, cfg.Project.LocalPath)
	matcher, err := ignore.Load(projectDir, cfg.Sync.IgnoreFiles)
	if err != nil {
		jsonErr(w, 500, err)
		return
	}
	files, err := scanner.Scan(projectDir, matcher)
	if err != nil {
		jsonErr(w, 500, err)
		return
	}
	current := make(map[string]string, len(files))
	byRel := make(map[string]scanner.File, len(files))
	for _, f := range files {
		current[f.RelPath] = f.Hash
		byRel[f.RelPath] = f
	}

	servers := cfg.EnabledServers()
	if req.Server != "" {
		srv, err := getServer(cfg, req.Server)
		if err != nil {
			jsonErr(w, 404, err)
			return
		}
		servers = []config.Server{srv}
	}

	var results []syncData
	for _, srv := range servers {
		res, err := apiSyncServer(configDir, cfg, srv, current, byRel, req)
		if err != nil {
			jsonErr(w, 500, fmt.Errorf("%s: %w", srv.Name, err))
			return
		}
		results = append(results, res)
	}

	if len(results) == 1 {
		jsonOK(w, results[0])
	} else {
		jsonOK(w, results)
	}
}

func apiSyncServer(configDir string, cfg *config.Config, srv config.Server, current map[string]string, byRel map[string]scanner.File, req syncRequest) (syncData, error) {
	res := syncData{Server: srv.Name, DryRun: req.DryRun}

	st, err := state.Load(configDir, srv.Name)
	if err != nil {
		return res, err
	}

	effectiveInclude := cfg.Sync.Include
	if len(srv.Include) > 0 {
		effectiveInclude = srv.Include
	}
	if len(req.Include) > 0 {
		effectiveInclude = req.Include
	}
	effectiveExclude := append(append([]string{}, cfg.Sync.Exclude...), srv.Exclude...)
	effectiveExclude = append(effectiveExclude, req.Exclude...)

	var toUpload []string

	if req.RetryFailed {
		fl, err := failed.Load(configDir, srv.Name)
		if err != nil {
			return res, err
		}
		if fl != nil {
			for _, rel := range fl.Files {
				if _, exists := current[rel]; exists {
					toUpload = append(toUpload, rel)
				}
			}
		}
	} else {
		isFirst := !st.FirstSyncDone || req.Full
		if isFirst {
			if cfg.FirstSync.Full || req.Full {
				toUpload = mapKeys(current)
			} else {
				toUpload = filterByInclude(mapKeys(current), effectiveInclude)
			}
		} else {
			diff := state.Diff(st, current)
			res.Deleted = nullSlice(diff.Deleted)
			toUpload = append(diff.New, diff.Changed...)
			if len(effectiveInclude) > 0 {
				toUpload = filterByInclude(toUpload, effectiveInclude)
			}
		}
		toUpload = filterByExclude(toUpload, effectiveExclude)

		// Milestone filtresi: sunucuda milestone varsa yalnızca o tarihten sonra
		// değişen (mtime) dosyalar sync edilir
		if ms, _ := milestone.Load(configDir, srv.Name); ms != nil {
			var kept []string
			for _, rel := range toUpload {
				if info, statErr := os.Stat(byRel[rel].AbsPath); statErr == nil && !info.ModTime().Before(ms.Date) {
					kept = append(kept, rel)
				}
			}
			toUpload = kept
		}
	}

	sort.Strings(toUpload)

	// Korunanları ayır
	var tasks []ftpclient.UploadTask
	for _, rel := range toUpload {
		if ftpclient.IsProtected(rel, cfg.Sync.Protect) {
			res.Skipped = append(res.Skipped, rel)
			continue
		}
		f := byRel[rel]
		tasks = append(tasks, ftpclient.UploadTask{LocalPath: f.AbsPath, RelPath: rel, Hash: current[rel]})
	}

	if req.DryRun {
		for _, t := range tasks {
			res.Uploaded = append(res.Uploaded, syncFileResult{Path: t.RelPath})
		}
		return res, nil
	}

	if len(tasks) == 0 {
		return res, nil
	}

	pool, err := ftpclient.NewPool(srv)
	if err != nil {
		return res, err
	}
	defer pool.Close()

	uploadResults := pool.Upload(tasks)
	successFiles := make(map[string]string)

	for _, r := range uploadResults {
		if r.Err != nil {
			res.Failed = append(res.Failed, syncFileResult{Path: r.RelPath, Attempts: r.Attempts, Error: r.Err.Error()})
		} else {
			res.Uploaded = append(res.Uploaded, syncFileResult{Path: r.RelPath, Attempts: r.Attempts})
			successFiles[r.RelPath] = r.Hash
		}
	}

	// Failed listesini güncelle
	var failedPaths []string
	for _, f := range res.Failed {
		failedPaths = append(failedPaths, f.Path)
	}
	if len(failedPaths) > 0 {
		_ = failed.Save(configDir, srv.Name, failedPaths)
	} else {
		failed.Clear(configDir, srv.Name)
	}

	// State güncelle
	for rel, hash := range successFiles {
		st.Files[rel] = hash
	}
	for rel := range st.Files {
		if _, exists := current[rel]; !exists {
			delete(st.Files, rel)
		}
	}
	st.FirstSyncDone = true
	_ = state.Save(configDir, st)

	if len(successFiles) > 0 {
		_, _ = release.Create(configDir, srv.Name, successFiles)
	}

	return res, nil
}

// ── GET /api/failed ───────────────────────────────────────────────────────────

func handleFailed(w http.ResponseWriter, r *http.Request, configDir string) {
	if r.Method != http.MethodGet {
		jsonErr(w, 405, fmt.Errorf("GET bekleniyor"))
		return
	}

	cfg, err := loadCfg(configDir)
	if err != nil {
		jsonErr(w, 500, err)
		return
	}

	filterName := r.URL.Query().Get("server")
	servers := cfg.EnabledServers()

	type failedData struct {
		Server    string   `json:"server"`
		CreatedAt string   `json:"created_at,omitempty"`
		Files     []string `json:"files"`
	}

	var results []failedData
	for _, srv := range servers {
		if filterName != "" && srv.Name != filterName {
			continue
		}
		fl, err := failed.Load(configDir, srv.Name)
		if err != nil || fl == nil {
			continue
		}
		results = append(results, failedData{
			Server:    srv.Name,
			CreatedAt: fl.CreatedAt.Format("2006-01-02 15:04:05"),
			Files:     fl.Files,
		})
	}
	jsonOK(w, results)
}

// ── GET /api/remote/ls ────────────────────────────────────────────────────────

type lsEntry struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Size uint64 `json:"size"`
	Date string `json:"date"`
}

func handleRemoteLs(w http.ResponseWriter, r *http.Request, configDir string) {
	if r.Method != http.MethodGet {
		jsonErr(w, 405, fmt.Errorf("GET bekleniyor"))
		return
	}

	cfg, err := loadCfg(configDir)
	if err != nil {
		jsonErr(w, 500, err)
		return
	}
	srv, err := getServer(cfg, r.URL.Query().Get("server"))
	if err != nil {
		jsonErr(w, 400, err)
		return
	}

	remotePath := srv.RemotePath
	if p := r.URL.Query().Get("path"); p != "" {
		remotePath = resolveRemotePath(srv, p)
	}
	recursive := r.URL.Query().Get("recursive") == "true"

	client, err := ftpclient.Connect(srv)
	if err != nil {
		jsonErr(w, 502, err)
		return
	}
	defer client.Close()

	entries, err := collectEntries(client, remotePath, recursive)
	if err != nil {
		jsonErr(w, 502, err)
		return
	}
	jsonOK(w, entries)
}

func collectEntries(client *ftpclient.Client, remotePath string, recursive bool) ([]lsEntry, error) {
	raw, err := client.List(remotePath)
	if err != nil {
		return nil, err
	}
	var out []lsEntry
	for _, e := range raw {
		if e.Name == "." || e.Name == ".." {
			continue
		}
		typ := "file"
		if e.Type == goftp.EntryTypeFolder {
			typ = "dir"
		} else if e.Type == goftp.EntryTypeLink {
			typ = "link"
		}
		out = append(out, lsEntry{Name: e.Name, Type: typ, Size: e.Size, Date: e.Time.Format("2006-01-02 15:04:05")})

		if recursive && e.Type == goftp.EntryTypeFolder {
			sub, err := collectEntries(client, remotePath+"/"+e.Name, true)
			if err == nil {
				for i := range sub {
					sub[i].Name = e.Name + "/" + sub[i].Name
				}
				out = append(out, sub...)
			}
		}
	}
	return out, nil
}

// ── GET /api/remote/cat ───────────────────────────────────────────────────────

func handleRemoteCat(w http.ResponseWriter, r *http.Request, configDir string) {
	if r.Method != http.MethodGet {
		jsonErr(w, 405, fmt.Errorf("GET bekleniyor"))
		return
	}

	cfg, err := loadCfg(configDir)
	if err != nil {
		jsonErr(w, 500, err)
		return
	}
	srv, err := getServer(cfg, r.URL.Query().Get("server"))
	if err != nil {
		jsonErr(w, 400, err)
		return
	}

	p := r.URL.Query().Get("path")
	if p == "" {
		jsonErr(w, 400, fmt.Errorf("path parametresi zorunlu"))
		return
	}
	maxKB := 10
	if v := r.URL.Query().Get("max_kb"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxKB = n
		}
	}

	client, err := ftpclient.Connect(srv)
	if err != nil {
		jsonErr(w, 502, err)
		return
	}
	defer client.Close()

	remotePath := resolveRemotePath(srv, p)
	data, err := client.Preview(remotePath, int64(maxKB)*1024)
	if err != nil {
		jsonErr(w, 502, err)
		return
	}

	truncated := len(data) == maxKB*1024
	jsonOK(w, map[string]any{
		"path":      remotePath,
		"content":   string(data),
		"truncated": truncated,
		"max_kb":    maxKB,
	})
}

// ── GET /api/remote/get ───────────────────────────────────────────────────────

func handleRemoteGet(w http.ResponseWriter, r *http.Request, configDir string) {
	if r.Method != http.MethodGet {
		jsonErr(w, 405, fmt.Errorf("GET bekleniyor"))
		return
	}

	cfg, err := loadCfg(configDir)
	if err != nil {
		jsonErr(w, 500, err)
		return
	}
	srv, err := getServer(cfg, r.URL.Query().Get("server"))
	if err != nil {
		jsonErr(w, 400, err)
		return
	}

	p := r.URL.Query().Get("path")
	if p == "" {
		jsonErr(w, 400, fmt.Errorf("path parametresi zorunlu"))
		return
	}

	// json=true ise base64 içinde döner, aksi hâlde raw dosya
	asJSON := r.URL.Query().Get("json") == "true"

	client, err := ftpclient.Connect(srv)
	if err != nil {
		jsonErr(w, 502, err)
		return
	}
	defer client.Close()

	remotePath := resolveRemotePath(srv, p)

	if asJSON {
		// Tüm içeriği belleğe alıp base64 ile döndür
		data, err := client.Preview(remotePath, 50*1024*1024) // 50 MB limit
		if err != nil {
			jsonErr(w, 502, err)
			return
		}
		jsonOK(w, map[string]any{
			"path":     remotePath,
			"filename": filepath.Base(p),
			"content":  base64.StdEncoding.EncodeToString(data),
			"size":     len(data),
		})
		return
	}

	// Raw dosya indirme — temp dosyaya yaz, sonra stream et
	tmp, err := os.CreateTemp("", "syncftp-*")
	if err != nil {
		jsonErr(w, 500, err)
		return
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	if err := client.Download(remotePath, tmp.Name()); err != nil {
		jsonErr(w, 502, err)
		return
	}

	info, err := tmp.Stat()
	if err != nil {
		jsonErr(w, 500, err)
		return
	}
	tmp.Seek(0, 0)
	filename := filepath.Base(p)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeContent(w, r, filename, info.ModTime(), tmp)
}

// ── DELETE /api/remote/rm ─────────────────────────────────────────────────────

func handleRemoteRm(w http.ResponseWriter, r *http.Request, configDir string) {
	if r.Method != http.MethodDelete {
		jsonErr(w, 405, fmt.Errorf("DELETE bekleniyor"))
		return
	}

	cfg, err := loadCfg(configDir)
	if err != nil {
		jsonErr(w, 500, err)
		return
	}
	srv, err := getServer(cfg, r.URL.Query().Get("server"))
	if err != nil {
		jsonErr(w, 400, err)
		return
	}

	p := r.URL.Query().Get("path")
	if p == "" {
		jsonErr(w, 400, fmt.Errorf("path parametresi zorunlu"))
		return
	}
	recursive := r.URL.Query().Get("recursive") == "true"

	client, err := ftpclient.Connect(srv)
	if err != nil {
		jsonErr(w, 502, err)
		return
	}
	defer client.Close()

	remotePath := resolveRemotePath(srv, p)

	// Dizin mi dosya mı?
	entries, listErr := client.List(remotePath)
	isDir := listErr == nil && entries != nil

	if isDir {
		if err := client.DeleteDir(remotePath, recursive); err != nil {
			jsonErr(w, 502, err)
			return
		}
	} else {
		if err := client.DeleteFile(remotePath); err != nil {
			jsonErr(w, 502, err)
			return
		}
	}

	jsonOK(w, map[string]any{"deleted": remotePath, "was_dir": isDir})
}

// ── GET /api/sync/stream (SSE canlı log) ─────────────────────────────────────

// sseWriter — io.Writer'ı SSE event akışına çevirir.
type sseWriter struct {
	w http.ResponseWriter
	f http.Flusher
}

func (s *sseWriter) Write(p []byte) (n int, err error) {
	text := strings.TrimRight(string(p), "\r\n")
	for _, line := range strings.Split(text, "\n") {
		fmt.Fprintf(s.w, "data: %s\n\n", line)
	}
	s.f.Flush()
	return len(p), nil
}

func handleSyncStream(w http.ResponseWriter, r *http.Request, configDir string) {
	if r.Method != http.MethodGet {
		jsonErr(w, 405, fmt.Errorf("GET bekleniyor"))
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		jsonErr(w, 500, fmt.Errorf("SSE bu sunucu tarafından desteklenmiyor"))
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	sw := &sseWriter{w: w, f: flusher}

	cfg, err := config.Load(configDir)
	if err != nil {
		fmt.Fprintf(sw, "config yüklenemedi: %v", err)
		return
	}

	srvName := r.URL.Query().Get("server")
	full := r.URL.Query().Get("full") == "true"
	dryRun := r.URL.Query().Get("dry_run") == "true"

	servers := cfg.EnabledServers()
	if srvName != "" {
		srv, err := getServer(cfg, srvName)
		if err != nil {
			fmt.Fprintf(sw, "sunucu bulunamadı: %v", err)
			return
		}
		servers = []config.Server{srv}
	}

	for _, srv := range servers {
		localDir := filepath.Join(configDir, srv.EffectiveLocalPath(cfg.Project))
		matcher, _ := ignore.Load(localDir, cfg.Sync.IgnoreFiles)
		files, err := scanner.Scan(localDir, matcher)
		if err != nil {
			fmt.Fprintf(sw, "[%s] scan hatası: %v", srv.Name, err)
			continue
		}
		current := make(map[string]string, len(files))
		byKey := make(map[string]scanner.File, len(files))
		for _, f := range files {
			current[f.RelPath] = f.Hash
			byKey[f.RelPath] = f
		}

		// --full flag'ini geçici olarak uygula
		origFull := flagFull
		flagFull = full
		_ = dryRun // dry-run SSE modunda henüz desteklenmiyor — gelecekte eklenebilir
		syncToServerResult(sw, configDir, cfg, srv, current, byKey, nil, nil, nil)
		flagFull = origFull
	}

	fmt.Fprintf(sw, "[sync tamamlandı]")
}

// ── GET /api/releases ─────────────────────────────────────────────────────────

func handleReleases(w http.ResponseWriter, r *http.Request, configDir string) {
	if r.Method != http.MethodGet {
		jsonErr(w, 405, fmt.Errorf("GET bekleniyor"))
		return
	}

	srvFilter := r.URL.Query().Get("server")
	limit := 10
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}

	releasesDir := filepath.Join(configDir, ".syncftp", "releases")
	type releaseEntry struct {
		Server    string `json:"server"`
		ID        string `json:"id"`
		CreatedAt string `json:"created_at"`
		FileCount int    `json:"file_count"`
	}

	var entries []releaseEntry
	servers, _ := os.ReadDir(releasesDir)
	for _, srvDir := range servers {
		if !srvDir.IsDir() {
			continue
		}
		if srvFilter != "" && srvDir.Name() != srvFilter {
			continue
		}
		runs, _ := os.ReadDir(filepath.Join(releasesDir, srvDir.Name()))
		// En yeniden eskiye
		for i := len(runs) - 1; i >= 0 && len(entries) < limit; i-- {
			run := runs[i]
			if !run.IsDir() {
				continue
			}
			manifestPath := filepath.Join(releasesDir, srvDir.Name(), run.Name(), "manifest.json")
			data, err := os.ReadFile(manifestPath)
			if err != nil {
				continue
			}
			var m release.Manifest
			if err := json.Unmarshal(data, &m); err != nil {
				continue
			}
			entries = append(entries, releaseEntry{
				Server:    srvDir.Name(),
				ID:        run.Name(),
				CreatedAt: m.CreatedAt.Format("2006-01-02 15:04:05"),
				FileCount: m.FileCount,
			})
		}
	}
	jsonOK(w, entries)
}

// ── POST /api/reload ──────────────────────────────────────────────────────────

func handleReload(w http.ResponseWriter, r *http.Request, configDir string) {
	if r.Method != http.MethodPost {
		jsonErr(w, 405, fmt.Errorf("POST bekleniyor"))
		return
	}
	cfg, err := config.Load(configDir)
	if err != nil {
		jsonErr(w, 500, fmt.Errorf("config geçersiz: %w", err))
		return
	}
	jsonOK(w, map[string]any{
		"project": cfg.Project.Name,
		"servers": len(cfg.Servers),
		"status":  "ok",
	})
}

// ── POST /api/trigger/github ──────────────────────────────────────────────────

func handleGitHubTrigger(w http.ResponseWriter, r *http.Request, configDir string) {
	if r.Method != http.MethodPost {
		jsonErr(w, 405, fmt.Errorf("POST bekleniyor"))
		return
	}

	event := r.Header.Get("X-GitHub-Event")
	if event != "push" {
		jsonOK(w, map[string]string{"status": "skipped", "event": event})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		jsonErr(w, 400, fmt.Errorf("body okunamadı: %w", err))
		return
	}

	// HMAC imzası doğrulama (opsiyonel — ?secret=xxx query param ile)
	if secret := r.URL.Query().Get("secret"); secret != "" {
		sig := r.Header.Get("X-Hub-Signature-256")
		if !verifyGitHubSignature(body, secret, sig) {
			jsonErr(w, 401, fmt.Errorf("geçersiz imza"))
			return
		}
	}

	var payload struct {
		Ref string `json:"ref"`
	}
	json.Unmarshal(body, &payload)

	// Branch filtresi
	if branch := r.URL.Query().Get("branch"); branch != "" {
		if payload.Ref != "refs/heads/"+branch {
			jsonOK(w, map[string]string{"status": "skipped", "ref": payload.Ref})
			return
		}
	}

	srvName := r.URL.Query().Get("server")
	jsonOK(w, map[string]string{"status": "accepted", "ref": payload.Ref})

	// Sync'i arka planda başlat
	go func() {
		cfg, err := config.Load(configDir)
		if err != nil {
			return
		}
		servers := cfg.EnabledServers()
		if srvName != "" {
			srv, err := getServer(cfg, srvName)
			if err != nil {
				return
			}
			servers = []config.Server{srv}
		}
		for _, srv := range servers {
			localDir := filepath.Join(configDir, srv.EffectiveLocalPath(cfg.Project))
			matcher, _ := ignore.Load(localDir, cfg.Sync.IgnoreFiles)
			files, err := scanner.Scan(localDir, matcher)
			if err != nil {
				continue
			}
			current := make(map[string]string, len(files))
			byKey := make(map[string]scanner.File, len(files))
			for _, f := range files {
				current[f.RelPath] = f.Hash
				byKey[f.RelPath] = f
			}
			syncToServer(configDir, cfg, srv, current, byKey, nil, nil)
		}
	}()
}

func verifyGitHubSignature(body []byte, secret, sig string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(sig))
}

// ── küçük yardımcılar ─────────────────────────────────────────────────────────

func nullSlice(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}


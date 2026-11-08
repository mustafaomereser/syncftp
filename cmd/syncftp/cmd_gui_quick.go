package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"syncftp/internal/config"
	ftpclient "syncftp/internal/ftp"
	"syncftp/internal/fsutil"
	"syncftp/internal/lang"
	"syncftp/internal/projects"
	"syncftp/internal/secret"
)

// Hızlı bağlantı (FileZilla'daki gibi): projesiz, tek seferlik bağlantı. Dosya uçları proje ister; bu
// yüzden "@quick" adında sanal bir proje vardır, bağlantıları bu oturumda açılanlardır ve her biri
// Dosyalar ekranına "@conn:<ad>" hedefiyle (browseTarget) girer — yani gezme, önizleme, düzenleme,
// yükleme, silme, zip mevcut kodla çalışır; sync/freeze/rollback hedefi reddeder.
//
// Geçmiş dosyasına ŞİFRE YAZILMAZ. "Şifreyi hatırla" işaretliyse şifre işletim sistemi anahtarlığında,
// projelerin kaydından AYRI bir hesapta durur (geçmişten silince başka projeyi etkilemesin).
const (
	quickProject    = "@quick"
	quickHistoryMax = 30
	quickDialLimit  = 30 * time.Second
)

// quickEntry geçmiş satırı — şifre alanı bilerek yok.
type quickEntry struct {
	Name          string    `json:"name"`
	Protocol      string    `json:"protocol"`
	Host          string    `json:"host"`
	Port          int       `json:"port"`
	User          string    `json:"user"`
	KeyFile       string    `json:"key_file,omitempty"`
	Passive       bool      `json:"passive"`
	TLSSkipVerify bool      `json:"tls_skip_verify,omitempty"`
	Remember      bool      `json:"remember"`
	LastUsed      time.Time `json:"last_used"`
}

type quickReq struct {
	quickEntry
	Password string `json:"password"`
}

var (
	quickMu       sync.Mutex
	quickSessions = map[string]config.Connection{}
)

func quickHistoryPath() string {
	base, err := os.UserConfigDir()
	if err != nil {
		base = os.TempDir()
	}
	return filepath.Join(base, "syncftp", "quick-history.json")
}

func loadQuickHistory() []quickEntry {
	out := []quickEntry{}
	if data, err := os.ReadFile(quickHistoryPath()); err == nil {
		_ = json.Unmarshal(data, &out)
	}
	return out
}

func saveQuickHistory(list []quickEntry) error {
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	p := quickHistoryPath()
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		return err
	}
	return fsutil.WriteFileAtomic(p, data, 0600)
}

func quickAccount(e quickEntry) string {
	return secret.Account("quick-"+e.Protocol, e.User, e.Host, e.Port)
}

// quickName bağlantının görünen ve hedefte kullanılan adı: aynı hesap tekrar açılınca aynı ad.
func quickName(e quickEntry) string {
	if e.User == "" {
		return fmt.Sprintf("%s:%d", e.Host, e.Port)
	}
	return fmt.Sprintf("%s@%s:%d", e.User, e.Host, e.Port)
}

// quickProjectInfo sanal proje: yalnızca bu oturumda açılmış bağlantılar, sunucu (profil) yok.
func quickProjectInfo() (projects.Project, *config.Config, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		base = os.TempDir()
	}
	dir := filepath.Join(base, "syncftp", "quick")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return projects.Project{}, nil, err
	}
	cfg := &config.Config{Servers: []config.Server{}}
	cfg.Project.Name = quickProject
	quickMu.Lock()
	for _, c := range quickSessions {
		cfg.Connections = append(cfg.Connections, c)
	}
	quickMu.Unlock()
	sort.Slice(cfg.Connections, func(i, j int) bool { return cfg.Connections[i].Name < cfg.Connections[j].Name })
	return projects.Project{Name: quickProject, Path: dir}, cfg, nil
}

func (sv *guiServer) registerQuickAPI(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/quick", sv.apiQuickGet)
	mux.HandleFunc("POST /api/quick/connect", sv.apiQuickConnect)
	mux.HandleFunc("POST /api/quick/close", sv.apiQuickClose)
	mux.HandleFunc("POST /api/quick/save", sv.apiQuickSave)
	mux.HandleFunc("POST /api/quick/forget", sv.apiQuickForget)
}

func (sv *guiServer) apiQuickGet(w http.ResponseWriter, r *http.Request) {
	quickMu.Lock()
	open := make([]string, 0, len(quickSessions))
	for n := range quickSessions {
		open = append(open, n)
	}
	quickMu.Unlock()
	sort.Strings(open)
	guiJSON(w, map[string]any{"history": loadQuickHistory(), "open": open})
}

// quickTryConnect bağlanmayı dener; sunucu cevap vermezse pencere süresiz beklemesin.
func quickTryConnect(srv config.Server) error {
	done := make(chan error, 1)
	go func() {
		c, err := ftpclient.Connect(srv)
		if err == nil {
			c.Close()
		}
		done <- err
	}()
	select {
	case err := <-done:
		return err
	case <-time.After(quickDialLimit):
		return fmt.Errorf(lang.L.RemoteNoAnswerFmt, int(quickDialLimit/time.Second))
	}
}

func (sv *guiServer) apiQuickConnect(w http.ResponseWriter, r *http.Request) {
	var req quickReq
	if err := guiDecode(r, &req); err != nil {
		guiErr(w, 400, err)
		return
	}
	e := req.quickEntry
	e.Host = strings.TrimSpace(e.Host)
	e.User = strings.TrimSpace(e.User)
	e.KeyFile = strings.TrimSpace(e.KeyFile)
	if e.Protocol == "" {
		e.Protocol = "ftp"
	}
	if e.Host == "" {
		guiErr(w, 400, errors.New(lang.L.QuickHostRequired))
		return
	}
	if e.Port == 0 {
		e.Port = 21
		if e.Protocol == "sftp" {
			e.Port = 22
		}
	}
	e.Name = quickName(e)

	// Boş şifre + hatırlanan kayıt → anahtarlıktan; anahtar dosyasıyla girişte şifre gerekmeyebilir
	pw := req.Password
	if pw == "" && e.Remember {
		got, err := secret.Get(quickAccount(e))
		if err != nil && e.KeyFile == "" {
			guiErr(w, 400, errors.New(lang.L.QuickPasswordMissing))
			return
		}
		pw = got
	}

	c := config.Connection{Name: e.Name, Protocol: e.Protocol, Host: e.Host, Port: e.Port, User: e.User, Password: pw,
		KeyFile: e.KeyFile, Passive: e.Passive, TLSSkipVerify: e.TLSSkipVerify}
	tmp := &config.Config{Connections: []config.Connection{c}}
	srv, _ := browseTarget(tmp, connTargetPrefix+c.Name)
	if err := quickTryConnect(srv); err != nil {
		guiErr(w, 502, err)
		return
	}

	quickMu.Lock()
	quickSessions[c.Name] = c
	quickMu.Unlock()

	warn := ""
	if e.Remember {
		if pw != "" {
			if err := secret.Set(quickAccount(e), pw); err != nil {
				e.Remember = false // anahtarlık yoksa düz metne düşmez: hatırlanmaz
				warn = lang.L.QuickKeyringFailed
			}
		}
	} else {
		secret.Delete(quickAccount(e))
	}

	e.LastUsed = time.Now()
	list := []quickEntry{e}
	for _, x := range loadQuickHistory() {
		if x.Name != e.Name && len(list) < quickHistoryMax {
			list = append(list, x)
		}
	}
	if err := saveQuickHistory(list); err != nil && warn == "" {
		warn = err.Error()
	}
	guiJSON(w, map[string]any{"name": c.Name, "warning": warn})
}

// apiQuickClose oturumu kapatır: bağlantı listeden düşer, önbellekteki soket hemen kapanır, geçmiş kalır.
func (sv *guiServer) apiQuickClose(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := guiDecode(r, &req); err != nil {
		guiErr(w, 400, err)
		return
	}
	quickMu.Lock()
	c, ok := quickSessions[req.Name]
	delete(quickSessions, req.Name)
	quickMu.Unlock()
	if ok {
		srv, _ := browseTarget(&config.Config{Connections: []config.Connection{c}}, connTargetPrefix+c.Name)
		key := remoteKey(quickProject, srv)
		remoteCacheMu.Lock()
		cc := remoteCache[key]
		delete(remoteCache, key)
		remoteCacheMu.Unlock()
		if cc != nil {
			go cc.client.Close()
		}
	}
	guiJSON(w, map[string]any{"ok": true})
}

// apiQuickSave açık bağlantıyı Bağlantı kütüphanesine yazar (şifre SaveLibrary ile anahtarlığa gider).
func (sv *guiServer) apiQuickSave(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := guiDecode(r, &req); err != nil {
		guiErr(w, 400, err)
		return
	}
	quickMu.Lock()
	c, ok := quickSessions[req.Name]
	quickMu.Unlock()
	if !ok {
		guiErr(w, 404, fmt.Errorf(lang.L.ServerNotFoundFmt, req.Name))
		return
	}
	lib, err := config.LoadLibrary()
	if err != nil {
		guiErr(w, 500, err)
		return
	}
	upsertLibConn(lib, c)
	if err := config.SaveLibrary(lib); err != nil {
		guiErr(w, 500, err)
		return
	}
	guiJSON(w, map[string]any{"ok": true})
}

// apiQuickForget geçmişten siler; hatırlanan şifre de anahtarlıktan kalkar.
func (sv *guiServer) apiQuickForget(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := guiDecode(r, &req); err != nil {
		guiErr(w, 400, err)
		return
	}
	list := loadQuickHistory()
	kept := make([]quickEntry, 0, len(list))
	for _, x := range list {
		if x.Name == req.Name {
			secret.Delete(quickAccount(x))
			continue
		}
		kept = append(kept, x)
	}
	if err := saveQuickHistory(kept); err != nil {
		guiErr(w, 500, err)
		return
	}
	guiJSON(w, map[string]any{"ok": true})
}

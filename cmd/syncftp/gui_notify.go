package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"syncftp/internal/lang"
)

// Masaüstü bildirimleri (Windows tepsi balonu). Panel penceresi açık ve GÖRÜNÜRKEN bildirim
// çıkmaz (panel zaten kendi kartını/toast'ını gösteriyor); pencere kapalı, küçültülmüş ya da
// arka plandaysa iş hatası ve watch hatası tepsiden bildirilir. Ayarlar → Genel'den kapatılabilir
// (notify_off), başarılı işler isteğe bağlı (notify_done).

// trayNotifyHook tepsi kurulunca atanır (yalnızca Windows); nil → bildirim yok.
var (
	trayNotifyMu   sync.RWMutex
	trayNotifyHook func(title, body string, isErr bool, hash string)
)

func setTrayNotifyHook(fn func(title, body string, isErr bool, hash string)) {
	trayNotifyMu.Lock()
	trayNotifyHook = fn
	trayNotifyMu.Unlock()
}

// notifyDesktop politika: kapalıysa / başarı bildirimi istenmiyorsa / panel görünürse çıkmaz.
func (sv *guiServer) notifyDesktop(title, body string, isErr bool, hash string) {
	trayNotifyMu.RLock()
	fn := trayNotifyHook
	trayNotifyMu.RUnlock()
	if fn == nil {
		return
	}
	st := loadGUIPersist()
	if st.NotifyOff || (!isErr && !st.NotifyDone) || sv.panelVisible() {
		return
	}
	if r := []rune(body); len(r) > 200 {
		body = string(r[:199]) + "…"
	}
	fn(title, body, isErr, hash)
}

// onJobDone her iş bitince çağrılır (jobManager.onDone).
func (sv *guiServer) onJobDone(j *job) {
	j.mu.Lock()
	status, errText, title, proj, srv := j.Status, j.Error, j.Title, j.Project, j.Server
	j.mu.Unlock()
	where := strings.Trim(proj+" · "+srv, " ·")
	hash := "#/jobs"
	if proj != "" {
		hash = "#/p/" + proj
	}
	switch status {
	case "error":
		body := where
		if errText != "" {
			body = strings.TrimSpace(where + "\n" + errText)
		}
		sv.notifyDesktop(fmt.Sprintf(lang.L.NotifyJobFailedFmt, title), body, true, hash)
	case "done":
		if j.Kind == "watch" { // watch'ın kendisi bitti (durduruldu) — bildirmeye değmez
			return
		}
		sv.notifyDesktop(fmt.Sprintf(lang.L.NotifyJobDoneFmt, title), where, false, hash)
	}
}

// onWatchSync watch'ın her turu sonrası: hata ya da yüklenemeyen dosya varsa bildirir
// (başarılı turlar sessiz — watch'ta her kayıtta bildirim çok gürültülü olur).
func (sv *guiServer) onWatchSync(project string, res serverSyncResult) {
	var body string
	switch {
	case res.err != nil:
		body = res.err.Error()
	case res.failed > 0:
		body = fmt.Sprintf(lang.L.NotifyFilesFailedFmt, res.failed)
	default:
		return
	}
	sv.notifyDesktop(fmt.Sprintf(lang.L.NotifyWatchFailedFmt, project, res.name), body, true, "#/p/"+project)
}

// ── Panel görünürlüğü: her pencere presence bağlantısını kimlikle açar, görünürlüğünü bildirir ──

type panelVis struct {
	mu  sync.Mutex
	vis map[string]bool
}

func (sv *guiServer) setPanelVisible(id string, visible bool, remove bool) {
	if id == "" {
		return
	}
	sv.pvis.mu.Lock()
	defer sv.pvis.mu.Unlock()
	if sv.pvis.vis == nil {
		sv.pvis.vis = map[string]bool{}
	}
	if remove {
		delete(sv.pvis.vis, id)
		return
	}
	sv.pvis.vis[id] = visible
}

// panelVisible en az bir panel penceresi görünür mü (kimliksiz eski sayfalar sayılmaz).
func (sv *guiServer) panelVisible() bool {
	sv.pvis.mu.Lock()
	defer sv.pvis.mu.Unlock()
	for _, v := range sv.pvis.vis {
		if v {
			return true
		}
	}
	return false
}

// apiPresenceVisible POST /api/presence/visible {id, visible}
func (sv *guiServer) apiPresenceVisible(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID      string `json:"id"`
		Visible bool   `json:"visible"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024)).Decode(&req); err != nil {
		guiErr(w, 400, err)
		return
	}
	sv.pvis.mu.Lock()
	_, known := sv.pvis.vis[req.ID]
	sv.pvis.mu.Unlock()
	if known { // yalnızca açık presence bağlantısı olan pencere
		sv.setPanelVisible(req.ID, req.Visible, false)
	}
	guiJSON(w, true)
}

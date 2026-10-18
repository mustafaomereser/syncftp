package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"syncftp/internal/projects"
)

// guiPersist — panel süreci yeniden başladığında geri yüklenen durum.
// <UserConfigDir>/syncftp/gui-state.json
type guiPersist struct {
	// Watches açık watch'lar: proje DİZİNİ → izlenen sunucular (eski kayıtlarda boş liste = tüm aktifler). Eskiden
	// anahtar proje adıydı ve her panel tüm haritayı kendi listesiyle eziyordu: farklı klasörlerde
	// açılan iki panel birbirinin watch'larını siliyor ya da aynı projeyi ikisi birden izliyordu.
	Watches map[string][]string `json:"watches"`
	// HookTunnel açılışta trycloudflare tünelini otomatik başlat
	HookTunnel bool `json:"hook_tunnel,omitempty"`
	// HookPort webhook dinleyicisinin portu (0 → varsayılan)
	HookPort int `json:"hook_port,omitempty"`
	// HookPublicURL kendi domain'i / named tunnel kullananlar için sabit public adres
	HookPublicURL string `json:"hook_public_url,omitempty"`
	// EditSaveTarget dosya düzenleyicide Ctrl+S'in hedefi: "" = her seferinde sor, "both" = sunucu +
	// yerel, "remote" = yalnızca sunucu (kaydetme penceresindeki "bir daha sorma")
	EditSaveTarget string `json:"edit_save_target,omitempty"`
}

var guiPersistMu sync.Mutex

func guiPersistPath() string { return filepath.Join(guiStateDir(), "gui-state.json") }

func loadGUIPersist() guiPersist {
	guiPersistMu.Lock()
	defer guiPersistMu.Unlock()
	st := guiPersist{Watches: map[string][]string{}}
	if data, err := os.ReadFile(guiPersistPath()); err == nil {
		_ = json.Unmarshal(data, &st)
	}
	if st.Watches == nil {
		st.Watches = map[string][]string{}
	}
	return st
}

// updateGUIPersist oku-değiştir-yaz (eş zamanlı güncellemeler birbirini ezmesin).
func updateGUIPersist(fn func(*guiPersist)) {
	guiPersistMu.Lock()
	defer guiPersistMu.Unlock()
	st := guiPersist{Watches: map[string][]string{}}
	if data, err := os.ReadFile(guiPersistPath()); err == nil {
		_ = json.Unmarshal(data, &st)
	}
	if st.Watches == nil {
		st.Watches = map[string][]string{}
	}
	fn(&st)
	_ = os.MkdirAll(guiStateDir(), 0700)
	if data, err := json.MarshalIndent(st, "", "  "); err == nil {
		tmp := guiPersistPath() + ".tmp"
		if os.WriteFile(tmp, data, 0600) == nil {
			_ = os.Rename(tmp, guiPersistPath())
		}
	}
}

// persistWatches açık watch listesini kaydeder. Süreç kapanırken (stopAll) çağrılmaz —
// çıkışta açık olan watch'lar bir sonraki açılışta geri gelir.
func (sv *guiServer) persistWatches() {
	cur := map[string][]string{}
	for _, w := range sv.watches.all() {
		cur[w.path] = append(cur[w.path], w.Server)
	}
	updateGUIPersist(func(st *guiPersist) {
		// Yalnızca bu panelin gördüğü projelerin kayıtları değişir; başka klasördeki panellerinkiler kalır
		for k := range st.Watches {
			if _, ok := projects.Find(k); ok {
				delete(st.Watches, k)
			}
		}
		for k, v := range cur {
			st.Watches[k] = v
		}
	})
}

// restoreWatches önceki oturumdan kalan watch'ları başlatır — yalnızca bu panelin klasöründeki
// projeler (Find taban klasörle sınırlı); proje silinmişse ya da başka süreç izliyorsa atlanır.
func (sv *guiServer) restoreWatches() {
	for key, servers := range loadGUIPersist().Watches {
		if p, ok := projects.Find(key); ok {
			_, _, _ = sv.startWatch(p.Path, servers)
		}
	}
}

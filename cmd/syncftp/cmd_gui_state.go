package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// guiPersist — panel süreci yeniden başladığında geri yüklenen durum.
// <UserConfigDir>/syncftp/gui-state.json
type guiPersist struct {
	// Watches açık watch'lar: proje → sunucular (boş liste = tüm aktif sunucular)
	Watches map[string][]string `json:"watches"`
	// HookTunnel açılışta trycloudflare tünelini otomatik başlat
	HookTunnel bool `json:"hook_tunnel,omitempty"`
	// HookPort webhook dinleyicisinin portu (0 → varsayılan)
	HookPort int `json:"hook_port,omitempty"`
	// HookPublicURL kendi domain'i / named tunnel kullananlar için sabit public adres
	HookPublicURL string `json:"hook_public_url,omitempty"`
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
		if w.explicit {
			cur[w.Project] = w.Servers
		} else {
			cur[w.Project] = []string{}
		}
	}
	updateGUIPersist(func(st *guiPersist) { st.Watches = cur })
}

// restoreWatches önceki oturumdan kalan watch'ları başlatır (proje silinmişse atlanır).
func (sv *guiServer) restoreWatches() {
	for project, servers := range loadGUIPersist().Watches {
		_, _ = sv.startWatch(project, servers)
	}
}

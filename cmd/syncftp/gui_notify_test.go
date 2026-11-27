package main

import (
	"errors"
	"testing"
)

func TestNotifyPolicy(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	type call struct {
		title, hash string
		isErr       bool
	}
	var got []call
	setTrayNotifyHook(func(title, body string, isErr bool, hash string) { got = append(got, call{title, hash, isErr}) })
	defer setTrayNotifyHook(nil)

	sv := &guiServer{}
	mk := func(kind, status string) *job {
		return &job{Kind: kind, Project: "acme", Server: "Prod", Title: "Sync", Status: status, Error: "boom"}
	}
	reset := func() { got = nil }

	// Varsayılan: hata bildirilir, başarı bildirilmez, durdurulan bildirilmez
	sv.onJobDone(mk("sync", "error"))
	sv.onJobDone(mk("sync", "done"))
	sv.onJobDone(mk("sync", "stopped"))
	if len(got) != 1 || !got[0].isErr || got[0].hash != "#/p/acme" {
		t.Fatalf("default policy: %+v", got)
	}

	// Panel görünürken hiçbir şey çıkmaz; arka plana geçince çıkar
	reset()
	sv.setPanelVisible("w1", true, false)
	sv.onJobDone(mk("sync", "error"))
	if len(got) != 0 {
		t.Fatalf("visible panel should suppress: %+v", got)
	}
	sv.setPanelVisible("w1", false, false)
	sv.onJobDone(mk("sync", "error"))
	if len(got) != 1 {
		t.Fatalf("hidden panel should notify: %+v", got)
	}
	sv.setPanelVisible("w1", false, true)

	// Başarılı işler açıkken: done bildirilir, watch'ın kendisinin bitişi bildirilmez
	reset()
	updateGUIPersist(func(st *guiPersist) { st.NotifyDone = true })
	sv.onJobDone(mk("sync", "done"))
	sv.onJobDone(mk("watch", "done"))
	if len(got) != 1 || got[0].isErr {
		t.Fatalf("notify_done: %+v", got)
	}

	// Watch turu: yalnızca hata / yüklenemeyen dosya
	reset()
	sv.onWatchSync("acme", serverSyncResult{name: "Prod", uploaded: 3})
	sv.onWatchSync("acme", serverSyncResult{name: "Prod", failed: 2})
	sv.onWatchSync("acme", serverSyncResult{name: "Prod", err: errors.New("x")})
	if len(got) != 2 {
		t.Fatalf("watch sync: %+v", got)
	}

	// Kapalı: hiçbir şey
	reset()
	updateGUIPersist(func(st *guiPersist) { st.NotifyOff = true })
	sv.onJobDone(mk("sync", "error"))
	sv.onWatchSync("acme", serverSyncResult{name: "Prod", failed: 1})
	if len(got) != 0 {
		t.Fatalf("notify_off: %+v", got)
	}
}

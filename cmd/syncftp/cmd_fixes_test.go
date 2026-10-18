package main

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"syncftp/internal/config"
	"syncftp/internal/release"
)

// Kapanan trigger kanalı debounce döngüsünü sonsuza dek döndürmemeli.
func TestWatchWorkerStopsWhenTriggerClosed(t *testing.T) {
	var syncs atomic.Int32
	wk := &watchWorker{trigger: make(chan struct{}, 1)}
	done := make(chan struct{})
	go func() {
		wk.runDebounced(func() { syncs.Add(1) }, 30*time.Millisecond)
		close(done)
	}()

	// Önce normal bir tur: tetik + sessizlik → bir sync
	wk.notify()
	time.Sleep(120 * time.Millisecond)
	if n := syncs.Load(); n != 1 {
		t.Fatalf("syncs = %d, want 1", n)
	}
	// Debounce sırasında kapanış: döngü dönmeli, bekleyen sync çalışmamalı
	wk.notify()
	time.Sleep(5 * time.Millisecond)
	close(wk.trigger)
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not return after trigger was closed")
	}
	if n := syncs.Load(); n != 1 {
		t.Fatalf("syncs after close = %d, want 1", n)
	}
}

type fakeRollbackRemote struct {
	failUpload map[string]bool
	calls      []string
}

func (f *fakeRollbackRemote) UploadRaw(localPath, relPath string) error {
	f.calls = append(f.calls, "up:"+relPath)
	if f.failUpload[relPath] {
		return errors.New("boom")
	}
	return nil
}

func (f *fakeRollbackRemote) DeleteFile(remotePath string) error {
	f.calls = append(f.calls, "rm:"+remotePath)
	return nil
}

// Yeni release kısmen başarısızsa daha eski release'ler geri alınmamalı.
func TestRollbackStopsAtFirstFailedRelease(t *testing.T) {
	dir := t.TempDir()
	srv := config.Server{Name: "prod", RemotePath: "/www"}
	r2 := &release.Manifest{Version: 2, ID: "r2", Backed: []string{"ok.txt", "bad.txt"}}
	r1 := &release.Manifest{Version: 2, ID: "r1", Backed: []string{"old.txt"}, New: []string{"n.txt"}}
	fake := &fakeRollbackRemote{failUpload: map[string]bool{"bad.txt": true}}

	var out strings.Builder
	rep, err := applyRollback(&out, dir, srv, []*release.Manifest{r2, r1}, fake, nil)
	if err != nil {
		t.Fatal(err)
	}
	if r1.RolledBack || r2.RolledBack {
		t.Fatalf("no release may be marked rolled back: r1=%v r2=%v", r1.RolledBack, r2.RolledBack)
	}
	if len(rep.Releases) != 0 || len(rep.Failed) != 1 {
		t.Fatalf("report = %+v", rep)
	}
	if want := []string{"up:ok.txt", "up:bad.txt"}; !reflect.DeepEqual(fake.calls, want) {
		t.Fatalf("calls = %v, want %v (older release must not be touched)", fake.calls, want)
	}

	// Başarılı zincir: hepsi geri alınır
	fake2 := &fakeRollbackRemote{}
	r3 := &release.Manifest{Version: 2, ID: "r3", Backed: []string{"a"}}
	r4 := &release.Manifest{Version: 2, ID: "r4", New: []string{"b"}}
	rep, err = applyRollback(&out, dir, srv, []*release.Manifest{r4, r3}, fake2, nil)
	if err != nil || len(rep.Releases) != 2 || !r3.RolledBack || !r4.RolledBack {
		t.Fatalf("full chain: rep=%+v err=%v", rep, err)
	}
}

// Failsafe temizlik: protect'e dokunmaz, include'u boşaltmaz, global listeyi başka sunucuda
// var olan yol için korur ve değişikliği diskteki taze config'e yazar (çağıranınkine değil).
func TestFailsafeCleanupSafety(t *testing.T) {
	dir := t.TempDir()
	for _, p := range []string{"a/keep.txt", "b/b-only.txt"} {
		full := filepath.Join(dir, filepath.FromSlash(p))
		os.MkdirAll(filepath.Dir(full), 0755)
		os.WriteFile(full, []byte("x"), 0644)
	}
	cfg := &config.Config{
		Project: config.Project{Name: "t", DefaultPath: "."},
		Sync: config.Sync{
			Protect: []string{".env"},
			Exclude: []string{"gone.txt", "b-only.txt"},
		},
		Servers: []config.Server{
			{Name: "a", Host: "h", Port: 21, RemotePath: "/", LocalPath: "a", Enabled: true,
				Include: []string{"missing.php"}, Exclude: []string{"x.txt", "keep.txt"}},
			{Name: "b", Host: "h", Port: 21, RemotePath: "/b", LocalPath: "b", Enabled: true},
		},
	}
	if err := config.Save(dir, cfg); err != nil {
		t.Fatal(err)
	}
	mem, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	var out strings.Builder
	failsafeCleanup(&out, dir, mem, "a")

	fresh, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(fresh.Sync.Protect, []string{".env"}) {
		t.Fatalf("protect changed: %v", fresh.Sync.Protect)
	}
	if !reflect.DeepEqual(fresh.Sync.Exclude, []string{"b-only.txt"}) {
		t.Fatalf("global exclude = %v", fresh.Sync.Exclude)
	}
	a := fresh.Servers[0]
	if !reflect.DeepEqual(a.Include, []string{"missing.php"}) {
		t.Fatalf("include must not be emptied: %v", a.Include)
	}
	if !reflect.DeepEqual(a.Exclude, []string{"keep.txt"}) {
		t.Fatalf("server exclude = %v", a.Exclude)
	}
	// Çağıranın bellekteki config'i değişmez (paralel goroutine'ler paylaşıyor)
	if len(mem.Sync.Exclude) != 2 || len(mem.Servers[0].Exclude) != 2 {
		t.Fatalf("caller cfg mutated: %v / %v", mem.Sync.Exclude, mem.Servers[0].Exclude)
	}
	if !strings.Contains(out.String(), "gone.txt") || !strings.Contains(out.String(), "x.txt") {
		t.Fatalf("warnings not written to w: %q", out.String())
	}
}

// Seçili dosya rollback'i: yalnızca o dosyalara dokunan release'ler plana girer, dosyalar hedef
// release'ten önceki hâline döner; release'ler kısmen geri alınmış işaretlenir, tam rollback
// aynı dosyaları ikinci kez işlemez.
func TestSelectiveRollback(t *testing.T) {
	dir := t.TempDir()
	srv := config.Server{Name: "prod", RemotePath: "/www"}
	r3 := &release.Manifest{Version: 2, ID: "r3", Backed: []string{"a.php"}} // a'yı yeniden değiştirdi
	r2 := &release.Manifest{Version: 2, ID: "r2", Backed: []string{"x.php"}} // a'ya dokunmadı
	r1 := &release.Manifest{Version: 2, ID: "r1", Backed: []string{"a.php", "b.php"}, New: []string{"c.php"}}
	list := []*release.Manifest{r3, r2, r1}

	plan, err := rollbackPlanFiles(list, "r1", []string{"a.php", "c.php"})
	if err != nil || len(plan) != 2 || plan[0].ID != "r3" || plan[1].ID != "r1" {
		t.Fatalf("plan: %v err=%v", plan, err)
	}
	fake := &fakeRollbackRemote{}
	var out strings.Builder
	if _, err := applyRollback(&out, dir, srv, plan, fake, []string{"a.php", "c.php"}); err != nil {
		t.Fatal(err)
	}
	if want := []string{"up:a.php", "up:a.php", "rm:/www/c.php"}; !reflect.DeepEqual(fake.calls, want) {
		t.Fatalf("calls = %v, want %v", fake.calls, want)
	}
	if !r3.RolledBack || r1.RolledBack || !r1.FileRolledBack("a.php") || r1.FileRolledBack("b.php") {
		t.Fatalf("işaretler: r3=%v r1=%v files=%v", r3.RolledBack, r1.RolledBack, r1.RolledBackFiles)
	}
	// Kalanların tam rollback'i yalnızca b.php'yi işler ve r1'i tamamen geri alınmış yapar
	fake2 := &fakeRollbackRemote{}
	if _, err := applyRollback(&out, dir, srv, []*release.Manifest{r1}, fake2, nil); err != nil {
		t.Fatal(err)
	}
	if want := []string{"up:b.php"}; !reflect.DeepEqual(fake2.calls, want) || !r1.RolledBack {
		t.Fatalf("tam: calls=%v rolled=%v", fake2.calls, r1.RolledBack)
	}
	// Seçilen dosya hedefte yoksa
	if _, err := rollbackPlanFiles(list, "r2", []string{"a.php"}); err == nil {
		t.Fatal("hedefte olmayan dosya hata vermeli")
	}
}

package release

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSummariesIndex(t *testing.T) {
	dir := t.TempDir()
	base := time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)
	for i, id := range []string{"r1", "r2", "r3"} {
		m := &Manifest{CreatedAt: base.Add(time.Duration(i) * time.Minute), Files: map[string]string{"a": "h"}, Backed: []string{"a"}}
		if _, err := Write(dir, "prod", id, m); err != nil {
			t.Fatal(err)
		}
	}
	// Index ilk Summaries'te kurulur (Write, index yokken yarım index yazmaz)
	if _, err := os.Stat(indexPath(dir, "prod")); err == nil {
		t.Fatal("index Summaries'ten önce oluşmamalı")
	}
	list, err := Summaries(dir, "prod")
	if err != nil || len(list) != 3 || list[0].ID != "r3" || !list[0].HasBackup || list[0].Backed != 1 {
		t.Fatalf("ilk liste: %+v %v", list, err)
	}
	if _, err := os.Stat(indexPath(dir, "prod")); err != nil {
		t.Fatal("index yazılmalı")
	}

	// Write güncellemesi index'e yansır (rollback işareti)
	m, _ := Load(dir, "prod", "r2")
	m.RolledBack = true
	if _, err := Write(dir, "prod", "r2", m); err != nil {
		t.Fatal(err)
	}
	// Discard index'ten siler
	Discard(dir, "prod", "r1")
	idx, _ := readIndex(dir, "prod")
	if !idx["r2"].RolledBack || len(idx) != 2 {
		t.Fatalf("index güncellenmedi: %+v", idx)
	}

	// Index'i bilmeyen biri klasör ekler / siler → Summaries onarır
	_ = os.MkdirAll(Dir(dir, "prod", "r9"), 0755)
	_ = os.WriteFile(filepath.Join(Dir(dir, "prod", "r9"), "manifest.json"),
		[]byte(`{"version":2,"id":"r9","created_at":"2026-10-18T13:00:00Z","files":{"x":"h"},"file_count":1,"new":["x"]}`), 0644)
	_ = os.RemoveAll(Dir(dir, "prod", "r3"))
	list, _ = Summaries(dir, "prod")
	if len(list) != 2 || list[0].ID != "r9" || list[0].New != 1 || list[1].ID != "r2" {
		t.Fatalf("onarım: %+v", list)
	}
}

func TestPruneUsesSummaries(t *testing.T) {
	dir := t.TempDir()
	base := time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)
	for i := 0; i < 4; i++ {
		id := string(rune('a' + i))
		_ = os.MkdirAll(FilesDir(dir, "prod", id), 0755)
		m := &Manifest{CreatedAt: base.Add(time.Duration(i) * time.Minute), Files: map[string]string{"f": "h"}, Backed: []string{"f"}}
		if _, err := Write(dir, "prod", id, m); err != nil {
			t.Fatal(err)
		}
	}
	if n := Prune(dir, "prod", 2); n != 2 {
		t.Fatalf("budanan %d, 2 bekleniyordu", n)
	}
	list, _ := Summaries(dir, "prod")
	restorable := 0
	for _, s := range list {
		if s.Restorable() {
			restorable++
		}
	}
	if restorable != 2 {
		t.Fatalf("geri alınabilir %d: %+v", restorable, list)
	}
	if _, err := os.Stat(FilesDir(dir, "prod", "a")); !os.IsNotExist(err) {
		t.Fatal("en eski yedek silinmeli")
	}
}

package release

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWriteListPrune(t *testing.T) {
	dir := t.TempDir()
	base := time.Now().Add(-time.Hour)
	for i := 0; i < 4; i++ {
		id := NewID(dir, "prod")
		os.MkdirAll(FilesDir(dir, "prod", id), 0755)
		os.WriteFile(filepath.Join(FilesDir(dir, "prod", id), "a.txt"), []byte("old"), 0644)
		m := &Manifest{CreatedAt: base.Add(time.Duration(i) * time.Minute), Files: map[string]string{"a.txt": "h"}, Backed: []string{"a.txt"}}
		if _, err := Write(dir, "prod", id, m); err != nil {
			t.Fatal(err)
		}
	}
	list, _ := List(dir, "prod")
	if len(list) != 4 {
		t.Fatalf("list = %d", len(list))
	}
	if !list[0].CreatedAt.After(list[3].CreatedAt) {
		t.Fatal("list must be newest first")
	}
	// Aynı saniyede üretilen ID'ler çakışmaz
	ids := map[string]bool{}
	for _, m := range list {
		if ids[m.ID] {
			t.Fatalf("duplicate id %s", m.ID)
		}
		ids[m.ID] = true
	}

	if n := Prune(dir, "prod", 2); n != 2 {
		t.Fatalf("pruned = %d", n)
	}
	list, _ = List(dir, "prod")
	for i, m := range list {
		want := i < 2
		if m.HasBackup() != want {
			t.Fatalf("release %d HasBackup = %v", i, m.HasBackup())
		}
		_, err := os.Stat(FilesDir(dir, "prod", m.ID))
		if want == os.IsNotExist(err) {
			t.Fatalf("release %d files dir state wrong: %v", i, err)
		}
	}
	// Yedeksiz (v1) manifest rollback edilemez
	v1 := &Manifest{Version: 1, Files: map[string]string{"x": "h"}}
	if v1.HasBackup() {
		t.Fatal("v1 must not be restorable")
	}
}

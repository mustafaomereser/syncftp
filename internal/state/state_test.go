package state

import (
	"os"
	"testing"
)

func TestDiff_NewFile(t *testing.T) {
	s := &State{Files: map[string]string{}}
	current := map[string]string{"index.php": "sha256:abc"}
	diff := Diff(s, current)

	if len(diff.New) != 1 || diff.New[0] != "index.php" {
		t.Errorf("beklenen yeni dosya: index.php, alınan: %v", diff.New)
	}
	if len(diff.Changed) != 0 || len(diff.Deleted) != 0 {
		t.Errorf("beklenmeyen değişiklik veya silme: %+v", diff)
	}
}

func TestDiff_ChangedFile(t *testing.T) {
	s := &State{Files: map[string]string{"index.php": "sha256:old"}}
	current := map[string]string{"index.php": "sha256:new"}
	diff := Diff(s, current)

	if len(diff.Changed) != 1 || diff.Changed[0] != "index.php" {
		t.Errorf("beklenen değişen dosya: index.php, alınan: %v", diff.Changed)
	}
	if len(diff.New) != 0 || len(diff.Deleted) != 0 {
		t.Errorf("beklenmeyen yeni veya silinme: %+v", diff)
	}
}

func TestDiff_DeletedFile(t *testing.T) {
	s := &State{Files: map[string]string{"old.php": "sha256:abc"}}
	current := map[string]string{}
	diff := Diff(s, current)

	if len(diff.Deleted) != 1 || diff.Deleted[0] != "old.php" {
		t.Errorf("beklenen silinen dosya: old.php, alınan: %v", diff.Deleted)
	}
}

func TestDiff_NoChange(t *testing.T) {
	s := &State{Files: map[string]string{"index.php": "sha256:same"}}
	current := map[string]string{"index.php": "sha256:same"}
	diff := Diff(s, current)

	if len(diff.New)+len(diff.Changed)+len(diff.Deleted) != 0 {
		t.Errorf("değişiklik olmamalıydı ama %+v alındı", diff)
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()

	s := &State{
		Version:       1,
		Server:        "test-server",
		FirstSyncDone: true,
		Files:         map[string]string{"app.js": "sha256:abc123"},
	}

	if err := Save(dir, s); err != nil {
		t.Fatalf("Save başarısız: %v", err)
	}

	loaded, err := Load(dir, "test-server")
	if err != nil {
		t.Fatalf("Load başarısız: %v", err)
	}

	if !loaded.FirstSyncDone {
		t.Error("FirstSyncDone false olmamalıydı")
	}
	if hash, ok := loaded.Files["app.js"]; !ok || hash != "sha256:abc123" {
		t.Errorf("beklenen hash sha256:abc123, alınan: %q", hash)
	}

	// Cleanup is handled by t.TempDir()
	_ = os.RemoveAll(dir)
}

func TestLoad_EmptyState(t *testing.T) {
	dir := t.TempDir()
	s, err := Load(dir, "nonexistent")
	if err != nil {
		t.Fatalf("Load başarısız: %v", err)
	}
	if s.FirstSyncDone {
		t.Error("yeni state'de FirstSyncDone true olmamalı")
	}
	if s.Files == nil {
		t.Error("Files nil olmamalı")
	}
}

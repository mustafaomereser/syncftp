package projects

import (
	"os"
	"path/filepath"
	"testing"
)

func mkProject(t *testing.T, name string) string {
	t.Helper()
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".syncftp"), 0700)
	os.WriteFile(filepath.Join(dir, ".syncftp", "syncftp.json"), []byte(`{"project":{"name":"`+name+`"}}`), 0600)
	return dir
}

func TestRegistry(t *testing.T) {
	pathOverride = filepath.Join(t.TempDir(), "projects.json")
	t.Cleanup(func() { pathOverride = "" })

	a := mkProject(t, "magaza")
	b := mkProject(t, "magaza") // aynı ad → -2

	pa, err := Add(a, "")
	if err != nil || pa.Name != "magaza" {
		t.Fatalf("add a: %+v %v", pa, err)
	}
	pb, _ := Add(b, "")
	if pb.Name != "magaza-2" {
		t.Fatalf("dup name = %s", pb.Name)
	}
	// Aynı yolu tekrar eklemek kopya üretmez
	Add(a, "")
	if n := len(List()); n != 2 {
		t.Fatalf("list = %d", n)
	}
	if _, err := Add(t.TempDir(), "x"); err == nil {
		t.Fatal("dir without config must be rejected")
	}
	if p, ok := Find("MAGAZA"); !ok || p.Path != pa.Path {
		t.Fatal("find by name (case-insensitive)")
	}
	if _, ok := Find(b); !ok {
		t.Fatal("find by path")
	}
	if err := Rename("magaza-2", "blog"); err != nil {
		t.Fatal(err)
	}
	if _, ok := Find("blog"); !ok {
		t.Fatal("rename")
	}
	if err := Remove("blog"); err != nil {
		t.Fatal(err)
	}
	// Touch kayıtsız projeyi ekler
	c := mkProject(t, "yeni")
	Touch(c)
	if p, ok := Find("yeni"); !ok || p.LastUsed.IsZero() {
		t.Fatalf("touch: %+v %v", p, ok)
	}
}

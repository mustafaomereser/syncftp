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

// Klasör modunda tarama 6 seviyeye iner; daha derindeki proje panelden eklenince kayda yazılır ve
// listede kalır (eskiden 2 seviyeydi, "proje bulunamadı (çok derinde…)" hatası veriyordu).
func TestBaseDeepProject(t *testing.T) {
	pathOverride = filepath.Join(t.TempDir(), "projects.json")
	root := t.TempDir()
	SetBase(root)
	t.Cleanup(func() { pathOverride = ""; SetBase("") })
	mid := filepath.Join(root, "a", "b", "dist") // 3. seviye → taramayla bulunur
	if err := os.MkdirAll(filepath.Join(mid, ".syncftp"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mid, ".syncftp", "syncftp.json"), []byte(`{"project":{"name":"orta"}}`), 0644); err != nil {
		t.Fatal(err)
	}
	if _, ok := Find("orta"); !ok {
		t.Fatal("3. seviyedeki proje taramayla bulunmalı")
	}
	deep := filepath.Join(root, "a", "b", "c", "d", "e", "f", "dist") // 7. seviye
	Invalidate()
	if err := os.MkdirAll(filepath.Join(deep, ".syncftp"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(deep, ".syncftp", "syncftp.json"), []byte(`{"project":{"name":"derin"}}`), 0644); err != nil {
		t.Fatal(err)
	}
	if _, ok := Find(deep); ok {
		t.Fatal("tarama 7. seviyeyi görmemeli (test varsayımı)")
	}
	p, err := Add(deep, "")
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "derin" {
		t.Fatalf("ad = %q", p.Name)
	}
	Invalidate()
	if _, ok := Find("derin"); !ok {
		t.Fatal("eklenen derin proje listede kalmalı")
	}
	Forget(deep)
	if _, ok := Find("derin"); ok {
		t.Fatal("Forget sonrası listeden düşmeli")
	}
}

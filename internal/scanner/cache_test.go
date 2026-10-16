package scanner

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"syncftp/internal/ignore"
)

func TestScanSkipsIgnoredDirsAndUsesCache(t *testing.T) {
	cacheDirOverride = t.TempDir()
	t.Cleanup(func() { cacheDirOverride = "" })

	dir := t.TempDir()
	write := func(rel, content string) {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		os.MkdirAll(filepath.Dir(p), 0755)
		os.WriteFile(p, []byte(content), 0644)
		old := time.Now().Add(-time.Hour) // "son 2 sn" kuralına takılmasın → cache'lenir
		os.Chtimes(p, old, old)
	}
	write(".gitignore", "vendor/\n/storage/logs/\n")
	write("index.php", "a")
	write("vendor/x/lib.php", "v")
	write("storage/logs/app.log", "l")

	m, _ := ignore.Load(dir, []string{".gitignore"})
	files, err := Scan(dir, m)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, f := range files {
		got[f.RelPath] = f.Hash
	}
	if _, ok := got["vendor/x/lib.php"]; ok {
		t.Fatal("vendor/ should be skipped")
	}
	if _, ok := got["storage/logs/app.log"]; ok {
		t.Fatal("storage/logs/ should be skipped")
	}
	if got["index.php"] == "" {
		t.Fatal("index.php missing")
	}

	// Cache yazıldı
	if _, err := os.Stat(cachePath(dir)); err != nil {
		t.Fatalf("cache not written: %v", err)
	}
	// Cache'teki hash'i boz: boyut+mtime aynıysa dosya yeniden okunmamalı → bozuk hash döner
	c := loadCache(dir)
	e := c.Files["index.php"]
	e.Hash = "sha256:cached"
	c.Files["index.php"] = e
	saveCache(dir, c)
	files, _ = Scan(dir, m)
	for _, f := range files {
		if f.RelPath == "index.php" && f.Hash != "sha256:cached" {
			t.Fatalf("expected cache hit, got %s", f.Hash)
		}
	}

	// İçerik + mtime değişince yeniden hash'lenir
	write("index.php", "changed")
	files, _ = Scan(dir, m)
	for _, f := range files {
		if f.RelPath == "index.php" && f.Hash == "sha256:cached" {
			t.Fatal("changed file must be rehashed")
		}
	}
}

func TestMatchDirTrailingSlashPattern(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("vendor/\nnode_modules\n"), 0644)
	m, _ := ignore.Load(dir, []string{".gitignore"})
	if m.Match("vendor") {
		t.Log("Match already handles dirs (library changed)")
	}
	for _, d := range []string{"vendor", "node_modules", "a/node_modules"} {
		if !m.MatchDir(d) {
			t.Errorf("MatchDir(%q) = false", d)
		}
	}
	if m.MatchDir("src") {
		t.Error("src must not match")
	}
}

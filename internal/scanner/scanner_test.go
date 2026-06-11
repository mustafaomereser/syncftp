package scanner

import (
	"os"
	"path/filepath"
	"testing"

	"syncftp/internal/ignore"
)

func TestScan_BasicFiles(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, dir, "index.php", "<?php echo 'hello';")
	writeFile(t, dir, "css/style.css", "body { margin: 0; }")
	writeFile(t, dir, "syncftp.config", "password = secret") // her zaman atlanmalı

	matcher := &ignore.Matcher{} // no-op matcher
	files, err := Scan(dir, matcher)
	if err != nil {
		t.Fatalf("Scan başarısız: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("2 dosya beklendi, %d alındı", len(files))
	}

	relPaths := make(map[string]bool)
	for _, f := range files {
		relPaths[f.RelPath] = true
		if f.Hash == "" || len(f.Hash) < 7 {
			t.Errorf("geçersiz hash: %q", f.Hash)
		}
	}

	if !relPaths["index.php"] {
		t.Error("index.php taranmış olmalıydı")
	}
	if !relPaths["css/style.css"] {
		t.Error("css/style.css taranmış olmalıydı")
	}
	if relPaths["syncftp.config"] {
		t.Error("syncftp.config taranmamalıydı")
	}
}

func TestScan_SkipsSyncftpDir(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "app.php", "<?php")
	writeFile(t, dir, ".syncftp/state/prod.json", "{}")

	matcher := &ignore.Matcher{}
	files, err := Scan(dir, matcher)
	if err != nil {
		t.Fatalf("Scan başarısız: %v", err)
	}

	for _, f := range files {
		if f.RelPath == ".syncftp/state/prod.json" {
			t.Error(".syncftp dizini taranmamalıydı")
		}
	}
}

func TestScan_HashConsistency(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "file.txt", "içerik")

	matcher := &ignore.Matcher{}
	files1, _ := Scan(dir, matcher)
	files2, _ := Scan(dir, matcher)

	if len(files1) == 0 || len(files2) == 0 {
		t.Fatal("tarama sonucu boş")
	}
	if files1[0].Hash != files2[0].Hash {
		t.Error("aynı dosyanın hash'i farklı olamaz")
	}
}

func writeFile(t *testing.T, dir, relPath, content string) {
	t.Helper()
	full := filepath.Join(dir, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

package fsutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSafeName(t *testing.T) {
	cases := map[string]string{
		"production":        "production",
		"Prod: public_html": "Prod_ public_html",
		`a/b\c*?"<>|`:       "a_b_c______",
		"trailing. ":        "trailing",
		"":                  "_",
		"con":               "con_",
		"LPT1.backup":       "LPT1.backup_",
		"çalışma":           "çalışma",
	}
	for in, want := range cases {
		if got := SafeName(in); got != want {
			t.Errorf("SafeName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestServerFileMigratesLegacyName(t *testing.T) {
	dir := t.TempDir()
	// Güvenli adlı sunucu: yol aynen döner
	if got := ServerFile(dir, "prod", ".json"); got != filepath.Join(dir, "prod.json") {
		t.Fatalf("plain name: %s", got)
	}
	// Eski (geçersiz karakter içermeyen ama değişen) ad: içerik yeni dosyaya taşınır
	legacy := filepath.Join(dir, "trailing..json")
	if err := os.WriteFile(legacy, []byte(`{"a":1}`), 0644); err != nil {
		t.Fatal(err)
	}
	p := ServerFile(dir, "trailing.", ".json")
	if p != filepath.Join(dir, "trailing.json") {
		t.Fatalf("safe path: %s", p)
	}
	data, err := os.ReadFile(p)
	if err != nil || string(data) != `{"a":1}` {
		t.Fatalf("migrated content: %q %v", data, err)
	}
	// Yeni dosya zaten varsa eskisi tekrar kopyalanmaz
	_ = os.WriteFile(p, []byte("new"), 0644)
	ServerFile(dir, "trailing.", ".json")
	if data, _ := os.ReadFile(p); string(data) != "new" {
		t.Fatalf("overwrote existing file: %q", data)
	}
}

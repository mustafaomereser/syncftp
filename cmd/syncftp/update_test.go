package main

import (
	"archive/zip"
	"crypto/ed25519"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

func TestVersionNewer(t *testing.T) {
	for _, c := range []struct {
		a, b string
		want bool
	}{{"2.0.1", "2.0.0", true}, {"2.0.10", "2.0.9", true}, {"2.0.0", "2.0.0", false}, {"1.9.9", "2.0.0", false}, {"2.1", "2.0.9", true}} {
		if got := versionNewer(c.a, c.b); got != c.want {
			t.Errorf("%s > %s = %v", c.a, c.b, got)
		}
	}
}

func TestVerifyManifest(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	pubB := base64.StdEncoding.EncodeToString(pub)
	m := []byte("abc123  *syncFTP-2.0.1-linux-amd64.zip\ndef  plain.zip\n")
	sig := []byte(base64.StdEncoding.EncodeToString(ed25519.Sign(priv, m)))
	got, err := verifyManifest(pubB, m, sig)
	if err != nil || got["syncFTP-2.0.1-linux-amd64.zip"] != "abc123" || got["plain.zip"] != "def" {
		t.Fatalf("%v %v", got, err)
	}
	if _, err := verifyManifest(pubB, append(m, 'x'), sig); err == nil {
		t.Fatal("değiştirilmiş liste kabul edildi")
	}
	other, _, _ := ed25519.GenerateKey(nil)
	if _, err := verifyManifest(base64.StdEncoding.EncodeToString(other), m, sig); err == nil {
		t.Fatal("yabancı anahtar kabul edildi")
	}
}

func writeZip(t *testing.T, p string, files map[string]string) {
	t.Helper()
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	z := zip.NewWriter(f)
	for n, c := range files {
		h := &zip.FileHeader{Name: n, Method: zip.Deflate}
		h.SetMode(0755)
		w, _ := z.CreateHeader(h)
		w.Write([]byte(c))
	}
	z.Close()
	f.Close()
}

func TestInstallPackage(t *testing.T) {
	d := t.TempDir()
	read := func(p string) string { b, _ := os.ReadFile(p); return string(b) }

	// Windows: exe doğrudan değişir, eskisi temizlenmek üzere kalır
	exe := filepath.Join(d, "a.exe")
	os.WriteFile(exe, []byte("old"), 0755)
	pkg := filepath.Join(d, "new.exe")
	os.WriteFile(pkg, []byte("new"), 0755)
	if err := installPackage("windows", pkg, exe); err != nil || read(exe) != "new" || read(exe+updateOldSuffix) != "old" {
		t.Fatalf("windows: %v", err)
	}

	// Linux: zip'ten yalnızca program; zip-slip dışarı yazamaz
	bin := filepath.Join(d, "syncftp")
	os.WriteFile(bin, []byte("old"), 0755)
	z := filepath.Join(d, "l.zip")
	writeZip(t, z, map[string]string{"syncftp": "newbin", "../../evil": "x", "install.sh": "y"})
	if err := installPackage("linux", z, bin); err != nil || read(bin) != "newbin" {
		t.Fatalf("linux: %v", err)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(d), "evil")); err == nil {
		t.Fatal("zip-slip")
	}

	// macOS: .app bütün olarak değişir
	mac := filepath.Join(d, "Apps", "syncFTP.app", "Contents", "MacOS")
	os.MkdirAll(mac, 0755)
	os.WriteFile(filepath.Join(mac, "syncftp"), []byte("old"), 0755)
	os.WriteFile(filepath.Join(mac, "stale"), []byte("x"), 0644)
	z2 := filepath.Join(d, "m.zip")
	writeZip(t, z2, map[string]string{"syncFTP.app/Contents/MacOS/syncftp": "newmac", "syncFTP.app/Contents/Info.plist": "p"})
	if err := installPackage("darwin", z2, filepath.Join(mac, "syncftp")); err != nil || read(filepath.Join(mac, "syncftp")) != "newmac" {
		t.Fatalf("mac: %v", err)
	}
	if _, err := os.Stat(filepath.Join(mac, "stale")); err == nil {
		t.Fatal("eski paket dosyası kaldı")
	}

	// Geçersiz paket: kurulu dosya değişmez
	bin2 := filepath.Join(d, "b", "syncftp")
	os.MkdirAll(filepath.Dir(bin2), 0755)
	os.WriteFile(bin2, []byte("keep"), 0755)
	z3 := filepath.Join(d, "bad.zip")
	writeZip(t, z3, map[string]string{"other": "x"})
	if err := installPackage("linux", z3, bin2); err == nil || read(bin2) != "keep" {
		t.Fatalf("bozuk paket: %v %q", err, read(bin2))
	}
}

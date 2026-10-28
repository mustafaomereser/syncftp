package ignore

import (
	"os"
	"path/filepath"
	"testing"
)

// Sync'in okuduğu kök ignore dosyaları yüklenmez; alt klasörlerdeki .gitignore'lar (boş klasörleri
// sunucuda var eden) yüklenir; "!/.gitignore" ile kullanıcı geri açabilir.
func TestIgnoreFilesThemselvesIgnored(t *testing.T) {
	dir := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(".gitignore", "*.log\n")
	write(".syncftp/syncftp.ignore", "")

	m, err := Load(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]bool{
		".gitignore":                         true,
		".syncftp/syncftp.ignore":            true,
		"storage/framework/cache/.gitignore": false,
		"app.log":                            true,
		"index.php":                          false,
	}
	for p, want := range cases {
		if got := m.Match(p); got != want {
			t.Errorf("Match(%q) = %v, want %v", p, got, want)
		}
	}

	// Kullanıcı açıkça isterse kök .gitignore yine yüklenir
	write(".syncftp/syncftp.ignore", "!/.gitignore\n")
	m, _ = Load(dir, nil)
	if m.Match(".gitignore") {
		t.Error(`"!/.gitignore" should re-include the root .gitignore`)
	}

	// Özel liste: yalnızca listedeki dosyalar elenir
	m, _ = Load(dir, []string{"deploy.ignore"})
	if !m.Match("deploy.ignore") || m.Match(".gitignore") {
		t.Errorf("custom list: deploy.ignore=%v .gitignore=%v", m.Match("deploy.ignore"), m.Match(".gitignore"))
	}
}

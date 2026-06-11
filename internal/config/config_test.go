package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_Basic(t *testing.T) {
	dir := t.TempDir()

	toml := `
[project]
name = "test-proje"
local_path = "."

[sync]
protect = [".env"]
include = []

[first_sync]
full = true

[[servers]]
name = "prod"
host = "ftp.example.com"
port = 21
user = "user"
remote_path = "/public_html"
passive = true
enabled = true
`
	if err := os.WriteFile(filepath.Join(dir, "syncftp.toml"), []byte(toml), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load başarısız: %v", err)
	}

	if cfg.Project.Name != "test-proje" {
		t.Errorf("proje adı yanlış: %q", cfg.Project.Name)
	}
	if len(cfg.Servers) != 1 {
		t.Fatalf("1 sunucu beklendi, %d alındı", len(cfg.Servers))
	}
	if cfg.Servers[0].Host != "ftp.example.com" {
		t.Errorf("host yanlış: %q", cfg.Servers[0].Host)
	}
}

func TestLoad_MergesPassword(t *testing.T) {
	dir := t.TempDir()

	toml := `
[project]
name = "test"
local_path = "."
[first_sync]
full = true
[[servers]]
name = "prod"
host = "ftp.example.com"
port = 21
user = "user"
remote_path = "/"
passive = true
enabled = true
`
	secret := `
[[servers]]
name = "prod"
password = "gizli123"
`
	os.WriteFile(filepath.Join(dir, "syncftp.toml"), []byte(toml), 0644)
	os.WriteFile(filepath.Join(dir, "syncftp.config"), []byte(secret), 0600)

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load başarısız: %v", err)
	}

	if cfg.Servers[0].Password != "gizli123" {
		t.Errorf("şifre merge edilmedi, alınan: %q", cfg.Servers[0].Password)
	}
}

func TestLoad_DefaultPort(t *testing.T) {
	dir := t.TempDir()

	toml := `
[project]
name = "test"
local_path = "."
[first_sync]
full = true
[[servers]]
name = "prod"
host = "ftp.example.com"
user = "user"
remote_path = "/"
enabled = true
`
	os.WriteFile(filepath.Join(dir, "syncftp.toml"), []byte(toml), 0644)

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load başarısız: %v", err)
	}

	if cfg.Servers[0].Port != 21 {
		t.Errorf("varsayılan port 21 olmalıydı, alınan: %d", cfg.Servers[0].Port)
	}
}

func TestEnabledServers(t *testing.T) {
	cfg := &Config{
		Servers: []Server{
			{Name: "prod", Enabled: true},
			{Name: "dev", Enabled: false},
			{Name: "staging", Enabled: true},
		},
	}
	enabled := cfg.EnabledServers()
	if len(enabled) != 2 {
		t.Errorf("2 aktif sunucu beklendi, %d alındı", len(enabled))
	}
}

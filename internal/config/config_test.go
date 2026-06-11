package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func writeJSON(t *testing.T, dir, name string, v any) {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), data, 0644); err != nil {
		t.Fatal(err)
	}
}

func baseConfig(name string) map[string]any {
	return map[string]any{
		"project":    map[string]any{"name": name, "local_path": "."},
		"sync":       map[string]any{"protect": []string{".env"}, "include": []string{}, "exclude": []string{}},
		"first_sync": map[string]any{"full": true},
		"servers": []map[string]any{
			{
				"name": "prod", "host": "ftp.example.com", "port": 21,
				"user": "user", "password": "gizli123", "remote_path": "/public_html",
				"passive": true, "enabled": true,
				"include": []string{}, "exclude": []string{},
			},
		},
	}
}

func TestLoad_Basic(t *testing.T) {
	dir := t.TempDir()
	writeJSON(t, dir, "syncftp.json", baseConfig("test-proje"))

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
	if cfg.Servers[0].Password != "gizli123" {
		t.Errorf("şifre yanlış: %q", cfg.Servers[0].Password)
	}
}

func TestLoad_DefaultPort(t *testing.T) {
	dir := t.TempDir()
	cfg := baseConfig("test")
	cfg["servers"].([]map[string]any)[0]["port"] = 0
	writeJSON(t, dir, "syncftp.json", cfg)

	loaded, err := Load(dir)
	if err != nil {
		t.Fatalf("Load başarısız: %v", err)
	}
	if loaded.Servers[0].Port != 21 {
		t.Errorf("varsayılan port 21 olmalıydı, alınan: %d", loaded.Servers[0].Port)
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

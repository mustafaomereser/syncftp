package config

import (
	"path/filepath"
	"testing"
)

func TestLibraryRoundtripAndDefaults(t *testing.T) {
	libraryPathOverride = filepath.Join(t.TempDir(), "library.json")
	t.Cleanup(func() { libraryPathOverride = "" })

	lib, err := LoadLibrary()
	if err != nil {
		t.Fatal(err)
	}
	if len(lib.Connections) != 0 || !lib.Defaults.FirstSyncFull || len(lib.Defaults.Sync.Protect) != 1 {
		t.Fatalf("missing file should give builtin defaults, got %+v", lib.Defaults)
	}

	// Şifresiz bağlantı (testte gerçek anahtarlığa yazılmasın)
	lib.Connections = append(lib.Connections, Connection{Name: "Host", Host: "ftp.example.com", User: "u", Protocol: "sftp"})
	lib.Defaults.IgnoreTemplate = "laravel"
	lib.Defaults.Sync.Exclude = []string{"node_modules/"}
	if err := lib.ValidateConnections(); err != nil {
		t.Fatal(err)
	}
	if err := SaveLibrary(lib); err != nil {
		t.Fatal(err)
	}
	got, err := LoadLibrary()
	if err != nil {
		t.Fatal(err)
	}
	c, ok := got.FindConnection("host")
	if !ok || c.Port != 22 {
		t.Fatalf("FindConnection case-insensitive + sftp default port: %+v %v", c, ok)
	}
	if got.Defaults.IgnoreTemplate != "laravel" {
		t.Fatalf("template lost: %+v", got.Defaults)
	}

	// Defaults yeni config'e derin kopya olarak uygulanır
	cfg := &Config{}
	got.Defaults.ApplyDefaults(cfg)
	cfg.Sync.Exclude[0] = "changed"
	if got.Defaults.Sync.Exclude[0] != "node_modules/" {
		t.Fatal("ApplyDefaults must not share slices with the library")
	}

	// Sunucuya bağlantı uygulanması
	srv := got.Defaults.NewServer("prod")
	srv.ApplyConnection(c)
	if srv.Host != "ftp.example.com" || !srv.IsSFTP() || srv.Port != 22 || srv.MaxConnections != 1 {
		t.Fatalf("ApplyConnection: %+v", srv)
	}

	// Yinelenen ad reddedilir
	got.Connections = append(got.Connections, Connection{Name: "HOST", Host: "x"})
	if got.ValidateConnections() == nil {
		t.Fatal("duplicate names must be rejected")
	}
}

func TestRedactedMasksAllSecrets(t *testing.T) {
	c := &Config{
		Servers:     []Server{{Name: "s", Password: "p"}},
		Connections: []Connection{{Name: "c", Password: "p2"}},
		Deploy:      []Deploy{{Name: "d", Secret: "hook"}},
	}
	c.Servers[0].TunnelPassword = "tp"
	r := c.Redacted(MaskStars, true)
	if r.Servers[0].Password != "***" || r.Servers[0].TunnelPassword != "***" || r.Connections[0].Password != "***" || r.Deploy[0].Secret != "***" {
		t.Fatalf("not masked: %+v", r)
	}
	if c.Servers[0].Password != "p" || c.Deploy[0].Secret != "hook" {
		t.Fatal("original changed")
	}
	if k := c.Redacted(MaskEmpty, false); k.Deploy[0].Secret != "hook" || k.Servers[0].Password != "" {
		t.Fatalf("deploySecrets=false: %+v", k)
	}
}

func TestExtractConnection(t *testing.T) {
	cfg := &Config{Servers: []Server{
		{Name: "prod", Host: "h", Port: 21, User: "u", Password: "p", RemotePath: "/a"},
		{Name: "stage", Host: "h", Port: 21, User: "u", Password: "p", RemotePath: "/b"},
		{Name: "other", Host: "x", User: "u", Password: "p"},
	}, Connections: []Connection{{Name: "other", Host: "y"}}}
	name, created, err := cfg.ExtractConnection("prod", "")
	if err != nil || !created || name != "prod" || cfg.Servers[0].Connection != "prod" {
		t.Fatalf("prod: %q %v %v", name, created, err)
	}
	// Aynı hesap → mevcut profile bağlanır
	if name, created, _ = cfg.ExtractConnection("stage", ""); created || name != "prod" {
		t.Fatalf("stage: %q %v", name, created)
	}
	// Ad çakışması → -2
	if name, _, _ = cfg.ExtractConnection("other", ""); name != "other-2" {
		t.Fatalf("other: %q", name)
	}
	if _, _, err = cfg.ExtractConnection("prod", ""); err == nil {
		t.Fatal("zaten profilli sunucu hata vermeli")
	}
	if len(cfg.Connections) != 3 {
		t.Fatalf("profil sayısı %d", len(cfg.Connections))
	}
}

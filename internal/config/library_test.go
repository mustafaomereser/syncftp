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

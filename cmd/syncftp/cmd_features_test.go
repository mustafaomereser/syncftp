package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"syncftp/internal/config"
	"syncftp/internal/ignore"
	"syncftp/internal/release"
	"syncftp/internal/state"
)

func TestRollbackPlan(t *testing.T) {
	now := time.Now()
	mk := func(id string, ago int, backed bool, rolled bool) *release.Manifest {
		m := &release.Manifest{Version: 2, ID: id, CreatedAt: now.Add(-time.Duration(ago) * time.Minute), RolledBack: rolled}
		if backed {
			m.Backed = []string{"a"}
		} else {
			m.Version = 1
		}
		return m
	}
	list := []*release.Manifest{mk("r4", 1, true, false), mk("r3", 2, true, true), mk("r2", 3, true, false), mk("r1", 4, false, false)}

	plan, err := rollbackPlan(list, "r2")
	if err != nil {
		t.Fatal(err)
	}
	if len(plan) != 2 || plan[0].ID != "r4" || plan[1].ID != "r2" {
		t.Fatalf("plan = %v", ids(plan))
	}
	if _, err := rollbackPlan(list, "r1"); err == nil {
		t.Fatal("release without backup must be refused")
	}
	if _, err := rollbackPlan(list, "r3"); err == nil {
		t.Fatal("already rolled back must be refused")
	}
	if _, err := rollbackPlan(list, "nope"); err == nil {
		t.Fatal("unknown id must be refused")
	}
}

func ids(ms []*release.Manifest) []string {
	var out []string
	for _, m := range ms {
		out = append(out, m.ID)
	}
	return out
}

func TestRemoteSkipFunc(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("vendor/\nnode_modules\n"), 0644)
	m, _ := ignore.Load(dir, []string{".gitignore"})

	skip := remoteSkipFunc(m, nil, []string{"uploads"})
	cases := map[string]bool{"vendor": true, "a/node_modules": true, "uploads": true, "uploads/x": true, "app": false, "uploads2": false}
	for rel, want := range cases {
		if got := skip(rel); got != want {
			t.Errorf("skip(%q) = %v, want %v", rel, got, want)
		}
	}
	// Include varken: include'u kapsayan ya da altındaki dizinlere girilir, diğerlerine girilmez
	skip = remoteSkipFunc(nil, []string{"app/Http", "public/"}, nil)
	cases = map[string]bool{"app": false, "app/Http": false, "app/Http/Controllers": false, "app/Models": true, "public": false, "resources": true}
	for rel, want := range cases {
		if got := skip(rel); got != want {
			t.Errorf("include skip(%q) = %v, want %v", rel, got, want)
		}
	}
}

func TestNextProtocol(t *testing.T) {
	p, port := nextProtocol("", 21)
	if p != "ftps" || port != 21 {
		t.Fatalf("ftp→%s:%d", p, port)
	}
	p, port = nextProtocol("ftps", 21)
	if p != "sftp" || port != 22 {
		t.Fatalf("ftps→%s:%d", p, port)
	}
	p, port = nextProtocol("sftp", 22)
	if p != "" || port != 21 {
		t.Fatalf("sftp→%q:%d", p, port)
	}
	if _, port = nextProtocol("ftps", 2121); port != 2121 {
		t.Fatal("custom port must be kept")
	}
}

func TestRestoreSecretsAndValidate(t *testing.T) {
	old := &config.Config{
		Servers:     []config.Server{{Name: "prod", Password: "p1", SSHTunnel: config.SSHTunnel{TunnelPassword: "t1"}}},
		Connections: []config.Connection{{Name: "host", Password: "c1"}},
	}
	in := &config.Config{
		Servers: []config.Server{
			{Name: "prod", Host: "h", RemotePath: "/", Password: keepSecret, SSHTunnel: config.SSHTunnel{TunnelPassword: keepSecret}},
			{Name: "new", Connection: "host", RemotePath: "/x", Password: "fresh"},
		},
		Connections: []config.Connection{{Name: "host", Password: keepSecret}},
	}
	restoreSecrets(in, old)
	if in.Servers[0].Password != "p1" || in.Servers[0].TunnelPassword != "t1" {
		t.Fatalf("server secrets not restored: %+v", in.Servers[0])
	}
	if in.Servers[1].Password != "fresh" || in.Connections[0].Password != "c1" {
		t.Fatal("fresh password overwritten or connection not restored")
	}
	if err := validateConfig(in); err != nil {
		t.Fatalf("valid config rejected: %v", err)
	}
	in.Servers[1].Connection = "missing"
	if err := validateConfig(in); err == nil {
		t.Fatal("missing connection must be rejected")
	}
	in.Servers[1].Connection = "host"
	in.Servers = append(in.Servers, config.Server{Name: "prod", Host: "h", RemotePath: "/"})
	if err := validateConfig(in); err == nil {
		t.Fatal("duplicate name must be rejected")
	}
}

func TestJobWriterLines(t *testing.T) {
	j := &job{subs: map[chan string]struct{}{}, done: make(chan struct{})}
	j.Write([]byte("bir\r\niki"))
	j.Write([]byte(" devam\n\r  3 bağlantı\r  5 bağlantı\n\x1b[32mrenk\x1b[0m\n"))
	j.finish(nil, nil)
	want := []string{"bir", "iki devam", "  5 bağlantı", "renk"}
	if strings.Join(j.lines, "|") != strings.Join(want, "|") {
		t.Fatalf("lines = %q", j.lines)
	}
	if j.Status != "done" {
		t.Fatal(j.Status)
	}
}

func TestPendingForFilters(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{
		Project: config.Project{DefaultPath: "."},
		Sync:    config.Sync{Exclude: []string{"docs"}, BlockExtensions: []string{"zip"}},
	}
	srv := config.Server{Name: "s"}
	current := map[string]string{"a.php": "1", "docs/x.md": "1", "b.zip": "1", "c.php": "2"}
	st := &state.State{Files: map[string]string{"c.php": "1", "gone.php": "1"}}
	n, c, d := pendingFor(dir, cfg, srv, current, st)
	if strings.Join(n, ",") != "a.php" || strings.Join(c, ",") != "c.php" || strings.Join(d, ",") != "gone.php" {
		t.Fatalf("new=%v changed=%v deleted=%v", n, c, d)
	}
}

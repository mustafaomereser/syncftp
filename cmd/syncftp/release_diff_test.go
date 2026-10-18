package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"syncftp/internal/config"
	"syncftp/internal/release"
)

func TestDiffLineOps(t *testing.T) {
	ops, ok := diffLineOps(strings.Split("a\nb\nc\nd", "\n"), strings.Split("a\nx\nc\nd\ne", "\n"))
	if !ok {
		t.Fatal("diff yok")
	}
	var got []string
	for _, o := range ops {
		got = append(got, string(o.op)+o.text)
	}
	if want := " a,-b,+x, c, d,+e"; strings.Join(got, ",") != want {
		t.Fatalf("got %q want %q", strings.Join(got, ","), want)
	}
}

func TestReleaseFileVersions(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{}
	srv := config.Server{Name: "prod"}
	write := func(p, s string) {
		_ = os.MkdirAll(filepath.Dir(p), 0755)
		if err := os.WriteFile(p, []byte(s), 0644); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(release.FilesDir(dir, "prod", "r1"), "a.txt"), "eski\n")
	write(filepath.Join(release.AfterDir(dir, "prod", "r1"), "a.txt"), "yeni\r\n")
	write(filepath.Join(dir, "b.txt"), "yerel\n") // after/ yok → hash tutarsa yerelden
	m := &release.Manifest{ID: "r1", Files: map[string]string{
		"a.txt": "x", "b.txt": contentHash([]byte("yerel\n")), "c.txt": "sha256:baska",
	}, Backed: []string{"a.txt"}, New: []string{"b.txt", "c.txt"}}
	write(filepath.Join(dir, "c.txt"), "sonradan değişti\n")

	v, err := releaseFileVersions(dir, cfg, srv, m, "a.txt")
	if err != nil || v.Before != "eski\n" || v.After != "yeni\n" || v.AfterSource != "release" || v.Status != "changed" {
		t.Fatalf("a: %+v %v", v, err)
	}
	if v, _ = releaseFileVersions(dir, cfg, srv, m, "b.txt"); !v.AfterOK || v.AfterSource != "local" || v.Status != "new" {
		t.Fatalf("b: %+v", v)
	}
	if v, _ = releaseFileVersions(dir, cfg, srv, m, "c.txt"); v.AfterOK {
		t.Fatalf("c: yerel değişmişken yeni hal gösterilmemeli: %+v", v)
	}
	if _, err = releaseFileVersions(dir, cfg, srv, m, "../x"); err == nil {
		t.Fatal("manifestte olmayan yol hata vermeli")
	}
}

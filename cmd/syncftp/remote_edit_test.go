package main

import (
	"strings"
	"testing"

	"syncftp/internal/config"
)

func TestOpenEditDoc(t *testing.T) {
	d := openEditDoc([]byte("a\r\nb\n"), []byte("a\nb\n"), true)
	if !d.Editable || d.Content != "a\nb\n" || !d.LocalSame {
		t.Fatalf("text: %+v", d)
	}
	if d := openEditDoc([]byte{'a', 0, 'b'}, nil, false); d.Editable || d.Reason != "binary" {
		t.Fatalf("binary: %+v", d)
	}
	if d := openEditDoc([]byte{'a', 0xff, 'b'}, nil, false); d.Editable || d.Reason != "encoding" {
		t.Fatalf("encoding: %+v", d)
	}
	if d := openEditDoc([]byte(strings.Repeat("x", editMaxBytes+1)), nil, false); d.Editable || d.Reason != "large" {
		t.Fatalf("large: %+v", d.Reason)
	}
	if d := openEditDoc([]byte("x"), []byte("y"), true); d.LocalSame {
		t.Fatal("different local reported same")
	}
}

func TestEditPaths(t *testing.T) {
	if got := editRel("/../src/a.php"); got != "src/a.php" {
		t.Fatalf("editRel: %q", got)
	}
	cfg := &config.Config{Project: config.Project{DefaultPath: "."}}
	srv := config.Server{LocalPath: "public"}
	p, err := editLocalPath("/proj", cfg, srv, "css/a.css")
	if err != nil || !strings.HasSuffix(strings.ReplaceAll(p, `\`, "/"), "/proj/public/css/a.css") {
		t.Fatalf("local path: %q %v", p, err)
	}
	if _, err := editLocalPath("/proj", cfg, srv, ""); err == nil {
		t.Fatal("root should be rejected")
	}
}

func TestSyncSkipReason(t *testing.T) {
	f := syncFilters{exclude: []string{"database/connections.php", "docs/"}}
	cases := map[string]string{"database/connections.php": "excluded", "docs": "excluded", "docs/a.md": "excluded", "index.php": ""}
	for rel, want := range cases {
		if got := syncSkipReason(rel, rel == "docs", nil, f); got != want {
			t.Errorf("%s: %q, want %q", rel, got, want)
		}
	}
	inc := syncFilters{include: []string{"public/"}}
	if got := syncSkipReason("src/a.php", false, nil, inc); got != "not_included" {
		t.Errorf("include: %q", got)
	}
	if got := syncSkipReason("public", true, nil, inc); got != "" {
		t.Errorf("include dir: %q", got)
	}
}

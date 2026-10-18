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

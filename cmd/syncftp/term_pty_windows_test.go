//go:build windows

package main

import (
	"os"
	"strings"
	"testing"
	"time"
)

// ConPTY katmanı: cmd.exe çıktısı okunabilmeli, girdi yazılabilmeli, çıkış Wait ile görülmeli.
func TestConPTYEchoAndInput(t *testing.T) {
	p, err := startTermPTY(`C:\Windows\System32\cmd.exe`, []string{"/q", "/k", "echo hello-pty"}, os.TempDir(), os.Environ(), 80, 25)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	out := make(chan string, 64)
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := p.Read(buf)
			if n > 0 {
				out <- string(buf[:n])
			}
			if err != nil {
				close(out)
				return
			}
		}
	}()
	var got strings.Builder
	deadline := time.After(10 * time.Second)
	wrote := false
	for {
		select {
		case s, ok := <-out:
			if !ok {
				t.Fatalf("output closed early: %q", got.String())
			}
			got.WriteString(s)
			if strings.Contains(got.String(), "hello-pty") && !wrote {
				wrote = true
				if _, err := p.Write([]byte("echo second-line\r")); err != nil {
					t.Fatal(err)
				}
			}
			if strings.Contains(got.String(), "second-line") && strings.Count(got.String(), "second-line") >= 2 {
				_, _ = p.Write([]byte("exit\r"))
				done := make(chan error, 1)
				go func() { done <- p.Wait() }()
				select {
				case <-done:
					return
				case <-time.After(5 * time.Second):
					t.Fatal("process did not exit")
				}
			}
		case <-deadline:
			t.Fatalf("timeout; output so far: %q", got.String())
		}
	}
}

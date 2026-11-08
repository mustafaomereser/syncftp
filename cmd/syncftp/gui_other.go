//go:build !windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// macOS'ta .app paketinden açıldıysa çift tıktır → panel açılır (terminalden çıplak binary kabuğu başlatır).
// Finder'dan açılan uygulamanın PATH'i çok kısadır: Homebrew araçları (claude, cloudflared…) bulunsun diye eklenir.
func ownsConsole() bool {
	if runtime.GOOS != "darwin" {
		return false
	}
	exe, err := os.Executable()
	if err != nil || !strings.Contains(exe, ".app/Contents/MacOS/") {
		return false
	}
	os.Setenv("PATH", os.Getenv("PATH")+":/opt/homebrew/bin:/usr/local/bin")
	return true
}

func hideConsole() {}

func spawnDetached(args ...string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	c := exec.Command(exe, args...)
	c.Stdin, c.Stdout, c.Stderr = nil, nil, nil
	if err := c.Start(); err != nil {
		return err
	}
	return c.Process.Release()
}

func openAppWindow(url, profileDir string) error {
	// macOS'ta tarayıcılar PATH'te değil, /Applications altında: eskiden hiçbiri bulunamıyor, panel sekmede açılıyordu
	for _, p := range chromiumCandidates() {
		c := exec.Command(p, "--app="+url, "--user-data-dir="+profileDir, "--window-size=1320,860", "--no-first-run", "--no-default-browser-check")
		if c.Start() == nil {
			return c.Process.Release()
		}
	}
	return openDefaultBrowser(url)
}

func chromiumCandidates() []string {
	var out []string
	if runtime.GOOS == "darwin" {
		home, _ := os.UserHomeDir()
		for _, app := range []string{"Google Chrome", "Microsoft Edge", "Brave Browser", "Chromium"} {
			for _, root := range []string{"/Applications", filepath.Join(home, "Applications")} {
				p := filepath.Join(root, app+".app", "Contents", "MacOS", app)
				if _, err := os.Stat(p); err == nil {
					out = append(out, p)
				}
			}
		}
		return out
	}
	for _, b := range []string{"google-chrome", "google-chrome-stable", "chromium", "chromium-browser", "microsoft-edge", "microsoft-edge-stable", "brave-browser"} {
		if p, err := exec.LookPath(b); err == nil {
			out = append(out, p)
		}
	}
	return out
}

func openDefaultBrowser(url string) error {
	if runtime.GOOS == "darwin" {
		return exec.Command("open", url).Start()
	}
	return exec.Command("xdg-open", url).Start()
}

func bringWindowToFront(string) bool { return false }

// revealInFileManager klasörü dosya yöneticisinde açar (dosyaysa bulunduğu klasörü).
func revealInFileManager(path string, selectFile bool) error {
	if selectFile {
		if runtime.GOOS == "darwin" {
			return exec.Command("open", "-R", path).Start()
		}
		path = filepath.Dir(path)
	}
	if runtime.GOOS == "darwin" {
		return exec.Command("open", path).Start()
	}
	return exec.Command("xdg-open", path).Start()
}

func closeAppWindows(string) {}

func hideCmdWindow(*exec.Cmd) {}

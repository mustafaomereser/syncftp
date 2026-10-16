//go:build !windows

package main

import (
	"os"
	"os/exec"
	"runtime"
)

// Windows dışında çift tık tespiti yok — GUI `syncftp gui` ile açılır.
func ownsConsole() bool { return false }

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
	for _, b := range []string{"google-chrome", "chromium", "chromium-browser", "microsoft-edge"} {
		if p, err := exec.LookPath(b); err == nil {
			c := exec.Command(p, "--app="+url, "--user-data-dir="+profileDir, "--window-size=1320,860", "--no-first-run")
			if c.Start() == nil {
				return c.Process.Release()
			}
		}
	}
	return openDefaultBrowser(url)
}

func openDefaultBrowser(url string) error {
	if runtime.GOOS == "darwin" {
		return exec.Command("open", url).Start()
	}
	return exec.Command("xdg-open", url).Start()
}

func bringWindowToFront(string) bool { return false }

func closeAppWindows(string) {}


func hideCmdWindow(*exec.Cmd) {}

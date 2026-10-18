//go:build !windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// moveToTrash dosya/klasörü çöp kutusuna taşır: macOS ~/.Trash, Linux `gio trash`
// (yoksa hata — çağıran yeniden adlandırmaya düşer).
func moveToTrash(path string) error {
	if runtime.GOOS == "darwin" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		dst := filepath.Join(home, ".Trash", filepath.Base(filepath.Dir(path))+"-"+filepath.Base(path)+"-"+time.Now().Format("150405"))
		return os.Rename(path, dst)
	}
	if gio, err := exec.LookPath("gio"); err == nil {
		if out, err := exec.Command(gio, "trash", path).CombinedOutput(); err != nil {
			return fmt.Errorf("gio trash: %v %s", err, out)
		}
		return nil
	}
	return fmt.Errorf("trash unavailable")
}

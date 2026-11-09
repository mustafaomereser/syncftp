//go:build cgo && (darwin || linux)

// syncftp-window: macOS ve Linux'ta panelin yerel penceresi. Ana program CGO'suz kalsın diye (Windows'tan
// çapraz derleniyor) ayrı programdır; yerel makinede derlenir (.github/workflows/native.yml).
//
//	syncftp-window --url <adres> --title <başlık>   pencereyi aç (openAppWindow)
//	syncftp-window                                  .app çift tık: paneli başlat, sonra pencereyi aç
//	syncftp-window --probe                          bu makinede çalışır mı (exit 0)
package main

/*
#cgo darwin CFLAGS: -x objective-c -fobjc-arc -Wno-deprecated-declarations
#cgo darwin LDFLAGS: -framework Cocoa -framework WebKit
#cgo linux pkg-config: gtk+-3.0 webkit2gtk-4.1
#include <stdlib.h>
#include "window.h"
*/
import "C"

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

func init() { runtime.LockOSThread() } // Cocoa/GTK ana iş parçacığında çalışmak zorunda

// bootstrap ana programı ayrık başlatır ve giriş adresini okur. Pipe değil dosya: bu pencere kapanınca
// panel yaşamaya devam ediyor (watch vb.), kapanan pipe'a yazdığında SIGPIPE ile ölürdü.
func bootstrap() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	f, err := os.CreateTemp("", "syncftp-url-")
	if err != nil {
		return "", err
	}
	defer os.Remove(f.Name())
	defer f.Close()
	c := exec.Command(filepath.Join(filepath.Dir(exe), "syncftp"), "gui", "--foreground", "--no-window")
	c.Stdout, c.Stderr = f, f
	if home, err := os.UserHomeDir(); err == nil {
		c.Dir = home
	}
	// Finder'dan açılan uygulamanın PATH'i çok kısa: Homebrew araçları bulunsun
	c.Env = append(os.Environ(), "PATH="+os.Getenv("PATH")+":/opt/homebrew/bin:/usr/local/bin")
	c.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := c.Start(); err != nil {
		return "", err
	}
	_ = c.Process.Release()
	for i := 0; i < 300; i++ {
		data, _ := os.ReadFile(f.Name())
		for _, line := range strings.Split(string(data), "\n") {
			if strings.Contains(line, "/login?t=") {
				return strings.TrimSpace(line), nil
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	return "", fmt.Errorf("panel did not start")
}

func main() {
	url := flag.String("url", "", "")
	title := flag.String("title", "syncFTP", "")
	probe := flag.Bool("probe", false, "")
	flag.Parse()
	if *probe {
		if C.window_probe() != 1 {
			os.Exit(1)
		}
		return
	}
	if *url == "" {
		u, err := bootstrap()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		*url = u
	}
	cu, ct, cs := C.CString(*url), C.CString(*title), C.CString(os.Getenv("SYNCFTP_WINDOW_SMOKE"))
	defer C.free(unsafe.Pointer(cu))
	defer C.free(unsafe.Pointer(ct))
	defer C.free(unsafe.Pointer(cs))
	C.window_run(cu, ct, cs)
}

//go:build windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

var (
	kernel32                    = windows.NewLazySystemDLL("kernel32.dll")
	user32                      = windows.NewLazySystemDLL("user32.dll")
	procGetConsoleProcessList   = kernel32.NewProc("GetConsoleProcessList")
	procGetConsoleWindow        = kernel32.NewProc("GetConsoleWindow")
	procFreeConsole             = kernel32.NewProc("FreeConsole")
	procShowWindow              = user32.NewProc("ShowWindow")
	procEnumWindows             = user32.NewProc("EnumWindows")
	procGetWindowTextW          = user32.NewProc("GetWindowTextW")
	procIsWindowVisible         = user32.NewProc("IsWindowVisible")
	procSetForegroundWindow     = user32.NewProc("SetForegroundWindow")
	procIsIconic                = user32.NewProc("IsIconic")
	procAllowSetForegroundWindow = user32.NewProc("AllowSetForegroundWindow")
	procPostMessageW            = user32.NewProc("PostMessageW")
)

// ownsConsole süreç kendi konsolunun tek sahibi mi → Explorer'dan çift tıkla açılmış demektir.
// Terminalden çalıştırılınca konsolu shell ile paylaşır (liste ≥ 2).
func ownsConsole() bool {
	var pids [4]uint32
	n, _, _ := procGetConsoleProcessList.Call(uintptr(unsafe.Pointer(&pids[0])), uintptr(len(pids)))
	return n == 1
}

// hideConsole çift tıkla açılan konsol penceresini gizler ve bırakır.
func hideConsole() {
	if hwnd, _, _ := procGetConsoleWindow.Call(); hwnd != 0 {
		procShowWindow.Call(hwnd, 0) // SW_HIDE
	}
	procFreeConsole.Call()
}

// spawnDetached exe'yi konsolsuz, bağımsız süreç olarak başlatır (terminal hemen döner).
func spawnDetached(args ...string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	c := exec.Command(exe, args...)
	c.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: windows.DETACHED_PROCESS | windows.CREATE_NEW_PROCESS_GROUP,
		HideWindow:    true,
	}
	if err := c.Start(); err != nil {
		return err
	}
	return c.Process.Release()
}

// appBrowser --app modunu destekleyen tarayıcı (önce Edge — Windows'ta her zaman var — sonra Chrome).
func appBrowser() string {
	for _, exe := range []string{"msedge.exe", "chrome.exe"} {
		for _, root := range []registry.Key{registry.LOCAL_MACHINE, registry.CURRENT_USER} {
			k, err := registry.OpenKey(root, `SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\`+exe, registry.QUERY_VALUE)
			if err != nil {
				continue
			}
			p, _, err := k.GetStringValue("")
			k.Close()
			if err == nil && fileExistsPlain(strings.Trim(p, `"`)) {
				return strings.Trim(p, `"`)
			}
		}
		sub := map[string]string{"msedge.exe": `Microsoft\Edge\Application`, "chrome.exe": `Google\Chrome\Application`}[exe]
		for _, env := range []string{"ProgramFiles(x86)", "ProgramFiles", "LocalAppData"} {
			if d := os.Getenv(env); d != "" {
				if p := filepath.Join(d, sub, exe); fileExistsPlain(p) {
					return p
				}
			}
		}
	}
	return ""
}

// openAppWindow paneli adres çubuksuz uygulama penceresi olarak açar.
func openAppWindow(url, profileDir string) error {
	b := appBrowser()
	if b == "" {
		return openDefaultBrowser(url)
	}
	c := exec.Command(b,
		"--app="+url,
		"--user-data-dir="+profileDir,
		"--window-size=1320,860",
		"--no-first-run",
		"--no-default-browser-check",
		"--disable-features=Translate",
	)
	c.SysProcAttr = &syscall.SysProcAttr{CreationFlags: windows.DETACHED_PROCESS | windows.CREATE_NEW_PROCESS_GROUP}
	if err := c.Start(); err != nil {
		return openDefaultBrowser(url)
	}
	return c.Process.Release()
}

func openDefaultBrowser(url string) error {
	return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
}

// closeAppWindows başlığı suffix ile biten tüm görünür pencerelere WM_CLOSE gönderir.
func closeAppWindows(suffix string) {
	var found []uintptr
	cb := syscall.NewCallback(func(hwnd, _ uintptr) uintptr {
		if v, _, _ := procIsWindowVisible.Call(hwnd); v == 0 {
			return 1
		}
		buf := make([]uint16, 512)
		n, _, _ := procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), 512)
		if n > 0 && strings.HasSuffix(windows.UTF16ToString(buf[:n]), suffix) {
			found = append(found, hwnd)
		}
		return 1
	})
	procEnumWindows.Call(cb, 0)
	for _, h := range found {
		procPostMessageW.Call(h, 0x0010, 0, 0) // WM_CLOSE
	}
}

// bringWindowToFront başlığı suffix ile biten görünür pencere varsa öne getirir.
func bringWindowToFront(suffix string) bool {
	var found uintptr
	cb := syscall.NewCallback(func(hwnd, _ uintptr) uintptr {
		if v, _, _ := procIsWindowVisible.Call(hwnd); v == 0 {
			return 1
		}
		buf := make([]uint16, 256)
		n, _, _ := procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), 256)
		if n > 0 && strings.HasSuffix(windows.UTF16ToString(buf[:n]), suffix) {
			found = hwnd
			return 0
		}
		return 1
	})
	procEnumWindows.Call(cb, 0)
	if found == 0 {
		return false
	}
	if ic, _, _ := procIsIconic.Call(found); ic != 0 {
		procShowWindow.Call(found, 9) // SW_RESTORE
	}
	procAllowSetForegroundWindow.Call(^uintptr(0)) // ASFW_ANY
	procSetForegroundWindow.Call(found)
	return true
}



// hideCmdWindow alt sürecin (git, cloudflared) konsol penceresi açmasını engeller.
func hideCmdWindow(c *exec.Cmd) {
	c.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: windows.CREATE_NO_WINDOW}
}

// ── Windows açılışında başlat (HKCU\...\Run) ─────────────────────────────────

const runKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`
const runValueName = "syncFTP"

func autostartEnabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()
	_, _, err = k.GetStringValue(runValueName)
	return err == nil
}

// setAutostart oturum açılınca paneli pencere açmadan (yalnızca tray) başlatır;
// kayıtlı watch'lar ve webhook dinleyicisi kaldığı yerden devam eder.
func setAutostart(on bool) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	if !on {
		err := k.DeleteValue(runValueName)
		if err == registry.ErrNotExist {
			return nil
		}
		return err
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	return k.SetStringValue(runValueName, `"`+exe+`" gui --foreground --background`)
}

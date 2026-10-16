//go:build windows

package main

import (
	"errors"
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

// pickFolder yerel klasör seçme diyaloğu (PowerShell + WinForms, ek bağımlılık yok).
func pickFolder(initial string) (string, error) {
	script := `Add-Type -AssemblyName System.Windows.Forms;` +
		`$f = New-Object System.Windows.Forms.FolderBrowserDialog;` +
		`$f.ShowNewFolderButton = $true;` +
		`if ($env:SYNCFTP_PICK_INIT) { $f.SelectedPath = $env:SYNCFTP_PICK_INIT };` +
		`$w = New-Object System.Windows.Forms.Form; $w.TopMost = $true;` +
		`if ($f.ShowDialog($w) -eq 'OK') { [Console]::OutputEncoding = [Text.Encoding]::UTF8; Write-Output $f.SelectedPath }`
	c := exec.Command("powershell", "-NoProfile", "-STA", "-NonInteractive", "-Command", script)
	c.Env = append(os.Environ(), "SYNCFTP_PICK_INIT="+initial)
	c.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: windows.CREATE_NO_WINDOW}
	out, err := c.Output()
	if err != nil {
		return "", err
	}
	p := strings.TrimSpace(string(out))
	if p == "" {
		return "", errors.New("cancelled")
	}
	return p, nil
}

func fileExistsPlain(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}

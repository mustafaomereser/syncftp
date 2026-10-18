//go:build windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

var (
	kernel32                     = windows.NewLazySystemDLL("kernel32.dll")
	user32                       = windows.NewLazySystemDLL("user32.dll")
	procGetConsoleProcessList    = kernel32.NewProc("GetConsoleProcessList")
	procGetConsoleWindow         = kernel32.NewProc("GetConsoleWindow")
	procFreeConsole              = kernel32.NewProc("FreeConsole")
	procShowWindow               = user32.NewProc("ShowWindow")
	procEnumWindows              = user32.NewProc("EnumWindows")
	procGetWindowTextW           = user32.NewProc("GetWindowTextW")
	procIsWindowVisible          = user32.NewProc("IsWindowVisible")
	procSetForegroundWindow      = user32.NewProc("SetForegroundWindow")
	procIsIconic                 = user32.NewProc("IsIconic")
	procAllowSetForegroundWindow = user32.NewProc("AllowSetForegroundWindow")
	procPostMessageW             = user32.NewProc("PostMessageW")
	procGetClassNameW            = user32.NewProc("GetClassNameW")
	procGetForegroundWindow      = user32.NewProc("GetForegroundWindow")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procAttachThreadInput        = user32.NewProc("AttachThreadInput")
	procBringWindowToTop         = user32.NewProc("BringWindowToTop")
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

// forceForeground pencereyi öne alır. Arka plandaki süreç (panel sunucusu) SetForegroundWindow'u
// tek başına çağırınca Windows yalnızca görev çubuğunu yakıp söndürür; ön plandaki pencerenin
// iş parçacığına girdi bağlanınca (AttachThreadInput) izin verilir.
func forceForeground(hwnd uintptr) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if ic, _, _ := procIsIconic.Call(hwnd); ic != 0 {
		procShowWindow.Call(hwnd, 9) // SW_RESTORE
	}
	fg, _, _ := procGetForegroundWindow.Call()
	fgThread, _, _ := procGetWindowThreadProcessId.Call(fg, 0)
	me := uintptr(windows.GetCurrentThreadId())
	if fgThread != 0 && fgThread != me {
		procAttachThreadInput.Call(me, fgThread, 1)
		defer procAttachThreadInput.Call(me, fgThread, 0)
	}
	procAllowSetForegroundWindow.Call(^uintptr(0)) // ASFW_ANY
	procBringWindowToTop.Call(hwnd)
	procSetForegroundWindow.Call(hwnd)
}

// explorerWindows görünür Explorer klasör pencereleri (CabinetWClass) → başlık.
func explorerWindows() map[uintptr]string {
	out := map[uintptr]string{}
	cb := syscall.NewCallback(func(hwnd, _ uintptr) uintptr {
		if v, _, _ := procIsWindowVisible.Call(hwnd); v == 0 {
			return 1
		}
		cls := make([]uint16, 64)
		n, _, _ := procGetClassNameW.Call(hwnd, uintptr(unsafe.Pointer(&cls[0])), 64)
		if windows.UTF16ToString(cls[:n]) != "CabinetWClass" {
			return 1
		}
		buf := make([]uint16, 512)
		n, _, _ = procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), 512)
		out[hwnd] = windows.UTF16ToString(buf[:n])
		return 1
	})
	procEnumWindows.Call(cb, 0)
	return out
}

// revealInFileManager klasörü (ya da dosyayı seçili olarak) Explorer'da açar ve pencereyi ÖNE alır —
// panel penceresinin arkasında açılıp "açılmadı" sanılıyordu. Explorer pencereyi kendi sürecinde
// açtığı için yeni CabinetWClass penceresi beklenir; aynı klasör zaten açıksa başlığından bulunur.
func revealInFileManager(path string, selectFile bool) error {
	before := explorerWindows()
	// `cmd /c start` üzerinden: kullanıcı denedi, Explorer böyle doğrudan önde açılıyor. Aşağıdaki
	// bekle + öne al yedek olarak kalıyor. CmdLine elle: /select,"…" Go'nun tırnaklamasıyla bozuluyor.
	cmd := exec.Command("cmd")
	line := `cmd /c start "" "` + path + `"`
	if selectFile {
		line = `cmd /c start "" explorer /select,"` + path + `"`
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: line, HideWindow: true, CreationFlags: windows.CREATE_NO_WINDOW}
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() {
		_ = cmd.Wait() // explorer.exe pencereyi açan sürece devredip hemen çıkar
		want := filepath.Base(path)
		if selectFile {
			want = filepath.Base(filepath.Dir(path))
		}
		for i := 0; i < 80; i++ { // ~8 sn (soğuk açılışta Explorer geç gelebiliyor)
			time.Sleep(100 * time.Millisecond)
			var match uintptr
			for h, title := range explorerWindows() {
				if _, old := before[h]; !old {
					match = h
					break
				}
				lt, lw := strings.ToLower(title), strings.ToLower(want)
				// yeni pencere çıkmadıysa mevcut pencere yeniden kullanılmıştır; Win11 başlığı "klasör - Dosya Gezgini"
				if i >= 15 && (lt == lw || strings.HasPrefix(lt, lw+" - ")) {
					match = h
				}
			}
			if match != 0 {
				forceForeground(match)
				return
			}
		}
	}()
	return nil
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

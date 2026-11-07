//go:build windows

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"unsafe"

	"github.com/jchv/go-webview2/pkg/edge"
	"github.com/spf13/cobra"
	"golang.org/x/sys/windows"
)

// Panel penceresi: Edge --app yerine exe'nin içinde çerçevesiz bir WebView2 penceresi. Windows'un
// başlık çubuğu yok; üst bar sayfada HTML (app.js renderTitlebar) — sürükleme WebView2'nin
// "non-client region" desteğiyle (CSS app-region: drag → başlık çubuğu gibi davranır: sürükle,
// çift tıkla büyüt, sağ tıkta sistem menüsü, Aero Snap). Pencere AYRI bir süreçte açılır
// (`syncftp panel-window`): WebView2 çökse de panel sunucusu ve tepsi yaşar, eski Edge penceresiyle
// aynı yaşam döngüsü. WebView2 yoksa / açılamazsa eski Edge --app penceresine düşülür.

var (
	pwURL, pwTitle, pwProfile string
)

func init() {
	panelWindowCmd.Flags().StringVar(&pwURL, "url", "", "")
	panelWindowCmd.Flags().StringVar(&pwTitle, "title", "syncFTP", "")
	panelWindowCmd.Flags().StringVar(&pwProfile, "profile", "", "")
	rootCmd.AddCommand(panelWindowCmd)
}

var panelWindowCmd = &cobra.Command{
	Use:    "panel-window",
	Hidden: true,
	Args:   cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := runPanelWindow(pwURL, pwTitle, pwProfile); err != nil {
			// WebView2 açılamadı → eski yol
			return openEdgeAppWindow(pwURL, pwProfile)
		}
		return nil
	},
}

// openAppWindow paneli açar: kendi WebView2 pencere sürecimiz (SYNCFTP_EDGE_APP=1 → eski Edge --app).
func openAppWindow(url, profileDir string) error {
	if os.Getenv("SYNCFTP_EDGE_APP") == "1" || !webview2Available() {
		return openEdgeAppWindow(url, profileDir)
	}
	exe, err := os.Executable()
	if err != nil {
		return openEdgeAppWindow(url, profileDir)
	}
	return spawnDetachedExe(exe, "panel-window", "--url", url, "--title", guiWindowTitle(), "--profile", filepath.Join(filepath.Dir(profileDir), "webview2"))
}

// webview2Available çalışma ortamı kurulu mu (EdgeUpdate kayıtları — makine ya da kullanıcı).
func webview2Available() bool {
	const id = `{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}`
	for _, k := range []struct {
		root windows.Handle
		path string
	}{
		{windows.HKEY_LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Microsoft\EdgeUpdate\Clients\` + id},
		{windows.HKEY_LOCAL_MACHINE, `SOFTWARE\Microsoft\EdgeUpdate\Clients\` + id},
		{windows.HKEY_CURRENT_USER, `Software\Microsoft\EdgeUpdate\Clients\` + id},
	} {
		var h windows.Handle
		p, _ := windows.UTF16PtrFromString(k.path)
		if windows.RegOpenKeyEx(k.root, p, 0, windows.KEY_READ, &h) == nil {
			var buf [64]uint16
			n := uint32(len(buf) * 2)
			name, _ := windows.UTF16PtrFromString("pv")
			err := windows.RegQueryValueEx(h, name, nil, nil, (*byte)(unsafe.Pointer(&buf[0])), &n)
			windows.RegCloseKey(h)
			if err == nil {
				if v := windows.UTF16ToString(buf[:]); v != "" && v != "0.0.0.0" {
					return true
				}
			}
		}
	}
	return false
}

var (
	wvUser32   = windows.NewLazySystemDLL("user32.dll")
	wvDwmapi   = windows.NewLazySystemDLL("dwmapi.dll")
	wvShell32  = windows.NewLazySystemDLL("shell32.dll")
	wvGdi32    = windows.NewLazySystemDLL("gdi32.dll")
	wvRegister = wvUser32.NewProc("RegisterClassExW")
	wvCreate   = wvUser32.NewProc("CreateWindowExW")
	wvDefProc  = wvUser32.NewProc("DefWindowProcW")
	wvShow     = wvUser32.NewProc("ShowWindow")
	wvUpdate   = wvUser32.NewProc("UpdateWindow")
	wvGetMsg   = wvUser32.NewProc("GetMessageW")
	wvTrans    = wvUser32.NewProc("TranslateMessage")
	wvDispatch = wvUser32.NewProc("DispatchMessageW")
	wvPostQuit = wvUser32.NewProc("PostQuitMessage")
	wvDestroy  = wvUser32.NewProc("DestroyWindow")
	wvClient   = wvUser32.NewProc("GetClientRect")
	wvSetPos   = wvUser32.NewProc("SetWindowPos")
	wvZoomed   = wvUser32.NewProc("IsZoomed")
	wvDpiWin   = wvUser32.NewProc("GetDpiForWindow")
	wvMetric   = wvUser32.NewProc("GetSystemMetricsForDpi")
	wvSendMsg  = wvUser32.NewProc("SendMessageW")
	wvPostMsg  = wvUser32.NewProc("PostMessageW")
	wvMonitor  = wvUser32.NewProc("MonitorFromWindow")
	wvMonInfo  = wvUser32.NewProc("GetMonitorInfoW")
	wvDpiCtx   = wvUser32.NewProc("SetProcessDpiAwarenessContext")
	wvLoadCur  = wvUser32.NewProc("LoadCursorW")
	wvFgWin    = wvUser32.NewProc("SetForegroundWindow")
	wvDwmAttr  = wvDwmapi.NewProc("DwmSetWindowAttribute")
	wvExtract  = wvShell32.NewProc("ExtractIconExW")
	wvShellEx  = wvShell32.NewProc("ShellExecuteW")
	wvBrush    = wvGdi32.NewProc("CreateSolidBrush")
)

const (
	wmDestroy       = 0x0002
	wmMove          = 0x0003
	wmSize          = 0x0005
	wmActivate      = 0x0006
	wmClose         = 0x0010
	wmGetMinMaxInfo = 0x0024
	wmSetIcon       = 0x0080
	wmNCCalcSize    = 0x0083
	wmNCHitTest     = 0x0084
	wmMoving        = 0x0216
	wmDpiChanged    = 0x02E0
	wmAppTitleState = 0x8000 + 7

	htTop     = 12
	htTopLeft = 13
	htTopRt   = 14

	swShow     = 5
	swMinimize = 6
	swMaximize = 3
	swRestore  = 9
)

// panelBg panelin arka planı (app.css --bg) — WebView2 yüklenirken beyaz parlama olmasın.
const panelBgR, panelBgG, panelBgB = 0x11, 0x11, 0x13

type wvRect struct{ Left, Top, Right, Bottom int32 }
type wvPoint struct{ X, Y int32 }
type wvMinMax struct {
	Reserved, MaxSize, MaxPosition, MinTrackSize, MaxTrackSize wvPoint
}
type wvMsg struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      wvPoint
}
type wvNCCalc struct {
	Rgrc [3]wvRect
	Pos  uintptr
}
type wvMonitorInfo struct {
	Size    uint32
	Monitor wvRect
	Work    wvRect
	Flags   uint32
}
type wvClass struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   windows.Handle
	Icon       windows.Handle
	Cursor     uintptr
	Background uintptr
	MenuName   *uint16
	ClassName  *uint16
	IconSm     windows.Handle
}

// panelWin tek pencere (süreç başına bir tane).
var panelWin struct {
	hwnd      uintptr
	chromium  *edge.Chromium
	frameless bool // sürükleme desteği doğrulandıysa başlık çubuğu kalkar; yoksa Windows çerçevesi kalır
}

// topInset: büyütülmemiş pencerede WebView üst kenardan bu kadar aşağıda başlar → o şerit
// pencerenin kendisine kalır, WM_NCHITTEST ile üst kenardan boyutlandırma çalışır.
func topInset(hwnd uintptr) int32 {
	if z, _, _ := wvZoomed.Call(hwnd); z != 0 || !panelWin.frameless {
		return 0
	}
	return int32(dpiScale(hwnd, 4))
}

func dpiScale(hwnd uintptr, v int) int {
	dpi, _, _ := wvDpiWin.Call(hwnd)
	if dpi == 0 {
		dpi = 96
	}
	return v * int(dpi) / 96
}

func resizeWebView() {
	c := panelWin.chromium
	if c == nil || panelWin.hwnd == 0 {
		return
	}
	ctrl := c.GetController()
	if ctrl == nil {
		return
	}
	var r wvRect
	wvClient.Call(panelWin.hwnd, uintptr(unsafe.Pointer(&r)))
	r.Top += topInset(panelWin.hwnd)
	comCall(uintptr(unsafe.Pointer(ctrl)), 6, uintptr(unsafe.Pointer(&r))) // ICoreWebView2Controller::put_Bounds
}

// ptrAt Win32'nin uintptr olarak verdiği adresi işaretçiye çevirir (lParam yapıları, COM sanal tablosu).
// Adres Go yığınında değil, işletim sisteminin belleğinde — taşınma riski yok.
func ptrAt[T any](p uintptr) *T { return *(**T)(unsafe.Pointer(&p)) }

// comCall COM nesnesinin sanal tablosundaki idx'inci yöntemi çağırır (internal tiplere ihtiyaç yok).
func comCall(obj uintptr, idx int, args ...uintptr) uintptr {
	vtbl := *ptrAt[uintptr](obj)
	fn := *ptrAt[uintptr](vtbl + uintptr(idx)*unsafe.Sizeof(uintptr(0)))
	r, _, _ := syscall.SyscallN(fn, append([]uintptr{obj}, args...)...)
	return r
}

// IID_ICoreWebView2Settings9 — put_IsNonClientRegionSupportEnabled (app-region: drag)
var iidSettings9 = windows.GUID{Data1: 0x0528A73B, Data2: 0xE92D, Data3: 0x49F4, Data4: [8]byte{0x92, 0x7A, 0xE5, 0x47, 0xDD, 0xDA, 0xA3, 0x7D}}

// enableNonClientRegions ICoreWebView2Settings9::put_IsNonClientRegionSupportEnabled(TRUE).
// Sanal tablo sırası (WebView2.h): Settings 3..20, Settings2 21-22, 3 23-24, 4 25-28, 5 29-30,
// 6 31-32, 7 33-34, 8 35-36, 9 37 (get) / 38 (put). Yazdıktan sonra okuyup doğrular.
func enableNonClientRegions(settings uintptr) bool {
	var s9 uintptr
	if hr := comCall(settings, 0, uintptr(unsafe.Pointer(&iidSettings9)), uintptr(unsafe.Pointer(&s9))); hr != 0 || s9 == 0 {
		return false
	}
	defer comCall(s9, 2) // Release
	if hr := comCall(s9, 38, 1); hr != 0 {
		return false
	}
	var v int32
	return comCall(s9, 37, uintptr(unsafe.Pointer(&v))) == 0 && v != 0
}

func wvWndProc(hwnd, msg, wp, lp uintptr) uintptr {
	switch msg {
	case wmNCCalcSize:
		if wp == 0 || !panelWin.frameless {
			break
		}
		// Başlık çubuğunu kaldır: varsayılan hesaplamadan sonra üst kenarı geri al. Yan/alt kenarlar
		// (Windows 10/11'de görünmez boyutlandırma alanı) ve DWM gölgesi kalır.
		p := ptrAt[wvNCCalc](lp)
		top := p.Rgrc[0].Top
		wvDefProc.Call(hwnd, msg, wp, lp)
		p.Rgrc[0].Top = top
		if z, _, _ := wvZoomed.Call(hwnd); z != 0 {
			// büyütülmüş pencere ekran dışına çerçeve kalınlığı kadar taşar
			dpi, _, _ := wvDpiWin.Call(hwnd)
			fy, _, _ := wvMetric.Call(33, dpi) // SM_CYFRAME
			pb, _, _ := wvMetric.Call(92, dpi) // SM_CXPADDEDBORDER
			p.Rgrc[0].Top += int32(fy + pb)
		}
		return 0
	case wmNCHitTest:
		r, _, _ := wvDefProc.Call(hwnd, msg, wp, lp)
		if r != 1 || !panelWin.frameless { // HTCLIENT değilse (yan kenarlar) varsayılan
			return r
		}
		var wr wvRect
		windowRect(hwnd, &wr)
		y := int32(int16(lp >> 16))
		x := int32(int16(lp & 0xffff))
		if z, _, _ := wvZoomed.Call(hwnd); z == 0 && y < wr.Top+int32(dpiScale(hwnd, 5)) {
			edge := int32(dpiScale(hwnd, 12))
			switch {
			case x < wr.Left+edge:
				return htTopLeft
			case x > wr.Right-edge:
				return htTopRt
			}
			return htTop
		}
		return r
	case wmSize:
		resizeWebView()
		state := "normal"
		if z, _, _ := wvZoomed.Call(hwnd); z != 0 {
			state = "max"
		}
		if panelWin.chromium != nil && panelWin.chromium.GetController() != nil {
			panelWin.chromium.Eval("window.__winState&&window.__winState('" + state + "')")
		}
		return 0
	case wmMove, wmMoving:
		if panelWin.chromium != nil && panelWin.chromium.GetController() != nil {
			_ = panelWin.chromium.NotifyParentWindowPositionChanged()
		}
	case wmActivate:
		if wp&0xffff != 0 && panelWin.chromium != nil && panelWin.chromium.GetController() != nil {
			panelWin.chromium.Focus()
		}
	case wmGetMinMaxInfo:
		mm := ptrAt[wvMinMax](lp)
		mm.MinTrackSize = wvPoint{int32(dpiScale(hwnd, 760)), int32(dpiScale(hwnd, 480))}
		return 0
	case wmDpiChanged:
		r := ptrAt[wvRect](lp)
		wvSetPos.Call(hwnd, 0, uintptr(r.Left), uintptr(r.Top), uintptr(r.Right-r.Left), uintptr(r.Bottom-r.Top), 0x0014) // NOZORDER|NOACTIVATE
		return 0
	case wmClose:
		wvDestroy.Call(hwnd)
		return 0
	case wmDestroy:
		wvPostQuit.Call(0)
		return 0
	}
	r, _, _ := wvDefProc.Call(hwnd, msg, wp, lp)
	return r
}

var wvGetWinRect = wvUser32.NewProc("GetWindowRect")

func windowRect(hwnd uintptr, r *wvRect) { wvGetWinRect.Call(hwnd, uintptr(unsafe.Pointer(r))) }

// onPanelMessage sayfadan gelen pencere komutları: win:min | win:max | win:close | open:<url>
func onPanelMessage(msg string) {
	h := panelWin.hwnd
	switch {
	case msg == "win:min":
		wvShow.Call(h, swMinimize)
	case msg == "win:max":
		if z, _, _ := wvZoomed.Call(h); z != 0 {
			wvShow.Call(h, swRestore)
		} else {
			wvShow.Call(h, swMaximize)
		}
	case msg == "win:close":
		wvPostMsg.Call(h, wmClose, 0, 0)
	case strings.HasPrefix(msg, "open:"):
		// harici bağlantılar varsayılan tarayıcıda (WebView2 yeni pencerede açardı)
		u := strings.TrimPrefix(msg, "open:")
		if strings.HasPrefix(u, "https://") || strings.HasPrefix(u, "http://") {
			op, _ := windows.UTF16PtrFromString("open")
			up, _ := windows.UTF16PtrFromString(u)
			wvShellEx.Call(0, uintptr(unsafe.Pointer(op)), uintptr(unsafe.Pointer(up)), 0, 0, 1)
		}
	}
}

// runPanelWindow pencereyi açar ve kapanana kadar mesaj döngüsünü çalıştırır.
func runPanelWindow(url, title, profile string) error {
	if url == "" {
		return errors.New("url")
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	wvDpiCtx.Call(^uintptr(3)) // DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2 (-4)
	_ = windows.CoInitializeEx(0, windows.COINIT_APARTMENTTHREADED)

	var inst windows.Handle
	_ = windows.GetModuleHandleEx(0, nil, &inst)
	cls, _ := windows.UTF16PtrFromString("SyncFTPPanel")
	cursor, _, _ := wvLoadCur.Call(0, 32512) // IDC_ARROW
	brush, _, _ := wvBrush.Call(uintptr(panelBgR) | uintptr(panelBgG)<<8 | uintptr(panelBgB)<<16)
	var large, small windows.Handle
	if exe, err := os.Executable(); err == nil {
		ep, _ := windows.UTF16PtrFromString(exe)
		wvExtract.Call(uintptr(unsafe.Pointer(ep)), 0, uintptr(unsafe.Pointer(&large)), uintptr(unsafe.Pointer(&small)), 1)
	}
	wc := wvClass{
		Size: uint32(unsafe.Sizeof(wvClass{})), Style: 0x0003, // CS_HREDRAW|CS_VREDRAW
		WndProc: syscall.NewCallback(wvWndProc), Instance: inst, Icon: large, IconSm: small,
		Cursor: cursor, Background: brush, ClassName: cls,
	}
	wvRegister.Call(uintptr(unsafe.Pointer(&wc)))

	// Boyut DPI'ya göre; açık ekranın çalışma alanına sığdır, ortala
	tp, _ := windows.UTF16PtrFromString(title)
	hwnd, _, _ := wvCreate.Call(0, uintptr(unsafe.Pointer(cls)), uintptr(unsafe.Pointer(tp)),
		0x00CF0000, // WS_OVERLAPPEDWINDOW (snap, gölge, sistem menüsü kalsın)
		0x80000000, 0x80000000, 1320, 860, 0, 0, uintptr(inst), 0)
	if hwnd == 0 {
		return errors.New("CreateWindowEx")
	}
	panelWin.hwnd = hwnd
	placeWindow(hwnd, 1320, 860)
	// Koyu çerçeve / gölge rengi (Windows 11): DWMWA_USE_IMMERSIVE_DARK_MODE = 20
	dark := int32(1)
	wvDwmAttr.Call(hwnd, 20, uintptr(unsafe.Pointer(&dark)), 4)

	c := edge.NewChromium()
	c.DataPath = profile
	c.MessageCallback = onPanelMessage
	c.SetPermission(edge.CoreWebView2PermissionKindClipboardRead, edge.CoreWebView2PermissionStateAllow)
	panelWin.chromium = c
	if !c.Embed(hwnd) {
		wvDestroy.Call(hwnd)
		panelWin.chromium = nil
		return errors.New("WebView2")
	}
	if ctrl := c.GetController(); ctrl != nil {
		if c2 := ctrl.GetICoreWebView2Controller2(); c2 != nil {
			_ = c2.PutDefaultBackgroundColor(edge.COREWEBVIEW2_COLOR{A: 255, R: panelBgR, G: panelBgG, B: panelBgB})
		}
	}
	nonClient := false
	if s, err := c.GetSettings(); err == nil && s != nil {
		_ = s.PutAreDefaultContextMenusEnabled(true) // input'larda kes/kopyala/yapıştır; diğer yerlerde panelin kendi menüsü
		_ = s.PutAreDevToolsEnabled(os.Getenv("SYNCFTP_DEVTOOLS") == "1")
		_ = s.PutIsStatusBarEnabled(false)
		nonClient = enableNonClientRegions(uintptr(unsafe.Pointer(s)))
	}
	// Destek doğrulandıysa başlık çubuğunu kaldır (SWP_FRAMECHANGED → WM_NCCALCSIZE yeniden)
	panelWin.frameless = nonClient
	wvSetPos.Call(hwnd, 0, 0, 0, 0, 0, 0x0027) // FRAMECHANGED|NOMOVE|NOSIZE|NOZORDER
	// Sayfa kendi üst barını çizsin (app.js): host + sürükleme desteği bilgisi
	c.Init(fmt.Sprintf("window.__SYNCFTP_FRAME={host:'webview2',drag:%t};", nonClient))
	resizeWebView()
	c.Navigate(url)
	wvShow.Call(hwnd, swShow)
	wvShow.Call(hwnd, 1) // SW_SHOWNORMAL: ilk çağrı STARTUPINFO ile ezilebilir (simge durumunda açılmasın)
	wvUpdate.Call(hwnd)
	_ = c.Show() // denetleyici, pencere gizliyken bağlandığı için görünmez başlar
	wvFgWin.Call(hwnd)

	var m wvMsg
	for {
		r, _, _ := wvGetMsg.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if r == 0 || int32(r) == -1 {
			break
		}
		wvTrans.Call(uintptr(unsafe.Pointer(&m)))
		wvDispatch.Call(uintptr(unsafe.Pointer(&m)))
	}
	return nil
}

// placeWindow pencereyi açıldığı ekranın çalışma alanında ortalar (DPI ölçekli, sığmazsa küçültür).
func placeWindow(hwnd uintptr, w, h int) {
	w, h = dpiScale(hwnd, w), dpiScale(hwnd, h)
	mon, _, _ := wvMonitor.Call(hwnd, 2) // MONITOR_DEFAULTTONEAREST
	mi := wvMonitorInfo{Size: uint32(unsafe.Sizeof(wvMonitorInfo{}))}
	wvMonInfo.Call(mon, uintptr(unsafe.Pointer(&mi)))
	ww, wh := int(mi.Work.Right-mi.Work.Left), int(mi.Work.Bottom-mi.Work.Top)
	if w > ww-40 {
		w = ww - 40
	}
	if h > wh-40 {
		h = wh - 40
	}
	x := int(mi.Work.Left) + (ww-w)/2
	y := int(mi.Work.Top) + (wh-h)/2
	wvSetPos.Call(hwnd, 0, uintptr(x), uintptr(y), uintptr(w), uintptr(h), 0x0014)
}

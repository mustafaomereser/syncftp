//go:build windows

package main

import (
	"fmt"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"

	"syncftp/internal/lang"
	"syncftp/internal/projects"
)

// ══════════════════════════════════════════════════════════════════════════════
// Sistem tepsisi — saf Win32, kütüphanesiz (agent-company'deki tepsiyle aynı yaklaşım).
//
// Görünmez bir mesaj penceresi + Shell_NotifyIcon. Sağ tıkta menü O ANDA kurulur
// (durum her açılışta taze) ve "sahibi çizer" (owner-draw) satırlardan oluşur:
// başlık (logo + sürüm/durum), Segoe ikonları, sağda rozet / yeşil nokta, üzerine
// gelince yuvarlak vurgu, kırmızı Çıkış. Çizim gui_traymenu_windows.go'da.
// Mesaj döngüsü pencereyle AYNI iş parçacığında koşmalı → runtime.LockOSThread.
// ══════════════════════════════════════════════════════════════════════════════

var (
	trUser32   = windows.NewLazySystemDLL("user32.dll")
	trShell32  = windows.NewLazySystemDLL("shell32.dll")
	trKernel32 = windows.NewLazySystemDLL("kernel32.dll")

	pRegisterClassEx  = trUser32.NewProc("RegisterClassExW")
	pCreateWindowEx   = trUser32.NewProc("CreateWindowExW")
	pDefWindowProc    = trUser32.NewProc("DefWindowProcW")
	pDestroyWindow    = trUser32.NewProc("DestroyWindow")
	pGetMessage       = trUser32.NewProc("GetMessageW")
	pTranslateMessage = trUser32.NewProc("TranslateMessage")
	pDispatchMessage  = trUser32.NewProc("DispatchMessageW")
	pPostQuitMessage  = trUser32.NewProc("PostQuitMessage")
	pPostMessage      = trUser32.NewProc("PostMessageW")
	pCreatePopupMenu  = trUser32.NewProc("CreatePopupMenu")
	pAppendMenu       = trUser32.NewProc("AppendMenuW")
	pTrackPopupMenu   = trUser32.NewProc("TrackPopupMenu")
	pDestroyMenu      = trUser32.NewProc("DestroyMenu")
	pGetCursorPos     = trUser32.NewProc("GetCursorPos")
	pSetForegroundWnd = trUser32.NewProc("SetForegroundWindow")
	pLoadIcon         = trUser32.NewProc("LoadIconW")
	pLoadCursor       = trUser32.NewProc("LoadCursorW")
	pCreateIconFromRe = trUser32.NewProc("CreateIconFromResourceEx")
	pDestroyIcon      = trUser32.NewProc("DestroyIcon")
	pOpenClipboard    = trUser32.NewProc("OpenClipboard")
	pEmptyClipboard   = trUser32.NewProc("EmptyClipboard")
	pSetClipboardData = trUser32.NewProc("SetClipboardData")
	pCloseClipboard   = trUser32.NewProc("CloseClipboard")

	pShellNotifyIcon = trShell32.NewProc("Shell_NotifyIconW")
	pExtractIconEx   = trShell32.NewProc("ExtractIconExW")

	pGetModuleHandle = trKernel32.NewProc("GetModuleHandleW")
	pGlobalAlloc     = trKernel32.NewProc("GlobalAlloc")
	pGlobalLock      = trKernel32.NewProc("GlobalLock")
	pGlobalUnlock    = trKernel32.NewProc("GlobalUnlock")
)

const (
	twmDestroy  = 0x0002
	twmClose    = 0x0010
	twmTrayIcon = 0x8000 + 1
	twmLBtnUp   = 0x0202
	twmRBtnUp   = 0x0205

	nimAdd    = 0x0000
	nimModify = 0x0001
	nimDelete = 0x0002

	nifMessage = 0x0001
	nifIcon    = 0x0002
	nifTip     = 0x0004
	nifInfo    = 0x0010

	niifError     = 0x0003
	niifUser      = 0x0004
	niifLargeIcon = 0x0020

	mfPopup = 0x0010

	tpmRightAlign  = 0x0008 // menü imlecin SOLUNA açılır (tepsi sağ altta)
	tpmBottomAlign = 0x0020
	tpmRightButton = 0x0002
	tpmReturnCmd   = 0x0100
)

// Menü kimlikleri (TPM_RETURNCMD seçileni doğrudan döndürür).
const (
	idOpen = 1 + iota
	idJobs
	idTunnel
	idTunnelCopy
	idHelp
	idQuit
	idWatchBase   = 100 // + watch sırası
	idProjectBase = 200 // + proje sırası
)

type tPoint struct{ X, Y int32 }

type tMsg struct {
	HWnd    windows.HWND
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      tPoint
}

type tWndClassEx struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     windows.Handle
	HIcon         windows.Handle
	HCursor       windows.Handle
	HbrBackground windows.Handle
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       windows.Handle
}

type tNotifyIconData struct {
	CbSize           uint32
	HWnd             windows.HWND
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            windows.Handle
	SzTip            [128]uint16
	DwState          uint32
	DwStateMask      uint32
	SzInfo           [256]uint16
	UVersion         uint32
	SzInfoTitle      [64]uint16
	DwInfoFlags      uint32
	GuidItem         windows.GUID
	HBalloonIcon     windows.Handle
}

type winTray struct {
	sv       *guiServer
	panelURL string

	hwnd       windows.HWND
	iconIdle   windows.Handle // tepsi: logo (16 px)
	iconActive windows.Handle // tepsi: logo + yeşil nokta (iş/watch çalışırken)
	bigIcon    windows.Handle // menü başlığı: logo (32 px)
	nid        tNotifyIconData
	nidMu      sync.Mutex
	active     bool

	// Açık menünün satırları ve yazı tipleri; yalnızca menü açıkken dolu.
	items []trayItem
	fonts trayFonts

	watchProjects [][2]string // proje, sunucu
	projectNames  []string
	once          sync.Once
}

func trayAvailable() bool { return true }

// runTray tepsiyi kurar ve sv.quit kapanana (ya da Çıkış seçilene) kadar bloklar.
func (sv *guiServer) runTray(panelURL string) {
	t := &winTray{sv: sv, panelURL: panelURL}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	// Menü yüksek DPI'da bulanık çizilmesin: yalnızca bu iş parçacığı monitör başına DPI farkında
	if pSetThreadDpiCtx.Find() == nil {
		pSetThreadDpiCtx.Call(dpiPerMonitor2)
	}
	enableDarkMenus()

	if err := t.create(); err != nil {
		<-sv.quit // tepsi kurulamadı (ör. masaüstü oturumu yok) — panel yine çalışır
		return
	}
	defer t.destroy()

	go func() {
		tk := time.NewTicker(2 * time.Second)
		defer tk.Stop()
		for {
			select {
			case <-tk.C:
				t.refreshIcon()
			case <-sv.quit:
				pPostMessage.Call(uintptr(t.hwnd), twmClose, 0, 0)
				return
			}
		}
	}()
	t.loop()
}

func (t *winTray) create() error {
	inst, _, _ := pGetModuleHandle.Call(0)
	className, _ := windows.UTF16PtrFromString("SyncFTPTray")
	cursor, _, _ := pLoadCursor.Call(0, 32512)
	wc := tWndClassEx{
		LpfnWndProc:   syscall.NewCallback(t.wndProc),
		HInstance:     windows.Handle(inst),
		HCursor:       windows.Handle(cursor),
		LpszClassName: className,
	}
	wc.CbSize = uint32(unsafe.Sizeof(wc))
	if r, _, e := pRegisterClassEx.Call(uintptr(unsafe.Pointer(&wc))); r == 0 {
		return fmt.Errorf("window class: %w", e)
	}
	hwnd, _, e := pCreateWindowEx.Call(0, uintptr(unsafe.Pointer(className)), uintptr(unsafe.Pointer(className)),
		0, 0, 0, 0, 0, 0, 0, inst, 0)
	if hwnd == 0 {
		return fmt.Errorf("window: %w", e)
	}
	t.hwnd = windows.HWND(hwnd)
	t.loadIcons()

	t.nid = tNotifyIconData{HWnd: t.hwnd, UID: 1, UFlags: nifMessage | nifIcon | nifTip,
		UCallbackMessage: twmTrayIcon, HIcon: t.iconIdle}
	t.nid.CbSize = uint32(unsafe.Sizeof(t.nid))
	t.setTip("syncFTP")
	if r, _, e := pShellNotifyIcon.Call(nimAdd, uintptr(unsafe.Pointer(&t.nid))); r == 0 {
		pDestroyWindow.Call(uintptr(t.hwnd))
		t.hwnd = 0
		return fmt.Errorf("notify icon: %w", e)
	}
	return nil
}

// loadIcons: exe'ye gömülü logo (rsrc.syso → logo.ico) 16 ve 32 px; ExtractIconEx gerçek
// 16'lık kareyi verir (bulanık küçültme yok). Aktif ikon: logo + yeşil nokta (gui_icon.go).
func (t *winTray) loadIcons() {
	if exe, err := os.Executable(); err == nil {
		if p, err := windows.UTF16PtrFromString(exe); err == nil {
			pExtractIconEx.Call(uintptr(unsafe.Pointer(p)), 0,
				uintptr(unsafe.Pointer(&t.bigIcon)), uintptr(unsafe.Pointer(&t.iconIdle)), 1)
		}
	}
	if t.iconIdle == 0 {
		t.iconIdle = iconFromICO(trayIconICO(false))
	}
	if t.bigIcon == 0 {
		t.bigIcon = t.iconIdle
	}
	if t.iconIdle == 0 {
		h, _, _ := pLoadIcon.Call(0, 32512)
		t.iconIdle = windows.Handle(h)
	}
	t.iconActive = iconFromICO(trayIconICO(true))
	if t.iconActive == 0 {
		t.iconActive = t.iconIdle
	}
}

// iconFromICO tek görüntülü (BMP) ICO baytlarından HICON üretir.
func iconFromICO(ico []byte) windows.Handle {
	if len(ico) < 22 {
		return 0
	}
	size := *(*uint32)(unsafe.Pointer(&ico[14]))
	off := *(*uint32)(unsafe.Pointer(&ico[18]))
	if int(off+size) > len(ico) {
		return 0
	}
	h, _, _ := pCreateIconFromRe.Call(uintptr(unsafe.Pointer(&ico[off])), uintptr(size), 1, 0x00030000, 0, 0, 0)
	return windows.Handle(h)
}

func (t *winTray) setTip(s string) {
	u := windows.StringToUTF16(s)
	if len(u) > len(t.nid.SzTip) {
		u = u[:len(t.nid.SzTip)]
		u[len(u)-1] = 0
	}
	t.nid.SzTip = [128]uint16{}
	copy(t.nid.SzTip[:], u)
}

// trayStatus tepsinin tek durum kaynağı: ikon, ipucu ve menü başlığı aynı metni gösterir.
// İş/watch sayısı + tünel (açık / açılıyor); hiçbiri yoksa "Boşta".
func (t *winTray) trayStatus() (string, bool) {
	jobs, watches := t.sv.jobs.running(), t.sv.watches.count()
	tun := t.sv.hooks.tunnelInfo()
	var parts []string
	if jobs > 0 || watches > 0 {
		parts = append(parts, fmt.Sprintf(lang.L.TrayStatusFmt, jobs, watches))
	}
	switch {
	case tun.Active && tun.URL != "":
		parts = append(parts, lang.L.TrayTunnelOnShort)
	case tun.Active:
		parts = append(parts, lang.L.TrayTunnelStartingShort)
	}
	if len(parts) == 0 {
		return lang.L.TrayIdle, false
	}
	return strings.Join(parts, "  ·  "), true
}

// refreshIcon durumuna göre ikon ve ipucunu günceller (her 2 sn + tünel/watch değişince hemen).
func (t *winTray) refreshIcon() {
	status, active := t.trayStatus()
	t.nidMu.Lock()
	defer t.nidMu.Unlock()
	if t.hwnd == 0 {
		return
	}
	t.setTip("syncFTP — " + status)
	t.nid.HIcon = t.iconIdle
	if active {
		t.nid.HIcon = t.iconActive
	}
	t.active = active
	pShellNotifyIcon.Call(nimModify, uintptr(unsafe.Pointer(&t.nid)))
}

// notify tepsiden Windows bildirimi (balon) gösterir. body boşsa başlık gövdeye de yazılır
// (Windows boş gövdeli balonu göstermiyor).
func (t *winTray) notify(title, body string, isErr bool) {
	t.nidMu.Lock()
	defer t.nidMu.Unlock()
	if t.hwnd == 0 {
		return
	}
	if body == "" {
		body, title = title, "syncFTP"
	}
	put := func(dst []uint16, s string) {
		for i := range dst {
			dst[i] = 0
		}
		u := windows.StringToUTF16(s)
		if len(u) > len(dst) {
			u = u[:len(dst)]
			u[len(u)-1] = 0
		}
		copy(dst, u)
	}
	put(t.nid.SzInfoTitle[:], title)
	put(t.nid.SzInfo[:], body)
	t.nid.DwInfoFlags = niifUser | niifLargeIcon
	t.nid.HBalloonIcon = t.bigIcon
	if isErr {
		t.nid.DwInfoFlags = niifError
		t.nid.HBalloonIcon = 0
	}
	t.nid.UFlags |= nifInfo
	pShellNotifyIcon.Call(nimModify, uintptr(unsafe.Pointer(&t.nid)))
	// sonraki ikon tazelemeleri aynı balonu yeniden göstermesin
	t.nid.UFlags &^= nifInfo
	t.nid.SzInfo = [256]uint16{}
	t.nid.SzInfoTitle = [64]uint16{}
}

func (t *winTray) destroy() {
	t.once.Do(func() {
		t.nidMu.Lock()
		defer t.nidMu.Unlock()
		if t.hwnd == 0 {
			return
		}
		pShellNotifyIcon.Call(nimDelete, uintptr(unsafe.Pointer(&t.nid)))
		pDestroyWindow.Call(uintptr(t.hwnd))
		t.hwnd = 0
		if t.iconActive != 0 && t.iconActive != t.iconIdle {
			pDestroyIcon.Call(uintptr(t.iconActive))
		}
	})
}

func (t *winTray) loop() {
	var m tMsg
	for {
		r, _, _ := pGetMessage.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if r == 0 || r == ^uintptr(0) {
			return
		}
		pTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		pDispatchMessage.Call(uintptr(unsafe.Pointer(&m)))
	}
}

func (t *winTray) wndProc(hwnd windows.HWND, message uint32, wParam, lParam uintptr) uintptr {
	switch message {
	case wmMeasureItem:
		t.measure((*tMeasureItem)(unsafe.Add(nil, lParam)))
		return 1
	case wmDrawItem:
		t.draw((*tDrawItem)(unsafe.Add(nil, lParam)))
		return 1
	case twmTrayIcon:
		switch uint32(lParam) & 0xffff {
		case twmRBtnUp:
			t.showMenu()
		case twmLBtnUp:
			t.openAt("")
		}
		return 0
	case twmClose:
		t.destroy()
		pPostQuitMessage.Call(0)
		return 0
	case twmDestroy:
		pPostQuitMessage.Call(0)
		return 0
	}
	r, _, _ := pDefWindowProc.Call(uintptr(hwnd), uintptr(message), wParam, lParam)
	return r
}

// openAt paneli verilen sayfada açar: pencere açıksa öne getirip sayfayı presence SSE
// üzerinden değiştirir; değilse login?next=… ile yeni pencere açar.
func (t *winTray) openAt(hash string) {
	if bringWindowToFront(guiWindowTitle()) {
		if hash != "" {
			t.sv.navigate(hash)
		}
		return
	}
	u := t.panelURL
	if hash != "" {
		u += "&next=" + url.QueryEscape(hash)
	}
	_ = openAppWindow(u, guiProfileDir())
}

// showMenu menüyü o anki durumla kurar ve gösterir.
func (t *winTray) showMenu() {
	menu, _, _ := pCreatePopupMenu.Call()
	if menu == 0 {
		return
	}
	defer pDestroyMenu.Call(menu)
	bg, _, _ := pCreateSolidBrush.Call(colBg)
	defer pDeleteObject.Call(bg)
	styleMenu(menu, bg)

	t.fonts = trayFonts{
		text:  newFont(t.scale(14), fwNormal, "Segoe UI"),
		title: newFont(t.scale(15), fwSemibold, "Segoe UI"),
		small: newFont(t.scale(12), fwNormal, "Segoe UI"),
		glyph: newFont(t.scale(15), fwNormal, glyphFace()),
	}
	defer func() {
		for _, f := range []uintptr{t.fonts.text, t.fonts.title, t.fonts.small, t.fonts.glyph} {
			pDeleteObject.Call(f)
		}
		t.fonts = trayFonts{}
		t.items = nil
	}()

	t.items = t.items[:0]
	add := func(m uintptr, it trayItem) {
		t.items = append(t.items, it)
		flags := uintptr(mfOwnerDraw)
		if it.kind != kindItem || it.grayed {
			flags |= mfDisabled
		}
		id := it.id
		if it.sub != 0 {
			flags |= mfPopup
			id = it.sub
		}
		pAppendMenu.Call(m, flags, id, uintptr(len(t.items)))
	}
	sep := func(m uintptr) { add(m, trayItem{kind: kindSep}) }

	jobs, watches := t.sv.jobs.running(), t.sv.watches.all()
	tun := t.sv.hooks.tunnelInfo()
	status, active := t.trayStatus()
	add(menu, trayItem{kind: kindHeader, text: "syncFTP", subText: "v" + appVersion + "  ·  " + status, dot: active})

	add(menu, trayItem{kind: kindItem, id: idOpen, glyph: 0xE8A7, text: lang.L.TrayOpen})

	// Projeler alt menüsü (owner-draw, aynı renkler)
	t.projectNames = t.projectNames[:0]
	if list := projects.List(); len(list) > 0 {
		sub, _, _ := pCreatePopupMenu.Call()
		styleMenu(sub, bg)
		watching := map[string]bool{}
		for _, w := range watches {
			watching[w.Project] = true
		}
		for i, p := range list {
			if i >= 30 {
				break
			}
			t.projectNames = append(t.projectNames, p.Name)
			add(sub, trayItem{kind: kindItem, id: uintptr(idProjectBase + i), glyph: 0xE8B7, text: p.Name, dot: watching[p.Name]})
		}
		add(menu, trayItem{kind: kindItem, sub: sub, glyph: 0xE8F1, text: lang.L.TrayProjects, arrow: true})
	}
	add(menu, trayItem{kind: kindItem, id: idJobs, glyph: 0xE945, text: lang.L.TrayJobs, badge: fmt.Sprint(jobs)})

	// Aktif watch'lar — tıkla → durdur
	t.watchProjects = t.watchProjects[:0]
	if len(watches) > 0 {
		sep(menu)
		for i, w := range watches {
			t.watchProjects = append(t.watchProjects, [2]string{w.Project, w.Server})
			add(menu, trayItem{kind: kindItem, id: uintptr(idWatchBase + i), glyph: 0xE7B3,
				text: w.Project + "  ·  " + w.Server, hint: lang.L.TrayStop, dot: true})
		}
	}

	// Webhook tüneli (CI/CD)
	sep(menu)
	switch {
	case tun.Active && tun.URL == "":
		// cloudflared başladı ama adres henüz gelmedi
		add(menu, trayItem{kind: kindItem, id: idTunnel, glyph: 0xE895, text: lang.L.TrayTunnelStarting, grayed: true})
	case tun.Active && tun.URL != "":
		add(menu, trayItem{kind: kindItem, id: idTunnel, glyph: 0xE774, text: lang.L.TrayTunnelClose, dot: true})
		add(menu, trayItem{kind: kindItem, id: idTunnelCopy, glyph: 0xE8C8, text: lang.L.TrayTunnelCopy})
	case findCloudflared() == "":
		add(menu, trayItem{kind: kindItem, id: idTunnel, glyph: 0xE774, text: lang.L.TrayTunnelNone, grayed: true})
	default:
		add(menu, trayItem{kind: kindItem, id: idTunnel, glyph: 0xE774, text: lang.L.TrayTunnelOpen})
	}
	add(menu, trayItem{kind: kindItem, id: idHelp, glyph: 0xE897, text: lang.L.TrayHelp})
	sep(menu)
	add(menu, trayItem{kind: kindItem, id: idQuit, glyph: 0xE7E8, text: lang.L.TrayQuit, danger: true})

	// Menü açılmadan önce pencere öne alınmalı; yoksa başka yere tıklayınca menü asılı kalır
	pSetForegroundWnd.Call(uintptr(t.hwnd))
	var pt tPoint
	pGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	cmd, _, _ := pTrackPopupMenu.Call(menu, tpmRightAlign|tpmBottomAlign|tpmRightButton|tpmReturnCmd,
		uintptr(pt.X), uintptr(pt.Y), 0, uintptr(t.hwnd), 0)

	switch {
	case cmd == idOpen:
		t.openAt("")
	case cmd == idJobs:
		t.openAt("#/jobs")
	case cmd == idHelp:
		t.openAt("#/help")
	case cmd == idTunnel:
		if tun.Active {
			t.sv.hooks.stopTunnel()
			t.refreshIcon()
			t.notify(lang.L.TrayTunnelClosed, "", false)
		} else {
			t.notify(lang.L.TrayTunnelStarting, "", false)
			go func() {
				u, err := t.sv.hooks.startTunnel()
				t.refreshIcon()
				if err != nil {
					t.notify(lang.L.TrayTunnelFailed, err.Error(), true)
					return
				}
				copyToClipboard(0, u)
				t.notify(lang.L.TrayTunnelOpened, fmt.Sprintf(lang.L.TrayTunnelOpenedFmt, u), false)
			}()
		}
	case cmd == idTunnelCopy:
		copyToClipboard(t.hwnd, tun.URL)
	case cmd == idQuit:
		t.sv.shutdown()
	case cmd >= idProjectBase && int(cmd-idProjectBase) < len(t.projectNames):
		t.openAt("#/p/" + url.PathEscape(t.projectNames[cmd-idProjectBase]))
	case cmd >= idWatchBase && cmd < idProjectBase && int(cmd-idWatchBase) < len(t.watchProjects):
		w := t.watchProjects[cmd-idWatchBase]
		t.sv.watches.stop(w[0], w[1])
		t.sv.persistWatches()
		t.refreshIcon()
	}
}

// copyToClipboard metni panoya koyar (CF_UNICODETEXT).
func copyToClipboard(hwnd windows.HWND, s string) {
	u := windows.StringToUTF16(s)
	if r, _, _ := pOpenClipboard.Call(uintptr(hwnd)); r == 0 {
		return
	}
	defer pCloseClipboard.Call()
	pEmptyClipboard.Call()
	h, _, _ := pGlobalAlloc.Call(0x0002 /*GMEM_MOVEABLE*/, uintptr(len(u)*2))
	if h == 0 {
		return
	}
	p, _, _ := pGlobalLock.Call(h)
	if p == 0 {
		return
	}
	copy(unsafe.Slice((*uint16)(unsafe.Add(nil, p)), len(u)), u)
	pGlobalUnlock.Call(h)
	pSetClipboardData.Call(13 /*CF_UNICODETEXT*/, h)
}

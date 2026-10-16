//go:build windows

package main

import (
	"os"
	"path/filepath"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Panelin renklerinde çizilen koyu tepsi menüsü (owner-draw).
// Her satırı biz çiziyoruz (ikon, metin, sağda rozet/nokta, üzerine gelince yuvarlak
// vurgu); çerçeve ve gölge Windows'un koyu menü temasıyla gelir, Windows 11'in yuvarlak
// köşeleri korunur.

var (
	trGdi32 = windows.NewLazySystemDLL("gdi32.dll")

	pCreateSolidBrush = trGdi32.NewProc("CreateSolidBrush")
	pCreateFont       = trGdi32.NewProc("CreateFontW")
	pDeleteObject     = trGdi32.NewProc("DeleteObject")
	pSelectObject     = trGdi32.NewProc("SelectObject")
	pSetBkMode        = trGdi32.NewProc("SetBkMode")
	pSetTextColor     = trGdi32.NewProc("SetTextColor")
	pRoundRect        = trGdi32.NewProc("RoundRect")
	pEllipse          = trGdi32.NewProc("Ellipse")
	pGetStockObject   = trGdi32.NewProc("GetStockObject")

	pFillRect         = trUser32.NewProc("FillRect")
	pDrawText         = trUser32.NewProc("DrawTextW")
	pDrawIconEx       = trUser32.NewProc("DrawIconEx")
	pSetMenuInfo      = trUser32.NewProc("SetMenuInfo")
	pGetDpiForWindow  = trUser32.NewProc("GetDpiForWindow")
	pSetThreadDpiCtx  = trUser32.NewProc("SetThreadDpiAwarenessContext")
	pRtlGetVersionDLL = windows.NewLazySystemDLL("ntdll.dll")
)

const (
	wmMeasureItem = 0x002C
	wmDrawItem    = 0x002B

	mfOwnerDraw = 0x0100
	mfDisabled  = 0x0002
	odsSelected = 0x0001

	mimStyle      = 0x00000010
	mimBackground = 0x00000002
	mnsNoCheck    = 0x80000000

	dtLeft        = 0x0000
	dtCenter      = 0x0001
	dtRight       = 0x0002
	dtVCenter     = 0x0004
	dtBottom      = 0x0008
	dtSingleLine  = 0x0020
	dtNoPrefix    = 0x0800
	dtEndEllipsis = 0x8000

	bkTransparent = 1
	nullPen       = 8
	diNormal      = 0x0003
	fwNormal      = 400
	fwSemibold    = 600
	clearTypeQ    = 5

	dpiPerMonitor2 = ^uintptr(3) // DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2 (-4)
)

// COLORREF (0x00BBGGRR) — app.css'teki koyu temayla aynı tonlar.
func rgb(r, g, b uint32) uintptr { return uintptr(r | g<<8 | b<<16) }

var (
	colBg       = rgb(0x1b, 0x1c, 0x1f) // --frame
	colHover    = rgb(0x2a, 0x2d, 0x31) // --surface-2
	colHoverRed = rgb(0x3a, 0x1e, 0x22)
	colLine     = rgb(0x2e, 0x31, 0x35) // --line
	colText     = rgb(0xed, 0xee, 0xf0) // --text
	colStrong   = rgb(0xff, 0xff, 0xff)
	colDim      = rgb(0xb0, 0xb4, 0xba) // --muted
	colFaint    = rgb(0x77, 0x7b, 0x84) // --faint
	colAccent   = rgb(0x3b, 0x9e, 0xff) // --accent
	colRed      = rgb(0xff, 0x63, 0x69) // --bad
	colGreen    = rgb(0x3d, 0xd6, 0x8c) // --ok
	colPill     = rgb(0x2e, 0x31, 0x35)
	colPillHot  = rgb(0x1e, 0x3a, 0x5c)
)

type trayKind int

const (
	kindHeader trayKind = iota
	kindItem
	kindSep
)

// trayItem çizilecek bir satır.
type trayItem struct {
	kind    trayKind
	id      uintptr
	sub     uintptr // alt menü tutamacı (Projeler)
	glyph   rune    // Segoe Fluent Icons / MDL2 kod noktası
	text    string
	subText string // yalnızca başlık: alt satır
	badge   string // sağda sayı rozeti
	hint    string // sağda soluk ipucu (üzerine gelince görünür)
	dot     bool   // sağda yeşil nokta
	danger  bool   // Çıkış: kırmızı
	grayed  bool
	arrow   bool // alt menü oku
}

type tRect struct{ Left, Top, Right, Bottom int32 }

type tMeasureItem struct {
	CtlType    uint32
	CtlID      uint32
	ItemID     uint32
	ItemWidth  uint32
	ItemHeight uint32
	ItemData   uintptr
}

type tDrawItem struct {
	CtlType    uint32
	CtlID      uint32
	ItemID     uint32
	ItemAction uint32
	ItemState  uint32
	HwndItem   uintptr
	HDC        uintptr
	RcItem     tRect
	ItemData   uintptr
}

type tMenuInfo struct {
	CbSize          uint32
	FMask           uint32
	DwStyle         uint32
	CyMax           uint32
	HbrBack         uintptr
	DwContextHelpID uint32
	DwMenuData      uintptr
}

type trayFonts struct{ text, title, small, glyph uintptr }

func newFont(px int32, weight int32, face string) uintptr {
	p, _ := windows.UTF16PtrFromString(face)
	h, _, _ := pCreateFont.Call(uintptr(-px), 0, 0, 0, uintptr(weight), 0, 0, 0,
		1 /*DEFAULT_CHARSET*/, 0, 0, clearTypeQ, 0, uintptr(unsafe.Pointer(p)))
	return h
}

// glyphFace Windows 11'de Segoe Fluent Icons, 10'da Segoe MDL2 Assets (kod noktaları aynı).
func glyphFace() string {
	if _, err := os.Stat(filepath.Join(os.Getenv("WINDIR"), "Fonts", "SegoeIcons.ttf")); err == nil {
		return "Segoe Fluent Icons"
	}
	return "Segoe MDL2 Assets"
}

// enableDarkMenus menü çerçevesini/gölgesini koyu temaya alır (Explorer'ın yolu):
// uxtheme #135 SetPreferredAppMode(ForceDark), #136 FlushMenuThemes. 1809 öncesi: atlanır.
func enableDarkMenus() {
	if v := windows.RtlGetVersion(); v == nil || v.BuildNumber < 17763 {
		return
	}
	h, err := windows.LoadLibrary("uxtheme.dll")
	if err != nil {
		return
	}
	if p, err := windows.GetProcAddressByOrdinal(h, 135); err == nil {
		syscall.SyscallN(p, 2)
	}
	if p, err := windows.GetProcAddressByOrdinal(h, 136); err == nil {
		syscall.SyscallN(p)
	}
}

func (t *winTray) scale(v int32) int32 {
	dpi := uintptr(96)
	if pGetDpiForWindow.Find() == nil {
		if d, _, _ := pGetDpiForWindow.Call(uintptr(t.hwnd)); d != 0 {
			dpi = d
		}
	}
	return int32((int64(v)*int64(dpi) + 48) / 96)
}

// Ölçüler (96 DPI piksel).
const (
	trayMenuWidth  = 280
	trayItemHeight = 36
	trayHeadHeight = 60
	traySepHeight  = 9
)

func (t *winTray) measure(mi *tMeasureItem) {
	i := int(mi.ItemData) - 1
	if i < 0 || i >= len(t.items) {
		return
	}
	mi.ItemWidth = uint32(t.scale(trayMenuWidth))
	switch t.items[i].kind {
	case kindHeader:
		mi.ItemHeight = uint32(t.scale(trayHeadHeight))
	case kindSep:
		mi.ItemHeight = uint32(t.scale(traySepHeight))
	default:
		mi.ItemHeight = uint32(t.scale(trayItemHeight))
	}
}

func gFill(hdc uintptr, r tRect, color uintptr) {
	b, _, _ := pCreateSolidBrush.Call(color)
	pFillRect.Call(hdc, uintptr(unsafe.Pointer(&r)), b)
	pDeleteObject.Call(b)
}

func gRound(hdc uintptr, r tRect, radius int32, color uintptr) {
	b, _, _ := pCreateSolidBrush.Call(color)
	pen, _, _ := pGetStockObject.Call(nullPen)
	ob, _, _ := pSelectObject.Call(hdc, b)
	op, _, _ := pSelectObject.Call(hdc, pen)
	// Boş kalemle RoundRect bir piksel eksik çizer; sağ ve alt telafi edilir.
	pRoundRect.Call(hdc, uintptr(r.Left), uintptr(r.Top), uintptr(r.Right+1), uintptr(r.Bottom+1), uintptr(radius), uintptr(radius))
	pSelectObject.Call(hdc, ob)
	pSelectObject.Call(hdc, op)
	pDeleteObject.Call(b)
}

func gDot(hdc uintptr, r tRect, color uintptr) {
	b, _, _ := pCreateSolidBrush.Call(color)
	pen, _, _ := pGetStockObject.Call(nullPen)
	ob, _, _ := pSelectObject.Call(hdc, b)
	op, _, _ := pSelectObject.Call(hdc, pen)
	pEllipse.Call(hdc, uintptr(r.Left), uintptr(r.Top), uintptr(r.Right+1), uintptr(r.Bottom+1))
	pSelectObject.Call(hdc, ob)
	pSelectObject.Call(hdc, op)
	pDeleteObject.Call(b)
}

func gText(hdc uintptr, s string, r tRect, font, color uintptr, flags uintptr) {
	p, err := windows.UTF16FromString(s)
	if err != nil {
		return
	}
	of, _, _ := pSelectObject.Call(hdc, font)
	pSetTextColor.Call(hdc, color)
	pDrawText.Call(hdc, uintptr(unsafe.Pointer(&p[0])), uintptr(len(p)-1), uintptr(unsafe.Pointer(&r)), flags|dtSingleLine|dtNoPrefix)
	pSelectObject.Call(hdc, of)
}

func (t *winTray) draw(di *tDrawItem) {
	i := int(di.ItemData) - 1
	if i < 0 || i >= len(t.items) {
		return
	}
	it := t.items[i]
	hdc, r := di.HDC, di.RcItem
	s := t.scale
	pSetBkMode.Call(hdc, bkTransparent)
	gFill(hdc, r, colBg)

	switch it.kind {
	case kindSep:
		mid := (r.Top + r.Bottom) / 2
		gFill(hdc, tRect{r.Left + s(12), mid, r.Right - s(12), mid + max32(1, s(1))}, colLine)
		return

	case kindHeader:
		pad, ic := s(14), s(30)
		top := r.Top + (r.Bottom-r.Top-ic)/2
		if t.bigIcon != 0 {
			pDrawIconEx.Call(hdc, uintptr(r.Left+pad), uintptr(top), uintptr(t.bigIcon), uintptr(ic), uintptr(ic), 0, 0, diNormal)
		}
		x := r.Left + pad + ic + s(12)
		mid := (r.Top + r.Bottom) / 2
		gText(hdc, it.text, tRect{x, r.Top, r.Right - pad, mid + s(1)}, t.fonts.title, colStrong, dtLeft|dtEndEllipsis|dtBottom)
		sx := x
		if it.dot {
			d := s(6)
			cy := mid + s(3) + s(8)
			gDot(hdc, tRect{sx, cy - d/2, sx + d, cy + d/2}, colGreen)
			sx += d + s(6)
		}
		gText(hdc, it.subText, tRect{sx, mid + s(3), r.Right - pad, mid + s(19)}, t.fonts.small, colDim, dtLeft|dtEndEllipsis)
		gFill(hdc, tRect{r.Left + s(12), r.Bottom - max32(1, s(1)), r.Right - s(12), r.Bottom}, colLine)
		return
	}

	hot := di.ItemState&odsSelected != 0 && !it.grayed
	if hot {
		bg := colHover
		if it.danger {
			bg = colHoverRed
		}
		gRound(hdc, tRect{r.Left + s(6), r.Top + s(2), r.Right - s(6), r.Bottom - s(2)}, s(8), bg)
	}
	fg, icon := colText, colDim
	switch {
	case it.grayed:
		fg, icon = colFaint, colFaint
	case it.danger:
		fg, icon = colRed, colRed // Çıkış her zaman kırmızı
		if hot {
			fg = colStrong
		}
	case hot:
		fg, icon = colStrong, colAccent
	}

	left := r.Left + s(18)
	if it.glyph != 0 {
		gText(hdc, string(it.glyph), tRect{left, r.Top, left + s(18), r.Bottom}, t.fonts.glyph, icon, dtCenter|dtVCenter)
	}
	right := r.Right - s(18)
	if it.arrow {
		right -= s(14) // Windows alt menü okunu bu boşluğa çizer
	}
	switch {
	case it.badge != "":
		w := s(10) + s(8)*int32(len(it.badge))
		h := s(20)
		top := r.Top + (r.Bottom-r.Top-h)/2
		pill := tRect{right - w, top, right, top + h}
		pc := colPill
		if hot {
			pc = colPillHot
		}
		gRound(hdc, pill, h, pc)
		gText(hdc, it.badge, pill, t.fonts.small, colStrong, dtCenter|dtVCenter)
		right = pill.Left - s(8)
	case hot && it.hint != "":
		gText(hdc, it.hint, tRect{right - s(80), r.Top, right, r.Bottom}, t.fonts.small, colRed, dtRight|dtVCenter)
		right -= s(84)
	case it.dot:
		d := s(8)
		cy := (r.Top + r.Bottom) / 2
		gDot(hdc, tRect{right - d, cy - d/2, right, cy + d/2}, colGreen)
		right -= d + s(8)
	}
	gText(hdc, it.text, tRect{left + s(28), r.Top, right, r.Bottom}, t.fonts.text, fg, dtLeft|dtVCenter|dtEndEllipsis)
}

// styleMenu menünün kendi arka planını (kenar payları) boyar ve onay işareti sütununu kaldırır.
func styleMenu(menu uintptr, bg uintptr) {
	mi := tMenuInfo{FMask: mimStyle | mimBackground, DwStyle: mnsNoCheck, HbrBack: bg}
	mi.CbSize = uint32(unsafe.Sizeof(mi))
	pSetMenuInfo.Call(menu, uintptr(unsafe.Pointer(&mi)))
}

var _ = pRtlGetVersionDLL

// max32 — pakette int tipli bir max (cmd_browser.go) yerleşik max'i gölgelediği için.
func max32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

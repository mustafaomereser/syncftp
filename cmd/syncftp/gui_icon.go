package main

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
)

// Projenin logosu (repo kökündeki logo.ico'nun kopyası — go:embed üst dizini göremez).
// İçinde 256/64/32/16 px PNG sıkıştırmalı görüntüler var.
//
//go:embed logo.ico
var logoICO []byte

// trayIconICO tray ikonu: boşta logo.ico'nun kendisi; active=true (iş/watch çalışıyor) →
// logonun 32 px hali + sağ altta yeşil nokta (32 bpp BMP ICO olarak yeniden kodlanır).
func trayIconICO(active bool) []byte {
	if !active {
		return logoICO
	}
	base := icoImage(logoICO, 32)
	if base == nil {
		return logoICO
	}
	const n = 32
	img := image.NewNRGBA(image.Rect(0, 0, n, n))
	draw.Draw(img, img.Bounds(), base, base.Bounds().Min, draw.Src)

	// Yeşil durum noktası (koyu halkalı, kenar yumuşatmalı)
	green := color.NRGBA{0x3d, 0xd6, 0x8c, 0xff}
	ring := color.NRGBA{0x0b, 0x0b, 0x0d, 0xff}
	const cx, cy = 24.5, 24.5
	for y := 16; y < n; y++ {
		for x := 16; x < n; x++ {
			d := math.Hypot(float64(x)+0.5-cx, float64(y)+0.5-cy)
			switch {
			case d <= 5.2:
				img.SetNRGBA(x, y, green)
			case d <= 7.0:
				img.SetNRGBA(x, y, ring)
			case d <= 7.8:
				a := uint8(255 * (7.8 - d) / 0.8)
				blendPixel(img, x, y, color.NRGBA{ring.R, ring.G, ring.B, a})
			}
		}
	}
	return encodeICO32(img)
}

// icoImage ICO içinden istenen boyuttaki (yoksa en yakın) PNG görüntüyü çözer.
func icoImage(ico []byte, want int) image.Image {
	if len(ico) < 6 {
		return nil
	}
	count := int(binary.LittleEndian.Uint16(ico[4:6]))
	best, bestDiff := -1, 1<<30
	for i := 0; i < count; i++ {
		e := ico[6+16*i:]
		if len(e) < 16 {
			break
		}
		w := int(e[0])
		if w == 0 {
			w = 256
		}
		if d := abs(w - want); d < bestDiff {
			best, bestDiff = i, d
		}
	}
	if best < 0 {
		return nil
	}
	e := ico[6+16*best:]
	size := binary.LittleEndian.Uint32(e[8:12])
	off := binary.LittleEndian.Uint32(e[12:16])
	if int(off+size) > len(ico) {
		return nil
	}
	img, err := png.Decode(bytes.NewReader(ico[off : off+size]))
	if err != nil {
		return nil
	}
	return img
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func blendPixel(img *image.NRGBA, x, y int, c color.NRGBA) {
	o := img.NRGBAAt(x, y)
	a := float64(c.A) / 255
	mix := func(f, b uint8) uint8 { return uint8(float64(f)*a + float64(b)*(1-a)) }
	img.SetNRGBA(x, y, color.NRGBA{mix(c.R, o.R), mix(c.G, o.G), mix(c.B, o.B), uint8(math.Max(float64(o.A), float64(c.A)))})
}

// encodeICO32 NRGBA görüntüyü tek görüntülü 32 bpp BMP ICO'ya çevirir.
func encodeICO32(img *image.NRGBA) []byte {
	n := img.Bounds().Dx()
	var bmp bytes.Buffer
	le := binary.LittleEndian
	hdr := struct {
		Size                                uint32
		Width, Height                       int32
		Planes, BitCount                    uint16
		Compression, SizeImage              uint32
		XPels, YPels, ClrUsed, ClrImportant int32
	}{40, int32(n), int32(n * 2), 1, 32, 0, 0, 0, 0, 0, 0}
	binary.Write(&bmp, le, hdr)
	for y := n - 1; y >= 0; y-- { // alttan üste
		for x := 0; x < n; x++ {
			c := img.NRGBAAt(x, y)
			bmp.Write([]byte{c.B, c.G, c.R, c.A})
		}
	}
	bmp.Write(make([]byte, n*n/8)) // AND maskesi (alfa kullanılır)

	var ico bytes.Buffer
	binary.Write(&ico, le, []uint16{0, 1, 1})
	ico.Write([]byte{byte(n), byte(n), 0, 0})
	binary.Write(&ico, le, []uint16{1, 32})
	binary.Write(&ico, le, []uint32{uint32(bmp.Len()), 6 + 16})
	ico.Write(bmp.Bytes())
	return ico.Bytes()
}

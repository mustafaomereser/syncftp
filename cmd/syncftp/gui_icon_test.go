package main

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestTrayIcons(t *testing.T) {
	idle := trayIconICO(false)
	if !bytes.Equal(idle, logoICO) || len(logoICO) < 1000 {
		t.Fatal("idle icon must be the embedded logo.ico")
	}
	if icoImage(logoICO, 32) == nil {
		t.Fatal("logo.ico has no decodable 32px image")
	}
	act := trayIconICO(true)
	if binary.LittleEndian.Uint16(act[2:4]) != 1 || binary.LittleEndian.Uint16(act[4:6]) != 1 || act[6] != 32 {
		t.Fatalf("active icon header wrong: % x", act[:8])
	}
	if want := 6 + 16 + 40 + 32*32*4 + 32*32/8; len(act) != want {
		t.Fatalf("active icon size %d, want %d", len(act), want)
	}
}

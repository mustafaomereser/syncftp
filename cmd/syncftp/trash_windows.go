//go:build windows

package main

import (
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

var procSHFileOperationW = windows.NewLazySystemDLL("shell32.dll").NewProc("SHFileOperationW")

// shFileOpStruct SHFILEOPSTRUCTW (64-bit'te doğal hizalama — Go düzeniyle aynı).
type shFileOpStruct struct {
	hwnd          uintptr
	wFunc         uint32
	pFrom         *uint16
	pTo           *uint16
	fFlags        uint16
	fAborted      int32
	hNameMappings uintptr
	progressTitle *uint16
}

// moveToTrash dosya/klasörü Geri Dönüşüm Kutusu'na taşır (oradan geri yüklenebilir).
func moveToTrash(path string) error {
	const (
		foDelete         = 3
		fofSilent        = 0x4
		fofNoConfirm     = 0x10
		fofAllowUndo     = 0x40
		fofNoErrorUI     = 0x400
		fofNoConfirmMkdr = 0x200
	)
	from, err := windows.UTF16FromString(path)
	if err != nil {
		return err
	}
	from = append(from, 0) // pFrom çift NUL ile biter
	op := shFileOpStruct{wFunc: foDelete, pFrom: &from[0], fFlags: fofSilent | fofNoConfirm | fofAllowUndo | fofNoErrorUI | fofNoConfirmMkdr}
	r, _, _ := procSHFileOperationW.Call(uintptr(unsafe.Pointer(&op)))
	if r != 0 || op.fAborted != 0 {
		return fmt.Errorf("SHFileOperation: 0x%x", r)
	}
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("SHFileOperation: %s hâlâ yerinde", path)
	}
	return nil
}

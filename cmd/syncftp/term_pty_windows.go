//go:build windows

package main

import (
	"errors"
	"os"
	"strings"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ConPTY tabanlı sözde terminal (Windows 10 1809+). Panel içi terminal (cmd_gui_term.go) syncftp
// shell'ini bunun içinde çalıştırır: readline / bubbletea TUI'ları gerçek bir konsol görür.
type winPTY struct {
	hpc     windows.Handle
	proc    windows.Handle
	in      *os.File // çocuğun stdin'ine yazılır
	out     *os.File // çocuğun çıktısı okunur
	closeMu sync.Once
	procMu  sync.Mutex // proc tutamacı: Close kapatırken Wait aynı tutamacı beklemesin
}

func startTermPTY(exe string, args []string, dir string, env []string, cols, rows int) (termPTY, error) {
	var inR, inW, outR, outW windows.Handle
	if err := windows.CreatePipe(&inR, &inW, nil, 0); err != nil {
		return nil, err
	}
	if err := windows.CreatePipe(&outR, &outW, nil, 0); err != nil {
		windows.CloseHandle(inR)
		windows.CloseHandle(inW)
		return nil, err
	}
	var hpc windows.Handle
	if err := windows.CreatePseudoConsole(windows.Coord{X: int16(cols), Y: int16(rows)}, inR, outW, 0, &hpc); err != nil {
		for _, h := range []windows.Handle{inR, inW, outR, outW} {
			windows.CloseHandle(h)
		}
		return nil, err
	}
	// Bu uçları ConPTY kopyaladı; bizde kalırsa çocuk çıkınca okuma EOF almaz
	windows.CloseHandle(inR)
	windows.CloseHandle(outW)

	p := &winPTY{hpc: hpc, in: os.NewFile(uintptr(inW), "pty-in"), out: os.NewFile(uintptr(outR), "pty-out")}

	attrs, err := windows.NewProcThreadAttributeList(1)
	if err != nil {
		p.Close()
		return nil, err
	}
	defer attrs.Delete()
	// HPCON değerin KENDİSİ öznitelik değeridir (işaretçisi değil)
	if err := attrs.Update(windows.PROC_THREAD_ATTRIBUTE_PSEUDOCONSOLE, *(*unsafe.Pointer)(unsafe.Pointer(&hpc)), unsafe.Sizeof(hpc)); err != nil {
		p.Close()
		return nil, err
	}
	si := windows.StartupInfoEx{ProcThreadAttributeList: attrs.List()}
	si.Cb = uint32(unsafe.Sizeof(si))
	// Std tutamaçları BOŞ olarak açıkça ver: yoksa çocuk ebeveynin konsol/boru tutamaçlarını devralıp
	// ConPTY yerine onlara yazıyor (panel gizli konsollu olduğu için shell anında kapanıyordu).
	si.Flags |= windows.STARTF_USESTDHANDLES

	cmdLine, err := windows.UTF16PtrFromString(windows.ComposeCommandLine(append([]string{exe}, args...)))
	if err != nil {
		p.Close()
		return nil, err
	}
	var dirp *uint16
	if dir != "" {
		if dirp, err = windows.UTF16PtrFromString(dir); err != nil {
			p.Close()
			return nil, err
		}
	}
	envBlock := envBlockUTF16(env)
	var pi windows.ProcessInformation
	err = windows.CreateProcess(nil, cmdLine, nil, nil, false,
		windows.EXTENDED_STARTUPINFO_PRESENT|windows.CREATE_UNICODE_ENVIRONMENT,
		&envBlock[0], dirp, &si.StartupInfo, &pi)
	if err != nil {
		p.Close()
		return nil, err
	}
	windows.CloseHandle(pi.Thread)
	p.proc = pi.Process
	return p, nil
}

// envBlockUTF16 "K=V\0K=V\0\0" biçiminde ortam bloğu.
func envBlockUTF16(env []string) []uint16 {
	var b []uint16
	for _, kv := range env {
		if strings.ContainsRune(kv, 0) {
			continue
		}
		b = append(b, windows.StringToUTF16(kv)...) // sonundaki \0 dahil
	}
	if len(b) == 0 {
		b = append(b, 0)
	}
	return append(b, 0)
}

func (p *winPTY) Read(b []byte) (int, error)  { return p.out.Read(b) }
func (p *winPTY) Write(b []byte) (int, error) { return p.in.Write(b) }

func (p *winPTY) Resize(cols, rows int) error {
	if cols <= 0 || rows <= 0 {
		return errors.New("size")
	}
	return windows.ResizePseudoConsole(p.hpc, windows.Coord{X: int16(cols), Y: int16(rows)})
}

// Wait çocuk süreç bitene kadar bekler.
// Tutamacın kopyasını bekler: Close asıl tutamacı (CloseHandle) beklerken kapatabilir.
func (p *winPTY) Wait() error {
	p.procMu.Lock()
	if p.proc == 0 {
		p.procMu.Unlock()
		return errors.New("not started")
	}
	var dup windows.Handle
	self := windows.CurrentProcess()
	err := windows.DuplicateHandle(self, p.proc, self, &dup, 0, false, windows.DUPLICATE_SAME_ACCESS)
	p.procMu.Unlock()
	if err != nil {
		return err
	}
	defer windows.CloseHandle(dup)
	_, err = windows.WaitForSingleObject(dup, windows.INFINITE)
	return err
}

// Close süreci sonlandırır ve konsolu kapatır. ClosePseudoConsole çıktı borusu boşalana kadar
// bloklayabildiği için okuma ucu en son kapatılır.
func (p *winPTY) Close() error {
	p.closeMu.Do(func() {
		p.procMu.Lock()
		if p.proc != 0 {
			_ = windows.TerminateProcess(p.proc, 0)
		}
		p.procMu.Unlock()
		_ = p.in.Close()
		done := make(chan struct{})
		go func() { windows.ClosePseudoConsole(p.hpc); close(done) }()
		go func() {
			buf := make([]byte, 4096)
			for {
				if _, err := p.out.Read(buf); err != nil {
					return
				}
			}
		}()
		<-done
		_ = p.out.Close()
		p.procMu.Lock()
		if p.proc != 0 {
			windows.CloseHandle(p.proc)
			p.proc = 0
		}
		p.procMu.Unlock()
	})
	return nil
}

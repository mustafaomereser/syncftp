//go:build !windows

package main

import (
	"os"
	"os/exec"

	"github.com/creack/pty"
)

// Unix sözde terminali (creack/pty) — panel içi terminal için.
type unixPTY struct {
	f   *os.File
	cmd *exec.Cmd
}

func startTermPTY(exe string, args []string, dir string, env []string, cols, rows int) (termPTY, error) {
	cmd := exec.Command(exe, args...)
	cmd.Dir = dir
	cmd.Env = env
	f, err := pty.StartWithSize(cmd, &pty.Winsize{Cols: uint16(cols), Rows: uint16(rows)})
	if err != nil {
		return nil, err
	}
	return &unixPTY{f: f, cmd: cmd}, nil
}

func (p *unixPTY) Read(b []byte) (int, error)  { return p.f.Read(b) }
func (p *unixPTY) Write(b []byte) (int, error) { return p.f.Write(b) }
func (p *unixPTY) Resize(cols, rows int) error {
	return pty.Setsize(p.f, &pty.Winsize{Cols: uint16(cols), Rows: uint16(rows)})
}
func (p *unixPTY) Wait() error { return p.cmd.Wait() }
func (p *unixPTY) Close() error {
	if p.cmd.Process != nil {
		_ = p.cmd.Process.Kill()
	}
	return p.f.Close()
}

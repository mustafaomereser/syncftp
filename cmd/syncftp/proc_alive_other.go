//go:build !windows

package main

import (
	"os"
	"syscall"
)

// processAlive PID'li süreç çalışıyor mu (watch kilidi).
func processAlive(pid int) bool {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return p.Signal(syscall.Signal(0)) == nil
}

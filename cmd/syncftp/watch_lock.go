package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"syncftp/internal/lang"
)

// Watch kilidi: aynı projeyi iki süreç (iki panel, panel + CLI `watch`) aynı anda izlerse her
// değişiklik iki kez yüklenir ve yarışır. watchLoop başlarken .syncftp/watch.lock'a kendi PID'ini
// yazar; dosyadaki PID canlı başka bir süreçse reddeder. Ölü süreçten kalan kilit devralınır.

func watchLockPath(dir string) string { return filepath.Join(dir, ".syncftp", "watch.lock") }

// watchLockOwner kilidi tutan canlı başka bir sürecin PID'i (yoksa 0).
func watchLockOwner(dir string) int {
	data, err := os.ReadFile(watchLockPath(dir))
	if err != nil {
		return 0
	}
	pid, _ := strconv.Atoi(strings.TrimSpace(string(data)))
	if pid <= 0 || pid == os.Getpid() || !processAlive(pid) {
		return 0
	}
	return pid
}

func acquireWatchLock(dir string) (release func(), err error) {
	if pid := watchLockOwner(dir); pid != 0 {
		return nil, fmt.Errorf(lang.L.WatchLockedFmt, pid)
	}
	p := watchLockPath(dir)
	_ = os.MkdirAll(filepath.Dir(p), 0700)
	if err := os.WriteFile(p, []byte(strconv.Itoa(os.Getpid())), 0600); err != nil {
		return nil, err
	}
	return func() {
		if data, err := os.ReadFile(p); err == nil && strings.TrimSpace(string(data)) == strconv.Itoa(os.Getpid()) {
			_ = os.Remove(p)
		}
	}, nil
}

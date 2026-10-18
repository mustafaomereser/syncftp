package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"syncftp/internal/lang"
)

// Watch kilidi: aynı sunucuyu iki süreç (iki panel, panel + CLI `watch`) aynı anda izlerse her
// değişiklik iki kez yüklenir ve yarışır. watchLoop başlarken her sunucu için
// .syncftp/locks/watch-<sunucu>.lock'a kendi PID'ini yazar; dosyadaki PID canlı başka bir süreçse
// reddeder. Ölü süreçten kalan kilit devralınır. Farklı sunucular ayrı süreçlerde izlenebilir.

func watchLockPath(dir, server string) string {
	safe := strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || r == ':' || r == '*' || r == '?' || r == '"' || r == '<' || r == '>' || r == '|' {
			return '_'
		}
		return r
	}, server)
	return filepath.Join(dir, ".syncftp", "locks", "watch-"+safe+".lock")
}

// watchLockOwner sunucunun kilidini tutan canlı başka bir sürecin PID'i (yoksa 0).
func watchLockOwner(dir, server string) int {
	data, err := os.ReadFile(watchLockPath(dir, server))
	if err != nil {
		return 0
	}
	pid, _ := strconv.Atoi(strings.TrimSpace(string(data)))
	if pid <= 0 || pid == os.Getpid() || !processAlive(pid) {
		return 0
	}
	return pid
}

// acquireWatchLock verilen sunucuların hepsini kilitler; biri başka süreçteyse hiçbirini almaz.
func acquireWatchLock(dir string, servers []string) (release func(), err error) {
	for _, s := range servers {
		if pid := watchLockOwner(dir, s); pid != 0 {
			return nil, fmt.Errorf("%s: "+lang.L.WatchLockedFmt, s, pid)
		}
	}
	me := strconv.Itoa(os.Getpid())
	// Proje başına tek kilit kullanan önceki sürümün dosyası (sahibi ölmüşse) temizlenir
	if old := filepath.Join(dir, ".syncftp", "watch.lock"); fileExistsPlain(old) {
		if data, err := os.ReadFile(old); err == nil {
			if pid, _ := strconv.Atoi(strings.TrimSpace(string(data))); pid <= 0 || !processAlive(pid) {
				_ = os.Remove(old)
			}
		}
	}
	var held []string
	release = func() {
		for _, p := range held {
			if data, err := os.ReadFile(p); err == nil && strings.TrimSpace(string(data)) == me {
				_ = os.Remove(p)
			}
		}
	}
	for _, s := range servers {
		p := watchLockPath(dir, s)
		_ = os.MkdirAll(filepath.Dir(p), 0700)
		if err := os.WriteFile(p, []byte(me), 0600); err != nil {
			release()
			return nil, err
		}
		held = append(held, p)
	}
	return release, nil
}

package ftp

import (
	"net"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// ioTimeout — tek bir okuma/yazma işleminin en fazla bekleyebileceği süre.
// jlaffaye/ftp yalnızca dial'a timeout koyar; takılan bir pasif veri bağlantısı
// (NAT, hosting kaprisi) Stor/Retr'yi sonsuza kadar bekletir → watch/sync kilitlenir.
const ioTimeout = 90 * time.Second

// deadlineConn her Read/Write öncesi deadline'ı yeniler (idle timeout).
// FTP istek-cevap protokolü olduğundan boşta bekleyen kontrol bağlantısında
// bekleyen Read olmaz → shell'de idle kalmak bağlantıyı düşürmez.
type deadlineConn struct {
	net.Conn
	timeout time.Duration
}

func (c *deadlineConn) Read(p []byte) (int, error) {
	_ = c.Conn.SetReadDeadline(time.Now().Add(c.timeout))
	return c.Conn.Read(p)
}

func (c *deadlineConn) Write(p []byte) (int, error) {
	_ = c.Conn.SetWriteDeadline(time.Now().Add(c.timeout))
	return c.Conn.Write(p)
}

func withDeadline(conn net.Conn, err error) (net.Conn, error) {
	if err != nil {
		return nil, err
	}
	return &deadlineConn{Conn: conn, timeout: ioTimeout}, nil
}

// writeDeadlineConn yalnızca yazmaya deadline koyar. SSH kütüphanesinin arka plan
// okuma döngüsü sürekli Read'de beklediğinden okuma deadline'ı idle bağlantıyı öldürür;
// ölü bağlantı tespiti sshKeepalive ile yapılır.
type writeDeadlineConn struct{ net.Conn }

func (c *writeDeadlineConn) Write(p []byte) (int, error) {
	_ = c.Conn.SetWriteDeadline(time.Now().Add(ioTimeout))
	return c.Conn.Write(p)
}

// sshKeepalive her 30 sn'de keepalive isteği yollar; cevap 30 sn içinde gelmezse
// bağlantıyı kapatır → bekleyen SFTP işlemleri hata ile döner (kilitlenme yerine).
func sshKeepalive(c *ssh.Client) {
	go func() {
		t := time.NewTicker(30 * time.Second)
		defer t.Stop()
		for range t.C {
			done := make(chan error, 1)
			go func() {
				_, _, err := c.SendRequest("keepalive@openssh.com", true, nil)
				done <- err
			}()
			select {
			case err := <-done:
				if err != nil {
					return // bağlantı kapanmış
				}
			case <-time.After(30 * time.Second):
				_ = c.Close()
				return
			}
		}
	}()
}

// connSet bir FTP oturumunun açtığı ağ bağlantıları (kontrol + pasif veri) — Abort kilitsiz
// kesebilsin diye. Kapanan bağlantı kendini çıkarır (her LIST/RETR yeni veri bağlantısı açar).
type connSet struct {
	mu      sync.Mutex
	conns   map[*trackedConn]struct{}
	aborted bool
}

type trackedConn struct {
	net.Conn
	set *connSet
}

func (t *trackedConn) Close() error {
	t.set.mu.Lock()
	delete(t.set.conns, t)
	t.set.mu.Unlock()
	return t.Conn.Close()
}

func (s *connSet) track(conn net.Conn, err error) (net.Conn, error) {
	if err != nil {
		return nil, err
	}
	t := &trackedConn{Conn: conn, set: s}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.aborted {
		_ = conn.Close()
		return nil, net.ErrClosed
	}
	if s.conns == nil {
		s.conns = map[*trackedConn]struct{}{}
	}
	s.conns[t] = struct{}{}
	return t, nil
}

func (s *connSet) closeAll() {
	s.mu.Lock()
	s.aborted = true
	list := make([]net.Conn, 0, len(s.conns))
	for t := range s.conns {
		list = append(list, t.Conn)
	}
	s.conns = nil
	s.mu.Unlock()
	for _, c := range list {
		_ = c.Close()
	}
}

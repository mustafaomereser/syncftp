package ftp

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/binary"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"syncftp/internal/config"
)

// startTestSSH bellek üstü SFTP + direct-tcpip (tünel) destekli SSH sunucusu başlatır.
// Her bağlantı aynı in-memory dosya sistemini görür.
func startTestSSH(t *testing.T, user, pass string) (host string, port int, stop func()) {
	t.Helper()
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	signer, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	cfg := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, p []byte) (*ssh.Permissions, error) {
			if c.User() == user && string(p) == pass {
				return nil, nil
			}
			return nil, os.ErrPermission
		},
	}
	cfg.AddHostKey(signer)
	handlers := sftp.InMemHandler()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	go func() {
		for {
			nc, err := ln.Accept()
			if err != nil {
				return
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				serveSSHConn(nc, cfg, handlers)
			}()
		}
	}()
	addr := ln.Addr().(*net.TCPAddr)
	return "127.0.0.1", addr.Port, func() { ln.Close() }
}

func serveSSHConn(nc net.Conn, cfg *ssh.ServerConfig, h sftp.Handlers) {
	sc, chans, reqs, err := ssh.NewServerConn(nc, cfg)
	if err != nil {
		nc.Close()
		return
	}
	defer sc.Close()
	go ssh.DiscardRequests(reqs)
	for nch := range chans {
		switch nch.ChannelType() {
		case "session":
			ch, creqs, err := nch.Accept()
			if err != nil {
				continue
			}
			go func() {
				for req := range creqs {
					ok := req.Type == "subsystem" && len(req.Payload) > 4 && string(req.Payload[4:]) == "sftp"
					req.Reply(ok, nil)
					if ok {
						srv := sftp.NewRequestServer(ch, h)
						go func() { _ = srv.Serve(); srv.Close() }()
					}
				}
			}()
		case "direct-tcpip":
			// payload: host string, port uint32, origin string, origin port uint32
			p := nch.ExtraData()
			hl := binary.BigEndian.Uint32(p[:4])
			host := string(p[4 : 4+hl])
			port := binary.BigEndian.Uint32(p[4+hl : 8+hl])
			target, err := net.Dial("tcp", net.JoinHostPort(host, strconv.Itoa(int(port))))
			if err != nil {
				nch.Reject(ssh.ConnectionFailed, err.Error())
				continue
			}
			ch, creqs, err := nch.Accept()
			if err != nil {
				target.Close()
				continue
			}
			go ssh.DiscardRequests(creqs)
			go func() { io.Copy(ch, target); ch.CloseWrite() }()
			go func() { io.Copy(target, ch); target.Close() }()
		default:
			nch.Reject(ssh.UnknownChannelType, "unsupported")
		}
	}
}

func withTempKnownHosts(t *testing.T) {
	t.Helper()
	old := knownHostsOverride
	knownHostsOverride = filepath.Join(t.TempDir(), "known_hosts")
	t.Cleanup(func() { knownHostsOverride = old })
}

func writeLocal(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, filepath.FromSlash(name))
	os.MkdirAll(filepath.Dir(p), 0755)
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestSFTPRoundTrip(t *testing.T) {
	withTempKnownHosts(t)
	host, port, stop := startTestSSH(t, "u", "p")
	defer stop()

	srv := config.Server{Name: "s", Protocol: "sftp", Host: host, Port: port, User: "u", Password: "p", RemotePath: "/site"}
	c, err := Connect(srv)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer c.Close()

	local := t.TempDir()
	f1 := writeLocal(t, local, "a/index.php", "line1\r\nline2\r\n")

	// İlk yükleme: dosya yok → BackupTo existed=false
	existed, err := c.BackupTo("a/index.php", filepath.Join(t.TempDir(), "bk"))
	if err != nil || existed {
		t.Fatalf("backup of missing file: existed=%v err=%v", existed, err)
	}
	if err := c.Upload(f1, "a/index.php"); err != nil {
		t.Fatalf("upload: %v", err)
	}
	// SFTP binary-safe: CRLF korunur, bayt bayt aynı
	got, err := c.Preview("/site/a/index.php", 1024)
	if err != nil || string(got) != "line1\r\nline2\r\n" {
		t.Fatalf("preview = %q, %v", got, err)
	}
	// Atomik upload geçici dosya bırakmamalı
	entries, _ := c.List("/site/a")
	for _, e := range entries {
		if strings.HasPrefix(e.Name, ".syncftp-") {
			t.Fatalf("temp file left behind: %s", e.Name)
		}
	}

	// İkinci yükleme: eski hal yedeklenir, yeni içerik üzerine yazılır
	writeLocal(t, local, "a/index.php", "v2")
	bk := filepath.Join(t.TempDir(), "bk2")
	existed, err = c.BackupTo("a/index.php", bk)
	if err != nil || !existed {
		t.Fatalf("backup existing: existed=%v err=%v", existed, err)
	}
	if b, _ := os.ReadFile(bk); string(b) != "line1\r\nline2\r\n" {
		t.Fatalf("backup content = %q", b)
	}
	if err := c.Upload(f1, "a/index.php"); err != nil {
		t.Fatalf("overwrite upload: %v", err)
	}
	if got, _ := c.Preview("/site/a/index.php", 1024); string(got) != "v2" {
		t.Fatalf("after overwrite = %q", got)
	}

	// Filtreli rekürsif listeleme: atlanan dizine girilmez
	f2 := writeLocal(t, local, "vendor/x/lib.php", "x")
	if err := c.Upload(f2, "vendor/x/lib.php"); err != nil {
		t.Fatal(err)
	}
	all, _ := c.ListRecursive("/site")
	if _, ok := all["vendor/x/lib.php"]; !ok {
		t.Fatalf("ListRecursive missing vendor file: %v", all)
	}
	filtered, _ := c.ListRecursiveFiltered("/site", func(rel string) bool { return rel == "vendor" }, nil)
	if _, ok := filtered["vendor/x/lib.php"]; ok {
		t.Fatalf("skipDir ignored: %v", filtered)
	}
	if filtered["a/index.php"] != 2 {
		t.Fatalf("size = %d", filtered["a/index.php"])
	}

	// Silme
	if err := c.DeleteFile("/site/a/index.php"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := c.Preview("/site/a/index.php", 10); !IsNotFound(err) {
		t.Fatalf("expected not found, got %v", err)
	}
	if err := c.DeleteDir("/site/vendor", true); err != nil {
		t.Fatalf("delete dir: %v", err)
	}

	// Host key kaydedildi (TOFU)
	data, err := os.ReadFile(KnownHostsPath())
	if err != nil || !strings.Contains(string(data), "ssh-ed25519") {
		t.Fatalf("known_hosts not written: %v %q", err, data)
	}
}

func TestSFTPHostKeyMismatch(t *testing.T) {
	withTempKnownHosts(t)
	host, port, stop := startTestSSH(t, "u", "p")
	srv := config.Server{Protocol: "sftp", Host: host, Port: port, User: "u", Password: "p", RemotePath: "/"}
	c, err := Connect(srv)
	if err != nil {
		t.Fatal(err)
	}
	c.Close()
	stop()

	// Aynı adres, farklı host key → reddedilmeli
	ln, err := net.Listen("tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		t.Skip("port reuse not possible:", err)
	}
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	signer, _ := ssh.NewSignerFromKey(priv)
	cfg := &ssh.ServerConfig{NoClientAuth: true}
	cfg.AddHostKey(signer)
	go func() {
		nc, err := ln.Accept()
		if err == nil {
			serveSSHConn(nc, cfg, sftp.InMemHandler())
		}
	}()
	defer ln.Close()
	if _, err := Connect(srv); err == nil || !strings.Contains(err.Error(), "UYUŞMUYOR") {
		t.Fatalf("expected host key mismatch, got %v", err)
	}
}

func TestSFTPThroughTunnel(t *testing.T) {
	withTempKnownHosts(t)
	// Hedef SFTP sunucusu + ayrı bir jump host
	tHost, tPort, stopTarget := startTestSSH(t, "app", "secret")
	defer stopTarget()
	jHost, jPort, stopJump := startTestSSH(t, "jump", "jp")
	defer stopJump()

	srv := config.Server{
		Protocol: "sftp", Host: tHost, Port: tPort, User: "app", Password: "secret", RemotePath: "/",
		SSHTunnel: config.SSHTunnel{TunnelHost: jHost, TunnelPort: jPort, TunnelUser: "jump", TunnelPassword: "jp"},
	}
	c, err := Connect(srv)
	if err != nil {
		t.Fatalf("connect via tunnel: %v", err)
	}
	defer c.Close()
	if c.tunnel == nil {
		t.Fatal("tunnel not used")
	}
	f := writeLocal(t, t.TempDir(), "t.txt", "through tunnel")
	if err := c.Upload(f, "t.txt"); err != nil {
		t.Fatalf("upload via tunnel: %v", err)
	}
	if got, _ := c.Preview("/t.txt", 100); string(got) != "through tunnel" {
		t.Fatalf("got %q", got)
	}
}

func TestBandwidthLimiter(t *testing.T) {
	withTempKnownHosts(t)
	host, port, stop := startTestSSH(t, "u", "p")
	defer stop()
	srv := config.Server{Protocol: "sftp", Host: host, Port: port, User: "u", Password: "p", RemotePath: "/"}
	c, err := ConnectLimited(srv, NewLimiter(64)) // 64 KB/s
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	f := writeLocal(t, t.TempDir(), "big.bin", strings.Repeat("x", 192*1024))
	start := time.Now()
	if err := c.Upload(f, "big.bin"); err != nil {
		t.Fatal(err)
	}
	// 192 KB @ 64 KB/s ≈ 2 sn (ilk burst 64 KB anında gider → en az ~2 sn)
	if el := time.Since(start); el < 1500*time.Millisecond {
		t.Fatalf("limiter too fast: %v", el)
	}
}

func TestPoolUploadWithBackup(t *testing.T) {
	withTempKnownHosts(t)
	host, port, stop := startTestSSH(t, "u", "p")
	defer stop()
	srv := config.Server{Name: "pool", Protocol: "sftp", Host: host, Port: port, User: "u", Password: "p", RemotePath: "/w", MaxConnections: 3}
	local := t.TempDir()
	var tasks []UploadTask
	for i := 0; i < 6; i++ {
		name := "f" + strconv.Itoa(i) + ".txt"
		tasks = append(tasks, UploadTask{LocalPath: writeLocal(t, local, name, "v1-"+name), RelPath: name})
	}
	bk1 := t.TempDir()
	for i := range tasks {
		tasks[i].BackupDir = bk1
	}
	pool, err := NewPool(srv)
	if err != nil {
		t.Fatal(err)
	}
	if pool.Size() != 3 {
		t.Fatalf("pool size = %d", pool.Size())
	}
	for _, r := range pool.Upload(tasks) {
		if r.Err != nil || !r.IsNew || r.BackedUp {
			t.Fatalf("first upload %s: err=%v new=%v backed=%v", r.RelPath, r.Err, r.IsNew, r.BackedUp)
		}
	}
	pool.Close()

	// İkinci tur: hepsi var → yedeklenir
	bk2 := t.TempDir()
	for i := range tasks {
		writeLocal(t, local, tasks[i].RelPath, "v2")
		tasks[i].BackupDir = bk2
	}
	pool, _ = NewPool(srv)
	defer pool.Close()
	for _, r := range pool.Upload(tasks) {
		if r.Err != nil || r.IsNew || !r.BackedUp {
			t.Fatalf("second upload %s: err=%v new=%v backed=%v", r.RelPath, r.Err, r.IsNew, r.BackedUp)
		}
		if b, _ := os.ReadFile(filepath.Join(bk2, r.RelPath)); string(b) != "v1-"+r.RelPath {
			t.Fatalf("backup of %s = %q", r.RelPath, b)
		}
	}
}

func TestHostSemaphoreSharedAcrossPools(t *testing.T) {
	srv := config.Server{Protocol: "ftp", Host: "h", Port: 21, User: "u"}
	s1 := semFor(srv, 2)
	s2 := semFor(config.Server{Protocol: "ftp", Host: "H", Port: 21, User: "u", MaxConnections: 5}, 5)
	if s1 != s2 {
		t.Fatal("same account should share semaphore")
	}
	s1.acquire()
	s1.acquire()
	if s2.tryAcquire() {
		t.Fatal("limit 2 exceeded")
	}
	s1.release(1)
	if !s2.tryAcquire() {
		t.Fatal("slot should be free")
	}
	s1.release(2)
}

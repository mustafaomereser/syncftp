package ftp

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"

	"syncftp/internal/config"
)

// ── SSH kimlik doğrulama ──────────────────────────────────────────────────────

// expandHome "~/" ile başlayan yolu kullanıcı dizinine çevirir.
func expandHome(p string) string {
	if strings.HasPrefix(p, "~/") || strings.HasPrefix(p, `~\`) {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, p[2:])
		}
	}
	return p
}

// sshAuthMethods anahtar dosyası (varsa) + şifre + keyboard-interactive yöntemlerini döner.
// Anahtar şifreliyse password passphrase olarak denenir.
func sshAuthMethods(password, keyFile string) ([]ssh.AuthMethod, error) {
	var methods []ssh.AuthMethod
	if keyFile != "" {
		data, err := os.ReadFile(expandHome(keyFile))
		if err != nil {
			return nil, fmt.Errorf("anahtar dosyası okunamadı (%s): %w", keyFile, err)
		}
		signer, err := ssh.ParsePrivateKey(data)
		var missing *ssh.PassphraseMissingError
		if errors.As(err, &missing) {
			if password == "" {
				return nil, fmt.Errorf("anahtar şifreli, passphrase için password alanını doldurun (%s)", keyFile)
			}
			signer, err = ssh.ParsePrivateKeyWithPassphrase(data, []byte(password))
		}
		if err != nil {
			return nil, fmt.Errorf("anahtar çözümlenemedi (%s): %w", keyFile, err)
		}
		methods = append(methods, ssh.PublicKeys(signer))
	}
	if password != "" {
		methods = append(methods, ssh.Password(password))
		methods = append(methods, ssh.KeyboardInteractive(func(_, _ string, questions []string, _ []bool) ([]string, error) {
			answers := make([]string, len(questions))
			for i := range answers {
				answers[i] = password
			}
			return answers, nil
		}))
	}
	if len(methods) == 0 {
		return nil, fmt.Errorf("SSH için şifre veya anahtar dosyası gerekli")
	}
	return methods, nil
}

// ── Host key doğrulama (TOFU) ─────────────────────────────────────────────────

var knownHostsMu sync.Mutex

// knownHostsOverride testlerde gerçek kullanıcı dosyasına yazılmasın diye.
var knownHostsOverride string

// KnownHostsPath syncFTP'nin kendi known_hosts dosyası: <UserConfigDir>/syncftp/known_hosts
func KnownHostsPath() string {
	if knownHostsOverride != "" {
		return knownHostsOverride
	}
	base, err := os.UserConfigDir()
	if err != nil {
		base = os.TempDir()
	}
	return filepath.Join(base, "syncftp", "known_hosts")
}

// hostKeyCallback önce ~/.ssh/known_hosts, sonra syncFTP known_hosts dosyasına bakar.
// Hiçbirinde kayıt yoksa anahtar syncFTP dosyasına eklenir (trust on first use).
// Kayıt var ama anahtar farklıysa bağlantı reddedilir (MITM koruması).
func hostKeyCallback() ssh.HostKeyCallback {
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		knownHostsMu.Lock()
		defer knownHostsMu.Unlock()

		own := KnownHostsPath()
		var files []string
		if home, err := os.UserHomeDir(); err == nil && knownHostsOverride == "" {
			if p := filepath.Join(home, ".ssh", "known_hosts"); fileExists(p) {
				files = append(files, p)
			}
		}
		if fileExists(own) {
			files = append(files, own)
		}
		if len(files) > 0 {
			cb, err := knownhosts.New(files...)
			if err == nil {
				err = cb(hostname, remote, key)
				if err == nil {
					return nil
				}
				var ke *knownhosts.KeyError
				if errors.As(err, &ke) && len(ke.Want) > 0 {
					return fmt.Errorf("host key UYUŞMUYOR (%s) — sunucu anahtarı değişmiş olabilir; güveniyorsanız %s içindeki satırı silin", hostname, own)
				}
			}
		}
		// Bilinmeyen host → kaydet
		if err := os.MkdirAll(filepath.Dir(own), 0700); err != nil {
			return err
		}
		f, err := os.OpenFile(own, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = fmt.Fprintln(f, knownhosts.Line([]string{knownhosts.Normalize(hostname)}, key))
		return err
	}
}

func fileExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}

// ── SSH bağlantı kurma ────────────────────────────────────────────────────────

// dialSSH addr'e SSH bağlantısı açar; via != nil ise bağlantı o SSH sunucusu üzerinden kurulur.
func dialSSH(addr, user string, auth []ssh.AuthMethod, via *ssh.Client) (*ssh.Client, error) {
	cfg := &ssh.ClientConfig{
		User:            user,
		Auth:            auth,
		HostKeyCallback: hostKeyCallback(),
		Timeout:         30 * time.Second,
	}
	var conn net.Conn
	var err error
	if via == nil {
		conn, err = net.DialTimeout("tcp", addr, 30*time.Second)
	} else {
		conn, err = via.Dial("tcp", addr)
	}
	if err != nil {
		return nil, err
	}
	conn = &writeDeadlineConn{Conn: conn}
	c, chans, reqs, err := ssh.NewClientConn(conn, addr, cfg)
	if err != nil {
		conn.Close()
		return nil, err
	}
	client := ssh.NewClient(c, chans, reqs)
	sshKeepalive(client)
	return client, nil
}

// dialTunnel sunucunun SSH tünelini (jump host) açar; tünel yoksa nil döner.
func dialTunnel(srv config.Server) (*ssh.Client, error) {
	if !srv.HasTunnel() {
		return nil, nil
	}
	port := srv.TunnelPort
	if port == 0 {
		port = 22
	}
	addr := net.JoinHostPort(srv.TunnelHost, fmt.Sprint(port))
	auth, err := sshAuthMethods(srv.TunnelPassword, srv.TunnelKeyFile)
	if err != nil {
		return nil, fmt.Errorf("tünel: %w", err)
	}
	user := srv.TunnelUser
	if user == "" {
		user = srv.User
	}
	c, err := dialSSH(addr, user, auth, nil)
	if err != nil {
		return nil, fmt.Errorf("SSH tüneli kurulamadı (%s@%s): %w", user, addr, err)
	}
	return c, nil
}

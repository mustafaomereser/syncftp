package ftp

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	goftp "github.com/jlaffaye/ftp"

	"syncftp/internal/config"
)

// Client wraps an active FTP connection for a single server.
type Client struct {
	conn   *goftp.ServerConn
	server config.Server
}

// Connect dials and authenticates an FTP connection.
func Connect(srv config.Server) (*Client, error) {
	addr := fmt.Sprintf("%s:%d", srv.Host, srv.Port)
	conn, err := goftp.Dial(addr, goftp.DialWithTimeout(30*time.Second))
	if err != nil {
		return nil, fmt.Errorf("bağlantı kurulamadı (%s): %w", addr, err)
	}
	if err := conn.Login(srv.User, srv.Password); err != nil {
		_ = conn.Quit()
		return nil, fmt.Errorf("giriş başarısız (%s@%s): %w", srv.User, addr, err)
	}
	return &Client{conn: conn, server: srv}, nil
}

// Close gracefully disconnects from the FTP server.
func (c *Client) Close() {
	_ = c.conn.Quit()
}

// Upload sends a local file to the remote server, creating any missing directories first.
// relPath is the forward-slash relative path (e.g. "css/style.css").
func (c *Client) Upload(localPath, relPath string) error {
	remoteFull := path.Join(c.server.RemotePath, relPath)
	remoteDir := path.Dir(remoteFull)

	if err := c.mkdirAll(remoteDir); err != nil {
		return fmt.Errorf("dizin oluşturulamadı (%s): %w", remoteDir, err)
	}

	f, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("dosya açılamadı: %w", err)
	}
	defer f.Close()

	if err := c.conn.Stor(remoteFull, f); err != nil {
		return fmt.Errorf("yükleme başarısız (%s): %w", remoteFull, err)
	}
	return nil
}

// mkdirAll ensures that the full remote path exists, creating directories one level at a time.
// Errors from MakeDir are silently ignored since the directory may already exist.
func (c *Client) mkdirAll(remotePath string) error {
	parts := strings.Split(strings.TrimPrefix(remotePath, "/"), "/")
	current := "/"
	for _, part := range parts {
		if part == "" {
			continue
		}
		current = path.Join(current, part)
		_ = c.conn.MakeDir(current)
	}
	return nil
}

// IsProtected returns true if relPath matches any protect pattern.
// Patterns ending with "/" are treated as directory prefixes.
func IsProtected(relPath string, protect []string) bool {
	for _, pattern := range protect {
		pattern = strings.TrimPrefix(pattern, "/")
		if strings.HasSuffix(pattern, "/") {
			// directory prefix match
			if strings.HasPrefix(relPath, pattern) {
				return true
			}
		} else if relPath == pattern {
			return true
		}
	}
	return false
}

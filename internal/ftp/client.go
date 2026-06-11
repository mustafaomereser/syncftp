package ftp

import (
	"fmt"
	"io"
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

// List returns entries at the given remote directory path.
func (c *Client) List(remotePath string) ([]*goftp.Entry, error) {
	entries, err := c.conn.List(remotePath)
	if err != nil {
		return nil, fmt.Errorf("listeleme başarısız (%s): %w", remotePath, err)
	}
	return entries, nil
}

// Download saves a remote file to localDest.
func (c *Client) Download(remotePath, localDest string) error {
	resp, err := c.conn.Retr(remotePath)
	if err != nil {
		return fmt.Errorf("dosya alınamadı (%s): %w", remotePath, err)
	}
	defer resp.Close()

	f, err := os.Create(localDest)
	if err != nil {
		return fmt.Errorf("yerel dosya oluşturulamadı: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp); err != nil {
		os.Remove(localDest)
		return fmt.Errorf("indirme başarısız: %w", err)
	}
	return nil
}

// DeleteFile removes a single remote file.
func (c *Client) DeleteFile(remotePath string) error {
	if err := c.conn.Delete(remotePath); err != nil {
		return fmt.Errorf("silme başarısız (%s): %w", remotePath, err)
	}
	return nil
}

// DeleteDir removes a remote directory.
// Set recursive=true to delete non-empty directories.
func (c *Client) DeleteDir(remotePath string, recursive bool) error {
	if recursive {
		if err := c.conn.RemoveDirRecur(remotePath); err != nil {
			return fmt.Errorf("dizin silme başarısız (%s): %w", remotePath, err)
		}
		return nil
	}
	if err := c.conn.RemoveDir(remotePath); err != nil {
		return fmt.Errorf("dizin silme başarısız (%s): %w", remotePath, err)
	}
	return nil
}

// Preview reads up to maxBytes from a remote file and returns the content.
func (c *Client) Preview(remotePath string, maxBytes int64) ([]byte, error) {
	resp, err := c.conn.Retr(remotePath)
	if err != nil {
		return nil, fmt.Errorf("dosya alınamadı (%s): %w", remotePath, err)
	}
	defer resp.Close()

	buf := make([]byte, maxBytes)
	n, err := io.ReadFull(resp, buf)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return nil, err
	}
	return buf[:n], nil
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

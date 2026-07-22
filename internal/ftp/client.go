package ftp

import (
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	goftp "github.com/jlaffaye/ftp"

	"syncftp/internal/config"
)

// Client wraps an active FTP connection for a single server.
// mu serialize eder — aynı bağlantıya eş zamanlı erişim
// "short response" hatalarına yol açar.
type Client struct {
	conn   *goftp.ServerConn
	server config.Server
	mu     sync.Mutex
}

// Connect dials and authenticates an FTP connection.
func Connect(srv config.Server) (*Client, error) {
	addr := fmt.Sprintf("%s:%d", srv.Host, srv.Port)

	opts := []goftp.DialOption{
		goftp.DialWithTimeout(30 * time.Second),
		goftp.DialWithDisabledEPSV(srv.DisableEPSV || srv.NATWorkaround),
	}

	if srv.NATWorkaround {
		serverHost := srv.Host
		opts = append(opts, goftp.DialWithDialFunc(func(network, address string) (net.Conn, error) {
			_, port, splitErr := net.SplitHostPort(address)
			if splitErr != nil {
				return net.Dial(network, address)
			}
			return net.DialTimeout(network, net.JoinHostPort(serverHost, port), 15*time.Second)
		}))
	}

	conn, err := goftp.Dial(addr, opts...)
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
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.conn.Quit()
}

// Upload sends a local file to the remote server.
func (c *Client) Upload(localPath, relPath string) error {
	remoteFull := path.Join(c.server.RemotePath, relPath)
	remoteDir := path.Dir(remoteFull)

	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.mkdirAllLocked(remoteDir); err != nil {
		return fmt.Errorf("dizin oluşturulamadı (%s): %w", remoteDir, err)
	}

	// Bazı PHP hosting sunucuları TYPE I'ı dosya bazında override edebiliyor.
	// Her upload öncesi binary modu açıkça set et.
	if err := c.conn.Type(goftp.TransferTypeBinary); err != nil {
		return fmt.Errorf("binary mod ayarlanamadı: %w", err)
	}

	f, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("dosya açılamadı: %w", err)
	}
	defer f.Close()

	// Metin dosyaları: \r\n → \n normalize et (PHP hosting ASCII mod sorunu).
	// Binary dosyalar (resim, PDF, ZIP vb.): normalizer ATLANIR — bozulmayı önler.
	// Önce uzantı (kesin), sonra içerik (null byte) kontrolü.
	var uploadReader io.Reader
	if isBinaryExt(relPath) || isBinaryContent(f) {
		uploadReader = f
	} else {
		uploadReader = newCRLFNormalizer(f)
	}
	if err := c.conn.Stor(remoteFull, uploadReader); err != nil {
		return fmt.Errorf("yükleme başarısız (%s): %w", remoteFull, err)
	}
	return nil
}

// binaryExts — uzantısından kesin binary olduğu bilinen dosya türleri.
// Bu dosyalarda içerik kontrolüne gerek kalmaz; ilk 8 KB'ında null byte
// bulunmayan nadir binary'ler (bazı GIF/ICO varyantları vb.) de yakalanır.
var binaryExts = map[string]bool{
	// görseller
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
	".ico": true, ".bmp": true, ".tif": true, ".tiff": true, ".avif": true, ".heic": true,
	// arşiv / belge
	".pdf": true, ".zip": true, ".rar": true, ".7z": true, ".gz": true, ".bz2": true,
	".tar": true, ".xz": true, ".xlsx": true, ".xls": true, ".docx": true, ".doc": true,
	".pptx": true, ".ppt": true, ".odt": true, ".ods": true,
	// medya
	".mp3": true, ".mp4": true, ".wav": true, ".ogg": true, ".flac": true, ".m4a": true,
	".avi": true, ".mov": true, ".mkv": true, ".webm": true,
	// font
	".woff": true, ".woff2": true, ".ttf": true, ".otf": true, ".eot": true,
	// çalıştırılabilir / veri
	".exe": true, ".dll": true, ".so": true, ".dylib": true, ".wasm": true,
	".sqlite": true, ".db": true, ".dat": true, ".bin": true, ".phar": true,
}

// isBinaryExt dosya uzantısı bilinen binary türlerinden biriyse true döner.
func isBinaryExt(relPath string) bool {
	return binaryExts[strings.ToLower(path.Ext(relPath))]
}

// isBinaryContent dosyanın ilk 8 KB'ını okur; null byte varsa binary döner.
// Okuma sonrası dosya başa sarılır, Upload akışı etkilenmez.
func isBinaryContent(f *os.File) bool {
	buf := make([]byte, 8192)
	n, _ := f.Read(buf)
	_, _ = f.Seek(0, io.SeekStart)
	for _, b := range buf[:n] {
		if b == 0 {
			return true
		}
	}
	return false
}

// crlfNormalizer \r\n → \n dönüşümü yapan io.Reader wrapper'ı.
type crlfNormalizer struct {
	r      io.Reader
	buf    []byte
	out    []byte
}

func newCRLFNormalizer(r io.Reader) *crlfNormalizer {
	return &crlfNormalizer{r: r, buf: make([]byte, 32*1024)}
}

func (c *crlfNormalizer) Read(p []byte) (int, error) {
	if len(c.out) > 0 {
		n := copy(p, c.out)
		c.out = c.out[n:]
		return n, nil
	}
	n, err := c.r.Read(c.buf)
	if n == 0 {
		return 0, err
	}
	// \r\n → \n
	src := c.buf[:n]
	dst := make([]byte, 0, n)
	for i := 0; i < len(src); i++ {
		if src[i] == '\r' && i+1 < len(src) && src[i+1] == '\n' {
			continue // \r'ı atla, \n gelecek
		}
		// \r tek başına (son byte olabilir) — bir sonraki chunk'ta \n gelebilir,
		// güvenli taraf: olduğu gibi bırak
		dst = append(dst, src[i])
	}
	written := copy(p, dst)
	if written < len(dst) {
		c.out = dst[written:]
	}
	return written, err
}

// mkdirAllLocked — mu alınmış durumda çağrılmalı.
func (c *Client) mkdirAllLocked(remotePath string) error {
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
// 550 hatası (yok/boş dizin) → boş liste döner, hata yok.
func (c *Client) List(remotePath string) ([]*goftp.Entry, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entries, err := c.conn.List(remotePath)
	if err != nil {
		// Bazı sunucular boş veya mevcut olmayan dizin için 550 döndürür.
		// Bu durumda boş liste döndür, fatal hata değil.
		if strings.Contains(err.Error(), "550") ||
			strings.Contains(err.Error(), "No such file") ||
			strings.Contains(err.Error(), "doesn't exist") ||
			strings.Contains(err.Error(), "not found") {
			return nil, nil
		}
		return nil, fmt.Errorf("listeleme başarısız (%s): %w", remotePath, err)
	}
	return entries, nil
}

// Download saves a remote file to localDest.
// Mutex okuma boyunca tutulur — Retr sonrası bırakmak data stream'i bozar.
func (c *Client) Download(remotePath, localDest string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

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
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.conn.Delete(remotePath); err != nil {
		return fmt.Errorf("silme başarısız (%s): %w", remotePath, err)
	}
	return nil
}

// DeleteDir removes a remote directory.
func (c *Client) DeleteDir(remotePath string, recursive bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()
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
// Mutex Retr + okuma boyunca tutulur — erken bırakmak "short response" üretir.
func (c *Client) Preview(remotePath string, maxBytes int64) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	resp, err := c.conn.Retr(remotePath)
	if err != nil {
		return nil, fmt.Errorf("dosya alınamadı (%s): %w", remotePath, err)
	}
	defer resp.Close()

	buf, err := io.ReadAll(io.LimitReader(resp, maxBytes))
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// ListRecursive remotePath altındaki tüm dosyaları rekürsif listeler.
// relPath → boyut (byte). Her List çağrısı ayrı mu ile serialize edilir.
func (c *Client) ListRecursive(remotePath string) (map[string]uint64, error) {
	return c.ListRecursiveProgress(remotePath, nil)
}

// ListRecursiveProgress ListRecursive ile aynıdır; progress nil değilse
// her yeni dosya bulunduğunda toplam sayıyla çağrılır (gerçek zamanlı sayaç).
func (c *Client) ListRecursiveProgress(remotePath string, progress func(count int)) (map[string]uint64, error) {
	if progress == nil {
		return c.ListRecursiveProgressPath(remotePath, nil)
	}
	return c.ListRecursiveProgressPath(remotePath, func(count int, _ string) { progress(count) })
}

// ListRecursiveProgressPath ListRecursiveProgress ile aynıdır; callback'e toplam
// sayıya ek olarak son bulunan dosyanın göreceli yolu da geçer (canlı dosya akışı).
func (c *Client) ListRecursiveProgressPath(remotePath string, progress func(count int, latest string)) (map[string]uint64, error) {
	result := make(map[string]uint64)
	var walk func(dir, rel string)
	walk = func(dir, rel string) {
		c.mu.Lock()
		entries, err := c.conn.List(dir)
		c.mu.Unlock()
		if err != nil {
			return
		}
		for _, e := range entries {
			if e.Name == "." || e.Name == ".." {
				continue
			}
			entryRel := e.Name
			if rel != "" {
				entryRel = rel + "/" + e.Name
			}
			if e.Type == goftp.EntryTypeFolder {
				walk(path.Join(dir, e.Name), entryRel)
			} else {
				result[entryRel] = e.Size
				if progress != nil {
					progress(len(result), entryRel)
				}
			}
		}
	}
	walk(remotePath, "")
	return result, nil
}

// Rename moves a remote file or directory from oldPath to newPath.
func (c *Client) Rename(oldPath, newPath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.conn.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("taşıma başarısız (%s → %s): %w", oldPath, newPath, err)
	}
	return nil
}

// IsProtected returns true if relPath matches any protect pattern.
func IsProtected(relPath string, protect []string) bool {
	for _, pattern := range protect {
		pattern = strings.TrimPrefix(pattern, "/")
		if strings.HasSuffix(pattern, "/") {
			if strings.HasPrefix(relPath, pattern) {
				return true
			}
		} else if relPath == pattern {
			return true
		}
	}
	return false
}

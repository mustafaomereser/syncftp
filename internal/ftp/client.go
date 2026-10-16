package ftp

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	goftp "github.com/jlaffaye/ftp"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/time/rate"

	"syncftp/internal/config"
)

// Client wraps an active FTP or SFTP connection for a single server.
// mu serialize eder — aynı bağlantıya eş zamanlı erişim
// "short response" hatalarına yol açar.
// Protokol: conn != nil → FTP, sftp != nil → SFTP. Tüm metotlar iki yolu da destekler;
// çağıranlar protokolden habersizdir (List her iki durumda goftp.Entry döner).
type Client struct {
	conn    *goftp.ServerConn
	sftp    *sftp.Client
	sshConn *ssh.Client // SFTP'nin taşıyıcı SSH bağlantısı
	tunnel  *ssh.Client // SSH tüneli (jump host); yoksa nil
	server  config.Server
	limiter *rate.Limiter // upload hız sınırı (nil = sınırsız); pool genelinde paylaşılır
	mu      sync.Mutex
}

// Connect dials and authenticates an FTP/SFTP connection.
// Sunucunun bant genişliği sınırı bu bağlantıya özel limiter ile uygulanır;
// pool genelinde paylaşılan limiter için ConnectLimited kullanılır.
func Connect(srv config.Server) (*Client, error) {
	return ConnectLimited(srv, nil)
}

// ConnectLimited Connect ile aynıdır; lim != nil ise upload'lar bu limiter ile kısılır.
func ConnectLimited(srv config.Server, lim *rate.Limiter) (*Client, error) {
	tunnel, err := dialTunnel(srv)
	if err != nil {
		return nil, err
	}
	var c *Client
	if srv.IsSFTP() {
		c, err = connectSFTP(srv, tunnel)
	} else {
		c, err = connectFTP(srv, tunnel)
	}
	if err != nil {
		if tunnel != nil {
			tunnel.Close()
		}
		return nil, err
	}
	c.tunnel = tunnel
	c.limiter = lim
	return c, nil
}

func connectFTP(srv config.Server, tunnel *ssh.Client) (*Client, error) {
	addr := fmt.Sprintf("%s:%d", srv.Host, srv.Port)

	opts := []goftp.DialOption{
		goftp.DialWithTimeout(30 * time.Second),
		goftp.DialWithDisabledEPSV(srv.DisableEPSV || srv.NATWorkaround),
	}
	if srv.IsFTPS() {
		opts = append(opts, goftp.DialWithExplicitTLS(&tls.Config{
			ServerName:         srv.Host,
			InsecureSkipVerify: srv.TLSSkipVerify, //nolint:gosec // kullanıcı tercihi (paylaşımlı hosting sertifikaları)
			MinVersion:         tls.VersionTLS12,
		}))
	}

	// Tüm bağlantılar (kontrol + pasif veri) buradan açılır: NAT düzeltmesi, SSH tüneli
	// ve idle deadline (takılan veri bağlantısı sync'i sonsuza kadar kilitlemesin) tek yerde.
	serverHost := srv.Host
	nat := srv.NATWorkaround
	opts = append(opts, goftp.DialWithDialFunc(func(network, address string) (net.Conn, error) {
		target := address
		if nat {
			if _, port, splitErr := net.SplitHostPort(address); splitErr == nil {
				target = net.JoinHostPort(serverHost, port)
			}
		}
		if tunnel != nil {
			return withDeadline(tunnel.Dial(network, target))
		}
		return withDeadline(net.DialTimeout(network, target, 30*time.Second))
	}))

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

func connectSFTP(srv config.Server, tunnel *ssh.Client) (*Client, error) {
	addr := net.JoinHostPort(srv.Host, fmt.Sprint(srv.Port))
	auth, err := sshAuthMethods(srv.Password, srv.KeyFile)
	if err != nil {
		return nil, err
	}
	sshConn, err := dialSSH(addr, srv.User, auth, tunnel)
	if err != nil {
		return nil, fmt.Errorf("bağlantı kurulamadı (sftp://%s@%s): %w", srv.User, addr, err)
	}
	sc, err := sftp.NewClient(sshConn)
	if err != nil {
		sshConn.Close()
		return nil, fmt.Errorf("SFTP alt sistemi açılamadı (%s): %w", addr, err)
	}
	return &Client{sftp: sc, sshConn: sshConn, server: srv}, nil
}

// Ping bağlantının canlı olduğunu yoldan bağımsız doğrular (FTP NOOP / SFTP Getwd).
func (c *Client) Ping() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.sftp != nil {
		_, err := c.sftp.Getwd()
		return err
	}
	return c.conn.NoOp()
}

// CanExec bağlantı SSH tabanlı mı (uzak komut çalıştırılabilir).
func (c *Client) CanExec() bool { return c.sshConn != nil }

// Exec SSH üzerinden uzak dizinde (RemotePath) komut çalıştırır; stdout+stderr w'ye akar.
// env değişkenleri komutun önüne export olarak eklenir (sshd AcceptEnv gerektirmez).
// Yalnızca SFTP sunucularında kullanılabilir (FTP'de SSH yok).
func (c *Client) Exec(command string, env map[string]string, w io.Writer) error {
	if c.sshConn == nil {
		return fmt.Errorf("uzak komut yalnızca SFTP (SSH) sunucularında çalışır")
	}
	sess, err := c.sshConn.NewSession()
	if err != nil {
		return fmt.Errorf("SSH oturumu açılamadı: %w", err)
	}
	defer sess.Close()
	sess.Stdout = w
	sess.Stderr = w
	var b strings.Builder
	for k, v := range env {
		fmt.Fprintf(&b, "export %s=%s; ", k, shellQuote(v))
	}
	if c.server.RemotePath != "" {
		fmt.Fprintf(&b, "cd %s && ", shellQuote(c.server.RemotePath))
	}
	b.WriteString(command)
	return sess.Run(b.String())
}

// shellQuote POSIX sh için tek tırnaklı güvenli değer.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// Close gracefully disconnects from the server.
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		_ = c.conn.Quit()
	}
	if c.sftp != nil {
		_ = c.sftp.Close()
	}
	if c.sshConn != nil {
		_ = c.sshConn.Close()
	}
	if c.tunnel != nil {
		_ = c.tunnel.Close()
	}
}

// Upload sends a local file to the remote server.
// Varsayılan atomik: önce aynı dizinde ".syncftp-<ad>.tmp" adıyla yüklenir, sonra
// rename ile yerine konur → bağlantı yarıda koparsa canlıda yarım dosya kalmaz.
// Rename başarısız olursa (üzerine rename'e izin vermeyen sunucu) doğrudan yazmaya düşer.
// server.DirectUpload = true → doğrudan hedefe yazılır.
func (c *Client) Upload(localPath, relPath string) error {
	remoteFull := path.Join(c.server.RemotePath, relPath)
	remoteDir := path.Dir(remoteFull)

	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.mkdirAllLocked(remoteDir); err != nil {
		return fmt.Errorf("dizin oluşturulamadı (%s): %w", remoteDir, err)
	}

	if c.server.DirectUpload {
		return c.storLocked(localPath, relPath, remoteFull)
	}

	tmp := path.Join(remoteDir, ".syncftp-"+path.Base(remoteFull)+".tmp")
	if err := c.storLocked(localPath, relPath, tmp); err != nil {
		c.removeQuietLocked(tmp)
		return err
	}
	if err := c.renameOverLocked(tmp, remoteFull); err != nil {
		// Atomik yer değiştirme olmadı → güvenli yol: doğrudan yaz, tmp'yi temizle
		c.removeQuietLocked(tmp)
		return c.storLocked(localPath, relPath, remoteFull)
	}
	return nil
}

// storLocked dosyayı remoteFull'e yazar (mu alınmış olmalı).
func (c *Client) storLocked(localPath, relPath, remoteFull string) error {
	f, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("dosya açılamadı: %w", err)
	}
	defer f.Close()

	// SFTP binary-safe → CRLF normalizasyonu yapılmaz, dosya bayt bayt gider
	// (calibrate tam boyut eşleşmesiyle yakalar).
	if c.sftp != nil {
		dst, err := c.sftp.OpenFile(remoteFull, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
		if err != nil {
			return fmt.Errorf("yükleme başarısız (%s): %w", remoteFull, err)
		}
		if _, err := io.Copy(dst, newLimitedReader(f, c.limiter)); err != nil {
			dst.Close()
			return fmt.Errorf("yükleme başarısız (%s): %w", remoteFull, err)
		}
		if err := dst.Close(); err != nil {
			return fmt.Errorf("yükleme başarısız (%s): %w", remoteFull, err)
		}
		return nil
	}

	// Bazı PHP hosting sunucuları TYPE I'ı dosya bazında override edebiliyor.
	// Her upload öncesi binary modu açıkça set et.
	if err := c.conn.Type(goftp.TransferTypeBinary); err != nil {
		return fmt.Errorf("binary mod ayarlanamadı: %w", err)
	}

	// Metin dosyaları: \r\n → \n normalize et (PHP hosting ASCII mod sorunu).
	// Binary dosyalar (resim, PDF, ZIP vb.): normalizer ATLANIR — bozulmayı önler.
	// Önce uzantı (kesin), sonra içerik (null byte) kontrolü.
	var uploadReader io.Reader
	if isBinaryExt(relPath) || isBinaryContent(f) {
		uploadReader = f
	} else {
		uploadReader = newCRLFNormalizer(f)
	}
	if err := c.conn.Stor(remoteFull, newLimitedReader(uploadReader, c.limiter)); err != nil {
		return fmt.Errorf("yükleme başarısız (%s): %w", remoteFull, err)
	}
	return nil
}

// renameOverLocked from'u to'nun üzerine taşır (to varsa değiştirilir).
func (c *Client) renameOverLocked(from, to string) error {
	if c.sftp != nil {
		// Standart SFTP rename hedef varken başarısız olur; OpenSSH uzantısı atomik üzerine yazar
		if _, ok := c.sftp.HasExtension("posix-rename@openssh.com"); ok {
			return c.sftp.PosixRename(from, to)
		}
		return c.sftp.Rename(from, to)
	}
	return c.conn.Rename(from, to)
}

func (c *Client) removeQuietLocked(p string) {
	if c.sftp != nil {
		_ = c.sftp.Remove(p)
	} else {
		_ = c.conn.Delete(p)
	}
}

// isNotFound uzak dosya yok hatası mı (FTP 550 / SFTP no such file).
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, os.ErrNotExist) {
		return true
	}
	s := err.Error()
	return strings.Contains(s, "550") || strings.Contains(s, "No such file") || strings.Contains(s, "No such directory") ||
		strings.Contains(s, "not found") || strings.Contains(s, "doesn't exist") ||
		strings.Contains(s, "does not exist")
}

// BackupTo uzak dosyanın (RemotePath'e göreceli relPath) mevcut halini dest'e indirir.
// existed=false → dosya sunucuda yoktu (rollback'te silinecek "yeni" dosya).
// Okunamayan ama var olan dosya hata döner — "yeni" sayılırsa rollback onu silerdi.
func (c *Client) BackupTo(relPath, dest string) (existed bool, err error) {
	remoteFull := path.Join(c.server.RemotePath, relPath)

	c.mu.Lock()
	defer c.mu.Unlock()

	var sizeErr error
	if c.sftp != nil {
		if _, sizeErr = c.sftp.Stat(remoteFull); IsNotFound(sizeErr) {
			return false, nil
		}
	} else {
		// SIZE bazı sunucularda (vsftpd) ASCII modda 550 döner → önce binary
		_ = c.conn.Type(goftp.TransferTypeBinary)
		if _, sizeErr = c.conn.FileSize(remoteFull); sizeErr != nil && IsNotFound(sizeErr) {
			return false, nil
		}
	}

	resp, err := c.retrLocked(remoteFull)
	if err != nil {
		// SIZE desteklenmiyor + RETR 550 → büyük ihtimalle dosya yok
		if sizeErr != nil && IsNotFound(err) {
			return false, nil
		}
		return true, fmt.Errorf("yedek alınamadı (%s): %w", remoteFull, err)
	}
	defer resp.Close()

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return true, err
	}
	f, err := os.Create(dest)
	if err != nil {
		return true, err
	}
	if _, err := io.Copy(f, resp); err != nil {
		f.Close()
		os.Remove(dest)
		return true, fmt.Errorf("yedek indirilemedi (%s): %w", remoteFull, err)
	}
	return true, f.Close()
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
	r   io.Reader
	buf []byte
	out []byte
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

// MakeDirAll uzak dizini (ara dizinlerle birlikte) oluşturur.
func (c *Client) MakeDirAll(remotePath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.mkdirAllLocked(remotePath)
}

// UploadTo yerel dosyayı mutlak uzak yola yükler (GUI dosya tarayıcısı; RemotePath dışı yollar için).
func (c *Client) UploadTo(localPath, remoteFull string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.mkdirAllLocked(path.Dir(remoteFull)); err != nil {
		return err
	}
	return c.storLocked(localPath, path.Base(remoteFull), remoteFull)
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
		if c.sftp != nil {
			_ = c.sftp.Mkdir(current)
		} else {
			_ = c.conn.MakeDir(current)
		}
	}
	return nil
}

// List returns entries at the given remote directory path.
// 550 hatası (yok/boş dizin) → boş liste döner, hata yok.
func (c *Client) List(remotePath string) ([]*goftp.Entry, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entries, err := c.listLocked(remotePath)
	if err != nil {
		// Bazı sunucular boş veya mevcut olmayan dizin için 550 döndürür.
		// Bu durumda boş liste döndür, fatal hata değil.
		if strings.Contains(err.Error(), "550") ||
			strings.Contains(err.Error(), "No such file") ||
			strings.Contains(err.Error(), "No such directory") ||
			strings.Contains(err.Error(), "doesn't exist") ||
			strings.Contains(err.Error(), "not found") {
			return nil, nil
		}
		return nil, fmt.Errorf("listeleme başarısız (%s): %w", remotePath, err)
	}
	return entries, nil
}

// listLocked — mu alınmış durumda çağrılmalı. SFTP'de os.FileInfo → goftp.Entry çevrilir.
func (c *Client) listLocked(remotePath string) ([]*goftp.Entry, error) {
	if c.sftp == nil {
		return c.conn.List(remotePath)
	}
	infos, err := c.sftp.ReadDir(remotePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("550 not found: %w", err)
		}
		return nil, err
	}
	out := make([]*goftp.Entry, 0, len(infos))
	for _, fi := range infos {
		e := &goftp.Entry{Name: fi.Name(), Size: uint64(fi.Size()), Time: fi.ModTime(), Type: goftp.EntryTypeFile}
		switch {
		case fi.Mode()&os.ModeSymlink != 0:
			e.Type = goftp.EntryTypeLink
		case fi.IsDir():
			e.Type = goftp.EntryTypeFolder
			e.Size = 0
		}
		out = append(out, e)
	}
	return out, nil
}

// retrLocked — mu alınmış durumda çağrılmalı; dosya okuyucu döner.
func (c *Client) retrLocked(remotePath string) (io.ReadCloser, error) {
	if c.sftp != nil {
		return c.sftp.Open(remotePath)
	}
	return c.conn.Retr(remotePath)
}

// Download saves a remote file to localDest.
// Mutex okuma boyunca tutulur — Retr sonrası bırakmak data stream'i bozar.
func (c *Client) Download(remotePath, localDest string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	resp, err := c.retrLocked(remotePath)
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
	var err error
	if c.sftp != nil {
		err = c.sftp.Remove(remotePath)
	} else {
		err = c.conn.Delete(remotePath)
	}
	if err != nil {
		return fmt.Errorf("silme başarısız (%s): %w", remotePath, err)
	}
	return nil
}

// DeleteDir removes a remote directory.
func (c *Client) DeleteDir(remotePath string, recursive bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.sftp != nil {
		var err error
		if recursive {
			err = c.sftp.RemoveAll(remotePath)
		} else {
			err = c.sftp.RemoveDirectory(remotePath)
		}
		if err != nil {
			return fmt.Errorf("dizin silme başarısız (%s): %w", remotePath, err)
		}
		return nil
	}
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

	resp, err := c.retrLocked(remotePath)
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
	return c.ListRecursiveFiltered(remotePath, nil, progress)
}

// ListRecursiveFiltered ListRecursiveProgressPath ile aynıdır; skipDir != nil ise
// göreceli yolu için true dönen dizinlere HİÇ girilmez (ignore/exclude edilmiş vendor,
// node_modules gibi dev ağaçlar için binlerce LIST komutu atılmaz).
func (c *Client) ListRecursiveFiltered(remotePath string, skipDir func(rel string) bool, progress func(count int, latest string)) (map[string]uint64, error) {
	result := make(map[string]uint64)
	var walk func(dir, rel string)
	walk = func(dir, rel string) {
		c.mu.Lock()
		entries, err := c.listLocked(dir)
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
				if skipDir != nil && skipDir(entryRel) {
					continue
				}
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
	var err error
	if c.sftp != nil {
		err = c.sftp.Rename(oldPath, newPath)
	} else {
		err = c.conn.Rename(oldPath, newPath)
	}
	if err != nil {
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

package ftp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"syncftp/internal/config"
)

// UploadTask is a single file to be uploaded.
type UploadTask struct {
	LocalPath string
	RelPath   string
	Hash      string
	BackupDir string // boş değilse upload öncesi uzak dosyanın eski hali BackupDir/RelPath'e indirilir
}

// UploadResult holds the outcome of one upload task.
type UploadResult struct {
	RelPath   string
	Hash      string
	Size      int64 // yüklenen dosyanın yerel boyutu (byte); hatada 0
	Err       error
	Attempts  int
	RetryErrs []error // her başarısız denemenin hatası (son hariç)
	BackedUp  bool    // eski uzak sürüm yedeklendi
	IsNew     bool    // dosya upload öncesi sunucuda yoktu (rollback'te silinir)
	BackupErr error   // yedek alınamadı (upload yine yapıldı; bu dosya geri alınamaz)
}

// ── Host başına bağlantı sınırı ───────────────────────────────────────────────
//
// Aynı hesaba (protokol+host+port+user) bağlanan birden fazla sunucu tanımı paralel
// sync edilince her pool kendi max_connections'ını açıyordu → toplam host limitini
// aşıp "421 Too many connections" alınıyordu. Artık aynı hesaba giden tüm pool'lar
// tek bir semaphore'u paylaşır; limit, O AN açık pool'ların en küçük max_connections'ıdır
// (pool kapanınca kaydı düşer → max_connections artırılınca yeniden başlatma gerekmez).

type hostSem struct {
	mu   sync.Mutex
	cond *sync.Cond
	regs map[int]int // açık pool'ların max_connections değeri → pool sayısı
	used int
}

// limitLocked kayıtlı en küçük limit (kayıt yoksa 1) — s.mu alınmış olmalı.
func (s *hostSem) limitLocked() int {
	lim := 0
	for l, n := range s.regs {
		if n > 0 && (lim == 0 || l < lim) {
			lim = l
		}
	}
	if lim <= 0 {
		lim = 1
	}
	return lim
}

// unregister semFor kaydını düşer (pool kapandı); limit büyüyebilir → bekleyenler uyanır.
func (s *hostSem) unregister(limit int) {
	s.mu.Lock()
	if s.regs[limit] > 1 {
		s.regs[limit]--
	} else {
		delete(s.regs, limit)
	}
	s.mu.Unlock()
	s.cond.Broadcast()
}

var (
	hostSemsMu sync.Mutex
	hostSems   = map[string]*hostSem{}
)

// HostKey aynı FTP/SFTP hesabını tanımlayan anahtar (bağlantı paylaşımı/limit için).
func HostKey(srv config.Server) string {
	return strings.ToLower(fmt.Sprintf("%s://%s@%s:%d", srv.ProtocolName(), srv.User, srv.Host, srv.Port))
}

// semFor hesabın semaphore'unu döner ve limit'i kaydeder; iş bitince unregister(limit) çağrılmalı.
func semFor(srv config.Server, limit int) *hostSem {
	hostSemsMu.Lock()
	defer hostSemsMu.Unlock()
	k := HostKey(srv)
	s, ok := hostSems[k]
	if !ok {
		s = &hostSem{regs: map[int]int{}}
		s.cond = sync.NewCond(&s.mu)
		hostSems[k] = s
	}
	s.mu.Lock()
	s.regs[limit]++
	s.mu.Unlock()
	return s
}

// acquire bir slot bekler (pool'un ilk bağlantısı — en az bir bağlantı garanti).
func (s *hostSem) acquire() {
	s.mu.Lock()
	for s.used >= s.limitLocked() {
		s.cond.Wait()
	}
	s.used++
	s.mu.Unlock()
}

// tryAcquire boş slot varsa alır (ek bağlantılar — başka pool'u bekletmesin).
func (s *hostSem) tryAcquire() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.used >= s.limitLocked() {
		return false
	}
	s.used++
	return true
}

func (s *hostSem) release(n int) {
	s.mu.Lock()
	s.used -= n
	if s.used < 0 {
		s.used = 0
	}
	s.mu.Unlock()
	s.cond.Broadcast()
}

// IsTooMany 421 / bağlantı limiti hatası mı.
func IsTooMany(err error) bool {
	if err == nil {
		return false
	}
	m := strings.ToLower(err.Error())
	return strings.Contains(m, "421") || strings.Contains(m, "too many") || strings.Contains(m, "maximum number")
}

// ── Pool ──────────────────────────────────────────────────────────────────────

// Pool manages N concurrent connections to the same server.
type Pool struct {
	clients    []*Client
	maxRetries int
	sem        *hostSem
	semLimit   int // semFor'a kaydedilen limit (Close'da düşülür)
	held       int
}

// NewPool opens up to srv.MaxConnections parallel connections.
// If some connections fail but at least one succeeds, the pool continues
// with the available connections. Returns an error only if zero connections
// could be established.
func NewPool(srv config.Server) (*Pool, error) {
	return NewPoolLimited(srv, srv.BandwidthLimit)
}

// NewPoolLimited NewPool ile aynıdır; kbps > 0 ise tüm bağlantılar tek bir
// limiter'ı paylaşır (sınır sunucu başına toplam upload hızıdır).
func NewPoolLimited(srv config.Server, kbps int) (*Pool, error) {
	n := srv.MaxConnections
	if n <= 0 {
		n = 1
	}
	maxRetries := srv.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}
	if maxRetries == 0 && srv.MaxRetries == 0 {
		maxRetries = 2
	}

	lim := NewLimiter(kbps)
	sem := semFor(srv, n)
	p := &Pool{maxRetries: maxRetries, sem: sem, semLimit: n}

	for i := 0; i < n; i++ {
		if i == 0 {
			sem.acquire()
		} else if !sem.tryAcquire() {
			break // host limiti başka pool'larca dolu — eldekiyle devam
		}
		p.held++
		c, err := connectRetry421(srv, lim, i == 0)
		if err != nil {
			sem.release(1)
			p.held--
			if i == 0 {
				sem.unregister(n)
				return nil, err
			}
			break // ek bağlantı açılamadı (çoğunlukla 421) — eldeki bağlantılarla devam
		}
		p.clients = append(p.clients, c)
	}
	return p, nil
}

// connectRetry421 — ilk bağlantı 421 (çok fazla bağlantı) alırsa artan beklemeyle
// 4 kez tekrar dener (başka bir istemci bağlantısını bırakabilir). Ek bağlantılar denemez.
func connectRetry421(srv config.Server, lim *rate.Limiter, retry bool) (*Client, error) {
	c, err := ConnectLimited(srv, lim)
	if !retry {
		return c, err
	}
	for attempt := 1; attempt <= 4 && IsTooMany(err); attempt++ {
		time.Sleep(time.Duration(attempt*2) * time.Second)
		c, err = ConnectLimited(srv, lim)
	}
	return c, err
}

// Size açık bağlantı sayısı.
func (p *Pool) Size() int { return len(p.clients) }

// Close shuts down all connections in the pool.
func (p *Pool) Close() {
	for _, c := range p.clients {
		c.Close()
	}
	if p.sem != nil && p.held > 0 {
		p.sem.release(p.held)
		p.held = 0
	}
	if p.sem != nil && p.semLimit > 0 {
		p.sem.unregister(p.semLimit)
		p.semLimit = 0
	}
}

// UploadCh distributes tasks across workers and streams results one-by-one via a channel.
// The channel is closed when all uploads finish.
func (p *Pool) UploadCh(tasks []UploadTask) <-chan UploadResult {
	ch := make(chan UploadResult, len(p.clients))
	if len(tasks) == 0 {
		close(ch)
		return ch
	}

	jobs := make(chan UploadTask, len(tasks))
	var wg sync.WaitGroup
	for _, c := range p.clients {
		wg.Add(1)
		go func(client *Client) {
			defer wg.Done()
			for task := range jobs {
				ch <- p.uploadWithRetry(client, task)
			}
		}(c)
	}
	for _, t := range tasks {
		jobs <- t
	}
	close(jobs)
	go func() {
		wg.Wait()
		close(ch)
	}()
	return ch
}

// Upload distributes tasks across all workers and returns one result per task.
// Failed uploads are retried up to MaxRetries times on the same worker before
// being reported as failed.
func (p *Pool) Upload(tasks []UploadTask) []UploadResult {
	out := make([]UploadResult, 0, len(tasks))
	for r := range p.UploadCh(tasks) {
		out = append(out, r)
	}
	return out
}

func (p *Pool) uploadWithRetry(client *Client, task UploadTask) UploadResult {
	var lastErr error
	var retryErrs []error

	// Yedek: denemelerden önce bir kez (retry'da tekrar indirilmez)
	var backedUp, isNew bool
	var backupErr error
	if task.BackupDir != "" {
		existed, err := client.BackupTo(task.RelPath, filepath.Join(task.BackupDir, filepath.FromSlash(task.RelPath)))
		switch {
		case err != nil:
			backupErr = err
		case existed:
			backedUp = true
		default:
			isNew = true
		}
	}

	maxAttempts := p.maxRetries + 1
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		lastErr = client.Upload(task.LocalPath, task.RelPath)
		if lastErr == nil {
			var size int64
			if fi, statErr := os.Stat(task.LocalPath); statErr == nil {
				size = fi.Size()
			}
			return UploadResult{
				RelPath:   task.RelPath,
				Hash:      task.Hash,
				Size:      size,
				Attempts:  attempt,
				RetryErrs: retryErrs,
				BackedUp:  backedUp,
				IsNew:     isNew,
				BackupErr: backupErr,
			}
		}
		retryErrs = append(retryErrs, lastErr)
	}
	return UploadResult{
		RelPath:   task.RelPath,
		Err:       lastErr,
		Attempts:  maxAttempts,
		RetryErrs: retryErrs[:len(retryErrs)-1], // son hata zaten Err'da, RetryErrs'da tekrar etme
		BackedUp:  backedUp,
		IsNew:     isNew,
	}
}

// ── Bağlantı limiti tespiti ───────────────────────────────────────────────────

// ProbeMaxConnections sunucunun aynı anda kabul ettiği en fazla bağlantıyı bulur:
// bağlantılar birer birer açılır (giriş + NOOP ile doğrulanır) ta ki sunucu
// reddedene ya da limit sınırına gelene kadar; sonra hepsi kapatılır.
// Bazı sunucular fazla bağlantıyı kabul edip hemen düşürür → sonda tüm
// bağlantılar tekrar NOOP ile yoklanır, yalnızca hayatta kalanlar sayılır.
// progress her başarılı bağlantıda çağrılır.
func ProbeMaxConnections(srv config.Server, limit int, progress func(n int)) (int, error) {
	if limit <= 0 {
		limit = 32
	}
	var clients []*Client
	defer func() {
		for _, c := range clients {
			c.Close()
		}
	}()
	var firstErr error
	for i := 0; i < limit; i++ {
		c, err := Connect(srv)
		if err != nil {
			firstErr = err
			break
		}
		if err := c.Ping(); err != nil {
			c.Close()
			firstErr = err
			break
		}
		clients = append(clients, c)
		if progress != nil {
			progress(len(clients))
		}
	}
	if len(clients) == 0 {
		return 0, firstErr
	}
	// Sunucu limiti aşan bağlantıları sessizce kapatmış olabilir → tekrar yokla
	time.Sleep(1500 * time.Millisecond)
	alive := 0
	for _, c := range clients {
		if err := c.Ping(); err == nil {
			alive++
		}
	}
	return alive, nil
}

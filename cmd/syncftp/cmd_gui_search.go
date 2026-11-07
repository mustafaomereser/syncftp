package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	goftp "github.com/jlaffaye/ftp"

	"syncftp/internal/config"
	ftpclient "syncftp/internal/ftp"
)

// searchDirTimeout tek klasörün listelenmesi için süre: cevap gelmezse bağlantı kesilir, klasör
// "yanıt vermedi" diye atlanır ve işçi yeni bağlantıyla devam eder. Eskiden sınır yoktu: SFTP'de
// okuma zaman aşımı olmadığı için (yalnızca keepalive) cPanel'in .cagefs gibi sanal klasörlerinde
// cevapsız kalan bir LIST aramayı sonsuza kadar kilitliyordu ("1263 klasör tarandı…" diye donuyordu).
var searchDirTimeout = 20 * time.Second

// searchResult arama özeti.
type searchResult struct {
	Scanned int      // listelenen klasör
	Skipped []string // yanıt vermediği / okunamadığı için atlanan klasörler (tam yol)
	ListErr string   // ilk listeleme hatası (uyarı)
}

// searchRemoteTree — uzak dosya aramasının TEK motoru: CLI tarayıcısındaki "/" araması
// (runSearchWorker) ve paneldeki "Alt klasörlerde ara" bunu kullanır.
// max_connections kadar (en az 2, en çok 8) paralel KENDİ bağlantısıyla klasörler taranır; adı q
// içeren her girdi bulunduğu an onHit'e verilir (onHit false dönerse arama durur). onHit / onProgress
// işçilerden, tek kilit altında çağrılır (içlerinde bekleme yapma).
//
//   - Bağlantılar paralel açılır, her biri hazır olduğu an işçi olur (ilk sonuç beklemeden gelir).
//   - Her LIST'in süresi sınırlı (searchDirTimeout): aşılırsa o bağlantı kesilir (Abort), klasör
//     atlananlara yazılır, işçi yeni bağlantı açıp devam eder.
//   - LIST bağlantı hatası verirse işçi yeniden bağlanır ve klasörü BİR kez daha dener (eskiden kopmuş
//     bağlantıyla devam edip kalan klasörleri sessizce atlıyordu).
//   - Sembolik bağlar izlenmez (döngü yok); iptalde süren LIST'ler Abort ile kesilir.
//
// Çağıranın açık bağlantısı kullanılmaz: takılan LIST'i kesmek bağlantıyı öldürür.
func searchRemoteTree(ctx context.Context, srv config.Server, base, q string,
	onHit func(dir string, e *goftp.Entry) bool, onProgress func(scanned int, current string)) (searchResult, error) {
	q = strings.ToLower(q)
	conns := srv.MaxConnections
	if conns < 2 {
		conns = 2
	}
	if conns > 8 {
		conns = 8
	}

	var (
		mu         sync.Mutex
		cond       = sync.NewCond(&mu)
		res        searchResult
		stack      = []string{base}
		pending    = 1 // kuyrukta + işlenmekte olan klasör
		stopped    = false
		finished   = false // arama bitti; geç gelen bağlantılar kapatılıp atılır
		workers    = 0     // çalışan işçi
		connecting = 0     // süren bağlantı denemesi
		connected  = 0
		firstErr   error
		live       = map[*ftpclient.Client]bool{} // açık bağlantılar (iptalde Abort, sonda Close)
		retried    = map[string]bool{}
		inflight   = map[string]time.Time{} // şu an listelenen klasörler
	)
	// En uzun süredir beklenen klasör (takılan varsa o görünür)
	oldest := func() string {
		cur, at := "", time.Time{}
		for d, t := range inflight {
			if cur == "" || t.Before(at) {
				cur, at = d, t
			}
		}
		return cur
	}
	stop := func() { stopped = true; cond.Broadcast() }

	// connect ctx'e bağlı bağlantı açar (iptalde beklemez); açılanı live'a ekler.
	connect := func() *ftpclient.Client {
		if ctx.Err() != nil {
			return nil
		}
		type r struct {
			c   *ftpclient.Client
			err error
		}
		ch := make(chan r, 1)
		go func() { c, err := ftpclient.Connect(srv); ch <- r{c, err} }()
		select {
		case x := <-ch:
			mu.Lock()
			defer mu.Unlock()
			if x.c == nil {
				if x.err != nil && firstErr == nil {
					firstErr = x.err
				}
				return nil
			}
			if finished || stopped {
				go x.c.Close()
				return nil
			}
			live[x.c] = true
			return x.c
		case <-ctx.Done():
			go func() {
				if x := <-ch; x.c != nil {
					x.c.Close()
				}
			}()
			return nil
		}
	}
	drop := func(c *ftpclient.Client) {
		mu.Lock()
		delete(live, c)
		mu.Unlock()
		c.Abort()
		go c.Close()
	}
	// list süre sınırlı LIST: timedOut → bağlantı Abort edildi, kullanılamaz.
	list := func(c *ftpclient.Client, dir string) (entries []*goftp.Entry, err error, timedOut bool) {
		type r struct {
			e   []*goftp.Entry
			err error
		}
		ch := make(chan r, 1)
		go func() { e, err := c.List(dir); ch <- r{e, err} }()
		timer := time.NewTimer(searchDirTimeout)
		defer timer.Stop()
		select {
		case x := <-ch:
			return x.e, x.err, false
		case <-timer.C:
			c.Abort() // takılan LIST'i çöz
			<-ch
			return nil, nil, true
		case <-ctx.Done():
			c.Abort()
			<-ch
			return nil, ctx.Err(), false
		}
	}

	work := func(c *ftpclient.Client) {
		defer func() {
			mu.Lock()
			workers--
			cond.Broadcast()
			mu.Unlock()
		}()
		for {
			mu.Lock()
			for len(stack) == 0 && pending > 0 && !stopped {
				cond.Wait()
			}
			if stopped || len(stack) == 0 {
				mu.Unlock()
				break
			}
			dir := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			inflight[dir] = time.Now()
			if onProgress != nil {
				onProgress(res.Scanned, oldest())
			}
			mu.Unlock()

			entries, lerr, timedOut := list(c, dir)
			broken := timedOut || (lerr != nil && !ftpclient.IsNotFound(lerr) && ctx.Err() == nil)
			if broken {
				drop(c)
				c = nil
			}

			mu.Lock()
			delete(inflight, dir)
			if broken && !stopped {
				if !timedOut && !retried[dir] {
					// Bağlantı hatası: yeni bağlantıyla bir kez daha (pending değişmez — klasör hâlâ bekliyor)
					retried[dir] = true
					stack = append(stack, dir)
					cond.Signal()
					mu.Unlock()
					if c = connect(); c == nil {
						break
					}
					continue
				}
				res.Skipped = append(res.Skipped, dir)
				if res.ListErr == "" && lerr != nil {
					res.ListErr = lerr.Error()
				}
			}
			res.Scanned++
			for _, e := range entries {
				if stopped {
					break
				}
				if e == nil || e.Name == "." || e.Name == ".." {
					continue
				}
				if strings.Contains(strings.ToLower(e.Name), q) && !onHit(dir, e) {
					stop()
					break
				}
				if e.Type == goftp.EntryTypeFolder {
					stack = append(stack, path.Join(dir, e.Name))
					pending++
				}
			}
			pending--
			if onProgress != nil {
				onProgress(res.Scanned, oldest())
			}
			if pending == 0 {
				cond.Broadcast()
			} else {
				cond.Signal()
			}
			done := stopped || pending == 0
			mu.Unlock()
			if done {
				break
			}
			if c == nil { // zaman aşımıyla kesildi → yeni bağlantıyla devam
				if c = connect(); c == nil {
					break
				}
			}
		}
		if c != nil {
			mu.Lock()
			delete(live, c)
			mu.Unlock()
			c.Close()
		}
	}

	mu.Lock()
	connecting = conns
	mu.Unlock()
	for i := 0; i < conns; i++ {
		go func() {
			c := connect()
			mu.Lock()
			connecting--
			if c != nil && !(finished || stopped || pending == 0) {
				connected++
				workers++
				mu.Unlock()
				work(c)
				return
			}
			if c != nil {
				delete(live, c)
				go c.Close()
			}
			cond.Broadcast()
			mu.Unlock()
		}()
	}

	// İptal: dur, süren LIST'leri kes (yoksa tıklanan dosyanın önizlemesi aramanın bağlantılarını bekliyordu)
	fin := make(chan struct{})
	defer close(fin)
	go func() {
		select {
		case <-ctx.Done():
		case <-fin:
			return
		}
		mu.Lock()
		stop()
		var all []*ftpclient.Client
		for c := range live {
			all = append(all, c)
		}
		mu.Unlock()
		for _, c := range all {
			c.Abort()
		}
	}()

	mu.Lock()
	// Bitiş: işçi kalmadı VE (iş bitti / durduruldu / bağlanacak bağlantı da kalmadı)
	for !(workers == 0 && (stopped || pending == 0 || connecting == 0)) {
		cond.Wait()
	}
	finished = true
	noConn := connected == 0
	var rest []*ftpclient.Client
	for c := range live {
		rest = append(rest, c)
	}
	// Bağlantı kalmadığı için taranamayan klasörler de atlanmış sayılır
	if !stopped && pending > 0 {
		res.Skipped = append(res.Skipped, stack...)
	}
	out := res
	mu.Unlock()
	for _, c := range rest {
		c.Close()
	}
	if noConn && firstErr != nil {
		return out, firstErr
	}
	return out, nil
}

const (
	searchMaxHits = 2000
	searchTimeout = 5 * time.Minute
)

// GET /api/p/{p}/remote/search?server=&path=&q= — SSE: hit / progress / done
func (sv *guiServer) apiRemoteSearch(w http.ResponseWriter, r *http.Request) {
	_, srv, ok := sv.remoteServer(w, r)
	if !ok {
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	flusher, _ := w.(http.Flusher)
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	send := func(event string, v any) {
		b, _ := json.Marshal(v)
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, b)
		if flusher != nil {
			flusher.Flush()
		}
	}
	if q == "" {
		send("done", map[string]any{})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), searchTimeout)
	defer cancel()
	start := time.Now()
	rel := func(p string) string { return strings.TrimPrefix(strings.TrimPrefix(p, srv.RemotePath), "/") }
	// onHit/onProgress arama motorunun kilidi altında çağrılır → HTTP yazımı (yavaş istemci) orada
	// yapılmaz; olaylar kanala konur, bu goroutine yazar. Kanal doluysa sonuç beklemez, düşmez: sonuçlar
	// için ayrı tampon (hits) tutulur, ilerleme olayı düşebilir.
	type sseEvent struct {
		name string
		v    any
	}
	events := make(chan sseEvent, 1024)
	var lastProg time.Time
	// Son durum (sayaç + şu an listelenen klasör) — saniyede bir ayrıca gönderilir: hiçbir klasör
	// bitmeden beklenirken de ekranda "… bekleniyor" görünsün
	var progMu sync.Mutex
	progN, progCur := 0, ""
	hits := 0
	var res searchResult
	var err error
	go func() {
		defer close(events)
		res, err = searchRemoteTree(ctx, srv, guiRemotePath(srv, r.URL.Query().Get("path")), q,
			func(dir string, e *goftp.Entry) bool {
				select {
				case events <- sseEvent{"hit", map[string]any{"path": rel(path.Join(dir, e.Name)), "dir": e.Type == goftp.EntryTypeFolder, "size": e.Size, "time": e.Time}}:
				case <-ctx.Done():
					return false
				}
				hits++
				return hits < searchMaxHits
			},
			func(n int, cur string) {
				progMu.Lock()
				progN = n
				progCur = ""
				if cur != "" {
					progCur = "/" + rel(cur)
				}
				progMu.Unlock()
				if time.Since(lastProg) > 120*time.Millisecond { // sayaç saniyede ~8 kez
					lastProg = time.Now()
					ev := map[string]any{"dirs": n}
					if cur != "" {
						ev["cur"] = "/" + rel(cur)
					}
					select {
					case events <- sseEvent{"progress", ev}:
					default: // sayaç düşebilir
					}
				}
			})
	}()
	tick := time.NewTicker(time.Second)
	defer tick.Stop()
loop:
	for {
		select {
		case ev, ok := <-events:
			if !ok {
				break loop
			}
			send(ev.name, ev.v)
		case <-tick.C:
			progMu.Lock()
			send("progress", map[string]any{"dirs": progN, "cur": progCur})
			progMu.Unlock()
		}
	}
	if err != nil {
		send("done", map[string]any{"error": err.Error()})
		return
	}
	skipped := []string{}
	for i, s := range res.Skipped {
		if i == 20 {
			break
		}
		skipped = append(skipped, "/"+rel(s))
	}
	send("done", map[string]any{"truncated": hits >= searchMaxHits || ctx.Err() == context.DeadlineExceeded,
		"dirs": res.Scanned, "warn": res.ListErr, "skipped": len(res.Skipped), "skipped_dirs": skipped,
		"ms": time.Since(start).Milliseconds()})
}

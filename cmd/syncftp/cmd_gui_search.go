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

// searchRemoteTree — uzak dosya aramasının TEK motoru: CLI tarayıcısındaki "/" araması
// (runSearchWorker) ve paneldeki "Alt klasörlerde ara" bunu kullanır.
// max_connections kadar (en az 2, en çok 8) paralel bağlantıyla klasörler taranır; adı q içeren her
// girdi bulunduğu an onHit'e verilir (onHit false dönerse arama durur). primary verilirse ilk bağlantı
// olarak o kullanılır (CLI'nın açık bağlantısı), kapatılmaz. onHit/onProgress işçilerden, tek kilit
// altında çağrılır.
//
// Bağlantılar PARALEL açılır ve her biri hazır olduğu an işçi olarak katılır — eskiden 8 bağlantının
// hepsi sırayla açılmadan tek klasöre bakılmıyordu (panelde ilk sonuç saniyeler sonra geliyordu; CLI
// açık bağlantıyla hemen başladığı için daha hızlı görünüyordu). İptalde süren LIST'ler Abort ile kesilir.
func searchRemoteTree(ctx context.Context, srv config.Server, primary *ftpclient.Client, base, q string,
	onHit func(dir string, e *goftp.Entry) bool, onProgress func(scanned int)) (scanned int, listErr string, err error) {
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
		stack      = []string{base}
		pending    = 1 // kuyrukta + işlenmekte olan klasör
		stopped    = false
		finished   = false // arama bitti; geç gelen bağlantılar kapatılıp atılır
		workers    = 0     // çalışan işçi
		connecting = 0     // süren bağlantı denemesi
		connected  = 0
		firstErr   error
		opened     []*ftpclient.Client // bu fonksiyonun açtıkları (iptalde Abort, sonda Close)
	)
	stop := func() { stopped = true; cond.Broadcast() }

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
				return
			}
			dir := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			mu.Unlock()

			entries, lerr := c.List(dir)

			mu.Lock()
			scanned++
			if lerr != nil && !stopped && !ftpclient.IsNotFound(lerr) && listErr == "" { // iptalin kesme hatası uyarı değil
				listErr = lerr.Error()
			}
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
				onProgress(scanned)
			}
			if pending == 0 {
				cond.Broadcast()
			} else {
				cond.Signal()
			}
			mu.Unlock()
		}
	}

	if primary != nil {
		workers++
		connected++
		go work(primary)
	}
	need := conns
	if primary != nil {
		need--
	}
	connecting = need
	for i := 0; i < need; i++ {
		go func() {
			var c *ftpclient.Client
			var cerr error
			if ctx.Err() == nil {
				c, cerr = ftpclient.Connect(srv)
			}
			mu.Lock()
			connecting--
			switch {
			case c == nil:
				if cerr != nil && firstErr == nil {
					firstErr = cerr
				}
			case finished || stopped || pending == 0:
				go c.Close() // geç kaldı — iş yok
			default:
				opened = append(opened, c)
				connected++
				workers++
				go work(c)
			}
			cond.Broadcast()
			mu.Unlock()
		}()
	}

	// İptal: dur, süren LIST'leri kes (yoksa tıklanan dosyanın önizlemesi aramanın bağlantılarını bekliyordu)
	go func() {
		<-ctx.Done()
		mu.Lock()
		stop()
		list := append([]*ftpclient.Client(nil), opened...)
		mu.Unlock()
		for _, c := range list {
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
	toClose := opened
	mu.Unlock()
	for _, c := range toClose {
		c.Close()
	}
	if noConn && firstErr != nil {
		return 0, "", firstErr
	}
	return scanned, listErr, nil
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
	// onHit/onProgress arama motorunun kilidi altında çağrılır → HTTP yazımı (yavaş istemci) orada
	// yapılmaz; olaylar kanala konur, bu goroutine yazar.
	type sseEvent struct {
		name string
		v    any
	}
	events := make(chan sseEvent, 256)
	var lastProg time.Time
	hits := 0
	var scanned int
	var listErr string
	var err error
	go func() {
		defer close(events)
		scanned, listErr, err = searchRemoteTree(ctx, srv, nil, guiRemotePath(srv, r.URL.Query().Get("path")), q,
			func(dir string, e *goftp.Entry) bool {
				rel := strings.TrimPrefix(strings.TrimPrefix(path.Join(dir, e.Name), srv.RemotePath), "/")
				select {
				case events <- sseEvent{"hit", map[string]any{"path": rel, "dir": e.Type == goftp.EntryTypeFolder, "size": e.Size, "time": e.Time}}:
				case <-ctx.Done():
					return false
				}
				hits++
				return hits < searchMaxHits
			},
			func(n int) {
				if time.Since(lastProg) > 120*time.Millisecond { // sayaç saniyede ~8 kez
					lastProg = time.Now()
					select {
					case events <- sseEvent{"progress", map[string]any{"dirs": n}}:
					default: // sayaç düşebilir
					}
				}
			})
	}()
	for ev := range events {
		send(ev.name, ev.v)
	}
	if err != nil {
		send("done", map[string]any{"error": err.Error()})
		return
	}
	send("done", map[string]any{"truncated": hits >= searchMaxHits || ctx.Err() == context.DeadlineExceeded,
		"dirs": scanned, "warn": listErr, "ms": time.Since(start).Milliseconds()})
}

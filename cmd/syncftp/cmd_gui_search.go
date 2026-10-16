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
	var clients, opened []*ftpclient.Client
	if primary != nil {
		clients = append(clients, primary)
	}
	for len(clients) < conns {
		c, cerr := ftpclient.Connect(srv)
		if cerr != nil {
			if len(clients) == 0 {
				return 0, "", cerr
			}
			break // sunucu daha fazla bağlantıya izin vermiyor — açılanlarla devam
		}
		clients = append(clients, c)
		opened = append(opened, c)
	}
	defer func() {
		for _, c := range opened {
			c.Close()
		}
	}()

	var (
		mu      sync.Mutex
		cond    = sync.NewCond(&mu)
		stack   = []string{base}
		pending = 1 // kuyrukta + işlenmekte olan klasör
		stopped = false
	)
	stop := func() { stopped = true; cond.Broadcast() }
	go func() {
		<-ctx.Done()
		mu.Lock()
		stop()
		mu.Unlock()
	}()

	var wg sync.WaitGroup
	for _, c := range clients {
		wg.Add(1)
		go func(c *ftpclient.Client) {
			defer wg.Done()
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
				if lerr != nil && !ftpclient.IsNotFound(lerr) && listErr == "" {
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
		}(c)
	}
	wg.Wait()
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
	var lastProg time.Time
	hits := 0
	scanned, listErr, err := searchRemoteTree(ctx, srv, nil, guiRemotePath(srv, r.URL.Query().Get("path")), q,
		func(dir string, e *goftp.Entry) bool {
			rel := strings.TrimPrefix(strings.TrimPrefix(path.Join(dir, e.Name), srv.RemotePath), "/")
			send("hit", map[string]any{"path": rel, "dir": e.Type == goftp.EntryTypeFolder, "size": e.Size, "time": e.Time})
			hits++
			return hits < searchMaxHits
		},
		func(n int) {
			if time.Since(lastProg) > 120*time.Millisecond { // sayaç saniyede ~8 kez
				lastProg = time.Now()
				send("progress", map[string]any{"dirs": n})
			}
		})
	if err != nil {
		send("done", map[string]any{"error": err.Error()})
		return
	}
	send("done", map[string]any{"truncated": hits >= searchMaxHits || ctx.Err() == context.DeadlineExceeded,
		"dirs": scanned, "warn": listErr, "ms": time.Since(start).Milliseconds()})
}

package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// apiJobsLive birden çok işin log akışını TEK SSE bağlantısında taşır.
//
// İş başına ayrı EventSource açılınca tarayıcının sunucu başına 6 bağlantı sınırı doluyordu
// (8 watch açınca panelin diğer istekleri sıraya girip sayfa "yükleniyor"da kalıyordu).
//
// ?ids=j1:120,j2:0 — iş kimliği : istemcinin o işten zaten aldığı satır sayısı (j.Lines sayacı);
// yeniden bağlanınca yalnızca eksik satırlar gönderilir.
// Olaylar: "line" {id, l}  ·  "done" iş anlık görüntüsü (id alanıyla)  ·  "gone" {id} (iş artık yok).
func (sv *guiServer) apiJobsLive(w http.ResponseWriter, r *http.Request) {
	fl, ok := w.(http.Flusher)
	if !ok {
		guiErr(w, 500, errors.New("sse"))
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)

	type ev struct {
		name string
		data string
	}
	out := make(chan ev, 1024)
	ctx := r.Context()
	send := func(e ev) {
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", e.name, e.data)
	}

	for _, part := range strings.Split(r.URL.Query().Get("ids"), ",") {
		id, have, _ := strings.Cut(strings.TrimSpace(part), ":")
		if id == "" {
			continue
		}
		j := sv.jobs.get(id)
		if j == nil {
			send(ev{"gone", mustJSON(map[string]string{"id": id})})
			continue
		}
		seen, _ := strconv.Atoi(have)

		ch := make(chan string, 512)
		j.mu.Lock()
		// Tamponda j.Lines-len(lines) … j.Lines aralığı duruyor; istemcinin görmediği kısmı gönder
		first := j.Lines - len(j.lines)
		from := seen - first
		if from < 0 {
			from = 0
		}
		var backlog []string
		if from < len(j.lines) {
			backlog = append(backlog, j.lines[from:]...)
		}
		j.subs[ch] = struct{}{}
		j.mu.Unlock()
		for _, l := range backlog {
			send(ev{"line", mustJSON(map[string]string{"id": id, "l": l})})
		}

		go func(j *job, ch chan string) {
			defer func() {
				j.mu.Lock()
				delete(j.subs, ch)
				j.mu.Unlock()
			}()
			line := func(l string) ev { return ev{"line", mustJSON(map[string]string{"id": j.ID, "l": l})} }
			for {
				select {
				case l := <-ch:
					select {
					case out <- line(l):
					case <-ctx.Done():
						return
					}
				case <-j.done:
					for drained := false; !drained; {
						select {
						case l := <-ch:
							select {
							case out <- line(l):
							case <-ctx.Done():
								return
							}
						default:
							drained = true
						}
					}
					select {
					case out <- ev{"done", mustJSON(j.snapshot())}:
					case <-ctx.Done():
					}
					return
				case <-ctx.Done():
					return
				}
			}
		}(j, ch)
	}
	fl.Flush()

	ping := time.NewTicker(15 * time.Second)
	defer ping.Stop()
	for {
		select {
		case e := <-out:
			send(e)
			// Biriken olayları tek flush ile gönder
			for drained := false; !drained; {
				select {
				case e2 := <-out:
					send(e2)
				default:
					drained = true
				}
			}
			fl.Flush()
		case <-ping.C:
			fmt.Fprint(w, ": ping\n\n")
			fl.Flush()
		case <-ctx.Done():
			return
		}
	}
}

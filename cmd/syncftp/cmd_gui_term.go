package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"syncftp/internal/lang"
	"syncftp/internal/projects"
)

// ══════════════════════════════════════════════════════════════════════════════
// Panel içi terminal: syncftp shell'i (argümansız syncftp) gerçek bir sözde terminalde
// (Windows: ConPTY, diğerleri: pty) çalıştırır; çıktı SSE ile xterm.js'e akar, tuşlar POST ile gelir.
// Oturum proje başına bir tane (proje yoksa genel); sayfadan ayrılınca süreç yaşamaya devam eder,
// geri gelince son 512 KB yeniden oynatılır.
// ══════════════════════════════════════════════════════════════════════════════

type termPTY interface {
	Read([]byte) (int, error)
	Write([]byte) (int, error)
	Resize(cols, rows int) error
	Wait() error
	Close() error
}

const termBacklogMax = 512 * 1024

type termSession struct {
	ID       string    `json:"id"`
	Project  string    `json:"project"`
	Dir      string    `json:"dir"`
	Started  time.Time `json:"started"`
	Exited   bool      `json:"exited"`
	ExitInfo string    `json:"exit_info,omitempty"`

	pty     termPTY
	mu      sync.Mutex
	backlog []byte
	subs    map[chan []byte]struct{}
	done    chan struct{}
}

var (
	termsMu sync.Mutex
	terms   = map[string]*termSession{} // id → oturum
)

func (sv *guiServer) registerTermAPI(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/terms", sv.apiTermList)
	mux.HandleFunc("POST /api/terms", sv.apiTermOpen)
	mux.HandleFunc("GET /api/terms/{id}/stream", sv.apiTermStream)
	mux.HandleFunc("POST /api/terms/{id}/input", sv.apiTermInput)
	mux.HandleFunc("POST /api/terms/{id}/resize", sv.apiTermResize)
	mux.HandleFunc("DELETE /api/terms/{id}", sv.apiTermClose)
}

func (s *termSession) pump() {
	buf := make([]byte, 32*1024)
	for {
		n, err := s.pty.Read(buf)
		if n > 0 {
			chunk := append([]byte(nil), buf[:n]...)
			s.mu.Lock()
			s.backlog = append(s.backlog, chunk...)
			if over := len(s.backlog) - termBacklogMax; over > 0 {
				s.backlog = append([]byte(nil), s.backlog[over:]...)
			}
			for ch := range s.subs {
				select {
				case ch <- chunk:
				default: // yavaş istemci: bu parçayı kaçırır (backlog'da duruyor)
				}
			}
			s.mu.Unlock()
		}
		if err != nil {
			return
		}
	}
}

func (s *termSession) wait() {
	err := s.pty.Wait()
	s.mu.Lock()
	s.Exited = true
	if err != nil {
		s.ExitInfo = err.Error()
	}
	s.mu.Unlock()
	time.Sleep(150 * time.Millisecond) // son çıktı parçaları yayılsın
	close(s.done)
	_ = s.pty.Close()
}

// apiTermOpen {project, cols, rows, restart} — projenin çalışan oturumunu döndürür, yoksa açar.
func (sv *guiServer) apiTermOpen(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Project string `json:"project"`
		Cols    int    `json:"cols"`
		Rows    int    `json:"rows"`
		Restart bool   `json:"restart"`
	}
	if err := guiDecode(r, &req); err != nil {
		guiErr(w, 400, err)
		return
	}
	dir := sv.startDir
	if req.Project != "" {
		p, ok := projects.Find(req.Project)
		if !ok {
			guiErr(w, 404, fmt.Errorf(lang.L.ProjectNotFoundFmt, req.Project))
			return
		}
		dir = p.Path
	}
	if dir == "" {
		dir, _ = os.Getwd()
	}
	termsMu.Lock()
	for id, s := range terms {
		if s.Project == req.Project {
			s.mu.Lock()
			exited := s.Exited
			s.mu.Unlock()
			if !exited && !req.Restart {
				termsMu.Unlock()
				guiJSON(w, s)
				return
			}
			_ = s.pty.Close()
			delete(terms, id)
		}
	}
	termsMu.Unlock()

	cols, rows := req.Cols, req.Rows
	if cols < 20 {
		cols = 120
	}
	if rows < 5 {
		rows = 30
	}
	exe, err := os.Executable()
	if err != nil {
		guiErr(w, 500, err)
		return
	}
	env := append(os.Environ(),
		"SYNCFTP_EMBEDDED_SHELL=1", // main(): ownsConsole → GUI açma, shell'de kal
		"TERM=xterm-256color", "COLORTERM=truecolor",
		"SYNCFTP_LANG="+lang.Current())
	pty, err := startTermPTY(exe, nil, dir, env, cols, rows)
	if err != nil {
		guiErr(w, 500, err)
		return
	}
	s := &termSession{ID: uuid4(), Project: req.Project, Dir: dir, Started: time.Now(), pty: pty,
		subs: map[chan []byte]struct{}{}, done: make(chan struct{})}
	termsMu.Lock()
	terms[s.ID] = s
	termsMu.Unlock()
	go s.pump()
	go s.wait()
	guiJSON(w, s)
}

func termByID(r *http.Request) *termSession {
	termsMu.Lock()
	defer termsMu.Unlock()
	return terms[r.PathValue("id")]
}

func (sv *guiServer) apiTermList(w http.ResponseWriter, r *http.Request) {
	termsMu.Lock()
	out := make([]*termSession, 0, len(terms))
	for _, s := range terms {
		out = append(out, s)
	}
	termsMu.Unlock()
	sort.Slice(out, func(i, j int) bool { return out[i].Started.Before(out[j].Started) })
	guiJSON(w, out)
}

// apiTermStream SSE: önce backlog, sonra canlı çıktı (base64); süreç bitince "exit" olayı.
func (sv *guiServer) apiTermStream(w http.ResponseWriter, r *http.Request) {
	s := termByID(r)
	if s == nil {
		guiErr(w, 404, errors.New("terminal"))
		return
	}
	fl, ok := w.(http.Flusher)
	if !ok {
		guiErr(w, 500, errors.New("stream"))
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	ch := make(chan []byte, 256)
	s.mu.Lock()
	back := append([]byte(nil), s.backlog...)
	s.subs[ch] = struct{}{}
	s.mu.Unlock()
	defer func() { s.mu.Lock(); delete(s.subs, ch); s.mu.Unlock() }()
	send := func(b []byte) {
		fmt.Fprintf(w, "data: %s\n\n", base64.StdEncoding.EncodeToString(b))
	}
	if len(back) > 0 {
		send(back)
	}
	fl.Flush()
	ping := time.NewTicker(20 * time.Second)
	defer ping.Stop()
	for {
		select {
		case b := <-ch:
			send(b)
			// bekleyen parçaları tek seferde boşalt
			for more := true; more; {
				select {
				case b2 := <-ch:
					send(b2)
				default:
					more = false
				}
			}
			fl.Flush()
		case <-s.done:
			for more := true; more; {
				select {
				case b2 := <-ch:
					send(b2)
				default:
					more = false
				}
			}
			fmt.Fprint(w, "event: exit\ndata: {}\n\n")
			fl.Flush()
			return
		case <-ping.C:
			fmt.Fprint(w, ": ping\n\n")
			fl.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func (sv *guiServer) apiTermInput(w http.ResponseWriter, r *http.Request) {
	s := termByID(r)
	if s == nil {
		guiErr(w, 404, errors.New("terminal"))
		return
	}
	var req struct {
		Data string `json:"data"`
	}
	if err := guiDecode(r, &req); err != nil {
		guiErr(w, 400, err)
		return
	}
	if _, err := s.pty.Write([]byte(req.Data)); err != nil {
		guiErr(w, 410, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (sv *guiServer) apiTermResize(w http.ResponseWriter, r *http.Request) {
	s := termByID(r)
	if s == nil {
		guiErr(w, 404, errors.New("terminal"))
		return
	}
	var req struct {
		Cols int `json:"cols"`
		Rows int `json:"rows"`
	}
	if err := guiDecode(r, &req); err != nil || req.Cols < 2 || req.Rows < 2 {
		guiErr(w, 400, errors.New("size"))
		return
	}
	_ = s.pty.Resize(req.Cols, req.Rows)
	w.WriteHeader(http.StatusNoContent)
}

func (sv *guiServer) apiTermClose(w http.ResponseWriter, r *http.Request) {
	s := termByID(r)
	if s == nil {
		guiErr(w, 404, errors.New("terminal"))
		return
	}
	termsMu.Lock()
	delete(terms, s.ID)
	termsMu.Unlock()
	_ = s.pty.Close()
	guiJSON(w, true)
}

// closeAllTerms panel kapanırken çağrılır (süreçler yetim kalmasın).
func closeAllTerms() {
	termsMu.Lock()
	defer termsMu.Unlock()
	for id, s := range terms {
		_ = s.pty.Close()
		delete(terms, id)
	}
}


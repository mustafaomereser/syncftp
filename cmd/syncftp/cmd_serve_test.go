package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"syncftp/internal/config"
)

func TestServeGuard(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	cases := []struct {
		host, origin, token, sent string
		want                      int
	}{
		{"127.0.0.1:8080", "", "", "", 200},
		{"localhost:8080", "http://localhost:3000", "", "", 200},
		{"[::1]:8080", "", "", "", 200},
		{"evil.example:8080", "", "", "", 403},                  // DNS rebinding
		{"127.0.0.1:8080", "https://evil.example", "", "", 403}, // başka site
		{"127.0.0.1:8080", "", "s3cret", "", 401},
		{"127.0.0.1:8080", "", "s3cret", "s3cret", 200},
	}
	for i, c := range cases {
		req := httptest.NewRequest("GET", "/api/servers", nil)
		req.Host = c.host
		if c.origin != "" {
			req.Header.Set("Origin", c.origin)
		}
		if c.sent != "" {
			req.Header.Set("X-Token", c.sent)
		}
		rec := httptest.NewRecorder()
		serveGuard(ok, c.token).ServeHTTP(rec, req)
		if rec.Code != c.want {
			t.Errorf("#%d %+v: kod %d, beklenen %d", i, c, rec.Code, c.want)
		}
		if c.origin != "" && c.want == 200 && rec.Header().Get("Access-Control-Allow-Origin") != c.origin {
			t.Errorf("#%d: CORS yerel Origin'e yansıtılmadı", i)
		}
	}
}

func TestServeRemotePath(t *testing.T) {
	srv := config.Server{RemotePath: "/public_html"}
	for in, want := range map[string]string{
		"a/b.php":              "/public_html/a/b.php",
		"../../etc/passwd":     "/public_html/etc/passwd",
		"/public_html/x.php":   "/public_html/x.php",
		"/etc/passwd":          "/public_html/etc/passwd",
		"/public_html/../etc":  "/public_html/etc",
		"/public_htmlX/secret": "/public_html/public_htmlX/secret",
	} {
		if got := serveRemotePath(srv, in); got != want {
			t.Errorf("%q → %q, beklenen %q", in, got, want)
		}
	}
}

package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"syncftp/internal/config"
)

func TestMatchTriggers(t *testing.T) {
	off := false
	list := []config.Trigger{
		{Name: "migrate", Paths: []string{"database/migrations/"}, Remote: []string{"php artisan migrate --force"}},
		{Name: "sql", Paths: []string{"*.sql"}, Run: []string{"echo sql"}},
		{Name: "her", Notify: "https://example.com/n"},
		{Name: "once", When: "before", Paths: []string{"composer.lock"}, Run: []string{"x"}},
		{Name: "kapali", Enabled: &off, Run: []string{"x"}},
		{Name: "eylemsiz", Paths: []string{"a"}},
	}
	files := []string{"database/migrations/2026_01_01_x.php", "app/a.php", "dump/seed.sql", "composer.lock"}
	got := map[string]int{}
	for _, f := range matchTriggers(list, false, files) {
		got[f.t.Name] = len(f.files)
	}
	want := map[string]int{"migrate": 1, "sql": 1, "her": 4}
	if len(got) != len(want) {
		t.Fatalf("after: %v", got)
	}
	for k, v := range want {
		if got[k] != v {
			t.Fatalf("after %s: %d, want %d (%v)", k, got[k], v, got)
		}
	}
	before := matchTriggers(list, true, files)
	if len(before) != 1 || before[0].t.Name != "once" {
		t.Fatalf("before: %+v", before)
	}
	if len(matchTriggers(list, false, []string{"app/a.php"})) != 1 { // yalnızca "her"
		t.Fatal("app/a.php yalnızca kalıpsız tetikleyiciyi çalıştırmalı")
	}
}

func TestTriggerExpand(t *testing.T) {
	v := triggerVars{project: "site", server: "prod", trigger: "m", files: []string{"a b.php", `c"d.sql`}}
	if got := v.expand("https://x/h?s={server}&f={first}&n={count}", true, false); got != "https://x/h?s=prod&f=a+b.php&n=2" {
		t.Fatalf("url: %s", got)
	}
	body := v.expand(`{"t":"{trigger}","files":{files_json},"txt":"{files}"}`, false, true)
	var m map[string]any
	if err := json.Unmarshal([]byte(body), &m); err != nil {
		t.Fatalf("gövde JSON değil: %s (%v)", body, err)
	}
	if m["txt"] != "a b.php\n"+`c"d.sql` || len(m["files"].([]any)) != 2 {
		t.Fatalf("gövde: %v", m)
	}
}

func TestTriggerHTTPAndValidate(t *testing.T) {
	var gotMethod, gotAuth, gotCT string
	var gotBody map[string]any
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotAuth, gotCT = r.Method, r.Header.Get("Authorization"), r.Header.Get("Content-Type")
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		if strings.Contains(r.URL.RawQuery, "fail") {
			w.WriteHeader(500)
		}
		_, _ = w.Write([]byte("tamam"))
	}))
	defer ts.Close()
	tr := config.Trigger{Name: "migrate", HTTPURL: ts.URL + "/deploy?token=gizli", HTTPHeaders: []string{"Authorization: Bearer abc"}}
	var out bytes.Buffer
	if err := runTrigger(&out, t.TempDir(), "site", config.Server{Name: "prod"}, tr, []string{"database/migrations/x.php"}); err != nil {
		t.Fatal(err)
	}
	if gotMethod != "POST" || gotAuth != "Bearer abc" || gotCT != "application/json" || gotBody["trigger"] != "migrate" || gotBody["count"] != float64(1) {
		t.Fatalf("istek: %s %s %s %v", gotMethod, gotAuth, gotCT, gotBody)
	}
	if strings.Contains(out.String(), "gizli") {
		t.Fatalf("log'a jeton sızdı: %s", out.String())
	}
	tr.HTTPURL = ts.URL + "/?fail=1"
	if err := runTrigger(&out, t.TempDir(), "site", config.Server{Name: "prod"}, tr, nil); err == nil {
		t.Fatal("HTTP 500 hata dönmeli")
	}

	bad := []config.Trigger{
		{Name: "", Run: []string{"x"}},
		{Name: "a", When: "sonra", Run: []string{"x"}},
		{Name: "a"},
		{Name: "a", HTTPURL: "ftp://x"},
		{Name: "a", HTTPURL: "https://x", HTTPMethod: "FETCH"},
		{Name: "a", HTTPURL: "https://x", HTTPHeaders: []string{"başlıksız"}},
	}
	for i, b := range bad {
		if config.ValidateTriggers([]config.Trigger{b}, "sync") == nil {
			t.Fatalf("geçersiz tetikleyici %d kabul edildi: %+v", i, b)
		}
	}
	if err := config.ValidateTriggers([]config.Trigger{{Name: "a", HTTPURL: "https://x/{server}/m", When: "before"}, {Name: "A", Run: []string{"y"}}}, "sync"); err == nil {
		t.Fatal("aynı ad (büyük/küçük harf) reddedilmeli")
	}
	if err := config.ValidateTriggers([]config.Trigger{{Name: "a", HTTPURL: "https://x/{server}/m", When: "before"}}, "sync"); err != nil {
		t.Fatal(err)
	}

	cfg := config.Config{Sync: config.Sync{Triggers: []config.Trigger{tr}}, Servers: []config.Server{{Name: "p", Triggers: []config.Trigger{tr}}}}
	red := cfg.Redacted(config.MaskStars, true)
	if red.Sync.Triggers[0].HTTPHeaders[0] != "Authorization: ***" || red.Servers[0].Triggers[0].HTTPHeaders[0] != "Authorization: ***" {
		t.Fatalf("maske: %v %v", red.Sync.Triggers[0].HTTPHeaders, red.Servers[0].Triggers[0].HTTPHeaders)
	}
	if cfg.Sync.Triggers[0].HTTPHeaders[0] != "Authorization: Bearer abc" {
		t.Fatal("Redacted asıl config'i değiştirdi")
	}
}

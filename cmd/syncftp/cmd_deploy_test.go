package main

import (
	"net/http/httptest"
	"strings"
	"testing"

	"syncftp/internal/config"
)

func TestVerifyHookProviders(t *testing.T) {
	body := []byte(`{"ref":"refs/heads/main"}`)
	d := config.Deploy{Secret: "s3cret"}

	// GitHub: HMAC-SHA256
	d.Provider = "github"
	r := httptest.NewRequest("POST", "/hook/p/t", nil)
	r.Header.Set("X-Hub-Signature-256", "sha256="+hmacSHA256Hex("s3cret", body))
	r.Header.Set("X-GitHub-Event", "push")
	if ev, err := verifyHook(d, r, body); err != nil || ev != "push" {
		t.Fatalf("github valid: %q %v", ev, err)
	}
	r.Header.Set("X-Hub-Signature-256", "sha256="+hmacSHA256Hex("wrong", body))
	if _, err := verifyHook(d, r, body); err == nil {
		t.Fatal("github bad signature accepted")
	}

	// GitLab: düz token
	d.Provider = "gitlab"
	r = httptest.NewRequest("POST", "/", nil)
	r.Header.Set("X-Gitlab-Token", "s3cret")
	r.Header.Set("X-Gitlab-Event", "Push Hook")
	if ev, err := verifyHook(d, r, body); err != nil || !isPushEvent("gitlab", ev) {
		t.Fatalf("gitlab: %q %v", ev, err)
	}

	// Gitea: önek'siz hex HMAC
	d.Provider = "gitea"
	r = httptest.NewRequest("POST", "/", nil)
	r.Header.Set("X-Gitea-Signature", hmacSHA256Hex("s3cret", body))
	r.Header.Set("X-Gitea-Event", "push")
	if _, err := verifyHook(d, r, body); err != nil {
		t.Fatalf("gitea: %v", err)
	}

	// Generic: header, query ya da Bearer
	d.Provider = "generic"
	r = httptest.NewRequest("POST", "/?token=s3cret", nil)
	if _, err := verifyHook(d, r, body); err != nil {
		t.Fatalf("generic query: %v", err)
	}
	r = httptest.NewRequest("POST", "/", nil)
	r.Header.Set("Authorization", "Bearer s3cret")
	if _, err := verifyHook(d, r, body); err != nil {
		t.Fatalf("generic bearer: %v", err)
	}
	r = httptest.NewRequest("POST", "/", nil)
	if _, err := verifyHook(d, r, body); err == nil {
		t.Fatal("generic without token accepted")
	}

	// Secret'sız tetikleyici asla kabul etmez
	if _, err := verifyHook(config.Deploy{Provider: "generic"}, httptest.NewRequest("POST", "/?token=", nil), body); err == nil {
		t.Fatal("empty secret must be rejected")
	}
}

func TestPushInfo(t *testing.T) {
	gh := `{"ref":"refs/heads/main","after":"abcdef1234567","head_commit":{"id":"abcdef1234567","message":"fix: login\n\nlong body"}}`
	b, c := pushInfo([]byte(gh))
	if b != "main" || c != "abcdef1 fix: login" {
		t.Fatalf("github: %q %q", b, c)
	}
	bb := `{"push":{"changes":[{"new":{"name":"develop","target":{"hash":"1234567890","message":"wip"}}}]}}`
	b, c = pushInfo([]byte(bb))
	if b != "develop" || c != "1234567 wip" {
		t.Fatalf("bitbucket: %q %q", b, c)
	}
	if !isPushEvent("bitbucket", "repo:push") || isPushEvent("github", "issues") {
		t.Fatal("event filter")
	}
}

func TestWebhookPayloadFormats(t *testing.T) {
	files := []string{"a.php", "b.css"}
	if p := webhookPayload("https://hooks.slack.com/services/x", "prod", 2, 0, files); !strings.HasPrefix(p, `{"text":`) || !strings.Contains(p, `a.php\nb.css`) {
		t.Fatalf("slack: %s", p)
	}
	if p := webhookPayload("https://discord.com/api/webhooks/1/x", "prod", 2, 1, files); !strings.HasPrefix(p, `{"content":`) {
		t.Fatalf("discord: %s", p)
	}
	if p := webhookPayload("https://example.com/hook", "prod", 2, 0, files); p != `{"server":"prod","uploaded":2,"failed":0,"files":["a.php","b.css"]}` {
		t.Fatalf("generic: %s", p)
	}
}

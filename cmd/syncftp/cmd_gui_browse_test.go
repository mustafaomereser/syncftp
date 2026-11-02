package main

import (
	"testing"

	"syncftp/internal/config"
)

func TestBrowseTarget(t *testing.T) {
	cfg := &config.Config{
		Connections: []config.Connection{{Name: "hesap", Protocol: "sftp", Host: "h", User: "u", Password: "p"}},
		Servers:     []config.Server{{Name: "prod", Host: "x", RemotePath: "/public_html"}},
	}

	// Profil olduğu gibi döner
	if s, ok := browseTarget(cfg, "prod"); !ok || s.RemotePath != "/public_html" {
		t.Fatalf("profile: %+v %v", s, ok)
	}

	// Bağlantı: kökten gezilir, bilgiler bağlantıdan, port protokole göre
	s, ok := browseTarget(cfg, connTargetPrefix+"hesap")
	if !ok || s.Name != "hesap" || s.RemotePath != "/" || s.Host != "h" || s.User != "u" || s.Password != "p" || s.Port != 22 {
		t.Fatalf("connection: %+v %v", s, ok)
	}

	// Öneksiz bağlantı adı ve bilinmeyen bağlantı bulunmaz
	if _, ok := browseTarget(cfg, "hesap"); ok {
		t.Fatal("connection name without the prefix must not resolve")
	}
	if _, ok := browseTarget(cfg, connTargetPrefix+"yok"); ok {
		t.Fatal("unknown connection must not resolve")
	}

	// Sync ve diğer profil işlerinin çözümleyicisi bağlantıyı kabul etmez
	if _, err := selectServers(cfg, []string{connTargetPrefix + "hesap"}); err == nil {
		t.Fatal("selectServers must reject a connection target")
	}
}

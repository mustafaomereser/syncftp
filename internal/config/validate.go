package config

import (
	"errors"
	"fmt"
	"strings"

	"syncftp/internal/lang"
)

// Config doğrulaması — panel (kaydet / içe aktar), CLI config ekranı ve bağlantı kütüphanesi ortak.
// Adlar büyük/küçük harf duyarsız benzersiz olmalı (aksi halde "Prod" ile "prod" karışır).

func validateConnections(conns []Connection) error {
	seen := map[string]bool{}
	for _, c := range conns {
		n := strings.ToLower(strings.TrimSpace(c.Name))
		if n == "" {
			return errors.New(lang.L.GuiValConnName)
		}
		if seen[n] {
			return fmt.Errorf(lang.L.GuiValDupFmt, c.Name)
		}
		seen[n] = true
		if strings.TrimSpace(c.Host) == "" {
			return fmt.Errorf(lang.L.GuiValHostFmt, c.Name)
		}
	}
	return nil
}

// Validate kaydetmeden önce config'in tutarlılığını denetler.
func (c *Config) Validate() error {
	if err := validateConnections(c.Connections); err != nil {
		return err
	}
	conns := map[string]bool{}
	for _, cn := range c.Connections {
		conns[cn.Name] = true // referans çözümü config.Load gibi birebir ad
	}
	seen := map[string]bool{}
	for _, s := range c.Servers {
		n := strings.ToLower(strings.TrimSpace(s.Name))
		if n == "" {
			return errors.New(lang.L.GuiValServerName)
		}
		if seen[n] {
			return fmt.Errorf(lang.L.GuiValDupFmt, s.Name)
		}
		seen[n] = true
		if s.Connection != "" && !conns[s.Connection] {
			return fmt.Errorf(lang.L.GuiValConnMissingFmt, s.Name, s.Connection)
		}
		if s.Connection == "" && strings.TrimSpace(s.Host) == "" {
			return fmt.Errorf(lang.L.GuiValHostFmt, s.Name)
		}
		if strings.TrimSpace(s.RemotePath) == "" {
			return fmt.Errorf(lang.L.GuiValRemoteFmt, s.Name)
		}
	}
	return nil
}

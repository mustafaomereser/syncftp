//go:build !windows

package main

// Windows dışında tray yok (macOS systray cgo ister); panel penceresi kapanınca
// süreç idle watchdog ile kapanır.
func trayAvailable() bool { return false }

func (sv *guiServer) runTray(string) { <-sv.quit }

func autostartEnabled() bool { return false }

func setAutostart(bool) error { return nil }

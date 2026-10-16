//go:build !windows

package lang

// systemUILangTurkish Unix'te dil yalnızca locale env değişkenlerinden algılanır.
func systemUILangTurkish() bool { return false }

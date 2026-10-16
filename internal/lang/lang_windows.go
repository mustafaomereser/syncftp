//go:build windows

package lang

import "syscall"

// systemUILangTurkish Windows arayüz dili Türkçe mi (kernel32 GetUserDefaultUILanguage).
func systemUILangTurkish() bool {
	dll, err := syscall.LoadDLL("kernel32.dll")
	if err != nil {
		return false
	}
	defer dll.Release()
	proc, err := dll.FindProc("GetUserDefaultUILanguage")
	if err != nil {
		return false
	}
	langID, _, _ := proc.Call()
	// LANGID: bits 0-9 = primary language; LANG_TURKISH = 0x1F (31)
	return langID&0x3FF == 0x1F
}

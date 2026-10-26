// Package fsutil küçük dosya sistemi yardımcıları.
package fsutil

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// WriteFileAtomic data'yı önce aynı dizinde benzersiz bir geçici dosyaya yazar, diske
// indirir, sonra rename ile yerine koyar → yazım yarıda kesilirse (çökme, disk dolu,
// eş zamanlı yazıcı) hedefte yarım/bozuk JSON kalmaz. Üst dizin yoksa hata döner.
func WriteFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	f, err := os.CreateTemp(dir, "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	tmp := f.Name()
	ok := false
	defer func() {
		if !ok {
			os.Remove(tmp)
		}
	}()
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp, perm); err != nil {
		return err
	}
	// Windows'ta hedef o an başka bir süreçte açıksa rename kısa süre reddedilebilir → birkaç deneme
	for attempt := 0; ; attempt++ {
		err = os.Rename(tmp, path)
		if err == nil || attempt >= 4 {
			break
		}
		time.Sleep(time.Duration(attempt+1) * 15 * time.Millisecond)
	}
	if err != nil {
		return err
	}
	ok = true
	return nil
}

// SafeName sunucu adı gibi kullanıcı metnini dosya/klasör adına çevirir: Windows'ta yasak
// karakterler (<>:"/\|?* ve kontrol karakterleri) '_' olur, sondaki nokta/boşluk atılır, ayrılmış
// adlara (CON, NUL, COM1…) '_' eklenir. Geçerli adlar AYNEN kalır (mevcut dosyalar etkilenmez).
// Neden: "Prod: public_html" sunucusunun state'i NTFS'te "Prod" dosyasının gizli veri akışına
// yazılıyor, release klasörü hiç oluşturulamıyor ve sync sonsuz döngüye giriyordu.
func SafeName(s string) string {
	out := strings.Map(func(r rune) rune {
		if r < 32 || strings.ContainsRune(`<>:"/\|?*`, r) {
			return '_'
		}
		return r
	}, s)
	out = strings.TrimRight(out, ". ")
	if out == "" {
		return "_"
	}
	base := strings.ToUpper(out)
	if i := strings.IndexByte(base, '.'); i >= 0 {
		base = base[:i]
	}
	switch base {
	case "CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9":
		out += "_"
	}
	return out
}

// ServerFile dir altındaki sunucuya ait dosyanın yolu (<SafeName>+ext). Ad güvenli hâline
// çevrildiyse ve yeni dosya henüz yoksa eski (ham adlı) dosyanın içeriği yeni yola kopyalanır —
// state / freeze / failed / milestone kayıtları kaybolmasın.
func ServerFile(dir, server, ext string) string {
	safe := SafeName(server)
	p := filepath.Join(dir, safe+ext)
	if safe == server {
		return p
	}
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		return p
	}
	if data, err := os.ReadFile(filepath.Join(dir, server+ext)); err == nil {
		if os.MkdirAll(dir, 0755) == nil {
			_ = WriteFileAtomic(p, data, 0644)
		}
	}
	return p
}

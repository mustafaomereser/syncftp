// Package fsutil küçük dosya sistemi yardımcıları.
package fsutil

import (
	"os"
	"path/filepath"
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

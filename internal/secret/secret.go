// Package secret şifreleri işletim sisteminin anahtarlığında saklar
// (Windows Credential Manager, macOS Keychain, Linux Secret Service).
package secret

import (
	"fmt"
	"strings"

	"github.com/zalando/go-keyring"
)

// Marker — JSON'da şifre yerine yazılan işaret; gerçek değer anahtarlıktadır.
const Marker = "keyring:"

const service = "syncftp"

// IsMarker değer anahtarlık işareti mi.
func IsMarker(v string) bool { return v == Marker }

// Account bir kimlik bilgisinin anahtarlıktaki adı: "<kind>://user@host:port".
// Sunucu adından bağımsızdır → yeniden adlandırma/proje taşıma kaydı kırmaz,
// aynı hesabı kullanan projeler aynı kaydı paylaşır.
func Account(kind, user, host string, port int) string {
	return strings.ToLower(fmt.Sprintf("%s://%s@%s:%d", kind, user, host, port))
}

// Get anahtarlıktan şifreyi okur.
func Get(account string) (string, error) {
	v, err := keyring.Get(service, account)
	if err != nil {
		return "", fmt.Errorf("anahtarlıkta bulunamadı (%s): %w", account, err)
	}
	return v, nil
}

// Set şifreyi anahtarlığa yazar (varsa günceller).
func Set(account, password string) error {
	return keyring.Set(service, account, password)
}

// Delete kaydı siler (yoksa hata vermez).
func Delete(account string) {
	_ = keyring.Delete(service, account)
}

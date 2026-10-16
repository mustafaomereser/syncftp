package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ══════════════════════════════════════════════════════════════════════════════
// Library — proje bağımsız, makine geneli ayarlar: <UserConfigDir>/syncftp/library.json
//
//   - Connections: kayıtlı bağlantılar (her projede "kütüphaneden" seçilebilir; push
//     --connection ile doğrudan kullanılabilir). Şifreler mümkünse OS anahtarlığına yazılır —
//     hesap adı projelerle AYNI (protokol://user@host:port), böylece projeye aktarılan
//     "keyring:" işaretli bağlantı ek bir şey yapmadan çözülür.
//   - Defaults: yeni proje oluşturulurken (init / GUI / AI) uygulanan global sync ayarları.
// ══════════════════════════════════════════════════════════════════════════════

type Library struct {
	Connections  []Connection `json:"connections"`
	Defaults     Defaults     `json:"defaults"`
	SecretErrors []string     `json:"-"`
}

// Defaults yeni projeye kopyalanan varsayılanlar.
type Defaults struct {
	Sync           Sync   `json:"sync"`
	FirstSyncFull  bool   `json:"first_sync_full,omitempty"`
	Keyring        bool   `json:"keyring,omitempty"`
	IgnoreTemplate string `json:"ignore_template,omitempty"` // laravel | wordpress | node | generic | ""
	MaxConnections int    `json:"max_connections,omitempty"` // yeni sunucu için (0 → 1)
	MaxRetries     int    `json:"max_retries,omitempty"`     // yeni sunucu için (0 → 2)
}

// BuiltinDefaults kütüphane dosyası hiç yokken kullanılan varsayılanlar — eski `init`
// sihirbazının yazdığı değerlerle aynı (.env korumalı, ilk sync tam, 1 bağlantı / 2 deneme).
func BuiltinDefaults() Defaults {
	return Defaults{Sync: Sync{Protect: []string{".env"}, Include: []string{}, Exclude: []string{}},
		FirstSyncFull: true, MaxConnections: 1, MaxRetries: 2}
}

// libraryPathOverride testlerde gerçek kullanıcı dosyasına yazılmasın diye.
var libraryPathOverride string

// LibraryPath kütüphane dosyasının yolu.
func LibraryPath() string {
	if libraryPathOverride != "" {
		return libraryPathOverride
	}
	base, err := os.UserConfigDir()
	if err != nil {
		base = os.TempDir()
	}
	return filepath.Join(base, "syncftp", "library.json")
}

// LoadLibrary kütüphaneyi okur; dosya yoksa boş kütüphane döner (hata değil).
// "keyring:" işaretli şifreler anahtarlıktan çözülür, çözülemeyenler SecretErrors'a yazılır.
func LoadLibrary() (*Library, error) {
	lib := &Library{Connections: []Connection{}, Defaults: BuiltinDefaults()}
	data, err := os.ReadFile(LibraryPath())
	if os.IsNotExist(err) {
		return lib, nil
	}
	if err != nil {
		return lib, err
	}
	if err := json.Unmarshal(data, lib); err != nil {
		return lib, fmt.Errorf("library.json: %w", err)
	}
	if lib.Connections == nil {
		lib.Connections = []Connection{}
	}
	tmp := &Config{}
	for i := range lib.Connections {
		c := &lib.Connections[i]
		if c.Port == 0 {
			c.Port = defaultPort(c.Protocol)
		}
		if c.TunnelHost != "" && c.TunnelPort == 0 {
			c.TunnelPort = 22
		}
		resolveSecrets(&c.Password, &c.SSHTunnel, c.Protocol, c.User, c.Host, c.Port, tmp)
	}
	lib.SecretErrors = tmp.SecretErrors
	return lib, nil
}

// SaveLibrary kütüphaneyi yazar (0600). Şifreler anahtarlığa yazılabiliyorsa JSON'a işaret düşer;
// anahtarlık yoksa düz metin kalır (proje config'iyle aynı davranış). Canlı yapı değişmez.
func SaveLibrary(lib *Library) error {
	out := *lib
	out.Connections = make([]Connection, len(lib.Connections))
	copy(out.Connections, lib.Connections)
	for i := range out.Connections {
		c := &out.Connections[i]
		storeSecrets(&c.Password, &c.SSHTunnel, c.Protocol, c.User, c.Host, c.Port)
	}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	p := LibraryPath()
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		return err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

// FindConnection ada göre (büyük/küçük harf duyarsız) kayıtlı bağlantıyı döndürür.
func (l *Library) FindConnection(name string) (Connection, bool) {
	for _, c := range l.Connections {
		if strings.EqualFold(c.Name, name) {
			return c, true
		}
	}
	return Connection{}, false
}

// ValidateConnections boş/yinelenen adları reddeder.
func (l *Library) ValidateConnections() error {
	seen := map[string]bool{}
	for _, c := range l.Connections {
		n := strings.ToLower(strings.TrimSpace(c.Name))
		if n == "" {
			return fmt.Errorf("connection name required")
		}
		if seen[n] {
			return fmt.Errorf("duplicate connection name: %s", c.Name)
		}
		if strings.TrimSpace(c.Host) == "" {
			return fmt.Errorf("%s: host required", c.Name)
		}
		seen[n] = true
	}
	return nil
}

// ApplyDefaults varsayılanları YENİ bir config'e uygular (mevcut projeyi ezmek için kullanılmaz).
func (d Defaults) ApplyDefaults(cfg *Config) {
	cfg.Sync = d.Sync.Clone()
	cfg.FirstSync.Full = d.FirstSyncFull
	cfg.Project.Keyring = d.Keyring
}

// NewServer varsayılan bağlantı/yeniden deneme sayılarıyla boş bir sunucu döndürür.
func (d Defaults) NewServer(name string) Server {
	mc, mr := d.MaxConnections, d.MaxRetries
	if mc <= 0 {
		mc = 1
	}
	if mr <= 0 {
		mr = 2
	}
	return Server{Name: name, Enabled: true, Passive: true, MaxConnections: mc, MaxRetries: mr,
		RemotePath: "/", Include: []string{}, Exclude: []string{}, Protect: []string{}}
}

// ApplyConnection bağlantının kimlik bilgilerini sunucuya kopyalar (profil referansı OLMADAN —
// proje kütüphaneye bağımlı kalmasın; şifre anahtarlıktaysa işaret aynı hesabı gösterir).
func (s *Server) ApplyConnection(c Connection) {
	s.Host, s.Port, s.User, s.Password = c.Host, c.Port, c.User, c.Password
	s.Passive, s.DisableEPSV, s.NATWorkaround = c.Passive, c.DisableEPSV, c.NATWorkaround
	s.Protocol, s.KeyFile, s.TLSSkipVerify = c.Protocol, c.KeyFile, c.TLSSkipVerify
	s.SSHTunnel = c.SSHTunnel
	if s.Port == 0 {
		s.Port = defaultPort(s.Protocol)
	}
}

// Clone listeleri paylaşmayan derin kopya.
func (s Sync) Clone() Sync {
	cp := func(v []string) []string {
		if v == nil {
			return nil
		}
		out := make([]string, len(v))
		copy(out, v)
		return out
	}
	s.Protect, s.Include, s.Exclude = cp(s.Protect), cp(s.Include), cp(s.Exclude)
	s.IgnoreFiles, s.BlockExtensions = cp(s.IgnoreFiles), cp(s.BlockExtensions)
	s.PreSync, s.PostSync = cp(s.PreSync), cp(s.PostSync)
	return s
}

// DefaultPort protokolün varsayılan portu (sftp → 22, diğerleri → 21).
func DefaultPort(protocol string) int { return defaultPort(protocol) }

// Normalize boş portları varsayılanla doldurur (sunucu ve SSH tüneli).
func (s *Server) Normalize() {
	if s.Port == 0 {
		s.Port = defaultPort(s.Protocol)
	}
	if s.TunnelHost != "" && s.TunnelPort == 0 {
		s.TunnelPort = 22
	}
}

// Resolve sunucuyu bağlantı profiliyle birlikte bağlanmaya hazır hale getirir: connection adı
// varsa (büyük/küçük harf duyarlı, config.Load ile aynı) kimlik bilgileri profilden gelir, portlar
// normalize edilir. config.Load ve paneldeki "Bağlantıyı test et" bunu kullanır.
func (s Server) Resolve(conns []Connection) Server {
	if s.Connection != "" {
		for _, c := range conns {
			if c.Name == s.Connection {
				s.ApplyConnection(c)
				break
			}
		}
	}
	s.Normalize()
	return s
}

// ConnectionFromServer sunucunun kimlik bilgilerinden bağlantı profili (ApplyConnection'ın tersi).
func ConnectionFromServer(name string, s Server) Connection {
	return Connection{Name: name, Host: s.Host, Port: s.Port, User: s.User, Password: s.Password,
		Passive: s.Passive, DisableEPSV: s.DisableEPSV, NATWorkaround: s.NATWorkaround, Protocol: s.Protocol,
		KeyFile: s.KeyFile, TLSSkipVerify: s.TLSSkipVerify, SSHTunnel: s.SSHTunnel}
}

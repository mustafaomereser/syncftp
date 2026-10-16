package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"syncftp/internal/secret"
)

// Connection holds reusable FTP credentials shared across multiple servers.
type Connection struct {
	Name          string `json:"name"`
	Host          string `json:"host"`
	Port          int    `json:"port"`
	User          string `json:"user"`
	Password      string `json:"password"`
	Passive       bool   `json:"passive"`
	DisableEPSV   bool   `json:"disable_epsv,omitempty"`
	NATWorkaround bool   `json:"nat_workaround,omitempty"`
	Protocol      string `json:"protocol,omitempty"`        // "ftp" (varsayılan) | "ftps" | "sftp"
	KeyFile       string `json:"key_file,omitempty"`        // SFTP özel anahtar dosyası (şifreli ise password passphrase olarak denenir)
	TLSSkipVerify bool   `json:"tls_skip_verify,omitempty"` // FTPS: sertifika doğrulamasını atla (paylaşımlı hosting'de hostname uyuşmazlığı)
	SSHTunnel
}

// Deploy — gelen webhook (GitHub/GitLab/Gitea/Bitbucket push ya da genel istek) ile
// tetiklenen deploy. URL: <public>/hook/<proje>/<name>. İsteğe bağlı önce `git pull`,
// sonra seçili sunuculara normal sync (yedek/rollback dahil).
type Deploy struct {
	Name     string   `json:"name"`               // URL'deki ad (benzersiz)
	Provider string   `json:"provider"`           // github | gitlab | gitea | bitbucket | generic
	Secret   string   `json:"secret"`             // imza/token doğrulama anahtarı
	Branch   string   `json:"branch,omitempty"`   // boş = her branch
	Servers  []string `json:"servers,omitempty"`  // boş = tüm aktif sunucular
	GitPull  bool     `json:"git_pull,omitempty"` // sync öncesi `git pull --ff-only`
	DryRun   bool     `json:"dry_run,omitempty"`  // yalnızca neyin yükleneceğini göster
	Enabled  bool     `json:"enabled"`
}

// DeployProviders geçerli sağlayıcılar.
var DeployProviders = []string{"github", "gitlab", "gitea", "bitbucket", "generic"}

// SSHTunnel — bağlantının üzerinden geçeceği SSH atlama sunucusu (jump host).
// TunnelHost boşsa tünel kullanılmaz. FTP'de kontrol + veri bağlantıları,
// SFTP'de SSH bağlantısının kendisi bu sunucu üzerinden açılır.
type SSHTunnel struct {
	TunnelHost     string `json:"tunnel_host,omitempty"`
	TunnelPort     int    `json:"tunnel_port,omitempty"`
	TunnelUser     string `json:"tunnel_user,omitempty"`
	TunnelPassword string `json:"tunnel_password,omitempty"`
	TunnelKeyFile  string `json:"tunnel_key_file,omitempty"`
}

// HasTunnel tünel tanımlı mı.
func (t SSHTunnel) HasTunnel() bool { return t.TunnelHost != "" }

type Config struct {
	Project     Project      `json:"project"`
	Sync        Sync         `json:"sync"`
	FirstSync   FirstSync    `json:"first_sync"`
	Connections []Connection `json:"connections,omitempty"`
	Servers     []Server     `json:"servers"`
	Deploy      []Deploy     `json:"deploy,omitempty"` // CI/CD: gelen webhook ile tetiklenen deploy'lar

	// SecretErrors anahtarlıktan okunamayan kayıtlar (Load doldurur, JSON'a yazılmaz).
	SecretErrors []string `json:"-"`
}

type Project struct {
	Name        string `json:"name"`
	Keyring     bool   `json:"keyring,omitempty"`      // şifreler sistem anahtarlığında (JSON'da "keyring:" işareti kalır)
	DefaultPath string `json:"default_path,omitempty"` // global varsayılan yerel dizin
	LocalPath   string `json:"local_path,omitempty"`   // deprecated: geriye dönük uyumluluk
}

// EffectiveLocalPath global varsayılan yerel dizini döndürür.
func (p Project) EffectiveLocalPath() string {
	if p.DefaultPath != "" {
		return p.DefaultPath
	}
	if p.LocalPath != "" {
		return p.LocalPath
	}
	return "."
}

type Sync struct {
	Protect         []string `json:"protect"`
	Include         []string `json:"include"`
	Exclude         []string `json:"exclude"`
	IgnoreFiles     []string `json:"ignore_files"`
	BandwidthLimit  int      `json:"bandwidth_limit,omitempty"`  // KB/s, sunucu başına upload hız sınırı (0 = sınırsız); sunucu değeri önceliklidir
	NoBackup        bool     `json:"no_backup,omitempty"`        // true → upload öncesi uzak yedek alınmaz (rollback devre dışı)
	BackupKeep      int      `json:"backup_keep,omitempty"`      // yedek dosyaları tutulacak son release sayısı (0 → 10)
	DirectUpload    bool     `json:"direct_upload,omitempty"`    // true → atomik (tmp+rename) upload kapalı, doğrudan hedefe yazılır              // nil/boş → [".gitignore","syncftp.ignore"] ikisi de yüklenir
	Webhook         string   `json:"webhook,omitempty"`          // sync sonrası POST atılacak URL
	BlockExtensions []string `json:"block_extensions,omitempty"` // yüklenmeyecek uzantılar: [".jpg",".pdf"]
	PreSync         []string `json:"pre_sync,omitempty"`         // upload öncesi çalıştırılacak komutlar; hata → o sunucunun sync'i iptal
	PostSync        []string `json:"post_sync,omitempty"`        // upload sonrası çalıştırılacak komutlar; hata sync'i etkilemez
}

type FirstSync struct {
	Full bool `json:"full"`
}

type Server struct {
	Name            string   `json:"name"`
	Connection      string   `json:"connection,omitempty"` // bağlantı profili adı; set edilince host/port/user/password/passive buradan gelir
	Host            string   `json:"host"`
	Port            int      `json:"port"`
	User            string   `json:"user"`
	Password        string   `json:"password"`
	RemotePath      string   `json:"remote_path"`
	LocalPath       string   `json:"local_path,omitempty"` // per-server yerel dizin; boşsa project.DefaultPath
	Passive         bool     `json:"passive"`
	DisableEPSV     bool     `json:"disable_epsv"`
	NATWorkaround   bool     `json:"nat_workaround"`
	Enabled         bool     `json:"enabled"`
	MaxConnections  int      `json:"max_connections"`
	MaxRetries      int      `json:"max_retries"`
	Include         []string `json:"include"`
	Exclude         []string `json:"exclude"`
	Protect         []string `json:"protect"`
	Webhook         string   `json:"webhook,omitempty"`          // sync sonrası POST atılacak URL (global webhook'u override eder)
	Minify          bool     `json:"minify,omitempty"`           // CSS/JS dosyalarını upload öncesi minify et
	Obfuscate       bool     `json:"obfuscate,omitempty"`        // JS dosyalarını minify üstüne obfuscate et (terser gerekir)
	BlockExtensions []string `json:"block_extensions,omitempty"` // yüklenmeyecek uzantılar (global listeye ek)
	PreSync         []string `json:"pre_sync,omitempty"`         // sunucuya özel pre-sync komutları (global listeye ek)
	PostSync        []string `json:"post_sync,omitempty"`        // sunucuya özel post-sync komutları (global listeye ek)
	Protocol        string   `json:"protocol,omitempty"`         // "ftp" (varsayılan) | "ftps" | "sftp"
	KeyFile         string   `json:"key_file,omitempty"`         // SFTP özel anahtar dosyası
	TLSSkipVerify   bool     `json:"tls_skip_verify,omitempty"`  // FTPS: sertifika doğrulamasını atla
	BandwidthLimit  int      `json:"bandwidth_limit,omitempty"`  // KB/s upload hız sınırı (0 = global değer)
	RemotePreSync   []string `json:"remote_pre_sync,omitempty"`  // SFTP: upload öncesi SUNUCUDA (SSH) çalışacak komutlar; hata → sync iptal
	RemotePostSync  []string `json:"remote_post_sync,omitempty"` // SFTP: upload sonrası SUNUCUDA çalışacak komutlar (migrate, cache clear, restart…)
	NoBackup        bool     `json:"no_backup,omitempty"`        // true → bu sunucuda upload öncesi yedek alınmaz
	DirectUpload    bool     `json:"direct_upload,omitempty"`    // true → atomik upload kapalı (rename desteklemeyen sunucular)
	SSHTunnel
}

// IsSFTP sunucu SFTP protokolü kullanıyorsa true.
func (s Server) IsSFTP() bool { return strings.EqualFold(s.Protocol, "sftp") }

// IsFTPS sunucu FTP over TLS (explicit) kullanıyorsa true.
func (s Server) IsFTPS() bool { return strings.EqualFold(s.Protocol, "ftps") }

// ProtocolName görüntüleme için protokol adı ("ftp" / "ftps" / "sftp").
func (s Server) ProtocolName() string {
	switch {
	case s.IsSFTP():
		return "sftp"
	case s.IsFTPS():
		return "ftps"
	}
	return "ftp"
}

// BackupEnabled upload öncesi uzak yedek alınacak mı (global ve sunucu NoBackup kapatabilir).
func (s Server) BackupEnabled(sy Sync) bool { return !sy.NoBackup && !s.NoBackup }

// EffectiveBackupKeep yedek dosyası tutulacak release sayısı.
func (sy Sync) EffectiveBackupKeep() int {
	if sy.BackupKeep > 0 {
		return sy.BackupKeep
	}
	return 10
}

// Protocols — geçerli protokol değerleri (TUI/GUI seçicileri için).
var Protocols = []string{"ftp", "ftps", "sftp"}

// EffectiveBandwidth KB/s cinsinden geçerli upload sınırı (sunucu > global, 0 = sınırsız).
func (s Server) EffectiveBandwidth(sy Sync) int {
	if s.BandwidthLimit > 0 {
		return s.BandwidthLimit
	}
	if sy.BandwidthLimit > 0 {
		return sy.BandwidthLimit
	}
	return 0
}

// defaultPort protokole göre varsayılan port.
func defaultPort(protocol string) int {
	if strings.EqualFold(protocol, "sftp") {
		return 22
	}
	return 21
}

// EffectiveLocalPath sunucunun kendi LocalPath'i varsa onu, yoksa proje defaultunu döndürür.
func (s Server) EffectiveLocalPath(proj Project) string {
	if s.LocalPath != "" {
		return s.LocalPath
	}
	return proj.EffectiveLocalPath()
}

// Load reads syncftp.json and returns the parsed config.
func Load(dir string) (*Config, error) {
	data, err := os.ReadFile(filepath.Join(dir, ".syncftp", "syncftp.json"))
	if err != nil {
		return nil, fmt.Errorf("syncftp.json okunamadı: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("syncftp.json parse hatası: %w", err)
	}

	// Bağlantı profili port defaults
	connMap := make(map[string]*Connection, len(cfg.Connections))
	for i := range cfg.Connections {
		if cfg.Connections[i].Port == 0 {
			cfg.Connections[i].Port = defaultPort(cfg.Connections[i].Protocol)
		}
		if cfg.Connections[i].TunnelHost != "" && cfg.Connections[i].TunnelPort == 0 {
			cfg.Connections[i].TunnelPort = 22
		}
		resolveSecrets(&cfg.Connections[i].Password, &cfg.Connections[i].SSHTunnel,
			cfg.Connections[i].Protocol, cfg.Connections[i].User, cfg.Connections[i].Host, cfg.Connections[i].Port, &cfg)
		connMap[cfg.Connections[i].Name] = &cfg.Connections[i]
	}
	for i := range cfg.Servers {
		if cfg.Servers[i].Connection == "" {
			port := cfg.Servers[i].Port
			if port == 0 {
				port = defaultPort(cfg.Servers[i].Protocol)
			}
			if cfg.Servers[i].TunnelHost != "" && cfg.Servers[i].TunnelPort == 0 {
				cfg.Servers[i].TunnelPort = 22
			}
			resolveSecrets(&cfg.Servers[i].Password, &cfg.Servers[i].SSHTunnel,
				cfg.Servers[i].Protocol, cfg.Servers[i].User, cfg.Servers[i].Host, port, &cfg)
		}
	}

	// Bağlantı referanslarını çöz: connection set edilmişse credential'lar oradan gelir
	for i := range cfg.Servers {
		if cfg.Servers[i].Connection != "" {
			if conn, ok := connMap[cfg.Servers[i].Connection]; ok {
				cfg.Servers[i].Host = conn.Host
				cfg.Servers[i].Port = conn.Port
				cfg.Servers[i].User = conn.User
				cfg.Servers[i].Password = conn.Password
				cfg.Servers[i].Passive = conn.Passive
				cfg.Servers[i].DisableEPSV = conn.DisableEPSV
				cfg.Servers[i].NATWorkaround = conn.NATWorkaround
				cfg.Servers[i].Protocol = conn.Protocol
				cfg.Servers[i].KeyFile = conn.KeyFile
				cfg.Servers[i].TLSSkipVerify = conn.TLSSkipVerify
				cfg.Servers[i].SSHTunnel = conn.SSHTunnel
			}
		}
		if cfg.Servers[i].Port == 0 {
			cfg.Servers[i].Port = defaultPort(cfg.Servers[i].Protocol)
		}
		if cfg.Servers[i].TunnelHost != "" && cfg.Servers[i].TunnelPort == 0 {
			cfg.Servers[i].TunnelPort = 22
		}
	}

	// Eski project.local_path → default_path göçü
	if cfg.Project.DefaultPath == "" && cfg.Project.LocalPath != "" {
		cfg.Project.DefaultPath = cfg.Project.LocalPath
	}
	cfg.Project.LocalPath = ""

	return &cfg, nil
}

// Save writes cfg back to syncftp.json in dir.
// Connection-referenced servers have their credential fields zeroed before
// serialization; they are re-resolved from the profile on next Load.
func Save(dir string, cfg *Config) error {
	// Deep-copy servers so we don't mutate the live config
	servers := make([]Server, len(cfg.Servers))
	copy(servers, cfg.Servers)
	for i := range servers {
		if servers[i].Connection != "" {
			servers[i].Host = ""
			servers[i].Port = 0
			servers[i].User = ""
			servers[i].Password = ""
			servers[i].Passive = false
			servers[i].DisableEPSV = false
			servers[i].NATWorkaround = false
			servers[i].Protocol = ""
			servers[i].KeyFile = ""
			servers[i].TLSSkipVerify = false
			servers[i].SSHTunnel = SSHTunnel{}
		}
	}
	conns := make([]Connection, len(cfg.Connections))
	copy(conns, cfg.Connections)

	// Anahtarlık modu: şifreler OS anahtarlığına yazılır, JSON'a yalnızca işaret düşer.
	// Anahtarlığa yazılamazsa (Linux'ta Secret Service yok vb.) düz metin kalır.
	if cfg.Project.Keyring {
		for i := range conns {
			storeSecrets(&conns[i].Password, &conns[i].SSHTunnel, conns[i].Protocol, conns[i].User, conns[i].Host, conns[i].Port)
		}
		for i := range servers {
			if servers[i].Connection == "" {
				storeSecrets(&servers[i].Password, &servers[i].SSHTunnel, servers[i].Protocol, servers[i].User, servers[i].Host, servers[i].Port)
			}
		}
	}

	out := *cfg
	out.Servers = servers
	out.Connections = conns
	data, err := json.MarshalIndent(&out, "", "  ")
	if err != nil {
		return err
	}
	p := filepath.Join(dir, ".syncftp", "syncftp.json")
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		return err
	}
	return os.WriteFile(p, data, 0600)
}

func (c *Config) EnabledServers() []Server {
	var out []Server
	for _, s := range c.Servers {
		if s.Enabled {
			out = append(out, s)
		}
	}
	return out
}

// protoKind anahtarlık hesap adı için protokol ("ftp"/"ftps"/"sftp").
func protoKind(p string) string {
	switch strings.ToLower(p) {
	case "sftp", "ftps":
		return strings.ToLower(p)
	}
	return "ftp"
}

func tunnelAccount(t *SSHTunnel, user string) string {
	tu := t.TunnelUser
	if tu == "" {
		tu = user
	}
	port := t.TunnelPort
	if port == 0 {
		port = 22
	}
	return secret.Account("ssh", tu, t.TunnelHost, port)
}

// resolveSecrets "keyring:" işaretli şifreleri anahtarlıktan okur. Okunamazsa işaret
// yerinde kalır (Save işareti korur) ve hata cfg.SecretErrors'a eklenir.
func resolveSecrets(password *string, t *SSHTunnel, protocol, user, host string, port int, cfg *Config) {
	if secret.IsMarker(*password) {
		if v, err := secret.Get(secret.Account(protoKind(protocol), user, host, port)); err == nil {
			*password = v
		} else {
			cfg.SecretErrors = append(cfg.SecretErrors, err.Error())
		}
	}
	if t.TunnelHost != "" && secret.IsMarker(t.TunnelPassword) {
		if v, err := secret.Get(tunnelAccount(t, user)); err == nil {
			t.TunnelPassword = v
		} else {
			cfg.SecretErrors = append(cfg.SecretErrors, err.Error())
		}
	}
}

// storeSecrets şifreleri anahtarlığa yazar, başarılıysa değeri işaretle değiştirir.
func storeSecrets(password *string, t *SSHTunnel, protocol, user, host string, port int) {
	if *password != "" && !secret.IsMarker(*password) && host != "" {
		if secret.Set(secret.Account(protoKind(protocol), user, host, port), *password) == nil {
			*password = secret.Marker
		}
	}
	if t.TunnelHost != "" && t.TunnelPassword != "" && !secret.IsMarker(t.TunnelPassword) {
		if secret.Set(tunnelAccount(t, user), t.TunnelPassword) == nil {
			t.TunnelPassword = secret.Marker
		}
	}
}

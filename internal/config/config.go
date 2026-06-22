package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
}

type Config struct {
	Project     Project      `json:"project"`
	Sync        Sync         `json:"sync"`
	FirstSync   FirstSync    `json:"first_sync"`
	Connections []Connection `json:"connections,omitempty"`
	Servers     []Server     `json:"servers"`
}

type Project struct {
	Name        string `json:"name"`
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
	IgnoreFiles     []string `json:"ignore_files"`              // nil/boş → [".gitignore","syncftp.ignore"] ikisi de yüklenir
	Webhook         string   `json:"webhook,omitempty"`         // sync sonrası POST atılacak URL
	BlockExtensions []string `json:"block_extensions,omitempty"` // yüklenmeyecek uzantılar: [".jpg",".pdf"]
	BlockFiles      []string `json:"block_files,omitempty"`      // yüklenmeyecek dosya adları: ["thumbs.db"]
}

type FirstSync struct {
	Full bool `json:"full"`
}

type Server struct {
	Name           string   `json:"name"`
	Connection     string   `json:"connection,omitempty"` // bağlantı profili adı; set edilince host/port/user/password/passive buradan gelir
	Host           string   `json:"host"`
	Port           int      `json:"port"`
	User           string   `json:"user"`
	Password       string   `json:"password"`
	RemotePath     string   `json:"remote_path"`
	LocalPath      string   `json:"local_path,omitempty"` // per-server yerel dizin; boşsa project.DefaultPath
	Passive        bool     `json:"passive"`
	DisableEPSV    bool     `json:"disable_epsv"`
	NATWorkaround  bool     `json:"nat_workaround"`
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
	BlockFiles      []string `json:"block_files,omitempty"`      // yüklenmeyecek dosya adları (global listeye ek)
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
			cfg.Connections[i].Port = 21
		}
		connMap[cfg.Connections[i].Name] = &cfg.Connections[i]
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
			}
		}
		if cfg.Servers[i].Port == 0 {
			cfg.Servers[i].Port = 21
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
		}
	}
	out := *cfg
	out.Servers = servers
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

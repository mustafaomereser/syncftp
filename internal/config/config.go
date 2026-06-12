package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Project   Project   `json:"project"`
	Sync      Sync      `json:"sync"`
	FirstSync FirstSync `json:"first_sync"`
	Servers   []Server  `json:"servers"`
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
	Protect     []string `json:"protect"`
	Include     []string `json:"include"`
	Exclude     []string `json:"exclude"`
	IgnoreFiles []string `json:"ignore_files"` // nil/boş → [".gitignore","syncftp.ignore"] ikisi de yüklenir
}

type FirstSync struct {
	Full bool `json:"full"`
}

type Server struct {
	Name           string   `json:"name"`
	Host           string   `json:"host"`
	Port           int      `json:"port"`
	User           string   `json:"user"`
	Password       string   `json:"password"`
	RemotePath     string   `json:"remote_path"`
	LocalPath      string   `json:"local_path,omitempty"` // per-server yerel dizin; boşsa project.DefaultPath
	Passive        bool     `json:"passive"`
	DisableEPSV    bool     `json:"disable_epsv"`
	NATWorkaround  bool     `json:"nat_workaround"`
	Enabled        bool     `json:"enabled"`
	MaxConnections int      `json:"max_connections"`
	MaxRetries     int      `json:"max_retries"`
	Include        []string `json:"include"`
	Exclude        []string `json:"exclude"`
	Protect        []string `json:"protect"`
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
	data, err := os.ReadFile(filepath.Join(dir, "syncftp.json"))
	if err != nil {
		return nil, fmt.Errorf("syncftp.json okunamadı: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("syncftp.json parse hatası: %w", err)
	}

	for i := range cfg.Servers {
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
func Save(dir string, cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "syncftp.json"), data, 0600)
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

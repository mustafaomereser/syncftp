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
	Name      string `json:"name"`
	LocalPath string `json:"local_path"`
}

type Sync struct {
	Protect []string `json:"protect"`
	Include []string `json:"include"`
	Exclude []string `json:"exclude"`
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
	Passive        bool     `json:"passive"`
	DisableEPSV    bool     `json:"disable_epsv"`    // EPSV'yi kapat, sadece PASV kullan
	NATWorkaround  bool     `json:"nat_workaround"`  // PASV yanıtındaki IP'yi yoksay, sunucu IP'sini kullan
	Enabled        bool     `json:"enabled"`
	MaxConnections int      `json:"max_connections"` // default 1
	MaxRetries     int      `json:"max_retries"`     // default 2
	Include        []string `json:"include"`
	Exclude        []string `json:"exclude"`
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

	if cfg.Project.LocalPath == "" {
		cfg.Project.LocalPath = "."
	}

	return &cfg, nil
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

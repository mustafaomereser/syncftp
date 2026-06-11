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
	Name       string   `json:"name"`
	Host       string   `json:"host"`
	Port       int      `json:"port"`
	User       string   `json:"user"`
	Password   string   `json:"password,omitempty"`
	RemotePath string   `json:"remote_path"`
	Passive    bool     `json:"passive"`
	Enabled    bool     `json:"enabled"`
	Include    []string `json:"include"`
	Exclude    []string `json:"exclude"`
}

type secretConfig struct {
	Servers []struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	} `json:"servers"`
}

// Load reads syncftp.json and merges passwords from syncftp.secrets.json if present.
func Load(dir string) (*Config, error) {
	mainPath := filepath.Join(dir, "syncftp.json")
	secretPath := filepath.Join(dir, "syncftp.secrets.json")

	data, err := os.ReadFile(mainPath)
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

	if data, err := os.ReadFile(secretPath); err == nil {
		var sec secretConfig
		if err := json.Unmarshal(data, &sec); err != nil {
			return nil, fmt.Errorf("syncftp.secrets.json parse hatası: %w", err)
		}
		pwMap := make(map[string]string, len(sec.Servers))
		for _, s := range sec.Servers {
			pwMap[s.Name] = s.Password
		}
		for i, srv := range cfg.Servers {
			if pw, ok := pwMap[srv.Name]; ok {
				cfg.Servers[i].Password = pw
			}
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

package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Project   Project   `toml:"project"`
	Sync      Sync      `toml:"sync"`
	FirstSync FirstSync `toml:"first_sync"`
	Servers   []Server  `toml:"servers"`
}

type Project struct {
	Name      string `toml:"name"`
	LocalPath string `toml:"local_path"`
}

type Sync struct {
	Protect []string `toml:"protect"`
	Include []string `toml:"include"`
}

type FirstSync struct {
	Full bool `toml:"full"`
}

type Server struct {
	Name       string `toml:"name"`
	Host       string `toml:"host"`
	Port       int    `toml:"port"`
	User       string `toml:"user"`
	Password   string `toml:"password"`
	RemotePath string `toml:"remote_path"`
	Passive    bool   `toml:"passive"`
	Enabled    bool   `toml:"enabled"`
}

type secretConfig struct {
	Servers []struct {
		Name     string `toml:"name"`
		Password string `toml:"password"`
	} `toml:"servers"`
}

// Load reads syncftp.toml and merges passwords from syncftp.config if present.
func Load(dir string) (*Config, error) {
	mainPath := filepath.Join(dir, "syncftp.toml")
	secretPath := filepath.Join(dir, "syncftp.config")

	var cfg Config
	if _, err := toml.DecodeFile(mainPath, &cfg); err != nil {
		return nil, fmt.Errorf("syncftp.toml okunamadı: %w", err)
	}

	for i := range cfg.Servers {
		if cfg.Servers[i].Port == 0 {
			cfg.Servers[i].Port = 21
		}
	}

	if _, err := os.Stat(secretPath); err == nil {
		var sec secretConfig
		if _, err := toml.DecodeFile(secretPath, &sec); err != nil {
			return nil, fmt.Errorf("syncftp.config okunamadı: %w", err)
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

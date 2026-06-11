package failed

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type List struct {
	Server    string    `json:"server"`
	CreatedAt time.Time `json:"created_at"`
	Files     []string  `json:"files"`
}

func dir(configDir string) string {
	return filepath.Join(configDir, ".syncftp", "failed")
}

func path(configDir, serverName string) string {
	return filepath.Join(dir(configDir), serverName+".json")
}

// Save writes the failed file list for a server. Overwrites any previous list.
func Save(configDir, serverName string, files []string) error {
	if err := os.MkdirAll(dir(configDir), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(List{
		Server:    serverName,
		CreatedAt: time.Now(),
		Files:     files,
	}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path(configDir, serverName), data, 0644)
}

// Load reads the failed file list for a server.
// Returns nil without error if no failed list exists.
func Load(configDir, serverName string) (*List, error) {
	data, err := os.ReadFile(path(configDir, serverName))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var l List
	if err := json.Unmarshal(data, &l); err != nil {
		return nil, err
	}
	return &l, nil
}

// Clear deletes the failed list for a server.
func Clear(configDir, serverName string) {
	_ = os.Remove(path(configDir, serverName))
}

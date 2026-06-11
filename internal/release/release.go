package release

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Manifest struct {
	Version   int               `json:"version"`
	CreatedAt time.Time         `json:"created_at"`
	Server    string            `json:"server"`
	FileCount int               `json:"file_count"`
	Files     map[string]string `json:"files"` // relPath -> "sha256:<hex>"
}

// Create writes a release manifest under <configDir>/.syncftp/releases/<server>/<timestamp>/.
// Returns the path to the release directory.
func Create(configDir, serverName string, files map[string]string) (string, error) {
	timestamp := time.Now().Format("20060102-150405")
	releaseDir := filepath.Join(configDir, ".syncftp", "releases", serverName, timestamp)

	if err := os.MkdirAll(releaseDir, 0755); err != nil {
		return "", fmt.Errorf("release dizini oluşturulamadı: %w", err)
	}

	manifest := Manifest{
		Version:   1,
		CreatedAt: time.Now(),
		Server:    serverName,
		FileCount: len(files),
		Files:     files,
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return "", err
	}

	manifestPath := filepath.Join(releaseDir, "manifest.json")
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		return "", err
	}

	return releaseDir, nil
}

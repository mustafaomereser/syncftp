package milestone

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Milestone per-server zaman işareti: "milestone sync" bu tarihten sonra
// değişen (mtime) dosyaları yükler. Dosya: .syncftp/milestones/<server>.json
type Milestone struct {
	Server string    `json:"server"`
	Date   time.Time `json:"date"`
}

// Load sunucunun milestone'unu okur. Dosya yoksa (nil, nil) döner.
func Load(configDir, serverName string) (*Milestone, error) {
	data, err := os.ReadFile(msPath(configDir, serverName))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var m Milestone
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// Save sunucunun milestone'unu yazar (varsa üzerine).
func Save(configDir, serverName string, date time.Time) error {
	dir := filepath.Join(configDir, ".syncftp", "milestones")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	m := Milestone{Server: serverName, Date: date}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(msPath(configDir, serverName), data, 0644)
}

// Clear sunucunun milestone'unu siler. Dosya yoksa hata dönmez.
func Clear(configDir, serverName string) error {
	err := os.Remove(msPath(configDir, serverName))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func msPath(configDir, serverName string) string {
	return filepath.Join(configDir, ".syncftp", "milestones", serverName+".json")
}

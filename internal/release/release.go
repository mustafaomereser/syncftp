package release

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Manifest — bir sync turunda yüklenen dosyaların kaydı.
// Version 2: rollback alanları (Backed/New/PrevHashes) eklendi; yedek dosyaları
// release dizininin files/ alt klasöründedir.
type Manifest struct {
	Version      int               `json:"version"`
	ID           string            `json:"id,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	Server       string            `json:"server"`
	FileCount    int               `json:"file_count"`
	Files        map[string]string `json:"files"`                 // relPath -> "sha256:<hex>" (yüklenen yerel sürüm)
	Backed       []string          `json:"backed,omitempty"`      // eski uzak sürümü files/ altında yedeklenenler
	New          []string          `json:"new,omitempty"`         // upload öncesi sunucuda olmayanlar (rollback'te silinir)
	NoBackup     []string          `json:"no_backup,omitempty"`   // yedek alınamayanlar (geri alınamaz)
	PrevHashes   map[string]string `json:"prev_hashes,omitempty"` // upload öncesi state hash'i (rollback state'i buna döndürür)
	BackupPurged bool              `json:"backup_purged,omitempty"`
	RolledBack   bool              `json:"rolled_back,omitempty"`
	RolledBackAt *time.Time        `json:"rolled_back_at,omitempty"`
}

// HasBackup rollback mümkün mü (yedek var ve henüz silinmemiş).
func (m *Manifest) HasBackup() bool {
	return m.Version >= 2 && !m.BackupPurged && (len(m.Backed) > 0 || len(m.New) > 0)
}

func serverDir(configDir, serverName string) string {
	return filepath.Join(configDir, ".syncftp", "releases", serverName)
}

// Dir release dizini.
func Dir(configDir, serverName, id string) string {
	return filepath.Join(serverDir(configDir, serverName), id)
}

// FilesDir yedek dosyalarının kökü.
func FilesDir(configDir, serverName, id string) string {
	return filepath.Join(Dir(configDir, serverName, id), "files")
}

// NewID zaman damgalı benzersiz release kimliği.
func NewID(configDir, serverName string) string {
	base := time.Now().Format("20060102-150405")
	id := base
	for i := 2; ; i++ {
		if _, err := os.Stat(Dir(configDir, serverName, id)); os.IsNotExist(err) {
			return id
		}
		id = fmt.Sprintf("%s-%d", base, i)
	}
}

// Create writes a release manifest (yedeksiz, v1 uyumlu kullanım).
// Returns the path to the release directory.
func Create(configDir, serverName string, files map[string]string) (string, error) {
	m := &Manifest{Files: files}
	return Write(configDir, serverName, NewID(configDir, serverName), m)
}

// Write manifest'i id dizinine yazar; eksik alanları doldurur.
func Write(configDir, serverName, id string, m *Manifest) (string, error) {
	dir := Dir(configDir, serverName, id)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("release dizini oluşturulamadı: %w", err)
	}
	if m.Version == 0 {
		m.Version = 2
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	m.ID = id
	m.Server = serverName
	m.FileCount = len(m.Files)
	sort.Strings(m.Backed)
	sort.Strings(m.New)
	sort.Strings(m.NoBackup)
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), data, 0644); err != nil {
		return "", err
	}
	return dir, nil
}

// Load tek bir release manifest'ini okur.
func Load(configDir, serverName, id string) (*Manifest, error) {
	data, err := os.ReadFile(filepath.Join(Dir(configDir, serverName, id), "manifest.json"))
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	if m.ID == "" {
		m.ID = id
	}
	return &m, nil
}

// List sunucunun release'lerini en yeniden eskiye döner.
func List(configDir, serverName string) ([]*Manifest, error) {
	entries, err := os.ReadDir(serverDir(configDir, serverName))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []*Manifest
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if m, err := Load(configDir, serverName, e.Name()); err == nil {
			out = append(out, m)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID > out[j].ID
		}
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out, nil
}

// Discard yarım kalan (hiç başarılı dosyası olmayan) release dizinini siler.
func Discard(configDir, serverName, id string) {
	_ = os.RemoveAll(Dir(configDir, serverName, id))
}

// Prune en yeni keep release dışındakilerin yedek dosyalarını siler (manifest kalır,
// geçmiş görünür ama rollback edilemez). Silinen release sayısını döner.
func Prune(configDir, serverName string, keep int) int {
	list, err := List(configDir, serverName)
	if err != nil {
		return 0
	}
	n, withBackup := 0, 0
	for _, m := range list {
		if !m.HasBackup() {
			continue
		}
		withBackup++
		if withBackup <= keep {
			continue
		}
		_ = os.RemoveAll(FilesDir(configDir, serverName, m.ID))
		m.BackupPurged = true
		if _, err := Write(configDir, serverName, m.ID, m); err == nil {
			n++
		}
	}
	return n
}

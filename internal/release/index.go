package release

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"syncftp/internal/fsutil"
)

// Release özeti (index.json): liste ve sayaçlar için binlerce manifest.json'u açmamak —
// panel her proje açılışında, Geçmiş sekmesi ve her sync'in Prune'u eskiden sunucunun TÜM
// manifestlerini okuyordu (5000 release'li sunucuda soğuk diskte saniyeler). Manifestin
// tamamı yalnızca gerçekten lazım olduğunda (rollback, detay) Load ile okunur.
//
// Tutarlılık: Write/Discard index'i günceller; index'te olmayan klasör (başka sürüm/süreç
// yazmış) ya da klasörü silinmiş kayıt Summaries'te fark edilip onarılır.

// Summary bir release'in liste satırı.
type Summary struct {
	ID           string    `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	Version      int       `json:"version"`
	FileCount    int       `json:"file_count"`
	Backed       int       `json:"backed"`
	New          int       `json:"new"`
	NoBackup     int       `json:"no_backup"`
	HasBackup    bool      `json:"has_backup"`
	BackupPurged bool      `json:"backup_purged,omitempty"`
	RolledBack   bool      `json:"rolled_back,omitempty"`
}

// Restorable Manifest.Restorable ile aynı.
func (s Summary) Restorable() bool { return s.HasBackup && !s.RolledBack }

// Summary manifestin liste satırı.
func (m *Manifest) Summary() Summary {
	return Summary{ID: m.ID, CreatedAt: m.CreatedAt, Version: m.Version, FileCount: m.FileCount,
		Backed: len(m.Backed), New: len(m.New), NoBackup: len(m.NoBackup), HasBackup: m.HasBackup(),
		BackupPurged: m.BackupPurged, RolledBack: m.RolledBack}
}

type indexFile struct {
	Version int       `json:"version"`
	Entries []Summary `json:"entries"`
}

var indexMu sync.Mutex

func indexPath(configDir, serverName string) string {
	return filepath.Join(serverDir(configDir, serverName), "index.json")
}

func readIndex(configDir, serverName string) (map[string]Summary, bool) {
	data, err := os.ReadFile(indexPath(configDir, serverName))
	if err != nil {
		return nil, false
	}
	var f indexFile
	if json.Unmarshal(data, &f) != nil || f.Version != 1 {
		return nil, false
	}
	m := make(map[string]Summary, len(f.Entries))
	for _, e := range f.Entries {
		m[e.ID] = e
	}
	return m, true
}

func writeIndex(configDir, serverName string, m map[string]Summary) {
	f := indexFile{Version: 1, Entries: sortedSummaries(m)}
	if data, err := json.Marshal(f); err == nil {
		_ = fsutil.WriteFileAtomic(indexPath(configDir, serverName), data, 0644)
	}
}

func sortedSummaries(m map[string]Summary) []Summary {
	out := make([]Summary, 0, len(m))
	for _, s := range m {
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID > out[j].ID
		}
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out
}

// Summaries sunucunun release özetleri, en yeniden eskiye. Index yoksa ya da klasörlerle
// uyuşmuyorsa eksikler manifestlerden okunup index yeniden yazılır.
func Summaries(configDir, serverName string) ([]Summary, error) {
	entries, err := os.ReadDir(serverDir(configDir, serverName))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	ids := make(map[string]bool, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			ids[e.Name()] = true
		}
	}
	indexMu.Lock()
	defer indexMu.Unlock()
	idx, _ := readIndex(configDir, serverName)
	if idx == nil {
		idx = map[string]Summary{}
	}
	dirty := false
	for id := range idx {
		if !ids[id] {
			delete(idx, id)
			dirty = true
		}
	}
	var missing []string
	for id := range ids {
		if _, ok := idx[id]; !ok {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		dirty = true
		res := make([]*Manifest, len(missing))
		var wg sync.WaitGroup
		next := make(chan int)
		for w := 0; w < 8 && w < len(missing); w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := range next {
					if m, err := Load(configDir, serverName, missing[i]); err == nil {
						res[i] = m
					}
				}
			}()
		}
		for i := range missing {
			next <- i
		}
		close(next)
		wg.Wait()
		for _, m := range res {
			if m != nil {
				idx[m.ID] = m.Summary()
			}
		}
	}
	if dirty {
		writeIndex(configDir, serverName, idx)
	}
	return sortedSummaries(idx), nil
}

// updateIndex tek kaydı ekler/günceller (s == nil → siler). Index henüz yoksa dokunmaz —
// yarım index yazılmasın; ilk Summaries çağrısı tamamını kurar.
func updateIndex(configDir, serverName, id string, s *Summary) {
	indexMu.Lock()
	defer indexMu.Unlock()
	idx, ok := readIndex(configDir, serverName)
	if !ok {
		return
	}
	if s == nil {
		delete(idx, id)
	} else {
		idx[id] = *s
	}
	writeIndex(configDir, serverName, idx)
}

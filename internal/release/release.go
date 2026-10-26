package release

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"syncftp/internal/fsutil"
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
	// RolledBackFiles seçili dosya rollback'inde geri alınanlar (release'in tamamı geri alınınca
	// RolledBack da işaretlenir). Tam rollback bu dosyaları yeniden işlemez.
	RolledBackFiles []string `json:"rolled_back_files,omitempty"`
}

// FileRolledBack dosya bu release'te (tek tek ya da tamamıyla) geri alındı mı.
func (m *Manifest) FileRolledBack(rel string) bool {
	if m.RolledBack {
		return true
	}
	for _, f := range m.RolledBackFiles {
		if f == rel {
			return true
		}
	}
	return false
}

// Pending henüz geri alınmamış yedekli ve yeni dosyalar.
func (m *Manifest) Pending() (backed, created []string) {
	for _, f := range m.Backed {
		if !m.FileRolledBack(f) {
			backed = append(backed, f)
		}
	}
	for _, f := range m.New {
		if !m.FileRolledBack(f) {
			created = append(created, f)
		}
	}
	return backed, created
}

// HasBackup rollback mümkün mü (yedek var ve henüz silinmemiş).
func (m *Manifest) HasBackup() bool {
	return m.Version >= 2 && !m.BackupPurged && (len(m.Backed) > 0 || len(m.New) > 0)
}

func serverDir(configDir, serverName string) string {
	return filepath.Join(configDir, ".syncftp", "releases", fsutil.SafeName(serverName))
}

// Dir release dizini.
func Dir(configDir, serverName, id string) string {
	return filepath.Join(serverDir(configDir, serverName), id)
}

// FilesDir yedek dosyalarının kökü.
func FilesDir(configDir, serverName, id string) string {
	return filepath.Join(Dir(configDir, serverName, id), "files")
}

// AfterDir yüklenen (yeni) sürümlerin kopyası — release diff'i (eski yedek ↔ yüklenen) için.
func AfterDir(configDir, serverName, id string) string {
	return filepath.Join(Dir(configDir, serverName, id), "after")
}

// AfterMaxBytes bundan büyük dosyaların yeni sürümü saklanmaz (diff zaten metin dosyaları için).
const AfterMaxBytes = 2 << 20

// SaveAfter yüklenen dosyanın kopyasını after/ altına yazar; büyük dosyaları sessizce atlar.
func SaveAfter(configDir, serverName, id, rel, src string) error {
	info, err := os.Stat(src)
	if err != nil || info.Size() > AfterMaxBytes {
		return err
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	dst := filepath.Join(AfterDir(configDir, serverName, id), filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// NewID zaman damgalı benzersiz release kimliği.
func NewID(configDir, serverName string) string {
	base := time.Now().Format("20060102-150405")
	id := base
	for i := 2; ; i++ {
		// Yalnızca "var" ise sonraki sonek denenir; başka her hata (geçersiz yol, erişim) döngüyü
		// bitirir — eskiden "yok" dışı hata sonsuz döngüye sokuyordu (sync CPU'yu yiyip asılı kalıyordu)
		if _, err := os.Stat(Dir(configDir, serverName, id)); err != nil {
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
	sum := m.Summary()
	updateIndex(configDir, serverName, id, &sum)
	return dir, nil
}

// Load tek bir release manifest'ini okur.
func Load(configDir, serverName, id string) (*Manifest, error) {
	p := filepath.Join(Dir(configDir, serverName, id), "manifest.json")
	st, err := os.Stat(p)
	if err != nil {
		return nil, err
	}
	// Binlerce release'i olan projelerde (panel her açılışta hepsini listeler) dosyayı yeniden
	// okumamak için: boyut + mtime aynıysa önbellekteki kopya döner.
	key := manifestKey{size: st.Size(), mod: st.ModTime().UnixNano()}
	cacheMu.Lock()
	if c, ok := cache[p]; ok && c.key == key {
		cacheMu.Unlock()
		m := *c.m
		return &m, nil
	}
	cacheMu.Unlock()
	data, err := os.ReadFile(p)
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
	cp := m
	cacheMu.Lock()
	cache[p] = cachedManifest{key: key, m: &cp}
	cacheMu.Unlock()
	return &m, nil
}

type manifestKey struct {
	size int64
	mod  int64
}

type cachedManifest struct {
	key manifestKey
	m   *Manifest
}

var (
	cacheMu sync.Mutex
	cache   = map[string]cachedManifest{}
)

// List sunucunun release'lerini en yeniden eskiye döner.
func List(configDir, serverName string) ([]*Manifest, error) {
	entries, err := os.ReadDir(serverDir(configDir, serverName))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	// Binlerce release'li sunucularda ilk (önbelleksiz) okuma paralel yapılır
	var ids []string
	for _, e := range entries {
		if e.IsDir() {
			ids = append(ids, e.Name())
		}
	}
	res := make([]*Manifest, len(ids))
	var wg sync.WaitGroup
	next := make(chan int)
	for w := 0; w < 8 && w < len(ids); w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range next {
				if m, err := Load(configDir, serverName, ids[i]); err == nil {
					res[i] = m
				}
			}
		}()
	}
	for i := range ids {
		next <- i
	}
	close(next)
	wg.Wait()
	out := make([]*Manifest, 0, len(res))
	for _, m := range res {
		if m != nil {
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
	updateIndex(configDir, serverName, id, nil)
}

// Prune en yeni keep release dışındakilerin yedek dosyalarını siler (manifest kalır,
// geçmiş görünür ama rollback edilemez). Silinen release sayısını döner.
func Prune(configDir, serverName string, keep int) int {
	list, err := Summaries(configDir, serverName) // özetten seç; yalnızca budanacaklar okunur
	if err != nil {
		return 0
	}
	n, withBackup := 0, 0
	for _, sum := range list {
		if !sum.HasBackup {
			continue
		}
		withBackup++
		if withBackup <= keep {
			continue
		}
		m, err := Load(configDir, serverName, sum.ID)
		if err != nil {
			continue
		}
		_ = os.RemoveAll(FilesDir(configDir, serverName, m.ID))
		_ = os.RemoveAll(AfterDir(configDir, serverName, m.ID))
		m.BackupPurged = true
		if _, err := Write(configDir, serverName, m.ID, m); err == nil {
			n++
		}
	}
	return n
}

// Restorable release hâlâ geri alınabilir mi (yedeği var ve daha önce geri alınmamış).
func (m *Manifest) Restorable() bool { return m.HasBackup() && !m.RolledBack }

// SortedFiles yüklenen dosyalar, sıralı.
func (m *Manifest) SortedFiles() []string {
	out := make([]string, 0, len(m.Files))
	for f := range m.Files {
		out = append(out, f)
	}
	sort.Strings(out)
	return out
}

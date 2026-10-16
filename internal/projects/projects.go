// Package projects — makine genelindeki syncFTP proje kaydı.
// <UserConfigDir>/syncftp/projects.json dosyasında tutulur; GUI ve
// `--project` flag'i projeleri buradan bulur.
package projects

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type Project struct {
	Name     string    `json:"name"`
	Path     string    `json:"path"` // mutlak yol (.syncftp/ klasörünü içeren dizin)
	AddedAt  time.Time `json:"added_at"`
	LastUsed time.Time `json:"last_used,omitempty"`
}

type registry struct {
	Projects []Project `json:"projects"`
}

var mu sync.Mutex

// pathOverride testlerde gerçek kayıt dosyasına yazılmasın diye.
var pathOverride string

// Path kayıt dosyasının yolu.
func Path() string {
	if pathOverride != "" {
		return pathOverride
	}
	base, err := os.UserConfigDir()
	if err != nil {
		base = os.TempDir()
	}
	return filepath.Join(base, "syncftp", "projects.json")
}

func load() registry {
	var r registry
	data, err := os.ReadFile(Path())
	if err == nil {
		_ = json.Unmarshal(data, &r)
	}
	return r
}

func save(r registry) error {
	p := Path()
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

func samePath(a, b string) bool {
	return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
}

// HasConfig dizinde syncFTP config'i var mı.
func HasConfig(dir string) bool {
	st, err := os.Stat(filepath.Join(dir, ".syncftp", "syncftp.json"))
	return err == nil && !st.IsDir()
}

// List kayıtlı projeleri son kullanılana göre sıralı döner.
func List() []Project {
	mu.Lock()
	defer mu.Unlock()
	r := load()
	sort.SliceStable(r.Projects, func(i, j int) bool {
		return r.Projects[i].LastUsed.After(r.Projects[j].LastUsed)
	})
	return r.Projects
}

// Add projeyi kaydeder (zaten kayıtlıysa adını günceller). name boşsa dizin adı kullanılır.
// Aynı adda başka bir proje varsa sonuna -2, -3 eklenir.
func Add(dir, name string) (Project, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return Project{}, err
	}
	if !HasConfig(abs) {
		return Project{}, fmt.Errorf("%s: .syncftp/syncftp.json yok (önce `syncftp init`)", abs)
	}
	if name == "" {
		name = configProjectName(abs)
	}
	if name == "" {
		name = filepath.Base(abs)
	}
	mu.Lock()
	defer mu.Unlock()
	r := load()
	for i, p := range r.Projects {
		if samePath(p.Path, abs) {
			if r.Projects[i].Name != name {
				r.Projects[i].Name = uniqueName(r, name, i)
			}
			return r.Projects[i], save(r)
		}
	}
	p := Project{Name: uniqueName(r, name, -1), Path: abs, AddedAt: time.Now()}
	r.Projects = append(r.Projects, p)
	return p, save(r)
}

func uniqueName(r registry, name string, self int) string {
	taken := func(n string) bool {
		for i, p := range r.Projects {
			if i != self && strings.EqualFold(p.Name, n) {
				return true
			}
		}
		return false
	}
	out := name
	for i := 2; taken(out); i++ {
		out = fmt.Sprintf("%s-%d", name, i)
	}
	return out
}

// Remove kaydı siler (proje dosyalarına dokunmaz). Ad veya yol kabul eder.
func Remove(nameOrPath string) error {
	mu.Lock()
	defer mu.Unlock()
	r := load()
	for i, p := range r.Projects {
		if strings.EqualFold(p.Name, nameOrPath) || samePath(p.Path, nameOrPath) {
			r.Projects = append(r.Projects[:i], r.Projects[i+1:]...)
			return save(r)
		}
	}
	return fmt.Errorf("proje bulunamadı: %s", nameOrPath)
}

// Rename projenin görünen adını değiştirir.
func Rename(nameOrPath, newName string) error {
	mu.Lock()
	defer mu.Unlock()
	r := load()
	for i, p := range r.Projects {
		if strings.EqualFold(p.Name, nameOrPath) || samePath(p.Path, nameOrPath) {
			r.Projects[i].Name = uniqueName(r, newName, i)
			return save(r)
		}
	}
	return fmt.Errorf("proje bulunamadı: %s", nameOrPath)
}

// Find ad (büyük/küçük harf duyarsız) veya yol ile projeyi bulur.
func Find(nameOrPath string) (Project, bool) {
	for _, p := range List() {
		if strings.EqualFold(p.Name, nameOrPath) || samePath(p.Path, nameOrPath) {
			return p, true
		}
	}
	return Project{}, false
}

// Touch projenin son kullanım zamanını günceller; kayıtlı değilse sessizce ekler.
func Touch(dir string) {
	abs, err := filepath.Abs(dir)
	if err != nil || !HasConfig(abs) {
		return
	}
	mu.Lock()
	r := load()
	for i, p := range r.Projects {
		if samePath(p.Path, abs) {
			r.Projects[i].LastUsed = time.Now()
			_ = save(r)
			mu.Unlock()
			return
		}
	}
	mu.Unlock()
	if p, err := Add(abs, ""); err == nil {
		mu.Lock()
		r = load()
		for i := range r.Projects {
			if r.Projects[i].Name == p.Name {
				r.Projects[i].LastUsed = time.Now()
			}
		}
		_ = save(r)
		mu.Unlock()
	}
}

// configProjectName config'deki project.name (yoksa boş) — kayıt varsayılan adı.
func configProjectName(dir string) string {
	data, err := os.ReadFile(filepath.Join(dir, ".syncftp", "syncftp.json"))
	if err != nil {
		return ""
	}
	var c struct {
		Project struct {
			Name string `json:"name"`
		} `json:"project"`
	}
	if json.Unmarshal(data, &c) != nil {
		return ""
	}
	return strings.TrimSpace(c.Project.Name)
}

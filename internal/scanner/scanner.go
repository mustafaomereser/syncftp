package scanner

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"syncftp/internal/ignore"
)

type File struct {
	RelPath string // relative to project root, forward-slash separated
	AbsPath string
	Hash    string // "sha256:<hex>"
	Size    int64
	ModTime time.Time
}

// Scan walks dir, applies ignore rules, and returns all non-ignored files with SHA256 hashes.
// The .syncftp and .git directories are always excluded.
//
// Performans:
//   - Ignore edilen dizinlere hiç girilmez (MatchDir — "vendor/" pattern'i "vendor" dizinini de eşler).
//   - Hash cache: boyut + mtime değişmemiş dosya yeniden okunmaz (<UserCacheDir>/syncftp/hashcache/).
//   - Cache'te olmayan dosyalar CPU sayısı kadar paralel hash'lenir.
func Scan(dir string, matcher *ignore.Matcher) ([]File, error) {
	var files []File

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		if rel == "." {
			return nil
		}

		// Always skip internal/VCS directories
		if rel == ".syncftp" || strings.HasPrefix(rel, ".syncftp/") ||
			rel == ".git" || strings.HasPrefix(rel, ".git/") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			if matcher.MatchDir(rel) {
				return filepath.SkipDir
			}
			return nil
		}
		if matcher.Match(rel) {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}
		files = append(files, File{
			RelPath: rel,
			AbsPath: path,
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	cache := loadCache(dir)
	var todo []int
	for i := range files {
		f := &files[i]
		if e, ok := cache.Files[f.RelPath]; ok && e.Size == f.Size && e.MTime == f.ModTime.UnixNano() {
			f.Hash = e.Hash
		} else {
			todo = append(todo, i)
		}
	}

	if len(todo) > 0 {
		if err := hashParallel(files, todo); err != nil {
			return nil, err
		}
	}

	// Cache'i güncelle (silinen dosyalar düşer). Son 2 sn içinde değişmiş dosyalar
	// cache'lenmez: aynı mtime içinde ikinci bir yazım fark edilmeden kalmasın.
	recent := time.Now().Add(-2 * time.Second)
	next := hashCache{Files: make(map[string]cacheEntry, len(files))}
	changed := len(todo) > 0 || len(cache.Files) != len(files)
	for _, f := range files {
		if f.ModTime.After(recent) {
			changed = true
			continue
		}
		next.Files[f.RelPath] = cacheEntry{Size: f.Size, MTime: f.ModTime.UnixNano(), Hash: f.Hash}
	}
	if changed {
		saveCache(dir, next)
	}

	return files, nil
}

func hashParallel(files []File, todo []int) error {
	workers := runtime.NumCPU()
	if workers > 8 {
		workers = 8
	}
	if workers > len(todo) {
		workers = len(todo)
	}
	jobs := make(chan int)
	var (
		wg       sync.WaitGroup
		errOnce  sync.Once
		firstErr error
	)
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				h, err := hashFile(files[i].AbsPath)
				if err != nil {
					errOnce.Do(func() { firstErr = err })
					continue
				}
				files[i].Hash = h
			}
		}()
	}
	for _, i := range todo {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	return firstErr
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}

// ── Hash cache ────────────────────────────────────────────────────────────────

type cacheEntry struct {
	Size  int64  `json:"s"`
	MTime int64  `json:"m"`
	Hash  string `json:"h"`
}

type hashCache struct {
	Files map[string]cacheEntry `json:"files"`
}

var cacheMu sync.Mutex

// cacheDirOverride testlerde gerçek kullanıcı cache'ine yazılmasın diye.
var cacheDirOverride string

// cachePath tarama dizinine özel cache dosyası (mutlak yolun sha1'i).
func cachePath(dir string) string {
	abs, err := filepath.Abs(dir)
	if err != nil {
		abs = dir
	}
	sum := sha1.Sum([]byte(strings.ToLower(filepath.Clean(abs))))
	return filepath.Join(CacheDir(), hex.EncodeToString(sum[:])+".json")
}

// CacheDir hash cache dosyalarının bulunduğu dizin (<UserCacheDir>/syncftp/hashcache).
func CacheDir() string {
	if cacheDirOverride != "" {
		return cacheDirOverride
	}
	base, err := os.UserCacheDir()
	if err != nil {
		base = os.TempDir()
	}
	return filepath.Join(base, "syncftp", "hashcache")
}

// ClearAllCache tüm projelerin hash cache dosyalarını siler.
func ClearAllCache() error {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	return os.RemoveAll(CacheDir())
}

func loadCache(dir string) hashCache {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	c := hashCache{Files: map[string]cacheEntry{}}
	data, err := os.ReadFile(cachePath(dir))
	if err != nil {
		return c
	}
	if json.Unmarshal(data, &c) != nil || c.Files == nil {
		return hashCache{Files: map[string]cacheEntry{}}
	}
	return c
}

func saveCache(dir string, c hashCache) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	p := cachePath(dir)
	if os.MkdirAll(filepath.Dir(p), 0700) != nil {
		return
	}
	data, err := json.Marshal(c)
	if err != nil {
		return
	}
	tmp := p + ".tmp"
	if os.WriteFile(tmp, data, 0600) != nil {
		return
	}
	_ = os.Rename(tmp, p)
}

// ClearCache dizinin hash cache'ini siler (sonraki tarama tüm dosyaları yeniden hash'ler).
func ClearCache(dir string) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	_ = os.Remove(cachePath(dir))
}

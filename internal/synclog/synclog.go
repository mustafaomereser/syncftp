package synclog

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"
)

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;?]*[a-zA-Z]`)

// StripANSI removes terminal color/style escape codes from s.
func StripANSI(s string) string { return ansiRe.ReplaceAllString(s, "") }

// Save writes content to .syncftp/logs/<server>/<timestamp>.txt and returns the path.
// ANSI escape codes are stripped so the log stays readable in any editor.
func Save(configDir, serverName, content string) (string, error) {
	logDir := Dir(configDir, serverName)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return "", err
	}
	// Aynı saniyede birden fazla log (sync + rollback) birbirini ezmesin
	base := time.Now().Format("20060102-150405")
	for i := 1; ; i++ {
		name := base + ".txt"
		if i > 1 {
			name = fmt.Sprintf("%s-%d.txt", base, i)
		}
		p := filepath.Join(logDir, name)
		f, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if os.IsExist(err) {
			continue
		}
		if err != nil {
			return "", err
		}
		_, werr := f.WriteString(StripANSI(content))
		if cerr := f.Close(); werr == nil {
			werr = cerr
		}
		return p, werr
	}
}

// Dir sunucunun log klasörü.
func Dir(configDir, serverName string) string {
	return filepath.Join(configDir, ".syncftp", "logs", serverName)
}

// Entry tek bir log dosyası.
type Entry struct {
	Server string    `json:"server"`
	File   string    `json:"file"`
	Size   int64     `json:"size"`
	Time   time.Time `json:"time"`
}

// List verilen sunucuların loglarını en yeniden eskiye döner (limit > 0 ise o kadar).
func List(configDir string, servers []string, limit int) []Entry {
	out := []Entry{}
	for _, s := range servers {
		ents, _ := os.ReadDir(Dir(configDir, s))
		for _, e := range ents {
			if e.IsDir() {
				continue
			}
			if info, err := e.Info(); err == nil {
				out = append(out, Entry{Server: s, File: e.Name(), Size: info.Size(), Time: info.ModTime()})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Time.After(out[j].Time) })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}

// Read log dosyasını okur (adlar dizin dışına çıkamaz); maxBytes > 0 ise yalnızca son kısmı.
func Read(configDir, serverName, file string, maxBytes int) (string, error) {
	data, err := os.ReadFile(filepath.Join(Dir(configDir, filepath.Base(serverName)), filepath.Base(file)))
	if err != nil {
		return "", err
	}
	if maxBytes > 0 && len(data) > maxBytes {
		data = data[len(data)-maxBytes:]
	}
	return string(data), nil
}

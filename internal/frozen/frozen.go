package frozen

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"syncftp/internal/fsutil"
)

// List holds the set of frozen relative paths for one server.
type List struct {
	Files map[string]bool
}

func Load(configDir, serverName string) (*List, error) {
	p := fsutil.ServerFile(filepath.Join(configDir, ".syncftp", "frozen"), serverName, ".json")
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return &List{Files: make(map[string]bool)}, nil
	}
	if err != nil {
		return nil, err
	}
	var paths []string
	if err := json.Unmarshal(data, &paths); err != nil {
		return nil, err
	}
	l := &List{Files: make(map[string]bool, len(paths))}
	for _, p := range paths {
		l.Files[p] = true
	}
	return l, nil
}

func Save(configDir, serverName string, paths []string) error {
	dir := filepath.Join(configDir, ".syncftp", "frozen")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(paths, "", "  ")
	if err != nil {
		return err
	}
	return fsutil.WriteFileAtomic(fsutil.ServerFile(dir, serverName, ".json"), data, 0600)
}

// IsFrozen yol freeze'li mi: listede birebir, ya da üst klasörlerinden biri "klasör/" olarak
// listede (klasör freeze'i — altındaki her dosya, sonradan eklenenler dahil). Eskiden klasör
// freeze'i sunucudaki tüm dosyaları tek tek listeye yazıyordu; binlerce dosyalı klasörde
// (node_modules, vendor, uploads) hem listeleme hem liste kilitleniyordu.
func IsFrozen(l *List, relPath string) bool {
	if l == nil {
		return false
	}
	return MatchSet(l.Files, relPath)
}

// MatchSet IsFrozen'ın küme üzerinde çalışan hâli (tarayıcı / panel freeze çekirdeği).
// relPath "x/" ile bitiyorsa klasörün kendisi sorulur.
func MatchSet(set map[string]bool, relPath string) bool {
	if set[relPath] {
		return true
	}
	for i := 0; i < len(relPath)-1; i++ {
		if relPath[i] == '/' && set[relPath[:i+1]] {
			return true
		}
	}
	return false
}

// Dirs listedeki klasör freeze'leri ("x/").
func Dirs(l *List) []string {
	if l == nil {
		return nil
	}
	var out []string
	for p, v := range l.Files {
		if v && strings.HasSuffix(p, "/") {
			out = append(out, p)
		}
	}
	return out
}

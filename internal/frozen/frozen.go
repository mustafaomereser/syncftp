package frozen

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// List holds the set of frozen relative paths for one server.
type List struct {
	Files map[string]bool
}

func Load(configDir, serverName string) (*List, error) {
	p := filepath.Join(configDir, ".syncftp", "frozen", serverName+".json")
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
	return os.WriteFile(filepath.Join(dir, serverName+".json"), data, 0600)
}

func IsFrozen(l *List, relPath string) bool {
	if l == nil {
		return false
	}
	return l.Files[relPath]
}

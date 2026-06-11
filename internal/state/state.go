package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type State struct {
	Version       int               `json:"version"`
	Server        string            `json:"server"`
	FirstSyncDone bool              `json:"first_sync_done"`
	LastSync      time.Time         `json:"last_sync"`
	Files         map[string]string `json:"files"` // relPath -> "sha256:<hex>"
}

type DiffResult struct {
	New     []string // in current but not in state
	Changed []string // in both but hash differs
	Deleted []string // in state but not in current (local delete)
}

// Load reads the sync state for the given server.
// Returns an empty state if no state file exists yet (first sync).
func Load(configDir, serverName string) (*State, error) {
	p := statePath(configDir, serverName)
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return &State{
			Version: 1,
			Server:  serverName,
			Files:   make(map[string]string),
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("state okunamadı (%s): %w", p, err)
	}

	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("state parse edilemedi: %w", err)
	}
	if s.Files == nil {
		s.Files = make(map[string]string)
	}
	return &s, nil
}

// Save writes the state to disk under <configDir>/.syncftp/state/<server>.json.
func Save(configDir string, s *State) error {
	dir := filepath.Join(configDir, ".syncftp", "state")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	s.LastSync = time.Now()
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(statePath(configDir, s.Server), data, 0644)
}

// Diff compares stored hashes against the current file scan.
func Diff(s *State, current map[string]string) DiffResult {
	var result DiffResult
	for path, hash := range current {
		stored, exists := s.Files[path]
		if !exists {
			result.New = append(result.New, path)
		} else if stored != hash {
			result.Changed = append(result.Changed, path)
		}
	}
	for path := range s.Files {
		if _, exists := current[path]; !exists {
			result.Deleted = append(result.Deleted, path)
		}
	}
	return result
}

func statePath(configDir, serverName string) string {
	return filepath.Join(configDir, ".syncftp", "state", serverName+".json")
}

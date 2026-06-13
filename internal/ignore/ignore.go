package ignore

import (
	"bufio"
	"os"
	"path/filepath"

	gitignore "github.com/sabhiram/go-gitignore"
)

// defaultIgnoreFiles is used when no explicit list is configured.
var defaultIgnoreFiles = []string{".gitignore", "syncftp.ignore"}

// Matcher checks whether a file path should be ignored.
type Matcher struct {
	gi *gitignore.GitIgnore
}

// Load reads the given ignore files from dir and merges their patterns.
// Pass nil to use the default list (.gitignore + syncftp.ignore).
// Pass an empty (non-nil) slice to use no ignore files at all.
// Files that do not exist are silently skipped.
func Load(dir string, ignoreFiles []string) (*Matcher, error) {
	if ignoreFiles == nil {
		ignoreFiles = defaultIgnoreFiles
	}
	if len(ignoreFiles) == 0 {
		return &Matcher{}, nil
	}
	var lines []string
	for _, name := range ignoreFiles {
		p := filepath.Join(dir, name)
		extra, err := readLines(p)
		if err != nil {
			return nil, err
		}
		lines = append(lines, extra...)
	}
	if len(lines) == 0 {
		return &Matcher{}, nil
	}
	return &Matcher{gi: gitignore.CompileIgnoreLines(lines...)}, nil
}

// readLines returns all lines from a file.
// If the file does not exist it returns nil, nil.
func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines, sc.Err()
}

// Match returns true if the given relative path should be excluded from sync.
func (m *Matcher) Match(relPath string) bool {
	if m.gi == nil {
		return false
	}
	return m.gi.MatchesPath(relPath)
}

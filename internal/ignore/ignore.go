package ignore

import (
	"os"
	"path/filepath"

	gitignore "github.com/sabhiram/go-gitignore"
)

// Matcher checks whether a file path should be ignored.
type Matcher struct {
	gi *gitignore.GitIgnore
}

// Load finds .gitignore or syncftp.ignore in dir (in that order of priority).
// If neither exists, returns a no-op Matcher that ignores nothing.
func Load(dir string) (*Matcher, error) {
	for _, name := range []string{".gitignore", "syncftp.ignore"} {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err == nil {
			gi, err := gitignore.CompileIgnoreFile(p)
			if err != nil {
				return nil, err
			}
			return &Matcher{gi: gi}, nil
		}
	}
	return &Matcher{}, nil
}

// Match returns true if the given relative path should be excluded from sync.
func (m *Matcher) Match(relPath string) bool {
	if m.gi == nil {
		return false
	}
	return m.gi.MatchesPath(relPath)
}

package synclog

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;?]*[a-zA-Z]`)

// StripANSI removes terminal color/style escape codes from s.
func StripANSI(s string) string { return ansiRe.ReplaceAllString(s, "") }

// Save writes content to .syncftp/logs/<server>/<timestamp>.txt and returns the path.
// ANSI escape codes are stripped so the log stays readable in any editor.
func Save(configDir, serverName, content string) (string, error) {
	logDir := filepath.Join(configDir, ".syncftp", "logs", serverName)
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

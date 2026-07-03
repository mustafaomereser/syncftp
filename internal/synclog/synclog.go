package synclog

import (
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
	p := filepath.Join(logDir, time.Now().Format("20060102-150405")+".txt")
	if err := os.WriteFile(p, []byte(StripANSI(content)), 0644); err != nil {
		return "", err
	}
	return p, nil
}

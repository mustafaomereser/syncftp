package scanner

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"

	"syncftp/internal/ignore"
)

type File struct {
	RelPath string // relative to project root, forward-slash separated
	AbsPath string
	Hash    string // "sha256:<hex>"
}

// Scan walks dir, applies ignore rules, and returns all non-ignored files with SHA256 hashes.
// The .syncftp directory and syncftp.config are always excluded.
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

		if matcher.Match(rel) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			return nil
		}

		hash, err := hashFile(path)
		if err != nil {
			return err
		}

		files = append(files, File{
			RelPath: rel,
			AbsPath: path,
			Hash:    hash,
		})
		return nil
	})

	return files, err
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

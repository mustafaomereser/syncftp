package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	ftpclient "syncftp/internal/ftp"
	"syncftp/internal/lang"
)

// minifiableExt CSS veya JS dosyası ise true döner.
func minifiableExt(relPath string) bool {
	ext := strings.ToLower(filepath.Ext(relPath))
	return ext == ".css" || ext == ".js"
}

// minifyCSS basit CSS küçültme: yorum kaldır, gereksiz boşlukları daralt.
func minifyCSS(data []byte) []byte {
	s := string(data)
	// /* ... */ yorumlarını kaldır
	for {
		start := strings.Index(s, "/*")
		if start == -1 {
			break
		}
		end := strings.Index(s[start:], "*/")
		if end == -1 {
			break
		}
		s = s[:start] + s[start+end+2:]
	}
	// Satır sonu ve tab → boşluk
	s = strings.NewReplacer("\r\n", " ", "\r", " ", "\n", " ", "\t", " ").Replace(s)
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	// { } : ; , öncesi/sonrası gereksiz boşlukları kaldır
	for _, ch := range []string{"{", "}", ":", ";", ","} {
		s = strings.ReplaceAll(s, " "+ch, ch)
		s = strings.ReplaceAll(s, ch+" ", ch)
	}
	return []byte(strings.TrimSpace(s))
}

// minifyJS temel JS küçültme: // satır yorumlarını ve fazla boşlukları kaldır.
// Not: string literal içi // veya /* yanlış işlenebilir — üretim için obfuscate=true ile terser kullanın.
func minifyJS(data []byte) []byte {
	lines := strings.Split(string(data), "\n")
	var out []string
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			continue
		}
		out = append(out, trimmed)
	}
	s := strings.Join(out, " ")
	// /* ... */ yorumlarını kaldır
	for {
		start := strings.Index(s, "/*")
		if start == -1 {
			break
		}
		end := strings.Index(s[start:], "*/")
		if end == -1 {
			break
		}
		s = s[:start] + " " + s[start+end+2:]
	}
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return []byte(strings.TrimSpace(s))
}

// terserAvailable PATH'te `terser` komutunun varlığını önbelleğe alır.
var terserAvailable *bool

func hasTerser() bool {
	if terserAvailable != nil {
		return *terserAvailable
	}
	_, err := exec.LookPath("terser")
	v := err == nil
	terserAvailable = &v
	return v
}

// obfuscateWithTerser terser komutunu stdin üzerinden çalıştırır.
func obfuscateWithTerser(data []byte) ([]byte, error) {
	cmd := exec.Command("terser", "--compress", "--mangle", "--")
	cmd.Stdin = bytes.NewReader(data)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ProcessMinify verilen görevleri minify/obfuscate ederek yeni görev listesi ve temp dosyaları döner.
// Hata durumunda o görev için orijinal task kullanılmaya devam eder.
// Çağıran defer ile tmpFiles'ı temizlemeli: for _, p := range tmpFiles { os.Remove(p) }
func ProcessMinify(tasks []ftpclient.UploadTask, minify, obfuscate bool, tmpDir string) (out []ftpclient.UploadTask, tmpFiles []string, minifiedCount, obfuscatedCount int) {
	if !minify && !obfuscate {
		return tasks, nil, 0, 0
	}

	terserWarned := false

	for _, t := range tasks {
		ext := strings.ToLower(filepath.Ext(t.RelPath))
		isCSS := ext == ".css"
		isJS := ext == ".js"

		if !isCSS && !isJS {
			out = append(out, t)
			continue
		}

		src, err := os.ReadFile(t.LocalPath)
		if err != nil {
			fmt.Printf(lang.L.SyncMinifyErrFmt, t.RelPath, err)
			out = append(out, t)
			continue
		}

		var processed []byte
		didMinify := false
		didObfuscate := false

		if minify {
			if isCSS {
				processed = minifyCSS(src)
			} else {
				processed = minifyJS(src)
			}
			didMinify = true
		} else {
			processed = src
		}

		if obfuscate && isJS {
			if !hasTerser() {
				if !terserWarned {
					fmt.Print(lang.L.SyncNoTerserWarn)
					terserWarned = true
				}
			} else {
				if obfOut, oErr := obfuscateWithTerser(processed); oErr != nil {
					fmt.Printf(lang.L.SyncObfuscErrFmt, t.RelPath, oErr)
				} else {
					processed = obfOut
					didObfuscate = true
				}
			}
		}

		// Temp dosya yaz — çakışmayı önlemek için RelPath'teki / → _
		safeRel := strings.ReplaceAll(t.RelPath, "/", "_")
		tmpPath := filepath.Join(tmpDir, safeRel)
		if wErr := os.WriteFile(tmpPath, processed, 0600); wErr != nil {
			fmt.Printf(lang.L.SyncMinifyErrFmt, t.RelPath, wErr)
			out = append(out, t)
			continue
		}
		tmpFiles = append(tmpFiles, tmpPath)
		out = append(out, ftpclient.UploadTask{LocalPath: tmpPath, RelPath: t.RelPath, Hash: t.Hash})
		if didMinify {
			minifiedCount++
		}
		if didObfuscate {
			obfuscatedCount++
		}
	}
	return out, tmpFiles, minifiedCount, obfuscatedCount
}

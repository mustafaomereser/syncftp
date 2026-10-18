package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"syncftp/internal/config"
	"syncftp/internal/lang"
	"syncftp/internal/release"
)

// Release dosya diff'i — CLI `rollback <id> --diff` ve panel Geçmiş → release → dosya ortak.
// Eski hal: release yedeği (files/, sunucudaki ham bayt) ya da yeni dosyaysa boş.
// Yeni hal: release'e kaydedilen kopya (after/); eski release'lerde yoksa yerel dosya —
// yalnızca hash'i manifestteki yüklenen hash'le AYNIYSA (sonradan değiştiyse yanıltmasın).

type releaseFileView struct {
	Path        string `json:"path"`
	Status      string `json:"status"` // changed | new | nobackup
	Before      string `json:"before"`
	After       string `json:"after"`
	BeforeOK    bool   `json:"before_ok"`
	AfterOK     bool   `json:"after_ok"`
	AfterSource string `json:"after_source,omitempty"` // release | local
	Binary      bool   `json:"binary,omitempty"`
}

// releaseFileVersions bir release'teki dosyanın eski / yeni içeriğini hazırlar (\r\n → \n).
func releaseFileVersions(dir string, cfg *config.Config, srv config.Server, m *release.Manifest, rel string) (releaseFileView, error) {
	rel = editRel(rel)
	hash, ok := m.Files[rel]
	if !ok || rel == "" {
		return releaseFileView{}, fmt.Errorf(lang.L.ReleaseFileNotFoundFmt, rel)
	}
	v := releaseFileView{Path: rel, Status: "nobackup"}
	inList := func(list []string) bool {
		for _, x := range list {
			if x == rel {
				return true
			}
		}
		return false
	}
	var before, after []byte
	switch {
	case inList(m.New):
		v.Status, v.BeforeOK = "new", true
	case inList(m.Backed):
		v.Status = "changed"
		if data, err := readCapped(filepath.Join(release.FilesDir(dir, srv.Name, m.ID), filepath.FromSlash(rel))); err == nil {
			before, v.BeforeOK = data, true
		}
	}
	if data, err := readCapped(filepath.Join(release.AfterDir(dir, srv.Name, m.ID), filepath.FromSlash(rel))); err == nil {
		after, v.AfterOK, v.AfterSource = data, true, "release"
	} else if lp, err := editLocalPath(dir, cfg, srv, rel); err == nil {
		if data, err := readCapped(lp); err == nil && contentHash(data) == hash {
			after, v.AfterOK, v.AfterSource = data, true, "local"
		}
	}
	if isBinaryBytes(before) || isBinaryBytes(after) {
		v.Binary = true
		return v, nil
	}
	v.Before = strings.ReplaceAll(string(before), "\r\n", "\n")
	v.After = strings.ReplaceAll(string(after), "\r\n", "\n")
	return v, nil
}

func readCapped(p string) ([]byte, error) {
	info, err := os.Stat(p)
	if err != nil {
		return nil, err
	}
	if info.Size() > editMaxBytes {
		return nil, fmt.Errorf("large")
	}
	return os.ReadFile(p)
}

func isBinaryBytes(b []byte) bool {
	return bytes.IndexByte(b[:min(len(b), 8000)], 0) >= 0
}

// writeReleaseDiff release'in dosyalarını (only boş değilse yalnızca onu) unified diff olarak yazar.
func writeReleaseDiff(w io.Writer, dir string, cfg *config.Config, srv config.Server, m *release.Manifest, only string) error {
	files := m.SortedFiles()
	if only != "" {
		files = []string{editRel(only)}
	}
	for _, rel := range files {
		v, err := releaseFileVersions(dir, cfg, srv, m, rel)
		if err != nil {
			return err
		}
		fmt.Fprintf(w, "\n── %s ──\n", rel)
		switch {
		case v.Binary:
			fmt.Fprintln(w, lang.L.ReleaseDiffBinary)
		case !v.BeforeOK:
			fmt.Fprintln(w, lang.L.ReleaseDiffNoBefore)
		case !v.AfterOK:
			fmt.Fprintln(w, lang.L.ReleaseDiffNoAfter)
		default:
			writeUnified(w, v.Before, v.After)
		}
	}
	return nil
}

// writeUnified satır bazlı LCS diff'i (±3 satır bağlam) yazar; çok büyük dosyada özet verir.
func writeUnified(w io.Writer, a, b string) {
	A, B := splitLines(a), splitLines(b)
	ops, ok := diffLineOps(A, B)
	if !ok {
		fmt.Fprintf(w, lang.L.ReleaseDiffTooBigFmt, len(A), len(B))
		return
	}
	const ctx = 3
	keep := make([]bool, len(ops))
	for i, o := range ops {
		if o.op != ' ' {
			for j := max(0, i-ctx); j <= min(len(ops)-1, i+ctx); j++ {
				keep[j] = true
			}
		}
	}
	gap := false
	for i, o := range ops {
		if !keep[i] {
			gap = true
			continue
		}
		if gap {
			fmt.Fprintln(w, "  …")
			gap = false
		}
		switch o.op {
		case '+':
			fmt.Fprintf(w, "\033[32m+ %s\033[0m\n", o.text)
		case '-':
			fmt.Fprintf(w, "\033[31m- %s\033[0m\n", o.text)
		default:
			fmt.Fprintf(w, "  %s\n", o.text)
		}
	}
	if !anyChange(ops) {
		fmt.Fprintln(w, lang.L.ReleaseDiffSame)
	}
}

type lineOp struct {
	op   byte // ' ' | '+' | '-'
	text string
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(strings.TrimSuffix(s, "\n"), "\n")
}

func anyChange(ops []lineOp) bool {
	for _, o := range ops {
		if o.op != ' ' {
			return true
		}
	}
	return false
}

// diffLineOps ortak baş/son kırpılmış LCS (panelin HL.diffLines'ı ile aynı yaklaşım); eşitlikte önce silme.
func diffLineOps(A, B []string) ([]lineOp, bool) {
	pre := 0
	for pre < len(A) && pre < len(B) && A[pre] == B[pre] {
		pre++
	}
	suf := 0
	for suf < len(A)-pre && suf < len(B)-pre && A[len(A)-1-suf] == B[len(B)-1-suf] {
		suf++
	}
	a, b := A[pre:len(A)-suf], B[pre:len(B)-suf]
	if len(a)*len(b) > 4_000_000 {
		return nil, false
	}
	n, m := len(a), len(b)
	dp := make([][]int32, n+1)
	for i := range dp {
		dp[i] = make([]int32, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if a[i] == b[j] {
				dp[i][j] = dp[i+1][j+1] + 1
			} else {
				dp[i][j] = dp[i+1][j]
				if dp[i][j+1] > dp[i][j] {
					dp[i][j] = dp[i][j+1]
				}
			}
		}
	}
	ops := make([]lineOp, 0, len(A)+len(B))
	for _, s := range A[:pre] {
		ops = append(ops, lineOp{' ', s})
	}
	i, j := 0, 0
	for i < n || j < m {
		switch {
		case i < n && j < m && a[i] == b[j]:
			ops = append(ops, lineOp{' ', a[i]})
			i++
			j++
		case i < n && (j == m || dp[i+1][j] >= dp[i][j+1]):
			ops = append(ops, lineOp{'-', a[i]})
			i++
		default:
			ops = append(ops, lineOp{'+', b[j]})
			j++
		}
	}
	for _, s := range A[len(A)-suf:] {
		ops = append(ops, lineOp{' ', s})
	}
	return ops, true
}

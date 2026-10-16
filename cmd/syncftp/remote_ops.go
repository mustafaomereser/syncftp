package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"

	goftp "github.com/jlaffaye/ftp"

	ftpclient "syncftp/internal/ftp"
)

// Uzak dosya işlemlerinin ortak çekirdeği — CLI `remote`, shell, dosya tarayıcısı (TUI), serve API ve
// panel aynı fonksiyonları kullanır; yalnızca çıktı biçimi (tablo / JSON / TUI) çağırana aittir.

// remoteEntries dizini listeler: "." / ".." atılır, önce klasörler, sonra büyük/küçük harf duyarsız ad.
func remoteEntries(c *ftpclient.Client, dir string) ([]*goftp.Entry, error) {
	raw, err := c.List(dir)
	if err != nil {
		return nil, err
	}
	out := make([]*goftp.Entry, 0, len(raw))
	for _, e := range raw {
		if e == nil || e.Name == "." || e.Name == ".." {
			continue
		}
		out = append(out, e)
	}
	sort.SliceStable(out, func(i, j int) bool {
		di, dj := out[i].Type == goftp.EntryTypeFolder, out[j].Type == goftp.EntryTypeFolder
		if di != dj {
			return di
		}
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out, nil
}

// entryKind "dir" / "link" / "file".
func entryKind(e *goftp.Entry) string {
	switch e.Type {
	case goftp.EntryTypeFolder:
		return "dir"
	case goftp.EntryTypeLink:
		return "link"
	}
	return "file"
}

// remoteIsDir yolun klasör olup olmadığını üst dizindeki kaydından anlar. Eskiden her yerde
// "List(p) hata vermediyse klasör" deniyordu: boş klasörler (550 → boş liste) dosya, bazı
// sunucularda List(dosya) dosyanın kendisini döndüğü için dosyalar klasör sanılıyordu.
func remoteIsDir(c *ftpclient.Client, p string) bool {
	p = path.Clean(p)
	if p == "/" || p == "." {
		return true
	}
	if entries, err := c.List(path.Dir(p)); err == nil {
		base := path.Base(p)
		for _, e := range entries {
			if e != nil && e.Name == base {
				return e.Type == goftp.EntryTypeFolder
			}
		}
	}
	// Üst dizin listelenemedi (izin vb.) → eski yönteme düş
	entries, err := c.List(p)
	return err == nil && len(entries) > 0
}

// remoteDelete dosya ya da klasör siler (türü kendisi bulur). recursive=false iken dolu klasör silinmez.
func remoteDelete(c *ftpclient.Client, p string, recursive bool) (isDir bool, err error) {
	if remoteIsDir(c, p) {
		return true, c.DeleteDir(p, recursive)
	}
	return false, c.DeleteFile(p)
}

// remoteMove taşır/yeniden adlandırır; hedefin üst klasörü yoksa oluşturulur (panel + shell taşıma).
func remoteMove(c *ftpclient.Client, from, to string) error {
	if path.Dir(to) != path.Dir(from) {
		_ = c.MakeDirAll(path.Dir(to))
	}
	return c.Rename(from, to)
}

// textPreview önizleme sonucu: ikili dosyada içerik boş, satır sonları \n.
type textPreview struct {
	Content   string `json:"content"`
	Binary    bool   `json:"binary"`
	Truncated bool   `json:"truncated"`
	Size      int    `json:"size"`
}

func makePreview(data []byte, max int) textPreview {
	p := textPreview{Size: len(data), Truncated: max > 0 && len(data) >= max}
	for _, b := range data[:min(len(data), 8000)] {
		if b == 0 {
			p.Binary = true
			return p
		}
	}
	p.Content = strings.ReplaceAll(string(data), "\r\n", "\n")
	return p
}

func remotePreview(c *ftpclient.Client, p string, max int) (textPreview, error) {
	data, err := c.Preview(p, int64(max))
	if err != nil {
		return textPreview{}, err
	}
	return makePreview(data, max), nil
}

func localPreview(file string, max int) (textPreview, bool) {
	f, err := os.Open(file)
	if err != nil {
		return textPreview{}, false
	}
	defer f.Close()
	data, _ := io.ReadAll(io.LimitReader(f, int64(max)))
	return makePreview(data, max), true
}

// printPreview önizlemeyi terminale yazar (CLI `remote cat` ve shell `cat`).
func printPreview(w io.Writer, p textPreview, truncNote string) {
	if p.Binary {
		fmt.Fprintf(w, "(binary, %s)\n", formatSize(uint64(p.Size)))
		return
	}
	fmt.Fprint(w, p.Content)
	if !strings.HasSuffix(p.Content, "\n") {
		fmt.Fprintln(w)
	}
	if p.Truncated && truncNote != "" {
		fmt.Fprint(w, truncNote)
	}
}

// serveRemoteDownload uzak dosyayı geçici dosyaya indirip ek olarak sunar (panel + serve API).
func serveRemoteDownload(w http.ResponseWriter, r *http.Request, c *ftpclient.Client, full string) error {
	tmp, err := os.CreateTemp("", "syncftp-dl-*")
	if err != nil {
		return err
	}
	tmp.Close()
	defer os.Remove(tmp.Name())
	if err := c.Download(full, tmp.Name()); err != nil {
		return err
	}
	f, err := os.Open(tmp.Name())
	if err != nil {
		return err
	}
	defer f.Close()
	st, _ := f.Stat()
	name := path.Base(full)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, strings.ReplaceAll(name, `"`, "")))
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeContent(w, r, name, st.ModTime(), f)
	return nil
}

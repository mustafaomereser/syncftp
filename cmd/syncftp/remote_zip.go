package main

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	goftp "github.com/jlaffaye/ftp"

	ftpclient "syncftp/internal/ftp"
	"syncftp/internal/lang"
)

// Sunucuda zip / zipten çıkar — CLI `remote zip|unzip` ve panel işleri (zip, unzip) aynı çekirdek.
// FTP'de sunucuda komut çalıştırılamadığı için iş yerelde yapılır: dosyalar indirilip arşive yazılır,
// arşiv yüklenir; çıkarırken arşiv indirilip her dosya tek tek yüklenir.

// unzipMaxTotal açılmış toplam boyut sınırı (zip bombası koruması).
const unzipMaxTotal = 8 << 30

// remoteZip baseDir içindeki srcs öğelerini (mutlak uzak yollar; dosya ya da klasör) outFull arşivine
// yazar. Arşivdeki adlar baseDir'e göreceli. Arşiv geçici dosyada kurulur, sona kadar sunucuya yazılmaz
// → iptal ya da hatada sunucuda hiçbir şey değişmez.
func remoteZip(ctx context.Context, c *ftpclient.Client, baseDir string, srcs []string, outFull string, w io.Writer) (files int, size int64, err error) {
	tmp, err := os.CreateTemp("", "syncftp-zip-*.zip")
	if err != nil {
		return 0, 0, err
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()
	zw := zip.NewWriter(tmp)
	rel := func(full string) string { return strings.TrimPrefix(strings.TrimPrefix(full, baseDir), "/") }

	addFile := func(full string, e *goftp.Entry) error {
		if full == outFull {
			return nil // arşivin eski hâli kendi içine girmesin
		}
		h := &zip.FileHeader{Name: rel(full), Method: zip.Deflate}
		if e != nil && !e.Time.IsZero() {
			h.Modified = e.Time
		}
		fw, err := zw.CreateHeader(h)
		if err != nil {
			return err
		}
		if _, err := c.RetrieveTo(full, fw); err != nil {
			return err
		}
		files++
		fmt.Fprintf(w, lang.L.RemoteZipItemFmt, rel(full))
		return nil
	}
	var walk func(dir string) error
	walk = func(dir string) error {
		if _, err := zw.Create(rel(dir) + "/"); err != nil { // boş klasörler de arşivde kalsın
			return err
		}
		entries, err := remoteEntries(c, dir)
		if err != nil && !ftpclient.IsNotFound(err) {
			return err
		}
		for _, e := range entries {
			if err := ctx.Err(); err != nil {
				return err
			}
			full := path.Join(dir, e.Name)
			if e.Type == goftp.EntryTypeFolder {
				if err := walk(full); err != nil {
					return err
				}
				continue
			}
			if err := addFile(full, e); err != nil {
				return err
			}
		}
		return nil
	}
	for _, src := range srcs {
		if err := ctx.Err(); err != nil {
			return files, 0, err
		}
		if remoteIsDir(c, src) {
			err = walk(src)
		} else {
			err = addFile(src, nil)
		}
		if err != nil {
			return files, 0, err
		}
	}
	if err := zw.Close(); err != nil {
		return files, 0, err
	}
	if err := tmp.Close(); err != nil {
		return files, 0, err
	}
	if err := ctx.Err(); err != nil {
		return files, 0, err
	}
	if st, err := os.Stat(tmp.Name()); err == nil {
		size = st.Size()
	}
	if err := c.UploadTo(tmp.Name(), outFull); err != nil {
		return files, 0, err
	}
	return files, size, nil
}

// uniqueRemoteName dir içinde name kullanılıyorsa "ad-2.zip", "ad-3.zip"… döner.
func uniqueRemoteName(c *ftpclient.Client, dir, name string) string {
	taken := map[string]bool{}
	if entries, err := remoteEntries(c, dir); err == nil {
		for _, e := range entries {
			taken[e.Name] = true
		}
	}
	if !taken[name] {
		return name
	}
	ext := path.Ext(name)
	stem := strings.TrimSuffix(name, ext)
	for i := 2; ; i++ {
		if n := fmt.Sprintf("%s-%d%s", stem, i, ext); !taken[n] {
			return n
		}
	}
}

// zipEntryRel arşiv içi adı güvenli göreceli yola çevirir: ters bölü, baştaki "/" ve ".." hedefin
// dışına çıkamaz (zip-slip). Boş dönerse girdi atlanır.
func zipEntryRel(name string) string {
	name = strings.ReplaceAll(name, `\`, "/")
	if strings.HasPrefix(name, "__MACOSX/") {
		return ""
	}
	return strings.TrimPrefix(path.Clean("/"+name), "/")
}

// remoteUnzip zipFull arşivini destDir'e çıkarır; var olan dosyaların üzerine yazar.
func remoteUnzip(ctx context.Context, c *ftpclient.Client, zipFull, destDir string, w io.Writer) (files int, err error) {
	tmp, err := os.CreateTemp("", "syncftp-unzip-*.zip")
	if err != nil {
		return 0, err
	}
	tmp.Close()
	defer os.Remove(tmp.Name())
	if err := c.Download(zipFull, tmp.Name()); err != nil {
		return 0, err
	}
	zr, err := zip.OpenReader(tmp.Name())
	if err != nil {
		return 0, fmt.Errorf(lang.L.RemoteUnzipBadFmt, path.Base(zipFull), err)
	}
	defer zr.Close()
	var total uint64
	for _, f := range zr.File {
		total += f.UncompressedSize64
	}
	if total > unzipMaxTotal {
		return 0, fmt.Errorf(lang.L.RemoteUnzipTooBigFmt, humanBytes(int64(total)))
	}
	if err := c.MakeDirAll(destDir); err != nil {
		return 0, err
	}
	for _, f := range zr.File {
		if err := ctx.Err(); err != nil {
			return files, err
		}
		rel := zipEntryRel(f.Name)
		if rel == "" {
			continue
		}
		target := path.Join(destDir, rel)
		mode := f.FileInfo().Mode()
		if f.FileInfo().IsDir() || strings.HasSuffix(f.Name, "/") {
			if err := c.MakeDirAll(target); err != nil {
				return files, err
			}
			continue
		}
		if mode&os.ModeSymlink != 0 {
			continue // bağlantılar sunucuya taşınmaz
		}
		if err := unzipOne(c, f, target); err != nil {
			return files, fmt.Errorf("%s: %w", rel, err)
		}
		files++
		fmt.Fprintf(w, lang.L.RemoteUnzipItemFmt, rel)
	}
	return files, nil
}

func unzipOne(c *ftpclient.Client, f *zip.File, target string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	local, cleanup, err := partToTemp(io.LimitReader(rc, unzipMaxTotal))
	if err != nil {
		return err
	}
	defer cleanup()
	return c.UploadTo(local, target)
}

// zipDefaultName seçilen öğelerden arşiv adı önerir: tek öğe → "ad.zip", çok öğe → "<klasör>.zip".
func zipDefaultName(baseDir string, srcs []string) string {
	if len(srcs) == 1 {
		return strings.TrimSuffix(path.Base(srcs[0]), path.Ext(path.Base(srcs[0]))) + ".zip"
	}
	if b := path.Base(baseDir); b != "/" && b != "." && b != "" {
		return b + ".zip"
	}
	return "archive.zip"
}

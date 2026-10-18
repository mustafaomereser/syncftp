package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"syncftp/internal/config"
	ftpclient "syncftp/internal/ftp"
	"syncftp/internal/lang"
	"syncftp/internal/state"
)

// Uzak dosya düzenleme: dosyanın TAMAMI okunur (önizlemedeki 256 KB sınırı yalnızca hızlı açılış
// içindi), kaydederken normal sync'in upload çekirdeği (runUploadSession) kullanılır → eski hal
// yedeklenir, release yazılır, Geri al ile dönülebilir.

const editMaxBytes = 4 << 20

type editDoc struct {
	Content     string `json:"content"`
	Hash        string `json:"hash"` // sunucudaki ham içeriğin sha256'sı — kaydederken çakışma kontrolü
	Size        int    `json:"size"`
	Editable    bool   `json:"editable"`
	Reason      string `json:"reason,omitempty"` // binary | large | encoding
	LocalExists bool   `json:"local_exists"`
	LocalSame   bool   `json:"local_same"`
	SaveTarget  string `json:"save_target,omitempty"` // panel tercihi (gui-state edit_save_target)
}

func contentHash(data []byte) string {
	h := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(h[:])
}

// editRel kullanıcıdan gelen yolu RemotePath'e göreceli, temiz state anahtarına çevirir ("" = kök).
func editRel(rel string) string {
	return strings.TrimPrefix(path.Clean("/"+strings.TrimPrefix(rel, "/")), "/")
}

// editLocalPath uzak dosyanın yereldeki karşılığı (sync'teki eşleme: local_path/rel).
func editLocalPath(dir string, cfg *config.Config, srv config.Server, rel string) (string, error) {
	root := filepath.Join(dir, srv.EffectiveLocalPath(cfg.Project))
	full := filepath.Join(root, filepath.FromSlash(rel))
	if r, err := filepath.Rel(root, full); err != nil || r == "." || strings.HasPrefix(r, "..") {
		return "", fmt.Errorf("path: %s", rel)
	}
	return full, nil
}

// readRemoteFull uzak dosyayı sınır+1 bayta kadar okur (sınırı aştıysa large).
func readRemoteFull(c *ftpclient.Client, full string) ([]byte, error) {
	return c.Preview(full, editMaxBytes+1)
}

func openEditDoc(data []byte, local []byte, localOK bool) editDoc {
	d := editDoc{Size: len(data), Hash: contentHash(data), LocalExists: localOK}
	switch {
	case len(data) > editMaxBytes:
		d.Reason = "large"
	case bytes.IndexByte(data[:min(len(data), 8000)], 0) >= 0:
		d.Reason = "binary"
	case !utf8.Valid(data):
		d.Reason = "encoding" // UTF-8 olmayan dosya tarayıcıdan geri yazılırsa bozulur
	default:
		d.Editable = true
		d.Content = strings.ReplaceAll(string(data), "\r\n", "\n")
		if localOK {
			d.LocalSame = strings.ReplaceAll(string(local), "\r\n", "\n") == d.Content
		}
	}
	return d
}

type editSaveOpts struct {
	Local bool // yerel dosyayı da yaz (state güncellenir → sonraki sync ezmez)
}

// saveRemoteEdit içeriği sunucuya (ve istenirse yerele) yazar. Yalnızca sunucu modunda state
// DEĞİŞMEZ: yerel dosya değişmedikçe sync düzenlemeyi ezmez.
func saveRemoteEdit(w io.Writer, dir string, cfg *config.Config, srv config.Server, rel, content string, o editSaveOpts) (string, error) {
	localFile, err := editLocalPath(dir, cfg, srv, rel)
	if err != nil {
		return "", err
	}
	data := []byte(content)
	src := localFile
	if o.Local {
		// Yereldeki satır sonu biçimi korunur (tarayıcı her şeyi \n'e çevirir)
		if old, err := os.ReadFile(localFile); err == nil && bytes.Contains(old, []byte("\r\n")) {
			data = []byte(strings.ReplaceAll(content, "\n", "\r\n"))
		}
		if err := os.MkdirAll(filepath.Dir(localFile), 0755); err != nil {
			return "", err
		}
		if err := os.WriteFile(localFile, data, 0644); err != nil {
			return "", err
		}
		fmt.Fprintf(w, lang.L.EditSavedLocalFmt, filepath.ToSlash(filepath.Join(srv.EffectiveLocalPath(cfg.Project), rel)))
	} else {
		tmp, err := os.CreateTemp("", "syncftp-edit-*")
		if err != nil {
			return "", err
		}
		_, werr := tmp.Write(data)
		tmp.Close()
		defer os.Remove(tmp.Name())
		if werr != nil {
			return "", werr
		}
		src = tmp.Name()
	}

	st, err := state.Load(dir, srv.Name)
	if err != nil {
		return "", err
	}
	target := st
	if !o.Local {
		cp := *st
		cp.Files = make(map[string]string, len(st.Files))
		for k, v := range st.Files {
			cp.Files[k] = v
		}
		target = &cp
	}
	task := ftpclient.UploadTask{LocalPath: src, RelPath: rel, Hash: contentHash(data)}
	out, err := runUploadSession(w, dir, cfg, srv, []ftpclient.UploadTask{task}, target, func(r ftpclient.UploadResult) { printUploadResult(w, r) })
	if err != nil {
		return "", err
	}
	for _, r := range out.results {
		if r.Err != nil {
			return "", r.Err
		}
	}
	if o.Local {
		if err := state.Save(dir, st); err != nil {
			return "", err
		}
	}
	return out.releaseID, nil
}

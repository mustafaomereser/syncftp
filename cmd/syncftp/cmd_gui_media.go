package main

import (
	"errors"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	ftpclient "syncftp/internal/ftp"
	"syncftp/internal/lang"
	"syncftp/internal/release"
)

// Panel medya önizlemesi: resim / video / ses / PDF dosyaları tarayıcıda gösterilir.
// Uzak dosya bir kez geçici klasöre indirilir (video ileri-geri sararken tarayıcı Range
// istekleri atar — her istekte yeniden indirilmesin), 5 dk kullanılmazsa silinir.
// Yanıtlar `Content-Security-Policy: sandbox` + nosniff ile gider: panelle aynı kökenden
// sunulan bir SVG / dosya script çalıştıramaz.

var mediaTypes = map[string]string{
	".png": "image", ".jpg": "image", ".jpeg": "image", ".gif": "image", ".webp": "image",
	".svg": "image", ".bmp": "image", ".ico": "image", ".avif": "image",
	".mp4": "video", ".webm": "video", ".ogv": "video", ".m4v": "video", ".mov": "video",
	".mp3": "audio", ".wav": "audio", ".ogg": "audio", ".oga": "audio", ".m4a": "audio",
	".flac": "audio", ".aac": "audio", ".opus": "audio",
	".pdf": "pdf",
}

// mediaKind dosya adından önizleme türü: image | video | audio | pdf | "".
func mediaKind(name string) string {
	return mediaTypes[strings.ToLower(path.Ext(name))]
}

func mediaContentType(name string) string {
	ext := strings.ToLower(path.Ext(name))
	switch ext { // Windows kayıt defteri bazılarını bilmiyor / yanlış biliyor
	case ".svg":
		return "image/svg+xml"
	case ".mov":
		return "video/quicktime"
	case ".m4v":
		return "video/mp4"
	case ".ogv":
		return "video/ogg"
	case ".oga", ".opus":
		return "audio/ogg"
	case ".m4a":
		return "audio/mp4"
	case ".avif":
		return "image/avif"
	}
	if t := mime.TypeByExtension(ext); t != "" {
		return t
	}
	return "application/octet-stream"
}

// serveMediaFile yerel dosyayı satır içi (Range destekli) sunar; medya değilse indirme olarak.
func serveMediaFile(w http.ResponseWriter, r *http.Request, file, name string) {
	f, err := os.Open(file)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil || st.IsDir() {
		guiErr(w, 404, errors.New("not a file"))
		return
	}
	h := w.Header()
	h.Set("X-Content-Type-Options", "nosniff")
	if mediaKind(name) != "pdf" { // Chrome'un PDF görüntüleyicisi sandbox'lı belgede açılmıyor
		h.Set("Content-Security-Policy", "sandbox")
	}
	h.Set("Cache-Control", "private, max-age=300")
	if mediaKind(name) != "" {
		h.Set("Content-Type", mediaContentType(name))
		h.Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, strings.ReplaceAll(name, `"`, "")))
	} else {
		h.Set("Content-Type", "application/octet-stream")
		h.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, strings.ReplaceAll(name, `"`, "")))
	}
	http.ServeContent(w, r, name, st.ModTime(), f)
}

type mediaEntry struct {
	once sync.Once
	file string
	err  error
	used time.Time
}

var (
	mediaMu      sync.Mutex
	mediaCache   = map[string]*mediaEntry{}
	mediaJanitor sync.Once
	mediaDir     string
)

// remoteMediaFile uzak dosyayı (aynı anahtar için bir kez) geçici klasöre indirir. Ayrı bir
// bağlantı açılır: büyük bir video inerken dosya tarayıcısının önbellekli bağlantısı bloklanmasın.
func remoteMediaFile(key string, connect func() (*ftpclient.Client, error), full string) (string, error) {
	mediaJanitor.Do(func() {
		// Zorla kapatılan panellerden kalan eski klasörler (çalışan başka panellerinkine dokunmamak için 1 gün)
		if old, _ := filepath.Glob(filepath.Join(os.TempDir(), "syncftp-media-*")); len(old) > 0 {
			for _, d := range old {
				if st, err := os.Stat(d); err == nil && time.Since(st.ModTime()) > 24*time.Hour {
					os.RemoveAll(d)
				}
			}
		}
		mediaDir, _ = os.MkdirTemp("", "syncftp-media-")
		go func() {
			for range time.Tick(time.Minute) {
				mediaMu.Lock()
				for k, e := range mediaCache {
					if time.Since(e.used) > 5*time.Minute {
						delete(mediaCache, k)
						if e.file != "" {
							os.Remove(e.file)
						}
					}
				}
				mediaMu.Unlock()
			}
		}()
	})
	mediaMu.Lock()
	e, ok := mediaCache[key]
	if !ok {
		e = &mediaEntry{}
		mediaCache[key] = e
	}
	e.used = time.Now()
	mediaMu.Unlock()

	e.once.Do(func() {
		tmp, err := os.CreateTemp(mediaDir, "m-*"+strings.ToLower(path.Ext(full)))
		if err != nil {
			e.err = err
			return
		}
		tmp.Close()
		c, err := connect()
		if err == nil {
			err = c.Download(full, tmp.Name())
			c.Close()
		}
		if err != nil {
			os.Remove(tmp.Name())
			e.err = err
			return
		}
		e.file = tmp.Name()
	})
	if e.err != nil { // başarısız indirme önbellekte kalmasın — tekrar denenebilsin
		mediaMu.Lock()
		if mediaCache[key] == e {
			delete(mediaCache, key)
		}
		mediaMu.Unlock()
	}
	return e.file, e.err
}

// clearMediaCache panel kapanırken geçici medya klasörünü siler.
func clearMediaCache() {
	mediaMu.Lock()
	defer mediaMu.Unlock()
	mediaCache = map[string]*mediaEntry{}
	if mediaDir != "" {
		os.RemoveAll(mediaDir)
		mediaDir = ""
	}
}

// GET /api/p/{p}/remote/media?server=&path=&v= — v önizleme penceresi başına değişir (taze indirme).
func (sv *guiServer) apiRemoteMedia(w http.ResponseWriter, r *http.Request) {
	p, _, srv, ok := sv.remoteServerCfg(w, r)
	if !ok {
		return
	}
	full := guiRemotePath(srv, r.URL.Query().Get("path"))
	key := remoteKey(p.Name, srv) + "|" + full + "|" + r.URL.Query().Get("v")
	file, err := remoteMediaFile(key, func() (*ftpclient.Client, error) { return ftpclient.Connect(srv) }, full)
	if err != nil {
		guiErr(w, 502, err)
		return
	}
	serveMediaFile(w, r, file, path.Base(full))
}

// GET /api/p/{p}/local/media?server=&path= — sunucunun yerel dizinindeki dosya.
func (sv *guiServer) apiLocalMedia(w http.ResponseWriter, r *http.Request) {
	p, cfg, srv, ok := sv.remoteServerCfg(w, r)
	if !ok {
		return
	}
	lp, err := editLocalPath(p.Path, cfg, srv, editRel(r.URL.Query().Get("path")))
	if err != nil {
		guiErr(w, 400, err)
		return
	}
	serveMediaFile(w, r, lp, filepath.Base(lp))
}

// GET /api/p/{p}/releases/{srv}/{id}/media?path=&side=before|after — release'in yedeği (eski
// hal) ya da yüklenen kopyası (yeni hal; yoksa hash'i tutan yerel dosya — releaseFileVersions gibi).
func (sv *guiServer) apiReleaseMedia(w http.ResponseWriter, r *http.Request) {
	p, cfg, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	srv, ok := findServer(cfg, r.PathValue("srv"))
	if !ok {
		guiErr(w, 404, fmt.Errorf(lang.L.ServerNotFoundFmt, r.PathValue("srv")))
		return
	}
	m, err := release.Load(p.Path, srv.Name, filepath.Base(r.PathValue("id")))
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	rel := editRel(r.URL.Query().Get("path"))
	hash, ok := m.Files[rel]
	if !ok || rel == "" {
		guiErr(w, 404, fmt.Errorf(lang.L.ReleaseFileNotFoundFmt, rel))
		return
	}
	name := path.Base(rel)
	if r.URL.Query().Get("side") == "before" {
		serveMediaFile(w, r, filepath.Join(release.FilesDir(p.Path, srv.Name, m.ID), filepath.FromSlash(rel)), name)
		return
	}
	after := filepath.Join(release.AfterDir(p.Path, srv.Name, m.ID), filepath.FromSlash(rel))
	if _, err := os.Stat(after); err == nil {
		serveMediaFile(w, r, after, name)
		return
	}
	lp, err := editLocalPath(p.Path, cfg, srv, rel)
	if err == nil {
		if data, rerr := os.ReadFile(lp); rerr == nil && contentHash(data) == hash {
			serveMediaFile(w, r, lp, name)
			return
		}
	}
	guiErr(w, 404, fmt.Errorf(lang.L.ReleaseFileNotFoundFmt, rel))
}

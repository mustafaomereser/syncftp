package main

import (
	"archive/zip"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"syncftp/internal/lang"
)

// Kendini güncelleme (CLI `syncftp update` + panel). Sürüm GitHub Releases'tan okunur. Kurulum yalnızca
// checksums.txt'nin imzası gömülü açık anahtarla doğrulanırsa ve dosyanın SHA256'sı tutarsa yapılır:
// imzayı atan anahtar GitHub'daki dosyalardan ayrı olduğu için hesap ele geçirilse bile sahte sürüm kurulmaz.
// Doğrulama ve açma biter bitmeden kurulu dosyaya dokunulmaz; değiştirme başarısızsa eskisi geri konur.

// updateSigningPublicKey release'leri imzalayan ed25519 anahtarının açık yarısı (base64).
const updateSigningPublicKey = "nPhNfUjHfaDBO2rvaSLS6smqgktTfdzp0TGep0IQmxQ="

const (
	updateMaxDownload = 256 << 20
	updateOldSuffix   = ".update-old"
)

// updateAPI test için değiştirilebilir (imza doğrulaması her durumda yapılır).
var updateAPI = "https://api.github.com/repos/mustafaomereser/syncftp/releases/latest"

var updateHTTP = &http.Client{Timeout: 10 * time.Minute}

type ghAsset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

type ghRelease struct {
	Tag    string    `json:"tag_name"`
	Body   string    `json:"body"`
	URL    string    `json:"html_url"`
	Assets []ghAsset `json:"assets"`
}

func (r *ghRelease) version() string { return strings.TrimPrefix(r.Tag, "v") }

func (r *ghRelease) asset(name string) (ghAsset, bool) {
	for _, a := range r.Assets {
		if a.Name == name {
			return a, true
		}
	}
	return ghAsset{}, false
}

func fetchLatestRelease() (*ghRelease, error) {
	req, err := http.NewRequest("GET", updateAPI, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "syncFTP/"+appVersion)
	resp, err := updateHTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub: HTTP %d", resp.StatusCode)
	}
	var r ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	return &r, nil
}

// versionNewer a, b'den yeni mi (noktalı sayılar; 2.0.10 > 2.0.9).
func versionNewer(a, b string) bool {
	x, y := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(x) || i < len(y); i++ {
		var p, q int
		if i < len(x) {
			p, _ = strconv.Atoi(x[i])
		}
		if i < len(y) {
			q, _ = strconv.Atoi(y[i])
		}
		if p != q {
			return p > q
		}
	}
	return false
}

// updateAssetName bu platformun release dosyası; paket yoksa "".
func updateAssetName(goos, goarch, version string) string {
	suffix := ""
	switch {
	case goos == "windows" && goarch == "amd64":
		suffix = "windows-amd64.exe"
	case goos == "darwin" && goarch == "arm64":
		suffix = "macos-apple-silicon.zip"
	case goos == "darwin" && goarch == "amd64":
		suffix = "macos-intel.zip"
	case goos == "linux" && (goarch == "amd64" || goarch == "arm64"):
		suffix = "linux-" + goarch + ".zip"
	default:
		return ""
	}
	return "syncFTP-" + version + "-" + suffix
}

func downloadTo(url, dest string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "syncFTP/"+appVersion)
	resp, err := updateHTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, path.Base(url))
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	n, err := io.Copy(f, io.LimitReader(resp.Body, updateMaxDownload+1))
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil && n > updateMaxDownload {
		err = errors.New(lang.L.UpdateBadSignature)
	}
	return err
}

// verifyManifest imzayı doğrular ve checksums.txt'teki ad → sha256 eşlemesini döndürür.
func verifyManifest(pub string, manifest, sigB64 []byte) (map[string]string, error) {
	key, err := base64.StdEncoding.DecodeString(pub)
	if err != nil || len(key) != ed25519.PublicKeySize {
		return nil, errors.New(lang.L.UpdateBadSignature)
	}
	sig, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(sigB64)))
	if err != nil || !ed25519.Verify(key, manifest, sig) {
		return nil, errors.New(lang.L.UpdateBadSignature)
	}
	out := map[string]string{}
	for _, line := range strings.Split(string(manifest), "\n") {
		f := strings.Fields(line)
		if len(f) == 2 {
			out[strings.TrimPrefix(f[1], "*")] = strings.ToLower(f[0])
		}
	}
	return out, nil
}

func fileSHA256(p string) (string, error) {
	f, err := os.Open(p)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// downloadVerified paketi tmp içine indirir, imza + özet tutmazsa hata döner.
func downloadVerified(r *ghRelease, name, tmp string) (string, error) {
	a, ok := r.asset(name)
	m, ok2 := r.asset("checksums.txt")
	s, ok3 := r.asset("checksums.txt.sig")
	if !ok || !ok2 || !ok3 {
		return "", errors.New(lang.L.UpdateUnsupported)
	}
	mp, sp, ap := filepath.Join(tmp, "checksums.txt"), filepath.Join(tmp, "checksums.txt.sig"), filepath.Join(tmp, name)
	for _, d := range [][2]string{{m.URL, mp}, {s.URL, sp}, {a.URL, ap}} {
		if err := downloadTo(d[0], d[1]); err != nil {
			return "", err
		}
	}
	manifest, err := os.ReadFile(mp)
	if err != nil {
		return "", err
	}
	sig, err := os.ReadFile(sp)
	if err != nil {
		return "", err
	}
	sums, err := verifyManifest(updateSigningPublicKey, manifest, sig)
	if err != nil {
		return "", err
	}
	got, err := fileSHA256(ap)
	if err != nil {
		return "", err
	}
	if want, ok := sums[name]; !ok || want != got {
		return "", errors.New(lang.L.UpdateBadSignature)
	}
	return ap, nil
}

// extractZip yalnızca keep'in true döndürdüğü girdileri dest altına açar (zip-slip korumalı, izinler korunur).
func extractZip(src, dest string, keep func(string) bool) error {
	z, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer z.Close()
	for _, e := range z.File {
		name := path.Clean("/" + e.Name)[1:]
		if name == "" || !keep(name) {
			continue
		}
		out := filepath.Join(dest, filepath.FromSlash(name))
		if e.FileInfo().IsDir() {
			if err := os.MkdirAll(out, 0755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
			return err
		}
		mode := os.FileMode(0644)
		if e.Mode()&0111 != 0 {
			mode = 0755
		}
		rc, err := e.Open()
		if err != nil {
			return err
		}
		f, err := os.OpenFile(out, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
		if err != nil {
			rc.Close()
			return err
		}
		_, err = io.Copy(f, rc)
		rc.Close()
		if cerr := f.Close(); err == nil {
			err = cerr
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// swapPath new'i target'ın yerine koyar; eskisi updateOldSuffix adıyla kalır (Windows çalışan exe'yi silemez,
// yeniden adlandırabilir), sonraki açılışta temizlenir. Yerleştirme başarısızsa eskisi geri konur.
func swapPath(newPath, target string) error {
	old := target + updateOldSuffix
	_ = os.RemoveAll(old)
	if err := os.Rename(target, old); err != nil {
		return err
	}
	if err := os.Rename(newPath, target); err != nil {
		_ = os.Rename(old, target)
		return err
	}
	return nil
}

// macBundleOf exe bir .app içindeyse paket kök yolunu döndürür.
func macBundleOf(exe string) string {
	const mark = ".app/Contents/MacOS/"
	if i := strings.Index(filepath.ToSlash(exe), mark); i >= 0 {
		return exe[:i+len(".app")]
	}
	return ""
}

// installPackage indirilmiş ve doğrulanmış paketi exe'nin yerine kurar.
func installPackage(goos, pkg, exe string) error {
	dir := filepath.Dir(exe)
	switch {
	case goos == "windows":
		staged := exe + ".update-new"
		defer os.Remove(staged)
		data, err := os.ReadFile(pkg)
		if err != nil {
			return err
		}
		if err := os.WriteFile(staged, data, 0755); err != nil {
			return err
		}
		return swapPath(staged, exe)
	case goos == "darwin" && macBundleOf(exe) != "":
		// .app bütün olarak değişir: imza ve Apple onayı pakete bağlı
		app := macBundleOf(exe)
		staging := app + ".update-new"
		_ = os.RemoveAll(staging)
		defer os.RemoveAll(staging)
		if err := extractZip(pkg, staging, func(n string) bool { return n == "syncFTP.app" || strings.HasPrefix(n, "syncFTP.app/") }); err != nil {
			return err
		}
		fresh := filepath.Join(staging, "syncFTP.app")
		if _, err := os.Stat(filepath.Join(fresh, "Contents", "MacOS", "syncftp")); err != nil {
			return err
		}
		if err := swapPath(fresh, app); err != nil {
			return err
		}
		_ = os.RemoveAll(app + updateOldSuffix) // macOS çalışan dosyaların klasörünü silmeye izin verir
		return nil
	default:
		want := "syncftp"
		if goos == "darwin" {
			want = "syncFTP.app/Contents/MacOS/syncftp"
		}
		staging, err := os.MkdirTemp(dir, ".syncftp-update-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(staging)
		if err := extractZip(pkg, staging, func(n string) bool { return n == want }); err != nil {
			return err
		}
		fresh := filepath.Join(staging, filepath.FromSlash(want))
		if _, err := os.Stat(fresh); err != nil {
			return err
		}
		if err := swapPath(fresh, exe); err != nil {
			return err
		}
		_ = os.Remove(exe + updateOldSuffix)
		return nil
	}
}

// applyUpdate release'i indirir, doğrular, kurar. Dönüşte kurulu program yenisidir, çalışan süreç eskisi.
func applyUpdate(r *ghRelease) error {
	name := updateAssetName(runtime.GOOS, runtime.GOARCH, r.version())
	if name == "" {
		return errors.New(lang.L.UpdateUnsupported)
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	if p, err := filepath.EvalSymlinks(exe); err == nil {
		exe = p
	}
	tmp, err := os.MkdirTemp("", "syncftp-update-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	pkg, err := downloadVerified(r, name, tmp)
	if err != nil {
		return err
	}
	return installPackage(runtime.GOOS, pkg, exe)
}

// cleanupUpdateLeftovers önceki güncellemeden kalanı siler. Eski panel penceresi hâlâ kapanıyor olabilir → birkaç deneme.
func cleanupUpdateLeftovers() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	targets := []string{exe + updateOldSuffix}
	if app := macBundleOf(exe); app != "" {
		targets = append(targets, app+updateOldSuffix)
	}
	for i := 0; i < 20; i++ {
		left := false
		for _, t := range targets {
			if _, err := os.Stat(t); err == nil {
				if os.RemoveAll(t) != nil {
					left = true
				}
			}
		}
		if !left {
			return
		}
		time.Sleep(3 * time.Second)
	}
}

var updateCheckOnly bool

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "syncFTP'yi son sürüme güncelle",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := fetchLatestRelease()
		if err != nil {
			return err
		}
		if !versionNewer(r.version(), appVersion) {
			fmt.Printf(lang.L.UpdateCurrentFmt, appVersion)
			return nil
		}
		fmt.Printf(lang.L.UpdateAvailableFmt, appVersion, r.version())
		if updateCheckOnly {
			return nil
		}
		if err := applyUpdate(r); err != nil {
			return err
		}
		fmt.Printf(lang.L.UpdateDoneFmt, r.version())
		return nil
	},
}

func init() {
	updateCmd.Flags().BoolVar(&updateCheckOnly, "check", false, "Yalnızca yeni sürüm var mı bak")
	rootCmd.AddCommand(updateCmd)
}

// ── Panel ─────────────────────────────────────────────────────────────────────

type updateState struct {
	mu       sync.Mutex
	release  *ghRelease
	checked  time.Time
	checkErr string
	busy     bool
}

var guiUpdate updateState

func (sv *guiServer) registerUpdateAPI(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/update", func(w http.ResponseWriter, r *http.Request) { guiJSON(w, guiUpdate.view()) })
	mux.HandleFunc("POST /api/update/check", func(w http.ResponseWriter, r *http.Request) {
		guiUpdate.check()
		guiJSON(w, guiUpdate.view())
	})
	mux.HandleFunc("PUT /api/update/settings", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Auto bool `json:"auto"`
		}
		if err := guiDecode(r, &req); err != nil {
			guiErr(w, 400, err)
			return
		}
		updateGUIPersist(func(st *guiPersist) { st.UpdateCheckOff = !req.Auto })
		guiJSON(w, guiUpdate.view())
	})
	mux.HandleFunc("POST /api/update/apply", sv.apiUpdateApply)
}

func (u *updateState) check() {
	r, err := fetchLatestRelease()
	u.mu.Lock()
	defer u.mu.Unlock()
	u.checked = time.Now()
	if err != nil {
		u.checkErr = err.Error()
		return
	}
	u.checkErr, u.release = "", r
}

func (u *updateState) view() map[string]any {
	u.mu.Lock()
	defer u.mu.Unlock()
	out := map[string]any{"current": appVersion, "available": false, "auto": !loadGUIPersist().UpdateCheckOff, "error": u.checkErr}
	if !u.checked.IsZero() {
		out["checked"] = u.checked
	}
	if r := u.release; r != nil {
		out["latest"] = r.version()
		out["notes"] = r.Body
		out["url"] = r.URL
		out["available"] = versionNewer(r.version(), appVersion)
		_, signed := r.asset("checksums.txt.sig")
		name := updateAssetName(runtime.GOOS, runtime.GOARCH, r.version())
		_, pkg := r.asset(name)
		out["installable"] = signed && pkg
	}
	return out
}

// updateLoop açılışta ve günde bir yeni sürüme bakar (Ayarlar'dan kapatılabilir).
func (sv *guiServer) updateLoop() {
	delay := 5 * time.Second
	for {
		select {
		case <-sv.quit:
			return
		case <-time.After(delay):
		}
		delay = 24 * time.Hour
		if !loadGUIPersist().UpdateCheckOff {
			guiUpdate.check()
		}
	}
}

// otherJobsRunning watch dışında çalışan iş sayısı — watch'lar yeniden başlayınca geri gelir, diğerleri kesilirdi.
func (sv *guiServer) otherJobsRunning() int {
	n := 0
	for _, j := range sv.jobs.list("") {
		if j["status"] == "running" && j["kind"] != "watch" {
			n++
		}
	}
	return n
}

func (sv *guiServer) apiUpdateApply(w http.ResponseWriter, r *http.Request) {
	u := &guiUpdate
	u.mu.Lock()
	rel, busy := u.release, u.busy
	if rel == nil || busy || !versionNewer(rel.version(), appVersion) {
		u.mu.Unlock()
		guiErr(w, 409, errors.New(lang.L.UpdateUnsupported))
		return
	}
	u.busy = true
	u.mu.Unlock()
	defer func() { u.mu.Lock(); u.busy = false; u.mu.Unlock() }()

	if sv.otherJobsRunning() > 0 {
		guiErr(w, 409, errors.New(lang.L.UpdateJobsRunning))
		return
	}
	if err := applyUpdate(rel); err != nil {
		guiErr(w, 502, err)
		return
	}
	// Yeni süreç bu panelin kapanmasını bekleyip kendi penceresini açar (runGUIForeground)
	args := []string{"gui", "--foreground"}
	if guiFlagNoWindow {
		args = append(args, "--no-window")
	}
	os.Setenv(updateRestartEnv, "1")
	if err := spawnDetached(args...); err != nil {
		guiErr(w, 500, err)
		return
	}
	guiJSON(w, map[string]any{"ok": true, "version": rel.version()})
	go func() {
		time.Sleep(700 * time.Millisecond)
		sv.shutdown()
	}()
}

const updateRestartEnv = "SYNCFTP_UPDATE_RESTART"

// waitForOldPanel güncelleme sonrası başlayan süreçte eski panelin kapanmasını bekler.
func waitForOldPanel() {
	if os.Getenv(updateRestartEnv) == "" {
		return
	}
	os.Unsetenv(updateRestartEnv)
	for i := 0; i < 100; i++ {
		if _, ok := readGUIInstance(); !ok {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
}

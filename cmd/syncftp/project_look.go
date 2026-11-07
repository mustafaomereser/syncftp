package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"syncftp/internal/config"
	"syncftp/internal/fsutil"
	"syncftp/internal/lang"
	"syncftp/internal/projects"
)

// Proje görünümü: renk + grup (config project.color / project.group) ve kapak resmi
// (.syncftp/banner.<uzantı>). Panelin "Görünüm" penceresi ve CLI `project look` aynı çekirdek.

var colorRe = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// bannerMax kapak resmi üst sınırı.
const bannerMax = 8 << 20

var bannerExts = map[string]string{"image/png": ".png", "image/jpeg": ".jpg", "image/webp": ".webp", "image/gif": ".gif"}

// projectLook nil alanlar değişmez; "" temizler.
type projectLook struct {
	Color *string `json:"color"`
	Group *string `json:"group"`
}

func setProjectLook(dir string, look projectLook) error {
	if look.Color == nil && look.Group == nil {
		return nil
	}
	cfg, err := config.Load(dir)
	if err != nil {
		return err
	}
	if look.Color != nil {
		c := strings.TrimSpace(*look.Color)
		if c != "" && !colorRe.MatchString(c) {
			return fmt.Errorf(lang.L.LookColorFmt, c)
		}
		cfg.Project.Color = strings.ToLower(c)
	}
	if look.Group != nil {
		g := strings.Join(strings.Fields(*look.Group), " ")
		if len([]rune(g)) > 40 {
			g = string([]rune(g)[:40])
		}
		cfg.Project.Group = g
	}
	if err := config.Save(dir, cfg); err != nil {
		return err
	}
	projects.Invalidate()
	return nil
}

// projectBanner varsa kapak resminin yolu ve değişme zamanı.
func projectBanner(dir string) (string, time.Time) {
	for _, ext := range []string{".png", ".jpg", ".webp", ".gif"} {
		p := filepath.Join(dir, ".syncftp", "banner"+ext)
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p, st.ModTime()
		}
	}
	return "", time.Time{}
}

// setProjectBanner resmi doğrular (içerikten tür: png/jpeg/webp/gif, ≤ 8 MB) ve eskisinin yerine yazar.
func setProjectBanner(dir string, src io.Reader) error {
	data, err := io.ReadAll(io.LimitReader(src, bannerMax+1))
	if err != nil {
		return err
	}
	if len(data) > bannerMax {
		return errors.New(lang.L.LookBannerTooBig)
	}
	ext, ok := bannerExts[http.DetectContentType(data)]
	if !ok {
		return errors.New(lang.L.LookBannerType)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".syncftp"), 0o755); err != nil {
		return err
	}
	if err := fsutil.WriteFileAtomic(filepath.Join(dir, ".syncftp", "banner"+ext), data, 0o644); err != nil {
		return err
	}
	for _, e := range []string{".png", ".jpg", ".webp", ".gif"} {
		if e != ext {
			_ = os.Remove(filepath.Join(dir, ".syncftp", "banner"+e))
		}
	}
	return nil
}

func removeProjectBanner(dir string) error {
	for _, e := range []string{".png", ".jpg", ".webp", ".gif"} {
		if err := os.Remove(filepath.Join(dir, ".syncftp", "banner"+e)); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

// ── Panel uçları ──────────────────────────────────────────────────────────────

// PATCH /api/p/{p}/look {color?, group?}
func (sv *guiServer) apiProjectLook(w http.ResponseWriter, r *http.Request) {
	p, _, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	var look projectLook
	if err := guiDecode(r, &look); err != nil {
		guiErr(w, 400, err)
		return
	}
	if err := setProjectLook(p.Path, look); err != nil {
		guiErr(w, 400, err)
		return
	}
	guiJSON(w, map[string]bool{"ok": true})
}

// GET / PUT / DELETE /api/p/{p}/banner — PUT gövdesi resmin kendisi
func (sv *guiServer) apiProjectBanner(w http.ResponseWriter, r *http.Request) {
	p, _, err := sv.project(r)
	if err != nil {
		guiErr(w, 404, err)
		return
	}
	switch r.Method {
	case http.MethodGet:
		path, mod := projectBanner(p.Path)
		if path == "" {
			http.NotFound(w, r)
			return
		}
		f, err := os.Open(path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer f.Close()
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Security-Policy", "sandbox")
		w.Header().Set("Cache-Control", "private, max-age=31536000") // adres ?v=<mtime> taşır
		http.ServeContent(w, r, filepath.Base(path), mod, f)
	case http.MethodPut:
		defer r.Body.Close()
		if err := setProjectBanner(p.Path, http.MaxBytesReader(w, r.Body, bannerMax+1)); err != nil {
			guiErr(w, 400, err)
			return
		}
		projects.Invalidate()
		guiJSON(w, map[string]bool{"ok": true})
	case http.MethodDelete:
		if err := removeProjectBanner(p.Path); err != nil {
			guiErr(w, 500, err)
			return
		}
		guiJSON(w, map[string]bool{"ok": true})
	}
}

// ── CLI ───────────────────────────────────────────────────────────────────────

var (
	lookColor, lookGroup, lookBanner string
	lookNoBanner                     bool
)

func init() {
	projectLookCmd.Flags().StringVar(&lookColor, "color", "", `Renk (#3b9eff); "none" temizler`)
	projectLookCmd.Flags().StringVar(&lookGroup, "group", "", `Grup adı; "none" temizler`)
	projectLookCmd.Flags().StringVar(&lookBanner, "banner", "", "Kapak resmi (png / jpg / webp / gif, ≤ 8 MB)")
	projectLookCmd.Flags().BoolVar(&lookNoBanner, "no-banner", false, "Kapak resmini kaldır")
	projectCmd.AddCommand(projectLookCmd)
}

var projectLookCmd = &cobra.Command{
	Use:   "look [ad|yol]",
	Short: "Projenin paneldeki görünümü: renk, grup, kapak resmi",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := os.Getwd()
		if len(args) == 1 {
			if p, ok := projects.Find(args[0]); ok {
				dir = p.Path
			} else if st, err := os.Stat(args[0]); err == nil && st.IsDir() {
				dir, _ = filepath.Abs(args[0])
			} else {
				return fmt.Errorf(lang.L.ProjectNotFoundFmt, args[0])
			}
		}
		var look projectLook
		val := func(s string) *string {
			if strings.EqualFold(s, "none") {
				s = ""
			}
			return &s
		}
		if cmd.Flags().Changed("color") {
			look.Color = val(lookColor)
		}
		if cmd.Flags().Changed("group") {
			look.Group = val(lookGroup)
		}
		if err := setProjectLook(dir, look); err != nil {
			return err
		}
		if lookBanner != "" {
			f, err := os.Open(lookBanner)
			if err != nil {
				return err
			}
			err = setProjectBanner(dir, f)
			f.Close()
			if err != nil {
				return err
			}
		} else if lookNoBanner {
			if err := removeProjectBanner(dir); err != nil {
				return err
			}
		}
		cfg, err := config.Load(dir)
		if err != nil {
			return err
		}
		banner, _ := projectBanner(dir)
		fmt.Printf(lang.L.LookShowFmt, cfg.Project.Name, orDash(cfg.Project.Color), orDash(cfg.Project.Group), orDash(banner))
		return nil
	},
}

func orDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

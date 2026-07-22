package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"syncftp/internal/config"
	"syncftp/internal/ignore"
	"syncftp/internal/lang"
	"syncftp/internal/milestone"
	"syncftp/internal/scanner"
	"syncftp/internal/synclog"
)

var (
	msFlagServer string
	msFlagAll    bool
	msFlagDate   string
	msFlagDry    bool
)

func init() {
	milestoneCmd.PersistentFlags().StringVarP(&msFlagServer, "server", "s", "", "Sadece bu sunucu")
	milestoneCmd.PersistentFlags().BoolVar(&msFlagAll, "all", false, "Tüm aktif sunucular (seçici açılmaz)")
	milestoneSetCmd.Flags().StringVar(&msFlagDate, "date", "", "Özel tarih: bugün, dün, 3d, 5h, 14:30, 20.07[.2026], 2026-07-20 [15:04]; verilmezse şimdi")
	milestoneSyncCmd.Flags().BoolVar(&msFlagDry, "dry-run", false, "Ne yükleneceğini göster, gerçekten yükleme")
	milestoneCmd.AddCommand(milestoneSetCmd, milestoneShowCmd, milestoneClearCmd, milestoneSyncCmd)
	rootCmd.AddCommand(milestoneCmd)
}

var milestoneCmd = &cobra.Command{
	Use:   "milestone",
	Short: "Sunucu başına zaman işareti: 'milestone sync' o tarihten sonra değişen dosyaları yükler",
	RunE:  runMilestoneShow,
}

var milestoneSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Milestone koy (varsayılan: şimdi; --date ile özel tarih)",
	RunE:  runMilestoneSet,
}

var milestoneShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Kayıtlı milestone'ları göster",
	RunE:  runMilestoneShow,
}

var milestoneClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Milestone'u sil",
	RunE:  runMilestoneClear,
}

var milestoneSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Milestone tarihinden sonra değişen (mtime) dosyaları yükle",
	RunE:  runMilestoneSync,
}

var msRelRe = regexp.MustCompile(`^(\d+)\s*([mhdw])$`)
var msDayMonthRe = regexp.MustCompile(`^(\d{1,2})\.(\d{1,2})$`)

// parseMilestoneDate esnek tarih çözümü (yerel saat dilimi, dil ayarından bağımsız):
//   şimdi / now                            → şu an
//   bugün / dün / today / yesterday        → o günün 00:00'ı
//   30m / 5h / 3d / 2w                     → o kadar süre öncesi
//   14:30                                  → bugün o saat
//   20.07 / 20.07.2026 [14:30[:05]]        → gün.ay[.yıl]
//   2026-07-20 [14:30[:05]]                → ISO
func parseMilestoneDate(s string) (time.Time, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	now := time.Now()

	switch s {
	case "now", "şimdi", "simdi":
		return now, nil
	case "today", "bugün", "bugun":
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local), nil
	case "yesterday", "dün", "dun":
		y := now.AddDate(0, 0, -1)
		return time.Date(y.Year(), y.Month(), y.Day(), 0, 0, 0, 0, time.Local), nil
	}

	if m := msRelRe.FindStringSubmatch(s); m != nil {
		n, _ := strconv.Atoi(m[1])
		switch m[2] {
		case "m":
			return now.Add(-time.Duration(n) * time.Minute), nil
		case "h":
			return now.Add(-time.Duration(n) * time.Hour), nil
		case "d":
			return now.AddDate(0, 0, -n), nil
		case "w":
			return now.AddDate(0, 0, -7*n), nil
		}
	}

	if t, err := time.ParseInLocation("15:04", s, time.Local); err == nil {
		return time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, time.Local), nil
	}

	if m := msDayMonthRe.FindStringSubmatch(s); m != nil {
		day, _ := strconv.Atoi(m[1])
		month, _ := strconv.Atoi(m[2])
		if month >= 1 && month <= 12 && day >= 1 && day <= 31 {
			return time.Date(now.Year(), time.Month(month), day, 0, 0, 0, 0, time.Local), nil
		}
	}

	for _, layout := range []string{
		"2006-01-02 15:04:05", "2006-01-02 15:04", "2006-01-02",
		"02.01.2006 15:04:05", "02.01.2006 15:04", "02.01.2006",
	} {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date: %q", s)
}

// milestoneServers --server / --all / picker ile hedef sunucuları seçer.
// (nil, nil) = kullanıcı iptal etti.
func milestoneServers(cfg *config.Config) ([]config.Server, error) {
	servers := cfg.EnabledServers()
	if msFlagServer != "" {
		for _, s := range servers {
			if s.Name == msFlagServer {
				return []config.Server{s}, nil
			}
		}
		return nil, fmt.Errorf("sunucu bulunamadı: %q (mevcut: %v)", msFlagServer, serverNames(servers))
	}
	if msFlagAll || len(servers) <= 1 {
		return servers, nil
	}
	return pickServerMultiTUI(servers, "")
}

func runMilestoneSet(cmd *cobra.Command, args []string) error {
	dir, _ := os.Getwd()
	cfg, err := config.Load(dir)
	if err != nil {
		return err
	}

	date := time.Now()
	if msFlagDate != "" {
		d, derr := parseMilestoneDate(msFlagDate)
		if derr != nil {
			fmt.Printf(lang.L.MilestoneDateErrFmt, msFlagDate)
			return nil
		}
		date = d
	}

	servers, err := milestoneServers(cfg)
	if err != nil {
		return err
	}
	if servers == nil {
		fmt.Println(lang.L.SyncCancelled)
		return nil
	}
	for _, srv := range servers {
		if err := milestone.Save(dir, srv.Name, date); err != nil {
			return err
		}
		fmt.Printf(lang.L.MilestoneSetFmt, srv.Name, date.Format("2006-01-02 15:04:05"))
	}
	return nil
}

func runMilestoneShow(cmd *cobra.Command, args []string) error {
	dir, _ := os.Getwd()
	cfg, err := config.Load(dir)
	if err != nil {
		return err
	}
	for _, srv := range cfg.Servers {
		if msFlagServer != "" && srv.Name != msFlagServer {
			continue
		}
		printMilestone(dir, srv.Name)
	}
	return nil
}

func runMilestoneClear(cmd *cobra.Command, args []string) error {
	dir, _ := os.Getwd()
	cfg, err := config.Load(dir)
	if err != nil {
		return err
	}
	servers, err := milestoneServers(cfg)
	if err != nil {
		return err
	}
	if servers == nil {
		fmt.Println(lang.L.SyncCancelled)
		return nil
	}
	for _, srv := range servers {
		if err := milestone.Clear(dir, srv.Name); err != nil {
			return err
		}
		fmt.Printf(lang.L.MilestoneClearedFmt, srv.Name)
	}
	return nil
}

func runMilestoneSync(cmd *cobra.Command, args []string) error {
	dir, _ := os.Getwd()
	cfg, err := config.Load(dir)
	if err != nil {
		return err
	}
	servers, err := milestoneServers(cfg)
	if err != nil {
		return err
	}
	if servers == nil {
		fmt.Println(lang.L.SyncCancelled)
		return nil
	}

	// syncToServerResult dry-run kararını paket seviyesindeki flagDryRun'dan okur
	flagDryRun = msFlagDry
	if msFlagDry {
		fmt.Println(lang.L.SyncDryRunNote)
	}

	for _, srv := range servers {
		ms, merr := milestone.Load(dir, srv.Name)
		if merr != nil || ms == nil {
			fmt.Printf(lang.L.MilestoneNotSetFmt, srv.Name)
			continue
		}

		localPath := srv.EffectiveLocalPath(cfg.Project)
		localDir := filepath.Join(dir, localPath)
		matcher, ierr := ignore.Load(localDir, cfg.Sync.IgnoreFiles)
		if ierr != nil {
			fmt.Printf("  ! %v\n", ierr)
			continue
		}
		files, serr := scanner.Scan(localDir, matcher)
		if serr != nil {
			fmt.Printf("  ! %v\n", serr)
			continue
		}
		current := make(map[string]string, len(files))
		byKey := make(map[string]scanner.File, len(files))
		for _, f := range files {
			current[f.RelPath] = f.Hash
			byKey[f.RelPath] = f
		}

		var buf strings.Builder
		buf.WriteString(fmt.Sprintf("══ %s (%s) ══\n", srv.Name, srv.Host))
		since := ms.Date
		result := syncToServerResult(&buf, dir, cfg, srv, current, byKey, nil, nil, &since)
		fmt.Print(buf.String())
		if !msFlagDry && (result.uploaded > 0 || result.failed > 0) {
			if p, lerr := synclog.Save(dir, srv.Name, buf.String()); lerr == nil {
				fmt.Printf(lang.L.SyncLogSavedFmt, p)
			}
		}
	}
	return nil
}

// milestoneFilterMtime milestone tarihinden önce değişmiş (mtime < since) dosyaları
// listeden çıkarır. status ve normal sync, milestone'u olan sunucularda bunu uygular.
func milestoneFilterMtime(paths []string, localDir string, since time.Time) (kept []string, dropped int) {
	for _, rel := range paths {
		info, err := os.Stat(filepath.Join(localDir, filepath.FromSlash(rel)))
		if err == nil && !info.ModTime().Before(since) {
			kept = append(kept, rel)
		} else {
			dropped++
		}
	}
	return kept, dropped
}

func printMilestone(configDir, serverName string) {
	ms, _ := milestone.Load(configDir, serverName)
	if ms == nil {
		fmt.Printf(lang.L.MilestoneNoneFmt, serverName)
		return
	}
	fmt.Printf(lang.L.MilestoneShowFmt, serverName, ms.Date.Format("2006-01-02 15:04:05"))
}

// ── shell komutu ──────────────────────────────────────────────────────────────

// cmdMilestone bağlı sunucu üzerinde çalışır: set/sync/clear bağlı sunucuya işler.
func (sh *shellState) cmdMilestone(args []string) {
	if len(args) == 0 || strings.ToLower(args[0]) == "show" {
		if sh.srv != nil {
			printMilestone(sh.configDir, sh.srv.Name)
			return
		}
		for _, s := range sh.cfg.Servers {
			printMilestone(sh.configDir, s.Name)
		}
		return
	}

	sub := strings.ToLower(args[0])
	rest := args[1:]

	if sh.srv == nil {
		fmt.Print(lang.L.MilestoneShellNoSrv)
		return
	}

	switch sub {
	case "set":
		date := time.Now()
		if len(rest) > 0 {
			raw := strings.Join(rest, " ")
			d, err := parseMilestoneDate(raw)
			if err != nil {
				fmt.Printf(lang.L.MilestoneDateErrFmt, raw)
				return
			}
			date = d
		}
		if err := milestone.Save(sh.configDir, sh.srv.Name, date); err != nil {
			fmt.Printf(lang.L.ShellErrFmt, err)
			return
		}
		fmt.Printf(lang.L.MilestoneSetFmt, sh.srv.Name, date.Format("2006-01-02 15:04:05"))

	case "clear":
		if err := milestone.Clear(sh.configDir, sh.srv.Name); err != nil {
			fmt.Printf(lang.L.ShellErrFmt, err)
			return
		}
		fmt.Printf(lang.L.MilestoneClearedFmt, sh.srv.Name)

	case "sync":
		dry := false
		for _, a := range rest {
			if a == "--dry-run" {
				dry = true
			}
		}
		ms, err := milestone.Load(sh.configDir, sh.srv.Name)
		if err != nil || ms == nil {
			fmt.Printf(lang.L.MilestoneNotSetFmt, sh.srv.Name)
			return
		}
		since := ms.Date
		fmt.Printf("\n── %s ──\n", sh.srv.Name)
		sh.shellSyncServer(*sh.srv, false, dry, &since)

	default:
		fmt.Print(lang.L.MilestoneUsage)
	}
}

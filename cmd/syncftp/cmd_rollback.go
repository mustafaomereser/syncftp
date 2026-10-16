package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"syncftp/internal/config"
	ftpclient "syncftp/internal/ftp"
	"syncftp/internal/lang"
	"syncftp/internal/release"
	"syncftp/internal/state"
	"syncftp/internal/synclog"
)

var (
	rbFlagServer string
	rbFlagList   bool
	rbFlagYes    bool
)

func init() {
	rollbackCmd.Flags().StringVarP(&rbFlagServer, "server", "s", "", "Sunucu adı")
	rollbackCmd.Flags().BoolVar(&rbFlagList, "list", false, "Release geçmişini listele, geri alma yapma")
	rollbackCmd.Flags().BoolVarP(&rbFlagYes, "yes", "y", false, "Onay sormadan geri al")
	rootCmd.AddCommand(rollbackCmd)
}

var rollbackCmd = &cobra.Command{
	Use:   "rollback [release-id]",
	Short: "Bir sync'i geri alır: eski dosyaları geri yükler, yeni eklenenleri siler",
	Long: `Seçilen release'i ve ondan SONRAKİ tüm release'leri en yeniden başlayarak geri alır
(yani sunucu o sync'ten önceki haline döner).

Her sync, yüklediği dosyaların sunucudaki eski halini .syncftp/releases/<sunucu>/<id>/files/
altına yedekler; sunucuda olmayan dosyaları "yeni" olarak kaydeder. Rollback yedekleri geri
yükler, yeni dosyaları siler ve state'i eski hash'lere çeker — sonraki sync yerel değişiklikleri
tekrar tespit eder.

Yedeklemeyi kapatmak: sync.no_backup / server.no_backup. Tutulacak yedek sayısı: sync.backup_keep (varsayılan 10).`,
	Args: cobra.MaximumNArgs(1),
	RunE: runRollbackCmd,
}

func runRollbackCmd(cmd *cobra.Command, args []string) error {
	dir, _ := os.Getwd()
	cfg, err := config.Load(dir)
	if err != nil {
		return err
	}
	srv, ok, err := resolveOneServer(cfg, rbFlagServer)
	if err != nil || !ok {
		return err
	}

	list, err := release.List(dir, srv.Name)
	if err != nil {
		return err
	}
	if len(list) == 0 {
		fmt.Printf(lang.L.RollbackNoReleasesFmt, srv.Name)
		return nil
	}

	if rbFlagList {
		printReleaseList(os.Stdout, list)
		return nil
	}

	target := ""
	if len(args) == 1 {
		target = args[0]
	} else {
		items := rollbackPickerItems(list)
		if len(items) == 0 {
			fmt.Printf(lang.L.RollbackNoReleasesFmt, srv.Name)
			return nil
		}
		target, err = RunPicker(fmt.Sprintf(lang.L.RollbackPickTitleFmt, srv.Name), lang.L.RollbackPickSubtitle, items)
		if err != nil || target == "" {
			fmt.Println(lang.L.SyncCancelled)
			return nil
		}
	}

	plan, err := rollbackPlan(list, target)
	if err != nil {
		return err
	}
	if !rbFlagYes {
		var sb strings.Builder
		for _, m := range plan {
			fmt.Fprintf(&sb, "%s  %s\n", m.ID, releaseDesc(m))
		}
		okConfirm, _ := RunConfirm(fmt.Sprintf(lang.L.RollbackConfirmFmt, len(plan), srv.Name), sb.String())
		if !okConfirm {
			fmt.Println(lang.L.SyncCancelled)
			return nil
		}
	}

	_, logPath, err := rollbackLogged(os.Stdout, dir, srv, plan)
	if err != nil {
		return err
	}
	if logPath != "" {
		fmt.Printf(lang.L.SyncLogSavedFmt, logPath)
	}
	return nil
}

// resolveOneServer ad verilmişse o sunucuyu (aktif/pasif fark etmez), verilmemişse
// tek aktif sunucuyu ya da seçiciden seçileni döner. ok=false → kullanıcı iptal etti.
func resolveOneServer(cfg *config.Config, name string) (config.Server, bool, error) {
	if name != "" {
		for _, s := range cfg.Servers {
			if s.Name == name {
				return s, true, nil
			}
		}
		return config.Server{}, false, fmt.Errorf(lang.L.ServerNotFoundFmt+" (%v)", name, serverNames(cfg.Servers))
	}
	enabled := cfg.EnabledServers()
	switch len(enabled) {
	case 0:
		return config.Server{}, false, fmt.Errorf("aktif sunucu yok")
	case 1:
		return enabled[0], true, nil
	}
	s, err := pickServerTUI(enabled)
	if err != nil || s == nil {
		fmt.Println(lang.L.SyncCancelled)
		return config.Server{}, false, nil
	}
	return *s, true, nil
}

func releaseDesc(m *release.Manifest) string {
	desc := fmt.Sprintf(lang.L.RollbackReleaseDescFmt, m.CreatedAt.Format("2006-01-02 15:04:05"), m.FileCount, len(m.Backed), len(m.New))
	switch {
	case m.RolledBack:
		desc += "  " + lang.L.RollbackTagRolledBack
	case !m.HasBackup():
		desc += "  " + lang.L.RollbackTagNoBackup
	}
	return desc
}

func printReleaseList(w io.Writer, list []*release.Manifest) {
	for _, m := range list {
		fmt.Fprintf(w, "  %-18s %s\n", m.ID, releaseDesc(m))
	}
}

// rollbackPlan hedef release ve ondan sonraki (henüz geri alınmamış) release'leri
// en yeniden eskiye döner. Aradaki herhangi biri yedeksizse geri dönüş güvenli değildir.
func rollbackPlan(list []*release.Manifest, target string) ([]*release.Manifest, error) {
	var plan []*release.Manifest
	for _, m := range list { // list en yeniden eskiye
		if m.RolledBack {
			if m.ID == target {
				return nil, fmt.Errorf(lang.L.RollbackAlreadyFmt, target)
			}
			continue
		}
		if !m.HasBackup() {
			return nil, fmt.Errorf(lang.L.RollbackNoBackupFmt, m.ID)
		}
		plan = append(plan, m)
		if m.ID == target {
			return plan, nil
		}
	}
	return nil, fmt.Errorf(lang.L.RollbackNotFoundFmt, target)
}

type rollbackReport struct {
	Restored int      `json:"restored"`
	Deleted  int      `json:"deleted"`
	Failed   []string `json:"failed"`
	Releases []string `json:"releases"`
}

// performRollback plan'daki release'leri sırayla (en yeni önce) geri alır.
// Yedeklenen dosyalar geri yüklenir, yeni dosyalar silinir, state eski hash'lere döner.
// Ortak çekirdek: CLI, shell ve GUI aynı fonksiyonu kullanır.
func performRollback(w io.Writer, configDir string, srv config.Server, plan []*release.Manifest) (rollbackReport, error) {
	var rep rollbackReport
	client, err := ftpclient.Connect(srv)
	if err != nil {
		return rep, err
	}
	defer client.Close()

	st, err := state.Load(configDir, srv.Name)
	if err != nil {
		return rep, err
	}

	for _, m := range plan {
		fmt.Fprintf(w, lang.L.RollbackReleaseFmt, m.ID, m.CreatedAt.Format("2006-01-02 15:04:05"))
		filesDir := release.FilesDir(configDir, srv.Name, m.ID)
		failedHere := 0
		for _, rel := range m.Backed {
			src := filepath.Join(filesDir, filepath.FromSlash(rel))
			if err := client.Upload(src, rel); err != nil {
				fmt.Fprintf(w, lang.L.RollbackFileErrFmt, rel, err)
				rep.Failed = append(rep.Failed, rel)
				failedHere++
				continue
			}
			fmt.Fprintf(w, lang.L.RollbackRestoredFmt, rel)
			rep.Restored++
			if h, ok := m.PrevHashes[rel]; ok {
				st.Files[rel] = h
			} else {
				delete(st.Files, rel)
			}
		}
		for _, rel := range m.New {
			// Zaten yoksa (sonraki bir release'in rollback'i sildiyse) başarı say
			if err := client.DeleteFile(path.Join(srv.RemotePath, rel)); err != nil && !ftpclient.IsNotFound(err) {
				fmt.Fprintf(w, lang.L.RollbackFileErrFmt, rel, err)
				rep.Failed = append(rep.Failed, rel)
				failedHere++
				continue
			}
			fmt.Fprintf(w, lang.L.RollbackDeletedFmt, rel)
			rep.Deleted++
			delete(st.Files, rel)
		}
		if len(m.NoBackup) > 0 {
			fmt.Fprintf(w, lang.L.RollbackSkippedNoBackupFmt, len(m.NoBackup))
		}
		// Kısmi hata: release geri alınmış sayılmaz (tekrar denenebilir)
		if failedHere == 0 {
			now := time.Now()
			m.RolledBack = true
			m.RolledBackAt = &now
			_, _ = release.Write(configDir, srv.Name, m.ID, m)
			rep.Releases = append(rep.Releases, m.ID)
		}
		_ = state.Save(configDir, st)
	}
	fmt.Fprintf(w, lang.L.RollbackDoneFmt, rep.Restored, rep.Deleted, len(rep.Failed))
	return rep, nil
}

// rollbackPickerItems geri alınabilir release'ler (CLI + shell seçicisi).
func rollbackPickerItems(list []*release.Manifest) []PickerItem {
	var items []PickerItem
	for _, m := range list {
		if m.Restorable() {
			items = append(items, PickerItem{Icon: "↺", Label: m.ID, Desc: releaseDesc(m), Value: m.ID})
		}
	}
	return items
}

// rollbackLogged geri alır ve çıktıyı sync log'una kaydeder (CLI, shell, panel ortak).
func rollbackLogged(w io.Writer, configDir string, srv config.Server, plan []*release.Manifest) (rollbackReport, string, error) {
	var buf strings.Builder
	rep, err := performRollback(io.MultiWriter(w, &buf), configDir, srv, plan)
	if err != nil {
		return rep, "", err
	}
	p, _ := synclog.Save(configDir, srv.Name, buf.String())
	return rep, p, nil
}

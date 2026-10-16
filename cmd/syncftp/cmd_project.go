package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"syncftp/internal/config"
	"syncftp/internal/lang"
	"syncftp/internal/projects"
)

// flagProject — global --project/-P: komutu kayıtlı başka bir projede çalıştırır.
var flagProject string

func init() {
	rootCmd.PersistentFlags().StringVarP(&flagProject, "project", "P", "", "Komutu kayıtlı projede çalıştır (ad veya yol) — bkz. syncftp project list")
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if flagProject != "" {
			p, ok := projects.Find(flagProject)
			if !ok {
				return fmt.Errorf(lang.L.ProjectNotFoundFmt, flagProject)
			}
			if err := os.Chdir(p.Path); err != nil {
				return err
			}
			lang.Init(p.Path)
		}
		// Config'i olan her dizin kayda girer / son kullanım güncellenir (GUI listesi için)
		if dir, err := os.Getwd(); err == nil && projects.HasConfig(dir) {
			projects.Touch(dir)
		}
		return nil
	}

	projectCmd.AddCommand(projectListCmd, projectAddCmd, projectRmCmd, projectRenameCmd)
	projectAddCmd.Flags().StringVar(&projectAddName, "name", "", "Görünen ad (varsayılan: klasör adı)")
	rootCmd.AddCommand(projectCmd)
}

var projectAddName string

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Kayıtlı projeleri yönetir (çoklu proje — GUI ve --project için)",
	Long: `syncFTP, config'i olan her projeyi makine genelindeki bir kayda ekler
(<UserConfigDir>/syncftp/projects.json). Kayıtlı projeler GUI'de listelenir ve
herhangi bir dizinden --project ile hedeflenebilir:

  syncftp -P magaza sync --all
  syncftp --project D:\siteler\blog status`,
	RunE: func(cmd *cobra.Command, args []string) error { return projectListCmd.RunE(cmd, args) },
}

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "Kayıtlı projeleri listeler",
	RunE: func(cmd *cobra.Command, args []string) error {
		list := projects.List()
		if len(list) == 0 {
			fmt.Print(lang.L.ProjectListEmpty)
			return nil
		}
		for _, p := range list {
			mark := " "
			if !projects.HasConfig(p.Path) {
				mark = "!"
			}
			used := "-"
			if !p.LastUsed.IsZero() {
				used = p.LastUsed.Format("2006-01-02 15:04")
			}
			fmt.Printf("  %s %-24s %-16s %s\n", mark, p.Name, used, p.Path)
		}
		return nil
	},
}

var projectAddCmd = &cobra.Command{
	Use:   "add [yol]",
	Short: "Projeyi kayda ekler (varsayılan: bulunulan dizin)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := os.Getwd()
		if len(args) == 1 {
			dir = args[0]
		}
		p, err := projects.Add(dir, projectAddName)
		if err != nil {
			return err
		}
		fmt.Printf(lang.L.ProjectAddedFmt, p.Name, p.Path)
		return nil
	},
}

var projectRmCmd = &cobra.Command{
	Use:   "rm <ad|yol>",
	Short: "Projeyi kayıttan çıkarır (dosyalara dokunmaz)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := projects.Remove(args[0]); err != nil {
			return err
		}
		fmt.Printf(lang.L.ProjectRemovedFmt, args[0])
		return nil
	},
}

var projectRenameCmd = &cobra.Command{
	Use:   "rename <ad|yol> <yeni-ad>",
	Short: "Projenin görünen adını değiştirir",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		p, ok := projects.Find(args[0])
		if !ok {
			return fmt.Errorf(lang.L.ProjectNotFoundFmt, args[0])
		}
		if err := renameProject(p.Path, args[1]); err != nil {
			return err
		}
		_ = projects.Rename(p.Path, args[1]) // kayıttaki görünen ad da aynı kalsın
		fmt.Printf(lang.L.ProjectRenamedFmt, args[0], args[1])
		return nil
	},
}

// renameProject projenin adını config'teki project.name'e yazar — CLI `project rename` ve panelin
// "Projeyi yeniden adlandır"ı ortak (panel klasör modunda adı buradan okur).
func renameProject(dir, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("name")
	}
	cfg, err := config.Load(dir)
	if err != nil {
		return err
	}
	cfg.Project.Name = name
	if err := config.Save(dir, cfg); err != nil {
		return err
	}
	projects.Invalidate()
	return nil
}

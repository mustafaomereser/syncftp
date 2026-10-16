package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"syncftp/internal/config"
	"syncftp/internal/lang"
	"syncftp/internal/synclog"
)

// syncftp logs — sync / rollback / calibrate logları (.syncftp/logs/<sunucu>/). Panelin Geçmiş
// sekmesindeki log listesiyle aynı kaynak (synclog.List / synclog.Read).

var logsFlagServer string
var logsFlagLimit int

var logsCmd = &cobra.Command{
	Use:   "logs [dosya]",
	Short: "Sync loglarını listele ya da birini göster",
	Long: `Argümansız: en yeni loglar (sunucu, dosya, boyut, tarih).
Dosya adı verilirse içeriğini yazar; "son" / "last" en yeni logu gösterir.

  syncftp logs
  syncftp logs --server production
  syncftp logs last
  syncftp logs 20260716-143012.txt --server production`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := os.Getwd()
		cfg, err := config.Load(dir)
		if err != nil {
			return err
		}
		file := ""
		if len(args) == 1 {
			file = args[0]
		}
		return showLogs(os.Stdout, dir, cfg, logsFlagServer, file, logsFlagLimit)
	},
}

func init() {
	logsCmd.Flags().StringVarP(&logsFlagServer, "server", "s", "", "Yalnızca bu sunucunun logları")
	logsCmd.Flags().IntVarP(&logsFlagLimit, "limit", "n", 20, "Listede en fazla kaç log")
	rootCmd.AddCommand(logsCmd)
}

// showLogs CLI `logs` ve shell `logs` ortak.
func showLogs(w io.Writer, dir string, cfg *config.Config, server, file string, limit int) error {
	var names []string
	for _, s := range cfg.Servers {
		if server == "" || s.Name == server {
			names = append(names, s.Name)
		}
	}
	if server != "" && len(names) == 0 {
		return fmt.Errorf(lang.L.ServerNotFoundFmt, server)
	}
	if file == "" {
		list := synclog.List(dir, names, limit)
		if len(list) == 0 {
			fmt.Fprintln(w, lang.L.LogsNone)
			return nil
		}
		for _, e := range list {
			fmt.Fprintf(w, "  %s  %-14s %-26s %8s\n", e.Time.Format("2006-01-02 15:04"), e.Server, e.File, humanBytes(e.Size))
		}
		fmt.Fprintln(w, lang.L.LogsHint)
		return nil
	}
	srvName := server
	if file == "last" || file == "son" {
		list := synclog.List(dir, names, 1)
		if len(list) == 0 {
			fmt.Fprintln(w, lang.L.LogsNone)
			return nil
		}
		srvName, file = list[0].Server, list[0].File
	} else if srvName == "" {
		// Ad verilip sunucu verilmediyse: dosyanın bulunduğu sunucu
		for _, e := range synclog.List(dir, names, 0) {
			if e.File == file {
				srvName = e.Server
				break
			}
		}
	}
	text, err := synclog.Read(dir, srvName, file, 0)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "── %s / %s ──\n%s", srvName, file, text)
	return nil
}

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"syncftp/internal/config"
	"syncftp/internal/lang"
	"syncftp/internal/state"
)

var (
	statusFlagInclude []string
	statusFlagExclude []string
	statusFlagServer  string
)

func init() {
	statusCmd.Flags().StringArrayVar(&statusFlagInclude, "include", nil, "Sadece bu yol/klasörleri göster (whitelist)")
	statusCmd.Flags().StringArrayVar(&statusFlagExclude, "exclude", nil, "Bu yol/klasörleri sonuçtan hariç tut")
	statusCmd.Flags().StringVarP(&statusFlagServer, "server", "s", "", "Sadece bu sunucunun durumunu göster")
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Değişen dosyaları listeler (hiçbir şey yüklemez)",
	RunE:  runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	dir, _ := os.Getwd()

	cfg, err := config.Load(dir)
	if err != nil {
		return err
	}

	// CLI --include varsa sync.include'u geçersiz kılar
	effectiveInclude := cfg.Sync.Include
	if len(statusFlagInclude) > 0 {
		effectiveInclude = statusFlagInclude
	}

	if len(effectiveInclude) > 0 {
		fmt.Printf(lang.L.StatusWhitelistFmt, len(effectiveInclude))
		for _, p := range effectiveInclude {
			fmt.Printf("  + %s\n", p)
		}
		fmt.Println()
	}
	if len(statusFlagExclude) > 0 {
		fmt.Printf(lang.L.StatusExcludeFmt, len(statusFlagExclude))
		for _, p := range statusFlagExclude {
			fmt.Printf("  - %s\n", p)
		}
		fmt.Println()
	}

	fmt.Printf(lang.L.StatusProjectFmt, cfg.Project.Name)

	servers := cfg.EnabledServers()
	if statusFlagServer != "" {
		var found []config.Server
		for _, s := range servers {
			if s.Name == statusFlagServer {
				found = append(found, s)
			}
		}
		if len(found) == 0 {
			return fmt.Errorf(lang.L.ServerNotFoundFmt, statusFlagServer)
		}
		servers = found
	}

	for _, srv := range servers {
		localPath := srv.EffectiveLocalPath(cfg.Project)
		current, _, err := scanForServer(dir, cfg, srv)
		if err != nil {
			fmt.Printf("── %s ──\n  %v\n\n", srv.Name, err)
			continue
		}

		st, err := state.Load(dir, srv.Name)
		if err != nil {
			fmt.Printf(lang.L.StatusStateErr, srv.Name, err)
			continue
		}

		if !st.FirstSyncDone {
			fmt.Print(lang.L.ResyncAutoMsg)
			runCalibrate(dir, srv, cfg)
			if reloaded, err2 := state.Load(dir, srv.Name); err2 == nil {
				st = reloaded
			}
		}

		// Sync'in yükleyeceği liste ile birebir aynı hesap (cmd_plan.go)
		pi := pendingChanges(dir, cfg, srv, current, st, statusFlagInclude, statusFlagExclude)

		fmt.Printf("── %s (%s) ──\n", srv.Name, srv.Host)
		fmt.Printf(lang.L.StatusDirFmt, filepath.Join(dir, localPath))
		fmt.Printf(lang.L.StatusFileFmt, len(current))
		if pi.plan.msDate != nil {
			fmt.Printf(lang.L.MilestoneFilterFmt, pi.plan.msDate.Format("2006-01-02 15:04"), pi.plan.msDropped)
		}

		if !st.FirstSyncDone {
			fmt.Println(lang.L.StatusNoFirstSync)
		} else if len(pi.newL)+len(pi.changed)+len(pi.deleted) == 0 {
			if len(effectiveInclude)+len(statusFlagExclude) > 0 {
				fmt.Println(lang.L.StatusFilteredUpToDate)
			} else {
				fmt.Println(lang.L.StatusUpToDate)
			}
		}

		printFileList(lang.L.StatusNewHeader, pi.newL)
		printFileList(lang.L.StatusChangedHeader, pi.changed)
		printFileList(lang.L.StatusDeletedHeader, pi.deleted)
		printFileList(lang.L.StatusFrozenHeader, pi.plan.frozen)
		printFileList(lang.L.StatusProtectedHeader, pi.plan.protected)
		printFileList(lang.L.StatusBlockedHeader, pi.plan.blocked)
		fmt.Println()
	}

	return nil
}

func printFileList(header string, items []string) {
	if len(items) == 0 {
		return
	}
	sort.Strings(items)
	fmt.Printf("%s (%d):\n", header, len(items))
	for _, item := range items {
		fmt.Printf("    %s\n", item)
	}
}

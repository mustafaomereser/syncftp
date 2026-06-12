package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"syncftp/internal/config"
	"syncftp/internal/ignore"
	"syncftp/internal/lang"
	"syncftp/internal/scanner"
	"syncftp/internal/state"
)

var (
	statusFlagInclude []string
	statusFlagExclude []string
)

func init() {
	statusCmd.Flags().StringArrayVar(&statusFlagInclude, "include", nil, "Sadece bu yol/klasörleri göster (whitelist)")
	statusCmd.Flags().StringArrayVar(&statusFlagExclude, "exclude", nil, "Bu yol/klasörleri sonuçtan hariç tut")
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

	projectDir := filepath.Join(dir, cfg.Project.LocalPath)

	matcher, err := ignore.Load(projectDir, cfg.Sync.IgnoreFiles)
	if err != nil {
		return fmt.Errorf("ignore dosyası yüklenemedi: %w", err)
	}

	files, err := scanner.Scan(projectDir, matcher)
	if err != nil {
		return fmt.Errorf("tarama başarısız: %w", err)
	}

	current := make(map[string]string, len(files))
	for _, f := range files {
		current[f.RelPath] = f.Hash
	}

	// CLI --include varsa TOML sync.include'u geçersiz kılar
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
	fmt.Printf(lang.L.StatusDirFmt, projectDir)
	fmt.Printf(lang.L.StatusFileFmt, len(files))
	fmt.Println()

	for _, srv := range cfg.EnabledServers() {
		st, err := state.Load(dir, srv.Name)
		if err != nil {
			fmt.Printf(lang.L.StatusStateErr, srv.Name, err)
			continue
		}

		diff := state.Diff(st, current)

		// Include önceliği: CLI > sunucu > global
		srvInclude := effectiveInclude
		if len(srv.Include) > 0 && len(statusFlagInclude) == 0 {
			srvInclude = srv.Include
		}
		// Exclude: global + sunucu + CLI
		srvExclude := append(append([]string{}, cfg.Sync.Exclude...), srv.Exclude...)
		srvExclude = append(srvExclude, statusFlagExclude...)

		newFiles := applyStatusFilters(diff.New, srvInclude, srvExclude)
		changedFiles := applyStatusFilters(diff.Changed, srvInclude, srvExclude)
		deletedFiles := applyStatusFilters(diff.Deleted, srvInclude, srvExclude)

		fmt.Printf("── %s (%s) ──\n", srv.Name, srv.Host)

		if !st.FirstSyncDone {
			fmt.Println(lang.L.StatusNoFirstSync)
		} else if len(newFiles)+len(changedFiles)+len(deletedFiles) == 0 {
			if len(effectiveInclude)+len(statusFlagExclude) > 0 {
				fmt.Println(lang.L.StatusFilteredUpToDate)
			} else {
				fmt.Println(lang.L.StatusUpToDate)
			}
		}

		printFileList(lang.L.StatusNewHeader, newFiles)
		printFileList(lang.L.StatusChangedHeader, changedFiles)
		printFileList(lang.L.StatusDeletedHeader, deletedFiles)
		fmt.Println()
	}

	return nil
}

func applyStatusFilters(paths, include, exclude []string) []string {
	result := paths
	if len(include) > 0 {
		var filtered []string
		for _, p := range result {
			for _, inc := range include {
				inc = filepath.ToSlash(inc)
				if p == inc || strings.HasPrefix(p, inc+"/") {
					filtered = append(filtered, p)
					break
				}
			}
		}
		result = filtered
	}
	if len(exclude) > 0 {
		var filtered []string
		for _, p := range result {
			excluded := false
			for _, exc := range exclude {
				exc = filepath.ToSlash(exc)
				if p == exc || strings.HasPrefix(p, exc+"/") {
					excluded = true
					break
				}
			}
			if !excluded {
				filtered = append(filtered, p)
			}
		}
		result = filtered
	}
	return result
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

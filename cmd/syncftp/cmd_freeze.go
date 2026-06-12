package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"syncftp/internal/config"
	"syncftp/internal/frozen"
	"syncftp/internal/ignore"
	"syncftp/internal/scanner"
)

var flagFreezeServer string

func init() {
	freezeCmd.Flags().StringVar(&flagFreezeServer, "server", "", "Server to configure freeze list for")
	rootCmd.AddCommand(freezeCmd)
}

var freezeCmd = &cobra.Command{
	Use:   "freeze",
	Short: "Mark files to never upload for a specific server",
	RunE:  runFreeze,
}

func runFreeze(cmd *cobra.Command, args []string) error {
	dir, _ := os.Getwd()
	cfg, err := config.Load(dir)
	if err != nil {
		return err
	}

	srv, err := resolveFreezeServer(cfg, flagFreezeServer)
	if err != nil || srv == nil {
		return err
	}

	items, err := buildFreezeItems(dir, cfg)
	if err != nil {
		return err
	}

	fl, err := frozen.Load(dir, srv.Name)
	if err != nil {
		return err
	}

	selected, err := RunFreezeList(
		fmt.Sprintf("Freeze list — %s", srv.Name),
		items,
		fl.Files,
	)
	if err != nil {
		return err
	}
	if selected == nil {
		fmt.Println("Cancelled.")
		return nil
	}

	if err := frozen.Save(dir, srv.Name, selected); err != nil {
		return fmt.Errorf("could not save freeze list: %w", err)
	}
	fmt.Printf("✓ %d files frozen for [%s]\n", len(selected), srv.Name)
	return nil
}

func resolveFreezeServer(cfg *config.Config, name string) (*config.Server, error) {
	servers := cfg.EnabledServers()
	if len(servers) == 0 {
		return nil, fmt.Errorf("no active servers")
	}
	if name != "" {
		for _, s := range servers {
			s2 := s
			if s.Name == name {
				return &s2, nil
			}
		}
		return nil, fmt.Errorf("server not found: %q", name)
	}
	if len(servers) == 1 {
		s := servers[0]
		return &s, nil
	}
	return pickServerTUI(servers)
}

func buildFreezeItems(configDir string, cfg *config.Config) ([]PickerItem, error) {
	projectDir := filepath.Join(configDir, cfg.Project.LocalPath)
	matcher, err := ignore.Load(projectDir)
	if err != nil {
		return nil, err
	}
	files, err := scanner.Scan(projectDir, matcher)
	if err != nil {
		return nil, err
	}
	sort.Slice(files, func(i, j int) bool { return files[i].RelPath < files[j].RelPath })

	items := make([]PickerItem, len(files))
	for i, f := range files {
		items[i] = PickerItem{Label: f.RelPath, Value: f.RelPath}
	}
	return items, nil
}

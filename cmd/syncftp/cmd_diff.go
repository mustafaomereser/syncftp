package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"syncftp/internal/config"
	ftpclient "syncftp/internal/ftp"
	"syncftp/internal/lang"
)

func init() {
	rootCmd.AddCommand(diffCmd)
}

var diffCmd = &cobra.Command{
	Use:   "diff [sunucuA] [sunucuB]",
	Short: "İki sunucunun FTP içeriğini karşılaştırır (production vs staging)",
	RunE:  runDiff,
}

func runDiff(cmd *cobra.Command, args []string) error {
	dir, _ := os.Getwd()
	cfg, err := config.Load(dir)
	if err != nil {
		return err
	}
	if len(cfg.Servers) < 2 {
		fmt.Print(lang.L.DiffNeedTwo)
		return nil
	}

	pickServer := func(title, exclude string) (*config.Server, error) {
		var items []PickerItem
		for _, s := range cfg.Servers {
			if s.Name == exclude {
				continue
			}
			items = append(items, PickerItem{Icon: "🖥", Label: s.Name, Desc: s.Host + s.RemotePath, Value: s.Name})
		}
		val, err := RunPicker(title, lang.L.DiffPickSub, items)
		if err != nil || val == "" {
			return nil, err
		}
		for i := range cfg.Servers {
			if cfg.Servers[i].Name == val {
				return &cfg.Servers[i], nil
			}
		}
		return nil, nil
	}

	findServer := func(name string) *config.Server {
		for i := range cfg.Servers {
			if cfg.Servers[i].Name == name {
				return &cfg.Servers[i]
			}
		}
		return nil
	}

	var srvA, srvB *config.Server
	if len(args) >= 1 {
		srvA = findServer(args[0])
		if srvA == nil {
			return fmt.Errorf("sunucu bulunamadı: %q", args[0])
		}
	} else {
		if srvA, err = pickServer(lang.L.DiffPickA, ""); err != nil || srvA == nil {
			return err
		}
	}
	if len(args) >= 2 {
		srvB = findServer(args[1])
		if srvB == nil {
			return fmt.Errorf("sunucu bulunamadı: %q", args[1])
		}
	} else {
		if srvB, err = pickServer(lang.L.DiffPickB, srvA.Name); err != nil || srvB == nil {
			return err
		}
	}
	if srvA.Name == srvB.Name {
		fmt.Print(lang.L.DiffNeedTwo)
		return nil
	}

	fmt.Printf(lang.L.DiffTitleFmt, srvA.Name, srvB.Name)

	listServer := func(srv *config.Server) (map[string]uint64, error) {
		fmt.Printf(lang.L.DiffListingFmt, srv.Name)
		client, err := ftpclient.Connect(*srv)
		if err != nil {
			return nil, err
		}
		defer client.Close()
		return client.ListRecursive(srv.RemotePath)
	}

	filesA, err := listServer(srvA)
	if err != nil {
		return err
	}
	filesB, err := listServer(srvB)
	if err != nil {
		return err
	}

	var onlyA, onlyB, sizeDiff []string
	same := 0
	for p, sa := range filesA {
		sb, ok := filesB[p]
		switch {
		case !ok:
			onlyA = append(onlyA, p)
		case sa != sb:
			sizeDiff = append(sizeDiff, p)
		default:
			same++
		}
	}
	for p := range filesB {
		if _, ok := filesA[p]; !ok {
			onlyB = append(onlyB, p)
		}
	}
	sort.Strings(onlyA)
	sort.Strings(onlyB)
	sort.Strings(sizeDiff)

	fmt.Println()
	if len(onlyA) > 0 {
		fmt.Printf(lang.L.DiffOnlyFmt, srvA.Name, len(onlyA))
		for _, p := range onlyA {
			fmt.Printf("    + %s\n", p)
		}
	}
	if len(onlyB) > 0 {
		fmt.Printf(lang.L.DiffOnlyFmt, srvB.Name, len(onlyB))
		for _, p := range onlyB {
			fmt.Printf("    + %s\n", p)
		}
	}
	if len(sizeDiff) > 0 {
		fmt.Printf(lang.L.DiffSizeHeaderFmt, len(sizeDiff))
		for _, p := range sizeDiff {
			fmt.Printf(lang.L.DiffSizeLineFmt, p,
				srvA.Name, humanBytes(int64(filesA[p])),
				srvB.Name, humanBytes(int64(filesB[p])))
		}
	}
	fmt.Printf(lang.L.DiffSameFmt, same)
	return nil
}

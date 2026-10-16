package main

import (
	"fmt"
	"io"
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
			return fmt.Errorf(lang.L.ServerNotFoundFmt, args[0])
		}
	} else {
		if srvA, err = pickServer(lang.L.DiffPickA, ""); err != nil || srvA == nil {
			return err
		}
	}
	if len(args) >= 2 {
		srvB = findServer(args[1])
		if srvB == nil {
			return fmt.Errorf(lang.L.ServerNotFoundFmt, args[1])
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

	_, err = diffServers(os.Stdout, *srvA, *srvB)
	return err
}

type serverDiffReport struct {
	A        string   `json:"a"`
	B        string   `json:"b"`
	OnlyA    []string `json:"only_a"`
	OnlyB    []string `json:"only_b"`
	SizeDiff []string `json:"size_diff"`
	Same     int      `json:"same"`
}

// diffServers iki sunucunun uzak içeriğini karşılaştırır (CLI + GUI ortak çekirdek).
func diffServers(w io.Writer, srvA, srvB config.Server) (serverDiffReport, error) {
	rep := serverDiffReport{A: srvA.Name, B: srvB.Name, OnlyA: []string{}, OnlyB: []string{}, SizeDiff: []string{}}
	fmt.Fprintf(w, lang.L.DiffTitleFmt, srvA.Name, srvB.Name)

	listServer := func(srv config.Server) (map[string]uint64, error) {
		fmt.Fprintf(w, lang.L.DiffListingFmt, srv.Name)
		client, err := ftpclient.Connect(srv)
		if err != nil {
			return nil, err
		}
		defer client.Close()
		return client.ListRecursive(srv.RemotePath)
	}

	filesA, err := listServer(srvA)
	if err != nil {
		return rep, err
	}
	filesB, err := listServer(srvB)
	if err != nil {
		return rep, err
	}

	for p, sa := range filesA {
		sb, ok := filesB[p]
		switch {
		case !ok:
			rep.OnlyA = append(rep.OnlyA, p)
		case sa != sb:
			rep.SizeDiff = append(rep.SizeDiff, p)
		default:
			rep.Same++
		}
	}
	for p := range filesB {
		if _, ok := filesA[p]; !ok {
			rep.OnlyB = append(rep.OnlyB, p)
		}
	}
	sort.Strings(rep.OnlyA)
	sort.Strings(rep.OnlyB)
	sort.Strings(rep.SizeDiff)

	fmt.Fprintln(w)
	if len(rep.OnlyA) > 0 {
		fmt.Fprintf(w, lang.L.DiffOnlyFmt, srvA.Name, len(rep.OnlyA))
		for _, p := range rep.OnlyA {
			fmt.Fprintf(w, "    + %s\n", p)
		}
	}
	if len(rep.OnlyB) > 0 {
		fmt.Fprintf(w, lang.L.DiffOnlyFmt, srvB.Name, len(rep.OnlyB))
		for _, p := range rep.OnlyB {
			fmt.Fprintf(w, "    + %s\n", p)
		}
	}
	if len(rep.SizeDiff) > 0 {
		fmt.Fprintf(w, lang.L.DiffSizeHeaderFmt, len(rep.SizeDiff))
		for _, p := range rep.SizeDiff {
			fmt.Fprintf(w, lang.L.DiffSizeLineFmt, p,
				srvA.Name, humanBytes(int64(filesA[p])),
				srvB.Name, humanBytes(int64(filesB[p])))
		}
	}
	fmt.Fprintf(w, lang.L.DiffSameFmt, rep.Same)
	return rep, nil
}

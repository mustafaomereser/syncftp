package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path"
	"strings"

	"github.com/spf13/cobra"

	"syncftp/internal/lang"
)

// `syncftp remote zip|unzip` — panelin zip / zipten çıkar işleriyle aynı çekirdek (remote_zip.go).

var (
	remoteZipOut  string
	remoteUnzipTo string
)

func init() {
	remoteZipCmd.Flags().StringVarP(&remoteZipOut, "out", "o", "", "Arşiv adı (varsayılan: tek öğede <ad>.zip, çok öğede <klasör>.zip); aynı klasöre yazılır")
	remoteUnzipCmd.Flags().StringVar(&remoteUnzipTo, "to", "", "Hedef klasör (varsayılan: arşivin yanında arşiv adıyla bir klasör)")
	remoteCmd.AddCommand(remoteZipCmd, remoteUnzipCmd)
}

var remoteZipCmd = &cobra.Command{
	Use:   "zip <path...>",
	Short: "Sunucudaki dosya / klasörleri zip arşivine sıkıştır (aynı klasördeki öğeler)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		srv, client, err := connectRemote()
		if err != nil {
			return err
		}
		defer client.Close()
		var srcs []string
		for _, a := range args {
			full := resolveRemotePath(srv, a)
			if path.Clean(full) == path.Clean(srv.RemotePath) {
				return errors.New("root")
			}
			srcs = append(srcs, full)
		}
		base := path.Dir(srcs[0])
		for _, s := range srcs[1:] {
			if path.Dir(s) != base {
				return fmt.Errorf("%s: %s", s, "all items must be in the same folder")
			}
		}
		name := zipDefaultName(base, srcs)
		if remoteZipOut != "" {
			if name, err = checkNewName(remoteZipOut); err != nil {
				return err
			}
		}
		if !strings.HasSuffix(strings.ToLower(name), ".zip") {
			name += ".zip"
		}
		out := path.Join(base, uniqueRemoteName(client, base, name))
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		n, size, err := remoteZip(ctx, client, base, srcs, out, os.Stdout)
		if err != nil {
			if ctx.Err() != nil {
				fmt.Print(lang.L.RemoteZipStopped)
				return nil
			}
			return err
		}
		fmt.Printf(lang.L.RemoteZipDoneFmt, n, out, humanBytes(size))
		return nil
	},
}

var remoteUnzipCmd = &cobra.Command{
	Use:   "unzip <zip-path>",
	Short: "Sunucudaki zip arşivini sunucuda çıkar (var olan dosyaların üzerine yazar)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		srv, client, err := connectRemote()
		if err != nil {
			return err
		}
		defer client.Close()
		zipFull := resolveRemotePath(srv, args[0])
		dest := strings.TrimSuffix(zipFull, path.Ext(zipFull))
		if remoteUnzipTo != "" {
			dest = resolveRemotePath(srv, remoteUnzipTo)
		}
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		n, err := remoteUnzip(ctx, client, zipFull, dest, os.Stdout)
		if err != nil {
			if ctx.Err() != nil {
				fmt.Printf(lang.L.RemoteUnzipStoppedFmt, n)
				return nil
			}
			return err
		}
		fmt.Printf(lang.L.RemoteUnzipDoneFmt, n, dest)
		return nil
	},
}

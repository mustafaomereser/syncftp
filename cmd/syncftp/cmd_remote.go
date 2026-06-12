package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/tabwriter"

	goftp "github.com/jlaffaye/ftp"
	"github.com/spf13/cobra"

	"syncftp/internal/config"
	ftpclient "syncftp/internal/ftp"
	"syncftp/internal/lang"
)

var remoteServer string

func init() {
	remoteCmd.PersistentFlags().StringVar(&remoteServer, "server", "", "Bağlanılacak sunucu adı (tek sunucu varsa otomatik seçilir)")

	remoteLsCmd.Flags().BoolVar(&remoteLsRecursive, "recursive", false, "Alt dizinleri de listele")
	remoteRmCmd.Flags().BoolVar(&remoteRmForce, "force", false, "Onay sormadan sil")
	remoteRmCmd.Flags().BoolVar(&remoteRmRecursive, "recursive", false, "Dizini içeriğiyle birlikte sil")
	remoteCatCmd.Flags().IntVar(&remoteCatMaxKB, "max-kb", 10, "Gösterilecek maksimum kilobayt")

	remoteCmd.AddCommand(remoteLsCmd, remoteGetCmd, remoteRmCmd, remoteCatCmd)
	rootCmd.AddCommand(remoteCmd)
}

var remoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "FTP sunucusundaki dosyaları listele, indir, sil, önizle",
}

// ── ls ────────────────────────────────────────────────────────────────────────

var remoteLsRecursive bool

var remoteLsCmd = &cobra.Command{
	Use:   "ls [path]",
	Short: "FTP sunucusundaki dosya ve dizinleri listele",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		srv, client, err := connectRemote()
		if err != nil {
			return err
		}
		defer client.Close()

		remotePath := srv.RemotePath
		if len(args) == 1 {
			remotePath = resolveRemotePath(srv, args[0])
		}

		fmt.Printf(lang.L.RemoteServerFmt, srv.Name, srv.Host)
		fmt.Printf(lang.L.RemoteDirFmt, remotePath)

		return listDir(client, remotePath, "", remoteLsRecursive)
	},
}

func listDir(client *ftpclient.Client, remotePath, indent string, recursive bool) error {
	entries, err := client.List(remotePath)
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, e := range entries {
		if e.Name == "." || e.Name == ".." {
			continue
		}
		switch e.Type {
		case goftp.EntryTypeFolder:
			fmt.Fprintf(w, "%s  d  %s/\t\t%s\n", indent, e.Name, e.Time.Format("2006-01-02 15:04"))
		case goftp.EntryTypeFile:
			fmt.Fprintf(w, "%s  f  %s\t%s\t%s\n", indent, e.Name, formatSize(e.Size), e.Time.Format("2006-01-02 15:04"))
		case goftp.EntryTypeLink:
			fmt.Fprintf(w, "%s  l  %s\t\t%s\n", indent, e.Name, e.Time.Format("2006-01-02 15:04"))
		}
	}
	w.Flush()

	if recursive {
		for _, e := range entries {
			if e.Type == goftp.EntryTypeFolder && e.Name != "." && e.Name != ".." {
				subPath := path.Join(remotePath, e.Name)
				fmt.Printf("\n%s  %s/\n", indent, e.Name)
				if err := listDir(client, subPath, indent+"  ", true); err != nil {
					fmt.Printf(lang.L.RemoteListErr, indent, err)
				}
			}
		}
	}
	return nil
}

// ── get ───────────────────────────────────────────────────────────────────────

var remoteGetCmd = &cobra.Command{
	Use:   "get [remote-path] [local-dest]",
	Short: "FTP'den dosya indir",
	Args:  cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		srv, client, err := connectRemote()
		if err != nil {
			return err
		}
		defer client.Close()

		var remotePath string
		if len(args) >= 1 {
			resolved := resolveRemotePath(srv, args[0])
			if entries, lsErr := client.List(resolved); lsErr == nil && entries != nil {
				var pickErr error
				remotePath, pickErr = pickRemoteFile(client, resolved)
				if pickErr != nil {
					return pickErr
				}
			} else {
				remotePath = resolved
			}
		} else {
			var pickErr error
			remotePath, pickErr = pickRemoteFile(client, srv.RemotePath)
			if pickErr != nil {
				return pickErr
			}
		}

		localDest := filepath.Base(remotePath)
		if len(args) == 2 {
			localDest = args[1]
		}

		fmt.Printf(lang.L.RemoteDownloading, remotePath, localDest)
		if err := client.Download(remotePath, localDest); err != nil {
			return err
		}

		info, _ := os.Stat(localDest)
		if info != nil {
			fmt.Printf(lang.L.RemoteDownloaded, formatSize(uint64(info.Size())))
		} else {
			fmt.Println(lang.L.RemoteDownloadedBare)
		}
		return nil
	},
}

// ── rm ────────────────────────────────────────────────────────────────────────

var (
	remoteRmForce     bool
	remoteRmRecursive bool
)

var remoteRmCmd = &cobra.Command{
	Use:   "rm [remote-path]",
	Short: "FTP'den dosya veya dizin sil",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		srv, client, err := connectRemote()
		if err != nil {
			return err
		}
		defer client.Close()

		var remotePath string
		if len(args) == 1 {
			resolved := resolveRemotePath(srv, args[0])
			if entries, lsErr := client.List(resolved); lsErr == nil && entries != nil {
				var pickErr error
				remotePath, pickErr = pickRemoteFile(client, resolved)
				if pickErr != nil {
					return pickErr
				}
			} else {
				remotePath = resolved
			}
		} else {
			var pickErr error
			remotePath, pickErr = pickRemoteFile(client, srv.RemotePath)
			if pickErr != nil {
				return pickErr
			}
		}

		// Hedefin dosya mı dizin mi olduğunu anla
		entries, err := client.List(remotePath)
		isDir := err == nil && entries != nil // List başarılıysa dizindir

		label := lang.L.RemoteFileLabel
		if isDir {
			label = lang.L.RemoteDirLabel
			if remoteRmRecursive {
				label = lang.L.RemoteDirRecLabel
			}
		}

		fmt.Printf(lang.L.RemoteDeleteLabel, label, remotePath)

		if !remoteRmForce {
			fmt.Print(lang.L.RemoteDeleteConfirm)
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer != "y" && answer != "yes" && answer != "e" && answer != "evet" {
				fmt.Println(lang.L.RemoteDeleteCancel)
				return nil
			}
		}

		if isDir {
			if err := client.DeleteDir(remotePath, remoteRmRecursive); err != nil {
				return err
			}
		} else {
			if err := client.DeleteFile(remotePath); err != nil {
				return err
			}
		}

		fmt.Printf(lang.L.RemoteDeletedFmt, remotePath)
		return nil
	},
}

// ── cat ───────────────────────────────────────────────────────────────────────

var remoteCatMaxKB int

var remoteCatCmd = &cobra.Command{
	Use:   "cat [remote-path]",
	Short: "FTP'deki dosyanın içeriğini görüntüle",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		srv, client, err := connectRemote()
		if err != nil {
			return err
		}
		defer client.Close()

		var remotePath string
		if len(args) == 1 {
			resolved := resolveRemotePath(srv, args[0])
			// Dizin verilmişse interaktif seçici aç
			if entries, lsErr := client.List(resolved); lsErr == nil && entries != nil {
				remotePath, err = pickRemoteFile(client, resolved)
				if err != nil {
					return err
				}
			} else {
				remotePath = resolved
			}
		} else {
			remotePath, err = pickRemoteFile(client, srv.RemotePath)
			if err != nil {
				return err
			}
		}
		maxBytes := int64(remoteCatMaxKB) * 1024

		fmt.Printf("── %s ──\n", remotePath)

		data, err := client.Preview(remotePath, maxBytes)
		if err != nil {
			return err
		}

		fmt.Print(string(data))
		if !strings.HasSuffix(string(data), "\n") {
			fmt.Println()
		}
		if int64(len(data)) == maxBytes {
			fmt.Printf(lang.L.RemoteCatTruncFmt, remoteCatMaxKB, args[0])
		}
		return nil
	},
}

// ── yardımcılar ───────────────────────────────────────────────────────────────

// connectRemote yükler config, sunucuyu seçer ve bağlantı açar.
// --server verilmişse onu kullanır, tek sunucu varsa otomatik seçer,
// birden fazla sunucu varsa interaktif menü gösterir.
func connectRemote() (config.Server, *ftpclient.Client, error) {
	dir, _ := os.Getwd()
	cfg, err := config.Load(dir)
	if err != nil {
		return config.Server{}, nil, err
	}

	servers := cfg.EnabledServers()
	if len(servers) == 0 {
		return config.Server{}, nil, fmt.Errorf("aktif sunucu yok")
	}

	var srv config.Server

	if remoteServer != "" {
		for _, s := range servers {
			if s.Name == remoteServer {
				srv = s
				break
			}
		}
		if srv.Name == "" {
			return config.Server{}, nil, fmt.Errorf("sunucu bulunamadı: %q (mevcut: %v)", remoteServer, serverNames(servers))
		}
	} else if len(servers) == 1 {
		srv = servers[0]
	} else {
		srv, err = pickServer(servers)
		if err != nil {
			return config.Server{}, nil, err
		}
	}

	fmt.Printf(lang.L.RemoteConnecting, srv.Name, srv.Host)
	client, err := ftpclient.Connect(srv)
	if err != nil {
		return config.Server{}, nil, err
	}
	fmt.Println()
	return srv, client, nil
}

// pickServer interaktif TUI ile sunucu seçtirir.
func pickServer(servers []config.Server) (config.Server, error) {
	srv, err := pickServerTUI(servers)
	if err != nil {
		return config.Server{}, err
	}
	if srv == nil {
		return config.Server{}, fmt.Errorf("%s", lang.L.RemoteNoServerSel)
	}
	return *srv, nil
}

// pickRemoteFile shows an interactive numbered file browser on the FTP server.
// The user can navigate into directories or select a file.
func pickRemoteFile(client *ftpclient.Client, startPath string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	current := startPath

	for {
		entries, err := client.List(current)
		if err != nil {
			return "", fmt.Errorf(lang.L.RemotePickDirErr, err)
		}

		// Filtrele . ve ..
		var items []goftp.Entry
		for _, e := range entries {
			if e.Name != "." && e.Name != ".." {
				items = append(items, *e)
			}
		}

		fmt.Printf("\n  %s\n", current)
		fmt.Println("  ─────────────────────────")
		if current != "/" && current != startPath {
			fmt.Println(lang.L.RemotePickUpDir)
		}
		for i, e := range items {
			switch e.Type {
			case goftp.EntryTypeFolder:
				fmt.Printf("  [%d] %s/\n", i+1, e.Name)
			default:
				fmt.Printf("  [%d] %s  (%s)\n", i+1, e.Name, formatSize(e.Size))
			}
		}
		fmt.Printf(lang.L.RemotePickPromptFmt, len(items))

		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)

		if line == "q" || line == "Q" {
			return "", fmt.Errorf("iptal edildi")
		}

		if line == "0" {
			current = path.Dir(current)
			if current == "." {
				current = "/"
			}
			continue
		}

		var n int
		if _, err := fmt.Sscanf(line, "%d", &n); err != nil || n < 1 || n > len(items) {
			fmt.Printf(lang.L.RemotePickInvalid, line)
			continue
		}

		chosen := items[n-1]
		selected := path.Join(current, chosen.Name)

		if chosen.Type == goftp.EntryTypeFolder {
			current = selected
			continue
		}

		fmt.Printf(lang.L.RemotePickSelected, selected)
		return selected, nil
	}
}

// resolveRemotePath makes a path absolute relative to the server's remote_path.
func resolveRemotePath(srv config.Server, p string) string {
	if strings.HasPrefix(p, "/") {
		return p
	}
	return path.Join(srv.RemotePath, p)
}

func formatSize(b uint64) string {
	switch {
	case b >= 1024*1024:
		return fmt.Sprintf("%.1f MB", float64(b)/1024/1024)
	case b >= 1024:
		return fmt.Sprintf("%.1f KB", float64(b)/1024)
	default:
		return fmt.Sprintf("%d B", b)
	}
}

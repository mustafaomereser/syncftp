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

		fmt.Printf("Sunucu : %s (%s)\n", srv.Name, srv.Host)
		fmt.Printf("Dizin  : %s\n\n", remotePath)

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
					fmt.Printf("%s  ! listelenemedi: %v\n", indent, err)
				}
			}
		}
	}
	return nil
}

// ── get ───────────────────────────────────────────────────────────────────────

var remoteGetCmd = &cobra.Command{
	Use:   "get <remote-path> [local-dest]",
	Short: "FTP'den dosya indir",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		srv, client, err := connectRemote()
		if err != nil {
			return err
		}
		defer client.Close()

		remotePath := resolveRemotePath(srv, args[0])

		localDest := filepath.Base(args[0])
		if len(args) == 2 {
			localDest = args[1]
		}

		fmt.Printf("İndiriliyor: %s → %s\n", remotePath, localDest)
		if err := client.Download(remotePath, localDest); err != nil {
			return err
		}

		info, _ := os.Stat(localDest)
		if info != nil {
			fmt.Printf("✓ İndirildi (%s)\n", formatSize(uint64(info.Size())))
		} else {
			fmt.Println("✓ İndirildi")
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
	Use:   "rm <remote-path>",
	Short: "FTP'den dosya veya dizin sil",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		srv, client, err := connectRemote()
		if err != nil {
			return err
		}
		defer client.Close()

		remotePath := resolveRemotePath(srv, args[0])

		// Hedefin dosya mı dizin mi olduğunu anla
		entries, err := client.List(remotePath)
		isDir := err == nil && entries != nil // List başarılıysa dizindir

		label := "dosya"
		if isDir {
			label = "dizin"
			if remoteRmRecursive {
				label = "dizin (içeriğiyle)"
			}
		}

		fmt.Printf("Silinecek %s: %s\n", label, remotePath)

		if !remoteRmForce {
			fmt.Print("Silmek istediğinizden emin misiniz? [y/N]: ")
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer != "y" && answer != "yes" && answer != "e" && answer != "evet" {
				fmt.Println("İptal edildi.")
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

		fmt.Printf("✓ Silindi: %s\n", remotePath)
		return nil
	},
}

// ── cat ───────────────────────────────────────────────────────────────────────

var remoteCatMaxKB int

var remoteCatCmd = &cobra.Command{
	Use:   "cat <remote-path>",
	Short: "FTP'deki dosyanın içeriğini görüntüle",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		srv, client, err := connectRemote()
		if err != nil {
			return err
		}
		defer client.Close()

		remotePath := resolveRemotePath(srv, args[0])
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
			fmt.Printf("\n[İlk %d KB gösterildi — tamamını indirmek için: syncftp remote get %s]\n", remoteCatMaxKB, args[0])
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

	fmt.Printf("Bağlanıyor: %s (%s)...\n", srv.Name, srv.Host)
	client, err := ftpclient.Connect(srv)
	if err != nil {
		return config.Server{}, nil, err
	}
	fmt.Println()
	return srv, client, nil
}

// pickServer shows a numbered list and returns the server the user picks.
func pickServer(servers []config.Server) (config.Server, error) {
	fmt.Println("Sunucu seçin:")
	for i, s := range servers {
		fmt.Printf("  [%d] %-20s %s\n", i+1, s.Name, s.Host)
	}
	fmt.Printf("Seçim [1-%d]: ", len(servers))

	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)

	var n int
	if _, err := fmt.Sscanf(line, "%d", &n); err != nil || n < 1 || n > len(servers) {
		return config.Server{}, fmt.Errorf("geçersiz seçim: %q", line)
	}
	fmt.Println()
	return servers[n-1], nil
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

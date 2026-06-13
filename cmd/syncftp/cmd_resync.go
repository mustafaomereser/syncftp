package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"syncftp/internal/config"
	ftpclient "syncftp/internal/ftp"
	"syncftp/internal/ignore"
	"syncftp/internal/lang"
	"syncftp/internal/scanner"
	"syncftp/internal/state"
)

var (
	resyncFlagServer string
	resyncFlagAll    bool
)

func init() {
	resyncCmd.Flags().StringVarP(&resyncFlagServer, "server", "s", "", "Sadece bu sunucuyu resync et")
	resyncCmd.Flags().BoolVar(&resyncFlagAll, "all", false, "Tüm aktif sunucuları resync et")
	rootCmd.AddCommand(resyncCmd)
}

var resyncCmd = &cobra.Command{
	Use:   "resync",
	Short: "Yerel dosyaları FTP boyutlarıyla karşılaştırıp state'i günceller (yükleme yapmaz)",
	RunE:  runResyncCmd,
}

func runResyncCmd(cmd *cobra.Command, args []string) error {
	dir, _ := os.Getwd()
	cfg, err := config.Load(dir)
	if err != nil {
		return err
	}

	servers := cfg.Servers
	if resyncFlagServer != "" {
		var found []config.Server
		for _, s := range servers {
			if s.Name == resyncFlagServer {
				found = append(found, s)
				break
			}
		}
		if len(found) == 0 {
			fmt.Print(lang.L.ResyncNoServers)
			return nil
		}
		servers = found
	} else if !resyncFlagAll {
		// varsayılan: sadece enabled sunucular
		var enabled []config.Server
		for _, s := range servers {
			if s.Enabled {
				enabled = append(enabled, s)
			}
		}
		servers = enabled
	}

	if len(servers) == 0 {
		fmt.Print(lang.L.ResyncNoServers)
		return nil
	}

	for _, srv := range servers {
		runResync(dir, srv, cfg)
	}
	return nil
}

// runResync yerel dosyaları FTP boyutlarıyla karşılaştırıp eşleşenleri state'e yazar.
// Yükleme yapmaz. Hata olursa sessizce devam eder (state değişmez).
func runResync(dir string, srv config.Server, cfg *config.Config) {
	localPath := srv.EffectiveLocalPath(cfg.Project)
	localDir := filepath.Join(dir, localPath)

	matcher, _ := ignore.Load(localDir, cfg.Sync.IgnoreFiles)
	files, err := scanner.Scan(localDir, matcher)
	if err != nil {
		return
	}

	current := make(map[string]string)
	byKey := make(map[string]scanner.File)
	for _, f := range files {
		current[f.RelPath] = f.Hash
		byKey[f.RelPath] = f
	}

	st, _ := state.Load(dir, srv.Name)

	fmt.Printf("  [%s] ", srv.Name)
	fmt.Print(lang.L.ResyncScanning)

	client, err := ftpclient.Connect(srv)
	if err != nil {
		fmt.Printf(lang.L.ResyncConnErr, err)
		return
	}
	defer client.Close()

	remoteFiles, err := client.ListRecursive(srv.RemotePath)
	if err != nil {
		fmt.Printf(lang.L.ResyncListErr, err)
		return
	}

	fmt.Printf(lang.L.ResyncFoundFmt, len(remoteFiles))

	matched := 0
	different := 0
	for key, hash := range current {
		f := byKey[key]
		localInfo, err := os.Stat(f.AbsPath)
		if err != nil {
			different++
			continue
		}
		localSize := uint64(localInfo.Size())
		remoteSize, exists := remoteFiles[key]
		if exists && remoteSize == localSize {
			st.Files[key] = hash
			matched++
		} else {
			different++
		}
	}

	fmt.Printf(lang.L.ResyncMatchedFmt, matched, different)

	st.FirstSyncDone = true
	_ = state.Save(dir, st)

	fmt.Printf(lang.L.ResyncDoneFmt, srv.Name)
}

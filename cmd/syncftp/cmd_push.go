package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"syncftp/internal/config"
	"syncftp/internal/failed"
	ftpclient "syncftp/internal/ftp"
	"syncftp/internal/ignore"
	"syncftp/internal/lang"
	"syncftp/internal/release"
	"syncftp/internal/scanner"
	"syncftp/internal/state"
)

var (
	pushHost      string
	pushPort      int
	pushUser      string
	pushPass      string
	pushRemote    string
	pushPassive   bool
	pushServer    string
	pushFull      bool
	pushDryRun    bool
	pushInclude   []string
	pushExclude   []string
	pushMaxConn   int
	pushMaxRetry  int
)

func init() {
	pushCmd.Flags().StringVar(&pushHost, "host", "", "FTP sunucu adresi")
	pushCmd.Flags().IntVar(&pushPort, "port", 21, "FTP port")
	pushCmd.Flags().StringVar(&pushUser, "user", "", "Kullanıcı adı")
	pushCmd.Flags().StringVar(&pushPass, "pass", "", "Şifre")
	pushCmd.Flags().StringVar(&pushRemote, "remote", "/", "Uzak hedef dizin")
	pushCmd.Flags().BoolVar(&pushPassive, "passive", true, "Passive mode")
	pushCmd.Flags().StringVar(&pushServer, "server", "", "syncftp.json'dan sunucu adı (--host yerine)")
	pushCmd.Flags().BoolVar(&pushFull, "full", false, "State'i yoksay, tüm dosyaları yükle")
	pushCmd.Flags().BoolVar(&pushDryRun, "dry-run", false, "Ne yükleneceğini göster, yükleme")
	pushCmd.Flags().StringArrayVar(&pushInclude, "include", nil, "Sadece bu yol/klasörleri yükle")
	pushCmd.Flags().StringArrayVar(&pushExclude, "exclude", nil, "Bu yol/klasörleri hariç tut")
	pushCmd.Flags().IntVar(&pushMaxConn, "connections", 1, "Paralel bağlantı sayısı")
	pushCmd.Flags().IntVar(&pushMaxRetry, "retries", 2, "Hata durumunda yeniden deneme sayısı")
	rootCmd.AddCommand(pushCmd)
}

var pushCmd = &cobra.Command{
	Use:   "push <local-dizin>",
	Short: "Bir dizini doğrudan FTP'ye yükler (config gerektirmez)",
	Long: `Herhangi bir dizini FTP sunucusuna senkronize eder.
syncftp.json olmadan da çalışır. State <local-dizin>/.syncftp/ altında saklanır.

Örnekler:
  syncftp push ./website --host ftp.example.com --user admin --pass gizli --remote /public_html
  syncftp push ./website --server production          (varsa syncftp.json'dan okur)
  syncftp push ./website --host ftp.example.com --user admin --pass gizli --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: runPush,
}

func runPush(cmd *cobra.Command, args []string) error {
	localDir, err := filepath.Abs(args[0])
	if err != nil {
		return fmt.Errorf("dizin yolu çözülemedi: %w", err)
	}
	if _, err := os.Stat(localDir); err != nil {
		return fmt.Errorf("dizin bulunamadı: %s", localDir)
	}

	srv, err := resolvePushServer(localDir)
	if err != nil {
		return err
	}

	fmt.Printf(lang.L.PushSourceFmt, localDir)
	fmt.Printf(lang.L.PushTargetFmt, srv.Host, srv.Port, srv.RemotePath)
	fmt.Println()

	// Tara
	matcher, err := ignore.Load(localDir, nil)
	if err != nil {
		return fmt.Errorf("ignore dosyası yüklenemedi: %w", err)
	}
	fmt.Print(lang.L.PushScanning)
	files, err := scanner.Scan(localDir, matcher)
	if err != nil {
		return fmt.Errorf("tarama başarısız: %w", err)
	}
	fmt.Printf(lang.L.PushScannedFmt, len(files))

	current := make(map[string]string, len(files))
	byRel := make(map[string]scanner.File, len(files))
	for _, f := range files {
		current[f.RelPath] = f.Hash
		byRel[f.RelPath] = f
	}

	// State: localDir'in kendi .syncftp/ klasöründe tutulur
	stateDir := localDir
	st, err := state.Load(stateDir, srv.Name)
	if err != nil {
		return err
	}

	// Yüklenecekleri belirle
	var toUpload []string

	if !st.FirstSyncDone || pushFull {
		label := lang.L.PushFirstPush
		if pushFull && st.FirstSyncDone {
			label = lang.L.PushFullPush
		}
		fmt.Printf(lang.L.PushFirstFmt, label)
		toUpload = mapKeys(current)
	} else {
		diff := state.Diff(st, current)
		if len(diff.Deleted) > 0 {
			sort.Strings(diff.Deleted)
			fmt.Print(lang.L.PushDeletedHeader)
			for _, p := range diff.Deleted {
				fmt.Printf("      - %s\n", p)
			}
		}
		toUpload = append(diff.New, diff.Changed...)
	}

	if len(pushInclude) > 0 {
		toUpload = filterByInclude(toUpload, pushInclude)
	}
	toUpload = filterByExclude(toUpload, pushExclude)
	sort.Strings(toUpload)

	if len(toUpload) == 0 {
		fmt.Println(lang.L.PushNoChange)
		return nil
	}

	fmt.Printf(lang.L.PushProcessingFmt, len(toUpload))

	if pushDryRun {
		fmt.Println(lang.L.PushDryRunHeader)
		for _, rel := range toUpload {
			fmt.Printf("    %s\n", rel)
		}
		return nil
	}

	fmt.Printf(lang.L.PushConnFmt, srv.MaxConnections, srv.MaxRetries)

	pool, err := ftpclient.NewPool(srv)
	if err != nil {
		return err
	}
	defer pool.Close()

	var tasks []ftpclient.UploadTask
	for _, rel := range toUpload {
		f := byRel[rel]
		tasks = append(tasks, ftpclient.UploadTask{LocalPath: f.AbsPath, RelPath: rel, Hash: current[rel]})
	}

	results := pool.Upload(tasks)

	successFiles := make(map[string]string)
	uploaded, failedCount := 0, 0
	var failedPaths []string

	for _, r := range results {
		if r.Err != nil {
			if r.Attempts > 1 {
				fmt.Printf(lang.L.PushAttemptsFmt, r.RelPath, r.Attempts, r.Err)
			} else {
				fmt.Printf(lang.L.PushUploadErrFmt, r.RelPath, r.Err)
			}
			failedPaths = append(failedPaths, r.RelPath)
			failedCount++
		} else {
			if r.Attempts > 1 {
				fmt.Printf(lang.L.PushAttemptOkFmt, r.RelPath, r.Attempts)
			} else {
				fmt.Printf(lang.L.PushUploadOkFmt, r.RelPath)
			}
			successFiles[r.RelPath] = r.Hash
			uploaded++
		}
	}

	fmt.Printf(lang.L.PushDoneFmt, uploaded, failedCount)

	if len(failedPaths) > 0 {
		if err := failed.Save(stateDir, srv.Name, failedPaths); err == nil {
			fmt.Printf(lang.L.PushFailedHint, args[0], srv.Name)
		}
	} else {
		failed.Clear(stateDir, srv.Name)
	}

	// State güncelle
	for rel, hash := range successFiles {
		st.Files[rel] = hash
	}
	for rel := range st.Files {
		if _, exists := current[rel]; !exists {
			delete(st.Files, rel)
		}
	}
	st.FirstSyncDone = true
	if err := state.Save(stateDir, st); err != nil {
		fmt.Printf(lang.L.PushStateErr, err)
	}

	if len(successFiles) > 0 {
		if relDir, err := release.Create(stateDir, srv.Name, successFiles); err == nil {
			fmt.Printf(lang.L.PushReleaseFmt, relDir)
		}
	}

	return nil
}

// resolvePushServer önce --server flag'ine bakar (syncftp.json'dan okur),
// yoksa --host/--user/--pass ile geçici bir Server oluşturur.
func resolvePushServer(localDir string) (config.Server, error) {
	if pushServer != "" {
		cwd, _ := os.Getwd()
		cfg, err := config.Load(cwd)
		if err != nil {
			return config.Server{}, fmt.Errorf("syncftp.json okunamadı (--server için gerekli): %w", err)
		}
		for _, s := range cfg.EnabledServers() {
			if s.Name == pushServer {
				// CLI flag'lerini override olarak uygula
				if pushMaxConn > 0 {
					s.MaxConnections = pushMaxConn
				}
				if pushMaxRetry >= 0 {
					s.MaxRetries = pushMaxRetry
				}
				if len(pushInclude) > 0 {
					s.Include = pushInclude
				}
				if len(pushExclude) > 0 {
					s.Exclude = pushExclude
				}
				return s, nil
			}
		}
		return config.Server{}, fmt.Errorf("sunucu bulunamadı: %q", pushServer)
	}

	if pushHost == "" {
		return config.Server{}, fmt.Errorf("--host veya --server gerekli\n\nÖrnek:\n  syncftp push %s --host ftp.example.com --user admin --pass gizli --remote /public_html", localDir)
	}
	if pushUser == "" {
		return config.Server{}, fmt.Errorf("--user gerekli")
	}

	// Ad-hoc server — isim olarak host+remote kombinasyonu
	name := strings.NewReplacer("/", "_", ".", "-", ":", "-").Replace(pushHost + pushRemote)

	maxConn := pushMaxConn
	if maxConn <= 0 {
		maxConn = 1
	}
	maxRetry := pushMaxRetry
	if maxRetry < 0 {
		maxRetry = 2
	}

	return config.Server{
		Name:           name,
		Host:           pushHost,
		Port:           pushPort,
		User:           pushUser,
		Password:       pushPass,
		RemotePath:     pushRemote,
		Passive:        pushPassive,
		Enabled:        true,
		MaxConnections: maxConn,
		MaxRetries:     maxRetry,
	}, nil
}

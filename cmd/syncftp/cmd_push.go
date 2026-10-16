package main

import (
	"fmt"
	"io"
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
	pushHost       string
	pushPort       int
	pushUser       string
	pushPass       string
	pushRemote     string
	pushPassive    bool
	pushServer     string
	pushConnection string
	pushProtocol   string
	pushKey        string
	pushFull       bool
	pushDryRun     bool
	pushInclude    []string
	pushExclude    []string
	pushMaxConn    int
	pushMaxRetry   int
)

func init() {
	pushCmd.Flags().StringVar(&pushHost, "host", "", "Sunucu adresi")
	pushCmd.Flags().IntVar(&pushPort, "port", 0, "Port (varsayılan protokole göre 21/22)")
	pushCmd.Flags().StringVar(&pushUser, "user", "", "Kullanıcı adı")
	pushCmd.Flags().StringVar(&pushPass, "pass", "", "Şifre")
	pushCmd.Flags().StringVar(&pushProtocol, "protocol", "", "ftp | ftps | sftp (varsayılan ftp)")
	pushCmd.Flags().StringVar(&pushKey, "key", "", "SFTP özel anahtar dosyası")
	pushCmd.Flags().StringVar(&pushRemote, "remote", "/", "Uzak hedef dizin")
	pushCmd.Flags().BoolVar(&pushPassive, "passive", true, "Passive mode")
	pushCmd.Flags().StringVar(&pushServer, "server", "", "syncftp.json'dan sunucu adı (--host yerine)")
	pushCmd.Flags().StringVar(&pushConnection, "connection", "", "Kayıtlı bağlantıyı kullan (syncftp global conn list)")
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
	Short: "Bir dizini doğrudan sunucuya yükler (config gerektirmez)",
	Long: `Herhangi bir dizini FTP / FTPS / SFTP sunucusuna senkronize eder.
syncftp.json olmadan da çalışır. State <local-dizin>/.syncftp/ altında saklanır.

Örnekler:
  syncftp push ./website --host ftp.example.com --user admin --pass gizli --remote /public_html
  syncftp push ./website --connection hostingim --remote /public_html   (kayıtlı bağlantı)
  syncftp push ./website --protocol sftp --host example.com --user deploy --key ~/.ssh/id_ed25519 --remote /var/www
  syncftp push ./website --server production          (varsa syncftp.json'dan okur)
  syncftp push ./website --host ftp.example.com --user admin --pass gizli --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: runPush,
}

// pushOpts tek seferlik yükleme seçenekleri (CLI ve panel ortak).
type pushOpts struct {
	Full, DryRun     bool
	Include, Exclude []string
}

func runPush(cmd *cobra.Command, args []string) error {
	localDir, err := filepath.Abs(args[0])
	if err != nil {
		return fmt.Errorf("dizin yolu çözülemedi: %w", err)
	}
	srv, err := resolvePushServer(localDir)
	if err != nil {
		return err
	}
	return doPush(os.Stdout, localDir, srv, pushOpts{Full: pushFull, DryRun: pushDryRun, Include: pushInclude, Exclude: pushExclude})
}

// doPush config'siz yüklemenin çekirdeği: localDir taranır, state localDir/.syncftp/ altında tutulur.
func doPush(w io.Writer, localDir string, srv config.Server, o pushOpts) error {
	if st, err := os.Stat(localDir); err != nil || !st.IsDir() {
		return fmt.Errorf("dizin bulunamadı: %s", localDir)
	}
	fmt.Fprintf(w, lang.L.PushSourceFmt, localDir)
	fmt.Fprintf(w, lang.L.PushTargetFmt, srv.Host, srv.Port, srv.RemotePath)
	fmt.Fprintln(w)

	matcher, err := ignore.Load(localDir, nil)
	if err != nil {
		return fmt.Errorf("ignore dosyası yüklenemedi: %w", err)
	}
	fmt.Fprint(w, lang.L.PushScanning)
	files, err := scanner.Scan(localDir, matcher)
	if err != nil {
		return fmt.Errorf("tarama başarısız: %w", err)
	}
	fmt.Fprintf(w, lang.L.PushScannedFmt, len(files))

	current := make(map[string]string, len(files))
	byRel := make(map[string]scanner.File, len(files))
	for _, f := range files {
		current[f.RelPath] = f.Hash
		byRel[f.RelPath] = f
	}

	stateDir := localDir
	st, err := state.Load(stateDir, srv.Name)
	if err != nil {
		return err
	}

	var toUpload []string
	if !st.FirstSyncDone || o.Full {
		label := lang.L.PushFirstPush
		if o.Full && st.FirstSyncDone {
			label = lang.L.PushFullPush
		}
		fmt.Fprintf(w, lang.L.PushFirstFmt, label)
		toUpload = mapKeys(current)
	} else {
		diff := state.Diff(st, current)
		if len(diff.Deleted) > 0 {
			sort.Strings(diff.Deleted)
			fmt.Fprint(w, lang.L.PushDeletedHeader)
			for _, p := range diff.Deleted {
				fmt.Fprintf(w, "      - %s\n", p)
			}
		}
		toUpload = append(diff.New, diff.Changed...)
	}

	if len(o.Include) > 0 {
		toUpload = filterByInclude(toUpload, o.Include)
	}
	toUpload = filterByExclude(toUpload, o.Exclude)
	sort.Strings(toUpload)

	if len(toUpload) == 0 {
		fmt.Fprintln(w, lang.L.PushNoChange)
		return nil
	}
	fmt.Fprintf(w, lang.L.PushProcessingFmt, len(toUpload))

	if o.DryRun {
		fmt.Fprintln(w, lang.L.PushDryRunHeader)
		for _, rel := range toUpload {
			fmt.Fprintf(w, "    %s\n", rel)
		}
		return nil
	}

	fmt.Fprintf(w, lang.L.PushConnFmt, srv.MaxConnections, srv.MaxRetries)
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
				fmt.Fprintf(w, lang.L.PushAttemptsFmt, r.RelPath, r.Attempts, r.Err)
			} else {
				fmt.Fprintf(w, lang.L.PushUploadErrFmt, r.RelPath, r.Err)
			}
			failedPaths = append(failedPaths, r.RelPath)
			failedCount++
		} else {
			if r.Attempts > 1 {
				fmt.Fprintf(w, lang.L.PushAttemptOkFmt, r.RelPath, r.Attempts)
			} else {
				fmt.Fprintf(w, lang.L.PushUploadOkFmt, r.RelPath)
			}
			successFiles[r.RelPath] = r.Hash
			uploaded++
		}
	}
	fmt.Fprintf(w, lang.L.PushDoneFmt, uploaded, failedCount)

	if len(failedPaths) > 0 {
		if err := failed.Save(stateDir, srv.Name, failedPaths); err == nil {
			fmt.Fprintf(w, lang.L.PushFailedHint, localDir, srv.Name)
		}
	} else {
		failed.Clear(stateDir, srv.Name)
	}

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
		fmt.Fprintf(w, lang.L.PushStateErr, err)
	}
	if len(successFiles) > 0 {
		if relDir, err := release.Create(stateDir, srv.Name, successFiles); err == nil {
			fmt.Fprintf(w, lang.L.PushReleaseFmt, relDir)
		}
	}
	if failedCount > 0 {
		return fmt.Errorf(lang.L.PushFailedErrFmt, failedCount)
	}
	return nil
}

// adhocServerName host+uzak dizinden state dosyası için kararlı bir ad üretir.
func adhocServerName(host, remote string) string {
	return strings.NewReplacer("/", "_", ".", "-", ":", "-").Replace(host + remote)
}

// resolvePushServer önce --server flag'ine bakar (syncftp.json'dan okur), sonra --connection
// (kütüphane), yoksa --host/--user/--pass ile geçici bir Server oluşturur.
func resolvePushServer(localDir string) (config.Server, error) {
	if pushServer != "" {
		cwd, _ := os.Getwd()
		cfg, err := config.Load(cwd)
		if err != nil {
			return config.Server{}, fmt.Errorf("syncftp.json okunamadı (--server için gerekli): %w", err)
		}
		for _, s := range cfg.EnabledServers() {
			if s.Name == pushServer {
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
		return config.Server{}, fmt.Errorf(lang.L.ServerNotFoundFmt, pushServer)
	}

	srv := config.Server{Passive: pushPassive, Enabled: true}
	if pushConnection != "" {
		lib, err := config.LoadLibrary()
		if err != nil {
			return config.Server{}, err
		}
		c, ok := lib.FindConnection(pushConnection)
		if !ok {
			return config.Server{}, fmt.Errorf(lang.L.GlobalConnNotFoundFmt, pushConnection)
		}
		srv.ApplyConnection(c)
	} else if pushHost == "" {
		return config.Server{}, fmt.Errorf("--host, --connection veya --server gerekli\n\nÖrnek:\n  syncftp push %s --host ftp.example.com --user admin --pass gizli --remote /public_html", localDir)
	}
	if pushProtocol != "" {
		if !validProtocol(pushProtocol) {
			return config.Server{}, fmt.Errorf("--protocol: ftp | ftps | sftp")
		}
		srv.Protocol = strings.ToLower(pushProtocol)
	}
	if pushHost != "" {
		srv.Host = pushHost
	}
	if pushUser != "" {
		srv.User = pushUser
	}
	if pushPass != "" {
		srv.Password = pushPass
	}
	if pushKey != "" {
		srv.KeyFile = pushKey
	}
	if pushPort > 0 {
		srv.Port = pushPort
	}
	if srv.User == "" {
		return config.Server{}, fmt.Errorf("--user gerekli")
	}
	srv.RemotePath = pushRemote
	finishAdhocServer(&srv, pushMaxConn, pushMaxRetry)
	return srv, nil
}

// finishAdhocServer varsayılan port / ad / bağlantı ve deneme sayılarını tamamlar.
func finishAdhocServer(srv *config.Server, maxConn, maxRetry int) {
	if srv.Port == 0 {
		srv.Port = map[bool]int{true: 22, false: 21}[srv.IsSFTP()]
	}
	if srv.RemotePath == "" {
		srv.RemotePath = "/"
	}
	if srv.HasTunnel() && srv.TunnelPort == 0 {
		srv.TunnelPort = 22
	}
	srv.Name = adhocServerName(srv.Host, srv.RemotePath)
	srv.Enabled = true
	if maxConn <= 0 {
		maxConn = 1
	}
	if maxRetry < 0 {
		maxRetry = 2
	}
	srv.MaxConnections, srv.MaxRetries = maxConn, maxRetry
}

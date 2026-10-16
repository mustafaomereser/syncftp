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
	ftpclient "syncftp/internal/ftp"
	"syncftp/internal/ignore"
	"syncftp/internal/lang"
	"syncftp/internal/state"
)

// ── detect-limit ──────────────────────────────────────────────────────────────

var (
	dlFlagServer string
	dlFlagAll    bool
	dlFlagMax    int
	dlFlagNoSave bool
)

func init() {
	detectLimitCmd.Flags().StringVarP(&dlFlagServer, "server", "s", "", "Sunucu adı")
	detectLimitCmd.Flags().BoolVar(&dlFlagAll, "all", false, "Tüm aktif sunucular")
	detectLimitCmd.Flags().IntVar(&dlFlagMax, "max", 32, "Denenecek en fazla bağlantı")
	detectLimitCmd.Flags().BoolVar(&dlFlagNoSave, "no-save", false, "Sonucu config'e yazma, sadece göster")
	rootCmd.AddCommand(detectLimitCmd)

	verifyCmd.Flags().StringVarP(&vfFlagServer, "server", "s", "", "Sunucu adı")
	verifyCmd.Flags().BoolVar(&vfFlagAll, "all", false, "Tüm aktif sunucular")
	verifyCmd.Flags().BoolVar(&vfFlagFix, "fix", false, "Sapan dosyaları state'ten düş → sonraki sync yeniden yükler")
	rootCmd.AddCommand(verifyCmd)

	secureCmd.Flags().BoolVar(&secFlagOff, "off", false, "Anahtarlık modunu kapat, şifreleri tekrar config'e yaz")
	rootCmd.AddCommand(secureCmd)
}

var detectLimitCmd = &cobra.Command{
	Use:   "detect-limit",
	Short: "Sunucunun eş zamanlı bağlantı limitini ölçer ve max_connections'a yazar",
	Long: `Bağlantılar birer birer açılır (giriş + NOOP ile doğrulanır) ta ki sunucu reddedene
ya da --max'a ulaşılana kadar. Sonra hepsi tekrar yoklanır — fazla bağlantıyı kabul edip
sessizce düşüren sunucular da doğru ölçülür. Sonuç aynı hesabı (host+port+kullanıcı)
paylaşan TÜM sunucuların max_connections alanına yazılır.

Not: o anda açık başka bağlantılar (shell, FileZilla…) ölçülen değeri düşürür.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := os.Getwd()
		cfg, err := config.Load(dir)
		if err != nil {
			return err
		}
		servers, err := toolServers(cfg, dlFlagServer, dlFlagAll)
		if err != nil || servers == nil {
			return err
		}
		_, err = detectAndSave(os.Stdout, dir, cfg, servers, dlFlagMax, func(n int) {
			fmt.Printf(lang.L.DetectProgressFmt, n)
		}, !dlFlagNoSave)
		return err
	},
}

// detectLimits her farklı hesabı bir kez ölçer, sonucu o hesabı kullanan tüm sunuculara yazar.
// Config değiştiyse true döner (kaydetmek çağıranın işi).
func detectLimits(w io.Writer, cfg *config.Config, servers []config.Server, maxProbe int, progress func(int)) bool {
	done := map[string]int{}
	changed := false
	for _, srv := range servers {
		key := ftpclient.HostKey(srv)
		if _, ok := done[key]; ok {
			continue
		}
		fmt.Fprintf(w, lang.L.DetectStartFmt, srv.Name, srv.ProtocolName(), srv.Host)
		n, err := ftpclient.ProbeMaxConnections(srv, maxProbe, progress)
		if progress != nil {
			fmt.Fprint(w, "\r                                              \r")
		}
		if err != nil && n == 0 {
			fmt.Fprintf(w, lang.L.DetectErrFmt, err)
			continue
		}
		done[key] = n
		if n >= maxProbe {
			fmt.Fprintf(w, lang.L.DetectHitMaxFmt, n)
		} else {
			fmt.Fprintf(w, lang.L.DetectResultFmt, n)
		}
		for i := range cfg.Servers {
			if ftpclient.HostKey(cfg.Servers[i]) == key && cfg.Servers[i].MaxConnections != n {
				fmt.Fprintf(w, lang.L.DetectApplyFmt, cfg.Servers[i].Name, cfg.Servers[i].MaxConnections, n)
				cfg.Servers[i].MaxConnections = n
				changed = true
			}
		}
	}
	return changed
}

// toolServers --server / --all / seçici mantığı (detect-limit, verify).
// nil, nil → kullanıcı iptal etti.
func toolServers(cfg *config.Config, name string, all bool) ([]config.Server, error) {
	if name != "" {
		s, ok, err := resolveOneServer(cfg, name)
		if err != nil || !ok {
			return nil, err
		}
		return []config.Server{s}, nil
	}
	enabled := cfg.EnabledServers()
	if len(enabled) == 0 {
		return nil, fmt.Errorf("aktif sunucu yok")
	}
	if all || len(enabled) == 1 {
		return enabled, nil
	}
	sel, err := pickServerMultiTUI(enabled, "")
	if err != nil {
		return nil, err
	}
	if len(sel) == 0 {
		fmt.Println(lang.L.SyncCancelled)
		return nil, nil
	}
	return sel, nil
}

// ── verify (drift tespiti) ────────────────────────────────────────────────────

var (
	vfFlagServer string
	vfFlagAll    bool
	vfFlagFix    bool
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Sunucuyu state ile karşılaştırır: elle değiştirilen/silinen dosyaları bulur",
	Long: `State'e göre güncel olması gereken her dosyanın sunucudaki boyutunu yerel dosyayla
karşılaştırır (CRLF farkı toleranslı). Sunucuda eksik ya da boyutu farklı dosyalar
"sapma" olarak raporlanır — biri sunucuda elle değişiklik yapmış ya da dosya silinmiş demektir.

--fix sapan dosyaları state'ten düşer; sonraki sync onları yeniden yükler.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := os.Getwd()
		cfg, err := config.Load(dir)
		if err != nil {
			return err
		}
		servers, err := toolServers(cfg, vfFlagServer, vfFlagAll)
		if err != nil || servers == nil {
			return err
		}
		for _, srv := range servers {
			if _, err := verifyServer(os.Stdout, dir, cfg, srv, vfFlagFix); err != nil {
				fmt.Printf(lang.L.VerifyErrFmt, srv.Name, err)
			}
		}
		return nil
	},
}

type verifyReport struct {
	Server   string   `json:"server"`
	Checked  int      `json:"checked"`
	OK       int      `json:"ok"`
	Missing  []string `json:"missing"`
	Modified []string `json:"modified"`
	Pending  int      `json:"pending"` // yerelde değişmiş, henüz sync edilmemiş (sapma değil)
	Fixed    bool     `json:"fixed"`
}

// verifyServer tek sunucu için drift kontrolü (CLI + GUI ortak çekirdek).
func verifyServer(w io.Writer, configDir string, cfg *config.Config, srv config.Server, fix bool) (verifyReport, error) {
	rep := verifyReport{Server: srv.Name, Missing: []string{}, Modified: []string{}}
	current, byKey, err := scanForServer(configDir, cfg, srv)
	if err != nil {
		return rep, err
	}
	st, err := state.Load(configDir, srv.Name)
	if err != nil {
		return rep, err
	}

	localDir := filepath.Join(configDir, srv.EffectiveLocalPath(cfg.Project))
	flt := effectiveFilters(cfg, srv, nil, nil) // sync ile aynı filtreler (cmd_plan.go)
	matcher, _ := ignore.Load(localDir, cfg.Sync.IgnoreFiles)
	// Bekleyenler sync'in yükleyeceği listeyle aynı hesap: korunan / freeze'li değişiklikler "bekleyen" sayılmaz
	pi := pendingChanges(configDir, cfg, srv, current, st, nil, nil)
	rep.Pending = len(pi.newL) + len(pi.changed)

	fmt.Fprintf(w, lang.L.VerifyStartFmt, srv.Name, srv.ProtocolName(), srv.Host)
	client, err := ftpclient.Connect(srv)
	if err != nil {
		return rep, err
	}
	defer client.Close()
	skip := remoteSkipFunc(matcher, flt.include, flt.exclude)
	remote, err := client.ListRecursiveFiltered(srv.RemotePath, skip, nil)
	if err != nil {
		return rep, err
	}

	for rel, stHash := range st.Files {
		hash, ok := current[rel]
		if !ok {
			continue // yerelde silinmiş — sync zaten raporluyor
		}
		if hash != stHash {
			continue // yerelde değişmiş: sapma değil (yüklenecekse rep.Pending'de)
		}
		rep.Checked++
		rsize, exists := remote[rel]
		if !exists {
			rep.Missing = append(rep.Missing, rel)
			continue
		}
		f := byKey[rel]
		lsize := uint64(f.Size)
		if rsize == lsize {
			rep.OK++
			continue
		}
		if crlf := countCRLF(f.AbsPath); crlf > 0 && rsize == lsize-crlf {
			rep.OK++
			continue
		}
		rep.Modified = append(rep.Modified, rel)
	}
	sort.Strings(rep.Missing)
	sort.Strings(rep.Modified)

	printCapped := func(title string, list []string, mark string) {
		if len(list) == 0 {
			return
		}
		fmt.Fprintf(w, title, len(list))
		for i, p := range list {
			if i == 50 {
				fmt.Fprintf(w, lang.L.VerifyMoreFmt, len(list)-50)
				break
			}
			fmt.Fprintf(w, "      %s %s\n", mark, p)
		}
	}
	printCapped(lang.L.VerifyMissingFmt, rep.Missing, "✗")
	printCapped(lang.L.VerifyModifiedFmt, rep.Modified, "~")
	if len(rep.Missing)+len(rep.Modified) == 0 {
		fmt.Fprintf(w, lang.L.VerifyCleanFmt, rep.OK)
	} else {
		fmt.Fprintf(w, lang.L.VerifySummaryFmt, rep.OK, len(rep.Missing), len(rep.Modified))
	}
	if rep.Pending > 0 {
		fmt.Fprintf(w, lang.L.VerifyPendingFmt, rep.Pending)
	}

	if fix && len(rep.Missing)+len(rep.Modified) > 0 {
		for _, p := range rep.Missing {
			delete(st.Files, p)
		}
		for _, p := range rep.Modified {
			delete(st.Files, p)
		}
		if err := state.Save(configDir, st); err != nil {
			return rep, err
		}
		rep.Fixed = true
		fmt.Fprintf(w, lang.L.VerifyFixedFmt, len(rep.Missing)+len(rep.Modified))
	} else if len(rep.Missing)+len(rep.Modified) > 0 {
		fmt.Fprint(w, lang.L.VerifyFixHint)
	}
	return rep, nil
}

// ── secure (anahtarlık) ───────────────────────────────────────────────────────

var secFlagOff bool

var secureCmd = &cobra.Command{
	Use:   "secure",
	Short: "Şifreleri işletim sisteminin anahtarlığına taşır (Credential Manager / Keychain)",
	Long: `Açıkken config'deki tüm şifreler (sunucu, bağlantı profili, SSH tüneli) Windows Credential
Manager / macOS Keychain / Linux Secret Service'e yazılır; syncftp.json'da yalnızca "keyring:"
işareti kalır. Kayıtlar hesap bazlıdır (protokol://kullanıcı@host:port) — sunucu adından
bağımsızdır, aynı hesabı kullanan projeler paylaşır.

--off şifreleri tekrar düz metin olarak config'e yazar.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := os.Getwd()
		cfg, err := config.Load(dir)
		if err != nil {
			return err
		}
		return setKeyringMode(os.Stdout, dir, cfg, !secFlagOff)
	},
}

// setKeyringMode anahtarlık modunu açar/kapatır ve config'i kaydeder (CLI + GUI ortak).
func setKeyringMode(w io.Writer, dir string, cfg *config.Config, on bool) error {
	if len(cfg.SecretErrors) > 0 {
		return fmt.Errorf(lang.L.SecureUnresolvedFmt, strings.Join(cfg.SecretErrors, "; "))
	}
	cfg.Project.Keyring = on
	if err := config.Save(dir, cfg); err != nil {
		return err
	}
	if !on {
		fmt.Fprint(w, lang.L.SecureOff)
		return nil
	}
	// Doğrula: yeniden yükle, hâlâ düz metin şifre var mı (anahtarlığa yazılamadıysa)
	raw, _ := os.ReadFile(filepath.Join(dir, ".syncftp", "syncftp.json"))
	reloaded, err := config.Load(dir)
	if err != nil {
		return err
	}
	if len(reloaded.SecretErrors) > 0 {
		return fmt.Errorf(lang.L.SecureUnresolvedFmt, strings.Join(reloaded.SecretErrors, "; "))
	}
	markers := strings.Count(string(raw), `"keyring:"`)
	fmt.Fprintf(w, lang.L.SecureOnFmt, markers)
	return nil
}

// detectAndSave bağlantı limitini ölçer, değiştiyse config'e yazar (CLI, shell, panel ortak).
func detectAndSave(w io.Writer, configDir string, cfg *config.Config, servers []config.Server, maxProbe int, progress func(int), save bool) (bool, error) {
	if maxProbe <= 0 {
		maxProbe = 32
	}
	changed := detectLimits(w, cfg, servers, maxProbe, progress)
	if changed && save {
		if err := config.Save(configDir, cfg); err != nil {
			return changed, err
		}
		fmt.Fprint(w, lang.L.DetectSaved)
	}
	return changed, nil
}

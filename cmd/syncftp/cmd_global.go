package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"syncftp/internal/config"
	"syncftp/internal/lang"
	"syncftp/internal/projects"
)

// ══════════════════════════════════════════════════════════════════════════════
// syncftp global — proje bağımsız ayarlar (config.Library): kayıtlı bağlantılar ve
// yeni proje varsayılanları. GUI'de Ayarlar sayfasındaki aynı veriler.
// ══════════════════════════════════════════════════════════════════════════════

var (
	gcProtocol string
	gcHost     string
	gcPort     int
	gcUser     string
	gcPass     string
	gcKey      string
	gcActive   bool
)

func init() {
	globalCmd.AddCommand(globalShowCmd, globalConnCmd, globalDefaultsCmd)
	globalConnCmd.AddCommand(globalConnListCmd, globalConnAddCmd, globalConnRmCmd, globalConnImportCmd)
	globalDefaultsCmd.AddCommand(globalDefFromCmd, globalDefTemplateCmd, globalDefResetCmd)
	f := globalConnAddCmd.Flags()
	f.StringVar(&gcProtocol, "protocol", "ftp", "ftp | ftps | sftp")
	f.StringVar(&gcHost, "host", "", "Sunucu adresi")
	f.IntVar(&gcPort, "port", 0, "Port (varsayılan 21/22)")
	f.StringVar(&gcUser, "user", "", "Kullanıcı adı")
	f.StringVar(&gcPass, "pass", "", "Şifre (anahtarlık varsa oraya yazılır)")
	f.StringVar(&gcKey, "key", "", "SFTP özel anahtar dosyası")
	f.BoolVar(&gcActive, "active", false, "Aktif mod (varsayılan passive)")
	rootCmd.AddCommand(globalCmd)
}

var globalCmd = &cobra.Command{
	Use:   "global",
	Short: "Proje bağımsız ayarlar: kayıtlı bağlantılar ve yeni proje varsayılanları",
	Long: `Makine geneli syncFTP ayarları (` + "`<config dizini>/syncftp/library.json`" + `).

  Kayıtlı bağlantılar  init --connection / push --connection / panelde "kütüphaneden" ile kullanılır.
  Varsayılanlar        yeni oluşturulan her projeye uygulanan global sync ayarları.

Örnekler:
  syncftp global
  syncftp global conn add hostingim --host ftp.example.com --user admin --pass gizli
  syncftp global conn import ../eski-proje
  syncftp global defaults from .
  syncftp global defaults template laravel`,
	RunE: func(cmd *cobra.Command, args []string) error { return showGlobal() },
}

var globalShowCmd = &cobra.Command{Use: "show", Short: "Kütüphaneyi ve varsayılanları göster",
	RunE: func(cmd *cobra.Command, args []string) error { return showGlobal() }}

func showGlobal() error {
	lib, err := config.LoadLibrary()
	if err != nil {
		return err
	}
	fmt.Printf(lang.L.GlobalFileFmt, config.LibraryPath())
	fmt.Println()
	printLibConnections(lib)
	fmt.Println()
	d := lib.Defaults
	fmt.Println(lang.L.GlobalDefaultsHeader)
	row := func(k string, v any) { fmt.Printf("    %-18s %v\n", k, v) }
	list := func(v []string) string {
		if len(v) == 0 {
			return "—"
		}
		return strings.Join(v, ", ")
	}
	row("include", list(d.Sync.Include))
	row("exclude", list(d.Sync.Exclude))
	row("protect", list(d.Sync.Protect))
	row("block_extensions", list(d.Sync.BlockExtensions))
	row("ignore_files", list(d.Sync.IgnoreFiles))
	row("pre_sync", list(d.Sync.PreSync))
	row("post_sync", list(d.Sync.PostSync))
	bw := "∞"
	if d.Sync.BandwidthLimit > 0 {
		bw = fmt.Sprintf("%d KB/s", d.Sync.BandwidthLimit)
	}
	row("bandwidth", bw)
	row("backup", map[bool]string{true: "off", false: fmt.Sprintf("on (keep %d)", d.Sync.EffectiveBackupKeep())}[d.Sync.NoBackup])
	row("direct_upload", d.Sync.DirectUpload)
	row("keyring", d.Keyring)
	row("ignore_template", orDefault(d.IgnoreTemplate, "—"))
	row("max_connections", d.MaxConnections)
	row("max_retries", d.MaxRetries)
	if len(lib.SecretErrors) > 0 {
		fmt.Printf(lang.L.GlobalSecretWarnFmt, strings.Join(lib.SecretErrors, "; "))
	}
	return nil
}

func printLibConnections(lib *config.Library) {
	fmt.Printf(lang.L.GlobalConnHeaderFmt, len(lib.Connections))
	if len(lib.Connections) == 0 {
		fmt.Println(lang.L.GlobalConnNone)
		return
	}
	for _, c := range lib.Connections {
		auth := "****"
		if c.Password == "" {
			auth = "—"
		}
		if c.KeyFile != "" {
			auth = "key " + c.KeyFile
		}
		fmt.Printf("    %-20s %s://%s@%s:%d  %s\n", c.Name, strings.ToLower(orDefault(c.Protocol, "ftp")), c.User, c.Host, c.Port, auth)
	}
}

var globalConnCmd = &cobra.Command{Use: "conn", Aliases: []string{"connections"}, Short: "Kayıtlı bağlantılar",
	RunE: func(cmd *cobra.Command, args []string) error {
		lib, err := config.LoadLibrary()
		if err != nil {
			return err
		}
		printLibConnections(lib)
		return nil
	}}

var globalConnListCmd = &cobra.Command{Use: "list", Short: "Kayıtlı bağlantıları listele",
	RunE: func(cmd *cobra.Command, args []string) error { return globalConnCmd.RunE(cmd, args) }}

var globalConnAddCmd = &cobra.Command{Use: "add <ad>", Short: "Bağlantı ekle / güncelle", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if gcHost == "" {
			return fmt.Errorf("--host")
		}
		if !validProtocol(gcProtocol) {
			return fmt.Errorf("--protocol: ftp | ftps | sftp")
		}
		lib, err := config.LoadLibrary()
		if err != nil {
			return err
		}
		c := config.Connection{Name: args[0], Protocol: strings.ToLower(gcProtocol), Host: gcHost, Port: gcPort,
			User: gcUser, Password: gcPass, KeyFile: gcKey, Passive: !gcActive}
		if c.Port == 0 {
			c.Port = config.DefaultPort(c.Protocol)
		}
		replaced := upsertLibConn(lib, c)
		if err := lib.ValidateConnections(); err != nil {
			return err
		}
		if err := config.SaveLibrary(lib); err != nil {
			return err
		}
		fmt.Printf(map[bool]string{true: lang.L.GlobalConnUpdatedFmt, false: lang.L.GlobalConnAddedFmt}[replaced], c.Name)
		return nil
	}}

// upsertLibConn aynı adlı bağlantıyı değiştirir ya da sona ekler; değiştirdiyse true.
func upsertLibConn(lib *config.Library, c config.Connection) bool {
	for i := range lib.Connections {
		if strings.EqualFold(lib.Connections[i].Name, c.Name) {
			lib.Connections[i] = c
			return true
		}
	}
	lib.Connections = append(lib.Connections, c)
	return false
}

var globalConnRmCmd = &cobra.Command{Use: "rm <ad>", Short: "Bağlantıyı sil", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		lib, err := config.LoadLibrary()
		if err != nil {
			return err
		}
		out := lib.Connections[:0]
		found := false
		for _, c := range lib.Connections {
			if strings.EqualFold(c.Name, args[0]) {
				found = true
				continue
			}
			out = append(out, c)
		}
		if !found {
			return fmt.Errorf(lang.L.GlobalConnNotFoundFmt, args[0])
		}
		lib.Connections = out
		if err := config.SaveLibrary(lib); err != nil {
			return err
		}
		fmt.Printf(lang.L.GlobalConnRemovedFmt, args[0])
		return nil
	}}

// loadProjectArg ad / yol / boş (bu dizin) ile bir projenin config'ini yükler.
func loadProjectArg(args []string) (string, *config.Config, error) {
	dir, _ := os.Getwd()
	if len(args) > 0 && args[0] != "" {
		if p, ok := projects.Find(args[0]); ok {
			dir = p.Path
		} else if abs, err := filepath.Abs(args[0]); err == nil {
			dir = abs
		}
	}
	cfg, err := config.Load(dir)
	return dir, cfg, err
}

// importConnections projedeki bağlantı profillerini ve profil kullanmayan sunucuların
// kimlik bilgilerini kütüphaneye ekler (aynı host+user+port zaten varsa atlanır). Eklenen sayısı.
func importConnections(lib *config.Library, cfg *config.Config) int {
	key := config.Connection.AccountKey
	have := map[string]bool{}
	names := map[string]bool{}
	for _, c := range lib.Connections {
		have[key(c)] = true
		names[strings.ToLower(c.Name)] = true
	}
	uniq := func(n string) string {
		base, i := n, 2
		for names[strings.ToLower(n)] {
			n = fmt.Sprintf("%s-%d", base, i)
			i++
		}
		names[strings.ToLower(n)] = true
		return n
	}
	added := 0
	add := func(c config.Connection) {
		if c.Host == "" || have[key(c)] {
			return
		}
		have[key(c)] = true
		c.Name = uniq(c.Name)
		lib.Connections = append(lib.Connections, c)
		added++
	}
	for _, c := range cfg.Connections {
		add(c)
	}
	for _, s := range cfg.Servers {
		if s.Connection != "" {
			continue
		}
		add(config.ConnectionFromServer(cfg.Project.Name+"-"+s.Name, s))
	}
	return added
}

var globalConnImportCmd = &cobra.Command{Use: "import [proje]", Short: "Bir projenin bağlantılarını kütüphaneye aktar", Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, cfg, err := loadProjectArg(args)
		if err != nil {
			return err
		}
		lib, err := config.LoadLibrary()
		if err != nil {
			return err
		}
		n := importConnections(lib, cfg)
		if n > 0 {
			if err := config.SaveLibrary(lib); err != nil {
				return err
			}
		}
		fmt.Printf(lang.L.GlobalConnImportedFmt, n, cfg.Project.Name)
		return nil
	}}

var globalDefaultsCmd = &cobra.Command{Use: "defaults", Short: "Yeni proje varsayılanları",
	RunE: func(cmd *cobra.Command, args []string) error { return showGlobal() }}

var globalDefFromCmd = &cobra.Command{Use: "from [proje]", Short: "Bir projenin global sync ayarlarını varsayılan yap", Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, cfg, err := loadProjectArg(args)
		if err != nil {
			return err
		}
		lib, err := config.LoadLibrary()
		if err != nil {
			return err
		}
		defaultsFromProject(&lib.Defaults, cfg)
		if err := config.SaveLibrary(lib); err != nil {
			return err
		}
		fmt.Printf(lang.L.GlobalDefaultsFromFmt, cfg.Project.Name)
		return nil
	}}

// defaultsFromProject projenin global sync ayarlarını (ve ilk sunucunun bağlantı/deneme sayısını) varsayılana kopyalar.
func defaultsFromProject(d *config.Defaults, cfg *config.Config) {
	d.Sync = cfg.Sync.Clone()
	d.FirstSyncFull = cfg.FirstSync.Full
	d.Keyring = cfg.Project.Keyring
	if len(cfg.Servers) > 0 {
		d.MaxConnections = cfg.Servers[0].MaxConnections
		d.MaxRetries = cfg.Servers[0].MaxRetries
	}
}

var globalDefTemplateCmd = &cobra.Command{Use: "template <laravel|wordpress|node|generic|none>", Short: "Varsayılan ignore şablonu", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		t := strings.ToLower(args[0])
		if _, ok := ignoreTemplates[t]; !ok && t != "none" {
			return fmt.Errorf("laravel | wordpress | node | generic | none")
		}
		lib, err := config.LoadLibrary()
		if err != nil {
			return err
		}
		lib.Defaults.IgnoreTemplate = map[bool]string{true: "", false: t}[t == "none"]
		if err := config.SaveLibrary(lib); err != nil {
			return err
		}
		fmt.Printf(lang.L.GlobalTemplateSetFmt, t)
		return nil
	}}

var globalDefResetCmd = &cobra.Command{Use: "reset", Short: "Varsayılanları fabrika ayarına döndür",
	RunE: func(cmd *cobra.Command, args []string) error {
		lib, err := config.LoadLibrary()
		if err != nil {
			return err
		}
		lib.Defaults = config.BuiltinDefaults()
		if err := config.SaveLibrary(lib); err != nil {
			return err
		}
		fmt.Println(lang.L.GlobalDefaultsReset)
		return nil
	}}

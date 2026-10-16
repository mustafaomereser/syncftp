package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"syncftp/internal/config"
	"syncftp/internal/lang"
	"syncftp/internal/projects"
)

var (
	initYes        bool
	initName       string
	initLocal      string
	initServerName string
	initProtocol   string
	initHost       string
	initPort       int
	initUser       string
	initPass       string
	initKey        string
	initRemote     string
	initConnection string
	initTemplate   string
	initNoServer   bool
)

func init() {
	f := initCmd.Flags()
	f.BoolVarP(&initYes, "yes", "y", false, "Soru sormadan oluştur (flag'ler + varsayılanlar)")
	f.StringVar(&initName, "name", "", "Proje adı (varsayılan: klasör adı)")
	f.StringVar(&initLocal, "local", "", "Varsayılan yerel dizin (varsayılan: .)")
	f.StringVar(&initServerName, "server-name", "", "İlk sunucunun adı (varsayılan: production)")
	f.StringVar(&initProtocol, "protocol", "", "ftp | ftps | sftp")
	f.StringVar(&initHost, "host", "", "Sunucu adresi")
	f.IntVar(&initPort, "port", 0, "Port (varsayılan protokole göre 21/22)")
	f.StringVar(&initUser, "user", "", "Kullanıcı adı")
	f.StringVar(&initPass, "pass", "", "Şifre")
	f.StringVar(&initKey, "key", "", "SFTP özel anahtar dosyası")
	f.StringVar(&initRemote, "remote", "", "Uzak dizin (varsayılan: /public_html)")
	f.StringVar(&initConnection, "connection", "", "Kayıtlı bağlantıyı kullan (syncftp global conn list)")
	f.StringVar(&initTemplate, "template", "", "Ignore şablonu: laravel | wordpress | node | generic | none")
	f.BoolVar(&initNoServer, "no-server", false, "Sunucusuz oluştur (sonra eklenir)")
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "syncftp.json dosyasını oluşturur",
	Long: `Bu dizinde syncFTP projesi oluşturur. Argümansız: interaktif sihirbaz.
Global varsayılanlar (syncftp global) yeni projeye uygulanır; kayıtlı bağlantılardan seçilebilir.

Örnekler:
  syncftp init
  syncftp init -y --connection hostingim --remote /public_html --template laravel
  syncftp init -y --protocol sftp --host example.com --user deploy --key ~/.ssh/id_ed25519 --remote /var/www
  syncftp init -y --no-server`,
	RunE: runInit,
}

// newProjectOpts createProject girdileri. Template: "" → global varsayılan, "none" → şablon yok.
type newProjectOpts struct {
	Name      string
	LocalPath string
	Template  string
	Server    *config.Server
}

// createProject dizinde yeni proje config'i oluşturur: global varsayılanlar + (varsa) ilk sunucu,
// ignore şablonu, .gitignore girdileri ve makine geneli proje kaydı. CLI init ve GUI ortak kullanır.
func createProject(w io.Writer, dir string, o newProjectOpts) (*config.Config, string, error) {
	if projects.HasConfig(dir) {
		return nil, "", fmt.Errorf(lang.L.InitExistsFmt, filepath.Join(dir, ".syncftp", "syncftp.json"))
	}
	lib, _ := config.LoadLibrary()
	name := strings.TrimSpace(o.Name)
	if name == "" {
		name = filepath.Base(dir)
	}
	local := strings.TrimSpace(o.LocalPath)
	if local == "" {
		local = "."
	}
	cfg := &config.Config{Project: config.Project{Name: name, DefaultPath: local}, Servers: []config.Server{}}
	lib.Defaults.ApplyDefaults(cfg)
	if o.Server != nil {
		cfg.Servers = append(cfg.Servers, *o.Server)
	}
	if err := config.Save(dir, cfg); err != nil {
		return nil, "", err
	}
	fmt.Fprint(w, lang.L.InitCreated+"\n")

	tmpl := o.Template
	if tmpl == "" {
		tmpl = lib.Defaults.IgnoreTemplate
	}
	if tmpl != "" && tmpl != "none" {
		writeIgnoreTemplateFile(w, filepath.Join(dir, ".syncftp"), tmpl)
	}
	addToIgnoreFile(w, dir)
	reg := name
	if p, err := projects.Add(dir, name); err == nil {
		reg = p.Name
	}
	return cfg, reg, nil
}

// initServerFromFlags flag'lerden (ve varsa kayıtlı bağlantıdan) ilk sunucuyu kurar; host yoksa nil.
func initServerFromFlags(lib *config.Library) (*config.Server, error) {
	if initNoServer {
		return nil, nil
	}
	name := initServerName
	if name == "" {
		name = "production"
	}
	srv := lib.Defaults.NewServer(name)
	if initConnection != "" {
		c, ok := lib.FindConnection(initConnection)
		if !ok {
			return nil, fmt.Errorf(lang.L.GlobalConnNotFoundFmt, initConnection)
		}
		srv.ApplyConnection(c)
	}
	if initProtocol != "" {
		if !validProtocol(initProtocol) {
			return nil, fmt.Errorf("--protocol: ftp | ftps | sftp")
		}
		srv.Protocol = strings.ToLower(initProtocol)
	}
	if initHost != "" {
		srv.Host = initHost
	}
	if initUser != "" {
		srv.User = initUser
	}
	if initPass != "" {
		srv.Password = initPass
	}
	if initKey != "" {
		srv.KeyFile = initKey
	}
	if initPort > 0 {
		srv.Port = initPort
	}
	if srv.Port == 0 {
		srv.Port = map[bool]int{true: 22, false: 21}[srv.IsSFTP()]
	}
	srv.RemotePath = initRemote
	if srv.RemotePath == "" {
		srv.RemotePath = "/public_html"
	}
	if srv.Host == "" {
		return nil, nil
	}
	return &srv, nil
}

func validProtocol(p string) bool {
	for _, x := range config.Protocols {
		if strings.EqualFold(x, p) {
			return true
		}
	}
	return false
}

func runInit(cmd *cobra.Command, args []string) error {
	dir, _ := os.Getwd()
	lib, err := config.LoadLibrary()
	if err != nil {
		fmt.Printf(lang.L.GlobalLoadWarnFmt, err)
	}

	// Etkileşimsiz: -y ya da sunucu bilgisi flag'le verilmişse
	if initYes || initHost != "" || initConnection != "" || initNoServer {
		srv, err := initServerFromFlags(lib)
		if err != nil {
			return err
		}
		tmpl := initTemplate
		if tmpl != "" && tmpl != "none" {
			if _, ok := ignoreTemplates[tmpl]; !ok {
				return fmt.Errorf("--template: laravel | wordpress | node | generic | none")
			}
		}
		_, reg, err := createProject(os.Stdout, dir, newProjectOpts{Name: initName, LocalPath: initLocal, Template: tmpl, Server: srv})
		if err != nil {
			return err
		}
		fmt.Printf(lang.L.InitReadyFmt, reg)
		return nil
	}

	if projects.HasConfig(dir) {
		return fmt.Errorf(lang.L.InitExistsFmt, filepath.Join(dir, ".syncftp", "syncftp.json"))
	}

	r := bufio.NewReader(os.Stdin)
	ask := func(prompt, def string) string {
		if def != "" {
			fmt.Printf("%s [%s]: ", prompt, def)
		} else {
			fmt.Printf("%s: ", prompt)
		}
		line, _ := r.ReadString('\n')
		line = strings.TrimSpace(line)
		if line == "" {
			return def
		}
		return line
	}

	fmt.Println(lang.L.InitWizardTitle)
	fmt.Println()

	projectName := ask(lang.L.InitProjectName, filepath.Base(dir))
	localPath := ask(lang.L.InitLocalDir, ".")

	fmt.Println()
	fmt.Println(lang.L.InitFTPHeader)
	serverName := ask(lang.L.InitServerName, "production")
	srv := lib.Defaults.NewServer(serverName)

	// Kayıtlı bağlantı varsa önce seçtir
	usedConn := false
	if len(lib.Connections) > 0 {
		items := make([]PickerItem, 0, len(lib.Connections)+1)
		for _, c := range lib.Connections {
			items = append(items, PickerItem{Icon: "🔗", Label: c.Name, Value: c.Name,
				Desc: fmt.Sprintf("%s://%s@%s:%d", strings.ToLower(orDefault(c.Protocol, "ftp")), c.User, c.Host, c.Port)})
		}
		items = append(items, PickerItem{Icon: "✎", Label: lang.L.InitConnManual, Value: ""})
		if choice, err := RunPicker(lang.L.InitConnPickTitle, lang.L.InitConnPickSub, items); err == nil && choice != "" {
			if c, ok := lib.FindConnection(choice); ok {
				srv.ApplyConnection(c)
				usedConn = true
				fmt.Printf(lang.L.InitUsingConnFmt, c.Name, c.User, c.Host, c.Port)
			}
		}
	}
	if !usedConn {
		srv.Host = ask(lang.L.InitHost, "")
		port := ask(lang.L.InitPort, "21")
		srv.User = ask(lang.L.InitUser, "")
		srv.Password = ask(lang.L.InitPassword, "")
		fmt.Sscanf(port, "%d", &srv.Port)
		if srv.Port == 0 {
			srv.Port = 21
		}
	}
	srv.RemotePath = ask(lang.L.InitRemotePath, "/public_html")

	tmpl := pickIgnoreTemplate(lib.Defaults.IgnoreTemplate)
	_, reg, err := createProject(os.Stdout, dir, newProjectOpts{Name: projectName, LocalPath: localPath, Template: tmpl, Server: &srv})
	if err != nil {
		return err
	}
	fmt.Println()
	fmt.Printf(lang.L.InitReadyFmt, reg)
	return nil
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

// ignoreTemplates proje tipine göre hazır .syncftp/syncftp.ignore içerikleri.
var ignoreTemplates = map[string]string{
	"laravel": `# Laravel
/vendor/
/node_modules/
/storage/*.key
/storage/logs/
/storage/framework/cache/
/storage/framework/sessions/
/storage/framework/views/
/bootstrap/cache/
.env
.env.*
/public/hot
/public/storage
npm-debug.log
yarn-error.log
/.phpunit.cache
`,
	"wordpress": `# WordPress
wp-config.php
/wp-content/uploads/
/wp-content/cache/
/wp-content/upgrade/
/wp-content/backup*/
*.log
.htaccess.bak
`,
	"node": `# Node.js
node_modules/
dist/
build/
.env
.env.*
npm-debug.log*
yarn-debug.log*
yarn-error.log*
.cache/
coverage/
`,
	"generic": `# Genel
.env
.env.*
*.log
*.tmp
*.bak
node_modules/
vendor/
.DS_Store
Thumbs.db
`,
}

// pickIgnoreTemplate proje tipini sorar; global varsayılan şablon listede önce gelir.
// Dönüş: şablon adı, "none" (yok) — iptalde de "none".
func pickIgnoreTemplate(def string) string {
	items := []PickerItem{
		{Icon: "🅻", Label: "Laravel", Value: "laravel", Desc: "vendor/, storage/, .env..."},
		{Icon: "🆆", Label: "WordPress", Value: "wordpress", Desc: "wp-config.php, uploads/, cache/..."},
		{Icon: "🅽", Label: "Node.js", Value: "node", Desc: "node_modules/, dist/, .env..."},
		{Icon: "·", Label: lang.L.InitTemplateGeneric, Value: "generic", Desc: ".env, *.log, node_modules/..."},
		{Icon: "✕", Label: lang.L.InitTemplateNone, Value: "none", Desc: lang.L.InitTemplateNoneDesc},
	}
	for i, it := range items {
		if it.Value == def && i > 0 {
			items[0], items[i] = items[i], items[0]
			items[0].Desc += "  ★"
			break
		}
	}
	choice, err := RunPicker(lang.L.InitTemplateTitle, lang.L.InitTemplateSub, items)
	if err != nil || choice == "" {
		return "none"
	}
	return choice
}

// writeIgnoreTemplateFile şablonu .syncftp/syncftp.ignore'a yazar (mevcut dosyanın üzerine yazmaz).
func writeIgnoreTemplateFile(w io.Writer, syncftpDir, name string) {
	tmpl, ok := ignoreTemplates[name]
	if !ok {
		return
	}
	ignorePath := filepath.Join(syncftpDir, "syncftp.ignore")
	if _, statErr := os.Stat(ignorePath); statErr == nil {
		return
	}
	if os.WriteFile(ignorePath, []byte(tmpl), 0644) == nil {
		fmt.Fprintf(w, lang.L.InitTemplateWrittenFmt, name)
	}
}

// addToIgnoreFile .gitignore'a .syncftp, syncftp.exe ve syncftp girdilerini ekler.
func addToIgnoreFile(w io.Writer, dir string) {
	entries := []string{".syncftp", "syncftp.exe", "syncftp"}
	gitignorePath := filepath.Join(dir, ".gitignore")

	content, _ := os.ReadFile(gitignorePath)
	existing := string(content)

	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	added := false
	for _, entry := range entries {
		if !strings.Contains(existing, entry) {
			if len(existing) > 0 && !strings.HasSuffix(existing, "\n") && !added {
				fmt.Fprintln(f)
			}
			fmt.Fprintln(f, entry)
			added = true
		}
	}
	if added {
		fmt.Fprintf(w, lang.L.InitIgnoreAdded, ".gitignore")
	} else {
		fmt.Fprintf(w, lang.L.InitIgnoreExists, ".gitignore")
	}
}

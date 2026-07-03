package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"syncftp/internal/lang"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "syncftp.json dosyasını oluşturur",
	Long:  "İnteraktif sihirbaz ile FTP sunucu bilgilerini girerek syncftp.json oluşturur.",
	RunE:  runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
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

	dir, _ := os.Getwd()
	projectName := ask(lang.L.InitProjectName, filepath.Base(dir))
	localPath := ask(lang.L.InitLocalDir, ".")

	fmt.Println()
	fmt.Println(lang.L.InitFTPHeader)
	serverName := ask(lang.L.InitServerName, "production")
	host := ask(lang.L.InitHost, "")
	port := ask(lang.L.InitPort, "21")
	user := ask(lang.L.InitUser, "")
	password := ask(lang.L.InitPassword, "")
	remotePath := ask(lang.L.InitRemotePath, "/public_html")

	portNum := 21
	fmt.Sscanf(port, "%d", &portNum)

	cfg := map[string]any{
		"project": map[string]any{
			"name":         projectName,
			"default_path": localPath,
		},
		"sync": map[string]any{
			"protect": []string{".env"},
			"include": []string{},
			"exclude": []string{},
		},
		"first_sync": map[string]any{
			"full": true,
		},
		"servers": []map[string]any{
			{
				"name":        serverName,
				"host":        host,
				"port":        portNum,
				"user":        user,
				"password":    password,
				"remote_path": remotePath,
				"passive":         true,
				"enabled":         true,
				"max_connections": 1,
				"max_retries":     2,
				"include":         []string{},
				"exclude":         []string{},
			},
		},
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON oluşturulamadı: %w", err)
	}
	syncftpDir := filepath.Join(dir, ".syncftp")
	if err := os.MkdirAll(syncftpDir, 0700); err != nil {
		return fmt.Errorf(".syncftp dizini oluşturulamadı: %w", err)
	}
	if err := os.WriteFile(filepath.Join(syncftpDir, "syncftp.json"), data, 0600); err != nil {
		return fmt.Errorf("syncftp.json yazılamadı: %w", err)
	}
	fmt.Println(lang.L.InitCreated)

	writeIgnoreTemplate(syncftpDir)
	addToIgnoreFile(dir)

	fmt.Println()
	fmt.Printf(lang.L.InitReadyFmt, projectName)
	return nil
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

// writeIgnoreTemplate proje tipini sorup .syncftp/syncftp.ignore şablonu yazar.
func writeIgnoreTemplate(syncftpDir string) {
	choice, err := RunPicker(lang.L.InitTemplateTitle, lang.L.InitTemplateSub, []PickerItem{
		{Icon: "🅻", Label: "Laravel", Value: "laravel", Desc: "vendor/, storage/, .env..."},
		{Icon: "🆆", Label: "WordPress", Value: "wordpress", Desc: "wp-config.php, uploads/, cache/..."},
		{Icon: "🅽", Label: "Node.js", Value: "node", Desc: "node_modules/, dist/, .env..."},
		{Icon: "·", Label: lang.L.InitTemplateGeneric, Value: "generic", Desc: ".env, *.log, node_modules/..."},
		{Icon: "✕", Label: lang.L.InitTemplateNone, Value: "", Desc: lang.L.InitTemplateNoneDesc},
	})
	if err != nil || choice == "" {
		return
	}
	tmpl, ok := ignoreTemplates[choice]
	if !ok {
		return
	}
	ignorePath := filepath.Join(syncftpDir, "syncftp.ignore")
	if _, statErr := os.Stat(ignorePath); statErr == nil {
		return // mevcut dosyanın üzerine yazma
	}
	if writeErr := os.WriteFile(ignorePath, []byte(tmpl), 0644); writeErr == nil {
		fmt.Printf(lang.L.InitTemplateWrittenFmt, choice)
	}
}

// addToIgnoreFile adds .syncftp, syncftp.exe and syncftp to .gitignore.
func addToIgnoreFile(dir string) {
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
			fmt.Fprintln(f, entry)
			added = true
		}
	}
	if added {
		fmt.Printf(lang.L.InitIgnoreAdded, ".gitignore")
	} else {
		fmt.Printf(lang.L.InitIgnoreExists, ".gitignore")
	}
}

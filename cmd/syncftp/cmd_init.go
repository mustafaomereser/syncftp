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
	if err := os.WriteFile("syncftp.json", data, 0600); err != nil {
		return fmt.Errorf("syncftp.json yazılamadı: %w", err)
	}
	fmt.Println(lang.L.InitCreated)

	addToIgnoreFile(dir)

	fmt.Println()
	fmt.Printf(lang.L.InitReadyFmt, projectName)
	return nil
}

// addToIgnoreFile adds syncftp.json and syncftp.exe to .gitignore or syncftp.ignore.
// If neither exists, creates syncftp.ignore.
func addToIgnoreFile(dir string) {
	block := "\n# syncFTP — do not commit\nsyncftp.json\nsyncftp.exe\n\n# syncFTP runtime\n.syncftp/\n.git/\n"

	for _, name := range []string{".gitignore", "syncftp.ignore"} {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err == nil {
			content, _ := os.ReadFile(p)
			if strings.Contains(string(content), "syncftp.json") {
				fmt.Printf(lang.L.InitIgnoreExists, name)
				return
			}
			f, err := os.OpenFile(p, os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				return
			}
			defer f.Close()
			fmt.Fprint(f, block)
			fmt.Printf(lang.L.InitIgnoreAdded, name)
			return
		}
	}

	content := "# syncFTP — do not commit\nsyncftp.json\nsyncftp.exe\n\n# syncFTP runtime\n.syncftp/\n.git/\n"
	if err := os.WriteFile(filepath.Join(dir, "syncftp.ignore"), []byte(content), 0644); err == nil {
		fmt.Println(lang.L.InitIgnoreCreated)
	}
}

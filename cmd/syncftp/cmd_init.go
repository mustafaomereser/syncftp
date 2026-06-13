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

	addToIgnoreFile(dir)

	fmt.Println()
	fmt.Printf(lang.L.InitReadyFmt, projectName)
	return nil
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

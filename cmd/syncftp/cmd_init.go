package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "syncftp.json ve syncftp.secrets.json dosyalarını oluşturur",
	Long:  "İnteraktif sihirbaz ile FTP sunucu bilgilerini girerek config dosyalarını oluşturur.",
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

	fmt.Println("=== syncFTP Kurulum Sihirbazı ===")
	fmt.Println()

	dir, _ := os.Getwd()
	projectName := ask("Proje adı", filepath.Base(dir))
	localPath := ask("Yerel dizin", ".")

	fmt.Println()
	fmt.Println("FTP Sunucu bilgileri:")
	serverName := ask("Sunucu adı", "production")
	host := ask("FTP host", "")
	port := ask("Port", "21")
	user := ask("Kullanıcı adı", "")
	password := ask("Şifre (syncftp.secrets.json'a kaydedilecek)", "")
	remotePath := ask("Uzak dizin", "/public_html")

	portNum := 21
	fmt.Sscanf(port, "%d", &portNum)

	mainCfg := map[string]any{
		"project": map[string]any{
			"name":       projectName,
			"local_path": localPath,
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
				"remote_path": remotePath,
				"passive":     true,
				"enabled":     true,
				"include":     []string{},
				"exclude":     []string{},
			},
		},
	}

	mainData, err := json.MarshalIndent(mainCfg, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON oluşturulamadı: %w", err)
	}
	if err := os.WriteFile("syncftp.json", mainData, 0644); err != nil {
		return fmt.Errorf("syncftp.json yazılamadı: %w", err)
	}
	fmt.Println("✓ syncftp.json oluşturuldu")

	secretCfg := map[string]any{
		"servers": []map[string]any{
			{"name": serverName, "password": password},
		},
	}
	secretData, err := json.MarshalIndent(secretCfg, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON oluşturulamadı: %w", err)
	}
	if err := os.WriteFile("syncftp.secrets.json", secretData, 0600); err != nil {
		return fmt.Errorf("syncftp.secrets.json yazılamadı: %w", err)
	}
	fmt.Println("✓ syncftp.secrets.json oluşturuldu (izinler: 600)")

	addToIgnoreFile(dir)

	fmt.Println()
	fmt.Printf("Hazır! 'syncftp sync' komutu ile %q projesini senkronize edebilirsiniz.\n", projectName)
	return nil
}

// addToIgnoreFile adds syncFTP-specific entries to .gitignore or syncftp.ignore.
// If neither exists, creates syncftp.ignore.
func addToIgnoreFile(dir string) {
	block := "\n# syncFTP — do not commit\nsyncftp.secrets.json\nsyncftp.json\nsyncftp.exe\n"

	for _, name := range []string{".gitignore", "syncftp.ignore"} {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err == nil {
			content, _ := os.ReadFile(p)
			if strings.Contains(string(content), "syncftp.secrets.json") {
				fmt.Printf("  (syncFTP girdileri zaten %s içinde)\n", name)
				return
			}
			f, err := os.OpenFile(p, os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				return
			}
			defer f.Close()
			fmt.Fprint(f, block)
			fmt.Printf("✓ syncftp.secrets.json, syncftp.json, syncftp.exe → %s'e eklendi\n", name)
			return
		}
	}

	content := "# syncFTP — do not commit\nsyncftp.secrets.json\nsyncftp.json\nsyncftp.exe\n"
	if err := os.WriteFile(filepath.Join(dir, "syncftp.ignore"), []byte(content), 0644); err == nil {
		fmt.Println("✓ syncftp.ignore oluşturuldu (syncftp.secrets.json, syncftp.json, syncftp.exe eklendi)")
	}
}

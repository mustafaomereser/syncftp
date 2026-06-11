package main

import (
	"bufio"
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
	Short: "syncftp.toml ve syncftp.config dosyalarını oluşturur",
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
	password := ask("Şifre (syncftp.config'e kaydedilecek)", "")
	remotePath := ask("Uzak dizin", "/public_html")

	tomlContent := fmt.Sprintf(`[project]
name = %q
local_path = %q

[sync]
# Uzak sunucuda asla üzerine yazılmayacak dosya/dizinler
protect = [
  ".env",
]

# Boşsa tüm değişen dosyalar senkronize edilir.
# Dolu ise yalnızca bu yollar senkronize edilir.
include = []

[first_sync]
# true = ilk sync'te tüm dosyaları gönder
# false = sadece include listesini gönder
full = true

[[servers]]
name = %q
host = %q
port = %s
user = %q
remote_path = %q
passive = true
enabled = true
`, projectName, localPath, serverName, host, port, user, remotePath)

	if err := os.WriteFile("syncftp.toml", []byte(tomlContent), 0644); err != nil {
		return fmt.Errorf("syncftp.toml yazılamadı: %w", err)
	}
	fmt.Println("✓ syncftp.toml oluşturuldu")

	secretContent := fmt.Sprintf(`# Bu dosya şifre içerir — asla git'e commit etmeyin.
[[servers]]
name = %q
password = %q
`, serverName, password)

	if err := os.WriteFile("syncftp.config", []byte(secretContent), 0600); err != nil {
		return fmt.Errorf("syncftp.config yazılamadı: %w", err)
	}
	fmt.Println("✓ syncftp.config oluşturuldu (izinler: 600)")

	addToIgnoreFile(dir)

	fmt.Println()
	fmt.Printf("Hazır! 'syncftp sync' komutu ile %q projesini senkronize edebilirsiniz.\n", projectName)
	return nil
}

// addToIgnoreFile adds "syncftp.config" to .gitignore or syncftp.ignore.
// If neither exists, creates syncftp.ignore.
func addToIgnoreFile(dir string) {
	for _, name := range []string{".gitignore", "syncftp.ignore"} {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err == nil {
			content, _ := os.ReadFile(p)
			if strings.Contains(string(content), "syncftp.config") {
				fmt.Printf("  (syncftp.config zaten %s içinde)\n", name)
				return
			}
			f, err := os.OpenFile(p, os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				return
			}
			defer f.Close()
			fmt.Fprintln(f, "\n# syncFTP credentials — do not commit")
			fmt.Fprintln(f, "syncftp.config")
			fmt.Printf("✓ syncftp.config → %s'e eklendi\n", name)
			return
		}
	}

	// Neither file exists — create syncftp.ignore
	content := "# syncFTP credentials — do not commit\nsyncftp.config\n"
	if err := os.WriteFile(filepath.Join(dir, "syncftp.ignore"), []byte(content), 0644); err == nil {
		fmt.Println("✓ syncftp.ignore oluşturuldu (syncftp.config eklendi)")
	}
}

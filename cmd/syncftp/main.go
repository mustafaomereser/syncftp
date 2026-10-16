package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"syncftp/internal/lang"
)

var rootCmd = &cobra.Command{
	Use:     "syncftp",
	Version: appVersion,
	Short:   "Yerel dosyaları birden fazla FTP/FTPS/SFTP sunucusuna senkronize eder",
	Long: `syncFTP — Değişen dosyaları SHA256 ile tespit eder ve seçili sunuculara dağıtır.

FTP, FTPS (TLS) ve SFTP desteklenir; isteğe bağlı SSH tüneli, bant genişliği sınırı,
atomik upload ve her sync'te otomatik yedek (rollback) vardır.

Sunucu tarafındaki config dosyaları (protect listesi) korunur, üzerine yazılmaz.
İlk sync'te mevcut dosyalar calibrate ile eşleştirilir; sonrasında yalnızca değişenler gider.

Argümansız çalıştırınca interaktif shell açılır; exe'ye çift tıklamak veya
'syncftp gui' masaüstü yönetim panelini açar.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runShell()
	},
}

func main() {
	dir, _ := os.Getwd()
	lang.Init(dir)

	// Explorer'dan çift tık: süreç konsolunun tek sahibi → konsolu gizle, GUI'yi aç.
	// Terminalden argümansız çalıştırma shell'de kalır.
	// Panel içi terminal (SYNCFTP_EMBEDDED_SHELL) ConPTY'de tek süreç olduğu için ownsConsole true döner — orada shell kalmalı.
	if len(os.Args) == 1 && os.Getenv("SYNCFTP_EMBEDDED_SHELL") == "" && ownsConsole() {
		hideConsole()
		if err := runGUIForeground(); err != nil {
			os.Exit(1)
		}
		return
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

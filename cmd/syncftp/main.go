package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "syncftp",
	Short: "Yerel dosyaları birden fazla FTP sunucusuna senkronize eder",
	Long: `syncFTP — Değişen dosyaları SHA256 ile tespit eder ve seçili FTP sunucularına dağıtır.

Sunucu tarafındaki config dosyaları (protect listesi) korunur, üzerine yazılmaz.
İlk sync'te tüm dosyalar gönderilir; sonrasında yalnızca değişenler.`,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"syncftp/internal/lang"
)

func init() {
	langCmd.Flags().Bool("global", false, "Kullanıcı geneli tercih olarak kaydet (panelle ortak; proje ayarı yoksa her yerde geçerli)")
	rootCmd.AddCommand(langCmd)
}

var langCmd = &cobra.Command{
	Use:   "lang [en|tr]",
	Short: "Show or change the display language",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runLang,
}

func runLang(cmd *cobra.Command, args []string) error {
	dir, _ := os.Getwd()
	global, _ := cmd.Flags().GetBool("global")
	changeLang(dir, args, global)
	return nil
}

// changeLang CLI `lang` ve shell `lang` ortak: argümansız → göster; en|tr → değiştir ve kaydet.
// global=true → kullanıcı geneli tercih (panelle ortak), değilse bu projenin .syncftp/lang dosyası.
func changeLang(dir string, args []string, global bool) {
	if len(args) == 0 {
		fmt.Printf(lang.L.LangCurrentFmt, lang.Current())
		return
	}
	l := args[0]
	if l != "en" && l != "tr" {
		fmt.Fprintf(os.Stderr, lang.L.LangInvalid, l)
		return
	}
	lang.Set(l)
	fmt.Printf(lang.L.LangSwitchedFmt, l)
	var err error
	if global {
		err = lang.SaveUser(l)
	} else {
		err = lang.Save(dir, l)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not save language preference: %v\n", err)
	} else if global {
		fmt.Print(lang.L.LangSavedUserFmt)
	} else {
		fmt.Print(lang.L.LangSavedFmt)
	}
}

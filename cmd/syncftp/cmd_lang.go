package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"syncftp/internal/lang"
)

func init() {
	rootCmd.AddCommand(langCmd)
}

var langCmd = &cobra.Command{
	Use:   "lang [en|tr]",
	Short: "Show or change the display language",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runLang,
}

func runLang(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		fmt.Printf(lang.L.LangCurrentFmt, lang.Current())
		return nil
	}

	l := args[0]
	if l != "en" && l != "tr" {
		fmt.Fprintf(os.Stderr, lang.L.LangInvalid, l)
		return nil
	}
	if l == lang.Current() {
		fmt.Printf(lang.L.LangAlreadyFmt, l)
		return nil
	}

	lang.Set(l)
	fmt.Printf(lang.L.LangSwitchedFmt, l)

	dir, _ := os.Getwd()
	if err := lang.Save(dir, l); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not save language preference: %v\n", err)
	} else {
		fmt.Print(lang.L.LangSavedFmt)
	}
	return nil
}

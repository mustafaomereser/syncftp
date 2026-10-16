package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	goftp "github.com/jlaffaye/ftp"

	ftpclient "syncftp/internal/ftp"
	"syncftp/internal/lang"
)

var treeGrey = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

// cmdTree — FTP dizinini ağaç görünümünde listeler.
// Kullanım: tree [path] [--max N]
//
//	--max N : klasör başına max N öğe göster (0 = sınırsız)
//	path    : başlangıç dizini (varsayılan: mevcut dizin)
func (sh *shellState) cmdTree(args []string) {
	if !sh.ensureConnected() {
		return
	}

	rootPath := sh.remoteCwd
	maxPerDir := -1 // -1 = henüz sorulmadı

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--max":
			if i+1 < len(args) {
				i++
				n, err := strconv.Atoi(args[i])
				if err == nil {
					maxPerDir = n
				}
			}
		default:
			if !strings.HasPrefix(args[i], "-") {
				rootPath = sh.resolvePath(args[i])
			}
		}
	}

	// --max verilmediyse interaktif sor
	if maxPerDir == -1 {
		items := []PickerItem{
			{Icon: "20", Label: "20 dosyaya kadar", Value: "20"},
			{Icon: "50", Label: "50 dosyaya kadar", Value: "50"},
			{Icon: "100", Label: "100 dosyaya kadar", Value: "100"},
			{Icon: "∞", Label: "Tümünü göster", Value: "0"},
		}
		val, err := RunPicker(lang.L.TreeMaxPromptTitle, lang.L.TreeMaxPromptSub, items)
		if err != nil || val == "" {
			return
		}
		n, _ := strconv.Atoi(val)
		maxPerDir = n
	}

	fmt.Println(rootPath)
	sh.printFTPTree(rootPath, "", maxPerDir)
	fmt.Println()
}

func (sh *shellState) printFTPTree(remotePath, prefix string, maxPerDir int) {
	writeFTPTree(os.Stdout, sh.client, remotePath, prefix, maxPerDir, treeGrey.Render)
}

// writeFTPTree uzak klasör ağacını yazar — shell `tree` ve panelin "Ağaç" işi ortak.
// grey dal karakterlerini boyar (panelde düz metin).
func writeFTPTree(w io.Writer, c *ftpclient.Client, remotePath, prefix string, maxPerDir int, grey func(...string) string) {
	entries, err := remoteEntries(c, remotePath) // ortak sıralı liste (remote_ops.go)
	if err != nil {
		fmt.Fprintf(w, "%s  %s\n", prefix, fmt.Sprintf(lang.L.TreeErrFmt, err))
		return
	}
	// Limit kontrolü: fazlasını kes, kalanı not olarak ekle
	remaining := 0
	if maxPerDir > 0 && len(entries) > maxPerDir {
		remaining = len(entries) - maxPerDir
		entries = entries[:maxPerDir]
	}
	for i, e := range entries {
		isLast := i == len(entries)-1 && remaining == 0
		connector := grey("├─")
		childPrefix := prefix + grey("│") + "  "
		if isLast {
			connector = grey("└─")
			childPrefix = prefix + "   "
		}
		if e.Type == goftp.EntryTypeFolder {
			fmt.Fprintf(w, "%s%s 📁 %s/\n", prefix, connector, e.Name)
			writeFTPTree(w, c, path.Join(remotePath, e.Name), childPrefix, maxPerDir, grey)
		} else {
			fmt.Fprintf(w, "%s%s 📄 %s  %s\n", prefix, connector, e.Name, treeFormatSize(e.Size))
		}
	}
	if remaining > 0 {
		fmt.Fprintf(w, "%s%s "+lang.L.TreeMoreFmt+"\n", prefix, grey("└─"), remaining)
	}
}

func treeFormatSize(bytes uint64) string {
	switch {
	case bytes >= 1<<20:
		return fmt.Sprintf("(%.1f MB)", float64(bytes)/float64(1<<20))
	case bytes >= 1<<10:
		return fmt.Sprintf("(%.1f KB)", float64(bytes)/float64(1<<10))
	default:
		return fmt.Sprintf("(%d B)", bytes)
	}
}

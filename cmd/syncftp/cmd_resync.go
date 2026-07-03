package main

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"syncftp/internal/config"
	ftpclient "syncftp/internal/ftp"
	"syncftp/internal/frozen"
	"syncftp/internal/ignore"
	"syncftp/internal/lang"
	"syncftp/internal/scanner"
	"syncftp/internal/state"
)

var (
	calibrateFlagServer string
	calibrateFlagAll    bool
)

func init() {
	calibrateCmd.Flags().StringVarP(&calibrateFlagServer, "server", "s", "", "Sadece bu sunucuyu kalibre et")
	calibrateCmd.Flags().BoolVar(&calibrateFlagAll, "all", false, "Tüm aktif sunucuları kalibre et")
	rootCmd.AddCommand(calibrateCmd)
}

var calibrateCmd = &cobra.Command{
	Use:   "calibrate",
	Short: "Yerel dosyaları FTP boyutlarıyla karşılaştırıp state'i günceller (yükleme yapmaz)",
	RunE:  runCalibrateCmd,
}

func runCalibrateCmd(cmd *cobra.Command, args []string) error {
	dir, _ := os.Getwd()
	cfg, err := config.Load(dir)
	if err != nil {
		return err
	}

	servers := cfg.Servers
	if calibrateFlagServer != "" {
		var found []config.Server
		for _, s := range servers {
			if s.Name == calibrateFlagServer {
				found = append(found, s)
				break
			}
		}
		if len(found) == 0 {
			fmt.Print(lang.L.ResyncNoServers)
			return nil
		}
		servers = found
	} else if !calibrateFlagAll {
		// varsayılan: sadece enabled sunucular
		var enabled []config.Server
		for _, s := range servers {
			if s.Enabled {
				enabled = append(enabled, s)
			}
		}
		servers = enabled
	}

	if len(servers) == 0 {
		fmt.Print(lang.L.ResyncNoServers)
		return nil
	}

	for _, srv := range servers {
		runCalibrate(dir, srv, cfg)
	}
	return nil
}

// countCRLF dosyadaki \r\n çiftlerini sayar (32 KB chunk'larla, bellek verimli).
func countCRLF(path string) uint64 {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()

	buf := make([]byte, 32*1024)
	var count uint64
	var prev byte
	for {
		n, err := f.Read(buf)
		for i := 0; i < n; i++ {
			if prev == '\r' && buf[i] == '\n' {
				count++
			}
			prev = buf[i]
		}
		if err != nil {
			break
		}
	}
	return count
}

// passesFilter sync ile aynı include/exclude mantığını uygular.
func passesFilter(relPath string, include, exclude []string) bool {
	if len(include) > 0 {
		matched := false
		for _, inc := range include {
			inc = strings.TrimSuffix(filepath.ToSlash(inc), "/")
			if relPath == inc || strings.HasPrefix(relPath, inc+"/") {
				matched = true
				break
			}
			if ok, _ := path.Match(inc, relPath); ok {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	for _, exc := range exclude {
		exc = strings.TrimSuffix(filepath.ToSlash(exc), "/")
		if relPath == exc || strings.HasPrefix(relPath, exc+"/") {
			return false
		}
		if ok, _ := path.Match(exc, relPath); ok {
			return false
		}
	}
	return true
}

// runCalibrate yerel dosyaları FTP boyutlarıyla karşılaştırıp eşleşenleri state'e yazar.
// Yükleme yapmaz. Hata olursa sessizce devam eder (state değişmez).
func runCalibrate(dir string, srv config.Server, cfg *config.Config) {
	localPath := srv.EffectiveLocalPath(cfg.Project)
	localDir := filepath.Join(dir, localPath)

	matcher, _ := ignore.Load(localDir, cfg.Sync.IgnoreFiles)
	files, err := scanner.Scan(localDir, matcher)
	if err != nil {
		return
	}

	// Sync ile aynı include/exclude filtreleri (server > global, exclude additive)
	effectiveInclude := cfg.Sync.Include
	if len(srv.Include) > 0 {
		effectiveInclude = srv.Include
	}
	effectiveExclude := append(append([]string{}, cfg.Sync.Exclude...), srv.Exclude...)

	current := make(map[string]string)
	byKey := make(map[string]scanner.File)
	excluded := 0
	for _, f := range files {
		if !passesFilter(f.RelPath, effectiveInclude, effectiveExclude) {
			excluded++
			continue
		}
		current[f.RelPath] = f.Hash
		byKey[f.RelPath] = f
	}

	// Ignore edilen dosya ve dizinleri say (matcher'ı kullanarak — hız için ignored dizinlere girme)
	var ignoredFiles int
	var ignoredDirs []string
	_ = filepath.WalkDir(localDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(localDir, p)
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil
		}
		if rel == ".syncftp" || strings.HasPrefix(rel, ".syncftp/") ||
			rel == ".git" || strings.HasPrefix(rel, ".git/") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if matcher.Match(rel) {
			if d.IsDir() {
				ignoredDirs = append(ignoredDirs, rel+"/")
				return filepath.SkipDir
			}
			ignoredFiles++
			return nil
		}
		return nil
	})
	ignoredCount := ignoredFiles

	st, _ := state.Load(dir, srv.Name)

	fmt.Printf("  [%s]\n", srv.Name)
	fmt.Printf(lang.L.ResyncLocalFmt, len(current))
	if len(ignoredDirs) > 0 {
		fmt.Printf(lang.L.ResyncIgnoreDirsFmt, len(ignoredDirs), strings.Join(ignoredDirs, ", "))
	}
	if ignoredCount > 0 {
		fmt.Printf(lang.L.ResyncIgnoreFilesFmt, ignoredCount)
	}
	if excluded > 0 {
		fmt.Printf(lang.L.ResyncFilteredFmt, excluded)
	}
	fmt.Print(lang.L.ResyncScanning)

	client, err := ftpclient.Connect(srv)
	if err != nil {
		fmt.Printf(lang.L.ResyncConnErr, err)
		return
	}
	defer client.Close()
	fmt.Print(lang.L.ResyncConnected)

	fmt.Printf("  %s\n", lang.L.ResyncListing)
	lastList := time.Now()
	remoteFiles, err := client.ListRecursiveProgress(srv.RemotePath, func(n int) {
		if time.Since(lastList) >= 80*time.Millisecond {
			fmt.Printf(lang.L.ResyncListProgressFmt, n)
			lastList = time.Now()
		}
	})
	fmt.Print("\r                                        \r")
	if err != nil {
		fmt.Printf(lang.L.ResyncListErr, err)
		return
	}
	fmt.Printf(lang.L.ResyncFoundFmt, len(remoteFiles))

	frozenList, _ := frozen.Load(dir, srv.Name)

	total := len(current)
	done := 0
	matched := 0
	different := 0
	frozenDiff := 0
	lastPrint := time.Now()
	for key, hash := range current {
		done++
		if done == total || time.Since(lastPrint) >= 80*time.Millisecond {
			fmt.Printf(lang.L.ResyncComparingFmt, done, total)
			lastPrint = time.Now()
		}

		f := byKey[key]
		localInfo, err := os.Stat(f.AbsPath)
		if err != nil {
			different++
			continue
		}
		localSize := uint64(localInfo.Size())
		remoteSize, exists := remoteFiles[key]
		if !exists {
			delete(st.Files, key)
			different++
			if frozen.IsFrozen(frozenList, key) {
				frozenDiff++
			}
			continue
		}
		// Boyut doğrudan eşleşiyorsa veya CRLF→LF farkı açıklıyorsa eşleşti say
		crlfCount := countCRLF(f.AbsPath)
		if remoteSize == localSize || (crlfCount > 0 && remoteSize == localSize-crlfCount) {
			st.Files[key] = hash
			matched++
		} else {
			// Boyut uyuşmadı — bir sonraki sync'in yeniden yüklemesi için state'ten sil
			delete(st.Files, key)
			different++
			if frozen.IsFrozen(frozenList, key) {
				frozenDiff++
			}
		}
	}

	fmt.Printf(lang.L.ResyncMatchedFmt, matched, different)
	if frozenDiff > 0 {
		fmt.Printf(lang.L.ResyncFrozenDiffFmt, frozenDiff)
	}

	st.FirstSyncDone = true
	_ = state.Save(dir, st)

	fmt.Printf(lang.L.ResyncDoneFmt, srv.Name)
}

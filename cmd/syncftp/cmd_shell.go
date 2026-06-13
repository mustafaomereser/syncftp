package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/chzyer/readline"

	"syncftp/internal/config"
	"syncftp/internal/failed"
	ftpclient "syncftp/internal/ftp"
	"syncftp/internal/frozen"
	"syncftp/internal/ignore"
	"syncftp/internal/lang"
	"syncftp/internal/release"
	"syncftp/internal/scanner"
	"syncftp/internal/state"
)

// shellState tuttuğu bağlantı + gezinme durumu
type shellState struct {
	cfg       config.Config
	configDir string
	srv       *config.Server
	client    *ftpclient.Client
	remoteCwd string
}

func runShell() error {
	dir, _ := os.Getwd()

	cfgPtr, err := config.Load(dir)
	if err != nil {
		return fmt.Errorf("syncftp.json okunamadı: %w\n\nİpucu: önce 'syncftp init' çalıştırın", err)
	}

	sh := &shellState{
		cfg:       *cfgPtr,
		configDir: dir,
	}

	// Tek sunucu varsa otomatik bağlan
	servers := cfgPtr.EnabledServers()
	if len(servers) == 1 {
		sh.connectServer(servers[0])
	}

	rl, err := readline.NewEx(&readline.Config{
		HistoryFile:     filepath.Join(dir, ".syncftp", "shell_history"),
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		return err
	}
	defer rl.Close()

	sh.printWelcome()

	if sh.srv == nil && len(servers) > 1 {
		fmt.Println(lang.L.ShellMultiServer)
		fmt.Println(lang.L.ShellServersLabel, serverNames(servers))
		fmt.Println()
	}

	ctrlCCount := 0
	for {
		rl.SetPrompt(sh.prompt())
		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			ctrlCCount++
			if ctrlCCount >= 2 {
				fmt.Println(lang.L.ShellExit)
				return nil
			}
			fmt.Println(lang.L.ShellCtrlCExitHint)
			continue
		}
		if err == io.EOF {
			break
		}

		ctrlCCount = 0
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := splitArgs(line)
		cmd := strings.ToLower(parts[0])
		args := parts[1:]

		switch cmd {
		case "exit", "quit", "q":
			fmt.Println(lang.L.ShellExit)
			return nil

		case "help", "?":
			sh.printHelp()

		case "lang":
			sh.cmdLang(args)

		case "freeze":
			sh.cmdFreeze(args)

		case "config":
			if err := runServerMgr(sh.configDir); err != nil {
				fmt.Printf(lang.L.ShellErrFmt, err)
			}
			// Config değişmiş olabilir — yeniden yükle
			if cfg, err := config.Load(sh.configDir); err == nil {
				sh.cfg = *cfg
			}

		case "clear", "cls":
			sh.cmdClear()

		case "servers":
			sh.cmdServers()

		case "server":
			sh.cmdServer(args)

		case "disconnect":
			sh.cmdDisconnect()

		case "pwd":
			sh.cmdPwd()

		case "ls":
			sh.cmdLs(args)

		case "tree":
			sh.cmdTree(args)

		case "cd":
			sh.cmdCd(args)

		case "cat":
			sh.cmdCat(args)

		case "get":
			sh.cmdGet(args)

		case "rm":
			sh.cmdRm(args)

		case "status":
			sh.cmdStatus()

		case "sync":
			sh.cmdSync(args)

		default:
			fmt.Printf(lang.L.ShellUnknownCmd, cmd)
		}
	}
	return nil
}

// ── prompt + welcome ──────────────────────────────────────────────────────────

func (sh *shellState) prompt() string {
	// \001 ve \002 readline'ın görünmez karakterleri doğru hesaplaması için gerekli
	reset := "\001\033[0m\002"
	bold := "\001\033[1m\002"
	cyan := "\001\033[36m\002"
	blue := "\001\033[34m\002"
	dim := "\001\033[2m\002"

	if sh.srv == nil {
		return bold + "syncftp" + reset + dim + "> " + reset
	}
	return bold + "syncftp" + reset +
		dim + " [" + reset +
		cyan + sh.srv.Name + reset +
		dim + ":" + reset +
		blue + sh.remoteCwd + reset +
		dim + "]" + reset +
		bold + "> " + reset
}

func (sh *shellState) printWelcome() {
	name := sh.cfg.Project.Name
	// Kutu genişliği 44 karakter içi
	inner := fmt.Sprintf("  syncFTP Shell  ·  %-22s", name)
	if len(inner) > 42 {
		inner = inner[:42]
	}
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════╗")
	fmt.Printf("║%-44s║\n", inner)
	fmt.Printf("║%-44s║\n", lang.L.ShellWelcomeHint)
	fmt.Println("╚══════════════════════════════════════════════╝")
	fmt.Println()
}

// ── bağlantı ──────────────────────────────────────────────────────────────────

func (sh *shellState) connectServer(srv config.Server) {
	if sh.client != nil {
		sh.client.Close()
		sh.client = nil
	}
	fmt.Printf(lang.L.ShellConnecting, srv.Name, srv.Host)
	client, err := ftpclient.Connect(srv)
	if err != nil {
		fmt.Printf(lang.L.ShellConnectErr, err)
		return
	}
	sh.srv = &srv
	sh.client = client
	sh.remoteCwd = srv.RemotePath
	fmt.Printf(" ✓\n")
}

func (sh *shellState) cmdDisconnect() {
	if sh.client == nil {
		fmt.Println(lang.L.ShellNotConnected)
		return
	}
	name := sh.srv.Name
	sh.client.Close()
	sh.client = nil
	sh.srv = nil
	sh.remoteCwd = ""
	fmt.Printf(lang.L.ShellDisconnected, name)
}

func (sh *shellState) ensureConnected() bool {
	if sh.client != nil {
		return true
	}
	fmt.Println(lang.L.ShellNotConnected)
	return false
}

// ── komutlar ──────────────────────────────────────────────────────────────────

func (sh *shellState) cmdFreeze(args []string) {
	if sh.srv == nil {
		fmt.Println("Not connected. Use 'server <name>' first.")
		return
	}
	items, err := buildFreezeItems(sh.configDir, &sh.cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fl, _ := frozen.Load(sh.configDir, sh.srv.Name)
	var preSelected map[string]bool
	if fl != nil {
		preSelected = fl.Files
	}

	selected, err := RunFreezeList(
		fmt.Sprintf("Freeze list — %s", sh.srv.Name),
		items,
		preSelected,
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	if selected == nil {
		fmt.Println("Cancelled.")
		return
	}
	if err := frozen.Save(sh.configDir, sh.srv.Name, selected); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("✓ %d files frozen for [%s]\n", len(selected), sh.srv.Name)
}

func (sh *shellState) cmdLang(args []string) {
	if len(args) == 0 {
		fmt.Printf(lang.L.LangCurrentFmt, lang.Current())
		return
	}
	l := args[0]
	if l != "en" && l != "tr" {
		fmt.Printf(lang.L.LangInvalid, l)
		return
	}
	if l == lang.Current() {
		fmt.Printf(lang.L.LangAlreadyFmt, l)
		return
	}
	lang.Set(l)
	fmt.Printf(lang.L.LangSwitchedFmt, l)
	if err := lang.Save(sh.configDir, l); err != nil {
		fmt.Printf("Warning: could not save: %v\n", err)
	} else {
		fmt.Print(lang.L.LangSavedFmt)
	}
}

func (sh *shellState) cmdClear() {
	fmt.Print("\033[H\033[2J\033[3J")
	sh.printWelcome()
}

func (sh *shellState) cmdServers() {
	servers := sh.cfg.EnabledServers()
	if len(servers) == 0 {
		fmt.Println(lang.L.ShellNoServers)
		return
	}
	current := ""
	if sh.srv != nil {
		current = sh.srv.Name
	}
	items := make([]PickerItem, len(servers))
	for i, s := range servers {
		icon := "🖥"
		desc := fmt.Sprintf("%s:%d  conn:%d  retry:%d", s.Host, s.Port, s.MaxConnections, s.MaxRetries)
		if s.Name == current {
			icon = "✅"
			desc += "  (bağlı)"
		}
		items[i] = PickerItem{Icon: icon, Label: s.Name, Desc: desc, Value: s.Name}
	}
	val, err := RunPicker(lang.L.ShellServersFmt, lang.L.ShellServersSubtitle, items)
	if err != nil || val == "" {
		return
	}
	if val == current {
		fmt.Printf(lang.L.ShellAlreadyConn, val)
		return
	}
	for _, s := range servers {
		if s.Name == val {
			sh.connectServer(s)
			return
		}
	}
}

func (sh *shellState) cmdServer(args []string) {
	servers := sh.cfg.EnabledServers()
	if len(args) == 0 {
		if len(servers) == 0 {
			fmt.Println(lang.L.ShellNoServers)
			return
		}
		srv, err := pickServerTUI(servers)
		if err != nil || srv == nil {
			return
		}
		sh.connectServer(*srv)
		return
	}
	name := args[0]
	for _, s := range servers {
		if s.Name == name {
			sh.connectServer(s)
			return
		}
	}
	fmt.Printf(lang.L.ShellServerNotFound, name)
}

func (sh *shellState) cmdPwd() {
	if !sh.ensureConnected() {
		return
	}
	fmt.Println(sh.remoteCwd)
}

// cmdLs — interaktif dosya tarayıcısı açar. Arrow key ile gezinilir.
func (sh *shellState) cmdLs(args []string) {
	if !sh.ensureConnected() {
		return
	}
	startPath := sh.remoteCwd
	if len(args) > 0 {
		startPath = sh.resolvePath(args[0])
	}

	result, err := RunBrowser(sh.client, startPath, sh.srv.RemotePath, sh.configDir, sh.srv)
	if err != nil {
		fmt.Printf(lang.L.ShellBrowserErr, err)
		return
	}
	sh.remoteCwd = result.CWD

	switch result.Action {
	case "delete":
		sh.browserDelete(result.Marked)
	case "move":
		sh.browserMove(result.Marked)
	case "tree":
		sh.cmdTree(nil)
	default:
		if result.Selected != "" {
			sh.fileActionMenu(result.Selected)
		}
	}
}

// browserDelete işaretli dosyaları onay alarak FTP'den siler.
func (sh *shellState) browserDelete(files []string) {
	if len(files) == 0 {
		return
	}

	// Dosya listesini confirm dialog'una taşı — terminale önceden yazdırma
	var listSb strings.Builder
	for _, f := range files {
		listSb.WriteString("  - " + f + "\n")
	}
	subtitle := strings.TrimRight(listSb.String(), "\n")

	ok, err := RunConfirm(
		fmt.Sprintf(lang.L.ShellDeleteTitle, len(files)),
		subtitle,
	)
	if err != nil || !ok {
		fmt.Println(lang.L.ShellCancelled)
		return
	}
	ok2, _ := RunConfirm(lang.L.ShellConfirmSure, lang.L.ShellConfirmSureBody)
	if !ok2 {
		fmt.Println(lang.L.ShellCancelled)
		return
	}

	fmt.Println()
	deleted, failed := 0, 0
	for _, f := range files {
		// Klasör mü dosya mı?
		entries, listErr := sh.client.List(f)
		isDir := listErr == nil && entries != nil
		if isDir {
			// Anlık feedback — klasör için işlem başladığını göster
			fmt.Printf("  ⟳ %s/", f)
			contents, _ := sh.client.ListRecursive(f)
			err = sh.client.DeleteDir(f, true)
			if err != nil {
				fmt.Printf("\r  ✗ %s/  (%v)\n", f, err)
				failed++
			} else {
				fmt.Printf("\r  ✓ %s/\n", f)
				paths := make([]string, 0, len(contents))
				for relPath := range contents {
					paths = append(paths, relPath)
				}
				sort.Strings(paths)
				for i, relPath := range paths {
					branch := "├─"
					if i == len(paths)-1 {
						branch = "└─"
					}
					fmt.Printf("       %s %s\n", branch, relPath)
				}
				deleted++
			}
		} else {
			fmt.Printf("  ⟳ %s", f)
			err = sh.client.DeleteFile(f)
			if err != nil {
				fmt.Printf("\r  ✗ %s  (%v)\n", f, err)
				failed++
			} else {
				fmt.Printf("\r  ✓ %s\n", f)
				deleted++
			}
		}
	}
	fmt.Printf(lang.L.ShellDeletedFmt, deleted, failed)
}

// browserMove işaretli dosyaları seçilen hedef dizine taşır.
func (sh *shellState) browserMove(files []string) {
	if len(files) == 0 {
		return
	}
	fmt.Printf(lang.L.ShellMovingFmt, len(files))

	destDir, err := RunBrowserPickDir(sh.client, sh.remoteCwd, sh.srv.RemotePath, sh.configDir, sh.srv)
	if err != nil {
		fmt.Printf(lang.L.ShellBrowserErr, err)
		return
	}
	if destDir == "" {
		fmt.Println(lang.L.ShellCancelled)
		return
	}

	fmt.Printf(lang.L.ShellMoveTarget, destDir)
	moved, failedCount := 0, 0
	for _, f := range files {
		filename := path.Base(f)
		newPath := path.Join(destDir, filename)
		if err := sh.client.Rename(f, newPath); err != nil {
			fmt.Printf("  ✗ %s  (%v)\n", filename, err)
			failedCount++
		} else {
			fmt.Printf("  ✓ %s  →  %s\n", filename, newPath)
			moved++
		}
	}
	fmt.Printf(lang.L.ShellMovedFmt, moved, failedCount)
}

func (sh *shellState) cmdCd(args []string) {
	if !sh.ensureConnected() {
		return
	}
	if len(args) == 0 {
		sh.remoteCwd = sh.srv.RemotePath
		return
	}
	target := args[0]
	if target == ".." {
		parent := sh.resolvePath("..")
		sh.remoteCwd = parent
		return
	}
	resolved := sh.resolvePath(target)
	entries, err := sh.client.List(resolved)
	if err != nil || entries == nil {
		fmt.Printf(lang.L.ShellDirNotFound, resolved)
		return
	}
	sh.remoteCwd = resolved
}

func (sh *shellState) cmdCat(args []string) {
	if !sh.ensureConnected() {
		return
	}

	maxKB := 10
	remaining := []string{}
	for i := 0; i < len(args); i++ {
		if (args[i] == "--max-kb" || args[i] == "-n") && i+1 < len(args) {
			fmt.Sscanf(args[i+1], "%d", &maxKB)
			i++
		} else {
			remaining = append(remaining, args[i])
		}
	}

	var filePath string
	if len(remaining) > 0 {
		resolved := sh.resolvePath(remaining[0])
		// Dizinse browser aç
		if entries, err := sh.client.List(resolved); err == nil && entries != nil {
			result, err := RunBrowser(sh.client, resolved, sh.srv.RemotePath, sh.configDir, sh.srv)
			if err != nil || result.Quit || result.Selected == "" {
				return
			}
			sh.remoteCwd = result.CWD
			filePath = result.Selected
		} else {
			filePath = resolved
		}
	} else {
		result, err := RunBrowser(sh.client, sh.remoteCwd, sh.srv.RemotePath, sh.configDir, sh.srv)
		if err != nil || result.Quit || result.Selected == "" {
			return
		}
		sh.remoteCwd = result.CWD
		filePath = result.Selected
	}

	sh.catFile(filePath, maxKB)
}

func (sh *shellState) catFile(filePath string, maxKB int) {
	maxBytes := int64(maxKB) * 1024
	data, err := sh.client.Preview(filePath, maxBytes)
	if err != nil {
		fmt.Printf(lang.L.ShellErrFmt, err)
		return
	}
	fmt.Printf("\n── %s ──\n", filePath)
	fmt.Print(string(data))
	if !strings.HasSuffix(string(data), "\n") {
		fmt.Println()
	}
	if int64(len(data)) == maxBytes {
		fmt.Printf(lang.L.ShellFileTruncFmt, maxKB, filePath)
	}
}

func (sh *shellState) cmdGet(args []string) {
	if !sh.ensureConnected() {
		return
	}

	var remotePath, localDest string

	if len(args) == 0 {
		result, err := RunBrowser(sh.client, sh.remoteCwd, sh.srv.RemotePath, sh.configDir, sh.srv)
		if err != nil || result.Quit || result.Selected == "" {
			return
		}
		sh.remoteCwd = result.CWD
		remotePath = result.Selected
		localDest = filepath.Base(remotePath)
	} else {
		resolved := sh.resolvePath(args[0])
		if entries, err := sh.client.List(resolved); err == nil && entries != nil {
			result, err := RunBrowser(sh.client, resolved, sh.srv.RemotePath, sh.configDir, sh.srv)
			if err != nil || result.Quit || result.Selected == "" {
				return
			}
			sh.remoteCwd = result.CWD
			remotePath = result.Selected
		} else {
			remotePath = resolved
		}
		if len(args) >= 2 {
			localDest = args[1]
		} else {
			localDest = filepath.Base(remotePath)
		}
	}

	sh.getFile(remotePath, localDest)
}

func (sh *shellState) getFile(remotePath, localDest string) {
	if localDest == "" {
		localDest = filepath.Base(remotePath)
	}
	fmt.Printf(lang.L.ShellDownloading, remotePath, localDest)
	if err := sh.client.Download(remotePath, localDest); err != nil {
		fmt.Printf(lang.L.ShellErrFmt, err)
		return
	}
	info, _ := os.Stat(localDest)
	if info != nil {
		fmt.Printf(lang.L.ShellDownloaded, formatSize(uint64(info.Size())))
	} else {
		fmt.Println(lang.L.ShellDownloadedBare)
	}
}

func (sh *shellState) cmdRm(args []string) {
	if !sh.ensureConnected() {
		return
	}

	force, recursive := false, false
	remaining := []string{}
	for _, a := range args {
		switch a {
		case "-f", "--force":
			force = true
		case "-r", "--recursive":
			recursive = true
		default:
			remaining = append(remaining, a)
		}
	}

	var remotePath string
	if len(remaining) == 0 {
		result, err := RunBrowser(sh.client, sh.remoteCwd, sh.srv.RemotePath, sh.configDir, sh.srv)
		if err != nil || result.Quit || result.Selected == "" {
			return
		}
		sh.remoteCwd = result.CWD
		remotePath = result.Selected
	} else {
		resolved := sh.resolvePath(remaining[0])
		if entries, err := sh.client.List(resolved); err == nil && entries != nil {
			result, err := RunBrowser(sh.client, resolved, sh.srv.RemotePath, sh.configDir, sh.srv)
			if err != nil || result.Quit || result.Selected == "" {
				return
			}
			sh.remoteCwd = result.CWD
			remotePath = result.Selected
		} else {
			remotePath = resolved
		}
	}

	entries, err := sh.client.List(remotePath)
	isDir := err == nil && entries != nil

	label := lang.L.RemoteFileLabel
	if isDir {
		if recursive {
			label = lang.L.RemoteDirRecLabel
		} else {
			label = lang.L.RemoteDirLabel
		}
	}

	if !force {
		confirmed, err := RunConfirm(
			lang.L.ShellDeleteConfTitle,
			fmt.Sprintf("%s silinecek: %s", label, remotePath),
		)
		if err != nil || !confirmed {
			fmt.Println(lang.L.ShellCancelShort)
			return
		}
	}

	if isDir {
		if err := sh.client.DeleteDir(remotePath, recursive); err != nil {
			fmt.Printf("Hata: %v\n", err)
			return
		}
	} else {
		if err := sh.client.DeleteFile(remotePath); err != nil {
			fmt.Printf(lang.L.ShellErrFmt, err)
			return
		}
	}
	fmt.Printf(lang.L.ShellDeletedShort, remotePath)
}

func (sh *shellState) cmdStatus() {
	cfg := sh.cfg
	dir := sh.configDir

	servers := cfg.EnabledServers()
	if len(servers) == 0 {
		fmt.Println(lang.L.ShellNoServers)
		return
	}
	if sh.srv != nil {
		servers = []config.Server{*sh.srv}
	}

	var entries []statusSrvEntry
	for _, srv := range servers {
		localPath := srv.EffectiveLocalPath(cfg.Project)
		localDir := filepath.Join(dir, localPath)

		matcher, err := ignore.Load(localDir, cfg.Sync.IgnoreFiles)
		if err != nil {
			fmt.Printf(lang.L.ShellErrFmt, err)
			continue
		}
		files, err := scanner.Scan(localDir, matcher)
		if err != nil {
			fmt.Printf(lang.L.ShellErrFmt, err)
			continue
		}
		current := make(map[string]string, len(files))
		for _, f := range files {
			current[f.RelPath] = f.Hash
		}

		st, err := state.Load(dir, srv.Name)
		if err != nil {
			fmt.Printf(lang.L.ShellStatusStateErr, srv.Name, err)
			continue
		}
		diff := state.Diff(st, current)

		// Include önceliği: sunucu > global
		srvInclude := cfg.Sync.Include
		if len(srv.Include) > 0 {
			srvInclude = srv.Include
		}
		srvInclude = normalizePatterns(srvInclude, localPath)

		srvExclude := append(append([]string{}, cfg.Sync.Exclude...), srv.Exclude...)
		srvExclude = normalizePatterns(srvExclude, localPath)

		added := applyStatusFilters(diff.New, srvInclude, srvExclude)
		changed := applyStatusFilters(diff.Changed, srvInclude, srvExclude)
		deleted := applyStatusFilters(diff.Deleted, srvInclude, srvExclude)

		// Frozen dosyaları ayır
		fl, _ := frozen.Load(dir, srv.Name)
		var frozenFiles []string
		if fl != nil {
			var addedClean, changedClean []string
			for _, f := range added {
				if frozen.IsFrozen(fl, f) {
					frozenFiles = append(frozenFiles, f)
				} else {
					addedClean = append(addedClean, f)
				}
			}
			for _, f := range changed {
				if frozen.IsFrozen(fl, f) {
					frozenFiles = append(frozenFiles, f)
				} else {
					changedClean = append(changedClean, f)
				}
			}
			added, changed = addedClean, changedClean
		}

		var totalSize int64
		for _, rel := range append(added, changed...) {
			if info, err := os.Stat(filepath.Join(localDir, filepath.FromSlash(rel))); err == nil {
				totalSize += info.Size()
			}
		}

		sort.Strings(added)
		sort.Strings(changed)
		sort.Strings(deleted)
		sort.Strings(frozenFiles)

		entries = append(entries, statusSrvEntry{
			name:     srv.Name,
			localDir: localDir,
			added:    added,
			changed:  changed,
			deleted:  deleted,
			frozen:   frozenFiles,
			size:     totalSize,
		})
	}

	if len(entries) == 0 {
		return
	}

	m := newStatusTUI(cfg.Project.Name, entries)
	p := tea.NewProgram(m, tea.WithAltScreen())
	final, _ := p.Run()
	if fm, ok := final.(statusTUI); ok && fm.syncServer != "" {
		for _, srv := range sh.cfg.EnabledServers() {
			if srv.Name == fm.syncServer {
				fmt.Printf("\n── %s ──\n", srv.Name)
				sh.shellSyncServer(srv, false, false)
				break
			}
		}
	}
}

func (sh *shellState) cmdSync(args []string) {
	full, dryRun, all := false, false, false
	serverName := ""
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--full":
			full = true
		case "--dry-run":
			dryRun = true
		case "--all":
			all = true
		case "--server":
			if i+1 < len(args) {
				serverName = args[i+1]
				i++
			}
		}
	}

	servers := sh.cfg.EnabledServers()
	if serverName != "" {
		filtered := []config.Server{}
		for _, s := range servers {
			if s.Name == serverName {
				filtered = append(filtered, s)
			}
		}
		servers = filtered
	} else if !all && len(servers) > 1 {
		// Bağlı sunucu varsa onu önceden işaretli aç
		connectedName := ""
		if sh.srv != nil {
			connectedName = sh.srv.Name
		}
		selected, err := pickServerMultiTUI(servers, connectedName)
		if err != nil || selected == nil {
			fmt.Println(lang.L.ShellSyncCancelled)
			return
		}
		servers = selected
	}
	if len(servers) == 0 {
		fmt.Println(lang.L.ShellNoServers)
		return
	}

	for _, srv := range servers {
		fmt.Printf("\n── %s ──\n", srv.Name)
		sh.shellSyncServer(srv, full, dryRun)
	}
}

func (sh *shellState) shellSyncServer(srv config.Server, full, dryRun bool) {
	cfg := sh.cfg
	dir := sh.configDir

	localPath := srv.EffectiveLocalPath(cfg.Project)
	localDir := filepath.Join(dir, localPath)

	matcher, err := ignore.Load(localDir, cfg.Sync.IgnoreFiles)
	if err != nil {
		fmt.Printf(lang.L.ShellErrFmt, err)
		return
	}
	fmt.Print(lang.L.ShellScanning)
	files, err := scanner.Scan(localDir, matcher)
	if err != nil {
		fmt.Printf(lang.L.ShellConnectErr, err)
		return
	}
	fmt.Printf(lang.L.ShellScannedFmt, len(files))

	current := make(map[string]string, len(files))
	byRel := make(map[string]scanner.File, len(files))
	for _, f := range files {
		current[f.RelPath] = f.Hash
		byRel[f.RelPath] = f
	}

	// Include önceliği: sunucu > global (CLI yok shell'de)
	effectiveInclude := cfg.Sync.Include
	if len(srv.Include) > 0 {
		effectiveInclude = srv.Include
	}
	effectiveInclude = normalizePatterns(effectiveInclude, localPath)

	effectiveExclude := append(append([]string{}, cfg.Sync.Exclude...), srv.Exclude...)
	effectiveExclude = normalizePatterns(effectiveExclude, localPath)

	effectiveProtect := append(append([]string{}, cfg.Sync.Protect...), srv.Protect...)
	effectiveProtect = normalizePatterns(effectiveProtect, localPath)

	st, err := state.Load(dir, srv.Name)
	if err != nil {
		fmt.Printf(lang.L.ShellStateReadErr, err)
		return
	}

	var toUpload []string
	if !st.FirstSyncDone || full {
		toUpload = mapKeys(current)
	} else {
		diff := state.Diff(st, current)
		toUpload = append(diff.New, diff.Changed...)
	}

	if len(effectiveInclude) > 0 {
		toUpload = filterByInclude(toUpload, effectiveInclude)
	}
	toUpload = filterByExclude(toUpload, effectiveExclude)

	// Frozen filtresi
	if fl, _ := frozen.Load(dir, srv.Name); fl != nil {
		var unfrozen []string
		frozenCount := 0
		for _, rel := range toUpload {
			if frozen.IsFrozen(fl, rel) {
				frozenCount++
			} else {
				unfrozen = append(unfrozen, rel)
			}
		}
		if frozenCount > 0 {
			fmt.Printf("  ❄ %d frozen (atlandı)\n", frozenCount)
		}
		toUpload = unfrozen
	}

	// Protect filtresi
	var filtered []string
	for _, rel := range toUpload {
		if ftpclient.IsProtected(rel, effectiveProtect) {
			continue
		}
		filtered = append(filtered, rel)
	}
	toUpload = filtered

	sort.Strings(toUpload)

	if dryRun {
		RunSyncTUI(srv.Name, len(toUpload), true, toUpload, nil)
		return
	}

	if len(toUpload) == 0 {
		RunSyncTUI(srv.Name, 0, false, nil, nil)
		return
	}

	pool, err := ftpclient.NewPool(srv)
	if err != nil {
		fmt.Printf(lang.L.ShellConnPoolErr, err)
		return
	}
	defer pool.Close()

	tasks := make([]ftpclient.UploadTask, 0, len(toUpload))
	for _, rel := range toUpload {
		f := byRel[rel]
		tasks = append(tasks, ftpclient.UploadTask{LocalPath: f.AbsPath, RelPath: rel, Hash: current[rel]})
	}

	ch := pool.UploadCh(tasks)
	tuiResults, _ := RunSyncTUI(srv.Name, len(tasks), false, nil, ch)

	successFiles := make(map[string]string)
	var failedPaths []string
	for _, r := range tuiResults {
		if r.Err != nil {
			failedPaths = append(failedPaths, r.RelPath)
		} else {
			successFiles[r.RelPath] = r.Hash
		}
	}

	if len(failedPaths) > 0 {
		failed.Save(dir, srv.Name, failedPaths)
	} else {
		failed.Clear(dir, srv.Name)
	}

	for rel, hash := range successFiles {
		st.Files[rel] = hash
	}
	for rel := range st.Files {
		if _, exists := current[rel]; !exists {
			delete(st.Files, rel)
		}
	}
	st.FirstSyncDone = true
	state.Save(dir, st)

	if len(successFiles) > 0 {
		if relDir, err := release.Create(dir, srv.Name, successFiles); err == nil {
			fmt.Printf(lang.L.ShellReleaseFmt, relDir)
		}
	}
}

// ── yardımcılar ───────────────────────────────────────────────────────────────

func (sh *shellState) resolvePath(p string) string {
	if strings.HasPrefix(p, "/") {
		// Mutlak yol — sadece temizle
		return path.Clean(p)
	}
	if p == ".." {
		parent := path.Dir(sh.remoteCwd)
		if parent == "." {
			return "/"
		}
		return parent
	}
	if p == "." || p == "" {
		return sh.remoteCwd
	}
	return path.Join(sh.remoteCwd, p)
}

func (sh *shellState) printHelp() {
	fmt.Println(lang.L.ShellHelp)
}

func splitArgs(s string) []string {
	return strings.Fields(s)
}

// fileActionMenu — dosya seçilince işlem menüsü açar.
func (sh *shellState) fileActionMenu(filePath string) {
	action, err := RunPicker(
		lang.L.ShellFileActionTitle,
		filePath,
		[]PickerItem{
			{Icon: "👁", Label: lang.L.ShellFileActionView, Desc: "cat", Value: "cat"},
			{Icon: "⬇", Label: lang.L.ShellFileActionGet, Desc: "get", Value: "get"},
			{Icon: "🗑", Label: lang.L.ShellFileActionDel, Desc: "rm", Value: "rm"},
			{Icon: "✕", Label: lang.L.ShellFileActionCancel, Desc: "", Value: ""},
		},
	)
	if err != nil || action == "" {
		return
	}
	switch action {
	case "cat":
		sh.catFile(filePath, 10)
	case "get":
		sh.getFile(filePath, "")
	case "rm":
		sh.deleteFile(filePath, false)
	}
}

// deleteFile — onay alıp dosyayı siler.
func (sh *shellState) deleteFile(remotePath string, force bool) {
	if !force {
		confirmed, err := RunConfirm(lang.L.ShellDeleteConfTitle, remotePath+" silinecek")
		if err != nil || !confirmed {
			fmt.Println(lang.L.ShellCancelShort)
			return
		}
	}
	if err := sh.client.DeleteFile(remotePath); err != nil {
		fmt.Printf(lang.L.ShellErrFmt, err)
		return
	}
	fmt.Printf(lang.L.ShellDeletedShort, remotePath)
}

// pickServerTUI — tek seçimli sunucu picker'ı.
func pickServerTUI(servers []config.Server) (*config.Server, error) {
	items := make([]PickerItem, len(servers))
	for i, s := range servers {
		items[i] = PickerItem{
			Icon:  "🖥",
			Label: s.Name,
			Desc:  fmt.Sprintf("%s:%d  (conn:%d)", s.Host, s.Port, s.MaxConnections),
			Value: s.Name,
		}
	}
	val, err := RunPicker(lang.L.ShellPickServerTitle, "", items)
	if err != nil || val == "" {
		return nil, err
	}
	for i, s := range servers {
		if s.Name == val {
			return &servers[i], nil
		}
	}
	return nil, nil
}

// ══════════════════════════════════════════════════════════════════════════════
// Status TUI
// ══════════════════════════════════════════════════════════════════════════════

type statusSrvEntry struct {
	name     string
	localDir string
	added    []string
	changed  []string
	deleted  []string
	frozen   []string // changed/added ama frozen — sync edilmeyecek
	size     int64
}

func (e statusSrvEntry) totalChanges() int {
	return len(e.added) + len(e.changed) + len(e.deleted)
}

type statusTUI struct {
	project    string
	entries    []statusSrvEntry
	cursor     int
	expanded   int // -1 = liste modu, >=0 = detay modu
	scroll     int
	searching  bool
	searchText string
	width      int
	height     int
	syncPending bool   // 's' sonrası onay bekleniyor
	syncServer  string // boş değilse TUI çıkışında bu sunucu sync edilecek
}

var (
	stTitle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5")).Padding(0, 1)
	stCursor   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	stName     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	stOk       = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	stCount    = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	stSize     = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	stHint     = lipgloss.NewStyle().Faint(true)
	stDiv      = lipgloss.NewStyle().Faint(true)
	stAdded    = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	stModified = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	stDeleted  = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
)

func newStatusTUI(project string, entries []statusSrvEntry) statusTUI {
	expanded := -1
	if len(entries) == 1 && entries[0].totalChanges() > 0 {
		expanded = 0
	}
	return statusTUI{project: project, entries: entries, expanded: expanded}
}

func (m statusTUI) Init() tea.Cmd { return nil }

func (m statusTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		if m.expanded >= 0 {
			// Arama modu
			if m.searching {
				switch msg.String() {
				case "esc":
					m.searching = false
					m.searchText = ""
					m.scroll = 0
				case "ctrl+c":
					m.searching = false
					m.searchText = ""
				case "backspace":
					r := []rune(m.searchText)
					if len(r) > 0 {
						m.searchText = string(r[:len(r)-1])
					}
					m.scroll = 0
				default:
					s := msg.String()
					for strings.HasPrefix(s, "alt+") {
						s = strings.TrimPrefix(s, "alt+")
					}
					r := []rune(s)
					if len(r) == 1 && r[0] >= 0x20 && r[0] != 0x7F {
						m.searchText += s
						m.scroll = 0
					}
				}
				return m, nil
			}
			// Sync onayı bekleniyor
			if m.syncPending {
				switch msg.String() {
				case "enter", "s", "S", "y", "Y":
					m.syncServer = m.entries[m.expanded].name
					return m, tea.Quit
				default:
					m.syncPending = false
				}
				return m, nil
			}
			// Detay modu normal tuşlar
			switch msg.String() {
			case "q", "Q", "esc", "left", "h":
				m.expanded = -1
				m.scroll = 0
				m.searchText = ""
			case "/":
				m.searching = true
				m.searchText = ""
				m.scroll = 0
			case "s", "S":
				if m.entries[m.expanded].totalChanges() > 0 {
					m.syncPending = true
				}
			case "up", "k":
				if m.scroll > 0 {
					m.scroll--
				}
			case "down", "j":
				e := m.entries[m.expanded]
				maxS := e.totalChanges() + len(e.frozen) - 1
				if m.scroll < maxS {
					m.scroll++
				}
			case "g":
				m.scroll = 0
			case "G":
				e := m.entries[m.expanded]
				maxS := e.totalChanges() + len(e.frozen) - 1
				if maxS > 0 {
					m.scroll = maxS
				}
			case "pgup":
				visRows := m.height - 7
				if visRows < 1 {
					visRows = 10
				}
				m.scroll -= visRows
				if m.scroll < 0 {
					m.scroll = 0
				}
			case "pgdown":
				e := m.entries[m.expanded]
				maxS := e.totalChanges() + len(e.frozen) - 1
				visRows := m.height - 7
				if visRows < 1 {
					visRows = 10
				}
				m.scroll += visRows
				if m.scroll > maxS {
					m.scroll = maxS
				}
			}
		} else {
			switch msg.String() {
			case "q", "Q", "esc", "ctrl+c":
				return m, tea.Quit
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.entries)-1 {
					m.cursor++
				}
			case "right", "l", "enter":
				if len(m.entries) > 0 {
					m.expanded = m.cursor
					m.scroll = 0
					m.searchText = ""
				}
			}
		}
	}
	return m, nil
}

func (m statusTUI) View() string {
	w := m.width
	if w < 60 {
		w = 80
	}
	h := m.height
	if h < 10 {
		h = 24
	}
	div := stDiv.Render(strings.Repeat("─", w))

	var b strings.Builder

	if m.expanded >= 0 {
		e := m.entries[m.expanded]
		sizeStr := ""
		if e.size > 0 {
			sizeStr = "  " + stSize.Render(formatSize(uint64(e.size)))
		}
		b.WriteString("\n")
		b.WriteString(stTitle.Render("  "+e.name) + stCount.Render(fmt.Sprintf("  %d değişiklik", e.totalChanges())) + sizeStr + "\n")
		if m.syncPending {
			b.WriteString(stCount.Render("  Sync başlatılsın?") + stHint.Render("  Enter/s = Evet  |  ESC = İptal") + "\n")
		} else if m.searching {
			b.WriteString(stHint.Render("  Backspace = sil  |  Esc = aramayı kapat") + "\n")
		} else {
			hintSync := ""
			if e.totalChanges() > 0 {
				hintSync = "  s = Sync  |"
			}
			b.WriteString(stHint.Render("  ↑↓/g/G kaydır  |  / = ara  |"+hintSync+"  ←/Esc = geri  |  q = çık") + "\n")
		}
		b.WriteString(div + "\n")

		// Tüm satırları oluştur
		var allLines []string
		for _, f := range e.added {
			allLines = append(allLines, stAdded.Render("  + ")+f)
		}
		for _, f := range e.changed {
			allLines = append(allLines, stModified.Render("  ~ ")+f)
		}
		for _, f := range e.deleted {
			allLines = append(allLines, stDeleted.Render("  - ")+f)
		}
		for _, f := range e.frozen {
			allLines = append(allLines, stHint.Render("  ❄ ")+stHint.Render(f))
		}

		// Arama filtresi: satırın görünür (ANSI soyulmuş) içeriğine göre filtrele
		lines := allLines
		if m.searchText != "" {
			q := strings.ToLower(m.searchText)
			var filtered []string
			for _, l := range allLines {
				if strings.Contains(strings.ToLower(l), q) {
					filtered = append(filtered, l)
				}
			}
			lines = filtered
		}

		bottomRows := 1 // div
		if m.searching {
			bottomRows = 2 // div + search input
		}
		visRows := h - 5 - bottomRows
		if visRows < 1 {
			visRows = 1
		}
		scroll := m.scroll
		if scroll > len(lines)-visRows {
			scroll = len(lines) - visRows
		}
		if scroll < 0 {
			scroll = 0
		}
		end := scroll + visRows
		if end > len(lines) {
			end = len(lines)
		}
		for _, l := range lines[scroll:end] {
			b.WriteString(l + "\n")
		}
		if len(lines) == 0 {
			if m.searchText != "" {
				b.WriteString(stHint.Render("  (eşleşme yok)") + "\n")
			} else {
				b.WriteString(stOk.Render("  ✓ güncel") + "\n")
			}
		}
		b.WriteString(div + "\n")
		if m.searching {
			cursor := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11")).Render("█")
			b.WriteString(stCount.Render("  / ") + m.searchText + cursor + "\n")
		} else if len(lines) > visRows {
			b.WriteString(stHint.Render(fmt.Sprintf("  %d/%d", scroll+1, len(lines))) + "\n")
		} else if m.searchText != "" {
			b.WriteString(stHint.Render(fmt.Sprintf("  %d sonuç", len(lines))) + "\n")
		}
	} else {
		b.WriteString("\n")
		b.WriteString(stTitle.Render("  Status — "+m.project) + "\n")
		b.WriteString(stHint.Render("  ↑↓ gezin  |  → = detay  |  q = çık") + "\n")
		b.WriteString(div + "\n\n")

		for i, e := range m.entries {
			n := e.totalChanges()
			var info string
			if n == 0 && len(e.frozen) == 0 {
				info = stOk.Render("✓ güncel")
			} else {
				sizeStr := ""
				if e.size > 0 {
					sizeStr = "  " + stSize.Render(formatSize(uint64(e.size)))
				}
				info = stCount.Render(fmt.Sprintf("%d değişiklik", n)) + sizeStr
				if len(e.frozen) > 0 {
					info += "  " + stHint.Render(fmt.Sprintf("❄ %d frozen", len(e.frozen)))
				}
				if n > 0 {
					info += stHint.Render("  →")
				}
			}

			if i == m.cursor {
				b.WriteString(stCursor.Render("▶ ") + stName.Render(e.name) + "  " + info + "\n")
			} else {
				b.WriteString("  " + stName.Render(e.name) + "  " + info + "\n")
			}
		}

		b.WriteString("\n" + div + "\n")
	}

	return b.String()
}

// pickServerMultiTUI — çoklu seçimli sunucu picker'ı (sync için).
// connectedName boş değilse o sunucu önceden işaretli + cursor üzerinde açılır.
func pickServerMultiTUI(servers []config.Server, connectedName string) ([]config.Server, error) {
	items := make([]PickerItem, len(servers))
	for i, s := range servers {
		items[i] = PickerItem{
			Icon:  "🖥",
			Label: s.Name,
			Desc:  fmt.Sprintf("%s:%d", s.Host, s.Port),
			Value: s.Name,
		}
	}
	var m multiPickerModel
	if connectedName != "" {
		m = newMultiPickerModelPreSelected(lang.L.ShellSyncServers, lang.L.ShellSyncPickSub, items, connectedName)
	} else {
		m = newMultiPickerModel(lang.L.ShellSyncServers, lang.L.ShellSyncPickSub, items)
	}
	vals, err := runMultiPickerRaw(m)
	if err != nil || vals == nil {
		return nil, err
	}
	nameSet := make(map[string]bool, len(vals))
	for _, v := range vals {
		nameSet[v] = true
	}
	var result []config.Server
	for _, s := range servers {
		if nameSet[s.Name] {
			result = append(result, s)
		}
	}
	return result, nil
}

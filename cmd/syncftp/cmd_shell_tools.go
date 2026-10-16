package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"syncftp/internal/config"
	"syncftp/internal/lang"
	"syncftp/internal/release"
	"syncftp/internal/synclog"
)

// shellTargetServer — bağlı sunucu varsa onu, yoksa seçiciyi kullanır (--server ad öncelikli).
func (sh *shellState) shellTargetServer(args []string) (config.Server, bool) {
	for i := 0; i < len(args); i++ {
		if args[i] == "--server" && i+1 < len(args) {
			s, ok, err := resolveOneServer(&sh.cfg, args[i+1])
			if err != nil {
				fmt.Printf(lang.L.ShellErrFmt, err)
			}
			return s, ok
		}
	}
	if sh.srv != nil {
		return *sh.srv, true
	}
	s, ok, err := resolveOneServer(&sh.cfg, "")
	if err != nil {
		fmt.Printf(lang.L.ShellErrFmt, err)
	}
	return s, ok
}

func hasFlag(args []string, flags ...string) bool {
	for _, a := range args {
		for _, f := range flags {
			if a == f {
				return true
			}
		}
	}
	return false
}

// firstPositional --flag ve değerleri dışındaki ilk argüman.
func firstPositional(args []string) string {
	for i := 0; i < len(args); i++ {
		if args[i] == "--server" {
			i++
			continue
		}
		if !strings.HasPrefix(args[i], "-") {
			return args[i]
		}
	}
	return ""
}

// cmdRollback — rollback [release-id] [--list] [--server ad]
func (sh *shellState) cmdRollback(args []string) {
	srv, ok := sh.shellTargetServer(args)
	if !ok {
		return
	}
	list, err := release.List(sh.configDir, srv.Name)
	if err != nil {
		fmt.Printf(lang.L.ShellErrFmt, err)
		return
	}
	if len(list) == 0 {
		fmt.Printf(lang.L.RollbackNoReleasesFmt, srv.Name)
		return
	}
	if hasFlag(args, "--list", "-l") {
		printReleaseList(os.Stdout, list)
		return
	}
	target := firstPositional(args)
	if target == "" {
		var items []PickerItem
		for _, m := range list {
			if m.RolledBack || !m.HasBackup() {
				continue
			}
			items = append(items, PickerItem{Icon: "↺", Label: m.ID, Desc: releaseDesc(m), Value: m.ID})
		}
		if len(items) == 0 {
			fmt.Printf(lang.L.RollbackNoReleasesFmt, srv.Name)
			return
		}
		target, _ = RunPicker(fmt.Sprintf(lang.L.RollbackPickTitleFmt, srv.Name), lang.L.RollbackPickSubtitle, items)
		if target == "" {
			fmt.Println(lang.L.ShellSyncCancelled)
			return
		}
	}
	plan, err := rollbackPlan(list, target)
	if err != nil {
		fmt.Printf(lang.L.ShellErrFmt, err)
		return
	}
	var sb strings.Builder
	for _, m := range plan {
		fmt.Fprintf(&sb, "%s  %s\n", m.ID, releaseDesc(m))
	}
	if okc, _ := RunConfirm(fmt.Sprintf(lang.L.RollbackConfirmFmt, len(plan), srv.Name), sb.String()); !okc {
		fmt.Println(lang.L.ShellSyncCancelled)
		return
	}
	var buf strings.Builder
	if _, err := performRollback(io.MultiWriter(os.Stdout, &buf), sh.configDir, srv, plan); err != nil {
		fmt.Printf(lang.L.ShellErrFmt, err)
		return
	}
	if p, err := synclog.Save(sh.configDir, srv.Name, buf.String()); err == nil {
		fmt.Printf(lang.L.SyncLogSavedFmt, p)
	}
}

// cmdVerify — verify [--fix] [--server ad]
func (sh *shellState) cmdVerify(args []string) {
	srv, ok := sh.shellTargetServer(args)
	if !ok {
		return
	}
	if _, err := verifyServer(os.Stdout, sh.configDir, &sh.cfg, srv, hasFlag(args, "--fix")); err != nil {
		fmt.Printf(lang.L.VerifyErrFmt, srv.Name, err)
	}
}

// cmdDetectLimit — detect-limit [--server ad] [--max N]
// Shell'in kendi bağlantısı ölçümü bir eksik gösterir → ölçüm boyunca kapatılır,
// sonra loop başında güvenli şekilde yeniden bağlanılır (pendingReconnect).
func (sh *shellState) cmdDetectLimit(args []string) {
	srv, ok := sh.shellTargetServer(args)
	if !ok {
		return
	}
	maxProbe := 32
	for i := 0; i < len(args); i++ {
		if args[i] == "--max" && i+1 < len(args) {
			fmt.Sscanf(args[i+1], "%d", &maxProbe)
			i++
		}
	}
	if sh.client != nil {
		sh.client.Close()
		sh.client = nil
		sh.pendingReconnect = true
	}
	if detectLimits(os.Stdout, &sh.cfg, []config.Server{srv}, maxProbe, func(n int) {
		fmt.Printf(lang.L.DetectProgressFmt, n)
	}) {
		if err := config.Save(sh.configDir, &sh.cfg); err != nil {
			fmt.Printf(lang.L.ShellErrFmt, err)
			return
		}
		// Bağlı sunucunun kopyasını da güncelle
		if sh.srv != nil {
			for _, s := range sh.cfg.Servers {
				if s.Name == sh.srv.Name {
					cp := s
					sh.srv = &cp
				}
			}
		}
		fmt.Print(lang.L.DetectSaved)
	}
}

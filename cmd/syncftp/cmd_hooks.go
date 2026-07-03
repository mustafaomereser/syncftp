package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"syncftp/internal/lang"
)

// runHooks komutları sırayla çalıştırır; çıktılarını girintiyle w'ye yazar.
// Bir komut hata verirse kalanlar çalıştırılmaz ve hata döner.
// Komutlar workDir içinde, sistem shell'i üzerinden çalışır (cmd /C veya sh -c).
func runHooks(w io.Writer, workDir string, cmds []string, extraEnv []string) error {
	for _, c := range cmds {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		fmt.Fprintf(w, lang.L.SyncHookRunFmt, c)

		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.Command("cmd", "/C", c)
		} else {
			cmd = exec.Command("sh", "-c", c)
		}
		cmd.Dir = workDir
		cmd.Env = append(os.Environ(), extraEnv...)

		out, err := cmd.CombinedOutput()
		for _, line := range strings.Split(strings.TrimRight(string(out), "\r\n"), "\n") {
			if line = strings.TrimRight(line, "\r"); line != "" {
				fmt.Fprintf(w, "      %s\n", line)
			}
		}
		if err != nil {
			return fmt.Errorf("%s: %w", c, err)
		}
	}
	return nil
}

// hookEnv sync bağlamını hook process'ine env var olarak taşır.
func hookEnv(serverName string, uploaded, failed int) []string {
	return []string{
		"SYNCFTP_SERVER=" + serverName,
		fmt.Sprintf("SYNCFTP_UPLOADED=%d", uploaded),
		fmt.Sprintf("SYNCFTP_FAILED=%d", failed),
	}
}

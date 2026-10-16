package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"syncftp/internal/config"
	ftpclient "syncftp/internal/ftp"
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

// runRemoteHooks komutları SFTP sunucusunda SSH ile, uzak dizinde (remote_path) sırayla çalıştırır.
// İlk hata kalanları durdurur. FTP/FTPS sunucularında SSH olmadığından atlanır (uyarı yazılır).
// Ortam: SYNCFTP_SERVER / SYNCFTP_UPLOADED / SYNCFTP_FAILED (yerel hook'larla aynı).
func runRemoteHooks(w io.Writer, srv config.Server, cmds []string, uploaded, failed int) error {
	var list []string
	for _, c := range cmds {
		if c = strings.TrimSpace(c); c != "" {
			list = append(list, c)
		}
	}
	if len(list) == 0 {
		return nil
	}
	if !srv.IsSFTP() {
		fmt.Fprintf(w, lang.L.RemoteHookSkipFmt, srv.Name)
		return nil
	}
	client, err := ftpclient.Connect(srv)
	if err != nil {
		return err
	}
	defer client.Close()
	env := map[string]string{}
	for _, kv := range hookEnv(srv.Name, uploaded, failed) {
		if k, v, ok := strings.Cut(kv, "="); ok {
			env[k] = v
		}
	}
	for _, c := range list {
		fmt.Fprintf(w, lang.L.RemoteHookRunFmt, c)
		if err := client.Exec(c, env, indentWriter6(w)); err != nil {
			return fmt.Errorf("%s: %w", c, err)
		}
	}
	return nil
}

// indentWriter6 uzak komut çıktısını yerel hook'larla aynı girintiyle (6 boşluk) yazar.
type indent6 struct {
	w       io.Writer
	partial string
}

func indentWriter6(w io.Writer) io.Writer { return &indent6{w: w} }

func (i *indent6) Write(p []byte) (int, error) {
	text := i.partial + strings.ReplaceAll(string(p), "\r\n", "\n")
	parts := strings.Split(text, "\n")
	i.partial = parts[len(parts)-1]
	for _, l := range parts[:len(parts)-1] {
		fmt.Fprintf(i.w, "      %s\n", l)
	}
	return len(p), nil
}

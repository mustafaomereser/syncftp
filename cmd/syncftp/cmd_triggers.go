package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"syncftp/internal/config"
	"syncftp/internal/lang"
	"syncftp/internal/pattern"
)

// Dosya tetikleyicileri — belirli dosyalar yüklendiğinde (ya da yüklenmeden hemen önce) yerel komut,
// SSH komutu, HTTP isteği ve Slack/Discord bildirimi. Sync çekirdeği (syncToServerOpts) çağırır;
// CLI `trigger test` ve panelin "Dene" işi aynı runTrigger'ı kullanır.

// firedTrigger eşleşen tetikleyici + onu tetikleyen dosyalar.
type firedTrigger struct {
	t     config.Trigger
	files []string
}

// serverTriggers global + sunucuya özel (global önce).
func serverTriggers(cfg *config.Config, srv config.Server) []config.Trigger {
	return append(append([]config.Trigger{}, cfg.Sync.Triggers...), srv.Triggers...)
}

// matchTriggers açık, zamanı tutan ve dosyalardan en az birine uyan tetikleyiciler.
func matchTriggers(list []config.Trigger, before bool, files []string) []firedTrigger {
	var out []firedTrigger
	for _, t := range list {
		if !t.IsEnabled() || t.IsBefore() != before || !t.HasAction() {
			continue
		}
		var hit []string
		for _, f := range files {
			if len(t.Paths) == 0 || pattern.Any(f, t.Paths) {
				hit = append(hit, f)
			}
		}
		if len(hit) > 0 {
			out = append(out, firedTrigger{t: t, files: hit})
		}
	}
	return out
}

// previewTriggers dry-run'da hangi tetikleyicilerin çalışacağını yazar (eylem yok).
func previewTriggers(w io.Writer, cfg *config.Config, srv config.Server, files []string) {
	for _, before := range []bool{true, false} {
		when := lang.L.TriggerAfter
		if before {
			when = lang.L.TriggerBefore
		}
		for _, f := range matchTriggers(serverTriggers(cfg, srv), before, files) {
			fmt.Fprintf(w, lang.L.TriggerDryFmt, f.t.Name, when, len(f.files))
		}
	}
}

// runTriggers eşleşen tetikleyicileri sırayla çalıştırır. before=true: ilk hata döner (sync iptal);
// after: hatalar yazılır, kalanlar yine çalışır, nil döner.
func runTriggers(w io.Writer, configDir string, cfg *config.Config, srv config.Server, before bool, files []string) error {
	for _, f := range matchTriggers(serverTriggers(cfg, srv), before, files) {
		fmt.Fprintf(w, lang.L.TriggerRunFmt, f.t.Name, len(f.files))
		if err := runTrigger(w, configDir, cfg.Project.Name, srv, f.t, f.files); err != nil {
			if before {
				fmt.Fprintf(w, lang.L.TriggerAbortFmt, f.t.Name, err)
				return err
			}
			fmt.Fprintf(w, lang.L.TriggerErrFmt, f.t.Name, err)
		}
	}
	return nil
}

// triggerEnvMax ortam değişkenine yazılan dosya listesi sınırı (Windows ortam bloğu ~32 K).
const triggerEnvMax = 300

// runTrigger tek tetikleyicinin eylemleri, sırayla: yerel komut → SSH → HTTP → bildirim. İlk hata durdurur.
func runTrigger(w io.Writer, configDir, project string, srv config.Server, t config.Trigger, files []string) error {
	listed := files
	if len(listed) > triggerEnvMax {
		listed = listed[:triggerEnvMax]
	}
	env := append(hookEnv(srv.Name, len(files), 0),
		"SYNCFTP_PROJECT="+project,
		"SYNCFTP_TRIGGER="+t.Name,
		fmt.Sprintf("SYNCFTP_TRIGGER_COUNT=%d", len(files)),
		"SYNCFTP_TRIGGER_FILES="+strings.Join(listed, "\n"),
	)
	if cmds := trimmed(t.Run); len(cmds) > 0 {
		if err := runHooks(w, configDir, cmds, env); err != nil {
			return err
		}
	}
	if cmds := trimmed(t.Remote); len(cmds) > 0 {
		if err := runRemoteCmds(w, srv, cmds, env); err != nil {
			return err
		}
	}
	vars := triggerVars{project: project, server: srv.Name, trigger: t.Name, files: files}
	if strings.TrimSpace(t.HTTPURL) != "" {
		if err := triggerHTTP(w, t, vars); err != nil {
			return err
		}
	}
	if strings.TrimSpace(t.Notify) != "" {
		if err := triggerNotify(w, t, vars); err != nil {
			return err
		}
	}
	return nil
}

func trimmed(list []string) []string {
	var out []string
	for _, s := range list {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// triggerVars HTTP adres/başlık/gövdesindeki yer tutucular.
type triggerVars struct {
	project, server, trigger string
	files                    []string
}

// expand {project} {server} {trigger} {count} {files} {files_json} {first}. inURL → değerler adres
// için kaçışlanır; inJSON → tırnak içinde kullanılacak şekilde (json string içeriği) kaçışlanır.
func (v triggerVars) expand(s string, inURL, inJSON bool) string {
	esc := func(x string) string {
		switch {
		case inURL:
			return url.QueryEscape(x)
		case inJSON:
			b, _ := json.Marshal(x)
			return string(b[1 : len(b)-1])
		}
		return x
	}
	first := ""
	if len(v.files) > 0 {
		first = v.files[0]
	}
	fj, _ := json.Marshal(v.files)
	if v.files == nil {
		fj = []byte("[]")
	}
	return strings.NewReplacer(
		"{project}", esc(v.project),
		"{server}", esc(v.server),
		"{trigger}", esc(v.trigger),
		"{count}", fmt.Sprint(len(v.files)),
		"{first}", esc(first),
		"{files_json}", string(fj),
		"{files}", esc(strings.Join(v.files, "\n")),
	).Replace(s)
}

// triggerClient — migrate gibi uzun süren uçlara yetecek kadar, ama sync'i sonsuza kilitlemeyecek kadar.
var triggerClient = &http.Client{Timeout: 2 * time.Minute}

func triggerHTTP(w io.Writer, t config.Trigger, v triggerVars) error {
	method := strings.ToUpper(strings.TrimSpace(t.HTTPMethod))
	if method == "" {
		method = http.MethodPost
	}
	target := v.expand(strings.TrimSpace(t.HTTPURL), true, false)
	body := strings.TrimSpace(t.HTTPBody)
	isJSON := body == "" || strings.HasPrefix(body, "{") || strings.HasPrefix(body, "[")
	if body == "" && method != http.MethodGet {
		b, _ := json.Marshal(map[string]any{"trigger": v.trigger, "project": v.project, "server": v.server, "count": len(v.files), "files": v.files})
		body = string(b)
	} else {
		body = v.expand(body, false, isJSON)
	}
	var rd io.Reader
	if method != http.MethodGet && body != "" {
		rd = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, target, rd)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "syncFTP")
	if rd != nil && isJSON {
		req.Header.Set("Content-Type", "application/json")
	}
	for _, h := range t.HTTPHeaders {
		if k, val, ok := strings.Cut(h, ":"); ok && strings.TrimSpace(k) != "" {
			req.Header.Set(strings.TrimSpace(k), v.expand(strings.TrimSpace(val), false, false))
		}
	}
	resp, err := triggerClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	fmt.Fprintf(w, lang.L.TriggerHTTPFmt, method, redactURL(target), resp.StatusCode)
	if s := strings.TrimSpace(string(out)); s != "" {
		for _, line := range strings.Split(s, "\n") {
			fmt.Fprintf(w, "        %s\n", strings.TrimRight(line, "\r"))
		}
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return nil
}

// redactURL log'a sorgu parametrelerindeki değerler (token=…) girmesin.
func redactURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.RawQuery == "" {
		return raw
	}
	q := u.Query()
	for k := range q {
		q.Set(k, "***")
	}
	u.RawQuery = strings.ReplaceAll(q.Encode(), "%2A", "*")
	return u.String()
}

// triggerNotify Slack / Discord / genel webhook — sync webhook'uyla aynı biçim seçimi.
func triggerNotify(w io.Writer, t config.Trigger, v triggerVars) error {
	summary := fmt.Sprintf("syncFTP · %s · %s — ⚡ %s (%d)", v.project, v.server, v.trigger, len(v.files))
	shown := v.files
	if len(shown) > 10 {
		shown = shown[:10]
	}
	if len(shown) > 0 {
		summary += "\n" + strings.Join(shown, "\n")
	}
	if len(v.files) > 10 {
		summary += fmt.Sprintf("\n… +%d", len(v.files)-10)
	}
	target := strings.TrimSpace(t.Notify)
	u := strings.ToLower(target)
	var payload string
	switch {
	case strings.Contains(u, "hooks.slack.com"):
		payload = fmt.Sprintf(`{"text":%s}`, jsonQuote(summary))
	case strings.Contains(u, "discord.com/api/webhooks") || strings.Contains(u, "discordapp.com/api/webhooks"):
		payload = fmt.Sprintf(`{"content":%s}`, jsonQuote(summary))
	default:
		b, _ := json.Marshal(map[string]any{"text": summary, "trigger": v.trigger, "project": v.project, "server": v.server, "files": v.files})
		payload = string(b)
	}
	resp, err := webhookClient.Post(target, "application/json", bytes.NewBufferString(payload))
	if err != nil {
		return err
	}
	resp.Body.Close()
	fmt.Fprintf(w, lang.L.TriggerNotifyFmt, resp.StatusCode)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return nil
}

// findTrigger ada göre (büyük/küçük harf duyarsız) — önce sunucunun, sonra global liste.
func findTrigger(cfg *config.Config, srv config.Server, name string) (config.Trigger, bool) {
	for _, list := range [][]config.Trigger{srv.Triggers, cfg.Sync.Triggers} {
		for _, t := range list {
			if strings.EqualFold(strings.TrimSpace(t.Name), strings.TrimSpace(name)) {
				return t, true
			}
		}
	}
	return config.Trigger{}, false
}

// ── CLI ───────────────────────────────────────────────────────────────────────

var trigFlagServer string

func init() {
	triggerCmd.PersistentFlags().StringVarP(&trigFlagServer, "server", "s", "", "Sunucu adı")
	triggerCmd.AddCommand(triggerListCmd, triggerTestCmd)
	rootCmd.AddCommand(triggerCmd)
}

var triggerCmd = &cobra.Command{
	Use:   "trigger",
	Short: "Dosya tetikleyicileri: belirli dosyalar yüklenince komut / SSH / HTTP / bildirim",
	Long: `Tetikleyiciler config'te sync.triggers (tüm sunucular) ve servers[].triggers altında durur:

  {"name": "migrate", "when": "after", "paths": ["database/migrations/"],
   "remote": ["php artisan migrate --force"],
   "http_url": "https://site.com/deploy/migrate?token=…", "notify": "https://hooks.slack.com/…"}

when: "after" (varsayılan, hata sync sonucunu bozmaz) | "before" (hata → o sunucunun sync'i iptal).
Komutlara SYNCFTP_TRIGGER, SYNCFTP_TRIGGER_COUNT, SYNCFTP_TRIGGER_FILES (satır satır) geçer.
HTTP adres/başlık/gövdesinde {files} {files_json} {count} {first} {server} {project} {trigger}.`,
}

var triggerListCmd = &cobra.Command{
	Use:   "list",
	Short: "Tetikleyicileri listele",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := os.Getwd()
		cfg, err := config.Load(dir)
		if err != nil {
			return err
		}
		servers := cfg.Servers
		if trigFlagServer != "" {
			s, ok, err := resolveOneServer(cfg, trigFlagServer)
			if err != nil || !ok {
				return err
			}
			servers = []config.Server{s}
		}
		for _, srv := range servers {
			list := serverTriggers(cfg, srv)
			if len(list) == 0 {
				fmt.Printf(lang.L.TriggerNoneFmt, srv.Name)
				continue
			}
			fmt.Printf("%s\n", srv.Name)
			for _, t := range list {
				when := lang.L.TriggerAfter
				if t.IsBefore() {
					when = lang.L.TriggerBefore
				}
				name := t.Name
				if !t.IsEnabled() {
					name += lang.L.TriggerOffSuffix
				}
				pats := strings.Join(t.Paths, ", ")
				if pats == "" {
					pats = "*"
				}
				fmt.Printf(lang.L.TriggerListRowFmt, name, when, pats)
			}
		}
		return nil
	},
}

var triggerTestCmd = &cobra.Command{
	Use:   "test <name> [file...]",
	Short: "Tetikleyiciyi sync olmadan çalıştır (dosya verilmezse boş listeyle)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := os.Getwd()
		cfg, err := config.Load(dir)
		if err != nil {
			return err
		}
		srv, ok, err := resolveOneServer(cfg, trigFlagServer)
		if err != nil || !ok {
			return err
		}
		t, found := findTrigger(cfg, srv, args[0])
		if !found {
			return fmt.Errorf(lang.L.TriggerNotFoundFmt, args[0])
		}
		files := args[1:]
		fmt.Printf(lang.L.TriggerRunFmt, t.Name, len(files))
		return runTrigger(os.Stdout, dir, cfg.Project.Name, srv, t, files)
	},
}

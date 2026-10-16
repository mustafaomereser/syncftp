package main

import (
	"bufio"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"syncftp/internal/config"
	"syncftp/internal/lang"
	"syncftp/internal/projects"
	"syncftp/internal/release"
	"syncftp/internal/state"
)

// ══════════════════════════════════════════════════════════════════════════════
// Yapay zekâ asistanı
//
// Panel, makinede yüklü AI kodlama araçlarını (Claude Code, OpenCode, Codex, Gemini…)
// algılar. Başsız (headless) modu doğrulanmış araçlarla sohbet PANEL İÇİNDE yürür:
// her mesaj aracı proje dizininde bir tur çalıştırır, oturum kimliğiyle sürdürülür,
// olaylar (metin / araç kullanımı / sonuç) SSE ile akar. Diğer araçlar yeni bir
// terminal penceresinde, syncFTP dokümanlarını okumaya yönlendiren bir ilk mesajla açılır.
//
// Asistanın bilgi kaynağı .syncftp/ai/ altında her oturumda yeniden üretilen markdown'lar:
// README (kurallar), CLI (cobra'dan), CONFIG (şema), PROJECT (şifresiz anlık durum).
// İzinler: dosya düzenleme + yalnızca `syncftp` komutları (rastgele shell yok).
// ══════════════════════════════════════════════════════════════════════════════

type aiAgent struct {
	ID    string   `json:"id"`
	Name  string   `json:"name"`
	Bins  []string `json:"-"`
	Chat  bool     `json:"chat"` // panel içi sohbet destekleniyor (başsız mod doğrulandı)
	Path  string   `json:"path,omitempty"`
	Found bool     `json:"found"`
}

func aiAgentCatalog() []aiAgent {
	return []aiAgent{
		{ID: "claude", Name: "Claude Code", Bins: []string{"claude"}, Chat: true},
		{ID: "opencode", Name: "OpenCode", Bins: []string{"opencode.cmd", "opencode"}, Chat: true},
		{ID: "codex", Name: "Codex CLI", Bins: []string{"codex.cmd", "codex"}},
		{ID: "gemini", Name: "Gemini CLI", Bins: []string{"gemini.cmd", "gemini"}},
		{ID: "qwen", Name: "Qwen Code", Bins: []string{"qwen.cmd", "qwen"}},
		{ID: "cursor-agent", Name: "Cursor Agent", Bins: []string{"cursor-agent.cmd", "cursor-agent"}},
		{ID: "aider", Name: "Aider", Bins: []string{"aider"}},
	}
}

func detectAIAgents() []aiAgent {
	out := aiAgentCatalog()
	for i := range out {
		for _, b := range out[i].Bins {
			if p, err := exec.LookPath(b); err == nil {
				out[i].Path, out[i].Found = p, true
				break
			}
		}
	}
	return out
}

// resolveCmdShim npm'in Windows .cmd sarmalayıcısını gerçek hedefe çevirir. .cmd üzerinden
// çalıştırınca cmd.exe argümanlardaki yeni satır, %, &, | gibi karakterleri yorumlar ve mesaj
// bozulur/kırpılır — sarmalayıcı atlanıp hedef (.exe ya da `node script.js`) doğrudan çalıştırılır.
func resolveCmdShim(p string) (bin string, pre []string) {
	if !strings.EqualFold(filepath.Ext(p), ".cmd") {
		return p, nil
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return p, nil
	}
	dir := filepath.Dir(p)
	for _, m := range shimTargetRe.FindAllStringSubmatch(string(data), -1) {
		target := filepath.Join(dir, strings.ReplaceAll(m[1], `\`, string(filepath.Separator)))
		if !fileExistsPlain(target) {
			continue
		}
		switch strings.ToLower(filepath.Ext(target)) {
		case ".exe":
			return target, nil
		case ".js", ".cjs", ".mjs", "":
			node := filepath.Join(dir, "node.exe")
			if !fileExistsPlain(node) {
				if np, err := exec.LookPath("node"); err == nil {
					node = np
				} else {
					continue
				}
			}
			return node, []string{target}
		}
	}
	return p, nil
}

var shimTargetRe = regexp.MustCompile(`"%(?:~)?dp0%\\?([^"%]+)"`)

func findAgent(id string) (aiAgent, bool) {
	for _, a := range detectAIAgents() {
		if a.ID == id && a.Found {
			return a, true
		}
	}
	return aiAgent{}, false
}

// ── Dokümanlar ────────────────────────────────────────────────────────────────

// writeAIDocs .syncftp/ai/ altına asistanın okuyacağı dokümanları yazar; README yolunu döner.
func writeAIDocs(p projects.Project) (string, error) {
	dir := filepath.Join(p.Path, ".syncftp", "ai")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	exe, _ := os.Executable()
	files := map[string]string{
		"README.md":  aiReadme(p, exe),
		"CLI.md":     aiCLIDoc(),
		"CONFIG.md":  aiConfigDoc,
		"PROJECT.md": aiProjectDoc(p),
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0600); err != nil {
			return "", err
		}
	}
	return filepath.Join(dir, "README.md"), nil
}

func aiReadme(p projects.Project, exe string) string {
	userLang := "Turkish"
	if lang.Current() == "en" {
		userLang = "English"
	}
	return fmt.Sprintf(`# syncFTP assistant guide

You are helping a user manage **syncFTP** — a tool that deploys changed files (SHA256 change
detection) from a local project to FTP / FTPS / SFTP servers, with backups/rollback, watch mode,
CI/CD webhooks and a desktop panel. The user talks to you from the syncFTP panel.

- Project: **%s**
- Project folder (your working directory): %s
- syncFTP executable: %s  (also on PATH as `+"`syncftp`"+`)
- Reply in **%s**. Be concise and practical.

## Files to read
- CLI.md — every command and flag (generated from the program itself)
- CONFIG.md — the .syncftp/syncftp.json schema with examples
- PROJECT.md — current servers, settings, state and history of THIS project (secrets removed)

## How to act
- To change settings, edit `+"`.syncftp/syncftp.json`"+` directly (keep valid JSON, keep existing fields
  you don't touch) — the panel picks changes up immediately. Re-read the file before editing.
- To inspect or run things use the CLI: `+"`syncftp status`, `syncftp sync --server NAME --dry-run`, `syncftp verify -s NAME`, `syncftp hook list`"+` …
  Always pass `+"`--server NAME` / `--all` / `--yes`"+` style flags: commands must not wait for interactive input.
- Prefer `+"`--dry-run`"+` first and show the user what would happen.

## Rules
1. NEVER print, echo or repeat passwords, secrets, tokens or private keys (password, tunnel_password,
   deploy[].secret). If the user asks to set a password, write it into the config without echoing it.
2. Ask for explicit confirmation in chat BEFORE anything destructive or outward-facing: a real
   (non dry-run) sync/deploy, rollback, deleting remote files (`+"`remote rm`"+`), removing servers,
   `+"`remote exec`"+` commands that change the server.
3. Don't edit files outside the project folder. Don't touch .syncftp/state, releases, backups.
4. After changing the config, run `+"`syncftp status`"+` (or a dry-run) to confirm it still loads.
`, p.Name, p.Path, exe, userLang)
}

func aiCLIDoc() string {
	var b strings.Builder
	b.WriteString("# syncFTP command reference\n\nGenerated from the program. Global flag: `-P, --project NAME` runs a command in another registered project.\n\n")
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		if c.Hidden || c.Name() == "help" || c.Name() == "completion" {
			return
		}
		if c != rootCmd {
			fmt.Fprintf(&b, "## %s\n\n%s\n\n```\n%s\n```\n", c.CommandPath(), c.Short, c.UseLine())
			if l := strings.TrimSpace(c.Long); l != "" {
				fmt.Fprintf(&b, "\n%s\n", l)
			}
			if f := strings.TrimRight(c.LocalFlags().FlagUsages(), "\n"); f != "" {
				fmt.Fprintf(&b, "\nFlags:\n```\n%s\n```\n", f)
			}
			b.WriteString("\n")
		}
		for _, sub := range c.Commands() {
			walk(sub)
		}
	}
	walk(rootCmd)
	return b.String()
}

// aiProjectDoc projenin şifresiz anlık özeti.
func aiProjectDoc(p projects.Project) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Project snapshot — %s\n\nGenerated %s. Secrets are removed.\n\n", p.Name, time.Now().Format("2006-01-02 15:04"))
	cfg, err := config.Load(p.Path)
	if err != nil {
		fmt.Fprintf(&b, "Config could not be loaded: %v\n", err)
		return b.String()
	}
	redact := cfg.Redacted(config.MaskStars, true)
	data, _ := json.MarshalIndent(redact, "", "  ")
	fmt.Fprintf(&b, "## Effective config (connection profiles resolved, secrets as ***)\n\n```json\n%s\n```\n\n## Servers\n\n", data)
	b.WriteString("| server | protocol | host | remote | local | enabled | first sync | last sync | tracked files | releases |\n|---|---|---|---|---|---|---|---|---|---|\n")
	for _, s := range cfg.Servers {
		st, _ := state.Load(p.Path, s.Name)
		last, tracked, first := "-", 0, false
		if st != nil {
			if !st.LastSync.IsZero() {
				last = st.LastSync.Format("2006-01-02 15:04")
			}
			tracked, first = len(st.Files), st.FirstSyncDone
		}
		rels, _ := release.List(p.Path, s.Name)
		fmt.Fprintf(&b, "| %s | %s | %s:%d | %s | %s | %v | %v | %s | %d | %d |\n",
			s.Name, s.ProtocolName(), s.Host, s.Port, s.RemotePath, s.EffectiveLocalPath(cfg.Project), s.Enabled, first, last, tracked, len(rels))
	}
	b.WriteString("\nUse `syncftp status` for pending changes and `syncftp rollback --list -s NAME` for history.\n")
	return b.String()
}

const aiConfigDoc = "# .syncftp/syncftp.json schema\n\n" +
	"```jsonc\n" +
	`{
  "project": { "name": "site", "default_path": ".", "keyring": false },
  "sync": {
    "protect": [".env"],            // never overwrite these on any server (dir: "storage/")
    "include": [],                  // whitelist (empty = everything)
    "exclude": ["docs"],            // blacklist (exact path or dir prefix)
    "ignore_files": null,           // null = .gitignore + .syncftp/syncftp.ignore
    "block_extensions": [".zip"],   // never upload these extensions
    "webhook": "",                  // POST after each sync (Slack/Discord URLs auto-formatted)
    "pre_sync": ["npm run build"],  // LOCAL commands before upload (failure aborts)
    "post_sync": [],                // LOCAL commands after upload
    "bandwidth_limit": 0,           // KB/s per server, 0 = unlimited
    "no_backup": false,             // true disables backups → no rollback
    "backup_keep": 10,              // releases that keep backup files
    "direct_upload": false          // true disables atomic temp+rename upload
  },
  "first_sync": { "full": false },  // false = first sync calibrates instead of uploading everything
  "connections": [                  // shared credential profiles (optional)
    { "name": "host1", "protocol": "sftp", "host": "example.com", "port": 22, "user": "deploy",
      "password": "", "key_file": "~/.ssh/id_ed25519", "passive": true }
  ],
  "servers": [
    {
      "name": "production",
      "connection": "host1",        // when set, host/port/user/password/protocol/key/tunnel come from the profile
      "protocol": "ftp",            // "ftp" | "ftps" | "sftp" (ignored when connection is set)
      "host": "", "port": 21, "user": "", "password": "",
      "key_file": "",               // SFTP private key; encrypted keys use password as passphrase
      "tls_skip_verify": false,     // FTPS only
      "passive": true, "disable_epsv": false, "nat_workaround": false,
      "tunnel_host": "", "tunnel_port": 22, "tunnel_user": "", "tunnel_password": "", "tunnel_key_file": "",
      "remote_path": "/public_html",
      "local_path": "",             // relative to project root; empty = project.default_path
      "enabled": true,
      "max_connections": 4,         // use "syncftp detect-limit -s NAME" to measure
      "max_retries": 2,
      "include": [], "exclude": [], "protect": [], "block_extensions": [],
      "webhook": "", "minify": false, "obfuscate": false,
      "pre_sync": [], "post_sync": [],
      "remote_pre_sync": [],        // SFTP only: commands run ON THE SERVER over SSH before upload
      "remote_post_sync": ["php artisan migrate --force"], // SFTP only: after upload
      "bandwidth_limit": 0, "no_backup": false, "direct_upload": false
    }
  ],
  "deploy": [                       // CI/CD: incoming webhooks → [git pull] → sync
    { "name": "main", "provider": "github", "secret": "<random hex>", "branch": "main",
      "servers": [], "git_pull": true, "dry_run": false, "enabled": true }
  ]
}
` + "```\n\n" +
	"Notes:\n" +
	"- Server names must be unique; a server with `connection` must reference an existing profile.\n" +
	"- `deploy[].provider`: github | gitlab | gitea | bitbucket | generic. Webhook URL: `<public>/hook/<project>/<name>` (see the CI/CD tab).\n" +
	"- Ignore rules (gitignore syntax) live in `.syncftp/syncftp.ignore`.\n" +
	"- Freeze lists: `syncftp freeze`; milestones: `syncftp milestone set --date ... -s NAME`.\n"

// ── Oturumlar ─────────────────────────────────────────────────────────────────

type aiMsg struct {
	ID     int    `json:"id"`
	Role   string `json:"role"` // user | assistant | tool | error | info
	Text   string `json:"text"`
	Tool   string `json:"tool,omitempty"`
	Detail string `json:"detail,omitempty"` // araç çıktısı (kısaltılmış)
	Time   int64  `json:"time"`
}

type aiSession struct {
	ID        string  `json:"id"`
	Project   string  `json:"project"`
	Agent     string  `json:"agent"`
	AgentName string  `json:"agent_name"`
	Busy      bool    `json:"busy"`
	Messages  []aiMsg `json:"messages"`
	Cost      float64 `json:"cost"`
	Created   int64   `json:"created"`
	Setup     bool    `json:"setup"`  // proje kurulum yardımcısı (proje henüz yok)
	Folder    string  `json:"folder"` // kurulum modunda hedef klasör

	agentSession string // aracın kendi oturum kimliği
	started      bool
	mu           sync.Mutex
	subs         map[chan aiMsg]struct{}
	cmd          *exec.Cmd
	seq          int
	dir          string
	bin          string
	readme       string
}

type aiManager struct {
	mu       sync.Mutex
	sessions map[string]*aiSession
}

func newAIManager() *aiManager { return &aiManager{sessions: map[string]*aiSession{}} }

func uuid4() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func (s *aiSession) emit(m aiMsg) {
	s.mu.Lock()
	s.seq++
	m.ID = s.seq
	m.Time = time.Now().UnixMilli()
	s.Messages = append(s.Messages, m)
	if len(s.Messages) > 1000 {
		s.Messages = s.Messages[len(s.Messages)-1000:]
	}
	for ch := range s.subs {
		select {
		case ch <- m:
		default:
		}
	}
	s.mu.Unlock()
}

func (s *aiSession) setBusy(b bool) {
	s.mu.Lock()
	s.Busy = b
	s.mu.Unlock()
	s.emit(aiMsg{Role: "state", Text: map[bool]string{true: "busy", false: "idle"}[b]})
}

func (s *aiSession) snapshot() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	return map[string]any{"id": s.ID, "project": s.Project, "agent": s.Agent, "agent_name": s.AgentName,
		"busy": s.Busy, "cost": s.Cost, "messages": append([]aiMsg(nil), s.Messages...)}
}

// aiSystemPrompt araca eklenen talimat (Claude: --append-system-prompt; diğerleri: ilk mesaja eklenir).
func aiSystemPrompt(readme string) string {
	return "You are the syncFTP assistant inside the syncFTP desktop panel. Before answering anything about this project, read " +
		readme + " and the files it lists (CLI.md, CONFIG.md, PROJECT.md in the same folder). Follow its rules strictly: never reveal secrets, " +
		"ask for confirmation in chat before destructive or outward-facing actions, use the syncftp CLI non-interactively, edit .syncftp/syncftp.json carefully."
}

// send kullanıcı mesajı için aracı bir tur çalıştırır (asenkron).
func (s *aiSession) send(text string) error {
	s.mu.Lock()
	if s.Busy {
		s.mu.Unlock()
		return errors.New(lang.L.AIBusy)
	}
	s.Busy = true
	first := !s.started
	s.started = true
	s.mu.Unlock()

	s.emit(aiMsg{Role: "user", Text: text})
	s.emit(aiMsg{Role: "state", Text: "busy"})

	var args []string
	switch s.Agent {
	case "claude":
		args = []string{"-p", text, "--output-format", "stream-json", "--verbose",
			"--permission-mode", "acceptEdits",
			// Windows'ta Claude Code komutları PowerShell aracıyla da çalıştırır → ikisine de izin
			"--allowedTools", "Bash(syncftp *) Bash(syncftp.exe *) PowerShell(syncftp *) PowerShell(syncftp.exe *) Read Edit Write Glob Grep",
			"--append-system-prompt", s.systemPrompt()}
		if first {
			args = append(args, "--session-id", s.agentSession)
		} else {
			args = append(args, "--resume", s.agentSession)
		}
	case "opencode":
		msg := text
		if first {
			msg = s.systemPrompt() + "\n\nUser: " + text
		}
		args = []string{"run", msg, "--format", "json"}
		if !first && s.agentSession != "" {
			args = append(args, "--session", s.agentSession)
		}
	default:
		s.setBusy(false)
		return errors.New("chat not supported for this agent")
	}

	bin, pre := resolveCmdShim(s.bin)
	cmd := exec.Command(bin, append(pre, args...)...)
	cmd.Dir = s.dir
	exeDir := ""
	if exe, err := os.Executable(); err == nil {
		exeDir = filepath.Dir(exe)
	}
	cmd.Env = append(os.Environ(), "PATH="+exeDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	hideCmdWindow(cmd)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		s.setBusy(false)
		return err
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr
	cmd.Stdin = nil
	if err := cmd.Start(); err != nil {
		s.setBusy(false)
		return err
	}
	s.mu.Lock()
	s.cmd = cmd
	s.mu.Unlock()

	go func() {
		sc := bufio.NewScanner(stdout)
		sc.Buffer(make([]byte, 1024*1024), 16*1024*1024)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" {
				continue
			}
			switch s.Agent {
			case "claude":
				s.handleClaudeEvent(line)
			case "opencode":
				s.handleOpenCodeEvent(line)
			}
		}
		werr := cmd.Wait()
		if werr != nil {
			msg := strings.TrimSpace(stderr.String())
			if msg == "" {
				msg = werr.Error()
			}
			if len(msg) > 2000 {
				msg = msg[len(msg)-2000:]
			}
			s.emit(aiMsg{Role: "error", Text: msg})
		}
		s.mu.Lock()
		s.cmd = nil
		s.mu.Unlock()
		s.setBusy(false)
	}()
	return nil
}

func (s *aiSession) stop() {
	s.mu.Lock()
	cmd := s.cmd
	s.mu.Unlock()
	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
}

func clip(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) > n {
		return s[:n] + "…"
	}
	return s
}

// toolSummary araç girdisinden tek satırlık özet (komut, dosya yolu…).
func toolSummary(input map[string]any) string {
	for _, k := range []string{"command", "file_path", "filePath", "path", "pattern", "url", "description"} {
		if v, ok := input[k].(string); ok && v != "" {
			return v
		}
	}
	b, _ := json.Marshal(input)
	return clip(string(b), 200)
}

func (s *aiSession) handleClaudeEvent(line string) {
	var ev struct {
		Type      string  `json:"type"`
		Subtype   string  `json:"subtype"`
		SessionID string  `json:"session_id"`
		IsError   bool    `json:"is_error"`
		Result    string  `json:"result"`
		Cost      float64 `json:"total_cost_usd"`
		Message   struct {
			Content []struct {
				Type    string          `json:"type"`
				Text    string          `json:"text"`
				Name    string          `json:"name"`
				Input   map[string]any  `json:"input"`
				Content json.RawMessage `json:"content"`
				IsError bool            `json:"is_error"`
			} `json:"content"`
		} `json:"message"`
		Denials []struct {
			ToolName  string         `json:"tool_name"`
			ToolInput map[string]any `json:"tool_input"`
		} `json:"permission_denials"`
	}
	if json.Unmarshal([]byte(line), &ev) != nil {
		return
	}
	switch ev.Type {
	case "assistant":
		for _, c := range ev.Message.Content {
			switch c.Type {
			case "text":
				if strings.TrimSpace(c.Text) != "" {
					s.emit(aiMsg{Role: "assistant", Text: c.Text})
				}
			case "tool_use":
				s.emit(aiMsg{Role: "tool", Tool: c.Name, Text: toolSummary(c.Input)})
			}
		}
	case "user":
		for _, c := range ev.Message.Content {
			if c.Type != "tool_result" {
				continue
			}
			var out string
			if json.Unmarshal(c.Content, &out) != nil {
				var parts []struct {
					Text string `json:"text"`
				}
				if json.Unmarshal(c.Content, &parts) == nil {
					for _, p := range parts {
						out += p.Text
					}
				}
			}
			role := "tool-result"
			if c.IsError {
				role = "tool-error"
			}
			s.emit(aiMsg{Role: role, Detail: clip(out, 4000)})
		}
	case "result":
		s.mu.Lock()
		s.Cost += ev.Cost
		s.mu.Unlock()
		for _, d := range ev.Denials {
			s.emit(aiMsg{Role: "error", Text: fmt.Sprintf(lang.L.AIDeniedFmt, d.ToolName, toolSummary(d.ToolInput))})
		}
		if ev.IsError && ev.Result != "" {
			s.emit(aiMsg{Role: "error", Text: ev.Result})
		}
		s.emit(aiMsg{Role: "info", Text: fmt.Sprintf("$%.4f", ev.Cost)})
	}
}

func (s *aiSession) handleOpenCodeEvent(line string) {
	var ev struct {
		Type      string `json:"type"`
		SessionID string `json:"sessionID"`
		Part      struct {
			Type  string  `json:"type"`
			Text  string  `json:"text"`
			Tool  string  `json:"tool"`
			Cost  float64 `json:"cost"`
			State struct {
				Status string         `json:"status"`
				Input  map[string]any `json:"input"`
				Output string         `json:"output"`
				Error  string         `json:"error"`
			} `json:"state"`
		} `json:"part"`
		Error any `json:"error"`
	}
	if json.Unmarshal([]byte(line), &ev) != nil {
		return
	}
	if ev.SessionID != "" {
		s.mu.Lock()
		s.agentSession = ev.SessionID
		s.mu.Unlock()
	}
	switch ev.Type {
	case "text":
		if strings.TrimSpace(ev.Part.Text) != "" {
			s.emit(aiMsg{Role: "assistant", Text: ev.Part.Text})
		}
	case "tool_use":
		s.emit(aiMsg{Role: "tool", Tool: ev.Part.Tool, Text: toolSummary(ev.Part.State.Input)})
		if ev.Part.State.Output != "" {
			s.emit(aiMsg{Role: "tool-result", Detail: clip(ev.Part.State.Output, 4000)})
		}
		if ev.Part.State.Error != "" {
			s.emit(aiMsg{Role: "tool-error", Detail: clip(ev.Part.State.Error, 4000)})
		}
	case "step_finish":
		if ev.Part.Cost > 0 {
			s.mu.Lock()
			s.Cost += ev.Part.Cost
			s.mu.Unlock()
		}
	case "error":
		b, _ := json.Marshal(ev.Error)
		s.emit(aiMsg{Role: "error", Text: string(b)})
	}
}

// ── Terminalde aç (headless desteklenmeyen araçlar ya da tercih) ─────────────

func openAgentTerminal(a aiAgent, dir, readme string) error {
	prompt := "Read " + filepath.ToSlash(readme) + " and the files it lists, then help me manage syncFTP for this project. Ask me what I want to do."
	var argLine string
	switch a.ID {
	case "opencode":
		argLine = `--prompt "` + prompt + `"`
	case "gemini", "qwen":
		argLine = `-i "` + prompt + `"`
	case "aider":
		argLine = `--read .syncftp/ai/README.md --read .syncftp/ai/CLI.md --read .syncftp/ai/CONFIG.md --read .syncftp/ai/PROJECT.md`
	default:
		argLine = `"` + prompt + `"`
	}
	if runtime.GOOS != "windows" {
		return errors.New("terminal launch is only implemented on Windows; run the agent in the project folder")
	}
	// .cmd başlatıcı: tırnak/özel karakter sorunlarını önler, pencere açık kalır
	launcher := filepath.Join(dir, ".syncftp", "ai", "start-"+a.ID+".cmd")
	exeDir := ""
	if exe, err := os.Executable(); err == nil {
		exeDir = filepath.Dir(exe)
	}
	content := "@echo off\r\ntitle syncFTP AI - " + a.Name + "\r\ncd /d \"" + dir + "\"\r\nset \"PATH=" + exeDir + ";%PATH%\"\r\n\"" + a.Path + "\" " + argLine + "\r\n"
	if err := os.WriteFile(launcher, []byte(content), 0600); err != nil {
		return err
	}
	c := exec.Command("cmd", "/c", "start", "", "cmd", "/k", launcher)
	c.Dir = dir
	return c.Start()
}

// ── HTTP uçları ───────────────────────────────────────────────────────────────

func (sv *guiServer) routesAI(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/ai/agents", func(w http.ResponseWriter, r *http.Request) { guiJSON(w, detectAIAgents()) })
	mux.HandleFunc("GET /api/ai/sessions", sv.apiAISessions)
	mux.HandleFunc("POST /api/p/{p}/ai/sessions", sv.apiAINew)
	mux.HandleFunc("POST /api/p/{p}/ai/terminal", sv.apiAITerminal)
	mux.HandleFunc("POST /api/ai/setup", sv.apiAISetup)
	mux.HandleFunc("GET /api/ai/sessions/{id}", sv.apiAIGet)
	mux.HandleFunc("POST /api/ai/sessions/{id}/send", sv.apiAISend)
	mux.HandleFunc("POST /api/ai/sessions/{id}/stop", sv.apiAIStop)
	mux.HandleFunc("DELETE /api/ai/sessions/{id}", sv.apiAIDelete)
	mux.HandleFunc("GET /api/ai/sessions/{id}/stream", sv.apiAIStream)
}

func (sv *guiServer) aiSession(id string) *aiSession {
	sv.ai.mu.Lock()
	defer sv.ai.mu.Unlock()
	return sv.ai.sessions[id]
}

func (sv *guiServer) apiAISessions(w http.ResponseWriter, r *http.Request) {
	sv.ai.mu.Lock()
	out := []map[string]any{}
	for _, s := range sv.ai.sessions {
		if p := r.URL.Query().Get("project"); p == "" || s.Project == p {
			snap := s.snapshot()
			delete(snap, "messages")
			snap["created"] = s.Created
			out = append(out, snap)
		}
	}
	sv.ai.mu.Unlock()
	sort.Slice(out, func(i, j int) bool { return out[i]["created"].(int64) > out[j]["created"].(int64) })
	guiJSON(w, out)
}

func (sv *guiServer) apiAINew(w http.ResponseWriter, r *http.Request) {
	p, ok := projects.Find(r.PathValue("p"))
	if !ok {
		guiErr(w, 404, fmt.Errorf(lang.L.ProjectNotFoundFmt, r.PathValue("p")))
		return
	}
	var req struct {
		Agent string `json:"agent"`
	}
	if err := guiDecode(r, &req); err != nil {
		guiErr(w, 400, err)
		return
	}
	a, ok := findAgent(req.Agent)
	if !ok || !a.Chat {
		guiErr(w, 400, errors.New(lang.L.AINoChat))
		return
	}
	readme, err := writeAIDocs(p)
	if err != nil {
		guiErr(w, 500, err)
		return
	}
	s := &aiSession{ID: uuid4(), Project: p.Name, Agent: a.ID, AgentName: a.Name, Messages: []aiMsg{},
		Created: time.Now().UnixMilli(), subs: map[chan aiMsg]struct{}{}, dir: p.Path, bin: a.Path, readme: readme}
	if a.ID == "claude" {
		s.agentSession = uuid4()
	}
	sv.ai.mu.Lock()
	sv.ai.sessions[s.ID] = s
	sv.ai.mu.Unlock()
	guiJSON(w, s.snapshot())
}

func (sv *guiServer) apiAITerminal(w http.ResponseWriter, r *http.Request) {
	p, ok := projects.Find(r.PathValue("p"))
	if !ok {
		guiErr(w, 404, errors.New("project"))
		return
	}
	var req struct {
		Agent string `json:"agent"`
	}
	if err := guiDecode(r, &req); err != nil {
		guiErr(w, 400, err)
		return
	}
	a, ok := findAgent(req.Agent)
	if !ok {
		guiErr(w, 404, errors.New("agent"))
		return
	}
	readme, err := writeAIDocs(p)
	if err != nil {
		guiErr(w, 500, err)
		return
	}
	if err := openAgentTerminal(a, p.Path, readme); err != nil {
		guiErr(w, 500, err)
		return
	}
	guiJSON(w, true)
}

func (sv *guiServer) apiAIGet(w http.ResponseWriter, r *http.Request) {
	s := sv.aiSession(r.PathValue("id"))
	if s == nil {
		guiErr(w, 404, errors.New("session"))
		return
	}
	guiJSON(w, s.snapshot())
}

func (sv *guiServer) apiAISend(w http.ResponseWriter, r *http.Request) {
	s := sv.aiSession(r.PathValue("id"))
	if s == nil {
		guiErr(w, 404, errors.New("session"))
		return
	}
	var req struct {
		Text string `json:"text"`
	}
	if err := guiDecode(r, &req); err != nil || strings.TrimSpace(req.Text) == "" {
		guiErr(w, 400, errors.New("text"))
		return
	}
	// Proje özeti her mesajda tazelenir (config panelden değişmiş olabilir)
	if s.Setup {
		_, _ = writeAISetupDocs(s.dir) // kayıtlı bağlantılar değişmiş olabilir (aynı README yolu)
	} else if p, ok := projects.Find(s.Project); ok {
		_, _ = writeAIDocs(p)
	}
	if err := s.send(strings.TrimSpace(req.Text)); err != nil {
		guiErr(w, 409, err)
		return
	}
	guiJSON(w, true)
}

func (sv *guiServer) apiAIStop(w http.ResponseWriter, r *http.Request) {
	if s := sv.aiSession(r.PathValue("id")); s != nil {
		s.stop()
	}
	guiJSON(w, true)
}

func (sv *guiServer) apiAIDelete(w http.ResponseWriter, r *http.Request) {
	if s := sv.aiSession(r.PathValue("id")); s != nil {
		s.stop()
		sv.ai.mu.Lock()
		delete(sv.ai.sessions, s.ID)
		sv.ai.mu.Unlock()
	}
	guiJSON(w, true)
}

func (sv *guiServer) apiAIStream(w http.ResponseWriter, r *http.Request) {
	s := sv.aiSession(r.PathValue("id"))
	if s == nil {
		guiErr(w, 404, errors.New("session"))
		return
	}
	fl, ok := w.(http.Flusher)
	if !ok {
		guiErr(w, 500, errors.New("sse"))
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	after := 0
	fmt.Sscanf(r.URL.Query().Get("after"), "%d", &after)

	ch := make(chan aiMsg, 256)
	s.mu.Lock()
	backlog := append([]aiMsg(nil), s.Messages...)
	s.subs[ch] = struct{}{}
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.subs, ch)
		s.mu.Unlock()
	}()
	send := func(m aiMsg) {
		b, _ := json.Marshal(m)
		fmt.Fprintf(w, "data: %s\n\n", b)
	}
	for _, m := range backlog {
		if m.ID > after {
			send(m)
		}
	}
	fl.Flush()
	t := time.NewTicker(20 * time.Second)
	defer t.Stop()
	for {
		select {
		case m := <-ch:
			send(m)
			fl.Flush()
		case <-t.C:
			_, _ = io.WriteString(w, ": ping\n\n")
			fl.Flush()
		case <-r.Context().Done():
			return
		case <-sv.quit:
			return
		}
	}
}

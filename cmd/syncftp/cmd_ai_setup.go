package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"syncftp/internal/config"
	"syncftp/internal/lang"
	"syncftp/internal/projects"
)

// ══════════════════════════════════════════════════════════════════════════════
// AI proje kurulum yardımcısı (YALNIZCA panel). Henüz proje olmayan bir klasörde oturum açar;
// asistan klasörü inceler (framework, build çıktısı), kayıtlı bağlantıları önerir ve projeyi
// normal CLI ile kurar: `syncftp init -y …`. Şifrelere hiç dokunmaz — kayıtlı bağlantı yoksa
// şifresiz oluşturur, kullanıcı şifreyi panelden girer.
// ══════════════════════════════════════════════════════════════════════════════

// systemPrompt oturum moduna göre sistem talimatı.
func (s *aiSession) systemPrompt() string {
	if s.Setup {
		return "You are the syncFTP project setup assistant inside the syncFTP desktop panel. Your job: turn the folder " + s.dir +
			" into a syncFTP project. First read " + s.readme + " and the files it lists (CLI.md, CONFIG.md, LIBRARY.md in the same folder) " +
			"and follow its rules strictly: never ask for, write or repeat passwords; create the project only with `syncftp init -y`; ask before anything outward-facing."
	}
	return aiSystemPrompt(s.readme)
}

// writeAISetupDocs hedef klasörün .syncftp/ai/ dizinine kurulum dokümanlarını yazar; README yolunu döner.
func writeAISetupDocs(dir string) (string, error) {
	aiDir := filepath.Join(dir, ".syncftp", "ai")
	if err := os.MkdirAll(aiDir, 0700); err != nil {
		return "", err
	}
	exe, _ := os.Executable()
	files := map[string]string{
		"README.md":  aiSetupReadme(dir, exe),
		"CLI.md":     aiCLIDoc(),
		"CONFIG.md":  aiConfigDoc,
		"LIBRARY.md": aiLibraryDoc(),
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(aiDir, name), []byte(content), 0600); err != nil {
			return "", err
		}
	}
	return filepath.Join(aiDir, "README.md"), nil
}

func aiSetupReadme(dir, exe string) string {
	userLang := "Turkish"
	if lang.Current() == "en" {
		userLang = "English"
	}
	existing := "No syncFTP config exists here yet."
	if projects.HasConfig(dir) {
		existing = "A syncFTP config ALREADY exists here — do not run init again; help the user review it (read .syncftp/syncftp.json) and make sure it is registered with `syncftp project add .`."
	}
	return fmt.Sprintf(`# syncFTP project setup guide

You help the user turn a local folder into a **syncFTP** project (syncFTP deploys changed files from a
local folder to FTP / FTPS / SFTP servers). The user talks to you from the syncFTP panel.

- Target folder (your working directory): %s
- syncFTP executable: %s  (also on PATH as `+"`syncftp`"+`)
- %s
- Reply in **%s**. Be concise, friendly and practical. Ask one or two questions at a time.

## Files to read
- LIBRARY.md — saved connections (names only, no secrets) and the defaults new projects get
- CLI.md — every command and flag (see `+"`syncftp init`"+`)
- CONFIG.md — the .syncftp/syncftp.json schema

## Steps
1. Inspect the folder (Glob/Read, do not modify): detect the stack and suggest an ignore template —
   composer.json + artisan → laravel; wp-config.php / wp-content → wordpress; package.json → node; otherwise generic.
   If the deployable output is a sub-folder (dist/, build/, public/ for a static build) suggest `+"`--local <folder>`"+`.
2. Ask where it should be deployed: which saved connection (LIBRARY.md) or which host/protocol/user,
   and the remote folder (e.g. /public_html, /var/www/site). Suggest a project name (folder name by default).
3. Show the exact command you will run and wait for the user's OK, then run ONE of:
   `+"```"+`
   syncftp init -y --name NAME --connection SAVED --remote /public_html --template laravel
   syncftp init -y --name NAME --protocol sftp --host HOST --user USER --key ~/.ssh/id_ed25519 --remote /var/www --template node
   syncftp init -y --name NAME --protocol ftp --host HOST --user USER --remote /public_html --template generic
   syncftp init -y --name NAME --no-server
   `+"```"+`
   init writes .syncftp/syncftp.json with the user's global defaults, adds .syncftp to .gitignore and
   registers the project in the panel — no extra step needed.
4. After init: if a saved connection was used, run `+"`syncftp status`"+` to check it connects. If the server
   was created WITHOUT a password, tell the user to open the project in the panel → Servers → the server →
   enter the password there (or set up an SSH key), then try "Dry-run".
5. Offer useful extras by editing .syncftp/syncftp.json (re-read it first, keep valid JSON): exclude
   build caches, protect .env, remote_post_sync commands (SFTP only) such as `+"`php artisan migrate --force`"+`.

## Rules
1. NEVER ask the user to type a password, token or private key into this chat, and never pass `+"`--pass`"+`.
   Passwords are entered in the panel's server form (or come from a saved connection).
2. Do not run a real sync/upload, rollback or any remote-changing command. Dry-runs are fine after the user agrees.
3. Do not modify files outside the target folder, and do not change the user's source files — only .syncftp/.
4. Do not run `+"`syncftp init`"+` without `+"`-y`"+` (interactive prompts would hang).
`, dir, exe, existing, userLang)
}

// aiLibraryDoc kayıtlı bağlantı adları (şifresiz) + yeni proje varsayılanları.
func aiLibraryDoc() string {
	lib, err := config.LoadLibrary()
	var b strings.Builder
	fmt.Fprintf(&b, "# Saved connections and new-project defaults\n\nGenerated %s. No secrets.\n\n", time.Now().Format("2006-01-02 15:04"))
	if err != nil {
		fmt.Fprintf(&b, "_(library could not be read: %v)_\n\n", err)
	}
	b.WriteString("## Saved connections — use with `syncftp init -y --connection NAME`\n\n")
	if len(lib.Connections) == 0 {
		b.WriteString("None. Ask the user for host / protocol / user, or suggest adding one in the panel (Settings → Connection library).\n\n")
	} else {
		b.WriteString("| name | protocol | user@host:port | auth |\n|---|---|---|---|\n")
		for _, c := range lib.Connections {
			auth := "password stored"
			if c.Password == "" {
				auth = "no password"
			}
			if c.KeyFile != "" {
				auth = "ssh key"
			}
			fmt.Fprintf(&b, "| %s | %s | %s@%s:%d | %s |\n", c.Name, orDefault(c.Protocol, "ftp"), c.User, c.Host, c.Port, auth)
		}
		b.WriteString("\n")
	}
	d := lib.Defaults
	b.WriteString("## Defaults applied to every new project (by init)\n\n")
	fmt.Fprintf(&b, "- ignore template: %s (override with --template)\n", orDefault(d.IgnoreTemplate, "none"))
	fmt.Fprintf(&b, "- include: %v\n- exclude: %v\n- protect: %v\n", d.Sync.Include, d.Sync.Exclude, d.Sync.Protect)
	fmt.Fprintf(&b, "- backups: %v (keep %d), bandwidth limit: %d KB/s\n", !d.Sync.NoBackup, d.Sync.EffectiveBackupKeep(), d.Sync.BandwidthLimit)
	fmt.Fprintf(&b, "- new server: max_connections %d, max_retries %d\n", d.MaxConnections, d.MaxRetries)
	return b.String()
}

// apiAISetup {agent, path} → kurulum oturumu.
func (sv *guiServer) apiAISetup(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Agent string `json:"agent"`
		Path  string `json:"path"`
	}
	if err := guiDecode(r, &req); err != nil {
		guiErr(w, 400, err)
		return
	}
	dir := filepath.Clean(strings.TrimSpace(req.Path))
	if st, err := os.Stat(dir); req.Path == "" || err != nil || !st.IsDir() {
		guiErr(w, 400, fmt.Errorf(lang.L.GuiDirNotFoundFmt, dir))
		return
	}
	a, ok := findAgent(req.Agent)
	if !ok || !a.Chat {
		guiErr(w, 400, errors.New(lang.L.AINoChat))
		return
	}
	readme, err := writeAISetupDocs(dir)
	if err != nil {
		guiErr(w, 500, err)
		return
	}
	s := &aiSession{ID: uuid4(), Project: filepath.Base(dir), Agent: a.ID, AgentName: a.Name, Messages: []aiMsg{},
		Created: time.Now().UnixMilli(), subs: map[chan aiMsg]struct{}{}, dir: dir, bin: a.Path, readme: readme,
		Setup: true, Folder: dir}
	if a.ID == "claude" {
		s.agentSession = uuid4()
	}
	sv.ai.mu.Lock()
	sv.ai.sessions[s.ID] = s
	sv.ai.mu.Unlock()
	guiJSON(w, s.snapshot())
}

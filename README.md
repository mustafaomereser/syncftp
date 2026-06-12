# syncFTP

A Go CLI tool that detects changed files via SHA256 hashing and distributes them to one or more FTP servers.

- **No git required** — change detection is hash-based, works in any directory
- **Multi-server** — deploy to production and staging in a single command
- **Server-side protection** — never overwrites `.env`, database configs, or any file you mark as protected
- **Freeze list** — per-server list of files that are permanently skipped even when changed locally
- **Parallel uploads** — configurable connection pool per server
- **Auto-retry** — failed uploads are retried automatically and saved for manual re-runs
- **Config-free mode** — sync any directory to any FTP without a project config (`syncftp push`)
- **HTTP API** — built-in local API server for PHP/web UI integration (`syncftp serve`)
- **Interactive shell** — run `syncftp` with no arguments for a full TUI shell with arrow-key file browser, server picker, and action menus
- **CRLF normalization** — line endings are normalized before upload so PHP hosting servers don't inject blank lines

The UI defaults to **English**. Switch languages with `syncftp lang tr` (or `lang tr` inside the shell). The preference is saved to `.syncftp/lang` and persists across restarts. Use `SYNCFTP_LANG=tr` to override for a single session without saving.

---

## Installation

Requires **Go 1.21+** → https://go.dev/dl/

```bash
git clone <repo>
cd syncftp
go mod tidy
go build -o syncftp.exe ./cmd/syncftp/   # Windows
go build -o syncftp     ./cmd/syncftp/   # Linux / macOS
```

---

## Quick Start

```bash
cd /path/to/your/project

syncftp init      # interactive wizard — creates syncftp.json
syncftp status    # show what has changed (nothing is uploaded)
syncftp sync      # upload changed files to all enabled servers
```

---

## Configuration — `syncftp.json`

Created by `syncftp init`. Stored with permission `600`. Added to `.gitignore` automatically.

```json
{
  "project": {
    "name": "my-project",
    "local_path": "."
  },
  "sync": {
    "protect": [".env", "config/database.php", "storage/"],
    "include": [],
    "exclude": ["vendor/", "node_modules/"],
    "ignore_files": []
  },
  "first_sync": {
    "full": false
  },
  "servers": [
    {
      "name": "production",
      "host": "ftp.example.com",
      "port": 21,
      "user": "ftpuser",
      "password": "secret",
      "remote_path": "/public_html",
      "passive": true,
      "enabled": true,
      "max_connections": 3,
      "max_retries": 2,
      "include": [],
      "exclude": []
    },
    {
      "name": "staging",
      "host": "ftp2.example.com",
      "port": 21,
      "user": "staginguser",
      "password": "secret2",
      "remote_path": "/staging",
      "passive": true,
      "enabled": true,
      "max_connections": 1,
      "max_retries": 2,
      "include": ["css/", "js/"],
      "exclude": []
    }
  ]
}
```

### Field Reference

| Field | Default | Description |
|---|---|---|
| `project.local_path` | `"."` | Local directory to scan, relative to config file |
| `sync.protect` | `[]` | Files/dirs that are **never** overwritten on the FTP server |
| `sync.include` | `[]` | Global whitelist — sync only these paths (empty = all) |
| `sync.exclude` | `[]` | Global blacklist — always skip these paths |
| `sync.ignore_files` | `[]` | Which ignore files to load — empty means both `.gitignore` and `syncftp.ignore` |
| `first_sync.full` | `false` | `true` = force upload everything on first sync; `false` = smart comparison (recommended) |
| `server.disable_epsv` | `false` | Disable EPSV, use PASV only — fixes some NAT/firewall setups |
| `server.nat_workaround` | `false` | Ignore the IP in PASV response, use server host instead |
| `server.max_connections` | `1` | Parallel FTP connections for this server |
| `server.max_retries` | `2` | Retry count on upload failure (0 = no retry) |
| `server.include` | `[]` | Per-server whitelist — overrides global `sync.include` |
| `server.exclude` | `[]` | Per-server blacklist — added on top of global `sync.exclude` |
| `server.enabled` | `true` | Set `false` to skip this server in all sync operations |

---

## Ignore Files

syncFTP loads **both** `.gitignore` and `syncftp.ignore` if they exist and merges their patterns. You can control which files are loaded via `sync.ignore_files` in `syncftp.json`.

| `sync.ignore_files` value | Behavior |
|---|---|
| `[]` or field absent | Load `.gitignore` + `syncftp.ignore` (default) |
| `[".gitignore"]` | Only `.gitignore` |
| `["syncftp.ignore"]` | Only `syncftp.ignore` |
| `[".gitignore", "syncftp.ignore"]` | Both (explicit) |

**Typical setup** — keep `.gitignore` for git, add FTP-specific ignores to `syncftp.ignore`:

```gitignore
# syncftp.ignore
vendor/
node_modules/
*.log
.DS_Store
uploads/
```

The `.syncftp/` metadata directory and `syncftp.json` itself are always excluded from scanning regardless of ignore rules.

---

## Commands

### `syncftp init`

Interactive wizard. Prompts for project name, FTP credentials, and writes `syncftp.json`. Also adds the config file and binary to `.gitignore`.

```
=== syncFTP Setup Wizard ===

Project name [my-project]:
Local directory [.]:

Server name [production]:
FTP host: ftp.example.com
Port [21]:
Username: ftpuser
Password: ****
Remote directory [/public_html]:
```

---

### `syncftp status`

Shows what has changed since the last sync. **Nothing is uploaded.**

```bash
syncftp status
syncftp status --include css
syncftp status --exclude vendor
```

```
Project : my-project
Path    : /home/user/my-project
Files   : 142

── production (ftp.example.com) ──
  + NEW (2):
      js/utils.js
      css/dark-mode.css
  ~ CHANGED (1):
      index.php
  - DELETED locally (not removed from FTP) (1):
      old-file.php
```

> Deleted files are reported but **never removed** from the FTP server — intentional safety behaviour.

---

### `syncftp sync`

Uploads changed files to all enabled servers.

```bash
syncftp sync                          # server picker opens if multiple servers configured
syncftp sync --all                    # skip picker, sync to all enabled servers
syncftp sync --server production      # one specific server
syncftp sync --full                   # ignore state, re-upload everything
syncftp sync --dry-run                # preview what would be uploaded
syncftp sync --retry-failed           # re-upload files that failed in the last run
syncftp sync --include css --include js/app.js
syncftp sync --exclude vendor --exclude tests
```

**Smart first sync** — on the very first run, syncFTP lists all files already on the FTP server and compares sizes with local files. Only missing or size-different files are uploaded. Safe to add to an already-deployed project. Use `--full` to force a complete re-upload.

**Example output — first sync:**

```
Scanning: /home/user/my-project
142 files found

══ production (ftp.example.com) ══
  First sync: scanning server, comparing existing files...
  139 files found on server
  Result: 139 up-to-date (skipped)  |  3 different/missing (uploading)
  ❄ 2 frozen (skipped)
  3 files to process
  Connection pool: 3 / Retry: 2
    ✓ js/utils.js
    ✓ css/dark-mode.css
    ✓ index.php
  Done: 3 uploaded, 0 protected, 0 failed
  Release: .syncftp/releases/production/20260612-143012
```

---

### `syncftp freeze`

Manages the **freeze list** for a server — files marked as frozen are permanently skipped during sync even if changed locally. Useful for files that differ intentionally between local and server (config overrides, server-specific assets, etc.).

```bash
syncftp freeze                        # server picker if multiple servers
syncftp freeze --server production    # specific server
```

Opens a full-screen TUI where you can browse all local files, toggle freeze with `Space`, filter by typing, and save with `Enter`.

```
🧊 Freeze list — production
  Type to filter  |  Space = freeze/unfreeze  |  a = toggle all  |  Enter = save  |  q = cancel
  ──────────────────────────────────────────────────────────────────
▶ [❄]  config/database.php
  [❄]  config/smtp.php
  [ ]  index.php
  [ ]  js/app.js
  ──────────────────────────────────────────────────────────────────
  ❄ 2 frozen  ·  4/142 shown
```

Freeze lists are stored per-server in `.syncftp/frozen/<server>.json`. To clear all freezes: open the TUI, press `a` to unmark all, then `Enter`.

You can also toggle freeze directly from the **file browser** (`ls` command or `syncftp remote ls`) by pressing `f` on any file or folder. Pressing `f` on a folder freezes/unfreezes all files inside it recursively. Frozen files show `❄` icon, folders with frozen contents show `❄📁`.

---

### `syncftp push`

Syncs **any local directory** to an FTP server without needing a `syncftp.json`. Useful for one-off deployments or scripting.

```bash
syncftp push ./website --host ftp.example.com --user admin --pass secret --remote /public_html
syncftp push ./website --host ftp.example.com --user admin --pass secret --remote /public_html --connections 3
syncftp push ./another-project --server production
syncftp push ./website --host ftp.example.com --user admin --pass secret --remote /public_html --dry-run
syncftp push ./website --host ftp.example.com --user admin --pass secret --remote /public_html --full
```

State is stored in `<local-dir>/.syncftp/` so subsequent runs only upload changed files.

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--host` | — | FTP server address (required unless `--server` is used) |
| `--port` | `21` | FTP port |
| `--user` | — | Username |
| `--pass` | — | Password |
| `--remote` | `"/"` | Remote target directory |
| `--passive` | `true` | Use passive mode |
| `--server` | — | Reuse a server from `syncftp.json` |
| `--full` | `false` | Re-upload all files, ignore state |
| `--dry-run` | `false` | Preview without uploading |
| `--include` | — | Whitelist (repeatable) |
| `--exclude` | — | Blacklist (repeatable) |
| `--connections` | `1` | Parallel FTP connections |
| `--retries` | `2` | Retry count on failure |

---

### `syncftp remote`

Browse, download, preview and delete files on the FTP server.

```bash
syncftp remote ls                          # open interactive file browser
syncftp remote ls css/                     # open browser starting at css/
syncftp remote ls --recursive              # plain recursive tree (no TUI)
syncftp remote ls --server staging

syncftp remote cat index.php               # preview first 10 KB
syncftp remote cat                         # open file browser to pick a file
syncftp remote cat error.log --max-kb 50

syncftp remote get index.php               # download to current directory
syncftp remote get                         # open file browser to pick a file
syncftp remote get css/style.css ~/tmp/

syncftp remote rm old-file.php             # delete file (TUI confirmation)
syncftp remote rm                          # open file browser to pick a file
syncftp remote rm old-file.php --force
syncftp remote rm cache/ --recursive
```

---

### `syncftp` (Interactive Shell)

Run `syncftp` with no arguments to open the interactive shell.

```bash
syncftp
```

#### Shell Commands

| Command | Description |
|---|---|
| `ls [path]` | Open arrow-key file browser |
| `cd <path>` | Change remote directory (`cd ..` to go up) |
| `cat [file]` | Preview file content |
| `get [file] [dest]` | Download file |
| `rm [-f] [-r] [file]` | Delete file/dir |
| `pwd` | Show current remote path |
| `status` | Show local changes per server |
| `sync [--all] [--full] [--dry-run] [--server name]` | Upload to FTP |
| `freeze [--server name]` | Manage freeze list for a server |
| `servers` | TUI server list |
| `server [name]` | Connect to a server |
| `lang [en\|tr]` | Show or change display language |
| `clear` / `cls` | Clear screen |
| `help` / `?` | Show command reference |
| `exit` / `quit` | Exit shell |

#### File Browser

Opened by `ls`, or automatically when `cat`/`get`/`rm` receive no path argument.

| Key | Action |
|---|---|
| `↑` / `↓` or `j` / `k` | Navigate |
| `Enter` | Enter directory / select file |
| `→` / `l` | Enter directory; on a file: load full-screen preview |
| `←` / `h` / `ESC` | Go up one directory |
| `g` / `G` | Jump to first / last item |
| `/` | Search — type to filter; `Enter` = recursive server-wide search |
| `Space` | Toggle mark on file or folder |
| `a` | Mark all / unmark all (disabled during search) |
| `f` | **Freeze/unfreeze** file; on a folder: toggle all files inside recursively |
| `d` | Delete all marked items (double confirmation) |
| `m` | Move all marked items — pick destination in second browser |
| `r` | Reconnect to server (after connection drop) |
| `q` | Close browser |

- Frozen files show `❄` icon; folders with frozen contents show `❄📁`
- Folder item counts loaded in background: `...` → `N items` or `1000+` or `?` on error
- Preview panel on terminals ≥ 120 chars wide — loaded on demand (`→`), not on every cursor move
- Recursive search with `/` + `Enter`: scans entire server, results show relative path
- Selecting a single file opens an action menu: **View / Download / Delete / Cancel**
- Command history saved to `.syncftp/shell_history`

#### Sync Progress (TUI)

When `sync` runs, a full-screen progress view is shown:

```
  ══ production ══

  ████████████████░░░░░░░░░░░░░░░░  52%  13/25

  ⠙ css/components/button.css

  ✓ index.php
  ✓ js/app.js
  ✗ img/hero.jpg: connection reset (3 attempts)
  ✓ css/components/button.css
```

---

### `syncftp lang`

Show or change the display language.

```bash
syncftp lang        # show current language
syncftp lang en     # switch to English
syncftp lang tr     # switch to Turkish
```

Preference saved to `.syncftp/lang`. Use `SYNCFTP_LANG=tr` to override for a single session without saving. Inside the shell: `lang en` / `lang tr`.

---

### `syncftp serve`

Starts a local HTTP API server.

```bash
syncftp serve              # http://127.0.0.1:8080
syncftp serve --port 9000
```

Only listens on `127.0.0.1`. CORS enabled for all origins.

#### API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/api/servers` | List configured servers (passwords excluded) |
| `GET` | `/api/status` | Changed files per server |
| `GET` | `/api/status?server=production` | Changed files for one server |
| `POST` | `/api/sync` | Run sync |
| `GET` | `/api/failed` | Files that failed in last sync |
| `GET` | `/api/remote/ls?server=&path=` | List remote directory |
| `GET` | `/api/remote/ls?server=&path=&recursive=true` | Recursive listing |
| `GET` | `/api/remote/cat?server=&path=&max_kb=10` | Preview file |
| `GET` | `/api/remote/get?server=&path=` | Download file (raw) |
| `GET` | `/api/remote/get?server=&path=&json=true` | Download file (base64 JSON) |
| `DELETE` | `/api/remote/rm?server=&path=` | Delete file |
| `DELETE` | `/api/remote/rm?server=&path=&recursive=true` | Delete directory |

**POST /api/sync body:**

```json
{
  "server":       "production",
  "full":         false,
  "dry_run":      false,
  "include":      ["css/"],
  "exclude":      ["vendor/"],
  "retry_failed": false
}
```

---

## Filtering System

Three independent mechanisms that work together:

### 1. `protect` — permanent server-side protection

Files in `sync.protect` are **never uploaded**. Use for server-specific files that must not be overwritten.

```json
"sync": { "protect": [".env", "config/database.php", "storage/"] }
```

### 2. Freeze list — per-server permanent skip

Files in a server's freeze list are skipped during sync even when changed locally. Managed via `syncftp freeze` or the `f` key in the file browser. Stored in `.syncftp/frozen/<server>.json`.

Unlike `protect` (which is global and config-based), freeze lists are:
- **Per-server** — a file can be frozen for production but not staging
- **Managed interactively** — no need to edit JSON manually
- **FTP-path aware** — toggled from the remote browser

### 3. `include` / `exclude` — config and CLI filters

```json
"sync": { "exclude": ["vendor/", "node_modules/", "tests/"] },
"servers": [
  { "name": "production", "include": ["css/", "js/", "index.php"] }
]
```

**Priority (include):** CLI `--include` → server `include` → global `sync.include`

**Priority (exclude):** global `sync.exclude` + server `exclude` + CLI `--exclude` (all combined)

### Summary

| Mechanism | Scope | Managed via |
|---|---|---|
| `sync.protect` | Global, all servers | `syncftp.json` |
| Freeze list | Per-server, file-level | `syncftp freeze` / browser `f` key |
| `sync.include` / `sync.exclude` | Global | `syncftp.json` |
| `server.include` / `server.exclude` | Per-server | `syncftp.json` |
| `--include` / `--exclude` flags | This run only | CLI |
| `sync.ignore_files` | File scanner | `syncftp.json` |

---

## State & Release Files

syncFTP creates a `.syncftp/` directory next to `syncftp.json`:

```
.syncftp/
├── state/
│   ├── production.json    # per-file hashes from last successful sync
│   └── staging.json
├── failed/
│   └── production.json    # files that failed last run (cleared on success)
├── frozen/
│   ├── production.json    # freeze list for production server
│   └── staging.json
├── releases/
│   └── production/
│       └── 20260612-143012/
│           └── manifest.json
└── lang                   # saved language preference ("en" or "tr")
```

`.syncftp/` is always excluded from scanning and never uploaded.

---

## Internal Structure

| Package | Role |
|---|---|
| `internal/config` | JSON config loading |
| `internal/ignore` | `.gitignore` + `syncftp.ignore` merged parser |
| `internal/scanner` | Directory walk, SHA256 hashing |
| `internal/state` | Per-server sync state (load / save / diff) |
| `internal/ftp` | FTP client, connection pool, CRLF normalizer |
| `internal/failed` | Failed file list persistence |
| `internal/release` | Release manifest writer |
| `internal/frozen` | Per-server freeze list (load / save) |
| `internal/lang` | i18n — English default, Turkish via `SYNCFTP_LANG=tr` |
| `cmd/syncftp` | CLI commands (init, status, sync, push, remote, serve, freeze, lang, shell) |

---

## Running Tests

```bash
go test ./...
go test ./internal/state/... -v
go test ./internal/scanner/...
```

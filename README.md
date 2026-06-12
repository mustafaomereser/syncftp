# syncFTP

A Go CLI tool that detects changed files via SHA256 hashing and distributes them to one or more FTP servers.

- **No git required** — change detection is hash-based, works in any directory
- **Multi-server** — deploy to production and staging in a single command
- **Server-side protection** — never overwrites `.env`, database configs, or any file you mark as protected
- **Parallel uploads** — configurable connection pool per server
- **Auto-retry** — failed uploads are retried automatically and saved for manual re-runs
- **Config-free mode** — sync any directory to any FTP without a project config (`syncftp push`)
- **HTTP API** — built-in local API server for PHP/web UI integration (`syncftp serve`)
- **Interactive shell** — run `syncftp` with no arguments for a full TUI shell with arrow-key file browser, server picker, and action menus

The UI defaults to **English**. Set `SYNCFTP_LANG=tr` to switch to Turkish.

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
    "exclude": []
  },
  "first_sync": {
    "full": true
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
      "exclude": ["vendor/"]
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
| `first_sync.full` | `false` | `true` = force upload everything on first sync; `false` = smart comparison (recommended) |
| `server.disable_epsv` | `false` | Disable EPSV, use PASV only — fixes some NAT/firewall setups |
| `server.nat_workaround` | `false` | Ignore the IP in PASV response, use server host instead — fixes NAT traversal issues |
| `server.max_connections` | `1` | Parallel FTP connections for this server |
| `server.max_retries` | `2` | Retry count on upload failure (0 = no retry) |
| `server.include` | `[]` | Per-server whitelist — overrides global `sync.include` |
| `server.exclude` | `[]` | Per-server blacklist — added on top of global `sync.exclude` |
| `server.enabled` | `true` | Set `false` to skip this server in all sync operations |

---

## Ignore Files

syncFTP uses the first file it finds, in this order:

1. `.gitignore` — used automatically if present
2. `syncftp.ignore` — create this if you don't use git

Format is identical to `.gitignore`:

```gitignore
node_modules/
*.log
.DS_Store
dist/
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
syncftp status                                        # all servers
syncftp status --include css                          # only show changes under css/
syncftp status --exclude vendor                       # hide vendor/ changes
syncftp status --include src --exclude src/__tests__  # combine
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
syncftp sync --dry-run                # preview what would be uploaded (TUI)
syncftp sync --retry-failed           # re-upload files that failed in the last run
```

**Smart first sync** — on the very first run, syncFTP does not blindly upload everything. Instead it lists all files already on the FTP server and compares sizes with local files. Only files that are missing or have a different size are uploaded. This is safe to use when syncFTP is added to an already-deployed project. Use `--full` to force a complete re-upload regardless.

**Filtering:**

```bash
# Whitelist — only sync specific paths (overrides syncftp.json include)
syncftp sync --include css
syncftp sync --include css --include js/app.js

# Blacklist — exclude specific paths for this run only
syncftp sync --exclude vendor
syncftp sync --exclude vendor --exclude tests

# Combine
syncftp sync --include src/components --exclude src/components/__tests__

# Always preview first
syncftp sync --include css --dry-run
syncftp sync --include css
```

**Example output — first sync (smart comparison):**

```
Scanning: /home/user/my-project
142 files found

══ production (ftp.example.com) ══
  First sync: scanning server, comparing existing files...
  139 files found on server
  Result: 139 up-to-date (skipped)  |  3 different/missing (uploading)
  3 files to process
  Connection pool: 3 / Retry: 2
    ✓ js/utils.js
    ✓ css/dark-mode.css
    ✓ index.php
  Done: 3 uploaded, 0 protected, 0 failed
  Release: .syncftp/releases/production/20260612-143012
```

**Example output — incremental sync:**

```
Scanning: /home/user/my-project
142 files found

══ production (ftp.example.com) ══
  3 files to process (1 protected)
  Connection pool: 3 / Retry: 2
    PROTECTED  .env
    ✓ js/utils.js
    ✓ css/dark-mode.css
    ✓ index.php (2nd attempt)
  Done: 3 uploaded, 1 protected, 0 failed
  Release: .syncftp/releases/production/20260612-143012
```

---

### `syncftp push`

Syncs **any local directory** to an FTP server without needing a `syncftp.json`. Useful for one-off deployments or scripting.

State is stored in `<local-dir>/.syncftp/` so subsequent runs only upload changed files.

```bash
# Ad-hoc — provide credentials directly
syncftp push ./website --host ftp.example.com --user admin --pass secret --remote /public_html

# Use a server already defined in syncftp.json (only changes the local source dir)
syncftp push ./another-project --server production

# With parallel connections
syncftp push ./website --host ftp.example.com --user admin --pass secret \
  --remote /public_html --connections 3

# Filtering
syncftp push ./website --host ftp.example.com --user admin --pass secret \
  --remote /public_html --include css --exclude vendor

# Preview first
syncftp push ./website --host ftp.example.com --user admin --pass secret \
  --remote /public_html --dry-run

# Force full re-upload
syncftp push ./website --host ftp.example.com --user admin --pass secret \
  --remote /public_html --full
```

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

When multiple servers are configured and `--server` is not provided, an interactive TUI server picker is shown.

```bash
syncftp remote ls                          # open interactive file browser
syncftp remote ls css/                     # open browser starting at css/
syncftp remote ls --recursive              # plain recursive tree (no TUI)
syncftp remote ls --server staging         # specific server

syncftp remote cat index.php               # preview first 10 KB
syncftp remote cat                         # open file browser to pick a file
syncftp remote cat error.log --max-kb 50   # preview first 50 KB

syncftp remote get index.php               # download to current directory
syncftp remote get                         # open file browser to pick a file
syncftp remote get css/style.css ~/tmp/    # download to specific path

syncftp remote rm old-file.php             # delete file (TUI confirmation)
syncftp remote rm                          # open file browser to pick a file
syncftp remote rm old-file.php --force     # delete without confirmation
syncftp remote rm cache/ --recursive       # delete directory and all contents
```

Paths without a leading `/` are resolved relative to the server's `remote_path`. Use an absolute path (starting with `/`) to reference any location on the server.

---

### `syncftp` (Interactive Shell)

Run `syncftp` with no arguments to open the interactive shell. A full TUI environment with arrow-key navigation, file browser, server picker, and action menus.

```bash
syncftp
```

```
╔══════════════════════════════════════════════╗
║  syncFTP Shell  ·  my-project                ║
║  'help' yazın, çıkmak için 'exit'            ║
╚══════════════════════════════════════════════╝

syncftp [production:/public_html]>
```

#### Shell Commands

| Command | Description |
|---|---|
| `ls [path]` | Open arrow-key file browser (navigate with ↑↓, enter dirs with →, go up with ←) |
| `cd <path>` | Change remote directory (`cd ..` to go up) |
| `cat [file]` | Preview file content — opens browser if no path given |
| `get [file] [dest]` | Download file — opens browser if no path given |
| `rm [-f] [-r] [file]` | Delete file/dir — opens browser if no path given, TUI confirmation |
| `pwd` | Show current remote path |
| `status` | Show local changes per server |
| `sync [--all] [--full] [--dry-run] [--server name]` | Upload to FTP — TUI progress view, multi-select picker if multiple servers |
| `servers` | TUI server list — select to connect |
| `server [name]` | Connect to a server (TUI picker if no name given) |
| `clear` / `cls` | Clear screen and show welcome banner |
| `help` / `?` | Show command reference |
| `exit` / `quit` | Exit shell |

#### File Browser

Opened by `ls`, or automatically when `cat`/`get`/`rm` receive no path argument.

```
  📁 /public_html/assets
  ↑↓ Navigate  |  Enter Open/Select  |  → Preview  |  ← Up  |  Space Mark  |  / Search  |  q Quit
  ────────────────────────────────────────────────────────────────────────────────────────────

 ▶ 📁  css/                              3 items        2026-06-11
    📁  js/                               12 items       2026-06-10
    📁  uploads/                          1000+ items    2026-06-09
    ·············································
    📄  style.css                         4.2 KB         2026-06-10
    📄  app.js                            18.7 KB        2026-06-09

  3 folders, 2 files  ·  1/5
```

| Key | Action |
|---|---|
| `↑` / `↓` or `j` / `k` | Navigate up/down |
| `Enter` | Enter directory / select file |
| `→` | Enter directory; on a **file**: load preview panel (on-demand, not on every cursor move) |
| `←` / `ESC` | Go up one directory |
| `g` / `G` | Jump to first / last item |
| `/` | Search — type to filter current directory; press **Enter** to search entire server recursively |
| `Space` | Toggle mark on file or folder |
| `a` | Mark all / unmark all visible items (disabled during search) |
| `d` | **Delete** all marked items (double confirmation) |
| `m` | **Move** all marked items — opens a second browser to pick destination folder |
| `q` | Close browser / cancel |

- **Folders first** (alphabetical), then files (alphabetical), dotted separator between them
- **Folder item counts** loaded in background: `...` → `N items` or `1000+ items` (1000+) or `?` on error
- **Preview panel** on terminals ≥ 120 chars wide — press `→` on a file to load; on-demand only, no FTP request on cursor moves
- **Recursive search** with `/` + `Enter`: scans all subdirectories on the server, results show the relative path so you know where each file lives
- **Mark & act**: `Space` marks files and folders, hint bar changes to `d=Sil  |  m=Taşı  |  a=Tümünü kaldır`
- **Move**: opens a second browser in "pick folder" mode — `Enter` selects the highlighted folder as destination (does not enter it); `→` still enters for navigation
- Selecting a single file in `ls` opens an action menu: **Cat / Get / Delete / Cancel**
- Command history saved to `.syncftp/shell_history` (arrow keys work)

#### Sync Progress (TUI)

When `sync` runs, a full-screen progress view is shown:

```
  ══ production ══

  ████████████████░░░░░░░░░░░░░░░░  52%  13/25
  
  ⠙ css/components/button.css

  ✓ index.php
  ✓ js/app.js
  ✓ css/style.css
  ✗ img/hero.jpg: connection reset (3 deneme)
  ✓ css/components/button.css
```

- Real-time per-file streaming as uploads complete
- Progress bar with percentage and file count
- Animated spinner showing current file
- Scrolling log of last 12 results (✓ success, ✗ failure with attempt count)
- Full summary on completion — press any key to return to shell

#### Server Picker (TUI)

Used by `servers`, `server`, `sync` (multi-select), and `remote` commands.

```
  Servers
  ─────────────────────────────────────────────────────
  Type to filter
  ─────────────────────────────────────────────────────

▶ [1] ✅  production    ftp.example.com:21  conn:3  (connected)
  [2] 🖥  staging       ftp2.example.com:21  conn:1
```

- Type any letter to **fuzzy-filter** the list instantly
- `Backspace` to clear search
- `↑↓` navigate · `Enter` select · `[1-9]` quick select · `q` cancel

For `sync` with multiple servers, a **multi-select picker** is shown (`Space` to toggle, `a` to select all, `Enter` to confirm).

---

### `syncftp serve`

Starts a local HTTP API server so external tools (PHP, scripts, web UIs) can interact with syncFTP via HTTP.

```bash
syncftp serve              # http://127.0.0.1:8080
syncftp serve --port 9000
```

Only listens on `127.0.0.1` (localhost). CORS is enabled for all origins so browser-based UIs work out of the box.

#### API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/api/servers` | List all configured servers (passwords excluded) |
| `GET` | `/api/status` | Changed files per server |
| `GET` | `/api/status?server=production` | Changed files for one server |
| `POST` | `/api/sync` | Run sync, returns results |
| `GET` | `/api/failed` | List files that failed in last sync |
| `GET` | `/api/failed?server=production` | Failed files for one server |
| `GET` | `/api/remote/ls?server=&path=` | List remote directory |
| `GET` | `/api/remote/ls?server=&path=&recursive=true` | Recursive listing |
| `GET` | `/api/remote/cat?server=&path=&max_kb=10` | Preview file content |
| `GET` | `/api/remote/get?server=&path=` | Download file (raw stream) |
| `GET` | `/api/remote/get?server=&path=&json=true` | Download file (base64 JSON) |
| `DELETE` | `/api/remote/rm?server=&path=` | Delete file |
| `DELETE` | `/api/remote/rm?server=&path=&recursive=true` | Delete directory recursively |

**All responses follow this envelope:**

```json
{ "ok": true,  "data": { ... } }
{ "ok": false, "error": "error message" }
```

**POST /api/sync — request body:**

```json
{
  "server":       "production",
  "full":         false,
  "dry_run":      false,
  "include":      ["css/", "js/"],
  "exclude":      ["vendor/"],
  "retry_failed": false
}
```

All fields are optional. Omitting `server` syncs all enabled servers.

**POST /api/sync — response:**

```json
{
  "ok": true,
  "data": {
    "server":   "production",
    "dry_run":  false,
    "uploaded": [{ "path": "css/style.css", "attempts": 1 }],
    "failed":   [{ "path": "img/hero.jpg",  "attempts": 3, "error": "connection reset" }],
    "skipped":  [".env"],
    "deleted_locally": ["old-page.php"]
  }
}
```

**PHP examples:**

```php
// List servers
$servers = json_decode(file_get_contents('http://127.0.0.1:8080/api/servers'));

// Check status
$status = json_decode(file_get_contents('http://127.0.0.1:8080/api/status?server=production'));

// Run sync
$ctx = stream_context_create(['http' => [
    'method'  => 'POST',
    'header'  => 'Content-Type: application/json',
    'content' => json_encode(['server' => 'production']),
]]);
$result = json_decode(file_get_contents('http://127.0.0.1:8080/api/sync', false, $ctx));

// List remote files
$files = json_decode(file_get_contents('http://127.0.0.1:8080/api/remote/ls?server=production&path=css/'));

// Preview a file
$res = json_decode(file_get_contents('http://127.0.0.1:8080/api/remote/cat?server=production&path=index.php'));
echo $res->data->content;

// Download a file (raw)
file_put_contents('local.php', file_get_contents(
    'http://127.0.0.1:8080/api/remote/get?server=production&path=index.php'
));

// Delete a file
$ctx = stream_context_create(['http' => ['method' => 'DELETE']]);
$res = json_decode(file_get_contents(
    'http://127.0.0.1:8080/api/remote/rm?server=production&path=old.php',
    false, $ctx
));
```

---

## Filtering System

syncFTP has three independent filtering mechanisms that work together.

### 1. `protect` — permanent server-side protection

Files in `sync.protect` are **never uploaded**, ever. Use this for server-specific files that must not be overwritten (`.env`, database configs, uploaded media directories).

```json
"sync": {
  "protect": [
    ".env",
    "config/database.php",
    "storage/"
  ]
}
```

A trailing `/` means directory prefix match — all files under that path are protected.

### 2. `include` / `exclude` — permanent filters in config

Control which files participate in every sync. Can be set globally or per-server.

```json
"sync": {
  "include": [],
  "exclude": ["vendor/", "node_modules/", "tests/"]
},
"servers": [
  {
    "name": "production",
    "include": ["css/", "js/", "index.php"],
    "exclude": []
  },
  {
    "name": "staging",
    "include": [],
    "exclude": ["vendor/"]
  }
]
```

**Priority (include):** CLI `--include` → server `include` → global `sync.include`

**Priority (exclude):** global `sync.exclude` + server `exclude` + CLI `--exclude` (all combined)

### 3. `--include` / `--exclude` flags — one-shot overrides

Applied only to the current command invocation. Do not affect state or future runs.

```bash
syncftp sync --include css --include js   # this run: only css/ and js/
syncftp sync --exclude tests              # this run: skip tests/
```

### Summary table

| Mechanism | Scope | Where |
|---|---|---|
| `sync.protect` | Permanent, all servers | `syncftp.json` |
| `sync.include` / `sync.exclude` | Permanent, global | `syncftp.json` |
| `server.include` / `server.exclude` | Permanent, per server | `syncftp.json` |
| `--include` / `--exclude` flags | This run only | CLI |

---

## Failed Files & Retry

When uploads fail after all retry attempts, syncFTP saves the list to `.syncftp/failed/<server>.json`.

```bash
# Re-upload only the files that failed last time
syncftp sync --retry-failed
syncftp sync --retry-failed --server production

# Or via the API
curl -X POST http://127.0.0.1:8080/api/sync \
  -H 'Content-Type: application/json' \
  -d '{"server":"production","retry_failed":true}'
```

The failed list is cleared automatically once all files upload successfully.

---

## Connection Pool & Retry

Each server can have its own connection pool and retry settings:

```json
{
  "name": "production",
  "max_connections": 3,
  "max_retries": 2
}
```

- `max_connections` — number of simultaneous FTP connections (default `1`)
- `max_retries` — how many times to retry a failed upload before giving up (default `2`, meaning 3 total attempts)

If some connections in the pool fail to open, syncFTP continues with however many opened successfully (minimum 1). A partial pool warning is printed.

---

## State & Release Files

syncFTP creates a `.syncftp/` directory next to `syncftp.json`:

```
.syncftp/
├── state/
│   ├── production.json    # per-file hashes from last successful sync
│   └── staging.json
├── failed/
│   └── production.json    # files that failed in the last run (cleared on success)
└── releases/
    └── production/
        └── 20260612-143012/
            └── manifest.json   # files and hashes for this release
```

`.syncftp/` is always excluded from scanning and never uploaded to the FTP server.

---

## Internal Structure

| Package | Role |
|---|---|
| `internal/config` | JSON config loading |
| `internal/ignore` | `.gitignore` / `syncftp.ignore` parser |
| `internal/scanner` | Directory walk, SHA256 hashing |
| `internal/state` | Per-server sync state (load / save / diff) |
| `internal/ftp` | FTP client, connection pool, upload, remote operations |
| `internal/failed` | Failed file list persistence |
| `internal/release` | Release manifest writer |
| `cmd/syncftp` | CLI commands (init, status, sync, push, remote, serve, shell) |

---

## Running Tests

```bash
go test ./...                           # all packages
go test ./internal/state/... -v         # verbose state tests
go test ./internal/scanner/...          # scanner tests
```

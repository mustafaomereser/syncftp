package lang

var En = StringSet{
	// Browser TUI
	BrowserLoading:           "  Loading...",
	BrowserSearchPrompt:      "  Enter=Search All  ESC=Cancel",
	BrowserSearching:         "  Searching: %q ...",
	BrowserSearchResults:     "  (full server, %d results)  ESC to clear",
	BrowserPickDirTitle:      "  SELECT TARGET DIRECTORY",
	BrowserPickDirEnter:      "Enter=Select This Dir",
	BrowserHintRight:         "→ Enter",
	BrowserHintLeft:          "← Up",
	BrowserHintQuit:          "q Cancel",
	BrowserMarkedHintDelete:  "d=Delete",
	BrowserMarkedHintMove:    "m=Move",
	BrowserMarkedHintFreeze:  "f=❄ Freeze",
	BrowserMarkedHintAll:     "a=Deselect all",
	BrowserMarkedHintCancel:  "ESC=Cancel",
	BrowserEscPending:        "  ESC again to exit",
	BrowserSearchClear:       "  ESC to clear",
	BrowserHintNav:           "  ↑↓ Navigate",
	BrowserHintEnter:         "Enter Open/Select",
	BrowserHintPreview:       "→ Preview",
	BrowserHintSpace:         "Space Mark",
	BrowserHintSearch:        "/ Search",
	BrowserHintQuitKey:       "q Quit",
	BrowserHintFreeze:        "f=❄ Freeze",
	BrowserHintTree:          "t=Tree",
	BrowserErrPrefix:         "  Error: ",
	BrowserReconnectHint:     "r = Reconnect  |  q = Quit",
	BrowserSelectThisPrefix:  "[ Select this dir: ",
	BrowserSelectThisSuffix:  " ]",
	BrowserCountMany:         "many items",
	BrowserFooterFmt:         "  %d dirs, %d files  ·  %d/%d",
	BrowserSearchMatchFmt:    "  search:%q %d matches",
	BrowserPreviewLoading:    "  Loading...",
	BrowserPreviewBack:       "  ESC/q back",
	BrowserPreviewScrollHint: "↑↓ scroll",
	BrowserPreviewPageHint:   "PgUp/PgDn",
	BrowserPreviewGGHint:     "g top  G bottom",
	BrowserPreviewEmpty:      "(empty file)",
	BrowserPreviewLineFmt:    "%d-%d / %d lines",
	BrowserConnLost:          "⚠ Connection lost",
	BrowserConnLostHint:      "Press r to reconnect",
	BrowserPreviewFailed:     "⚠ Preview unavailable",
	BrowserReconnectFailed:   "reconnect failed: %w",
	BrowserMarkedCount:       "  ✓ %d marked",
	BrowserSearchCancelHint:  "   ESC=cancel",
	BrowserSearchTimeout:     "⚠ timeout (5m) — %d partial results",
	BrowserSearchConnErr:     "⚠ connection error: %s",
	BrowserSearchOpenErr:     "connection %d failed to open: %v",
	BrowserSearchDoneFmt:     "%d results (%.1fs)",
	BrowserSearchOpsHint:     "  ↑↓ navigate  Space=mark  Enter=preview  → preview  d=delete marked  m=move marked  f=freeze  ESC=clear",
	BrowserSearchNoResults:   "  (no results found)",

	// Picker / Confirm TUI
	PickerFilterHint:   "  Type to filter",
	PickerMatchFmt:     "  (%d matches)",
	PickerNoMatch:      "  (no match — Backspace to clear)",
	PickerNavHint:      "↑↓/jk navigate   Enter select   [1-9] quick   Backspace clear   q cancel",
	MultiPickerNavHint: "↑↓ navigate   Space mark   a all   Enter confirm   q cancel",
	ConfirmYes:         "Yes",
	ConfirmNo:          "No, cancel",

	// Sync TUI
	SyncDryRunTitle:      "  [DRY RUN] Files to upload:",
	SyncDryRunFmt:        "  %d files — press any key",
	SyncNoChange:         "  ✓ No changes — up to date",
	SyncAnyKey:           "  Press any key",
	SyncConnecting:       "connecting...",
	SyncRetryFmt:         " (%d attempts)",
	SyncRetryOkFmt:       " (succeeded on attempt %d)",
	SyncDoneFmt:          "  Done: %s uploaded",
	SyncDoneFailFmt:      ", %s failed",
	SyncAttemptDetailFmt: "attempt %d: %v",
	SyncScrollHint:       "  ↑↓/j/k scroll · g=top G=bottom · PgUp/PgDn · any other key = exit",
	SyncFailedHeader:     "failed files",

	// Shell REPL
	ShellWelcomeHint:      "  Type 'help', 'exit' to quit",
	ShellMultiServer:      "Multiple servers — to connect: server <name>",
	ShellServersLabel:     "Servers:",
	ShellCtrlCHint:        "(Ctrl+C — type 'exit' to quit)",
	ShellCtrlCExitHint:    "Press Ctrl+C again to exit — or type 'exit'",
	ShellExit:             "Goodbye.",
	ShellUnknownCmd:       "Unknown command: %q  (type 'help')\n",
	ShellConnecting:       "Connecting: %s (%s)...",
	ShellConnectErr:       " error: %v\n",
	ShellNotConnected:     "Not connected. Use 'server <name>' to connect.",
	ShellNoServers:        "No active servers.",
	ShellServersFmt:       "Servers",
	ShellServersSubtitle:  "Select to connect — q to close",
	ShellAlreadyConn:      "Already connected: %s\n",
	ShellDisconnected:     "Disconnected: %s\n",
	ShellServerNotFound:   "Server not found: %q\n",
	ShellBrowserErr:       "Browser error: %v\n",
	ShellDeleteTitle:      "Delete %d items",
	ShellDeleteSubtitle:   "This cannot be undone. Files will be permanently deleted from the FTP server.",
	ShellCancelled:        "Cancelled.",
	ShellConfirmSure:      "Are you sure?",
	ShellConfirmSureBody:  "Select Yes to confirm.",
	ShellDeletedFmt:       "\n%d deleted, %d failed\n",
	ShellMovingFmt:        "\n%d files to move. Select destination (Enter = select this dir):\n",
	ShellMoveTarget:       "Destination: %s\n\n",
	ShellMovedFmt:         "\n%d moved, %d failed\n",
	ShellDirNotFound:      "Directory not found: %s\n",
	ShellDownloading:      "Downloading: %s → %s\n",
	ShellDownloaded:       "✓ Downloaded (%s)\n",
	ShellDownloadedBare:   "✓ Downloaded",
	ShellErrFmt:           "Error: %v\n",
	ShellDeleteConfTitle:  "Delete Confirmation",
	ShellCancelShort:      "Cancelled.",
	ShellDeletedShort:     "✓ Deleted: %s\n",
	ShellPickServerTitle:  "Select Server",
	ShellSyncServers:          "Servers to Sync",
	ShellCalibrateServers:     "Servers to Calibrate",
	ShellSyncPickSub:      "Space to mark, Enter to confirm",
	ShellSyncCancelled:    "Cancelled.",
	ShellFileActionTitle:  "File Action",
	ShellFileActionView:   "View Contents",
	ShellFileActionGet:    "Download",
	ShellFileActionDel:    "Delete",
	ShellFileActionCancel: "Cancel",
	ShellFileTruncFmt:     "[First %d KB — to get all: get %s]\n",
	ShellScanning:         "Scanning...",
	ShellScannedFmt:       " %d files\n",
	ShellConnPoolErr:      "  Connection error: %v\n",
	ShellReleaseFmt:       "  Release: %s\n",
	ShellStateReadErr:     "  Could not read state: %v\n",
	ShellStatusDeleted:    "  - %s (deleted locally)\n",
	ShellStatusUpToDate:   "  Up to date",
	ShellStatusNoChange:   "  %d changes\n",
	ShellStatusStateErr:   "[%s] could not read state: %v\n",

	TreeMaxPromptTitle: "Tree view",
	TreeMaxPromptSub:   "Max files per directory?",
	TreeSkippedFmt:     "[%d items — skipped, increase with --max]",
	TreeErrFmt:         "[error: %v]",

	ShellHelp: `
Remote server commands:
  ls [dir]                Interactive file browser (↑↓ navigate, → enter/preview, ← back)
                            Space = mark  |  d = delete marked  |  m = move marked
                            f = freeze/unfreeze file or whole folder  |  / = search
                            t = tree view of current directory
  tree [path] [--max N]   Show FTP directory as tree (dirs first, files with sizes)
                            --max N : skip dirs with > N items; show first N + "+X more..."
                            omit --max for interactive prompt (20 / 50 / 100 / all)
  cd <dir>                Change remote directory  (cd .. to go up)
  cat [file]              View file contents (opens browser if no arg)
  get [file] [dest]       Download file (opens browser if no arg)
  rm [-f] [-r] [file]     Delete file/dir  (-f no confirmation, -r recursive)
  pwd                     Show remote directory

Sync:
  status                  Show local changes per server (TUI)
  sync [--all] [--full] [--dry-run] [--server name]   Upload to FTP
  calibrate [--all] [--server name]   Compare local sizes with FTP, update state (no upload)
  freeze [--server name]  Manage freeze list — mark files to never upload

Server management:
  servers                 Server list
  server [name]           Select / connect to server
  disconnect              Close current FTP connection (prompt returns to no-server mode)
  config                  Manage servers and connection profiles — full TUI
                            List: ↑↓ navigate | Enter/e edit | Space enable/disable | d delete | n new
                            ── Global Settings: project-wide default dir, protect, include, exclude
                            ── Connection Profiles: shared FTP credentials (host/port/user/pass)
                               Multiple servers can reference the same profile.
                               Edit profile once → all servers using it are updated.
                            ── Server edit (17 fields):
                               Connection = select a profile (b/Enter to pick)
                               When a profile is active, credential fields are read-only (↳ profile)
                               Local dir: per-server local directory
                                 "(project default)" = inherits from Global Settings
                               Include/Exclude/Protect: b = local file browser
                                 n = enter custom path (glob patterns, paths outside project root)
                            s = save  |  q = cancel

Other:
  lang [en|tr]            Show or change display language (auto-detected from OS on first run)
  clear / cls             Clear screen
  help / ?                This help
  exit / quit             Exit`,

	// lang command
	LangCurrentFmt:  "Language: %s\n",
	LangSwitchedFmt: "Language set to: %s\n",
	LangSavedFmt:    "Saved to .syncftp/lang\n",
	LangInvalid:     "Unknown language: %q — use 'en' or 'tr'\n",
	LangAlreadyFmt:  "Already using %s\n",

	// cmd_sync
	SyncCancelled:      "Cancelled.",
	SyncDryRunNote:     "[DRY RUN] Nothing will be uploaded",
	SyncWhitelistFmt:   "Whitelist (%d paths): only these will be synced\n",
	SyncExcludeFmt:     "Exclude (%d paths): these will be skipped\n",
	SyncServerErrFmt:   "  ERROR: %v\n",
	SyncNoChange2:      "  No changes",
	SyncDeletedHeader:  "  ! DELETED files (kept on FTP):\n",
	SyncProcessingFmt:  "  %d files to process",
	SyncProtectedFmt:   " (%d protected)",
	SyncProtectedLabel: "    PROTECTED  %s\n",
	SyncUploadLabel:    "    WILL UPLOAD %s\n",
	SyncPoolFmt:        "  Connection pool: %d / Retry: %d\n",
	SyncDoneFullFmt:    "  Done: %d uploaded, %d protected, %d errors\n",
	SyncFailedSavedFmt: "  ! %d failed files saved to .syncftp/failed/%s.json — retry with: syncftp sync --retry-failed\n",
	SyncFailedSaveErr:  "  ! Could not save failed list: %v\n",
	SyncStateErr:       "  ! Could not save state: %v\n",
	SyncReleaseErr:     "  ! Could not create release: %v\n",
	SyncReleaseFmt:     "  Release: %s\n",
	SyncFullFlag:       "  Full sync (--full): all files will be uploaded",
	SyncRetryNoFiles:   "  No failed files to retry",
	SyncRetryModeFmt:   "  Retry mode: %d failed files (%s)\n",
	SyncRetrySkipFmt:   "  ! %s no longer exists locally, skipping\n",
	SyncAttemptsFmt:    "    ✗ %s (%d attempts): %v\n",
	SyncAttemptOkFmt:   "    ✓ %s (succeeded on attempt %d)\n",
	SyncUploadOkFmt:    "    ✓ %s\n",
	SyncUploadErrFmt:   "    ✗ %s: %v\n",

	// resync
	ResyncScanning:     "  Connecting to server...",
	ResyncConnected:    " connected\n",
	ResyncListing:      "Listing remote files...",
	ResyncLocalFmt:     "  Local:  %d files scanned\n",
	ResyncConnErr:      "  ! Could not connect to server (%v) — resync skipped\n",
	ResyncListErr:      "  ! Could not read remote dir (%v) — resync skipped\n",
	ResyncFoundFmt:     "  Remote: %d files found\n",
	ResyncComparingFmt:   "  Comparing: %d / %d\n",
	ResyncMatchedFmt:     "  Matched: %d (size OK)  |  Different/missing: %d\n",
	ResyncDoneFmt:        "  [%s] calibrate done\n",
	ResyncAutoMsg:        "  First run: running calibrate (server comparison)...\n",
	ResyncNoServers:      "No matching servers found.\n",
	ResyncIgnoreDirsFmt:  "  Ignore: %d dir(s) skipped → %s\n",
	ResyncIgnoreFilesFmt: "  Ignore: %d file(s) skipped\n",
	ResyncFilteredFmt:    "  Filter (include/exclude): %d file(s) out of scope\n",
	ResyncFrozenDiffFmt:  "  ❄ %d frozen file(s) differ/missing (will be skipped on sync)\n",

	// cmd_status
	StatusWhitelistFmt:     "Whitelist (%d paths): only these will be shown\n",
	StatusExcludeFmt:       "Exclude (%d paths): these will be hidden\n",
	StatusProjectFmt:       "Project: %s\n",
	StatusDirFmt:           "Dir    : %s\n",
	StatusFileFmt:          "Files  : %d\n",
	StatusStateErr:         "[%s] Could not load state: %v\n\n",
	StatusNoFirstSync:      "  No first sync yet",
	StatusUpToDate:         "  No changes — server up to date",
	StatusFilteredUpToDate: "  No changes in specified filter scope",
	StatusNewHeader:        "  + NEW",
	StatusChangedHeader:    "  ~ CHANGED",
	StatusDeletedHeader:    "  - DELETED (not removed from FTP, info only)",

	// cmd_push
	PushSourceFmt:     "Source  : %s\n",
	PushTargetFmt:     "Target  : %s:%d%s\n",
	PushScanning:      "Scanning...",
	PushScannedFmt:    " %d files\n\n",
	PushFirstPush:     "First push",
	PushFullPush:      "Full push (--full)",
	PushFirstFmt:      "  %s — all files will be uploaded\n",
	PushDeletedHeader: "  ! DELETED files (kept on FTP):\n",
	PushNoChange:      "  No changes — target up to date",
	PushProcessingFmt: "  %d files to process\n",
	PushDryRunHeader:  "\n[DRY RUN] Files to upload:",
	PushConnFmt:       "  Connections: %d / Retry: %d\n",
	PushAttemptsFmt:   "    ✗ %s (%d attempts): %v\n",
	PushAttemptOkFmt:  "    ✓ %s (attempt %d)\n",
	PushUploadOkFmt:   "    ✓ %s\n",
	PushUploadErrFmt:  "    ✗ %s: %v\n",
	PushDoneFmt:       "\n  Done: %d uploaded, %d errors\n",
	PushFailedHint:    "  ! Failed files saved — retry: syncftp push %s --server %s --full\n",
	PushStateErr:      "  ! Could not save state: %v\n",
	PushReleaseFmt:    "  Release: %s\n",

	// cmd_remote
	RemoteServerFmt:      "Server : %s (%s)\n",
	RemoteDirFmt:         "Dir    : %s\n\n",
	RemoteConnecting:     "Connecting: %s (%s)...\n",
	RemoteDownloading:    "Downloading: %s → %s\n",
	RemoteDownloaded:     "✓ Downloaded (%s)\n",
	RemoteDownloadedBare: "✓ Downloaded",
	RemoteDeleteLabel:    "Deleting %s: %s\n",
	RemoteDeleteConfirm:  "Are you sure you want to delete? [y/N]: ",
	RemoteDeleteCancel:   "Cancelled.",
	RemoteDeletedFmt:     "✓ Deleted: %s\n",
	RemoteCatTruncFmt:    "\n[First %d KB shown — to download all: syncftp remote get %s]\n",
	RemotePickDirErr:     "could not list directory: %w",
	RemotePickUpDir:      "  [0] .. (parent dir)",
	RemotePickPromptFmt:  "  Select [0-%d, q=cancel]: ",
	RemotePickInvalid:    "  Invalid selection: %q\n",
	RemotePickSelected:   "  Selected: %s\n\n",
	RemoteFileLabel:      "file",
	RemoteDirLabel:       "dir",
	RemoteDirRecLabel:    "dir (with contents)",
	RemoteListErr:        "%s  ! could not list: %v\n",
	RemoteNoServerSel:    "no server selected",

	// failsafe cleanup
	SyncCleanupRemovedFmt: "  ⚠ %q no longer exists locally, removed from list\n",
	SyncConfigSaveErr:     "  ! Config could not be saved: %v\n",

	// status TUI
	StatusDetailChangesCountFmt: "  %d changes",
	StatusSyncConfirm:           "  Start sync?",
	StatusSyncHint:              "  Enter/s = Yes  |  ESC = Cancel",
	StatusSearchHint:            "  Backspace = clear  |  Esc = close search",
	StatusDetailNavHint:         "  ↑↓/g/G scroll  |  / = search  |%s  ←/Esc = back  |  q = quit",
	StatusDetailSyncPart:        "  s = Sync  |",
	StatusNoMatch:               "  (no matches)",
	StatusUpToDateShort:         "  ✓ up to date",
	StatusScrollFmt:             "  %d/%d",
	StatusResultsFmt:            "  %d results",
	StatusListTitleFmt:          "  Status — %s",
	StatusListNavHint:           "  ↑↓ navigate  |  → = details  |  q = quit",
	StatusOkShort:               "✓ up to date",
	StatusChangesListFmt:        "%d changes",
	StatusFrozenFmt:             "❄ %d frozen",

	// cmd_init
	InitWizardTitle:   "=== syncFTP Setup Wizard ===",
	InitProjectName:   "Project name",
	InitLocalDir:      "Local directory",
	InitFTPHeader:     "FTP Server details:",
	InitServerName:    "Server name",
	InitHost:          "FTP host",
	InitPort:          "Port",
	InitUser:          "Username",
	InitPassword:      "Password",
	InitRemotePath:    "Remote directory",
	InitCreated:       "✓ syncftp.json created (permissions: 600)",
	InitReadyFmt:      "Ready! Run 'syncftp sync' to synchronize project %q.\n",
	InitIgnoreExists:  "  (syncFTP entries already in %s)\n",
	InitIgnoreAdded:   "✓ syncftp.json, syncftp.exe → added to %s\n",
	InitIgnoreCreated: "✓ syncftp.ignore created (syncftp.json, syncftp.exe added)",

	// Config TUI — server manager
	CfgSrvTitle:          "⚙  Server Settings",
	CfgSrvNavHint:        "  ↑↓ navigate  |  Enter/e = edit  |  Space = enable/disable  |  d = delete  |  n = new  |  q = quit",
	CfgDeleteConfirmFmt:  "[%s] delete?  y = yes, any other = cancel",
	CfgGlobalLabel:       "⚙ Global Settings",
	CfgGlobalHint:        "  (protect, include, exclude, ignore_files)",
	CfgConnProfilesLabel: "🔗 Connection Profiles",
	CfgConnCountFmt:      "(%d profiles)",
	CfgNewServerLabel:    "+ Add new server",
	CfgSrvCountFmt:       "  %d servers  |  %d connection profiles",
	CfgSaveErr:           "Save error: %v\n",
	CfgUpdatedFmt:        "✓ [%s] updated\n",
	CfgAddedFmt:          "✓ [%s] added\n",
	CfgDeletedFmt:        "✓ [%s] deleted\n",
	CfgEnabledLabel:      "enabled",
	CfgDisabledLabel:     "disabled",

	// Config TUI — server edit
	CfgSrvEditTitle:   "Edit Server",
	CfgSrvNewTitle:    "New Server",
	CfgSrvEditNavHint: "  ↑↓ navigate  |  Enter/b = edit/select  |  Space = bool toggle  |  s = save  |  q = cancel",
	CfgSrvSaveHint:    "  s = save  |  q/Esc = cancel",
	CfgSrvManual:      "(manual)",
	CfgSrvProjectDef:  "(project default)",
	CfgSrvConnChange:  "[b=change]",
	CfgSrvConnSelect:  "[b=select profile]",
	CfgSrvBrowseHint:  "[b=browse]",

	// Config TUI — server field labels
	CfgFldName:        "Name",
	CfgFldConnection:  "Connection",
	CfgFldHost:        "Host",
	CfgFldPort:        "Port",
	CfgFldUser:        "User",
	CfgFldPassword:    "Password",
	CfgFldRemotePath:  "Remote path",
	CfgFldLocalPath:   "Local path",
	CfgFldEnabled:     "Enabled",
	CfgFldPassive:     "Passive mode",
	CfgFldDisableEPSV: "Disable EPSV",
	CfgFldNAT:         "NAT workaround",
	CfgFldMaxConn:     "Max connections",
	CfgFldMaxRetry:    "Max retry",
	CfgFldInclude:     "Include",
	CfgFldExclude:     "Exclude",
	CfgFldProtect:     "Protect",

	// Config TUI — global settings
	CfgGlobalEditTitle:    "⚙  Global Settings",
	CfgGlobalEditNavHint:  "  ↑↓ navigate  |  Enter = edit/toggle  |  b = browse local dir  |  s = save  |  q = cancel",
	CfgGlobalDefaultEmpty: "(empty = working directory)",
	CfgGlobalIgnoreDefVal: "(default: .gitignore + syncftp.ignore)",
	CfgGlobalIgnoreNone:   "(none — ignore files not loaded)",
	CfgGlobalIgnoreHint:   "  b=browse to add files  |  Enter=manage  |  s=save  |  q=cancel",

	// Config TUI — global field labels
	CfgGFldDefaultPath: "Default directory",
	CfgGFldProtect:     "Protect",
	CfgGFldInclude:     "Include (global)",
	CfgGFldExclude:     "Exclude (global)",
	CfgGFldIgnore:      "Ignore files",

	CfgValueEmpty: "(empty)",

	// Config TUI — connection manager
	CfgConnMgrTitle:   "🔗  Connection Profiles",
	CfgConnMgrNavHint: "  ↑↓ navigate  |  Enter/e = edit  |  d = delete  |  n = new  |  q = quit",
	CfgNewConnLabel:   "+ Add new profile",
	CfgConnTotalFmt:   "  %d profiles",

	// Config TUI — connection edit
	CfgConnEditTitle:   "🔗  Edit Profile",
	CfgConnNewTitle:    "🔗  New Profile",
	CfgConnEditNavHint: "  ↑↓ navigate  |  Enter = edit/toggle  |  Space = bool toggle  |  s = save  |  q = cancel",

	// Config TUI — connection picker
	CfgConnPickerTitle:      "Select Connection Profile",
	CfgConnPickerSub:        "FTP credentials will be taken from this profile",
	CfgConnPickerCurrentFmt: "Current: %s",
	CfgConnPickerManual:     "(manual)",
	CfgConnPickerManualDesc: "Enter credentials manually",

	// Config TUI — local browser
	CfgLocalNavHint:      "  Space=mark  |  n=custom path  |  Enter/→=enter  |  ←/Esc=back  |  a=all  |  s=save  |  q=cancel",
	CfgLocalDirPickHint:  "  ↑↓ navigate  |  Enter/→=enter dir  |  Space/s=select this dir  |  ←/Esc=back  |  q=cancel",
	CfgLocalCustomHint:   "  You can enter a path outside the project dir (e.g. ../shared/config.php)",
	CfgLocalEmpty:        "  (empty folder)",
	CfgLocalSelectFmt:    "  Current: %s  |  Space or s = select this dir",
	CfgLocalPathLabel:    "  Path: ",
	CfgLocalCustomSave:   "  Enter=add  |  Esc=cancel",
	CfgLocalMarkedFmt:    "  ✓ %d marked  +%d custom paths  |  n=add custom path",
	CfgLocalMarkedSimple: "  ✓ %d marked  |  n=add custom path",

	// Config TUI — ignore picker
	CfgIgnoreTitle:   "⚙  Ignore files",
	CfgIgnoreNavHint: "  Space = select/deselect  |  s = save  |  q = cancel",
	CfgIgnoreNote:    "  If none selected, ignore files will not be loaded.",
}

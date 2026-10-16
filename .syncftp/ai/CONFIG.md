# .syncftp/syncftp.json schema

```jsonc
{
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
```

Notes:
- Server names must be unique; a server with `connection` must reference an existing profile.
- `deploy[].provider`: github | gitlab | gitea | bitbucket | generic. Webhook URL: `<public>/hook/<project>/<name>` (see the CI/CD tab).
- Ignore rules (gitignore syntax) live in `.syncftp/syncftp.ignore`.
- Freeze lists: `syncftp freeze`; milestones: `syncftp milestone set --date ... -s NAME`.

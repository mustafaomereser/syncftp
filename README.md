# syncFTP

Yerel projelerdeki değişen dosyaları tespit edip birden fazla FTP sunucusuna dağıtan CLI aracı.

- Değişiklik tespiti SHA256 hash ile yapılır — git gerektirmez
- Birden fazla FTP sunucusunu tek komutla günceller
- Sunucu tarafındaki config dosyalarını (`.env`, `database.php` vb.) korur, üzerine yazmaz
- Şifreler ayrı dosyada tutulur, ana config git'e commit edilebilir

---

## Kurulum

**Go 1.21+** gereklidir. https://go.dev/dl/

```bash
git clone <repo>
cd syncftp
go mod tidy
go build -o syncftp.exe ./cmd/syncftp/   # Windows
go build -o syncftp ./cmd/syncftp/       # Linux / macOS
```

---

## Hızlı Başlangıç

```bash
# 1. Proje dizinine girin
cd /path/to/your/project

# 2. Config oluşturun (sihirbaz açılır)
syncftp init

# 3. Neyin değiştiğini görün
syncftp status

# 4. FTP'ye yükleyin
syncftp sync
```

---

## Config Dosyaları

syncFTP iki ayrı dosya kullanır:

### `syncftp.toml` — Ana config (git'e commit edilebilir)

```toml
[project]
name = "my-project"
local_path = "."          # hangi klasör taransın (. = bulunduğun yer)

[sync]
# Bu dosyalar/dizinler FTP'de asla üzerine yazılmaz
protect = [
  ".env",
  "config/database.php",
  "storage/",
]

# Boşsa tüm değişen dosyalar senkronize edilir.
# Dolu ise sadece bu yollar gönderilir.
include = []

[first_sync]
# true  → ilk sync'te tüm dosyaları gönder
# false → sadece include listesini gönder
full = true

[[servers]]
name       = "production"
host       = "ftp.example.com"
port       = 21
user       = "ftpuser"
remote_path = "/public_html"
passive    = true
enabled    = true

[[servers]]
name       = "staging"
host       = "ftp2.example.com"
port       = 21
user       = "staginguser"
remote_path = "/staging"
passive    = true
enabled    = true
```

### `syncftp.config` — Şifreler (git'e commit ETMEYİN)

`syncftp init` komutu bu dosyayı oluşturur ve otomatik olarak `.gitignore`'a ekler.

```toml
[[servers]]
name     = "production"
password = "gizli_sifre_1"

[[servers]]
name     = "staging"
password = "gizli_sifre_2"
```

> Şifreler `name` alanı eşleştirilerek `syncftp.toml`'daki sunuculara merge edilir.

---

## Ignore Dosyaları

Proje kökünde hangisi varsa o kullanılır (öncelik sırası):

1. `.gitignore` — zaten varsa otomatik kullanılır
2. `syncftp.ignore` — git kullanmıyorsanız bu dosyayı oluşturun

Format `.gitignore` ile birebir aynıdır:

```gitignore
node_modules/
*.log
.DS_Store
dist/
uploads/
```

---

## Komutlar

### `syncftp init`

İnteraktif sihirbaz. Proje adı, FTP bilgileri sorulur ve `syncftp.toml` + `syncftp.config` oluşturulur.

```
=== syncFTP Kurulum Sihirbazı ===

Proje adı [my-project]:
Yerel dizin [.]:

Sunucu adı [production]:
FTP host: ftp.example.com
Port [21]:
Kullanıcı adı: ftpuser
Şifre: ****
Uzak dizin [/public_html]:
```

---

### `syncftp status`

Değişen dosyaları listeler. **Hiçbir şey yüklemez.**

```bash
syncftp status                            # tüm değişiklikleri göster
syncftp status --include css              # sadece css/ altındaki değişiklikleri göster
syncftp status --exclude vendor           # vendor/ hariç göster
syncftp status --include src --exclude src/__tests__   # kombine
```

```
Proje : my-project
Dizin : C:\projeler\my-project
Dosya : 142 adet

── production (ftp.example.com) ──
  + YENİ (2):
    js/utils.js
    css/dark-mode.css
  ~ DEĞİŞEN (1):
    index.php
  - SİLİNEN (FTP'den silinmez, sadece bilgi) (1):
    old-file.php
```

---

### `syncftp sync`

Değişen dosyaları tüm aktif sunuculara yükler.

```bash
syncftp sync                        # tüm sunuculara
syncftp sync --server production    # sadece production'a
syncftp sync --full                 # tüm dosyaları (state'i yoksay)
syncftp sync --dry-run              # ne yükleneceğini göster, yükleme

# Whitelist: sadece belirtilen dosya/klasörü sync et
syncftp sync --include css
syncftp sync --include css --include js/app.js   # birden fazla yol

# Exclude: belirtilen dosya/klasörü bu sync'ten hariç tut (tek seferlik)
syncftp sync --exclude vendor
syncftp sync --exclude vendor --exclude tests

# Kombine kullanım
syncftp sync --include src/components --exclude src/components/__tests__
syncftp sync --include css --dry-run              # önce dry-run ile kontrol
```

> **`--include` ile TOML `include` farkı:** `--include` sadece o anki çalıştırma için geçerlidir ve `syncftp.toml`'daki `sync.include`'u geçersiz kılar. Kalıcı whitelist için TOML'u kullanın.

> **`--exclude` ile `protect` farkı:** `protect` (TOML) kalıcı ve sunucu tarafındaki dosyaları korumak içindir. `--exclude` tek seferlik ve geçici hariç tutma içindir.

**Örnek çıktı:**

```
Taranıyor: C:\projeler\my-project
142 dosya bulundu

══ production (ftp.example.com) ══
  3 dosya işlenecek
    ✓ js/utils.js
    ✓ css/dark-mode.css
    ✓ index.php
    KORUNUYOR  .env
  Tamamlandı: 3 yüklendi, 1 korundu, 0 hata
  Release: .syncftp\releases\production\20260611-143012
```

---

## Protect (Koruma) Nasıl Çalışır?

`sync.protect` listesindeki dosya ve dizinler FTP'de **hiçbir zaman güncellenmez**. Bu sayede sunucu tarafındaki özel config dosyaları korunur.

```toml
[sync]
protect = [
  ".env",              # tam dosya adı eşleşmesi
  "config/app.php",    # alt dizindeki dosya
  "storage/",          # dizin sonu / ile dizin eşleşmesi (tüm içeriği korur)
]
```

---

## Include / Exclude Nasıl Çalışır?

syncFTP'de filtrelemeyi iki farklı şekilde yapabilirsiniz:

### Kalıcı whitelist — `syncftp.toml`

Her sync'te sadece belirli dosya/klasörleri göndermek istiyorsanız TOML'a ekleyin:

```toml
[sync]
include = [
  "public/",     # sadece public/ klasörü
  "index.php",   # ve index.php
]
```

`include` boşsa tüm değişen dosyalar senkronize edilir.

### Tek seferlik — `--include` ve `--exclude` flag'leri

O anki sync için geçici filtre uygulamak istiyorsanız CLI flag'lerini kullanın:

```bash
# Sadece css/ ve js/ klasörlerini bu sefer sync et
syncftp sync --include css --include js

# vendor/ ve tests/ hariç tüm değişiklikleri sync et
syncftp sync --exclude vendor --exclude tests

# src/components altını sync et ama test dosyalarını atla
syncftp sync --include src/components --exclude src/components/__tests__

# Önce ne yükleneceğini gör, sonra gerçekten yükle
syncftp sync --include css --dry-run
syncftp sync --include css
```

`--include` verilirse TOML'daki `sync.include`'u geçersiz kılar. `--exclude` ise her durumda ek olarak uygulanır.

---

## İç Yapı

| Dizin | Görevi |
|---|---|
| `internal/config` | TOML okuma, `syncftp.config` şifre merge'i |
| `internal/ignore` | `.gitignore` / `syncftp.ignore` parser |
| `internal/scanner` | Dosya ağacı tarama, SHA256 hesaplama |
| `internal/state` | Per-server sync durumu (`.syncftp/state/`) |
| `internal/ftp` | FTP bağlantı, upload, dizin oluşturma |
| `internal/release` | Release manifest (`.syncftp/releases/`) |
| `cmd/syncftp` | CLI komutları (init, status, sync) |

---

## State ve Release Dosyaları

syncFTP proje kökünde `.syncftp/` dizini oluşturur:

```
.syncftp/
├── state/
│   ├── production.json    # hangi dosyaların hangi hash ile yüklendiği
│   └── staging.json
└── releases/
    └── production/
        └── 20260611-143012/
            └── manifest.json   # o release'teki dosyalar ve hash'leri
```

> `.syncftp/` dizini her zaman taramadan hariç tutulur — FTP'ye gönderilmez.

---

## Testler

```bash
go test ./...                          # tüm testler
go test ./internal/state/...           # sadece state testleri
go test ./internal/scanner/... -v      # verbose çıktı
```

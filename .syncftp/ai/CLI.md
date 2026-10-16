# syncFTP command reference

Generated from the program. Global flag: `-P, --project NAME` runs a command in another registered project.

## syncftp calibrate

Yerel dosyaları FTP boyutlarıyla karşılaştırıp state'i günceller (yükleme yapmaz)

```
syncftp calibrate [flags]
```

Flags:
```
      --all             Tüm aktif sunucuları kalibre et
  -s, --server string   Sadece bu sunucuyu kalibre et
```

## syncftp config

Server ayarlarını yönet (ekle, düzenle, sil)

```
syncftp config [flags]
```

Flags:
```
      --export   Config'i şifresiz JSON olarak stdout'a yaz
```

## syncftp detect-limit

Sunucunun eş zamanlı bağlantı limitini ölçer ve max_connections'a yazar

```
syncftp detect-limit [flags]
```

Bağlantılar birer birer açılır (giriş + NOOP ile doğrulanır) ta ki sunucu reddedene
ya da --max'a ulaşılana kadar. Sonra hepsi tekrar yoklanır — fazla bağlantıyı kabul edip
sessizce düşüren sunucular da doğru ölçülür. Sonuç aynı hesabı (host+port+kullanıcı)
paylaşan TÜM sunucuların max_connections alanına yazılır.

Not: o anda açık başka bağlantılar (shell, FileZilla…) ölçülen değeri düşürür.

Flags:
```
      --all             Tüm aktif sunucular
      --max int         Denenecek en fazla bağlantı (default 32)
      --no-save         Sonucu config'e yazma, sadece göster
  -s, --server string   Sunucu adı
```

## syncftp diff

İki sunucunun FTP içeriğini karşılaştırır (production vs staging)

```
syncftp diff [sunucuA] [sunucuB]
```

## syncftp freeze

Mark files to never upload for a specific server

```
syncftp freeze [flags]
```

Flags:
```
      --server string   Server to configure freeze list for
```

## syncftp global

Proje bağımsız ayarlar: kayıtlı bağlantılar ve yeni proje varsayılanları

```
syncftp global
```

Makine geneli syncFTP ayarları (`<config dizini>/syncftp/library.json`).

  Kayıtlı bağlantılar  init --connection / push --connection / panelde "kütüphaneden" ile kullanılır.
  Varsayılanlar        yeni oluşturulan her projeye uygulanan global sync ayarları.

Örnekler:
  syncftp global
  syncftp global conn add hostingim --host ftp.example.com --user admin --pass gizli
  syncftp global conn import ../eski-proje
  syncftp global defaults from .
  syncftp global defaults template laravel

## syncftp global conn

Kayıtlı bağlantılar

```
syncftp global conn
```

## syncftp global conn add

Bağlantı ekle / güncelle

```
syncftp global conn add <ad> [flags]
```

Flags:
```
      --active            Aktif mod (varsayılan passive)
      --host string       Sunucu adresi
      --key string        SFTP özel anahtar dosyası
      --pass string       Şifre (anahtarlık varsa oraya yazılır)
      --port int          Port (varsayılan 21/22)
      --protocol string   ftp | ftps | sftp (default "ftp")
      --user string       Kullanıcı adı
```

## syncftp global conn import

Bir projenin bağlantılarını kütüphaneye aktar

```
syncftp global conn import [proje]
```

## syncftp global conn list

Kayıtlı bağlantıları listele

```
syncftp global conn list
```

## syncftp global conn rm

Bağlantıyı sil

```
syncftp global conn rm <ad>
```

## syncftp global defaults

Yeni proje varsayılanları

```
syncftp global defaults
```

## syncftp global defaults from

Bir projenin global sync ayarlarını varsayılan yap

```
syncftp global defaults from [proje]
```

## syncftp global defaults reset

Varsayılanları fabrika ayarına döndür

```
syncftp global defaults reset
```

## syncftp global defaults template

Varsayılan ignore şablonu

```
syncftp global defaults template <laravel|wordpress|node|generic|none>
```

## syncftp global show

Kütüphaneyi ve varsayılanları göster

```
syncftp global show
```

## syncftp gui

Masaüstü yönetim panelini açar (projeler, sync, rollback, watch, ayarlar)

```
syncftp gui [flags]
```

Tüm kayıtlı projeleri tek pencereden yönetmek için grafik arayüz.

Exe'ye çift tıklamak da paneli açar. Terminalden argümansız 'syncftp' her zamanki
shell'i başlatır; 'syncftp gui' paneli ayrı süreçte açar ve terminali serbest bırakır.

Panel yalnızca 127.0.0.1'den, oturuma özel token ile erişilebilir.

Flags:
```
      --background   Yalnızca tray'de başla, pencere açma (Windows açılışında başlat bunu kullanır)
      --foreground   Paneli bu süreçte çalıştır (terminal bloklanır)
      --no-window    Pencere açma, giriş adresini yazdır (uzak/headless kullanım; --foreground ile)
```

## syncftp hook

CI/CD: GitHub/GitLab/Gitea/Bitbucket push webhook'u ile otomatik deploy

```
syncftp hook
```

Projelerin config'indeki "deploy" tetikleyicileri için webhook dinleyicisi.

Push → https://<public>/hook/<proje>/<tetikleyici> → imza doğrulama → [git pull] → sync.
Tetikleyiciler panelden (CI/CD sekmesi) ya da syncftp.json'daki "deploy" dizisinden tanımlanır.
Panel (tray) açıkken dinleyici zaten çalışır; GUI'siz makinelerde 'syncftp hook serve' kullanın.

## syncftp hook list

Tüm projelerdeki deploy tetikleyicilerini ve URL yollarını listeler

```
syncftp hook list
```

## syncftp hook serve

Webhook dinleyicisini başlatır (GUI'siz sunucular için)

```
syncftp hook serve [flags]
```

Flags:
```
      --addr string   Dinlenecek adres (varsayılan 127.0.0.1:47821; VPS'te doğrudan açmak için 0.0.0.0:47821)
      --tunnel        cloudflared quick tunnel aç (trycloudflare public URL)
```

## syncftp init

syncftp.json dosyasını oluşturur

```
syncftp init [flags]
```

Bu dizinde syncFTP projesi oluşturur. Argümansız: interaktif sihirbaz.
Global varsayılanlar (syncftp global) yeni projeye uygulanır; kayıtlı bağlantılardan seçilebilir.

Örnekler:
  syncftp init
  syncftp init -y --connection hostingim --remote /public_html --template laravel
  syncftp init -y --protocol sftp --host example.com --user deploy --key ~/.ssh/id_ed25519 --remote /var/www
  syncftp init -y --no-server

Flags:
```
      --connection string    Kayıtlı bağlantıyı kullan (syncftp global conn list)
      --host string          Sunucu adresi
      --key string           SFTP özel anahtar dosyası
      --local string         Varsayılan yerel dizin (varsayılan: .)
      --name string          Proje adı (varsayılan: klasör adı)
      --no-server            Sunucusuz oluştur (sonra eklenir)
      --pass string          Şifre
      --port int             Port (varsayılan protokole göre 21/22)
      --protocol string      ftp | ftps | sftp
      --remote string        Uzak dizin (varsayılan: /public_html)
      --server-name string   İlk sunucunun adı (varsayılan: production)
      --template string      Ignore şablonu: laravel | wordpress | node | generic | none
      --user string          Kullanıcı adı
  -y, --yes                  Soru sormadan oluştur (flag'ler + varsayılanlar)
```

## syncftp lang

Show or change the display language

```
syncftp lang [en|tr]
```

## syncftp milestone

Sunucu başına zaman işareti: 'milestone sync' o tarihten sonra değişen dosyaları yükler

```
syncftp milestone
```

Flags:
```
      --all             Tüm aktif sunucular (seçici açılmaz)
  -s, --server string   Sadece bu sunucu
```

## syncftp milestone clear

Milestone'u sil

```
syncftp milestone clear
```

## syncftp milestone set

Milestone koy (varsayılan: şimdi; --date ile özel tarih)

```
syncftp milestone set [flags]
```

Flags:
```
      --date string   Özel tarih: bugün, dün, 3d, 5h, 14:30, 20.07[.2026], 2026-07-20 [15:04]; verilmezse şimdi
```

## syncftp milestone show

Kayıtlı milestone'ları göster

```
syncftp milestone show
```

## syncftp milestone sync

Milestone tarihinden sonra değişen (mtime) dosyaları yükle

```
syncftp milestone sync [flags]
```

Flags:
```
      --dry-run   Ne yükleneceğini göster, gerçekten yükleme
```

## syncftp project

Kayıtlı projeleri yönetir (çoklu proje — GUI ve --project için)

```
syncftp project
```

syncFTP, config'i olan her projeyi makine genelindeki bir kayda ekler
(<UserConfigDir>/syncftp/projects.json). Kayıtlı projeler GUI'de listelenir ve
herhangi bir dizinden --project ile hedeflenebilir:

  syncftp -P magaza sync --all
  syncftp --project D:\siteler\blog status

## syncftp project add

Projeyi kayda ekler (varsayılan: bulunulan dizin)

```
syncftp project add [yol] [flags]
```

Flags:
```
      --name string   Görünen ad (varsayılan: klasör adı)
```

## syncftp project list

Kayıtlı projeleri listeler

```
syncftp project list
```

## syncftp project rename

Projenin görünen adını değiştirir

```
syncftp project rename <ad|yol> <yeni-ad>
```

## syncftp project rm

Projeyi kayıttan çıkarır (dosyalara dokunmaz)

```
syncftp project rm <ad|yol>
```

## syncftp push

Bir dizini doğrudan sunucuya yükler (config gerektirmez)

```
syncftp push <local-dizin> [flags]
```

Herhangi bir dizini FTP / FTPS / SFTP sunucusuna senkronize eder.
syncftp.json olmadan da çalışır. State <local-dizin>/.syncftp/ altında saklanır.

Örnekler:
  syncftp push ./website --host ftp.example.com --user admin --pass gizli --remote /public_html
  syncftp push ./website --connection hostingim --remote /public_html   (kayıtlı bağlantı)
  syncftp push ./website --protocol sftp --host example.com --user deploy --key ~/.ssh/id_ed25519 --remote /var/www
  syncftp push ./website --server production          (varsa syncftp.json'dan okur)
  syncftp push ./website --host ftp.example.com --user admin --pass gizli --dry-run

Flags:
```
      --connection string     Kayıtlı bağlantıyı kullan (syncftp global conn list)
      --connections int       Paralel bağlantı sayısı (default 1)
      --dry-run               Ne yükleneceğini göster, yükleme
      --exclude stringArray   Bu yol/klasörleri hariç tut
      --full                  State'i yoksay, tüm dosyaları yükle
      --host string           Sunucu adresi
      --include stringArray   Sadece bu yol/klasörleri yükle
      --key string            SFTP özel anahtar dosyası
      --pass string           Şifre
      --passive               Passive mode (default true)
      --port int              Port (varsayılan protokole göre 21/22)
      --protocol string       ftp | ftps | sftp (varsayılan ftp)
      --remote string         Uzak hedef dizin (default "/")
      --retries int           Hata durumunda yeniden deneme sayısı (default 2)
      --server string         syncftp.json'dan sunucu adı (--host yerine)
      --user string           Kullanıcı adı
```

## syncftp remote

FTP sunucusundaki dosyaları listele, indir, sil, önizle

```
syncftp remote
```

Flags:
```
      --server string   Bağlanılacak sunucu adı (tek sunucu varsa otomatik seçilir)
```

## syncftp remote cat

FTP'deki dosyanın içeriğini görüntüle

```
syncftp remote cat [remote-path] [flags]
```

Flags:
```
      --max-kb int   Gösterilecek maksimum kilobayt (default 10)
```

## syncftp remote exec

SFTP sunucusunda SSH ile komut çalıştırır (uzak dizinde)

```
syncftp remote exec <komut>
```

Komut, sunucunun remote_path dizininde SSH üzerinden çalışır; çıktı canlı akar.
Yalnızca protocol=sftp sunucularda kullanılabilir.

  syncftp remote exec -s production "php artisan migrate --force"
  syncftp remote exec -s production "composer install --no-dev && php artisan optimize"

## syncftp remote get

FTP'den dosya indir

```
syncftp remote get [remote-path] [local-dest]
```

## syncftp remote ls

FTP sunucusundaki dosya ve dizinleri listele

```
syncftp remote ls [path] [flags]
```

Flags:
```
      --recursive   Alt dizinleri de listele
```

## syncftp remote rm

FTP'den dosya veya dizin sil

```
syncftp remote rm [remote-path] [flags]
```

Flags:
```
      --force       Onay sormadan sil
      --recursive   Dizini içeriğiyle birlikte sil
```

## syncftp rollback

Bir sync'i geri alır: eski dosyaları geri yükler, yeni eklenenleri siler

```
syncftp rollback [release-id] [flags]
```

Seçilen release'i ve ondan SONRAKİ tüm release'leri en yeniden başlayarak geri alır
(yani sunucu o sync'ten önceki haline döner).

Her sync, yüklediği dosyaların sunucudaki eski halini .syncftp/releases/<sunucu>/<id>/files/
altına yedekler; sunucuda olmayan dosyaları "yeni" olarak kaydeder. Rollback yedekleri geri
yükler, yeni dosyaları siler ve state'i eski hash'lere çeker — sonraki sync yerel değişiklikleri
tekrar tespit eder.

Yedeklemeyi kapatmak: sync.no_backup / server.no_backup. Tutulacak yedek sayısı: sync.backup_keep (varsayılan 10).

Flags:
```
      --list            Release geçmişini listele, geri alma yapma
  -s, --server string   Sunucu adı
  -y, --yes             Onay sormadan geri al
```

## syncftp secure

Şifreleri işletim sisteminin anahtarlığına taşır (Credential Manager / Keychain)

```
syncftp secure [flags]
```

Açıkken config'deki tüm şifreler (sunucu, bağlantı profili, SSH tüneli) Windows Credential
Manager / macOS Keychain / Linux Secret Service'e yazılır; syncftp.json'da yalnızca "keyring:"
işareti kalır. Kayıtlar hesap bazlıdır (protokol://kullanıcı@host:port) — sunucu adından
bağımsızdır, aynı hesabı kullanan projeler paylaşır.

--off şifreleri tekrar düz metin olarak config'e yazar.

Flags:
```
      --off   Anahtarlık modunu kapat, şifreleri tekrar config'e yaz
```

## syncftp serve

Yerel HTTP API sunucusu başlatır (PHP/web arayüzleri için)

```
syncftp serve [flags]
```

Flags:
```
      --port int   Dinlenecek port (default 8080)
```

## syncftp status

Değişen dosyaları listeler (hiçbir şey yüklemez)

```
syncftp status [flags]
```

Flags:
```
      --exclude stringArray   Bu yol/klasörleri sonuçtan hariç tut
      --include stringArray   Sadece bu yol/klasörleri göster (whitelist)
  -s, --server string         Sadece bu sunucunun durumunu göster
```

## syncftp sync

Değişen dosyaları FTP sunucularına yükler

```
syncftp sync [flags]
```

Flags:
```
      --all                   Tüm aktif sunuculara sync et (seçici açılmaz)
      --dry-run               Ne yükleneceğini göster, gerçekten yükleme
      --exclude stringArray   Bu yol/klasörleri bu sync'ten hariç tut (tek seferlik)
      --full                  State'i yoksay, tüm dosyaları yükle
      --include stringArray   Sadece bu yol/klasörleri sync et (whitelist; JSON include'u geçersiz kılar)
      --retry-failed          Önceki sync'te başarısız olan dosyaları tekrar yükle
      --server string         Sadece belirtilen sunucuya yükle
```

## syncftp verify

Sunucuyu state ile karşılaştırır: elle değiştirilen/silinen dosyaları bulur

```
syncftp verify [flags]
```

State'e göre güncel olması gereken her dosyanın sunucudaki boyutunu yerel dosyayla
karşılaştırır (CRLF farkı toleranslı). Sunucuda eksik ya da boyutu farklı dosyalar
"sapma" olarak raporlanır — biri sunucuda elle değişiklik yapmış ya da dosya silinmiş demektir.

--fix sapan dosyaları state'ten düşer; sonraki sync onları yeniden yükler.

Flags:
```
      --all             Tüm aktif sunucular
      --fix             Sapan dosyaları state'ten düş → sonraki sync yeniden yükler
  -s, --server string   Sunucu adı
```

## syncftp watch

Dosya değişikliklerini izle, değişince otomatik sync yap

```
syncftp watch [flags]
```

Flags:
```
      --all             Tüm aktif sunucuları izle
      --server string   Sadece belirtilen sunucuyu izle
```


# Changelog

## 2.1.0

### tr
- Yerelde silinen dosyaları sunucudan da silme (profil ya da proje başına açılır, varsayılan kapalı): her dosya önce yedeklenir, Geçmiş'te görünür ve geri alınabilir; freeze / protect listesindekilere ve ignore edilip diskte duran dosyalara dokunulmaz
- Windows bildirimleri: panel kapalı ya da arka plandayken başarısız işler ve izleyici hataları tepsiden bildirilir, tıklayınca panel o projede açılır; Ayarlar → Genel'den kapatılabilir, başarılı işler için de açılabilir
- Yardım: soldaki içindekiler okunan bölümü işaretliyor (scroll spy); aramada eşleşen metin vurgulanıyor, içindekiler ve komut listesi daralıyor

### en
- Optionally delete locally removed files on the server (per profile or per project, off by default): each file is backed up first, shows up in History and can be rolled back; freeze / protect entries and ignored files still on disk are never touched
- Windows notifications: while the panel is closed or in the background, failed jobs and watch errors are notified from the tray; clicking opens the panel on that project; can be turned off, or on for successful jobs, under Settings → General
- Help: the table of contents follows the section you read (scroll spy); search highlights matches and narrows the contents and command list

## 2.0.9

### tr
- Performans: izleyiciler açıkken panel boşta sürekli yeniden çiziliyordu (bir çekirdeğin ~%20'si); artık ~%2
- İzleyici göstergesi yanıp sönmek yerine sabit nokta; proje kartındaki "watch" rozeti "izleniyor" oldu
- Pencere küçültülmüşken proje listesi arka planda tazelenmiyor; tepsi ikonu yalnızca durum değişince güncelleniyor

### en
- Performance: with watches running the panel kept repainting while idle (~20% of a core); now ~2%
- The watch indicator is a steady dot instead of a pulsing one
- The project list is not refreshed while the window is minimized; the tray icon updates only when its state changes

## 2.0.8

### tr
- Filtreler sekmeli: Dahil et · Hariç tut · Koru · Uzantılar · Yok sayma dosyaları (her sekmede kayıt sayısı)
- Türkçe arayüzde terimler Türkçe: izleyiciler (watch), dahil et / hariç tut / koru / yok say
- Hook'lar açıklamalı: ne zaman çalıştığı, hata verirse ne olduğu, örnek komutlar ve komutlara verilen değişkenler
- CI/CD sekmesi: üstte dışarıdan erişim durumu ve tek tıkla tünel, boşken kurulum adımları; tünel kapalıyken yerel adres artık "public adres" diye gösterilmiyor

### en
- Filters are now tabs: Include · Exclude · Protect · Extensions · Ignore files (with counts)
- Hooks explain when they run, what a failure does, example commands and the variables commands receive
- CI/CD tab: reachability and a one-click tunnel at the top, setup steps when empty; the local address is no longer shown as the public address while the tunnel is off

## 2.0.7

### tr
- Sağ alttaki bildirimler ve iş kartları tutup aşağı ya da sağa atılarak kapatılabiliyor (log penceresi açılmadan)

### en
- Drag a notification or job card in the bottom right down or to the right to dismiss it without opening its log

## 2.0.6

### tr
- Güncellemeler artık saatte bir denetleniyor
- Çok sayıda watch ya da iş açıkken panel "yükleniyor"da kalmıyor (tüm iş logları tek bağlantıdan akıyor)
- Durdurulan watch, konsoldaki watch listesinden hemen kalkıyor

### en
- Updates are now checked every hour
- The panel no longer hangs on "loading" with many watches or jobs running (all job logs share one connection)
- A stopped watch disappears from the watch switcher in the console right away

## 2.0.5

### tr
- Freeze listesinde Gözat artık sunucudaki dosyaları gösteriyor
- Freeze listesinde Liste | Klasör görünümü
- Profil kartında "geri alınabilir" rozeti ve Dosyalar düğmesi yeniden çalışıyor

### en
- Browse in the freeze list now shows the files on the server
- List | Folders view in the freeze list
- The "can roll back" badge and the Files button on profile cards work again

## 2.0.4

### tr
- Proje kartlarında sync bekleyen dosya sayısı rozeti
- Tüm projeler sayfasında "Değişiklikler" görünümü: bütün projelerde bekleyen dosyalar tek ekranda, toplu sync
- Profil kartındaki her bilgi tıklanabilir: yerel klasör Explorer'da açılır, sunucu adresi bağlantı ayarlarına, uzak dizin Dosyalar'a, son sync Geçmiş'e götürür
- Proje başlığındaki ve proje kartındaki klasör yoluna tıklayınca klasör açılır
- Profil kartına sağ tık ⋯ menüsünü açar
- Yedek rozetine basınca o profilde yedek açılır, açıkken onay sorup kapatır
- Doğrula, Ağaç, Karşılaştır ve SSH komut işleri durdurulabiliyor; Doğrula listeleme sırasında ilerleme gösteriyor
- Süren işin kartı tıklanınca kaybolmuyor, log penceresi kapanınca yeniden açılabiliyor
- Renkli rozetler (yedekli, yedek kapalı, başarısız dosyalar) artık kendi renginde

### en
- Pending-file badge on project cards
- "Changes" view on the projects page: pending files across all projects on one screen, with bulk sync
- Every detail on a profile card is clickable: the local folder opens in the file manager, the server address opens the connection settings, the remote folder opens Files, the last sync opens History
- Clicking the folder path in the project header or on a project card opens the folder
- Right-click a profile card for its ⋯ menu
- Click the backup badge to turn backups on for that profile, or (after a confirmation) off
- Verify, Tree, Compare and SSH command jobs can be stopped; Verify shows progress while listing
- A running job's card stays in the dock after you open its log
- Colored badges (backed up, no backup, failed files) now show their colors

## 2.0.3

### tr
- macOS ve Linux'ta panel artık gerçek uygulama penceresinde açılıyor, üst bar ve Hızlı bağlan orada da var
- Güncellemeden sonra "Yenilikler" penceresi; Ayarlar → Sistem'den de açılır
- Yeni sürüm açılamazsa güncelleme kendiliğinden geri alınıyor
- Mac klavyesinde silme tuşu seçilen dosyaları siliyor
- Panel her açılışta aynı adresi kullanıyor; görünüm tercihleri ve editör sekmeleri artık kaybolmuyor
- Hızlı bağlantı sayfasında AI asistan bu sayfayı proje sanmıyor

### en
- On macOS and Linux the panel now opens in a real app window, with the top bar and Quick connect
- A "What's new" window after updating; also available in Settings → System
- If a new version fails to start, the update rolls back by itself
- The delete key on Mac keyboards deletes the selected files
- The panel keeps the same address between launches, so view preferences and editor tabs are no longer lost
- The AI assistant no longer treats the Quick connect page as a project

## 2.0.1

### tr
- Tek tıkla güncelleme: yeni sürüm çıkınca üst barda Güncelle düğmesi

### en
- One-click updates: an Update button appears in the top bar when a new version is out

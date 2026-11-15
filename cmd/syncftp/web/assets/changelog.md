# Changelog

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

'use strict';

/* Yardım sayfası (#/help[/bölüm]) — görev odaklı rehber + cobra'dan üretilen CLI referansı. */

Object.assign(T.tr, { navHelp: 'Yardım', helpTitle: 'Yardım', helpSub: 'syncFTP nasıl kullanılır — panel ve komut satırı', helpSearch: 'Yardımda ara…', helpCli: 'Komut satırı referansı', helpShell: 'İnteraktif shell komutları', helpGlobal: 'Genel flag\'ler', helpToc: 'İçindekiler', helpNoMatch: 'Eşleşen bölüm yok' });
Object.assign(T.en, { navHelp: 'Help', helpTitle: 'Help', helpSub: 'How to use syncFTP — panel and command line', helpSearch: 'Search help…', helpCli: 'Command line reference', helpShell: 'Interactive shell commands', helpGlobal: 'Global flags', helpToc: 'Contents', helpNoMatch: 'No matching section' });

// Her bölüm: [id, ikon, {tr:[başlık, html], en:[başlık, html]}]
const HELP = [
  ['start', 'play', {
    tr: ['Başlarken', `<ol>
      <li><b>Proje ekle</b> — sol menüde <kbd>+</kbd> → proje klasörünü seçin. syncFTP config'i yoksa sizin için boş bir tane oluşturulur. Terminalde <code>syncftp init</code> yaptığınız projeler listede kendiliğinden görünür.</li>
      <li><b>Sunucu ekle</b> — <i>Sunucular → Sunucu ekle</i>. Protokol (FTP / FTPS / SFTP), host, kullanıcı, şifre ya da SSH anahtarı ve <b>uzak dizin</b> (örn. <code>/public_html</code>) yeterli. <i>Bağlantıyı test et</i> ile kaydetmeden deneyin.</li>
      <li><b>İlk sync</b> — <i>Genel bakış → Sync</i>. Sunucu daha önce elle doldurulmuşsa ilk sync önce <b>calibrate</b> çalıştırır: dosya boyutlarını karşılaştırıp aynı olanları "güncel" sayar, yalnızca farklı olanları yükler.</li>
      <li>Sonrasında her sync <b>sadece değişen dosyaları</b> yükler (SHA256 ile tespit).</li></ol>`],
    en: ['Getting started', `<ol>
      <li><b>Add a project</b> — <kbd>+</kbd> in the sidebar → choose the project folder. If it has no syncFTP config, an empty one is created. Projects where you ran <code>syncftp init</code> show up automatically.</li>
      <li><b>Add a server</b> — <i>Servers → Add server</i>. Protocol (FTP / FTPS / SFTP), host, user, password or SSH key and the <b>remote path</b> (e.g. <code>/public_html</code>). Use <i>Test connection</i> before saving.</li>
      <li><b>First sync</b> — <i>Overview → Sync</i>. If the server already has files, the first sync runs <b>calibrate</b> first: it compares file sizes, treats identical ones as up to date and uploads only the rest.</li>
      <li>After that every sync uploads <b>only changed files</b> (detected with SHA256).</li></ol>`],
  }],
  ['calibrate', 'scan', {
    tr: ['Calibrate: sunucudakini tanıtmak', `<p>Sunucu daha önce elle (ya da başka bir araçla) doldurulduysa <b>calibrate</b> yerel dosyaları sunucudaki boyutlarla karşılaştırır (CRLF farkını hesaba katar) ve aynı olanları "güncel" sayar — hiçbir şey yüklemez. İlk sync bunu kendiliğinden yapar; state kaybolduysa kart menüsünden elle çalıştırın.</p>
      <ul><li>Log penceresinde o an ne yaptığı canlı görünür: yerel tarama, bağlanma, listelenen dosya sayısı ve en son yol, karşılaştırma.</li>
      <li><b>Durdur</b> calibrate'i anında keser — sunucu cevap vermiyor olsa bile; state'e dokunulmaz.</li>
      <li>Otomatik calibrate başarısız olursa (sunucuya ulaşılamadı vb.) o sunucunun sync'i iptal edilir; eskisi gibi her dosya yeniden yüklenmez.</li>
      <li>Freeze'li ve ignore edilen klasörlere sunucuda girilmez.</li></ul>`],
    en: ['Calibrate: matching what is on the server', `<p>If the server was filled by hand (or another tool), <b>calibrate</b> compares local files with the server's file sizes (CRLF-aware) and marks identical ones as up to date — nothing is uploaded. The first sync does this automatically; run it from the card menu if the state was lost.</p>
      <ul><li>The log window shows what it is doing live: local scan, connecting, files listed with the latest path, comparing.</li>
      <li><b>Stop</b> cancels it at once — even when the server is not answering; the state is left untouched.</li>
      <li>If the automatic calibrate fails (server unreachable etc.), that server's sync is aborted instead of re-uploading every file.</li>
      <li>Frozen and ignored folders are not walked on the server.</li></ul>`],
  }],
  ['sync', 'upload', {
    tr: ['Değişiklikleri görmek ve sync', `<ul>
      <li>Sunucu kartındaki <span class="degisim" style="cursor:default"><b class="n">+yeni</b> <b class="c">~değişen</b> <b class="d">−silinen</b></span> çipine tıklayın → dosya listesi. Bir dosyaya tıklayınca <b>yerel ↔ sunucu farkı</b> renkli olarak açılır.</li>
      <li><b>Dry-run</b>: neyin yükleneceğini gösterir, hiçbir şey yüklemez. <b>Tam sync</b>: state'i yok sayıp her şeyi yükler.</li>
      <li>Yüklenemeyen dosyalar kırmızı rozetle görünür → tıklayıp <b>Başarısızları tekrar dene</b>.</li>
      <li>Yerelde silinen dosyalar sunucudan <b>silinmez</b>, yalnızca raporlanır (güvenli taraf).</li>
      <li>Her iş sağ altta canlı görünür; tıklayınca tam log açılır. Loglar <i>Geçmiş</i> sekmesinde saklanır. Durdurulabilen işlerde (sync, calibrate, watch) log penceresinde <b>Durdur</b> düğmesi vardır. Sync durdurulunca yeni dosya başlatılmaz, açık aktarımlar kesilir; yüklenenler kaydedilir (geri alınabilir), kalanlar bekleyen olarak kalır ve bir sonraki sync'te gider.</li>
      <li>Sunucuya bağlanılamazsa hata satırı log'a yazılır; CLI'da <code>sync</code> 1 koduyla çıkar (script / CI başarısızlığı görür).</li></ul>`],
    en: ['Seeing changes and syncing', `<ul>
      <li>Click the <span class="degisim" style="cursor:default"><b class="n">+new</b> <b class="c">~changed</b> <b class="d">−deleted</b></span> chip on a server card → file list. Clicking a file opens a coloured <b>local ↔ server diff</b>.</li>
      <li><b>Dry-run</b> shows what would be uploaded without uploading. <b>Full sync</b> ignores the state and uploads everything.</li>
      <li>Files that failed show a red badge → click it and <b>Retry failed</b>.</li>
      <li>Locally deleted files are <b>not</b> deleted from the server, only reported (safe side).</li>
      <li>Every job appears live at the bottom right; click it for the full log. Logs are kept under <i>History</i>. Jobs that can be cancelled (sync, calibrate, watch) have a <b>Stop</b> button in their log window. Stopping a sync starts no new files and cuts the running transfers; what was uploaded is recorded (and can be undone), the rest stays pending for the next sync.</li>
      <li>If the server cannot be reached, the error is written to the log; on the CLI <code>sync</code> exits with code 1 (scripts and CI see the failure).</li></ul>`],
  }],
  ['filters', 'sliders', {
    tr: ['Filtreler: neyin gideceğini belirlemek', `<ul>
      <li><b>Ignore</b> (<code>.gitignore</code> + <i>Ayarlar → Ignore kuralları</i>): bu dosyalar hiç taranmaz; <code>vendor/</code>, <code>node_modules/</code> gibi klasörlere yerelde de sunucuda da girilmez. Hazır şablonlar (Laravel, WordPress, Node) var.</li>
      <li><b>Include</b>: yalnızca bu yollar sync edilir. <b>Exclude</b>: bu yollar atlanır. <b>Protect</b>: sunucudaki bu dosyaların (örn. <code>.env</code>) üzerine asla yazılmaz. <i>Ağaçtan seç</i> ile dosya ağacından işaretleyebilirsiniz.</li>
      <li><b>Freeze</b>: sunucu başına "değişse de yükleme" listesi (kart menüsü ya da Dosyalar sekmesinde ❄). Bir <b>klasörü</b> freeze'lemek tek bir <code>klasör/</code> kaydıdır: altındaki her dosya (sonradan eklenenler dahil) yüklenmez, sunucuda listeleme yapılmadığı için binlerce dosyalı klasörde bile anında olur ve calibrate o klasöre girmez. Freeze'li klasörün içindeki tek bir dosya ayrıca çözülemez — önce klasörü çözün.</li>
      <li><b>Engellenen uzantılar</b>: <code>.zip .psd</code> gibi uzantılar hiç yüklenmez.</li>
      <li><b>Milestone</b>: bir tarih koyun; o tarihten önce değişmiş dosyalar sync dışında kalır. <i>Milestone sync</i> yalnızca sonrasını yükler.</li>
      <li>Dosyalar sekmesindeki <b>ignore / exclude / include dışı</b> rozetleri sync'in o dosyayı hiç yüklemeyeceğini gösterir. Bu filtreler <b>yerel taraf</b> içindir; sunucudaki kopyaya dokunulmaz, silinmez.</li>
      <li>Yerelde artık olmayan tam dosya yolları include / exclude listelerinden sync öncesi otomatik çıkarılır. Güvenlik için <b>protect</b> listelerine hiç dokunulmaz ve include listesi asla boşaltılmaz (boş include = her şey).</li></ul>`],
    en: ['Filters: deciding what gets uploaded', `<ul>
      <li><b>Ignore</b> (<code>.gitignore</code> + <i>Settings → Ignore rules</i>): never scanned; folders like <code>vendor/</code> and <code>node_modules/</code> are not walked locally or on the server. Templates for Laravel, WordPress, Node are included.</li>
      <li><b>Include</b>: only these paths are synced. <b>Exclude</b>: these are skipped. <b>Protect</b>: these server files (e.g. <code>.env</code>) are never overwritten. Use <i>Pick from tree</i> to tick them in the file tree.</li>
      <li><b>Freeze</b>: per-server "never upload even if changed" list (card menu, or ❄ in the Files tab). Freezing a <b>folder</b> is a single <code>folder/</code> entry: nothing inside it (including files added later) is uploaded; nothing is listed on the server, so it is instant even for huge folders, and calibrate skips that folder. A file inside a frozen folder cannot be unfrozen on its own — unfreeze the folder.</li>
      <li><b>Blocked extensions</b>: extensions like <code>.zip .psd</code> are never uploaded.</li>
      <li><b>Milestone</b>: set a date; files changed before it stay out of syncs. <i>Milestone sync</i> uploads only what changed after it.</li>
      <li>The <b>ignore / exclude / not included</b> badges in the Files tab mark files that sync will never upload. These filters apply to the <b>local side</b>; the copy on the server is left alone, never deleted.</li>
      <li>Exact file paths that no longer exist locally are removed from include / exclude before a sync. For safety <b>protect</b> lists are never touched and an include list is never emptied (empty include = everything).</li></ul>`],
  }],
  ['watch', 'eye', {
    tr: ['Watch: kaydedince otomatik yükle', `<p><i>Genel bakış → Watch başlat</i>. Dosya kaydettiğiniz anda (işletim sistemi olayıyla, CPU harcamadan) değişen dosyalar yüklenir. Ignore edilen klasörler izlenmez, editör geçici dosyaları yok sayılır.</p>
      <p>Pencereyi kapatsanız da watch tray'de çalışmaya devam eder; syncFTP yeniden açıldığında açık watch'lar kaldığı yerden başlar. Tray menüsünden tek tıkla durdurabilirsiniz.</p>`],
    en: ['Watch: upload on save', `<p><i>Overview → Start watch</i>. Changed files are uploaded the moment you save (OS events, no CPU while idle). Ignored folders are not watched and editor temp files are skipped.</p>
      <p>Watch keeps running in the tray when you close the window, and running watches resume when syncFTP starts again. Stop them from the tray menu with one click.</p>`],
  }],
  ['rollback', 'undo', {
    tr: ['Geri alma (rollback) ve güvenli yükleme', `<ul>
      <li>Her sync, değiştirdiği dosyaların <b>sunucudaki eski halini yedekler</b>. <i>Geçmiş</i> sekmesinde bir release'te <b>Geri al</b> → o sync ve sonrakiler geri alınır: eski dosyalar bayt bayt geri yüklenir. Yeni eklenen dosyalar varsayılan olarak <b>sunucuda kalır</b>; onay penceresindeki <i>Sonradan eklenen dosyaları da sil</i> kutusu işaretlenince (silinecekler listelenir) silinir — CLI'da <code>--delete-new</code>. Bir release kısmen geri alınamazsa daha eskilerine dokunulmaz.</li>
      <li><b>Ne değişti?</b> Release'in <i>Detaylar</i>'ı solda değişen dosyaları klasör ağacında, sağda seçtiğiniz dosyanın <b>farkını</b> gösterir (sunucudaki önceki hal → yüklenen hal; ayrıca önceki / yüklenen hal sekmeleri). Her release 2 MB'a kadar dosyaların yüklenen halini de saklar. CLI: <code>syncftp rollback &lt;id&gt; -s &lt;sunucu&gt; --diff[=yol]</code>.</li>
      <li><b>Seçili dosyaları geri al</b>: ağaçta dosyaları (ya da klasörün tamamını) işaretleyip <i>Seçilenleri geri al</i> → yalnızca o dosyalar release'ten önceki hâline döner (release'in eklediği dosyalar yalnızca kutu işaretlenirse silinir). Aynı dosyalar daha yeni release'lerde de değiştiyse onlarda da geri alınır; release'ler "kısmen geri alındı" olarak işaretlenir ve sonraki tam rollback bu dosyalara tekrar dokunmaz. CLI: <code>syncftp rollback &lt;id&gt; --files a.php,b.css</code>.</li>
      <li>Yüklemeler <b>atomiktir</b>: dosya önce geçici adla yüklenip yerine konur; bağlantı koparsa canlıda yarım dosya kalmaz.</li>
      <li>Yarıda kalan sync (elektrik, kopma) tekrar başlatılınca <b>kaldığı yerden</b> devam eder.</li>
      <li><i>Doğrula</i>: sunucuda elle değiştirilen ya da silinen dosyaları bulur; <i>Doğrula ve düzelt</i> sonraki sync'te onları yeniden yükletir.</li></ul>`],
    en: ['Rollback and safe uploads', `<ul>
      <li>Every sync <b>backs up the previous server version</b> of the files it replaces. In <i>History</i>, <b>Roll back</b> a release → that sync and every newer one are undone: old files come back byte for byte. Newly added files <b>stay on the server</b> by default; tick <i>Also delete the newly added files</i> in the confirmation (it lists them) to delete them — <code>--delete-new</code> on the CLI. If a release can only be partly rolled back, older ones are left untouched.</li>
      <li><b>What changed?</b> A release's <i>Details</i> shows its changed files as a folder tree on the left and the <b>diff</b> of the selected file on the right (previous server version → uploaded; plus before / uploaded tabs). Each release keeps the uploaded version of files up to 2 MB. CLI: <code>syncftp rollback &lt;id&gt; -s &lt;server&gt; --diff[=path]</code>.</li>
      <li><b>Roll back selected files</b>: tick files (or a whole folder) in the tree → <i>Roll back selected</i> → only those files return to their version before the release (files it added are deleted only if you tick the box). Newer releases that changed the same files are rolled back for them too; releases are marked "partly rolled back" and a later full rollback does not touch those files again. CLI: <code>syncftp rollback &lt;id&gt; --files a.php,b.css</code>.</li>
      <li>Uploads are <b>atomic</b>: written under a temp name and renamed into place, so a dropped connection never leaves a half file live.</li>
      <li>An interrupted sync <b>resumes</b> where it stopped.</li>
      <li><i>Verify</i> finds files changed or deleted on the server by hand; <i>Verify & fix</i> makes the next sync re-upload them.</li></ul>`],
  }],
  ['files', 'folder', {
    tr: ['Dosyalar: gezinme ve sunucudaki dosyayı düzenleme', `<ul>
      <li><b>Dosyalar</b> sekmesi sunucuyu gezer: önizleme (renkli; HTML ayrıca güvenli çerçevede), <b>yerelle karşılaştır</b>, indir, yükle, klasör, yeniden adlandır / <b>taşı</b> (uzak klasör seçiciyle), çoklu seçim, freeze, alt klasörlerde canlı arama, ağaç görünümü.</li>
      <li><b>Düzenle</b>: dosyanın tamamı (4 MB'a kadar; binary ve UTF-8 olmayan dosyalar yalnızca görüntülenir) renkli bir editörde açılır — Tab / Shift+Tab girinti, <kbd>Ctrl</kbd>+<kbd>S</kbd> kaydet.</li>
      <li>Kaydederken: <b>Sunucu + yerel</b> (yerel dosya da yazılır, CRLF korunur, state güncellenir) ya da <b>Yalnızca sunucu</b> (state değişmez; yereldeki dosya değişmedikçe sync düzenlemeyi ezmez). "Bir daha sorma" ile seçim hatırlanır; <i>Ayarlar → Genel</i>'den ya da editörün altındaki <i>Tekrar sor</i>'dan değiştirilir.</li>
      <li>Kayıt normal upload'tan geçer: sunucudaki eski hal yedeklenir, <i>Geçmiş</i>'ten geri alınabilir. Dosyayı açtığınızdan beri sunucuda biri değiştirdiyse <b>üzerine yaz</b> ya da <b>sunucudakini aç</b> diye sorulur.</li>
      <li><b>Biçimlendir</b>: JS / CSS / JSON / HTML / XML dosyalarında önizleme ve karşılaştırmanın üstünde <i>Orijinal | Biçimlendir</i> seçimi vardır. Sıkıştırılmış (minify) dosyalar kendiliğinden okunur hâlde açılır. Yalnızca görünümdür — sunucudaki ya da yereldeki dosya değişmez; internet gerekmez.</li>
      <li>Önizlemede kısayollar: <kbd>S</kbd> düzenle, <kbd>R</kbd> yerelle karşılaştır, <kbd>D</kbd> indir. Düzenleme / karşılaştırmada <kbd>Esc</kbd> önizlemeye döner.</li></ul>`],
    en: ['Files: browsing and editing server files', `<ul>
      <li>The <b>Files</b> tab browses the server: preview (highlighted; HTML also in a sandboxed frame), <b>compare with local</b>, download, upload, new folder, rename / <b>move</b> (with a remote folder picker), multi-select, freeze, live recursive search, tree view.</li>
      <li><b>Edit</b>: the whole file (up to 4 MB; binary and non-UTF-8 files are view-only) opens in a highlighted editor — Tab / Shift+Tab indent, <kbd>Ctrl</kbd>+<kbd>S</kbd> saves.</li>
      <li>On save: <b>Server + local</b> (the local file is written too, CRLF kept, state updated) or <b>Server only</b> (state unchanged; sync won't overwrite the edit unless the local file changes). "Don't ask again" remembers the choice; change it in <i>Settings → General</i> or with <i>Ask again</i> under the editor.</li>
      <li>Saving goes through the normal upload: the old server version is backed up and can be undone from <i>History</i>. If someone changed the file on the server since you opened it, you are asked to <b>overwrite</b> or <b>open the server version</b>.</li>
      <li><b>Beautify</b>: for JS / CSS / JSON / HTML / XML files the preview and compare views have an <i>Original | Beautify</i> switch. Minified files open in the readable form automatically. It is view-only — the file on the server or on disk is not changed; works offline.</li>
      <li>Preview shortcuts: <kbd>S</kbd> edit, <kbd>R</kbd> compare with local, <kbd>D</kbd> download. <kbd>Esc</kbd> in edit / compare returns to the preview.</li></ul>`],
  }],
  ['conns', 'link', {
    tr: ['Bağlantı profilleri ve kütüphane', `<ul>
      <li><b>Bağlantı profili</b>: aynı FTP/SFTP hesabını kullanan sunucular (production, staging…) bilgileri tek yerden paylaşır. Profil seçili sunucu formunda kimlik alanları yerine profilin özeti ve <i>Profili düzenle</i> görünür.</li>
      <li><b>Sunucudan al</b> (Bağlantılar sekmesi) ya da sunucu formundaki <b>Profile çevir</b>: elle girilmiş bilgileri profile taşır; aynı hesap zaten profilse sunucu ona bağlanır. CLI: <code>syncftp config extract-conn &lt;sunucu&gt;</code>.</li>
      <li><b>Kütüphane</b> (<i>Ayarlar → Bağlantı kütüphanesi</i>): projeler arası kayıtlı bağlantılar. <i>Kütüphaneden</i> ile projeye ekleyin, <i>Kütüphaneye kaydet</i> ile kaydedin — kütüphanede olanlar <b>Kütüphanede</b> rozetiyle görünür; aynı adlı başka hesabın üzerine yazmadan önce sorulur.</li></ul>`],
    en: ['Connection profiles and library', `<ul>
      <li><b>Connection profile</b>: servers on the same FTP/SFTP account (production, staging…) share the details from one place. A server using a profile shows the profile's summary and <i>Edit profile</i> instead of credential fields.</li>
      <li><b>From a server</b> (Connections tab) or <b>Make it a profile</b> in the server form moves hand-entered details into a profile; if a profile for the same account exists, the server is linked to it. CLI: <code>syncftp config extract-conn &lt;server&gt;</code>.</li>
      <li><b>Library</b> (<i>Settings → Connection library</i>): connections saved across projects. Add with <i>From library</i>, save with <i>Save to library</i> — entries already in it show an <b>In library</b> badge; overwriting a same-named entry of another account asks first.</li></ul>`],
  }],
  ['cicd', 'zap', {
    tr: ['CI/CD: git push ile otomatik deploy', `<ol>
      <li>Proje → <b>CI/CD</b> sekmesi → <b>Tetikleyici ekle</b>: sağlayıcı (GitHub, GitLab, Gitea, Bitbucket ya da genel), branch (örn. <code>main</code>), hedef sunucular, isterseniz <b>önce git pull</b>.</li>
      <li><b>Tüneli başlat</b>: syncFTP, Cloudflare'in ücretsiz <i>trycloudflare</i> tüneliyle size herkese açık bir https adresi verir (cloudflared yoksa panel indirir). Panel asla dışarı açılmaz — yalnızca imzalı webhook istekleri kabul edilir.</li>
      <li><b>Kurulum</b> düğmesindeki adımlarla URL ve secret'ı sağlayıcının webhook ayarına yapıştırın.</li>
      <li>Artık her push'ta: imza doğrulanır → branch kontrol edilir → <code>git pull</code> → sunuculara sync (yedekli, rollback'li).</li></ol>
      <p>trycloudflare adresi her başlatmada değişir; kalıcı adres için kendi domain'inizi girin. GUI'siz bir sunucuda: <code>syncftp hook serve --tunnel</code>.</p>`],
    en: ['CI/CD: deploy on git push', `<ol>
      <li>Project → <b>CI/CD</b> tab → <b>Add trigger</b>: provider (GitHub, GitLab, Gitea, Bitbucket or generic), branch (e.g. <code>main</code>), target servers, optionally <b>git pull first</b>.</li>
      <li><b>Start tunnel</b>: syncFTP gets a public https address through Cloudflare's free <i>trycloudflare</i> tunnel (the panel downloads cloudflared if missing). The panel itself is never exposed — only signed webhook requests are accepted.</li>
      <li>Follow the steps under <b>Setup</b> to paste the URL and secret into your provider's webhook settings.</li>
      <li>From now on every push: signature verified → branch checked → <code>git pull</code> → sync to servers (with backup and rollback).</li></ol>
      <p>The trycloudflare address changes on every start; enter your own domain for a permanent one. On a server without GUI: <code>syncftp hook serve --tunnel</code>.</p>`],
  }],
  ['ssh', 'tunnel', {
    tr: ['SFTP, SSH komutları ve tünel', `<ul>
      <li><b>SFTP</b> dosyaları SSH üzerinden gönderir — FTP'si olmayan VPS'lere de deploy edebilirsiniz. Şifre ya da SSH anahtarı (<code>~/.ssh/id_ed25519</code>) kullanılır; sunucu anahtarı ilk bağlantıda kaydedilir, değişirse bağlantı reddedilir.</li>
      <li><b>Uzak komutlar</b> (yalnızca SFTP): sunucu düzenleme → <i>SSH komutları</i>. <code>remote_post_sync</code> ile deploy sonrası sunucuda <code>php artisan migrate --force</code>, <code>composer install --no-dev</code>, cache temizleme vb. çalışır. Kart menüsündeki <i>Sunucuda komut çalıştır</i> ile tek seferlik komut çalıştırabilirsiniz.</li>
      <li><b>FTPS</b>: TLS ile şifreli FTP. Paylaşımlı hosting sertifikası host adıyla uyuşmuyorsa <i>Sertifika doğrulamasını atla</i>.</li>
      <li><b>SSH tüneli</b>: sunucuya doğrudan erişilemiyorsa bağlantı bir atlama sunucusu (bastion) üzerinden açılır.</li></ul>`],
    en: ['SFTP, SSH commands and tunnel', `<ul>
      <li><b>SFTP</b> sends files over SSH — deploy to VPSs without FTP. Password or SSH key (<code>~/.ssh/id_ed25519</code>); the host key is recorded on first connect and a changed key is refused.</li>
      <li><b>Remote commands</b> (SFTP only): edit server → <i>SSH commands</i>. With <code>remote_post_sync</code>, <code>php artisan migrate --force</code>, <code>composer install --no-dev</code>, cache clears etc. run on the server after a deploy. Use <i>Run command on server</i> in the card menu for one-off commands.</li>
      <li><b>FTPS</b>: FTP encrypted with TLS. Tick <i>Skip certificate check</i> if a shared hosting certificate does not match the host name.</li>
      <li><b>SSH tunnel</b>: if the server is not directly reachable, connect through a jump host (bastion).</li></ul>`],
  }],
  ['security', 'lock', {
    tr: ['Güvenlik', `<ul>
      <li><i>Ayarlar → Güvenlik</i>: şifreleri <b>Windows Credential Manager</b>'a taşıyın — <code>syncftp.json</code>'da yalnızca <code>keyring:</code> işareti kalır.</li>
      <li>Panel yalnızca bu bilgisayardan (127.0.0.1) ve oturuma özel anahtarla açılır; şifreler sayfaya hiç gönderilmez.</li>
      <li><code>.syncftp/</code> klasörü <code>.gitignore</code>'a otomatik eklenir.</li>
      <li><b>HTTP API</b> (<code>syncftp serve</code> ve proje ayarlarındaki HTTP API) yalnızca localhost'tan gelen istekleri kabul eder; başka sitelerin tarayıcıdan isteklerini reddeder. İsterseniz <code>--token</code> ile anahtar zorunlu olur. GitHub tetikleyicisi <code>SYNCFTP_GITHUB_SECRET</code> ister.</li>
      <li>Config, state ve freeze dosyaları atomik yazılır; aynı sunucuya sync / düzenleme / calibrate aynı anda çalışmaz (watch, panel, CI/CD ve API dahil).</li></ul>`],
    en: ['Security', `<ul>
      <li><i>Settings → Security</i>: move passwords to <b>Windows Credential Manager</b> — only a <code>keyring:</code> marker stays in <code>syncftp.json</code>.</li>
      <li>The panel only opens from this computer (127.0.0.1) with a per-session key; passwords are never sent to the page.</li>
      <li>The <code>.syncftp/</code> folder is added to <code>.gitignore</code> automatically.</li>
      <li>The <b>HTTP API</b> (<code>syncftp serve</code> and the HTTP API in project settings) only accepts requests from localhost and refuses cross-site browser requests. <code>--token</code> makes a key mandatory. The GitHub trigger requires <code>SYNCFTP_GITHUB_SECRET</code>.</li>
      <li>Config, state and freeze files are written atomically; syncs, edits and calibrations of the same server never overlap (watch, panel, CI/CD and API included).</li></ul>`],
  }],
  ['perf', 'gauge', {
    tr: ['Hız ve bağlantı limiti', `<ul>
      <li><b>Bağlantı limitini ölç</b> (kart menüsü): sunucunun aynı anda kabul ettiği en fazla bağlantıyı bulur ve kaydeder. Aynı hesaba bağlanan sunucular bu limiti paylaşır — "421 Too many connections" alınmaz.</li>
      <li><b>Hız sınırı</b> (KB/s): yüklemelerin interneti tıkamasını önler; sunucu başına toplamdır.</li>
      <li>Değişmeyen dosyalar yeniden okunmaz (önbellek) ve tarama paralel yapılır — büyük projelerde kontrol saniyeler sürer.</li>
      <li>Release listesi ve sayaçlar sunucu başına küçük bir özet dosyasından okunur; binlerce release'li projeler de anında açılır.</li>
      <li>Binlerce dosyalı ve yüklenmesi gerekmeyen klasörleri (<code>node_modules/</code>, <code>uploads/</code>) ignore'a ekleyin ya da freeze'leyin — calibrate ve listeleme o klasörlere girmez.</li></ul>`],
    en: ['Speed and connection limit', `<ul>
      <li><b>Measure connection limit</b> (card menu): finds and saves the most simultaneous connections the server accepts. Servers on the same account share it — no more "421 Too many connections".</li>
      <li><b>Speed limit</b> (KB/s) keeps uploads from saturating your connection; it is the total per server.</li>
      <li>Unchanged files are not re-read (cache) and scanning runs in parallel — checks take seconds even on large projects.</li>
      <li>Release lists and counters come from a small per-server summary file, so projects with thousands of releases open instantly.</li>
      <li>Add huge folders you never upload (<code>node_modules/</code>, <code>uploads/</code>) to ignore or freeze them — calibrate and listing skip them.</li></ul>`],
  }],
  ['tray', 'power', {
    tr: ['Tray ve açılışta başlatma', `<p>syncFTP sağ alttaki sistem tepsisinde çalışır: sol tık paneli açar, sağ tık menüsünde durum, projeler, aktif watch'lar, webhook tüneli ve çıkış vardır. <b>Windows açılışında başlat</b> seçeneği <i>Ayarlar → Genel</i>'dedir. Pencereyi kapatmak syncFTP'yi kapatmaz — watch'lar ve CI/CD dinleyicisi çalışmaya devam eder. Tamamen kapatmak için tray'den <i>Çık</i> ya da sol alttaki <i>Kapat</i>.</p>`],
    en: ['Tray and start with Windows', `<p>syncFTP lives in the system tray: left click opens the panel; the right-click menu shows status, projects, active watches, the webhook tunnel and quit. <b>Start with Windows</b> is under <i>Settings → General</i>. Closing the window does not quit syncFTP — watches and the CI/CD listener keep running. Quit from the tray or with <i>Quit</i> at the bottom left.</p>`],
  }],
  ['project', 'folder', {
    tr: ['Projeyi kaldırma ve veri konumları', `<ul>
      <li><b>Projeyi syncFTP'den çıkar</b>: <i>Proje ayarları → Tehlikeli bölge</i> (CLI: <code>syncftp project delete</code>). <code>.syncftp</code> klasörü (config, state, yedekler, loglar) <b>çöp kutusuna</b> taşınır — geri yüklerseniz proje geri gelir. Proje dosyalarına, sunuculara ve anahtarlıktaki şifrelere dokunulmaz. Çalışan iş varsa ya da başka bir süreç projeyi izliyorsa reddedilir.</li>
      <li><b>Veri konumları</b>: <i>Ayarlar → Sistem</i> syncFTP'nin diske yazdığı her yeri (ayar dosyaları, önbellek, geçici dosyalar, projelerin <code>.syncftp</code> klasörleri) kopyala / aç düğmeleriyle listeler.</li>
      <li>Sunucu, bağlantı ve proje ayarları formlarında kaydetmeden çıkmaya çalışırsanız sorulur.</li></ul>`],
    en: ['Removing a project and data locations', `<ul>
      <li><b>Remove project from syncFTP</b>: <i>Project settings → Danger zone</i> (CLI: <code>syncftp project delete</code>). The <code>.syncftp</code> folder (config, state, backups, logs) moves to the <b>trash</b> — restore it to bring the project back. Project files, servers and keyring passwords are not touched. Refused while a job runs or another process watches the project.</li>
      <li><b>Data locations</b>: <i>Settings → System</i> lists everything syncFTP writes (settings files, cache, temp files, each project's <code>.syncftp</code>) with copy / open buttons.</li>
      <li>Server, connection and project settings forms ask before you leave them with unsaved changes.</li></ul>`],
  }],
];

async function pageHelp(section) {
  const help = await GET('/api/help').catch(() => ({ commands: [], shell: '', global_flags: '' }));
  const L = LANG === 'en' ? 'en' : 'tr';
  bind((root) => {
    const q = $('#help-q', root);
    q.oninput = () => {
      const v = q.value.toLowerCase().trim();
      let any = false;
      $$('[data-hsec]', root).forEach((sec) => {
        const show = !v || sec.textContent.toLowerCase().includes(v);
        sec.hidden = !show;
        any = any || show;
      });
      $('#help-none', root).hidden = any;
    };
    if (section) { const el = $(`#h-${CSS.escape(section)}`, root); if (el) setTimeout(() => el.scrollIntoView({ block: 'start' }), 30); }
    $$('[data-hlink]', root).forEach((a) => (a.onclick = (e) => { e.preventDefault(); $(`#h-${a.dataset.hlink}`, root).scrollIntoView({ behavior: 'smooth', block: 'start' }); }));
  });
  const toc = [...HELP.map(([id, ic, c]) => [id, ic, c[L][0]]), ['cli', 'log', t('helpCli')]];
  return `<div class="sayfa"><div class="baslik-satir"><div><h1>${icon('log')}${t('helpTitle')}</h1><p>${t('helpSub')} · v${esc(help.version || '')}</p></div></div>
    <div class="arac-cubuk"><div class="ara">${icon('search')}<input type="text" id="help-q" placeholder="${t('helpSearch')}"></div></div>
    <div class="yardim">
      <nav class="yardim-toc"><div class="nav-label">${t('helpToc')}</div>${toc.map(([id, ic, title]) => `<a href="#" data-hlink="${id}">${icon(ic)}${esc(title)}</a>`).join('')}</nav>
      <div class="yardim-ic">
        ${HELP.map(([id, ic, c]) => `<section class="bolum yardim-bolum" id="h-${id}" data-hsec><h3>${icon(ic)}${c[L][0]}</h3><div class="yardim-metin">${c[L][1]}</div></section>`).join('')}
        <section class="bolum yardim-bolum" id="h-cli" data-hsec><h3>${icon('log')}${t('helpCli')}</h3>
          ${help.global_flags ? `<div class="yardim-cmd"><b>${t('helpGlobal')}</b><pre>${esc(help.global_flags)}</pre></div>` : ''}
          ${help.commands.map((c) => `<details class="yardim-cmd"><summary><code>${esc(c.path)}</code> <span class="muted">${esc(c.short)}</span></summary>
            <pre>${esc(c.use)}${c.long ? '\n\n' + esc(c.long) : ''}${c.flags ? '\n\n' + esc(c.flags) : ''}</pre></details>`).join('')}
          ${help.shell ? `<details class="yardim-cmd"><summary><b>${t('helpShell')}</b></summary><pre>${esc(help.shell.trim())}</pre></details>` : ''}
        </section>
        <div class="bos" id="help-none" hidden>${t('helpNoMatch')}</div>
      </div>
    </div></div>`;
}

'use strict';

/* ═══════════════════════════════════════════════════════════════════════════
   syncFTP panel — vanilla JS, bağımlılıksız.
   Router: #/ · #/jobs · #/p/<proje>[/<sekme>[/<alt>]]
   ═══════════════════════════════════════════════════════════════════════════ */

// ── i18n ─────────────────────────────────────────────────────────────────────
const T = {
  tr: {
    brandSub: 'Dağıtım paneli', navAll: 'Tüm projeler', navJobs: 'İşler', navProjects: 'Projeler',
    addProject: 'Proje ekle', navNoProjects: 'Bu klasörde proje yok', watching: 'Watch', language: 'Dil', quit: 'Kapat',
    loading: 'Yükleniyor…', moreRows: '{0} öğe daha göster', stFrozen: 'Freeze', stProtected: 'Korunan', stBlocked: 'Engelli', stUploadSize: 'yüklenecek: {0}', treeView: 'Ağaç', treeHint: 'Bu klasörün altını ağaç olarak göster (shell tree)', treeAsk: 'Klasör başına en fazla kaç öğe gösterilsin?', treeAll: 'Tümü', ignoreNoFile: 'ignore_files listesinde syncFTP dosyası yok — yalnızca .gitignore okunuyor', viewCode: 'Kod', viewRendered: 'Önizleme', viewRenderedLocal: 'Önizleme (yerel)', viewRenderedRemote: 'Önizleme (sunucu)', scanningLocal: 'Taranıyor…', refresh: 'Yenile', save: 'Kaydet', cancel: 'Vazgeç', close: 'Kapat', delete: 'Sil',
    edit: 'Düzenle', back: 'Geri', yes: 'Evet', confirm: 'Onayla', browse: 'Gözat', add: 'Ekle', open: 'Aç',
    saved: 'Kaydedildi', copied: 'Kopyalandı', never: 'hiç', none: 'yok', all: 'Tümü', manual: '(manuel giriş)',
    projectsTitle: 'Projeler', projectsSub: `{0} klasöründeki syncFTP projeleri (CLI gibi: panel açıldığı klasörü esas alır)`,
    noProjects: 'Henüz proje yok', noProjectsSub: 'Bu klasörde (ya da alt klasörlerinde) .syncftp/ yok. Buraya proje kurun; terminalde `syncftp init` çalıştırdığınız alt klasörler de burada kendiliğinden görünür.',
    serversN: '{0} sunucu', enabledN: '{0} aktif', configError: 'Config okunamadı',
    addProjectTitle: 'Proje ekle', folder: 'Proje klasörü', folderHint: 'Panelin açıldığı klasör ya da altındaki bir klasör (.syncftp/ yoksa oluşturulur)',
    displayName: 'Görünen ad', displayNameHint: 'Boş bırakılırsa klasör adı kullanılır',
    createConfigTitle: 'Config oluşturulsun mu?', createConfigBody: 'Bu klasörde syncFTP config\'i yok. Boş bir config oluşturup projeyi ekleyeyim mi? Ardından sunucu ekleyebilirsiniz.',
    create: 'Oluştur',
    tabOverview: 'Genel bakış', tabServers: 'Sunucular', tabConnections: 'Bağlantılar', tabSettings: 'Proje ayarları',
    tabFiles: 'Dosyalar', tabHistory: 'Geçmiş',
    sumServers: 'Sunucular', sumPending: 'Bekleyen değişiklik', sumLastSync: 'Son sync', sumWatch: 'Watch',
    syncAll: 'Tümünü sync et', dryRunAll: 'Dry-run', more: 'Diğer',
    calibrateAll: 'Tümünü calibrate et', calibrateHint: 'Sunucudaki dosyaları boyutla eşleştir, state\'i kur (yükleme yok)',
    verifyAll: 'Tümünü doğrula', verifyHint: 'Sunucuda elle değiştirilen/silinen dosyaları bul',
    detectAll: 'Bağlantı limitlerini ölç', detectHint: 'Sunucunun kabul ettiği en fazla bağlantıyı bulup max_connections\'a yazar',
    watchStart: 'Watch başlat', watchStop: 'Watch durdur', watchOn: 'Açık', watchOff: 'Kapalı',
    watchSince: '{0} beri · {1} sync', watchStarted: 'Watch başladı', watchStopped: 'Watch durduruldu',
    noServers: 'Bu projede sunucu yok', noServersSub: 'İlk sunucuyu ekleyin — FTP, FTPS veya SFTP; isterseniz SSH tüneli üzerinden.',
    addServer: 'Sunucu ekle', disabled: 'pasif', busy: 'çalışıyor', tunnel: 'tünel',
    sync: 'Sync', dryRun: 'Dry-run', dryRunHint: 'Neyin yükleneceğini göster, yükleme yapma',
    fullSync: 'Tam sync', fullSyncHint: 'State\'i yok say, bütün dosyaları yükle', fullSyncConfirm: '{0} üzerindeki BÜTÜN dosyalar yeniden yüklenecek. Devam edilsin mi?',
    retryFailed: 'Başarısızları tekrar dene', retryFailedHint: 'Son sync\'te yüklenemeyen {0} dosya',
    calibrate: 'Calibrate', verify: 'Doğrula', verifyFix: 'Doğrula ve düzelt', verifyFixHint: 'Sapan dosyalar sonraki sync\'te yeniden yüklenir',
    detect: 'Bağlantı limitini ölç', milestone: 'Milestone', milestoneSync: 'Milestone sync', milestoneSyncHint: 'Sadece milestone tarihinden sonra değişenleri yükle',
    freeze: 'Freeze listesi', freezeHint: 'Hiç yüklenmeyecek dosyalar', rollback: 'Geri al (rollback)', rollbackHint: 'Önceki bir sync\'in öncesine dön',
    browseFiles: 'Sunucu dosyaları', firstSyncNote: 'İlk sync henüz yapılmadı — sync başlayınca calibrate otomatik çalışıp mevcut dosyaları eşleştirir.',
    pendingChip: '<b class="n">+{0}</b> <b class="c">~{1}</b> <b class="d">−{2}</b>', upToDate: 'Güncel',
    filesTracked: '{0} dosya', maxConn: '{0} bağlantı', speedLimit: '{0} KB/s', noBackup: 'yedek kapalı', backupOn: 'yedekli',
    failedN: '{0} başarısız', frozenN: '{0} frozen', rollbackN: '{0} geri alınabilir',
    statusTitle: '{0} — bekleyen değişiklikler', newFiles: 'Yeni', changedFiles: 'Değişen', deletedFiles: 'Yerelde silinen',
    deletedNote: 'Yerelde silinen dosyalar sunucudan silinmez, yalnızca raporlanır.', emptyList: 'Liste boş',
    milestoneTitle: '{0} — milestone', milestoneBody: 'Milestone, bu tarihten önce değişmiş dosyaları sync dışında bırakır (kalıcı filtre). "Milestone sync" yalnızca tarihten sonra değişenleri yükler.',
    milestoneCurrent: 'Mevcut: {0}', milestoneDate: 'Tarih', milestoneDateHint: 'boş = şimdi · bugün · dün · 3d · 5h · 14:30 · 20.07 · 20.07.2026 15:04 · 2026-07-20',
    milestoneSet: 'Ayarla', milestoneClear: 'Kaldır', milestoneSaved: 'Milestone: {0}', milestoneCleared: 'Milestone kaldırıldı',
    freezeTitle: '{0} — freeze listesi', freezeBody: 'Her satıra bir dosya yolu (yerel dizine göreceli). Bu dosyalar değişse bile yüklenmez.',
    freezeSaved: '{0} dosya frozen',
    serversTitle: 'Sunucular', colName: 'Ad', colProtocol: 'Protokol', colHost: 'Host', colRemote: 'Uzak dizin', colLocal: 'Yerel dizin', colEnabled: 'Aktif',
    duplicate: 'Kopyala', deleteServerConfirm: '"{0}" sunucusu config\'ten silinsin mi? (State ve yedekler diskte kalır.)',
    serverNew: 'Yeni sunucu', serverEdit: 'Sunucuyu düzenle',
    secGeneral: 'Genel', secConnection: 'Bağlantı', secConnectionDesc: 'Bir bağlantı profili seçin ya da bilgileri elle girin.',
    secTunnel: 'SSH tüneli', secTunnelDesc: 'Sunucuya doğrudan erişilemiyorsa bağlantı bu SSH sunucusu (jump host) üzerinden açılır. Boş = tünel yok.',
    secPaths: 'Dizinler', secPerf: 'Performans', secFilters: 'Filtreler', secFiltersDesc: 'Her satıra bir yol. Sunucu include\'u globali geçersiz kılar; exclude/protect globale eklenir.',
    secUpload: 'Yükleme', secHooks: 'Hook\'lar', secHooksDesc: 'Her satıra bir komut, proje dizininde sırayla çalışır. Global hook\'lardan sonra çalışırlar.',
    fName: 'Ad', fEnabled: 'Aktif', fEnabledHint: 'Pasif sunucular "tümünü sync et"e dahil olmaz',
    fConnection: 'Bağlantı profili', fProtocol: 'Protokol', fHost: 'Host', fPort: 'Port', fUser: 'Kullanıcı', fPassword: 'Şifre',
    fKeyFile: 'SSH anahtar dosyası', fKeyFileHint: 'SFTP için; şifreliyse şifre alanı passphrase olarak denenir',
    fPassive: 'Passive mode', fEpsv: 'EPSV kapalı', fEpsvHint: 'Bazı eski sunucular için', fNat: 'NAT düzeltmesi', fNatHint: 'Pasif IP yerine host adresine bağlan',
    fTls: 'Sertifika doğrulamasını atla', fTlsHint: 'FTPS — paylaşımlı hosting sertifikası host adıyla uyuşmuyorsa',
    fTunnelHost: 'Tünel host', fTunnelPort: 'Tünel port', fTunnelUser: 'Tünel kullanıcı', fTunnelUserHint: 'Boş = sunucu kullanıcısı', fTunnelPass: 'Tünel şifre', fTunnelKey: 'Tünel anahtar dosyası',
    fRemote: 'Uzak dizin', fRemoteHint: 'Örn. /public_html', fLocal: 'Yerel dizin', fLocalHint: 'Proje köküne göreceli; boş = proje varsayılanı ({0})',
    fMaxConn: 'Eş zamanlı bağlantı', fMaxConnHint: 'Paralel upload sayısı; "Ölç" sunucunun limitini bulur', measure: 'Ölç',
    fRetries: 'Tekrar deneme', fBandwidth: 'Hız sınırı (KB/s)', fBandwidthHint: '0 = global değer ({0})',
    fInclude: 'Include (yalnızca bunlar)', fExclude: 'Exclude (hariç)', fProtect: 'Protect (asla üzerine yazma)', fBlock: 'Engellenen uzantılar', fBlockHint: 'Örn. .zip .psd — boşluk veya satırla ayırın',
    fMinify: 'CSS/JS minify', fMinifyHint: 'Yorum ve boşlukları kaldırır', fObfuscate: 'JS obfuscate', fObfuscateHint: 'terser gerekir (npm i -g terser)',
    fNoBackup: 'Yedeği kapat', fNoBackupHint: 'Upload öncesi eski dosya yedeklenmez — bu sunucuda rollback yapılamaz',
    fDirect: 'Doğrudan yükle', fDirectHint: 'Atomik (geçici ad + rename) yükleme kapalı — rename desteklemeyen sunucular için',
    fWebhook: 'Webhook URL', fWebhookHint: 'Her sync sonrası özet POST edilir; boş = global',
    fPre: 'pre_sync', fPost: 'post_sync',
    testConn: 'Bağlantıyı test et', testOk: 'Bağlantı başarılı — {0} ms, {1} öğe', testing: 'Test ediliyor…',
    lockedByProfile: 'Bağlantı bilgileri "{0}" profilinden geliyor. Değiştirmek için Bağlantılar sekmesini kullanın.',
    connectionsTitle: 'Bağlantı profilleri', connectionsSub: 'Aynı FTP/SFTP hesabını kullanan sunucular bilgileri tek yerden paylaşır.',
    noConnections: 'Bağlantı profili yok', noConnectionsSub: 'Birden fazla sunucu aynı hesaba bağlanıyorsa bilgileri bir profilde toplayın.',
    addConnection: 'Profil ekle', connNew: 'Yeni bağlantı profili', connEdit: 'Profili düzenle', usedBy: 'Kullanan',
    connInUse: '"{0}" profilini kullanan sunucular var: {1}', deleteConnConfirm: '"{0}" profili silinsin mi?',
    settingsTitle: 'Proje ayarları', secProject: 'Proje', fProjectName: 'Proje adı (config)', fDefaultPath: 'Varsayılan yerel dizin', fDefaultPathHint: 'Kendi yerel dizini olmayan sunucular buradan okur',
    secGlobal: 'Global sync ayarları', secGlobalDesc: 'Tüm sunuculara uygulanır; sunucu ayarları bunları genişletir.',
    fIgnoreFiles: 'Ignore dosyaları', fIgnoreFilesHint: 'Boş = .gitignore + .syncftp/syncftp.ignore',
    fFirstFull: 'İlk sync\'te her şeyi yükle', fFirstFullHint: 'Kapalıyken ilk sync calibrate ile mevcut dosyaları eşleştirir',
    fBackupKeep: 'Tutulacak yedek', fBackupKeepHint: 'Son kaç sync\'in yedeği saklansın (rollback derinliği)',
    secSecurity: 'Güvenlik', keyringTitle: 'Şifreleri sistem anahtarlığında sakla', keyringHint: 'Windows Credential Manager / macOS Keychain. syncftp.json\'da yalnızca "keyring:" işareti kalır.',
    keyringOn: 'Şifreler anahtarlığa taşındı', keyringOff: 'Şifreler tekrar config\'te', secretErrors: 'Bazı şifreler anahtarlıktan okunamadı',
    secRegistry: 'Proje kaydı', registryDesc: 'Panelde görünen ad ve kayıt. Kayıttan çıkarmak dosyalara dokunmaz.',
    renameProject: 'Yeniden adlandır', removeProject: 'Listeden çıkar', removeProjectConfirm: '"{0}" panel listesinden çıkarılsın mı? Proje dosyaları silinmez.',
    filesTitle: 'Sunucu dosyaları', pickServer: 'Sunucu', filter: 'Bu klasörde filtrele…', colSize: 'Boyut', colDate: 'Tarih',
    emptyDir: 'Klasör boş', preview: 'Önizle', download: 'İndir', deleteRemoteConfirm: '"{0}" sunucudan KALICI olarak silinsin mi?', deleteDirConfirm: '"{0}" klasörü ve İÇİNDEKİ HER ŞEY sunucudan kalıcı olarak silinsin mi?',
    deleted: 'Silindi', binaryFile: 'Binary dosya — önizleme yok', truncated: '(ilk 256 KB)',
    historyTitle: 'Geçmiş', releases: 'Sync geçmişi (release)', releasesSub: 'Her sync, değiştirdiği dosyaların eski halini yedekler. Bir release\'i geri almak onu ve ondan sonrakileri geri alır.',
    colWhen: 'Zaman', colServer: 'Sunucu', colFiles: 'Dosya', colReplaced: 'Değişen', colNew: 'Yeni', colStatus: 'Durum',
    rolledBack: 'geri alındı', noBackupTag: 'yedek yok', restorable: 'geri alınabilir', details: 'Ayrıntı', undo: 'Geri al',
    noReleases: 'Henüz release yok', rollbackTitle: 'Geri alma — {0}',
    rollbackBody: 'Aşağıdaki {0} release EN YENİDEN başlayarak geri alınacak: değişen dosyaların eski hali geri yüklenir, yeni eklenen dosyalar sunucudan silinir.',
    rollbackBlocked: 'Bu noktaya dönülemez: aradaki "{0}" release\'inin yedeği yok.', rollbackGo: 'Geri al',
    logs: 'Loglar', noLogs: 'Log yok', releaseTitle: 'Release {0}',
    jobsTitle: 'İşler', jobsSub: 'Bu oturumda çalışan ve tamamlanan işler', noJobs: 'Henüz iş yok', activeWatches: 'Aktif watch\'lar',
    colJob: 'İş', colProject: 'Proje', colStarted: 'Başladı', colDuration: 'Süre',
    stRunning: 'çalışıyor', stDone: 'tamam', stError: 'hata', stop: 'Durdur',
    jobDone: '{0} tamamlandı', jobFailed: '{0}: {1}', jobStarted: '{0} başladı', jobLog: 'İş günlüğü', resultLine: '↑ {0} yüklendi · {1} hata · {2}',
    quitConfirm: 'syncFTP tamamen kapatılsın mı? Çalışan watch\'lar ve webhook dinleyicisi de durur. (Pencereyi kapatırsanız tray\'de çalışmaya devam eder.)', quitDone: 'Panel kapandı — bu pencereyi kapatabilirsiniz.',
    sessionLost: 'Panel bağlantısı koptu', ago: '{0} önce', justNow: 'az önce', min: 'dk', hour: 'sa', day: 'gün',
    copyPath: 'Yolu kopyala', notFound: 'Bulunamadı', minutes: '{0} dk', seconds: '{0} sn',
    compare: 'Yerelle karşılaştır', diffTitle: 'Fark', diffLocal: 'Yerel', diffRemote: 'Sunucu', diffSame: 'Yerel ve sunucu sürümü aynı',
    diffRemoteMissing: 'Sunucuda yok — sync bu dosyayı yeni olarak yükler',
    secRemote: 'SSH komutları (sunucuda)', secRemoteDesc: 'Yalnızca SFTP sunucularında: SSH ile uzak dizinde çalışır. pre hata verirse sync iptal olur; post deploy sonrası çalışır (migrate, cache, restart…).', fRemotePre: 'remote_pre_sync', fRemotePost: 'remote_post_sync',
    whTest: 'Test et', whTestOk: 'Webhook gönderildi: {0}', ignoreTitle: 'Ignore kuralları (.syncftp/syncftp.ignore)', ignoreDesc: 'Gitignore sözdizimi. .gitignore ile birleştirilir; ignore edilen klasörlere yerelde ve sunucuda hiç girilmez.',
    ignoreTemplate: 'Şablon ekle…', ignoreSaved: 'Ignore kuralları kaydedildi', gitignoreNote: '.gitignore ({0} satır) de uygulanıyor',
    treePick: 'Ağaçtan seç', failedTitle: '{0} — başarısız dosyalar', failedBody: 'Son sync\'te yüklenemeyen dosyalar. Tekrar denemek yalnızca bunları yükler.',
    compareServers: 'Sunucuları karşılaştır', compareServersHint: 'İki sunucunun uzak içeriğini karşılaştırır (sadece birinde olan / boyutu farklı)', compareA: 'Sunucu A', compareB: 'Sunucu B', compareGo: 'Karşılaştır',
    upload: 'Yükle', newFolder: 'Yeni klasör', folderName: 'Klasör adı', rename: 'Yeniden adlandır / taşı', newPath: 'Yeni yol (uzak dizine göreceli)',
    moveTo: 'Taşı', moveTarget: 'Hedef klasör (uzak dizine göreceli)', selectedN: '{0} seçili', clearSel: 'Seçimi temizle',
    freezeToggle: 'Freeze / unfreeze', frozenBadge: 'frozen', searchAll: 'Alt klasörlerde ara', searchTitle: '"{0}" araması', searchNone: 'Sonuç yok', searching: 'Aranıyor…', searchingN: '{0} sonuç · {1} klasör tarandı…', searchDoneN: '{0} sonuç · {1} klasör · {2} sn', searchTruncated: 'Çok fazla sonuç (ilk 2000) ya da 5 dakika sınırı doldu — aramayı daraltın.',
    uploaded: '{0} dosya yüklendi', created: 'Oluşturuldu', moved: 'Taşındı', frozenAdded: '{0} dosya freeze listesine eklendi', frozenRemoved: '{0} dosya freeze listesinden çıkarıldı',
    deleteManyConfirm: '{0} öğe sunucudan KALICI olarak silinsin mi? (klasörler içerikleriyle)', runCmd: 'Sunucuda komut çalıştır', runCmdHint: 'SSH ile uzak dizinde komut çalıştırır', diffLocalMissing: 'Yerelde yok', diffTooBig: 'Dosya karşılaştırma için çok büyük — tam görünüm gösteriliyor',
    diffHint: 'Sunucu → yerel: <span class="bad-t">−</span> sunucudan gidecek, <span class="ok-t">+</span> sync ile gelecek', foldN: '⋯ {0} aynı satır', clickToDiff: 'Farkı görmek için dosyaya tıklayın',
  },
  en: {
    brandSub: 'Deploy panel', navAll: 'All projects', navJobs: 'Jobs', navProjects: 'Projects',
    addProject: 'Add project', navNoProjects: 'No project in this folder', watching: 'Watch', language: 'Language', quit: 'Quit',
    loading: 'Loading…', moreRows: 'Show {0} more', stFrozen: 'Frozen', stProtected: 'Protected', stBlocked: 'Blocked', stUploadSize: 'to upload: {0}', treeView: 'Tree', treeHint: 'Show everything below this folder as a tree (shell tree)', treeAsk: 'Show at most how many entries per folder?', treeAll: 'All', ignoreNoFile: 'ignore_files has no syncFTP file — only .gitignore is read', viewCode: 'Code', viewRendered: 'Preview', viewRenderedLocal: 'Preview (local)', viewRenderedRemote: 'Preview (server)', scanningLocal: 'Scanning…', refresh: 'Refresh', save: 'Save', cancel: 'Cancel', close: 'Close', delete: 'Delete',
    edit: 'Edit', back: 'Back', yes: 'Yes', confirm: 'Confirm', browse: 'Browse', add: 'Add', open: 'Open',
    saved: 'Saved', copied: 'Copied', never: 'never', none: 'none', all: 'All', manual: '(manual entry)',
    projectsTitle: 'Projects', projectsSub: `syncFTP projects in {0} (like the CLI: the panel works on the folder it was opened in)`,
    noProjects: 'No projects yet', noProjectsSub: 'This folder (and its subfolders) has no .syncftp/. Set a project up here; subfolders where you run `syncftp init` show up automatically.',
    serversN: '{0} servers', enabledN: '{0} enabled', configError: 'Config could not be read',
    addProjectTitle: 'Add project', folder: 'Project folder', folderHint: 'The folder the panel was opened in, or one below it (.syncftp/ is created if missing)',
    displayName: 'Display name', displayNameHint: 'Defaults to the folder name',
    createConfigTitle: 'Create config?', createConfigBody: 'This folder has no syncFTP config. Create an empty one and add the project? You can add servers next.',
    create: 'Create',
    tabOverview: 'Overview', tabServers: 'Servers', tabConnections: 'Connections', tabSettings: 'Project settings',
    tabFiles: 'Files', tabHistory: 'History',
    sumServers: 'Servers', sumPending: 'Pending changes', sumLastSync: 'Last sync', sumWatch: 'Watch',
    syncAll: 'Sync all', dryRunAll: 'Dry-run', more: 'More',
    calibrateAll: 'Calibrate all', calibrateHint: 'Match server files by size and build state (no upload)',
    verifyAll: 'Verify all', verifyHint: 'Find files changed/deleted on the server by hand',
    detectAll: 'Measure connection limits', detectHint: 'Find the most connections the server accepts and save to max_connections',
    watchStart: 'Start watch', watchStop: 'Stop watch', watchOn: 'On', watchOff: 'Off',
    watchSince: 'since {0} · {1} syncs', watchStarted: 'Watch started', watchStopped: 'Watch stopped',
    noServers: 'No servers in this project', noServersSub: 'Add the first server — FTP, FTPS or SFTP, optionally through an SSH tunnel.',
    addServer: 'Add server', disabled: 'disabled', busy: 'running', tunnel: 'tunnel',
    sync: 'Sync', dryRun: 'Dry-run', dryRunHint: 'Show what would be uploaded, upload nothing',
    fullSync: 'Full sync', fullSyncHint: 'Ignore state, upload every file', fullSyncConfirm: 'EVERY file on {0} will be re-uploaded. Continue?',
    retryFailed: 'Retry failed', retryFailedHint: '{0} file(s) failed last time',
    calibrate: 'Calibrate', verify: 'Verify', verifyFix: 'Verify & fix', verifyFixHint: 'Drifted files are re-uploaded on the next sync',
    detect: 'Measure connection limit', milestone: 'Milestone', milestoneSync: 'Milestone sync', milestoneSyncHint: 'Upload only files changed after the milestone',
    freeze: 'Freeze list', freezeHint: 'Files that are never uploaded', rollback: 'Roll back', rollbackHint: 'Return to before an earlier sync',
    browseFiles: 'Server files', firstSyncNote: 'No first sync yet — when a sync starts, calibrate runs automatically and matches existing files.',
    pendingChip: '<b class="n">+{0}</b> <b class="c">~{1}</b> <b class="d">−{2}</b>', upToDate: 'Up to date',
    filesTracked: '{0} files', maxConn: '{0} connections', speedLimit: '{0} KB/s', noBackup: 'no backup', backupOn: 'backed up',
    failedN: '{0} failed', frozenN: '{0} frozen', rollbackN: '{0} restorable',
    statusTitle: '{0} — pending changes', newFiles: 'New', changedFiles: 'Changed', deletedFiles: 'Deleted locally',
    deletedNote: 'Locally deleted files are not deleted from the server, only reported.', emptyList: 'Empty',
    milestoneTitle: '{0} — milestone', milestoneBody: 'A milestone keeps files changed before its date out of syncs (persistent filter). "Milestone sync" uploads only files changed after it.',
    milestoneCurrent: 'Current: {0}', milestoneDate: 'Date', milestoneDateHint: 'empty = now · today · yesterday · 3d · 5h · 14:30 · 20.07 · 20.07.2026 15:04 · 2026-07-20',
    milestoneSet: 'Set', milestoneClear: 'Clear', milestoneSaved: 'Milestone: {0}', milestoneCleared: 'Milestone cleared',
    freezeTitle: '{0} — freeze list', freezeBody: 'One file path per line (relative to the local folder). These files are never uploaded, even when changed.',
    freezeSaved: '{0} file(s) frozen',
    serversTitle: 'Servers', colName: 'Name', colProtocol: 'Protocol', colHost: 'Host', colRemote: 'Remote path', colLocal: 'Local path', colEnabled: 'Enabled',
    duplicate: 'Duplicate', deleteServerConfirm: 'Remove server "{0}" from the config? (State and backups stay on disk.)',
    serverNew: 'New server', serverEdit: 'Edit server',
    secGeneral: 'General', secConnection: 'Connection', secConnectionDesc: 'Pick a connection profile or enter the details by hand.',
    secTunnel: 'SSH tunnel', secTunnelDesc: 'If the server is not directly reachable, connect through this SSH server (jump host). Empty = no tunnel.',
    secPaths: 'Folders', secPerf: 'Performance', secFilters: 'Filters', secFiltersDesc: 'One path per line. Server include overrides global; exclude/protect add to global.',
    secUpload: 'Upload', secHooks: 'Hooks', secHooksDesc: 'One command per line, run in the project folder in order, after global hooks.',
    fName: 'Name', fEnabled: 'Enabled', fEnabledHint: 'Disabled servers are skipped by "sync all"',
    fConnection: 'Connection profile', fProtocol: 'Protocol', fHost: 'Host', fPort: 'Port', fUser: 'User', fPassword: 'Password',
    fKeyFile: 'SSH key file', fKeyFileHint: 'For SFTP; if encrypted, the password is tried as passphrase',
    fPassive: 'Passive mode', fEpsv: 'Disable EPSV', fEpsvHint: 'For some older servers', fNat: 'NAT workaround', fNatHint: 'Connect to the host address instead of the passive IP',
    fTls: 'Skip certificate check', fTlsHint: 'FTPS — when a shared hosting certificate does not match the host name',
    fTunnelHost: 'Tunnel host', fTunnelPort: 'Tunnel port', fTunnelUser: 'Tunnel user', fTunnelUserHint: 'Empty = server user', fTunnelPass: 'Tunnel password', fTunnelKey: 'Tunnel key file',
    fRemote: 'Remote path', fRemoteHint: 'e.g. /public_html', fLocal: 'Local path', fLocalHint: 'Relative to the project root; empty = project default ({0})',
    fMaxConn: 'Simultaneous connections', fMaxConnHint: 'Parallel uploads; "Measure" finds the server limit', measure: 'Measure',
    fRetries: 'Retries', fBandwidth: 'Speed limit (KB/s)', fBandwidthHint: '0 = global value ({0})',
    fInclude: 'Include (only these)', fExclude: 'Exclude', fProtect: 'Protect (never overwrite)', fBlock: 'Blocked extensions', fBlockHint: 'e.g. .zip .psd — separate by space or line',
    fMinify: 'Minify CSS/JS', fMinifyHint: 'Removes comments and whitespace', fObfuscate: 'Obfuscate JS', fObfuscateHint: 'Requires terser (npm i -g terser)',
    fNoBackup: 'Disable backup', fNoBackupHint: 'Old files are not backed up before upload — no rollback on this server',
    fDirect: 'Direct upload', fDirectHint: 'Atomic (temp name + rename) upload off — for servers without rename',
    fWebhook: 'Webhook URL', fWebhookHint: 'A summary is POSTed after each sync; empty = global',
    fPre: 'pre_sync', fPost: 'post_sync',
    testConn: 'Test connection', testOk: 'Connected — {0} ms, {1} entries', testing: 'Testing…',
    lockedByProfile: 'Connection details come from the "{0}" profile. Use the Connections tab to change them.',
    connectionsTitle: 'Connection profiles', connectionsSub: 'Servers that use the same FTP/SFTP account share the details from one place.',
    noConnections: 'No connection profiles', noConnectionsSub: 'If several servers connect to the same account, keep the details in a profile.',
    addConnection: 'Add profile', connNew: 'New connection profile', connEdit: 'Edit profile', usedBy: 'Used by',
    connInUse: 'Profile "{0}" is used by: {1}', deleteConnConfirm: 'Delete profile "{0}"?',
    settingsTitle: 'Project settings', secProject: 'Project', fProjectName: 'Project name (config)', fDefaultPath: 'Default local folder', fDefaultPathHint: 'Servers without their own local folder read from here',
    secGlobal: 'Global sync settings', secGlobalDesc: 'Apply to every server; server settings extend them.',
    fIgnoreFiles: 'Ignore files', fIgnoreFilesHint: 'Empty = .gitignore + .syncftp/syncftp.ignore',
    fFirstFull: 'Upload everything on first sync', fFirstFullHint: 'When off, the first sync matches existing files with calibrate',
    fBackupKeep: 'Backups to keep', fBackupKeepHint: 'How many recent syncs keep their backup (rollback depth)',
    secSecurity: 'Security', keyringTitle: 'Store passwords in the system keyring', keyringHint: 'Windows Credential Manager / macOS Keychain. Only a "keyring:" marker stays in syncftp.json.',
    keyringOn: 'Passwords moved to the keyring', keyringOff: 'Passwords are in the config again', secretErrors: 'Some passwords could not be read from the keyring',
    secRegistry: 'Project registry', registryDesc: 'The name shown in the panel. Removing from the list does not touch any files.',
    renameProject: 'Rename', removeProject: 'Remove from list', removeProjectConfirm: 'Remove "{0}" from the panel list? Project files are not deleted.',
    filesTitle: 'Server files', pickServer: 'Server', filter: 'Filter this folder…', colSize: 'Size', colDate: 'Date',
    emptyDir: 'Folder is empty', preview: 'Preview', download: 'Download', deleteRemoteConfirm: 'PERMANENTLY delete "{0}" from the server?', deleteDirConfirm: 'PERMANENTLY delete folder "{0}" and EVERYTHING inside it from the server?',
    deleted: 'Deleted', binaryFile: 'Binary file — no preview', truncated: '(first 256 KB)',
    historyTitle: 'History', releases: 'Sync history (releases)', releasesSub: 'Every sync backs up the previous version of the files it changes. Rolling back a release also rolls back every newer one.',
    colWhen: 'When', colServer: 'Server', colFiles: 'Files', colReplaced: 'Replaced', colNew: 'New', colStatus: 'Status',
    rolledBack: 'rolled back', noBackupTag: 'no backup', restorable: 'restorable', details: 'Details', undo: 'Roll back',
    noReleases: 'No releases yet', rollbackTitle: 'Roll back — {0}',
    rollbackBody: 'The following {0} release(s) will be undone, NEWEST first: replaced files get their previous version back, newly added files are deleted from the server.',
    rollbackBlocked: 'Cannot return here: release "{0}" in between has no backup.', rollbackGo: 'Roll back',
    logs: 'Logs', noLogs: 'No logs', releaseTitle: 'Release {0}',
    jobsTitle: 'Jobs', jobsSub: 'Jobs running and finished in this session', noJobs: 'No jobs yet', activeWatches: 'Active watches',
    colJob: 'Job', colProject: 'Project', colStarted: 'Started', colDuration: 'Duration',
    stRunning: 'running', stDone: 'done', stError: 'error', stop: 'Stop',
    jobDone: '{0} finished', jobFailed: '{0}: {1}', jobStarted: '{0} started', jobLog: 'Job log', resultLine: '↑ {0} uploaded · {1} failed · {2}',
    quitConfirm: 'Quit syncFTP completely? Running watches and the webhook listener stop too. (Closing the window keeps it running in the tray.)', quitDone: 'Panel stopped — you can close this window.',
    sessionLost: 'Lost connection to the panel', ago: '{0} ago', justNow: 'just now', min: 'min', hour: 'h', day: 'd',
    copyPath: 'Copy path', notFound: 'Not found', minutes: '{0} min', seconds: '{0} s',
    compare: 'Compare with local', diffTitle: 'Diff', diffLocal: 'Local', diffRemote: 'Server', diffSame: 'Local and server versions are identical',
    diffRemoteMissing: 'Not on the server — sync uploads it as a new file',
    secRemote: 'SSH commands (on the server)', secRemoteDesc: 'SFTP servers only: run over SSH in the remote folder. If pre fails the sync is aborted; post runs after the deploy (migrate, cache, restart…).', fRemotePre: 'remote_pre_sync', fRemotePost: 'remote_post_sync',
    whTest: 'Test', whTestOk: 'Webhook sent: {0}', ignoreTitle: 'Ignore rules (.syncftp/syncftp.ignore)', ignoreDesc: 'Gitignore syntax. Merged with .gitignore; ignored folders are never walked locally or on the server.',
    ignoreTemplate: 'Add template…', ignoreSaved: 'Ignore rules saved', gitignoreNote: '.gitignore ({0} lines) is applied too',
    treePick: 'Pick from tree', failedTitle: '{0} — failed files', failedBody: 'Files that could not be uploaded in the last sync. Retrying uploads only these.',
    compareServers: 'Compare servers', compareServersHint: 'Compare the remote contents of two servers (only on one side / size differs)', compareA: 'Server A', compareB: 'Server B', compareGo: 'Compare',
    upload: 'Upload', newFolder: 'New folder', folderName: 'Folder name', rename: 'Rename / move', newPath: 'New path (relative to remote root)',
    moveTo: 'Move', moveTarget: 'Target folder (relative to remote root)', selectedN: '{0} selected', clearSel: 'Clear selection',
    freezeToggle: 'Freeze / unfreeze', frozenBadge: 'frozen', searchAll: 'Search subfolders', searchTitle: 'Search "{0}"', searchNone: 'No results', searching: 'Searching…', searchingN: '{0} results · {1} folders scanned…', searchDoneN: '{0} results · {1} folders · {2}s', searchTruncated: 'Too many results (first 2000) or the 5-minute limit was hit — narrow the search.',
    uploaded: '{0} file(s) uploaded', created: 'Created', moved: 'Moved', frozenAdded: '{0} file(s) added to the freeze list', frozenRemoved: '{0} file(s) removed from the freeze list',
    deleteManyConfirm: 'PERMANENTLY delete {0} item(s) from the server? (folders with their contents)', runCmd: 'Run command on server', runCmdHint: 'Runs a command over SSH in the remote folder', diffLocalMissing: 'Missing locally', diffTooBig: 'File too large to diff — showing full view',
    diffHint: 'Server → local: <span class="bad-t">−</span> leaves the server, <span class="ok-t">+</span> arrives with sync', foldN: '⋯ {0} unchanged lines', clickToDiff: 'Click a file to see the diff',
  },
};
let LANG = 'tr';
const t = (k, ...a) => (T[LANG][k] ?? T.en[k] ?? k).replace(/\{(\d)\}/g, (_, i) => a[i] ?? '');

// ── yardımcılar ──────────────────────────────────────────────────────────────
const $ = (s, el = document) => el.querySelector(s);
const $$ = (s, el = document) => [...el.querySelectorAll(s)];
const esc = (s) => String(s ?? '').replace(/[&<>"']/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
const icon = (n, cls = '') => `<svg class="i ${cls}"><use href="#i-${n}"/></svg>`;
const enc = encodeURIComponent;

function bytes(n) {
  n = Number(n) || 0;
  if (n >= 1 << 30) return (n / (1 << 30)).toFixed(1) + ' GB';
  if (n >= 1 << 20) return (n / (1 << 20)).toFixed(1) + ' MB';
  if (n >= 1 << 10) return (n / (1 << 10)).toFixed(0) + ' KB';
  return n + ' B';
}
function ago(d) {
  if (!d) return t('never');
  const s = (Date.now() - new Date(d).getTime()) / 1000;
  if (s < 50) return t('justNow');
  if (s < 3600) return t('ago', Math.round(s / 60) + ' ' + t('min'));
  if (s < 86400) return t('ago', Math.round(s / 3600) + ' ' + t('hour'));
  return t('ago', Math.round(s / 86400) + ' ' + t('day'));
}
function fmtDate(d) {
  if (!d) return '';
  const x = new Date(d);
  if (isNaN(x)) return String(d);
  const p = (n) => String(n).padStart(2, '0');
  return `${x.getFullYear()}-${p(x.getMonth() + 1)}-${p(x.getDate())} ${p(x.getHours())}:${p(x.getMinutes())}`;
}
function dur(ms) {
  if (ms == null) return '';
  const s = Math.round(ms / 1000);
  return s >= 60 ? t('minutes', Math.floor(s / 60)) + ' ' + (s % 60) + 's' : t('seconds', s);
}
const lines = (v) => (v || []).join('\n');
const unlines = (s) => String(s || '').split(/\r?\n/).map((x) => x.trim()).filter(Boolean);
const unwords = (s) => String(s || '').split(/[\s,]+/).map((x) => x.trim()).filter(Boolean);

async function api(method, url, body) {
  const opt = { method, headers: {} };
  if (body !== undefined) { opt.headers['Content-Type'] = 'application/json'; opt.body = JSON.stringify(body); }
  let r;
  try { r = await fetch(url, opt); } catch (e) { throw new Error(t('sessionLost')); }
  let j = null;
  try { j = await r.json(); } catch { /* boş */ }
  if (!r.ok || !j || j.ok === false) throw new Error((j && j.error) || `HTTP ${r.status}`);
  return j.data;
}
const GET = (u) => api('GET', u);

// Config — Go tarafında omitempty olan alanlar (connections, …) yoksa boş değerlerle doldurulur.
async function getConfig(P) {
  const c = await GET(`/api/p/${enc(P)}/config`);
  c.project ??= {}; c.sync ??= {}; c.first_sync ??= {};
  c.servers ??= []; c.connections ??= [];
  return c;
}

function toast(msg, kind = 'ok') {
  const el = document.createElement('div');
  el.className = 'toast ' + kind;
  el.innerHTML = icon(kind === 'bad' ? 'alert' : 'check') + `<span>${esc(msg)}</span>`;
  $('#toastlar').appendChild(el);
  setTimeout(() => el.remove(), kind === 'bad' ? 7000 : 3500);
}
const fail = (e) => toast(e.message || String(e), 'bad');

// ── modal ────────────────────────────────────────────────────────────────────
function modal({ title, body, size = '', actions = [], onClose }) {
  const root = document.createElement('div');
  root.className = 'modal-arka';
  root.innerHTML = `<div class="modal ${size}" role="dialog"><div class="modal-bas"><h3>${title}</h3>
    <button class="ikon-btn" data-x>${icon('x')}</button></div><div class="modal-ic">${body}</div>
    ${actions.length ? '<div class="modal-alt"></div>' : ''}</div>`;
  const close = () => {
    if (!S.modals.delete(close)) return; // iki kez kapanmasın
    root.remove(); document.removeEventListener('keydown', onKey); onClose && onClose();
  };
  S.modals.add(close);
  // Esc yalnızca en üstteki modalı kapatır
  const onKey = (e) => { if (e.key === 'Escape' && $('#modal-kok').lastElementChild === root) close(); };
  document.addEventListener('keydown', onKey);
  root.addEventListener('mousedown', (e) => { if (e.target === root) close(); });
  $('[data-x]', root).onclick = close;
  const alt = $('.modal-alt', root);
  for (const a of actions) {
    if (a.left) { const s = document.createElement('div'); s.className = 'sol'; s.innerHTML = a.left; alt.appendChild(s); continue; }
    const b = document.createElement('button');
    b.className = 'btn ' + (a.cls || '');
    b.innerHTML = a.label;
    b.onclick = async () => {
      if (!a.run) return close();
      b.disabled = true;
      try { if ((await a.run(root)) !== false) close(); } catch (e) { fail(e); } finally { b.disabled = false; }
    };
    alt.appendChild(b);
  }
  $('#modal-kok').appendChild(root);
  const f = $('input,textarea,select', root);
  if (f) setTimeout(() => f.focus(), 30);
  return { root, close };
}
// loadModal: veri beklerken boş tıklama hissi olmasın — önce yükleniyor modalı açılır.
// fetcher hata verirse modal kapanır ve hata gösterilir; kullanıcı beklerken kapattıysa null döner.
async function loadModal(title, size, fetcher) {
  const ph = modal({ title, size, body: `<div class="yukleniyor" style="padding:28px 0">${icon('refresh', 'spin')} ${t('loading')}</div>` });
  try {
    const d = await fetcher();
    if (!ph.root.isConnected) return null;
    // Yeni modal eklendikten SONRA (aynı görev içinde) kapatılır → arada boş kare çizilmez
    queueMicrotask(() => ph.close());
    return { d };
  } catch (e) {
    ph.close();
    fail(e);
    return null;
  }
}

function confirmBox(title, text, okLabel = t('confirm'), danger = false) {
  return new Promise((res) => {
    let done = false;
    modal({
      title, body: `<p>${text}</p>`,
      actions: [
        { label: t('cancel'), cls: 'hayalet', run: () => { done = true; res(false); } },
        { label: okLabel, cls: danger ? 'tehlike' : 'birincil', run: () => { done = true; res(true); } },
      ],
      onClose: () => { if (!done) res(false); },
    });
  });
}

// Açılır menü: items = [{label, hint, icon, run, danger}] | 'hr' | {title}
function openMenu(anchor, items) {
  closeMenus();
  const m = document.createElement('div');
  m.className = 'menu';
  for (const it of items) {
    if (!it) continue;
    if (it === 'hr') { m.appendChild(document.createElement('hr')); continue; }
    if (it.title) { const d = document.createElement('div'); d.className = 'menu-baslik'; d.textContent = it.title; m.appendChild(d); continue; }
    const b = document.createElement('button');
    if (it.danger) b.className = 'tehlike';
    b.innerHTML = `${icon(it.icon || 'play')}<span>${esc(it.label)}${it.hint ? `<small>${esc(it.hint)}</small>` : ''}</span>`;
    b.onclick = (e) => { e.stopPropagation(); closeMenus(); it.run(); };
    m.appendChild(b);
  }
  const kap = anchor.closest('.menu-kap') || anchor.parentElement;
  kap.appendChild(m);
  const r = m.getBoundingClientRect();
  if (r.bottom > window.innerHeight - 10) { m.style.top = 'auto'; m.style.bottom = 'calc(100% + 6px)'; }
}
function closeMenus() { $$('.menu').forEach((m) => m.remove()); }
document.addEventListener('click', (e) => { if (!e.target.closest('.menu') && !e.target.closest('[data-menu]')) closeMenus(); });

// ── durum ────────────────────────────────────────────────────────────────────
const S = { meta: null, projects: [], jobs: new Map(), renderSeq: 0, modals: new Set() };

// ── iş (job) yönetimi ───────────────────────────────────────────────────────
function trackJob(id, opts = {}) {
  if (S.jobs.has(id)) return S.jobs.get(id);
  const j = { id, title: opts.title || id, server: opts.server || '', project: opts.project || '', status: 'running', last: '', lines: [], es: null, shownUntil: 0 };
  S.jobs.set(id, j);
  const es = new EventSource(`/api/jobs/${enc(id)}/stream`);
  j.es = es;
  es.onmessage = (e) => {
    j.lines.push(e.data);
    if (e.data.trim()) j.last = e.data;
    if (j.onLine) j.onLine(e.data);
    renderDock();
  };
  es.addEventListener('done', (e) => {
    es.close();
    try {
      const snap = JSON.parse(e.data);
      Object.assign(j, { status: snap.status, error: snap.error, result: snap.result, title: snap.title, server: snap.server, project: snap.project, ended: snap.ended, started: snap.started });
    } catch { j.status = 'done'; }
    j.shownUntil = Date.now() + 9000;
    if (j.onDone) j.onDone(j);
    const name = `${j.title}${j.server ? ' · ' + j.server : ''}`;
    if (j.status === 'error') toast(t('jobFailed', name, j.error || ''), 'bad');
    else if (opts.quiet !== true) toast(t('jobDone', name) + resultSuffix(j.result));
    renderDock();
    onJobFinished(j);
  });
  es.onerror = () => { /* tarayıcı yeniden bağlanır; done gelmişse zaten kapalı */ };
  renderDock();
  return j;
}
function resultSuffix(r) {
  if (!r || r.uploaded == null) return '';
  return ' — ' + t('resultLine', r.uploaded, r.failed, bytes(r.bytes));
}
async function startJobs(project, req, label) {
  const res = await api('POST', `/api/p/${enc(project)}/jobs`, req);
  for (const e of res.errors || []) toast(e, 'bad');
  for (const id of res.jobs || []) trackJob(id, { project, title: label });
  setTimeout(refreshJobsMeta, 300);
  return res.jobs || [];
}
function renderDock() {
  const el = $('#isler');
  const now = Date.now();
  const visible = [...S.jobs.values()].filter((j) => j.status === 'running' || j.shownUntil > now).slice(-5);
  el.innerHTML = visible.map((j) => {
    const st = j.status === 'running' ? 'busy' : j.status === 'error' ? 'bad' : 'ok';
    const txt = j.status === 'running' ? t('stRunning') : j.status === 'error' ? t('stError') : t('stDone');
    return `<div class="is" data-job="${esc(j.id)}"><div class="ust"><span class="pill ${st}"><span class="nokta"></span>${txt}</span>
      <b>${esc(j.title)}${j.server ? ' · ' + esc(j.server) : ''}</b></div>
      <div class="son">${esc(j.status === 'error' ? j.error || '' : j.last || '…')}</div>
      ${j.status === 'running' ? '<div class="bar"><i></i></div>' : ''}</div>`;
  }).join('');
  $$('[data-job]', el).forEach((d) => (d.onclick = () => showJobLog(d.dataset.job)));
  const running = [...S.jobs.values()].filter((j) => j.status === 'running' && j.title !== 'Watch').length;
  $('#nav-jobs').textContent = running ? running : '';
}
setInterval(renderDock, 3000);

function logClass(l) {
  if (/══|^\s*\[.*\]\s*$|──/.test(l)) return 'l-head';
  // Özet satırları: "… 0 hata / 0 errors / 0 failed" → başarı
  if (/(tamam|done)/i.test(l) && /\b0 (hata|errors?|failed)/i.test(l)) return 'l-ok';
  if (/✗|! |hata|error|failed|başarısız|UYUŞMUYOR|iptal/i.test(l)) return 'l-bad';
  if (/✓|tamam|done|geri yüklendi|restored|eşleşti|matched/i.test(l)) return 'l-ok';
  if (/↺|❄|⏱|milestone|frozen|yedek|backup/i.test(l)) return 'l-vio';
  if (/protected|korunan|skip|atla|dry/i.test(l)) return 'l-warn';
  return '';
}
const logLine = (l) => `<div class="${logClass(l)}">${esc(l) || '&nbsp;'}</div>`;

async function showJobLog(id) {
  let snap;
  try { snap = await GET(`/api/jobs/${enc(id)}`); } catch (e) { return fail(e); }
  // Satırlar tek kaynaktan gelir: iş takip ediliyorsa SSE ile toplanan j.lines (backlog dahil);
  // yeni takip başlıyorsa backlog onLine üzerinden akar → arada satır kaybolmaz.
  let j = S.jobs.get(id);
  const fresh = !j;
  if (fresh) j = trackJob(id, { title: snap.title, server: snap.server, project: snap.project, quiet: true });
  const header = () => {
    const st = j.status === 'running' ? 'busy' : j.status === 'error' ? 'bad' : 'ok';
    const txt = j.status === 'running' ? t('stRunning') : j.status === 'error' ? t('stError') : t('stDone');
    return `<span class="pill ${st}"><span class="nokta"></span>${txt}</span> ${esc(snap.title)}${snap.server ? ' · ' + esc(snap.server) : ''}`;
  };
  const m = modal({
    title: header(), size: 'genis',
    body: `<div class="log" id="joblog">${(fresh ? [] : j.lines).map(logLine).join('')}</div>`,
    actions: [
      { left: `<span class="mono">${esc(snap.project || '')} · ${fmtDate(snap.started)}</span>` },
      ...(snap.kind === 'watch' && j.status === 'running' ? [{ label: icon('stop') + t('stop'), cls: 'tehlike', run: async () => { await api('DELETE', `/api/jobs/${enc(id)}`); toast(t('watchStopped')); } }] : []),
      { label: t('close'), cls: 'birincil' },
    ],
    onClose: () => { j.onLine = null; j.onDone = null; },
  });
  const box = $('#joblog', m.root);
  box.scrollTop = box.scrollHeight;
  j.onLine = (l) => {
    const stick = box.scrollTop + box.clientHeight >= box.scrollHeight - 30;
    box.insertAdjacentHTML('beforeend', logLine(l));
    if (stick) box.scrollTop = box.scrollHeight;
  };
  j.onDone = () => { $('.modal-bas h3', m.root).innerHTML = header(); };
}

function onJobFinished(j) {
  // Açık sayfa bu projeye aitse veriyi tazele
  const r = parseRoute();
  if (r.project && r.project === j.project && ['overview', 'history', 'servers', 'cicd', 'files'].includes(r.tab)) render(true);
  if (r.page === 'jobs') render(true);
}

async function refreshJobsMeta() {
  try {
    const list = await GET('/api/jobs');
    for (const x of list) if (x.status === 'running' && !S.jobs.has(x.id)) trackJob(x.id, { title: x.title, server: x.server, project: x.project, quiet: true });
  } catch { /* yoksay */ }
}

// ── router ───────────────────────────────────────────────────────────────────
function parseRoute() {
  const parts = location.hash.replace(/^#\/?/, '').split('/').map(decodeURIComponent);
  if (parts[0] === 'p' && parts[1]) return { page: 'project', project: parts[1], tab: parts[2] || 'overview', sub: parts[3] ?? null, sub2: parts[4] ?? null };
  if (parts[0] === 'jobs') return { page: 'jobs' };
  if (parts[0] === 'help') return { page: 'help', sub: parts[1] || '' };
  if (parts[0] === 'settings') return { page: 'settings' };
  if (parts[0] === 'terminal') return { page: 'terminal', sub: parts.length > 1 ? parts[1] : null };
  return { page: 'home' };
}
const go = (h) => { if (location.hash === h) render(); else location.hash = h; };
const pHash = (p, tab, sub) => `#/p/${enc(p)}${tab ? '/' + tab : ''}${sub != null ? '/' + enc(sub) : ''}`;
window.addEventListener('hashchange', () => render());

async function render(soft = false) {
  const seq = ++S.renderSeq;
  const r = parseRoute();
  if (r.project) S.lastProject = r.project;
  document.body.classList.toggle('terminal-modu', r.page === 'terminal');
  $$('.nav a').forEach((a) => a.classList.toggle('aktif',
    !!((a.dataset.nav === 'home' && r.page === 'home') || (a.dataset.nav === 'jobs' && r.page === 'jobs') || (a.dataset.nav === 'help' && r.page === 'help') || (a.dataset.nav === 'settings' && r.page === 'settings') || (a.dataset.nav === 'terminal' && r.page === 'terminal') || (a.dataset.proj && a.dataset.proj === r.project))));
  const root = $('#sayfa');
  if (!soft) {
    [...S.modals].forEach((c) => c());
    const tabsEl = r.page === 'project' && root.dataset.proj === r.project ? $('.sayfa > .tabs', root) : null;
    if (tabsEl) {
      const activeTab = r.tab === 'server' ? 'servers' : r.tab === 'connection' ? 'connections' : r.tab;
      $$('a', tabsEl).forEach((a) => a.classList.toggle('aktif', a.getAttribute('href') === pHash(r.project, activeTab)));
      while (tabsEl.nextSibling) tabsEl.nextSibling.remove();
      tabsEl.insertAdjacentHTML('afterend', `<div class="yukleniyor">${icon('refresh', 'spin')} ${t('loading')}</div>`);
      const aks = $('#proj-aks', root);
      if (aks) aks.innerHTML = '';
    } else root.innerHTML = `<div class="yukleniyor">${icon('refresh', 'spin')} ${t('loading')}</div>`;
  }
  try {
    let html;
    if (r.page === 'home') html = await pageHome();
    else if (r.page === 'jobs') html = await pageJobs();
    else if (r.page === 'help') html = await pageHelp(r.sub);
    else if (r.page === 'settings') html = await pageSettings();
    else if (r.page === 'terminal') html = await pageTerminal(r.sub);
    else html = await pageProject(r);
    if (seq !== S.renderSeq) return;
    if (typeof html === 'string') root.innerHTML = html;
    root.dataset.proj = r.page === 'project' ? r.project : '';
    if (root._bind) { const b = root._bind; root._bind = null; b(root); }
  } catch (e) {
    if (seq !== S.renderSeq) return;
    root.innerHTML = `<div class="bos"><div class="buyuk">${icon('alert')}</div><h3>${esc(e.message)}</h3></div>`;
  }
}
const bind = (fn) => { $('#sayfa')._bind = fn; };

// ── sidebar ──────────────────────────────────────────────────────────────────
async function loadProjects() {
  try { S.projects = await GET('/api/projects'); } catch (e) { S.projects = []; }
  const el = $('#proje-listesi');
  const r = parseRoute();
  el.innerHTML = S.projects.map((p) => `<a href="${pHash(p.name)}" data-proj="${esc(p.name)}" class="${r.project === p.name ? 'aktif' : ''}">
      ${p.watching ? `<span class="busy" style="display:flex"><span class="nokta"></span></span>` : icon(p.ok ? 'folder' : 'alert', p.ok ? '' : 'bad-t')}
      <span class="ad">${esc(p.name)}</span><span class="sag">${p.ok ? p.enabled : ''}</span></a>`).join('')
    || `<div class="nav-bos">${icon('folder')}<span data-t="navNoProjects">${t('navNoProjects')}</span>
        <button class="btn kucuk" id="nav-bos-ekle">${icon('plus')}<span data-t="addProject">${t('addProject')}</span></button></div>`;
  const nb = $('#nav-bos-ekle');
  if (nb) nb.onclick = addProjectFlow;
  $('#sb-watch').textContent = S.projects.filter((p) => p.watching).length;
}

function applyLang() {
  document.documentElement.lang = LANG;
  $$('[data-t]').forEach((e) => (e.textContent = t(e.dataset.t)));
  $$('[data-tt]').forEach((e) => (e.title = t(e.dataset.tt)));
  $$('#dil-sec button').forEach((b) => b.classList.toggle('aktif', b.dataset.l === LANG));
}

// ═══ Sayfa: projeler ══════════════════════════════════════════════════════════
async function pageHome() {
  await loadProjects();
  const ps = S.projects;
  bind((root) => {
    $$('[data-add]', root).forEach((b) => (b.onclick = addProjectFlow));
    $$('.pkart', root).forEach((c) => (c.onclick = () => go(pHash(c.dataset.p))));
  });
  if (!ps.length) {
    return `<div class="sayfa"><div class="baslik-satir"><div><h1>${t('projectsTitle')}</h1><p>${esc(t('projectsSub', (S.meta && S.meta.base) || ''))}</p></div></div>
      <div class="bos"><div class="buyuk">${icon('folder', 'lg')}</div><h3>${t('noProjects')}</h3><p>${esc(t('noProjectsSub'))}</p>
      <div class="aksiyonlar" style="justify-content:center;flex-wrap:wrap"><button class="btn birincil" data-add>${icon('plus')}${t('addProject')}</button>
        <button class="btn ai-btn" onclick="aiSetupStart()">${icon('spark')}${t('aiSetupNav')}</button>
        <a class="btn" href="#/settings">${icon('sliders')}${t('navSettings')}</a></div></div></div>`;
  }
  return `<div class="sayfa"><div class="baslik-satir"><div><h1>${t('projectsTitle')}</h1><p>${esc(t('projectsSub', (S.meta && S.meta.base) || ''))}</p></div>
    <div class="aksiyonlar"><button class="btn birincil" data-add>${icon('plus')}${t('addProject')}</button></div></div>
    <div class="izgara">${ps.map((p) => `<div class="pkart" data-p="${esc(p.name)}">
      <h3>${icon(p.ok ? 'folder' : 'alert', p.ok ? '' : 'bad-t')}${esc(p.name)}${p.watching ? `<span class="pill busy"><span class="nokta"></span>watch</span>` : ''}</h3>
      <div class="yol" title="${esc(p.path)}">${esc(p.path)}</div>
      <div class="alt">${p.ok ? `<span>${t('serversN', p.servers)}</span><span>·</span><span>${t('enabledN', p.enabled)}</span>${p.keyring ? `<span>·</span><span class="acc-t">${icon('lock')}</span>` : ''}` : `<span class="bad-t">${t('configError')}</span>`}
      <span style="margin-left:auto" class="faint">${p.last_used ? ago(p.last_used) : ''}</span></div></div>`).join('')}</div></div>`;
}

async function addProjectFlow() {
  const lib = await GET('/api/library').catch(() => ({ connections: [], defaults: {}, templates: [] }));
  const defT = (lib.defaults || {}).ignore_template || '';
  const tmplOpts = [['', t('apTemplateDefault', defT ? (TEMPLATE_LABEL[defT] || defT) : t('defTemplateNone'))], ['none', t('defTemplateNone')],
    ...(lib.templates || []).map((k) => [k, TEMPLATE_LABEL[k] || k])];
  const conns = lib.connections || [];
  const m = modal({
    title: t('addProjectTitle'), size: 'orta',
    body: `<div class="form"><div class="alan"><label>${t('folder')}</label><div class="girdi-grup"><input type="text" id="ap-path" placeholder="D:\\projeler\\site" spellcheck="false" value="${esc((S.meta && S.meta.base) || '')}">
      <button class="btn" id="ap-browse">${icon('folder')}${t('browse')}</button></div><small>${t('folderHint')}</small></div>
      <div class="alan"><label>${t('displayName')}</label><input type="text" id="ap-name"><small>${t('displayNameHint')}</small></div>
      <div class="alanlar"><div class="alan"><label>${t('apTemplate')}</label><select id="ap-tmpl">${tmplOpts.map(([v, l]) => `<option value="${esc(v)}">${esc(l)}</option>`).join('')}</select></div>
        <div class="alan"><label>${t('apFirstServer')}</label><select id="ap-conn"><option value="">${t('apNoServer')}</option>${conns.map((c) => `<option value="${esc(c.name)}">${esc(c.name)} — ${esc(connURL(c))}</option>`).join('')}</select></div></div>
      <div class="alan gizli" id="ap-remote-box"><label>${t('apRemote')}</label><input type="text" id="ap-remote" class="mono" value="/public_html" spellcheck="false"></div></div>`,
    actions: [
      { label: icon('spark') + t('aiSetupNav'), cls: 'hayalet ai-btn', run: (root) => { aiSetupStart($('#ap-path', root).value.trim()); } },
      { label: t('cancel'), cls: 'hayalet' },
      { label: icon('plus') + t('add'), cls: 'birincil', run: async (root) => {
        const path = $('#ap-path', root).value.trim(), name = $('#ap-name', root).value.trim();
        let res = await api('POST', '/api/projects', { path, name });
        if (res && res.needs_create) {
          if (!(await confirmBox(t('createConfigTitle'), esc(t('createConfigBody')), t('create')))) return false;
          const connection = $('#ap-conn', root).value;
          res = await api('POST', '/api/projects', { path, name, create: true, template: $('#ap-tmpl', root).value, connection, remote: $('#ap-remote', root).value.trim() });
          await loadProjects();
          go(connection ? pHash(res.name) : pHash(res.name, 'server', 'new'));
          return;
        }
        await loadProjects();
        go(pHash(res.name));
      } },
    ],
  });
  const cs = $('#ap-conn', m.root);
  cs.onchange = () => $('#ap-remote-box', m.root).classList.toggle('gizli', !cs.value);
  $('#ap-browse', m.root).onclick = async () => {
    const picked = await pickPath({ mode: 'dir', initial: $('#ap-path', m.root).value.trim() });
    if (picked) $('#ap-path', m.root).value = picked;
  };
}

// ═══ Sayfa: proje ════════════════════════════════════════════════════════════
async function pageProject(r) {
  if (!S.projects.length) await loadProjects();
  const proj = S.projects.find((p) => p.name === r.project);
  if (!proj) throw new Error(t('notFound') + ': ' + r.project);
  const tabs = [
    ['overview', 'layers', t('tabOverview')], ['servers', 'server', t('tabServers')], ['connections', 'link', t('tabConnections')],
    ['files', 'folder', t('tabFiles')], ['history', 'history', t('tabHistory')], ['cicd', 'zap', t('tabCicd')], ['settings', 'sliders', t('tabSettings')],
  ];
  const activeTab = r.tab === 'server' ? 'servers' : r.tab === 'connection' ? 'connections' : r.tab;
  const head = `<div class="baslik-satir"><div><h1>${esc(proj.name)}</h1>
      <div class="yol">${icon('folder')}<span>${esc(proj.path)}</span><button class="ikon-btn" style="width:24px;height:24px" data-copy="${esc(proj.path)}" title="${t('copyPath')}">${icon('copy')}</button></div></div>
      <div class="aksiyonlar"><span id="proj-aks" class="aksiyonlar"></span>
        <button class="btn kucuk ai-btn" onclick="aiToggle(true)">${icon('spark')}${t('aiNav')}</button>
        <a class="ikon-btn" href="#/help/${({ overview: 'sync', servers: 'start', server: 'start', connections: 'ssh', connection: 'ssh', files: 'filters', history: 'rollback', cicd: 'cicd', settings: 'filters' })[r.tab] || 'start'}" title="${t('navHelp')}">${icon('help')}</a></div></div>
    <div class="tabs">${tabs.map(([k, ic, l]) => `<a href="${pHash(proj.name, k)}" class="${activeTab === k ? 'aktif' : ''}">${icon(ic)}${l}</a>`).join('')}</div>`;
  let body;
  switch (r.tab) {
    case 'servers': body = await tabServers(proj); break;
    case 'server': body = await tabServerEdit(proj, r.sub); break;
    case 'connections': body = await tabConnections(proj); break;
    case 'connection': body = await tabConnEdit(proj, r.sub); break;
    case 'settings': body = await tabSettings(proj); break;
    case 'files': body = await tabFiles(proj, r.sub, r.sub2); break;
    case 'history': body = await tabHistory(proj, r.sub); break;
    case 'cicd': body = await tabCicd(proj); break;
    default: body = await tabOverview(proj);
  }
  const inner = $('#sayfa')._bind;
  bind((root) => {
    $$('[data-copy]', root).forEach((b) => (b.onclick = () => { navigator.clipboard.writeText(b.dataset.copy); toast(t('copied')); }));
    inner && inner(root);
  });
  return `<div class="sayfa">${head}${body}</div>`;
}

// ── Genel bakış ──────────────────────────────────────────────────────────────
async function tabOverview(proj) {
  const ov = await GET(`/api/p/${enc(proj.name)}/overview?scan=0`);
  const servers = ov.servers || [];
  servers.forEach((s) => { s.scanning = s.enabled; });
  const enabled = servers.filter((s) => s.enabled);
  const pending = enabled.reduce((a, s) => a + s.new + s.changed, 0);
  const seq = S.renderSeq;
  const last = servers.map((s) => s.last_sync).filter(Boolean).sort().pop();
  const w = ov.watch;

  bind((root) => {
    const P = proj.name;
    const aks = $('#proj-aks', root);
    aks.innerHTML = `<button class="btn ${w ? 'tehlike' : ''}" id="b-watch">${icon(w ? 'stop' : 'eye')}${w ? t('watchStop') : t('watchStart')}</button>
      <button class="ikon-btn" id="b-refresh" title="${t('refresh')}">${icon('refresh')}</button>`;
    $('#b-refresh', root).onclick = () => render();
    $('#b-watch', root).onclick = async () => {
      try {
        if (w) { await api('DELETE', `/api/p/${enc(P)}/watch`); toast(t('watchStopped')); }
        else { const gw = await api('POST', `/api/p/${enc(P)}/watch`, {}); trackJob(gw.job_id, { project: P, title: 'Watch', quiet: true }); toast(t('watchStarted')); }
        await loadProjects(); render(true);
      } catch (e) { fail(e); }
    };
    const sa = $('#b-syncall', root);
    if (sa) sa.onclick = () => startJobs(P, { kind: 'sync' }, t('sync')).catch(fail);
    const dr = $('#b-dryall', root);
    if (dr) dr.onclick = () => startJobs(P, { kind: 'sync', dry_run: true }, t('dryRun')).catch(fail);
    const mo = $('#b-moreall', root);
    if (mo) mo.onclick = () => openMenu(mo, [
      { icon: 'scan', label: t('calibrateAll'), hint: t('calibrateHint'), run: () => startJobs(P, { kind: 'calibrate' }, t('calibrate')).catch(fail) },
      { icon: 'shield', label: t('verifyAll'), hint: t('verifyHint'), run: () => startJobs(P, { kind: 'verify' }, t('verify')).catch(fail) },
      { icon: 'gauge', label: t('detectAll'), hint: t('detectHint'), run: () => startJobs(P, { kind: 'detect' }, t('detect')).catch(fail) },
      servers.length > 1 ? 'hr' : null,
      servers.length > 1 ? { icon: 'layers', label: t('compareServers'), hint: t('compareServersHint'), run: () => compareServersModal(P, servers) } : null,
    ]);
    $$('[data-add-server]', root).forEach((b) => (b.onclick = () => go(pHash(P, 'server', 'new'))));
    $$('.kart[data-s]', root).forEach((card) => bindServerCard(card, P, servers.find((s) => s.name === card.dataset.s)));
    // Yerel tarama sunucu başına arka planda; biten kart yerinde güncellenir
    const sumPending = () => {
      const el = $('#ov-pending', root);
      if (!el) return;
      const busy = enabled.some((s) => s.scanning);
      const n = enabled.reduce((a, s) => a + (s.scan_error ? 0 : s.new + s.changed), 0);
      el.className = n ? 'warn-t' : '';
      el.innerHTML = busy && !n ? icon('refresh', 'spin') : String(n);
    };
    enabled.forEach((s) => {
      GET(`/api/p/${enc(P)}/scan?server=${enc(s.name)}`).catch((e) => ({ scan_error: e.message })).then((info) => {
        if (seq !== S.renderSeq) return;
        Object.assign(s, info, { scanning: false });
        const old = $(`.kart[data-s="${CSS.escape(s.name)}"]`, root);
        if (!old || !document.body.contains(old)) return;
        const tmp = document.createElement('div');
        tmp.innerHTML = serverCard(s);
        const card = tmp.firstElementChild;
        old.replaceWith(card);
        bindServerCard(card, P, s);
        sumPending();
      });
    });
    sumPending();
  });

  if (!servers.length) {
    return `<div class="bos"><div class="buyuk">${icon('server', 'lg')}</div><h3>${t('noServers')}</h3><p>${t('noServersSub')}</p>
      <button class="btn birincil" data-add-server>${icon('plus')}${t('addServer')}</button></div>`;
  }
  const secretWarn = (ov.secret_errors || []).length ? `<div class="kart" style="margin-bottom:12px"><div class="not bad">${icon('alert')}${t('secretErrors')}: ${esc(ov.secret_errors.join('; '))}</div></div>` : '';
  return `${secretWarn}<div class="ozet">
      <div class="kutu"><span>${icon('server')}${t('sumServers')}</span><b>${enabled.length}<small> / ${servers.length}</small></b></div>
      <div class="kutu"><span>${icon('upload')}${t('sumPending')}</span><b id="ov-pending" class="${pending ? 'warn-t' : ''}">${pending}</b></div>
      <div class="kutu"><span>${icon('clock')}${t('sumLastSync')}</span><b style="font-size:20px">${last ? ago(last) : t('never')}</b></div>
      <div class="kutu"><span>${icon('eye')}${t('sumWatch')}</span><b style="font-size:20px" class="${w ? 'acc-t' : ''}">${w ? t('watchOn') : t('watchOff')}</b>
        ${w ? `<span style="margin-top:2px">${t('watchSince', fmtDate(w.started).slice(11), w.syncs)}</span>` : ''}</div>
    </div>
    <div class="aksiyonlar" style="margin-bottom:14px">
      <button class="btn mavi" id="b-syncall">${icon('upload')}${t('syncAll')}</button>
      <button class="btn" id="b-dryall">${icon('eye')}${t('dryRunAll')}</button>
      <span class="menu-kap"><button class="btn" id="b-moreall" data-menu>${icon('dots')}${t('more')}</button></span>
      <span style="margin-left:auto"></span>
      <button class="btn hayalet" data-add-server>${icon('plus')}${t('addServer')}</button>
    </div>
    <div class="kartlar">${servers.map(serverCard).join('')}</div>`;
}

function serverCard(s) {
  const pend = s.new + s.changed + s.deleted;
  const busy = s.busy ? `<span class="pill busy"><span class="nokta"></span>${t('busy')}</span>` : '';
  return `<div class="kart ${s.enabled ? '' : 'pasif'}" data-s="${esc(s.name)}">
    <div>
      <div class="ust"><h3>${esc(s.name)}</h3><span class="pill mono ${s.protocol}">${s.protocol.toUpperCase()}</span>
        ${s.enabled ? '' : `<span class="pill">${t('disabled')}</span>`}${busy}
        ${s.tunnel ? `<span class="pill vio" title="${esc(s.tunnel)}">${icon('tunnel')}${t('tunnel')}</span>` : ''}
        ${s.connection ? `<span class="pill">${icon('link')}${esc(s.connection)}</span>` : ''}</div>
      <div class="meta">
        <span>${icon('server')}<span class="mono">${esc(s.user ? s.user + '@' : '')}${esc(s.host)}:${s.port}</span></span>
        <span>${icon('folder')}<span class="mono">${esc(s.local_path)}</span> <em>→</em> <span class="mono">${esc(s.remote_path)}</span></span>
        <span>${icon('zap')}${t('maxConn', s.max_connections)}${s.bandwidth ? ' · ' + t('speedLimit', s.bandwidth) : ''}</span>
        <span>${icon('clock')}${s.last_sync ? ago(s.last_sync) : t('never')}</span>
      </div>
    </div>
    <div class="butonlar">
      <button class="btn kucuk mavi" data-act="sync" ${s.busy ? 'disabled' : ''}>${icon('upload')}${t('sync')}</button>
      <span class="menu-kap"><button class="ikon-btn" data-act="menu" data-menu>${icon('dots')}</button></span>
    </div>
    <div class="alt">
      ${s.scanning ? `<span class="rozet">${icon('refresh', 'spin')}${t('scanningLocal')}</span>` : ''}
      ${s.enabled && !s.scanning && !s.scan_error ? (pend ? `<span class="degisim" data-act="status">${t('pendingChip', s.new, s.changed, s.deleted)}</span>` : `<span class="rozet ok-t">${icon('check')}${t('upToDate')}</span>`) : ''}
      ${s.files ? `<span class="rozet">${icon('file')}${t('filesTracked', s.files)}</span>` : ''}
      ${s.milestone ? `<button class="rozet" data-act="milestone">${icon('clock')}${esc(s.milestone)}</button>` : ''}
      ${s.frozen ? `<button class="rozet" data-act="freeze">${icon('snow')}${t('frozenN', s.frozen)}</button>` : ''}
      ${s.failed ? `<button class="rozet bad-t" data-act="retry">${icon('alert')}${t('failedN', s.failed)}</button>` : ''}
      ${s.rollbackable ? `<button class="rozet" data-act="history">${icon('undo')}${t('rollbackN', s.rollbackable)}</button>` : ''}
      <span class="rozet">${icon('shield')}${s.backup ? t('backupOn') : t('noBackup')}</span>
    </div>
    ${s.enabled && !s.first_sync_done ? `<div class="not">${icon('scan')}${t('firstSyncNote')}</div>` : ''}
    ${s.scan_error ? `<div class="not bad">${icon('alert')}${esc(s.scan_error)}</div>` : ''}
  </div>`;
}

function bindServerCard(card, P, s) {
  const act = (sel, fn) => { const b = $(`[data-act="${sel}"]`, card); if (b) b.onclick = (e) => { e.stopPropagation(); fn(b); }; };
  const job = (req, label) => startJobs(P, { servers: [s.name], ...req }, label).catch(fail);
  act('sync', () => job({ kind: 'sync' }, t('sync')));
  act('status', () => statusModal(P, s));
  act('milestone', () => milestoneModal(P, s));
  act('freeze', () => freezeModal(P, s));
  act('retry', () => failedModal(P, s));
  act('history', () => go(pHash(P, 'history', s.name)));
  act('menu', (b) => openMenu(b, [
    { icon: 'eye', label: t('dryRun'), hint: t('dryRunHint'), run: () => job({ kind: 'sync', dry_run: true }, t('dryRun')) },
    { icon: 'upload', label: t('fullSync'), hint: t('fullSyncHint'), run: async () => { if (await confirmBox(t('fullSync'), esc(t('fullSyncConfirm', s.name)), t('fullSync'), true)) job({ kind: 'sync', full: true }, t('fullSync')); } },
    s.failed ? { icon: 'refresh', label: t('retryFailed'), hint: t('retryFailedHint', s.failed), run: () => job({ kind: 'sync', retry_failed: true }, t('retryFailed')) } : null,
    { icon: 'sliders', label: t('syncOnce'), hint: 'include / exclude', run: () => filteredSyncModal(P, s) },
    'hr',
    { icon: 'scan', label: t('calibrate'), hint: t('calibrateHint'), run: () => job({ kind: 'calibrate' }, t('calibrate')) },
    { icon: 'shield', label: t('verify'), hint: t('verifyHint'), run: () => job({ kind: 'verify' }, t('verify')) },
    { icon: 'shield', label: t('verifyFix'), hint: t('verifyFixHint'), run: () => job({ kind: 'verify', fix: true }, t('verifyFix')) },
    { icon: 'gauge', label: t('detect'), hint: t('detectHint'), run: () => job({ kind: 'detect' }, t('detect')) },
    'hr',
    { icon: 'clock', label: t('milestone') + '…', hint: s.milestone || '', run: () => milestoneModal(P, s) },
    s.milestone ? { icon: 'upload', label: t('milestoneSync'), hint: t('milestoneSyncHint'), run: () => job({ kind: 'milestone-sync' }, t('milestoneSync')) } : null,
    { icon: 'snow', label: t('freeze') + '…', hint: t('freezeHint'), run: () => freezeModal(P, s) },
    { icon: 'undo', label: t('rollback') + '…', hint: t('rollbackHint'), run: () => go(pHash(P, 'history', s.name)) },
    'hr',
    s.protocol === 'sftp' ? { icon: 'terminal', label: t('runCmd') + '…', hint: t('runCmdHint'), run: () => execModal(P, s) } : null,
    { icon: 'folder', label: t('browseFiles'), run: () => go(pHash(P, 'files', s.name)) },
    { icon: 'edit', label: t('edit'), run: () => go(pHash(P, 'server', s.name)) },
  ]));
}

async function statusModal(P, s) {
  const lm = await loadModal(t('statusTitle', esc(s.name)), 'orta', () => GET(`/api/p/${enc(P)}/status?server=${enc(s.name)}`));
  if (!lm) return;
  const st = lm.d;
  const list = (arr, mark, click) => arr.length ? `<div class="liste-kutu">${arr.map((x) => `<div ${click ? `class="tikla" data-diff="${esc(x)}"` : ''}>${mark} ${esc(x)}</div>`).join('')}</div>` : `<p class="faint">${t('emptyList')}</p>`;
  // Sync'in yükleyeceği liste ile aynı hesap: atlanacaklar (freeze / protect / engelli uzantı) ayrı sekmede
  const tabs = [['new', t('newFiles'), st.new, '+'], ['changed', t('changedFiles'), st.changed, '~'], ['deleted', t('deletedFiles'), st.deleted, '−'],
    ['frozen', t('stFrozen'), st.frozen || [], '❄'], ['protected', t('stProtected'), st.protected || [], '🔒'], ['blocked', t('stBlocked'), st.blocked || [], '✗']]
    .filter(([k, , a], i) => i < 3 || a.length);
  const first = Math.max(0, tabs.findIndex(([, , a]) => a.length)); // ilk dolu sekmede aç
  const m = modal({
    title: t('statusTitle', esc(s.name)), size: 'orta',
    body: `<div class="tabs" style="margin-bottom:12px">${tabs.map(([k, l, a], i) => `<a href="javascript:void 0" data-st="${k}" class="${i === first ? 'aktif' : ''}">${l} <span class="sayac">${st.counts[k]}</span></a>`).join('')}</div>
      <p class="ipucu" style="margin:-4px 0 10px">${t('clickToDiff')}${st.bytes ? ' · ' + t('stUploadSize', bytes(st.bytes)) : ''}</p>
      ${tabs.map(([k, , a, mk], i) => `<div data-stp="${k}" ${i !== first ? 'hidden' : ''}>${list(a, mk, k === 'new' || k === 'changed')}${k === 'deleted' ? `<p class="ipucu" style="margin-top:10px">${t('deletedNote')}</p>` : ''}</div>`).join('')}`,
    actions: [{ label: t('close'), cls: 'hayalet' }, { label: icon('upload') + t('sync'), cls: 'mavi', run: () => startJobs(P, { kind: 'sync', servers: [s.name] }, t('sync')) }],
  });
  $$('[data-diff]', m.root).forEach((d) => (d.onclick = () => compareModal(P, s.name, d.dataset.diff)));
  $$('[data-st]', m.root).forEach((a) => (a.onclick = () => {
    $$('[data-st]', m.root).forEach((x) => x.classList.toggle('aktif', x === a));
    $$('[data-stp]', m.root).forEach((p) => (p.hidden = p.dataset.stp !== a.dataset.st));
  }));
}

function milestoneModal(P, s) {
  modal({
    title: t('milestoneTitle', esc(s.name)),
    body: `<p>${t('milestoneBody')}</p>${s.milestone ? `<p><b>${t('milestoneCurrent', esc(s.milestone))}</b></p>` : ''}
      <div class="alan"><label>${t('milestoneDate')}</label><input type="text" id="ms-date" placeholder="${LANG === 'tr' ? 'bugün' : 'today'}"><small>${t('milestoneDateHint')}</small></div>`,
    actions: [
      ...(s.milestone ? [{ label: t('milestoneClear'), cls: 'tehlike', run: async () => { await api('DELETE', `/api/p/${enc(P)}/milestones?server=${enc(s.name)}`); toast(t('milestoneCleared')); render(true); } }] : []),
      { label: t('cancel'), cls: 'hayalet' },
      { label: t('milestoneSet'), cls: 'birincil', run: async (root) => {
        const d = await api('PUT', `/api/p/${enc(P)}/milestones`, { servers: [s.name], date: $('#ms-date', root).value });
        toast(t('milestoneSaved', d)); render(true);
      } },
    ],
  });
}

function execModal(P, s) {
  const recent = JSON.parse(localStorageGet('syncftp-cmds') || '[]');
  modal({
    title: `${icon('terminal')}${t('runCmd')} — ${esc(s.name)}`, size: 'orta',
    body: `<p class="ipucu" style="margin:0 0 10px">${t('runCmdHint')} <code>${esc(s.remote_path)}</code></p>
      <div class="alan"><input type="text" class="mono" id="cmd" placeholder="php artisan migrate --force"></div>
      ${recent.length ? `<div class="liste-kutu" style="margin-top:10px">${recent.map((c) => `<div class="tikla" data-cmd="${esc(c)}">$ ${esc(c)}</div>`).join('')}</div>` : ''}`,
    actions: [{ label: t('cancel'), cls: 'hayalet' }, { label: icon('play') + t('runCmd'), cls: 'birincil', run: async (root) => {
      const command = $('#cmd', root).value.trim();
      if (!command) return false;
      localStorageSet('syncftp-cmds', JSON.stringify([command, ...recent.filter((c) => c !== command)].slice(0, 8)));
      const ids = await startJobs(P, { kind: 'exec', servers: [s.name], command }, t('runCmd'));
      if (ids[0]) showJobLog(ids[0]);
    } }],
  }).root.querySelectorAll('[data-cmd]').forEach((d) => (d.onclick = () => { d.closest('.modal').querySelector('#cmd').value = d.dataset.cmd; }));
}
const localStorageGet = (k) => { try { return localStorage.getItem(k); } catch { return null; } };
const localStorageSet = (k, v) => { try { localStorage.setItem(k, v); } catch { /* gizli mod */ } };

async function failedModal(P, s) {
  const lm = await loadModal(`${icon('alert')}${t('failedTitle', esc(s.name))}`, 'orta', () => GET(`/api/p/${enc(P)}/failed?server=${enc(s.name)}`));
  if (!lm) return;
  const files = lm.d || [];
  modal({
    title: `${icon('alert')}${t('failedTitle', esc(s.name))}`, size: 'orta',
    body: `<p>${t('failedBody')}</p><div class="liste-kutu">${files.map((f) => `<div class="bad-t">✗ ${esc(f)}</div>`).join('') || `<div class="faint">${t('emptyList')}</div>`}</div>`,
    actions: [{ label: t('close'), cls: 'hayalet' }, { label: icon('refresh') + t('retryFailed'), cls: 'mavi', run: () => startJobs(P, { kind: 'sync', servers: [s.name], retry_failed: true }, t('retryFailed')) }],
  });
}

function compareServersModal(P, servers) {
  const opts = servers.map((s) => [s.name, s.name]);
  modal({
    title: `${icon('layers')}${t('compareServers')}`,
    body: `<p>${t('compareServersHint')}</p><div class="alanlar">${fSelect('a', t('compareA'), servers[0] ? servers[0].name : '', opts)}${fSelect('b', t('compareB'), servers[1] ? servers[1].name : '', opts)}</div>`,
    actions: [{ label: t('cancel'), cls: 'hayalet' }, { label: icon('layers') + t('compareGo'), cls: 'birincil', run: async (root) => {
      const a = $('[data-k="a"]', root).value, b = $('[data-k="b"]', root).value;
      if (a === b) { toast(t('compareServersHint'), 'bad'); return false; }
      const ids = await startJobs(P, { kind: 'diff', servers: [a, b] }, t('compareServers'));
      if (ids[0]) showJobLog(ids[0]);
    } }],
  });
}

async function freezeModal(P, s) {
  const lm = await loadModal(t('freezeTitle', esc(s.name)), 'orta', () => GET(`/api/p/${enc(P)}/frozen?server=${enc(s.name)}`));
  if (!lm) return;
  const files = lm.d || [];
  const proj = S.projects.find((x) => x.name === P) || {};
  const m = modal({
    title: t('freezeTitle', esc(s.name)), size: 'orta',
    body: `<p>${t('freezeBody')}</p><div class="form" id="fz-form" data-p="${esc(P)}" data-ppath="${esc(proj.path || '')}" data-srv="${esc(s.name)}">${fArea('files', t('freezeFiles'), files, { tree: 'files' })}</div>`,
    actions: [{ label: t('cancel'), cls: 'hayalet' }, { label: t('save'), cls: 'birincil', run: async (root) => {
      const n = await api('PUT', `/api/p/${enc(P)}/frozen`, { server: s.name, files: readForm($('#fz-form', root), {}).files || [] });
      toast(t('freezeSaved', n)); render(true);
    } }],
  });
  bindFormTools(m.root);
}

// ── Sunucular listesi ────────────────────────────────────────────────────────
async function tabServers(proj) {
  const cfg = await getConfig(proj.name);
  const P = proj.name;
  bind((root) => {
    $('#proj-aks', root).innerHTML = `<button class="btn birincil" id="b-add">${icon('plus')}${t('addServer')}</button>`;
    $('#b-add', root).onclick = () => go(pHash(P, 'server', 'new'));
    $$('[data-row]', root).forEach((tr) => {
      const name = tr.dataset.row;
      tr.onclick = (e) => { if (!e.target.closest('button,label')) go(pHash(P, 'server', name)); };
    });
    $$('[data-toggle]', root).forEach((inp) => (inp.onchange = async () => {
      const s = cfg.servers.find((x) => x.name === inp.dataset.toggle);
      s.enabled = inp.checked;
      try { await api('PUT', `/api/p/${enc(P)}/config`, cfg); toast(t('saved')); loadProjects(); } catch (e) { inp.checked = !inp.checked; fail(e); }
    }));
    $$('[data-dup]', root).forEach((b) => (b.onclick = async () => {
      const src = cfg.servers.find((x) => x.name === b.dataset.dup);
      const copy = JSON.parse(JSON.stringify(src));
      let n = 2; while (cfg.servers.some((x) => x.name === `${src.name}-${n}`)) n++;
      copy.name = `${src.name}-${n}`; copy.enabled = false;
      cfg.servers.push(copy);
      try { await api('PUT', `/api/p/${enc(P)}/config`, cfg); go(pHash(P, 'server', copy.name)); } catch (e) { fail(e); }
    }));
    $$('[data-del]', root).forEach((b) => (b.onclick = async () => {
      if (!(await confirmBox(t('delete'), esc(t('deleteServerConfirm', b.dataset.del)), t('delete'), true))) return;
      cfg.servers = cfg.servers.filter((x) => x.name !== b.dataset.del);
      try { await api('PUT', `/api/p/${enc(P)}/config`, cfg); toast(t('saved')); render(true); loadProjects(); } catch (e) { fail(e); }
    }));
  });
  if (!cfg.servers.length) {
    return `<div class="bos"><div class="buyuk">${icon('server', 'lg')}</div><h3>${t('noServers')}</h3><p>${t('noServersSub')}</p>
      <button class="btn birincil" onclick="location.hash='${pHash(P, 'server', 'new')}'">${icon('plus')}${t('addServer')}</button></div>`;
  }
  return `<div class="tablo-kap"><table class="tablo"><thead><tr><th>${t('colName')}</th><th>${t('colProtocol')}</th><th>${t('colHost')}</th><th>${t('colLocal')} → ${t('colRemote')}</th><th>${t('colEnabled')}</th><th></th></tr></thead><tbody>
    ${cfg.servers.map((s) => {
      const conn = s.connection ? cfg.connections.find((c) => c.name === s.connection) : null;
      const proto = ((conn ? conn.protocol : s.protocol) || 'ftp').toLowerCase();
      const host = conn ? conn.host : s.host;
      return `<tr data-row="${esc(s.name)}" class="tikla"><td><div class="ad-hucre">${icon('server')}<span><b>${esc(s.name)}</b></span></div></td>
        <td><span class="pill mono ${proto}">${proto.toUpperCase()}</span></td>
        <td class="mono">${esc(host)}${conn ? ` <span class="faint">(${esc(conn.name)})</span>` : ''}</td>
        <td class="mono faint">${esc(s.local_path || cfg.project.default_path || '.')} → ${esc(s.remote_path)}</td>
        <td><label class="switch"><input type="checkbox" data-toggle="${esc(s.name)}" ${s.enabled ? 'checked' : ''}><i></i></label></td>
        <td class="sag"><button class="ikon-btn" data-dup="${esc(s.name)}" title="${t('duplicate')}">${icon('copy')}</button>
          <button class="ikon-btn tehlike" data-del="${esc(s.name)}" title="${t('delete')}">${icon('trash')}</button></td></tr>`;
    }).join('')}</tbody></table></div>`;
}

// ── form yardımcıları ────────────────────────────────────────────────────────
// data-k="alan" data-ty="text|int|bool|list|words" — readForm bunları nesneye yazar.
// browse: 'dir' | 'file' (mutlak yol) | 'reldir' (proje köküne göreceli klasör); whtest: webhook test düğmesi
const fText = (k, label, val, { hint = '', type = 'text', wide = false, ph = '', ty = 'text', dis = false, mono = false, browse = '', whtest = false } = {}) => {
  const input = `<input type="${type}" data-k="${k}" data-ty="${ty}" value="${esc(val ?? '')}" placeholder="${esc(ph)}" ${dis ? 'disabled' : ''} spellcheck="false" ${mono ? 'class="mono"' : ''} autocomplete="off">`;
  const btns = (browse && !dis ? `<button class="btn" data-browse="${browse}" data-for="${k}">${icon(browse === 'file' ? 'file' : 'folder')}${t('browse')}</button>` : '') +
    (whtest ? `<button class="btn" data-whtest="${k}">${icon('zap')}${t('whTest')}</button>` : '');
  return `<div class="alan ${wide ? 'genis' : ''}"><label>${label}</label>${btns ? `<div class="girdi-grup">${input}${btns}</div>` : input}${hint ? `<small>${hint}</small>` : ''}</div>`;
};
const fNum = (k, label, val, o = {}) => fText(k, label, val ?? 0, { ...o, type: 'number', ty: 'int' });
// tree: 'paths' | 'protect' → etiketin yanında yerel ağaç seçici düğmesi
// chips: tek tıkla eklenen hazır kalıplar (dizi) — 'exclude' | 'protect' | 'block' kısa adları da olur
const PATTERN_CHIPS = {
  exclude: ['node_modules/', 'vendor/', '.git/', '.env', '*.log', 'storage/logs/', 'dist/', 'build/', '.DS_Store', 'Thumbs.db'],
  include: ['public/', 'src/', 'app/', 'dist/', 'build/'],
  protect: ['.env', 'wp-config.php', 'config.php', '.htaccess', 'uploads/', 'storage/'],
  block: ['.psd', '.zip', '.sql', '.bak', '.mp4'],
};
// fArea — liste alanı (textarea YOK): her girdi ayrı satır (ikon + ✕). Ekleme: "Gözat" (yerel ağaçtan
// dosya/klasör işaretle), tek satırlık kalıp kutusu (Enter) ve hazır kalıp çipleri. Değer gizli input'ta
// tutulur (data-k / data-ty list|words) — readForm aynen okur.
const ledIcon = (v) => (/[*?{]/.test(v) ? icon('search', 'faint') : v.endsWith('/') || !/\.[^/]+$/.test(v) ? icon('folder', 'acc-t') : icon('file', 'faint'));
const ledItem = (v) => `<div class="led-item" data-v="${esc(v)}">${ledIcon(v)}<span class="mono">${esc(v)}</span><button class="ikon-btn led-del" data-led-del title="${t('remove')}">${icon('x')}</button></div>`;
const fArea = (k, label, val, { hint = '', ty = 'list', wide = true, tree = '', chips = null, ph = '' } = {}) => {
  const items = (val || []).filter(Boolean);
  const list = typeof chips === 'string' ? PATTERN_CHIPS[chips] : chips;
  const placeholder = ph || (ty === 'words' ? '.jpg' : tree ? t('ledPh') : t('ledPhCmd'));
  return `<div class="alan ${wide ? 'genis' : ''}"><label>${label}</label>
    <div class="liste-ed" data-led="${k}" data-ledty="${ty}">
      <input type="hidden" data-k="${k}" data-ty="${ty}" value="${esc(ty === 'words' ? items.join(' ') : items.join('\n'))}">
      <div class="led-items">${items.length ? items.map(ledItem).join('') : `<div class="led-bos">${t('ledEmpty')}</div>`}</div>
      <div class="led-ekle">${tree ? `<button class="btn kucuk" data-tree="${tree}" data-for="${k}">${icon('folder')}${t('browse')}</button>` : ''}
        <input type="text" class="mono" data-led-input placeholder="${esc(placeholder)}" spellcheck="false" autocomplete="off">
        <button class="btn kucuk" data-led-add>${icon('plus')}${t('ledAdd')}</button></div>
      ${list ? `<div class="cipler">${list.map((c) => `<button class="cip" data-chip="${esc(c)}" data-for="${k}">+ ${esc(c)}</button>`).join('')}</div>` : ''}
    </div>${hint ? `<small>${hint}</small>` : ''}</div>`;
};
// Liste editörü yardımcıları (event delegation — form ne zaman render edilirse edilsin çalışır)
function ledBox(el) { return el.closest('[data-led]'); }
function ledGet(box) { const h = box.querySelector('input[type=hidden]'); return box.dataset.ledty === 'words' ? unwords(h.value) : unlines(h.value); }
function ledSet(box, arr) {
  const uniq = [...new Set(arr.map((x) => String(x).trim()).filter(Boolean))];
  box.querySelector('input[type=hidden]').value = box.dataset.ledty === 'words' ? uniq.join(' ') : uniq.join('\n');
  box.querySelector('.led-items').innerHTML = uniq.length ? uniq.map(ledItem).join('') : `<div class="led-bos">${t('ledEmpty')}</div>`;
}
function ledAddFromInput(box) {
  const inp = box.querySelector('[data-led-input]');
  const vals = box.dataset.ledty === 'words' ? unwords(inp.value) : [inp.value.trim()];
  if (!vals.filter(Boolean).length) return;
  ledSet(box, [...ledGet(box), ...vals]);
  inp.value = '';
  inp.focus();
}
document.addEventListener('click', (e) => {
  const del = e.target.closest('[data-led-del]');
  if (del) { e.preventDefault(); const box = ledBox(del); ledSet(box, ledGet(box).filter((v) => v !== del.closest('.led-item').dataset.v)); return; }
  const add = e.target.closest('[data-led-add]');
  if (add) { e.preventDefault(); ledAddFromInput(ledBox(add)); return; }
  const chip = e.target.closest('[data-chip]');
  if (chip) { e.preventDefault(); const box = ledBox(chip); ledSet(box, [...ledGet(box), chip.dataset.chip]); }
});
document.addEventListener('keydown', (e) => {
  if (e.key === 'Enter' && e.target.matches && e.target.matches('[data-led-input]')) { e.preventDefault(); ledAddFromInput(ledBox(e.target)); }
});
const fBool = (k, label, val, hint = '', dis = false) =>
  `<label class="anahtar"><span class="metin"><b>${label}</b>${hint ? `<small>${hint}</small>` : ''}</span><span class="switch"><input type="checkbox" data-k="${k}" data-ty="bool" ${val ? 'checked' : ''} ${dis ? 'disabled' : ''}><i></i></span></label>`;
const fSelect = (k, label, val, opts, { hint = '', dis = false } = {}) =>
  `<div class="alan"><label>${label}</label><select data-k="${k}" data-ty="text" ${dis ? 'disabled' : ''}>${opts.map(([v, l]) => `<option value="${esc(v)}" ${v === (val ?? '') ? 'selected' : ''}>${esc(l)}</option>`).join('')}</select>${hint ? `<small>${hint}</small>` : ''}</div>`;
const section = (icn, title, desc, inner) => `<div class="bolum"><h3>${icon(icn)}${title}</h3>${desc ? `<p class="aciklama">${desc}</p>` : '<div style="height:10px"></div>'}${inner}</div>`;

// Form içindeki Gözat / ağaç / webhook test düğmelerini bağlar.
// Bağlam form kökündeki data-p (proje), data-ppath (proje yolu), data-srv (sunucu) özniteliklerinden okunur.
function bindFormTools(root) {
  const ctx = root.closest('[data-ppath]') || root.querySelector('[data-ppath]') || root;
  const P = ctx.dataset.p, ppath = ctx.dataset.ppath || '', srv = ctx.dataset.srv || '';
  const input = (k) => root.querySelector(`[data-k="${k}"]`);
  $$('[data-browse]', root).forEach((b) => (b.onclick = async (e) => {
    e.preventDefault();
    const inp = input(b.dataset.for), kind = b.dataset.browse;
    let initial = inp.value.trim();
    if (kind === 'reldir') initial = !initial || initial === '.' ? ppath : (/^([a-zA-Z]:|[\\/])/.test(initial) ? initial : ppath + '\\' + initial.replace(/\//g, '\\'));
    if (kind === 'file' && !initial && S.meta) initial = S.meta.home + '\\.ssh';
    if (kind === 'file' && initial && /\.[a-z0-9]+$/i.test(initial)) initial = initial.replace(/[\\/][^\\/]*$/, '');
    const picked = await pickPath({ mode: kind === 'file' ? 'file' : 'dir', initial });
    if (picked == null) return;
    inp.value = kind === 'reldir' ? relToProject(ppath, picked) : picked;
    inp.dispatchEvent(new Event('change'));
  }));
  // Ağaç kökü: proje (data-p) → projenin yerel dizini; yoksa data-rootfrom="#input" (ör. hızlı yüklemedeki kaynak);
  // o da yoksa (varsayılanlar) bir kez örnek klasör sorulur ve hatırlanır.
  const treeRoot = async () => {
    if (P) return '';
    const from = ctx.dataset.rootfrom && document.querySelector(ctx.dataset.rootfrom);
    if (from) {
      if (from.value.trim()) return from.value.trim();
      toast(t('treeNeedRoot'), 'bad');
      return null;
    }
    if (ctx.dataset.root) return ctx.dataset.root;
    const picked = await pickPath({ mode: 'dir', title: t('treePickRoot') });
    if (picked) ctx.dataset.root = picked;
    return picked;
  };
  $$('[data-tree]', root).forEach((b) => (b.onclick = async (e) => {
    e.preventDefault();
    const box = ledBox(b);
    const r = await treeRoot();
    if (r === null) return;
    const res = await pickLocalTree({ P, root: r, server: srv, current: ledGet(box), mode: b.dataset.tree });
    if (res) ledSet(box, res);
  }));
  $$('[data-whtest]', root).forEach((b) => (b.onclick = async (e) => {
    e.preventDefault();
    const url = input(b.dataset.whtest).value.trim();
    if (!url) return;
    b.disabled = true;
    try { toast(t('whTestOk', await api('POST', `/api/p/${enc(P)}/webhook-test`, { url }))); } catch (err) { fail(err); } finally { b.disabled = false; }
  }));
}

function readForm(root, obj) {
  $$('[data-k]', root).forEach((el) => {
    if (el.disabled && el.dataset.keepDisabled !== '1') return;
    const k = el.dataset.k, ty = el.dataset.ty;
    let v;
    if (ty === 'bool') v = el.checked;
    else if (ty === 'int') v = parseInt(el.value, 10) || 0;
    else if (ty === 'list') v = unlines(el.value);
    else if (ty === 'words') v = unwords(el.value);
    else v = el.value.trim();
    const path = k.split('.');
    let o = obj;
    for (let i = 0; i < path.length - 1; i++) o = o[path[i]] ??= {};
    o[path[path.length - 1]] = v;
  });
  return obj;
}

const protoOpts = [['', 'FTP'], ['ftps', 'FTPS (TLS)'], ['sftp', 'SFTP (SSH)']];
function connFields(src, locked, prefix = '') {
  const dis = !!locked;
  const proto = (src.protocol || '').toLowerCase();
  return `<div class="alanlar uc">
      ${fSelect(prefix + 'protocol', t('fProtocol'), proto, protoOpts, { dis })}
      ${fText(prefix + 'host', t('fHost'), src.host, { dis, mono: true, ph: 'ftp.example.com' })}
      ${fNum(prefix + 'port', t('fPort'), src.port || (proto === 'sftp' ? 22 : 21), { dis })}
      ${fText(prefix + 'user', t('fUser'), src.user, { dis })}
      ${fText(prefix + 'password', t('fPassword'), src.password, { type: 'password', dis })}
      ${fText(prefix + 'key_file', t('fKeyFile'), src.key_file, { dis, mono: true, ph: '~/.ssh/id_ed25519', hint: t('fKeyFileHint'), browse: 'file' })}
    </div><div class="alanlar" style="margin-top:14px">
      ${fBool(prefix + 'passive', t('fPassive'), src.passive ?? true, '', dis)}
      ${fBool(prefix + 'tls_skip_verify', t('fTls'), src.tls_skip_verify, t('fTlsHint'), dis)}
      ${fBool(prefix + 'disable_epsv', t('fEpsv'), src.disable_epsv, t('fEpsvHint'), dis)}
      ${fBool(prefix + 'nat_workaround', t('fNat'), src.nat_workaround, t('fNatHint'), dis)}
    </div>`;
}
function tunnelFields(src, locked) {
  const dis = !!locked;
  return `<div class="alanlar uc">
    ${fText('tunnel_host', t('fTunnelHost'), src.tunnel_host, { dis, mono: true, ph: 'bastion.example.com' })}
    ${fNum('tunnel_port', t('fTunnelPort'), src.tunnel_port || 22, { dis })}
    ${fText('tunnel_user', t('fTunnelUser'), src.tunnel_user, { dis, hint: t('fTunnelUserHint') })}
    ${fText('tunnel_password', t('fTunnelPass'), src.tunnel_password, { type: 'password', dis })}
    ${fText('tunnel_key_file', t('fTunnelKey'), src.tunnel_key_file, { dis, mono: true, wide: false, ph: '~/.ssh/id_ed25519', browse: 'file' })}
  </div>`;
}
// Protokol değişince port varsayılandaysa çevir
function bindProtoPort(root) {
  const sel = $('[data-k="protocol"]', root), port = $('[data-k="port"]', root);
  if (!sel || !port) return;
  sel.onchange = () => {
    const p = parseInt(port.value, 10);
    if (sel.value === 'sftp' && (p === 21 || !p)) port.value = 22;
    if (sel.value !== 'sftp' && (p === 22 || !p)) port.value = 21;
  };
}

// ── Sunucu düzenleyici ───────────────────────────────────────────────────────
async function tabServerEdit(proj, name) {
  const P = proj.name;
  const cfg = await getConfig(P);
  const isNew = name === 'new';
  const idx = cfg.servers.findIndex((s) => s.name === name);
  if (!isNew && idx < 0) throw new Error(t('notFound') + ': ' + name);
  const s = isNew ? { name: '', enabled: true, passive: true, max_connections: 4, max_retries: 2, remote_path: '/' } : cfg.servers[idx];
  const conn = s.connection ? cfg.connections.find((c) => c.name === s.connection) : null;
  const defPath = cfg.project.default_path || '.';
  const globalBw = cfg.sync.bandwidth_limit ? cfg.sync.bandwidth_limit + ' KB/s' : '∞';

  bind((root) => {
    $('#proj-aks', root).innerHTML = '';
    bindProtoPort(root);
    bindFormTools($('#srv-form', root));
    $('[data-k="connection"]', root).onchange = (e) => {
      // Profil değişince credential alanlarını yeniden çiz
      const c = cfg.connections.find((x) => x.name === e.target.value);
      readForm($('#srv-form', root), s);
      s.connection = e.target.value;
      $('#conn-box', root).innerHTML = connBox(c);
      $('#tunnel-box', root).innerHTML = tunnelFields(c || s, !!c);
      bindProtoPort(root);
      bindFormTools($('#srv-form', root));
    };
    const collect = () => {
      const out = JSON.parse(JSON.stringify(s));
      readForm($('#srv-form', root), out);
      if (out.connection) { // profil kullanılıyorsa credential'lar profilden gelir
        for (const k of ['host', 'port', 'user', 'password', 'protocol', 'key_file', 'passive', 'disable_epsv', 'nat_workaround', 'tls_skip_verify', 'tunnel_host', 'tunnel_port', 'tunnel_user', 'tunnel_password', 'tunnel_key_file']) delete out[k];
      }
      if (out.protocol === '') delete out.protocol;
      return out;
    };
    $('#b-save', root).onclick = async () => {
      const out = collect();
      const next = JSON.parse(JSON.stringify(cfg));
      if (isNew) next.servers.push(out); else next.servers[idx] = out;
      try { await api('PUT', `/api/p/${enc(P)}/config`, next); toast(t('saved')); await loadProjects(); go(pHash(P, 'servers')); } catch (e) { fail(e); }
    };
    $('#b-test', root).onclick = async (e) => {
      const b = e.currentTarget; b.disabled = true; const old = b.innerHTML; b.innerHTML = icon('refresh', 'spin') + t('testing');
      try { const r = await api('POST', `/api/p/${enc(P)}/test`, { server: collect(), connections: cfg.connections }); toast(t('testOk', r.ms, r.entries) + (r.list_error ? ' — ' + r.list_error : ''), r.list_error ? 'bad' : 'ok'); }
      catch (err) { fail(err); } finally { b.disabled = false; b.innerHTML = old; }
    };
    const det = $('#b-detect', root);
    if (det) det.onclick = () => startJobs(P, { kind: 'detect', servers: [s.name] }, t('detect')).then((ids) => {
      const j = S.jobs.get(ids[0]);
      if (j) j.onDone = (jj) => { if (jj.result && jj.result[s.name]) $('[data-k="max_connections"]', root).value = jj.result[s.name]; };
    }).catch(fail);
    const del = $('#b-del', root);
    if (del) del.onclick = async () => {
      if (!(await confirmBox(t('delete'), esc(t('deleteServerConfirm', s.name)), t('delete'), true))) return;
      const next = JSON.parse(JSON.stringify(cfg));
      next.servers.splice(idx, 1);
      try { await api('PUT', `/api/p/${enc(P)}/config`, next); toast(t('saved')); await loadProjects(); go(pHash(P, 'servers')); } catch (e) { fail(e); }
    };
  });

  function connBox(c) {
    return c ? `<div class="kilit" style="margin-bottom:14px">${icon('lock')}${esc(t('lockedByProfile', c.name))}</div>${connFields(c, true)}` : connFields(s, false);
  }

  return `<a class="faint" href="${pHash(P, 'servers')}" style="display:inline-flex;gap:6px;align-items:center;margin-bottom:12px">${icon('left')}${t('back')}</a>
  <h2 style="margin-top:0">${isNew ? t('serverNew') : t('serverEdit') + ' — ' + esc(s.name)}</h2>
  <div class="form" id="srv-form" data-p="${esc(P)}" data-ppath="${esc(proj.path)}" data-srv="${esc(isNew ? '' : s.name)}">
    ${section('server', t('secGeneral'), '', `<div class="alanlar">${fText('name', t('fName'), s.name, { ph: 'production' })}${fBool('enabled', t('fEnabled'), s.enabled, t('fEnabledHint'))}</div>`)}
    ${section('link', t('secConnection'), t('secConnectionDesc'), `<div class="alanlar" style="margin-bottom:14px">${fSelect('connection', t('fConnection'), s.connection || '', [['', t('manual')], ...cfg.connections.map((c) => [c.name, `${c.name} — ${c.host}`])])}</div><div id="conn-box">${connBox(conn)}</div>`)}
    ${section('tunnel', t('secTunnel'), t('secTunnelDesc'), `<div id="tunnel-box">${tunnelFields(conn || s, !!conn)}</div>`)}
    ${section('folder', t('secPaths'), '', `<div class="alanlar">${fText('remote_path', t('fRemote'), s.remote_path, { mono: true, hint: t('fRemoteHint') })}${fText('local_path', t('fLocal'), s.local_path, { mono: true, ph: defPath, hint: t('fLocalHint', esc(defPath)), browse: 'reldir' })}</div>`)}
    ${section('zap', t('secPerf'), '', `<div class="alanlar uc">
        <div class="alan"><label>${t('fMaxConn')}</label><div class="girdi-grup"><input type="number" data-k="max_connections" data-ty="int" value="${s.max_connections || 1}" min="1">${isNew ? '' : `<button class="btn" id="b-detect">${icon('gauge')}${t('measure')}</button>`}</div><small>${t('fMaxConnHint')}</small></div>
        ${fNum('max_retries', t('fRetries'), s.max_retries)}
        ${fNum('bandwidth_limit', t('fBandwidth'), s.bandwidth_limit, { hint: t('fBandwidthHint', globalBw) })}</div>`)}
    ${section('upload', t('secUpload'), '', `<div class="alanlar">
        ${fBool('minify', t('fMinify'), s.minify, t('fMinifyHint'))}${fBool('obfuscate', t('fObfuscate'), s.obfuscate, t('fObfuscateHint'))}
        ${fBool('no_backup', t('fNoBackup'), s.no_backup, t('fNoBackupHint'))}${fBool('direct_upload', t('fDirect'), s.direct_upload, t('fDirectHint'))}
        ${fText('webhook', t('fWebhook'), s.webhook, { wide: true, mono: true, hint: t('fWebhookHint'), ph: 'https://…', whtest: true })}</div>`)}
    ${section('sliders', t('secFilters'), t('secFiltersDesc'), `<div class="alanlar">
        ${fArea('include', t('fInclude'), s.include, { wide: false, tree: 'paths', chips: 'include' })}${fArea('exclude', t('fExclude'), s.exclude, { wide: false, tree: 'paths', chips: 'exclude' })}
        ${fArea('protect', t('fProtect'), s.protect, { wide: false, tree: 'protect', chips: 'protect' })}${fArea('block_extensions', t('fBlock'), s.block_extensions, { ty: 'words', wide: false, hint: t('fBlockHint'), chips: 'block' })}</div>`)}
    ${section('log', t('secHooks'), t('secHooksDesc'), `<div class="alanlar">${fArea('pre_sync', t('fPre'), s.pre_sync, { wide: false, rows: 3 })}${fArea('post_sync', t('fPost'), s.post_sync, { wide: false, rows: 3 })}</div>`)}
    ${section('terminal', t('secRemote'), t('secRemoteDesc'), `<div class="alanlar">${fArea('remote_pre_sync', t('fRemotePre'), s.remote_pre_sync, { wide: false, rows: 3 })}${fArea('remote_post_sync', t('fRemotePost'), s.remote_post_sync, { wide: false, rows: 3, hint: 'php artisan migrate --force' })}</div>`)}
  </div>
  <div class="form-alt"><div class="sol">${isNew ? '' : `<button class="btn tehlike" id="b-del">${icon('trash')}${t('delete')}</button>`}</div>
    <button class="btn" id="b-test">${icon('test')}${t('testConn')}</button>
    <button class="btn hayalet" onclick="history.back()">${t('cancel')}</button>
    <button class="btn birincil" id="b-save">${icon('check')}${t('save')}</button></div>`;
}

// ── Bağlantı profilleri ──────────────────────────────────────────────────────
async function tabConnections(proj) {
  const P = proj.name;
  const cfg = await getConfig(P);
  bind((root) => {
    $('#proj-aks', root).innerHTML = `<button class="btn" id="b-lib">${icon('link')}${t('connFromLib')}</button><button class="btn birincil" id="b-add">${icon('plus')}${t('addConnection')}</button>`;
    $('#b-add', root).onclick = () => go(pHash(P, 'connection', 'new'));
    $('#b-lib', root).onclick = () => connFromLibraryModal(P).catch(fail);
    $$('[data-row]', root).forEach((tr) => (tr.onclick = () => go(pHash(P, 'connection', tr.dataset.row))));
    $$('[data-tolib]', root).forEach((b) => (b.onclick = async (e) => {
      e.stopPropagation();
      try { const r = await api('POST', `/api/p/${enc(P)}/connections/to-library`, { connection: b.dataset.tolib }); toast(t('connToLibDone', r.name)); } catch (err) { fail(err); }
    }));
  });
  if (!cfg.connections.length) {
    return `<p class="muted" style="margin-top:-6px">${t('connectionsSub')}</p><div class="bos"><div class="buyuk">${icon('link', 'lg')}</div><h3>${t('noConnections')}</h3><p>${t('noConnectionsSub')}</p>
      <div class="aksiyonlar" style="justify-content:center"><button class="btn" onclick="connFromLibraryModal('${esc(P)}').catch(fail)">${icon('link')}${t('connFromLib')}</button>
      <button class="btn birincil" onclick="location.hash='${pHash(P, 'connection', 'new')}'">${icon('plus')}${t('addConnection')}</button></div></div>`;
  }
  return `<p class="muted" style="margin-top:-6px">${t('connectionsSub')}</p><div class="tablo-kap"><table class="tablo"><thead><tr><th>${t('colName')}</th><th>${t('colProtocol')}</th><th>${t('colHost')}</th><th>${t('usedBy')}</th><th></th></tr></thead><tbody>
    ${cfg.connections.map((c) => {
      const used = cfg.servers.filter((s) => s.connection === c.name).map((s) => s.name);
      const proto = (c.protocol || 'ftp').toLowerCase();
      return `<tr data-row="${esc(c.name)}" class="tikla"><td><div class="ad-hucre">${icon('link')}<b>${esc(c.name)}</b></div></td><td><span class="pill mono ${proto}">${proto.toUpperCase()}</span></td>
        <td class="mono">${esc(c.user ? c.user + '@' : '')}${esc(c.host)}:${c.port}${c.tunnel_host ? ` <span class="pill vio">${icon('tunnel')}${t('tunnel')}</span>` : ''}</td><td class="muted">${esc(used.join(', ') || '—')}</td>
        <td class="sag"><button class="btn kucuk hayalet" data-tolib="${esc(c.name)}" title="${t('connToLib')}">${icon('download')}${t('connToLib')}</button></td></tr>`;
    }).join('')}</tbody></table></div>`;
}

async function tabConnEdit(proj, name) {
  const P = proj.name;
  const cfg = await getConfig(P);
  const isNew = name === 'new';
  const idx = cfg.connections.findIndex((c) => c.name === name);
  if (!isNew && idx < 0) throw new Error(t('notFound') + ': ' + name);
  const c = isNew ? { name: '', passive: true, port: 21 } : cfg.connections[idx];
  bind((root) => {
    $('#proj-aks', root).innerHTML = '';
    bindProtoPort(root);
    bindFormTools($('#conn-form', root));
    $('#b-save', root).onclick = async () => {
      const out = readForm($('#conn-form', root), JSON.parse(JSON.stringify(c)));
      if (out.protocol === '') delete out.protocol;
      const next = JSON.parse(JSON.stringify(cfg));
      if (isNew) next.connections.push(out);
      else {
        next.connections[idx] = out;
        if (out.name !== c.name) next.servers.forEach((s) => { if (s.connection === c.name) s.connection = out.name; });
      }
      try { await api('PUT', `/api/p/${enc(P)}/config`, next); toast(t('saved')); go(pHash(P, 'connections')); } catch (e) { fail(e); }
    };
    $('#b-test', root).onclick = async (e) => {
      const b = e.currentTarget; b.disabled = true;
      const out = readForm($('#conn-form', root), JSON.parse(JSON.stringify(c)));
      const probe = { ...out, name: '__test__', remote_path: '/', connection: '' };
      try { const r = await api('POST', `/api/p/${enc(P)}/test`, { server: probe, connections: [] }); toast(t('testOk', r.ms, r.entries)); }
      catch (err) { fail(err); } finally { b.disabled = false; }
    };
    const del = $('#b-del', root);
    if (del) del.onclick = async () => {
      const used = cfg.servers.filter((s) => s.connection === c.name).map((s) => s.name);
      if (used.length) return toast(t('connInUse', c.name, used.join(', ')), 'bad');
      if (!(await confirmBox(t('delete'), esc(t('deleteConnConfirm', c.name)), t('delete'), true))) return;
      const next = JSON.parse(JSON.stringify(cfg));
      next.connections.splice(idx, 1);
      try { await api('PUT', `/api/p/${enc(P)}/config`, next); toast(t('saved')); go(pHash(P, 'connections')); } catch (e) { fail(e); }
    };
  });
  return `<a class="faint" href="${pHash(P, 'connections')}" style="display:inline-flex;gap:6px;align-items:center;margin-bottom:12px">${icon('left')}${t('back')}</a>
    <h2 style="margin-top:0">${isNew ? t('connNew') : t('connEdit') + ' — ' + esc(c.name)}</h2>
    <div class="form" id="conn-form" data-p="${esc(P)}" data-ppath="${esc(proj.path)}">
      ${section('link', t('secGeneral'), '', `<div class="alanlar">${fText('name', t('fName'), c.name, { ph: 'my-host' })}</div>`)}
      ${section('server', t('secConnection'), '', connFields(c, false))}
      ${section('tunnel', t('secTunnel'), t('secTunnelDesc'), tunnelFields(c, false))}
    </div>
    <div class="form-alt"><div class="sol">${isNew ? '' : `<button class="btn tehlike" id="b-del">${icon('trash')}${t('delete')}</button>`}</div>
      <button class="btn" id="b-test">${icon('test')}${t('testConn')}</button>
      <button class="btn hayalet" onclick="history.back()">${t('cancel')}</button>
      <button class="btn birincil" id="b-save">${icon('check')}${t('save')}</button></div>`;
}

// ── Ayarlar ──────────────────────────────────────────────────────────────────
async function tabSettings(proj) {
  const P = proj.name;
  const cfg = await getConfig(P);
  const sy = cfg.sync || {};
  bind((root) => {
    $('#proj-aks', root).innerHTML = '';
    bindFormTools($('#set-form', root));
    $('#ign-tmpl', root).onchange = (e) => {
      const v = tmpls[e.target.value];
      if (v) { const ta = $('#ign-text', root); ta.value = (ta.value.trim() ? ta.value.trimEnd() + '\n\n' : '') + v.trim() + '\n'; }
      e.target.value = '';
    };
    const ignSel = $('#ign-dir', root);
    const ignShow = () => {
      $('#ign-text', root).value = ignCur.content || '';
      $('#ign-text', root).disabled = !ignCur.file;
      $('#ign-save', root).disabled = !ignCur.file;
      $('#ign-file', root).textContent = ignCur.file ? `${ignCur.dir === '.' ? '' : ignCur.dir + '/'}${ignCur.file}` : t('ignoreNoFile');
    };
    if (ignSel) ignSel.onchange = () => { ignCur = ignTargets[+ignSel.value]; ignShow(); };
    $('#ign-save', root).onclick = async () => {
      try {
        await api('PUT', `/api/p/${enc(P)}/ignore`, { dir: ignCur.dir, content: $('#ign-text', root).value });
        ignCur.content = $('#ign-text', root).value;
        toast(t('ignoreSaved'));
      } catch (e) { fail(e); }
    };
    $('#b-save', root).onclick = async () => {
      const next = JSON.parse(JSON.stringify(cfg));
      readForm($('#set-form', root), next);
      if (!next.sync.ignore_files || !next.sync.ignore_files.length) next.sync.ignore_files = null;
      try { await api('PUT', `/api/p/${enc(P)}/config`, next); toast(t('saved')); render(true); } catch (e) { fail(e); }
    };
    $('#kr', root).onchange = async (e) => {
      const on = e.target.checked;
      try { await api('POST', `/api/p/${enc(P)}/keyring`, { on }); toast(on ? t('keyringOn') : t('keyringOff')); loadProjects(); }
      catch (err) { e.target.checked = !on; fail(err); }
    };
    $('#b-rename', root).onclick = async () => {
      const m = modal({
        title: t('renameProject'), body: `<div class="alan"><label>${t('displayName')}</label><input type="text" id="rn" value="${esc(P)}"></div>`,
        actions: [{ label: t('cancel'), cls: 'hayalet' }, { label: t('save'), cls: 'birincil', run: async (r) => {
          const nn = $('#rn', r).value.trim();
          await api('PATCH', `/api/projects/${enc(P)}`, { name: nn });
          await loadProjects();
          const np = S.projects.find((x) => x.path === proj.path);
          go(pHash(np ? np.name : nn, 'settings'));
        } }],
      });
      return m;
    };
    bindProjTools(root, P);
  });
  // Sync'in okuduğu ignore dosyaları — farklı yerel dizinli her sunucu grubu için bir tane
  const ignTargets = (await GET(`/api/p/${enc(P)}/ignore`).catch(() => ({ targets: [] }))).targets || [];
  let ignCur = ignTargets[0] || { dir: '.', file: '', content: '', gitignore: '', servers: [] };
  const ign = ignCur;
  const tmpls = await GET('/api/ignore-templates').catch(() => ({}));
  const gitLines = (ign.gitignore || '').split('\n').filter((l) => l.trim() && !l.trim().startsWith('#')).length;
  return `<div class="form" id="set-form" data-p="${esc(P)}" data-ppath="${esc(proj.path)}" data-srv="">
    ${section('folder', t('secProject'), '', `<div class="alanlar">${fText('project.name', t('fProjectName'), cfg.project.name)}${fText('project.default_path', t('fDefaultPath'), cfg.project.default_path, { mono: true, ph: '.', hint: t('fDefaultPathHint'), browse: 'reldir' })}</div>`)}
    ${section('sliders', t('secGlobal'), t('secGlobalDesc'), `<div class="alanlar">
        ${fArea('sync.include', t('fInclude'), sy.include, { wide: false, tree: 'paths', chips: 'include' })}${fArea('sync.exclude', t('fExclude'), sy.exclude, { wide: false, tree: 'paths', chips: 'exclude' })}
        ${fArea('sync.protect', t('fProtect'), sy.protect, { wide: false, tree: 'protect', chips: 'protect' })}${fArea('sync.block_extensions', t('fBlock'), sy.block_extensions, { ty: 'words', wide: false, hint: t('fBlockHint'), chips: 'block' })}
        ${fArea('sync.ignore_files', t('fIgnoreFiles'), sy.ignore_files, { wide: false, tree: 'paths', ph: '.gitignore', hint: t('fIgnoreFilesHint') })}
        ${fText('sync.webhook', t('fWebhook'), sy.webhook, { mono: true, ph: 'https://…', whtest: true })}
        ${fNum('sync.bandwidth_limit', t('fBandwidth'), sy.bandwidth_limit, { hint: '0 = ∞' })}
        ${fNum('sync.backup_keep', t('fBackupKeep'), sy.backup_keep || 10, { hint: t('fBackupKeepHint') })}
        ${fBool('sync.no_backup', t('fNoBackup'), sy.no_backup, t('fNoBackupHint'))}
        ${fBool('sync.direct_upload', t('fDirect'), sy.direct_upload, t('fDirectHint'))}
        ${fBool('first_sync.full', t('fFirstFull'), cfg.first_sync && cfg.first_sync.full, t('fFirstFullHint'))}
      </div>`)}
    ${section('log', t('secHooks'), t('secHooksDesc'), `<div class="alanlar">${fArea('sync.pre_sync', t('fPre'), sy.pre_sync, { wide: false, rows: 3 })}${fArea('sync.post_sync', t('fPost'), sy.post_sync, { wide: false, rows: 3 })}</div>`)}
    <div class="bolum"><h3>${icon('x')}${t('ignoreTitle')}</h3><p class="aciklama">${t('ignoreDesc')}${gitLines ? ' · ' + t('gitignoreNote', gitLines) : ''}</p>
      <div class="aksiyonlar" style="margin-bottom:10px">${ignTargets.length > 1 ? `<select id="ign-dir" style="width:auto">${ignTargets.map((g, i) => `<option value="${i}">${esc(g.dir)} — ${esc((g.servers || []).join(', '))}</option>`).join('')}</select>` : ''}
        <span class="mono faint" id="ign-file" style="font-size:12px">${ign.file ? esc(`${ign.dir === '.' ? '' : ign.dir + '/'}${ign.file}`) : t('ignoreNoFile')}</span>
        <select id="ign-tmpl" style="width:auto"><option value="">${t('ignoreTemplate')}</option>${Object.keys(tmpls).map((k) => `<option value="${esc(k)}">${esc(k)}</option>`).join('')}</select>
        <span style="margin-left:auto"></span><button class="btn" id="ign-save">${icon('check')}${t('save')}</button></div>
      <textarea id="ign-text" rows="10" spellcheck="false" ${ign.file ? '' : 'disabled'}>${esc(ign.content || '')}</textarea></div>
    <div class="bolum"><h3>${icon('lock')}${t('secSecurity')}</h3><div style="height:10px"></div>
      <label class="anahtar"><span class="metin"><b>${t('keyringTitle')}</b><small>${t('keyringHint')}</small></span><span class="switch"><input type="checkbox" id="kr" ${cfg.project.keyring ? 'checked' : ''}><i></i></span></label></div>
    ${projToolsHTML()}
    <div class="bolum"><h3>${icon('layers')}${t('secRegistry')}</h3><p class="aciklama">${t('registryDesc')}</p>
      <div class="aksiyonlar"><button class="btn" id="b-rename">${icon('edit')}${t('renameProject')}</button></div></div>
  </div>
  <div class="form-alt"><button class="btn birincil" id="b-save">${icon('check')}${t('save')}</button></div>`;
}

// ── Dosyalar ─────────────────────────────────────────────────────────────────
async function tabFiles(proj, srvName, relPath) {
  const P = proj.name;
  const cfg = await getConfig(P);
  if (!cfg.servers.length) return `<div class="bos"><h3>${t('noServers')}</h3></div>`;
  const srvObj = cfg.servers.find((s) => s.name === srvName) || cfg.servers.find((s) => s.enabled) || cfg.servers[0];
  const srv = srvObj.name;
  const rel = relPath || '';
  const q = `server=${enc(srv)}`;
  let data;
  try { data = await GET(`/api/p/${enc(P)}/remote/ls?${q}&path=${enc(rel)}`); }
  catch (e) { data = { error: e.message, entries: [] }; }
  const frozenSet = new Set(await GET(`/api/p/${enc(P)}/frozen?${q}`).catch(() => []));
  const hashFor = (p) => `${pHash(P, 'files', srv)}/${enc(p)}`;
  const full = (name) => (rel ? rel + '/' : '') + name;
  const crumbs = rel.split('/').filter(Boolean);
  const entries = data.entries || [];
  const sel = new Set();

  // Freeze ekle/çıkar: dosyalar doğrudan, klasörler yereldeki dosyalarına açılır
  // CLI tarayıcısıyla aynı çekirdek: klasör → sunucudaki dosyaları; hepsi frozen ise çözer
  async function toggleFreeze(items) {
    try {
      const r = await api('POST', `/api/p/${enc(P)}/frozen/toggle?${q}`, { items: items.map((it) => ({ path: it.path, dir: it.type === 'dir' })) });
      toast(r.frozen ? t('frozenAdded', r.n) : t('frozenRemoved', r.n));
    } catch (e) { fail(e); }
    render(true);
  }
  async function removeItems(items) {
    const msg = items.length === 1 ? (items[0].type === 'dir' ? t('deleteDirConfirm', items[0].path) : t('deleteRemoteConfirm', items[0].path)) : t('deleteManyConfirm', items.length);
    if (!(await confirmBox(t('delete'), esc(msg), t('delete'), true))) return;
    try {
      for (const it of items) await api('DELETE', `/api/p/${enc(P)}/remote/rm?${q}&path=${enc(it.path)}&dir=${it.type === 'dir'}`);
      toast(t('deleted'));
    } catch (err) { fail(err); }
    render(true);
  }
  function moveItems(items) {
    modal({
      title: `${icon('folder')}${t('moveTo')}`,
      body: `<div class="alan"><label>${t('moveTarget')}</label><input type="text" class="mono" id="mv" value="${esc(rel)}"></div>`,
      actions: [{ label: t('cancel'), cls: 'hayalet' }, { label: t('moveTo'), cls: 'birincil', run: async (root) => {
        const target = $('#mv', root).value.trim().replace(/^\/+|\/+$/g, '');
        for (const it of items) {
          const name = it.path.split('/').pop();
          await api('POST', `/api/p/${enc(P)}/remote/rename?${q}&from=${enc(it.path)}&to=${enc((target ? target + '/' : '') + name)}`);
        }
        toast(t('moved')); render(true);
      } }],
    });
  }
  function renameItem(it) {
    modal({
      title: `${icon('edit')}${t('rename')}`,
      body: `<div class="alan"><label>${t('newPath')}</label><input type="text" class="mono" id="rn" value="${esc(it.path)}"></div>`,
      actions: [{ label: t('cancel'), cls: 'hayalet' }, { label: t('save'), cls: 'birincil', run: async (root) => {
        const to = $('#rn', root).value.trim().replace(/^\/+/, '');
        if (!to || to === it.path) return;
        await api('POST', `/api/p/${enc(P)}/remote/rename?${q}&from=${enc(it.path)}&to=${enc(to)}`);
        toast(t('moved')); render(true);
      } }],
    });
  }
  // Alt klasörlerde ara — CLI'daki gibi sonuçlar bulundukça akar (SSE); kapatınca arama durur
  function searchAll(term) {
    let es = null, n = 0, dirs = 0, done = false;
    const t0 = Date.now();
    const m = modal({
      title: `${icon('search')}${t('searchTitle', esc(term))}`, size: 'orta',
      body: `<div class="ara-durum" id="sr-st"><span>${icon('refresh', 'spin')}</span><span id="sr-txt">${t('searching')}</span></div>
        <div class="liste-kutu" id="sr-list" hidden></div><p class="faint" id="sr-none" hidden>${t('searchNone')}</p><p class="ipucu" id="sr-warn" hidden></p>`,
      actions: [{ label: t('close'), cls: 'birincil' }],
      onClose: () => { if (es) es.close(); },
    });
    const list = $('#sr-list', m.root), txt = $('#sr-txt', m.root);
    const status = () => {
      const sec = ((Date.now() - t0) / 1000).toFixed(1);
      txt.textContent = done ? t('searchDoneN', n, dirs, sec) : t('searchingN', n, dirs);
    };
    // Gelen sonuçlar kare başına toplu eklenir (binlerce sonuçta DOM tek tek şişmesin)
    let buf = [], raf = 0;
    const flushHits = () => {
      raf = 0;
      if (!buf.length) return;
      list.hidden = false;
      list.insertAdjacentHTML('beforeend', buf.map((h) => `<div class="tikla" data-hit="${esc(h.path)}" data-dir="${h.dir ? 1 : 0}">${icon(h.dir ? 'folder' : 'file', h.dir ? 'acc-t' : 'faint')} ${esc(h.path)}${h.dir ? '' : ` <span class="faint">${bytes(h.size)}</span>`}</div>`).join(''));
      buf = [];
      status();
    };
    list.onclick = (e) => {
      const d = e.target.closest('[data-hit]');
      if (!d) return;
      m.close();
      const p = d.dataset.hit;
      if (d.dataset.dir === '1') go(hashFor(p));
      else { const parent = p.split('/').slice(0, -1).join('/'); go(parent ? hashFor(parent) : pHash(P, 'files', srv)); previewFile(P, srv, p); }
    };
    es = new EventSource(`/api/p/${enc(P)}/remote/search?${q}&path=${enc(rel)}&q=${enc(term)}`);
    es.addEventListener('hit', (e) => { n++; buf.push(JSON.parse(e.data)); if (!raf) raf = requestAnimationFrame(flushHits); });
    es.addEventListener('progress', (e) => { dirs = JSON.parse(e.data).dirs; if (!raf) status(); });
    es.addEventListener('done', (e) => {
      const d = JSON.parse(e.data);
      es.close(); es = null; done = true;
      if (d.dirs) dirs = d.dirs;
      flushHits();
      $('#sr-st span', m.root).innerHTML = d.error ? icon('alert', 'bad-t') : icon('check', 'ok-t');
      if (d.error) { txt.textContent = d.error; txt.className = 'bad-t'; return; }
      status();
      if (!n) $('#sr-none', m.root).hidden = false;
      const w = $('#sr-warn', m.root);
      if (d.truncated) { w.hidden = false; w.textContent = t('searchTruncated'); }
      else if (d.warn) { w.hidden = false; w.textContent = d.warn; }
    });
    es.onerror = () => { if (es && !done) { es.close(); es = null; done = true; flushHits(); status(); } };
  }

  bind((root) => {
    $('#proj-aks', root).innerHTML = `<button class="ikon-btn" id="b-rf" title="${t('refresh')}">${icon('refresh')}</button>`;
    $('#b-rf', root).onclick = () => render(true);
    $('#f-srv', root).onchange = (e) => go(pHash(P, 'files', e.target.value));
    $$('[data-crumb]', root).forEach((b) => (b.onclick = () => go(b.dataset.crumb === '' ? pHash(P, 'files', srv) : hashFor(b.dataset.crumb))));
    const filter = $('#f-filter', root);
    filter.oninput = () => applyFilter(filter.value);
    filter.onkeydown = (e) => { if (e.key === 'Enter' && filter.value.trim()) searchAll(filter.value.trim()); };
    $('#f-search', root).onclick = () => { if (filter.value.trim()) searchAll(filter.value.trim()); else filter.focus(); };
    const up = $('#f-upload-input', root);
    $('#f-upload', root).onclick = () => up.click();
    up.onchange = async () => {
      if (!up.files.length) return;
      const fd = new FormData();
      [...up.files].forEach((f) => fd.append('files', f, f.name));
      const b = $('#f-upload', root); b.disabled = true; b.innerHTML = icon('refresh', 'spin') + t('upload');
      try {
        const r = await fetch(`/api/p/${enc(P)}/remote/upload?${q}&path=${enc(rel)}`, { method: 'POST', body: fd }).then((x) => x.json());
        if (!r.ok) throw new Error(r.error);
        toast(t('uploaded', r.data.length));
      } catch (err) { fail(err); }
      render(true);
    };
    // Ağaç: shell `tree` ile aynı çekirdek (iş olarak çalışır, çıktı canlı akar)
    $('#f-tree', root).onclick = () => {
      const m = modal({
        title: `${icon('layers')}${t('treeView')}`,
        body: `<p>${t('treeAsk')}</p><div class="aksiyonlar" style="flex-wrap:wrap">${[[20, '20'], [50, '50'], [100, '100'], [0, t('treeAll')]].map(([n, l]) => `<button class="btn" data-tmax="${n}">${l}</button>`).join('')}</div>`,
        actions: [{ label: t('cancel'), cls: 'hayalet' }],
      });
      $$('[data-tmax]', m.root).forEach((b) => (b.onclick = async () => {
        m.close();
        try {
          const ids = await startJobs(P, { kind: 'tree', servers: [srv], path: rel, max: +b.dataset.tmax }, t('treeView'));
          if (ids[0]) showJobLog(ids[0]);
        } catch (e) { fail(e); }
      }));
    };
    $('#f-mkdir', root).onclick = () => modal({
      title: `${icon('folder')}${t('newFolder')}`, body: `<div class="alan"><label>${t('folderName')}</label><input type="text" class="mono" id="nf"></div>`,
      actions: [{ label: t('cancel'), cls: 'hayalet' }, { label: t('create'), cls: 'birincil', run: async (r2) => {
        const n = $('#nf', r2).value.trim().replace(/^\/+|\/+$/g, '');
        if (!n) return false;
        await api('POST', `/api/p/${enc(P)}/remote/mkdir?${q}&path=${enc(full(n))}`);
        toast(t('created')); render(true);
      } }],
    });
    // Satırlar parça parça çizilir (5000 girdili klasör tek seferde ~5 MB HTML + 1 sn yerleşim
    // demekti); filtre / tümünü seç TÜM liste üzerinde çalışır, olaylar tbody'de delegasyonla.
    const bar = $('#f-selbar', root);
    const body = $('#f-body', root);
    const more = $('#f-more', root);
    const typeOf = new Map(entries.map((e) => [full(e.name), e.type]));
    let view = entries, shown = 0;
    const PAGE = 300;
    const renderMore = () => {
      if (!body) return;
      const next = view.slice(shown, shown + PAGE);
      body.insertAdjacentHTML('beforeend', next.map((e) => rowHTML(e, sel.has(full(e.name)))).join(''));
      shown += next.length;
      more.hidden = shown >= view.length;
      if (!more.hidden) more.textContent = t('moreRows', view.length - shown);
    };
    const applyFilter = (v) => {
      v = v.toLowerCase();
      view = v ? entries.filter((e) => e.name.toLowerCase().includes(v)) : entries;
      if (body) { body.innerHTML = ''; shown = 0; renderMore(); }
      refreshBar();
    };
    if (more) {
      more.onclick = renderMore;
      new IntersectionObserver((ents) => { if (ents.some((x) => x.isIntersecting) && document.body.contains(more)) renderMore(); }, { rootMargin: '600px' }).observe(more);
    }
    const items = () => [...sel].map((p) => ({ path: p, type: typeOf.get(p) }));
    const allBox = $('#f-all', root); // bağlantı hatası / boş klasörde tablo yok
    const refreshBar = () => {
      bar.hidden = !sel.size;
      $('[data-selcount]', bar).textContent = t('selectedN', sel.size);
      if (allBox) {
        const inView = view.filter((e) => sel.has(full(e.name))).length;
        allBox.checked = !!view.length && inView === view.length;
        allBox.indeterminate = inView > 0 && inView < view.length;
      }
    };
    if (allBox) allBox.onchange = () => {
      const on = allBox.checked;
      view.forEach((e) => (on ? sel.add(full(e.name)) : sel.delete(full(e.name))));
      $$('[data-entry] input[type=checkbox]', body).forEach((c) => (c.checked = on));
      refreshBar();
    };
    $('[data-sel="clear"]', bar).onclick = () => { sel.clear(); $$('[data-entry] input[type=checkbox]', root).forEach((c) => (c.checked = false)); refreshBar(); };
    $('[data-sel="del"]', bar).onclick = () => removeItems(items());
    $('[data-sel="move"]', bar).onclick = () => moveItems(items());
    $('[data-sel="freeze"]', bar).onclick = () => toggleFreeze(items());
    if (body) {
      body.onchange = (e) => {
        const tr = e.target.closest('[data-entry]');
        if (!tr || e.target.type !== 'checkbox') return;
        const p = full(tr.dataset.entry);
        if (e.target.checked) sel.add(p); else sel.delete(p);
        refreshBar();
      };
      body.onclick = (e) => {
        const tr = e.target.closest('[data-entry]');
        if (!tr) return;
        const p = full(tr.dataset.entry), typ = tr.dataset.type, it = { path: p, type: typ };
        const btn = e.target.closest('button');
        if (btn) {
          const act = btn.dataset.act;
          if (act === 'pv') previewFile(P, srv, p);
          else if (act === 'cmp') compareModal(P, srv, p);
          else if (act === 'dl') location.href = `/api/p/${enc(P)}/remote/get?${q}&path=${enc(p)}`;
          else if (act === 'ren') renameItem(it);
          else if (act === 'fz') toggleFreeze([it]);
          else if (act === 'rm') removeItems([it]);
          return;
        }
        if (e.target.closest('input,label')) return;
        if (typ === 'dir') go(hashFor(p)); else previewFile(P, srv, p);
      };
      renderMore();
    }
  });

  const frozenUnder = (p) => { const pre = p + '/'; for (const f of frozenSet) if (f.startsWith(pre)) return true; return false; };
  function rowHTML(e, checked) {
    const p = full(e.name);
    const fz = e.type === 'dir' ? frozenUnder(p) : frozenSet.has(p);
    return `<tr data-entry="${esc(e.name)}" data-type="${e.type}" class="tikla"><td class="sec-hucre"><label><input type="checkbox" ${checked ? 'checked' : ''}></label></td>
      <td><div class="ad-hucre">${icon(e.type === 'dir' ? 'folder' : 'file', e.type === 'dir' ? 'acc-t' : 'faint')}<span>${esc(e.name)}</span>${fz ? `<span class="pill" style="height:18px;font-size:11px">${icon('snow')}${t('frozenBadge')}</span>` : ''}</div></td>
      <td class="sag mono faint">${e.type === 'dir' ? '' : bytes(e.size)}</td><td class="mono faint">${e.time && !e.time.startsWith('0001') ? fmtDate(e.time) : ''}</td>
      <td class="sag" style="white-space:nowrap">${e.type === 'dir' ? '' : `<button class="ikon-btn" data-act="pv" title="${t('preview')}">${icon('eye')}</button><button class="ikon-btn" data-act="cmp" title="${t('compare')}">${icon('layers')}</button><button class="ikon-btn" data-act="dl" title="${t('download')}">${icon('download')}</button>`}
        <button class="ikon-btn ${fz ? 'acc-t' : ''}" data-act="fz" title="${t('freezeToggle')}">${icon('snow')}</button>
        <button class="ikon-btn" data-act="ren" title="${t('rename')}">${icon('edit')}</button>
        <button class="ikon-btn tehlike" data-act="rm" title="${t('delete')}">${icon('trash')}</button></td></tr>`;
  }
  return `<div class="arac-cubuk">
      <select id="f-srv">${cfg.servers.map((s) => `<option value="${esc(s.name)}" ${s.name === srv ? 'selected' : ''}>${esc(s.name)}</option>`).join('')}</select>
      <div class="ara">${icon('search')}<input type="text" id="f-filter" placeholder="${t('filter')}"></div>
      <button class="btn" id="f-search" title="${t('searchAll')}">${icon('search')}${t('searchAll')}</button>
      <button class="btn" id="f-tree" title="${t('treeHint')}">${icon('layers')}${t('treeView')}</button>
      <button class="btn" id="f-mkdir">${icon('folder')}${t('newFolder')}</button>
      <button class="btn mavi" id="f-upload">${icon('upload')}${t('upload')}</button><input type="file" id="f-upload-input" multiple hidden></div>
    <div class="yol-cubuk"><button data-crumb="">${esc(data.root || '/')}</button>${crumbs.map((c, i) => `<span>/</span><button data-crumb="${esc(crumbs.slice(0, i + 1).join('/'))}">${esc(c)}</button>`).join('')}</div>
    <div class="secim-cubuk" id="f-selbar" hidden><b data-selcount></b>
      <button class="btn kucuk" data-sel="move">${icon('folder')}${t('moveTo')}</button>
      <button class="btn kucuk" data-sel="freeze">${icon('snow')}${t('freezeToggle')}</button>
      <button class="btn kucuk tehlike" data-sel="del">${icon('trash')}${t('delete')}</button>
      <button class="btn kucuk hayalet" data-sel="clear">${t('clearSel')}</button></div>
    ${data.error ? `<div class="kart"><div class="not bad">${icon('alert')}${esc(data.error)}</div></div>` : !entries.length ? `<div class="bos" style="padding:30px">${t('emptyDir')}</div>` :
    `<div class="tablo-kap"><table class="tablo"><thead><tr><th class="sec-hucre"><label><input type="checkbox" id="f-all"></label></th><th>${t('colName')}</th><th class="sag">${t('colSize')}</th><th>${t('colDate')}</th><th></th></tr></thead><tbody id="f-body">
      </tbody></table></div><button class="btn hayalet tam-gen" id="f-more" hidden></button>`}`;
}

// Satır numaralı, renklendirilmiş kod görünümü
const isHTMLFile = (name) => /\.(html?|xhtml|svg)$/i.test(name || '');
function htmlFrame(text) {
  // sandbox="" → script çalışmaz, form gönderilmez, panelin çerezine/DOM'una erişemez
  return `<iframe class="html-onizle" sandbox="" referrerpolicy="no-referrer" srcdoc="${esc(text)}"></iframe>`;
}
// Sekmeli görünüm: [[anahtar, etiket, html], ...] → ilk sekme açık
function viewTabs(views) {
  if (views.length === 1) return views[0][2];
  return `<div class="tabs" style="margin-bottom:12px">${views.map(([k, l], i) => `<a href="javascript:void 0" data-vt="${k}" class="${i ? '' : 'aktif'}">${l}</a>`).join('')}</div>
    ${views.map(([k, , h], i) => `<div data-vtp="${k}" ${i ? 'hidden' : ''}>${h}</div>`).join('')}`;
}
function bindViewTabs(root) {
  $$('[data-vt]', root).forEach((a) => (a.onclick = () => {
    $$('[data-vt]', root).forEach((x) => x.classList.toggle('aktif', x === a));
    $$('[data-vtp]', root).forEach((p) => (p.hidden = p.dataset.vtp !== a.dataset.vt));
  }));
}

function codeView(text, name) {
  const rows = HL.hlLines(text, name);
  return `<div class="kod"><table>${rows.map((l, i) => `<tr><td class="no">${i + 1}</td><td class="l">${l || ' '}</td></tr>`).join('')}</table></div>`;
}

// Sunucu → yerel diff (sync'in sunucuya getireceği değişiklik), iki taraf da renklendirilmiş
function diffView(remote, local, name) {
  const ops = HL.diffLines(remote, local);
  if (!ops) return { html: `<p class="faint">${t('diffTooBig')}</p>${codeView(local, name)}`, add: 0, del: 0 };
  const R = HL.hlLines(remote, name), L = HL.hlLines(local, name);
  let add = 0, del = 0;
  const rows = HL.hunks(ops).map((o) => {
    if (o.fold) return `<tr class="fold"><td colspan="4">${t('foldN', o.fold)}</td></tr>`;
    if (o.op === '+') { add++; return `<tr class="add"><td class="no"></td><td class="no">${o.b}</td><td class="isaret">+</td><td class="l">${L[o.bi] || ' '}</td></tr>`; }
    if (o.op === '-') { del++; return `<tr class="del"><td class="no">${o.a}</td><td class="no"></td><td class="isaret">−</td><td class="l">${R[o.ai] || ' '}</td></tr>`; }
    return `<tr><td class="no">${o.a}</td><td class="no">${o.b}</td><td class="isaret"></td><td class="l">${L[o.bi] || ' '}</td></tr>`;
  }).join('');
  return { html: `<div class="kod" style="max-height:58vh"><table>${rows}</table></div>`, add, del };
}

async function compareModal(P, srv, rel) {
  const q = `server=${enc(srv)}&path=${enc(rel)}`;
  const lm = await loadModal(`${icon('file')}<span class="mono" style="font-size:13px">${esc(rel)}</span>`, 'genis', () => Promise.all([
    GET(`/api/p/${enc(P)}/local/cat?${q}`).catch(() => ({ exists: false })),
    GET(`/api/p/${enc(P)}/remote/cat?${q}`).then((d) => ({ exists: true, ...d })).catch(() => ({ exists: false })),
  ]));
  if (!lm) return;
  const [loc, rem] = lm.d;
  const binary = loc.binary || rem.binary;
  let views = {}, summary = '';
  if (binary) {
    views.diff = `<div class="bos">${t('binaryFile')}</div>`;
  } else if (!rem.exists && loc.exists) {
    views.diff = `<p class="ok-t">${t('diffRemoteMissing')}</p>${codeView(loc.content, rel)}`;
  } else if (!loc.exists) {
    views.diff = `<p class="bad-t">${t('diffLocalMissing')}</p>`;
  } else if (loc.content === rem.content) {
    views.diff = `<div class="bos" style="padding:28px">${icon('check')} ${t('diffSame')}</div>`;
  } else {
    const d = diffView(rem.content, loc.content, rel);
    views.diff = `<p class="ipucu" style="margin:0 0 10px">${t('diffHint')}</p>${d.html}`;
    summary = `<span class="fark-ozet"><span class="ok-t">+${d.add}</span><span class="bad-t">−${d.del}</span></span>`;
  }
  if (loc.exists && !loc.binary) views.local = codeView(loc.content, rel);
  if (rem.exists && !rem.binary) views.remote = codeView(rem.content, rel);
  if (isHTMLFile(rel)) {
    if (loc.exists && !loc.binary) views.localHtml = htmlFrame(loc.content);
    if (rem.exists && !rem.binary) views.remoteHtml = htmlFrame(rem.content);
  }
  const tabs = [['diff', t('diffTitle')], ['local', t('diffLocal')], ['remote', t('diffRemote')],
    ['localHtml', t('viewRenderedLocal')], ['remoteHtml', t('viewRenderedRemote')]].filter(([k]) => views[k]);
  const m = modal({
    title: `${icon('file')}<span class="mono" style="font-size:13px">${esc(rel)}</span> ${summary}`, size: 'genis',
    body: `<div class="tabs" style="margin-bottom:12px">${tabs.map(([k, l], i) => `<a href="javascript:void 0" data-cv="${k}" class="${i ? '' : 'aktif'}">${l}</a>`).join('')}</div>
      ${tabs.map(([k], i) => `<div data-cvp="${k}" ${i ? 'hidden' : ''}>${views[k]}</div>`).join('')}`,
    actions: [{ left: `<span class="mono">${esc(srv)}</span>` }, { label: t('close'), cls: 'birincil' }],
  });
  $$('[data-cv]', m.root).forEach((a) => (a.onclick = () => {
    $$('[data-cv]', m.root).forEach((x) => x.classList.toggle('aktif', x === a));
    $$('[data-cvp]', m.root).forEach((p) => (p.hidden = p.dataset.cvp !== a.dataset.cv));
  }));
}

async function previewFile(P, srv, p) {
  const lm = await loadModal(`${icon('file')}<span class="mono" style="font-size:13px">${esc(p)}</span>`, 'genis', () => GET(`/api/p/${enc(P)}/remote/cat?server=${enc(srv)}&path=${enc(p)}`));
  if (!lm) return;
  const d = lm.d;
  const pm = modal({
    title: `${icon('file')}<span class="mono" style="font-size:13px">${esc(p)}</span>`, size: 'genis',
    body: d.binary ? `<div class="bos">${t('binaryFile')}</div>`
      : viewTabs([['code', t('viewCode'), codeView(d.content, p)], ...(isHTMLFile(p) ? [['html', t('viewRendered'), htmlFrame(d.content)]] : [])]),
    actions: [{ left: `${bytes(d.size)} ${d.truncated ? t('truncated') : ''}` },
      ...(d.binary ? [] : [{ label: icon('layers') + t('compare'), cls: '', run: () => { compareModal(P, srv, p); } }]),
      { label: icon('download') + t('download'), cls: '', run: () => { location.href = `/api/p/${enc(P)}/remote/get?server=${enc(srv)}&path=${enc(p)}`; } },
      { label: t('close'), cls: 'birincil' }],
  });
  bindViewTabs(pm.root);
}

// ── Geçmiş (release + rollback + log) ────────────────────────────────────────
async function tabHistory(proj, srvFilter) {
  const P = proj.name;
  const q = srvFilter ? `?server=${enc(srvFilter)}` : '';
  const [rels, logs, cfg] = await Promise.all([GET(`/api/p/${enc(P)}/releases${q}`), GET(`/api/p/${enc(P)}/logs${q}`), getConfig(P)]);
  bind((root) => {
    $('#proj-aks', root).innerHTML = '';
    $('#h-srv', root).onchange = (e) => go(e.target.value ? pHash(P, 'history', e.target.value) : pHash(P, 'history'));
    $$('[data-rel]', root).forEach((b) => (b.onclick = () => releaseDetail(P, b.dataset.srv, b.dataset.rel)));
    $$('[data-undo]', root).forEach((b) => (b.onclick = () => rollbackFlow(P, b.dataset.srv, b.dataset.undo)));
    $$('[data-log]', root).forEach((tr) => (tr.onclick = async () => {
      const lm = await loadModal(`${icon('log')}<span class="mono" style="font-size:13px">${esc(tr.dataset.srv)} / ${esc(tr.dataset.log)}</span>`, 'genis', () => GET(`/api/p/${enc(P)}/logs/${enc(tr.dataset.srv)}/${enc(tr.dataset.log)}`));
      if (!lm || lm.d == null) return;
      const txt = lm.d;
      modal({ title: `${icon('log')}<span class="mono" style="font-size:13px">${esc(tr.dataset.srv)} / ${esc(tr.dataset.log)}</span>`, size: 'genis',
        body: `<div class="log">${txt.split('\n').map(logLine).join('')}</div>`, actions: [{ label: t('close'), cls: 'birincil' }] });
    }));
  });
  const relRows = rels.map((r) => {
    const status = r.rolled_back ? `<span class="pill">${t('rolledBack')}</span>` : r.has_backup ? `<span class="pill ok"><span class="nokta"></span>${t('restorable')}</span>` : `<span class="pill">${t('noBackupTag')}</span>`;
    return `<tr><td class="mono">${fmtDate(r.created_at)}</td><td>${esc(r.server)}</td><td class="sag">${r.files}</td><td class="sag">${r.backed}</td><td class="sag">${r.new}</td><td>${status}</td>
      <td class="sag" style="white-space:nowrap"><button class="btn kucuk hayalet" data-rel="${esc(r.id)}" data-srv="${esc(r.server)}">${t('details')}</button>
      ${r.has_backup && !r.rolled_back ? `<button class="btn kucuk tehlike" data-undo="${esc(r.id)}" data-srv="${esc(r.server)}">${icon('undo')}${t('undo')}</button>` : ''}</td></tr>`;
  }).join('');
  return `<div class="arac-cubuk"><select id="h-srv"><option value="">${t('all')}</option>${cfg.servers.map((s) => `<option value="${esc(s.name)}" ${s.name === srvFilter ? 'selected' : ''}>${esc(s.name)}</option>`).join('')}</select></div>
    <h2 style="margin-top:6px">${icon('undo')}${t('releases')}</h2><p class="muted" style="margin-top:-6px">${t('releasesSub')}</p>
    ${rels.length ? `<div class="tablo-kap"><table class="tablo"><thead><tr><th>${t('colWhen')}</th><th>${t('colServer')}</th><th class="sag">${t('colFiles')}</th><th class="sag">${t('colReplaced')}</th><th class="sag">${t('colNew')}</th><th>${t('colStatus')}</th><th></th></tr></thead><tbody>${relRows}</tbody></table></div>`
      : `<div class="bos" style="padding:30px">${t('noReleases')}</div>`}
    <h2>${icon('log')}${t('logs')}</h2>
    ${logs.length ? `<div class="tablo-kap"><table class="tablo"><thead><tr><th>${t('colWhen')}</th><th>${t('colServer')}</th><th>${t('colName')}</th><th class="sag">${t('colSize')}</th></tr></thead><tbody>
      ${logs.map((l) => `<tr class="tikla" data-log="${esc(l.file)}" data-srv="${esc(l.server)}"><td class="mono">${fmtDate(l.time)}</td><td>${esc(l.server)}</td><td class="mono faint">${esc(l.file)}</td><td class="sag mono faint">${bytes(l.size)}</td></tr>`).join('')}</tbody></table></div>`
      : `<div class="bos" style="padding:30px">${t('noLogs')}</div>`}`;
}

async function releaseDetail(P, srv, id) {
  const d = await GET(`/api/p/${enc(P)}/releases/${enc(srv)}/${enc(id)}`).catch(fail);
  if (!d) return;
  const m = d.manifest;
  const sets = { backed: new Set(m.backed || []), nw: new Set(m.new || []) };
  modal({
    title: `${icon('history')}${t('releaseTitle', esc(id))} · ${esc(srv)}`, size: 'orta',
    body: `<p class="mono faint">${fmtDate(m.created_at)} · ${t('colFiles')}: ${d.files.length} · ${t('colReplaced')}: ${sets.backed.size} · ${t('colNew')}: ${sets.nw.size}</p>
      <div class="liste-kutu">${d.files.map((f) => `<div><span class="${sets.nw.has(f) ? 'ok-t' : sets.backed.has(f) ? 'warn-t' : 'faint'}">${sets.nw.has(f) ? '+' : sets.backed.has(f) ? '~' : '·'}</span> ${esc(f)}</div>`).join('')}</div>`,
    actions: [{ label: t('close'), cls: 'birincil' }],
  });
}

async function rollbackFlow(P, srv, id) {
  const rels = await GET(`/api/p/${enc(P)}/releases?server=${enc(srv)}`).catch(fail);
  if (!rels) return;
  const plan = [];
  let blocked = '';
  for (const r of rels) { // en yeniden eskiye
    if (r.rolled_back) continue;
    if (!r.has_backup) { blocked = r.id; break; }
    plan.push(r);
    if (r.id === id) break;
  }
  if (blocked && !plan.some((r) => r.id === id)) return toast(t('rollbackBlocked', blocked), 'bad');
  modal({
    title: `${icon('undo')}${t('rollbackTitle', esc(srv))}`, size: 'orta',
    body: `<p>${t('rollbackBody', plan.length)}</p><div class="liste-kutu">${plan.map((r) => `<div><b>${esc(r.id)}</b> <span class="faint">· ${fmtDate(r.created_at)} · ~${r.backed} +${r.new}</span></div>`).join('')}</div>`,
    actions: [{ label: t('cancel'), cls: 'hayalet' }, { label: icon('undo') + t('rollbackGo'), cls: 'tehlike', run: async () => {
      const ids = await startJobs(P, { kind: 'rollback', servers: [srv], release: id }, t('rollback'));
      if (ids[0]) showJobLog(ids[0]);
    } }],
  });
}

// ═══ Sayfa: işler ═════════════════════════════════════════════════════════════
async function pageJobs() {
  const [jobs, watches] = await Promise.all([GET('/api/jobs'), GET('/api/watches')]);
  bind((root) => {
    $$('[data-jid]', root).forEach((tr) => (tr.onclick = () => showJobLog(tr.dataset.jid)));
    $$('[data-wstop]', root).forEach((b) => (b.onclick = async (e) => {
      e.stopPropagation();
      await api('DELETE', `/api/p/${enc(b.dataset.wstop)}/watch`).catch(fail);
      toast(t('watchStopped')); await loadProjects(); render(true);
    }));
  });
  const st = (s) => s === 'running' ? `<span class="pill busy"><span class="nokta"></span>${t('stRunning')}</span>` : s === 'error' ? `<span class="pill bad"><span class="nokta"></span>${t('stError')}</span>` : `<span class="pill ok"><span class="nokta"></span>${t('stDone')}</span>`;
  return `<div class="sayfa"><div class="baslik-satir"><div><h1>${t('jobsTitle')}</h1><p>${t('jobsSub')}</p></div></div>
    ${watches.length ? `<h2 style="margin-top:0">${icon('eye')}${t('activeWatches')}</h2><div class="kartlar" style="margin-bottom:24px">${watches.map((w) => `<div class="kart"><div><div class="ust"><h3>${esc(w.project)}</h3><span class="pill busy"><span class="nokta"></span>watch</span></div>
      <div class="meta"><span>${icon('server')}${esc(w.servers.join(', '))}</span><span>${icon('clock')}${t('watchSince', fmtDate(w.started).slice(11), w.syncs)}</span>${w.last_sync ? `<span>${icon('upload')}${esc(w.last_sync)}</span>` : ''}</div></div>
      <div class="butonlar"><button class="btn kucuk" data-jid="${esc(w.job_id)}">${icon('log')}${t('jobLog')}</button><button class="btn kucuk tehlike" data-wstop="${esc(w.project)}">${icon('stop')}${t('stop')}</button></div></div>`).join('')}</div>` : ''}
    ${jobs.length ? `<div class="tablo-kap"><table class="tablo"><thead><tr><th>${t('colJob')}</th><th>${t('colProject')}</th><th>${t('colServer')}</th><th>${t('colStatus')}</th><th>${t('colStarted')}</th><th class="sag">${t('colDuration')}</th></tr></thead><tbody>
      ${jobs.map((j) => `<tr class="tikla" data-jid="${esc(j.id)}"><td><b>${esc(j.title)}</b></td><td>${esc(j.project)}</td><td class="muted">${esc(j.server)}</td><td>${st(j.status)}${j.error ? ` <span class="bad-t" style="font-size:12px">${esc(j.error.slice(0, 80))}</span>` : ''}</td>
        <td class="mono faint">${fmtDate(j.started)}</td><td class="sag mono faint">${j.ended ? dur(new Date(j.ended) - new Date(j.started)) : ''}</td></tr>`).join('')}</tbody></table></div>`
      : `<div class="bos">${t('noJobs')}</div>`}</div>`;
}

// ═══ Başlangıç ════════════════════════════════════════════════════════════════
async function boot() {
  try {
    S.meta = await GET('/api/meta');
    LANG = S.meta.lang === 'en' ? 'en' : 'tr';
    if (S.meta.base) document.title = 'syncFTP — ' + S.meta.base;
  } catch { /* oturum yoksa sayfa zaten 401 */ }
  applyLang();
  $('#sb-ver').textContent = S.meta ? 'v' + S.meta.version : '';
  $$('#dil-sec button').forEach((b) => (b.onclick = async () => {
    try { await api('PUT', '/api/lang', { lang: b.dataset.l }); LANG = b.dataset.l; applyLang(); render(true); } catch (e) { fail(e); }
  }));
  $('#btn-proje-ekle').onclick = addProjectFlow;
  $('#nav-ai').onclick = () => aiToggle();
  $('#btn-kapat').onclick = async () => {
    if (!(await confirmBox(t('quit'), esc(t('quitConfirm')), t('quit'), true))) return;
    await api('POST', '/api/quit').catch(() => {});
    document.body.innerHTML = `<div style="display:grid;place-items:center;height:100vh;color:var(--muted)">${esc(t('quitDone'))}</div>`;
  };
  await loadProjects();
  await refreshJobsMeta();
  render();

  // Varlık bağlantısı: pencere açık olduğu sürece süreç yaşar (küçültülünce de kopmaz)
  const presence = new EventSource('/api/presence');
  presence.addEventListener('bye', () => { presence.close(); window.close(); });
  presence.addEventListener('navigate', (e) => { if (e.data && e.data.startsWith('#/')) location.hash = e.data; });

  // Yedek heartbeat + bağlantı kopma uyarısı
  let lost = 0;
  setInterval(async () => {
    try {
      const ping = await GET('/api/ping');
      if (lost) { lost = 0; }
      // Sunucu tarafında başlayan işler (webhook deploy'ları, watch sync'leri) da dock'ta canlı görünsün
      const known = [...S.jobs.values()].filter((j) => j.status === 'running').length;
      if (ping.jobs > known) refreshJobsMeta();
    } catch { if (++lost === 2) toast(t('sessionLost'), 'bad'); }
  }, 5000);
  setInterval(loadProjects, 20000);
}
// pickers.js / cicd.js app.js'ten sonra yüklenir → boot tüm script'ler hazır olunca
window.addEventListener('DOMContentLoaded', boot);

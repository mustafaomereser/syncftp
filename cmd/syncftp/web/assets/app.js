'use strict';

/* ═══════════════════════════════════════════════════════════════════════════
   syncFTP panel — vanilla JS, bağımlılıksız.
   Router: #/ · #/jobs · #/p/<proje>[/<sekme>[/<alt>]]
   ═══════════════════════════════════════════════════════════════════════════ */

// ── i18n ─────────────────────────────────────────────────────────────────────
const T = {
  tr: {
    navAll: 'Tüm projeler', navJobs: 'İşler', navProjects: 'Projeler',
    addProject: 'Proje ekle', navNoProjects: 'Bu klasörde proje yok', watching: 'Watch', language: 'Dil', quit: 'Kapat',
    loading: 'Yükleniyor…', shortcutKey: 'Kısayol: {0}', skip_ignored: 'ignore', skip_excluded: 'exclude', skip_not_included: 'include dışı', skipHint_ignored: 'Yerelde ignore dosyalarında (.gitignore / syncftp.ignore): sync bu yolu taramaz, sunucuya yüklemez. Sunucudaki kopyaya dokunulmaz, silinmez.', skipHint_excluded: 'Exclude yerel taraf içindir: yereldeki bu dosya sunucuya yüklenmez. Sunucudaki kopyaya dokunulmaz, silinmez.', skipHint_not_included: 'Include listesinde değil: yereldeki bu dosya sunucuya yüklenmez. Sunucudaki kopyaya dokunulmaz, silinmez.', watchAll: 'Tümünü izle', watchStopAll: 'Tüm watch\'ları durdur', watchAllHint: 'Projedeki her aktif profil için ayrı watch başlatır', watchServerHint: 'Bu profili izle: yerelde değişen dosyalar otomatik yüklenir (CLI: watch -s)', watchingTag: 'izleniyor', watchCount: '{0}/{1} profil', watchSyncs: '{0} otomatik sync', mgSync: 'Sync', mgCheck: 'Kontrol', mgHistory: 'Zaman ve geri alma', mgServer: 'Profil', showMore: '{0} tane daha göster', foldAdvanced: 'Gelişmiş', foldMore: 'Diğer', foldDefault: 'varsayılan', foldOff: 'kapalı', foldConn: 'bağlantı', foldRetry: 'deneme', foldBlock: 'engelli uzantı', setTabGeneral: 'Genel', setTabLibrary: 'Bağlantı kütüphanesi', setTabDefaults: 'Yeni proje varsayılanları', setTabProjects: 'Projeler', setTabHooks: 'Webhook / CI', setTabSecurity: 'Host anahtarları', setTabSystem: 'Sistem', deleteRecursive: 'İçeriğiyle birlikte sil', deleteRecursiveHint: 'Kapalıyken yalnızca boş klasörler silinir (CLI: remote rm --recursive)', stFilter: 'Filtrele', stFilterHint: 'Yalnızca bu liste için — config değişmez (CLI: status --include / --exclude)', moreRows: '{0} öğe daha göster', stFrozen: 'Freeze', stProtected: 'Korunan', stBlocked: 'Engelli', stUploadSize: 'yüklenecek: {0}', treeView: 'Ağaç', treeHint: 'Bu klasörün altını ağaç olarak göster (shell tree)', treeAsk: 'Klasör başına en fazla kaç öğe gösterilsin?', treeAll: 'Tümü', ignoreNoFile: 'ignore_files listesinde syncFTP dosyası yok — yalnızca .gitignore okunuyor', viewCode: 'Kod', fmtRaw: 'Orijinal', fmtPretty: 'Biçimlendir', fmtMinified: 'Dosya sıkıştırılmış (minify) — okunur hâli gösteriliyor; sunucudaki dosya değişmez', fmtHint: 'Yalnızca görünüm: sunucudaki ya da yereldeki dosya değişmez', fmtFailed: 'Biçimlendirilemedi: {0}', viewRendered: 'Önizleme', viewRenderedLocal: 'Önizleme (yerel)', viewRenderedRemote: 'Önizleme (sunucu)', scanningLocal: 'Taranıyor…', refresh: 'Yenile', save: 'Kaydet', cancel: 'Vazgeç', close: 'Kapat', delete: 'Sil',
    edit: 'Düzenle', back: 'Geri', yes: 'Evet', confirm: 'Onayla', browse: 'Gözat', add: 'Ekle', open: 'Aç',
    saved: 'Kaydedildi', copied: 'Kopyalandı', never: 'hiç', none: 'yok', all: 'Tümü', manual: '(manuel giriş)',
    projectsTitle: 'Projeler', projectsSub: `{0} klasöründeki syncFTP projeleri (CLI gibi: panel açıldığı klasörü esas alır)`,
    noProjects: 'Henüz proje yok', noProjectsSub: 'Bu klasörde (ya da alt klasörlerinde) .syncftp/ yok. Buraya proje kurun; terminalde `syncftp init` çalıştırdığınız alt klasörler de burada kendiliğinden görünür.',
    serversN: '{0} profil', enabledN: '{0} aktif', configError: 'Config okunamadı',
    addProjectTitle: 'Proje ekle', folder: 'Proje klasörü', folderHint: 'Panelin açıldığı klasör ya da altındaki bir klasör (.syncftp/ yoksa oluşturulur)',
    displayName: 'Görünen ad', displayNameHint: 'Boş bırakılırsa klasör adı kullanılır',
    createConfigTitle: 'Config oluşturulsun mu?', createConfigBody: 'Bu klasörde syncFTP config\'i yok. Boş bir config oluşturup projeyi ekleyeyim mi? Ardından sunucu ekleyebilirsiniz.',
    create: 'Oluştur',
    tabOverview: 'Genel bakış', tabServers: 'Profiller', tabConnections: 'Bağlantılar', tabSettings: 'Proje ayarları',
    tabFiles: 'Dosyalar', tabHistory: 'Geçmiş',
    sumServers: 'Profiller', sumPending: 'Bekleyen değişiklik', sumLastSync: 'Son sync', sumWatch: 'Watch',
    syncAll: 'Tümünü sync et', dryRunAll: 'Dry-run', more: 'Diğer',
    calibrateAll: 'Tümünü calibrate et', calibrateHint: 'Sunucudaki dosyaları boyutla eşleştir, state\'i kur (yükleme yok)',
    verifyAll: 'Tümünü doğrula', verifyHint: 'Sunucuda elle değiştirilen/silinen dosyaları bul',
    detectAll: 'Bağlantı limitlerini ölç', detectHint: 'Sunucunun kabul ettiği en fazla bağlantıyı bulup max_connections\'a yazar',
    watchStart: 'Watch başlat', watchStop: 'Watch durdur', watchOff: 'Kapalı',
    watchSince: '{0} beri · {1} sync', watchStarted: 'Watch başladı', watchStopped: 'Watch durduruldu', noWatches: 'Çalışan watch yok',
    noServers: 'Bu projede profil yok', noServersSub: 'İlk profili ekleyin: yerel klasörün sunucuda nereye yükleneceğini belirler — FTP, FTPS veya SFTP; isterseniz SSH tüneli üzerinden.',
    addServer: 'Profil ekle', disabled: 'pasif', busy: 'çalışıyor', tunnel: 'tünel',
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
    filesConnTip: 'Bu bağlantıyla gez: {0}', filesConnHint: 'Bağlantıyla geziyorsun: kök dizinden başlar, sync\'e bağlı değildir. Freeze, yerelle karşılaştırma ve düzenleme yalnızca profillerde.', statusTitle: '{0} — bekleyen değişiklikler', stViewList: 'Liste', stViewTree: 'Klasör', stViewHint: 'Liste ya da klasör ağacı (ls gibi) olarak göster', newFiles: 'Yeni', changedFiles: 'Değişen', deletedFiles: 'Yerelde silinen',
    deletedNote: 'Yerelde silinen dosyalar sunucudan silinmez, yalnızca raporlanır.', emptyList: 'Liste boş',
    milestoneTitle: '{0} — milestone', milestoneBody: 'Milestone, bu tarihten önce değişmiş dosyaları sync dışında bırakır (kalıcı filtre). "Milestone sync" yalnızca tarihten sonra değişenleri yükler.',
    milestoneCurrent: 'Mevcut: {0}', milestoneDate: 'Tarih', milestonePh: 'bugün', milestoneDateHint: 'boş = şimdi · bugün · dün · 3d · 5h · 14:30 · 20.07 · 20.07.2026 15:04 · 2026-07-20',
    milestoneSet: 'Ayarla', milestoneClear: 'Kaldır', milestoneSaved: 'Milestone: {0}', milestoneCleared: 'Milestone kaldırıldı',
    freezeTitle: '{0} — freeze listesi', freezeBody: 'Her satıra bir dosya yolu (yerel dizine göreceli). Bu dosyalar değişse bile yüklenmez.',
    freezeSaved: 'Freeze listesi kaydedildi ({0} kayıt)',
    colName: 'Ad', colProtocol: 'Protokol', colHost: 'Host', colRemote: 'Uzak dizin', colLocal: 'Yerel dizin', colEnabled: 'Aktif',
    duplicate: 'Kopyala', deleteServerConfirm: '"{0}" profili config\'ten silinsin mi? (State ve yedekler diskte kalır.)',
    serverNew: 'Yeni profil', serverEdit: 'Profili düzenle',
    secGeneral: 'Genel', secConnection: 'Bağlantı', secConnectionDesc: 'Bir bağlantı seçin ya da bilgileri elle girin.',
    secTunnel: 'SSH tüneli', secTunnelDesc: 'Sunucuya doğrudan erişilemiyorsa bağlantı bu SSH sunucusu (jump host) üzerinden açılır. Boş = tünel yok.',
    secPaths: 'Dizinler', secPerf: 'Performans', secFilters: 'Filtreler', secFiltersDesc: 'Her satıra bir yol. Profilin include\'u globali geçersiz kılar; exclude/protect globale eklenir.',
    secUpload: 'Yükleme', secHooks: 'Hook\'lar', secHooksDesc: 'Her satıra bir komut, proje dizininde sırayla çalışır. Global hook\'lardan sonra çalışırlar.',
    fName: 'Ad', fEnabled: 'Aktif', fEnabledHint: 'Pasif profiller "tümünü sync et"e dahil olmaz',
    fConnection: 'Bağlantı', fProtocol: 'Protokol', fHost: 'Host', fPort: 'Port', fUser: 'Kullanıcı', fPassword: 'Şifre',
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
    lockedByProfile: 'Bilgiler "{0}" bağlantısından geliyor.', editProfile: 'Bağlantıyı düzenle',
    tunnelByProfile: 'Tünel ayarı da "{0}" bağlantısından geliyor.',
    connectionsSub: 'Aynı FTP/SFTP hesabını kullanan profiller bilgileri tek yerden paylaşır.',
    noConnections: 'Bağlantı yok', noConnectionsSub: 'Birden fazla profil aynı hesaba bağlanıyorsa bilgileri bir bağlantıda toplayın.',
    addConnection: 'Bağlantı ekle', connNew: 'Yeni bağlantı', connEdit: 'Bağlantıyı düzenle', usedBy: 'Kullanan',
    connInUse: '"{0}" bağlantısını kullanan profiller var: {1}', deleteConnConfirm: '"{0}" bağlantısı silinsin mi?',
    secProject: 'Proje', fProjectName: 'Proje adı (config)', fDefaultPath: 'Varsayılan yerel dizin', fDefaultPathHint: 'Kendi yerel dizini olmayan profiller buradan okur',
    secGlobal: 'Global sync ayarları', secGlobalDesc: 'Tüm profillere uygulanır; profil ayarları bunları genişletir.',
    fIgnoreFiles: 'Ignore dosyaları', fIgnoreFilesHint: 'Boş = .gitignore + .syncftp/syncftp.ignore',
    fFirstFull: 'İlk sync\'te her şeyi yükle', fFirstFullHint: 'Kapalıyken ilk sync calibrate ile mevcut dosyaları eşleştirir',
    fBackupKeep: 'Tutulacak yedek', fBackupKeepHint: 'Son kaç sync\'in yedeği saklansın (rollback derinliği)',
    secSecurity: 'Güvenlik', keyringTitle: 'Şifreleri sistem anahtarlığında sakla', keyringHint: 'Windows Credential Manager / macOS Keychain. syncftp.json\'da yalnızca "keyring:" işareti kalır.',
    keyringOn: 'Şifreler anahtarlığa taşındı', keyringOff: 'Şifreler tekrar config\'te', secretErrors: 'Bazı şifreler anahtarlıktan okunamadı',
    secRegistry: 'Proje kaydı', registryDesc: 'Panelde görünen ad ve kayıt. Kayıttan çıkarmak dosyalara dokunmaz.',
    renameProject: 'Yeniden adlandır', 
    filter: 'Bu klasörde filtrele…', colSize: 'Boyut', colDate: 'Tarih',
    emptyDir: 'Klasör boş', preview: 'Önizle', download: 'İndir', deleteRemoteConfirm: '"{0}" sunucudan KALICI olarak silinsin mi?', deleteDirConfirm: '"{0}" klasörü ve İÇİNDEKİ HER ŞEY sunucudan kalıcı olarak silinsin mi?',
    deleted: 'Silindi', binaryFile: 'Binary dosya — önizleme yok', viewMedia: 'Görüntü', mediaFailed: 'Bu dosya tarayıcıda gösterilemiyor', mediaSameSize: 'Yerel ve sunucudaki dosyanın boyutu aynı', mediaNoLocal: 'Yerelde yok', mediaNoRemote: 'Sunucuda yok', dropHere: 'Yüklemek için bırak', dropInto: '{0} klasörüne yüklemek için bırak', dropHint: 'Dosya / klasörleri sürükleyip buraya bırakabilirsin', truncated: '(ilk 256 KB)',
    releases: 'Sync geçmişi (release)', releasesSub: 'Her sync, değiştirdiği dosyaların eski halini yedekler. Bir release\'i geri almak onu ve ondan sonrakileri geri alır.',
    colWhen: 'Zaman', colServer: 'Profil', colFiles: 'Dosya', colReplaced: 'Değişen', colNew: 'Yeni', colStatus: 'Durum',
    rolledBack: 'geri alındı', noBackupTag: 'yedek yok', restorable: 'geri alınabilir', details: 'Ayrıntı', undo: 'Geri al',
    noReleases: 'Henüz release yok', rollbackTitle: 'Geri alma — {0}',
    rollbackBody: 'Aşağıdaki {0} release EN YENİDEN başlayarak geri alınacak: değişen dosyaların eski hali geri yüklenir. Yeni eklenen dosyalar, aşağıdaki kutuyu işaretlemezseniz sunucuda kalır.',
    rollbackGo: 'Geri al',
    logs: 'Loglar', noLogs: 'Log yok', releaseTitle: 'Release {0}',
    relPickFile: 'Farkını görmek için soldan bir dosya seçin', relSelected: '{0} dosya seçili', relNotRestorable: 'Bu release geri alınamaz (yedeği yok ya da geri alınmış)', relRollbackSel: 'Seçilenleri geri al ({0})', relRollbackAll: 'Tümünü geri al', relDelete: 'Sil', relDeleteAll: 'Tüm yedekleri sil', relDeleteTitle: '{0} release silinsin mi?', relDeleteTitleAll: '{0} — tüm release\'ler silinsin mi?', relDeleteBody: 'Kayıt ve yedek dosyaları silinir; bu release\'ler geçmişten düşer ve geri alınamaz. <b>Sunucudaki dosyalara dokunulmaz.</b>', relDeleteKeep: 'Yalnızca yedek dosyalarını sil — release geçmişte kalsın', relDeletedN: '{0} release silindi', relPurgedN: '{0} release\'in yedeği silindi', relPartly: 'kısmen geri alındı ({0} dosya)', relRolledFile: 'Bu dosya geri alındı', relNoBackupFile: 'Yedek yok — geri alınamaz', noFiles: 'Dosya yok',
    rbSelBody: '{0} dosya {1} release\'inden önceki hâline döner.', rbDeleteNew: 'Sonradan eklenen {0} dosyayı da sunucudan sil', rbDeleteNewHint: 'Kapalıyken bu release\'lerin yeni eklediği dosyalar sunucuda kalır; sonra ağaçtan seçip silebilirsiniz.', rbSelChain: 'Bu dosyalar daha yeni release\'lerde de değiştiği için onlarda da bu dosyalar geri alınır:',
    relBefore: 'Önceki hal', relAfter: 'Yüklenen hal', relDiffHint: 'Sunucudaki önceki hal → bu sync\'te yüklenen hal', relFileHint: 'Farkı görmek için dosyaya tıklayın',
    relNoBefore: 'Önceki hal yok — yedek alınamamış ya da eski olduğu için silinmiş.', relNoAfter: 'Yüklenen hal saklanmamış ve yerel dosya o zamandan beri değişmiş.', relNewFile: 'Bu sync\'te sunucuya yeni eklenen dosya', relAfterLocal: 'Yüklenen hal yerel dosyadan (hash aynı)',
    jobsTitle: 'İşler', jobsSub: 'Bu oturumda çalışan ve tamamlanan işler', noJobs: 'Henüz iş yok', activeWatches: 'Aktif watch\'lar',
    colJob: 'İş', colProject: 'Proje', colStarted: 'Başladı', colDuration: 'Süre',
    stRunning: 'çalışıyor', stDone: 'tamam', stStopped: 'durduruldu', stError: 'hata', stop: 'Durdur', jobStopping: 'Durduruluyor…',
    jobDone: '{0} tamamlandı', jobFailed: '{0}: {1}', jobLog: 'İş günlüğü', resultLine: '↑ {0} yüklendi · {1} hata · {2}',
    quitConfirm: 'syncFTP tamamen kapatılsın mı? Çalışan watch\'lar ve webhook dinleyicisi de durur. (Pencereyi kapatırsanız tray\'de çalışmaya devam eder.)', quitDone: 'Panel kapandı — bu pencereyi kapatabilirsiniz.',
    sessionLost: 'Panel bağlantısı koptu', ago: '{0} önce', justNow: 'az önce', min: 'dk', hour: 'sa', day: 'gün',
    copyPath: 'Yolu kopyala', notFound: 'Bulunamadı', minutes: '{0} dk', seconds: '{0} sn',
    compare: 'Yerelle karşılaştır', diffTitle: 'Fark', diffLocal: 'Yerel', diffRemote: 'Sunucu', diffSame: 'Yerel ve sunucu sürümü aynı',
    diffRemoteMissing: 'Sunucuda yok — sync bu dosyayı yeni olarak yükler',
    secRemote: 'SSH komutları (sunucuda)', secRemoteDesc: 'Yalnızca SFTP sunucularında: SSH ile uzak dizinde çalışır. pre hata verirse sync iptal olur; post deploy sonrası çalışır (migrate, cache, restart…).', fRemotePre: 'remote_pre_sync', fRemotePost: 'remote_post_sync',
    whTest: 'Test et', whTestOk: 'Webhook gönderildi: {0}', ignoreTitle: 'Ignore kuralları (.syncftp/syncftp.ignore)', ignoreDesc: 'Gitignore sözdizimi. .gitignore ile birleştirilir; ignore edilen klasörlere yerelde ve sunucuda hiç girilmez.',
    ignoreTemplate: 'Şablon ekle…', ignoreSaved: 'Ignore kuralları kaydedildi', gitignoreNote: '.gitignore ({0} satır) de uygulanıyor',
    failedTitle: '{0} — başarısız dosyalar', failedBody: 'Son sync\'te yüklenemeyen dosyalar. Tekrar denemek yalnızca bunları yükler.',
    compareServers: 'Profilleri karşılaştır', compareServersHint: 'İki profilin uzak içeriğini karşılaştırır (sadece birinde olan / boyutu farklı)', compareA: 'Profil A', compareB: 'Profil B', compareGo: 'Karşılaştır',
    upload: 'Yükle', newFolder: 'Yeni klasör', folderName: 'Klasör adı', rename: 'Yeniden adlandır / taşı', newPath: 'Yeni yol (uzak dizine göreceli)',
    moveTo: 'Taşı', selectedN: '{0} seçili', clearSel: 'Seçimi temizle',
    freezeToggle: 'Freeze / unfreeze', frozenBadge: 'frozen', searchAll: 'Alt klasörlerde ara', searchTitle: '"{0}" araması', searchNone: 'Sonuç yok', searching: 'Aranıyor…', searchingN: '{0} sonuç · {1} klasör tarandı…', searchDoneN: '{0} sonuç · {1} klasör · {2} sn', searchTruncated: 'Çok fazla sonuç (ilk 2000) ya da 5 dakika sınırı doldu — aramayı daraltın.',
    uploaded: '{0} dosya yüklendi', pkTitle: 'Hızlı arama', pkPlaceholder: 'Proje, sayfa ya da eylem ara…', pkPlaceholderShort: 'Ara…', pkGroupPages: 'Sayfalar', pkGroupProjects: 'Projeler', pkGroupActions: 'Eylemler', pkEmpty: 'Sonuç yok', pkMove: 'gez', pkOpen: 'aç', tbJobsN: '{0} iş çalışıyor', tbMin: 'Simge durumuna küçült', tbMax: 'Ekranı kapla', tbRestore: 'Önceki boyut', tbClose: 'Kapat', cmRename: 'Yeniden adlandır', cmNoEnabled: 'Aktif profil yok', cmBack: 'Geri', cmForward: 'İleri', cmJobs: 'İşler', cmSettings: 'Ayarlar', cmHelp: 'Yardım', cmPage: 'Sayfa', cmCopy: 'Kopyala', cmCopied: '{0} kopyalandı', cmCopyFail: 'Kopyalanamadı; elle seçip Ctrl+C ile kopyala', cmText: 'Metin', cmSelection: 'Seçili metin', cmBrowserNote: 'Tarayıcı menüsü için <kbd>Shift</kbd> + sağ tık', cmOpen: 'Aç', cmUploadHere: 'Bu klasöre yükle…', cmCopyMenu: 'Kopyala', cmCopyName: 'Adı', cmCopyRel: 'Göreceli yolu', cmCopyRemote: 'Sunucudaki tam yolu', cmName: 'Ad', cmPath: 'Yol', cmSelect: 'Seç', cmUnselect: 'Seçimi kaldır', cmSelectAll: 'Tümünü seç', cmDeleteN: '{0} öğeyi sil', cmUp: 'Üst klasöre çık', cmProjectSettings: 'Proje ayarları', cmTerminal: 'Terminalde aç', cmCopyFolder: 'Proje klasörünün yolunu kopyala', cmJobLog: 'Logu aç', cmNotStoppable: 'Bu iş durdurulamıyor', secTriggers: 'Tetikleyiciler', secTriggersDesc: 'Belirli dosyalar yüklenince (ya da yüklenmeden hemen önce) çalışır: örneğin migrations klasörü değişince migrate. Proje ayarlarındaki tetikleyiciler bu profilde de çalışır.', secTriggersDescGlobal: 'Tüm profillerde çalışır; profile özel olanlar profil düzenleme sayfasında. Belirli dosyalar yüklenince (ya da hemen önce) komut, SSH, HTTP isteği ya da bildirim.', trAfter: 'Yüklemeden sonra', trBefore: 'Yüklemeden önce', trWhen: 'Ne zaman', trWhenHint: '"Önce" hata verirse o profilin sync\'i iptal olur; "sonra" hatası sync sonucunu bozmaz.', trPaths: 'Hangi dosyalar', trPathsHint: 'Bu kalıplardan birine uyan dosya yüklenirse çalışır (include/exclude ile aynı kurallar). Boş = her yüklemede.', trRun: 'Yerel komutlar', trRunHint: 'Proje klasöründe çalışır; SYNCFTP_TRIGGER_FILES ortamda.', trRemote: 'SSH komutları (sunucuda)', trRemoteHint: 'Uzak dizinde çalışır (php artisan migrate --force).', trRemoteFtp: 'Yalnızca SFTP profillerinde çalışır; FTP\'de atlanır.', trHttpUrl: 'HTTP isteği — adres', trHttpMethod: 'Yöntem', trHttpHeaders: 'Başlıklar', trHttpBody: 'Gövde', trHttpBodyHint: 'Boşsa JSON özet gider (trigger, server, project, count, files).', trVarsHint: 'Yer tutucular: {files} {files_json} {count} {first} {server} {project} {trigger}', trNotify: 'Bildirim (Slack / Discord / webhook)', trActRun: 'komut', trActNotify: 'bildirim', trNoAction: 'eylem yok', trUnnamed: '(adsız)', trAnyFile: 'her dosya', trEmpty: 'Henüz tetikleyici yok. Aşağıdan hazır bir şablonla ekle.', trAdd: 'Tetikleyici ekle', trDelConfirm: '"{0}" tetikleyicisi silinsin mi? (Kaydet\'e basınca kalıcı olur)', trTest: 'Dene', trTestHint: 'Sync yapmadan bu tetikleyicinin eylemlerini çalıştır (formdaki hâliyle)', trTestServer: 'Profil', trTestFiles: 'Örnek dosyalar', trTestNote: 'Eylemler gerçekten çalışır (komut, SSH, HTTP, bildirim). Dosyalar yalnızca ortama ve yer tutuculara geçer, hiçbir şey yüklenmez.', trTestNoServer: 'Önce profili kaydet', trTpl_migrateSsh: 'Şablon: Laravel migrate (SSH)', trTpl_migrateHttp: 'Şablon: migrate (HTTP isteği)', trTpl_composer: 'Şablon: composer install (SSH)', trTpl_cache: 'Şablon: cache temizle (SSH)', trTpl_npm: 'Şablon: önce npm run build', trTpl_notify: 'Şablon: Slack / Discord bildirimi', trTpl_blank: 'Boş tetikleyici', edDirect: 'Doğrudan sunucuya · yedeksiz', edDirectHint: 'Bağlantıyla açıldı: yerel kopya ve yedek (Geri al) yok, kayıt doğrudan sunucudaki dosyaya yazılır. Sunucudaki dosya açıldıktan sonra değiştiyse yine sorulur.', zipSel: 'Zip\'le', zipTitle: 'Zip arşivi oluştur', zipName: 'Arşiv adı', zipHint: '{0} öğe bu klasörde tek arşive sıkıştırılır; aynı adlı dosya varsa sonuna -2 eklenir. FTP\'de sunucuda komut çalıştırılamadığı için dosyalar indirilip arşivlenir, arşiv geri yüklenir.', unzip: 'Zipten çıkar', unzipTo: '{0}/ klasörüne', unzipHere: 'Bulunduğu klasöre', unzipHint: 'Var olan dosyaların üzerine yazılır. Arşiv indirilir, içindeki dosyalar tek tek yüklenir.', upWaiting: 'sırada…', upDone: 'yüklendi', upCount: '{0} / {1} dosya', created: 'Oluşturuldu', moved: 'Taşındı', frozenAdded: '{0} öğe freeze listesine eklendi', frozenRemoved: '{0} kayıt freeze listesinden çıkarıldı',
    deleteManyConfirm: '{0} öğe sunucudan KALICI olarak silinsin mi? (klasörler içerikleriyle)', runCmd: 'Sunucuda komut çalıştır', runCmdHint: 'SSH ile uzak dizinde komut çalıştırır', diffLocalMissing: 'Yerelde yok', diffTooBig: 'Dosya karşılaştırma için çok büyük — tam görünüm gösteriliyor',
    unsavedTitle: 'Kaydedilmemiş değişiklikler', unsavedBody: 'Bu formda kaydedilmemiş değişiklikler var. Sayfadan çıkılırsa kaybolacak.', unsavedLeave: 'Kaydetmeden çık',
    diffHint: 'Sunucu → yerel: <span class="bad-t">−</span> sunucudan gidecek, <span class="ok-t">+</span> sync ile gelecek', foldN: '⋯ {0} aynı satır', clickToDiff: 'Farkı görmek için dosyaya tıklayın',
  },
  en: {
    navAll: 'All projects', navJobs: 'Jobs', navProjects: 'Projects',
    addProject: 'Add project', navNoProjects: 'No project in this folder', watching: 'Watch', language: 'Language', quit: 'Quit',
    loading: 'Loading…', shortcutKey: 'Shortcut: {0}', skip_ignored: 'ignored', skip_excluded: 'excluded', skip_not_included: 'not included', skipHint_ignored: 'Matched by the local ignore files (.gitignore / syncftp.ignore): sync doesn’t scan or upload it. The copy on the server is left alone, never deleted.', skipHint_excluded: 'Exclude applies to the local side: this local file is not uploaded. The copy on the server is left alone, never deleted.', skipHint_not_included: 'Not in the include list: this local file is not uploaded. The copy on the server is left alone, never deleted.', watchAll: 'Watch all', watchStopAll: 'Stop all watches', watchAllHint: 'Starts a separate watch for every enabled profile', watchServerHint: 'Watch this profile: locally changed files upload automatically (CLI: watch -s)', watchingTag: 'watching', watchCount: '{0}/{1} profiles', watchSyncs: '{0} auto syncs', mgSync: 'Sync', mgCheck: 'Check', mgHistory: 'Time & rollback', mgServer: 'Profile', showMore: 'Show {0} more', foldAdvanced: 'Advanced', foldMore: 'More', foldDefault: 'default', foldOff: 'off', foldConn: 'connections', foldRetry: 'retries', foldBlock: 'blocked ext', setTabGeneral: 'General', setTabLibrary: 'Connection library', setTabDefaults: 'New project defaults', setTabProjects: 'Projects', setTabHooks: 'Webhook / CI', setTabSecurity: 'Host keys', setTabSystem: 'System', deleteRecursive: 'Delete with contents', deleteRecursiveHint: 'When off, only empty folders are deleted (CLI: remote rm --recursive)', stFilter: 'Filter', stFilterHint: 'Only for this list — config is not changed (CLI: status --include / --exclude)', moreRows: 'Show {0} more', stFrozen: 'Frozen', stProtected: 'Protected', stBlocked: 'Blocked', stUploadSize: 'to upload: {0}', treeView: 'Tree', treeHint: 'Show everything below this folder as a tree (shell tree)', treeAsk: 'Show at most how many entries per folder?', treeAll: 'All', ignoreNoFile: 'ignore_files has no syncFTP file — only .gitignore is read', viewCode: 'Code', fmtRaw: 'Original', fmtPretty: 'Beautify', fmtMinified: 'File is minified — showing a readable version; the file on the server is not changed', fmtHint: 'View only: the file on the server or on disk is not changed', fmtFailed: 'Could not format: {0}', viewRendered: 'Preview', viewRenderedLocal: 'Preview (local)', viewRenderedRemote: 'Preview (server)', scanningLocal: 'Scanning…', refresh: 'Refresh', save: 'Save', cancel: 'Cancel', close: 'Close', delete: 'Delete',
    edit: 'Edit', back: 'Back', yes: 'Yes', confirm: 'Confirm', browse: 'Browse', add: 'Add', open: 'Open',
    saved: 'Saved', copied: 'Copied', never: 'never', none: 'none', all: 'All', manual: '(manual entry)',
    projectsTitle: 'Projects', projectsSub: `syncFTP projects in {0} (like the CLI: the panel works on the folder it was opened in)`,
    noProjects: 'No projects yet', noProjectsSub: 'This folder (and its subfolders) has no .syncftp/. Set a project up here; subfolders where you run `syncftp init` show up automatically.',
    serversN: '{0} profiles', enabledN: '{0} enabled', configError: 'Config could not be read',
    addProjectTitle: 'Add project', folder: 'Project folder', folderHint: 'The folder the panel was opened in, or one below it (.syncftp/ is created if missing)',
    displayName: 'Display name', displayNameHint: 'Defaults to the folder name',
    createConfigTitle: 'Create config?', createConfigBody: 'This folder has no syncFTP config. Create an empty one and add the project? You can add servers next.',
    create: 'Create',
    tabOverview: 'Overview', tabServers: 'Profiles', tabConnections: 'Connections', tabSettings: 'Project settings',
    tabFiles: 'Files', tabHistory: 'History',
    sumServers: 'Profiles', sumPending: 'Pending changes', sumLastSync: 'Last sync', sumWatch: 'Watch',
    syncAll: 'Sync all', dryRunAll: 'Dry-run', more: 'More',
    calibrateAll: 'Calibrate all', calibrateHint: 'Match server files by size and build state (no upload)',
    verifyAll: 'Verify all', verifyHint: 'Find files changed/deleted on the server by hand',
    detectAll: 'Measure connection limits', detectHint: 'Find the most connections the server accepts and save to max_connections',
    watchStart: 'Start watch', watchStop: 'Stop watch', watchOff: 'Off',
    watchSince: 'since {0} · {1} syncs', watchStarted: 'Watch started', watchStopped: 'Watch stopped', noWatches: 'No watch is running',
    noServers: 'No profiles in this project', noServersSub: 'Add the first profile: it sets where the local folder is uploaded on the server — FTP, FTPS or SFTP, optionally through an SSH tunnel.',
    addServer: 'Add profile', disabled: 'disabled', busy: 'running', tunnel: 'tunnel',
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
    filesConnTip: 'Browse with this connection: {0}', filesConnHint: 'Browsing with a connection: starts at the root and is not tied to sync. Freeze, compare with local and editing are for profiles only.', statusTitle: '{0} — pending changes', stViewList: 'List', stViewTree: 'Folders', stViewHint: 'Show as a flat list or as a folder tree (like ls)', newFiles: 'New', changedFiles: 'Changed', deletedFiles: 'Deleted locally',
    deletedNote: 'Locally deleted files are not deleted from the server, only reported.', emptyList: 'Empty',
    milestoneTitle: '{0} — milestone', milestoneBody: 'A milestone keeps files changed before its date out of syncs (persistent filter). "Milestone sync" uploads only files changed after it.',
    milestoneCurrent: 'Current: {0}', milestoneDate: 'Date', milestonePh: 'today', milestoneDateHint: 'empty = now · today · yesterday · 3d · 5h · 14:30 · 20.07 · 20.07.2026 15:04 · 2026-07-20',
    milestoneSet: 'Set', milestoneClear: 'Clear', milestoneSaved: 'Milestone: {0}', milestoneCleared: 'Milestone cleared',
    freezeTitle: '{0} — freeze list', freezeBody: 'One file path per line (relative to the local folder). These files are never uploaded, even when changed.',
    freezeSaved: 'Freeze list saved ({0} entries)',
    colName: 'Name', colProtocol: 'Protocol', colHost: 'Host', colRemote: 'Remote path', colLocal: 'Local path', colEnabled: 'Enabled',
    duplicate: 'Duplicate', deleteServerConfirm: 'Remove profile "{0}" from the config? (State and backups stay on disk.)',
    serverNew: 'New profile', serverEdit: 'Edit profile',
    secGeneral: 'General', secConnection: 'Connection', secConnectionDesc: 'Pick a connection or enter the details by hand.',
    secTunnel: 'SSH tunnel', secTunnelDesc: 'If the server is not directly reachable, connect through this SSH server (jump host). Empty = no tunnel.',
    secPaths: 'Folders', secPerf: 'Performance', secFilters: 'Filters', secFiltersDesc: 'One path per line. Profile include overrides global; exclude/protect add to global.',
    secUpload: 'Upload', secHooks: 'Hooks', secHooksDesc: 'One command per line, run in the project folder in order, after global hooks.',
    fName: 'Name', fEnabled: 'Enabled', fEnabledHint: 'Disabled profiles are skipped by "sync all"',
    fConnection: 'Connection', fProtocol: 'Protocol', fHost: 'Host', fPort: 'Port', fUser: 'User', fPassword: 'Password',
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
    lockedByProfile: 'Details come from the "{0}" connection.', editProfile: 'Edit connection',
    tunnelByProfile: 'The tunnel setting also comes from the "{0}" connection.',
    connectionsSub: 'Profiles that use the same FTP/SFTP account share the details from one place.',
    noConnections: 'No connections', noConnectionsSub: 'If several profiles connect to the same account, keep the details in a connection.',
    addConnection: 'Add connection', connNew: 'New connection', connEdit: 'Edit connection', usedBy: 'Used by',
    connInUse: 'Connection "{0}" is used by: {1}', deleteConnConfirm: 'Delete connection "{0}"?',
    secProject: 'Project', fProjectName: 'Project name (config)', fDefaultPath: 'Default local folder', fDefaultPathHint: 'Profiles without their own local folder read from here',
    secGlobal: 'Global sync settings', secGlobalDesc: 'Apply to every profile; profile settings extend them.',
    fIgnoreFiles: 'Ignore files', fIgnoreFilesHint: 'Empty = .gitignore + .syncftp/syncftp.ignore',
    fFirstFull: 'Upload everything on first sync', fFirstFullHint: 'When off, the first sync matches existing files with calibrate',
    fBackupKeep: 'Backups to keep', fBackupKeepHint: 'How many recent syncs keep their backup (rollback depth)',
    secSecurity: 'Security', keyringTitle: 'Store passwords in the system keyring', keyringHint: 'Windows Credential Manager / macOS Keychain. Only a "keyring:" marker stays in syncftp.json.',
    keyringOn: 'Passwords moved to the keyring', keyringOff: 'Passwords are in the config again', secretErrors: 'Some passwords could not be read from the keyring',
    secRegistry: 'Project registry', registryDesc: 'The name shown in the panel. Removing from the list does not touch any files.',
    renameProject: 'Rename', 
    filter: 'Filter this folder…', colSize: 'Size', colDate: 'Date',
    emptyDir: 'Folder is empty', preview: 'Preview', download: 'Download', deleteRemoteConfirm: 'PERMANENTLY delete "{0}" from the server?', deleteDirConfirm: 'PERMANENTLY delete folder "{0}" and EVERYTHING inside it from the server?',
    deleted: 'Deleted', binaryFile: 'Binary file — no preview', viewMedia: 'View', mediaFailed: 'This file cannot be shown in the browser', mediaSameSize: 'Local and server files have the same size', mediaNoLocal: 'Not on local', mediaNoRemote: 'Not on server', dropHere: 'Drop to upload', dropInto: 'Drop to upload into {0}', dropHint: 'You can drag and drop files / folders here', truncated: '(first 256 KB)',
    releases: 'Sync history (releases)', releasesSub: 'Every sync backs up the previous version of the files it changes. Rolling back a release also rolls back every newer one.',
    colWhen: 'When', colServer: 'Profile', colFiles: 'Files', colReplaced: 'Replaced', colNew: 'New', colStatus: 'Status',
    rolledBack: 'rolled back', noBackupTag: 'no backup', restorable: 'restorable', details: 'Details', undo: 'Roll back',
    noReleases: 'No releases yet', rollbackTitle: 'Roll back — {0}',
    rollbackBody: 'The following {0} release(s) will be undone, NEWEST first: replaced files get their previous version back. Newly added files stay on the server unless you tick the box below.',
    rollbackGo: 'Roll back',
    logs: 'Logs', noLogs: 'No logs', releaseTitle: 'Release {0}',
    relPickFile: 'Pick a file on the left to see its diff', relSelected: '{0} file(s) selected', relNotRestorable: 'This release cannot be rolled back (no backup or already rolled back)', relRollbackSel: 'Roll back selected ({0})', relRollbackAll: 'Roll back all', relDelete: 'Delete', relDeleteAll: 'Delete all backups', relDeleteTitle: 'Delete {0} release(s)?', relDeleteTitleAll: '{0} — delete all releases?', relDeleteBody: 'The records and their backup files are deleted; these releases leave the history and cannot be rolled back. <b>Files on the server are not touched.</b>', relDeleteKeep: 'Only delete the backup files — keep the releases in history', relDeletedN: '{0} release(s) deleted', relPurgedN: 'Backups of {0} release(s) deleted', relPartly: 'partly rolled back ({0} files)', relRolledFile: 'This file was rolled back', relNoBackupFile: 'No backup — cannot be rolled back', noFiles: 'No files',
    rbSelBody: '{0} file(s) return to their version before release {1}.', rbDeleteNew: 'Also delete the {0} newly added file(s) from the server', rbDeleteNewHint: 'When off, files these releases added stay on the server; you can select and delete them later from the tree.', rbSelChain: 'These files also changed in newer releases, so they are rolled back there too:',
    relBefore: 'Before', relAfter: 'Uploaded', relDiffHint: 'Previous server version → version uploaded by this sync', relFileHint: 'Click a file to see its diff',
    relNoBefore: 'No previous version — the backup is missing or was purged as old.', relNoAfter: 'The uploaded version was not stored and the local file has changed since.', relNewFile: 'File newly added to the server by this sync', relAfterLocal: 'Uploaded version taken from the local file (same hash)',
    jobsTitle: 'Jobs', jobsSub: 'Jobs running and finished in this session', noJobs: 'No jobs yet', activeWatches: 'Active watches',
    colJob: 'Job', colProject: 'Project', colStarted: 'Started', colDuration: 'Duration',
    stRunning: 'running', stDone: 'done', stStopped: 'stopped', stError: 'error', stop: 'Stop', jobStopping: 'Stopping…',
    jobDone: '{0} finished', jobFailed: '{0}: {1}', jobLog: 'Job log', resultLine: '↑ {0} uploaded · {1} failed · {2}',
    quitConfirm: 'Quit syncFTP completely? Running watches and the webhook listener stop too. (Closing the window keeps it running in the tray.)', quitDone: 'Panel stopped — you can close this window.',
    sessionLost: 'Lost connection to the panel', ago: '{0} ago', justNow: 'just now', min: 'min', hour: 'h', day: 'd',
    copyPath: 'Copy path', notFound: 'Not found', minutes: '{0} min', seconds: '{0} s',
    compare: 'Compare with local', diffTitle: 'Diff', diffLocal: 'Local', diffRemote: 'Server', diffSame: 'Local and server versions are identical',
    diffRemoteMissing: 'Not on the server — sync uploads it as a new file',
    secRemote: 'SSH commands (on the server)', secRemoteDesc: 'SFTP servers only: run over SSH in the remote folder. If pre fails the sync is aborted; post runs after the deploy (migrate, cache, restart…).', fRemotePre: 'remote_pre_sync', fRemotePost: 'remote_post_sync',
    whTest: 'Test', whTestOk: 'Webhook sent: {0}', ignoreTitle: 'Ignore rules (.syncftp/syncftp.ignore)', ignoreDesc: 'Gitignore syntax. Merged with .gitignore; ignored folders are never walked locally or on the server.',
    ignoreTemplate: 'Add template…', ignoreSaved: 'Ignore rules saved', gitignoreNote: '.gitignore ({0} lines) is applied too',
    failedTitle: '{0} — failed files', failedBody: 'Files that could not be uploaded in the last sync. Retrying uploads only these.',
    compareServers: 'Compare profiles', compareServersHint: 'Compare the remote contents of two profiles (only on one side / size differs)', compareA: 'Profile A', compareB: 'Profile B', compareGo: 'Compare',
    upload: 'Upload', newFolder: 'New folder', folderName: 'Folder name', rename: 'Rename / move', newPath: 'New path (relative to remote root)',
    moveTo: 'Move', selectedN: '{0} selected', clearSel: 'Clear selection',
    freezeToggle: 'Freeze / unfreeze', frozenBadge: 'frozen', searchAll: 'Search subfolders', searchTitle: 'Search "{0}"', searchNone: 'No results', searching: 'Searching…', searchingN: '{0} results · {1} folders scanned…', searchDoneN: '{0} results · {1} folders · {2}s', searchTruncated: 'Too many results (first 2000) or the 5-minute limit was hit — narrow the search.',
    uploaded: '{0} file(s) uploaded', pkTitle: 'Quick search', pkPlaceholder: 'Search projects, pages or actions…', pkPlaceholderShort: 'Search…', pkGroupPages: 'Pages', pkGroupProjects: 'Projects', pkGroupActions: 'Actions', pkEmpty: 'No results', pkMove: 'move', pkOpen: 'open', tbJobsN: '{0} running', tbMin: 'Minimize', tbMax: 'Maximize', tbRestore: 'Restore', tbClose: 'Close', cmRename: 'Rename', cmNoEnabled: 'No enabled profile', cmBack: 'Back', cmForward: 'Forward', cmJobs: 'Jobs', cmSettings: 'Settings', cmHelp: 'Help', cmPage: 'Page', cmCopy: 'Copy', cmCopied: '{0} copied', cmCopyFail: 'Could not copy; select it and press Ctrl+C', cmText: 'Text', cmSelection: 'Selected text', cmBrowserNote: '<kbd>Shift</kbd> + right click for the browser menu', cmOpen: 'Open', cmUploadHere: 'Upload into this folder…', cmCopyMenu: 'Copy', cmCopyName: 'Name', cmCopyRel: 'Relative path', cmCopyRemote: 'Full path on the server', cmName: 'Name', cmPath: 'Path', cmSelect: 'Select', cmUnselect: 'Unselect', cmSelectAll: 'Select all', cmDeleteN: 'Delete {0} items', cmUp: 'Up one folder', cmProjectSettings: 'Project settings', cmTerminal: 'Open in terminal', cmCopyFolder: 'Copy the project folder path', cmJobLog: 'Open log', cmNotStoppable: 'This job cannot be stopped', secTriggers: 'Triggers', secTriggersDesc: 'Run when certain files are uploaded (or right before): e.g. migrate when the migrations folder changes. Triggers in the project settings also run for this profile.', secTriggersDescGlobal: 'Run for every profile; profile-only ones live on the profile edit page. When certain files are uploaded (or right before): a command, SSH, an HTTP request or a notification.', trAfter: 'After upload', trBefore: 'Before upload', trWhen: 'When', trWhenHint: 'If "before" fails, that profile\'s sync is aborted; an "after" failure does not change the sync result.', trPaths: 'Which files', trPathsHint: 'Fires when an uploaded file matches one of these patterns (same rules as include/exclude). Empty = every upload.', trRun: 'Local commands', trRunHint: 'Run in the project folder; SYNCFTP_TRIGGER_FILES is in the environment.', trRemote: 'SSH commands (on the server)', trRemoteHint: 'Run in the remote folder (php artisan migrate --force).', trRemoteFtp: 'SFTP profiles only; skipped on FTP.', trHttpUrl: 'HTTP request — URL', trHttpMethod: 'Method', trHttpHeaders: 'Headers', trHttpBody: 'Body', trHttpBodyHint: 'Empty sends a JSON summary (trigger, server, project, count, files).', trVarsHint: 'Placeholders: {files} {files_json} {count} {first} {server} {project} {trigger}', trNotify: 'Notification (Slack / Discord / webhook)', trActRun: 'command', trActNotify: 'notification', trNoAction: 'no action', trUnnamed: '(unnamed)', trAnyFile: 'any file', trEmpty: 'No triggers yet. Add one from a template below.', trAdd: 'Add trigger', trDelConfirm: 'Delete the "{0}" trigger? (Takes effect when you save)', trTest: 'Try', trTestHint: 'Run this trigger\'s actions without a sync (as it is in the form)', trTestServer: 'Profile', trTestFiles: 'Sample files', trTestNote: 'The actions really run (command, SSH, HTTP, notification). The files only go into the environment and placeholders; nothing is uploaded.', trTestNoServer: 'Save the profile first', trTpl_migrateSsh: 'Template: Laravel migrate (SSH)', trTpl_migrateHttp: 'Template: migrate (HTTP request)', trTpl_composer: 'Template: composer install (SSH)', trTpl_cache: 'Template: clear cache (SSH)', trTpl_npm: 'Template: npm run build first', trTpl_notify: 'Template: Slack / Discord notification', trTpl_blank: 'Blank trigger', edDirect: 'Straight to the server · no backup', edDirectHint: 'Opened through a connection: there is no local copy and no backup (Undo), saving writes the server file directly. You are still asked if the server file changed after you opened it.', zipSel: 'Zip', zipTitle: 'Create a zip archive', zipName: 'Archive name', zipHint: '{0} item(s) are packed into one archive in this folder; if the name is taken, -2 is appended. FTP cannot run commands on the server, so the files are downloaded, archived and the archive is uploaded back.', unzip: 'Unzip', unzipTo: 'Into {0}/', unzipHere: 'Into this folder', unzipHint: 'Existing files are overwritten. The archive is downloaded and its files are uploaded one by one.', upWaiting: 'queued…', upDone: 'uploaded', upCount: '{0} / {1} files', created: 'Created', moved: 'Moved', frozenAdded: '{0} item(s) added to the freeze list', frozenRemoved: '{0} entr(y/ies) removed from the freeze list',
    deleteManyConfirm: 'PERMANENTLY delete {0} item(s) from the server? (folders with their contents)', runCmd: 'Run command on server', runCmdHint: 'Runs a command over SSH in the remote folder', diffLocalMissing: 'Missing locally', diffTooBig: 'File too large to diff — showing full view',
    unsavedTitle: 'Unsaved changes', unsavedBody: 'This form has unsaved changes. They will be lost if you leave the page.', unsavedLeave: 'Leave without saving',
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
const unlines = (s) => String(s || '').split(/\r?\n/).map((x) => x.trim()).filter(Boolean);
const unwords = (s) => String(s || '').split(/[\s,]+/).map((x) => x.trim()).filter(Boolean);

async function api(method, url, body) {
  const opt = { method, headers: {} };
  if (body !== undefined) { opt.headers['Content-Type'] = 'application/json'; opt.body = JSON.stringify(body); }
  let r;
  try { r = await fetch(url, opt); } catch (e) { throw new Error(t('sessionLost')); }
  let j = null;
  try { j = await r.json(); } catch { /* boş */ }
  if (!r.ok || !j || j.ok === false) throw Object.assign(new Error((j && j.error) || `HTTP ${r.status}`), { status: r.status, body: j });
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
  el.onclick = () => el.remove(); // tıklayınca kapanır
  $('#toastlar').appendChild(el);
  setTimeout(() => el.remove(), kind === 'bad' ? 7000 : 3500);
}
const fail = (e) => toast(e.message || String(e), 'bad');

// ── modal ────────────────────────────────────────────────────────────────────
// guard: kullanıcı kapatmak istediğinde (Esc / ✕ / dışarı tık) çağrılır; false dönerse modal açık kalır
// actions[].key: tek tuşlu kısayol (ör. 's') — en üstteki modalda, yazı alanı dışında ve Ctrl/Alt'sız çalışır
function modal({ title, body, size = '', actions = [], onClose, guard }) {
  const root = document.createElement('div');
  root.className = 'modal-arka';
  root.innerHTML = `<div class="modal ${size}" role="dialog"><div class="modal-bas"><h3>${title}</h3>
    <button class="ikon-btn" data-x>${icon('x')}</button></div><div class="modal-ic">${body}</div>
    ${actions.length ? '<div class="modal-alt"></div>' : ''}</div>`;
  const close = () => {
    if (!S.modals.delete(close)) return; // iki kez kapanmasın
    root.remove(); document.removeEventListener('keydown', onKey); if (typeof dpClose === 'function') dpClose(); onClose && onClose();
  };
  S.modals.add(close);
  const tryClose = async () => { if (!guard || (await guard()) !== false) close(); };
  const keys = {};
  // Esc yalnızca en üstteki modalı kapatır; kısayollar da yalnızca en üsttekinde
  const onKey = (e) => {
    if ($('#modal-kok').lastElementChild !== root) return;
    if (e.key === 'Escape') return tryClose();
    if (e.ctrlKey || e.metaKey || e.altKey || e.repeat || e.target.closest('input, textarea, select, [contenteditable]')) return;
    const b = keys[e.key.toLowerCase()];
    if (b && !b.disabled) { e.preventDefault(); b.click(); }
  };
  document.addEventListener('keydown', onKey);
  root.addEventListener('mousedown', (e) => { if (e.target === root) tryClose(); });
  $('[data-x]', root).onclick = tryClose;
  const alt = $('.modal-alt', root);
  for (const a of actions) {
    if ('left' in a) { const s = document.createElement('div'); s.className = 'sol'; s.innerHTML = a.left; alt.appendChild(s); continue; }
    const b = document.createElement('button');
    b.className = 'btn ' + (a.cls || '');
    b.innerHTML = a.label + (a.key ? `<kbd class="kisayol">${esc(a.key.toUpperCase())}</kbd>` : '');
    if (a.key) { keys[a.key.toLowerCase()] = b; b.title = t('shortcutKey', a.key.toUpperCase()); }
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
  // Menü düğmenin altına sığmıyorsa daha çok yer olan tarafa açılır ve o taraftaki boşlukla
  // sınırlanır (kaydırılır) — yukarı açılınca ekranın üstünden taşıp tıklanamaz hale geliyordu
  const a = kap.getBoundingClientRect();
  const below = window.innerHeight - a.bottom - 16, above = a.top - 16;
  const h = m.scrollHeight;
  if (h > below && above > below) { m.style.top = 'auto'; m.style.bottom = 'calc(100% + 6px)'; m.style.maxHeight = `${Math.max(120, above)}px`; }
  else m.style.maxHeight = `${Math.max(120, below)}px`;
}
function closeMenus() { $$('.menu').forEach((m) => m.remove()); }
document.addEventListener('click', (e) => { if (!e.target.closest('.menu') && !e.target.closest('[data-menu]')) closeMenus(); });
// Kopyala düğmeleri: data-copy="metin" taşıyan her öğe — sayfa/modal fark etmez, tek yerden
document.addEventListener('click', (e) => {
  const b = e.target.closest('[data-copy]');
  if (!b) return;
  e.preventDefault();
  navigator.clipboard.writeText(b.dataset.copy).then(() => toast(t('copied')), fail);
});

// ── durum ────────────────────────────────────────────────────────────────────
const S = { meta: null, projects: [], jobs: new Map(), renderSeq: 0, modals: new Set(), lsCache: new Map() };

// ── iş (job) yönetimi ───────────────────────────────────────────────────────
function trackJob(id, opts = {}) {
  if (S.jobs.has(id)) return S.jobs.get(id);
  // onDone: açık log penceresinin işleyicisi (pencere kapanınca null); doneCbs: diğer bekleyenler
  // (ör. "Ölç" sonucu forma yazılır) — log penceresi açılıp kapansa da kaybolmaz
  const j = { id, kind: opts.kind || '', title: opts.title || id, server: opts.server || '', project: opts.project || '', status: 'running', last: '', lines: [], es: null, shownUntil: 0, doneCbs: [] };
  S.jobs.set(id, j);
  const es = new EventSource(`/api/jobs/${enc(id)}/stream`);
  j.es = es;
  es.onmessage = (e) => {
    j.lines.push(e.data);
    // Günlerce süren watch işlerinde bellek şişmesin: son JOB_MAX_LINES satır tutulur
    if (j.lines.length > JOB_MAX_LINES + 500) j.lines.splice(0, j.lines.length - JOB_MAX_LINES);
    if (e.data.trim()) j.last = e.data;
    if (j.onLine) j.onLine(e.data);
    scheduleDock();
  };
  es.addEventListener('done', (e) => {
    es.close();
    try {
      const snap = JSON.parse(e.data);
      Object.assign(j, { kind: snap.kind || j.kind, status: snap.status, error: snap.error, result: snap.result, title: snap.title, server: snap.server, project: snap.project, ended: snap.ended, started: snap.started });
    } catch { j.status = 'done'; }
    j.shownUntil = Date.now() + 9000;
    j.finishedAt = Date.now();
    if (j.onDone) j.onDone(j);
    for (const cb of j.doneCbs.splice(0)) { try { cb(j); } catch (err) { console.error(err); } }
    // Sonuç sağ alttaki kartta görünür (ayrıca toast YOK — aynı bildirim iki kez çıkıyordu);
    // kartı olmayan watch işleri için toast kalır. Hata kartı daha uzun durur.
    if (j.status === 'error') j.shownUntil = Date.now() + 15000;
    if (isWatchJob(j)) {
      const name = `${j.title}${j.server ? ' · ' + j.server : ''}`;
      if (j.status === 'error') toast(t('jobFailed', name, j.error || ''), 'bad');
      else if (opts.quiet !== true) toast(t('jobDone', name));
    }
    renderDock();
    onJobFinished(j);
  });
  es.onerror = () => { /* tarayıcı yeniden bağlanır; done gelmişse zaten kapalı */ };
  renderDock();
  return j;
}
// İş durumu → [rozet sınıfı, etiket] (dock, iş penceresi ve İşler sayfası ortak)
function jobPill(s) {
  if (s === 'running') return ['busy', t('stRunning')];
  if (s === 'error') return ['bad', t('stError')];
  if (s === 'stopped') return ['warn', t('stStopped')];
  return ['ok', t('stDone')];
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
// Watch süresiz çalışan bir iş: sağ alttaki kutuda durmaz (her sync satırında git-gel yapıyordu),
// sol alttaki "Watch" satırında sayılır, konsolu oradan açılır.
const isWatchJob = (j) => j.kind === 'watch';
const JOB_MAX_LINES = 5000;
const JOB_KEEP_MS = 30 * 60 * 1000;
// SSE her log satırında dock'u yeniden çizmesin — kare başına en fazla bir kez
let dockRaf = 0;
function scheduleDock() { if (!dockRaf) dockRaf = requestAnimationFrame(() => { dockRaf = 0; renderDock(); }); }
// Biteli 30 dk'dan uzun olan işler bellekten atılır (çalışanlar ve log penceresi açık olanlar kalır)
function pruneJobs() {
  const now = Date.now();
  for (const [id, j] of S.jobs) if (j.status !== 'running' && j.finishedAt && now - j.finishedAt > JOB_KEEP_MS && !j.onLine) S.jobs.delete(id);
}
function renderDock() {
  const el = $('#isler');
  const now = Date.now();
  const visible = [...S.jobs.values()].filter((j) => !isWatchJob(j) && !j.dismissed && (j.status === 'running' || j.shownUntil > now)).slice(-5);
  // Kartlar iş kimliğiyle yerinde güncellenir: eskiden her çizimde (3 sn'de bir + her log satırında)
  // innerHTML baştan yazılıyor, giriş animasyonu yeniden oynuyor → kart sürekli gelip gidiyordu
  const keep = new Set(visible.map((j) => j.id));
  for (const d of [...el.children]) if (!keep.has(d.dataset.job)) d.remove();
  visible.forEach((j, i) => {
    const [st, txt] = jobPill(j.status);
    // Son satır ayrıca yazılır — her log satırında ilerleme çubuğunun animasyonu baştan başlamasın
    const html = `<div class="ust"><span class="pill ${st}"><span class="nokta"></span>${txt}</span>
      <b>${esc(j.title)}${j.server ? ' · ' + esc(j.server) : ''}</b></div>
      <div class="son"></div>
      ${j.status === 'running' ? '<div class="bar"><i></i></div>' : ''}`;
    let d = [...el.children].find((c) => c.dataset.job === j.id);
    if (!d) {
      d = document.createElement('div');
      d.className = 'is';
      d.dataset.job = j.id;
      // Tıklayınca log açılır ve kart kapanır (iş sürüyorsa İşler sayfasında / sayacında görünmeye devam eder)
      d.onclick = () => { j.dismissed = true; renderDock(); showJobLog(j.id); };
    }
    if (d.__html !== html) { d.innerHTML = html; d.__html = html; }
    const son = j.status === 'error' ? j.error || '' : j.status === 'running' ? j.last || '…' : (resultSuffix(j.result).replace(/^ — /, '') || j.last || '…');
    const sonEl = d.querySelector('.son');
    if (sonEl.textContent !== son) sonEl.textContent = son;
    if (el.children[i] !== d) el.insertBefore(d, el.children[i] || null);
  });
  document.documentElement.style.setProperty('--dock-h', el.offsetHeight ? el.offsetHeight + 8 + 'px' : '0px');
  const running = [...S.jobs.values()].filter((j) => j.status === 'running' && !isWatchJob(j)).length;
  $('#nav-jobs').textContent = running ? running : '';
  if (window.tbUpdate) window.tbUpdate();
  renderWatchRow();
}
// Watch aç/kapa — servers boşsa tüm aktif sunucular; server boşsa projenin tüm watch'ları
async function watchStart(P, servers) {
  try {
    const res = await api('POST', `/api/p/${enc(P)}/watch`, { servers });
    for (const w of res.started || []) trackJob(w.job_id, { kind: 'watch', project: P, server: w.server, title: t('watching'), quiet: true });
    for (const e of res.errors || []) toast(e, 'bad');
    if ((res.started || []).length) toast(t('watchStarted') + ': ' + res.started.map((w) => w.server).join(', '));
  } catch (e) { fail(e); }
  await loadProjects(); render(true);
}
async function watchStop(P, server = '') {
  try { await api('DELETE', `/api/p/${enc(P)}/watch${server ? `?server=${enc(server)}` : ''}`); toast(t('watchStopped') + (server ? ': ' + server : '')); } catch (e) { fail(e); }
  await loadProjects(); render(true);
}
function runningWatches() { return [...S.jobs.values()].filter((j) => isWatchJob(j) && j.status === 'running'); }
function renderWatchRow() {
  const row = $('#sb-watch-row');
  if (!row) return;
  const ws = runningWatches();
  const n = Math.max(ws.length, S.projects.reduce((a, p) => a + (p.watched || []).length, 0));
  $('#sb-watch').innerHTML = n ? `<span class="busy" style="display:inline-flex"><span class="nokta"></span></span>${n}` : '0';
  row.classList.toggle('aktif', n > 0);
  row.title = ws.map((j) => `${j.project} · ${j.server}: ${j.last || '…'}`).join('\n');
}
async function openWatches(anchor) {
  await refreshJobsMeta();
  const ws = runningWatches();
  if (!ws.length) return toast(t('noWatches'));
  // Birden fazlaysa ilki açılır; pencerenin üstündeki pill'lerle diğerlerine geçilir
  showJobLog(ws[0].id);
}
const watchLabel = (j) => `${j.project}${j.server ? ' · ' + j.server : ''}`;
setInterval(() => { pruneJobs(); renderDock(); }, 3000);

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
  if (fresh) j = trackJob(id, { kind: snap.kind, title: snap.title, server: snap.server, project: snap.project, quiet: true });
  const header = () => {
    const [st, txt] = jobPill(j.status);
    return `<span class="pill ${st}"><span class="nokta"></span>${txt}</span> ${esc(snap.title)}${snap.server ? ' · ' + esc(snap.server) : ''}`;
  };
  // Watch penceresi: açık watch'lar arasında geçiş pill'leri (tıkla → o watch'ın konsolu)
  const others = snap.kind === 'watch' ? runningWatches() : [];
  const switcher = others.length > 1 ? `<div class="watch-gecis">${others.map((w) => `<button class="pill${w.id === id ? ' aktif' : ''}" data-wsw="${esc(w.id)}" title="${esc(w.last || '…')}"><span class="nokta"></span>${esc(watchLabel(w))}</button>`).join('')}</div>` : '';
  const m = modal({
    title: header(), size: 'genis',
    body: `${switcher}<div class="log" id="joblog">${(fresh ? [] : j.lines).map(logLine).join('')}</div>`,
    actions: [
      { left: `<span class="mono">${esc(snap.project || '')} · ${fmtDate(snap.started)}</span>` },
      // Durdurulabilen her iş (watch, calibrate…) — pencere açık kalır, durma satırı log'a düşer
      ...(snap.cancelable && j.status === 'running' ? [{ label: icon('stop') + t('stop'), cls: 'tehlike', run: async () => {
        await api('DELETE', `/api/jobs/${enc(id)}`);
        toast(snap.kind === 'watch' ? t('watchStopped') : t('jobStopping'));
        return snap.kind === 'watch';
      } }] : []),
      { label: t('close'), cls: 'birincil' },
    ],
    onClose: () => { j.onLine = null; j.onDone = null; },
  });
  $$('[data-wsw]', m.root).forEach((b) => (b.onclick = () => { if (b.dataset.wsw !== id) { m.close(); showJobLog(b.dataset.wsw); } }));
  const box = $('#joblog', m.root);
  box.scrollTop = box.scrollHeight;
  j.onLine = (l) => {
    const stick = box.scrollTop + box.clientHeight >= box.scrollHeight - 30;
    box.insertAdjacentHTML('beforeend', logLine(l));
    if (stick) box.scrollTop = box.scrollHeight;
  };
  j.onDone = () => {
    $('.modal-bas h3', m.root).innerHTML = header();
    $$('.modal-alt .tehlike', m.root).forEach((b) => b.remove()); // bitti → Durdur anlamsız
  };
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
    for (const x of list) if (x.status === 'running' && !S.jobs.has(x.id)) trackJob(x.id, { kind: x.kind, title: x.title, server: x.server, project: x.project, quiet: true });
  } catch { /* yoksay */ }
}

// ── router ───────────────────────────────────────────────────────────────────
function parseRoute() {
  const parts = location.hash.replace(/^#\/?/, '').split('/').map(decodeURIComponent);
  if (parts[0] === 'p' && parts[1]) return { page: 'project', project: parts[1], tab: parts[2] || 'overview', sub: parts[3] ?? null, sub2: parts[4] ?? null };
  if (parts[0] === 'jobs') return { page: 'jobs' };
  if (parts[0] === 'help') return { page: 'help', sub: parts[1] || '' };
  if (parts[0] === 'settings') return { page: 'settings', sub: parts[1] || 'general' };
  if (parts[0] === 'terminal') return { page: 'terminal', sub: parts.length > 1 ? parts[1] : null };
  return { page: 'home' };
}
const go = (h) => { if (location.hash === h) render(); else location.hash = h; };
const pHash = (p, tab, sub) => `#/p/${enc(p)}${tab ? '/' + tab : ''}${sub != null ? '/' + enc(sub) : ''}`;
// Kaydedilmemiş form koruması: trackDirty ile izlenen formda (sunucu / bağlantı / proje ayarları) değişiklik
// varken başka sayfaya geçilirse sorulur; vazgeçilirse adres eski haline döner, sayfa yeniden çizilmez.
function trackDirty(form) {
  if (!form) return;
  const d = { form, hash: location.hash, on: false };
  S.dirty = d;
  // Yalnızca config alanları (data-k) sayılır — anında kaydedilen anahtarlar / ignore editörü sayılmaz
  const mark = (e) => { if (S.dirty === d && e.target.closest('[data-k]')) d.on = true; };
  form.addEventListener('input', mark);
  form.addEventListener('change', mark);
}
const clearDirty = () => { S.dirty = null; };
window.addEventListener('hashchange', async () => {
  const d = S.dirty;
  if (d && d.on && d.form.isConnected && location.hash !== d.hash) {
    const target = location.hash;
    history.replaceState(null, '', d.hash);
    if (!(await confirmBox(t('unsavedTitle'), esc(t('unsavedBody')), t('unsavedLeave'), true))) return;
    if (S.dirty === d) S.dirty = null;
    location.hash = target;
    return;
  }
  render();
});

async function render(soft = false) {
  const seq = ++S.renderSeq;
  const r = parseRoute();
  if (r.project) S.lastProject = r.project;
  document.body.classList.toggle('terminal-modu', r.page === 'terminal');
  if (window.tbUpdate) window.tbUpdate(); // kendi pencere barımız (titlebar.js)
  $$('.nav a').forEach((a) => a.classList.toggle('aktif',
    !!((a.dataset.nav === 'home' && r.page === 'home') || (a.dataset.nav === 'jobs' && r.page === 'jobs') || (a.dataset.nav === 'help' && r.page === 'help') || (a.dataset.nav === 'settings' && r.page === 'settings') || (a.dataset.nav === 'terminal' && r.page === 'terminal') || (a.dataset.proj && a.dataset.proj === r.project))));
  const root = $('#sayfa');
  if (!soft) {
    S.dirty = null;
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
    else if (r.page === 'settings') html = await pageSettings(r.sub);
    else if (r.page === 'terminal') html = await pageTerminal(r.sub);
    else html = await pageProject(r);
    if (seq !== S.renderSeq) return;
    if (typeof html === 'string') root.innerHTML = html;
    $('#frame')?.classList.remove('surukle'); // sürükle-bırak katmanı yarıda kaldıysa yeni sayfaya taşınmasın
    root.dataset.proj = r.page === 'project' ? r.project : '';
    if (root._bind) { const b = root._bind; root._bind = null; b(root); }
    // Sayfa değişimiyle birlikte açılacak modal (ör. arama sonucundan önizleme): hashchange'in
    // tetiklediği render modalları kapatır → modal ancak çizim bittikten SONRA açılabilir
    if (S.afterRender) { const f = S.afterRender; S.afterRender = null; f(); }
  } catch (e) {
    if (seq !== S.renderSeq) return;
    S.afterRender = null;
    root.innerHTML = `<div class="bos"><div class="buyuk">${icon('alert')}</div><h3>${esc(e.message)}</h3></div>`;
  }
}
const bind = (fn) => { $('#sayfa')._bind = fn; };

// ── sidebar ──────────────────────────────────────────────────────────────────
// fresh: sunucu proje taramasını hemen tazelesin (normalde önbellekli, tam tarama dakikada bir)
async function loadProjects(fresh = false) {
  try { S.projects = await GET('/api/projects' + (fresh ? '?fresh=1' : '')); } catch (e) { S.projects = []; }
  const el = $('#proje-listesi');
  const r = parseRoute();
  el.innerHTML = S.projects.map((p) => `<a href="${pHash(p.name)}" data-proj="${esc(p.name)}" class="${r.project === p.name ? 'aktif' : ''}">
      ${p.watching ? `<span class="busy" style="display:flex"><span class="nokta"></span></span>` : icon(p.ok ? 'folder' : 'alert', p.ok ? '' : 'bad-t')}
      <span class="ad">${esc(p.name)}</span><span class="sag">${p.ok ? p.enabled : ''}</span></a>`).join('')
    || `<div class="nav-bos">${icon('folder')}<span data-t="navNoProjects">${t('navNoProjects')}</span>
        <button class="btn kucuk" id="nav-bos-ekle">${icon('plus')}<span data-t="addProject">${t('addProject')}</span></button></div>`;
  const nb = $('#nav-bos-ekle');
  if (nb) nb.onclick = addProjectFlow;
  renderWatchRow();
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
      <div class="alan gizli" id="ap-remote-box"><label>${t('apRemote')}</label><div class="girdi-grup"><input type="text" id="ap-remote" class="mono" value="/public_html" spellcheck="false">
        <button class="btn" id="ap-remote-pick" type="button">${icon('folder')}${t('browse')}</button></div></div></div>`,
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
  // Uzak dizin → Gözat: seçili kütüphane bağlantısında gezinir (proje henüz yok — sunucu formundaki seçiciyle aynı)
  $('#ap-remote-pick', m.root).onclick = async () => {
    if (!cs.value) return;
    const inp = $('#ap-remote', m.root);
    const picked = await pickRemoteDir({ initial: inp.value.trim() || '/', list: (path, create) => api('POST', '/api/library/browse-dirs', { name: cs.value, path, create }) });
    if (picked) inp.value = picked;
  };
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
        <a class="ikon-btn" href="#/help/${({ overview: 'sync', servers: 'start', server: 'start', connections: 'conns', connection: 'conns', files: 'files', history: 'rollback', cicd: 'cicd', settings: 'project' })[r.tab] || 'start'}" title="${t('navHelp')}">${icon('help')}</a></div></div>
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
  // Watch sunucu başına (CLI `watch -s` gibi): her kartın kendi aç/kapa düğmesi var
  const ws = ov.watches || [];
  const watched = new Set(ws.map((x) => x.server));
  servers.forEach((s) => { s.watching = watched.has(s.name); });
  const allWatched = enabled.length > 0 && enabled.every((s) => s.watching);
  const syncs = ws.reduce((a, x) => a + (x.syncs || 0), 0);

  bind((root) => {
    const P = proj.name;
    const aks = $('#proj-aks', root);
    aks.innerHTML = `${enabled.length > 1 ? `<button class="btn ${allWatched ? 'tehlike' : ''}" id="b-watch" title="${t('watchAllHint')}">${icon(allWatched ? 'stop' : 'eye')}${allWatched ? t('watchStopAll') : t('watchAll')}</button>` : ''}
      <button class="ikon-btn" id="b-refresh" title="${t('refresh')}">${icon('refresh')}</button>`;
    $('#b-refresh', root).onclick = () => render();
    const bw = $('#b-watch', root);
    if (bw) bw.onclick = () => (allWatched ? watchStop(P) : watchStart(P, []));
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
    // Dosyalar sekmesi için ön yükleme: varsayılan sunucunun kök listesi (ve sunucu bağlantısı) hazır
    // olsun, sekmeye ilk girişte FTP bağlantı + LIST beklenmesin
    const fsrv = (servers.find((x) => x.enabled) || servers[0] || {}).name;
    if (fsrv && !S.lsCache.has(`${P}|${fsrv}|`)) {
      GET(`/api/p/${enc(P)}/remote/ls?server=${enc(fsrv)}&path=`).then((d) => S.lsCache.set(`${P}|${fsrv}|`, d)).catch(() => {});
    }
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
      <div class="kutu"><span>${icon('eye')}${t('sumWatch')}</span><b style="font-size:20px" class="${ws.length ? 'acc-t' : ''}">${ws.length ? t('watchCount', ws.length, enabled.length) : t('watchOff')}</b>
        ${ws.length ? `<span style="margin-top:2px">${t('watchSyncs', syncs)}</span>` : ''}</div>
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
        ${s.connection ? `<span class="pill">${icon('link')}${esc(s.connection)}</span>` : ''}
        ${s.watching ? `<span class="pill busy"><span class="nokta"></span>${t('watchingTag')}</span>` : ''}</div>
      <div class="meta">
        <span>${icon('server')}<span class="mono">${esc(s.user ? s.user + '@' : '')}${esc(s.host)}:${s.port}</span></span>
        <span>${icon('folder')}<span class="mono">${esc(s.local_path)}</span> <em>→</em> <span class="mono">${esc(s.remote_path)}</span></span>
        <span>${icon('zap')}${t('maxConn', s.max_connections)}${s.bandwidth ? ' · ' + t('speedLimit', s.bandwidth) : ''}</span>
        <span>${icon('clock')}${s.last_sync ? ago(s.last_sync) : t('never')}</span>
      </div>
    </div>
    <div class="butonlar">
      <button class="btn kucuk mavi" data-act="sync" ${s.busy ? 'disabled' : ''}>${icon('upload')}${t('sync')}</button>
      ${s.enabled ? `<button class="btn kucuk ${s.watching ? 'tehlike' : ''}" data-act="watch" title="${t('watchServerHint')}">${icon(s.watching ? 'stop' : 'eye')}${s.watching ? t('watchStop') : t('watchStart')}</button>` : ''}
      <button class="btn kucuk" data-act="files" title="${t('browseFiles')}">${icon('folder')}${t('tabFiles')}</button>
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
  act('watch', () => (s.watching ? watchStop(P, s.name) : watchStart(P, [s.name])));
  act('files', () => go(pHash(P, 'files', s.name)));
  act('status', () => statusModal(P, s));
  act('milestone', () => milestoneModal(P, s));
  act('freeze', () => freezeModal(P, s));
  act('retry', () => failedModal(P, s));
  act('history', () => go(pHash(P, 'history', s.name)));
  act('menu', (b) => openMenu(b, [
    { title: t('mgSync') },
    { icon: 'eye', label: t('dryRun'), hint: t('dryRunHint'), run: () => job({ kind: 'sync', dry_run: true }, t('dryRun')) },
    { icon: 'upload', label: t('fullSync'), hint: t('fullSyncHint'), run: async () => { if (await confirmBox(t('fullSync'), esc(t('fullSyncConfirm', s.name)), t('fullSync'), true)) job({ kind: 'sync', full: true }, t('fullSync')); } },
    s.failed ? { icon: 'refresh', label: t('retryFailed'), hint: t('retryFailedHint', s.failed), run: () => job({ kind: 'sync', retry_failed: true }, t('retryFailed')) } : null,
    { icon: 'sliders', label: t('syncOnce'), hint: 'include / exclude', run: () => filteredSyncModal(P, s) },
    { title: t('mgCheck') },
    { icon: 'scan', label: t('calibrate'), hint: t('calibrateHint'), run: () => job({ kind: 'calibrate' }, t('calibrate')) },
    { icon: 'shield', label: t('verify'), hint: t('verifyHint'), run: () => job({ kind: 'verify' }, t('verify')) },
    { icon: 'shield', label: t('verifyFix'), hint: t('verifyFixHint'), run: () => job({ kind: 'verify', fix: true }, t('verifyFix')) },
    { icon: 'gauge', label: t('detect'), hint: t('detectHint'), run: () => job({ kind: 'detect' }, t('detect')) },
    { title: t('mgHistory') },
    { icon: 'clock', label: t('milestone') + '…', hint: s.milestone || '', run: () => milestoneModal(P, s) },
    s.milestone ? { icon: 'upload', label: t('milestoneSync'), hint: t('milestoneSyncHint'), run: () => job({ kind: 'milestone-sync' }, t('milestoneSync')) } : null,
    { icon: 'snow', label: t('freeze') + '…', hint: t('freezeHint'), run: () => freezeModal(P, s) },
    { icon: 'undo', label: t('rollback') + '…', hint: t('rollbackHint'), run: () => go(pHash(P, 'history', s.name)) },
    { title: t('mgServer') },
    s.protocol === 'sftp' ? { icon: 'terminal', label: t('runCmd') + '…', hint: t('runCmdHint'), run: () => execModal(P, s) } : null,
    { icon: 'folder', label: t('browseFiles'), run: () => go(pHash(P, 'files', s.name)) },
    { icon: 'edit', label: t('edit'), run: () => go(pHash(P, 'server', s.name)) },
  ]));
}

// flt: tek seferlik include/exclude (CLI status --include / --exclude) — "Filtreli sync" ile aynı liste editörleri
async function statusModal(P, s, flt = { include: [], exclude: [] }) {
  const fq = `&include=${enc(flt.include.join('\n'))}&exclude=${enc(flt.exclude.join('\n'))}`;
  const proj = S.projects.find((x) => x.name === P) || {};
  const fltOn = flt.include.length + flt.exclude.length > 0;
  const lm = await loadModal(t('statusTitle', esc(s.name)), 'orta', () => GET(`/api/p/${enc(P)}/status?server=${enc(s.name)}${fq}`));
  if (!lm) return;
  const st = lm.d;
  // Görünüm: düz liste ya da klasör ağacı (ls gibi) — tercih hatırlanır, iki görünüm de çizilir
  let view = localStorageGet('syncftp-st-view') === 'tree' ? 'tree' : 'list';
  const flat = (arr, mark, click) => `<div class="liste-kutu">${arr.map((x) => `<div ${click ? `class="tikla" data-diff="${esc(x)}"` : ''}>${mark} ${esc(x)}</div>`).join('')}</div>`;
  const list = (arr, mark, click) => arr.length
    ? `<div data-stv="list" ${view !== 'list' ? 'hidden' : ''}>${flat(arr, mark, click)}</div><div data-stv="tree" ${view !== 'tree' ? 'hidden' : ''}>${pathTreeHTML(arr, mark, click)}</div>`
    : `<p class="faint">${t('emptyList')}</p>`;
  // Sync'in yükleyeceği liste ile aynı hesap: atlanacaklar (freeze / protect / engelli uzantı) ayrı sekmede
  const tabs = [['new', t('newFiles'), st.new, '+'], ['changed', t('changedFiles'), st.changed, '~'], ['deleted', t('deletedFiles'), st.deleted, '−'],
    ['frozen', t('stFrozen'), st.frozen || [], '❄'], ['protected', t('stProtected'), st.protected || [], '🔒'], ['blocked', t('stBlocked'), st.blocked || [], '✗']]
    .filter(([k, , a], i) => i < 3 || a.length);
  const first = Math.max(0, tabs.findIndex(([, , a]) => a.length)); // ilk dolu sekmede aç
  const m = modal({
    title: t('statusTitle', esc(s.name)), size: 'orta',
    body: `<div class="st-ust"><div class="tabs">${tabs.map(([k, l, a], i) => `<a href="javascript:void 0" data-st="${k}" class="${i === first ? 'aktif' : ''}">${l} <span class="sayac">${st.counts[k]}</span></a>`).join('')}</div>
        <span class="seg st-gorunum" title="${t('stViewHint')}"><button data-view="list" class="${view === 'list' ? 'aktif' : ''}">${icon('log')}${t('stViewList')}</button><button data-view="tree" class="${view === 'tree' ? 'aktif' : ''}">${icon('folder')}${t('stViewTree')}</button></span></div>
      <details class="st-filtre" style="margin:-4px 0 10px" ${fltOn ? 'open' : ''}><summary class="faint" style="cursor:pointer">${icon('sliders')} ${t('stFilter')}${fltOn ? ` <span class="pill">${flt.include.length + flt.exclude.length}</span>` : ''}</summary>
        <p class="ipucu" style="margin:6px 0 8px">${t('stFilterHint')}</p>
        <div class="form" id="st-form" data-p="${esc(P)}" data-ppath="${esc(proj.path || '')}" data-srv="${esc(s.name)}"><div class="alanlar">
          ${fArea('include', t('syncOnceInc'), flt.include, { wide: false, tree: 'paths', chips: 'include' })}${fArea('exclude', t('syncOnceExc'), flt.exclude, { wide: false, tree: 'paths', chips: 'exclude' })}</div></div>
        <div class="aksiyonlar" style="margin-top:8px;justify-content:flex-end"><button class="btn kucuk" id="st-flt">${icon('sliders')}${t('stFilter')}</button></div></details>
      <p class="ipucu" style="margin:-4px 0 10px">${t('clickToDiff')}${st.bytes ? ' · ' + t('stUploadSize', bytes(st.bytes)) : ''}</p>
      ${tabs.map(([k, , a, mk], i) => `<div data-stp="${k}" ${i !== first ? 'hidden' : ''}>${list(a, mk, k === 'new' || k === 'changed')}${k === 'deleted' ? `<p class="ipucu" style="margin-top:10px">${t('deletedNote')}</p>` : ''}</div>`).join('')}`,
    actions: [{ label: t('close'), cls: 'hayalet' }, { label: icon('upload') + t('sync'), cls: 'mavi', run: () => startJobs(P, { kind: 'sync', servers: [s.name] }, t('sync')) }],
  });
  bindFormTools(m.root);
  $$('[data-diff]', m.root).forEach((d) => (d.onclick = () => compareModal(P, s.name, d.dataset.diff)));
  $$('[data-view]', m.root).forEach((b) => (b.onclick = () => {
    view = b.dataset.view;
    localStorageSet('syncftp-st-view', view);
    $$('[data-view]', m.root).forEach((x) => x.classList.toggle('aktif', x === b));
    $$('[data-stv]', m.root).forEach((x) => (x.hidden = x.dataset.stv !== view));
  }));
  $('#st-flt', m.root).onclick = () => {
    const f = readForm($('#st-form', m.root), {});
    m.close();
    statusModal(P, s, { include: f.include || [], exclude: f.exclude || [] });
  };
  $$('[data-st]', m.root).forEach((a) => (a.onclick = () => {
    $$('[data-st]', m.root).forEach((x) => x.classList.toggle('aktif', x === a));
    $$('[data-stp]', m.root).forEach((p) => (p.hidden = p.dataset.stp !== a.dataset.st));
  }));
}

// pathTreeHTML yol listesini klasör ağacı olarak çizer (ls gibi): önce klasörler, tek alt klasörlü
// zincirler birleşik (a/b/c), klasörde altındaki dosya sayısı. Küçük listede hepsi açık; büyükte
// (200+ dosya) yalnızca ilk seviye. click → dosya satırı data-diff taşır (karşılaştırma açılır).
function pathTreeHTML(paths, mark, click) {
  const root = { dirs: {}, files: [] };
  for (const p of paths) {
    const parts = p.split('/');
    let n = root;
    for (const d of parts.slice(0, -1)) n = n.dirs[d] || (n.dirs[d] = { dirs: {}, files: [] });
    n.files.push({ path: p, name: parts[parts.length - 1] });
  }
  const openAll = paths.length <= 200;
  const count = (n) => n.files.length + Object.values(n.dirs).reduce((a, x) => a + count(x), 0);
  const draw = (n, depth) => Object.keys(n.dirs).sort((a, b) => a.localeCompare(b)).map((name) => {
    let label = name, node = n.dirs[name];
    while (!node.files.length && Object.keys(node.dirs).length === 1) {
      const only = Object.keys(node.dirs)[0];
      label += '/' + only;
      node = node.dirs[only];
    }
    return `<details ${openAll || depth === 0 ? 'open' : ''}><summary>${icon('folder', 'acc-t')}<span>${esc(label)}/</span><span class="faint rt-say">${count(node)}</span></summary>
      <div class="rt-ic">${draw(node, depth + 1)}</div></details>`;
  }).join('') + n.files.sort((a, b) => a.name.localeCompare(b.name)).map((f) =>
    `<div class="rt-dosya${click ? ' tikla' : ''}" ${click ? `data-diff="${esc(f.path)}"` : ''} title="${esc(f.path)}"><span class="mono faint">${mark}</span>${icon('file', 'faint')}<span class="rt-ad">${esc(f.name)}</span></div>`).join('');
  return `<div class="liste-kutu rt-agac st-agac">${draw(root, 0)}</div>`;
}

function milestoneModal(P, s) {
  modal({
    title: t('milestoneTitle', esc(s.name)),
    body: `<p>${t('milestoneBody')}</p>${s.milestone ? `<p><b>${t('milestoneCurrent', esc(s.milestone))}</b></p>` : ''}
      <div class="alan"><label>${t('milestoneDate')}</label><input type="text" id="ms-date" data-date placeholder="${t('milestonePh')}"><small>${t('milestoneDateHint')}</small></div>`,
    actions: [
      ...(s.milestone ? [{ label: t('milestoneClear'), cls: 'tehlike', run: async () => { await api('DELETE', `/api/p/${enc(P)}/milestones?server=${enc(s.name)}`); toast(t('milestoneCleared')); render(true); } }] : []),
      { label: t('cancel'), cls: 'hayalet' },
      { label: t('milestoneSet'), cls: 'birincil', run: async (root) => {
        const d = await api('PUT', `/api/p/${enc(P)}/milestones`, { servers: [s.name], date: $('#ms-date', root).value });
        toast(t('milestoneSaved', d)); render(true);
      } },
    ],
  });
  initDatePickers(document.getElementById('modal-kok').lastElementChild);
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
  const files = (lm.d && lm.d.files) || [];
  const when = lm.d && lm.d.created_at && !String(lm.d.created_at).startsWith('0001') ? ` <span class="faint">(${fmtDate(lm.d.created_at)})</span>` : '';
  modal({
    title: `${icon('alert')}${t('failedTitle', esc(s.name))}`, size: 'orta',
    body: `<p>${t('failedBody')}${when}</p><div class="liste-kutu">${files.map((f) => `<div class="bad-t">✗ ${esc(f)}</div>`).join('') || `<div class="faint">${t('emptyList')}</div>`}</div>`,
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
    $$('[data-add-server]', root).forEach((b) => (b.onclick = () => go(pHash(P, 'server', 'new'))));
    // Her işlem config'i YENİDEN okur, kopyasını değiştirip yazar: sayfa çizildikten sonra başka yerde
    // (AI / CLI / başka sekme) yapılan değişiklikler ezilmez, başarısız işlem ekrandaki kopyayı bozmaz
    const mutate = async (fn) => { const next = await getConfig(P); const r = fn(next); await api('PUT', `/api/p/${enc(P)}/config`, next); return r; };
    $$('[data-toggle]', root).forEach((inp) => (inp.onchange = async () => {
      const on = inp.checked;
      try {
        await mutate((c) => { const s = c.servers.find((x) => x.name === inp.dataset.toggle); if (!s) throw new Error(t('notFound') + ': ' + inp.dataset.toggle); s.enabled = on; });
        toast(t('saved')); loadProjects();
      } catch (e) { inp.checked = !on; fail(e); }
    }));
    $$('[data-dup]', root).forEach((b) => (b.onclick = async () => {
      try {
        const name = await mutate((c) => {
          const src = c.servers.find((x) => x.name === b.dataset.dup);
          if (!src) throw new Error(t('notFound') + ': ' + b.dataset.dup);
          const copy = JSON.parse(JSON.stringify(src));
          let n = 2; while (c.servers.some((x) => x.name === `${src.name}-${n}`)) n++;
          copy.name = `${src.name}-${n}`; copy.enabled = false;
          c.servers.push(copy);
          return copy.name;
        });
        go(pHash(P, 'server', name));
      } catch (e) { fail(e); }
    }));
    $$('[data-del]', root).forEach((b) => (b.onclick = async () => {
      if (!(await confirmBox(t('delete'), esc(t('deleteServerConfirm', b.dataset.del)), t('delete'), true))) return;
      try {
        await mutate((c) => { c.servers = c.servers.filter((x) => x.name !== b.dataset.del); });
        toast(t('saved')); render(true); loadProjects();
      } catch (e) { fail(e); }
    }));
  });
  if (!cfg.servers.length) {
    return `<div class="bos"><div class="buyuk">${icon('server', 'lg')}</div><h3>${t('noServers')}</h3><p>${t('noServersSub')}</p>
      <button class="btn birincil" data-add-server>${icon('plus')}${t('addServer')}</button></div>`;
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
  const h = box.querySelector('input[type=hidden]');
  h.value = box.dataset.ledty === 'words' ? uniq.join(' ') : uniq.join('\n');
  h.dispatchEvent(new Event('change', { bubbles: true })); // kaydedilmemiş form koruması duysun
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
// section(..., {fold, open, summary, proto}): fold → başlığa tıklayınca açılan bölüm (nadiren
// değişen ayarlar kapalı gelir, özet satırı ne ayarlı olduğunu söyler); proto → yalnızca o protokolde görünür
const section = (icn, title, desc, inner, o = {}) => {
  const pa = o.proto ? ` data-proto="${o.proto}"` : '';
  const body = `${desc ? `<p class="aciklama">${desc}</p>` : '<div style="height:10px"></div>'}${inner}`;
  if (!o.fold) return `<div class="bolum"${pa}><h3>${icon(icn)}${title}</h3>${body}</div>`;
  return `<details class="bolum katlanir"${pa} ${o.open ? 'open' : ''}><summary><h3>${icon(icn)}${title}</h3>${o.summary ? `<span class="kat-ozet">${o.summary}</span>` : ''}${icon('right', 'ok-ac')}</summary><div class="katlanir-ic">${body}</div></details>`;
};
// ── Sağ tık menüleri (contextmenu.js) ─────────────────────────────────────────
const filesCtxLive = () => { const F = S.filesCtx; return F && F.root.isConnected && parseRoute().tab === 'files' ? F : null; };
const remoteFull = (F, p) => (F.remoteRoot.replace(/\/+$/, '') + '/' + p).replace(/\/{2,}/g, '/');
function copyMenu(F, name, p) {
  return { label: t('cmCopyMenu'), icon: 'copy', children: [
    { label: t('cmCopyName'), hint: name.length > 24 ? name.slice(0, 23) + '…' : name, action: () => CM.copy(name, t('cmName')) },
    { label: t('cmCopyRel'), action: () => CM.copy(p, t('cmPath')) },
    { label: t('cmCopyRemote'), action: () => CM.copy(remoteFull(F, p), t('cmPath')) },
  ] };
}
// Dosya / klasör satırı: tıklanan öğeye göre seçenekler; seçili birden fazla öğenin üstünde → toplu işlemler
CM.register('#f-body tr[data-entry]', (tr) => {
  const F = filesCtxLive();
  if (!F) return null;
  const name = tr.dataset.entry, typ = tr.dataset.type, p = F.full(name), it = { path: p, type: typ };
  const isDir = typ === 'dir', isZip = !isDir && /\.zip$/i.test(name);
  if (F.sel.size > 1 && F.sel.has(p)) {
    const its = F.items();
    return [{ heading: t('selectedN', its.length) },
      { label: t('moveTo'), icon: 'folder', action: () => F.moveItems(its) },
      { label: t('zipSel'), icon: 'archive', action: () => F.zipItems(its) },
      !F.viaConn && { label: t('freezeToggle'), icon: 'snow', action: () => F.toggleFreeze(its) },
      '-', { label: t('clearSel'), icon: 'x', action: () => F.clearSel() },
      '-', { label: t('cmDeleteN', its.length), icon: 'trash', danger: true, action: () => F.removeItems(its) }];
  }
  const checked = F.sel.has(p);
  return [{ heading: name },
    isDir ? { label: t('cmOpen'), icon: 'folder', action: () => go(F.hashFor(p)) } : { label: t('preview'), icon: 'eye', action: () => F.previewFile(p) },
    !isDir && { label: t('edEdit'), icon: 'code', action: () => F.editFile(p) },
    !isDir && !F.viaConn && { label: t('compare'), icon: 'layers', action: () => F.compareModal(p) },
    !isDir && { label: t('download'), icon: 'download', action: () => F.download(p) },
    isDir && { label: t('cmUploadHere'), icon: 'upload', action: () => F.uploadInto(p) },
    '-',
    isZip && { label: t('unzip'), icon: 'archive', action: () => F.unzipItem(it) },
    { label: t('zipSel'), icon: 'archive', action: () => F.zipItems([it]) },
    { label: t('moveTo'), icon: 'folder', action: () => F.moveItems([it]) },
    { label: t('cmRename'), icon: 'edit', action: () => F.renameItem(it) },
    !F.viaConn && { label: t('freezeToggle'), icon: 'snow', action: () => F.toggleFreeze([it]) },
    '-',
    copyMenu(F, name, p),
    { label: checked ? t('cmUnselect') : t('cmSelect'), icon: 'check', action: () => F.toggleRow(name) },
    '-',
    { label: t('delete'), icon: 'trash', danger: true, action: () => F.removeItems([it]) }];
});
// Dosyalar sekmesinde boş alan / araç çubuğu: bulunulan klasörün işlemleri
CM.register('#sayfa', (el, ctx) => {
  const F = filesCtxLive();
  if (!F || ctx.target.closest('a[href], .modal')) return null;
  const here = F.rel || '/';
  return [{ heading: here === '/' ? F.remoteRoot : here },
    { label: t('newFolder'), icon: 'folder', action: () => F.click('#f-mkdir') },
    { label: t('pkNewFile'), icon: 'file', action: () => F.click('#f-mkfile') },
    { label: t('upload'), icon: 'upload', action: () => F.click('#f-upload') },
    '-',
    { label: t('searchAll'), icon: 'search', action: () => { const f = $('#f-filter', F.root); f.focus(); } },
    { label: t('treeView'), icon: 'layers', action: () => F.click('#f-tree') },
    { label: t('refresh'), icon: 'refresh', action: () => F.reload() },
    F.rel && { label: t('cmUp'), icon: 'left', action: () => go(F.rel.includes('/') ? F.hashFor(F.rel.split('/').slice(0, -1).join('/')) : pHash(F.P, 'files', F.srv)) },
    '-',
    $('#f-all', F.root) && { label: t('cmSelectAll'), icon: 'check', action: () => F.selectAll() },
    F.sel.size && { label: t('clearSel'), icon: 'x', action: () => F.clearSel() },
    { label: t('cmCopyRemote'), icon: 'copy', action: () => CM.copy(F.rel ? remoteFull(F, F.rel) : F.remoteRoot, t('cmPath')) }];
});
// Proje (kenar çubuğu + ana sayfadaki kart) — ikisi aynı menü
function projectMenu(P) {
  const proj = S.projects.find((x) => x.name === P);
  const ok = !proj || proj.ok !== false;
  return [{ heading: P },
    ok && { label: t('syncAll'), icon: 'upload', disabled: proj && !proj.enabled && t('cmNoEnabled'), action: () => startJobs(P, { kind: 'sync' }, t('sync')) },
    ok && { label: t('dryRun'), icon: 'eye', disabled: proj && !proj.enabled && t('cmNoEnabled'), action: () => startJobs(P, { kind: 'sync', dry_run: true }, t('dryRun')) },
    ok && (proj && proj.watching
      ? { label: t('watchStop'), icon: 'stop', action: () => watchStop(P) }
      : { label: t('watchStart'), icon: 'play', disabled: proj && !proj.enabled && t('cmNoEnabled'), action: () => watchStart(P, []) }),
    '-',
    { label: t('tabOverview'), icon: 'layers', action: () => go(pHash(P)) },
    { label: t('tabFiles'), icon: 'folder', action: () => go(pHash(P, 'files')) },
    { label: t('tabHistory'), icon: 'history', action: () => go(pHash(P, 'history')) },
    { label: t('tabServers'), icon: 'server', action: () => go(pHash(P, 'servers')) },
    { label: t('cmProjectSettings'), icon: 'sliders', action: () => go(pHash(P, 'settings')) },
    '-',
    { label: t('cmTerminal'), icon: 'terminal', action: () => go('#/terminal/' + enc(P)) },
    proj && proj.path && { label: t('cmCopyFolder'), icon: 'copy', action: () => CM.copy(proj.path, t('cmPath')) }];
}
CM.register('a[data-proj]', (a) => projectMenu(a.dataset.proj));
CM.register('.pkart[data-p]', (c) => projectMenu(c.dataset.p));
// Sağ alttaki iş kartı
CM.register('#isler .is[data-job]', (d) => {
  const j = S.jobs.get(d.dataset.job);
  if (!j) return null;
  return [{ heading: j.title + (j.server ? ' · ' + j.server : '') },
    { label: t('cmJobLog'), icon: 'log', action: () => d.click() },
    j.status === 'running' && { label: t('stop'), icon: 'stop', danger: true, disabled: !j.cancelable && t('cmNotStoppable'), action: () => api('DELETE', `/api/jobs/${enc(j.id)}`).then(() => toast(t('jobStopping'))) },
    { label: t('close'), icon: 'x', action: () => { j.dismissed = true; renderDock(); } }];
});

// ── Tetikleyiciler (config.Trigger, cmd_triggers.go) ─────────────────────────
// Kart başına mevcut alan bileşenleri, anahtarlar "<prefix>.<i>.<alan>" → readForm diziye çevirir.
// Ekle / sil kartları formdan okuyup yeniden çizer (indeksler hep 0..n-1 kalır).
const TRIGGER_TEMPLATES = {
  migrateSsh: { name: 'migrate', when: 'after', paths: ['database/migrations/'], remote: ['php artisan migrate --force'] },
  migrateHttp: { name: 'migrate', when: 'after', paths: ['database/migrations/'], http_url: 'https://example.com/deploy/migrate?token=…', http_method: 'POST' },
  composer: { name: 'composer', when: 'after', paths: ['composer.json', 'composer.lock'], remote: ['composer install --no-dev --optimize-autoloader'] },
  cache: { name: 'cache', when: 'after', paths: ['config/', 'routes/'], remote: ['php artisan optimize:clear'] },
  npm: { name: 'build', when: 'before', paths: ['resources/js/', 'resources/css/'], run: ['npm run build'] },
  notify: { name: 'bildirim', when: 'after', paths: [], notify: '' },
  blank: { name: '', when: 'after', paths: [] },
};
function triggerCard(prefix, tr, i, o) {
  const k = (f) => `${prefix}.${i}.${f}`;
  const on = tr.enabled !== false;
  const whenOpts = [['after', t('trAfter')], ['before', t('trBefore')]];
  const methods = ['POST', 'GET', 'PUT', 'PATCH', 'DELETE'].map((m) => [m, m]);
  const acts = [n2(tr.run) && t('trActRun'), n2(tr.remote) && 'SSH', tr.http_url && 'HTTP', tr.notify && t('trActNotify')].filter(Boolean).join(' · ') || t('trNoAction');
  return `<div class="tetik-kart ${on ? '' : 'kapali'}" data-tidx="${i}">
    <div class="tetik-bas">
      <span class="tetik-ad">${icon('zap', 'acc-t')}<b>${esc(tr.name || t('trUnnamed'))}</b><span class="faint">${esc((tr.paths || []).join(', ') || t('trAnyFile'))} → ${esc(acts)}</span></span>
      <span class="aksiyonlar"><button class="btn kucuk" data-ttest="${i}" title="${t('trTestHint')}">${icon('play')}${t('trTest')}</button>
        <button class="ikon-btn tehlike" data-tdel="${i}" title="${t('delete')}">${icon('trash')}</button></span></div>
    <div class="alanlar">${fText(k('name'), t('fName'), tr.name, { ph: 'migrate' })}${fSelect(k('when'), t('trWhen'), tr.when || 'after', whenOpts, { hint: t('trWhenHint') })}
      ${fBool(k('enabled'), t('fEnabled'), on)}</div>
    <div class="alanlar">${fArea(k('paths'), t('trPaths'), tr.paths, { tree: 'paths', chips: ['database/migrations/', '*.sql', 'composer.lock', 'package.json', 'config/', '.env'], hint: t('trPathsHint') })}</div>
    <div class="alanlar">${fArea(k('run'), t('trRun'), tr.run, { wide: false, hint: t('trRunHint') })}${fArea(k('remote'), t('trRemote'), tr.remote, { wide: false, hint: o.sftp ? t('trRemoteHint') : t('trRemoteFtp') })}</div>
    <div class="alanlar">${fText(k('http_url'), t('trHttpUrl'), tr.http_url, { mono: true, wide: true, ph: 'https://site.com/deploy/migrate?token=…', hint: t('trVarsHint') })}
      ${fSelect(k('http_method'), t('trHttpMethod'), (tr.http_method || 'POST').toUpperCase(), methods)}</div>
    <div class="alanlar">${fArea(k('http_headers'), t('trHttpHeaders'), tr.http_headers, { wide: false, ph: 'Authorization: Bearer …' })}
      <div class="alan"><label>${t('trHttpBody')}</label><textarea data-k="${k('http_body')}" data-ty="text" class="mono" rows="4" spellcheck="false" placeholder='{"files": {files_json}}'>${esc(tr.http_body || '')}</textarea><small>${t('trHttpBodyHint')}</small></div></div>
    <div class="alanlar">${fText(k('notify'), t('trNotify'), tr.notify, { mono: true, wide: true, ph: 'https://hooks.slack.com/…', whtest: true })}</div>
  </div>`;
}
const n2 = (a) => (a || []).filter((x) => String(x).trim()).length;
function triggersInner(prefix, list, o) {
  return `<input type="hidden" data-k="${prefix}" data-ty="emptylist">
    ${(list || []).map((tr, i) => triggerCard(prefix, tr, i, o)).join('') || `<div class="bos" style="padding:14px">${t('trEmpty')}</div>`}
    <div class="aksiyonlar"><select data-ttmpl style="width:auto">${Object.keys(TRIGGER_TEMPLATES).map((x) => `<option value="${x}">${t('trTpl_' + x)}</option>`).join('')}</select>
      <button class="btn" data-tadd>${icon('plus')}${t('trAdd')}</button></div>`;
}
function triggersSection(prefix, list, o = {}) {
  const sum = (list || []).length ? (list || []).map((x) => esc(x.name)).join(', ') : t('foldDefault');
  return section('zap', t('secTriggers'), o.global ? t('secTriggersDescGlobal') : t('secTriggersDesc'),
    `<div class="tetikler" data-tprefix="${prefix}" data-tservers="${esc(JSON.stringify(o.servers || []))}" data-tsftp="${o.sftp ? 1 : 0}">${triggersInner(prefix, list, o)}</div>`,
    { fold: true, open: (list || []).length > 0, summary: sum });
}
// Formdaki tetikleyici listesini oku (kaydedilmemiş hâliyle)
function readTriggers(box) {
  const obj = readForm(box, {});
  let v = obj;
  for (const part of box.dataset.tprefix.split('.')) v = v ? v[part] : undefined;
  return (v || []).filter(Boolean);
}
function redrawTriggers(box, list) {
  box.innerHTML = triggersInner(box.dataset.tprefix, list, { sftp: box.dataset.tsftp === '1' });
  bindFormTools(box);
  box.dispatchEvent(new Event('change', { bubbles: true })); // kaydedilmemiş form koruması
}
document.addEventListener('click', async (e) => {
  const box = e.target.closest && e.target.closest('.tetikler');
  if (!box) return;
  const add = e.target.closest('[data-tadd]'), del = e.target.closest('[data-tdel]'), test = e.target.closest('[data-ttest]');
  if (!add && !del && !test) return;
  e.preventDefault();
  const list = readTriggers(box);
  if (add) {
    const tpl = JSON.parse(JSON.stringify(TRIGGER_TEMPLATES[$('[data-ttmpl]', box).value] || TRIGGER_TEMPLATES.blank));
    // aynı ad varsa sonuna sayı
    let name = tpl.name, n = 2;
    while (name && list.some((x) => (x.name || '').toLowerCase() === name.toLowerCase())) name = `${tpl.name}-${n++}`;
    tpl.name = name;
    redrawTriggers(box, [...list, tpl]);
    const cards = $$('.tetik-kart', box);
    if (cards.length) { cards[cards.length - 1].scrollIntoView({ block: 'nearest' }); $('input[data-k$=".name"]', cards[cards.length - 1]).focus(); }
    return;
  }
  const i = +(del || test).dataset[del ? 'tdel' : 'ttest'];
  if (del) {
    if (!(await confirmBox(t('delete'), t('trDelConfirm', list[i].name || t('trUnnamed')), t('delete'), true))) return;
    list.splice(i, 1);
    redrawTriggers(box, list);
    return;
  }
  // Dene: formdaki (kaydedilmemiş) tanımla, sync olmadan; örnek dosyalar sorulur
  const form = box.closest('[data-p]');
  const P = form.dataset.p;
  const servers = JSON.parse(box.dataset.tservers || '[]');
  const res = await new Promise((done) => {
    modal({
      title: `${icon('play')}${t('trTest')} — ${esc(list[i].name || t('trUnnamed'))}`, onClose: () => done(null),
      body: `<div class="form">${servers.length > 1 ? fSelect('srv', t('trTestServer'), servers[0], servers.map((x) => [x, x])) : ''}
        ${fArea('files', t('trTestFiles'), (list[i].paths || []).filter((x) => !/[*?]/.test(x)).slice(0, 1).map((x) => (x.endsWith('/') ? x + 'ornek.php' : x)), { ph: 'database/migrations/2026_01_01_ornek.php' })}
        <p class="ipucu">${t('trTestNote')}</p></div>`,
      actions: [{ label: t('cancel'), cls: 'hayalet' }, { label: icon('play') + t('trTest'), cls: 'mavi', run: (r) => done(readForm(r, {})) }],
    });
  });
  if (!res) return;
  const srv = res.srv || servers[0];
  if (!srv) { fail(new Error(t('trTestNoServer'))); return; }
  const tr = { ...list[i] };
  if (tr.http_method === 'POST') delete tr.http_method;
  try {
    await startJobs(P, { kind: 'trigger', servers: [srv], target: tr.name, trigger: tr, paths: res.files || [] }, `${t('trTest')}: ${tr.name}`);
  } catch (err) { fail(err); }
});
// Protokole özel alanlar: FTP'de SSH anahtarı, SFTP'de passive/EPSV/NAT görünmez (değerleri korunur)
const PROTO_FIELDS = { key_file: 'sftp', passive: 'ftp ftps', disable_epsv: 'ftp ftps', nat_workaround: 'ftp ftps', tls_skip_verify: 'ftps' };
function applyProtoVis(root) {
  $$('select[data-k$="protocol"]', root).forEach((sel) => {
    const prefix = sel.dataset.k.slice(0, -'protocol'.length);
    const scope = sel.closest('.form, .modal, form') || root;
    const update = () => {
      const pr = sel.value || 'ftp';
      for (const [k, ps] of Object.entries(PROTO_FIELDS)) {
        const el = $(`[data-k="${prefix}${k}"]`, scope);
        const box = el && el.closest('.alan, .anahtar');
        if (box) box.hidden = !ps.split(' ').includes(pr);
      }
      $$('[data-proto]', scope).forEach((b) => (b.hidden = !b.dataset.proto.split(' ').includes(pr)));
    };
    if (!sel.dataset.pv) { sel.dataset.pv = '1'; sel.addEventListener('change', update); }
    update();
  });
}

// Form içindeki Gözat / ağaç / webhook test düğmelerini bağlar.
// Bağlam form kökündeki data-p (proje), data-ppath (proje yolu), data-srv (sunucu) özniteliklerinden okunur.
function bindFormTools(root) {
  applyProtoVis(root);
  const ctx = root.closest('[data-ppath]') || root.querySelector('[data-ppath]') || root;
  const P = ctx.dataset.p, ppath = ctx.dataset.ppath || '', srv = ctx.dataset.srv || '';
  const input = (k) => root.querySelector(`[data-k="${k}"]`);
  $$('[data-browse]:not([data-browse="remote"])', root).forEach((b) => (b.onclick = async (e) => {
    e.preventDefault();
    const inp = input(b.dataset.for), kind = b.dataset.browse;
    let initial = inp.value.trim();
    if (kind === 'reldir') initial = !initial || initial === '.' ? ppath : (/^([a-zA-Z]:|[\\/])/.test(initial) ? initial : ppath + '\\' + initial.replace(/\//g, '\\'));
    if (kind === 'file' && !initial && S.meta) initial = S.meta.home + '\\.ssh';
    if (kind === 'file' && initial && /\.[a-z0-9]+$/i.test(initial)) initial = initial.replace(/[\\/][^\\/]*$/, '');
    const picked = await pickPath({ mode: kind === 'file' ? 'file' : 'dir', initial });
    if (picked == null) return;
    inp.value = kind === 'reldir' ? relToProject(ppath, picked) : picked;
    inp.dispatchEvent(new Event('change', { bubbles: true }));
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
    else if (ty === 'emptylist') v = [];
    else v = el.value.trim();
    // "triggers.0.name" → dizi (sonraki parça sayıysa)
    const path = k.split('.');
    let o = obj;
    for (let i = 0; i < path.length - 1; i++) o = o[path[i]] ??= (/^\d+$/.test(path[i + 1]) ? [] : {});
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
  applyProtoVis(root);
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
  const savedManual = !isNew && !s.connection && !!s.host; // kayıtlı hali profilsiz → "Profile çevir"
  const defPath = cfg.project.default_path || '.';
  const globalBw = cfg.sync.bandwidth_limit ? cfg.sync.bandwidth_limit + ' KB/s' : '∞';

  bind((root) => {
    $('#proj-aks', root).innerHTML = '';
    bindProtoPort(root);
    bindFormTools($('#srv-form', root));
    trackDirty($('#srv-form', root));
    // Kayıtlı bilgiler profile taşınır (formdaki kaydedilmemiş değişiklikler değil) → sayfa yeniden yüklenir
    $('#srv-form', root).addEventListener('click', async (e) => {
      const b = e.target.closest('[data-toprofile]');
      if (!b) return;
      b.disabled = true;
      try { await extractConn(P, name); render(true); } catch (err) { fail(err); b.disabled = false; }
    });
    $('[data-k="connection"]', root).onchange = (e) => {
      // Profil değişince credential alanlarını yeniden çiz
      const c = cfg.connections.find((x) => x.name === e.target.value);
      readForm($('#srv-form', root), s);
      s.connection = e.target.value;
      $('#conn-box', root).innerHTML = connBox(c);
      $('#tunnel-box', root).innerHTML = tunnelBox(c);
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
      try { await api('PUT', `/api/p/${enc(P)}/config`, next); clearDirty(); toast(t('saved')); await loadProjects(); go(pHash(P, 'servers')); } catch (e) { fail(e); }
    };
    // Uzak dizin → Gözat: formdaki bilgilerle bağlanıp sunucudaki klasörlerden seçtirir (kaydetmeden)
    const rb = $('[data-browse="remote"]', root);
    if (rb) rb.onclick = async (e) => {
      e.preventDefault();
      const inp = $('[data-k="remote_path"]', root);
      const picked = await pickRemoteDir({ initial: inp.value.trim() || '/', list: (path, create) => api('POST', `/api/p/${enc(P)}/browse-dirs`, { server: collect(), connections: cfg.connections, path, create }) });
      if (picked != null) inp.value = picked;
    };
    $('#b-test', root).onclick = async (e) => {
      const b = e.currentTarget; b.disabled = true; const old = b.innerHTML; b.innerHTML = icon('refresh', 'spin') + t('testing');
      try { const r = await api('POST', `/api/p/${enc(P)}/test`, { server: collect(), connections: cfg.connections }); toast(t('testOk', r.ms, r.entries) + (r.list_error ? ' — ' + r.list_error : ''), r.list_error ? 'bad' : 'ok'); }
      catch (err) { fail(err); } finally { b.disabled = false; b.innerHTML = old; }
    };
    const det = $('#b-detect', root);
    if (det) det.onclick = () => startJobs(P, { kind: 'detect', servers: [s.name] }, t('detect')).then((ids) => {
      const j = S.jobs.get(ids[0]);
      if (j) j.doneCbs.push((jj) => { const inp = $('[data-k="max_connections"]', root); if (inp && jj.result && jj.result[s.name]) inp.value = jj.result[s.name]; });
    }).catch(fail);
    const del = $('#b-del', root);
    if (del) del.onclick = async () => {
      if (!(await confirmBox(t('delete'), esc(t('deleteServerConfirm', s.name)), t('delete'), true))) return;
      const next = JSON.parse(JSON.stringify(cfg));
      next.servers.splice(idx, 1);
      try { await api('PUT', `/api/p/${enc(P)}/config`, next); clearDirty(); toast(t('saved')); await loadProjects(); go(pHash(P, 'servers')); } catch (e) { fail(e); }
    };
  });

  function connBox(c) {
    // Profil seçiliyse alanlar hiç çizilmez (collect zaten profilden gelenleri siler) — yalnızca özet
    if (!c) return connFields(s, false) + (savedManual ? `<div class="aksiyonlar" style="margin-top:12px"><button class="btn kucuk" data-toprofile title="${esc(t('connToProfileHint'))}">${icon('link')}${t('connToProfile')}</button></div>` : '');
    const proto = (c.protocol || 'ftp').toLowerCase();
    return `<div class="kilit">${icon('lock')}<span style="flex:1">${esc(t('lockedByProfile', c.name))}
      <span class="mono" style="display:block;color:var(--muted);margin-top:2px">${proto}://${esc(c.user ? c.user + '@' : '')}${esc(c.host)}:${c.port || (proto === 'sftp' ? 22 : 21)}</span></span>
      <a class="btn kucuk" href="${pHash(P, 'connection', c.name)}">${icon('edit')}${t('editProfile')}</a></div>`;
  }
  function tunnelBox(c) {
    return c ? `<div class="kilit">${icon('lock')}${esc(t('tunnelByProfile', c.name))}${c.tunnel_host ? ` <span class="mono">${esc(c.tunnel_host)}</span>` : ''}</div>` : tunnelFields(s, false);
  }

  const n = (a) => (a || []).filter(Boolean).length;
  const onOff = (pairs) => pairs.filter(([v]) => v).map(([, l]) => l).join(' · ') || t('foldDefault');
  const tun = (conn || s).tunnel_host;
  const sumPerf = `${s.max_connections || 1} ${t('foldConn')} · ${s.max_retries || 0} ${t('foldRetry')} · ${s.bandwidth_limit ? s.bandwidth_limit + ' KB/s' : globalBw}`;
  const sumUpload = onOff([[s.minify, 'minify'], [s.obfuscate, 'obfuscate'], [s.no_backup, t('fNoBackup')], [s.direct_upload, t('fDirect')], [s.webhook, 'webhook']]);
  const sumFilt = `include ${n(s.include)} · exclude ${n(s.exclude)} · protect ${n(s.protect)} · ${t('foldBlock')} ${n(s.block_extensions)}`;
  const sumHooks = `pre ${n(s.pre_sync)} · post ${n(s.post_sync)}`;
  const sumRemote = `pre ${n(s.remote_pre_sync)} · post ${n(s.remote_post_sync)}`;
  return `<a class="faint" href="${pHash(P, 'servers')}" style="display:inline-flex;gap:6px;align-items:center;margin-bottom:12px">${icon('left')}${t('back')}</a>
  <h2 style="margin-top:0">${isNew ? t('serverNew') : t('serverEdit') + ' — ' + esc(s.name)}</h2>
  <div class="form" id="srv-form" data-p="${esc(P)}" data-ppath="${esc(proj.path)}" data-srv="${esc(isNew ? '' : s.name)}">
    ${section('server', t('secGeneral'), '', `<div class="alanlar">${fText('name', t('fName'), s.name, { ph: 'production' })}${fBool('enabled', t('fEnabled'), s.enabled, t('fEnabledHint'))}</div>`)}
    ${section('link', t('secConnection'), t('secConnectionDesc'), `<div class="alanlar" style="margin-bottom:14px">${fSelect('connection', t('fConnection'), s.connection || '', [['', t('manual')], ...cfg.connections.map((c) => [c.name, `${c.name} — ${c.host}`])])}</div><div id="conn-box">${connBox(conn)}</div>`)}
    ${section('folder', t('secPaths'), '', `<div class="alanlar">${fText('remote_path', t('fRemote'), s.remote_path, { mono: true, hint: t('fRemoteHint'), browse: 'remote' })}${fText('local_path', t('fLocal'), s.local_path, { mono: true, ph: defPath, hint: t('fLocalHint', esc(defPath)), browse: 'reldir' })}</div>`)}
    <div class="ara-baslik">${t('foldAdvanced')}</div>
    ${section('tunnel', t('secTunnel'), t('secTunnelDesc'), `<div id="tunnel-box">${tunnelBox(conn)}</div>`, { fold: true, open: !!tun, summary: tun ? esc(tun) : t('foldOff') })}
    ${section('zap', t('secPerf'), '', `<div class="alanlar uc">
        <div class="alan"><label>${t('fMaxConn')}</label><div class="girdi-grup"><input type="number" data-k="max_connections" data-ty="int" value="${s.max_connections || 1}" min="1">${isNew ? '' : `<button class="btn" id="b-detect">${icon('gauge')}${t('measure')}</button>`}</div><small>${t('fMaxConnHint')}</small></div>
        ${fNum('max_retries', t('fRetries'), s.max_retries)}
        ${fNum('bandwidth_limit', t('fBandwidth'), s.bandwidth_limit, { hint: t('fBandwidthHint', globalBw) })}</div>`, { fold: true, summary: sumPerf })}
    ${section('upload', t('secUpload'), '', `<div class="alanlar">
        ${fBool('minify', t('fMinify'), s.minify, t('fMinifyHint'))}${fBool('obfuscate', t('fObfuscate'), s.obfuscate, t('fObfuscateHint'))}
        ${fBool('no_backup', t('fNoBackup'), s.no_backup, t('fNoBackupHint'))}${fBool('direct_upload', t('fDirect'), s.direct_upload, t('fDirectHint'))}
        ${fText('webhook', t('fWebhook'), s.webhook, { wide: true, mono: true, hint: t('fWebhookHint'), ph: 'https://…', whtest: true })}</div>`, { fold: true, summary: sumUpload })}
    ${section('sliders', t('secFilters'), t('secFiltersDesc'), `<div class="alanlar">
        ${fArea('include', t('fInclude'), s.include, { wide: false, tree: 'paths', chips: 'include' })}${fArea('exclude', t('fExclude'), s.exclude, { wide: false, tree: 'paths', chips: 'exclude' })}
        ${fArea('protect', t('fProtect'), s.protect, { wide: false, tree: 'protect', chips: 'protect' })}${fArea('block_extensions', t('fBlock'), s.block_extensions, { ty: 'words', wide: false, hint: t('fBlockHint'), chips: 'block' })}</div>`, { fold: true, summary: sumFilt })}
    ${section('log', t('secHooks'), t('secHooksDesc'), `<div class="alanlar">${fArea('pre_sync', t('fPre'), s.pre_sync, { wide: false })}${fArea('post_sync', t('fPost'), s.post_sync, { wide: false })}</div>`, { fold: true, summary: sumHooks })}
    ${section('terminal', t('secRemote'), t('secRemoteDesc'), `<div class="alanlar">${fArea('remote_pre_sync', t('fRemotePre'), s.remote_pre_sync, { wide: false })}${fArea('remote_post_sync', t('fRemotePost'), s.remote_post_sync, { wide: false, hint: 'php artisan migrate --force' })}</div>`, { fold: true, summary: sumRemote, proto: 'sftp' })}
    ${triggersSection('triggers', s.triggers, { servers: isNew ? [] : [s.name], sftp: (conn || s).protocol === 'sftp' })}
  </div>
  <div class="form-alt"><div class="sol">${isNew ? '' : `<button class="btn tehlike" id="b-del">${icon('trash')}${t('delete')}</button>`}</div>
    <button class="btn" id="b-test">${icon('test')}${t('testConn')}</button>
    <button class="btn hayalet" onclick="history.back()">${t('cancel')}</button>
    <button class="btn birincil" id="b-save">${icon('check')}${t('save')}</button></div>`;
}

// ── Bağlantı profilleri ──────────────────────────────────────────────────────
async function tabConnections(proj) {
  const P = proj.name;
  const [cfg, libSt] = await Promise.all([getConfig(P), GET(`/api/p/${enc(P)}/connections/library`).catch(() => ({}))]);
  const manual = cfg.servers.filter((s) => !s.connection && s.host);
  bind((root) => {
    // Boş durumda aynı düğmeler ortadaki kartta — üstte tekrarlanmaz
    $('#proj-aks', root).innerHTML = !cfg.connections.length ? '' : `${manual.length ? `<button class="btn" data-fromsrv>${icon('server')}${t('connFromServer')}</button>` : ''}<button class="btn" data-fromlib>${icon('link')}${t('connFromLib')}</button><button class="btn birincil" data-addconn>${icon('plus')}${t('addConnection')}</button>`;
    $$('[data-addconn]', root).forEach((b) => (b.onclick = () => go(pHash(P, 'connection', 'new'))));
    $$('[data-fromlib]', root).forEach((b) => (b.onclick = () => connFromLibraryModal(P).catch(fail)));
    $$('[data-fromsrv]', root).forEach((b) => (b.onclick = () => connFromServerModal(P, manual)));
    $$('[data-row]', root).forEach((tr) => (tr.onclick = () => go(pHash(P, 'connection', tr.dataset.row))));
    $$('[data-connfiles]', root).forEach((b) => (b.onclick = (e) => { e.stopPropagation(); go(pHash(P, 'files', CONN_TARGET + b.dataset.connfiles)); }));
    $$('[data-tolib]', root).forEach((b) => (b.onclick = async (e) => {
      e.stopPropagation();
      const name = b.dataset.tolib;
      if ((libSt[name] || {}).name_taken && !(await confirmBox(t('connLibOverwriteTitle'), esc(t('connLibOverwrite', name)), t('overwrite'), true))) return;
      try { const r = await api('POST', `/api/p/${enc(P)}/connections/to-library`, { connection: name }); toast(t('connToLibDone', r.name)); render(true); } catch (err) { fail(err); }
    }));
  });
  if (!cfg.connections.length) {
    return `<p class="muted" style="margin-top:-6px">${t('connectionsSub')}</p><div class="bos"><div class="buyuk">${icon('link', 'lg')}</div><h3>${t('noConnections')}</h3><p>${t('noConnectionsSub')}</p>
      <div class="aksiyonlar" style="justify-content:center">${manual.length ? `<button class="btn" data-fromsrv>${icon('server')}${t('connFromServer')}</button>` : ''}<button class="btn" data-fromlib>${icon('link')}${t('connFromLib')}</button>
      <button class="btn birincil" data-addconn>${icon('plus')}${t('addConnection')}</button></div></div>`;
  }
  return `<p class="muted" style="margin-top:-6px">${t('connectionsSub')}</p><div class="tablo-kap"><table class="tablo"><thead><tr><th>${t('colName')}</th><th>${t('colProtocol')}</th><th>${t('colHost')}</th><th>${t('usedBy')}</th><th></th></tr></thead><tbody>
    ${cfg.connections.map((c) => {
      const used = cfg.servers.filter((s) => s.connection === c.name).map((s) => s.name);
      const proto = (c.protocol || 'ftp').toLowerCase();
      return `<tr data-row="${esc(c.name)}" class="tikla"><td><div class="ad-hucre">${icon('link')}<b>${esc(c.name)}</b></div></td><td><span class="pill mono ${proto}">${proto.toUpperCase()}</span></td>
        <td class="mono">${esc(c.user ? c.user + '@' : '')}${esc(c.host)}:${c.port}${c.tunnel_host ? ` <span class="pill vio">${icon('tunnel')}${t('tunnel')}</span>` : ''}</td><td class="muted">${esc(used.join(', ') || '—')}</td>
        <td class="sag" style="white-space:nowrap"><button class="btn kucuk hayalet" data-connfiles="${esc(c.name)}" title="${esc(t('filesConnTip', `${c.user ? c.user + '@' : ''}${c.host}`))}">${icon('folder')}${t('tabFiles')}</button> ${(() => {
          const lib = (libSt[c.name] || {}).library;
          return lib ? `<span class="pill ok" ${lib !== c.name ? `title="${esc(t('connInLibAs', lib))}"` : ''}>${icon('check')}${t('connInLib')}</span>`
            : `<button class="btn kucuk hayalet" data-tolib="${esc(c.name)}">${icon('download')}${t('connToLib')}</button>`;
        })()}</td></tr>`;
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
    trackDirty($('#conn-form', root));
    $('#b-save', root).onclick = async () => {
      const out = readForm($('#conn-form', root), JSON.parse(JSON.stringify(c)));
      if (out.protocol === '') delete out.protocol;
      const next = JSON.parse(JSON.stringify(cfg));
      if (isNew) next.connections.push(out);
      else {
        next.connections[idx] = out;
        if (out.name !== c.name) next.servers.forEach((s) => { if (s.connection === c.name) s.connection = out.name; });
      }
      try { await api('PUT', `/api/p/${enc(P)}/config`, next); clearDirty(); toast(t('saved')); go(pHash(P, 'connections')); } catch (e) { fail(e); }
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
      try { await api('PUT', `/api/p/${enc(P)}/config`, next); clearDirty(); toast(t('saved')); go(pHash(P, 'connections')); } catch (e) { fail(e); }
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
    trackDirty($('#set-form', root));
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
          clearDirty();
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
  const n = (a) => (a || []).filter(Boolean).length;
  const sumFilt = `include ${n(sy.include)} · exclude ${n(sy.exclude)} · protect ${n(sy.protect)} · ${t('foldBlock')} ${n(sy.block_extensions)}`;
  const sumUp = [sy.webhook && 'webhook', sy.bandwidth_limit && sy.bandwidth_limit + ' KB/s', sy.no_backup ? t('fNoBackup') : `${t('fBackupKeep')}: ${sy.backup_keep || 10}`, sy.direct_upload && t('fDirect'), cfg.first_sync && cfg.first_sync.full && t('fFirstFull')].filter(Boolean).join(' · ');
  return `<div class="form" id="set-form" data-p="${esc(P)}" data-ppath="${esc(proj.path)}" data-srv="">
    ${section('folder', t('secProject'), '', `<div class="alanlar">${fText('project.name', t('fProjectName'), cfg.project.name)}${fText('project.default_path', t('fDefaultPath'), cfg.project.default_path, { mono: true, ph: '.', hint: t('fDefaultPathHint'), browse: 'reldir' })}</div>`)}
    <div class="ara-baslik">${t('secGlobal')}</div>
    ${section('sliders', t('secFilters'), t('secGlobalDesc'), `<div class="alanlar">
        ${fArea('sync.include', t('fInclude'), sy.include, { wide: false, tree: 'paths', chips: 'include' })}${fArea('sync.exclude', t('fExclude'), sy.exclude, { wide: false, tree: 'paths', chips: 'exclude' })}
        ${fArea('sync.protect', t('fProtect'), sy.protect, { wide: false, tree: 'protect', chips: 'protect' })}${fArea('sync.block_extensions', t('fBlock'), sy.block_extensions, { ty: 'words', wide: false, hint: t('fBlockHint'), chips: 'block' })}
        ${fArea('sync.ignore_files', t('fIgnoreFiles'), sy.ignore_files, { wide: false, tree: 'paths', ph: '.gitignore', hint: t('fIgnoreFilesHint') })}
      </div>`, { fold: true, summary: sumFilt })}
    ${section('upload', t('secUpload'), '', `<div class="alanlar">
        ${fText('sync.webhook', t('fWebhook'), sy.webhook, { mono: true, ph: 'https://…', whtest: true })}
        ${fNum('sync.bandwidth_limit', t('fBandwidth'), sy.bandwidth_limit, { hint: '0 = ∞' })}
        ${fNum('sync.backup_keep', t('fBackupKeep'), sy.backup_keep || 10, { hint: t('fBackupKeepHint') })}
        ${fBool('sync.no_backup', t('fNoBackup'), sy.no_backup, t('fNoBackupHint'))}
        ${fBool('sync.direct_upload', t('fDirect'), sy.direct_upload, t('fDirectHint'))}
        ${fBool('first_sync.full', t('fFirstFull'), cfg.first_sync && cfg.first_sync.full, t('fFirstFullHint'))}
      </div>`, { fold: true, summary: sumUp })}
    ${section('log', t('secHooks'), t('secHooksDesc'), `<div class="alanlar">${fArea('sync.pre_sync', t('fPre'), sy.pre_sync, { wide: false })}${fArea('sync.post_sync', t('fPost'), sy.post_sync, { wide: false })}</div>`, { fold: true, summary: `pre ${n(sy.pre_sync)} · post ${n(sy.post_sync)}` })}
    ${triggersSection('sync.triggers', sy.triggers, { servers: cfg.servers.map((x) => x.name), global: true, sftp: true })}
    <div class="ara-baslik">${t('foldMore')}</div>
    ${section('x', t('ignoreTitle'), `${t('ignoreDesc')}${gitLines ? ' · ' + t('gitignoreNote', gitLines) : ''}`, `
      <div class="aksiyonlar" style="margin-bottom:10px">${ignTargets.length > 1 ? `<select id="ign-dir" style="width:auto">${ignTargets.map((g, i) => `<option value="${i}">${esc(g.dir)} — ${esc((g.servers || []).join(', '))}</option>`).join('')}</select>` : ''}
        <span class="mono faint" id="ign-file" style="font-size:12px">${ign.file ? esc(`${ign.dir === '.' ? '' : ign.dir + '/'}${ign.file}`) : t('ignoreNoFile')}</span>
        <select id="ign-tmpl" style="width:auto"><option value="">${t('ignoreTemplate')}</option>${Object.keys(tmpls).map((k) => `<option value="${esc(k)}">${esc(k)}</option>`).join('')}</select>
        <span style="margin-left:auto"></span><button class="btn" id="ign-save">${icon('check')}${t('save')}</button></div>
      <textarea id="ign-text" rows="10" spellcheck="false" ${ign.file ? '' : 'disabled'}>${esc(ign.content || '')}</textarea>`, { fold: true, summary: ign.file ? esc(ign.file) : '' })}
    ${section('lock', t('secSecurity'), '', `
      <label class="anahtar"><span class="metin"><b>${t('keyringTitle')}</b><small>${t('keyringHint')}</small></span><span class="switch"><input type="checkbox" id="kr" ${cfg.project.keyring ? 'checked' : ''}><i></i></span></label>`, { fold: true, summary: cfg.project.keyring ? 'keyring' : t('foldOff') })}
    ${projToolsHTML()}
    ${section('layers', t('secRegistry'), t('registryDesc'), `<div class="aksiyonlar"><button class="btn" id="b-rename">${icon('edit')}${t('renameProject')}</button></div>`, { fold: true })}
    ${projRemoveHTML()}
  </div>
  <div class="form-alt"><button class="btn birincil" id="b-save">${icon('check')}${t('save')}</button></div>`;
}

// ── Dosyalar ─────────────────────────────────────────────────────────────────
// Dosya tarayıcısında profil yerine doğrudan bağlantıyla gezme: "server" değeri "@conn:<ad>" (browseTarget).
// Bağlantının uzak dizini ve yerel karşılığı yok → kökten gezilir; freeze / karşılaştırma / düzenleme yalnızca profilde.
const CONN_TARGET = '@conn:';
const isConnTarget = (s) => String(s || '').startsWith(CONN_TARGET);

async function tabFiles(proj, srvName, relPath) {
  const P = proj.name;
  const cfg = await getConfig(P);
  const conns = cfg.connections || [];
  if (!cfg.servers.length && !conns.length) return `<div class="bos"><h3>${t('noServers')}</h3></div>`;
  const connObj = isConnTarget(srvName) ? conns.find((c) => CONN_TARGET + c.name === srvName) : null;
  const srvObj = connObj ? null : cfg.servers.find((s) => s.name === srvName) || cfg.servers.find((s) => s.enabled) || cfg.servers[0];
  const srv = srvObj ? srvObj.name : CONN_TARGET + (connObj || conns[0]).name;
  const viaConn = isConnTarget(srv);
  const rel = relPath || '';
  const q = `server=${enc(srv)}`;
  // Daha önce açılan klasör önbellekten anında çizilir, sunucudan arka planda tazelenir; fark varsa
  // sayfa sessizce yenilenir (FTP listesi ağ gecikmesi kadar sürer — her girişte beklemeyelim)
  const lsKey = `${P}|${srv}|${rel}`;
  const fetchLs = () => GET(`/api/p/${enc(P)}/remote/ls?${q}&path=${enc(rel)}`);
  let data = S.lsCache.get(lsKey);
  const fromCache = !!data;
  if (!data) {
    try { data = await fetchLs(); S.lsCache.set(lsKey, data); }
    catch (e) { data = { error: e.message, entries: [] }; }
  }
  const reload = () => { S.lsCache.delete(lsKey); render(true); };
  const frozenSet = new Set(viaConn ? [] : await GET(`/api/p/${enc(P)}/frozen?${q}`).catch(() => []));
  const hashFor = (p) => `${pHash(P, 'files', srv)}/${enc(p)}`;
  const full = (name) => (rel ? rel + '/' : '') + name;
  const crumbs = rel.split('/').filter(Boolean);
  const entries = data.entries || [];
  const sel = new Set();

  // Freeze ekle/çıkar — CLI tarayıcısıyla aynı çekirdek: klasör tek "klasör/" kaydı (altındaki her şey);
  // hepsi frozen ise çözer
  async function toggleFreeze(items) {
    try {
      const r = await api('POST', `/api/p/${enc(P)}/frozen/toggle?${q}`, { items: items.map((it) => ({ path: it.path, dir: it.type === 'dir' })) });
      toast(r.frozen ? t('frozenAdded', r.n) : t('frozenRemoved', r.n));
    } catch (e) { fail(e); }
    reload();
  }
  async function removeItems(items) {
    const msg = items.length === 1 ? (items[0].type === 'dir' ? t('deleteDirConfirm', items[0].path) : t('deleteRemoteConfirm', items[0].path)) : t('deleteManyConfirm', items.length);
    const hasDir = items.some((it) => it.type === 'dir');
    // CLI `remote rm --recursive` karşılığı: işaret kalkarsa yalnızca boş klasörler silinir
    const recursive = await new Promise((res) => {
      const m = modal({
        title: t('delete'), onClose: () => res(null),
        body: `<p>${esc(msg)}</p>${hasDir ? `<label class="anahtar"><span class="metin"><b>${t('deleteRecursive')}</b><small>${t('deleteRecursiveHint')}</small></span><span class="switch"><input type="checkbox" id="rm-rec" checked><i></i></span></label>` : ''}`,
        actions: [{ label: t('cancel'), cls: 'hayalet' }, { label: t('delete'), cls: 'tehlike', run: (r) => { const c = $('#rm-rec', r); res(c ? c.checked : true); } }],
      });
      return m;
    });
    if (recursive === null) return;
    // Silme bir iş: kendi bağlantısında çalışır (liste donmaz), sağ altta canlı görünür, Durdur ile
    // kesilir; bitince bu klasör tazelenir (onJobFinished). Silinmekte olan satırlar soluklaşır.
    try {
      const label = items.length === 1 ? `${t('delete')}: ${items[0].path}` : `${t('delete')}: ${t('selectedN', items.length)}`;
      await startJobs(P, { kind: 'rm', servers: [srv], paths: items.map((it) => it.path), recursive }, label);
      const names = new Set(items.map((it) => it.path.split('/').pop()));
      $$('#f-body tr[data-entry]').forEach((tr) => { if (names.has(tr.dataset.entry)) tr.classList.add('siliniyor'); });
      S.lsCache.delete(`${P}|${srv}|${rel}`);
    } catch (err) { fail(err); }
  }
  // Zip / zipten çıkar: CLI `remote zip|unzip` ile aynı çekirdek (remote_zip.go), iş olarak çalışır —
  // sağ altta canlı, Durdur'la kesilir, bitince bu klasör tazelenir (onJobFinished).
  async function zipItems(items) {
    const base = (p) => p.split('/').pop();
    const def = items.length === 1 ? base(items[0].path).replace(/\.[^.]*$/, '') + '.zip' : (rel ? base(rel) : 'archive') + '.zip';
    const name = await new Promise((res) => {
      modal({
        title: t('zipTitle'), onClose: () => res(null),
        body: `<div class="form"><div class="alan"><label>${t('zipName')}</label><input type="text" id="zip-ad" value="${esc(def)}" spellcheck="false"></div></div>
          <p class="ipucu">${esc(t('zipHint', items.length))}</p>`,
        actions: [{ label: t('cancel'), cls: 'hayalet' }, { label: t('zipSel'), cls: 'mavi', run: (r) => res($('#zip-ad', r).value.trim() || def) }],
      });
    });
    if (name === null) return;
    try {
      await startJobs(P, { kind: 'zip', servers: [srv], paths: items.map((it) => it.path), target: name }, `${t('zipSel')}: ${name}`);
      S.lsCache.delete(`${P}|${srv}|${rel}`);
    } catch (err) { fail(err); }
  }
  async function unzipItem(it) {
    const name = it.path.split('/').pop();
    const folder = name.replace(/\.zip$/i, '');
    const dir = it.path.split('/').slice(0, -1).join('/');
    const target = await new Promise((res) => {
      modal({
        title: `${t('unzip')} — ${esc(name)}`, onClose: () => res(null),
        body: `<div class="ed-secim">
            <label><input type="radio" name="uz" value="sub" checked><span><b>${esc(t('unzipTo', folder))}</b></span></label>
            <label><input type="radio" name="uz" value="here"><span><b>${t('unzipHere')}</b></span></label></div>
          <p class="ipucu">${t('unzipHint')}</p>`,
        actions: [{ label: t('cancel'), cls: 'hayalet' }, { label: t('unzip'), cls: 'mavi', run: (r) => res($('input[name=uz]:checked', r).value === 'here' ? dir : (dir ? dir + '/' : '') + folder) }],
      });
    });
    if (target === null) return;
    try {
      await startJobs(P, { kind: 'unzip', servers: [srv], paths: [it.path], target }, `${t('unzip')}: ${name}`);
      S.lsCache.delete(`${P}|${srv}|${rel}`);
    } catch (err) { fail(err); }
  }
  // Taşı: sunucudaki klasörlerden gezerek seç (pickRemoteDir; yol kutusu elle yazmaya da izin verir).
  // Seçicide "/" = sunucunun uzak dizini; liste dosya tarayıcısının remote/ls ucundan gelir.
  async function moveItems(items) {
    const toRel = (p) => String(p || '').replace(/^\/+|\/+$/g, '');
    let created = false; // seçicide klasör/dosya oluşturulduysa vazgeçilse de liste tazelenir
    const target = await pickRemoteDir({
      title: `${t('moveTo')} — ${items.length === 1 ? esc(items[0].path) : t('selectedN', items.length)}`,
      initial: '/' + rel,
      list: async (path, create) => {
        const r = toRel(path);
        let name = '';
        if (create) {
          name = (await api('POST', `/api/p/${enc(P)}/remote/create?${q}&path=${enc(r)}`, create)).name;
          created = true;
          S.lsCache.delete(`${P}|${srv}|${r}`);
        }
        const d = await GET(`/api/p/${enc(P)}/remote/ls?${q}&path=${enc(r)}`);
        if (d.error) throw new Error(d.error);
        const es = d.entries || [];
        return { path: '/' + r, dirs: es.filter((e) => e.type === 'dir').map((e) => e.name), files: es.filter((e) => e.type !== 'dir').map((e) => e.name), created: name };
      },
    });
    if (target == null) { if (created) reload(); return; }
    const to = toRel(target);
    try {
      for (const it of items) {
        const name = it.path.split('/').pop();
        const dest = (to ? to + '/' : '') + name;
        if (dest === it.path) continue;
        await api('POST', `/api/p/${enc(P)}/remote/rename?${q}&from=${enc(it.path)}&to=${enc(dest)}`);
      }
      toast(t('moved'));
    } catch (e) { fail(e); }
    reload();
  }
  function renameItem(it) {
    modal({
      title: `${icon('edit')}${t('rename')}`,
      body: `<div class="alan"><label>${t('newPath')}</label><input type="text" class="mono" id="rn" value="${esc(it.path)}"></div>`,
      actions: [{ label: t('cancel'), cls: 'hayalet' }, { label: t('save'), cls: 'birincil', run: async (root) => {
        const to = $('#rn', root).value.trim().replace(/^\/+/, '');
        if (!to || to === it.path) return;
        await api('POST', `/api/p/${enc(P)}/remote/rename?${q}&from=${enc(it.path)}&to=${enc(to)}`);
        toast(t('moved')); reload();
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
      else {
        const parent = p.split('/').slice(0, -1).join('/');
        S.afterRender = () => previewFile(P, srv, p);
        go(parent ? hashFor(parent) : pHash(P, 'files', srv));
      }
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
    $('#b-rf', root).onclick = () => reload();
    $$('[data-fsrv]', root).forEach((sel) => (sel.onchange = (e) => { if (e.target.value) go(pHash(P, 'files', e.target.value)); }));
    $$('[data-crumb]', root).forEach((b) => (b.onclick = () => go(b.dataset.crumb === '' ? pHash(P, 'files', srv) : hashFor(b.dataset.crumb))));
    const filter = $('#f-filter', root);
    // Binlerce girdide her tuşta yeniden çizmesin: yazma durunca (120 ms) filtrele
    let fltTimer = 0;
    filter.oninput = () => { clearTimeout(fltTimer); fltTimer = setTimeout(() => applyFilter(filter.value), 120); };
    filter.onkeydown = (e) => { if (e.key === 'Enter' && filter.value.trim()) searchAll(filter.value.trim()); };
    $('#f-search', root).onclick = () => { if (filter.value.trim()) searchAll(filter.value.trim()); else filter.focus(); };
    const up = $('#f-upload-input', root);
    $('#f-upload', root).onclick = () => up.click();
    // items: [{file, path}] — path hedef klasöre göreceli (sürüklenen klasörlerde alt yol); dir: hedef klasör
    const uploadItems = async (items, dir) => {
      if (!items.length) return;
      const fd = new FormData();
      // yol kendi dosyasından ÖNCE gelmeli — sunucu parçaları akış hâlinde okur
      items.forEach((it) => { fd.append('paths', it.path); fd.append('files', it.file, it.file.name); });
      const b = $('#f-upload', root);
      const label = (pct) => { b.innerHTML = icon('refresh', 'spin') + t('upload') + (pct != null ? ` ${pct}%` : ''); };
      b.disabled = true; label();
      const ghost = uploadGhosts(items, dir);
      try {
        // api() JSON gövde gönderir; dosya için FormData + XHR (yükleme yüzdesi) — hata işleme api() ile aynı
        const r = await new Promise((res, rej) => {
          const x = new XMLHttpRequest();
          x.open('POST', `/api/p/${enc(P)}/remote/upload?${q}&path=${enc(dir)}`);
          x.upload.onprogress = (e) => { if (e.lengthComputable && b.isConnected) { label(Math.floor((e.loaded / e.total) * 100)); ghost.progress(e.loaded); } };
          x.onerror = () => rej(new Error(t('sessionLost')));
          x.onload = () => {
            let d = null;
            try { d = JSON.parse(x.responseText); } catch { /* boş / JSON değil */ }
            if (x.status >= 400 || !d || d.ok === false) rej(new Error((d && d.error) || `HTTP ${x.status}`)); else res(d);
          };
          x.send(fd);
        });
        ghost.progress(Infinity);
        toast(t('uploaded', (r.data || []).length));
      } catch (err) { ghost.failed(); fail(err); }
      if (dir !== rel) S.lsCache.delete(`${P}|${srv}|${dir}`);
      reload();
    };
    up.onchange = () => uploadItems([...up.files].map((f) => ({ file: f, path: f.name })), rel);
    bindDropUpload(root, (items, dirName) => uploadItems(items, dirName ? full(dirName) : rel));
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
        toast(t('created')); reload();
      } }],
    });
    // Yeni (boş) dosya → oluşturulup doğrudan editörde açılır (seçicilerdeki "Yeni dosya" ile aynı uç)
    $('#f-mkfile', root).onclick = () => modal({
      title: `${icon('file')}${t('pkNewFile')}`, body: `<div class="alan"><label>${t('pkNewFilePh')}</label><input type="text" class="mono" id="nfl"></div>`,
      actions: [{ label: t('cancel'), cls: 'hayalet' }, { label: t('create'), cls: 'birincil', run: async (r2) => {
        const n = $('#nfl', r2).value.trim();
        if (!n) return false;
        const r = await api('POST', `/api/p/${enc(P)}/remote/create?${q}&path=${enc(rel)}`, { kind: 'file', name: n });
        toast(t('pkCreated', r.name)); reload();
        editFile(P, srv, full(r.name));
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
    // Önceki çizimin gözlemcisi bırakılır (her yeniden çizimde biri daha birikiyordu)
    if (S.filesObs) { S.filesObs.disconnect(); S.filesObs = null; }
    if (more) {
      more.onclick = renderMore;
      S.filesObs = new IntersectionObserver((ents) => { if (ents.some((x) => x.isIntersecting) && document.body.contains(more)) renderMore(); }, { rootMargin: '600px' });
      S.filesObs.observe(more);
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
    $('[data-sel="zip"]', bar).onclick = () => zipItems(items());
    if (!viaConn) $('[data-sel="freeze"]', bar).onclick = () => toggleFreeze(items());
    // Sağ tık menüsü (contextmenu.js) bu çizimin işlevlerini buradan kullanır — satır düğmeleriyle aynı yollar
    S.filesCtx = {
      root, P, srv, rel, viaConn, sel, items, full, hashFor, reload, q, remoteRoot: data.root || '/',
      previewFile: (p) => previewFile(P, srv, p), editFile: (p) => editFile(P, srv, p), compareModal: (p) => compareModal(P, srv, p),
      download: (p) => { location.href = `/api/p/${enc(P)}/remote/get?${q}&path=${enc(p)}`; },
      renameItem, toggleFreeze, removeItems, moveItems, zipItems, unzipItem,
      uploadInto: (dir) => {
        const inp = document.createElement('input');
        inp.type = 'file'; inp.multiple = true;
        inp.onchange = () => uploadItems([...inp.files].map((f) => ({ file: f, path: f.name })), dir);
        inp.click();
      },
      click: (id) => { const b = $(id, root); if (b) b.click(); },
      toggleRow: (name) => { const c = $$('#f-body tr[data-entry]', root).find((r) => r.dataset.entry === name)?.querySelector('input[type=checkbox]'); if (c) c.click(); },
      selectAll: () => { if (allBox && !allBox.checked) allBox.click(); },
      clearSel: () => $('[data-sel="clear"]', bar).click(),
    };
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
          else if (act === 'ed') editFile(P, srv, p);
          else if (act === 'cmp') compareModal(P, srv, p);
          else if (act === 'dl') location.href = `/api/p/${enc(P)}/remote/get?${q}&path=${enc(p)}`;
          else if (act === 'ren') renameItem(it);
          else if (act === 'fz') toggleFreeze([it]);
          else if (act === 'rm') removeItems([it]);
          else if (act === 'unzip') unzipItem(it);
          return;
        }
        if (e.target.closest('input,label')) return;
        if (typ === 'dir') go(hashFor(p)); else previewFile(P, srv, p);
      };
      renderMore();
    }
    if (fromCache) {
      const seq = S.renderSeq;
      fetchLs().then((fresh) => {
        if (JSON.stringify(fresh) === JSON.stringify(data)) return;
        S.lsCache.set(lsKey, fresh);
        if (seq === S.renderSeq) render(true);
      }).catch(() => {});
    }
  });

  // Sync'e hiç girmeyen dosya/klasör: ignore dosyası / exclude / include dışı (syncSkipReason)
  const skipBadge = (k) => (k ? `<span class="pill" style="height:18px;font-size:11px" title="${t('skipHint_' + k)}">${icon('x')}${t('skip_' + k)}</span>` : '');
  // frozen.MatchSet ile aynı: kendisi, "klasör/" kaydı ya da üst klasörlerinden biri; klasörde altında kayıt da
  const frozenUnder = (p) => { const pre = p + '/'; for (const f of frozenSet) if (f.startsWith(pre)) return true; return false; };
  const frozenSelf = (p, dir) => {
    if (frozenSet.has(dir ? p + '/' : p)) return true;
    for (let i = p.indexOf('/'); i >= 0; i = p.indexOf('/', i + 1)) if (frozenSet.has(p.slice(0, i + 1))) return true;
    return false;
  };
  // Yükleme önizlemesi: bulunulan klasöre gidenler üst seviye adlarıyla (klasör / dosya) listede soluk
  // satır olarak hemen görünür, bir klasöre bırakılanlar o klasörün satırında sayaç olarak. Sunucu
  // parçaları sırayla okuyup yükler (TCP geri basıncı) → gönderilen bayt ≈ yüklenen dosyalar; satır
  // dosyalarının hepsi gidince tam görünür. Bitince liste zaten yeniden yüklenir.
  function uploadGhosts(items, dir) {
    const body = () => $('#f-body', document);
    const groups = new Map();
    let off = 0;
    const into = dir !== rel ? dir.split('/').pop() : '';
    for (const it of items) {
      const seg = it.path.split('/');
      const key = into || seg[0];
      if (!groups.has(key)) groups.set(key, { name: key, dir: !!into || seg.length > 1, size: 0, ends: [], done: 0, row: null });
      const g = groups.get(key);
      off += it.file.size + it.path.length * 2 + it.file.name.length + 320; // parça başlıkları için pay
      g.ends.push(off);
      g.size += it.file.size;
    }
    if (!$('#f-body')) {
      const bos = $('#sayfa .bos');
      if (bos) bos.outerHTML = `<div class="tablo-kap"><table class="tablo"><thead><tr><th class="sec-hucre"></th><th>${t('colName')}</th><th class="sag">${t('colSize')}</th><th>${t('colDate')}</th><th></th></tr></thead><tbody id="f-body"></tbody></table></div>`;
    }
    const tb = body();
    const status = (g) => (g.dir ? t('upCount', g.done, g.ends.length) : g.done ? t('upDone') : t('upWaiting'));
    // bırakma sırasıyla en üstte: sondan başa afterbegin
    for (const g of [...groups.values()].reverse()) {
      let row = tb &&[...tb.querySelectorAll('tr[data-entry]')].find((r) => r.dataset.entry === g.name);
      if (!row && tb && !into) {
        tb.insertAdjacentHTML('afterbegin', `<tr data-ghost class="yukleme-satir"><td class="sec-hucre"></td>
          <td><div class="ad-hucre">${icon(g.dir ? 'folder' : 'file', g.dir ? 'acc-t' : 'faint')}<span>${esc(g.name)}</span></div></td>
          <td class="sag mono faint">${bytes(g.size)}</td><td class="mono faint" data-upst></td><td></td></tr>`);
        row = tb.firstElementChild;
      } else if (row) {
        row.classList.add('yukleme-satir');
        const c = row.children[3];
        if (c) c.setAttribute('data-upst', '');
      }
      g.row = row;
      const st = row && row.querySelector('[data-upst]');
      if (st) st.textContent = status(g);
    }
    const progress = (loaded) => {
      for (const g of groups.values()) {
        let n = 0;
        while (n < g.ends.length && g.ends[n] <= loaded) n++;
        if (loaded === Infinity) n = g.ends.length;
        if (n === g.done) continue;
        g.done = n;
        if (!g.row || !g.row.isConnected) continue;
        const st = g.row.querySelector('[data-upst]');
        if (st) st.textContent = status(g);
        if (n === g.ends.length) g.row.classList.add('yuklendi');
      }
    };
    const failed = () => { for (const g of groups.values()) if (g.row && g.done < g.ends.length) g.row.classList.add('yukleme-hata'); };
    return { progress, failed };
  }
  function rowHTML(e, checked) {
    const p = full(e.name);
    const fz = frozenSelf(p, e.type === 'dir') || (e.type === 'dir' && frozenUnder(p));
    return `<tr data-entry="${esc(e.name)}" data-type="${e.type}" class="tikla"><td class="sec-hucre"><label><input type="checkbox" ${checked ? 'checked' : ''}></label></td>
      <td><div class="ad-hucre">${icon(e.type === 'dir' ? 'folder' : 'file', e.type === 'dir' ? 'acc-t' : 'faint')}<span>${esc(e.name)}</span>${fz ? `<span class="pill" style="height:18px;font-size:11px">${icon('snow')}${t('frozenBadge')}</span>` : ''}${skipBadge(e.skip)}</div></td>
      <td class="sag mono faint">${e.type === 'dir' ? '' : bytes(e.size)}</td><td class="mono faint">${e.time && !e.time.startsWith('0001') ? fmtDate(e.time) : ''}</td>
      <td class="sag satir-aks" style="white-space:nowrap">${e.type === 'dir' ? '' : `<button class="ikon-btn" data-act="pv" title="${t('preview')}">${icon('eye')}</button><button class="ikon-btn" data-act="ed" title="${t('edEdit')}">${icon('code')}</button>${viaConn ? '' : `<button class="ikon-btn" data-act="cmp" title="${t('compare')}">${icon('layers')}</button>`}<button class="ikon-btn" data-act="dl" title="${t('download')}">${icon('download')}</button>${/\.zip$/i.test(e.name) ? `<button class="ikon-btn" data-act="unzip" title="${t('unzip')}">${icon('archive')}</button>` : ''}`}
        ${viaConn ? '' : `<button class="ikon-btn ${fz ? 'acc-t' : ''}" data-act="fz" title="${t('freezeToggle')}">${icon('snow')}</button>`}
        <button class="ikon-btn" data-act="ren" title="${t('rename')}">${icon('edit')}</button>
        <button class="ikon-btn tehlike" data-act="rm" title="${t('delete')}">${icon('trash')}</button></td></tr>`;
  }
  // Hedef seçimi iki ayrı grup: profiller ve bağlantılar; kalabalık grup açılır listeye döner
  const targetGroup = (label, items) => (items.length ? `<div class="f-hedef-grup"><span class="f-hedef-et">${esc(label)}</span>${items.length <= 6
    ? `<span class="seg srv-seg">${items.map((x) => `<a href="${pHash(P, 'files', x.value)}" class="${x.value === srv ? 'aktif' : ''}" title="${esc(x.title)}">${icon(x.icon)}${esc(x.name)}</a>`).join('')}</span>`
    : `<select data-fsrv>${items.some((x) => x.value === srv) ? '' : '<option value="" selected>—</option>'}${items.map((x) => `<option value="${esc(x.value)}" ${x.value === srv ? 'selected' : ''}>${esc(x.name)}</option>`).join('')}</select>`}</div>` : '');
  const profileItems = cfg.servers.map((s) => ({ value: s.name, name: s.name, icon: 'server', title: s.remote_path }));
  const connItems = conns.map((c) => ({ value: CONN_TARGET + c.name, name: c.name, icon: 'link', title: t('filesConnTip', `${c.user ? c.user + '@' : ''}${c.host}`) }));
  return `<div class="f-hedefler">${targetGroup(t('tabServers'), profileItems)}${targetGroup(t('tabConnections'), connItems)}</div>
    <div class="arac-cubuk">
      <div class="ara">${icon('search')}<input type="text" id="f-filter" placeholder="${t('filter')}"></div>
      <button class="btn" id="f-search" title="${t('searchAll')}">${icon('search')}${t('searchAll')}</button>
      <button class="btn" id="f-tree" title="${t('treeHint')}">${icon('layers')}${t('treeView')}</button>
      <button class="btn" id="f-mkdir">${icon('folder')}${t('newFolder')}</button>
      <button class="btn" id="f-mkfile">${icon('file')}${t('pkNewFile')}</button>
      <button class="btn mavi" id="f-upload" title="${t('dropHint')}">${icon('upload')}${t('upload')}</button><input type="file" id="f-upload-input" multiple hidden></div>
    ${viaConn ? `<p class="ipucu" style="margin:0 0 8px">${icon('link')} ${t('filesConnHint')}</p>` : ''}
    <div class="yol-cubuk"><button data-crumb="">${esc(data.root || '/')}</button>${crumbs.map((c, i) => `<span>/</span><button data-crumb="${esc(crumbs.slice(0, i + 1).join('/'))}">${esc(c)}</button>`).join('')}</div>
    <div class="secim-cubuk" id="f-selbar" hidden><b data-selcount></b>
      <button class="btn kucuk" data-sel="move">${icon('folder')}${t('moveTo')}</button>
      <button class="btn kucuk" data-sel="zip">${icon('archive')}${t('zipSel')}</button>
      ${viaConn ? '' : `<button class="btn kucuk" data-sel="freeze">${icon('snow')}${t('freezeToggle')}</button>`}
      <button class="btn kucuk tehlike" data-sel="del">${icon('trash')}${t('delete')}</button>
      <button class="btn kucuk hayalet" data-sel="clear">${t('clearSel')}</button></div>
    ${data.error ? `<div class="kart"><div class="not bad">${icon('alert')}${esc(data.error)}</div></div>` : !entries.length ? `<div class="bos" style="padding:30px">${t('emptyDir')}</div>` :
    `<div class="tablo-kap"><table class="tablo"><thead><tr><th class="sec-hucre"><label><input type="checkbox" id="f-all"></label></th><th>${t('colName')}</th><th class="sag">${t('colSize')}</th><th>${t('colDate')}</th><th></th></tr></thead><tbody id="f-body">
      </tbody></table></div><button class="btn hayalet tam-gen" id="f-more" hidden></button>`}`;
}

// Dosyalar sekmesine sürükle-bırak yükleme: dosyalar ve klasörler (alt yapısıyla) bulunulan klasöre,
// bir klasör satırının üstüne bırakılırsa o klasöre. onDrop(items [{file, path}], dirName|'').
// Bırakma alanı içerik kadar değil, çerçevenin görünen alanının tamamı (#frame — klasör kısa olsa da
// sayfanın altına kadar). Çerçeve sayfalar arasında aynı öğe → olaylar özellik olarak atanır (birikmez),
// dosyalar sayfası değilse hiçbir şey yapmaz.
function bindDropUpload(root, onDrop) {
  const zone = $('#frame') || root;
  const active = () => !!$('#f-upload', root);
  const hasFiles = (e) => [...(e.dataTransfer?.types || [])].includes('Files');
  let depth = 0;
  const clear = () => { depth = 0; zone.classList.remove('surukle'); $$('tr.surukle-hedef', root).forEach((x) => x.classList.remove('surukle-hedef')); };
  const readEntry = async (entry, prefix, out) => {
    if (entry.isFile) {
      const file = await new Promise((res, rej) => entry.file(res, rej));
      out.push({ file, path: prefix + file.name });
    } else if (entry.isDirectory) {
      const reader = entry.createReader();
      // readEntries her çağrıda en fazla ~100 girdi döner — boş gelene kadar oku
      for (;;) {
        const batch = await new Promise((res, rej) => reader.readEntries(res, rej));
        if (!batch.length) break;
        for (const ch of batch) await readEntry(ch, prefix + entry.name + '/', out);
      }
    }
  };
  zone.ondragenter = (e) => {
    if (!active() || !hasFiles(e)) return;
    e.preventDefault();
    depth++;
    // Katman position:fixed → çerçeve kaydırılmış olsa da görünen alanı kaplar
    const r = zone.getBoundingClientRect();
    zone.style.setProperty('--dz-t', r.top + 'px'); zone.style.setProperty('--dz-l', r.left + 'px');
    zone.style.setProperty('--dz-w', r.width + 'px'); zone.style.setProperty('--dz-h', r.height + 'px');
    zone.dataset.drop = t('dropHere');
    zone.classList.add('surukle');
  };
  zone.ondragover = (e) => {
    if (!active() || !hasFiles(e)) return;
    e.preventDefault();
    e.dataTransfer.dropEffect = 'copy';
    const row = e.target.closest && e.target.closest('tr[data-type="dir"]');
    $$('tr.surukle-hedef', root).forEach((x) => { if (x !== row) x.classList.remove('surukle-hedef'); });
    if (row) row.classList.add('surukle-hedef');
    zone.dataset.drop = row ? t('dropInto', row.dataset.entry) : t('dropHere');
  };
  zone.ondragleave = (e) => {
    if (!active() || !hasFiles(e)) return;
    if (--depth <= 0) clear();
  };
  zone.ondrop = async (e) => {
    if (!active() || !hasFiles(e)) { clear(); return; }
    e.preventDefault();
    const row = e.target.closest && e.target.closest('tr[data-type="dir"]');
    const dirName = row ? row.dataset.entry : '';
    clear();
    // Girdiler olay sırasında alınmalı (await sonrası DataTransfer boşalır)
    const entries = [...e.dataTransfer.items].map((it) => it.webkitGetAsEntry && it.webkitGetAsEntry()).filter(Boolean);
    const plain = [...e.dataTransfer.files];
    let items = [];
    try {
      if (entries.length) for (const en of entries) await readEntry(en, '', items);
      else items = plain.map((f) => ({ file: f, path: f.name }));
    } catch (err) { fail(err); return; }
    onDrop(items, dirName);
  };
}

// Medya önizlemesi (resim / video / ses / PDF) — sunucu satır içi, sandbox CSP ile sunar (cmd_gui_media.go)
const MEDIA_EXT = {
  image: /\.(png|jpe?g|gif|webp|svg|bmp|ico|avif)$/i,
  video: /\.(mp4|webm|ogv|m4v|mov)$/i,
  audio: /\.(mp3|wav|ogg|oga|m4a|flac|aac|opus)$/i,
  pdf: /\.pdf$/i,
};
const mediaKind = (name) => Object.keys(MEDIA_EXT).find((k) => MEDIA_EXT[k].test(name || '')) || '';
function mediaView(kind, url) {
  const u = esc(url);
  switch (kind) {
    case 'image': return `<div class="medya medya-resim"><img src="${u}" alt="" onerror="this.parentElement.classList.add('hata')"><span class="medya-hata">${t('mediaFailed')}</span></div>`;
    case 'video': return `<div class="medya"><video src="${u}" controls preload="metadata"></video></div>`;
    case 'audio': return `<div class="medya medya-ses"><audio src="${u}" controls preload="metadata"></audio></div>`;
    case 'pdf': return `<div class="medya"><iframe class="medya-pdf" src="${u}"></iframe></div>`;
  }
  return `<div class="bos">${t('binaryFile')}</div>`;
}
// Pencere kapanınca oynayan video/ses dursun (öğe DOM'dan kalkınca tarayıcı bazen çalmaya devam ediyor)
function stopMedia(root) { $$('video,audio', root).forEach((m) => { try { m.pause(); m.removeAttribute('src'); m.load(); } catch { /* */ } }); }

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

// ── Biçimlendirme (unminify) — yalnızca görünüm; js-beautify gömülü, ilk kullanımda yüklenir ──
function beautifyKind(name) {
  const e = (/\.([a-z0-9]+)$/i.exec(name || '') || [, ''])[1].toLowerCase();
  if (['js', 'mjs', 'cjs', 'jsx', 'ts', 'tsx'].includes(e)) return 'js';
  if (['json', 'map', 'webmanifest'].includes(e)) return 'json';
  if (['css', 'scss', 'less'].includes(e)) return 'css';
  if (['html', 'htm', 'xhtml', 'xml', 'svg'].includes(e)) return 'html';
  return '';
}
// Uzun satırlar ya da satır başına çok karakter → sıkıştırılmış sayılır (otomatik biçimlendirilir)
function looksMinified(text) {
  if (!text || text.length < 1500) return false;
  const lines = text.split('\n');
  return text.length / lines.length > 250 || lines.some((l) => l.length > 2000);
}
let beautifierP = null;
function loadBeautifier() {
  return (beautifierP ||= new Promise((ok, no) => {
    const s = document.createElement('script');
    s.src = '/assets/vendor/beautifier.min.js';
    s.onload = () => (window.beautifier ? ok(window.beautifier) : no(new Error('beautifier')));
    s.onerror = () => { beautifierP = null; no(new Error('beautifier')); };
    document.head.appendChild(s);
  }));
}
async function beautifyText(text, name) {
  const kind = beautifyKind(name);
  if (kind === 'json') { try { return JSON.stringify(JSON.parse(text), null, 2); } catch { /* geçersiz JSON → JS biçimleyici */ } }
  const b = await loadBeautifier();
  const o = { indent_size: 2, max_preserve_newlines: 2, wrap_line_length: 0 };
  if (kind === 'css') return b.css(text, o);
  if (kind === 'html') return b.html(text, { ...o, indent_inner_html: true });
  return b.js(text, o);
}
// fmtCodeView: biçimlendirilebilen dosyada üstte "Orijinal | Biçimlendir" seçimi; sıkıştırılmışsa
// biçimlendirilmiş açılır. bindFmt(root) modal çizildikten sonra çağrılır.
const FMT = new Map();
let fmtSeq = 0;
function fmtCodeView(text, name) {
  if (!beautifyKind(name)) return codeView(text, name);
  const id = 'fmt' + ++fmtSeq;
  FMT.set(id, { text, name, pretty: null });
  const min = looksMinified(text);
  return `<div data-fmt="${id}" data-auto="${min ? 1 : ''}">
    <div class="fmt-bar"><span class="seg"><button data-fmtv="raw" class="aktif">${t('fmtRaw')}</button><button data-fmtv="pretty">${icon('code')}${t('fmtPretty')}</button></span>
      <span class="ipucu" data-fmtnote>${min ? t('fmtMinified') : t('fmtHint')}</span></div>
    <div data-fmtbody>${codeView(text, name)}</div></div>`;
}
function bindFmt(root) {
  $$('[data-fmt]', root).forEach((box) => {
    const st = FMT.get(box.dataset.fmt);
    if (!st) return;
    FMT.delete(box.dataset.fmt); // modal kapanınca metin bellekte kalmasın (durum kutunun kapanışında)
    const show = async (v) => {
      $$('[data-fmtv]', box).forEach((b) => b.classList.toggle('aktif', b.dataset.fmtv === v));
      const body = $('[data-fmtbody]', box);
      if (v === 'raw') { body.innerHTML = codeView(st.text, st.name); return; }
      if (st.pretty == null) {
        body.innerHTML = `<div class="yukleniyor" style="padding:20px 0">${icon('refresh', 'spin')} ${t('loading')}</div>`;
        try { st.pretty = await beautifyText(st.text, st.name); } catch (e) { toast(t('fmtFailed', e.message || e), 'bad'); return show('raw'); }
      }
      if (box.isConnected) body.innerHTML = codeView(st.pretty, st.name);
    };
    $$('[data-fmtv]', box).forEach((b) => (b.onclick = () => show(b.dataset.fmtv)));
    if (box.dataset.auto) show('pretty');
  });
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
  const kind = mediaKind(rel);
  if (kind && (loc.exists || rem.exists)) {
    // Medya: yan yana yerel ↔ sunucu (boyut farkı başlıkta)
    const v = Date.now();
    const side = (label, d, url, missing) => `<div class="medya-yan"><div class="medya-bas"><b>${label}</b><span class="faint mono">${d.exists ? bytes(d.size) : ''}</span></div>
      ${d.exists ? mediaView(kind, url) : `<div class="bos">${t(missing)}</div>`}</div>`;
    const same = loc.exists && rem.exists && loc.size === rem.size;
    views.media = `${same ? `<p class="ipucu" style="margin:0 0 10px">${icon('check')} ${t('mediaSameSize')}</p>` : ''}<div class="medya-ikili">
      ${side(t('diffLocal'), loc, `/api/p/${enc(P)}/local/media?${q}&v=${v}`, 'mediaNoLocal')}
      ${side(t('diffRemote'), rem, `/api/p/${enc(P)}/remote/media?${q}&v=${v}`, 'mediaNoRemote')}</div>`;
  }
  if (binary) {
    if (!views.media) views.diff = `<div class="bos">${t('binaryFile')}</div>`;
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
  if (loc.exists && !loc.binary) views.local = fmtCodeView(loc.content, rel);
  if (rem.exists && !rem.binary) views.remote = fmtCodeView(rem.content, rel);
  if (isHTMLFile(rel)) {
    if (loc.exists && !loc.binary) views.localHtml = htmlFrame(loc.content);
    if (rem.exists && !rem.binary) views.remoteHtml = htmlFrame(rem.content);
  }
  const tabs = [['media', t('viewMedia')], ['diff', t('diffTitle')], ['local', t('diffLocal')], ['remote', t('diffRemote')],
    ['localHtml', t('viewRenderedLocal')], ['remoteHtml', t('viewRenderedRemote')]].filter(([k]) => views[k]);
  const m = modal({
    title: `${icon('file')}<span class="mono" style="font-size:13px">${esc(rel)}</span> ${summary}`, size: 'genis',
    body: `<div class="tabs" style="margin-bottom:12px">${tabs.map(([k, l], i) => `<a href="javascript:void 0" data-cv="${k}" class="${i ? '' : 'aktif'}">${l}</a>`).join('')}</div>
      ${tabs.map(([k], i) => `<div data-cvp="${k}" ${i ? 'hidden' : ''}>${views[k]}</div>`).join('')}`,
    actions: [{ left: `<span class="mono">${esc(srv)}</span>` }, { label: t('close'), cls: 'birincil' }],
    onClose: () => stopMedia(m.root),
  });
  $$('[data-cv]', m.root).forEach((a) => (a.onclick = () => {
    $$('[data-cv]', m.root).forEach((x) => x.classList.toggle('aktif', x === a));
    $$('[data-cvp]', m.root).forEach((p) => (p.hidden = p.dataset.cvp !== a.dataset.cv));
  }));
  bindFmt(m.root);
}

async function previewFile(P, srv, p) {
  const lm = await loadModal(`${icon('file')}<span class="mono" style="font-size:13px">${esc(p)}</span>`, 'genis', () => GET(`/api/p/${enc(P)}/remote/cat?server=${enc(srv)}&path=${enc(p)}`));
  if (!lm) return;
  const d = lm.d;
  const kind = mediaKind(p);
  const media = kind ? mediaView(kind, `/api/p/${enc(P)}/remote/media?server=${enc(srv)}&path=${enc(p)}&v=${Date.now()}`) : '';
  let body;
  if (d.binary) body = media || `<div class="bos">${t('binaryFile')}</div>`;
  else body = viewTabs([...(media ? [['media', t('viewMedia'), media]] : []), ['code', t('viewCode'), fmtCodeView(d.content, p)],
    ...(isHTMLFile(p) && kind !== 'image' ? [['html', t('viewRendered'), htmlFrame(d.content)]] : [])]);
  const pm = modal({
    title: `${icon('file')}<span class="mono" style="font-size:13px">${esc(p)}</span>`, size: 'genis',
    body, onClose: () => stopMedia(pm.root),
    actions: [{ left: `${bytes(d.size)} ${d.truncated && !d.binary ? t('truncated') : ''}` },
      ...((d.binary && !kind) || isConnTarget(srv) ? [] : [{ label: icon('layers') + t('compare'), cls: '', key: 'r', run: () => { compareModal(P, srv, p); return false; } }]),
      ...(d.binary ? [] : [
        // Önizleme altta açık kalır (false) → Esc üstteki pencereyi kapatıp buraya döner; kaydedildiyse tazelenir
        { label: icon('code') + t('edEdit'), cls: '', key: 's', run: () => { editFile(P, srv, p, { onClose: (saved) => { if (saved && pm.root.isConnected) { pm.close(); previewFile(P, srv, p); } } }); return false; } }]),
      { label: icon('download') + t('download'), cls: '', key: 'd', run: () => { location.href = `/api/p/${enc(P)}/remote/get?server=${enc(srv)}&path=${enc(p)}`; } },
      { label: t('close'), cls: 'birincil' }],
  });
  bindViewTabs(pm.root);
  bindFmt(pm.root);
}

// ── Geçmiş (release + rollback + log) ────────────────────────────────────────
async function tabHistory(proj, srvFilter) {
  const P = proj.name;
  const q = srvFilter ? `?server=${enc(srvFilter)}` : '';
  const [rels, logs, cfg] = await Promise.all([GET(`/api/p/${enc(P)}/releases${q}`), GET(`/api/p/${enc(P)}/logs${q}`), getConfig(P)]);
  bind((root) => {
    $('#proj-aks', root).innerHTML = '';
    const hs = $('#h-srv', root); // tek sunucuda filtre çizilmez
    if (hs) hs.onchange = (e) => go(e.target.value ? pHash(P, 'history', e.target.value) : pHash(P, 'history'));
    $$('[data-rel]', root).forEach((b) => (b.onclick = () => releaseDetail(P, b.dataset.srv, b.dataset.rel)));
    $$('[data-more]', root).forEach((b) => (b.onclick = () => { $$(`[data-extra="${b.dataset.more}"]`, root).forEach((tr) => (tr.hidden = false)); b.parentElement.remove(); }));
    $$('[data-undo]', root).forEach((b) => (b.onclick = () => rollbackFlow(P, b.dataset.srv, b.dataset.undo)));
    $$('[data-reldel]', root).forEach((b) => (b.onclick = () => deleteReleasesFlow(P, b.dataset.srv, [b.dataset.reldel])));
    const da = $('#h-delall', root);
    if (da) da.onclick = async () => {
      // Filtre yoksa release'i olan her sunucu için ayrı onay (CLI: rollback --delete --all -s <sunucu>)
      const servers = srvFilter ? [srvFilter] : [...new Set(rels.map((r) => r.server))];
      for (const s of servers) if (!(await deleteReleasesFlow(P, s, null))) break;
    };
    $$('[data-log]', root).forEach((tr) => (tr.onclick = async () => {
      const lm = await loadModal(`${icon('log')}<span class="mono" style="font-size:13px">${esc(tr.dataset.srv)} / ${esc(tr.dataset.log)}</span>`, 'genis', () => GET(`/api/p/${enc(P)}/logs/${enc(tr.dataset.srv)}/${enc(tr.dataset.log)}`));
      if (!lm || lm.d == null) return;
      const txt = lm.d;
      modal({ title: `${icon('log')}<span class="mono" style="font-size:13px">${esc(tr.dataset.srv)} / ${esc(tr.dataset.log)}</span>`, size: 'genis',
        body: `<div class="log">${txt.split('\n').map(logLine).join('')}</div>`, actions: [{ label: t('close'), cls: 'birincil' }] });
    }));
  });
  const LIM = 25;
  const more = (n, id) => (n > LIM ? `<div class="daha-fazla"><button class="btn kucuk" data-more="${id}">${icon('down2')}${t('showMore', n - LIM)}</button></div>` : '');
  const relRows = rels.map((r, i) => {
    const status = r.rolled_back ? `<span class="pill">${t('rolledBack')}</span>` : r.partly_rolled_back ? `<span class="pill warn">${t('relPartly', r.partly_rolled_back)}</span>` : r.has_backup ? `<span class="pill ok"><span class="nokta"></span>${t('restorable')}</span>` : `<span class="pill">${t('noBackupTag')}</span>`;
    return `<tr ${i >= LIM ? 'hidden data-extra="rel"' : ''}><td class="mono">${fmtDate(r.created_at)}</td><td>${esc(r.server)}</td><td class="sag">${r.files}</td><td class="sag">${r.backed}</td><td class="sag">${r.new}</td><td>${status}</td>
      <td class="sag" style="white-space:nowrap"><button class="btn kucuk hayalet" data-rel="${esc(r.id)}" data-srv="${esc(r.server)}">${t('details')}</button>
      ${r.has_backup && !r.rolled_back ? `<button class="btn kucuk tehlike" data-undo="${esc(r.id)}" data-srv="${esc(r.server)}">${icon('undo')}${t('undo')}</button>` : ''}
      <button class="ikon-btn tehlike" data-reldel="${esc(r.id)}" data-srv="${esc(r.server)}" title="${t('relDelete')}">${icon('trash')}</button></td></tr>`;
  }).join('');
  return `${cfg.servers.length > 1 ? `<div class="arac-cubuk"><select id="h-srv"><option value="">${t('all')}</option>${cfg.servers.map((s) => `<option value="${esc(s.name)}" ${s.name === srvFilter ? 'selected' : ''}>${esc(s.name)}</option>`).join('')}</select></div>` : ''}
    <div style="display:flex;align-items:center;gap:10px;flex-wrap:wrap"><h2 style="margin-top:6px">${icon('undo')}${t('releases')}</h2>${rels.length ? `<button class="btn kucuk tehlike" id="h-delall" style="margin-left:auto">${icon('trash')}${t('relDeleteAll')}</button>` : ''}</div>
    <p class="muted" style="margin-top:-6px">${t('releasesSub')}</p>
    ${rels.length ? `<div class="tablo-kap"><table class="tablo"><thead><tr><th>${t('colWhen')}</th><th>${t('colServer')}</th><th class="sag">${t('colFiles')}</th><th class="sag">${t('colReplaced')}</th><th class="sag">${t('colNew')}</th><th>${t('colStatus')}</th><th></th></tr></thead><tbody>${relRows}</tbody></table></div>${more(rels.length, 'rel')}`
      : `<div class="bos" style="padding:30px">${t('noReleases')}</div>`}
    <h2>${icon('log')}${t('logs')}</h2>
    ${logs.length ? `<div class="tablo-kap"><table class="tablo"><thead><tr><th>${t('colWhen')}</th><th>${t('colServer')}</th><th>${t('colName')}</th><th class="sag">${t('colSize')}</th></tr></thead><tbody>
      ${logs.map((l, i) => `<tr class="tikla" ${i >= LIM ? 'hidden data-extra="log"' : ''} data-log="${esc(l.file)}" data-srv="${esc(l.server)}"><td class="mono">${fmtDate(l.time)}</td><td>${esc(l.server)}</td><td class="mono faint">${esc(l.file)}</td><td class="sag mono faint">${bytes(l.size)}</td></tr>`).join('')}</tbody></table></div>${more(logs.length, 'log')}`
      : `<div class="bos" style="padding:30px">${t('noLogs')}</div>`}`;
}

// Release penceresi: solda değişen dosyalar klasör ağacında (onay kutulu), sağda seçili dosyanın
// diff'i; seçilenleri ya da release'in tamamını geri al (CLI: rollback <id> [--files a,b] [--diff]).
async function releaseDetail(P, srv, id) {
  const title = `${icon('history')}${t('releaseTitle', esc(id))} · ${esc(srv)}`;
  const lm = await loadModal(title, 'genis', () => GET(`/api/p/${enc(P)}/releases/${enc(srv)}/${enc(id)}`));
  if (!lm) return;
  const d = lm.d, m = d.manifest;
  const items = d.items || [];
  const can = (it) => d.restorable && !it.rolled_back && it.status !== 'nobackup'; // seçilebilir mi
  const nB = (m.backed || []).length, nN = (m.new || []).length;
  const partly = (m.rolled_back_files || []).length;
  const state = m.rolled_back ? `<span class="pill">${t('rolledBack')}</span>`
    : partly ? `<span class="pill warn">${t('relPartly', partly)}</span>` : '';

  const rm = modal({
    title: `${title} ${state}`, size: 'genis',
    body: `<p class="mono faint" style="margin:0 0 10px">${fmtDate(m.created_at)} · ${t('colFiles')}: ${items.length} · ${t('colReplaced')}: ${nB} · ${t('colNew')}: ${nN}</p>
      <div class="rt-grid"><div class="rt-agac" data-tree>${relTreeHTML(items, can)}</div>
        <div class="rt-sag" data-pane><div class="bos" style="padding:40px">${t('relPickFile')}</div></div></div>
      <div class="rt-alt"><span class="faint" data-selcount></span><span style="margin-left:auto"></span>
        ${d.restorable ? `<button class="btn tehlike" data-rbsel disabled>${icon('undo')}<span></span></button>
        <button class="btn hayalet" data-rball>${icon('undo')}${t('relRollbackAll')}</button>` : ''}</div>`,
    actions: [{ label: icon('trash') + t('relDelete'), cls: 'tehlike', run: () => { deleteReleasesFlow(P, srv, [id], () => rm.close()); return false; } },
      { label: t('close'), cls: 'birincil' }],
    onClose: () => stopMedia(rm.root),
  });
  const tree = $('[data-tree]', rm.root), pane = $('[data-pane]', rm.root);
  const selected = () => $$('input[data-rsel]:checked', tree).map((c) => c.dataset.rsel);
  const sync = () => {
    // Klasör kutuları: altındaki seçilebilir dosyaların hepsi seçiliyse işaretli, bir kısmıysa belirsiz
    $$('input[data-rdir]', tree).forEach((c) => {
      const kids = $$('input[data-rsel]:not(:disabled)', c.closest('details'));
      const on = kids.filter((k) => k.checked).length;
      c.checked = kids.length > 0 && on === kids.length;
      c.indeterminate = on > 0 && on < kids.length;
      c.disabled = kids.length === 0;
    });
    const n = selected().length;
    $('[data-selcount]', rm.root).textContent = d.restorable ? t('relSelected', n) : t('relNotRestorable');
    const b = $('[data-rbsel]', rm.root);
    if (b) { b.disabled = n === 0; $('span', b).textContent = t('relRollbackSel', n); }
  };
  tree.addEventListener('change', (e) => {
    const c = e.target;
    if (c.dataset.rdir !== undefined) $$('input[data-rsel]:not(:disabled)', c.closest('details')).forEach((k) => (k.checked = c.checked));
    sync();
  });
  // Dosya satırına tıklayınca (kutu dışında) sağda diff
  const show = async (row) => {
    $$('.rt-dosya.aktif', tree).forEach((x) => x.classList.remove('aktif'));
    row.classList.add('aktif');
    const rel = row.dataset.rf;
    pane.innerHTML = `<div class="bos" style="padding:40px">${icon('refresh', 'spin')}</div>`;
    try {
      const v = await GET(`/api/p/${enc(P)}/releases/${enc(srv)}/${enc(id)}/file?path=${enc(rel)}`);
      if (!pane.isConnected || !row.classList.contains('aktif')) return;
      const f = releaseFileHTML(v, rel, `/api/p/${enc(P)}/releases/${enc(srv)}/${enc(id)}/media?path=${enc(rel)}`);
      pane.innerHTML = `<div class="rt-sag-bas"><span class="mono">${esc(rel)}</span>${f.summary}</div>${f.html}`;
      bindViewTabs(pane);
    } catch (e) { pane.innerHTML = `<p class="bad-t">${esc(e.message)}</p>`; }
  };
  tree.addEventListener('click', (e) => {
    const row = e.target.closest('.rt-dosya');
    if (row && !e.target.closest('input')) show(row);
  });
  const rb = $('[data-rbsel]', rm.root);
  if (rb) rb.onclick = () => rollbackFlow(P, srv, id, selected(), () => rm.close());
  const ra = $('[data-rball]', rm.root);
  if (ra) ra.onclick = () => rollbackFlow(P, srv, id, [], () => rm.close());
  sync();
  const first = $('.rt-dosya', tree);
  if (first) show(first);
}

// deleteReleasesFlow release'leri (ids=null → sunucunun hepsini) ya da yalnızca yedeklerini siler —
// CLI `rollback --delete [--all] [--keep-history]` ile aynı çekirdek. Onaylanıp silindiyse true.
function deleteReleasesFlow(P, srv, ids, onDone) {
  return new Promise((res) => {
    let ok = false;
    const m = modal({
      title: `${icon('trash')}${ids ? t('relDeleteTitle', ids.length) : t('relDeleteTitleAll', esc(srv))}`, size: 'orta',
      body: `<p>${t('relDeleteBody')}</p>${ids && ids.length === 1 ? `<p class="mono faint">${esc(srv)} · ${esc(ids[0])}</p>` : ''}
        <label class="anahtar"><span class="metin"><b>${t('relDeleteKeep')}</b></span><span class="switch"><input type="checkbox" data-keep><i></i></span></label>`,
      onClose: () => res(ok),
      actions: [{ label: t('cancel'), cls: 'hayalet' },
        { label: icon('trash') + t('delete'), cls: 'tehlike', run: async () => {
          const keep = $('[data-keep]', m.root).checked;
          try {
            const r = await api('POST', `/api/p/${enc(P)}/releases/${enc(srv)}/delete`, { ids: ids || [], all: !ids, keep_history: keep });
            ok = true;
            toast(t(keep ? 'relPurgedN' : 'relDeletedN', r.n));
            onDone && onDone();
            render(true);
          } catch (e) { fail(e); }
        } }],
    });
  });
}

// relTreeHTML dosya yollarından klasör ağacı; tek alt klasörlü zincirler birleştirilir (a/b/c).
function relTreeHTML(items, can) {
  const root = { dirs: {}, files: [] };
  for (const it of items) {
    const parts = it.path.split('/');
    let n = root;
    for (const p of parts.slice(0, -1)) n = n.dirs[p] || (n.dirs[p] = { dirs: {}, files: [] });
    n.files.push({ ...it, name: parts[parts.length - 1] });
  }
  const mark = { changed: ['~', 'warn-t'], new: ['+', 'ok-t'], nobackup: ['·', 'faint'] };
  const count = (n) => n.files.length + Object.values(n.dirs).reduce((a, x) => a + count(x), 0);
  const draw = (n) => {
    const dirs = Object.keys(n.dirs).sort().map((name) => {
      let label = name, node = n.dirs[name];
      while (!node.files.length && Object.keys(node.dirs).length === 1) { // a → b → c zinciri
        const only = Object.keys(node.dirs)[0];
        label += '/' + only;
        node = node.dirs[only];
      }
      return `<details open><summary><input type="checkbox" data-rdir>${icon('folder', 'acc-t')}<span>${esc(label)}</span><span class="faint rt-say">${count(node)}</span></summary>
        <div class="rt-ic">${draw(node)}</div></details>`;
    }).join('');
    const files = n.files.sort((a, b) => a.name.localeCompare(b.name)).map((f) => {
      const [sym, cls] = mark[f.status] || mark.nobackup;
      const tip = f.rolled_back ? t('relRolledFile') : f.status === 'nobackup' ? t('relNoBackupFile') : '';
      return `<div class="rt-dosya${f.rolled_back ? ' geri' : ''}" data-rf="${esc(f.path)}" ${tip ? `title="${esc(tip)}"` : ''}>
        <input type="checkbox" data-rsel="${esc(f.path)}" ${can(f) ? '' : 'disabled'}><span class="${cls} mono">${sym}</span><span class="rt-ad">${esc(f.name)}</span></div>`;
    }).join('');
    return dirs + files;
  };
  return items.length ? draw(root) : `<div class="bos">${t('noFiles')}</div>`;
}

// releaseFileHTML release dosyasının önceki ↔ yüklenen hâli (diff / önceki / yüklenen sekmeleri).
function releaseFileHTML(v, rel, mediaURL) {
  let diff, summary = '';
  const kind = mediaURL && mediaKind(rel);
  if (kind) {
    // Medya: önceki ↔ yüklenen hâl yan yana
    const side = (label, ok, sideName, missing) => `<div class="medya-yan"><div class="medya-bas"><b>${label}</b></div>
      ${ok ? mediaView(kind, `${mediaURL}&side=${sideName}`) : `<div class="bos">${t(missing)}</div>`}</div>`;
    const media = `<div class="medya-ikili">${v.status === 'new' ? `<div class="medya-yan"><div class="medya-bas"><b>${t('relBefore')}</b></div><div class="bos ok-t">${t('relNewFile')}</div></div>`
      : side(t('relBefore'), v.before_ok, 'before', 'relNoBefore')}${side(t('relAfter'), v.after_ok, 'after', 'relNoAfter')}</div>`;
    if (v.binary || kind !== 'image') return { html: media, summary };
    diff = media;
  } else if (v.binary) diff = `<div class="bos">${t('binaryFile')}</div>`;
  else if (!v.before_ok) diff = `<p class="warn-t">${t('relNoBefore')}</p>${v.after_ok ? codeView(v.after, rel) : ''}`;
  else if (!v.after_ok) diff = `<p class="warn-t">${t('relNoAfter')}</p>`;
  else if (v.status === 'new') diff = `<p class="ok-t">${t('relNewFile')}</p>${codeView(v.after, rel)}`;
  else if (v.before === v.after) diff = `<div class="bos" style="padding:28px">${icon('check')} ${t('diffSame')}</div>`;
  else {
    const d = diffView(v.before, v.after, rel);
    diff = `<p class="ipucu" style="margin:0 0 10px">${t('relDiffHint')}${v.after_source === 'local' ? ' · ' + t('relAfterLocal') : ''}</p>${d.html}`;
    summary = ` <span class="fark-ozet"><span class="ok-t">+${d.add}</span><span class="bad-t">−${d.del}</span></span>`;
  }
  const views = [['diff', t('diffTitle'), diff]];
  if (!v.binary && v.before_ok && v.status !== 'new') views.push(['before', t('relBefore'), codeView(v.before, rel)]);
  if (!v.binary && v.after_ok) views.push(['after', t('relAfter'), codeView(v.after, rel)]);
  return { html: viewTabs(views), summary };
}

// rollbackFlow onay + iş: plan sunucudan (CLI ile aynı rollbackPlan / rollbackPlanFiles).
// files boş → release ve ondan yeni her release; dolu → yalnızca o dosyalar, release'ten önceki hâline.
async function rollbackFlow(P, srv, id, files = [], onStart) {
  const title = `${icon('undo')}${t('rollbackTitle', esc(srv))}`;
  const lm = await loadModal(title, 'orta', () => GET(`/api/p/${enc(P)}/releases/${enc(srv)}/${enc(id)}/plan${files.length ? '?files=' + enc(files.join('\n')) : ''}`));
  if (!lm) return;
  const plan = (lm.d && lm.d.releases) || [];
  const newFiles = (lm.d && lm.d.new_files) || [];
  const others = plan.filter((r) => r.id !== id);
  const body = files.length
    ? `<p>${t('rbSelBody', files.length, esc(id))}</p><div class="liste-kutu" style="max-height:30vh">${files.map((f) => `<div class="mono">${esc(f)}</div>`).join('')}</div>
      ${others.length ? `<p class="ipucu" style="margin-top:10px">${t('rbSelChain')}</p><div class="liste-kutu">${others.map((r) => `<div><b>${esc(r.id)}</b> <span class="faint">· ${fmtDate(r.created_at)}</span></div>`).join('')}</div>` : ''}`
    : `<p>${t('rollbackBody', plan.length)}</p><div class="liste-kutu">${plan.map((r) => `<div><b>${esc(r.id)}</b> <span class="faint">· ${fmtDate(r.created_at)} · ~${r.backed} +${r.new}</span></div>`).join('')}</div>`;
  // Yeni eklenen dosyalar varsayılan olarak sunucuda BIRAKILIR; silmek bilinçli seçim — liste gösterilir
  const newBox = newFiles.length ? `<div class="rb-yeni">${fBool('delete_new', t('rbDeleteNew', newFiles.length), false, t('rbDeleteNewHint'))}
      <div class="liste-kutu" data-newlist hidden style="max-height:24vh;margin-top:8px">${newFiles.map((f) => `<div class="mono bad-t">− ${esc(f)}</div>`).join('')}</div></div>` : '';
  const cm = modal({
    title, size: 'orta', body: body + newBox,
    actions: [{ label: t('cancel'), cls: 'hayalet' }, { label: icon('undo') + (files.length ? t('relRollbackSel', files.length) : t('rollbackGo')), cls: 'tehlike', run: async (root) => {
      const del = $('[data-k="delete_new"]', root);
      const ids = await startJobs(P, { kind: 'rollback', servers: [srv], release: id, files, delete_new: !!(del && del.checked) }, t('rollback'));
      if (onStart) onStart();
      const j = ids[0] && S.jobs.get(ids[0]);
      if (j) j.doneCbs.push(() => { if (/\/history/.test(location.hash)) render(true); }); // Geçmiş durumları tazelensin
      if (ids[0]) showJobLog(ids[0]);
    } }],
  });
  const del = $('[data-k="delete_new"]', cm.root);
  if (del) del.onchange = () => { $('[data-newlist]', cm.root).hidden = !del.checked; };
}

// ═══ Sayfa: işler ═════════════════════════════════════════════════════════════
async function pageJobs() {
  let [jobs, watches] = await Promise.all([GET('/api/jobs'), GET('/api/watches')]);
  bind((root) => {
    $$('[data-jid]', root).forEach((tr) => (tr.onclick = () => showJobLog(tr.dataset.jid)));
    $$('[data-wstop]', root).forEach((b) => (b.onclick = async (e) => {
      e.stopPropagation();
      await watchStop(b.dataset.wstop, b.dataset.wsrv);
    }));
  });
  const st = (s) => { const [c, l] = jobPill(s); return `<span class="pill ${c}"><span class="nokta"></span>${l}</span>`; };
  return `<div class="sayfa"><div class="baslik-satir"><div><h1>${t('jobsTitle')}</h1><p>${t('jobsSub')}</p></div></div>
    ${watches.length ? `<h2 style="margin-top:0">${icon('eye')}${t('activeWatches')}</h2><div class="kartlar" style="margin-bottom:24px">${watches.map((w) => `<div class="kart"><div><div class="ust"><h3>${esc(w.project)} · ${esc(w.server)}</h3><span class="pill busy"><span class="nokta"></span>watch</span></div>
      <div class="meta"><span>${icon('server')}${esc(w.servers.join(', '))}</span><span>${icon('clock')}${t('watchSince', fmtDate(w.started).slice(11), w.syncs)}</span>${w.last_sync ? `<span>${icon('upload')}${esc(w.last_sync)}</span>` : ''}</div></div>
      <div class="butonlar"><button class="btn kucuk" data-jid="${esc(w.job_id)}">${icon('log')}${t('jobLog')}</button><button class="btn kucuk tehlike" data-wstop="${esc(w.project)}" data-wsrv="${esc(w.server)}">${icon('stop')}${t('stop')}</button></div></div>`).join('')}</div>` : ''}
    ${(jobs = jobs.filter((j) => !(j.kind === 'watch' && j.status === 'running'))).length ? `<div class="tablo-kap"><table class="tablo"><thead><tr><th>${t('colJob')}</th><th>${t('colProject')}</th><th>${t('colServer')}</th><th>${t('colStatus')}</th><th>${t('colStarted')}</th><th class="sag">${t('colDuration')}</th></tr></thead><tbody>
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
  $('#sb-watch-row').onclick = (e) => openWatches(e.currentTarget);
  $('#btn-kapat').onclick = async () => {
    if (!(await confirmBox(t('quit'), esc(t('quitConfirm')), t('quit'), true))) return;
    await api('POST', '/api/quit').catch(() => {});
    document.body.innerHTML = `<div style="display:grid;place-items:center;height:100vh;color:var(--muted)">${esc(t('quitDone'))}</div>`;
    if (window.__frameClose) setTimeout(window.__frameClose, 600);
  };
  await loadProjects();
  await refreshJobsMeta();
  render();

  // Varlık bağlantısı: pencere açık olduğu sürece süreç yaşar (küçültülünce de kopmaz)
  const presence = new EventSource('/api/presence');
  presence.addEventListener('bye', () => { presence.close(); if (window.__frameClose) window.__frameClose(); else window.close(); });
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

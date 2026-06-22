package lang

var Tr = StringSet{
	// Browser TUI
	BrowserLoading:           "  Yükleniyor...",
	BrowserSearchPrompt:      "  Enter=Tümünü Ara  ESC=İptal",
	BrowserSearching:         "  Aranıyor: %q ...",
	BrowserSearchResults:     "  (tüm sunucu, %d sonuç)  ESC temizle",
	BrowserPickDirTitle:      "  HEDEF DİZİN SEÇ",
	BrowserPickDirEnter:      "Enter=Bu Klasörü Seç",
	BrowserHintRight:         "→ Gir",
	BrowserHintLeft:          "← Üst",
	BrowserHintQuit:          "q İptal",
	BrowserMarkedHintDelete:  "d=Sil",
	BrowserMarkedHintMove:    "m=Taşı",
	BrowserMarkedHintFreeze:  "f=❄ Dondur",
	BrowserMarkedHintAll:     "a=Tümünü kaldır",
	BrowserMarkedHintCancel:  "ESC=İptal",
	BrowserEscPending:        "  ESC tekrar = çık",
	BrowserSearchClear:       "  ESC temizle",
	BrowserHintNav:           "  ↑↓ Gezin",
	BrowserHintEnter:         "Enter Gir/Seç",
	BrowserHintPreview:       "→ Önizle",
	BrowserHintSpace:         "Space İşaretle",
	BrowserHintSearch:        "/ Ara",
	BrowserHintQuitKey:       "q Çık",
	BrowserHintFreeze:        "f=❄ Dondur",
	BrowserHintTree:          "t=Tree",
	BrowserErrPrefix:         "  Hata: ",
	BrowserReconnectHint:     "r = Yeniden Bağlan  |  q = Çık",
	BrowserSelectThisPrefix:  "[ Bu dizini seç: ",
	BrowserSelectThisSuffix:  " ]",
	BrowserCountMany:         "çok boyutlu",
	BrowserFooterFmt:         "  %d klasör, %d dosya  ·  %d/%d",
	BrowserSearchMatchFmt:    "  ara:%q %d eşleşme",
	BrowserPreviewLoading:    "  Yükleniyor...",
	BrowserPreviewBack:       "  ESC/q geri",
	BrowserPreviewScrollHint: "↑↓ kaydır",
	BrowserPreviewPageHint:   "PgUp/PgDn",
	BrowserPreviewGGHint:     "g başa  G sona",
	BrowserPreviewEmpty:      "(boş dosya)",
	BrowserPreviewLineFmt:    "%d-%d / %d satır",
	BrowserConnLost:          "⚠ Bağlantı kesildi",
	BrowserConnLostHint:      "r tuşuna basarak yeniden bağlanabilirsiniz",
	BrowserPreviewFailed:     "⚠ Önizleme alınamadı",
	BrowserReconnectFailed:   "yeniden bağlanılamadı: %w",
	BrowserMarkedCount:       "  ✓ %d işaretli",
	BrowserSearchCancelHint:  "   ESC=iptal",
	BrowserSearchTimeout:     "⚠ zaman aşımı (5dk) — %d kısmi sonuç",
	BrowserSearchConnErr:     "⚠ bağlantı hatası: %s",
	BrowserSearchOpenErr:     "bağlantı %d açılamadı: %v",
	BrowserSearchDoneFmt:     "%d sonuç (%.1fs)",
	BrowserSearchOpsHint:     "  ↑↓ gezin  Space=işaretle  Enter=önizle  → önizle  d=sil  m=taşı  f=freeze  ESC=temizle",
	BrowserSearchNoResults:   "  (sonuç bulunamadı)",

	// Picker / Confirm TUI
	PickerFilterHint:   "  Harf yaz → filtrele",
	PickerMatchFmt:     "  (%d eşleşme)",
	PickerNoMatch:      "  (eşleşme yok — Backspace ile sil)",
	PickerNavHint:      "↑↓/jk gezin   Enter seç   [1-9] hızlı   Backspace sil   q iptal",
	MultiPickerNavHint: "↑↓ gezin   Space işaretle   a hepsini   Enter onayla   q iptal",
	ConfirmYes:         "Evet",
	ConfirmNo:          "Hayır, iptal",

	// Sync TUI
	SyncDryRunTitle:      "  [DRY RUN] Yüklenecekler:",
	SyncDryRunFmt:        "  %d dosya — herhangi bir tuşa basın",
	SyncNoChange:         "  ✓ Değişiklik yok — güncel",
	SyncAnyKey:           "  Herhangi bir tuşa basın",
	SyncConnecting:       "bağlanıyor...",
	SyncRetryFmt:         " (%d deneme)",
	SyncRetryOkFmt:       " (%d. denemede)",
	SyncDoneFmt:          "  Tamamlandı: %s yüklendi",
	SyncDoneFailFmt:      ", %s hata",
	SyncAttemptDetailFmt: "deneme %d: %v",
	SyncScrollHint:       "  ↑↓/j/k kaydır · g=üst G=alt · PgYu/PgAş · diğer tuş = çıkış",
	SyncFailedHeader:     "yüklenemeyen dosyalar",

	// Shell REPL
	ShellWelcomeHint:      "  'help' yazın, çıkmak için 'exit'",
	ShellMultiServer:      "Birden fazla sunucu var — bağlanmak için: server <ad>",
	ShellServersLabel:     "Sunucular:",
	ShellCtrlCHint:        "(Ctrl+C — çıkmak için 'exit')",
	ShellCtrlCExitHint:    "Çıkmak için Ctrl+C'ye tekrar basın — ya da 'exit' yazın",
	ShellExit:             "Görüşürüz.",
	ShellUnknownCmd:       "Bilinmeyen komut: %q  (yardım için 'help')\n",
	ShellConnecting:       "Bağlanıyor: %s (%s)...",
	ShellConnectErr:       " hata: %v\n",
	ShellNotConnected:     "Bağlı değil. 'server <ad>' ile bağlanın.",
	ShellNoServers:        "Aktif sunucu yok.",
	ShellServersFmt:       "Sunucular",
	ShellServersSubtitle:  "Seçince bağlanır — q ile sadece kapat",
	ShellAlreadyConn:      "Zaten bağlı: %s\n",
	ShellDisconnected:     "Bağlantı kesildi: %s\n",
	ShellServerNotFound:   "Sunucu bulunamadı: %q\n",
	ShellBrowserErr:       "Browser hatası: %v\n",
	ShellDeleteTitle:      "%d öğeyi sil",
	ShellDeleteSubtitle:   "Bu işlem geri alınamaz. FTP sunucusundan kalıcı olarak silinir.",
	ShellCancelled:        "İptal edildi.",
	ShellConfirmSure:      "Emin misiniz?",
	ShellConfirmSureBody:  "Onaylamak için Evet seçin.",
	ShellDeletedFmt:       "\n%d silindi, %d başarısız\n",
	ShellMovingFmt:        "\n%d dosya taşınacak. Hedef klasörü seçin (Enter = bu klasörü seç):\n",
	ShellMoveTarget:       "Hedef: %s\n\n",
	ShellMovedFmt:         "\n%d taşındı, %d başarısız\n",
	ShellDirNotFound:      "Dizin bulunamadı: %s\n",
	ShellDownloading:      "İndiriliyor: %s → %s\n",
	ShellDownloaded:       "✓ İndirildi (%s)\n",
	ShellDownloadedBare:   "✓ İndirildi",
	ShellErrFmt:           "Hata: %v\n",
	ShellDeleteConfTitle:  "Silme Onayı",
	ShellCancelShort:      "İptal.",
	ShellDeletedShort:     "✓ Silindi: %s\n",
	ShellPickServerTitle:  "Sunucu Seç",
	ShellSyncServers:          "Sync Edilecek Sunucular",
	ShellCalibrateServers:     "Kalibre Edilecek Sunucular",
	ShellSyncPickSub:      "Space ile işaretle, Enter ile onayla",
	ShellSyncCancelled:    "İptal.",
	ShellFileActionTitle:  "Dosya İşlemi",
	ShellFileActionView:   "İçeriği Görüntüle",
	ShellFileActionGet:    "İndir",
	ShellFileActionDel:    "Sil",
	ShellFileActionCancel: "İptal",
	ShellFileTruncFmt:     "[İlk %d KB — tamamı için: get %s]\n",
	ShellScanning:         "Taranıyor...",
	ShellScannedFmt:       " %d dosya\n",
	ShellConnPoolErr:      "  Bağlantı hatası: %v\n",
	ShellReleaseFmt:       "  Release: %s\n",
	ShellStateReadErr:     "  State okunamadı: %v\n",
	ShellStatusDeleted:    "  - %s (yerel silinmiş)\n",
	ShellStatusUpToDate:   "  Güncel",
	ShellStatusNoChange:   "  %d değişiklik\n",
	ShellStatusStateErr:   "[%s] state okunamadı: %v\n",

	TreeMaxPromptTitle: "Tree görünümü",
	TreeMaxPromptSub:   "Klasör başına max dosya sayısı?",
	TreeSkippedFmt:     "[%d öğe — atlandı, --max ile artırın]",
	TreeErrFmt:         "[hata: %v]",

	ShellHelp: `
Uzak sunucu komutları:
  ls [dizin]              İnteraktif dosya tarayıcısı (↑↓ gezin, → gir/önizle, ← çık)
                            Space = işaretle  |  d = işaretlileri sil  |  m = taşı
                            f = dosya/klasörü freeze/unfreeze  |  / = ara
                            t = mevcut dizini tree olarak göster
  tree [yol] [--max N]    FTP dizinini ağaç görünümünde listeler (önce klasörler, dosya boyutlarıyla)
                            --max N : N'den fazla öğeli klasörlerde ilk N + "+X daha..." göster
                            --max verilmezse etkileşimli seçim (20 / 50 / 100 / tümü)
  cd <dizin>              Uzak dizin değiştir  (cd .. ile üste çık)
  cat [dosya]             Dosya içeriğini görüntüle (arg verilmezse tarayıcı açılır)
  get [dosya] [hedef]     Dosya indir (arg verilmezse tarayıcı açılır)
  rm [-f] [-r] [dosya]    Dosya/dizin sil  (-f onay istemez, -r klasör içeriğiyle)
  pwd                     Uzak dizini göster

Senkronizasyon:
  status                  Yerel değişiklikleri göster (TUI)
  sync [--all] [--full] [--dry-run] [--server ad]   FTP'ye yükle
  calibrate [--all] [--server ad]   Yerel boyutları FTP ile karşılaştır, state güncelle (yükleme yok)
  freeze [--server ad]    Freeze listesi — yüklenmesin istenen dosyaları işaretle

Sunucu yönetimi:
  servers                 Sunucu listesi
  server [ad]             Sunucu seç / bağlan
  disconnect              Mevcut FTP bağlantısını kes (prompt sunucusuz moda döner)
  config                  Sunucu ve bağlantı profili yönetimi — TUI
                            Liste: ↑↓ gezin | Enter/e düzenle | Space aç/kapat | d sil | n yeni
                            ── Global Ayarlar: proje geneli varsayılan dizin, protect, include, exclude
                            ── Bağlantı Profilleri: paylaşılan FTP kimlik bilgileri (host/port/kullanıcı/şifre)
                               Birden fazla sunucu aynı profili kullanabilir.
                               Profil değişince onu kullanan tüm sunucular güncellenir.
                            ── Sunucu düzenle (17 alan):
                               Connection = profil seç (b veya Enter ile picker açılır)
                               Profil seçiliyse credential alanları read-only gösterilir (↳ profil)
                               Yerel dizin: sunucuya özel yerel dizin
                                 "(project default)" = Global Ayarlar'dan miras alır
                               Include/Exclude/Protect: b = yerel dosya tarayıcısı
                                 n = özel yol gir (glob kalıpları, proje dışı yollar)
                            s = kaydet  |  q = iptal

Diğer:
  lang [en|tr]            Dili göster veya değiştir (ilk çalışmada OS dilinden otomatik algılanır)
  clear / cls             Ekranı temizle
  help / ?                Bu yardım
  exit / quit             Çık`,

	// lang command
	LangCurrentFmt:  "Dil: %s\n",
	LangSwitchedFmt: "Dil değiştirildi: %s\n",
	LangSavedFmt:    ".syncftp/lang dosyasına kaydedildi\n",
	LangInvalid:     "Bilinmeyen dil: %q — 'en' veya 'tr' kullanın\n",
	LangAlreadyFmt:  "Zaten %s dili kullanılıyor\n",

	// cmd_sync
	SyncCancelled:      "İptal.",
	SyncDryRunNote:     "[DRY RUN] Hiçbir şey yüklenmeyecek",
	SyncWhitelistFmt:   "Whitelist (%d yol): yalnızca bu yollar sync edilecek\n",
	SyncExcludeFmt:     "Exclude (%d yol): bu yollar bu sync'ten hariç tutulacak\n",
	SyncServerErrFmt:   "  HATA: %v\n",
	SyncNoChange2:      "  Değişiklik yok",
	SyncDeletedHeader:  "  ! SİLİNEN dosyalar (FTP'de bırakıldı):\n",
	SyncProcessingFmt:  "  %d dosya işlenecek",
	SyncProtectedFmt:   " (%d korunuyor)",
	SyncProtectedLabel: "    KORUNUYOR  %s\n",
	SyncUploadLabel:    "    YÜKLENECEK %s\n",
	SyncPoolFmt:        "  Bağlantı havuzu: %d / Retry: %d\n",
	SyncDoneFullFmt:    "  Tamamlandı: %d yüklendi, %d korundu, %d hata\n",
	SyncFailedSavedFmt: "  ! %d başarısız dosya .syncftp/failed/%s.json'a kaydedildi — tekrar denemek için: syncftp sync --retry-failed\n",
	SyncFailedSaveErr:  "  ! Failed listesi kaydedilemedi: %v\n",
	SyncStateErr:       "  ! State kaydedilemedi: %v\n",
	SyncReleaseErr:     "  ! Release oluşturulamadı: %v\n",
	SyncReleaseFmt:     "  Release: %s\n",
	SyncFullFlag:       "  Tam sync (--full): tüm dosyalar yüklenecek",
	SyncRetryNoFiles:   "  Yeniden denenecek başarısız dosya yok",
	SyncRetryModeFmt:   "  Retry modu: %d başarısız dosya (%s)\n",
	SyncRetrySkipFmt:   "  ! %s artık local'de yok, atlanıyor\n",
	SyncAttemptsFmt:    "    ✗ %s (%d deneme): %v\n",
	SyncAttemptOkFmt:   "    ✓ %s (%d. denemede başarılı)\n",
	SyncUploadOkFmt:    "    ✓ %s\n",
	SyncUploadErrFmt:   "    ✗ %s: %v\n",

	// resync
	ResyncScanning:     "  Sunucuya bağlanılıyor...",
	ResyncConnected:    " bağlandı\n",
	ResyncListing:      "Uzak dosyalar listeleniyor...",
	ResyncLocalFmt:     "  Yerel:  %d dosya tarandı\n",
	ResyncConnErr:      "  ! Sunucuya bağlanılamadı (%v) — resync atlandı\n",
	ResyncListErr:      "  ! Uzak dizin okunamadı (%v) — resync atlandı\n",
	ResyncFoundFmt:     "  Uzak:   %d dosya bulundu\n",
	ResyncComparingFmt:   "  Karşılaştırılıyor: %d / %d\n",
	ResyncMatchedFmt:     "  Eşleşti: %d (boyut OK)  |  Farklı/eksik: %d\n",
	ResyncDoneFmt:        "  [%s] kalibre tamamlandı\n",
	ResyncAutoMsg:        "  İlk çalıştırma: kalibre yapılıyor (sunucu karşılaştırması)...\n",
	ResyncNoServers:      "Eşleşen sunucu bulunamadı.\n",
	ResyncIgnoreDirsFmt:  "  Ignore: %d klasör atlandı → %s\n",
	ResyncIgnoreFilesFmt: "  Ignore: %d dosya atlandı\n",
	ResyncFilteredFmt:    "  Filtre (include/exclude): %d dosya kapsam dışı\n",
	ResyncFrozenDiffFmt:  "  ❄ %d frozen dosya farklı/eksik (sync sırasında atlanacak)\n",

	// cmd_status
	StatusWhitelistFmt:     "Whitelist (%d yol): yalnızca bu yollar gösterilecek\n",
	StatusExcludeFmt:       "Exclude (%d yol): bu yollar sonuçtan hariç tutulacak\n",
	StatusProjectFmt:       "Proje : %s\n",
	StatusDirFmt:           "Dizin : %s\n",
	StatusFileFmt:          "Dosya : %d adet\n",
	StatusStateErr:         "[%s] State yüklenemedi: %v\n\n",
	StatusNoFirstSync:      "  Henüz ilk sync yapılmadı",
	StatusUpToDate:         "  Değişiklik yok — sunucu güncel",
	StatusFilteredUpToDate: "  Belirtilen filtre kapsamında değişiklik yok",
	StatusNewHeader:        "  + YENİ",
	StatusChangedHeader:    "  ~ DEĞİŞEN",
	StatusDeletedHeader:    "  - SİLİNEN (FTP'den silinmez, sadece bilgi)",

	// cmd_push
	PushSourceFmt:     "Kaynak  : %s\n",
	PushTargetFmt:     "Hedef   : %s:%d%s\n",
	PushScanning:      "Taranıyor...",
	PushScannedFmt:    " %d dosya\n\n",
	PushFirstPush:     "İlk push",
	PushFullPush:      "Tam push (--full)",
	PushFirstFmt:      "  %s — tüm dosyalar yüklenecek\n",
	PushDeletedHeader: "  ! SİLİNEN dosyalar (FTP'de bırakıldı):\n",
	PushNoChange:      "  Değişiklik yok — hedef güncel",
	PushProcessingFmt: "  %d dosya işlenecek\n",
	PushDryRunHeader:  "\n[DRY RUN] Yüklenecekler:",
	PushConnFmt:       "  Bağlantı: %d / Retry: %d\n",
	PushAttemptsFmt:   "    ✗ %s (%d deneme): %v\n",
	PushAttemptOkFmt:  "    ✓ %s (%d. denemede)\n",
	PushUploadOkFmt:   "    ✓ %s\n",
	PushUploadErrFmt:  "    ✗ %s: %v\n",
	PushDoneFmt:       "\n  Tamamlandı: %d yüklendi, %d hata\n",
	PushFailedHint:    "  ! Başarısız dosyalar kaydedildi — tekrar: syncftp push %s --server %s --full\n",
	PushStateErr:      "  ! State kaydedilemedi: %v\n",
	PushReleaseFmt:    "  Release: %s\n",

	// cmd_remote
	RemoteServerFmt:      "Sunucu : %s (%s)\n",
	RemoteDirFmt:         "Dizin  : %s\n\n",
	RemoteConnecting:     "Bağlanıyor: %s (%s)...\n",
	RemoteDownloading:    "İndiriliyor: %s → %s\n",
	RemoteDownloaded:     "✓ İndirildi (%s)\n",
	RemoteDownloadedBare: "✓ İndirildi",
	RemoteDeleteLabel:    "Silinecek %s: %s\n",
	RemoteDeleteConfirm:  "Silmek istediğinizden emin misiniz? [y/N]: ",
	RemoteDeleteCancel:   "İptal edildi.",
	RemoteDeletedFmt:     "✓ Silindi: %s\n",
	RemoteCatTruncFmt:    "\n[İlk %d KB gösterildi — tamamını indirmek için: syncftp remote get %s]\n",
	RemotePickDirErr:     "dizin listelenemedi: %w",
	RemotePickUpDir:      "  [0] .. (üst dizin)",
	RemotePickPromptFmt:  "  Seçim [0-%d, q=iptal]: ",
	RemotePickInvalid:    "  Geçersiz seçim: %q\n",
	RemotePickSelected:   "  Seçildi: %s\n\n",
	RemoteFileLabel:      "dosya",
	RemoteDirLabel:       "dizin",
	RemoteDirRecLabel:    "dizin (içeriğiyle)",
	RemoteListErr:        "%s  ! listelenemedi: %v\n",
	RemoteNoServerSel:    "sunucu seçilmedi",

	// failsafe cleanup
	SyncCleanupRemovedFmt: "  ⚠ %q artık diskte yok, listeden çıkarıldı\n",
	SyncConfigSaveErr:     "  ! Config kaydedilemedi: %v\n",

	// webhook
	SyncWebhookSentFmt: "  ✓ Webhook: %s (HTTP %d)\n",
	SyncWebhookErrFmt:  "  ! Webhook hatası: %v\n",

	// block filter
	SyncBlockedFmt: "  ✗ %d engellendi (uzantı/dosya filtresi)\n",

	// minify / obfuscate
	SyncMinifiedFmt:   "  ✓ %d dosya küçültüldü (minify)\n",
	SyncObfuscatedFmt: "  ✓ %d dosya obfuscate edildi\n",
	SyncMinifyErrFmt:  "  ! Minify hatası (%s): %v — orijinal yükleniyor\n",
	SyncObfuscErrFmt:  "  ! Obfuscate hatası (%s): %v — minified yükleniyor\n",
	SyncNoTerserWarn:  "  ⚠ terser PATH'te bulunamadı — obfuscation atlandı\n",

	// sync summary TUI
	SyncSummaryTitle:   "  Sync Özeti",
	SyncSummaryNavHint: "  ↑↓ gezin  |  → = detay  |  q = çık",
	SyncSummaryOk:      "✓",
	SyncSummaryErrFmt:  "HATA: %v",

	// status TUI
	StatusDetailChangesCountFmt: "  %d değişiklik",
	StatusSyncConfirm:           "  Sync başlatılsın?",
	StatusSyncHint:              "  Enter/s = Evet  |  ESC = İptal",
	StatusSearchHint:            "  Backspace = sil  |  Esc = aramayı kapat",
	StatusDetailNavHint:         "  ↑↓/g/G kaydır  |  / = ara  |%s  ←/Esc = geri  |  q = çık",
	StatusDetailSyncPart:        "  s = Sync  |",
	StatusNoMatch:               "  (eşleşme yok)",
	StatusUpToDateShort:         "  ✓ güncel",
	StatusScrollFmt:             "  %d/%d",
	StatusResultsFmt:            "  %d sonuç",
	StatusListTitleFmt:          "  Status — %s",
	StatusListNavHint:           "  ↑↓ gezin  |  → = detay  |  q = çık",
	StatusOkShort:               "✓ güncel",
	StatusChangesListFmt:        "%d değişiklik",
	StatusFrozenFmt:             "❄ %d frozen",

	// cmd_init
	InitWizardTitle:   "=== syncFTP Kurulum Sihirbazı ===",
	InitProjectName:   "Proje adı",
	InitLocalDir:      "Yerel dizin",
	InitFTPHeader:     "FTP Sunucu bilgileri:",
	InitServerName:    "Sunucu adı",
	InitHost:          "FTP host",
	InitPort:          "Port",
	InitUser:          "Kullanıcı adı",
	InitPassword:      "Şifre",
	InitRemotePath:    "Uzak dizin",
	InitCreated:       "✓ syncftp.json oluşturuldu (izinler: 600)",
	InitReadyFmt:      "Hazır! 'syncftp sync' komutu ile %q projesini senkronize edebilirsiniz.\n",
	InitIgnoreExists:  "  (syncFTP girdileri zaten %s içinde)\n",
	InitIgnoreAdded:   "✓ syncftp.json, syncftp.exe → %s'e eklendi\n",
	InitIgnoreCreated: "✓ syncftp.ignore oluşturuldu (syncftp.json, syncftp.exe eklendi)",

	// Config TUI — server manager
	CfgSrvTitle:          "⚙  Server Ayarları",
	CfgSrvNavHint:        "  ↑↓ gezin  |  Enter/e = düzenle  |  Space = aç/kapat  |  d = sil  |  n = yeni  |  q = çık",
	CfgDeleteConfirmFmt:  "[%s] silinsin mi?  y = evet, diğer = iptal",
	CfgGlobalLabel:       "⚙ Global Ayarlar",
	CfgGlobalHint:        "  (protect, include, exclude, ignore_files)",
	CfgConnProfilesLabel: "🔗 Bağlantı Profilleri",
	CfgConnCountFmt:      "(%d profil)",
	CfgNewServerLabel:    "+ Yeni server ekle",
	CfgSrvCountFmt:       "  %d server  |  %d bağlantı profili",
	CfgSaveErr:           "Kayıt hatası: %v\n",
	CfgUpdatedFmt:        "✓ [%s] güncellendi\n",
	CfgAddedFmt:          "✓ [%s] eklendi\n",
	CfgDeletedFmt:        "✓ [%s] silindi\n",
	CfgEnabledLabel:      "aktif",
	CfgDisabledLabel:     "devre dışı",

	// Config TUI — server edit
	CfgSrvEditTitle:   "Server Düzenle",
	CfgSrvNewTitle:    "Yeni Server",
	CfgSrvEditNavHint: "  ↑↓ gezin  |  Enter/b = düzenle/seç  |  Space = bool toggle  |  s = kaydet  |  q = iptal",
	CfgSrvSaveHint:    "  s = kaydet  |  q/Esc = iptal",
	CfgSrvManual:      "(manuel)",
	CfgSrvProjectDef:  "(project default)",
	CfgSrvConnChange:  "[b=değiştir]",
	CfgSrvConnSelect:  "[b=profil seç]",
	CfgSrvBrowseHint:  "[b=gözat]",

	// Config TUI — server field labels
	CfgFldName:        "Ad",
	CfgFldConnection:  "Connection",
	CfgFldHost:        "Host",
	CfgFldPort:        "Port",
	CfgFldUser:        "Kullanıcı",
	CfgFldPassword:    "Şifre",
	CfgFldRemotePath:  "Uzak dizin",
	CfgFldLocalPath:   "Yerel dizin",
	CfgFldEnabled:     "Aktif",
	CfgFldPassive:     "Passive mode",
	CfgFldDisableEPSV: "EPSV devre dışı",
	CfgFldNAT:         "NAT workaround",
	CfgFldMaxConn:     "Max bağlantı",
	CfgFldMaxRetry:    "Max retry",
	CfgFldInclude:     "Include",
	CfgFldExclude:     "Exclude",
	CfgFldProtect:    "Protect",
	CfgFldWebhook:    "Webhook URL",
	CfgFldMinify:     "CSS/JS küçült",
	CfgFldObfuscate:  "JS gizle (obfuscate)",
	CfgFldBlockExts:  "Engellenen uzantılar",
	CfgFldBlockFiles: "Engellenen dosyalar",

	// Config TUI — global settings
	CfgGlobalEditTitle:    "⚙  Global Ayarlar",
	CfgGlobalEditNavHint:  "  ↑↓ gezin  |  Enter = düzenle/aç  |  b = yerel dosya gözat  |  s = kaydet  |  q = iptal",
	CfgGlobalDefaultEmpty: "(boş = çalışma dizini)",
	CfgGlobalIgnoreDefVal: "(varsayılan: .gitignore + syncftp.ignore)",
	CfgGlobalIgnoreNone:   "(hiçbiri — ignore kullanılmaz)",
	CfgGlobalIgnoreHint:   "  b=gözat ile ekle  |  Enter=yönet  |  s=kaydet  |  q=iptal",

	// Config TUI — global field labels
	CfgGFldDefaultPath: "Varsayılan dizin",
	CfgGFldProtect:     "Protect",
	CfgGFldInclude:     "Include (global)",
	CfgGFldExclude:     "Exclude (global)",
	CfgGFldIgnore:      "Ignore files",
	CfgGFldWebhook:     "Webhook URL (global)",
	CfgGFldBlockExts:   "Engellenen uzantılar (global)",
	CfgGFldBlockFiles:  "Engellenen dosyalar (global)",

	CfgValueEmpty: "(boş)",

	// Config TUI — connection manager
	CfgConnMgrTitle:   "🔗  Bağlantı Profilleri",
	CfgConnMgrNavHint: "  ↑↓ gezin  |  Enter/e = düzenle  |  d = sil  |  n = yeni  |  q = çık",
	CfgNewConnLabel:   "+ Yeni profil ekle",
	CfgConnTotalFmt:   "  %d profil",

	// Config TUI — connection edit
	CfgConnEditTitle:   "🔗  Profil Düzenle",
	CfgConnNewTitle:    "🔗  Yeni Profil",
	CfgConnEditNavHint: "  ↑↓ gezin  |  Enter = düzenle/toggle  |  Space = bool toggle  |  s = kaydet  |  q = iptal",

	// Config TUI — connection picker
	CfgConnPickerTitle:      "Bağlantı Profili Seç",
	CfgConnPickerSub:        "Bağlantı bilgileri buradan alınacak",
	CfgConnPickerCurrentFmt: "Mevcut: %s",
	CfgConnPickerManual:     "(manuel)",
	CfgConnPickerManualDesc: "Bilgileri elle gir",

	// Config TUI — local browser
	CfgLocalNavHint:      "  Space=işaretle  |  n=özel yol  |  Enter/→=gir  |  ←/Esc=çık  |  a=tümü  |  s=kaydet  |  q=iptal",
	CfgLocalDirPickHint:  "  ↑↓ gezin  |  Enter/→=klasöre gir  |  Space/s=bu dizini seç  |  ←/Esc=çık  |  q=iptal",
	CfgLocalCustomHint:   "  Dizin dışı dosya/klasör yolu girebilirsiniz (örn: ../shared/config.php)",
	CfgLocalEmpty:        "  (boş klasör)",
	CfgLocalSelectFmt:    "  Şu an: %s  |  Space veya s = bu dizini seç",
	CfgLocalPathLabel:    "  Yol: ",
	CfgLocalCustomSave:   "  Enter=ekle  |  Esc=iptal",
	CfgLocalMarkedFmt:    "  ✓ %d işaretli  +%d özel yol  |  n=özel yol ekle",
	CfgLocalMarkedSimple: "  ✓ %d işaretli  |  n=özel yol ekle",

	// Config TUI — ignore picker
	CfgIgnoreTitle:   "⚙  Ignore dosyaları",
	CfgIgnoreNavHint: "  Space = seç/kaldır  |  s = kaydet  |  q = iptal",
	CfgIgnoreNote:    "  Hiçbirini seçmezsen ignore kullanılmaz.",
}

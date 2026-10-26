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
	ShellSyncServers:      "Sync Edilecek Sunucular",
	ShellCalibrateServers: "Kalibre Edilecek Sunucular",
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
  watch [--all] [--server ad]   Dosya değişikliklerini izle, değişince otomatik sync yap
  calibrate [--all] [--server ad]   Yerel boyutları FTP ile karşılaştır, state güncelle (yükleme yok)
  milestone [set [tarih] | sync | clear]   Bağlı sunucuya zaman işareti koy;
                            'milestone sync' sadece o tarihten sonra değişen dosyaları yükler
                            tarih: bugün, dün, 3d, 5h, 14:30, 20.07[.2026],
                            2026-07-20 [15:04] (boş = şimdi)
  freeze [--server ad]    Freeze listesi — yüklenmesin istenen dosyaları işaretle
  logs [dosya | son]      Sync loglarını listele / göster (bağlıysan o sunucunun)
  rollback [id] [--list]  Bir sync'i geri al: eski dosyalar geri yüklenir, yeni eklenenler silinir
                            id verilmezse release seçici açılır (bağlı sunucu / seçici)
  verify [--fix]          Sunucuda elle değiştirilen/silinen dosyaları bul (drift);
                            --fix sonraki sync'te yeniden yüklenmelerini sağlar
  detect-limit [--max N]  Sunucunun eş zamanlı bağlantı limitini ölçer, max_connections'a yazar
  exec <komut>            Bağlı SFTP sunucusunda SSH ile komut çalıştır (örn. php artisan migrate)

Sunucu yönetimi:
  servers                 Sunucu listesi
  server [ad]             Sunucu seç / bağlan
  disconnect              Mevcut FTP bağlantısını kes (prompt sunucusuz moda döner)
  config                  Sunucu ve bağlantı profili yönetimi — TUI
  config --export         Config'i şifresiz JSON olarak yazdır (paylaşılabilir)
                            Liste: ↑↓ gezin | Enter/e düzenle | Space aç/kapat | d sil | n yeni
                            ── Global Ayarlar: varsayılan dizin, protect, include, exclude,
                               webhook URL, engellenen uzantılar
                            ── Bağlantı Profilleri: paylaşılan FTP kimlik bilgileri (host/port/kullanıcı/şifre)
                               Birden fazla sunucu aynı profili kullanabilir.
                               Profil değişince onu kullanan tüm sunucular güncellenir.
                            ── Sunucu düzenle (21 alan):
                               Connection = profil seç (b veya Enter ile picker açılır)
                               Profil seçiliyse credential alanları read-only gösterilir (↳ profil)
                               Yerel dizin: sunucuya özel yerel dizin
                                 "(project default)" = Global Ayarlar'dan miras alır
                               Include/Exclude/Protect: b = yerel dosya tarayıcısı
                                 n = özel yol gir (glob kalıpları, proje dışı yollar)
                               Minify CSS/JS: yorum ve boşlukları kaldırır (pure Go, bağımlılık yok)
                               Obfuscate JS: terser --compress --mangle çalıştırır (gerekli: npm i -g terser)
                               Engellenen uzantılar: uzantıya göre dosya atla (.jpg .zip .pdf vb.)
                               Webhook URL: her sync sonrası HTTP POST ile özet gönder
                            s = kaydet  |  q = iptal

Diğer:
  serve [--port N]        Yerel HTTP API sunucusu başlatır (varsayılan: 127.0.0.1:8080)
                            GET  /api/servers, /api/status, /api/failed, /api/releases
                            POST /api/sync         (senkron, JSON sonuç döner)
                            GET  /api/sync/stream  (SSE canlı log)
                            POST /api/reload       (config doğrula)
                            POST /api/trigger/github?server=&branch=  (SYNCFTP_GITHUB_SECRET)
                            GET/DELETE /api/remote/ls|cat|get|rm
  gui                     Masaüstü yönetim panelini aç (ayrı pencere)
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
	SyncCanceled:       "  ■ Sync durduruldu",
	SyncCanceledFmt:    "  ■ Sync durduruldu: %d dosya yüklendi, %d dosya bekliyor\n",
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
	ResyncScanning:       "  Sunucuya bağlanılıyor...",
	ResyncConnected:      " bağlandı\n",
	ResyncListing:        "Uzak dosyalar listeleniyor...",
	ResyncLocalFmt:       "  Yerel:  %d dosya tarandı\n",
	ResyncConnErr:        "  ! Sunucuya bağlanılamadı (%v) — resync atlandı\n",
	ResyncListErr:        "  ! Uzak dizin okunamadı (%v) — resync atlandı\n",
	ResyncFoundFmt:       "  Uzak:   %d dosya bulundu\n",
	ResyncComparingFmt:   "  Karşılaştırılıyor: %d / %d\n",
	ResyncMatchedFmt:     "  Eşleşti: %d (boyut OK)  |  Farklı/eksik: %d\n",
	ResyncDoneFmt:        "  [%s] kalibre tamamlandı\n",
	ResyncAutoMsg:        "  İlk çalıştırma: kalibre yapılıyor (sunucu karşılaştırması)...\n",
	ResyncNoServers:      "Eşleşen sunucu bulunamadı.\n",
	ResyncIgnoreDirsFmt:  "  Ignore: %d klasör atlandı → %s\n",
	ResyncIgnoreFilesFmt: "  Ignore: %d dosya atlandı\n",
	ResyncFilteredFmt:    "  Filtre (include/exclude): %d dosya kapsam dışı\n",
	ResyncFrozenDiffFmt:  "  ❄ %d frozen dosya farklı/eksik (sync sırasında atlanacak)\n",
	ResyncListTUIFmt:     "  Listeleniyor: %d dosya   (→ dosyaları göster | ← gizle | ctrl+c iptal)",
	ResyncListCancelled:  "  ! Listeleme iptal edildi — calibrate atlandı\n",
	ResyncCancelled:      "  ! Calibrate durduruldu — state değişmedi\n",
	ResyncLocalScanning:  "  Yerel dosyalar taranıyor…\n",
	ResyncConnectedLine:  "  ✓ Bağlandı\n",
	ResyncListLiveFmt:    "    … %d dosya listelendi · %s\n",
	ResyncCompareLiveFmt: "    … karşılaştırılıyor %d / %d\n",
	ResyncLogFilesHeader: "\n── Taranan uzak dosyalar (%d) ──\n",

	MilestoneSetFmt:     "  ✓ Milestone [%s]: %s\n",
	MilestoneShowFmt:    "  %s: %s\n",
	MilestoneNoneFmt:    "  %s: (milestone yok)\n",
	MilestoneClearedFmt: "  ✓ Milestone silindi: %s\n",
	MilestoneNotSetFmt:  "  ⚠ %s: milestone tanımlı değil — önce 'milestone set' çalıştır\n",
	MilestoneDateErrFmt: "  ✗ Tarih anlaşılamadı: %q (örn: bugün, dün, 3d, 5h, 14:30, 20.07.2026, 2026-07-20 15:04)\n",
	MilestoneSinceFmt:   "  ⏱ Milestone: %s sonrası değişen dosyalar sync edilecek\n",
	MilestoneFilterFmt:  "  ⏱ Milestone filtresi (%s): %d dosya tarih öncesi — atlandı\n",
	MilestoneShellNoSrv: "  Önce bir sunucuya bağlan (server <ad>) — milestone bağlı sunucuya işler.\n",
	MilestoneUsage:      "  Kullanım: milestone [set [tarih] | sync [--dry-run] | clear | show]\n  Tarih: bugün, dün, 3d, 5h, 14:30, 20.07[.2026], 2026-07-20 [15:04] (boş = şimdi)\n",

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
	// watch
	WatchStartFmt:  "  ✓ [%s] izleniyor: %s\n",
	WatchReady:     "Değişiklikler bekleniyor. Durdurmak için Ctrl+C.\n",
	WatchChangeFmt: "  ⟳ [%s] değişiklik tespit edildi, sync başlatılıyor...\n",
	WatchErrFmt:    "  ! Watch hatası: %v\n",
	WatchNoServers: "İzlenecek aktif sunucu bulunamadı.\n",

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
	CfgFldProtect:     "Protect",
	CfgFldWebhook:     "Webhook URL",
	CfgFldMinify:      "CSS/JS küçült",
	CfgFldObfuscate:   "JS gizle (obfuscate)",
	CfgFldBlockExts:   "Engellenen uzantılar",

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

	// Sync log + rapor
	SyncLogSavedFmt: "  Log: %s\n",
	SyncLogHint:     "  ·  s = log kaydet",
	SyncReportFmt:   "  ↑ %s · %s",

	// Hook'lar (pre_sync / post_sync)
	SyncHookRunFmt:   "  ⚙ hook: %s\n",
	SyncHookAbortFmt: "  ✗ pre_sync hook başarısız, sync iptal: %v\n",
	SyncHookErrFmt:   "  ! post_sync hook hatası: %v\n",

	// Status — frozen
	StatusProtectedHeader: "  🔒 Korunan (protect — yüklenmez)",
	StatusBlockedHeader:   "  ✗ Engelli uzantı (yüklenmez)",
	StatusFrozenHeader:    "  ❄ Frozen (sync'te atlanır)",

	// Calibrate — canlı listeleme sayacı
	ResyncListProgressFmt: "\r  Listeleniyor: %d dosya",

	// Init — ignore şablonu
	InitTemplateTitle:      "Ignore şablonu",
	InitTemplateSub:        ".syncftp/syncftp.ignore için şablon seç",
	InitTemplateGeneric:    "Genel",
	InitTemplateNone:       "Yok",
	InitTemplateNoneDesc:   "Şablon oluşturma",
	InitTemplateWrittenFmt: "  ✓ .syncftp/syncftp.ignore oluşturuldu (%s)\n",

	// Diff komutu
	DiffPickA:         "Birinci sunucu",
	DiffPickB:         "İkinci sunucu",
	DiffPickSub:       "Karşılaştırılacak sunucuyu seç",
	DiffNeedTwo:       "  Karşılaştırma için iki farklı sunucu gerekir.\n",
	DiffTitleFmt:      "── %s ↔ %s ──\n",
	DiffListingFmt:    "  %s listeleniyor...\n",
	DiffOnlyFmt:       "  Sadece %s üzerinde (%d):\n",
	DiffSizeHeaderFmt: "  Boyut farklı (%d):\n",
	DiffSizeLineFmt:   "    ~ %s  (%s: %s · %s: %s)\n",
	DiffSameFmt:       "  ✓ %d dosya aynı\n",

	// Upload çekirdeği (yedek / limit)
	SyncBandwidthFmt:     "  Hız sınırı: %d KB/s\n",
	SyncPoolLimitedFmt:   "  ! %d/%d bağlantı açılabildi (sunucu/host bağlantı limiti)\n",
	SyncBackupErrFmt:     "  ! Yedek alınamadı (%s): %v — bu dosya geri alınamaz\n",
	SyncBackupSummaryFmt: "  ↺ Yedek: %d dosyanın eski hali saklandı, %d yeni dosya — geri almak için: syncftp rollback --server %s\n",

	// Watch
	WatchStopped: "\nWatch durduruldu.\n",

	// Rollback
	RollbackNoReleasesFmt:      "  %s için geri alınabilir release yok.\n",
	RollbackPickTitleFmt:       "Rollback — %s",
	RollbackPickSubtitle:       "Seçilen release VE ondan sonraki tüm release'ler geri alınır",
	RollbackConfirmFmt:         "%[2]s üzerinde %[1]d release geri alınsın mı?",
	RollbackReleaseDescFmt:     "%s · %d dosya · %d değişen · %d yeni",
	RollbackTagRolledBack:      "[geri alındı]",
	RollbackTagNoBackup:        "[yedek yok]",
	RollbackAlreadyFmt:         "%s zaten geri alınmış",
	RollbackNoBackupFmt:        "%s release'inin yedeği yok (yedek kapalı ya da silinmiş) — bu noktanın gerisine dönülemez",
	RollbackNothingSelectedFmt: "%s release'inde seçilen dosyalardan geri alınabilecek yok (zaten geri alınmış ya da bu release'te değil)",
	RollbackNotFoundFmt:        "release bulunamadı: %s",
	RollbackReleaseFmt:         "\n  ↺ %s  (%s)\n",
	RollbackRestoredFmt:        "    ↺ geri yüklendi  %s\n",
	RollbackDeletedFmt:         "    ✗ silindi        %s\n",
	RollbackFileErrFmt:         "    ! %s: %v\n",
	RollbackSkippedNoBackupFmt: "    ! %d dosyanın yedeği yoktu, olduğu gibi bırakıldı\n",
	RollbackDoneFmt:            "\n  Rollback tamam: %d geri yüklendi, %d silindi, %d hata\n",

	// detect-limit / verify / secure / project
	DetectStartFmt:      "  [%s] bağlantı limiti ölçülüyor (%s://%s)...\n",
	DetectProgressFmt:   "\r    %d bağlantı açık...",
	DetectErrFmt:        "    ! bağlanılamadı: %v\n",
	DetectResultFmt:     "    ✓ sunucu aynı anda %d bağlantı kabul ediyor\n",
	DetectHitMaxFmt:     "    ✓ %d bağlantı limitsiz açıldı (daha fazlasını denemek için --max'ı artırın)\n",
	DetectApplyFmt:      "    → %s: max_connections %d → %d\n",
	DetectSaved:         "  ✓ Config'e kaydedildi.\n",
	VerifyStartFmt:      "  [%s] sunucu doğrulanıyor (%s://%s)...\n",
	VerifyErrFmt:        "  [%s] ! %v\n",
	VerifyMissingFmt:    "    Sunucuda eksik (%d):\n",
	VerifyModifiedFmt:   "    Sunucuda değişmiş (%d):\n",
	VerifyMoreFmt:       "      … +%d tane daha\n",
	VerifyCleanFmt:      "    ✓ Sapma yok — %d dosya eşleşiyor\n",
	VerifySummaryFmt:    "    %d tamam · %d eksik · %d değişmiş\n",
	VerifyPendingFmt:    "    (%d yerel değişiklik henüz sync edilmemiş — sapma değil)\n",
	VerifyFixedFmt:      "    ✓ %d dosya state'ten düşürüldü — sonraki sync yeniden yükler\n",
	VerifyFixHint:       "    İpucu: --fix ile bu dosyalar sonraki sync'te yeniden yüklenir\n",
	SecureOnFmt:         "  ✓ Anahtarlık modu açık — %d şifre işletim sistemi anahtarlığına taşındı\n",
	SecureOff:           "  ✓ Anahtarlık modu kapalı — şifreler tekrar syncftp.json'da\n",
	SecureUnresolvedFmt: "bazı şifreler anahtarlıktan okunamadı: %s",
	ProjectNotFoundFmt:  "proje bulunamadı: %s (bkz: syncftp project list)",
	ProjectListEmpty:    "  Kayıtlı proje yok. Proje dizininde `syncftp project add` çalıştırın.\n",
	ProjectAddedFmt:     "  ✓ %s → %s\n",
	ProjectRemovedFmt:   "  ✓ %s kayıttan çıkarıldı\n",
	ProjectRenamedFmt:   "  ✓ %s → %s\n",

	// Config TUI — protokol / tünel / yedek alanları
	CfgFldProtocol:        "Protokol (ftp/ftps/sftp)",
	CfgFldKeyFile:         "SSH anahtar dosyası",
	CfgFldTLSSkip:         "FTPS: sertifika atla",
	CfgFldTunnelHost:      "SSH tünel host",
	CfgFldTunnelPort:      "SSH tünel port",
	CfgFldTunnelUser:      "SSH tünel kullanıcı",
	CfgFldTunnelPass:      "SSH tünel şifre",
	CfgFldTunnelKey:       "SSH tünel anahtar",
	CfgFldBandwidth:       "Hız sınırı (KB/s)",
	CfgFldNoBackup:        "Yedeği kapat (rollback)",
	CfgFldDirectUpload:    "Doğrudan yükle (tmp+rename yok)",
	CfgGFldBandwidth:      "Hız sınırı (KB/s, sunucu başı)",
	CfgGFldNoBackup:       "Yedeği kapat (rollback)",
	CfgGFldBackupKeep:     "Tutulacak yedek (release)",
	CfgGFldDirectUpload:   "Doğrudan yükle (tmp+rename yok)",
	CfgGFldKeyring:        "Şifreler OS anahtarlığında",
	CfgBandwidthGlobal:    "(global)",
	CfgBandwidthUnlimited: "sınırsız",

	// Shell / GUI
	ShellReconnectErrFmt: "  ! Yeniden bağlanılamadı: %v\n",
	ShellReconnectedFmt:  "  ✓ %s yeniden bağlandı\n",
	GuiLaunched:          "  ✓ Yönetim paneli ayrı pencerede açıldı\n",
	GuiStartFmt:          "syncFTP paneli → %s\n",
	GuiNoBrowser:         "Edge/Chrome bulunamadı — varsayılan tarayıcıda açılıyor\n",

	// GUI backend
	GuiBusyFmt:             "%s üzerinde zaten çalışan bir iş var (%s)",
	GuiPathRequired:        "klasör yolu gerekli",
	GuiDirNotFoundFmt:      "klasör bulunamadı: %s",
	LocalDirMissingFmt:     "yerel klasör yok: %s (%s) — sunucunun yerel dizinini düzeltin",
	GuiJobTree:             "Ağaç",
	GuiJobDelete:           "Silme",
	RemoteDelItemFmt:       "  ✗ %s\n",
	RemoteDelDoneFmt:       "\n  ✓ %d dosya, %d klasör silindi\n",
	RemoteDelStoppedFmt:    "\n  ! Durduruldu — %d dosya, %d klasör silindi, kalanlara dokunulmadı\n",
	TreeMoreFmt:            "+%d daha...",
	LangSavedUserFmt:       "Kullanıcı geneli tercih olarak kaydedildi (panelle ortak)\n",
	LogsNone:               "  Henüz log yok (her değişiklikli sync / rollback .syncftp/logs/ altına kaydedilir)",
	LogsHint:               "\n  Göstermek için: syncftp logs <dosya>  ·  en yenisi: syncftp logs son",
	ServerNotFoundFmt:      "sunucu bulunamadı: %q",
	NoServerSelected:       "sunucu seçilmedi",
	GuiIgnoreNoFile:        "ignore_files listesinde syncFTP'ye özel bir dosya yok (yalnızca .gitignore okunuyor) — önce listeye .syncftp/syncftp.ignore ekleyin",
	GuiOutsideBaseFmt:      "%s, panelin açıldığı klasörün (%s) dışında — o klasördeki syncftp ile yönetin",
	GuiValConnName:         "her bağlantı profilinin bir adı olmalı",
	GuiValDupFmt:           "ad iki kez kullanılmış: %s",
	GuiValServerName:       "her sunucunun bir adı olmalı",
	GuiValConnMissingFmt:   "%s: %q bağlantı profili yok",
	GuiValHostFmt:          "%s: host gerekli (ya da bağlantı profili seçin)",
	GuiValRemoteFmt:        "%s: uzak dizin gerekli",
	GuiJobSync:             "Sync",
	GuiJobDryRun:           "Dry-run",
	GuiJobMilestoneSync:    "Milestone sync",
	GuiJobCalibrate:        "Calibrate",
	GuiJobVerify:           "Doğrula",
	GuiJobEdit:             "Düzenle",
	GuiSessionInvalidHTML:  "Oturum geçersiz — paneli <code>syncftp gui</code> ile yeniden açın.",
	GuiNeedOneServer:       "bu işlem için tam olarak bir sunucu seçilmeli",
	GuiKnownHostChanged:    "known_hosts dosyası değişmiş — liste yenilendi, tekrar deneyin",
	ServeBannerFmt:         "syncFTP API → http://%s\n\n",
	ServeTokenOn:           "Token zorunlu: X-Token başlığı, Authorization: Bearer ya da ?token=",
	ServeEndpoints:         "Endpoint'ler:",
	ServeStopHint:          "\nDurdurmak için Ctrl+C",
	ServeForbidden:         "izin verilmeyen Host / Origin (yalnızca yerel istekler)",
	ServeUnauthorized:      "token geçersiz ya da eksik",
	ServeMethodFmt:         "%s bekleniyor",
	ServePathRequired:      "path parametresi zorunlu",
	ServeServerRequiredFmt: "server parametresi zorunlu (mevcut: %s)",
	ServeBadJSONFmt:        "geçersiz JSON: %v",
	ServeSSEUnsupported:    "SSE bu sunucu tarafından desteklenmiyor",
	ServeConfigLoadFmt:     "config yüklenemedi: %v",
	ServeConfigInvalidFmt:  "config geçersiz: %v",
	ServeSyncDone:          "[sync tamamlandı]",
	ServeBodyReadFmt:       "body okunamadı: %v",
	ServeSecretRequired:    "imza doğrulaması zorunlu: SYNCFTP_GITHUB_SECRET ortam değişkenini ayarlayın",
	ServeBadSignature:      "geçersiz imza",
	EditConflict:           "sunucudaki dosya sen açtıktan sonra değişmiş",
	EditSavedLocalFmt:      "  ✓ yerel dosya güncellendi: %s\n",
	GuiJobRollback:         "Rollback",
	GuiJobDetect:           "Bağlantı limiti ölç",
	GuiJobWatch:            "Watch",
	GuiScannedFmt:          "  Tarandı: %s → %d dosya\n\n",
	GuiNoMilestoneFmt:      "%s için milestone yok",
	GuiJobNotCancelable:    "bu iş iptal edilemez",
	GuiWatchRunning:        "bu projede watch zaten çalışıyor",
	WatchLockedFmt:         "bu proje başka bir syncftp sürecinde zaten izleniyor (PID %d)",

	// Tray / CI-CD deploy
	TrayOpen:             "Paneli aç",
	TrayIdle:             "Boşta",
	TrayStatusFmt:        "%d iş · %d watch",
	TrayWatchesHeader:    "Aktif watch'lar (tıkla → durdur)",
	TrayStopWatchFmt:     "%s  (%s)",
	TrayAutostart:        "Windows açılışında başlat",
	TrayQuit:             "Çıkış",
	TrayHookTunnel:       "webhook tüneli açık",
	GuiJobDeploy:         "Deploy",
	GuiJobInstall:        "cloudflared kurulumu",
	DeployStartFmt:       "══ deploy %s / %s ══  %s\n",
	DeployGitPull:        "  git pull --ff-only\n",
	DeployGitPullErrFmt:  "  ✗ git pull başarısız: %v — deploy iptal\n",
	DeploySrvErrFmt:      "  ✗ %s: %v\n",
	DeployDone:           "  ✓ deploy tamamlandı\n",
	DeployQueuedRun:      "\n  ↻ deploy sırasında yeni push geldi — tekrar çalışıyor\n",
	DeployNoCloudflared:  "cloudflared bulunamadı — panelden (CI/CD sekmesi) kurun ya da PATH'e ekleyin",
	DeployDownloadingFmt: "  indiriliyor: %s\n",
	DeployInstalledFmt:   "  ✓ kuruldu: %s (%s)\n",
	DeployNoTriggers:     "  Deploy tetikleyicisi yok. Panelden (CI/CD sekmesi) ya da syncftp.json'daki \"deploy\" dizisinden ekleyin.\n",
	DeployListeningFmt:   "syncFTP webhook dinleyicisi → %s\n",
	DeployTunnelFmt:      "Public URL (trycloudflare) → %s\n",
	DeployCtrlC:          "\nWebhook bekleniyor — durdurmak için Ctrl+C.\n",
	WebhookTestFmt:       "syncFTP test bildirimi (%s)",

	// GUI 2.1
	GuiJobDiff: "Sunucuları karşılaştır",

	// SSH uzak komutlar
	RemoteHookSkipFmt: "  ! [%s] uzak komutlar atlandı — yalnızca SFTP (SSH) sunucularında çalışır\n",
	RemoteHookRunFmt:  "  ⚙ ssh $ %s\n",
	ShellExecUsage:    "  Kullanım: exec <komut>  — bağlı SFTP sunucusunda (remote_path içinde) çalışır\n",
	ShellExecNoSSH:    "  ! exec için SFTP (SSH) bağlantısı gerekir — bu sunucu FTP kullanıyor\n",
	GuiJobExec:        "SSH komutu",
	CfgFldRemotePre:   "Uzak pre_sync (SSH)",
	CfgFldRemotePost:  "Uzak post_sync (SSH)",

	// AI asistan (yalnızca GUI)
	AIBusy:      "asistan hâlâ cevap veriyor — bekleyin ya da durdurun",
	AINoChat:    "bu araç panel içi sohbeti desteklemiyor — terminalde açın",
	AIDeniedFmt: "izin verilmedi: %s %s (yalnızca dosya düzenleme ve syncftp komutlarına izin var)",

	// Tray menü
	TrayProjects: "Projeler",
	TrayJobs:     "İşler",

	// Tray menü (owner-draw)
	TrayStop:        "durdur",
	TrayTunnelOpen:  "Webhook tünelini aç",
	TrayTunnelClose: "Webhook tünelini kapat",
	TrayTunnelCopy:  "Tünel adresini kopyala",
	TrayTunnelNone:  "Tünel: cloudflared kurulu değil",
	TrayHelp:        "Yardım",

	// Tray tünel bildirimi
	TrayTunnelStarting:  "Webhook tüneli açılıyor…",
	TrayTunnelOpened:    "Webhook tüneli açıldı",
	TrayTunnelOpenedFmt: "%s\nAdres panoya kopyalandı.",
	TrayTunnelFailed:    "Webhook tüneli açılamadı",
	TrayTunnelClosed:    "Webhook tüneli kapatıldı",

	// Tray tünel durumu
	TrayTunnelOnShort:       "tünel açık",
	TrayTunnelStartingShort: "tünel açılıyor…",

	// Global ayarlar / init / push
	InitExistsFmt:            "burada zaten bir syncFTP projesi var: %s",
	GlobalConnNotFoundFmt:    "kayıtlı bağlantı bulunamadı: %q (syncftp global conn list)",
	ConnHasProfileFmt:        "%q zaten %q bağlantı profilini kullanıyor",
	ConnExtractedFmt:         "%q sunucusunun bilgileri %q bağlantı profiline taşındı\n",
	ConnExtractReusedFmt:     "%q sunucusu aynı hesabı tutan %q profiline bağlandı\n",
	FreezeByParentFmt:        "%s, üst klasörü %s freeze'li olduğu için freeze'li — önce klasörü çözün",
	NoActiveServer:           "aktif sunucu yok",
	SyncFrozenSkippedFmt:     "  ❄ %d frozen (atlandı)\n",
	SyncFailedListReadFmt:    "failed listesi okunamadı: %w",
	SyncCalibrateAbortFmt:    "  ✗ Calibrate başarısız, bu sunucunun sync'i iptal edildi (aksi halde tüm dosyalar yeniden yüklenirdi): %v\n",
	SyncCleanupIncludeKept:   "  ⚠ Include listesindeki yollar diskte yok ama liste boşalırsa her şey yüklenir — liste olduğu gibi bırakıldı\n",
	RollbackStoppedFmt:       "  ! %s tamamen geri alınamadı — daha eski release'lere dokunulmadı (tekrar deneyin)\n",
	RollbackKeptNewFmt:       "  · %d yeni eklenen dosya sunucuda bırakıldı (silmek için: CLI --delete-new, panelde onay penceresindeki \"sonradan eklenen dosyaları da sil\" kutusu)\n",
	RollbackWillDeleteFmt:    "\n%d yeni eklenen dosya sunucudan SİLİNECEK:\n  %s\n",
	RollbackWillKeepFmt:      "\n%d yeni eklenen dosya sunucuda bırakılacak (silmek için --delete-new)\n",
	ReleaseFileNotFoundFmt:   "release'te böyle bir dosya yok: %s",
	ReleaseDiffBinary:        "  (binary dosya — fark gösterilmez)",
	ReleaseDeletedFmt:        "  ✓ %d release silindi\n",
	ReleasePurgedFmt:         "  ✓ %d release'in yedeği silindi — geçmişte durur, geri alınamaz\n",
	ReleaseBusyFmt:           "%s üzerinde şu an bir işlem sürüyor — bitince tekrar dene",
	ReleaseDeleteConfirmFmt:  "%[2]s üzerinde %[1]d release silinsin mi?",
	ReleaseDeleteBodyFull:    "Kayıt ve yedek dosyaları tamamen silinir; bu release'ler geçmişten düşer ve geri alınamaz. Sunucudaki dosyalara dokunulmaz.",
	ReleaseDeleteBodyPurge:   "Yalnızca yedek dosyaları silinir; release'ler geçmişte durur ama geri alınamaz. Sunucudaki dosyalara dokunulmaz.",
	ReleaseDeleteNeedID:      "silinecek release'i ver (ya da --all)",
	ReleaseDiffNoBefore:      "  (eski hal yok — yedek alınamamış ya da silinmiş)",
	ReleaseDiffNoAfter:       "  (yüklenen hal saklanmamış ve yerel dosya o zamandan beri değişmiş)",
	ReleaseDiffTooBigFmt:     "  (fark hesaplanamayacak kadar büyük: %d / %d satır)\n",
	ReleaseDiffSame:          "  (içerik aynı)",
	SyncServersFailedFmt:     "%d / %d sunucuda sync başarısız",
	ProjectNoConfigFmt:       "%s bir syncFTP projesi değil (.syncftp/syncftp.json yok)",
	ProjectNotListedFmt:      "%s klasöründe proje oluşturuldu ama listeye eklenemedi — klasöre erişim izni olmayabilir; paneli yeniden açıp tekrar deneyin",
	ProjectRemovedTrashFmt:   "%s syncFTP'den çıkarıldı — .syncftp klasörü çöp kutusunda (geri yüklersen proje geri gelir)\n",
	ProjectRemovedRenamedFmt: "%s syncFTP'den çıkarıldı — çöp kutusu kullanılamadı, .syncftp yeniden adlandırıldı: %s\n",
	ProjectDeleteConfirmFmt:  "%s syncFTP'den çıkarılsın mı?",
	ProjectDeleteConfirmBody: ".syncftp klasörü (config, state, release yedekleri, loglar) çöp kutusuna taşınır.\nProje dosyalarına ve sunuculara dokunulmaz; anahtarlıktaki şifreler silinmez.",
	GlobalLoadWarnFmt:        "  ⚠ global ayarlar okunamadı: %v\n",
	InitConnManual:           "Elle gir",
	InitConnPickTitle:        "Bağlantı",
	InitConnPickSub:          "Kayıtlı bir bağlantı seçin ya da elle girin",
	InitUsingConnFmt:         "  ✓ Kayıtlı bağlantı: %s (%s@%s:%d)\n",
	GlobalFileFmt:            "  Global ayarlar: %s\n",
	GlobalDefaultsHeader:     "  Yeni proje varsayılanları",
	GlobalSecretWarnFmt:      "\n  ⚠ Anahtarlıktan okunamayan şifreler: %s\n",
	GlobalConnHeaderFmt:      "  Kayıtlı bağlantılar (%d)\n",
	GlobalConnNone:           "    — yok (syncftp global conn add <ad> --host …)",
	GlobalConnAddedFmt:       "  ✓ Bağlantı eklendi: %s\n",
	GlobalConnUpdatedFmt:     "  ✓ Bağlantı güncellendi: %s\n",
	GlobalConnRemovedFmt:     "  ✓ Bağlantı silindi: %s\n",
	GlobalConnImportedFmt:    "  ✓ %d bağlantı aktarıldı (%s)\n",
	GlobalDefaultsFromFmt:    "  ✓ Varsayılanlar %s projesinden alındı\n",
	GlobalTemplateSetFmt:     "  ✓ Varsayılan ignore şablonu: %s\n",
	GlobalDefaultsReset:      "  ✓ Varsayılanlar sıfırlandı",
	PushFailedErrFmt:         "%d dosya yüklenemedi",
}

package lang

import (
	"os"
	"path/filepath"
	"strings"
)

// StringSet holds all user-visible strings for one locale.
type StringSet struct {
	// ── Browser TUI ──────────────────────────────────────────────────────────
	BrowserLoading           string // "Yükleniyor..."
	BrowserSearchPrompt      string // "  Enter=Tümünü Ara  ESC=İptal"
	BrowserSearching         string // "  Aranıyor: %q ..."
	BrowserSearchResults     string // "  (tüm sunucu, %d sonuç)  ESC temizle"
	BrowserPickDirTitle      string // "  HEDEF DİZİN SEÇ"
	BrowserPickDirEnter      string // "Enter=Bu Klasörü Seç"
	BrowserHintRight         string // "→ Gir"
	BrowserHintLeft          string // "← Üst"
	BrowserHintQuit          string // "q İptal"
	BrowserMarkedHintDelete  string // "d=Sil"
	BrowserMarkedHintMove    string // "m=Taşı"
	BrowserMarkedHintFreeze  string // "f=❄ Freeze"
	BrowserMarkedHintAll     string // "a=Tümünü kaldır"
	BrowserMarkedHintCancel  string // "ESC=İptal"
	BrowserEscPending        string // "ESC tekrar = çık"
	BrowserSearchClear       string // "  ESC temizle"
	BrowserHintNav           string // "  ↑↓ Gezin"
	BrowserHintEnter         string // "Enter Gir/Seç"
	BrowserHintPreview       string // "→ Önizle"
	BrowserHintSpace         string // "Space İşaretle"
	BrowserHintSearch        string // "/ Ara"
	BrowserHintQuitKey       string // "q Çık"
	BrowserHintFreeze        string // "f=❄ Freeze"
	BrowserHintTree          string // "t=Tree"
	BrowserErrPrefix         string // "  Hata: "
	BrowserReconnectHint     string // "r = Yeniden Bağlan  |  q = Çık"
	BrowserSelectThisPrefix  string // "[ Bu dizini seç: "
	BrowserSelectThisSuffix  string // " ]"
	BrowserCountMany         string // "çok boyutlu"  (folder with >999 items)
	BrowserFooterFmt         string // "  %d klasör, %d dosya  ·  %d/%d"
	BrowserSearchMatchFmt    string // "  ara:%q %d eşleşme"
	BrowserPreviewLoading    string // "  Yükleniyor..."
	BrowserPreviewBack       string // "  ESC/q geri"
	BrowserPreviewScrollHint string // "↑↓ kaydır"
	BrowserPreviewPageHint   string // "PgUp/PgDn"
	BrowserPreviewGGHint     string // "g başa  G sona"
	BrowserPreviewEmpty      string // "(boş dosya)"
	BrowserPreviewLineFmt    string // "%d-%d / %d satır"
	BrowserConnLost          string // "⚠ Bağlantı kesildi"
	BrowserConnLostHint      string // "r tuşuna basarak yeniden bağlanabilirsiniz"
	BrowserPreviewFailed     string // "⚠ Önizleme alınamadı"
	BrowserReconnectFailed   string // "yeniden bağlanılamadı: %w"  (error wrap prefix — keep %w)
	BrowserMarkedCount       string // "  ✓ %d işaretli"
	BrowserSearchCancelHint  string // "   ESC=iptal"
	BrowserSearchTimeout     string // "⚠ zaman aşımı (5dk) — %d kısmi sonuç"
	BrowserSearchConnErr     string // "⚠ bağlantı hatası: %s"
	BrowserSearchOpenErr     string // "bağlantı %d açılamadı: %v"
	BrowserSearchDoneFmt     string // "%d sonuç (%.1fs)"
	BrowserSearchOpsHint     string // arama sonuçlarında yapılabilecek işlemler
	BrowserSearchNoResults   string // hiç sonuç bulunamadı

	// ── Picker / Confirm TUI ─────────────────────────────────────────────────
	PickerFilterHint   string // "  Harf yaz → filtrele"
	PickerMatchFmt     string // "  (%d eşleşme)"
	PickerNoMatch      string // "  (eşleşme yok — Backspace ile sil)"
	PickerNavHint      string // "↑↓/jk gezin   Enter seç   [1-9] hızlı   Backspace sil   q iptal"
	MultiPickerNavHint string // "↑↓ gezin   Space işaretle   a hepsini   Enter onayla   q iptal"
	ConfirmYes         string // "Evet"
	ConfirmNo          string // "Hayır, iptal"

	// ── Sync TUI ─────────────────────────────────────────────────────────────
	SyncDryRunTitle      string // "  [DRY RUN] Yüklenecekler:"
	SyncDryRunFmt        string // "  %d dosya — herhangi bir tuşa basın"
	SyncNoChange         string // "  ✓ Değişiklik yok — güncel"
	SyncAnyKey           string // "  Herhangi bir tuşa basın"
	SyncConnecting       string // "bağlanıyor..."
	SyncRetryFmt         string // " (%d deneme)"
	SyncRetryOkFmt       string // " (%d. denemede)"
	SyncDoneFmt          string // "  Tamamlandı: %s yüklendi"
	SyncDoneFailFmt      string // ", %s hata"
	SyncAttemptDetailFmt string // "deneme %d: %v"
	SyncScrollHint       string // kaydırma ipucu (sync bittikten sonra)
	SyncFailedHeader     string // başarısız dosyalar bölüm başlığı

	// ── Shell REPL ───────────────────────────────────────────────────────────
	ShellWelcomeHint      string // "  'help' yazın, çıkmak için 'exit'"
	ShellMultiServer      string // "Birden fazla sunucu var — bağlanmak için: server <ad>"
	ShellServersLabel     string // "Sunucular:"
	ShellCtrlCHint        string // "(Ctrl+C — çıkmak için 'exit')"
	ShellCtrlCExitHint    string // "Çıkmak için Ctrl+C'ye tekrar basın — ya da 'exit' yazın"
	ShellExit             string // "Görüşürüz."
	ShellUnknownCmd       string // "Bilinmeyen komut: %q  (yardım için 'help')"
	ShellConnecting       string // "Bağlanıyor: %s (%s)..."
	ShellConnectErr       string // " hata: %v\n"
	ShellNotConnected     string // "Bağlı değil. 'server <ad>' ile bağlanın."
	ShellNoServers        string // "Aktif sunucu yok."
	ShellServersFmt       string // "Sunucu listesi"
	ShellServersSubtitle  string // "Seçince bağlanır — q ile sadece kapat"
	ShellAlreadyConn      string // "Zaten bağlı: %s\n"
	ShellDisconnected     string // "Bağlantı kesildi: %s\n"
	ShellServerNotFound   string // "Sunucu bulunamadı: %q\n"
	ShellBrowserErr       string // "Browser hatası: %v\n"
	ShellDeleteTitle      string // "%d dosyayı sil"
	ShellDeleteSubtitle   string // "Bu işlem geri alınamaz. FTP sunucusundan kalıcı olarak silinir."
	ShellCancelled        string // "İptal edildi."
	ShellConfirmSure      string // "Emin misiniz?"
	ShellConfirmSureBody  string // "Onaylamak için Evet seçin."
	ShellDeletedFmt       string // "\n%d silindi, %d başarısız\n"
	ShellMovingFmt        string // "\n%d dosya taşınacak. Hedef klasörü seçin (Enter = bu klasörü seç):\n"
	ShellMoveTarget       string // "Hedef: %s\n\n"
	ShellMovedFmt         string // "\n%d taşındı, %d başarısız\n"
	ShellDirNotFound      string // "Dizin bulunamadı: %s\n"
	ShellDownloading      string // "İndiriliyor: %s → %s\n"
	ShellDownloaded       string // "✓ İndirildi (%s)\n"
	ShellDownloadedBare   string // "✓ İndirildi"
	ShellErrFmt           string // "Hata: %v\n"
	ShellDeleteConfTitle  string // "Silme Onayı"
	ShellCancelShort      string // "İptal."
	ShellDeletedShort     string // "✓ Silindi: %s\n"
	ShellPickServerTitle  string // "Sunucu Seç"
	ShellSyncServers      string // "Sync Edilecek Sunucular"
	ShellCalibrateServers string // "Kalibre Edilecek Sunucular"
	ShellSyncPickSub      string // "Space ile işaretle, Enter ile onayla"
	ShellSyncCancelled    string // "İptal."
	ShellFileActionTitle  string // "Dosya İşlemi"
	ShellFileActionView   string // "İçeriği Görüntüle"
	ShellFileActionGet    string // "İndir"
	ShellFileActionDel    string // "Sil"
	ShellFileActionCancel string // "İptal"
	ShellFileTruncFmt     string // "[İlk %d KB — tamamı için: get %s]\n"
	ShellScanning         string // "Taranıyor..."
	ShellScannedFmt       string // " %d dosya\n"
	ShellConnPoolErr      string // "  Bağlantı hatası: %v\n"
	ShellReleaseFmt       string // "  Release: %s\n"
	ShellStateReadErr     string // "  State okunamadı: %v\n"
	ShellStatusDeleted    string // "  - %s (yerel silinmiş)\n"
	ShellStatusUpToDate   string // "  Güncel"
	ShellStatusNoChange   string // "  %d değişiklik\n"
	ShellStatusStateErr   string // "[%s] state okunamadı: %v\n"

	// ── Tree ─────────────────────────────────────────────────────────────────
	TreeMaxPromptTitle string // "Tree görünümü"
	TreeMaxPromptSub   string // "Klasör başına max dosya sayısı?"
	TreeSkippedFmt     string // "[%d öğe — atlandı, --max ile artırın]"
	TreeErrFmt         string // "[hata: %v]"

	// ── Help text ────────────────────────────────────────────────────────────
	ShellHelp string

	// ── cmd_sync ─────────────────────────────────────────────────────────────
	SyncCancelled      string // "İptal."
	SyncDryRunNote     string // "[DRY RUN] Hiçbir şey yüklenmeyecek"
	SyncWhitelistFmt   string // "Whitelist (%d yol): yalnızca bu yollar sync edilecek\n"
	SyncExcludeFmt     string // "Exclude (%d yol): bu yollar bu sync'ten hariç tutulacak\n"
	SyncServerErrFmt   string // "  HATA: %v\n"
	SyncNoChange2      string // "  Değişiklik yok"
	SyncDeletedHeader  string // "  ! SİLİNEN dosyalar (FTP'de bırakıldı):\n"
	SyncProcessingFmt  string // "  %d dosya işlenecek"
	SyncProtectedFmt   string // " (%d korunuyor)"
	SyncProtectedLabel string // "    KORUNUYOR  %s\n"
	SyncUploadLabel    string // "    YÜKLENECEK %s\n"
	SyncPoolFmt        string // "  Bağlantı havuzu: %d / Retry: %d\n"
	SyncDoneFullFmt    string // "  Tamamlandı: %d yüklendi, %d korundu, %d hata\n"
	SyncFailedSavedFmt string // "  ! %d başarısız dosya .syncftp/failed/%s.json'a kaydedildi — tekrar denemek için: syncftp sync --retry-failed\n"
	SyncFailedSaveErr  string // "  ! Failed listesi kaydedilemedi: %v\n"
	SyncStateErr       string // "  ! State kaydedilemedi: %v\n"
	SyncReleaseErr     string // "  ! Release oluşturulamadı: %v\n"
	SyncReleaseFmt     string // "  Release: %s\n"
	SyncFullFlag       string // "  Tam sync (--full): tüm dosyalar yüklenecek"
	SyncRetryNoFiles   string // "  Yeniden denenecek başarısız dosya yok"
	SyncRetryModeFmt   string // "  Retry modu: %d başarısız dosya (%s)\n"
	SyncRetrySkipFmt   string // "  ! %s artık local'de yok, atlanıyor\n"
	SyncAttemptsFmt    string // "    ✗ %s (%d deneme): %v\n"
	SyncAttemptOkFmt   string // "    ✓ %s (%d. denemede başarılı)\n"
	SyncUploadOkFmt    string // "    ✓ %s\n"
	SyncUploadErrFmt   string // "    ✗ %s: %v\n"

	// ── resync ───────────────────────────────────────────────────────────────
	ResyncScanning       string // "  Sunucuya bağlanılıyor..."
	ResyncConnected      string // " bağlandı\n"
	ResyncListing        string // "  Uzak dosyalar listeleniyor...\n"
	ResyncLocalFmt       string // "  Yerel: %d dosya tarandı\n"
	ResyncConnErr        string // "  ! Sunucuya bağlanılamadı (%v) — resync atlandı\n"
	ResyncListErr        string // "  ! Uzak dizin okunamadı (%v) — resync atlandı\n"
	ResyncFoundFmt       string // "  Uzak:   %d dosya bulundu\n"
	ResyncComparingFmt   string // "  Karşılaştırılıyor: %d / %d\n"
	ResyncMatchedFmt     string // "  Eşleşti: %d (boyut OK)  |  Farklı/eksik: %d\n"
	ResyncDoneFmt        string // "  [%s] calibrate tamamlandı\n"
	ResyncAutoMsg        string // "  İlk çalıştırma: calibrate yapılıyor...\n"
	ResyncNoServers      string // "Eşleşen sunucu bulunamadı.\n"
	ResyncIgnoreDirsFmt  string // "  Ignore: %d klasör atlandı → %s\n"
	ResyncIgnoreFilesFmt string // "  Ignore: %d dosya atlandı\n"
	ResyncFilteredFmt    string // "  Filtre (include/exclude): %d dosya kapsam dışı\n"
	ResyncFrozenDiffFmt  string // "  ❄ %d frozen dosya farklı/eksik (sync sırasında atlanacak)\n"
	ResyncListTUIFmt     string // "  Listeleniyor: %d dosya   (→ dosyaları göster | ← gizle | ctrl+c iptal)"
	ResyncListCancelled  string // "  ! Listeleme iptal edildi — calibrate atlandı\n"
	ResyncLogFilesHeader string // "\n── Taranan uzak dosyalar (%d) ──\n"

	// ── cmd_milestone ────────────────────────────────────────────────────────
	MilestoneSetFmt     string // "  ✓ Milestone [%s]: %s\n"
	MilestoneShowFmt    string // "  %s: %s\n"
	MilestoneNoneFmt    string // "  %s: (milestone yok)\n"
	MilestoneClearedFmt string // "  ✓ Milestone silindi: %s\n"
	MilestoneNotSetFmt  string // "  ⚠ %s: milestone tanımlı değil — önce 'milestone set' çalıştır\n"
	MilestoneDateErrFmt string // "  ✗ Tarih anlaşılamadı: %q (format: 2006-01-02 [15:04[:05]])\n"
	MilestoneSinceFmt   string // "  ⏱ Milestone: %s sonrası değişen dosyalar sync edilecek\n"
	MilestoneFilterFmt  string // "  ⏱ Milestone filtresi (%s): %d dosya tarih öncesi — atlandı\n"
	MilestoneShellNoSrv string // "  Önce bir sunucuya bağlan (server <ad>) — milestone bağlı sunucuya işler.\n"
	MilestoneUsage      string // "  Kullanım: milestone [set [tarih] | sync [--dry-run] | clear | show]\n"

	// ── cmd_status ───────────────────────────────────────────────────────────
	StatusWhitelistFmt     string // "Whitelist (%d yol): yalnızca bu yollar gösterilecek\n"
	StatusExcludeFmt       string // "Exclude (%d yol): bu yollar sonuçtan hariç tutulacak\n"
	StatusProjectFmt       string // "Proje : %s\n"
	StatusDirFmt           string // "Dizin : %s\n"
	StatusFileFmt          string // "Dosya : %d adet\n"
	StatusStateErr         string // "[%s] State yüklenemedi: %v\n\n"
	StatusNoFirstSync      string // "  Henüz ilk sync yapılmadı"
	StatusUpToDate         string // "  Değişiklik yok — sunucu güncel"
	StatusFilteredUpToDate string // "  Belirtilen filtre kapsamında değişiklik yok"
	StatusNewHeader        string // "  + YENİ"
	StatusChangedHeader    string // "  ~ DEĞİŞEN"
	StatusDeletedHeader    string // "  - SİLİNEN (FTP'den silinmez, sadece bilgi)"

	// ── cmd_push ─────────────────────────────────────────────────────────────
	PushSourceFmt     string // "Kaynak  : %s\n"
	PushTargetFmt     string // "Hedef   : %s:%d%s\n"
	PushScanning      string // "Taranıyor..."
	PushScannedFmt    string // " %d dosya\n\n"
	PushFirstPush     string // "İlk push"
	PushFullPush      string // "Tam push (--full)"
	PushFirstFmt      string // "  %s — tüm dosyalar yüklenecek\n"
	PushDeletedHeader string // "  ! SİLİNEN dosyalar (FTP'de bırakıldı):\n"
	PushNoChange      string // "  Değişiklik yok — hedef güncel"
	PushProcessingFmt string // "  %d dosya işlenecek\n"
	PushDryRunHeader  string // "\n[DRY RUN] Yüklenecekler:"
	PushConnFmt       string // "  Bağlantı: %d / Retry: %d\n"
	PushAttemptsFmt   string // "    ✗ %s (%d deneme): %v\n"
	PushAttemptOkFmt  string // "    ✓ %s (%d. denemede)\n"
	PushUploadOkFmt   string // "    ✓ %s\n"
	PushUploadErrFmt  string // "    ✗ %s: %v\n"
	PushDoneFmt       string // "\n  Tamamlandı: %d yüklendi, %d hata\n"
	PushFailedHint    string // "  ! Başarısız dosyalar kaydedildi — tekrar: syncftp push %s --server %s --full\n"
	PushStateErr      string // "  ! State kaydedilemedi: %v\n"
	PushReleaseFmt    string // "  Release: %s\n"

	// ── cmd_remote ───────────────────────────────────────────────────────────
	RemoteServerFmt      string // "Sunucu : %s (%s)\n"
	RemoteDirFmt         string // "Dizin  : %s\n\n"
	RemoteConnecting     string // "Bağlanıyor: %s (%s)...\n"
	RemoteDownloading    string // "İndiriliyor: %s → %s\n"
	RemoteDownloaded     string // "✓ İndirildi (%s)\n"
	RemoteDownloadedBare string // "✓ İndirildi"
	RemoteDeleteLabel    string // "Silinecek %s: %s\n"
	RemoteDeleteConfirm  string // "Silmek istediğinizden emin misiniz? [y/N]: "
	RemoteDeleteCancel   string // "İptal edildi."
	RemoteDeletedFmt     string // "✓ Silindi: %s\n"
	RemoteCatTruncFmt    string // "\n[İlk %d KB gösterildi — tamamını indirmek için: syncftp remote get %s]\n"
	RemotePickDirErr     string // "dizin listelenemedi: %w"
	RemotePickUpDir      string // "  [0] .. (üst dizin)"
	RemotePickPromptFmt  string // "  Seçim [0-%d, q=iptal]: "
	RemotePickInvalid    string // "  Geçersiz seçim: %q\n"
	RemotePickSelected   string // "  Seçildi: %s\n\n"
	RemoteFileLabel      string // "dosya"
	RemoteDirLabel       string // "dizin"
	RemoteDirRecLabel    string // "dizin (içeriğiyle)"
	RemoteListErr        string // "%s  ! listelenemedi: %v\n"
	RemoteNoServerSel    string // "sunucu seçilmedi"

	// ── lang command ─────────────────────────────────────────────────────────
	LangCurrentFmt  string // "Language: %s\n"
	LangSwitchedFmt string // "Language set to: %s\n"
	LangSavedFmt    string // "Saved to .syncftp/lang\n"
	LangInvalid     string // "Unknown language: %q — use 'en' or 'tr'\n"
	LangAlreadyFmt  string // "Already using %s\n"

	// ── failsafe cleanup ─────────────────────────────────────────────────────
	SyncCleanupRemovedFmt string // "  ⚠ %q no longer exists locally, removed from list\n"
	SyncConfigSaveErr     string // "  ! Config could not be saved: %v\n"

	// ── webhook ──────────────────────────────────────────────────────────────
	SyncWebhookSentFmt string // "  ✓ Webhook: %s (HTTP %d)\n"
	SyncWebhookErrFmt  string // "  ! Webhook hatası: %v\n"

	// ── block filter ─────────────────────────────────────────────────────────
	SyncBlockedFmt string // "  ✗ %d blocked (uzantı/dosya filtresi)\n"

	// ── minify / obfuscate ───────────────────────────────────────────────────
	SyncMinifiedFmt   string // "  ✓ %d dosya minified\n"
	SyncObfuscatedFmt string // "  ✓ %d dosya obfuscated\n"
	SyncMinifyErrFmt  string // "  ! Minify hatası (%s): %v — orijinal yükleniyor\n"
	SyncObfuscErrFmt  string // "  ! Obfuscate hatası (%s): %v — minified yükleniyor\n"
	SyncNoTerserWarn  string // "  ⚠ terser PATH'te bulunamadı — obfuscation atlandı\n"

	// ── sync summary TUI ─────────────────────────────────────────────────────
	SyncSummaryTitle   string // "  Sync Özeti"
	SyncSummaryNavHint string // "  ↑↓ gezin  |  → = detay  |  q = çık"
	SyncSummaryOk      string // "✓"
	SyncSummaryErrFmt  string // "HATA: %v"

	// ── status TUI ───────────────────────────────────────────────────────────
	StatusDetailChangesCountFmt string // "  %d changes"  (detail view title)
	StatusSyncConfirm           string // "  Start sync?"
	StatusSyncHint              string // "  Enter/s = Yes  |  ESC = Cancel"
	StatusSearchHint            string // "  Backspace = clear  |  Esc = close"
	StatusDetailNavHint         string // "  ↑↓/g/G scroll  |  / = search  |%s  ←/Esc = back  |  q = quit"
	StatusDetailSyncPart        string // "  s = Sync  |"
	StatusNoMatch               string // "  (no matches)"
	StatusUpToDateShort         string // "  ✓ up to date"
	StatusScrollFmt             string // "  %d/%d"
	StatusResultsFmt            string // "  %d results"
	StatusListTitleFmt          string // "  Status — %s"
	StatusListNavHint           string // "  ↑↓ navigate  |  → = details  |  q = quit"
	StatusOkShort               string // "✓ up to date"
	StatusChangesListFmt        string // "%d changes"
	StatusFrozenFmt             string // "❄ %d frozen"

	// ── watch ────────────────────────────────────────────────────────────────
	WatchStartFmt  string // "  ✓ [%s] izleniyor: %s\n"
	WatchReady     string // "Değişiklikler bekleniyor. Durdurmak için Ctrl+C.\n"
	WatchChangeFmt string // "  ⟳ [%s] değişiklik tespit edildi, sync başlatılıyor...\n"
	WatchErrFmt    string // "  ! Watch hatası: %v\n"
	WatchNoServers string // "İzlenecek aktif sunucu bulunamadı.\n"

	// ── cmd_init ─────────────────────────────────────────────────────────────
	InitWizardTitle   string // "=== syncFTP Kurulum Sihirbazı ==="
	InitProjectName   string // "Proje adı"
	InitLocalDir      string // "Yerel dizin"
	InitFTPHeader     string // "FTP Sunucu bilgileri:"
	InitServerName    string // "Sunucu adı"
	InitHost          string // "FTP host"
	InitPort          string // "Port"
	InitUser          string // "Kullanıcı adı"
	InitPassword      string // "Şifre"
	InitRemotePath    string // "Uzak dizin"
	InitCreated       string // "✓ syncftp.json oluşturuldu (izinler: 600)"
	InitReadyFmt      string // "Hazır! 'syncftp sync' komutu ile %q projesini senkronize edebilirsiniz.\n"
	InitIgnoreExists  string // "  (syncFTP girdileri zaten %s içinde)\n"
	InitIgnoreAdded   string // "✓ syncftp.json, syncftp.exe → %s'e eklendi\n"
	InitIgnoreCreated string // "✓ syncftp.ignore oluşturuldu (syncftp.json, syncftp.exe eklendi)"

	// ── Config TUI — server manager ──────────────────────────────────────────
	CfgSrvTitle          string // "⚙  Server Ayarları"
	CfgSrvNavHint        string // "  ↑↓ gezin  |  Enter/e = düzenle  |  Space = aç/kapat  |  d = sil  |  n = yeni  |  q = çık"
	CfgDeleteConfirmFmt  string // "[%s] silinsin mi?  y = evet, diğer = iptal"
	CfgGlobalLabel       string // "⚙ Global Ayarlar"
	CfgGlobalHint        string // "  (protect, include, exclude, ignore_files)"
	CfgConnProfilesLabel string // "🔗 Bağlantı Profilleri"
	CfgConnCountFmt      string // "(%d profil)"
	CfgNewServerLabel    string // "+ Yeni server ekle"
	CfgSrvCountFmt       string // "  %d server  |  %d bağlantı profili"
	CfgSaveErr           string // "Kayıt hatası: %v\n"
	CfgUpdatedFmt        string // "✓ [%s] güncellendi\n"
	CfgAddedFmt          string // "✓ [%s] eklendi\n"
	CfgDeletedFmt        string // "✓ [%s] silindi\n"
	CfgEnabledLabel      string // "aktif"
	CfgDisabledLabel     string // "devre dışı"

	// ── Config TUI — server edit ─────────────────────────────────────────────
	CfgSrvEditTitle   string // "Server Düzenle"
	CfgSrvNewTitle    string // "Yeni Server"
	CfgSrvEditNavHint string // "  ↑↓ gezin  |  Enter/b = düzenle/seç  |  Space = bool toggle  |  s = kaydet  |  q = iptal"
	CfgSrvSaveHint    string // "  s = kaydet  |  q/Esc = iptal"
	CfgSrvManual      string // "(manuel)"
	CfgSrvProjectDef  string // "(project default)"
	CfgSrvConnChange  string // "[b=değiştir]"
	CfgSrvConnSelect  string // "[b=profil seç]"
	CfgSrvBrowseHint  string // "[b=gözat]"

	// ── Config TUI — server field labels ─────────────────────────────────────
	CfgFldName        string // "Ad"
	CfgFldConnection  string // "Connection"
	CfgFldHost        string // "Host"
	CfgFldPort        string // "Port"
	CfgFldUser        string // "Kullanıcı"
	CfgFldPassword    string // "Şifre"
	CfgFldRemotePath  string // "Uzak dizin"
	CfgFldLocalPath   string // "Yerel dizin"
	CfgFldEnabled     string // "Aktif"
	CfgFldPassive     string // "Passive mode"
	CfgFldDisableEPSV string // "EPSV devre dışı"
	CfgFldNAT         string // "NAT workaround"
	CfgFldMaxConn     string // "Max bağlantı"
	CfgFldMaxRetry    string // "Max retry"
	CfgFldInclude     string // "Include"
	CfgFldExclude     string // "Exclude"
	CfgFldProtect     string // "Protect"
	CfgFldWebhook     string // "Webhook URL"
	CfgFldMinify      string // "Minify CSS/JS"
	CfgFldObfuscate   string // "Obfuscate JS"
	CfgFldBlockExts   string // "Block extensions"

	// ── Config TUI — global settings ─────────────────────────────────────────
	CfgGlobalEditTitle    string // "⚙  Global Ayarlar"
	CfgGlobalEditNavHint  string // nav hint
	CfgGlobalDefaultEmpty string // "(boş = çalışma dizini)"
	CfgGlobalIgnoreDefVal string // "(varsayılan: .gitignore + syncftp.ignore)"
	CfgGlobalIgnoreNone   string // "(hiçbiri — ignore kullanılmaz)"
	CfgGlobalIgnoreHint   string // long hint at bottom

	// ── Config TUI — global field labels ─────────────────────────────────────
	CfgGFldDefaultPath string // "Varsayılan dizin"
	CfgGFldProtect     string // "Protect"
	CfgGFldInclude     string // "Include (global)"
	CfgGFldExclude     string // "Exclude (global)"
	CfgGFldIgnore      string // "Ignore files"
	CfgGFldWebhook     string // "Webhook URL (global)"
	CfgGFldBlockExts   string // "Block extensions (global)"

	CfgValueEmpty string // "(boş)" — boş alan göstergesi

	// ── Config TUI — connection manager ──────────────────────────────────────
	CfgConnMgrTitle   string // "🔗  Bağlantı Profilleri" (tam ekran başlığı)
	CfgConnMgrNavHint string // nav hint
	CfgNewConnLabel   string // "+ Yeni profil ekle"
	CfgConnTotalFmt   string // "  %d profil"

	// ── Config TUI — connection edit ─────────────────────────────────────────
	CfgConnEditTitle   string // "🔗  Profil Düzenle"
	CfgConnNewTitle    string // "🔗  Yeni Profil"
	CfgConnEditNavHint string // nav hint

	// ── Config TUI — connection picker ───────────────────────────────────────
	CfgConnPickerTitle      string // "Bağlantı Profili Seç"
	CfgConnPickerSub        string // "Bağlantı bilgileri buradan alınacak"
	CfgConnPickerCurrentFmt string // "Mevcut: %s"
	CfgConnPickerManual     string // "(manuel)"
	CfgConnPickerManualDesc string // "Bilgileri elle gir"

	// ── Config TUI — local browser ───────────────────────────────────────────
	CfgLocalNavHint      string // normal mod nav hint
	CfgLocalDirPickHint  string // dirPick mod nav hint
	CfgLocalCustomHint   string // özel yol giriş ipucu
	CfgLocalEmpty        string // "  (boş klasör)"
	CfgLocalSelectFmt    string // "  Şu an: %s  |  Space veya s = bu dizini seç"
	CfgLocalPathLabel    string // "  Yol: "
	CfgLocalCustomSave   string // "  Enter=ekle  |  Esc=iptal"
	CfgLocalMarkedFmt    string // "  ✓ %d işaretli  +%d özel yol  |  n=özel yol ekle"
	CfgLocalMarkedSimple string // "  ✓ %d işaretli  |  n=özel yol ekle"

	// ── Config TUI — ignore picker ────────────────────────────────────────────
	CfgIgnoreTitle   string // "⚙  Ignore dosyaları"
	CfgIgnoreNavHint string // "  Space = seç/kaldır  |  s = kaydet  |  q = iptal"
	CfgIgnoreNote    string // "  Hiçbirini seçmezsen ignore kullanılmaz."

	// ── Sync log + rapor ──────────────────────────────────────────────────────
	SyncLogSavedFmt string // "  Log: %s\n"
	SyncLogHint     string // "  ·  s = log kaydet"
	SyncReportFmt   string // "  ↑ %s · %s" (boyut, süre)

	// ── Hook'lar (pre_sync / post_sync) ───────────────────────────────────────
	SyncHookRunFmt   string // "  ⚙ hook: %s\n"
	SyncHookAbortFmt string // "  ✗ pre_sync hook başarısız, sync iptal: %v\n"
	SyncHookErrFmt   string // "  ! post_sync hook hatası: %v\n"

	// ── Status — frozen ───────────────────────────────────────────────────────
	StatusProtectedHeader string
	StatusBlockedHeader   string
	StatusFrozenHeader    string // "  ❄ Frozen (sync'te atlanır)"

	// ── Calibrate — canlı listeleme sayacı ────────────────────────────────────
	ResyncListProgressFmt string // "\r  Listeleniyor: %d dosya"

	// ── Init — ignore şablonu ─────────────────────────────────────────────────
	InitTemplateTitle      string // "Ignore şablonu"
	InitTemplateSub        string // ".syncftp/syncftp.ignore için şablon seç"
	InitTemplateGeneric    string // "Genel"
	InitTemplateNone       string // "Yok"
	InitTemplateNoneDesc   string // "Şablon oluşturma"
	InitTemplateWrittenFmt string // "  ✓ .syncftp/syncftp.ignore oluşturuldu (%s)\n"

	// ── Diff komutu ───────────────────────────────────────────────────────────
	DiffPickA         string // "Birinci sunucu"
	DiffPickB         string // "İkinci sunucu"
	DiffPickSub       string // "Karşılaştırılacak sunucuyu seç"
	DiffNeedTwo       string // "  Karşılaştırma için iki farklı sunucu gerekir.\n"
	DiffTitleFmt      string // "── %s ↔ %s ──\n"
	DiffListingFmt    string // "  %s listeleniyor...\n"
	DiffOnlyFmt       string // "  Sadece %s üzerinde (%d):\n"
	DiffSizeHeaderFmt string // "  Boyut farklı (%d):\n"
	DiffSizeLineFmt   string // "    ~ %s  (%s: %s · %s: %s)\n"
	DiffSameFmt       string // "  ✓ %d dosya aynı\n"

	// ── Upload çekirdeği (yedek / limit) ────────────────────────────
	SyncBandwidthFmt     string // "  Hız sınırı: %d KB/s\n"
	SyncPoolLimitedFmt   string // "  ! %d/%d bağlantı açılabildi (sunucu/host bağlantı limiti)\n"
	SyncBackupErrFmt     string // "  ! Yedek alınamadı (%s): %v — bu dosya geri alınamaz\n"
	SyncBackupSummaryFmt string // "  ↺ Yedek: %d dosyanın eski hali saklandı, %d yeni dosya — geri almak için: syncftp rollback --server %s\n"

	// ── Watch ───────────────────────────────────────────────────────
	WatchStopped string // "\nWatch durduruldu.\n"

	// ── Rollback ────────────────────────────────────────────────────
	RollbackNoReleasesFmt      string // "  %s için geri alınabilir release yok.\n"
	RollbackPickTitleFmt       string // "Rollback — %s"
	RollbackPickSubtitle       string // "Seçilen release VE ondan sonraki tüm release'ler geri alınır"
	RollbackConfirmFmt         string // "%[2]s üzerinde %[1]d release geri alınsın mı?"
	RollbackReleaseDescFmt     string // "%s · %d dosya · %d değişen · %d yeni"
	RollbackTagRolledBack      string // "[geri alındı]"
	RollbackTagNoBackup        string // "[yedek yok]"
	RollbackAlreadyFmt         string // "%s zaten geri alınmış"
	RollbackNoBackupFmt        string // "%s release'inin yedeği yok (yedek kapalı ya da silinmiş) — bu noktanın gerisine dönülemez"
	RollbackNotFoundFmt        string // "release bulunamadı: %s"
	RollbackReleaseFmt         string // "\n  ↺ %s  (%s)\n"
	RollbackRestoredFmt        string // "    ↺ geri yüklendi  %s\n"
	RollbackDeletedFmt         string // "    ✗ silindi        %s\n"
	RollbackFileErrFmt         string // "    ! %s: %v\n"
	RollbackSkippedNoBackupFmt string // "    ! %d dosyanın yedeği yoktu, olduğu gibi bırakıldı\n"
	RollbackDoneFmt            string // "\n  Rollback tamam: %d geri yüklendi, %d silindi, %d hata\n"

	// ── detect-limit / verify / secure / project ────────────────────
	DetectStartFmt      string // "  [%s] bağlantı limiti ölçülüyor (%s://%s)...\n"
	DetectProgressFmt   string // "\r    %d bağlantı açık..."
	DetectErrFmt        string // "    ! bağlanılamadı: %v\n"
	DetectResultFmt     string // "    ✓ sunucu aynı anda %d bağlantı kabul ediyor\n"
	DetectHitMaxFmt     string // "    ✓ %d bağlantı limitsiz açıldı (daha fazlasını denemek için --max'ı artırın)\n"
	DetectApplyFmt      string // "    → %s: max_connections %d → %d\n"
	DetectSaved         string // "  ✓ Config'e kaydedildi.\n"
	VerifyStartFmt      string // "  [%s] sunucu doğrulanıyor (%s://%s)...\n"
	VerifyErrFmt        string // "  [%s] ! %v\n"
	VerifyMissingFmt    string // "    Sunucuda eksik (%d):\n"
	VerifyModifiedFmt   string // "    Sunucuda değişmiş (%d):\n"
	VerifyMoreFmt       string // "      … +%d tane daha\n"
	VerifyCleanFmt      string // "    ✓ Sapma yok — %d dosya eşleşiyor\n"
	VerifySummaryFmt    string // "    %d tamam · %d eksik · %d değişmiş\n"
	VerifyPendingFmt    string // "    (%d yerel değişiklik henüz sync edilmemiş — sapma değil)\n"
	VerifyFixedFmt      string // "    ✓ %d dosya state'ten düşürüldü — sonraki sync yeniden yükler\n"
	VerifyFixHint       string // "    İpucu: --fix ile bu dosyalar sonraki sync'te yeniden yüklenir\n"
	SecureOnFmt         string // "  ✓ Anahtarlık modu açık — %d şifre işletim sistemi anahtarlığına taşındı\n"
	SecureOff           string // "  ✓ Anahtarlık modu kapalı — şifreler tekrar syncftp.json'da\n"
	SecureUnresolvedFmt string // "bazı şifreler anahtarlıktan okunamadı: %s"
	ProjectNotFoundFmt  string // "proje bulunamadı: %s (bkz: syncftp project list)"
	ProjectListEmpty    string // "  Kayıtlı proje yok. Proje dizininde `syncftp project add` çalıştırın.\n"
	ProjectAddedFmt     string // "  ✓ %s → %s\n"
	ProjectRemovedFmt   string // "  ✓ %s kayıttan çıkarıldı\n"
	ProjectRenamedFmt   string // "  ✓ %s → %s\n"

	// ── Config TUI — protokol / tünel / yedek alanları ──────────────
	CfgFldProtocol        string // "Protokol (ftp/ftps/sftp)"
	CfgFldKeyFile         string // "SSH anahtar dosyası"
	CfgFldTLSSkip         string // "FTPS: sertifika atla"
	CfgFldTunnelHost      string // "SSH tünel host"
	CfgFldTunnelPort      string // "SSH tünel port"
	CfgFldTunnelUser      string // "SSH tünel kullanıcı"
	CfgFldTunnelPass      string // "SSH tünel şifre"
	CfgFldTunnelKey       string // "SSH tünel anahtar"
	CfgFldBandwidth       string // "Hız sınırı (KB/s)"
	CfgFldNoBackup        string // "Yedeği kapat (rollback)"
	CfgFldDirectUpload    string // "Doğrudan yükle (tmp+rename yok)"
	CfgGFldBandwidth      string // "Hız sınırı (KB/s, sunucu başı)"
	CfgGFldNoBackup       string // "Yedeği kapat (rollback)"
	CfgGFldBackupKeep     string // "Tutulacak yedek (release)"
	CfgGFldDirectUpload   string // "Doğrudan yükle (tmp+rename yok)"
	CfgGFldKeyring        string // "Şifreler OS anahtarlığında"
	CfgBandwidthGlobal    string // "(global)"
	CfgBandwidthUnlimited string // "sınırsız"

	// ── Shell / GUI ─────────────────────────────────────────────────
	ShellReconnectErrFmt string // "  ! Yeniden bağlanılamadı: %v\n"
	ShellReconnectedFmt  string // "  ✓ %s yeniden bağlandı\n"
	GuiLaunched          string // "  ✓ Yönetim paneli ayrı pencerede açıldı\n"
	GuiStartFmt          string // "syncFTP paneli → %s\n"
	GuiNoBrowser         string // "Edge/Chrome bulunamadı — varsayılan tarayıcıda açılıyor\n"

	// ── GUI backend ─────────────────────────────────────────────────
	GuiBusyFmt           string // "%s üzerinde zaten çalışan bir iş var (%s)"
	GuiPathRequired      string // "klasör yolu gerekli"
	GuiDirNotFoundFmt    string // "klasör bulunamadı: %s"
	LocalDirMissingFmt   string // "yerel klasör yok: %s (%s) — sunucunun yerel dizinini düzeltin"
	GuiIgnoreNoFile      string
	LogsNone             string
	LogsHint             string
	ServerNotFoundFmt    string // "sunucu bulunamadı: %q"
	NoServerSelected     string
	GuiOutsideBaseFmt    string // "%s panelin açıldığı klasörün (%s) dışında"
	GuiValConnName       string // "her bağlantı profilinin bir adı olmalı"
	GuiValDupFmt         string // "ad iki kez kullanılmış: %s"
	GuiValServerName     string // "her sunucunun bir adı olmalı"
	GuiValConnMissingFmt string // "%s: %q bağlantı profili yok"
	GuiValHostFmt        string // "%s: host gerekli (ya da bağlantı profili seçin)"
	GuiValRemoteFmt      string // "%s: uzak dizin gerekli"
	GuiJobSync           string // "Sync"
	GuiJobDryRun         string // "Dry-run"
	GuiJobMilestoneSync  string // "Milestone sync"
	GuiJobCalibrate      string // "Calibrate"
	GuiJobVerify         string // "Doğrula"
	GuiJobRollback       string // "Rollback"
	GuiJobDetect         string // "Bağlantı limiti ölç"
	GuiJobWatch          string // "Watch"
	GuiScannedFmt        string // "  Tarandı: %s → %d dosya\n\n"
	GuiNoMilestoneFmt    string // "%s için milestone yok"
	GuiJobNotCancelable  string // "bu iş iptal edilemez"
	GuiWatchRunning      string // "bu projede watch zaten çalışıyor"

	// ── Tray / CI-CD deploy ─────────────────────────────────────────
	TrayOpen             string // "Paneli aç"
	TrayIdle             string // "Boşta"
	TrayStatusFmt        string // "%d iş · %d watch"
	TrayWatchesHeader    string // "Aktif watch'lar (tıkla → durdur)"
	TrayStopWatchFmt     string // "■  %s  (%s)"
	TrayAutostart        string // "Windows açılışında başlat"
	TrayQuit             string // "syncFTP'den çık"
	TrayHookTunnel       string // "webhook tüneli açık"
	GuiJobDeploy         string // "Deploy"
	GuiJobInstall        string // "cloudflared kurulumu"
	DeployStartFmt       string // "══ deploy %s / %s ══  %s\n"
	DeployGitPull        string // "  git pull --ff-only\n"
	DeployGitPullErrFmt  string // "  ✗ git pull başarısız: %v — deploy iptal\n"
	DeploySrvErrFmt      string // "  ✗ %s: %v\n"
	DeployDone           string // "  ✓ deploy tamamlandı\n"
	DeployQueuedRun      string // "\n  ↻ deploy sırasında yeni push geldi — tekrar çalışıyor\n"
	DeployNoCloudflared  string // "cloudflared bulunamadı — panelden (CI/CD sekmesi) kurun ya da PATH'e ekleyin"
	DeployDownloadingFmt string // "  indiriliyor: %s\n"
	DeployInstalledFmt   string // "  ✓ kuruldu: %s (%s)\n"
	DeployNoTriggers     string // "  Deploy tetikleyicisi yok. Panelden (CI/CD sekmesi) ya da syncftp.json'daki \"deploy\" dizisinden ekleyin.\n"
	DeployListeningFmt   string // "syncFTP webhook dinleyicisi → %s\n"
	DeployTunnelFmt      string // "Public URL (trycloudflare) → %s\n"
	DeployCtrlC          string // "\nWebhook bekleniyor — durdurmak için Ctrl+C.\n"
	WebhookTestFmt       string // "syncFTP test bildirimi (%s)"

	// ── GUI 2.1 ─────────────────────────────────────────────────────
	GuiJobDiff string // "Sunucuları karşılaştır"

	// ── SSH uzak komutlar ───────────────────────────────────────────
	RemoteHookSkipFmt string // "  ! [%s] uzak komutlar atlandı — yalnızca SFTP (SSH) sunucularında çalışır\n"
	RemoteHookRunFmt  string // "  ⚙ ssh $ %s\n"
	ShellExecUsage    string // "  Kullanım: exec <komut>  — bağlı SFTP sunucusunda (remote_path içinde) çalışır\n"
	ShellExecNoSSH    string // "  ! exec için SFTP (SSH) bağlantısı gerekir — bu sunucu FTP kullanıyor\n"
	GuiJobExec        string // "SSH komutu"
	CfgFldRemotePre   string // "Uzak pre_sync (SSH)"
	CfgFldRemotePost  string // "Uzak post_sync (SSH)"

	// ── AI asistan (yalnızca GUI) ───────────────────────────────────
	AIBusy      string // "asistan hâlâ cevap veriyor — bekleyin ya da durdurun"
	AINoChat    string // "bu araç panel içi sohbeti desteklemiyor — terminalde açın"
	AIDeniedFmt string // "izin verilmedi: %s %s (yalnızca dosya düzenleme ve syncftp komutlarına izin var)"

	// ── Tray menü ───────────────────────────────────────────────────
	TrayProjects string // "Projeler"
	TrayJobs     string // "İşler"

	// ── Tray menü (owner-draw) ──────────────────────────────────────
	TrayStop        string // "durdur"
	TrayTunnelOpen  string // "Webhook tünelini aç"
	TrayTunnelClose string // "Webhook tünelini kapat"
	TrayTunnelCopy  string // "Tünel adresini kopyala"
	TrayTunnelNone  string // "Tünel: cloudflared kurulu değil"
	TrayHelp        string // "Yardım"

	// ── Tray tünel bildirimi ────────────────────────────────────────
	TrayTunnelStarting  string // "Webhook tüneli açılıyor…"
	TrayTunnelOpened    string // "Webhook tüneli açıldı"
	TrayTunnelOpenedFmt string // "%s\nAdres panoya kopyalandı."
	TrayTunnelFailed    string // "Webhook tüneli açılamadı"
	TrayTunnelClosed    string // "Webhook tüneli kapatıldı"

	// ── Tray tünel durumu ───────────────────────────────────────────
	TrayTunnelOnShort       string // "tünel açık"
	TrayTunnelStartingShort string // "tünel açılıyor…"

	// ── Global ayarlar / init / push ────────────────────────────────
	InitExistsFmt         string // "burada zaten bir syncFTP projesi var: %s"
	GlobalConnNotFoundFmt string // "kayıtlı bağlantı bulunamadı: %q (syncftp global conn list)"
	GlobalLoadWarnFmt     string // "  ⚠ global ayarlar okunamadı: %v\n"
	InitConnManual        string // "Elle gir"
	InitConnPickTitle     string // "Bağlantı"
	InitConnPickSub       string // "Kayıtlı bir bağlantı seçin ya da elle girin"
	InitUsingConnFmt      string // "  ✓ Kayıtlı bağlantı: %s (%s@%s:%d)\n"
	GlobalFileFmt         string // "  Global ayarlar: %s\n"
	GlobalDefaultsHeader  string // "  Yeni proje varsayılanları"
	GlobalSecretWarnFmt   string // "\n  ⚠ Anahtarlıktan okunamayan şifreler: %s\n"
	GlobalConnHeaderFmt   string // "  Kayıtlı bağlantılar (%d)\n"
	GlobalConnNone        string // "    — yok (syncftp global conn add <ad> --host …)"
	GlobalConnAddedFmt    string // "  ✓ Bağlantı eklendi: %s\n"
	GlobalConnUpdatedFmt  string // "  ✓ Bağlantı güncellendi: %s\n"
	GlobalConnRemovedFmt  string // "  ✓ Bağlantı silindi: %s\n"
	GlobalConnImportedFmt string // "  ✓ %d bağlantı aktarıldı (%s)\n"
	GlobalDefaultsFromFmt string // "  ✓ Varsayılanlar %s projesinden alındı\n"
	GlobalTemplateSetFmt  string // "  ✓ Varsayılan ignore şablonu: %s\n"
	GlobalDefaultsReset   string // "  ✓ Varsayılanlar sıfırlandı"
	PushFailedErrFmt      string // "%d dosya yüklenemedi"
}

// L is the active locale. Defaults to English.
var L = En

var currentLang = "en"

// detectSystemLang returns "tr" or "en" based on the OS locale.
// Used as the default when no saved preference or env override is present.
func detectSystemLang() string {
	// Unix/cross-platform: standard locale env vars
	for _, env := range []string{"LANGUAGE", "LC_ALL", "LC_MESSAGES", "LANG"} {
		if v := os.Getenv(env); v != "" {
			first := strings.SplitN(v, ":", 2)[0] // LANGUAGE may be "tr:en"
			if strings.HasPrefix(strings.ToLower(first), "tr") {
				return "tr"
			}
			return "en"
		}
	}
	// Windows: GetUserDefaultUILanguage (lang_windows.go); diğer platformlarda false
	if systemUILangTurkish() {
		return "tr"
	}
	return "en"
}

// Init reads language preference: .syncftp/lang file, then SYNCFTP_LANG env var.
// Falls back to OS system language when no preference is saved.
// Call once from main() before any output.
func Init(configDir string) {
	l := detectSystemLang()
	if data, err := os.ReadFile(filepath.Join(configDir, ".syncftp", "lang")); err == nil {
		if saved := strings.TrimSpace(string(data)); saved != "" {
			l = saved
		}
	}
	if env := os.Getenv("SYNCFTP_LANG"); env != "" {
		l = env
	}
	Set(l)
}

// Set switches the active language at runtime.
func Set(l string) {
	switch strings.ToLower(l) {
	case "tr":
		L = Tr
		currentLang = "tr"
	default:
		L = En
		currentLang = "en"
	}
}

// Current returns the active language code ("en" or "tr").
func Current() string { return currentLang }

// Save writes the language preference to .syncftp/lang.
func Save(configDir, l string) error {
	dir := filepath.Join(configDir, ".syncftp")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "lang"), []byte(l), 0600)
}

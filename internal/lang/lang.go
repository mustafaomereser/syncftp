package lang

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
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
	PickerFilterHint    string // "  Harf yaz → filtrele"
	PickerMatchFmt      string // "  (%d eşleşme)"
	PickerNoMatch       string // "  (eşleşme yok — Backspace ile sil)"
	PickerNavHint       string // "↑↓/jk gezin   Enter seç   [1-9] hızlı   Backspace sil   q iptal"
	MultiPickerNavHint  string // "↑↓ gezin   Space işaretle   a hepsini   Enter onayla   q iptal"
	ConfirmYes          string // "Evet"
	ConfirmNo           string // "Hayır, iptal"

	// ── Sync TUI ─────────────────────────────────────────────────────────────
	SyncDryRunTitle   string // "  [DRY RUN] Yüklenecekler:"
	SyncDryRunFmt     string // "  %d dosya — herhangi bir tuşa basın"
	SyncNoChange      string // "  ✓ Değişiklik yok — güncel"
	SyncAnyKey        string // "  Herhangi bir tuşa basın"
	SyncConnecting    string // "bağlanıyor..."
	SyncRetryFmt      string // " (%d deneme)"
	SyncRetryOkFmt    string // " (%d. denemede)"
	SyncDoneFmt          string // "  Tamamlandı: %s yüklendi"
	SyncDoneFailFmt      string // ", %s hata"
	SyncAttemptDetailFmt string // "deneme %d: %v"
	SyncScrollHint       string // kaydırma ipucu (sync bittikten sonra)
	SyncFailedHeader     string // başarısız dosyalar bölüm başlığı

	// ── Shell REPL ───────────────────────────────────────────────────────────
	ShellWelcomeHint     string // "  'help' yazın, çıkmak için 'exit'"
	ShellMultiServer     string // "Birden fazla sunucu var — bağlanmak için: server <ad>"
	ShellServersLabel    string // "Sunucular:"
	ShellCtrlCHint       string // "(Ctrl+C — çıkmak için 'exit')"
	ShellCtrlCExitHint   string // "Çıkmak için Ctrl+C'ye tekrar basın — ya da 'exit' yazın"
	ShellExit            string // "Görüşürüz."
	ShellUnknownCmd      string // "Bilinmeyen komut: %q  (yardım için 'help')"
	ShellConnecting      string // "Bağlanıyor: %s (%s)..."
	ShellConnectErr      string // " hata: %v\n"
	ShellNotConnected    string // "Bağlı değil. 'server <ad>' ile bağlanın."
	ShellNoServers       string // "Aktif sunucu yok."
	ShellServersFmt      string // "Sunucu listesi"
	ShellServersSubtitle string // "Seçince bağlanır — q ile sadece kapat"
	ShellAlreadyConn     string // "Zaten bağlı: %s\n"
	ShellDisconnected    string // "Bağlantı kesildi: %s\n"
	ShellServerNotFound  string // "Sunucu bulunamadı: %q\n"
	ShellBrowserErr      string // "Browser hatası: %v\n"
	ShellDeleteTitle     string // "%d dosyayı sil"
	ShellDeleteSubtitle  string // "Bu işlem geri alınamaz. FTP sunucusundan kalıcı olarak silinir."
	ShellCancelled       string // "İptal edildi."
	ShellConfirmSure     string // "Emin misiniz?"
	ShellConfirmSureBody string // "Onaylamak için Evet seçin."
	ShellDeletedFmt      string // "\n%d silindi, %d başarısız\n"
	ShellMovingFmt       string // "\n%d dosya taşınacak. Hedef klasörü seçin (Enter = bu klasörü seç):\n"
	ShellMoveTarget      string // "Hedef: %s\n\n"
	ShellMovedFmt        string // "\n%d taşındı, %d başarısız\n"
	ShellDirNotFound     string // "Dizin bulunamadı: %s\n"
	ShellDownloading     string // "İndiriliyor: %s → %s\n"
	ShellDownloaded      string // "✓ İndirildi (%s)\n"
	ShellDownloadedBare  string // "✓ İndirildi"
	ShellErrFmt          string // "Hata: %v\n"
	ShellDeleteConfTitle string // "Silme Onayı"
	ShellCancelShort     string // "İptal."
	ShellDeletedShort    string // "✓ Silindi: %s\n"
	ShellPickServerTitle string // "Sunucu Seç"
	ShellSyncServers     string // "Sync Edilecek Sunucular"
	ShellSyncPickSub     string // "Space ile işaretle, Enter ile onayla"
	ShellSyncCancelled   string // "İptal."
	ShellFileActionTitle string // "Dosya İşlemi"
	ShellFileActionView  string // "İçeriği Görüntüle"
	ShellFileActionGet   string // "İndir"
	ShellFileActionDel   string // "Sil"
	ShellFileActionCancel string // "İptal"
	ShellFileTruncFmt    string // "[İlk %d KB — tamamı için: get %s]\n"
	ShellScanning        string // "Taranıyor..."
	ShellScannedFmt      string // " %d dosya\n"
	ShellConnPoolErr     string // "  Bağlantı hatası: %v\n"
	ShellReleaseFmt      string // "  Release: %s\n"
	ShellStateReadErr    string // "  State okunamadı: %v\n"
	ShellStatusDeleted   string // "  - %s (yerel silinmiş)\n"
	ShellStatusUpToDate  string // "  Güncel"
	ShellStatusNoChange  string // "  %d değişiklik\n"
	ShellStatusStateErr  string // "[%s] state okunamadı: %v\n"

	// ── Tree ─────────────────────────────────────────────────────────────────
	TreeMaxPromptTitle string // "Tree görünümü"
	TreeMaxPromptSub   string // "Klasör başına max dosya sayısı?"
	TreeSkippedFmt     string // "[%d öğe — atlandı, --max ile artırın]"
	TreeErrFmt         string // "[hata: %v]"

	// ── Help text ────────────────────────────────────────────────────────────
	ShellHelp string

	// ── cmd_sync ─────────────────────────────────────────────────────────────
	SyncCancelled        string // "İptal."
	SyncDryRunNote       string // "[DRY RUN] Hiçbir şey yüklenmeyecek"
	SyncWhitelistFmt     string // "Whitelist (%d yol): yalnızca bu yollar sync edilecek\n"
	SyncExcludeFmt       string // "Exclude (%d yol): bu yollar bu sync'ten hariç tutulacak\n"
	SyncServerErrFmt     string // "  HATA: %v\n"
	SyncNoChange2        string // "  Değişiklik yok"
	SyncDeletedHeader    string // "  ! SİLİNEN dosyalar (FTP'de bırakıldı):\n"
	SyncProcessingFmt    string // "  %d dosya işlenecek"
	SyncProtectedFmt     string // " (%d korunuyor)"
	SyncProtectedLabel   string // "    KORUNUYOR  %s\n"
	SyncUploadLabel      string // "    YÜKLENECEK %s\n"
	SyncPoolFmt          string // "  Bağlantı havuzu: %d / Retry: %d\n"
	SyncDoneFullFmt      string // "  Tamamlandı: %d yüklendi, %d korundu, %d hata\n"
	SyncFailedSavedFmt   string // "  ! %d başarısız dosya .syncftp/failed/%s.json'a kaydedildi — tekrar denemek için: syncftp sync --retry-failed\n"
	SyncFailedSaveErr    string // "  ! Failed listesi kaydedilemedi: %v\n"
	SyncStateErr         string // "  ! State kaydedilemedi: %v\n"
	SyncReleaseErr       string // "  ! Release oluşturulamadı: %v\n"
	SyncReleaseFmt       string // "  Release: %s\n"
	SyncFullFlag         string // "  Tam sync (--full): tüm dosyalar yüklenecek"
	SyncRetryNoFiles     string // "  Yeniden denenecek başarısız dosya yok"
	SyncRetryModeFmt     string // "  Retry modu: %d başarısız dosya (%s)\n"
	SyncRetrySkipFmt     string // "  ! %s artık local'de yok, atlanıyor\n"
	SyncAttemptsFmt      string // "    ✗ %s (%d deneme): %v\n"
	SyncAttemptOkFmt     string // "    ✓ %s (%d. denemede başarılı)\n"
	SyncUploadOkFmt      string // "    ✓ %s\n"
	SyncUploadErrFmt     string // "    ✗ %s: %v\n"

	// ── resync ───────────────────────────────────────────────────────────────
	ResyncScanning    string // "  Resync: sunucu taranıyor...\n"
	ResyncConnErr     string // "  ! Sunucuya bağlanılamadı (%v) — resync atlandı\n"
	ResyncListErr     string // "  ! Uzak dizin okunamadı (%v) — resync atlandı\n"
	ResyncFoundFmt    string // "  Sunucuda %d dosya bulundu\n"
	ResyncMatchedFmt  string // "  %d eşleşti (boyut OK), %d farklı/eksik\n"
	ResyncDoneFmt     string // "  [%s] resync tamamlandı\n"
	ResyncAutoMsg     string // "  İlk çalıştırma: resync yapılıyor (sunucu karşılaştırması)...\n"
	ResyncNoServers   string // "Eşleşen sunucu bulunamadı.\n"

	// ── cmd_status ───────────────────────────────────────────────────────────
	StatusWhitelistFmt    string // "Whitelist (%d yol): yalnızca bu yollar gösterilecek\n"
	StatusExcludeFmt      string // "Exclude (%d yol): bu yollar sonuçtan hariç tutulacak\n"
	StatusProjectFmt      string // "Proje : %s\n"
	StatusDirFmt          string // "Dizin : %s\n"
	StatusFileFmt         string // "Dosya : %d adet\n"
	StatusStateErr        string // "[%s] State yüklenemedi: %v\n\n"
	StatusNoFirstSync     string // "  Henüz ilk sync yapılmadı"
	StatusUpToDate        string // "  Değişiklik yok — sunucu güncel"
	StatusFilteredUpToDate string // "  Belirtilen filtre kapsamında değişiklik yok"
	StatusNewHeader       string // "  + YENİ"
	StatusChangedHeader   string // "  ~ DEĞİŞEN"
	StatusDeletedHeader   string // "  - SİLİNEN (FTP'den silinmez, sadece bilgi)"

	// ── cmd_push ─────────────────────────────────────────────────────────────
	PushSourceFmt      string // "Kaynak  : %s\n"
	PushTargetFmt      string // "Hedef   : %s:%d%s\n"
	PushScanning       string // "Taranıyor..."
	PushScannedFmt     string // " %d dosya\n\n"
	PushFirstPush      string // "İlk push"
	PushFullPush       string // "Tam push (--full)"
	PushFirstFmt       string // "  %s — tüm dosyalar yüklenecek\n"
	PushDeletedHeader  string // "  ! SİLİNEN dosyalar (FTP'de bırakıldı):\n"
	PushNoChange       string // "  Değişiklik yok — hedef güncel"
	PushProcessingFmt  string // "  %d dosya işlenecek\n"
	PushDryRunHeader   string // "\n[DRY RUN] Yüklenecekler:"
	PushConnFmt        string // "  Bağlantı: %d / Retry: %d\n"
	PushAttemptsFmt    string // "    ✗ %s (%d deneme): %v\n"
	PushAttemptOkFmt   string // "    ✓ %s (%d. denemede)\n"
	PushUploadOkFmt    string // "    ✓ %s\n"
	PushUploadErrFmt   string // "    ✗ %s: %v\n"
	PushDoneFmt        string // "\n  Tamamlandı: %d yüklendi, %d hata\n"
	PushFailedHint     string // "  ! Başarısız dosyalar kaydedildi — tekrar: syncftp push %s --server %s --full\n"
	PushStateErr       string // "  ! State kaydedilemedi: %v\n"
	PushReleaseFmt     string // "  Release: %s\n"

	// ── cmd_remote ───────────────────────────────────────────────────────────
	RemoteServerFmt     string // "Sunucu : %s (%s)\n"
	RemoteDirFmt        string // "Dizin  : %s\n\n"
	RemoteConnecting    string // "Bağlanıyor: %s (%s)...\n"
	RemoteDownloading   string // "İndiriliyor: %s → %s\n"
	RemoteDownloaded    string // "✓ İndirildi (%s)\n"
	RemoteDownloadedBare string // "✓ İndirildi"
	RemoteDeleteLabel   string // "Silinecek %s: %s\n"
	RemoteDeleteConfirm string // "Silmek istediğinizden emin misiniz? [y/N]: "
	RemoteDeleteCancel  string // "İptal edildi."
	RemoteDeletedFmt    string // "✓ Silindi: %s\n"
	RemoteCatTruncFmt   string // "\n[İlk %d KB gösterildi — tamamını indirmek için: syncftp remote get %s]\n"
	RemotePickDirErr    string // "dizin listelenemedi: %w"
	RemotePickUpDir     string // "  [0] .. (üst dizin)"
	RemotePickPromptFmt string // "  Seçim [0-%d, q=iptal]: "
	RemotePickInvalid   string // "  Geçersiz seçim: %q\n"
	RemotePickSelected  string // "  Seçildi: %s\n\n"
	RemoteFileLabel     string // "dosya"
	RemoteDirLabel      string // "dizin"
	RemoteDirRecLabel   string // "dizin (içeriğiyle)"
	RemoteListErr       string // "%s  ! listelenemedi: %v\n"
	RemoteNoServerSel   string // "sunucu seçilmedi"

	// ── lang command ─────────────────────────────────────────────────────────
	LangCurrentFmt  string // "Language: %s\n"
	LangSwitchedFmt string // "Language set to: %s\n"
	LangSavedFmt    string // "Saved to .syncftp/lang\n"
	LangInvalid     string // "Unknown language: %q — use 'en' or 'tr'\n"
	LangAlreadyFmt  string // "Already using %s\n"

	// ── failsafe cleanup ─────────────────────────────────────────────────────
	SyncCleanupRemovedFmt string // "  ⚠ %q no longer exists locally, removed from list\n"
	SyncConfigSaveErr     string // "  ! Config could not be saved: %v\n"

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

	// ── cmd_init ─────────────────────────────────────────────────────────────
	InitWizardTitle     string // "=== syncFTP Kurulum Sihirbazı ==="
	InitProjectName     string // "Proje adı"
	InitLocalDir        string // "Yerel dizin"
	InitFTPHeader       string // "FTP Sunucu bilgileri:"
	InitServerName      string // "Sunucu adı"
	InitHost            string // "FTP host"
	InitPort            string // "Port"
	InitUser            string // "Kullanıcı adı"
	InitPassword        string // "Şifre"
	InitRemotePath      string // "Uzak dizin"
	InitCreated         string // "✓ syncftp.json oluşturuldu (izinler: 600)"
	InitReadyFmt        string // "Hazır! 'syncftp sync' komutu ile %q projesini senkronize edebilirsiniz.\n"
	InitIgnoreExists    string // "  (syncFTP girdileri zaten %s içinde)\n"
	InitIgnoreAdded     string // "✓ syncftp.json, syncftp.exe → %s'e eklendi\n"
	InitIgnoreCreated   string // "✓ syncftp.ignore oluşturuldu (syncftp.json, syncftp.exe eklendi)"

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
	// Windows: GetUserDefaultUILanguage via kernel32.dll (no extra deps needed)
	if runtime.GOOS == "windows" {
		if dll, err := syscall.LoadDLL("kernel32.dll"); err == nil {
			defer dll.Release()
			if proc, err := dll.FindProc("GetUserDefaultUILanguage"); err == nil {
				langID, _, _ := proc.Call()
				// LANGID: bits 0-9 = primary language; LANG_TURKISH = 0x1F (31)
				if langID&0x3FF == 0x1F {
					return "tr"
				}
			}
		}
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

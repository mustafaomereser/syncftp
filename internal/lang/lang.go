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

	// ── smartFirstSync ───────────────────────────────────────────────────────
	SmartSyncScanning   string // "  İlk sync: sunucu taranıyor, mevcut dosyalar karşılaştırılıyor..."
	SmartSyncConnErr    string // "  ! Sunucuya bağlanılamadı (%v) — tüm dosyalar yüklenecek\n"
	SmartSyncListErr    string // "  ! Uzak dizin okunamadı (%v) — tüm dosyalar yüklenecek\n"
	SmartSyncFoundFmt   string // "  Sunucuda %d dosya bulundu\n"
	SmartSyncResultFmt  string // "  Sonuç: %d güncel (yüklenmeyecek)  |  %d farklı/eksik (yüklenecek)\n"

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
}

// L is the active locale. Defaults to English.
var L = En

var currentLang = "en"

// Init reads language preference: .syncftp/lang file, then SYNCFTP_LANG env var.
// Call once from main() before any output.
func Init(configDir string) {
	l := "en"
	// 1. saved preference
	if data, err := os.ReadFile(filepath.Join(configDir, ".syncftp", "lang")); err == nil {
		l = strings.TrimSpace(string(data))
	}
	// 2. env var overrides file
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

// ── English ──────────────────────────────────────────────────────────────────

var En = StringSet{
	// Browser TUI
	BrowserLoading:           "  Loading...",
	BrowserSearchPrompt:      "  Enter=Search All  ESC=Cancel",
	BrowserSearching:         "  Searching: %q ...",
	BrowserSearchResults:     "  (full server, %d results)  ESC to clear",
	BrowserPickDirTitle:      "  SELECT TARGET DIRECTORY",
	BrowserPickDirEnter:      "Enter=Select This Dir",
	BrowserHintRight:         "→ Enter",
	BrowserHintLeft:          "← Up",
	BrowserHintQuit:          "q Cancel",
	BrowserMarkedHintDelete:  "d=Delete",
	BrowserMarkedHintMove:    "m=Move",
	BrowserMarkedHintFreeze:  "f=❄ Freeze",
	BrowserMarkedHintAll:     "a=Deselect all",
	BrowserMarkedHintCancel:  "ESC=Cancel",
	BrowserEscPending:        "  ESC again to exit",
	BrowserSearchClear:       "  ESC to clear",
	BrowserHintNav:           "  ↑↓ Navigate",
	BrowserHintEnter:         "Enter Open/Select",
	BrowserHintPreview:       "→ Preview",
	BrowserHintSpace:         "Space Mark",
	BrowserHintSearch:        "/ Search",
	BrowserHintQuitKey:       "q Quit",
	BrowserHintFreeze:        "f=❄ Freeze",
	BrowserHintTree:          "t=Tree",
	BrowserErrPrefix:         "  Error: ",
	BrowserReconnectHint:     "r = Reconnect  |  q = Quit",
	BrowserSelectThisPrefix:  "[ Select this dir: ",
	BrowserSelectThisSuffix:  " ]",
	BrowserCountMany:         "many items",
	BrowserFooterFmt:         "  %d dirs, %d files  ·  %d/%d",
	BrowserSearchMatchFmt:    "  search:%q %d matches",
	BrowserPreviewLoading:    "  Loading...",
	BrowserPreviewBack:       "  ESC/q back",
	BrowserPreviewScrollHint: "↑↓ scroll",
	BrowserPreviewPageHint:   "PgUp/PgDn",
	BrowserPreviewGGHint:     "g top  G bottom",
	BrowserPreviewEmpty:      "(empty file)",
	BrowserPreviewLineFmt:    "%d-%d / %d lines",
	BrowserConnLost:          "⚠ Connection lost",
	BrowserConnLostHint:      "Press r to reconnect",
	BrowserPreviewFailed:     "⚠ Preview unavailable",
	BrowserReconnectFailed:   "reconnect failed: %w",
	BrowserMarkedCount:       "  ✓ %d marked",
	BrowserSearchCancelHint:  "   ESC=cancel",
	BrowserSearchTimeout:     "⚠ timeout (5m) — %d partial results",
	BrowserSearchConnErr:     "⚠ connection error: %s",
	BrowserSearchOpenErr:     "connection %d failed to open: %v",
	BrowserSearchDoneFmt:     "%d results (%.1fs)",
	BrowserSearchOpsHint:     "  ↑↓ navigate  Space=mark  Enter=preview  → preview  d=delete marked  m=move marked  f=freeze  ESC=clear",
	BrowserSearchNoResults:   "  (no results found)",

	// Picker / Confirm TUI
	PickerFilterHint:   "  Type to filter",
	PickerMatchFmt:     "  (%d matches)",
	PickerNoMatch:      "  (no match — Backspace to clear)",
	PickerNavHint:      "↑↓/jk navigate   Enter select   [1-9] quick   Backspace clear   q cancel",
	MultiPickerNavHint: "↑↓ navigate   Space mark   a all   Enter confirm   q cancel",
	ConfirmYes:         "Yes",
	ConfirmNo:          "No, cancel",

	// Sync TUI
	SyncDryRunTitle: "  [DRY RUN] Files to upload:",
	SyncDryRunFmt:   "  %d files — press any key",
	SyncNoChange:    "  ✓ No changes — up to date",
	SyncAnyKey:      "  Press any key",
	SyncConnecting:  "connecting...",
	SyncRetryFmt:    " (%d attempts)",
	SyncRetryOkFmt:  " (succeeded on attempt %d)",
	SyncDoneFmt:          "  Done: %s uploaded",
	SyncDoneFailFmt:      ", %s failed",
	SyncAttemptDetailFmt: "attempt %d: %v",

	// Shell REPL
	ShellWelcomeHint:     "  Type 'help', 'exit' to quit",
	ShellMultiServer:     "Multiple servers — to connect: server <name>",
	ShellServersLabel:    "Servers:",
	ShellCtrlCHint:       "(Ctrl+C — type 'exit' to quit)",
	ShellCtrlCExitHint:   "Press Ctrl+C again to exit — or type 'exit'",
	ShellExit:            "Goodbye.",
	ShellUnknownCmd:      "Unknown command: %q  (type 'help')\n",
	ShellConnecting:      "Connecting: %s (%s)...",
	ShellConnectErr:      " error: %v\n",
	ShellNotConnected:    "Not connected. Use 'server <name>' to connect.",
	ShellNoServers:       "No active servers.",
	ShellServersFmt:      "Servers",
	ShellServersSubtitle: "Select to connect — q to close",
	ShellAlreadyConn:     "Already connected: %s\n",
	ShellDisconnected:    "Disconnected: %s\n",
	ShellServerNotFound:  "Server not found: %q\n",
	ShellBrowserErr:      "Browser error: %v\n",
	ShellDeleteTitle:     "Delete %d items",
	ShellDeleteSubtitle:  "This cannot be undone. Files will be permanently deleted from the FTP server.",
	ShellCancelled:       "Cancelled.",
	ShellConfirmSure:     "Are you sure?",
	ShellConfirmSureBody: "Select Yes to confirm.",
	ShellDeletedFmt:      "\n%d deleted, %d failed\n",
	ShellMovingFmt:       "\n%d files to move. Select destination (Enter = select this dir):\n",
	ShellMoveTarget:      "Destination: %s\n\n",
	ShellMovedFmt:        "\n%d moved, %d failed\n",
	ShellDirNotFound:     "Directory not found: %s\n",
	ShellDownloading:     "Downloading: %s → %s\n",
	ShellDownloaded:      "✓ Downloaded (%s)\n",
	ShellDownloadedBare:  "✓ Downloaded",
	ShellErrFmt:          "Error: %v\n",
	ShellDeleteConfTitle: "Delete Confirmation",
	ShellCancelShort:     "Cancelled.",
	ShellDeletedShort:    "✓ Deleted: %s\n",
	ShellPickServerTitle: "Select Server",
	ShellSyncServers:     "Servers to Sync",
	ShellSyncPickSub:     "Space to mark, Enter to confirm",
	ShellSyncCancelled:   "Cancelled.",
	ShellFileActionTitle: "File Action",
	ShellFileActionView:  "View Contents",
	ShellFileActionGet:   "Download",
	ShellFileActionDel:   "Delete",
	ShellFileActionCancel: "Cancel",
	ShellFileTruncFmt:    "[First %d KB — to get all: get %s]\n",
	ShellScanning:        "Scanning...",
	ShellScannedFmt:      " %d files\n",
	ShellConnPoolErr:     "  Connection error: %v\n",
	ShellReleaseFmt:      "  Release: %s\n",
	ShellStateReadErr:    "  Could not read state: %v\n",
	ShellStatusDeleted:   "  - %s (deleted locally)\n",
	ShellStatusUpToDate:  "  Up to date",
	ShellStatusNoChange:  "  %d changes\n",
	ShellStatusStateErr:  "[%s] could not read state: %v\n",

	TreeMaxPromptTitle: "Tree view",
	TreeMaxPromptSub:   "Max files per directory?",
	TreeSkippedFmt:     "[%d items — skipped, increase with --max]",
	TreeErrFmt:         "[error: %v]",

	ShellHelp: `
Remote server commands:
  ls [dir]                Interactive file browser (↑↓ navigate, → enter/preview, ← back)
                            Space = mark  |  d = delete marked  |  m = move marked
                            f = freeze/unfreeze file or whole folder  |  / = search
                            t = tree view of current directory
  tree [path] [--max N]   Show FTP directory as tree (dirs first, files with sizes)
                            --max N : skip dirs with > N items; show first N + "+X more..."
                            omit --max for interactive prompt (20 / 50 / 100 / all)
  cd <dir>                Change remote directory  (cd .. to go up)
  cat [file]              View file contents (opens browser if no arg)
  get [file] [dest]       Download file (opens browser if no arg)
  rm [-f] [-r] [file]     Delete file/dir  (-f no confirmation, -r recursive)
  pwd                     Show remote directory

Sync:
  status                  Show local changes per server
  sync [--all] [--full] [--dry-run] [--server name]   Upload to FTP
  freeze [--server name]  Manage freeze list — mark files to never upload

Server management:
  servers                 Server list
  server [name]           Select / connect to server
  disconnect              Close current FTP connection (prompt returns to no-server mode)
  config                  Add, edit or delete servers — full TUI field navigator
                            ↑↓ = select field  |  Enter = edit  |  Space = toggle bool
                            b = local file browser for include / exclude / protect / local dir
                            s = save  |  q = cancel
                            Global Settings → Default dir: project-wide local directory
                            Server edit → Local dir: per-server local directory
                              "(project default)" = inherits from global Default dir

Other:
  lang [en|tr]            Show or change display language
  clear / cls             Clear screen
  help / ?                This help
  exit / quit             Exit`,

	// lang command (En)
	LangCurrentFmt:  "Language: %s\n",
	LangSwitchedFmt: "Language set to: %s\n",
	LangSavedFmt:    "Saved to .syncftp/lang\n",
	LangInvalid:     "Unknown language: %q — use 'en' or 'tr'\n",
	LangAlreadyFmt:  "Already using %s\n",

	// cmd_sync
	SyncCancelled:        "Cancelled.",
	SyncDryRunNote:       "[DRY RUN] Nothing will be uploaded",
	SyncWhitelistFmt:     "Whitelist (%d paths): only these will be synced\n",
	SyncExcludeFmt:       "Exclude (%d paths): these will be skipped\n",
	SyncServerErrFmt:     "  ERROR: %v\n",
	SyncNoChange2:        "  No changes",
	SyncDeletedHeader:    "  ! DELETED files (kept on FTP):\n",
	SyncProcessingFmt:    "  %d files to process",
	SyncProtectedFmt:     " (%d protected)",
	SyncProtectedLabel:   "    PROTECTED  %s\n",
	SyncUploadLabel:      "    WILL UPLOAD %s\n",
	SyncPoolFmt:          "  Connection pool: %d / Retry: %d\n",
	SyncDoneFullFmt:      "  Done: %d uploaded, %d protected, %d errors\n",
	SyncFailedSavedFmt:   "  ! %d failed files saved to .syncftp/failed/%s.json — retry with: syncftp sync --retry-failed\n",
	SyncFailedSaveErr:    "  ! Could not save failed list: %v\n",
	SyncStateErr:         "  ! Could not save state: %v\n",
	SyncReleaseErr:       "  ! Could not create release: %v\n",
	SyncReleaseFmt:       "  Release: %s\n",
	SyncFullFlag:         "  Full sync (--full): all files will be uploaded",
	SyncRetryNoFiles:     "  No failed files to retry",
	SyncRetryModeFmt:     "  Retry mode: %d failed files (%s)\n",
	SyncRetrySkipFmt:     "  ! %s no longer exists locally, skipping\n",
	SyncAttemptsFmt:      "    ✗ %s (%d attempts): %v\n",
	SyncAttemptOkFmt:     "    ✓ %s (succeeded on attempt %d)\n",
	SyncUploadOkFmt:      "    ✓ %s\n",
	SyncUploadErrFmt:     "    ✗ %s: %v\n",

	// smartFirstSync
	SmartSyncScanning:  "  First sync: scanning server, comparing existing files...",
	SmartSyncConnErr:   "  ! Could not connect to server (%v) — uploading all files\n",
	SmartSyncListErr:   "  ! Could not read remote dir (%v) — uploading all files\n",
	SmartSyncFoundFmt:  "  Found %d files on server\n",
	SmartSyncResultFmt: "  Result: %d up to date (skipped)  |  %d different/missing (to upload)\n",

	// cmd_status
	StatusWhitelistFmt:     "Whitelist (%d paths): only these will be shown\n",
	StatusExcludeFmt:       "Exclude (%d paths): these will be hidden\n",
	StatusProjectFmt:       "Project: %s\n",
	StatusDirFmt:           "Dir    : %s\n",
	StatusFileFmt:          "Files  : %d\n",
	StatusStateErr:         "[%s] Could not load state: %v\n\n",
	StatusNoFirstSync:      "  No first sync yet",
	StatusUpToDate:         "  No changes — server up to date",
	StatusFilteredUpToDate: "  No changes in specified filter scope",
	StatusNewHeader:        "  + NEW",
	StatusChangedHeader:    "  ~ CHANGED",
	StatusDeletedHeader:    "  - DELETED (not removed from FTP, info only)",

	// cmd_push
	PushSourceFmt:      "Source  : %s\n",
	PushTargetFmt:      "Target  : %s:%d%s\n",
	PushScanning:       "Scanning...",
	PushScannedFmt:     " %d files\n\n",
	PushFirstPush:      "First push",
	PushFullPush:       "Full push (--full)",
	PushFirstFmt:       "  %s — all files will be uploaded\n",
	PushDeletedHeader:  "  ! DELETED files (kept on FTP):\n",
	PushNoChange:       "  No changes — target up to date",
	PushProcessingFmt:  "  %d files to process\n",
	PushDryRunHeader:   "\n[DRY RUN] Files to upload:",
	PushConnFmt:        "  Connections: %d / Retry: %d\n",
	PushAttemptsFmt:    "    ✗ %s (%d attempts): %v\n",
	PushAttemptOkFmt:   "    ✓ %s (attempt %d)\n",
	PushUploadOkFmt:    "    ✓ %s\n",
	PushUploadErrFmt:   "    ✗ %s: %v\n",
	PushDoneFmt:        "\n  Done: %d uploaded, %d errors\n",
	PushFailedHint:     "  ! Failed files saved — retry: syncftp push %s --server %s --full\n",
	PushStateErr:       "  ! Could not save state: %v\n",
	PushReleaseFmt:     "  Release: %s\n",

	// cmd_remote
	RemoteServerFmt:      "Server : %s (%s)\n",
	RemoteDirFmt:         "Dir    : %s\n\n",
	RemoteConnecting:     "Connecting: %s (%s)...\n",
	RemoteDownloading:    "Downloading: %s → %s\n",
	RemoteDownloaded:     "✓ Downloaded (%s)\n",
	RemoteDownloadedBare: "✓ Downloaded",
	RemoteDeleteLabel:    "Deleting %s: %s\n",
	RemoteDeleteConfirm:  "Are you sure you want to delete? [y/N]: ",
	RemoteDeleteCancel:   "Cancelled.",
	RemoteDeletedFmt:     "✓ Deleted: %s\n",
	RemoteCatTruncFmt:    "\n[First %d KB shown — to download all: syncftp remote get %s]\n",
	RemotePickDirErr:     "could not list directory: %w",
	RemotePickUpDir:      "  [0] .. (parent dir)",
	RemotePickPromptFmt:  "  Select [0-%d, q=cancel]: ",
	RemotePickInvalid:    "  Invalid selection: %q\n",
	RemotePickSelected:   "  Selected: %s\n\n",
	RemoteFileLabel:      "file",
	RemoteDirLabel:       "dir",
	RemoteDirRecLabel:    "dir (with contents)",
	RemoteListErr:        "%s  ! could not list: %v\n",
	RemoteNoServerSel:    "no server selected",

	// cmd_init
	InitWizardTitle:   "=== syncFTP Setup Wizard ===",
	InitProjectName:   "Project name",
	InitLocalDir:      "Local directory",
	InitFTPHeader:     "FTP Server details:",
	InitServerName:    "Server name",
	InitHost:          "FTP host",
	InitPort:          "Port",
	InitUser:          "Username",
	InitPassword:      "Password",
	InitRemotePath:    "Remote directory",
	InitCreated:       "✓ syncftp.json created (permissions: 600)",
	InitReadyFmt:      "Ready! Run 'syncftp sync' to synchronize project %q.\n",
	InitIgnoreExists:  "  (syncFTP entries already in %s)\n",
	InitIgnoreAdded:   "✓ syncftp.json, syncftp.exe → added to %s\n",
	InitIgnoreCreated: "✓ syncftp.ignore created (syncftp.json, syncftp.exe added)",
}

// ── Turkish ───────────────────────────────────────────────────────────────────

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
	SyncDryRunTitle: "  [DRY RUN] Yüklenecekler:",
	SyncDryRunFmt:   "  %d dosya — herhangi bir tuşa basın",
	SyncNoChange:    "  ✓ Değişiklik yok — güncel",
	SyncAnyKey:      "  Herhangi bir tuşa basın",
	SyncConnecting:  "bağlanıyor...",
	SyncRetryFmt:    " (%d deneme)",
	SyncRetryOkFmt:  " (%d. denemede)",
	SyncDoneFmt:          "  Tamamlandı: %s yüklendi",
	SyncDoneFailFmt:      ", %s hata",
	SyncAttemptDetailFmt: "deneme %d: %v",

	// Shell REPL
	ShellWelcomeHint:     "  'help' yazın, çıkmak için 'exit'",
	ShellMultiServer:     "Birden fazla sunucu var — bağlanmak için: server <ad>",
	ShellServersLabel:    "Sunucular:",
	ShellCtrlCHint:       "(Ctrl+C — çıkmak için 'exit')",
	ShellCtrlCExitHint:   "Çıkmak için Ctrl+C'ye tekrar basın — ya da 'exit' yazın",
	ShellExit:            "Görüşürüz.",
	ShellUnknownCmd:      "Bilinmeyen komut: %q  (yardım için 'help')\n",
	ShellConnecting:      "Bağlanıyor: %s (%s)...",
	ShellConnectErr:      " hata: %v\n",
	ShellNotConnected:    "Bağlı değil. 'server <ad>' ile bağlanın.",
	ShellNoServers:       "Aktif sunucu yok.",
	ShellServersFmt:      "Sunucular",
	ShellServersSubtitle: "Seçince bağlanır — q ile sadece kapat",
	ShellAlreadyConn:     "Zaten bağlı: %s\n",
	ShellDisconnected:    "Bağlantı kesildi: %s\n",
	ShellServerNotFound:  "Sunucu bulunamadı: %q\n",
	ShellBrowserErr:      "Browser hatası: %v\n",
	ShellDeleteTitle:     "%d öğeyi sil",
	ShellDeleteSubtitle:  "Bu işlem geri alınamaz. FTP sunucusundan kalıcı olarak silinir.",
	ShellCancelled:       "İptal edildi.",
	ShellConfirmSure:     "Emin misiniz?",
	ShellConfirmSureBody: "Onaylamak için Evet seçin.",
	ShellDeletedFmt:      "\n%d silindi, %d başarısız\n",
	ShellMovingFmt:       "\n%d dosya taşınacak. Hedef klasörü seçin (Enter = bu klasörü seç):\n",
	ShellMoveTarget:      "Hedef: %s\n\n",
	ShellMovedFmt:        "\n%d taşındı, %d başarısız\n",
	ShellDirNotFound:     "Dizin bulunamadı: %s\n",
	ShellDownloading:     "İndiriliyor: %s → %s\n",
	ShellDownloaded:      "✓ İndirildi (%s)\n",
	ShellDownloadedBare:  "✓ İndirildi",
	ShellErrFmt:          "Hata: %v\n",
	ShellDeleteConfTitle: "Silme Onayı",
	ShellCancelShort:     "İptal.",
	ShellDeletedShort:    "✓ Silindi: %s\n",
	ShellPickServerTitle: "Sunucu Seç",
	ShellSyncServers:     "Sync Edilecek Sunucular",
	ShellSyncPickSub:     "Space ile işaretle, Enter ile onayla",
	ShellSyncCancelled:   "İptal.",
	ShellFileActionTitle: "Dosya İşlemi",
	ShellFileActionView:  "İçeriği Görüntüle",
	ShellFileActionGet:   "İndir",
	ShellFileActionDel:   "Sil",
	ShellFileActionCancel: "İptal",
	ShellFileTruncFmt:    "[İlk %d KB — tamamı için: get %s]\n",
	ShellScanning:        "Taranıyor...",
	ShellScannedFmt:      " %d dosya\n",
	ShellConnPoolErr:     "  Bağlantı hatası: %v\n",
	ShellReleaseFmt:      "  Release: %s\n",
	ShellStateReadErr:    "  State okunamadı: %v\n",
	ShellStatusDeleted:   "  - %s (yerel silinmiş)\n",
	ShellStatusUpToDate:  "  Güncel",
	ShellStatusNoChange:  "  %d değişiklik\n",
	ShellStatusStateErr:  "[%s] state okunamadı: %v\n",

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
  status                  Yerel değişiklikleri göster
  sync [--all] [--full] [--dry-run] [--server ad]   FTP'ye yükle
  freeze [--server ad]    Freeze listesi — yüklenmesin istenen dosyaları işaretle

Sunucu yönetimi:
  servers                 Sunucu listesi
  server [ad]             Sunucu seç / bağlan
  disconnect              Mevcut FTP bağlantısını kes (prompt sunucusuz moda döner)
  config                  Sunucu ekle, düzenle, sil — TUI alan gezgini
                            ↑↓ = alan seç  |  Enter = düzenle  |  Space = bool toggle
                            b = include / exclude / protect için yerel dosya tarayıcısı
                            s = kaydet  |  q = iptal
                            Global Ayarlar → Varsayılan dizin: proje geneli yerel dizin
                            Sunucu düzenle → Yerel dizin: sunucuya özel yerel dizin
                              "(project default)" = global varsayılan dizinden miras alır

Diğer:
  lang [en|tr]            Dili göster veya değiştir
  clear / cls             Ekranı temizle
  help / ?                Bu yardım
  exit / quit             Çık`,

	// lang command (Tr)
	LangCurrentFmt:  "Dil: %s\n",
	LangSwitchedFmt: "Dil değiştirildi: %s\n",
	LangSavedFmt:    ".syncftp/lang dosyasına kaydedildi\n",
	LangInvalid:     "Bilinmeyen dil: %q — 'en' veya 'tr' kullanın\n",
	LangAlreadyFmt:  "Zaten %s dili kullanılıyor\n",

	// cmd_sync
	SyncCancelled:        "İptal.",
	SyncDryRunNote:       "[DRY RUN] Hiçbir şey yüklenmeyecek",
	SyncWhitelistFmt:     "Whitelist (%d yol): yalnızca bu yollar sync edilecek\n",
	SyncExcludeFmt:       "Exclude (%d yol): bu yollar bu sync'ten hariç tutulacak\n",
	SyncServerErrFmt:     "  HATA: %v\n",
	SyncNoChange2:        "  Değişiklik yok",
	SyncDeletedHeader:    "  ! SİLİNEN dosyalar (FTP'de bırakıldı):\n",
	SyncProcessingFmt:    "  %d dosya işlenecek",
	SyncProtectedFmt:     " (%d korunuyor)",
	SyncProtectedLabel:   "    KORUNUYOR  %s\n",
	SyncUploadLabel:      "    YÜKLENECEK %s\n",
	SyncPoolFmt:          "  Bağlantı havuzu: %d / Retry: %d\n",
	SyncDoneFullFmt:      "  Tamamlandı: %d yüklendi, %d korundu, %d hata\n",
	SyncFailedSavedFmt:   "  ! %d başarısız dosya .syncftp/failed/%s.json'a kaydedildi — tekrar denemek için: syncftp sync --retry-failed\n",
	SyncFailedSaveErr:    "  ! Failed listesi kaydedilemedi: %v\n",
	SyncStateErr:         "  ! State kaydedilemedi: %v\n",
	SyncReleaseErr:       "  ! Release oluşturulamadı: %v\n",
	SyncReleaseFmt:       "  Release: %s\n",
	SyncFullFlag:         "  Tam sync (--full): tüm dosyalar yüklenecek",
	SyncRetryNoFiles:     "  Yeniden denenecek başarısız dosya yok",
	SyncRetryModeFmt:     "  Retry modu: %d başarısız dosya (%s)\n",
	SyncRetrySkipFmt:     "  ! %s artık local'de yok, atlanıyor\n",
	SyncAttemptsFmt:      "    ✗ %s (%d deneme): %v\n",
	SyncAttemptOkFmt:     "    ✓ %s (%d. denemede başarılı)\n",
	SyncUploadOkFmt:      "    ✓ %s\n",
	SyncUploadErrFmt:     "    ✗ %s: %v\n",

	// smartFirstSync
	SmartSyncScanning:  "  İlk sync: sunucu taranıyor, mevcut dosyalar karşılaştırılıyor...",
	SmartSyncConnErr:   "  ! Sunucuya bağlanılamadı (%v) — tüm dosyalar yüklenecek\n",
	SmartSyncListErr:   "  ! Uzak dizin okunamadı (%v) — tüm dosyalar yüklenecek\n",
	SmartSyncFoundFmt:  "  Sunucuda %d dosya bulundu\n",
	SmartSyncResultFmt: "  Sonuç: %d güncel (yüklenmeyecek)  |  %d farklı/eksik (yüklenecek)\n",

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
	PushSourceFmt:      "Kaynak  : %s\n",
	PushTargetFmt:      "Hedef   : %s:%d%s\n",
	PushScanning:       "Taranıyor...",
	PushScannedFmt:     " %d dosya\n\n",
	PushFirstPush:      "İlk push",
	PushFullPush:       "Tam push (--full)",
	PushFirstFmt:       "  %s — tüm dosyalar yüklenecek\n",
	PushDeletedHeader:  "  ! SİLİNEN dosyalar (FTP'de bırakıldı):\n",
	PushNoChange:       "  Değişiklik yok — hedef güncel",
	PushProcessingFmt:  "  %d dosya işlenecek\n",
	PushDryRunHeader:   "\n[DRY RUN] Yüklenecekler:",
	PushConnFmt:        "  Bağlantı: %d / Retry: %d\n",
	PushAttemptsFmt:    "    ✗ %s (%d deneme): %v\n",
	PushAttemptOkFmt:   "    ✓ %s (%d. denemede)\n",
	PushUploadOkFmt:    "    ✓ %s\n",
	PushUploadErrFmt:   "    ✗ %s: %v\n",
	PushDoneFmt:        "\n  Tamamlandı: %d yüklendi, %d hata\n",
	PushFailedHint:     "  ! Başarısız dosyalar kaydedildi — tekrar: syncftp push %s --server %s --full\n",
	PushStateErr:       "  ! State kaydedilemedi: %v\n",
	PushReleaseFmt:     "  Release: %s\n",

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
}

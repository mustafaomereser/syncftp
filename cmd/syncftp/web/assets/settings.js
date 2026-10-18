'use strict';

/* Uygulama ayarları (#/settings) — makine geneli: dil, Windows ile başlat, webhook dinleyicisi +
   tünel, projeler (her birinin sync/global config ayarlarına kısayol), SSH host anahtarları,
   hash önbelleği, veri klasörleri, hakkında. Proje bazlı config projenin "Ayarlar" sekmesinde. */

Object.assign(T.tr, {
  navSettings: 'Ayarlar', setTitle: 'Ayarlar', setSub: 'Uygulama geneli ayarlar — proje bazlı sync ayarları her projenin Proje ayarları sekmesinde',
  setGeneral: 'Genel', setLang: 'Dil', setLangHint: 'Panel ve log mesajlarının dili',
  setEditSave: 'Dosya düzenleyicide kaydet', setEditSaveHint: 'Ctrl+S basınca nereye kaydedilsin. "Her seferinde sor" değişiklik farkını ve seçenekleri gösterir.', setEditAsk: 'Her seferinde sor', setEditBoth: 'Sunucuya ve yerele', setEditRemote: 'Yalnızca sunucuya',
  setAutostart: 'Windows açılışında başlat', setAutostartHint: 'Oturum açılınca syncFTP pencere açmadan tepside başlar; açık watch\'lar ve webhook dinleyicisi kaldığı yerden devam eder.',
  setAutostartOn: 'Windows açılışında başlayacak', setAutostartOff: 'Otomatik başlatma kapatıldı',
  setTrayOn: 'Tepsi simgesi açık — pencereyi kapatsanız da syncFTP tepside çalışmaya devam eder.',
  setHooks: 'CI/CD webhook dinleyicisi', setHooksDesc: 'Git sağlayıcılarından gelen push\'ları dinleyen yerel sunucu. Tetikleyicileri projenin CI/CD sekmesinden eklersiniz.',
  setPort: 'Dinleyici portu', setPortHint: '0 = varsayılan ({0}). Değişiklik syncFTP yeniden başlatılınca geçerli olur.',
  setPortActive: 'Şu an dinlenen port: {0}', setPortRestart: 'Yeni port bir sonraki açılışta geçerli olacak',
  setProjects: 'Projeler', setProjectsDesc: 'Sync filtreleri, yedek, bant genişliği, hook\'lar, ignore ve anahtarlık gibi global config ayarları projeye özeldir — buradan her projenin ayarlarına geçin.',
  setOpenSettings: 'Ayarlarını aç', setServersN: '{0} sunucu · {1} aktif', setBroken: 'config okunamıyor',
  setHostKeys: 'SSH host anahtarları', setHostKeysDesc: 'SFTP / SSH tüneli ile bağlandığınız sunucuların ilk bağlantıda kaydedilen anahtarları (TOFU). Sunucu anahtarı gerçekten değiştiyse kaydı silin; bir sonraki bağlantıda yeniden öğrenilir.',
  setNoHostKeys: 'Henüz kayıtlı anahtar yok', setHostDelConfirm: '{0} kaydı silinsin mi? Bir sonraki bağlantıda anahtar yeniden kaydedilir.',
  colHost: 'Host', colKeyType: 'Tür', colFingerprint: 'Parmak izi',
  setCache: 'Hash önbelleği', setCacheDesc: 'Değişmeyen dosyaları yeniden hash\'lememek için boyut + değişiklik zamanı saklanır. Temizlerseniz ilk tarama biraz uzun sürer, sonuç değişmez.',
  setCacheSize: '{0} · {1} proje dizini', setCacheClear: 'Önbelleği temizle', setCacheCleared: 'Önbellek temizlendi',
  setAbout: 'Hakkında', setVersion: 'Sürüm', setExe: 'Program', setPlatform: 'Platform',
  setOpenFolder: 'Klasörü aç',
  setLocs: "Veri konumları", setLocsDesc: "syncFTP'nin diske yazdığı her şey. Anahtarlık modu açıksa şifreler bu dosyalarda değil, Windows Kimlik Bilgisi Yöneticisi'nde durur.",
  setLocApp: 'Uygulama ayarları', setLocCache: 'Önbellek (silinebilir)', setLocTemp: 'Geçici dosyalar', setLocProjects: 'Projeler',
  setLocMissing: 'henüz yok', setLocReveal: 'Dosyayı klasöründe göster',
  loc_config_dir: 'Ayar klasörü', locHint_config_dir: 'Aşağıdaki uygulama dosyalarının hepsi burada',
  loc_library: 'Bağlantı kütüphanesi', locHint_library: 'Kayıtlı bağlantılar ve yeni proje varsayılanları',
  loc_gui_state: 'Panel durumu', locHint_gui_state: 'Açık watch\'lar (panel açılınca geri gelir), webhook portu / tüneli, editör tercihi',
  loc_gui_lang: 'Dil tercihi', locHint_gui_lang: 'Panelin ve CLI\'ın dili',
  loc_registry: 'CLI proje kaydı', locHint_registry: 'syncftp project / --project için kayıtlı projeler',
  loc_known_hosts: 'SFTP host anahtarları', locHint_known_hosts: 'İlk bağlantıda öğrenilen sunucu anahtarları (~/.ssh/known_hosts da okunur)',
  loc_cloudflared: 'cloudflared', locHint_cloudflared: 'Webhook tüneli için indirilen program',
  loc_hashcache: 'Hash önbelleği', locHint_hashcache: 'Değişmeyen dosyalar yeniden okunmasın diye boyut + tarih + hash',
  loc_panel_profile: 'Panel penceresi profili', locHint_panel_profile: 'Panelin açıldığı tarayıcı penceresinin profili',
  loc_temp: 'Sistem Temp klasörü', locHint_temp: 'İndirme / yükleme / düzenleme sırasında syncftp-dl-*, syncftp-up-*, syncftp-edit-* açılır, iş bitince silinir',
  locHint_project: 'Config, state, release yedekleri (rollback), loglar, freeze / failed / milestone, watch kilidi, AI dokümanları — .gitignore\'da',
});
Object.assign(T.en, {
  navSettings: 'Settings', setTitle: 'Settings', setSub: 'App-wide settings — per-project sync settings live in each project\'s Project settings tab',
  setGeneral: 'General', setLang: 'Language', setLangHint: 'Language of the panel and log messages',
  setEditSave: 'Saving in the file editor', setEditSaveHint: 'Where Ctrl+S saves. "Ask every time" shows the diff and the options.', setEditAsk: 'Ask every time', setEditBoth: 'Server and local', setEditRemote: 'Server only',
  setAutostart: 'Start with Windows', setAutostartHint: 'When you sign in, syncFTP starts in the tray without opening a window; open watches and the webhook listener resume.',
  setAutostartOn: 'Will start with Windows', setAutostartOff: 'Autostart disabled',
  setTrayOn: 'Tray icon is on — closing the window keeps syncFTP running in the tray.',
  setHooks: 'CI/CD webhook listener', setHooksDesc: 'Local server that receives pushes from Git providers. Add triggers in a project\'s CI/CD tab.',
  setPort: 'Listener port', setPortHint: '0 = default ({0}). Takes effect after syncFTP restarts.',
  setPortActive: 'Currently listening on: {0}', setPortRestart: 'The new port takes effect on next start',
  setProjects: 'Projects', setProjectsDesc: 'Sync filters, backups, bandwidth, hooks, ignore and keyring are per-project config — jump to each project\'s settings from here.',
  setOpenSettings: 'Open settings', setServersN: '{0} servers · {1} enabled', setBroken: 'config unreadable',
  setHostKeys: 'SSH host keys', setHostKeysDesc: 'Keys of SFTP / SSH tunnel hosts saved on first connect (TOFU). If a server key really changed, delete its entry; it is learned again on next connect.',
  setNoHostKeys: 'No saved keys yet', setHostDelConfirm: 'Delete the entry for {0}? The key will be saved again on next connect.',
  colHost: 'Host', colKeyType: 'Type', colFingerprint: 'Fingerprint',
  setCache: 'Hash cache', setCacheDesc: 'Size + modification time are stored so unchanged files are not hashed again. Clearing it only makes the next scan slower.',
  setCacheSize: '{0} · {1} project folders', setCacheClear: 'Clear cache', setCacheCleared: 'Cache cleared',
  setAbout: 'About', setVersion: 'Version', setExe: 'Program', setPlatform: 'Platform',
  setOpenFolder: 'Open folder',
  setLocs: 'Data locations', setLocsDesc: 'Everything syncFTP writes to disk. With keyring mode on, passwords are kept in the OS credential store, not in these files.',
  setLocApp: 'App settings', setLocCache: 'Cache (safe to delete)', setLocTemp: 'Temporary files', setLocProjects: 'Projects',
  setLocMissing: 'not created yet', setLocReveal: 'Show the file in its folder',
  loc_config_dir: 'Settings folder', locHint_config_dir: 'All the app files below live here',
  loc_library: 'Connection library', locHint_library: 'Saved connections and new-project defaults',
  loc_gui_state: 'Panel state', locHint_gui_state: 'Open watches (restored on start), webhook port / tunnel, editor preference',
  loc_gui_lang: 'Language', locHint_gui_lang: 'Language of the panel and the CLI',
  loc_registry: 'CLI project registry', locHint_registry: 'Projects registered for syncftp project / --project',
  loc_known_hosts: 'SFTP host keys', locHint_known_hosts: 'Server keys learned on first connect (~/.ssh/known_hosts is read too)',
  loc_cloudflared: 'cloudflared', locHint_cloudflared: 'Program downloaded for the webhook tunnel',
  loc_hashcache: 'Hash cache', locHint_hashcache: 'Size + time + hash so unchanged files are not read again',
  loc_panel_profile: 'Panel window profile', locHint_panel_profile: 'Browser profile of the panel window',
  loc_temp: 'System temp folder', locHint_temp: 'syncftp-dl-*, syncftp-up-*, syncftp-edit-* are created during downloads / uploads / edits and removed afterwards',
  locHint_project: 'Config, state, release backups (rollback), logs, freeze / failed / milestone, watch lock, AI docs — in .gitignore',
});

async function pageSettings(sub = "general") {
  const [s, hooks, lib] = await Promise.all([GET('/api/settings'), GET('/api/hooks').catch(() => ({ tunnel: {} })), GET('/api/library')]);
  if (!S.projects.length) await loadProjects();

  bind((root) => {
    $$('[data-open]', root).forEach((b) => (b.onclick = () => api('POST', '/api/settings/open', { what: b.dataset.open }).catch(fail)));
    $$('#set-lang button', root).forEach((b) => (b.onclick = async () => {
      try { await api('PUT', '/api/lang', { lang: b.dataset.l }); LANG = b.dataset.l; applyLang(); render(true); } catch (e) { fail(e); }
    }));
    $('#set-editsave', root).onchange = async (e) => {
      try { await api('PUT', '/api/settings', { edit_save_target: e.target.value }); toast(t('saved')); } catch (err) { fail(err); }
    };
    const as = $('#set-autostart', root);
    if (as) as.onchange = async () => {
      try { await api('PUT', '/api/settings', { autostart: as.checked }); toast(as.checked ? t('setAutostartOn') : t('setAutostartOff')); }
      catch (e) { as.checked = !as.checked; fail(e); }
    };
    $('#set-port-save', root).onclick = async () => {
      const v = parseInt($('#set-port', root).value, 10) || 0;
      try {
        await api('PUT', '/api/settings', { hook_port: v });
        toast((v || s.hook_default_port) !== s.hook_port_active ? t('setPortRestart') : t('saved'));
        render(true);
      } catch (e) { fail(e); }
    };
    bindTunnelCtl(root, hooks, 'set');
    $$('[data-hkdel]', root).forEach((b) => (b.onclick = async () => {
      if (!(await confirmBox(t('delete'), esc(t('setHostDelConfirm', b.dataset.host)), t('delete'), true))) return;
      try { await api('DELETE', `/api/settings/known-hosts/${b.dataset.hkdel}?fp=${encodeURIComponent(b.dataset.fp)}`); render(true); } catch (e) { fail(e); }
    }));
    $('#set-cache-clear', root).onclick = async () => {
      try { await api('POST', '/api/settings/cache-clear'); toast(t('setCacheCleared')); render(true); } catch (e) { fail(e); }
    };
    $('#set-add-proj', root).onclick = addProjectFlow;
    bindLibSection(root, lib, () => render(true));
    bindDefaultsSection(root, lib, () => render(true));
    bindFormTools(root);
  });

  const pathRow = (label, val, open) => `<div class="alan"><label>${label}</label><div class="girdi-grup"><input type="text" class="mono" value="${esc(val)}" readonly>
      <button class="btn" data-copy="${esc(val)}" title="${t('copyPath')}">${icon('copy')}</button>${open ? `<button class="btn" data-open="${open}" title="${t('setOpenFolder')}">${icon('folder')}</button>` : ''}</div></div>`;

  const general = section('sliders', t('setGeneral'), '', `
    <label class="anahtar"><span class="metin"><b>${t('setLang')}</b><small>${t('setLangHint')}</small></span>
      <span class="seg" id="set-lang"><button data-l="tr" class="${LANG === 'tr' ? 'aktif' : ''}">Türkçe</button><button data-l="en" class="${LANG === 'en' ? 'aktif' : ''}">English</button></span></label>
    ${s.autostart_supported ? `<label class="anahtar"><span class="metin"><b>${t('setAutostart')}</b><small>${t('setAutostartHint')}</small></span>
      <span class="switch"><input type="checkbox" id="set-autostart" ${s.autostart ? 'checked' : ''}><i></i></span></label>` : ''}
    <label class="anahtar"><span class="metin"><b>${t('setEditSave')}</b><small>${t('setEditSaveHint')}</small></span>
      <select id="set-editsave" style="width:auto">${[['', 'setEditAsk'], ['both', 'setEditBoth'], ['remote', 'setEditRemote']].map(([v, k]) => `<option value="${v}" ${(s.edit_save_target || '') === v ? 'selected' : ''}>${t(k)}</option>`).join('')}</select></label>
    ${s.tray ? `<p class="ipucu">${icon('check', 'ok-t')} ${t('setTrayOn')}</p>` : ''}`);

  const hooksSec = section('zap', t('setHooks'), t('setHooksDesc'), `
    <div class="alanlar">
      <div class="alan"><label>${t('setPort')}</label><div class="girdi-grup"><input type="number" id="set-port" value="${s.hook_port === s.hook_default_port ? 0 : s.hook_port}" min="0" max="65535">
        <button class="btn" id="set-port-save">${icon('check')}${t('save')}</button></div>
        <small>${t('setPortHint', s.hook_default_port)}${s.hook_port_active ? ' · ' + t('setPortActive', s.hook_port_active) : ''}</small></div>
      <div class="alan"><label>${t('ciPublic')}</label><div class="girdi-grup"><input type="text" class="mono" value="${esc(hooks.public_url || '')}" readonly><button class="btn" data-copy="${esc(hooks.public_url || '')}">${icon('copy')}</button></div></div>
    </div>
    <div class="ayrac"></div>
    ${tunnelCtlHTML(hooks, 'set')}`);

  const projSec = `<div class="bolum"><h3>${icon('layers')}${t('setProjects')}</h3><p class="aciklama">${t('setProjectsDesc')}</p>
    ${S.projects.length ? `<div class="tablo-kap"><table class="tablo"><tbody>${S.projects.map((p) => `<tr>
        <td><div class="ad-hucre">${icon(p.ok ? 'folder' : 'alert', p.ok ? 'acc-t' : 'bad-t')}<b>${esc(p.name)}</b>${p.keyring ? ` <span class="pill">${icon('lock')}keyring</span>` : ''}${p.watching ? ` <span class="pill ok"><span class="nokta"></span>watch</span>` : ''}</div>
          <div class="mono faint" style="font-size:12px;margin-top:2px">${esc(p.path)}</div></td>
        <td class="faint">${p.ok ? t('setServersN', p.servers, p.enabled) : `<span class="bad-t">${t('setBroken')}</span>`}</td>
        <td class="sag"><a class="btn kucuk" href="${pHash(p.name, 'settings')}">${icon('sliders')}${t('setOpenSettings')}</a></td></tr>`).join('')}</tbody></table></div>` : ''}
    <div class="aksiyonlar" style="margin-top:10px"><button class="btn" id="set-add-proj">${icon('plus')}${t('addProject')}</button></div></div>`;

  const hk = s.known_hosts || [];
  const hostSec = `<div class="bolum"><h3>${icon('shield')}${t('setHostKeys')}</h3><p class="aciklama">${t('setHostKeysDesc')}</p>
    ${hk.length ? `<div class="tablo-kap"><table class="tablo"><thead><tr><th>${t('colHost')}</th><th>${t('colKeyType')}</th><th>${t('colFingerprint')}</th><th></th></tr></thead><tbody>
      ${hk.map((h) => `<tr><td class="mono">${esc(h.hosts)}</td><td class="mono faint">${esc(h.type)}</td><td class="mono faint" style="font-size:12px;word-break:break-all">${esc(h.fingerprint)}</td>
        <td class="sag"><button class="ikon-btn tehlike" data-hkdel="${h.line}" data-fp="${esc(h.fingerprint)}" data-host="${esc(h.hosts)}" title="${t('delete')}">${icon('trash')}</button></td></tr>`).join('')}</tbody></table></div>`
      : `<div class="bos" style="padding:18px">${t('setNoHostKeys')}</div>`}
    <div class="alanlar" style="margin-top:12px">${pathRow('known_hosts', s.known_hosts_path, 'known_hosts')}</div></div>`;

  const cacheSec = section('gauge', t('setCache'), t('setCacheDesc'), `
    <div class="aksiyonlar"><span class="pill mono">${t('setCacheSize', bytes(s.cache_bytes), s.cache_files)}</span><span style="margin-left:auto"></span>
      <button class="btn" data-open="hashcache">${icon('folder')}${t('setOpenFolder')}</button><button class="btn tehlike" id="set-cache-clear" ${s.cache_files ? '' : 'disabled'}>${icon('trash')}${t('setCacheClear')}</button></div>`);

  const about = section('help', t('setAbout'), '', `<div class="alanlar">
    <div class="alan"><label>${t('setVersion')}</label><input type="text" class="mono" value="v${esc(s.version)}" readonly></div>
    <div class="alan"><label>${t('setPlatform')}</label><input type="text" class="mono" value="${esc(s.os)} · ${esc(s.go)}" readonly></div>
    ${pathRow(t('setExe'), s.exe, '')}</div>`);

  // Veri konumları: Go'daki dataLocations listesi (Klasörü aç da yalnızca bunları açar)
  const locs = s.locations || [];
  const byKey = Object.fromEntries(locs.map((l) => [l.key, l]));
  const locBtns = (l) => `<span class="konum-aks">${l.exists ? '' : `<span class="pill">${t('setLocMissing')}</span>`}
      <button class="ikon-btn" data-copy="${esc(l.path)}" title="${t('copyPath')}">${icon('copy')}</button><button class="ikon-btn" data-open="${esc(l.key)}" title="${t(l.file && l.exists ? 'setLocReveal' : 'setOpenFolder')}">${icon('folder')}</button></span>`;
  // Satır tek satır: ad (açıklama tooltip'te) · kök klasöre göreceli yol (tam yol tooltip'te) · düğmeler
  const locRow = (l, root) => {
    const rel = root && l.path.startsWith(root) ? l.path.slice(root.length).replace(/^[\\/]/, '') : l.path;
    return `<div class="konum-satir"><span class="konum-ad" title="${esc(t(l.project ? 'locHint_project' : 'locHint_' + l.key))}">${esc(l.project || t('loc_' + l.key))}</span>
      <span class="konum-yol mono" title="${esc(l.path)}">${esc(rel)}</span>${locBtns(l)}</div>`;
  };
  // Grup: başlıkta ortak kök klasör (kopyala/aç), altında o klasöre göreceli satırlar
  const locGroup = (title, rootKey, rows) => {
    const root = byKey[rootKey];
    return `<div class="konum-grup"><div class="konum-bas"><span class="konum-baslik">${title}</span>
      ${root ? `<span class="konum-yol mono" title="${esc(root.path)}">${esc(root.path)}</span>${locBtns(root)}` : ''}</div>
      ${rows.map((l) => locRow(l, root && root.path)).join('')}</div>`;
  };
  const projLocs = locs.filter((l) => l.group === 'project');
  const locSec = section('folder', t('setLocs'), t('setLocsDesc'),
    locGroup(t('setLocApp'), 'config_dir', locs.filter((l) => l.group === 'app' && l.key !== 'config_dir')) +
    locGroup(t('setLocCache'), 'cache_dir', locs.filter((l) => l.group === 'cache' && l.key !== 'cache_dir')) +
    locGroup(t('setLocTemp'), '', locs.filter((l) => l.group === 'temp')) +
    (projLocs.length ? `<details class="konum-grup konum-kat"><summary class="konum-bas"><span class="konum-baslik">${t('setLocProjects')} · ${projLocs.length}</span>
      <span class="faint" style="font-size:12px">${esc(t('locHint_project'))}</span>${icon('right', 'ok-ac')}</summary>
      ${projLocs.map((l) => locRow(l, '')).join('')}</details>` : ''));

  // Tek sayfada 8 bölüm alt alta kafa karıştırıyordu → her iş kendi alt sayfasında
  const tabs = [
    ['general', 'sliders', t('setTabGeneral'), general],
    ['library', 'link', t('setTabLibrary'), libSectionHTML(lib)],
    ['defaults', 'layers', t('setTabDefaults'), defaultsSectionHTML(lib)],
    ['projects', 'folder', t('setTabProjects'), projSec],
    ['hooks', 'zap', t('setTabHooks'), hooksSec],
    ['security', 'shield', t('setTabSecurity'), hostSec],
    ['system', 'gauge', t('setTabSystem'), cacheSec + about + locSec],
  ];
  const cur = tabs.find((x) => x[0] === sub) || tabs[0];
  return `<div class="sayfa"><div class="baslik-satir"><div><h1>${t('setTitle')}</h1><div class="yol">${esc(t('setSub'))}</div></div></div>
    <div class="tabs" style="margin-bottom:16px">${tabs.map(([k, ic, label]) => `<a href="#/settings/${k}" class="${k === cur[0] ? 'aktif' : ''}">${icon(ic)}${label}</a>`).join('')}</div>
    ${tabs.map(([k, , , html]) => `<div class="form" ${k === cur[0] ? '' : 'hidden'}>${html}</div>`).join('')}</div>`;
}

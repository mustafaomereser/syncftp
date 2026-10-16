'use strict';

/* Proje bağımsız panel parçaları (CLI `syncftp global` ile aynı veri):
   - Bağlantı kütüphanesi (Ayarlar) + düzenleme modalı
   - Yeni proje varsayılanları (Ayarlar)
   - Proje araçları: config dışa/içe aktar, varsayılan olarak kaydet, HTTP API (serve),
     kütüphaneden bağlantı ekle / kütüphaneye kaydet */

Object.assign(T.tr, {
  freezeFiles: 'Dondurulan dosyalar', ledEmpty: 'Boş', ledAdd: 'Ekle', ledPh: 'ya da kalıp yaz: *.log, cache/ … (Enter)', ledPhCmd: 'komut yaz, Enter ile ekle', remove: 'Kaldır',
  treeNeedRoot: 'Önce yerel klasörü seçin', treePickRoot: 'Ağaç için örnek bir proje klasörü seçin', defTreeHint: 'Ağaçtan seç: örnek bir proje klasöründen dosya/klasör işaretleyin; ya da alttaki hazır kalıplara tıklayın.',
  libTitle: 'Bağlantı kütüphanesi', libDesc: 'Proje bağımsız kayıtlı FTP / FTPS / SFTP bağlantıları. Yeni proje eklerken ve sunucu eklerken tek tıkla seçilir. Şifreler mümkünse Windows Kimlik Bilgileri Yöneticisi\'nde saklanır.',
  libEmpty: 'Henüz kayıtlı bağlantı yok', libAdd: 'Bağlantı ekle', libImport: 'Projeden aktar', libImportPick: 'Bağlantıları aktarılacak proje', libImported: '{0} bağlantı aktarıldı',
  libEditTitle: 'Bağlantı', libNewTitle: 'Yeni bağlantı', libName: 'Ad', libNameHint: 'init --connection / push --connection ile de kullanılır',
  libTestRemote: 'Test için uzak dizin', libDeleteConfirm: '{0} bağlantısı kütüphaneden silinsin mi? Projelerdeki kopyalar etkilenmez.',
  libTunnel: 'SSH tüneli (opsiyonel)', libSaved: 'Kütüphane kaydedildi',
  defTitle: 'Yeni proje varsayılanları', defDesc: 'Yeni oluşturulan her projeye (panel, AI yardımcısı ve CLI init) kopyalanan global sync ayarları. Mevcut projeler etkilenmez.',
  defTemplate: 'Ignore şablonu', defTemplateNone: 'Yok', defMaxConn: 'Yeni sunucu: max bağlantı', defMaxRetry: 'Yeni sunucu: max deneme',
  defFromProject: 'Projeden al', defFromPick: 'Ayarları alınacak proje', defReset: 'Fabrika ayarı', defResetConfirm: 'Varsayılanlar fabrika ayarına dönsün mü?', defSaved: 'Varsayılanlar kaydedildi',
  pushNoLib: 'Kayıtlı bağlantı yok — elle girin ya da Ayarlar\'dan ekleyin',
  ptTools: 'Config araçları', ptToolsDesc: 'Bu projenin config\'ini dışa aktarın (şifreler çıkarılır), başka bir projeden ya da JSON\'dan ayar alın, ya da bu projenin global ayarlarını yeni projeler için varsayılan yapın.',
  ptExport: 'Dışa aktar (JSON)', ptImport: 'İçe aktar…', ptAsDefaults: 'Yeni projeler için varsayılan yap', ptAsDefaultsDone: 'Varsayılanlar bu projeden alındı',
  ptImportTitle: 'Ayarları içe aktar', ptFrom: 'Kaynak', ptFromProject: 'Başka proje', ptFromJson: 'JSON yapıştır', ptParts: 'Alınacak bölümler',
  ptPartSync: 'Global sync ayarları', ptPartServers: 'Sunucular', ptPartConns: 'Bağlantı profilleri', ptPartDeploy: 'CI/CD tetikleyicileri',
  ptModeMerge: 'Birleştir (aynı adlıyı güncelle, yenileri ekle)', ptModeReplace: 'Değiştir (bu projedekileri sil)', ptImported: 'İçe aktarıldı',
  ptServe: 'HTTP API (syncftp serve)', ptServeDesc: 'Bu proje için yerel REST API (127.0.0.1) — PHP/web panellerinden status/sync/remote uçlarını çağırmak için. Panel kapanınca durur.',
  ptServeOn: 'API çalışıyor', ptServeStart: 'API\'yi başlat', ptServeStop: 'Durdur', ptServePort: 'Port',
  connFromLib: 'Kütüphaneden', connFromLibTitle: 'Kütüphaneden bağlantı ekle', connToLib: 'Kütüphaneye kaydet', connToLibDone: '{0} kütüphaneye kaydedildi', connAddedFromLib: '{0} bağlantı profili eklendi',
  apTemplate: 'Ignore şablonu', apTemplateDefault: 'Varsayılan ({0})', apFirstServer: 'İlk sunucu (opsiyonel)', apNoServer: 'Sunucusuz — sonra eklerim', apRemote: 'Uzak dizin',
  syncOnce: 'Tek seferlik filtreli sync…', syncOnceTitle: 'Filtreli sync', syncOnceDesc: 'Yalnızca bu çalıştırma için — config değişmez (CLI: sync --include / --exclude).', syncOnceInc: 'Sadece bunlar (include)', syncOnceExc: 'Bunları hariç tut (exclude)',
});
Object.assign(T.en, {
  freezeFiles: 'Frozen files', ledEmpty: 'Empty', ledAdd: 'Add', ledPh: 'or type a pattern: *.log, cache/ … (Enter)', ledPhCmd: 'type a command, Enter to add', remove: 'Remove',
  treeNeedRoot: 'Choose the local folder first', treePickRoot: 'Pick a sample project folder for the tree', defTreeHint: 'Pick from tree: tick files/folders in a sample project folder, or click the ready-made patterns below.',
  libTitle: 'Connection library', libDesc: 'Project-independent saved FTP / FTPS / SFTP connections. Pick them with one click when adding a project or a server. Passwords are stored in the OS keyring when available.',
  libEmpty: 'No saved connections yet', libAdd: 'Add connection', libImport: 'Import from project', libImportPick: 'Project to import connections from', libImported: '{0} connection(s) imported',
  libEditTitle: 'Connection', libNewTitle: 'New connection', libName: 'Name', libNameHint: 'Also usable with init --connection / push --connection',
  libTestRemote: 'Remote folder for test', libDeleteConfirm: 'Remove {0} from the library? Copies inside projects are not affected.',
  libTunnel: 'SSH tunnel (optional)', libSaved: 'Library saved',
  defTitle: 'New project defaults', defDesc: 'Global sync settings copied into every new project (panel, AI assistant and CLI init). Existing projects are not changed.',
  defTemplate: 'Ignore template', defTemplateNone: 'None', defMaxConn: 'New server: max connections', defMaxRetry: 'New server: max retries',
  defFromProject: 'Take from project', defFromPick: 'Project to take settings from', defReset: 'Factory defaults', defResetConfirm: 'Reset defaults to factory settings?', defSaved: 'Defaults saved',
  pushNoLib: 'No saved connections — enter one manually or add it in Settings',
  ptTools: 'Config tools', ptToolsDesc: 'Export this project\'s config (passwords removed), take settings from another project or JSON, or make this project\'s global settings the default for new projects.',
  ptExport: 'Export (JSON)', ptImport: 'Import…', ptAsDefaults: 'Use as defaults for new projects', ptAsDefaultsDone: 'Defaults taken from this project',
  ptImportTitle: 'Import settings', ptFrom: 'Source', ptFromProject: 'Another project', ptFromJson: 'Paste JSON', ptParts: 'Sections to import',
  ptPartSync: 'Global sync settings', ptPartServers: 'Servers', ptPartConns: 'Connection profiles', ptPartDeploy: 'CI/CD triggers',
  ptModeMerge: 'Merge (update same names, add new ones)', ptModeReplace: 'Replace (remove this project\'s entries)', ptImported: 'Imported',
  ptServe: 'HTTP API (syncftp serve)', ptServeDesc: 'Local REST API (127.0.0.1) for this project — call status/sync/remote endpoints from PHP/web panels. Stops when the panel closes.',
  ptServeOn: 'API running', ptServeStart: 'Start API', ptServeStop: 'Stop', ptServePort: 'Port',
  connFromLib: 'From library', connFromLibTitle: 'Add connection from library', connToLib: 'Save to library', connToLibDone: '{0} saved to the library', connAddedFromLib: 'Connection profile {0} added',
  apTemplate: 'Ignore template', apTemplateDefault: 'Default ({0})', apFirstServer: 'First server (optional)', apNoServer: 'No server — add later', apRemote: 'Remote folder',
  syncOnce: 'One-off filtered sync…', syncOnceTitle: 'Filtered sync', syncOnceDesc: 'Only for this run — config is not changed (CLI: sync --include / --exclude).', syncOnceInc: 'Only these (include)', syncOnceExc: 'Exclude these',
});

const TEMPLATE_LABEL = { laravel: 'Laravel', wordpress: 'WordPress', node: 'Node.js', generic: 'Genel / Generic' };
const connURL = (c) => `${(c.protocol || 'ftp').toLowerCase()}://${c.user ? c.user + '@' : ''}${c.host}:${c.port || ''}`;

// ── Bağlantı kütüphanesi ─────────────────────────────────────────────────────
function libSectionHTML(lib) {
  const conns = lib.connections || [];
  return `<div class="bolum"><h3>${icon('link')}${t('libTitle')}</h3><p class="aciklama">${t('libDesc')}</p>
    ${conns.length ? `<div class="tablo-kap"><table class="tablo"><tbody>${conns.map((c, i) => `<tr>
        <td><div class="ad-hucre">${icon('server', 'acc-t')}<b>${esc(c.name)}</b><span class="pill mono">${esc((c.protocol || 'ftp').toUpperCase())}</span>${c.tunnel_host ? `<span class="pill">${icon('tunnel')}SSH</span>` : ''}</div></td>
        <td class="mono faint">${esc(connURL(c))}</td>
        <td class="faint">${c.key_file ? icon('lock') + ' ' + esc(c.key_file) : c.password ? '••••' : ''}</td>
        <td class="sag"><button class="ikon-btn" data-libedit="${i}" title="${t('edit')}">${icon('edit')}</button>
          <button class="ikon-btn tehlike" data-libdel="${i}" title="${t('delete')}">${icon('trash')}</button></td></tr>`).join('')}</tbody></table></div>`
      : `<div class="bos" style="padding:18px">${t('libEmpty')}</div>`}
    ${(lib.secret_errors || []).length ? `<div class="not bad">${icon('alert')}${esc(lib.secret_errors.join('; '))}</div>` : ''}
    <div class="aksiyonlar" style="margin-top:10px"><button class="btn mavi" id="lib-add">${icon('plus')}${t('libAdd')}</button>
      ${S.projects.length ? `<button class="btn" id="lib-import">${icon('download')}${t('libImport')}</button>` : ''}</div></div>`;
}

function bindLibSection(root, lib, rerender) {
  const conns = lib.connections || [];
  const save = async (next) => { await api('PUT', '/api/library', { connections: next }); toast(t('libSaved')); rerender(); };
  $('#lib-add', root).onclick = () => libConnModal(null, conns, save);
  $$('[data-libedit]', root).forEach((b) => (b.onclick = () => libConnModal(+b.dataset.libedit, conns, save)));
  $$('[data-libdel]', root).forEach((b) => (b.onclick = async () => {
    const c = conns[+b.dataset.libdel];
    if (!(await confirmBox(t('delete'), esc(t('libDeleteConfirm', c.name)), t('delete'), true))) return;
    try { await save(conns.filter((_, i) => i !== +b.dataset.libdel)); } catch (e) { fail(e); }
  }));
  const imp = $('#lib-import', root);
  if (imp) imp.onclick = async () => {
    const p = await pickProjectModal(t('libImportPick'));
    if (!p) return;
    try { const r = await api('POST', `/api/library/import/${enc(p)}`); toast(t('libImported', r.added)); rerender(); } catch (e) { fail(e); }
  };
}

function pickProjectModal(title) {
  return new Promise((res) => {
    let done = false;
    modal({
      title, body: `<div class="liste-kutu">${S.projects.map((p) => `<div class="tikla pk-satir" data-p="${esc(p.name)}">${icon('folder', 'acc-t')} <b>${esc(p.name)}</b> <span class="faint mono" style="font-size:12px">${esc(p.path)}</span></div>`).join('')}</div>`,
      actions: [{ label: t('cancel'), cls: 'hayalet', run: () => { done = true; res(null); } }],
      onClose: () => { if (!done) res(null); },
    }).root.querySelectorAll('[data-p]').forEach((el) => (el.onclick = () => { done = true; res(el.dataset.p); el.closest('.modal-arka').remove(); }));
  });
}

// libConnModal idx=null → yeni. save(nextConnections) kaydeder.
function libConnModal(idx, conns, save) {
  const isNew = idx == null;
  const src = isNew ? { passive: true } : conns[idx];
  const m = modal({
    title: `${icon('link')}${isNew ? t('libNewTitle') : t('libEditTitle') + ' · ' + esc(src.name)}`, size: 'genis',
    body: `<div class="form" id="lib-form" data-p="" data-ppath="" data-srv="">
      <div class="alanlar">${fText('name', t('libName'), src.name, { hint: t('libNameHint') })}</div><div style="height:12px"></div>
      ${connFields(src, false)}
      <details style="margin-top:14px"${src.tunnel_host ? ' open' : ''}><summary class="faint" style="cursor:pointer">${icon('tunnel')} ${t('libTunnel')}</summary><div style="height:10px"></div>${tunnelFields(src, false)}</details>
      <div class="alanlar" style="margin-top:14px"><div class="alan"><label>${t('libTestRemote')}</label><div class="girdi-grup"><input type="text" class="mono" id="lib-remote" value="/" spellcheck="false"><button class="btn" id="lib-test">${icon('test')}${t('testConn')}</button></div><small id="lib-test-out"></small></div></div>
    </div>`,
    actions: [{ label: t('cancel'), cls: 'hayalet' }, { label: icon('check') + t('save'), cls: 'birincil', run: async (root) => {
      const c = readForm($('#lib-form', root), {});
      if (!c.name) { toast(t('libName') + '?', 'bad'); return false; }
      const next = conns.slice();
      if (isNew) next.push(c); else next[idx] = { ...conns[idx], ...c };
      await save(next);
    } }],
  });
  bindProtoPort(m.root);
  bindFormTools(m.root);
  $('#lib-test', m.root).onclick = async (e) => {
    const out = $('#lib-test-out', m.root);
    e.target.disabled = true; out.innerHTML = `${icon('refresh', 'spin')} ${t('testing')}`;
    try {
      const c = readForm($('#lib-form', m.root), {});
      const r = await api('POST', '/api/library/test', { connection: c, remote: $('#lib-remote', m.root).value.trim() });
      out.innerHTML = `<span class="ok-t">${icon('check')} ${t('testOk', r.ms, r.entries)}</span>${r.list_error ? ` <span class="warn-t">${esc(r.list_error)}</span>` : ''}`;
    } catch (err) { out.innerHTML = `<span class="bad-t">${esc(err.message)}</span>`; }
    e.target.disabled = false;
  };
}

// ── Yeni proje varsayılanları ────────────────────────────────────────────────
function defaultsSectionHTML(lib) {
  const d = lib.defaults || {}, sy = d.sync || {};
  const tmpl = [['', t('defTemplateNone')], ...(lib.templates || []).map((k) => [k, TEMPLATE_LABEL[k] || k])];
  return `<div class="bolum"><h3>${icon('sliders')}${t('defTitle')}</h3><p class="aciklama">${t('defDesc')}</p>
    <p class="ipucu" style="margin:-4px 0 12px">${icon('folder')} ${t('defTreeHint')}</p><div class="form" id="def-form" data-p="" data-ppath="" data-srv=""><div class="alanlar">
      ${fArea('sync.include', t('fInclude'), sy.include, { wide: false, tree: 'paths', chips: 'include' })}${fArea('sync.exclude', t('fExclude'), sy.exclude, { wide: false, tree: 'paths', chips: 'exclude' })}
      ${fArea('sync.protect', t('fProtect'), sy.protect, { wide: false, tree: 'protect', chips: 'protect' })}${fArea('sync.block_extensions', t('fBlock'), sy.block_extensions, { ty: 'words', wide: false, hint: t('fBlockHint'), chips: 'block' })}
      ${fArea('sync.ignore_files', t('fIgnoreFiles'), sy.ignore_files, { wide: false, tree: 'paths', ph: '.gitignore', hint: t('fIgnoreFilesHint') })}
      ${fSelect('ignore_template', t('defTemplate'), d.ignore_template || '', tmpl)}
      ${fText('sync.webhook', t('fWebhook'), sy.webhook, { mono: true, ph: 'https://…' })}
      ${fNum('sync.bandwidth_limit', t('fBandwidth'), sy.bandwidth_limit, { hint: '0 = ∞' })}
      ${fNum('sync.backup_keep', t('fBackupKeep'), sy.backup_keep || 10, { hint: t('fBackupKeepHint') })}
      ${fNum('max_connections', t('defMaxConn'), d.max_connections || 1)}
      ${fNum('max_retries', t('defMaxRetry'), d.max_retries || 2)}
      ${fArea('sync.pre_sync', t('fPre'), sy.pre_sync, { wide: false, rows: 2 })}${fArea('sync.post_sync', t('fPost'), sy.post_sync, { wide: false, rows: 2 })}
    </div><div style="height:12px"></div><div class="alanlar">
      ${fBool('sync.no_backup', t('fNoBackup'), sy.no_backup, t('fNoBackupHint'))}
      ${fBool('sync.direct_upload', t('fDirect'), sy.direct_upload, t('fDirectHint'))}
      ${fBool('first_sync_full', t('fFirstFull'), d.first_sync_full, t('fFirstFullHint'))}
      ${fBool('keyring', t('keyringTitle'), d.keyring, t('keyringHint'))}
    </div></div>
    <div class="aksiyonlar" style="margin-top:12px">${S.projects.length ? `<button class="btn" id="def-from">${icon('download')}${t('defFromProject')}</button>` : ''}
      <button class="btn hayalet" id="def-reset">${icon('undo')}${t('defReset')}</button><span style="margin-left:auto"></span>
      <button class="btn birincil" id="def-save">${icon('check')}${t('save')}</button></div></div>`;
}

function bindDefaultsSection(root, lib, rerender) {
  $('#def-save', root).onclick = async () => {
    const d = readForm($('#def-form', root), {});
    d.sync = { ...(lib.defaults.sync || {}), ...d.sync };
    try { await api('PUT', '/api/library', { defaults: d }); toast(t('defSaved')); rerender(); } catch (e) { fail(e); }
  };
  $('#def-reset', root).onclick = async () => {
    if (!(await confirmBox(t('defReset'), esc(t('defResetConfirm')), t('defReset'), true))) return;
    try { await api('PUT', '/api/library', { defaults: lib.builtin }); toast(t('defSaved')); rerender(); } catch (e) { fail(e); }
  };
  const from = $('#def-from', root);
  if (from) from.onclick = async () => {
    const p = await pickProjectModal(t('defFromPick'));
    if (!p) return;
    try { await api('POST', `/api/library/defaults-from/${enc(p)}`); toast(t('defSaved')); rerender(); } catch (e) { fail(e); }
  };
}

// ── Proje araçları (Proje ayarları sekmesi) ──────────────────────────────────
function projToolsHTML() {
  return `<div class="bolum"><h3>${icon('copy')}${t('ptTools')}</h3><p class="aciklama">${t('ptToolsDesc')}</p>
      <div class="aksiyonlar"><button class="btn" id="pt-export">${icon('download')}${t('ptExport')}</button>
        <button class="btn" id="pt-import">${icon('upload')}${t('ptImport')}</button>
        <button class="btn" id="pt-defaults">${icon('sliders')}${t('ptAsDefaults')}</button></div></div>
    <div class="bolum"><h3>${icon('zap')}${t('ptServe')}</h3><p class="aciklama">${t('ptServeDesc')}</p><div id="pt-serve"></div></div>`;
}

function bindProjTools(root, P) {
  $('#pt-export', root).onclick = () => { location.href = `/api/p/${enc(P)}/config/export`; };
  $('#pt-defaults', root).onclick = async () => {
    try { await api('POST', `/api/library/defaults-from/${enc(P)}`); toast(t('ptAsDefaultsDone')); } catch (e) { fail(e); }
  };
  $('#pt-import', root).onclick = () => importModal(P);
  const box = $('#pt-serve', root);
  const draw = async () => {
    const s = await GET(`/api/p/${enc(P)}/serve`).catch(() => ({ active: false }));
    box.innerHTML = s.active
      ? `<div class="aksiyonlar"><span class="pill ok"><span class="nokta"></span>${t('ptServeOn')}</span><span class="mono">${esc(s.url)}</span>
          <button class="ikon-btn" data-copy2="${esc(s.url)}">${icon('copy')}</button><span style="margin-left:auto"></span><button class="btn tehlike" id="srv-stop">${icon('stop')}${t('ptServeStop')}</button></div>`
      : `<div class="aksiyonlar"><label style="display:flex;gap:8px;align-items:center">${t('ptServePort')} <input type="number" id="srv-port" value="8080" style="width:110px"></label>
          <span style="margin-left:auto"></span><button class="btn mavi" id="srv-start">${icon('play')}${t('ptServeStart')}</button></div>`;
    const c = $('[data-copy2]', box); if (c) c.onclick = () => { navigator.clipboard.writeText(c.dataset.copy2); toast(t('copied')); };
    const st = $('#srv-start', box); if (st) st.onclick = async () => { try { await api('POST', `/api/p/${enc(P)}/serve`, { on: true, port: parseInt($('#srv-port', box).value, 10) || 8080 }); draw(); } catch (e) { fail(e); } };
    const sp = $('#srv-stop', box); if (sp) sp.onclick = async () => { await api('POST', `/api/p/${enc(P)}/serve`, { on: false }).catch(fail); draw(); };
  };
  draw();
}

function importModal(P) {
  const others = S.projects.filter((p) => p.name !== P);
  const m = modal({
    title: `${icon('upload')}${t('ptImportTitle')}`, size: 'orta',
    body: `<div class="form">
      <div class="alan"><label>${t('ptFrom')}</label><select id="im-src">${others.map((p) => `<option value="p:${esc(p.name)}">${t('ptFromProject')}: ${esc(p.name)}</option>`).join('')}<option value="json">${t('ptFromJson')}</option></select></div>
      <div class="alan gizli" id="im-json-box"><textarea id="im-json" rows="8" class="mono" spellcheck="false" placeholder="{ &quot;sync&quot;: { … } }"></textarea></div>
      <div class="alan"><label>${t('ptParts')}</label>
        ${[['sync', t('ptPartSync'), true], ['servers', t('ptPartServers')], ['connections', t('ptPartConns')], ['deploy', t('ptPartDeploy')]].map(([k, l, on]) =>
          `<label style="display:flex;gap:8px;align-items:center;margin:4px 0"><input type="checkbox" data-part="${k}" ${on ? 'checked' : ''}> ${l}</label>`).join('')}</div>
      <div class="alan"><select id="im-mode"><option value="merge">${t('ptModeMerge')}</option><option value="replace">${t('ptModeReplace')}</option></select></div></div>`,
    actions: [{ label: t('cancel'), cls: 'hayalet' }, { label: icon('check') + t('ptImport').replace('…', ''), cls: 'birincil', run: async (root) => {
      const src = $('#im-src', root).value;
      const req = { parts: $$('[data-part]', root).filter((c) => c.checked).map((c) => c.dataset.part), mode: $('#im-mode', root).value };
      if (src === 'json') req.json = $('#im-json', root).value; else req.from_project = src.slice(2);
      await api('POST', `/api/p/${enc(P)}/config/import`, req);
      toast(t('ptImported'));
      render(true);
    } }],
  });
  const sel = $('#im-src', m.root);
  sel.onchange = () => $('#im-json-box', m.root).classList.toggle('gizli', sel.value !== 'json');
  sel.onchange();
}

// Bağlantılar sekmesi: kütüphaneden ekle
async function connFromLibraryModal(P) {
  const lib = await GET('/api/library');
  const conns = lib.connections || [];
  const m = modal({
    title: `${icon('link')}${t('connFromLibTitle')}`,
    body: conns.length ? `<div class="liste-kutu">${conns.map((c) => `<div class="tikla pk-satir" data-n="${esc(c.name)}">${icon('server', 'acc-t')} <b>${esc(c.name)}</b> <span class="faint mono" style="font-size:12px">${esc(connURL(c))}</span></div>`).join('')}</div>`
      : `<p class="faint">${t('pushNoLib')}</p>`,
    actions: [{ label: t('close'), cls: 'hayalet' }],
  });
  $$('[data-n]', m.root).forEach((el) => (el.onclick = async () => {
    try { const r = await api('POST', `/api/p/${enc(P)}/connections/from-library`, { name: el.dataset.n }); toast(t('connAddedFromLib', r.name)); m.close(); render(true); } catch (e) { fail(e); }
  }));
}

// Sunucu kartı menüsü: tek seferlik include/exclude ile sync
function filteredSyncModal(P, s) {
  const proj = S.projects.find((x) => x.name === P) || {};
  const m = modal({
    title: `${icon('upload')}${t('syncOnceTitle')} · ${esc(s.name)}`, size: 'orta',
    body: `<p class="aciklama">${t('syncOnceDesc')}</p><div class="form" id="fs-form" data-p="${esc(P)}" data-ppath="${esc(proj.path || '')}" data-srv="${esc(s.name)}"><div class="alanlar">
      ${fArea('include', t('syncOnceInc'), [], { wide: false, rows: 5, tree: 'paths', chips: 'include' })}${fArea('exclude', t('syncOnceExc'), [], { wide: false, rows: 5, tree: 'paths', chips: 'exclude' })}</div>
      <div style="height:10px"></div>${fBool('dry_run', t('dryRun'), false, t('dryRunHint'))}</div>`,
    actions: [{ label: t('cancel'), cls: 'hayalet' }, { label: icon('upload') + t('sync'), cls: 'mavi', run: async (root) => {
      const f = readForm($('#fs-form', root), {});
      await startJobs(P, { kind: 'sync', servers: [s.name], include: f.include, exclude: f.exclude, dry_run: f.dry_run }, t('syncOnceTitle'));
    } }],
  });
  bindFormTools(m.root);
}

'use strict';

/* ═══════════════════════════════════════════════════════════════════════════
   Dosyalar → Editör modu (#/p/<proje>/edit/<profil>/<klasör>)
   Solda sunucudaki klasör ağacı (tıkladıkça yüklenir), sağda sekmeli editör. Her sekme
   createEditor (editor.js) — kaydetme, yedek, çakışma, "sunucu + yerel" seçimi aynı.
   Sekmeler bellekte yaşar (WS): sayfadan çıkıp dönünce kaydedilmemiş değişiklikler de durur;
   açık sekmeler + açık klasörler localStorage'da → panel yeniden açılınca sekmeler geri gelir.
   Kısayollar: Ctrl+S kaydet, Ctrl+W / Alt+W sekmeyi kapat, Ctrl+Tab / Ctrl+Shift+Tab gez.
   editor.js'ten sonra yüklenir.
   ═══════════════════════════════════════════════════════════════════════════ */

Object.assign(T.tr, {
  wsOpen: 'Editör', wsOpenHint: 'Editör modu: solda dosyalar, sağda sekmeli editör',
  wsList: 'Liste', wsListHint: 'Dosya listesine dön', wsRefresh: 'Ağacı yenile',
  wsEmpty: 'Soldan bir dosya seç', wsEmptyHint: 'Her dosya yeni bir sekmede açılır. Ctrl+S kaydet · Ctrl+W / Alt+W sekmeyi kapat · Ctrl+Tab sekmeler arası',
  wsClose: 'Kapat', wsCloseOthers: 'Diğerlerini kapat', wsCloseAll: 'Tümünü kapat', wsUnsavedN: '{0} sekmede kaydedilmemiş değişiklik var. Yine de kapatılsın mı?',
  wsReadonly: 'Bu dosya editörde açılamaz', wsLoadFail: 'Açılamadı', wsReload: 'Sunucudakini yeniden yükle',
  wsReloadAsk: 'Kaydedilmemiş değişiklikler gider. Sunucudaki hâli yüklensin mi?', wsRetry: 'Tekrar dene', wsSetRoot: 'Ağacı buradan başlat',
});
Object.assign(T.en, {
  wsOpen: 'Editor', wsOpenHint: 'Editor mode: files on the left, tabbed editor on the right',
  wsList: 'List', wsListHint: 'Back to the file list', wsRefresh: 'Refresh tree',
  wsEmpty: 'Pick a file on the left', wsEmptyHint: 'Each file opens in its own tab. Ctrl+S save · Ctrl+W / Alt+W close tab · Ctrl+Tab switch tabs',
  wsClose: 'Close', wsCloseOthers: 'Close others', wsCloseAll: 'Close all', wsUnsavedN: '{0} tab(s) have unsaved changes. Close anyway?',
  wsReadonly: 'This file can\'t be opened in the editor', wsLoadFail: 'Could not open', wsReload: 'Reload from the server',
  wsReloadAsk: 'Unsaved changes will be lost. Load the server version?', wsRetry: 'Try again', wsSetRoot: 'Start the tree here',
});

// Canlı çalışma alanları: "proje|profil" → { tabs: [{rel, host, ed, state, dirty}], active, open: Set, cache: Map }
const WS = new Map();
const wsStoreKey = (key) => `syncftp-ws:${key}`;
function wsLoadStored(key) {
  try { return JSON.parse(localStorage.getItem(wsStoreKey(key)) || 'null'); } catch { return null; }
}
function wsStore(w) {
  try { localStorage.setItem(wsStoreKey(w.key), JSON.stringify({ tabs: w.tabs.map((x) => x.rel), active: w.active, open: [...w.open] })); } catch { /* */ }
}
const wsBase = (p) => p.split('/').pop();
const wsDir = (p) => (p.includes('/') ? p.slice(0, p.lastIndexOf('/')) : '');
// Etkin çalışma alanı (Editör penceresi açık)
const wsLive = () => { const w = S.wsCur; return w && w.el && w.el.isConnected ? w : null; };

// Editör penceresini aç: Dosyalar'da bulunulan profil/bağlantı ve klasör (ağacın kökü o klasör)
async function openWorkspace(P, srvName, relDir) {
  let cfg;
  try { cfg = await getConfig(P); } catch (e) { return fail(e); }
  const conns = cfg.connections || [];
  if (!cfg.servers.length && !conns.length) return toast(t('noServers'), 'bad');
  const connObj = isConnTarget(srvName) ? conns.find((c) => CONN_TARGET + c.name === srvName) : null;
  const srvObj = connObj ? null : cfg.servers.find((s) => s.name === srvName) || cfg.servers.find((s) => s.enabled) || cfg.servers[0];
  const srv = srvObj ? srvObj.name : CONN_TARGET + (connObj || conns[0]).name;
  const key = `${P}|${srv}`;
  let w = WS.get(key);
  if (!w) {
    const st = wsLoadStored(key) || {};
    w = { key, P, srv, tabs: [], active: null, open: new Set(st.open || []), cache: new Map(), el: null };
    for (const rel of st.tabs || []) w.tabs.push({ rel, host: null, ed: null, state: 'new', dirty: false });
    w.active = w.tabs.some((x) => x.rel === st.active) ? st.active : (w.tabs[0] || {}).rel || null;
    WS.set(key, w);
  }
  // Ağacın kökü: Dosyalar'da bulunulan klasör (üst klasöre ağacın başlığından çıkılır)
  const dir = relDir || '';
  w.root = dir;

  const opts = [
    ...cfg.servers.map((s) => [s.name, s.name, t('tabServers')]),
    ...conns.map((c) => [CONN_TARGET + c.name, c.name, t('tabConnections')]),
  ];
  const groups = [...new Set(opts.map((o) => o[2]))];
  const select = `<select id="ws-srv" class="ws-srv" title="${t('tabServers')}">${groups.map((g) => `<optgroup label="${esc(g)}">${opts.filter((o) => o[2] === g).map(([v, l]) => `<option value="${esc(v)}" ${v === srv ? 'selected' : ''}>${esc(l)}</option>`).join('')}</optgroup>`).join('')}</select>`;

  // Kapatınca bir şey kaybolmaz: sekmeler (kaydedilmemiş değişiklikler dahil) bellekte kalır, tekrar açınca yerinde
  const m = modal({
    title: `${icon('code')}${t('wsOpen')} <span class="faint" style="font-weight:500">· ${esc(P)}</span>`, size: 'ws-modal',
    body: `<div class="ws-ust">${select}
      <span class="ws-ara"></span>
      <button class="btn kucuk hayalet" id="ws-refresh" title="${t('wsRefresh')}">${icon('refresh')}</button>
      <button class="btn kucuk" id="ws-save" title="Ctrl+S">${icon('check')}${t('edSave')}</button>
      <button class="btn kucuk" id="ws-list" title="${t('wsListHint')}">${icon('folder')}${t('wsList')}</button></div>
    <div class="ws" id="ws"><div class="ws-sol"><div class="ws-kok" id="ws-kok"></div><div class="ws-agac" id="ws-agac"></div></div>
      <div class="ws-sag"><div class="ws-sekmeler" id="ws-tabs"></div><div class="ws-govde" id="ws-body"></div></div></div>`,
    onClose: () => {
      // Editörler canlı kalır (DOM'dan ayrılır); tekrar açınca geri takılır
      for (const tab of w.tabs) if (tab.host) tab.host.remove();
      w.el = null;
      if (S.wsCur === w) S.wsCur = null;
    },
  });
  wsMount(m.root, w, m);
}

function wsMount(root, w, m) {
  const { P, srv } = w;
  w.el = $('#ws', root);
  S.wsCur = w;
  const tree = $('#ws-agac', root), tabsEl = $('#ws-tabs', root), body = $('#ws-body', root), kok = $('#ws-kok', root);
  w.body = body;
  // Önceki açılıştan kalan editörler (DOM'dan ayrılmış ama canlı) geri takılır
  for (const tab of w.tabs) if (tab.host) body.appendChild(tab.host);
  const empty = document.createElement('div');
  empty.className = 'ws-bos';
  empty.innerHTML = `${icon('code')}<b>${t('wsEmpty')}</b><small>${t('wsEmptyHint')}</small>`;
  body.appendChild(empty);

  // ── ağaç ──
  async function ls(path) {
    if (w.cache.has(path)) return w.cache.get(path);
    const lsKey = `${P}|${srv}|${path}`;
    let data = S.lsCache.get(lsKey);
    if (!data) { data = await GET(`/api/p/${enc(P)}/remote/ls?server=${enc(srv)}&path=${enc(path)}`); S.lsCache.set(lsKey, data); }
    const entries = (data.entries || []).slice().sort((a, b) => (a.type === 'dir') !== (b.type === 'dir') ? (a.type === 'dir' ? -1 : 1) : a.name.localeCompare(b.name));
    w.cache.set(path, entries);
    return entries;
  }
  function rows(path, depth) {
    const entries = w.cache.get(path);
    if (!entries) return w.open.has(path) || path === w.root ? `<div class="ws-satir ws-yuk" style="--d:${depth}">${icon('refresh', 'spin')}</div>` : '';
    if (entries.error) return `<div class="ws-satir ws-hata" style="--d:${depth}">${esc(entries.error)}</div>`;
    return entries.map((e) => {
      const p = (path ? path + '/' : '') + e.name;
      const isDir = e.type === 'dir', open = isDir && w.open.has(p);
      const skip = e.skip ? ` ws-soluk" title="${esc(t('skip_' + e.skip) || e.skip)}` : '';
      return `<div class="ws-satir${isDir ? ' ws-klasor' : ''}${!isDir && p === w.active ? ' aktif' : ''}${skip}" style="--d:${depth}" data-p="${esc(p)}" ${isDir ? 'data-dir' : ''}>
        ${isDir ? `<span class="ws-ok${open ? ' acik' : ''}">${icon('right')}</span>` : '<span class="ws-ok"></span>'}${icon(isDir ? 'folder' : 'file')}<span class="ws-ad">${esc(e.name)}</span></div>
        ${open ? rows(p, depth + 1) : ''}`;
    }).join('');
  }
  // Ağacın kökü w.root (Dosyalar'da bulunulan klasör); başlıkta yol + bir üst klasöre çık
  const drawKok = () => {
    const parts = w.root.split('/').filter(Boolean);
    kok.innerHTML = `<button class="ikon-btn" data-up title="${t('cmUp')}" ${w.root ? '' : 'disabled'}>${icon('left')}</button>
      <span class="ws-kok-yol" title="/${esc(w.root)}"><a href="javascript:void 0" data-kok="">/</a>${parts.map((x, i) => `${i ? '<span>/</span>' : ''}<a href="javascript:void 0" data-kok="${esc(parts.slice(0, i + 1).join('/'))}">${esc(x)}</a>`).join('')}</span>`;
  };
  const drawTree = () => { drawKok(); const st = tree.scrollTop; tree.innerHTML = rows(w.root, 0); tree.scrollTop = st; };
  const setRoot = (p) => { w.root = p; tree.scrollTop = 0; loadOpen().then(() => { if (w.active) markActive(w.active); }); };
  kok.onclick = (e) => {
    if (e.target.closest('[data-up]')) return setRoot(wsDir(w.root));
    const a = e.target.closest('[data-kok]');
    if (a && a.dataset.kok !== w.root) setRoot(a.dataset.kok);
  };
  async function ensure(path) {
    if (w.cache.has(path)) return;
    try { await ls(path); } catch (e) { w.cache.set(path, Object.assign([], { error: e.message })); }
  }
  async function loadOpen() {
    drawTree();
    await ensure(w.root);
    // Açık klasörler sığdan derine yüklenir (üstü kapalı olan atlanır)
    const open = [...w.open].sort((a, b) => a.split('/').length - b.split('/').length);
    for (const p of open) {
      const parent = wsDir(p);
      if (!(w.root === '' || p.startsWith(w.root + '/'))) continue; // kökün dışında
      if (parent !== w.root && !w.open.has(parent)) continue;
      drawTree();
      await ensure(p);
    }
    drawTree();
  }
  tree.onclick = async (e) => {
    const r = e.target.closest('.ws-satir[data-p]');
    if (!r) return;
    const p = r.dataset.p;
    if (r.hasAttribute('data-dir')) {
      if (w.open.has(p)) w.open.delete(p); else w.open.add(p);
      wsStore(w);
      drawTree();
      if (w.open.has(p)) { await ensure(p); drawTree(); }
    } else openTab(p);
  };

  // ── sekmeler ──
  const find = (rel) => w.tabs.find((x) => x.rel === rel);
  // Uzun süren yüklemeler (load) sayfadan çıkıp dönülünce hep güncel mount'u kullanır
  w.drawTabs = () => drawTabs();
  function drawTabs() {
    tabsEl.innerHTML = w.tabs.map((x) => `<div class="ws-sekme${x.rel === w.active ? ' aktif' : ''}" data-tab="${esc(x.rel)}" title="${esc(x.rel)}">
      ${icon(x.state === 'ro' ? 'eye' : 'file')}<span>${esc(wsBase(x.rel))}</span>${x.dirty ? `<span class="ed-nokta" title="${t('edModified')}"></span>` : ''}
      <button class="ws-x" data-x title="${t('wsClose')}">${icon('x')}</button></div>`).join('');
    const a = $('.ws-sekme.aktif', tabsEl);
    if (a) a.scrollIntoView({ block: 'nearest', inline: 'nearest' });
    empty.hidden = !!w.active;
    $('#ws-save', root).disabled = !w.active || !(find(w.active) || {}).ed;
  }
  tabsEl.onclick = (e) => {
    const tab = e.target.closest('.ws-sekme');
    if (!tab) return;
    if (e.target.closest('[data-x]')) closeTab(tab.dataset.tab); else activate(tab.dataset.tab);
  };
  tabsEl.onauxclick = (e) => { const tab = e.target.closest('.ws-sekme'); if (tab && e.button === 1) { e.preventDefault(); closeTab(tab.dataset.tab); } };
  tabsEl.onmousedown = (e) => { if (e.button === 1) e.preventDefault(); }; // orta tık kaydırma imleci açmasın

  async function load(tab) {
    if (!tab.host) { tab.host = document.createElement('div'); tab.host.className = 'ws-pano'; w.body.appendChild(tab.host); }
    tab.state = 'loading';
    tab.host.innerHTML = `<div class="yukleniyor">${icon('refresh', 'spin')} ${t('loading')}</div>`;
    let doc;
    try { doc = await GET(`/api/p/${enc(P)}/remote/edit?server=${enc(srv)}&path=${enc(tab.rel)}`); }
    catch (e) {
      tab.state = 'err';
      tab.host.innerHTML = `<div class="ws-bilgi"><div class="not bad">${icon('alert')}${t('wsLoadFail')}: ${esc(e.message)}</div>
        <button class="btn kucuk" data-retry>${icon('refresh')}${t('wsRetry')}</button></div>`;
      $('[data-retry]', tab.host).onclick = () => load(tab);
      return;
    }
    if (!doc.editable) {
      // Resim / PDF / binary / büyük dosya: sekmede bilgi + önizleme penceresi
      tab.state = 'ro';
      tab.host.innerHTML = `<div class="ws-bilgi"><div class="not">${icon('alert')}${t('wsReadonly')}: ${esc(edNotEditableReason(doc))}</div>
        <button class="btn" data-pv>${icon('eye')}${t('preview')}</button></div>`;
      $('[data-pv]', tab.host).onclick = () => previewFile(P, srv, tab.rel);
      w.drawTabs();
      return;
    }
    tab.state = 'ed';
    tab.ed = createEditor(tab.host, {
      P, srv, rel: tab.rel, doc,
      // Sayfadan çıkıp dönünce sekme çubuğu yeniden kurulur → hep güncel mount'un çizicisi (w.drawTabs)
      onDirty: (d) => { tab.dirty = d; w.drawTabs(); },
      onBusy: (b) => { const s = $('#ws-save'); if (s && tab.rel === w.active) s.disabled = b; },
    });
    w.drawTabs();
    if (tab.rel === w.active) tab.ed.focus();
  }

  function activate(rel) {
    const tab = find(rel);
    if (!tab) return;
    w.active = rel;
    for (const x of w.tabs) if (x.host) x.host.hidden = x.rel !== rel;
    if (tab.state === 'new') load(tab);
    else if (tab.ed) tab.ed.focus();
    markActive(rel);
    wsStore(w);
    drawTabs();
  }
  // Ağaçta etkin dosyayı vurgula; klasörü kapalıysa (kökün altındaysa) oraya kadar aç
  function markActive(rel) {
    $$('.ws-satir.aktif', tree).forEach((x) => x.classList.remove('aktif'));
    const mark = () => {
      const row = [...tree.querySelectorAll('.ws-satir[data-p]')].find((x) => x.dataset.p === rel);
      if (row) { row.classList.add('aktif'); row.scrollIntoView({ block: 'nearest' }); }
      return !!row;
    };
    if (mark() || !(w.root === '' || rel.startsWith(w.root + '/'))) return;
    const parts = wsDir(rel).split('/').filter(Boolean);
    parts.forEach((_, i) => { const d = parts.slice(0, i + 1).join('/'); if (w.root === '' || d.startsWith(w.root + '/')) w.open.add(d); });
    loadOpen().then(mark);
  }
  function openTab(rel) {
    if (!find(rel)) {
      // Yeni sekme etkin sekmenin hemen sağına
      const i = w.tabs.findIndex((x) => x.rel === w.active);
      w.tabs.splice(i < 0 ? w.tabs.length : i + 1, 0, { rel, host: null, ed: null, state: 'new', dirty: false });
    }
    activate(rel);
  }
  async function closeTab(rel, ask = true) {
    const tab = find(rel);
    if (!tab) return false;
    if (ask && tab.dirty && !(await edConfirmDiscard())) return false;
    const i = w.tabs.indexOf(tab);
    w.tabs.splice(i, 1);
    if (tab.host) tab.host.remove();
    if (w.active === rel) {
      const next = w.tabs[i] || w.tabs[i - 1];
      w.active = null;
      if (next) activate(next.rel);
    }
    wsStore(w);
    drawTabs();
    return true;
  }
  async function closeMany(list) {
    const dirtyN = list.filter((x) => x.dirty).length;
    if (dirtyN && !(await confirmBox(t('edDiscardTitle'), t('wsUnsavedN', dirtyN), t('edDiscard'), true))) return;
    for (const x of list.slice()) await closeTab(x.rel, false);
  }
  Object.assign(w, { find, openTab, closeTab, closeMany, setRoot, close: () => m.close() });
  w.save = () => { const tab = find(w.active); if (tab && tab.ed) tab.ed.save(); };
  w.closeActive = () => { if (w.active) closeTab(w.active); };
  w.cycle = (dir) => {
    if (w.tabs.length < 2) return;
    const i = w.tabs.findIndex((x) => x.rel === w.active);
    activate(w.tabs[(i + dir + w.tabs.length) % w.tabs.length].rel);
  };



  // ── araç çubuğu ──
  // Başka profil / bağlantı: pencere o hedefin kökünde yeniden açılır
  $('#ws-srv', root).onchange = (e) => { const v = e.target.value; m.close(); openWorkspace(P, v, ''); };
  $('#ws-save', root).onclick = () => w.save();
  $('#ws-refresh', root).onclick = () => {
    for (const k of [...S.lsCache.keys()]) if (k.startsWith(`${P}|${srv}|`)) S.lsCache.delete(k);
    w.cache.clear();
    loadOpen();
  };
  $('#ws-list', root).onclick = () => {
    const d = w.active ? wsDir(w.active) : w.root;
    m.close();
    go(`${pHash(P, 'files', srv)}${d ? '/' + enc(d) : ''}`);
  };

  drawTabs();
  for (const x of w.tabs) if (x.host) x.host.hidden = x.rel !== w.active;
  loadOpen().then(() => { if (w.active) activate(w.active); });
}

// Sağ tık menüleri: sekme ve ağaç satırı (bir kez kaydedilir, etkin çalışma alanına bakar)
CM.register('#ws-tabs .ws-sekme', (el) => {
  const w = wsLive(); if (!w) return null;
  const { find, closeTab, closeMany, P, srv } = w;
  const rel = el.dataset.tab, tab = find(rel);
  return [
    { heading: rel },
    { label: t('wsClose'), icon: 'x', hint: 'Ctrl+W', action: () => closeTab(rel) },
    { label: t('wsCloseOthers'), action: () => closeMany(w.tabs.filter((x) => x.rel !== rel)), disabled: w.tabs.length < 2 },
    { label: t('wsCloseAll'), action: () => closeMany(w.tabs) },
    '-',
    { label: t('wsReload'), icon: 'refresh', disabled: !tab || !tab.ed, action: async () => { if (!tab.dirty || await confirmBox(t('wsReload'), t('wsReloadAsk'), t('wsReload'), true)) tab.ed.reload().catch(fail); } },
    { label: t('preview'), icon: 'eye', action: () => previewFile(P, srv, rel) },
    { label: t('cmCopyRel'), icon: 'copy', action: () => CM.copy(rel) },
  ];
});
CM.register('#ws-agac .ws-satir[data-p]', (el) => {
  const w = wsLive(); if (!w) return null;
  const { openTab, P, srv } = w;
  const p = el.dataset.p, isDir = el.hasAttribute('data-dir');
  return [
    { heading: p },
    isDir ? { label: t('cmOpen'), icon: 'folder', action: () => el.click() } : { label: t('cmOpen'), icon: 'code', action: () => openTab(p) },
    !isDir && { label: t('preview'), icon: 'eye', action: () => previewFile(P, srv, p) },
    isDir && { label: t('wsSetRoot'), icon: 'folder', action: () => w.setRoot(p) },
    { label: t('wsList'), icon: 'folder', action: () => { w.close(); go(`${pHash(P, 'files', srv)}${(isDir ? p : wsDir(p)) ? '/' + enc(isDir ? p : wsDir(p)) : ''}`); } },
    '-',
    { label: t('cmCopyRel'), icon: 'copy', action: () => CM.copy(p) },
  ].filter(Boolean);
});


// Kısayollar — yalnızca Editör penceresi en üstteyken (üstünde kaydet / onay penceresi yokken)
document.addEventListener('keydown', (e) => {
  const w = wsLive();
  if (!w || $('#modal-kok').lastElementChild !== w.el.closest('.modal-arka')) return;
  // Yazarken Esc pencereyi kapatmasın (ağaçta / araç çubuğundayken Esc kapatır)
  if (e.key === 'Escape' && e.target.closest('.ed-alan')) { e.stopPropagation(); return; }
  const k = e.key.toLowerCase();
  const mod = e.ctrlKey || e.metaKey;
  if (mod && k === 's') {
    // Editörün kendi alanındaysa onun işleyicisi kaydeder
    if (!e.target.closest('.ed-alan')) { e.preventDefault(); w.save(); }
    return;
  }
  if ((mod && k === 'w') || (e.altKey && k === 'w')) { e.preventDefault(); e.stopPropagation(); w.closeActive(); return; }
  if (mod && (e.key === 'Tab' || e.key === 'PageDown' || e.key === 'PageUp')) {
    e.preventDefault(); e.stopPropagation();
    w.cycle(e.key === 'PageUp' || (e.key === 'Tab' && e.shiftKey) ? -1 : 1);
  }
}, true);

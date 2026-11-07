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
  wsReloadAsk: 'Kaydedilmemiş değişiklikler gider. Sunucudaki hâli yüklensin mi?', wsRetry: 'Tekrar dene', wsSetRoot: 'Ağacı buradan başlat', wsReadonlyShort: 'Yalnızca görüntüleme', wsCollapseAll: 'Tümünü daralt', wsPin: 'Sabitle', wsUnpin: 'Sabitlemeyi kaldır', wsCloseRight: 'Sağdakileri kapat', wsKeepPinned: 'sabitliler kalır', wsSearchPh: 'Bu klasörde ara…', wsSearchClear: 'Aramayı temizle',
});
Object.assign(T.en, {
  wsOpen: 'Editor', wsOpenHint: 'Editor mode: files on the left, tabbed editor on the right',
  wsList: 'List', wsListHint: 'Back to the file list', wsRefresh: 'Refresh tree',
  wsEmpty: 'Pick a file on the left', wsEmptyHint: 'Each file opens in its own tab. Ctrl+S save · Ctrl+W / Alt+W close tab · Ctrl+Tab switch tabs',
  wsClose: 'Close', wsCloseOthers: 'Close others', wsCloseAll: 'Close all', wsUnsavedN: '{0} tab(s) have unsaved changes. Close anyway?',
  wsReadonly: 'This file can\'t be opened in the editor', wsLoadFail: 'Could not open', wsReload: 'Reload from the server',
  wsReloadAsk: 'Unsaved changes will be lost. Load the server version?', wsRetry: 'Try again', wsSetRoot: 'Start the tree here', wsReadonlyShort: 'View only', wsCollapseAll: 'Collapse all', wsPin: 'Pin', wsUnpin: 'Unpin', wsCloseRight: 'Close tabs to the right', wsKeepPinned: 'pinned stay', wsSearchPh: 'Search this folder…', wsSearchClear: 'Clear search',
});

// Canlı çalışma alanları: "proje|profil" → { tabs: [{rel, host, ed, state, dirty}], active, open: Set, cache: Map }
const WS = new Map();
const wsStoreKey = (key) => `syncftp-ws:${key}`;
function wsLoadStored(key) {
  try { return JSON.parse(localStorage.getItem(wsStoreKey(key)) || 'null'); } catch { return null; }
}
function wsStore(w) {
  try { localStorage.setItem(wsStoreKey(w.key), JSON.stringify({ tabs: w.tabs.map((x) => x.rel), pinned: w.tabs.filter((x) => x.pinned).map((x) => x.rel), active: w.active, open: [...w.open] })); } catch { /* */ }
}
const wsBase = (p) => p.split('/').pop();
// Sabitli sekmeler hep başta (kendi aralarındaki sıra korunur)
const wsSortPinned = (w) => { w.tabs = [...w.tabs.filter((x) => x.pinned), ...w.tabs.filter((x) => !x.pinned)]; };
const WS_PIN_SVG = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 4h6l-1 6 3 3v2H7v-2l3-3-1-6zM12 15v6"/></svg>';
const WS_COLLAPSE_SVG = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M7 20l5-5 5 5M7 4l5 5 5-5"/></svg>';
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
    const pinned = new Set(st.pinned || []);
    for (const rel of st.tabs || []) w.tabs.push({ rel, host: null, ed: null, state: 'new', dirty: false, pinned: pinned.has(rel) });
    wsSortPinned(w);
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
    <div class="ws" id="ws"><div class="ws-sol"><div class="ws-kok" id="ws-kok"></div>
      <div class="ws-ara-kutu">${icon('search')}<input id="ws-q" type="search" placeholder="${esc(t('wsSearchPh'))}" autocomplete="off" spellcheck="false"><button class="ws-x" id="ws-q-x" title="${t('wsSearchClear')}" hidden>${icon('x')}</button></div>
      <div class="ws-agac" id="ws-agac"></div><div class="ws-sonuc" id="ws-sonuc" hidden></div></div>
      <div class="ws-sag"><div class="ws-sekmeler" id="ws-tabs"></div><div class="ws-govde" id="ws-body"></div></div></div>`,
    onClose: () => {
      // Editörler canlı kalır (DOM'dan ayrılır); tekrar açınca geri takılır
      if (w.stopSearch) w.stopSearch();
      for (const tab of w.tabs) if (tab.host) { stopMedia(tab.host); tab.host.remove(); }
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
  // Freeze listesi (yalnızca profilde; bağlantıyla gezerken sync yok)
  const frozenSelf = (p, dir) => {
    const F = w.frozen;
    if (!F || !F.size) return false;
    if (F.has(dir ? p + '/' : p)) return true;
    for (let i = p.indexOf('/'); i >= 0; i = p.indexOf('/', i + 1)) if (F.has(p.slice(0, i + 1))) return true;
    return false;
  };
  const frozenUnder = (p) => { const F = w.frozen; if (!F) return false; const pre = p + '/'; for (const f of F) if (f.startsWith(pre)) return true; return false; };
  const loadFrozen = async () => {
    if (isConnTarget(srv)) { w.frozen = new Set(); return; }
    w.frozen = new Set(await GET(`/api/p/${enc(P)}/frozen?server=${enc(srv)}`).catch(() => []));
  };
  function rows(path, depth) {
    const entries = w.cache.get(path);
    if (!entries) return w.open.has(path) || path === w.root ? `<div class="ws-satir ws-yuk" style="--d:${depth}">${icon('refresh', 'spin')}</div>` : '';
    if (entries.error) return `<div class="ws-satir ws-hata" style="--d:${depth}">${esc(entries.error)}</div>`;
    return entries.map((e) => {
      const p = (path ? path + '/' : '') + e.name;
      const isDir = e.type === 'dir', open = isDir && w.open.has(p);
      const skip = e.skip ? ' ws-soluk' : '';
      // Rozetler Dosyalar sekmesiyle aynı: freeze (kendisi / üst klasörü / klasörün altında kayıt) + ignore / exclude / include dışı
      const fz = frozenSelf(p, isDir) || (isDir && frozenUnder(p));
      const tags = `${fz ? `<span class="ws-rozet ws-rozet-fz" title="${esc(t('frozenBadge'))}">${icon('snow')}</span>` : ''}${e.skip ? `<span class="ws-rozet" title="${esc(t('skipHint_' + e.skip))}">${esc(t('skip_' + e.skip))}</span>` : ''}`;
      return `<div class="ws-satir${isDir ? ' ws-klasor' : ''}${!isDir && p === w.active ? ' aktif' : ''}${skip}" style="--d:${depth}" data-p="${esc(p)}" ${isDir ? 'data-dir' : ''}>
        ${isDir ? `<span class="ws-ok${open ? ' acik' : ''}">${icon('right')}</span>` : '<span class="ws-ok"></span>'}${icon(isDir ? 'folder' : 'file')}<span class="ws-ad">${esc(e.name)}</span>${tags}</div>
        ${open ? rows(p, depth + 1) : ''}`;
    }).join('');
  }
  // Ağacın kökü w.root (Dosyalar'da bulunulan klasör); başlıkta yol + bir üst klasöre çık
  const drawKok = () => {
    const parts = w.root.split('/').filter(Boolean);
    kok.innerHTML = `<button class="ikon-btn" data-up title="${t('cmUp')}" ${w.root ? '' : 'disabled'}>${icon('left')}</button>
      <button class="ikon-btn ws-daralt" data-collapse title="${t('wsCollapseAll')}" ${[...w.open].some((d) => w.root === '' || d.startsWith(w.root + '/')) ? '' : 'disabled'}>${WS_COLLAPSE_SVG}</button>
      <span class="ws-kok-yol" title="/${esc(w.root)}"><a href="javascript:void 0" data-kok="">/</a>${parts.map((x, i) => `${i ? '<span>/</span>' : ''}<a href="javascript:void 0" data-kok="${esc(parts.slice(0, i + 1).join('/'))}">${esc(x)}</a>`).join('')}</span>`;
  };
  const drawTree = () => { drawKok(); const st = tree.scrollTop; tree.innerHTML = rows(w.root, 0); tree.scrollTop = st; };
  const setRoot = (p) => { w.root = p; tree.scrollTop = 0; loadOpen().then(() => { if (w.active) markActive(w.active); }); };
  kok.onclick = (e) => {
    if (e.target.closest('[data-up]')) return setRoot(wsDir(w.root));
    if (e.target.closest('[data-collapse]')) {
      // Tümünü daralt: kökün altındaki açık klasörlerin hepsi kapanır
      for (const d of [...w.open]) if (w.root === '' || d.startsWith(w.root + '/')) w.open.delete(d);
      wsStore(w); tree.scrollTop = 0; drawTree();
      return;
    }
    const a = e.target.closest('[data-kok]');
    if (a && a.dataset.kok !== w.root) setRoot(a.dataset.kok);
  };
  async function ensure(path) {
    if (w.cache.has(path)) return;
    try { await ls(path); } catch (e) { w.cache.set(path, Object.assign([], { error: e.message })); }
  }
  async function loadOpen() {
    if (!w.frozen) await loadFrozen();
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

  // ── arama (bulunulan klasörün altındaki tüm alt klasörler; motor Dosyalar'daki "Alt klasörlerde ara" ile aynı) ──
  const qIn = $('#ws-q', root), qX = $('#ws-q-x', root), sonuc = $('#ws-sonuc', root);
  let es = null, qTimer = 0;
  const stopSearch = () => { if (es) { es.close(); es = null; } };
  const showTree = () => { stopSearch(); sonuc.hidden = true; sonuc.innerHTML = ''; tree.hidden = false; qX.hidden = !qIn.value; };
  function runSearch(term) {
    stopSearch();
    tree.hidden = true; sonuc.hidden = false; qX.hidden = false;
    let n = 0, dirs = 0, curPath = '', curSince = 0, done = false;
    const t0 = Date.now();
    sonuc.innerHTML = `<div class="ws-ara-durum"><span data-sp>${icon('refresh', 'spin')}</span><span data-st>${t('searching')}</span></div><div class="ws-ara-bekle mono" data-cur hidden></div><div data-list></div><div class="ws-ara-not" data-not hidden></div>`;
    const st = $('[data-st]', sonuc), list = $('[data-list]', sonuc), curEl = $('[data-cur]', sonuc), notEl = $('[data-not]', sonuc);
    const status = () => { st.textContent = done ? t('searchDoneN', n, dirs, ((Date.now() - t0) / 1000).toFixed(1)) : t('searchingN', n, dirs); };
    let buf = [], raf = 0;
    const flush = () => {
      raf = 0;
      if (!buf.length) return;
      const base = w.root ? w.root + '/' : '';
      list.insertAdjacentHTML('beforeend', buf.map((h) => {
        const shown = h.path.startsWith(base) ? h.path.slice(base.length) : h.path;
        return `<div class="ws-satir${h.dir ? ' ws-klasor' : ''}" style="--d:0" data-hit="${esc(h.path)}" ${h.dir ? 'data-dir' : ''} title="${esc(h.path)}"><span class="ws-ok"></span>${icon(h.dir ? 'folder' : 'file')}<span class="ws-ad">${esc(wsBase(shown))}</span><span class="ws-ara-yol">${esc(wsDir(shown))}</span></div>`;
      }).join(''));
      buf = [];
      status();
    };
    es = new EventSource(`/api/p/${enc(P)}/remote/search?server=${enc(srv)}&path=${enc(w.root)}&q=${enc(term)}`);
    es.addEventListener('hit', (e) => { n++; buf.push(JSON.parse(e.data)); if (!raf) raf = requestAnimationFrame(flush); });
    es.addEventListener('progress', (e) => {
      const d = JSON.parse(e.data);
      dirs = d.dirs; if (!raf) status();
      if (d.cur !== curPath) { curPath = d.cur || ''; curSince = Date.now(); }
      const show = curPath && Date.now() - curSince > 2000;
      curEl.hidden = !show;
      if (show) curEl.textContent = t('searchWaiting', curPath);
    });
    es.addEventListener('done', (e) => {
      const d = JSON.parse(e.data);
      stopSearch(); done = true; curEl.hidden = true;
      if (d.dirs) dirs = d.dirs;
      flush();
      $('[data-sp]', sonuc).innerHTML = d.error ? icon('alert', 'bad-t') : icon('check', 'ok-t');
      if (d.error) { st.textContent = d.error; st.className = 'bad-t'; return; }
      status();
      if (!n) list.innerHTML = `<div class="faint ws-ara-bos">${t('searchNone')}</div>`;
      const msgs = [];
      if (d.truncated) msgs.push(esc(t('searchTruncated')));
      if (d.skipped) msgs.push(`${esc(t('searchSkipped', d.skipped))}<br><span class="mono">${(d.skipped_dirs || []).map(esc).join('<br>')}</span>`);
      else if (d.warn) msgs.push(esc(d.warn));
      if (msgs.length) { notEl.hidden = false; notEl.innerHTML = msgs.join('<br><br>'); }
    });
    es.onerror = () => { if (es && !done) { stopSearch(); done = true; flush(); status(); } };
  }
  qIn.oninput = () => {
    clearTimeout(qTimer);
    const v = qIn.value.trim();
    qX.hidden = !qIn.value;
    if (!v) return showTree();
    if (v.length >= 2) qTimer = setTimeout(() => runSearch(v), 450); // yazma durunca
  };
  qIn.onkeydown = (e) => {
    if (e.key === 'Enter' && qIn.value.trim()) { clearTimeout(qTimer); runSearch(qIn.value.trim()); }
    if (e.key === 'Escape' && qIn.value) { e.stopPropagation(); qIn.value = ''; showTree(); }
  };
  qX.onclick = () => { qIn.value = ''; showTree(); qIn.focus(); };
  sonuc.onclick = (e) => {
    const r = e.target.closest('[data-hit]');
    if (!r) return;
    const p = r.dataset.hit;
    if (r.hasAttribute('data-dir')) { qIn.value = ''; showTree(); setRoot(p); }
    else openTab(p);
  };
  w.stopSearch = stopSearch;

  // ── sekmeler ──
  const find = (rel) => w.tabs.find((x) => x.rel === rel);
  // Uzun süren yüklemeler (load) sayfadan çıkıp dönülünce hep güncel mount'u kullanır
  w.drawTabs = () => drawTabs();
  function drawTabs() {
    tabsEl.innerHTML = w.tabs.map((x) => `<div class="ws-sekme${x.rel === w.active ? ' aktif' : ''}${x.pinned ? ' sabit' : ''}" data-tab="${esc(x.rel)}" title="${esc(x.rel)}">
      ${icon(x.state === 'ro' ? 'eye' : 'file')}<span>${esc(wsBase(x.rel))}</span>${x.dirty ? `<span class="ed-nokta" title="${t('edModified')}"></span>` : ''}
      ${x.pinned ? `<button class="ws-x ws-pin" data-unpin title="${t('wsUnpin')}">${WS_PIN_SVG}</button>` : `<button class="ws-x" data-x title="${t('wsClose')}">${icon('x')}</button>`}</div>`).join('');
    const a = $('.ws-sekme.aktif', tabsEl);
    if (a) a.scrollIntoView({ block: 'nearest', inline: 'nearest' });
    empty.hidden = !!w.active;
    $('#ws-save', root).disabled = !w.active || !(find(w.active) || {}).ed;
  }
  let dragJustEnded = false;
  tabsEl.onclick = (e) => {
    const tab = e.target.closest('.ws-sekme');
    if (!tab || dragJustEnded) return;
    if (e.target.closest('[data-unpin]')) setPinned(tab.dataset.tab, false);
    else if (e.target.closest('[data-x]')) closeTab(tab.dataset.tab);
    else activate(tab.dataset.tab);
  };
  // Orta tık kapatır (sabitli sekme hariç — VS Code gibi, yanlışlıkla kapanmasın)
  tabsEl.onauxclick = (e) => { const tab = e.target.closest('.ws-sekme'); if (tab && e.button === 1) { e.preventDefault(); if (!(find(tab.dataset.tab) || {}).pinned) closeTab(tab.dataset.tab); } };
  // Tekerlek sekme çubuğunu yatay kaydırır
  tabsEl.onwheel = (e) => { if (tabsEl.scrollWidth > tabsEl.clientWidth && Math.abs(e.deltaY) > Math.abs(e.deltaX)) { e.preventDefault(); tabsEl.scrollLeft += e.deltaY; } };
  // Sürükleyerek sırala: fareyle tut, 5 px'ten fazla kayınca sekme imleci takip eder, diğerleri yer
  // değiştirir; bırakınca sıra kaydedilir. Sabitliler kendi grubunda kalır.
  tabsEl.onmousedown = (e) => {
    if (e.button === 1) { e.preventDefault(); return; } // orta tık kaydırma imleci açmasın
    if (e.button !== 0 || e.target.closest('button')) return;
    const el = e.target.closest('.ws-sekme');
    if (!el) return;
    const startX = e.clientX;
    let dragging = false;
    const onMove = (ev) => {
      if (!dragging) {
        if (Math.abs(ev.clientX - startX) < 5) return;
        dragging = true;
        el.classList.add('surukle');
        document.body.classList.add('ws-sekme-surukle');
      }
      const pinned = el.classList.contains('sabit');
      const sibs = [...tabsEl.children].filter((x) => x !== el && x.classList.contains('sabit') === pinned);
      // İmlecin solundaki son kardeşin arkasına
      let before = null;
      for (const x of sibs) { const r = x.getBoundingClientRect(); if (ev.clientX < r.left + r.width / 2) { before = x; break; } }
      if (!before) {
        const last = sibs[sibs.length - 1];
        before = last ? last.nextElementSibling : (pinned ? tabsEl.querySelector('.ws-sekme:not(.sabit)') : null);
        if (before === el) before = el.nextElementSibling;
      }
      if (before !== el && el.nextElementSibling !== before) tabsEl.insertBefore(el, before);
      // kenara yaklaşınca çubuk kayar
      const r = tabsEl.getBoundingClientRect();
      if (ev.clientX < r.left + 30) tabsEl.scrollLeft -= 12; else if (ev.clientX > r.right - 30) tabsEl.scrollLeft += 12;
    };
    const onUp = () => {
      document.removeEventListener('mousemove', onMove);
      document.removeEventListener('mouseup', onUp);
      if (!dragging) return;
      el.classList.remove('surukle');
      document.body.classList.remove('ws-sekme-surukle');
      const order = [...tabsEl.children].map((x) => x.dataset.tab);
      w.tabs.sort((a, b) => order.indexOf(a.rel) - order.indexOf(b.rel));
      wsStore(w);
      dragJustEnded = true; setTimeout(() => { dragJustEnded = false; }, 0); // bırakmanın click'i sekmeyi değiştirmesin
      activate(el.dataset.tab);
    };
    document.addEventListener('mousemove', onMove);
    document.addEventListener('mouseup', onUp);
  };
  function setPinned(rel, on) {
    const tab = find(rel);
    if (!tab) return;
    tab.pinned = on;
    // Sabitlenen sabitlilerin sonuna, çözülen sabitlenmemişlerin başına
    w.tabs.splice(w.tabs.indexOf(tab), 1);
    const at = w.tabs.filter((x) => x.pinned).length;
    w.tabs.splice(at, 0, tab);
    wsStore(w);
    drawTabs();
  }
  // Etkin sekmeyi bir sağa / sola taşı (Ctrl+Shift+PgUp/PgDn), kendi grubunda
  function moveActive(dir) {
    const i = w.tabs.findIndex((x) => x.rel === w.active);
    const j = i + dir;
    if (i < 0 || j < 0 || j >= w.tabs.length || w.tabs[j].pinned !== w.tabs[i].pinned) return;
    [w.tabs[i], w.tabs[j]] = [w.tabs[j], w.tabs[i]];
    wsStore(w);
    drawTabs();
  }

  async function load(tab) {
    if (!tab.host) { tab.host = document.createElement('div'); tab.host.className = 'ws-pano'; w.body.appendChild(tab.host); }
    tab.state = 'loading';
    tab.host.innerHTML = `<div class="yukleniyor">${icon('refresh', 'spin')} ${t('loading')}</div>`;
    // Resim / video / ses / PDF: düzenleme yok, sekmenin içinde doğrudan önizleme
    const kind = mediaKind(tab.rel);
    if (kind) {
      tab.state = 'ro';
      tab.host.innerHTML = `<div class="ws-onizle">${mediaView(kind, `/api/p/${enc(P)}/remote/media?server=${enc(srv)}&path=${enc(tab.rel)}&v=${Date.now()}`)}</div>`;
      w.drawTabs();
      return;
    }
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
      // Düzenlenemeyen dosya sekmede önizlenir: metinse (büyük / UTF-8 değil) salt okunur kod görünümü,
      // binary ise bilgi + indir
      tab.state = 'ro';
      const note = `<div class="ws-ro-not">${icon('eye')}${t('wsReadonlyShort')}: ${esc(edNotEditableReason(doc))}</div>`;
      let pv = null;
      if (doc.reason !== 'binary') pv = await GET(`/api/p/${enc(P)}/remote/cat?server=${enc(srv)}&path=${enc(tab.rel)}`).catch(() => null);
      tab.host.innerHTML = pv && !pv.binary
        ? `${note}<div class="ws-onizle ws-kod">${fmtCodeView(pv.content, tab.rel)}</div>`
        : `${note}<div class="ws-bilgi"><div class="bos">${t('binaryFile')}</div>
            <a class="btn" href="/api/p/${enc(P)}/remote/get?server=${enc(srv)}&path=${enc(tab.rel)}" download>${icon('download')}${t('download')}</a></div>`;
      if (pv && !pv.binary) bindFmt(tab.host);
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
    for (const x of w.tabs) if (x.host) { const hide = x.rel !== rel; if (hide && !x.host.hidden) $$('video,audio', x.host).forEach((m) => m.pause()); x.host.hidden = hide; }
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
      w.tabs.splice(Math.max(i + 1, w.tabs.filter((x) => x.pinned).length), 0, { rel, host: null, ed: null, state: 'new', dirty: false, pinned: false });
    }
    activate(rel);
  }
  async function closeTab(rel, ask = true) {
    const tab = find(rel);
    if (!tab) return false;
    if (ask && tab.dirty && !(await edConfirmDiscard())) return false;
    const i = w.tabs.indexOf(tab);
    w.tabs.splice(i, 1);
    if (tab.host) { stopMedia(tab.host); tab.host.remove(); }
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
  Object.assign(w, { find, openTab, closeTab, closeMany, setRoot, setPinned, moveActive, close: () => m.close() });
  w.save = () => { const tab = find(w.active); if (tab && tab.ed) tab.ed.save(); };
  w.closeActive = () => { const tab = find(w.active); if (tab && !tab.pinned) closeTab(w.active); };
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
    w.frozen = null;
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
    { label: tab && tab.pinned ? t('wsUnpin') : t('wsPin'), icon: 'lock', action: () => w.setPinned(rel, !(tab && tab.pinned)) },
    '-',
    { label: t('wsClose'), icon: 'x', hint: tab && tab.pinned ? '' : 'Ctrl+W', action: () => closeTab(rel) },
    { label: t('wsCloseOthers'), action: () => closeMany(w.tabs.filter((x) => x.rel !== rel && !x.pinned)), disabled: !w.tabs.some((x) => x.rel !== rel && !x.pinned) },
    { label: t('wsCloseRight'), action: () => closeMany(w.tabs.slice(w.tabs.findIndex((x) => x.rel === rel) + 1).filter((x) => !x.pinned)), disabled: !w.tabs.slice(w.tabs.findIndex((x) => x.rel === rel) + 1).some((x) => !x.pinned) },
    { label: t('wsCloseAll'), hint: w.tabs.some((x) => x.pinned) ? t('wsKeepPinned') : '', action: () => closeMany(w.tabs.filter((x) => !x.pinned)) },
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
  if (mod && e.shiftKey && (e.key === 'PageDown' || e.key === 'PageUp')) {
    e.preventDefault(); e.stopPropagation();
    w.moveActive(e.key === 'PageUp' ? -1 : 1);
    return;
  }
  if (mod && (e.key === 'Tab' || e.key === 'PageDown' || e.key === 'PageUp')) {
    e.preventDefault(); e.stopPropagation();
    w.cycle(e.key === 'PageUp' || (e.key === 'Tab' && e.shiftKey) ? -1 : 1);
  }
}, true);

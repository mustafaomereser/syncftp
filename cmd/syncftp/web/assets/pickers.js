'use strict';

/* ═══════════════════════════════════════════════════════════════════════════
   Panel içi seçiciler (native diyalog yok — odak/pencere sorunu yaşanmaz)
   pickPath({mode:'dir'|'file', initial, title})       → Promise<string|null>
   pickLocalTree({P, server, current, title, mode})    → Promise<string[]|null>
     mode: 'paths' (include/exclude/protect — klasör ya da dosya) | 'files' (freeze — klasör "klasör/" olarak, altındaki her şey)
   app.js'ten sonra yüklenir; T / modal / api / icon / esc yardımcılarını kullanır.
   ═══════════════════════════════════════════════════════════════════════════ */

Object.assign(T.tr, {
  pkTitleDir: 'Klasör seç', pkTitleFile: 'Dosya seç', pkDrives: 'Sürücüler', pkHome: 'Ev dizini', pkUp: 'Üst klasör',
  pkSelectThis: 'Bu klasörü seç', pkRemoteTitle: 'Sunucuda klasör seç', pkNoSubdirs: 'Alt klasör yok', pkSelect: 'Seç', pkGo: 'Git', pkEmpty: 'Klasör boş', pkProject: 'syncFTP projesi',
  pkExplorerHint: 'Klasöre tıklayınca içine girer, .. ile geri çıkar. Onay kutusuyla işaretleyin. Klavye: ↑↓ gez, Enter gir, Space işaretle, Backspace geri.', pkRoot: 'Proje kökü',
  pkHidden: 'Gizlileri göster', pkTreeTitle: 'Yerel dosyalardan seç', pkIgnored: 'ignore', pkSelected: '{0} seçili',
  
  pkFreezeHint: 'Klasör işaretlenirse klasörün kendisi freeze\'lenir: altındaki her dosya (sonradan eklenenler dahil) yüklenmez.',
  pkNewFolder: 'Yeni klasör', pkNewFile: 'Yeni dosya', pkNewFolderPh: 'Klasör adı', pkNewFilePh: 'Dosya adı (ör. index.php)', pkCreate: 'Oluştur', pkCreated: '{0} oluşturuldu',
});
Object.assign(T.en, {
  pkTitleDir: 'Choose folder', pkTitleFile: 'Choose file', pkDrives: 'Drives', pkHome: 'Home', pkUp: 'Parent folder',
  pkSelectThis: 'Select this folder', pkRemoteTitle: 'Choose a server folder', pkNoSubdirs: 'No subfolders', pkSelect: 'Select', pkGo: 'Go', pkEmpty: 'Folder is empty', pkProject: 'syncFTP project',
  pkExplorerHint: 'Click a folder to open it, .. to go back. Tick checkboxes to select. Keys: ↑↓ move, Enter open, Space select, Backspace back.', pkRoot: 'Project root',
  pkHidden: 'Show hidden', pkTreeTitle: 'Pick from local files', pkIgnored: 'ignored', pkSelected: '{0} selected',
  
  pkFreezeHint: 'Checking a folder freezes the folder itself: nothing inside it (including files added later) is uploaded.',
  pkNewFolder: 'New folder', pkNewFile: 'New file', pkNewFolderPh: 'Folder name', pkNewFilePh: 'File name (e.g. index.php)', pkCreate: 'Create', pkCreated: '{0} created',
});

// ── Seçicilerde "Yeni klasör / Yeni dosya" ───────────────────────────────────
// Araç çubuğunda iki düğme; tıklayınca listenin üstünde satır içi ad kutusu açılır (Enter oluşturur,
// Esc kapatır — Esc pencereyi kapatmaz). create(kind, name) sunucuya / diske yazar ve listeyi tazeler;
// hata kutunun yanında kalır, ad düzeltilip tekrar denenir. kinds: ['dir'] ya da ['dir','file'].
const pkNewButtons = (kinds = ['dir', 'file']) => kinds.map((k) =>
  `<button class="btn kucuk" data-pknew="${k}" title="${t(k === 'dir' ? 'pkNewFolder' : 'pkNewFile')}">${icon('plus')}${icon(k === 'dir' ? 'folder' : 'file')}<span class="pk-yeni-et">${t(k === 'dir' ? 'pkNewFolder' : 'pkNewFile')}</span></button>`).join('');
const pkNewRow = () => `<div class="pk-yeni" data-pknewrow hidden><span data-pknewicon></span>
  <input type="text" spellcheck="false" autocomplete="off" data-pknewname>
  <button class="btn kucuk birincil" data-pknewok>${t('pkCreate')}</button>
  <button class="ikon-btn" data-pknewx title="${t('cancel')}">${icon('x')}</button></div>`;
function bindPkNew(root, create) {
  const row = $('[data-pknewrow]', root), inp = $('[data-pknewname]', root), ok = $('[data-pknewok]', root);
  if (!row) return { hide() {} };
  let kind = 'dir';
  const hide = () => { row.hidden = true; inp.value = ''; };
  const submit = async () => {
    const name = inp.value.trim();
    if (!name) { inp.focus(); return; }
    ok.disabled = true;
    try { await create(kind, name); hide(); toast(t('pkCreated', name)); }
    catch (e) { toast(e.message, 'bad'); inp.focus(); inp.select(); }
    finally { ok.disabled = false; }
  };
  $$('[data-pknew]', root).forEach((b) => (b.onclick = () => {
    kind = b.dataset.pknew;
    $('[data-pknewicon]', root).innerHTML = icon(kind === 'dir' ? 'folder' : 'file', 'acc-t');
    inp.placeholder = t(kind === 'dir' ? 'pkNewFolderPh' : 'pkNewFilePh');
    row.hidden = false; inp.focus();
  }));
  ok.onclick = submit;
  $('[data-pknewx]', root).onclick = hide;
  // Pencerenin Esc / tek tuş kısayolları bu kutuda çalışmasın
  inp.onkeydown = (e) => {
    e.stopPropagation();
    if (e.key === 'Enter') { e.preventDefault(); submit(); } else if (e.key === 'Escape') { e.preventDefault(); hide(); }
  };
  return { hide };
}

// ── Dosya sistemi seçici ─────────────────────────────────────────────────────
function pickPath({ mode = 'dir', initial = '', title } = {}) {
  return new Promise((resolve) => {
    // Boşsa panelin açıldığı dizinden başla (CLI'daki gibi: çift tıkta exe'nin klasörü)
    let cur = initial || (S.meta && S.meta.start_dir) || '';
    let chosenFile = '';
    let hidden = false;
    let done = false;
    let home = '', sep = '\\';
    const m = modal({
      title: `${icon(mode === 'dir' ? 'folder' : 'file')}${title || (mode === 'dir' ? t('pkTitleDir') : t('pkTitleFile'))}`,
      size: 'orta',
      body: `<div class="girdi-grup" style="margin-bottom:10px">
          <button class="ikon-btn" data-pk="up" title="${t('pkUp')}">${icon('left')}</button>
          <button class="ikon-btn" data-pk="drives" title="${t('pkDrives')}">${icon('server')}</button>
          <button class="ikon-btn" data-pk="home" title="${t('pkHome')}">${icon('layers')}</button>
          <input type="text" class="mono" data-pk="path" spellcheck="false">
          <button class="btn" data-pk="go">${t('pkGo')}</button></div>
        <div class="pk-arac" data-pk="newbar">${pkNewButtons(mode === 'dir' ? ['dir'] : ['dir', 'file'])}</div>
        ${pkNewRow()}
        <div class="liste-kutu pk-liste" data-pk="list" style="height:46vh;max-height:none"></div>
        <label style="display:flex;gap:8px;align-items:center;margin-top:10px;font-size:13px;color:var(--muted)">
          <input type="checkbox" data-pk="hidden"> ${t('pkHidden')}</label>`,
      actions: [
        { left: '<span data-pk="info"></span>' },
        { label: t('cancel'), cls: 'hayalet', run: () => { done = true; resolve(null); } },
        { label: mode === 'dir' ? icon('check') + t('pkSelectThis') : icon('check') + t('pkSelect'), cls: 'birincil', run: () => {
          const v = mode === 'dir' ? cur : chosenFile;
          if (!v) return false;
          done = true; resolve(v);
        } },
      ],
      onClose: () => { if (!done) resolve(null); },
    });
    const $p = (k) => $(`[data-pk="${k}"]`, m.root);
    const join = (a, b) => (a.endsWith(sep) ? a + b : a + sep + b);

    // Klasör yoksa / okunamıyorsa üst klasörlere, en sonda sürücü listesine düşer (boş liste kalmasın).
    // quiet: ilk açılışta kayıtlı yol silinmiş olabilir — sessizce en yakın var olan üst klasör açılır.
    async function load(path, quiet = false) {
      let d;
      try { d = await GET(`/api/fs/list?path=${enc(path)}&files=${mode === 'file' ? 1 : 0}&hidden=${hidden ? 1 : 0}`); }
      catch (e) {
        if (!path) { toast(e.message, 'bad'); return; }
        if (!quiet) toast(e.message, 'bad');
        const up = path.replace(/[\\/]+$/, '').replace(/[\\/][^\\/]*$/, '');
        return load(up && up !== path ? up : '', true);
      }
      cur = d.path; home = d.home || home; sep = d.sep || sep; chosenFile = '';
      $p('path').value = cur;
      $p('newbar').hidden = !cur; // sürücü listesinde oluşturulamaz
      $p('info').innerHTML = d.has_config ? `<span class="pill ok"><span class="nokta"></span>${t('pkProject')}</span>` : '';
      const list = $p('list');
      if (!d.entries.length) { list.innerHTML = `<div class="faint" style="padding:10px 12px">${t('pkEmpty')}</div>`; return; }
      list.innerHTML = d.entries.map((e) => `<div class="tikla pk-satir" data-name="${esc(e.name)}" data-dir="${e.dir ? 1 : 0}">
        ${icon(e.dir ? 'folder' : 'file', e.dir ? 'acc-t' : 'faint')} ${esc(e.name)}${e.dir ? '' : ` <span class="faint">${bytes(e.size)}</span>`}</div>`).join('');
      $$('.pk-satir', list).forEach((row) => {
        const full = cur ? join(cur, row.dataset.name) : row.dataset.name;
        row.onclick = () => {
          if (row.dataset.dir === '1') { load(full); return; }
          $$('.pk-satir', list).forEach((r) => r.classList.remove('secili'));
          row.classList.add('secili');
          chosenFile = full;
          $p('path').value = full;
        };
        row.ondblclick = () => {
          if (row.dataset.dir === '1') return;
          chosenFile = full; done = true; resolve(full); m.close();
        };
      });
    }
    // Yeni klasör → içine girilir (klasör seçerken genelde yeni klasörü seçmek istenir); yeni dosya → seçili gelir
    bindPkNew(m.root, async (kind, name) => {
      const r = await api('POST', '/api/fs/create', { root: cur, path: '', kind, name });
      if (kind === 'dir') { await load(r.path); return; }
      await load(cur);
      const row = $$('.pk-satir', $p('list')).find((x) => x.dataset.name === r.name);
      if (row) { row.click(); row.scrollIntoView({ block: 'nearest' }); }
    });
    $p('up').onclick = () => GET(`/api/fs/list?path=${enc(cur)}`).then((d) => load(d.parent || '')).catch(() => load(''));
    $p('drives').onclick = () => load('');
    $p('home').onclick = () => load(home);
    $p('go').onclick = () => load($p('path').value.trim());
    $p('path').onkeydown = (e) => { if (e.key === 'Enter') load($p('path').value.trim()); };
    $p('hidden').onchange = (e) => { hidden = e.target.checked; load(cur); };
    load(cur, true).then(() => { if (!cur) GET('/api/fs/list?path=').then((d) => { home = d.home; }); });
  });
}

// Uzak klasör seçici — sunucu formundaki "Uzak dizin → Gözat", proje ekleme, dosya taşıma.
// list(path, create) → {path, dirs, files?}; create ({kind, name}) verilirse çağıran önce onu
// sunucuda oluşturur sonra listeler. Klasöre tıkla → içine gir, ".." → üst; dosyalar yalnızca görünür.
function pickRemoteDir({ list, initial = '/', title }) {
  return new Promise((resolve) => {
    let cur = initial || '/', done = false, busy = 0;
    const m = modal({
      title: `${icon('server')}${title || t('pkRemoteTitle')}`, size: 'orta',
      body: `<div class="girdi-grup" style="margin-bottom:10px">
          <button class="ikon-btn" data-rp="up" title="${t('pkUp')}">${icon('left')}</button>
          <input type="text" class="mono" data-rp="path" spellcheck="false">
          <button class="btn" data-rp="go">${t('pkGo')}</button></div>
        <div class="pk-arac">${pkNewButtons()}</div>
        ${pkNewRow()}
        <div class="liste-kutu pk-liste" data-rp="list" style="height:46vh;max-height:none"></div>`,
      actions: [
        { label: t('cancel'), cls: 'hayalet', run: () => { done = true; resolve(null); } },
        { label: icon('check') + t('pkSelectThis'), cls: 'birincil', run: () => { done = true; resolve(cur); } },
      ],
      onClose: () => { if (!done) resolve(null); },
    });
    const $r = (k) => $(`[data-rp="${k}"]`, m.root);
    const parent = (p) => (p === '/' ? '/' : p.replace(/\/[^/]*$/, '') || '/');
    const child = (n) => (cur === '/' ? '' : cur) + '/' + n;
    function draw(d, mark = '') {
      cur = d.path; $r('path').value = cur;
      const rows = (cur === '/' ? [] : [['..', parent(cur)]]).concat(d.dirs.map((n) => [n, child(n)]));
      const files = d.files || [];
      $r('list').innerHTML = rows.length || files.length
        ? rows.map(([n, full]) => `<div class="tikla pk-satir${n === mark ? ' secili' : ''}" data-full="${esc(full)}">${icon(n === '..' ? 'left' : 'folder', n === '..' ? 'faint' : 'acc-t')} ${esc(n)}</div>`).join('')
          + files.map((n) => `<div class="pk-satir pk-dosya${n === mark ? ' secili' : ''}">${icon('file', 'faint')} <span class="faint">${esc(n)}</span></div>`).join('')
        : `<div class="faint" style="padding:10px 12px">${t('pkNoSubdirs')}</div>`;
      $$('.pk-satir[data-full]', $r('list')).forEach((row) => (row.onclick = () => load(row.dataset.full)));
      const sel = $('.pk-satir.secili', $r('list'));
      if (sel) sel.scrollIntoView({ block: 'nearest' });
    }
    async function load(path) {
      const seq = ++busy;
      $r('list').innerHTML = `<div class="faint" style="padding:10px 12px">${icon('refresh', 'spin')} ${t('loading')}</div>`;
      let d;
      try { d = await list(path); } catch (e) {
        if (seq !== busy) return;
        $r('list').innerHTML = `<div class="bad-t" style="padding:10px 12px">${esc(e.message)}</div>`;
        return;
      }
      if (seq !== busy) return;
      draw(d);
    }
    // Yeni klasör → içine girilir; yeni dosya → listede işaretli görünür
    bindPkNew(m.root, async (kind, name) => {
      const d = await list(cur, { kind, name });
      ++busy;
      if (kind === 'dir') await load(child(d.created || name));
      else draw(d, d.created || name);
    });
    $r('up').onclick = () => load(parent(cur));
    $r('go').onclick = () => load($r('path').value.trim() || '/');
    $r('path').onkeydown = (e) => { if (e.key === 'Enter') load($r('path').value.trim() || '/'); };
    load(cur);
  });
}

// Proje köküne göreceli yol ("." = kök); proje dışındaysa mutlak yol olduğu gibi döner.
function relToProject(projectPath, abs) {
  const norm = (s) => s.replace(/[\\/]+$/, '').toLowerCase();
  const a = abs.replace(/[\\/]+$/, ''), root = projectPath.replace(/[\\/]+$/, '');
  if (norm(a) === norm(root)) return '.';
  if (norm(a).startsWith(norm(root) + '\\') || norm(a).startsWith(norm(root) + '/')) return a.slice(root.length + 1).replace(/\\/g, '/');
  return abs;
}

// ── Yerel gezgin seçici (CLI'daki localBrowser gibi) ─────────────────────────
// Klasöre tıkla / Enter → içine gir; ".." / Backspace / ← → üst klasör; onay kutusu / Space → işaretle.
// Üstte yol çubuğu (tıklanabilir), altta seçilenler (✕ ile çıkarılır). Listede olup ağaçta karşılığı
// olmayan kalıplar (glob vb.) korunur.
// root verilirse proje yerine o mutlak klasör gezilir (hızlı yükleme / varsayılanlar).
// mode: 'paths' (include/exclude) | 'protect' (klasörlere / eklenir) | 'files' (freeze: klasör "klasör/" olarak)
function pickLocalTree({ P, root = '', server = '', current = [], title, mode = 'paths' }) {
  return new Promise((resolve) => {
    const sel = new Set((current || []).map((x) => x.replace(/\/+$/, '')).filter(Boolean));
    const dirMarks = new Set((current || []).filter((x) => x.endsWith('/')).map((x) => x.replace(/\/+$/, '')));
    let cwd = '', rows = [], cursor = 0, done = false, baseName = root || '';
    const m = modal({
      title: `${icon('folder')}${title || t('pkTreeTitle')}`, size: 'orta',
      body: `<p class="ipucu" style="margin:0 0 8px">${mode === 'files' ? t('pkFreezeHint') : t('pkExplorerHint')}</p>
        <div class="pk-arac pk-arac-yol"><div class="gz-yol" data-crumbs></div>${pkNewButtons()}</div>
        ${pkNewRow()}
        <div class="liste-kutu gezgin" data-list tabindex="0" style="height:44vh;max-height:none"></div>
        <div class="gz-secili" data-picked></div>`,
      actions: [
        { left: '<span data-count></span>' },
        { label: t('cancel'), cls: 'hayalet', run: () => { done = true; resolve(null); } },
        { label: icon('check') + t('save'), cls: 'birincil', run: async () => {
          // protect ve freeze ('files'): klasör "klasör/" olarak kaydedilir (altındaki her şey)
          const out = [...sel].map((p) => (dirMarks.has(p) && mode !== 'paths' ? p + '/' : p));
          done = true; resolve(out.sort());
        } },
      ],
      onClose: () => { if (!done) resolve(null); },
    });
    const list = $('[data-list]', m.root), crumbs = $('[data-crumbs]', m.root), picked = $('[data-picked]', m.root);
    const join = (a, b) => (a ? a + '/' + b : b);

    function drawPicked() {
      $('[data-count]', m.root).textContent = t('pkSelected', sel.size);
      picked.innerHTML = sel.size ? [...sel].sort().map((p) => `<span class="gz-cip" data-p="${esc(p)}">${icon(dirMarks.has(p) ? 'folder' : /[*?{]/.test(p) ? 'search' : 'file')}${esc(p)}<button data-unpick title="${t('remove')}">${icon('x')}</button></span>`).join('') : '';
      $$('[data-unpick]', picked).forEach((b) => (b.onclick = () => {
        const p = b.closest('[data-p]').dataset.p;
        sel.delete(p); dirMarks.delete(p); drawPicked(); drawRows();
      }));
    }
    function toggle(r) {
      if (!r || r.up) return;
      if (sel.has(r.path)) { sel.delete(r.path); dirMarks.delete(r.path); }
      else { sel.add(r.path); if (r.dir) dirMarks.add(r.path); }
      drawPicked(); drawRows();
    }
    function drawRows() {
      list.innerHTML = rows.map((r, i) => r.up
        ? `<div class="gz-satir ${i === cursor ? 'imlec' : ''}" data-i="${i}"><span class="gz-kutu"></span>${icon('left', 'faint')}<span class="ad">..</span></div>`
        : `<div class="gz-satir ${i === cursor ? 'imlec' : ''} ${sel.has(r.path) ? 'secili' : ''}" data-i="${i}">
            <input type="checkbox" class="gz-kutu" ${sel.has(r.path) ? 'checked' : ''}>
            ${icon(r.dir ? 'folder' : 'file', r.dir ? 'acc-t' : 'faint')}
            <span class="ad ${r.ignored ? 'faint' : ''}">${esc(r.name)}${r.dir ? '/' : ''}</span>
            ${r.ignored ? `<span class="pill" style="height:18px;font-size:11px">${t('pkIgnored')}</span>` : ''}
            <span class="sag faint">${r.dir ? icon('right') : bytes(r.size)}</span></div>`).join('')
        || `<div class="faint" style="padding:10px 12px">${t('pkEmpty')}</div>`;
      $$('.gz-satir', list).forEach((el) => {
        const r = rows[+el.dataset.i];
        const cb = $('input', el);
        if (cb) cb.onclick = (e) => { e.stopPropagation(); cursor = +el.dataset.i; toggle(r); };
        el.onclick = () => { cursor = +el.dataset.i; if (r.up) open(parentOf(cwd)); else if (r.dir) open(r.path); else toggle(r); };
      });
      const cur = $('.imlec', list);
      if (cur) cur.scrollIntoView({ block: 'nearest' });
    }
    const parentOf = (p) => (p.includes('/') ? p.slice(0, p.lastIndexOf('/')) : '');
    function drawCrumbs() {
      const parts = cwd ? cwd.split('/') : [];
      const rootLabel = root ? root.split(/[\\/]/).filter(Boolean).pop() : t('pkRoot');
      crumbs.innerHTML = `<button data-c="">${icon('folder')}${esc(rootLabel)}</button>` +
        parts.map((p, i) => `<span class="faint">/</span><button data-c="${esc(parts.slice(0, i + 1).join('/'))}">${esc(p)}</button>`).join('');
      $$('[data-c]', crumbs).forEach((b) => (b.onclick = () => open(b.dataset.c)));
    }
    async function open(rel, focusName) {
      let d;
      try { d = await GET(root ? `/api/fs/tree?root=${enc(root)}&path=${enc(rel)}` : `/api/p/${enc(P)}/local/ls?server=${enc(server)}&path=${enc(rel)}`); }
      catch (e) { toast(e.message, 'bad'); return; }
      const prev = cwd;
      cwd = rel;
      rows = (rel ? [{ up: true }] : []).concat(d.entries.map((e) => ({ ...e, path: join(rel, e.name) })));
      cursor = 0;
      // Yukarı çıkınca geldiğimiz klasör imleçte kalsın (CLI gibi)
      const back = focusName || (prev.startsWith(rel ? rel + '/' : '') && prev !== rel ? prev.slice(rel ? rel.length + 1 : 0).split('/')[0] : '');
      if (back) { const i = rows.findIndex((r) => r.name === back); if (i >= 0) cursor = i; }
      drawCrumbs(); drawRows();
      list.focus();
    }
    // Ok tuşları listeyi yeniden çizmez — yalnızca imleç sınıfı taşınır (büyük klasörde her tuşta tam çizim ağırdı)
    function moveCursor(i) {
      if (!rows.length) return;
      cursor = Math.max(0, Math.min(rows.length - 1, i));
      const old = $('.imlec', list);
      if (old) old.classList.remove('imlec');
      const el = $(`.gz-satir[data-i="${cursor}"]`, list);
      if (el) { el.classList.add('imlec'); el.scrollIntoView({ block: 'nearest' }); }
    }
    list.onkeydown = (e) => {
      const r = rows[cursor];
      if (e.key === 'ArrowDown') moveCursor(cursor + 1);
      else if (e.key === 'ArrowUp') moveCursor(cursor - 1);
      else if (e.key === 'Enter' || e.key === 'ArrowRight') { if (r && r.up) open(parentOf(cwd)); else if (r && r.dir) open(r.path); else if (e.key === 'Enter') toggle(r); }
      else if (e.key === ' ') toggle(r);
      else if (e.key === 'Backspace' || e.key === 'ArrowLeft') { if (cwd) open(parentOf(cwd)); }
      else return;
      e.preventDefault();
    };
    // Yeni öğe bulunulan klasörde oluşturulur, imleç ona gelir (seçmek için Space / onay kutusu)
    bindPkNew(m.root, async (kind, name) => {
      const r = root
        ? await api('POST', '/api/fs/create', { root, path: cwd, kind, name })
        : await api('POST', `/api/p/${enc(P)}/local/create?server=${enc(server)}&path=${enc(cwd)}`, { kind, name });
      await open(cwd, r.name);
    });
    drawPicked();
    open('');
  });
}

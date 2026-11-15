// Tüm projeler → "Değişiklikler" görünümü: her projenin her aktif profilinde sync'in yükleyeceği dosyalar
// tek sayfada. Veri proje sayfasındakiyle aynı uçlardan (overview?scan=0 + status), dosyaya tıklayınca
// var olan karşılaştırma penceresi, "Ayrıntı" var olan durum penceresi açılır.
Object.assign(T.tr, {
  viewChanges: 'Değişiklikler', chSummary: '{0} projede {1} dosya bekliyor', chScanning: 'Projeler taranıyor…', chNone: 'Bütün projeler güncel',
  chUpToDate: 'Güncel: {0}', chSyncAll: 'Hepsini sync', chSyncAllConfirm: '{0} projede bekleyen değişiklikler sunuculara yüklenecek ({1} dosya). Devam edilsin mi?',
  chBadgeTip: '{0} dosya sync bekliyor', chMore: '+{0} dosya daha', chDetails: 'Ayrıntı', chSyncProject: 'Projeyi sync', chOpen: 'Projeyi aç', chNoProfiles: 'Aktif profil yok',
});
Object.assign(T.en, {
  viewChanges: 'Changes', chSummary: '{1} file(s) pending in {0} project(s)', chScanning: 'Scanning projects…', chNone: 'All projects are up to date',
  chUpToDate: 'Up to date: {0}', chSyncAll: 'Sync all', chSyncAllConfirm: 'Pending changes in {0} project(s) will be uploaded ({1} files). Continue?',
  chBadgeTip: '{0} file(s) waiting to sync', chMore: '+{0} more files', chDetails: 'Details', chSyncProject: 'Sync project', chOpen: 'Open project', chNoProfiles: 'No enabled profile',
});

// Proje kartlarındaki bekleyen dosya rozeti: proje sayfasındaki sayının aynısı (aktif profillerin yeni +
// değişen dosyaları). Son bilinen değer anında çizilir, arka planda tazelenir (aynı anda 3 proje).
const PENDING = new Map();
function pendingBadgeHTML(P) {
  const n = PENDING.get(P);
  return `<span class="pend-rozet" data-pend="${esc(P)}">${n ? pendingBadgeInner(n) : ''}</span>`;
}
const pendingBadgeInner = (n) => `<span class="pill warn" title="${esc(t('chBadgeTip', n))}">${icon('upload')}${n}</span>`;
function loadPendingBadges(root) {
  const seq = S.renderSeq;
  const queue = (S.projects || []).filter((p) => p.ok && p.enabled).map((p) => p.name);
  const worker = async () => {
    for (let P; (P = queue.shift());) {
      let n = 0;
      try {
        const ov = await GET(`/api/p/${enc(P)}/overview`);
        n = (ov.servers || []).filter((s) => s.enabled && !s.scan_error).reduce((a, s) => a + s.new + s.changed, 0);
      } catch { continue; }
      if (seq !== S.renderSeq) return;
      PENDING.set(P, n);
      $$(`[data-pend="${CSS.escape(P)}"]`, root).forEach((el) => (el.innerHTML = n ? pendingBadgeInner(n) : ''));
    }
  };
  for (let i = 0; i < 3; i++) worker();
}

const CH_SHOW = 40; // profil başına listede gösterilen dosya (fazlası "Ayrıntı" penceresinde)

function homeViewSwitch(view) {
  const b = (v, ic, l) => `<a role="button" data-hv="${v}" class="${view === v ? 'aktif' : ''}" title="${l}">${icon(ic)}${l}</a>`;
  return `<span class="seg srv-seg">${b('grid', 'layers', t('viewGrid'))}${b('list', 'log', t('viewList'))}${b('changes', 'upload', t('viewChanges'))}</span>`;
}

async function pageChanges(ps) {
  const projs = projectsByGroup(ps).filter((p) => p.ok && p.enabled);
  const seq = S.renderSeq;
  const res = new Map(); // proje → [{server, st}]
  bind((root) => {
    $$('[data-add]', root).forEach((b) => (b.onclick = addProjectFlow));
    $$('[data-hv]', root).forEach((b) => (b.onclick = () => { lsSet('syncftp-home-view', b.dataset.hv); render(true); }));
    $('#ch-yenile', root).onclick = () => render(true);
    const summary = () => {
      const pend = [...res.entries()].map(([P, list]) => [P, list.reduce((a, x) => a + x.st.counts.new + x.st.counts.changed, 0)]).filter(([, n]) => n);
      const total = pend.reduce((a, [, n]) => a + n, 0);
      const done = res.size === projs.length;
      $('#ch-ozet', root).innerHTML = !done ? `${icon('refresh', 'spin')} ${t('chScanning')} <span class="faint">${res.size} / ${projs.length}</span>`
        : total ? `<b class="warn-t">${esc(t('chSummary', pend.length, total))}</b>` : `<b class="ok-t">${icon('check')} ${t('chNone')}</b>`;
      const all = $('#ch-hepsi', root);
      all.hidden = !done || !total;
      all.onclick = async () => {
        if (!(await confirmBox(t('chSyncAll'), esc(t('chSyncAllConfirm', pend.length, total)), t('chSyncAll')))) return;
        for (const [P] of pend) {
          const servers = res.get(P).filter((x) => x.st.counts.new + x.st.counts.changed).map((x) => x.server);
          startJobs(P, { kind: 'sync', servers }, t('sync')).catch(fail);
        }
      };
      const ok = [...res.entries()].filter(([, list]) => !list.some((x) => x.st.counts.new + x.st.counts.changed + x.st.counts.deleted)).map(([P]) => P);
      $('#ch-guncel', root).innerHTML = done && ok.length ? `${icon('check', 'ok-t')} ${esc(t('chUpToDate', ok.join(', ')))}` : '';
    };
    const draw = (P, list, err) => {
      const sec = $(`.ch-proje[data-p="${CSS.escape(P)}"]`, root);
      if (!sec) return;
      const n = list.reduce((a, x) => a + x.st.counts.new + x.st.counts.changed + x.st.counts.deleted, 0);
      if (!err && !n) { sec.remove(); return; } // güncel projeler altta tek satırda
      sec.querySelector('.ch-say').innerHTML = err ? '' : `<span class="pill warn">${list.reduce((a, x) => a + x.st.counts.new + x.st.counts.changed, 0)}</span>`;
      sec.querySelector('.ch-ic').innerHTML = err ? `<div class="not bad">${icon('alert')}${esc(err)}</div>` : list.filter((x) => x.st.counts.new + x.st.counts.changed + x.st.counts.deleted).map(({ server, st }) => {
        const items = [...st.new.map((f) => ['+', f, true]), ...st.changed.map((f) => ['~', f, true]), ...st.deleted.map((f) => ['−', f, false])].sort((a, b) => a[1].localeCompare(b[1]));
        const c = st.counts, rest = c.new + c.changed + c.deleted - Math.min(items.length, CH_SHOW);
        return `<div class="ch-profil" data-s="${esc(server)}"><div class="ch-profil-bas">${icon('server')}<b>${esc(server)}</b>
            <span class="faint mono">${t('pendingChip', c.new, c.changed, c.deleted)}</span>${st.bytes ? `<span class="faint">· ${bytes(st.bytes)}</span>` : ''}
            <span style="margin-left:auto"></span><button class="btn kucuk hayalet" data-ch="details">${t('chDetails')}</button>
            ${c.new + c.changed ? `<button class="btn kucuk mavi" data-ch="sync">${icon('upload')}${t('sync')}</button>` : ''}</div>
          <div class="liste-kutu ch-liste">${items.slice(0, CH_SHOW).map(([m, f, click]) => `<div ${click ? `class="tikla" data-diff="${esc(f)}"` : 'class="faint"'}><span class="mono ch-isaret ch-${m === '+' ? 'yeni' : m === '~' ? 'deg' : 'sil'}">${m}</span> ${esc(f)}</div>`).join('')}
            ${rest > 0 ? `<div class="tikla faint" data-ch="details">${t('chMore', rest)}</div>` : ''}</div></div>`;
      }).join('');
      sec.querySelector('[data-ch="psync"]').hidden = !!err || !list.some((x) => x.st.counts.new + x.st.counts.changed);
      $$('.ch-profil', sec).forEach((pr) => {
        const s = pr.dataset.s;
        $$('[data-diff]', pr).forEach((d) => (d.onclick = () => compareModal(P, s, d.dataset.diff)));
        $$('[data-ch="details"]', pr).forEach((b) => (b.onclick = () => statusModal(P, { name: s })));
        const sb = pr.querySelector('[data-ch="sync"]');
        if (sb) sb.onclick = () => startJobs(P, { kind: 'sync', servers: [s] }, t('sync')).catch(fail);
      });
    };
    $$('.ch-proje', root).forEach((sec) => {
      const P = sec.dataset.p;
      sec.querySelector('[data-ch="open"]').onclick = () => go(pHash(P));
      sec.querySelector('[data-ch="psync"]').onclick = () => {
        const servers = (res.get(P) || []).filter((x) => x.st.counts.new + x.st.counts.changed).map((x) => x.server);
        if (servers.length) startJobs(P, { kind: 'sync', servers }, t('sync')).catch(fail);
      };
    });
    // Tarama yerel diski okur: aynı anda en fazla 3 proje (büyük projeler diski boğmasın)
    const queue = [...projs];
    const worker = async () => {
      for (let p; (p = queue.shift());) {
        let list = [], err = '';
        try {
          const ov = await GET(`/api/p/${enc(p.name)}/overview?scan=0`);
          const servers = (ov.servers || []).filter((s) => s.enabled);
          list = await Promise.all(servers.map(async (s) => ({ server: s.name, st: await GET(`/api/p/${enc(p.name)}/status?server=${enc(s.name)}`) })));
        } catch (e) { err = e.message; }
        if (seq !== S.renderSeq) return;
        res.set(p.name, list);
        draw(p.name, list, err);
        summary();
      }
    };
    summary();
    for (let i = 0; i < 3; i++) worker();
  });
  return `<div class="sayfa"><div class="baslik-satir"><div><h1>${t('projectsTitle')}</h1><p>${esc(t('projectsSub', (S.meta && S.meta.base) || ''))}</p></div>
    <div class="aksiyonlar">${homeViewSwitch('changes')}<button class="btn birincil" data-add>${icon('plus')}${t('addProject')}</button></div></div>
    <div class="aksiyonlar ch-ust"><span id="ch-ozet"></span><span style="margin-left:auto"></span>
      <button class="ikon-btn" id="ch-yenile" title="${t('refresh')}">${icon('refresh')}</button>
      <button class="btn mavi" id="ch-hepsi" hidden>${icon('upload')}${t('chSyncAll')}</button></div>
    ${projs.map((p) => `<section class="ch-proje" data-p="${esc(p.name)}" ${p.color ? `style="--pc:${esc(p.color)}"` : ''}>
      <div class="ch-proje-bas"><span class="p-ikon" ${p.color ? `style="color:${esc(p.color)}"` : ''}>${icon('folder')}</span><b>${esc(p.name)}</b><span class="ch-say">${icon('refresh', 'spin')}</span>
        <span style="margin-left:auto"></span><button class="btn kucuk hayalet" data-ch="open">${t('chOpen')}</button>
        <button class="btn kucuk mavi" data-ch="psync" hidden>${icon('upload')}${t('chSyncProject')}</button></div>
      <div class="ch-ic"></div></section>`).join('')}
    <p class="faint ch-guncel" id="ch-guncel"></p></div>`;
}

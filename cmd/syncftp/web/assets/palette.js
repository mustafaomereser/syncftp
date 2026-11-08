// Hızlı arama (Ctrl+K): sayfalar, projeler, proje sekmeleri ve proje eylemleri tek kutuda.
// Kendi pencere barımızdaki "Ara…" kutusu da bunu açar (titlebar.js); Edge penceresinde de çalışır.
(() => {
  let box = null;

  const fold = (s) => String(s || '').toLocaleLowerCase(LANG === 'tr' ? 'tr' : 'en')
    .replace(/[ıi̇]/g, 'i').replace(/ş/g, 's').replace(/ğ/g, 'g').replace(/ü/g, 'u').replace(/ö/g, 'o').replace(/ç/g, 'c');

  // Boşlukla ayrılan her kelime ayrı eşleşmeli ("medya geçmiş"); toplam puan. -1 = eşleşmedi.
  function score(text, q) {
    if (!q) return 0;
    let total = 0;
    for (const tok of q.split(/\s+/).filter(Boolean)) {
      const s = scoreTok(fold(text), tok);
      if (s < 0) return -1;
      total += s;
    }
    return total;
  }
  // Tam alt dize önce (kelime başında daha iyi); yoksa sırayla geçen harfler — ama harfler çok
  // dağınıksa (yayılma > 3 × uzunluk) eşleşme sayılmaz ("dry" → "History" gelmesin).
  function scoreTok(t, q) {
    const idx = t.indexOf(q);
    if (idx >= 0) return 1000 - idx - (idx === 0 || /[\s·/:-]/.test(t[idx - 1]) ? 0 : 50);
    let s = 0, pos = -1, run = 0, first = -1;
    for (const ch of q) {
      const i = t.indexOf(ch, pos + 1);
      if (i < 0) return -1;
      if (first < 0) first = i;
      run = i === pos + 1 ? run + 1 : 0;
      s += 10 + run * 5 - Math.min(i - pos, 10);
      pos = i;
    }
    return pos - first + 1 > q.length * 3 ? -1 : s;
  }

  function commands() {
    const out = [
      { group: t('pkGroupPages'), icon: 'layers', label: t('projectsTitle'), run: () => go('#/') },
      { group: t('pkGroupPages'), icon: 'zap', label: t('cmJobs'), run: () => go('#/jobs') },
      { group: t('pkGroupPages'), icon: 'sliders', label: t('cmSettings'), run: () => go('#/settings') },
      { group: t('pkGroupPages'), icon: 'help', label: t('cmHelp'), run: () => go('#/help') },
      { group: t('pkGroupPages'), icon: 'terminal', label: 'Terminal', run: () => go('#/terminal') },
      { group: t('pkGroupPages'), icon: 'plus', label: t('addProject'), run: () => addProjectFlow() },
      { group: t('pkGroupPages'), icon: 'zap', label: t('qcTitle'), hint: t('qcHint'), run: () => openQuickConnect() },
    ];
    const tabs = [['', 'layers', 'tabOverview'], ['files', 'folder', 'tabFiles'], ['history', 'history', 'tabHistory'], ['servers', 'server', 'tabServers'],
      ['connections', 'link', 'tabConnections'], ['cicd', 'zap', 'tabCicd'], ['settings', 'sliders', 'tabSettings']];
    const cur = parseRoute().project;
    // Açık proje önce
    const projs = [...(S.projects || [])].sort((a, b) => (b.name === cur) - (a.name === cur));
    for (const p of projs) {
      const P = p.name;
      out.push({ group: t('pkGroupProjects'), icon: p.ok ? 'folder' : 'alert', label: P, hint: p.path, run: () => go(pHash(P)) });
      for (const [tab, ic, key] of tabs.slice(1)) out.push({ group: t('pkGroupProjects'), icon: ic, label: `${P} · ${t(key)}`, run: () => go(pHash(P, tab)) });
      out.push({ group: t('pkGroupProjects'), icon: 'code', label: `${P} · ${t('tabFiles')} · ${t('wsOpen')}`, run: () => openWorkspace(P) });
      if (p.ok && p.enabled) {
        out.push({ group: t('pkGroupActions'), icon: 'upload', label: `${P} · ${t('syncAll')}`, run: () => startJobs(P, { kind: 'sync' }, t('sync')) });
        out.push({ group: t('pkGroupActions'), icon: 'eye', label: `${P} · ${t('dryRun')}`, run: () => startJobs(P, { kind: 'sync', dry_run: true }, t('dryRun')) });
        out.push(p.watching
          ? { group: t('pkGroupActions'), icon: 'stop', label: `${P} · ${t('watchStop')}`, run: () => watchStop(P) }
          : { group: t('pkGroupActions'), icon: 'play', label: `${P} · ${t('watchStart')}`, run: () => watchStart(P, []) });
      }
      out.push({ group: t('pkGroupActions'), icon: 'terminal', label: `${P} · ${t('cmTerminal')}`, run: () => go('#/terminal/' + enc(P)) });
    }
    return out;
  }

  function close() {
    if (!box) return;
    box.remove();
    box = null;
  }

  function open() {
    if (box) { box.querySelector('input').select(); return; }
    const all = commands();
    box = document.createElement('div');
    box.className = 'pk-arka';
    box.innerHTML = `<div class="pk" role="dialog" aria-label="${esc(t('pkTitle'))}">
      <div class="pk-ara">${icon('search')}<input type="text" placeholder="${esc(t('pkPlaceholder'))}" spellcheck="false" autocomplete="off"><kbd>Esc</kbd></div>
      <div class="pk-liste" role="listbox"></div>
      <div class="pk-alt"><span><kbd>↑</kbd><kbd>↓</kbd> ${t('pkMove')}</span><span><kbd>Enter</kbd> ${t('pkOpen')}</span></div></div>`;
    document.body.appendChild(box);
    const input = box.querySelector('input'), list = box.querySelector('.pk-liste');
    let shown = [], sel = 0;
    const draw = () => {
      const q = fold(input.value.trim());
      shown = q
        ? all.map((c) => ({ c, s: Math.max(score(c.label, q), score(c.hint || '', q) - 200) })).filter((x) => x.s >= 0).sort((a, b) => b.s - a.s).slice(0, 60).map((x) => x.c)
        : all.filter((c) => c.group !== t('pkGroupProjects') || !c.label.includes(' · ')).slice(0, 40);
      sel = Math.min(sel, Math.max(0, shown.length - 1));
      let last = '';
      list.innerHTML = shown.length ? shown.map((c, i) => {
        const head = !q && c.group !== last ? `<div class="pk-grup">${esc((last = c.group))}</div>` : '';
        return `${head}<div class="pk-oge ${i === sel ? 'aktif' : ''}" data-i="${i}" role="option">${icon(c.icon)}<span class="pk-etiket">${esc(c.label)}</span>${c.hint ? `<span class="pk-ipucu">${esc(c.hint)}</span>` : ''}</div>`;
      }).join('') : `<div class="pk-bos">${t('pkEmpty')}</div>`;
      const a = list.querySelector('.aktif');
      if (a) a.scrollIntoView({ block: 'nearest' });
    };
    const run = (i) => {
      const c = shown[i];
      if (!c) return;
      close();
      try { const r = c.run(); if (r && r.catch) r.catch(fail); } catch (e) { fail(e); }
    };
    input.addEventListener('input', () => { sel = 0; draw(); });
    input.addEventListener('keydown', (e) => {
      if (e.key === 'ArrowDown') { e.preventDefault(); sel = (sel + 1) % Math.max(1, shown.length); draw(); }
      else if (e.key === 'ArrowUp') { e.preventDefault(); sel = (sel - 1 + shown.length) % Math.max(1, shown.length); draw(); }
      else if (e.key === 'Enter') { e.preventDefault(); run(sel); }
      else if (e.key === 'Escape') { e.preventDefault(); e.stopPropagation(); close(); }
    });
    list.addEventListener('mousemove', (e) => { const o = e.target.closest('.pk-oge'); if (o && +o.dataset.i !== sel) { sel = +o.dataset.i; list.querySelectorAll('.pk-oge').forEach((x) => x.classList.toggle('aktif', x === o)); } });
    list.addEventListener('click', (e) => { const o = e.target.closest('.pk-oge'); if (o) run(+o.dataset.i); });
    box.addEventListener('mousedown', (e) => { if (e.target === box) close(); });
    draw();
    input.focus();
  }

  document.addEventListener('keydown', (e) => {
    if ((e.ctrlKey || e.metaKey) && !e.altKey && !e.shiftKey && e.key.toLowerCase() === 'k') { e.preventDefault(); if (box) close(); else open(); }
  }, true);
  window.addEventListener('hashchange', close);
  window.openPalette = open;
})();

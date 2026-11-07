// Panelin sağ tık menüsü (acomod panelindeki context-menu.js'in vanilla uyarlaması).
//
// Kayıt:  CM.register(selector, (el, ctx) => items | null | false)
//   null / []  → burada özel bir şey yok, yukarı doğru aramaya devam (sonunda sayfa menüsü)
//   false      → geri çağırma işi kendisi halletti; tarayıcı menüsü de açılmaz
// Öğe: { label, icon, action(ctx), disabled (true | 'neden'), hint, danger, children: [...] }
//   '-' ayraç, { heading } küçük başlık, { note } alt not.
// ctx: { target, event, selection, close }
//
// Tarayıcı menüsü şunlarda kalır: input / textarea / select / xterm, Shift (ya da Ctrl/Alt) + sağ tık,
// video / ses / iframe, data-cm="none". Seçili metin varsa "Kopyala". Klavye: Menü tuşu / Shift+F10,
// oklar, Home/End, Enter/Boşluk, Esc, harfle atlama.
(() => {
  const NATIVE = 'input, textarea, select, option, [contenteditable=""], [contenteditable="true"], .xterm, iframe, video, audio, embed, object, [data-cm="none"]';
  const registry = [];
  const itemOf = new WeakMap();
  let state = null; // { stack, ctx, focusBefore, keyboard }
  let hoverTimer = 0;
  let suppressUntil = 0;

  const textOf = (el) => (el ? el.textContent.replace(/\s+/g, ' ').trim() : '');
  const isNative = (el) => !el || !el.closest || !!el.closest(NATIVE);
  const inMenu = (el) => !!(el && el.closest && el.closest('.cm-menu'));
  const selectedText = () => { const s = window.getSelection && window.getSelection(); return s && !s.isCollapsed ? s.toString().trim() : ''; };

  function copyText(text, what) {
    const done = () => toast(t('cmCopied', what || t('cmText')));
    if (navigator.clipboard && window.isSecureContext) return navigator.clipboard.writeText(text).then(done, () => toast(t('cmCopyFail'), 'bad'));
    const a = document.createElement('textarea');
    a.value = text; a.style.cssText = 'position:fixed;top:-1000px;opacity:0';
    document.body.appendChild(a); a.select();
    let ok = false;
    try { ok = document.execCommand('copy'); } catch { /* */ }
    a.remove();
    ok ? done() : toast(t('cmCopyFail'), 'bad');
  }

  function normalize(items) {
    const out = [];
    for (const it of items || []) {
      if (!it) continue;
      if (it === '-' || it.divider) { if (out.length && !out[out.length - 1].divider) out.push({ divider: true }); continue; }
      out.push(it);
    }
    while (out.length && out[0].divider) out.shift();
    while (out.length && out[out.length - 1].divider) out.pop();
    return out;
  }

  function build(items, level) {
    const menu = document.createElement('div');
    menu.className = 'cm-menu';
    menu.setAttribute('role', 'menu');
    menu.tabIndex = -1;
    menu.dataset.level = level;
    for (const it of normalize(items)) {
      if (it.divider) { const s = document.createElement('div'); s.className = 'cm-sep'; menu.appendChild(s); continue; }
      if (it.heading != null) { const h = document.createElement('div'); h.className = 'cm-heading'; h.textContent = it.heading; menu.appendChild(h); continue; }
      if (it.note != null) { const n = document.createElement('div'); n.className = 'cm-note'; n.innerHTML = it.note; menu.appendChild(n); continue; }
      const kids = it.children && normalize(it.children).length ? it.children : null;
      const el = document.createElement('button');
      el.type = 'button';
      el.className = 'cm-item' + (it.danger ? ' tehlike' : '');
      el.setAttribute('role', 'menuitem');
      el.tabIndex = -1;
      if (it.disabled) {
        el.setAttribute('aria-disabled', 'true');
        if (typeof it.disabled === 'string') el.title = it.disabled; // tooltip.js balona çevirir
      }
      if (kids) el.setAttribute('aria-haspopup', 'menu');
      el.innerHTML = `<span class="cm-ikon">${it.icon ? icon(it.icon) : ''}</span><span class="cm-etiket"></span>${it.hint ? `<span class="cm-ipucu"></span>` : ''}${kids ? `<span class="cm-ok">${icon('right')}</span>` : ''}`;
      el.querySelector('.cm-etiket').textContent = it.label;
      if (it.hint) el.querySelector('.cm-ipucu').textContent = it.hint;
      itemOf.set(el, { ...it, children: kids });
      menu.appendChild(el);
    }
    menu.addEventListener('click', (e) => { const el = e.target.closest('.cm-item'); if (el) { e.preventDefault(); activate(el, false); } });
    menu.addEventListener('mousemove', onHover);
    menu.addEventListener('contextmenu', (e) => e.preventDefault());
    return menu;
  }

  function place(menu, x, y) {
    const vw = document.documentElement.clientWidth, vh = window.innerHeight;
    const w = menu.offsetWidth, h = menu.offsetHeight;
    let left = x, top = y;
    if (left + w > vw - 8) left = x - w >= 8 ? x - w : Math.max(8, vw - w - 8);
    if (top + h > vh - 8) top = y - h >= 8 ? y - h : Math.max(8, vh - h - 8);
    menu.style.left = Math.max(8, left) + 'px';
    menu.style.top = Math.max(8, top) + 'px';
  }
  function placeSub(menu, owner) {
    const vw = document.documentElement.clientWidth, vh = window.innerHeight;
    const r = owner.getBoundingClientRect(), w = menu.offsetWidth, h = menu.offsetHeight;
    let left = r.right - 2;
    if (left + w > vw - 8) left = Math.max(8, r.left - w + 2);
    let top = r.top - 5;
    if (top + h > vh - 8) top = Math.max(8, vh - h - 8);
    menu.style.left = left + 'px';
    menu.style.top = top + 'px';
  }

  const itemsIn = (menu) => [...menu.querySelectorAll(':scope > .cm-item')];
  const levelOf = (menu) => Number(menu.dataset.level || 0);
  function focusItem(el) {
    if (!el) return;
    itemsIn(el.parentElement).forEach((i) => i.classList.toggle('aktif', i === el));
    el.focus({ preventScroll: true });
  }
  function closeSubs(level) {
    while (state && state.stack.length > level + 1) state.stack.pop().remove();
  }
  function openSub(el, focusFirst) {
    const it = itemOf.get(el);
    if (!state || !it || !it.children || it.disabled) return;
    const level = levelOf(el.parentElement);
    const next = state.stack[level + 1];
    if (next && next._owner === el) { if (focusFirst) focusItem(itemsIn(next)[0]); return; }
    closeSubs(level);
    const sub = build(it.children, level + 1);
    sub._owner = el;
    document.body.appendChild(sub);
    placeSub(sub, el);
    sub.classList.add('acik');
    state.stack.push(sub);
    if (focusFirst) focusItem(itemsIn(sub)[0]);
  }
  function close(restoreFocus) {
    if (!state) return;
    clearTimeout(hoverTimer);
    const old = state;
    state = null;
    old.stack.forEach((m) => m.remove());
    if (restoreFocus && old.focusBefore && document.contains(old.focusBefore)) { try { old.focusBefore.focus({ preventScroll: true }); } catch { /* */ } }
  }
  function activate(el, keyboard) {
    const it = itemOf.get(el);
    if (!it || !state) return;
    if (it.disabled) { if (typeof it.disabled === 'string') toast(it.disabled, 'warn'); return; }
    if (it.children) return openSub(el, true);
    const ctx = state.ctx;
    close(keyboard);
    if (typeof it.action === 'function') { try { const r = it.action(ctx); if (r && r.catch) r.catch(fail); } catch (err) { fail(err); } }
  }
  function onHover(e) {
    const el = e.target.closest('.cm-item');
    if (!el || !state || document.activeElement === el) return;
    focusItem(el);
    const level = levelOf(el.parentElement);
    clearTimeout(hoverTimer);
    hoverTimer = setTimeout(() => {
      if (!state || !el.isConnected) return;
      if (itemOf.get(el) && itemOf.get(el).children) openSub(el, false);
      else closeSubs(level);
    }, 140);
  }
  function open(items, x, y, o = {}) {
    close(false);
    items = normalize(items);
    if (!items.length) return;
    state = { stack: [], ctx: { close: () => close(false), ...(o.ctx || {}) }, focusBefore: document.activeElement, keyboard: !!o.keyboard };
    const menu = build(items, 0);
    document.body.appendChild(menu);
    place(menu, x, y);
    state.stack.push(menu);
    requestAnimationFrame(() => menu.classList.add('acik'));
    if (o.keyboard) focusItem(itemsIn(menu)[0]); else menu.focus({ preventScroll: true });
  }

  // ── İçerik ──────────────────────────────────────────────────────────────────
  function pageItems() {
    return [
      { label: t('cmBack'), icon: 'left', hint: 'Alt+←', action: () => history.back() },
      { label: t('cmForward'), icon: 'right', hint: 'Alt+→', action: () => history.forward() },
      { label: t('refresh'), icon: 'refresh', hint: 'F5', action: () => render(true) },
      '-',
      { label: t('cmJobs'), icon: 'zap', action: () => go('#/jobs') },
      { label: t('cmSettings'), icon: 'sliders', action: () => go('#/settings') },
      { label: t('cmHelp'), icon: 'help', action: () => go('#/help') },
    ];
  }
  function register(selector, fn) {
    const entry = { selector, fn };
    registry.push(entry);
    return () => { const i = registry.indexOf(entry); if (i > -1) registry.splice(i, 1); };
  }
  // En yakın kayıtlı öğe kazanır (aynı öğede sonra kaydedilen önce)
  function specific(target, ctx) {
    for (let el = target; el && el.nodeType === 1 && el !== document.documentElement; el = el.parentElement) {
      for (let i = registry.length - 1; i >= 0; i--) {
        const r = registry[i];
        let ok = false;
        try { ok = el.matches(r.selector); } catch { /* */ }
        if (!ok) continue;
        const items = r.fn(el, ctx);
        if (items === false) return false;
        if (items && normalize(items).length) return items;
      }
    }
    return null;
  }
  function compose(target, event) {
    const selection = selectedText();
    const ctx = { target, event, selection };
    const own = specific(target, ctx);
    if (own === false) return false;
    const groups = [];
    if (selection) groups.push([{ label: t('cmCopy'), icon: 'copy', hint: 'Ctrl+C', action: () => copyText(selection, t('cmSelection')) }]);
    if (own) groups.push(own);
    let items = [];
    groups.forEach((g) => { items = items.concat(['-'], g); });
    if (groups.length) items.push('-', { label: t('cmPage'), icon: 'layers', children: pageItems() });
    else items = pageItems();
    items.push('-', { note: t('cmBrowserNote') });
    return { items, ctx };
  }
  function openFor(target, x, y, o = {}) {
    const built = compose(target, o.event || null);
    if (built === false) return;
    open(built.items, x, y, { keyboard: o.keyboard, ctx: built.ctx });
  }

  // ── Olaylar ─────────────────────────────────────────────────────────────────
  document.addEventListener('contextmenu', (e) => {
    if (inMenu(e.target) || Date.now() < suppressUntil) { e.preventDefault(); return; }
    if (e.defaultPrevented) return;
    if (e.shiftKey || e.ctrlKey || e.metaKey || e.altKey) { close(false); return; } // tarayıcı menüsü
    const target = e.target.nodeType === 1 ? e.target : e.target.parentElement;
    if (isNative(target)) { close(false); return; }
    e.preventDefault();
    let x = e.clientX, y = e.clientY, keyboard = false;
    if (!x && !y) { const r = target.getBoundingClientRect(); x = r.left + Math.min(24, r.width / 2); y = Math.min(r.bottom, innerHeight - 40); keyboard = true; }
    openFor(target, x, y, { event: e, keyboard });
  });
  document.addEventListener('keydown', (e) => {
    if (state) return onKey(e);
    if (!(e.key === 'ContextMenu' || (e.key === 'F10' && e.shiftKey && !e.ctrlKey && !e.altKey))) return;
    let target = document.activeElement || document.body;
    if (isNative(target)) return;
    if (target === document.documentElement) target = document.body;
    const r = target.getBoundingClientRect(), body = target === document.body;
    e.preventDefault();
    suppressUntil = Date.now() + 1000; // bazı sistemlerde ardından tarayıcının kendi olayı gelir
    openFor(target, body ? innerWidth / 2 - 120 : r.left + Math.min(24, r.width / 2), body ? 120 : Math.max(8, Math.min(r.bottom, innerHeight - 40)), { event: e, keyboard: true });
  }, true);
  function onKey(e) {
    const a = document.activeElement;
    const menu = a && a.closest && a.closest('.cm-menu') && state.stack.includes(a.closest('.cm-menu')) ? a.closest('.cm-menu') : state.stack[state.stack.length - 1];
    const items = itemsIn(menu), i = items.indexOf(a), level = levelOf(menu);
    const stop = () => { e.preventDefault(); e.stopPropagation(); e.stopImmediatePropagation(); };
    switch (e.key) {
      case 'ArrowDown': stop(); focusItem(items[(i + 1) % items.length]); break;
      case 'ArrowUp': stop(); focusItem(items[i <= 0 ? items.length - 1 : i - 1]); break;
      case 'Home': stop(); focusItem(items[0]); break;
      case 'End': stop(); focusItem(items[items.length - 1]); break;
      case 'ArrowRight': stop(); if (i > -1 && itemOf.get(items[i]).children) openSub(items[i], true); break;
      case 'ArrowLeft': stop(); if (level > 0) { const o = menu._owner; closeSubs(level - 1); focusItem(o); } break;
      case 'Escape': stop(); if (level > 0) { const o = menu._owner; closeSubs(level - 1); focusItem(o); } else close(true); break;
      case 'Enter': case ' ': stop(); if (i > -1) activate(items[i], true); break;
      case 'Tab': case 'ContextMenu': case 'F10': stop(); close(true); break;
      default:
        if (e.key.length === 1 && !e.ctrlKey && !e.metaKey && !e.altKey) {
          stop();
          const ch = e.key.toLocaleLowerCase(LANG === 'tr' ? 'tr' : 'en');
          const order = items.slice(i + 1).concat(items.slice(0, i + 1));
          const hit = order.find((x) => textOf(x.querySelector('.cm-etiket')).toLocaleLowerCase(LANG === 'tr' ? 'tr' : 'en').startsWith(ch));
          if (hit) focusItem(hit);
        }
    }
  }
  document.addEventListener('pointerdown', (e) => { if (state && !inMenu(e.target)) close(false); }, true);
  window.addEventListener('scroll', (e) => { if (state && !inMenu(e.target)) close(false); }, true);
  window.addEventListener('resize', () => close(false));
  window.addEventListener('blur', () => close(false));
  window.addEventListener('hashchange', () => close(false));

  window.CM = { register, open, close: () => close(false), copy: copyText, isOpen: () => !!state };
})();

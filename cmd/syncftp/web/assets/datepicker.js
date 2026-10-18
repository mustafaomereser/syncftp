'use strict';

/* ═══════════════════════════════════════════════════════════════════════════
   Tarih/saat seçici — tarayıcının yerleşik date input'u yerine panelin kendi bileşeni.
   datePicker(input, {time: true})  → input'un yanına takvim düğmesi ekler; açılır pencere:
     ay başlığı (‹ ›), haftanın günleri (Pzt…), gün ızgarası, saat:dakika, hızlı seçimler,
     Temizle / Tamam. Değer "YYYY-MM-DD HH:MM" yazılır (CLI'daki parseMilestoneDate'in kabul
     ettiği biçim); kutuya elle "bugün", "3d", "14:30" yazmak da çalışmaya devam eder.
   initDatePickers(root) → root içindeki [data-date] input'larına bağlar.
   app.js'ten sonra yüklenir (T / icon / esc / $ / $$).
   ═══════════════════════════════════════════════════════════════════════════ */

Object.assign(T.tr, {
  dpOpen: 'Takvimden seç', dpNow: 'Şimdi', dpToday: 'Bugün', dpYesterday: 'Dün', dp1h: '1 saat önce', dp3d: '3 gün önce', dp1w: '1 hafta önce',
  dpClear: 'Temizle', dpOk: 'Tamam', dpTime: 'Saat', dpPrev: 'Önceki ay', dpNext: 'Sonraki ay',
});
Object.assign(T.en, {
  dpOpen: 'Pick from calendar', dpNow: 'Now', dpToday: 'Today', dpYesterday: 'Yesterday', dp1h: '1 hour ago', dp3d: '3 days ago', dp1w: '1 week ago',
  dpClear: 'Clear', dpOk: 'Done', dpTime: 'Time', dpPrev: 'Previous month', dpNext: 'Next month',
});

const DP = { open: null };
const dpPad = (n) => String(n).padStart(2, '0');
const dpFormat = (d, time) => `${d.getFullYear()}-${dpPad(d.getMonth() + 1)}-${dpPad(d.getDate())}${time ? ` ${dpPad(d.getHours())}:${dpPad(d.getMinutes())}` : ''}`;
const dpSameDay = (a, b) => a && b && a.getFullYear() === b.getFullYear() && a.getMonth() === b.getMonth() && a.getDate() === b.getDate();

// Kutudaki değeri tarihe çevirir: 2026-10-18 [14:30] ya da 18.10.2026 [14:30]; olmazsa null
function dpParse(v) {
  v = (v || '').trim();
  let m = v.match(/^(\d{4})-(\d{1,2})-(\d{1,2})(?:[ T](\d{1,2}):(\d{2}))?/);
  if (m) return new Date(+m[1], m[2] - 1, +m[3], +(m[4] || 0), +(m[5] || 0));
  m = v.match(/^(\d{1,2})\.(\d{1,2})\.(\d{4})(?:\s+(\d{1,2}):(\d{2}))?/);
  if (m) return new Date(+m[3], m[2] - 1, +m[1], +(m[4] || 0), +(m[5] || 0));
  return null;
}

function dpClose() {
  if (!DP.open) return;
  DP.open.el.remove();
  document.removeEventListener('mousedown', DP.open.outside, true);
  window.removeEventListener('resize', DP.open.place);
  DP.open = null;
}

function dpShow(input, { time = true } = {}) {
  if (DP.open && DP.open.input === input) return dpClose();
  dpClose();
  const now = new Date();
  let sel = dpParse(input.value);
  let view = new Date((sel || now).getFullYear(), (sel || now).getMonth(), 1);
  let hh = sel ? sel.getHours() : now.getHours(), mm = sel ? sel.getMinutes() : now.getMinutes();
  const el = document.createElement('div');
  el.className = 'dp';
  el.setAttribute('role', 'dialog');
  document.body.appendChild(el);

  const months = (i) => new Date(2000, i, 1).toLocaleDateString(LANG === 'tr' ? 'tr-TR' : 'en-GB', { month: 'long' });
  const days = LANG === 'tr' ? ['Pt', 'Sa', 'Ça', 'Pe', 'Cu', 'Ct', 'Pz'] : ['Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa', 'Su'];
  const write = () => { if (sel) { sel.setHours(hh, mm, 0, 0); input.value = dpFormat(sel, time); input.dispatchEvent(new Event('input', { bubbles: true })); } };
  const quick = [
    ['dpNow', () => new Date()],
    ['dpToday', () => { const d = new Date(); d.setHours(0, 0, 0, 0); return d; }],
    ['dpYesterday', () => { const d = new Date(); d.setDate(d.getDate() - 1); d.setHours(0, 0, 0, 0); return d; }],
    ['dp1h', () => new Date(Date.now() - 3600e3)],
    ['dp3d', () => new Date(Date.now() - 3 * 86400e3)],
    ['dp1w', () => new Date(Date.now() - 7 * 86400e3)],
  ];

  function draw() {
    const first = (view.getDay() + 6) % 7; // pazartesi başlangıçlı
    const start = new Date(view.getFullYear(), view.getMonth(), 1 - first);
    const cells = [];
    for (let i = 0; i < 42; i++) {
      const d = new Date(start.getFullYear(), start.getMonth(), start.getDate() + i);
      const cls = [d.getMonth() !== view.getMonth() && 'disari', dpSameDay(d, now) && 'bugun', dpSameDay(d, sel) && 'secili', d > now && 'gelecek'].filter(Boolean).join(' ');
      cells.push(`<button type="button" class="dp-gun ${cls}" data-d="${d.getFullYear()}-${d.getMonth()}-${d.getDate()}">${d.getDate()}</button>`);
    }
    el.innerHTML = `<div class="dp-bas">
        <button type="button" class="ikon-btn" data-nav="-1" title="${t('dpPrev')}">${icon('left')}</button>
        <b>${esc(months(view.getMonth()))} ${view.getFullYear()}</b>
        <button type="button" class="ikon-btn" data-nav="1" title="${t('dpNext')}">${icon('right')}</button></div>
      <div class="dp-hafta">${days.map((d) => `<span>${d}</span>`).join('')}</div>
      <div class="dp-izgara">${cells.join('')}</div>
      ${time ? `<div class="dp-saat"><span>${icon('clock')}${t('dpTime')}</span>
        <input type="number" min="0" max="23" data-h value="${dpPad(hh)}"><b>:</b><input type="number" min="0" max="59" step="5" data-m value="${dpPad(mm)}"></div>` : ''}
      <div class="dp-hizli">${quick.map(([k], i) => `<button type="button" class="cip" data-q="${i}">${t(k)}</button>`).join('')}</div>
      <div class="dp-alt"><button type="button" class="btn kucuk hayalet" data-clear>${t('dpClear')}</button><button type="button" class="btn kucuk birincil" data-ok>${icon('check')}${t('dpOk')}</button></div>`;
    $$('[data-nav]', el).forEach((b) => (b.onclick = () => { view = new Date(view.getFullYear(), view.getMonth() + +b.dataset.nav, 1); draw(); }));
    $$('[data-d]', el).forEach((b) => (b.onclick = () => {
      const [y, mo, d] = b.dataset.d.split('-').map(Number);
      sel = new Date(y, mo, d);
      if (mo !== view.getMonth()) view = new Date(y, mo, 1);
      write(); draw();
    }));
    $$('[data-q]', el).forEach((b) => (b.onclick = () => {
      sel = quick[+b.dataset.q][1]();
      hh = sel.getHours(); mm = sel.getMinutes();
      view = new Date(sel.getFullYear(), sel.getMonth(), 1);
      write(); dpClose(); input.focus();
    }));
    const h = $('[data-h]', el), m = $('[data-m]', el);
    if (h) {
      const clamp = (v, max) => Math.min(max, Math.max(0, parseInt(v, 10) || 0));
      h.onchange = () => { hh = clamp(h.value, 23); h.value = dpPad(hh); if (!sel) sel = new Date(now.getFullYear(), now.getMonth(), now.getDate()); write(); draw(); };
      m.onchange = () => { mm = clamp(m.value, 59); m.value = dpPad(mm); if (!sel) sel = new Date(now.getFullYear(), now.getMonth(), now.getDate()); write(); draw(); };
    }
    $('[data-clear]', el).onclick = () => { input.value = ''; input.dispatchEvent(new Event('input', { bubbles: true })); dpClose(); input.focus(); };
    $('[data-ok]', el).onclick = () => { write(); dpClose(); input.focus(); };
    place();
  }
  // Kutunun altına, sığmazsa üstüne; yatayda ekrandan taşmaz (modal içinde de çalışır — fixed)
  function place() {
    const r = input.getBoundingClientRect();
    const w = el.offsetWidth, hgt = el.offsetHeight;
    let top = r.bottom + 6;
    if (top + hgt > innerHeight - 8) top = r.top - hgt - 6 > 8 ? r.top - hgt - 6 : innerHeight - hgt - 8; // iki yana da sığmazsa ekrana yasla
    el.style.top = `${Math.max(8, top)}px`;
    el.style.left = `${Math.min(Math.max(8, r.left), innerWidth - w - 8)}px`;
  }
  const outside = (e) => { if (!el.contains(e.target) && e.target !== input && !e.target.closest('[data-dp-btn]')) dpClose(); };
  document.addEventListener('mousedown', outside, true);
  window.addEventListener('resize', place);
  el.addEventListener('keydown', (e) => { if (e.key === 'Escape') { e.stopPropagation(); dpClose(); input.focus(); } });
  DP.open = { el, input, outside, place };
  draw();
}

function datePicker(input, opts = {}) {
  if (input.dataset.dpBound) return;
  input.dataset.dpBound = '1';
  input.autocomplete = 'off';
  const btn = document.createElement('button');
  btn.type = 'button';
  btn.className = 'btn';
  btn.dataset.dpBtn = '1';
  btn.title = t('dpOpen');
  btn.innerHTML = icon('clock');
  const wrap = document.createElement('div');
  wrap.className = 'girdi-grup';
  input.parentNode.insertBefore(wrap, input);
  wrap.append(input, btn);
  btn.onclick = (e) => { e.preventDefault(); dpShow(input, opts); };
  // Kutu açıkken Esc yalnızca takvimi kapatır (modal kapanmasın); kutuya yazınca takvim güncellenir
  input.addEventListener('keydown', (e) => {
    if (e.key === 'Escape' && DP.open && DP.open.input === input) { e.stopPropagation(); dpClose(); }
    if (e.key === 'ArrowDown' && e.altKey) { e.preventDefault(); dpShow(input, opts); }
  });
}

function initDatePickers(root) {
  $$('input[data-date]', root).forEach((i) => datePicker(i, { time: i.dataset.date !== 'day' }));
}
// Sayfa / modal değişince açık takvim kalmasın
window.addEventListener('hashchange', dpClose);

'use strict';

/* ═══════════════════════════════════════════════════════════════════════════
   Tooltip — tarayıcının yerleşik title balonu yerine Bootstrap tarzı balon (oklu, üstte; yer yoksa
   altta; ekrandan taşmaz). Paneldeki TÜM [title] / [data-tip] öğeleri için tek delegasyon: üstüne
   gelince title data-tip'e taşınır (yerleşik balon çıkmasın), kodun sonradan yazdığı title da
   bir sonraki üzerine gelişte okunur. Klavye odağında da açılır. Çok satırlı metin (\n) korunur.
   ═══════════════════════════════════════════════════════════════════════════ */

const TIP = { el: null, target: null, timer: 0, raf: 0 };

function tipHide() {
  clearTimeout(TIP.timer);
  cancelAnimationFrame(TIP.raf);
  TIP.target = null;
  if (TIP.el) TIP.el.classList.remove('acik');
}

function tipShow(t) {
  if (!TIP.el) {
    TIP.el = document.createElement('div');
    TIP.el.className = 'tip';
    TIP.el.setAttribute('role', 'tooltip');
    TIP.el.innerHTML = '<div class="tip-ic"></div><span class="tip-ok"></span>';
    document.body.appendChild(TIP.el);
  }
  const el = TIP.el;
  $('.tip-ic', el).textContent = t.dataset.tip;
  el.classList.remove('alt');
  el.style.left = '0px'; el.style.top = '0px';
  const r = t.getBoundingClientRect(), w = el.offsetWidth, h = el.offsetHeight;
  let top = r.top - h - 8;
  if (top < 6) { top = r.bottom + 8; el.classList.add('alt'); } // üstte yer yok → altta
  const cx = r.left + r.width / 2;
  const left = Math.min(Math.max(6, cx - w / 2), innerWidth - w - 6);
  el.style.left = `${left}px`; el.style.top = `${top}px`;
  $('.tip-ok', el).style.left = `${Math.min(Math.max(8, cx - left), w - 8)}px`;
  el.classList.add('acik');
  // Hedef yeniden çizimle DOM'dan kalkarsa balon asılı kalmasın
  // Açıkken kod title'ı güncellerse (ör. watch satırı) yerleşik balon çıkmasın, metin tazelensin
  const watch = () => {
    if (!TIP.target || !TIP.target.isConnected) return tipHide();
    if (TIP.target.hasAttribute('title')) {
      const v = TIP.target.getAttribute('title');
      TIP.target.removeAttribute('title');
      if (v) { TIP.target.dataset.tip = v; $('.tip-ic', el).textContent = v; }
    }
    TIP.raf = requestAnimationFrame(watch);
  };
  TIP.raf = requestAnimationFrame(watch);
}

const tipTarget = (node) => (node && node.closest ? node.closest('[title], [data-tip]') : null);

function tipEnter(t, delay) {
  if (t === TIP.target) return;
  tipHide();
  if (!t) return;
  if (t.hasAttribute('title')) {
    const v = t.getAttribute('title');
    t.removeAttribute('title');
    if (v) t.dataset.tip = v;
  }
  if (!t.dataset.tip) return;
  TIP.target = t;
  TIP.timer = setTimeout(() => TIP.target === t && tipShow(t), delay);
}

document.addEventListener('mouseover', (e) => tipEnter(tipTarget(e.target), 120));
document.addEventListener('mouseout', (e) => { if (TIP.target && !TIP.target.contains(e.relatedTarget)) tipHide(); });
document.addEventListener('focusin', (e) => { if (e.target.matches(':focus-visible')) tipEnter(tipTarget(e.target), 0); });
document.addEventListener('focusout', tipHide);
document.addEventListener('mousedown', tipHide, true);
document.addEventListener('keydown', (e) => { if (e.key === 'Escape') tipHide(); });
document.addEventListener('scroll', tipHide, true);
window.addEventListener('hashchange', tipHide);
window.addEventListener('blur', tipHide);

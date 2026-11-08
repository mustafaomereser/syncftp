// Sürüm kontrolü: yeni sürüm varsa üst barda (bar yoksa sol menüde) "Güncelle"; onaylanınca indirir, kurar, yeniden başlar.
Object.assign(T.tr, {
  upTitle: 'Güncelleme', upBadge: 'Güncelle {0}', upBody: 'syncFTP {0} çıktı (sende {1}).', upGo: 'Güncelle ve yeniden başlat', upLater: 'Sonra',
  upBusy: 'İndiriliyor ve doğrulanıyor…', upDone: 'Güncellendi. syncFTP yeniden başlıyor, yeni pencere açılınca bu pencereyi kapatabilirsin.', upNoPkg: 'Bu sürümde platformun için kurulabilir paket yok, Release sayfasından indirebilirsin.',
  upOpenRelease: 'Release sayfası', upSection: 'Güncellemeler', upAuto: 'Yeni sürümleri otomatik denetle', upAutoHint: 'Açılışta ve günde bir GitHub\'a bakar; kurulum her zaman senin onayınla',
  upCheck: 'Şimdi denetle', upLatest: 'Son sürümü kullanıyorsun ({0})', upFound: 'Yeni sürüm var: {0}', upFail: 'Denetlenemedi: {0}',
});
Object.assign(T.en, {
  upTitle: 'Update', upBadge: 'Update {0}', upBody: 'syncFTP {0} is out (you have {1}).', upGo: 'Update and restart', upLater: 'Later',
  upBusy: 'Downloading and verifying…', upDone: 'Updated. syncFTP is restarting; you can close this window once the new one opens.', upNoPkg: 'No installable package for your platform in this release; you can download it from the Release page.',
  upOpenRelease: 'Release page', upSection: 'Updates', upAuto: 'Check for new versions automatically', upAutoHint: 'Checks GitHub at startup and once a day; installing always needs your approval',
  upCheck: 'Check now', upLatest: 'You are on the latest version ({0})', upFound: 'New version available: {0}', upFail: 'Check failed: {0}',
});

const UP = { info: null };

function upDraw() {
  const u = UP.info, on = !!(u && u.available), label = on ? t('upBadge', u.latest) : '';
  const b = document.querySelector('#titlebar .tb-guncel');
  if (b) { b.hidden = !on; b.lastElementChild.textContent = label; }
  let a = document.getElementById('nav-update');
  if (!a) {
    const s = document.querySelector('a[data-nav="settings"]');
    if (!s) return;
    s.insertAdjacentHTML('beforebegin', `<a href="#" id="nav-update" hidden>${icon('upload')}<span class="ad"></span></a>`);
    a = document.getElementById('nav-update');
    a.onclick = (e) => { e.preventDefault(); openUpdate(); };
  }
  a.hidden = !on || !!b;
  a.querySelector('.ad').textContent = label;
}

async function upRefresh(force) {
  try { UP.info = force ? await api('POST', '/api/update/check') : await GET('/api/update'); upDraw(); } catch { /* sessiz */ }
  return UP.info;
}

function openUpdate() {
  const u = UP.info;
  if (!u || !u.available || document.querySelector('.up-durum')) return;
  const relLink = `<a href="${esc(u.url || '#')}" target="_blank" rel="noopener">${t('upOpenRelease')}</a>`;
  const m = modal({
    title: `${icon('upload')}${t('upTitle')}`,
    body: `<p>${esc(t('upBody', u.latest, u.current))}</p>${u.notes ? `<pre class="up-not">${esc(u.notes)}</pre>` : ''}<p class="ipucu">${relLink}</p><div class="qc-durum up-durum" hidden></div>`,
    actions: [{ label: t('upLater'), cls: 'hayalet' }, ...(u.installable ? [{ label: t('upGo'), cls: 'birincil', focus: true, run: () => go_() }] : [])],
  });
  const st = $('.up-durum', m.root);
  const show = (msg, kind) => { st.hidden = false; st.className = 'qc-durum up-durum' + (kind === 'bad' ? ' hata' : ''); st.innerHTML = `${icon(kind === 'bad' ? 'alert' : kind === 'ok' ? 'check' : 'refresh', kind ? '' : 'spin')}<span>${esc(msg)}</span>`; };
  if (!u.installable) show(t('upNoPkg'), 'bad');
  async function go_() {
    show(t('upBusy'));
    try {
      await api('POST', '/api/update/apply');
      show(t('upDone'), 'ok');
      $$('.modal-alt button', m.root).forEach((b) => (b.disabled = true));
      UP.info = null;
    } catch (e) { show(e.message, 'bad'); }
    return false;
  }
}

// Ayarlar → Sistem bölümü
function updateSectionHTML() {
  const u = UP.info || {};
  return section('upload', t('upSection'), '', `${fBool('up-auto', t('upAuto'), u.auto !== false, t('upAutoHint'))}
    <div class="aksiyonlar" style="margin-top:10px"><button class="btn" id="up-check">${icon('refresh')}${t('upCheck')}</button><span class="faint" id="up-sonuc"></span></div>`);
}
function bindUpdateSection(root) {
  const c = $('[data-k="up-auto"]', root), b = $('#up-check', root), out = $('#up-sonuc', root);
  if (!c || !b) return;
  c.onchange = async () => { try { await api('PUT', '/api/update/settings', { auto: c.checked }); toast(t('saved')); } catch (e) { fail(e); } };
  b.onclick = async () => {
    b.disabled = true;
    const u = await upRefresh(true);
    b.disabled = false;
    if (!u) return;
    out.textContent = u.error ? t('upFail', u.error) : u.available ? t('upFound', u.latest) : t('upLatest', u.current);
    if (!u.error && u.available) openUpdate();
  };
}

// Sunucu kontrolü açılıştan 5 sn sonra bitiyor
window.addEventListener('DOMContentLoaded', () => { setTimeout(upRefresh, 9000); setTimeout(upDraw, 300); setInterval(upRefresh, 3600 * 1000); });

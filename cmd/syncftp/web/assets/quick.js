// Hızlı bağlantı (FileZilla'daki gibi): pencere barındaki düğme + Ctrl+K. Projesiz bağlantı sanal "@quick"
// projesine eklenir (cmd_gui_quick.go) ve VAR OLAN Dosyalar ekranı (tabFiles) "@conn:<ad>" hedefiyle açılır.
Object.assign(T.tr, {
  qcTitle: 'Hızlı bağlan', qcHint: 'Proje açmadan bir sunucuya bağlan', qcProtocol: 'Protokol', qcHost: 'Sunucu', qcHostPh: 'ftp.ornek.com ya da sftp://ornek.com:22',
  qcPort: 'Port', qcUser: 'Kullanıcı', qcPass: 'Şifre', qcPassKept: 'Kayıtlı şifre kullanılır', qcKey: 'Özel anahtar dosyası', qcKeyPh: 'Boşsa şifreyle girilir',
  qcRemember: 'Şifreyi hatırla', qcRememberHint: 'Windows Kimlik Bilgileri\'nde saklanır; geçmiş dosyasına yazılmaz', qcPassive: 'Pasif mod', qcTls: 'TLS sertifikasını doğrulama',
  qcConnect: 'Bağlan', qcConnecting: 'Bağlanıyor…', qcHistory: 'Geçmiş bağlantılar', qcNoHistory: 'Henüz geçmiş bağlantı yok', qcForget: 'Geçmişten sil (hatırlanan şifre de silinir)',
  qcOpen: 'açık', qcSaved: 'şifre kayıtlı', qcClose: 'Bağlantıyı kes', qcSaveLib: 'Kütüphaneye kaydet', qcSavedLib: 'Bağlantı kütüphanesine kaydedildi',
  qcNoSession: 'Açık hızlı bağlantı yok', qcNoSessionHint: 'Panel yeniden başlayınca hızlı bağlantılar kapanır, geçmişten tek tıkla tekrar açabilirsin.',
});
Object.assign(T.en, {
  qcTitle: 'Quick connect', qcHint: 'Connect to a server without a project', qcProtocol: 'Protocol', qcHost: 'Host', qcHostPh: 'ftp.example.com or sftp://example.com:22',
  qcPort: 'Port', qcUser: 'User', qcPass: 'Password', qcPassKept: 'Saved password will be used', qcKey: 'Private key file', qcKeyPh: 'Empty = password login',
  qcRemember: 'Remember password', qcRememberHint: 'Kept in the OS credential store, never in the history file', qcPassive: 'Passive mode', qcTls: 'Skip TLS certificate check',
  qcConnect: 'Connect', qcConnecting: 'Connecting…', qcHistory: 'Recent connections', qcNoHistory: 'No recent connections yet', qcForget: 'Remove from history (also deletes the saved password)',
  qcOpen: 'open', qcSaved: 'password saved', qcClose: 'Disconnect', qcSaveLib: 'Save to library', qcSavedLib: 'Saved to the connection library',
  qcNoSession: 'No open quick connection', qcNoSessionHint: 'Quick connections close when the panel restarts; reopen one from the history in one click.',
});

const QUICK_P = '@quick';
const qcDefaultPort = (p) => (p === 'sftp' ? 22 : 21);

async function openQuickConnect() {
  if (document.querySelector('.qc-form')) return;
  let data = { history: [], open: [] };
  try { data = await GET('/api/quick'); } catch (e) { return fail(e); }
  const f = (k) => $(`[data-k="${k}"]`, m.root);
  const m = modal({
    title: `${icon('zap')}${t('qcTitle')}`, size: 'qc-modal',
    body: `<div class="form qc-form">
      <div class="qc-satir">${fSelect('protocol', t('qcProtocol'), 'ftp', config_PROTOS())}${fText('host', t('qcHost'), '', { ph: t('qcHostPh') })}${fNum('port', t('qcPort'), 21)}</div>
      <div class="qc-satir iki">${fText('user', t('qcUser'), '')}${fText('password', t('qcPass'), '', { type: 'password' })}</div>
      <div data-qc="key">${fText('key_file', t('qcKey'), '', { ph: t('qcKeyPh'), browse: 'file' })}</div>
      ${fBool('remember', t('qcRemember'), false, t('qcRememberHint'))}
      <div data-qc="ftp">${fBool('passive', t('qcPassive'), true)}</div>
      <div data-qc="tls">${fBool('tls_skip_verify', t('qcTls'), false)}</div>
      <div class="qc-durum" id="qc-durum" hidden></div></div>
      <h4 class="qc-baslik">${t('qcHistory')}</h4><div class="qc-gecmis" id="qc-gecmis"></div>`,
    actions: [{ label: t('cancel'), cls: 'hayalet' }, { label: t('qcConnect'), cls: 'birincil', run: () => connect() }],
  });
  let usingSaved = false;
  const vis = () => {
    const p = f('protocol').value;
    $('[data-qc="key"]', m.root).hidden = p !== 'sftp';
    $('[data-qc="ftp"]', m.root).hidden = p === 'sftp';
    $('[data-qc="tls"]', m.root).hidden = p !== 'ftps';
  };
  const status = (msg, bad) => {
    const d = $('#qc-durum', m.root);
    d.hidden = !msg; d.className = 'qc-durum' + (bad ? ' hata' : '');
    d.innerHTML = msg ? `${icon(bad ? 'alert' : 'refresh', bad ? '' : 'spin')}<span>${esc(msg)}</span>` : '';
  };
  const passHint = () => { f('password').placeholder = usingSaved ? t('qcPassKept') : ''; };
  // sftp://kullanici@host:22 yapıştırılırsa alanlara dağıt
  const parseHost = () => {
    const x = /^\s*(?:(ftps?|sftp):\/\/)?(?:([^@/\s]+)@)?([^:/\s]+)(?::(\d+))?\/?\s*$/i.exec(f('host').value);
    if (!x || !(x[1] || x[2] || x[4])) return;
    if (x[1]) { f('protocol').value = x[1].toLowerCase(); vis(); f('port').value = qcDefaultPort(f('protocol').value); }
    if (x[2]) f('user').value = x[2];
    f('host').value = x[3];
    if (x[4]) f('port').value = x[4];
  };
  async function connect(over) {
    parseHost();
    const body = over || {
      protocol: f('protocol').value, host: f('host').value.trim(), port: parseInt(f('port').value, 10) || 0, user: f('user').value.trim(), password: f('password').value,
      key_file: f('protocol').value === 'sftp' ? f('key_file').value.trim() : '', passive: f('passive').checked, tls_skip_verify: f('tls_skip_verify').checked, remember: f('remember').checked,
    };
    status(t('qcConnecting'));
    try {
      const r = await api('POST', '/api/quick/connect', body);
      if (r.warning) toast(r.warning, 'bad');
      m.close();
      go(pHash(QUICK_P, 'files', CONN_TARGET + r.name));
    } catch (e) { status(e.message, true); }
    return false; // başarıda pencereyi kendimiz kapattık, hatada açık kalsın
  }
  const fillFrom = (h) => {
    f('protocol').value = h.protocol; vis();
    f('host').value = h.host; f('port').value = h.port; f('user').value = h.user; f('password').value = '';
    f('key_file').value = h.key_file || ''; f('passive').checked = !!h.passive; f('tls_skip_verify').checked = !!h.tls_skip_verify; f('remember').checked = !!h.remember;
    usingSaved = !!h.remember; passHint(); status('');
  };
  const drawHistory = () => {
    $('#qc-gecmis', m.root).innerHTML = !data.history.length ? `<div class="faint qc-bos">${t('qcNoHistory')}</div>` : data.history.map((h, i) => `<div class="qc-kayit" data-i="${i}" role="button" tabindex="0">
      <span class="rozet">${esc(h.protocol.toUpperCase())}</span><b>${esc(h.name)}</b>
      ${data.open.includes(h.name) ? `<span class="rozet ok">${t('qcOpen')}</span>` : ''}${h.remember ? `<span class="faint qc-not">${t('qcSaved')}</span>` : ''}
      <span class="faint qc-zaman">${esc(new Date(h.last_used).toLocaleString())}</span>
      <button class="ikon-btn" data-forget="${i}" title="${esc(t('qcForget'))}">${icon('x')}</button></div>`).join('');
  };
  const useRow = (h) => {
    if (data.open.includes(h.name)) { m.close(); return go(pHash(QUICK_P, 'files', CONN_TARGET + h.name)); }
    fillFrom(h);
    if (h.remember) return connect({ ...h, password: '' });
    f('password').focus();
  };
  $('#qc-gecmis', m.root).addEventListener('click', async (e) => {
    const fb = e.target.closest('[data-forget]');
    if (fb) {
      e.stopPropagation();
      const h = data.history[+fb.dataset.forget];
      try { await api('POST', '/api/quick/forget', { name: h.name }); data.history.splice(+fb.dataset.forget, 1); drawHistory(); } catch (er) { status(er.message, true); }
      return;
    }
    const row = e.target.closest('.qc-kayit');
    if (row) useRow(data.history[+row.dataset.i]);
  });
  $('#qc-gecmis', m.root).addEventListener('keydown', (e) => {
    const row = e.target.closest('.qc-kayit');
    if (row && e.target === row && e.key === 'Enter') { e.preventDefault(); useRow(data.history[+row.dataset.i]); }
  });
  f('protocol').onchange = () => { const pr = f('port').value; if (pr === '21' || pr === '22' || !pr) f('port').value = qcDefaultPort(f('protocol').value); vis(); };
  f('host').onblur = parseHost;
  f('password').oninput = () => { usingSaved = false; passHint(); };
  $('.qc-form', m.root).addEventListener('keydown', (e) => { if (e.key === 'Enter' && e.target.matches('input')) { e.preventDefault(); connect(); } });
  bindFormTools(m.root);
  vis(); drawHistory();
  setTimeout(() => f('host').focus(), 40); // modal ilk alana (protokol) odaklanır; yazılacak yer sunucu
}
const config_PROTOS = () => [['ftp', 'FTP'], ['ftps', 'FTPS'], ['sftp', 'SFTP']];

// Sayfa: açık hızlı bağlantının dosyaları — gövde proje sayfasındaki tabFiles'ın aynısı
async function pageQuick(r) {
  const cfg = await getConfig(QUICK_P);
  const has = cfg.connections.some((c) => CONN_TARGET + c.name === r.sub);
  if (!has) {
    bind((root) => { $('#qc-ac', root).onclick = openQuickConnect; });
    return `<div class="sayfa"><div class="bos"><h3>${t('qcNoSession')}</h3><p>${t('qcNoSessionHint')}</p><button class="btn birincil" id="qc-ac">${icon('zap')}${t('qcTitle')}</button></div></div>`;
  }
  const c = cfg.connections.find((x) => CONN_TARGET + x.name === r.sub);
  const body = await tabFiles({ name: QUICK_P, path: '' }, r.sub, r.sub2);
  const inner = $('#sayfa')._bind;
  bind((root) => {
    if (inner) inner(root);
    $('#qc-save', root).onclick = async () => { try { await api('POST', '/api/quick/save', { name: c.name }); toast(t('qcSavedLib')); } catch (e) { fail(e); } };
    $('#qc-close', root).onclick = async () => {
      try {
        await api('POST', '/api/quick/close', { name: c.name });
        const rest = cfg.connections.filter((x) => x.name !== c.name);
        go(rest.length ? pHash(QUICK_P, 'files', CONN_TARGET + rest[0].name) : '#/');
      } catch (e) { fail(e); }
    };
  });
  return `<div class="sayfa"><div class="baslik-satir"><div><h1>${esc(c.name)}</h1><div class="yol">${icon('zap')}<span>${t('qcTitle')} · ${esc((c.protocol || 'ftp').toUpperCase())}</span></div></div>
    <div class="aksiyonlar"><span id="proj-aks" class="aksiyonlar"></span><button class="btn" id="qc-save">${icon('link')}${t('qcSaveLib')}</button><button class="btn tehlike" id="qc-close">${icon('x')}${t('qcClose')}</button></div></div>${body}</div>`;
}

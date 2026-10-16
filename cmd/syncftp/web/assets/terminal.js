'use strict';

/* Panel içi terminal (#/terminal[/proje]) — syncftp shell'i gerçek bir sözde terminalde (ConPTY/pty)
   çalışır, xterm.js ile çizilir. Oturum proje başına bir tane; sayfadan ayrılınca süreç yaşar,
   geri dönünce kaldığı yerden görünür (sunucu son 512 KB'ı yeniden oynatır). */

Object.assign(T.tr, {
  navTerminal: 'Terminal', termTitle: 'Terminal', termGeneral: 'Genel (proje yok)', termRestart: 'Yeniden başlat', termClose: 'Kapat',
  termExited: 'Oturum bitti', termExitedHint: 'Yeniden başlatmak için Enter\'a basın ya da "Yeniden başlat"a tıklayın.', termOpening: 'syncftp başlatılıyor…',
  termHint: 'CLI\'daki syncftp shell\'inin aynısı — help ile komutları görün. Çıkış: iki kez Ctrl+C.',
});
Object.assign(T.en, {
  navTerminal: 'Terminal', termTitle: 'Terminal', termGeneral: 'General (no project)', termRestart: 'Restart', termClose: 'Close',
  termExited: 'Session ended', termExitedHint: 'Press Enter or click "Restart" to start again.', termOpening: 'starting syncftp…',
  termHint: 'Same as the syncftp shell in the CLI — type help for commands. Exit: Ctrl+C twice.',
});

// key: proje adı ('' = genel) → { id, term, fit, host, es, exited, queue }
const TERMS = { sessions: new Map(), current: null };

const TERM_THEME = {
  background: '#111113', foreground: '#edeef0', cursor: '#3b9eff', cursorAccent: '#111113', selectionBackground: '#3b9eff55',
  black: '#1b1c1f', red: '#ff6369', green: '#3dd68c', yellow: '#ffca16', blue: '#3b9eff', magenta: '#bf7af0', cyan: '#4ccce6', white: '#edeef0',
  brightBlack: '#6f6e77', brightRed: '#ff8589', brightGreen: '#70e1a8', brightYellow: '#ffe066', brightBlue: '#70b8ff', brightMagenta: '#d19dff', brightCyan: '#7ee0f0', brightWhite: '#ffffff',
};

function termB64ToBytes(b64) {
  const bin = atob(b64);
  const out = new Uint8Array(bin.length);
  for (let i = 0; i < bin.length; i++) out[i] = bin.charCodeAt(i);
  return out;
}

async function termEnsure(key, restart = false) {
  let s = TERMS.sessions.get(key);
  if (s && !restart && !s.exited) return s;
  if (!s) {
    const host = document.createElement('div');
    host.className = 'term-host';
    const term = new Terminal({
      fontFamily: getComputedStyle(document.documentElement).getPropertyValue('--mono') || 'Consolas, monospace',
      fontSize: 13, lineHeight: 1.15, cursorBlink: true, allowProposedApi: true, scrollback: 5000, theme: TERM_THEME,
    });
    const fit = new FitAddon.FitAddon();
    term.loadAddon(fit);
    s = { key, id: null, term, fit, host, es: null, exited: false, queue: '', timer: null, opened: false };
    TERMS.sessions.set(key, s);
    term.onData((d) => {
      if (s.exited) { if (d === '\r') termEnsure(key, true).then(() => termAttach(key)); return; }
      s.queue += d;
      if (!s.timer) s.timer = setTimeout(() => termFlush(s), 8);
    });
    term.onResize(({ cols, rows }) => { if (s.id && !s.exited) api('POST', `/api/terms/${enc(s.id)}/resize`, { cols, rows }).catch(() => {}); });
  }
  if (s.es) { s.es.close(); s.es = null; }
  if (restart) { s.term.reset(); }
  s.exited = false;
  const dims = s.opened ? { cols: s.term.cols, rows: s.term.rows } : { cols: 120, rows: 30 };
  const info = await api('POST', '/api/terms', { project: key, restart, ...dims });
  s.id = info.id;
  s.dir = info.dir;
  s.es = new EventSource(`/api/terms/${enc(s.id)}/stream`);
  s.es.onmessage = (e) => s.term.write(termB64ToBytes(e.data));
  s.es.addEventListener('exit', () => {
    s.exited = true;
    s.es.close(); s.es = null;
    s.term.write(`\r\n\x1b[2m── ${t('termExited')} · ${t('termExitedHint')}\x1b[0m\r\n`);
    termDrawBar();
  });
  return s;
}

async function termFlush(s) {
  s.timer = null;
  const data = s.queue;
  s.queue = '';
  if (!data || !s.id) return;
  try { await api('POST', `/api/terms/${enc(s.id)}/input`, { data }); } catch { /* süreç bitti */ }
}

function termAttach(key) {
  const s = TERMS.sessions.get(key);
  const box = $('#term-kap');
  if (!s || !box) return;
  TERMS.current = key;
  box.replaceChildren(s.host);
  if (!s.opened) { s.term.open(s.host); s.opened = true; }
  requestAnimationFrame(() => { try { s.fit.fit(); } catch { /* görünmüyor */ } s.term.focus(); });
  termDrawBar();
}

function termDrawBar() {
  const s = TERMS.sessions.get(TERMS.current);
  const dir = $('#term-dir');
  if (dir && s) dir.textContent = s.dir || '';
  const st = $('#term-durum');
  if (st && s) st.innerHTML = s.exited ? `<span class="pill bad"><span class="nokta"></span>${t('termExited')}</span>` : `<span class="pill ok"><span class="nokta"></span>syncftp</span>`;
}

window.addEventListener('resize', () => {
  const s = TERMS.sessions.get(TERMS.current);
  if (s && s.opened && document.body.contains(s.host)) { try { s.fit.fit(); } catch { /* yoksay */ } }
});

async function pageTerminal(sub) {
  if (!S.projects.length) await loadProjects();
  // Proje: URL'deki, yoksa son açık terminal, yoksa son gezilen proje
  let key = sub != null ? sub : (TERMS.current ?? (S.lastProject || ''));
  if (key && !S.projects.some((p) => p.name === key)) key = '';
  bind(async (root) => {
    const sel = $('#term-proj', root);
    sel.onchange = () => go(sel.value ? `#/terminal/${enc(sel.value)}` : '#/terminal/');
    $('#term-restart', root).onclick = async () => { try { await termEnsure(key, true); termAttach(key); } catch (e) { fail(e); } };
    $('#term-close', root).onclick = async () => {
      const s = TERMS.sessions.get(key);
      if (s) {
        if (s.es) s.es.close();
        if (s.id) api('DELETE', `/api/terms/${enc(s.id)}`).catch(() => {});
        s.term.dispose();
        TERMS.sessions.delete(key);
      }
      TERMS.current = null;
      go('#/');
    };
    try { await termEnsure(key); termAttach(key); } catch (e) { $('#term-kap', root).innerHTML = `<div class="bos">${icon('alert')} ${esc(e.message)}</div>`; }
  });
  return `<div class="term-sayfa">
    <div class="term-bar">
      <h1>${icon('terminal')}${t('termTitle')}</h1>
      <select id="term-proj"><option value="">${t('termGeneral')}</option>${S.projects.map((p) => `<option value="${esc(p.name)}" ${p.name === key ? 'selected' : ''}>${esc(p.name)}</option>`).join('')}</select>
      <span id="term-durum"></span><span class="mono faint term-dir" id="term-dir"></span>
      <span style="margin-left:auto" class="faint term-ipucu">${t('termHint')}</span>
      <button class="btn kucuk" id="term-restart">${icon('refresh')}${t('termRestart')}</button>
      <button class="ikon-btn" id="term-close" title="${t('termClose')}">${icon('x')}</button>
    </div>
    <div class="term-kap" id="term-kap"><div class="yukleniyor">${icon('refresh', 'spin')} ${t('termOpening')}</div></div>
  </div>`;
}

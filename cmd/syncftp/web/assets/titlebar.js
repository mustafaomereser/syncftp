// Kendi pencere barımız — yalnızca syncFTP'nin WebView2 penceresinde (gui_webview_windows.go,
// window.__SYNCFTP_FRAME). Windows'un başlık çubuğu yok: bu bar sürükleme alanı (CSS app-region:
// drag → WebView2 başlık çubuğu gibi davranır: sürükle, çift tıkla büyüt, sağ tıkta sistem menüsü,
// Aero Snap), sayfa konumu, geri/ileri, çalışan iş sayısı ve pencere düğmeleri. Edge --app'te yok.
(() => {
  const F = window.__SYNCFTP_FRAME;
  if (!F || !F.drag) return;
  const post = (m) => { try { window.chrome.webview.postMessage(m); } catch { /* */ } };
  window.__frameClose = () => post('win:close');

  const svg = (d) => `<svg width="10" height="10" viewBox="0 0 10 10" fill="none" stroke="currentColor" stroke-width="1">${d}</svg>`;
  const ICON = {
    min: svg('<path d="M0 5.5h10"/>'),
    max: svg('<rect x=".5" y=".5" width="9" height="9"/>'),
    restore: svg('<rect x=".5" y="2.5" width="7" height="7"/><path d="M2.5 2.5V.5h7v7h-2"/>'),
    close: svg('<path d="M.5.5l9 9M9.5.5l-9 9"/>'),
  };
  const TAB = { overview: 'tabOverview', servers: 'tabServers', server: 'tabServers', connections: 'tabConnections', connection: 'tabConnections', files: 'tabFiles', history: 'tabHistory', cicd: 'tabCicd', settings: 'tabSettings' };

  function where() {
    const r = parseRoute();
    switch (r.page) {
      case 'home': return [t('projectsTitle')];
      case 'jobs': return [t('cmJobs')];
      case 'settings': return [t('cmSettings')];
      case 'help': return [t('cmHelp')];
      case 'terminal': return ['Terminal', r.sub || ''];
      case 'project': if (r.project === QUICK_P) return [t('qcTitle'), r.sub ? r.sub.replace(/^@conn:/, '') : ''];
        return [r.project, TAB[r.tab] ? t(TAB[r.tab]) : r.tab, r.tab === 'files' && r.sub ? r.sub.replace(/^@conn:/, '') : ''];
    }
    return [];
  }

  // app.js render() ve renderDock() çağırır
  window.tbUpdate = () => {
    const bar = document.getElementById('titlebar');
    if (!bar) return;
    const parts = where().filter(Boolean);
    const yer = bar.querySelector('.tb-yer');
    const txt = parts.join('  ·  ');
    if (yer.textContent !== txt) yer.textContent = txt;
    const running = typeof S !== 'undefined' ? [...S.jobs.values()].filter((j) => j.status === 'running' && !isWatchJob(j)).length : 0;
    const jb = bar.querySelector('.tb-is');
    jb.hidden = !running;
    jb.querySelector('span').textContent = running ? t('tbJobsN', running) : '';
    const nav = window.navigation;
    bar.querySelector('[data-tb="back"]').disabled = !!nav && 'canGoBack' in nav && !nav.canGoBack;
    bar.querySelector('[data-tb="fwd"]').disabled = !!nav && 'canGoForward' in nav && !nav.canGoForward;
    for (const [k, key] of [['back', 'cmBack'], ['fwd', 'cmForward'], ['min', 'tbMin'], ['close', 'tbClose']]) bar.querySelector(`[data-tb="${k}"]`).title = t(key);
    const qc = bar.querySelector('.tb-hizli');
    qc.title = t('qcHint');
    if (qc.lastElementChild.textContent !== t('qcTitle')) qc.lastElementChild.textContent = t('qcTitle');
    const ara = bar.querySelector('.tb-ara span');
    if (ara.textContent !== t('pkPlaceholderShort')) ara.textContent = t('pkPlaceholderShort');
    bar.querySelector('[data-tb="max"]').title = t(document.body.classList.contains('pencere-buyuk') ? 'tbRestore' : 'tbMax');
  };
  window.__winState = (st) => {
    document.body.classList.toggle('pencere-buyuk', st === 'max');
    const b = document.querySelector('#titlebar [data-tb="max"]');
    if (b) b.innerHTML = st === 'max' ? ICON.restore : ICON.max;
    window.tbUpdate();
  };

  window.addEventListener('DOMContentLoaded', () => {
    document.body.classList.add('cerceveli');
    const bar = document.createElement('div');
    bar.id = 'titlebar';
    bar.innerHTML = `<div class="tb-sol"><img src="/logo.png" alt=""><b>syncFTP</b><span class="tb-yer"></span></div>
      <div class="tb-nav"><button type="button" data-tb="back">${icon('left')}</button><button type="button" data-tb="fwd">${icon('right')}</button></div>
      <div class="tb-bosluk"></div>
      <button type="button" class="tb-ara" data-tb="search">${icon('search')}<span></span><kbd>Ctrl K</kbd></button>
      <div class="tb-bosluk"></div>
      <button type="button" class="tb-is" hidden>${icon('refresh', 'spin')}<span></span></button>
      <button type="button" class="tb-guncel" data-tb="update" hidden>${icon('upload')}<span></span></button>
      <button type="button" class="tb-hizli" data-tb="quick">${icon('zap')}<span></span></button>
      <div class="tb-pencere"><button type="button" data-tb="min">${ICON.min}</button><button type="button" data-tb="max">${ICON.max}</button><button type="button" data-tb="close" class="kapat">${ICON.close}</button></div>`;
    document.body.prepend(bar);
    bar.addEventListener('click', (e) => {
      const b = e.target.closest('button');
      if (!b) return;
      if (b.classList.contains('tb-is')) return go('#/jobs');
      switch (b.dataset.tb) {
        case 'back': history.back(); break;
        case 'fwd': history.forward(); break;
        case 'min': post('win:min'); break;
        case 'max': post('win:max'); break;
        case 'close': post('win:close'); break;
        case 'search': if (window.openPalette) window.openPalette(); break;
        case 'quick': openQuickConnect(); break;
        case 'update': openUpdate(); break;
      }
    });
    window.addEventListener('hashchange', () => window.tbUpdate());
    window.tbUpdate();
  });

  // Harici bağlantılar varsayılan tarayıcıda (WebView2 bunları kendi çıplak penceresinde açardı)
  document.addEventListener('click', (e) => {
    const a = e.target.closest && e.target.closest('a[href]');
    if (!a || e.defaultPrevented) return;
    let u;
    try { u = new URL(a.getAttribute('href'), location.href); } catch { return; }
    if ((u.protocol === 'http:' || u.protocol === 'https:') && u.origin !== location.origin) {
      e.preventDefault();
      post('open:' + u.href);
    }
  }, true);
})();

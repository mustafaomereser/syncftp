'use strict';

/* CI/CD sekmesi — deploy tetikleyicileri (gelen webhook), webhook dinleyicisi + trycloudflare tüneli. */

Object.assign(T.tr, {
  tabCicd: 'CI/CD',
  ciIntro: 'Git sağlayıcınız her push\'ta bu bilgisayardaki syncFTP\'ye webhook gönderir; syncFTP isterse önce <code>git pull</code> yapar, sonra seçili sunuculara deploy eder (yedek ve rollback dahil). Panel tray\'de açık kaldığı sürece çalışır; sunucu makinelerde <code>syncftp hook serve --tunnel</code>.',
  ciListener: 'Webhook dinleyicisi', ciLocal: 'Yerel adres', ciPublic: 'Public adres', ciNotPublic: 'Henüz dışarıdan erişilemiyor — tüneli başlatın ya da kendi adresinizi girin',
  ciTunnel: 'trycloudflare tüneli', ciTunnelOn: 'açık', ciTunnelOff: 'kapalı', ciTunnelStart: 'Tüneli başlat', ciTunnelStop: 'Tüneli durdur', ciTunnelStarting: 'Tünel açılıyor…', ciTunnelOpened: 'Tünel açıldı — adres panoya kopyalandı: {0}', ciTunnelClosed: 'Tünel kapatıldı',
  ciInstall: 'cloudflared indir ve başlat', ciInstallHint: 'cloudflared bulunamadı. Resmi sürüm GitHub\'dan (cloudflare/cloudflared) indirilir — ~40 MB.',
  ciTunnelNote: 'trycloudflare adresi her başlatmada değişir; adres değişince sağlayıcıdaki webhook URL\'sini güncelleyin. Kalıcı adres için kendi domain\'inizi (Cloudflare named tunnel, ngrok, reverse proxy) aşağıya girin.',
  ciCustomURL: 'Kalıcı public adres (opsiyonel)', ciCustomHint: 'Örn. https://deploy.sirketim.com — 127.0.0.1:{0} adresine yönlenmeli', ciAutoTunnel: 'Panel açılınca tüneli otomatik başlat',
  ciTriggers: 'Deploy tetikleyicileri', ciNoTriggers: 'Henüz tetikleyici yok', ciNoTriggersSub: 'Bir tetikleyici ekleyin, URL\'yi ve secret\'ı GitHub/GitLab webhook ayarına yapıştırın.',
  ciAdd: 'Tetikleyici ekle', ciEdit: 'Tetikleyiciyi düzenle', ciNew: 'Yeni tetikleyici', ciProvider: 'Sağlayıcı', ciBranch: 'Branch', ciBranchHint: 'Boş = her branch (örn. main)',
  ciServers: 'Hedef sunucular', ciAllServers: 'boş = tüm aktif sunucular', ciGitPull: 'Önce git pull', ciGitPullHint: 'Proje klasöründe git pull --ff-only çalıştırır (repo klonu olmalı)',
  ciDryRun: 'Dry-run', ciDryRunHint: 'Hiçbir şey yükleme, sadece neyin gideceğini logla', ciSecret: 'Secret', ciRegen: 'Yeni secret', ciShow: 'Göster',
  ciRunNow: 'Şimdi deploy et', ciSetup: 'Kurulum', ciDeliveries: 'Son webhook istekleri', ciNoDeliveries: 'Henüz istek gelmedi',
  ciColEvent: 'Olay', ciColResult: 'Sonuç', ciDeleteConfirm: '"{0}" tetikleyicisi silinsin mi?', ciNameHint: 'URL\'de kullanılır: harf, rakam, - ve _',
  ciSetupTitle: '{0} kurulumu', ciEnabled: 'Aktif', ciDeployStarted: 'Deploy başladı', ciDeployTitle: 'Deploy', ciDupName: 'Bu adla bir tetikleyici zaten var: {0}',
  ciGitHub: '<ol><li>Repo → <b>Settings → Webhooks → Add webhook</b></li><li><b>Payload URL</b>: aşağıdaki URL</li><li><b>Content type</b>: <code>application/json</code></li><li><b>Secret</b>: aşağıdaki secret</li><li><b>Just the push event</b> → Add webhook</li></ol>',
  ciGitLab: '<ol><li>Proje → <b>Settings → Webhooks → Add new webhook</b></li><li><b>URL</b>: aşağıdaki URL</li><li><b>Secret token</b>: aşağıdaki secret</li><li><b>Trigger</b>: Push events (isteğe bağlı branch filtresi)</li></ol>',
  ciGitea: '<ol><li>Repo → <b>Settings → Webhooks → Add Webhook → Gitea</b></li><li><b>Target URL</b>: aşağıdaki URL, <b>POST</b>, <code>application/json</code></li><li><b>Secret</b>: aşağıdaki secret</li><li><b>Trigger On</b>: Push Events</li></ol>',
  ciBitbucket: '<ol><li>Repo → <b>Repository settings → Webhooks → Add webhook</b></li><li><b>URL</b>: aşağıdaki URL, <b>Secret</b>: aşağıdaki secret</li><li><b>Triggers</b>: Repository push</li></ol>',
  ciGeneric: '<p>Herhangi bir CI\'dan (GitHub Actions, GitLab CI, Jenkins…) ya da elle tetikleyin:</p>',
});
Object.assign(T.en, {
  tabCicd: 'CI/CD',
  ciIntro: 'Your Git provider sends a webhook to syncFTP on this computer on every push; syncFTP optionally runs <code>git pull</code> and then deploys to the selected servers (with backup and rollback). It works while the panel stays in the tray; on server machines use <code>syncftp hook serve --tunnel</code>.',
  ciListener: 'Webhook listener', ciLocal: 'Local address', ciPublic: 'Public address', ciNotPublic: 'Not reachable from outside yet — start the tunnel or enter your own address',
  ciTunnel: 'trycloudflare tunnel', ciTunnelOn: 'on', ciTunnelOff: 'off', ciTunnelStart: 'Start tunnel', ciTunnelStop: 'Stop tunnel', ciTunnelStarting: 'Opening tunnel…', ciTunnelOpened: 'Tunnel is open — address copied: {0}', ciTunnelClosed: 'Tunnel closed',
  ciInstall: 'Download cloudflared and start', ciInstallHint: 'cloudflared not found. The official release is downloaded from GitHub (cloudflare/cloudflared) — ~40 MB.',
  ciTunnelNote: 'The trycloudflare address changes on every start; update the webhook URL at your provider when it does. For a permanent address enter your own domain (Cloudflare named tunnel, ngrok, reverse proxy) below.',
  ciCustomURL: 'Permanent public address (optional)', ciCustomHint: 'e.g. https://deploy.mycompany.com — must forward to 127.0.0.1:{0}', ciAutoTunnel: 'Start the tunnel when the panel starts',
  ciTriggers: 'Deploy triggers', ciNoTriggers: 'No triggers yet', ciNoTriggersSub: 'Add a trigger and paste its URL and secret into your GitHub/GitLab webhook settings.',
  ciAdd: 'Add trigger', ciEdit: 'Edit trigger', ciNew: 'New trigger', ciProvider: 'Provider', ciBranch: 'Branch', ciBranchHint: 'Empty = any branch (e.g. main)',
  ciServers: 'Target servers', ciAllServers: 'empty = all enabled servers', ciGitPull: 'git pull first', ciGitPullHint: 'Runs git pull --ff-only in the project folder (must be a git clone)',
  ciDryRun: 'Dry-run', ciDryRunHint: 'Upload nothing, only log what would be deployed', ciSecret: 'Secret', ciRegen: 'New secret', ciShow: 'Show',
  ciRunNow: 'Deploy now', ciSetup: 'Setup', ciDeliveries: 'Recent webhook requests', ciNoDeliveries: 'No requests yet',
  ciColEvent: 'Event', ciColResult: 'Result', ciDeleteConfirm: 'Delete trigger "{0}"?', ciNameHint: 'Used in the URL: letters, digits, - and _',
  ciSetupTitle: '{0} setup', ciEnabled: 'Enabled', ciDeployStarted: 'Deploy started', ciDeployTitle: 'Deploy', ciDupName: 'A trigger with this name already exists: {0}',
  ciGitHub: '<ol><li>Repo → <b>Settings → Webhooks → Add webhook</b></li><li><b>Payload URL</b>: the URL below</li><li><b>Content type</b>: <code>application/json</code></li><li><b>Secret</b>: the secret below</li><li><b>Just the push event</b> → Add webhook</li></ol>',
  ciGitLab: '<ol><li>Project → <b>Settings → Webhooks → Add new webhook</b></li><li><b>URL</b>: the URL below</li><li><b>Secret token</b>: the secret below</li><li><b>Trigger</b>: Push events (optional branch filter)</li></ol>',
  ciGitea: '<ol><li>Repo → <b>Settings → Webhooks → Add Webhook → Gitea</b></li><li><b>Target URL</b>: the URL below, <b>POST</b>, <code>application/json</code></li><li><b>Secret</b>: the secret below</li><li><b>Trigger On</b>: Push Events</li></ol>',
  ciBitbucket: '<ol><li>Repo → <b>Repository settings → Webhooks → Add webhook</b></li><li><b>URL</b>: the URL below, <b>Secret</b>: the secret below</li><li><b>Triggers</b>: Repository push</li></ol>',
  ciGeneric: '<p>Trigger from any CI (GitHub Actions, GitLab CI, Jenkins…) or by hand:</p>',
});

function genSecret() {
  const b = new Uint8Array(24);
  crypto.getRandomValues(b);
  return [...b].map((x) => x.toString(16).padStart(2, '0')).join('');
}

const PROVIDER_LABEL = { github: 'GitHub', gitlab: 'GitLab', gitea: 'Gitea', bitbucket: 'Bitbucket', generic: 'Generic / CI' };

function copyBtn(text) {
  return `<button class="ikon-btn" data-copy="${esc(text)}" title="${t('copyPath')}">${icon('copy')}</button>`;
}

// toggleTunnel tüneli açar/kapatır. Açılış 30 sn'ye kadar sürebilir: düğme "açılıyor…" gösterir,
// adres gelince panoya kopyalanır ve bildirilir. (CI/CD sekmesi + Ayarlar sayfası ortak)
async function toggleTunnel(btn, tun, hooks) {
  btn.disabled = true;
  try {
    if (tun.active) {
      await api('POST', '/api/hooks/tunnel', { on: false });
      toast(t('ciTunnelClosed'));
    } else {
      btn.innerHTML = `${icon('refresh', 'spin')}${t('ciTunnelStarting')}`;
      const r = await api('POST', '/api/hooks/tunnel', { on: true, install: !hooks.cloudflared });
      if (r.job) { trackJob(r.job, { title: t('ciInstall') }); showJobLog(r.job); }
      else if (r.url) {
        navigator.clipboard.writeText(r.url).catch(() => {});
        toast(t('ciTunnelOpened', r.url));
      }
    }
  } catch (e) { fail(e); }
  render(true);
}

// Tünel denetimi (durum + aç/kapa + kalıcı adres + otomatik başlat) — CI/CD sekmesi ve Ayarlar → Webhook
// aynı parçayı kullanır; id öneki öğe kimliklerini ayırır (ci-tunnel / set-tunnel …).
function tunnelPillHTML(tun) {
  if (tun.active && tun.url) return `<span class="pill ok"><span class="nokta"></span>${t('ciTunnelOn')}</span>`;
  if (tun.active) return `<span class="pill busy"><span class="nokta"></span>${t('ciTunnelStarting')}</span>`;
  return `<span class="pill"><span class="nokta"></span>${t('ciTunnelOff')}</span>`;
}
function tunnelCtlHTML(hooks, id) {
  const tun = hooks.tunnel || {};
  const localPort = (hooks.local_url || '').split(':').pop();
  return `<div class="aksiyonlar" style="margin-bottom:10px"><b style="font-weight:600">${t('ciTunnel')}</b> ${tunnelPillHTML(tun)}
      ${tun.url ? `<span class="mono faint">${esc(tun.url)}</span>` : ''}${tun.error && !tun.active ? `<span class="bad-t" style="font-size:12px">${esc(tun.error)}</span>` : ''}
      <span style="margin-left:auto"></span>
      <button class="btn ${tun.active ? 'tehlike' : 'mavi'}" id="${id}-tunnel">${icon(tun.active ? 'stop' : 'play')}${tun.active ? t('ciTunnelStop') : hooks.cloudflared ? t('ciTunnelStart') : t('ciInstall')}</button></div>
    ${hooks.cloudflared ? `<p class="ipucu mono">${esc(hooks.cloudflared)}</p>` : !tun.active ? `<p class="ipucu">${t('ciInstallHint')}</p>` : ''}
    <p class="ipucu">${t('ciTunnelNote')}</p>
    <div class="alanlar" style="margin-top:12px">
      <div class="alan"><label>${t('ciCustomURL')}</label><input type="text" class="mono" id="${id}-custom" value="${esc(hooks.custom_url || '')}" placeholder="https://…">${localPort ? `<small>${t('ciCustomHint', esc(localPort))}</small>` : ''}</div>
      <div class="alan"><label>&nbsp;</label>
        <label class="anahtar"><span class="metin"><b>${t('ciAutoTunnel')}</b></span><span class="switch"><input type="checkbox" id="${id}-auto" ${hooks.auto_tunnel ? 'checked' : ''}><i></i></span></label></div>
    </div>
    <div class="aksiyonlar" style="margin-top:10px;justify-content:flex-end"><button class="btn" id="${id}-save-url">${icon('check')}${t('save')}</button></div>`;
}
function bindTunnelCtl(root, hooks, id) {
  const tb = $(`#${id}-tunnel`, root);
  if (tb) tb.onclick = () => toggleTunnel(tb, hooks.tunnel || {}, hooks);
  const sv = $(`#${id}-save-url`, root);
  if (sv) sv.onclick = async () => {
    try {
      await api('PUT', '/api/hooks/settings', { custom_url: $(`#${id}-custom`, root).value, auto_tunnel: $(`#${id}-auto`, root).checked });
      toast(t('saved')); render(true);
    } catch (e) { fail(e); }
  };
}

async function tabCicd(proj) {
  const P = proj.name;
  const [hooks, dep, cfg] = await Promise.all([GET('/api/hooks'), GET(`/api/p/${enc(P)}/deploy`), getConfig(P)]);
  const isLocalOnly = dep.local_only;

  bind((root) => {
    $('#proj-aks', root).innerHTML = `<button class="ikon-btn" id="ci-rf" title="${t('refresh')}">${icon('refresh')}</button><button class="btn birincil" id="ci-add">${icon('plus')}${t('ciAdd')}</button>`;
    $('#ci-rf', root).onclick = () => render(true);
    $('#ci-add', root).onclick = () => triggerModal(P, null);
    bindTunnelCtl(root, hooks, 'ci');
    $$('[data-trig]', root).forEach((card) => {
      const name = card.dataset.trig;
      const d = cfg.deploy.find((x) => x.name === name);
      const it = dep.triggers.find((x) => x.name === name);
      const on = (sel, fn) => { const b = $(sel, card); if (b) b.onclick = fn; };
      on('[data-act="run"]', async () => {
        try { const r = await api('POST', `/api/p/${enc(P)}/deploy/${enc(name)}/run`); if (r.job) { trackJob(r.job, { kind: 'deploy', project: P, title: t('ciDeployTitle') }); toast(t('ciDeployStarted')); } } catch (e) { fail(e); }
      });
      on('[data-act="setup"]', () => setupModal(d, it.url));
      on('[data-act="edit"]', () => triggerModal(P, d));
      on('[data-act="del"]', async () => {
        if (!(await confirmBox(t('delete'), esc(t('ciDeleteConfirm', name)), t('delete'), true))) return;
        try { await mutateDeploy(P, (list) => list.filter((x) => x.name !== name)); render(true); } catch (e) { fail(e); }
      });
      on('[data-act="show"]', (e) => { const s = $('[data-secret]', card); s.textContent = s.textContent.includes('•') ? d.secret : '•'.repeat(16); });
      const tg = $('[data-act="toggle"]', card);
      if (tg) tg.onchange = async () => {
        const v = tg.checked;
        try {
          await mutateDeploy(P, (list) => { const x = list.find((y) => y.name === name); if (!x) throw new Error(t('notFound') + ': ' + name); x.enabled = v; return list; });
          toast(t('saved'));
        } catch (e) { tg.checked = !v; fail(e); }
      };
    });
    $$('[data-djob]', root).forEach((tr) => (tr.onclick = () => showJobLog(tr.dataset.djob)));
  });

  const listenerCard = `<div class="bolum" style="margin-bottom:18px"><h3>${icon('link')}${t('ciListener')}</h3><p class="aciklama">${t('ciIntro')}</p>
    <div class="alanlar">
      <div class="alan"><label>${t('ciLocal')}</label><div class="girdi-grup"><input type="text" class="mono" value="${esc(hooks.local_url)}" readonly>${copyBtn(hooks.local_url)}</div></div>
      <div class="alan"><label>${t('ciPublic')}</label><div class="girdi-grup"><input type="text" class="mono" value="${esc(hooks.public_url)}" readonly>${copyBtn(hooks.public_url)}</div>
        ${isLocalOnly ? `<small class="warn-t">${t('ciNotPublic')}</small>` : ''}</div>
    </div>
    <div class="ayrac"></div>
    ${tunnelCtlHTML(hooks, 'ci')}
  </div>`;

  const trigCards = dep.triggers.length ? `<div class="kartlar">${dep.triggers.map((d) => `<div class="kart ${d.enabled ? '' : 'pasif'}" data-trig="${esc(d.name)}">
      <div><div class="ust"><h3>${esc(d.name)}</h3><span class="pill vio">${esc(PROVIDER_LABEL[d.provider] || d.provider)}</span>
        ${d.branch ? `<span class="pill mono">${esc(d.branch)}</span>` : ''}${d.git_pull ? '<span class="pill">git pull</span>' : ''}${d.dry_run ? `<span class="pill warn">${t('ciDryRun')}</span>` : ''}</div>
        <div class="meta"><span>${icon('server')}${esc((d.servers && d.servers.length) ? d.servers.join(', ') : t('ciAllServers'))}</span></div>
        <div class="girdi-grup" style="margin-top:10px"><input type="text" class="mono" value="${esc(d.url)}" readonly style="font-size:12px">${copyBtn(d.url)}</div>
        <div class="aksiyonlar" style="margin-top:6px;font-size:12px"><span class="faint">${t('ciSecret')}:</span><code data-secret>${'•'.repeat(16)}</code>
          <button class="btn hayalet kucuk" data-act="show" style="height:24px">${icon('eye')}${t('ciShow')}</button>${copyBtn(d.secret)}</div></div>
      <div class="butonlar" style="flex-direction:column;align-items:flex-end;gap:8px">
        <label class="switch" title="${t('ciEnabled')}"><input type="checkbox" data-act="toggle" ${d.enabled ? 'checked' : ''}><i></i></label>
        <div class="aksiyonlar"><button class="btn kucuk mavi" data-act="run">${icon('upload')}${t('ciRunNow')}</button>
          <button class="btn kucuk" data-act="setup">${icon('log')}${t('ciSetup')}</button>
          <button class="ikon-btn" data-act="edit" title="${t('edit')}">${icon('edit')}</button>
          <button class="ikon-btn tehlike" data-act="del" title="${t('delete')}">${icon('trash')}</button></div></div>
    </div>`).join('')}</div>` : `<div class="bos"><div class="buyuk">${icon('zap', 'lg')}</div><h3>${t('ciNoTriggers')}</h3><p>${t('ciNoTriggersSub')}</p></div>`;

  const st = (s) => ({ accepted: 'ok', ignored: '', rejected: 'bad', error: 'bad' }[s] || '');
  const deliveries = dep.deliveries.length ? `<div class="tablo-kap"><table class="tablo"><thead><tr><th>${t('colWhen')}</th><th>${t('colName')}</th><th>${t('ciColEvent')}</th><th>${t('ciBranch')}</th><th>${t('ciColResult')}</th></tr></thead><tbody>
      ${dep.deliveries.map((d) => `<tr ${d.job_id ? `class="tikla" data-djob="${esc(d.job_id)}"` : ''}><td class="mono">${fmtDate(d.time)}</td><td>${esc(d.trigger)}</td><td class="mono faint">${esc(d.event || '')}</td>
        <td class="mono">${esc(d.branch || '')} <span class="faint">${esc(d.commit || '')}</span></td><td><span class="pill ${st(d.status)}"><span class="nokta"></span>${esc(d.status)}</span> <span class="faint" style="font-size:12px">${esc(d.detail || '')}</span></td></tr>`).join('')}</tbody></table></div>`
    : `<div class="bos" style="padding:24px">${t('ciNoDeliveries')}</div>`;

  return `${listenerCard}<h2 style="margin-top:6px">${icon('zap')}${t('ciTriggers')}</h2>${trigCards}<h2>${icon('history')}${t('ciDeliveries')}</h2>${deliveries}`;
}

// Tetikleyici listesini güncel config üzerinde değiştirip yazar (sayfa açıldıktan sonra başka yerde
// yapılan config değişiklikleri ezilmesin). fn(list) yeni listeyi döner.
async function mutateDeploy(P, fn) {
  const next = await getConfig(P);
  next.deploy = fn(next.deploy || []);
  await api('PUT', `/api/p/${enc(P)}/config`, next);
}

async function triggerModal(P, d) {
  const isNew = !d;
  const cfg = await getConfig(P).catch((e) => { fail(e); return null; });
  if (!cfg) return;
  const cur = d ? JSON.parse(JSON.stringify(d)) : { name: 'main', provider: 'github', secret: genSecret(), branch: 'main', servers: [], git_pull: true, dry_run: false, enabled: true };
  const srvChecks = cfg.servers.map((s) => `<label class="anahtar" style="padding:7px 10px"><span class="metin"><b>${esc(s.name)}</b></span><span class="switch"><input type="checkbox" data-trsrv="${esc(s.name)}" ${(cur.servers || []).includes(s.name) ? 'checked' : ''}><i></i></span></label>`).join('');
  modal({
    title: `${icon('zap')}${isNew ? t('ciNew') : t('ciEdit')}`, size: 'orta',
    body: `<div class="form">
      <div class="alanlar">
        ${fText('name', t('fName'), cur.name, { hint: t('ciNameHint'), mono: true })}
        ${fSelect('provider', t('ciProvider'), cur.provider, Object.entries(PROVIDER_LABEL))}
        ${fText('branch', t('ciBranch'), cur.branch, { hint: t('ciBranchHint'), mono: true })}
        <div class="alan"><label>${t('ciSecret')}</label><div class="girdi-grup"><input type="text" class="mono" data-k="secret" data-ty="text" value="${esc(cur.secret)}"><button class="btn" data-regen>${icon('refresh')}${t('ciRegen')}</button></div></div>
      </div>
      <div class="alan"><label>${t('ciServers')} <span class="faint">(${t('ciAllServers')})</span></label><div class="alanlar">${srvChecks}</div></div>
      <div class="alanlar">${fBool('git_pull', t('ciGitPull'), cur.git_pull, t('ciGitPullHint'))}${fBool('dry_run', t('ciDryRun'), cur.dry_run, t('ciDryRunHint'))}${fBool('enabled', t('ciEnabled'), cur.enabled)}</div>
    </div>`,
    actions: [{ label: t('cancel'), cls: 'hayalet' }, { label: icon('check') + t('save'), cls: 'birincil', run: async (root) => {
      const out = readForm($('.form', root), cur);
      out.name = out.name.replace(/[^A-Za-z0-9_-]/g, '-');
      out.servers = $$('[data-trsrv]', root).filter((c) => c.checked).map((c) => c.dataset.trsrv);
      if (!out.name) return false;
      // Hata fırlarsa modal açık kalır ve hata gösterilir (aynı adlı tetikleyici → config yazılmaz)
      await mutateDeploy(P, (list) => {
        if (list.some((x) => x.name === out.name && (!d || x.name !== d.name))) throw new Error(t('ciDupName', out.name));
        const i = d ? list.findIndex((x) => x.name === d.name) : -1;
        if (i >= 0) list[i] = out; else list.push(out);
        return list;
      });
      toast(t('saved')); render(true);
    } }],
  }).root.querySelector('[data-regen]').onclick = (e) => { e.preventDefault(); e.currentTarget.parentElement.querySelector('input').value = genSecret(); };
}

function setupModal(d, url) {
  const guide = { github: t('ciGitHub'), gitlab: t('ciGitLab'), gitea: t('ciGitea'), bitbucket: t('ciBitbucket'), generic: t('ciGeneric') }[d.provider] || '';
  const curl = `curl -X POST -H "X-Syncftp-Token: ${d.secret}" ${url}`;
  modal({
    title: `${icon('log')}${t('ciSetupTitle', esc(PROVIDER_LABEL[d.provider] || d.provider))}`, size: 'orta',
    body: `<div class="kurulum">${guide}</div>
      <div class="alan" style="margin-top:12px"><label>URL</label><div class="girdi-grup"><input type="text" class="mono" value="${esc(url)}" readonly>${copyBtn(url)}</div></div>
      <div class="alan" style="margin-top:10px"><label>${t('ciSecret')}</label><div class="girdi-grup"><input type="text" class="mono" value="${esc(d.secret)}" readonly>${copyBtn(d.secret)}</div></div>
      ${d.provider === 'generic' ? `<div class="alan" style="margin-top:10px"><label>curl</label><div class="girdi-grup"><input type="text" class="mono" value="${esc(curl)}" readonly>${copyBtn(curl)}</div></div>` : ''}`,
    actions: [{ label: t('close'), cls: 'birincil' }],
  });
}

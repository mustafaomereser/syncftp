'use strict';

/* Yapay zekâ asistanı çekmecesi — yüklü AI araçlarından biriyle (Claude Code, OpenCode)
   panel içinde sohbet. */

Object.assign(T.tr, {
  aiTitle: 'Yapay zekâ asistanı', aiNav: 'AI asistan', aiPick: 'Hangi araçla açalım?', aiPickSub: 'Bu bilgisayarda yüklü yapay zekâ araçları. Asistan syncFTP dokümanlarını ve projenin şifresiz özetini okuyarak yardım eder; isterseniz sizin yerinize ayar yapar.',
  aiChat: 'Panel içi sohbet', aiNotInstalled: 'yüklü değil', aiProject: 'Proje',
  aiNew: 'Yeni sohbet', aiStop: 'Durdur', aiSend: 'Gönder', aiPlaceholder: 'Sorun ya da yaptırın… (Enter gönderir, Shift+Enter yeni satır)',
  aiThinking: 'düşünüyor…', aiNoProject: 'Önce bir proje ekleyin', aiNoAgents: 'Yüklü yapay zekâ aracı bulunamadı. Claude Code ya da OpenCode kurun.',
  aiPerms: 'İzinler: dosya düzenleme + yalnızca syncftp komutları. Gerçek sync, rollback ve silme gibi işlemlerden önce size sorar.',
  aiS1: 'Bu projede neler var, özetler misin?', aiS2: 'Yeni bir SFTP sunucusu eklemek istiyorum, adım adım sorarak ekle', aiS3: 'GitHub\'a push yapınca otomatik deploy olsun, kur', aiS4: 'Neden bu dosyalar yükleniyor? Bekleyen değişiklikleri açıkla',
  aiS5: 'vendor ve node_modules\'ü sync dışında bırak, .env\'yi koru', aiS6: 'Deploy sonrası sunucuda php artisan migrate çalışsın',
  aiCost: 'maliyet', aiChanged: 'Asistan bir değişiklik yaptı — sayfa yenilendi',
  aiSetupNav: 'AI ile proje kur', aiSetupTitle: 'Proje kurulum yardımcısı', aiSetupSub: 'Asistan klasörü inceler (Laravel, WordPress, Node…), size nereye deploy edeceğinizi sorar ve projeyi sizin yerinize kurar. Şifre istemez — kayıtlı bağlantınızı kullanır ya da şifreyi sonra panelden girersiniz.',
  aiSetupFolder: 'Proje klasörü', aiSetupLink: 'Yeni proje kur (AI)', aiSetupBack: 'Mevcut projeyle sohbet', aiSetupNeedFolder: 'Önce proje klasörünü seçin',
  aiSS1: 'Bu klasörü incele ve syncFTP projesi olarak kur', aiSS2: 'Kayıtlı bağlantılarımdan biriyle kur', aiSS3: 'Bunu SFTP ile sunucudaki /var/www klasörüne deploy etmek istiyorum', aiSS4: 'Önce ne yapacağını anlat, sonra kur',
  aiPermsSetup: 'İzinler: klasörü okuma + yalnızca .syncftp/ altına yazma ve syncftp komutları. Şifre sormaz; kurmadan önce komutu gösterip onayınızı alır.',
});
Object.assign(T.en, {
  aiTitle: 'AI assistant', aiNav: 'AI assistant', aiPick: 'Which tool should we use?', aiPickSub: 'AI tools installed on this computer. The assistant reads the syncFTP docs and a secret-free summary of the project; it can also change settings for you.',
  aiChat: 'In-panel chat', aiNotInstalled: 'not installed', aiProject: 'Project',
  aiNew: 'New chat', aiStop: 'Stop', aiSend: 'Send', aiPlaceholder: 'Ask or ask it to do something… (Enter sends, Shift+Enter new line)',
  aiThinking: 'thinking…', aiNoProject: 'Add a project first', aiNoAgents: 'No AI tool found. Install Claude Code or OpenCode.',
  aiPerms: 'Permissions: file edits + syncftp commands only. It asks you before real syncs, rollbacks and deletions.',
  aiS1: 'What is in this project? Give me a summary', aiS2: 'I want to add a new SFTP server — ask me step by step and add it', aiS3: 'Deploy automatically when I push to GitHub — set it up', aiS4: 'Why are these files being uploaded? Explain the pending changes',
  aiS5: 'Keep vendor and node_modules out of sync and protect .env', aiS6: 'Run php artisan migrate on the server after each deploy',
  aiCost: 'cost', aiChanged: 'The assistant made a change — page refreshed',
  aiSetupNav: 'Set up a project with AI', aiSetupTitle: 'Project setup assistant', aiSetupSub: 'The assistant inspects the folder (Laravel, WordPress, Node…), asks where to deploy and sets the project up for you. It never asks for passwords — it uses a saved connection or you enter the password in the panel later.',
  aiSetupFolder: 'Project folder', aiSetupLink: 'Set up a new project (AI)', aiSetupBack: 'Chat about an existing project', aiSetupNeedFolder: 'Choose the project folder first',
  aiSS1: 'Inspect this folder and set it up as a syncFTP project', aiSS2: 'Set it up with one of my saved connections', aiSS3: 'I want to deploy this over SFTP to /var/www on my server', aiSS4: 'Explain what you will do first, then set it up',
  aiPermsSetup: 'Permissions: read the folder + write only under .syncftp/ and syncftp commands. It never asks for passwords and shows the command before running it.',
});

const AI = { open: false, project: '', session: null, es: null, lastId: 0, agents: null, setupPath: '' };

function aiCurrentProject() {
  const r = parseRoute();
  if (r.project) return r.project;
  return AI.project || (S.projects[0] && S.projects[0].name) || '';
}

function aiToggle(force) {
  AI.open = force ?? !AI.open;
  let el = $('#ai-drawer');
  if (!AI.open) { if (el) el.classList.remove('acik'); return; }
  if (!el) {
    el = document.createElement('aside');
    el.id = 'ai-drawer';
    el.className = 'ai-drawer';
    document.body.appendChild(el);
  }
  requestAnimationFrame(() => el.classList.add('acik'));
  const p = aiCurrentProject();
  if (AI.session && (AI.session.setup || AI.session.project === p)) aiRenderChat();
  else { AI.project = p; aiRenderPicker(); }
}

// aiSetupStart kurulum yardımcısını açar (path verilirse klasör hazır gelir).
function aiSetupStart(path) {
  if (AI.session && AI.session.setup && (!path || AI.session.folder === path)) { aiToggle(true); return; }
  if (AI.es) AI.es.close();
  AI.session = null;
  AI.setupPath = path || AI.setupPath || (S.meta && S.meta.start_dir) || '';
  AI.open = false;
  aiToggle(true);
  aiRenderPicker(true);
}

async function aiRenderPicker(setup = false) {
  const el = $('#ai-drawer');
  el.innerHTML = `<div class="ai-bas"><h3>${icon('spark')}${t('aiTitle')}</h3><button class="ikon-btn" data-ai-close>${icon('x')}</button></div>
    <div class="ai-govde"><div class="yukleniyor">${icon('refresh', 'spin')} ${t('loading')}</div></div>`;
  $('[data-ai-close]', el).onclick = () => aiToggle(false);
  if (!S.projects.length) await loadProjects();
  // Yalnızca panel içinde sohbet edilebilen araçlar
  const agents = (await GET('/api/ai/agents').catch(() => [])).filter((a) => a.chat);
  AI.agents = agents;
  const found = agents.filter((a) => a.found);
  const body = $('.ai-govde', el);
  setup = setup || !S.projects.length;
  const head = setup
    ? `<h4 class="ai-h4">${icon('plus')} ${t('aiSetupTitle')}</h4><p class="ipucu" style="margin:0 0 12px">${t('aiSetupSub')}</p>
      <div class="alan" style="margin-bottom:14px"><label>${t('aiSetupFolder')}</label><div class="girdi-grup"><input type="text" class="mono" id="ai-folder" value="${esc(AI.setupPath)}" spellcheck="false">
        <button class="btn" id="ai-folder-pick">${icon('folder')}</button></div></div>
      ${S.projects.length ? `<p style="margin:0 0 14px"><a href="javascript:void 0" id="ai-to-proj" class="acc-t">${icon('left')} ${t('aiSetupBack')}</a></p>` : ''}`
    : `<div class="alan" style="margin-bottom:8px"><label>${t('aiProject')}</label><select id="ai-proj">${S.projects.map((p) => `<option ${p.name === AI.project ? 'selected' : ''}>${esc(p.name)}</option>`).join('')}</select></div>
      <p style="margin:0 0 14px"><a href="javascript:void 0" id="ai-to-setup" class="acc-t">${icon('plus')} ${t('aiSetupLink')}</a></p>`;
  body.innerHTML = `${head}
    <h4 class="ai-h4">${t('aiPick')}</h4><p class="ipucu" style="margin:0 0 12px">${setup ? '' : t('aiPickSub')}</p>
    ${found.length ? '' : `<div class="bos" style="padding:20px">${t('aiNoAgents')}</div>`}
    <div class="ai-araclar">${agents.map((a) => `<div class="ai-arac ${a.found ? '' : 'yok'}" data-agent="${esc(a.id)}">
        <div class="ai-arac-ust"><b>${esc(a.name)}</b>${a.found ? `<span class="pill ok"><span class="nokta"></span>${t('aiChat')}</span>` : `<span class="pill">${t('aiNotInstalled')}</span>`}</div>
        ${a.found ? `<div class="mono faint ai-yol">${esc(a.path)}</div><div class="aksiyonlar" style="margin-top:8px">
          <button class="btn kucuk mavi" data-ai-chat>${icon('spark')}${t('aiChat')}</button>
</div>` : ''}
      </div>`).join('')}</div>
    <p class="ipucu" style="margin-top:14px">${icon('shield')} ${t(setup ? 'aiPermsSetup' : 'aiPerms')}</p>`;
  const ps = $('#ai-proj', body); if (ps) ps.onchange = (e) => { AI.project = e.target.value; };
  const ts = $('#ai-to-setup', body); if (ts) ts.onclick = () => aiRenderPicker(true);
  const tp = $('#ai-to-proj', body); if (tp) tp.onclick = () => aiRenderPicker(false);
  const fp = $('#ai-folder-pick', body);
  if (fp) fp.onclick = async () => { const v = await pickPath({ mode: 'dir', initial: $('#ai-folder', body).value.trim() }); if (v) { $('#ai-folder', body).value = v; AI.setupPath = v; } };
  $$('[data-agent]', body).forEach((card) => {
    const id = card.dataset.agent;
    const chat = $('[data-ai-chat]', card);
    if (chat) chat.onclick = async () => {
      chat.disabled = true;
      try {
        if (setup) {
          const folder = $('#ai-folder', body).value.trim();
          if (!folder) { toast(t('aiSetupNeedFolder'), 'bad'); chat.disabled = false; return; }
          AI.setupPath = folder;
          AI.session = await api('POST', '/api/ai/setup', { agent: id, path: folder });
        } else AI.session = await api('POST', `/api/p/${enc(AI.project)}/ai/sessions`, { agent: id });
        AI.lastId = 0; aiRenderChat();
      }
      catch (e) { fail(e); chat.disabled = false; }
    };
  });
}

function aiRenderChat() {
  const el = $('#ai-drawer');
  const s = AI.session;
  const sugs = s.setup ? ['aiSS1', 'aiSS2', 'aiSS3', 'aiSS4'] : ['aiS1', 'aiS2', 'aiS3', 'aiS4', 'aiS5', 'aiS6'];
  el.innerHTML = `<div class="ai-bas"><h3>${icon('spark')}${esc(s.agent_name)} <span class="faint" style="font-weight:400;font-size:13px" title="${esc(s.folder || '')}">· ${s.setup ? t('aiSetupTitle') + ' · ' : ''}${esc(s.project)}</span></h3>
      <span class="ai-maliyet faint mono" id="ai-cost"></span>
      <button class="ikon-btn" data-ai-new title="${t('aiNew')}">${icon('plus')}</button>
      <button class="ikon-btn" data-ai-close>${icon('x')}</button></div>
    <div class="ai-akis" id="ai-akis"><div class="ai-oneriler" id="ai-oneriler">
      ${sugs.map((k) => `<button class="ai-oneri" data-sug="${esc(t(k))}">${esc(t(k))}</button>`).join('')}
      <p class="ipucu" style="margin:8px 2px 0">${icon('shield')} ${t(s.setup ? 'aiPermsSetup' : 'aiPerms')}</p></div></div>
    <div class="ai-giris"><textarea id="ai-text" rows="2" placeholder="${t('aiPlaceholder')}"></textarea>
      <button class="btn mavi" id="ai-send">${icon('upload')}</button><button class="btn tehlike" id="ai-stop" hidden>${icon('stop')}</button></div>`;
  $('[data-ai-close]', el).onclick = () => aiToggle(false);
  $('[data-ai-new]', el).onclick = async () => {
    if (AI.es) AI.es.close();
    api('DELETE', `/api/ai/sessions/${enc(s.id)}`).catch(() => {});
    AI.session = null;
    aiRenderPicker(s.setup);
  };
  const ta = $('#ai-text', el);
  const send = async (text) => {
    text = (text ?? ta.value).trim();
    if (!text) return;
    ta.value = '';
    try { await api('POST', `/api/ai/sessions/${enc(s.id)}/send`, { text }); } catch (e) { fail(e); }
  };
  $('#ai-send', el).onclick = () => send();
  $('#ai-stop', el).onclick = () => api('POST', `/api/ai/sessions/${enc(s.id)}/stop`).catch(fail);
  ta.onkeydown = (e) => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); send(); } };
  $$('[data-sug]', el).forEach((b) => (b.onclick = () => send(b.dataset.sug)));
  ta.focus();

  // Akış: mevcut mesajlar + canlı
  if (AI.es) AI.es.close();
  AI.lastId = 0;
  AI.busy = false;
  AI.changed = false;
  AI.es = new EventSource(`/api/ai/sessions/${enc(s.id)}/stream`);
  AI.es.onmessage = (e) => { try { aiOnMsg(JSON.parse(e.data)); } catch { /* yoksay */ } };
}

function aiOnMsg(m) {
  if (m.id <= AI.lastId) return;
  AI.lastId = m.id;
  const box = $('#ai-akis');
  if (!box) return;
  const sug = $('#ai-oneriler');
  if (sug && m.role === 'user') sug.remove();
  const stick = box.scrollTop + box.clientHeight >= box.scrollHeight - 40;
  const add = (html) => box.insertAdjacentHTML('beforeend', html);
  switch (m.role) {
    case 'user': add(`<div class="ai-msg ai-user">${esc(m.text)}</div>`); break;
    case 'assistant': add(`<div class="ai-msg ai-asst">${mdLite(m.text)}</div>`); break;
    case 'tool':
      if (/edit|write/i.test(m.tool) || /syncftp(\.exe)?\s+(sync|rollback|secure|detect|milestone|freeze|remote|project|hook|init|global)/.test(m.text)) AI.changed = true;
      add(`<details class="ai-tool"><summary>${icon('terminal')}<b>${esc(m.tool)}</b> <span class="mono">${esc(m.text)}</span></summary><pre class="ai-tool-out" hidden></pre></details>`); break;
    case 'tool-result': case 'tool-error': {
      const last = [...box.querySelectorAll('.ai-tool')].pop();
      if (last && m.detail) { const pre = last.querySelector('.ai-tool-out'); pre.hidden = false; pre.textContent = m.detail; if (m.role === 'tool-error') last.classList.add('hata'); }
      break;
    }
    case 'error': add(`<div class="ai-msg ai-hata">${icon('alert')} ${esc(m.text)}</div>`); break;
    case 'info': { const c = $('#ai-cost'); if (c) { AI.cost = (AI.cost || 0) + (parseFloat(String(m.text).replace('$', '')) || 0); c.textContent = `${t('aiCost')} $${AI.cost.toFixed(3)}`; } break; }
    case 'state': {
      AI.busy = m.text === 'busy';
      $('#ai-send').hidden = AI.busy; $('#ai-stop').hidden = !AI.busy;
      let ty = $('#ai-typing');
      if (AI.busy && !ty) add(`<div class="ai-typing" id="ai-typing"><span></span><span></span><span></span> ${t('aiThinking')}</div>`);
      if (!AI.busy) {
        if (ty) ty.remove();
        // Asistan ayar/komut çalıştırdıysa açık sayfayı ve proje listesini tazele
        if (AI.changed) { AI.changed = false; loadProjects(); render(true); toast(t('aiChanged')); }
      }
      break;
    }
  }
  const ty = $('#ai-typing');
  if (ty && ty !== box.lastElementChild) box.appendChild(ty);
  if (stick || m.role === 'user') box.scrollTop = box.scrollHeight;
}

// Hafif markdown: kod blokları (renklendirilmiş), satır içi kod, kalın, başlık, liste, tablo
function mdLite(src) {
  const out = [];
  const lines = String(src).replace(/\r\n/g, '\n').split('\n');
  let i = 0;
  const inline = (s) => esc(s)
    .replace(/`([^`]+)`/g, '<code>$1</code>')
    .replace(/\*\*([^*]+)\*\*/g, '<b>$1</b>')
    .replace(/(^|[^*])\*([^*\n]+)\*/g, '$1<i>$2</i>')
    .replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<u title="$2">$1</u>');
  while (i < lines.length) {
    const l = lines[i];
    const fence = l.match(/^```(\w*)/);
    if (fence) {
      const code = [];
      i++;
      while (i < lines.length && !lines[i].startsWith('```')) code.push(lines[i++]);
      i++;
      const ext = { php: 'a.php', js: 'a.js', ts: 'a.ts', json: 'a.json', jsonc: 'a.json', css: 'a.css', html: 'a.html', sql: 'a.sql', yaml: 'a.yml', yml: 'a.yml', sh: 'a.sh', bash: 'a.sh', go: 'a.go', ini: 'a.ini', env: '.env' }[fence[1]] || '';
      out.push(`<pre class="ai-kod">${HL.hlLines(code.join('\n'), ext).join('\n')}</pre>`);
      continue;
    }
    if (/^\s*\|/.test(l)) {
      const rows = [];
      while (i < lines.length && /^\s*\|/.test(lines[i])) rows.push(lines[i++]);
      const cells = rows.filter((r) => !/^\s*\|[\s:|-]+\|\s*$/.test(r)).map((r) => r.trim().replace(/^\||\|$/g, '').split('|').map((c) => inline(c.trim())));
      out.push(`<div class="ai-tablo"><table>${cells.map((r, ri) => `<tr>${r.map((c) => (ri ? `<td>${c}</td>` : `<th>${c}</th>`)).join('')}</tr>`).join('')}</table></div>`);
      continue;
    }
    if (/^\s*([-*]|\d+\.)\s+/.test(l)) {
      const ordered = /^\s*\d+\./.test(l);
      const items = [];
      while (i < lines.length && /^\s*([-*]|\d+\.)\s+/.test(lines[i])) items.push(lines[i++].replace(/^\s*([-*]|\d+\.)\s+/, ''));
      out.push(`<${ordered ? 'ol' : 'ul'}>${items.map((x) => `<li>${inline(x)}</li>`).join('')}</${ordered ? 'ol' : 'ul'}>`);
      continue;
    }
    const h = l.match(/^(#{1,4})\s+(.*)/);
    if (h) { out.push(`<p class="ai-baslik">${inline(h[2])}</p>`); i++; continue; }
    if (!l.trim()) { i++; continue; }
    const para = [];
    while (i < lines.length && lines[i].trim() && !/^```|^\s*\||^\s*([-*]|\d+\.)\s+|^#{1,4}\s/.test(lines[i])) para.push(lines[i++]);
    out.push(`<p>${para.map(inline).join('<br>')}</p>`);
  }
  return out.join('');
}

'use strict';

/* ═══════════════════════════════════════════════════════════════════════════
   Uzak dosya düzenleyici — editFile(P, server, path)
   Dosyanın TAMAMI gelir (GET /remote/edit, 4 MB'a kadar; önizlemedeki 256 KB sınırı yok).
   Editör: şeffaf textarea + altında renklendirilmiş <pre> (HL.hlLines) + satır numaraları.
   Tab/Shift+Tab girinti, Enter girintiyi korur, Ctrl+S kaydet. Kaydederken değişiklikler fark olarak
   gösterilir; "sunucu + yerel" (varsayılan) ya da "yalnızca sunucu" seçilir. Sunucu tarafı normal
   sync'in upload çekirdeğini kullanır → eski hal yedeklenir, Geçmiş'ten geri alınabilir.
   Sunucudaki dosya açıldıktan sonra değiştiyse (hash tutmuyor) çakışma penceresi açılır.
   app.js + highlight.js'ten sonra yüklenir.
   ═══════════════════════════════════════════════════════════════════════════ */

Object.assign(T.tr, {
  edEdit: 'Düzenle', edSave: 'Kaydet', edSaving: 'Kaydediliyor…', edSaved: 'Kaydedildi', edModified: 'değişti',
  edLine: 'Satır {0}, Sütun {1}', edNoHl: 'büyük dosya — renklendirme kapalı',
  edLocalSame: 'Yerel dosya aynı', edLocalDiff: 'Yerel dosya farklı', edLocalMissing: 'Yerelde yok',
  edNotEditable: 'Bu dosya düzenlenemez', edReasonBinary: 'binary dosya (resim, arşiv…)', edReasonLarge: '4 MB\'tan büyük',
  edReasonEncoding: 'UTF-8 değil — panelden kaydedilirse karakterler bozulur',
  edSaveTitle: 'Değişiklikleri kaydet', edNoChanges: 'Değişiklik yok',
  edBoth: 'Sunucuya ve yerele kaydet', edBothHint: 'Yereldeki dosya da güncellenir; sonraki sync düzenlemeyi ezmez.',
  edBothLocalDiff: 'Yereldeki dosya sunucudakinden farklıydı — yerelde senkronize edilmemiş değişiklikler varsa kaybolur.',
  edRemoteOnly: 'Yalnızca sunucuya kaydet', edRemoteOnlyHint: 'Acil düzeltmeler için. Yereldeki dosya değişmez; yerelde bu dosyayı değiştirip sync edersen sunucudaki düzenleme ezilir.',
  edBackupNote: 'Sunucudaki eski hal yedeklenir, Geçmiş\'ten geri alabilirsin.',
  edDiscardTitle: 'Kaydedilmemiş değişiklikler', edDiscardText: 'Değişiklikler kaydedilmedi. Yine de kapatılsın mı?', edDiscard: 'Kaydetmeden kapat',
  edConflictTitle: 'Sunucudaki dosya değişmiş', edConflictText: 'Sen açtıktan sonra dosya sunucuda değişti (başka biri ya da bir sync). Aşağıda sunucudaki güncel hal → senin sürümün farkı var.',
  edOverwrite: 'Benimkini yine de kaydet', edReload: 'Sunucudakini aç (değişikliklerim gider)',
});
Object.assign(T.en, {
  edEdit: 'Edit', edSave: 'Save', edSaving: 'Saving…', edSaved: 'Saved', edModified: 'modified',
  edLine: 'Ln {0}, Col {1}', edNoHl: 'large file — highlighting off',
  edLocalSame: 'Local file is the same', edLocalDiff: 'Local file differs', edLocalMissing: 'Not in local folder',
  edNotEditable: 'This file can\'t be edited', edReasonBinary: 'binary file (image, archive…)', edReasonLarge: 'larger than 4 MB',
  edReasonEncoding: 'not UTF-8 — saving from the panel would corrupt characters',
  edSaveTitle: 'Save changes', edNoChanges: 'No changes',
  edBoth: 'Save to server and local', edBothHint: 'The local file is updated too, so the next sync won\'t overwrite the edit.',
  edBothLocalDiff: 'The local file differed from the server — unsynced local changes will be lost.',
  edRemoteOnly: 'Save to server only', edRemoteOnlyHint: 'For quick fixes. The local file stays as is; if you change it locally and sync, the edit on the server is overwritten.',
  edBackupNote: 'The old version on the server is backed up; you can undo it from History.',
  edDiscardTitle: 'Unsaved changes', edDiscardText: 'Your changes aren\'t saved. Close anyway?', edDiscard: 'Close without saving',
  edConflictTitle: 'The file changed on the server', edConflictText: 'The file changed on the server after you opened it (someone else or a sync). Below: the current server version → your version.',
  edOverwrite: 'Save mine anyway', edReload: 'Open the server version (discard mine)',
});

const ED_HL_MAX = 150000; // bundan büyükte her tuşta renklendirme yavaşlar → düz metin

async function editFile(P, srv, rel) {
  const title = (dirty) => `${icon('code')}<span class="mono" style="font-size:13px">${esc(rel)}</span>${dirty ? `<span class="ed-nokta" title="${t('edModified')}"></span>` : ''}`;
  const q = `server=${enc(srv)}&path=${enc(rel)}`;
  const lm = await loadModal(title(false), 'genis', () => GET(`/api/p/${enc(P)}/remote/edit?${q}`));
  if (!lm) return;
  let doc = lm.d;
  if (!doc.editable) {
    const why = { binary: t('edReasonBinary'), large: t('edReasonLarge'), encoding: t('edReasonEncoding') }[doc.reason] || doc.reason;
    modal({ title: title(false), body: `<div class="not">${icon('alert')}${t('edNotEditable')}: ${esc(why)}</div>`, actions: [{ label: t('close'), cls: 'birincil' }] });
    return;
  }

  let orig = doc.content;
  const localPill = () => doc.local_exists
    ? (doc.local_same ? `<span class="pill ok"><span class="nokta"></span>${t('edLocalSame')}</span>` : `<span class="pill warn"><span class="nokta"></span>${t('edLocalDiff')}</span>`)
    : `<span class="pill">${t('edLocalMissing')}</span>`;
  const m = modal({
    title: title(false), size: 'genis ed-modal',
    body: `<div class="ed"><div class="ed-no"><pre></pre></div><div class="ed-alan"><pre class="ed-hl"></pre>
        <textarea spellcheck="false" autocomplete="off" autocapitalize="off" wrap="off"></textarea></div></div>
      <div class="ed-durum"><span data-pos></span><span data-info></span><span data-local>${localPill()}</span></div>`,
    actions: [{ left: `<span class="mono">${esc(srv)}</span>` },
      { label: t('close'), cls: 'hayalet', run: async () => (dirty() ? confirmDiscard() : true) },
      { label: icon('check') + t('edSave'), cls: 'birincil', run: () => { save(); return false; } }],
    guard: async () => (dirty() ? confirmDiscard() : true),
  });
  const ta = $('textarea', m.root), hl = $('.ed-hl', m.root), gut = $('.ed-no pre', m.root);
  const titleEl = $('.modal-bas h3', m.root);
  ta.value = orig;
  const dirty = () => ta.value !== orig;
  const confirmDiscard = () => confirmBox(t('edDiscardTitle'), t('edDiscardText'), t('edDiscard'), true);

  // Girinti birimi dosyadan: sekme ile girintili satır varsa \t, yoksa en küçük boşluk girintisi (2/4)
  const unit = (() => {
    if (/^\t/m.test(orig)) return '\t';
    const sp = (orig.match(/^ +(?=\S)/gm) || []).map((s) => s.length).filter((n) => n >= 2);
    return ' '.repeat(sp.length ? Math.min(4, Math.min(...sp)) : 4);
  })();

  let lines = -1, wasDirty = false, raf = 0;
  function paint() {
    raf = 0;
    const v = ta.value;
    const big = v.length > ED_HL_MAX;
    hl.innerHTML = (big ? esc(v) : HL.hlLines(v, rel).join('\n')) + '\n ';
    const n = v.split('\n').length;
    if (n !== lines) { lines = n; gut.textContent = Array.from({ length: n }, (_, i) => i + 1).join('\n') + '\n '; }
    $('[data-info]', m.root).textContent = `${bytes(new Blob([v]).size)}${big ? ' · ' + t('edNoHl') : ''}`;
    const d = dirty();
    if (d !== wasDirty) { wasDirty = d; titleEl.innerHTML = title(d); }
    sync();
  }
  const queue = () => { if (!raf) raf = requestAnimationFrame(paint); };
  function sync() {
    hl.style.transform = `translate(${-ta.scrollLeft}px, ${-ta.scrollTop}px)`;
    gut.style.transform = `translateY(${-ta.scrollTop}px)`;
  }
  function pos() {
    const before = ta.value.slice(0, ta.selectionStart);
    const ln = before.split('\n').length, col = ta.selectionStart - before.lastIndexOf('\n');
    $('[data-pos]', m.root).textContent = t('edLine', ln, col);
  }
  // execCommand('insertText') geri al (Ctrl+Z) geçmişini korur
  const insert = (text) => { if (!document.execCommand('insertText', false, text)) ta.setRangeText(text, ta.selectionStart, ta.selectionEnd, 'end'); };

  ta.addEventListener('input', () => { queue(); pos(); });
  ta.addEventListener('scroll', sync);
  ['keyup', 'click', 'select'].forEach((ev) => ta.addEventListener(ev, pos));
  ta.addEventListener('keydown', (e) => {
    if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 's') { e.preventDefault(); save(); return; }
    if (e.key === 'Tab') {
      e.preventDefault();
      const v = ta.value, s = ta.selectionStart, en = ta.selectionEnd;
      if (s === en && !e.shiftKey) return insert(unit);
      // Seçili satırların hepsine girinti ekle / kaldır
      const ls = v.lastIndexOf('\n', s - 1) + 1;
      const le = en > s && v[en - 1] === '\n' ? en - 1 : (v.indexOf('\n', en) + 1 || v.length + 1) - 1;
      const block = v.slice(ls, le).split('\n');
      const out = block.map((l) => {
        if (!e.shiftKey) return unit + l;
        if (l.startsWith('\t')) return l.slice(1);
        return l.replace(new RegExp(`^ {1,${unit === '\t' ? 4 : unit.length}}`), '');
      }).join('\n');
      ta.setSelectionRange(ls, le);
      insert(out);
      ta.setSelectionRange(ls, ls + out.length);
      return;
    }
    if (e.key === 'Enter' && !e.ctrlKey && !e.altKey) {
      const v = ta.value, s = ta.selectionStart;
      const ind = v.slice(v.lastIndexOf('\n', s - 1) + 1, s).match(/^[ \t]*/)[0];
      if (ind) { e.preventDefault(); insert('\n' + ind); }
    }
  });
  paint(); pos();
  ta.focus(); ta.setSelectionRange(0, 0); ta.scrollTop = 0; sync();

  let saving = false;
  async function save() {
    if (saving) return;
    if (!dirty()) { toast(t('edNoChanges')); return; }
    const content = ta.value;
    const choice = await saveDialog(content);
    if (!choice) return ta.focus();
    await put(content, choice.local, false);
  }

  async function put(content, local, force) {
    saving = true;
    const btn = $$('.modal-alt .btn', m.root).pop();
    const label = btn.innerHTML;
    btn.disabled = true; btn.innerHTML = `${icon('refresh', 'spin')}${t('edSaving')}`;
    let again = false;
    try {
      const res = await api('PUT', `/api/p/${enc(P)}/remote/edit`, { server: srv, path: rel, content, base_hash: doc.hash, local, force });
      orig = content;
      doc.hash = res.hash;
      if (local) doc.local_exists = true;
      doc.local_same = local;
      $('[data-local]', m.root).innerHTML = localPill();
      toast(`${t('edSaved')}: ${rel}`);
      paint();
    } catch (e) {
      if (e.status === 409 && e.body && e.body.conflict) again = (await conflict(e.body.conflict, content)) === 'overwrite';
      else fail(e);
    } finally {
      saving = false; btn.disabled = false; btn.innerHTML = label; ta.focus();
    }
    if (again) await put(content, local, true);
  }

  function saveDialog(content) {
    return new Promise((res) => {
      let done = false;
      const d = diffView(orig, content, rel);
      const warnLocal = doc.local_exists && !doc.local_same;
      const dm = modal({
        title: `${icon('check')}${t('edSaveTitle')} <span class="fark-ozet"><span class="ok-t">+${d.add}</span><span class="bad-t">−${d.del}</span></span>`, size: 'genis',
        body: `${d.html}
          <div class="ed-secim">
            <label><input type="radio" name="ed-hedef" value="both" checked><span><b>${t('edBoth')}</b><small>${t('edBothHint')}</small>
              ${warnLocal ? `<small class="warn-t">${icon('alert')}${t('edBothLocalDiff')}</small>` : ''}</span></label>
            <label><input type="radio" name="ed-hedef" value="remote"><span><b>${t('edRemoteOnly')}</b><small>${t('edRemoteOnlyHint')}</small></span></label>
          </div>`,
        actions: [{ left: `<span class="faint">${t('edBackupNote')}</span>` },
          { label: t('cancel'), cls: 'hayalet', run: () => { done = true; res(null); } },
          { label: icon('check') + t('edSave'), cls: 'birincil', run: (root) => { done = true; res({ local: $('input[name=ed-hedef]:checked', root).value === 'both' }); } }],
        onClose: () => { if (!done) res(null); },
      });
      // Enter = Kaydet (modal ilk input'a, yani seçeneğe odaklanır)
      dm.root.addEventListener('keydown', (e) => {
        if (e.key === 'Enter' && !e.target.matches('button')) { e.preventDefault(); $('.modal-alt .birincil', dm.root).click(); }
      });
    });
  }

  // Çözer: 'overwrite' | 'reload' | undefined (iptal)
  function conflict(cur, mine) {
    return new Promise((res) => {
      const body = cur.editable
        ? `<p class="muted" style="margin-top:0">${t('edConflictText')}</p>${diffView(cur.content, mine, rel).html}`
        : `<p class="muted" style="margin-top:0">${t('edConflictText')}</p>`;
      modal({
        title: `${icon('alert')}${t('edConflictTitle')}`, size: 'genis', body,
        actions: [
          { label: t('cancel'), cls: 'hayalet', run: () => res() },
          ...(cur.editable ? [{ label: icon('refresh') + t('edReload'), cls: '', run: () => {
            doc = { ...doc, hash: cur.hash, size: cur.size }; orig = cur.content; ta.value = cur.content; paint(); pos(); res('reload');
          } }] : []),
          { label: icon('check') + t('edOverwrite'), cls: 'tehlike', run: () => res('overwrite') },
        ],
        onClose: () => res(),
      });
    });
  }
}

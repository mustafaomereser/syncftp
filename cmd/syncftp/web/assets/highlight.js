'use strict';

/* ═══════════════════════════════════════════════════════════════════════════
   Hafif syntax highlighter + satır diff — bağımlılıksız (panel offline çalışır).
   hlLines(code, filename) → satır başına HTML dizisi (çok satırlı yorum/string
   satırlara bölünse de span'ler her satırda kapanıp yeniden açılır).
   diffLines(a, b) → [{op: ' '|'+'|'-', a: satırNo, b: satırNo, text}]
   ═══════════════════════════════════════════════════════════════════════════ */

(function () {
  const esc = (s) => s.replace(/[&<>"]/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;' }[c]));
  const kw = (words) => new RegExp('\\b(?:' + words.split(' ').join('|') + ')\\b', 'y');

  const JS_KW = 'await async break case catch class const continue debugger default delete do else export extends finally for from function if import in instanceof let new of return static super switch this throw try typeof var void while with yield true false null undefined interface type enum implements private public protected readonly as';
  const PHP_KW = 'abstract and array as break callable case catch class clone const continue declare default do echo else elseif empty enddeclare endfor endforeach endif endswitch endwhile enum extends final finally fn for foreach function global goto if implements include include_once instanceof insteadof interface isset list match namespace new or print private protected public readonly require require_once return static switch throw trait try unset use var while xor yield true false null self parent';
  const GO_KW = 'break case chan const continue default defer else fallthrough for func go goto if import interface map package range return select struct switch type var true false nil iota';
  const PY_KW = 'and as assert async await break class continue def del elif else except finally for from global if import in is lambda nonlocal not or pass raise return try while with yield True False None self';
  const SQL_KW = 'select from where insert into values update set delete create table alter drop index primary key foreign references join left right inner outer on group by order having limit offset as and or not null is in like between distinct union all exists case when then else end default unique auto_increment int varchar text datetime timestamp';

  // Kural: [regex(sticky), sınıf]. İlk eşleşen kazanır; hiçbiri eşleşmezse tek karakter düz metin.
  const C_COMMENTS = [[/\/\*[\s\S]*?(?:\*\/|$)/y, 'c'], [/\/\/[^\n]*/y, 'c']];
  const STRINGS = [[/"(?:\\[\s\S]|[^"\\\n])*"?/y, 's'], [/'(?:\\[\s\S]|[^'\\\n])*'?/y, 's']];
  const NUM = [/\b(?:0x[\da-fA-F]+|\d+(?:\.\d+)?(?:e[+-]?\d+)?)\b/y, 'n'];
  const CALL = [/\b[A-Za-z_$][\w$]*(?=\s*\()/y, 'f'];

  const LANGS = {
    js: [...C_COMMENTS, [/`(?:\\[\s\S]|[^`\\])*`?/y, 's'], ...STRINGS, [kw(JS_KW), 'k'], NUM, CALL, [/\b[A-Z][\w$]*/y, 't']],
    php: [...C_COMMENTS, [/#[^\n]*/y, 'c'], [/<\?php|\?>/y, 'k'], ...STRINGS, [/\$[A-Za-z_]\w*/y, 'v'], [kw(PHP_KW), 'k'], NUM, CALL, [/\b[A-Z]\w*/y, 't']],
    go: [...C_COMMENTS, [/`[^`]*`?/y, 's'], ...STRINGS, [kw(GO_KW), 'k'], NUM, CALL, [/\b[A-Z]\w*/y, 't']],
    py: [[/#[^\n]*/y, 'c'], [/"""[\s\S]*?(?:"""|$)/y, 's'], [/'''[\s\S]*?(?:'''|$)/y, 's'], ...STRINGS, [kw(PY_KW), 'k'], NUM, CALL],
    css: [[/\/\*[\s\S]*?(?:\*\/|$)/y, 'c'], [/\/\/[^\n]*/y, 'c'], ...STRINGS, [/@[\w-]+/y, 'k'], [/#[\da-fA-F]{3,8}\b/y, 'n'],
      [/--[\w-]+|[\w-]+(?=\s*:(?!:))/y, 'p'], [/-?\d*\.?\d+(?:px|em|rem|%|vh|vw|s|ms|deg|fr)?\b/y, 'n'], [/\$[\w-]+/y, 'v'], [/[.#][\w-]+/y, 't'], [/!important/y, 'k']],
    html: [[/<!--[\s\S]*?(?:-->|$)/y, 'c'], [/\{\{[\s\S]*?\}\}|\{!![\s\S]*?!!\}|@\w+/y, 'v'], [/<\?php[\s\S]*?(?:\?>|$)/y, 'v'],
      [/<\/?[\w:-]+/y, 'k'], [/\/?>/y, 'k'], [/[\w:@.-]+(?==)/y, 'p'], ...STRINGS],
    json: [[/"(?:\\.|[^"\\])*"(?=\s*:)/y, 'p'], ...STRINGS, [/\b(?:true|false|null)\b/y, 'k'], NUM],
    sql: [[/--[^\n]*/y, 'c'], [/\/\*[\s\S]*?(?:\*\/|$)/y, 'c'], ...STRINGS, [/`[^`]*`/y, 'v'], [new RegExp('\\b(?:' + SQL_KW.split(' ').join('|') + ')\\b', 'iy'), 'k'], NUM],
    yaml: [[/#[^\n]*/y, 'c'], [/^[ \t-]*[\w.-]+(?=\s*:)/my, 'p'], ...STRINGS, [/\b(?:true|false|null|yes|no)\b/y, 'k'], NUM],
    ini: [[/[#;][^\n]*/y, 'c'], [/^\s*\[[^\]\n]*\]/my, 't'], [/^\s*[\w.-]+(?=\s*=)/my, 'p'], ...STRINGS, [/\b(?:true|false|null|on|off)\b/iy, 'k'], NUM],
    sh: [[/#[^\n]*/y, 'c'], ...STRINGS, [/\$\{?[\w]+\}?/y, 'v'], [kw('if then else elif fi for while do done case esac function return export local in echo exit'), 'k'], NUM],
    md: [[/^#{1,6}[^\n]*/my, 'k'], [/```[\s\S]*?(?:```|$)/y, 's'], [/`[^`\n]*`/y, 's'], [/\*\*[^*\n]+\*\*/y, 't'], [/\[[^\]\n]*\]\([^)\n]*\)/y, 'v'], [/^\s*[-*+] /my, 'p']],
  };
  const EXT = {
    js: 'js', mjs: 'js', cjs: 'js', jsx: 'js', ts: 'js', tsx: 'js', php: 'php', phtml: 'php', go: 'go', py: 'py',
    css: 'css', scss: 'css', sass: 'css', less: 'css', html: 'html', htm: 'html', xml: 'html', svg: 'html', vue: 'html', twig: 'html',
    json: 'json', lock: 'json', webmanifest: 'json', sql: 'sql', yml: 'yaml', yaml: 'yaml', ini: 'ini', env: 'ini', conf: 'ini', toml: 'ini',
    htaccess: 'ini', sh: 'sh', bash: 'sh', md: 'md', markdown: 'md',
  };

  function langOf(name) {
    const base = (name || '').split('/').pop().toLowerCase();
    if (base.endsWith('.blade.php')) return 'html';
    if (base === '.htaccess' || base.startsWith('.env')) return 'ini';
    const m = base.match(/\.([a-z0-9]+)$/);
    return m ? EXT[m[1]] || null : null;
  }

  function tokenize(code, rules) {
    const out = [];
    let i = 0, plain = '';
    const flush = () => { if (plain) { out.push(['', plain]); plain = ''; } };
    while (i < code.length) {
      let hit = null;
      for (const [re, cls] of rules) {
        re.lastIndex = i;
        const m = re.exec(code);
        if (m && m[0].length) { hit = [cls, m[0]]; break; }
      }
      if (hit) { flush(); out.push(hit); i += hit[1].length; }
      else { plain += code[i]; i++; }
    }
    flush();
    return out;
  }

  // Token'ları satırlara böl; her satırda span kapanıp yeniden açılır.
  function hlLines(code, name) {
    const lang = langOf(name);
    if (!lang || code.length > 400000) return code.split('\n').map(esc);
    const lines = [''];
    for (const [cls, text] of tokenize(code, LANGS[lang])) {
      const parts = text.split('\n');
      parts.forEach((part, idx) => {
        if (idx > 0) lines.push('');
        if (part) lines[lines.length - 1] += cls ? `<span class="h-${cls}">${esc(part)}</span>` : esc(part);
      });
    }
    return lines;
  }

  // Satır diff — LCS (Myers'ın sade dinamik programlama hali). Büyük dosyalarda
  // ortak baş/son kısım kırpılır; kalan çok büyükse null döner (çağıran tam görünüme düşer).
  function diffLines(aText, bText) {
    const a = aText.split('\n'), b = bText.split('\n');
    let start = 0;
    while (start < a.length && start < b.length && a[start] === b[start]) start++;
    let endA = a.length - 1, endB = b.length - 1;
    while (endA >= start && endB >= start && a[endA] === b[endB]) { endA--; endB--; }
    const A = a.slice(start, endA + 1), B = b.slice(start, endB + 1);
    if (A.length * B.length > 6e6) return null;
    const n = A.length, m = B.length;
    const dp = Array.from({ length: n + 1 }, () => new Uint32Array(m + 1));
    for (let i = n - 1; i >= 0; i--) for (let j = m - 1; j >= 0; j--)
      dp[i][j] = A[i] === B[j] ? dp[i + 1][j + 1] + 1 : Math.max(dp[i + 1][j], dp[i][j + 1]);
    const ops = [];
    for (let k = 0; k < start; k++) ops.push({ op: ' ', a: k + 1, b: k + 1, ai: k, bi: k });
    let i = 0, j = 0;
    while (i < n || j < m) {
      if (i < n && j < m && A[i] === B[j]) { ops.push({ op: ' ', a: start + i + 1, b: start + j + 1, ai: start + i, bi: start + j }); i++; j++; }
      // Eşitlikte önce silme: alışılmış diff düzeni (− satırları + satırlarından önce)
      else if (i < n && (j >= m || dp[i + 1][j] >= dp[i][j + 1])) { ops.push({ op: '-', a: start + i + 1, ai: start + i }); i++; }
      else { ops.push({ op: '+', b: start + j + 1, bi: start + j }); j++; }
    }
    for (let k = endA + 1, q = endB + 1; k < a.length; k++, q++) ops.push({ op: ' ', a: k + 1, b: q + 1, ai: k, bi: q });
    return ops;
  }

  // Değişikliklerin etrafında ctx satır bağlam bırakıp aradaki ortak blokları katlar.
  function hunks(ops, ctx = 3) {
    const keep = new Uint8Array(ops.length);
    ops.forEach((o, i) => { if (o.op !== ' ') for (let k = Math.max(0, i - ctx); k <= Math.min(ops.length - 1, i + ctx); k++) keep[k] = 1; });
    const out = [];
    let skipped = 0;
    ops.forEach((o, i) => {
      if (keep[i]) { if (skipped) { out.push({ fold: skipped }); skipped = 0; } out.push(o); }
      else skipped++;
    });
    if (skipped) out.push({ fold: skipped });
    return out;
  }

  window.HL = { hlLines, diffLines, hunks, langOf };
})();

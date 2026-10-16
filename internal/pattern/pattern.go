// Package pattern — include / exclude / protect kalıplarının TEK eşleştiricisi. Sync, status,
// calibrate, verify ve protect kontrolü bunu kullanır (eskiden üç ayrı kural vardı: sync'in
// filtresi "node_modules/" ve "*.log" gibi kalıpları hiç tanımıyordu).
//
// Kurallar (yollar ve kalıplar "/" ayraçlı, göreceli):
//   - baştaki "/" yok sayılır; sondaki "/" klasör demektir ("vendor/" → vendor ve altındaki her şey)
//   - joker yoksa: tam yol ya da klasör öneki ("app" → app, app/x.php)
//   - joker varsa (* ? [ ): yolun tamamıyla eşleşir; kalıpta "/" yoksa her derinlikteki dosya adıyla da
//     ("*.log" → a.log, logs/b.log) — .gitignore alışkanlığıyla aynı
package pattern

import (
	"path"
	"strings"
)

func clean(p string) (pat string, dirOnly bool) {
	p = strings.ReplaceAll(strings.TrimSpace(p), "\\", "/")
	p = strings.TrimPrefix(p, "/")
	dirOnly = strings.HasSuffix(p, "/")
	return strings.TrimSuffix(p, "/"), dirOnly
}

func hasGlob(p string) bool { return strings.ContainsAny(p, "*?[") }

// Match rel yolu kalıba uyuyor mu.
func Match(rel, pat string) bool {
	p, _ := clean(pat)
	if p == "" {
		return false
	}
	rel = strings.TrimPrefix(rel, "/")
	if !hasGlob(p) {
		return rel == p || strings.HasPrefix(rel, p+"/")
	}
	if ok, _ := path.Match(p, rel); ok {
		return true
	}
	// Klasör kalıbı: "build*/" → build-x/a.js; ya da yolun herhangi bir üst klasörü eşleşiyorsa
	for d := path.Dir(rel); d != "." && d != "/"; d = path.Dir(d) {
		if ok, _ := path.Match(p, d); ok {
			return true
		}
	}
	if !strings.Contains(p, "/") {
		if ok, _ := path.Match(p, path.Base(rel)); ok {
			return true
		}
		for _, seg := range strings.Split(path.Dir(rel), "/") {
			if ok, _ := path.Match(p, seg); ok {
				return true
			}
		}
	}
	return false
}

// Any yol kalıplardan birine uyuyor mu.
func Any(rel string, pats []string) bool {
	for _, p := range pats {
		if Match(rel, p) {
			return true
		}
	}
	return false
}

// MayContain dir klasörünün altında kalıba uyan bir şey OLABİLİR mi (uzak taramada include dışı
// klasörlere hiç girmemek için). Jokerli kalıpta bilinemez → true.
func MayContain(dir, pat string) bool {
	p, _ := clean(pat)
	if p == "" || hasGlob(p) {
		return true
	}
	dir = strings.TrimPrefix(dir, "/")
	return dir == p || strings.HasPrefix(p, dir+"/") || strings.HasPrefix(dir, p+"/")
}

// Exact kalıp diskte doğrudan kontrol edilebilir bir tam yol mu (joker ve klasör işareti yok).
func Exact(pat string) bool {
	p := strings.TrimSpace(pat)
	return !hasGlob(p) && !strings.ContainsAny(p, "{") && !strings.HasSuffix(p, "/")
}

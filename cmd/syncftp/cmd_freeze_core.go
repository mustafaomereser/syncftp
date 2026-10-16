package main

import (
	"path"
	"sort"
	"strings"

	ftpclient "syncftp/internal/ftp"
)

// Freeze aç/kapa — CLI tarayıcısı (f tuşu) ve panelin dosya tarayıcısı bunu kullanır.
// Freeze listesi DOSYA yolu tutar (sync dosya yolu bekler): klasör hedefleri sunucudaki dosyalarına
// açılır (ListRecursive). Hedeflerin hepsi zaten frozen ise kaldırılır — klasörde kendisi ve önekiyle
// başlayan tüm yollar (artık olmayan dosyaların kayıtları da) — değilse eklenir. Klasör, altında tek
// bir frozen dosya varsa "frozen" sayılır (tarayıcıdaki ❄📁 ikonuyla aynı) → f onu çözer.

type freezeTarget struct {
	Rel string `json:"path"` // RemotePath'e göreceli
	Dir bool   `json:"dir"`
}

// freezeIsFrozen dosya → listede mi; klasör → kendisi ya da altında frozen dosya var mı.
func freezeIsFrozen(set map[string]bool, t freezeTarget) bool {
	if set[t.Rel] {
		return true
	}
	if !t.Dir {
		return false
	}
	prefix := t.Rel + "/"
	for p, v := range set {
		if v && strings.HasPrefix(p, prefix) {
			return true
		}
	}
	return false
}

// freezeToggle set'i yerinde değiştirir. listFolder klasörün altındaki dosyaları (klasöre göreceli)
// döner; yalnızca freeze eklenirken çağrılır. Dönen n: eklenen/kaldırılan dosya sayısı.
func freezeToggle(set map[string]bool, targets []freezeTarget, listFolder func(rel string) ([]string, error)) (frozen bool, n int, err error) {
	all := len(targets) > 0
	for _, t := range targets {
		if !freezeIsFrozen(set, t) {
			all = false
			break
		}
	}
	if all { // kaldır
		for _, t := range targets {
			if set[t.Rel] {
				delete(set, t.Rel)
				n++
			}
			if t.Dir {
				prefix := t.Rel + "/"
				for p := range set {
					if strings.HasPrefix(p, prefix) {
						delete(set, p)
						n++
					}
				}
			}
		}
		return false, n, nil
	}
	for _, t := range targets {
		if !t.Dir {
			if !set[t.Rel] {
				set[t.Rel] = true
				n++
			}
			continue
		}
		files, lerr := listFolder(t.Rel)
		if lerr != nil {
			return true, n, lerr
		}
		for _, f := range files {
			full := f
			if t.Rel != "" {
				full = t.Rel + "/" + f
			}
			if !set[full] {
				set[full] = true
				n++
			}
		}
	}
	return true, n, nil
}

// remoteFolderLister sunucudaki klasörün dosyalarını (klasöre göreceli) listeler.
func remoteFolderLister(c *ftpclient.Client, root string) func(rel string) ([]string, error) {
	return func(rel string) ([]string, error) {
		files, err := c.ListRecursive(path.Join(root, rel))
		if err != nil {
			return nil, err
		}
		out := make([]string, 0, len(files))
		for f := range files {
			out = append(out, f)
		}
		return out, nil
	}
}

func frozenSetList(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for p, v := range set {
		if v {
			out = append(out, p)
		}
	}
	sort.Strings(out)
	return out
}

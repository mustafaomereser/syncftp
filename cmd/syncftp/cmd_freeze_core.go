package main

import (
	"fmt"
	"sort"
	"strings"

	"syncftp/internal/frozen"
	"syncftp/internal/lang"
)

// Freeze aç/kapa — CLI tarayıcısı (f tuşu) ve panelin dosya tarayıcısı bunu kullanır.
// Klasör freeze'i tek kayıttır: "klasör/" → altındaki her dosya (sonradan eklenenler dahil) freeze'li
// sayılır (frozen.MatchSet). Eskiden klasör sunucudaki tüm dosyalarına açılıyordu (ListRecursive);
// binlerce dosyalı klasörde freeze imkânsız hale geliyordu. Çözerken klasörün kendisi ve altındaki
// tek tek dosya kayıtları (eski usul) birlikte silinir. Klasör, altında tek bir frozen dosya varsa da
// "frozen" görünür (tarayıcıdaki ❄📁) → f onu çözer.

type freezeTarget struct {
	Rel string `json:"path"` // RemotePath'e göreceli
	Dir bool   `json:"dir"`
}

func (t freezeTarget) key() string {
	if t.Dir {
		return strings.TrimSuffix(t.Rel, "/") + "/"
	}
	return t.Rel
}

// freezeIsFrozen hedef (kendisi ya da üst klasörü) freeze'li mi; klasörde altında freeze'li dosya da sayılır.
func freezeIsFrozen(set map[string]bool, t freezeTarget) bool {
	if frozen.MatchSet(set, t.key()) {
		return true
	}
	return t.Dir && hasFrozenUnder(set, t.key())
}

func hasFrozenUnder(set map[string]bool, prefix string) bool {
	for p, v := range set {
		if v && strings.HasPrefix(p, prefix) {
			return true
		}
	}
	return false
}

// frozenByParent hedefi freeze'leyen üst klasör ("a/b/"), yoksa "".
func frozenByParent(set map[string]bool, rel string) string {
	for i := 0; i < len(rel)-1; i++ {
		if rel[i] == '/' && set[rel[:i+1]] && rel[:i+1] != rel {
			return rel[:i+1]
		}
	}
	return ""
}

// freezeToggle set'i yerinde değiştirir. Hepsi freeze'liyse kaldırır, değilse ekler.
// Dönen n: eklenen/kaldırılan kayıt sayısı. Üst klasörü freeze'li bir öğe tek başına çözülemez.
func freezeToggle(set map[string]bool, targets []freezeTarget) (frozenNow bool, n int, err error) {
	all := len(targets) > 0
	for _, t := range targets {
		if !freezeIsFrozen(set, t) {
			all = false
			break
		}
	}
	if all { // kaldır
		for _, t := range targets {
			if parent := frozenByParent(set, t.key()); parent != "" {
				return true, n, fmt.Errorf(lang.L.FreezeByParentFmt, t.Rel, parent)
			}
			k := t.key()
			if set[k] {
				delete(set, k)
				n++
			}
			if t.Dir {
				for p := range set {
					if strings.HasPrefix(p, k) {
						delete(set, p)
						n++
					}
				}
			}
		}
		return false, n, nil
	}
	for _, t := range targets {
		k := t.key()
		if frozen.MatchSet(set, k) {
			continue // zaten (üst klasörden) freeze'li
		}
		if t.Dir { // altındaki tek tek kayıtlar artık gereksiz
			for p := range set {
				if strings.HasPrefix(p, k) {
					delete(set, p)
				}
			}
		}
		set[k] = true
		n++
	}
	return true, n, nil
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

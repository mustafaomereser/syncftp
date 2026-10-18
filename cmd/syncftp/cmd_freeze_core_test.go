package main

import (
	"strings"
	"testing"
)

func TestFreezeToggleFolderAndUnfreezeByPrefix(t *testing.T) {
	// eski usul tek tek dosya kayıtları klasör freeze'inde temizlenir, tek "app/" kalır
	set := map[string]bool{"app/a.php": true}
	isFrozen, n, err := freezeToggle(set, []freezeTarget{{Rel: "app", Dir: true}, {Rel: "x.txt"}})
	if err == nil && isFrozen { // app zaten "frozen" görünüyordu ama x.txt değildi → hepsi eklenir
		if strings.Join(frozenSetList(set), ",") != "app/,x.txt" || n != 2 {
			t.Fatalf("freeze: n=%d set=%v", n, set)
		}
	} else {
		t.Fatalf("freeze: frozen=%v err=%v", isFrozen, err)
	}
	// klasör altındaki dosya üst klasörden freeze'li, tek başına çözülemez
	if !freezeIsFrozen(set, freezeTarget{Rel: "app/sub/b.php"}) {
		t.Fatal("alt dosya freeze'li görünmeli")
	}
	if _, _, err := freezeToggle(set, []freezeTarget{{Rel: "app/sub/b.php"}}); err == nil {
		t.Fatal("üst klasörü freeze'li dosya çözülememeli")
	}
	// klasörü çöz → "app/" ve altındaki eski kayıtlar gider
	set["app/old.php"] = true
	isFrozen, n, _ = freezeToggle(set, []freezeTarget{{Rel: "app", Dir: true}})
	if isFrozen || n != 2 || strings.Join(frozenSetList(set), ",") != "x.txt" {
		t.Fatalf("unfreeze: frozen=%v n=%d set=%v", isFrozen, n, set)
	}
}

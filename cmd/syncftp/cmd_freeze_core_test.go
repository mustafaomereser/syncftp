package main

import (
	"strings"
	"testing"
)

func TestFreezeToggleFolderAndUnfreezeByPrefix(t *testing.T) {
	set := map[string]bool{}
	list := func(rel string) ([]string, error) { return []string{"a.php", "sub/b.php"}, nil }
	// frozen değil → sunucudaki dosyalar eklenir
	frozen, n, err := freezeToggle(set, []freezeTarget{{Rel: "app", Dir: true}}, list)
	if err != nil || !frozen || n != 2 || !set["app/a.php"] || !set["app/sub/b.php"] {
		t.Fatalf("freeze: frozen=%v n=%d set=%v err=%v", frozen, n, set, err)
	}
	// altında frozen var (❄📁) → çöz; yerelde/sunucuda artık olmayan eski kayıt da önekle gider
	set["app/old-deleted.php"] = true
	frozen, n, _ = freezeToggle(set, []freezeTarget{{Rel: "app", Dir: true}}, list)
	if frozen || n != 3 || len(set) != 0 {
		t.Fatalf("unfreeze: frozen=%v n=%d set=%v", frozen, n, set)
	}
	// dosya
	freezeToggle(set, []freezeTarget{{Rel: "x.txt"}}, nil)
	if strings.Join(frozenSetList(set), ",") != "x.txt" {
		t.Fatalf("file: %v", set)
	}
}

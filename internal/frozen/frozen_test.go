package frozen

import "testing"

func TestMatchSet(t *testing.T) {
	set := map[string]bool{"node_modules/": true, "src/a.php": true, "public/uploads/": true}
	cases := map[string]bool{
		"node_modules/x/y.js":   true,
		"node_modules/":         true,
		"node_modules":          false, // klasörün kendisi "x/" ile sorulur
		"src/a.php":             true,
		"src/b.php":             false,
		"public/uploads/1.jpg":  true,
		"public/index.php":      false,
		"node_modules2/x.js":    false,
		"public/uploads2/x.jpg": false,
	}
	for p, want := range cases {
		if got := MatchSet(set, p); got != want {
			t.Errorf("%s: %v, want %v", p, got, want)
		}
	}
}

package pattern

import "testing"

func TestMatch(t *testing.T) {
	cases := []struct {
		rel, pat string
		want     bool
	}{
		{"node_modules/a/b.js", "node_modules/", true},
		{"node_modules", "node_modules/", true},
		{"src/node_modules/x.js", "node_modules/", false}, // jokersiz kalıp köke göreceli
		{"app/x.php", "app", true},
		{"application/x.php", "app", false},
		{"a.log", "*.log", true},
		{"storage/logs/b.log", "*.log", true},
		{"storage/logs/b.txt", "storage/logs/", true},
		{"storage/logs/b.txt", "/storage/logs/", true},
		{".env", ".env", true},
		{"config/.env", ".env", false},
		{"build-1/a.js", "build*/", true},
		{"src/a.min.js", "src/*.min.js", true},
		{"src/sub/a.min.js", "src/*.min.js", false},
		{"x/.DS_Store", ".DS_Store", false},
		{"x/.DS_Store", "*.DS_Store", true},
		{"cache/tmp/x", "tmp", false},
		{"a/tmp/x", "tm?", true},
	}
	for _, c := range cases {
		if got := Match(c.rel, c.pat); got != c.want {
			t.Errorf("Match(%q, %q) = %v, want %v", c.rel, c.pat, got, c.want)
		}
	}
}

func TestMayContain(t *testing.T) {
	if !MayContain("public", "public/css/") || !MayContain("public/css", "public/") || MayContain("vendor", "public/") || !MayContain("x", "*.php") {
		t.Fatal("MayContain")
	}
}

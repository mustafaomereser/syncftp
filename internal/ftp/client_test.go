package ftp

import "testing"

func TestIsProtected_ExactMatch(t *testing.T) {
	protect := []string{".env", "config/database.php"}
	if !IsProtected(".env", protect) {
		t.Error(".env korunmalıydı")
	}
	if !IsProtected("config/database.php", protect) {
		t.Error("config/database.php korunmalıydı")
	}
}

func TestIsProtected_DirectoryPrefix(t *testing.T) {
	protect := []string{"storage/"}
	if !IsProtected("storage/logs/app.log", protect) {
		t.Error("storage/ dizini altındaki dosya korunmalıydı")
	}
	if IsProtected("storagebackup/file.txt", protect) {
		t.Error("sadece prefix uyan dizin korunmamalıydı")
	}
}

func TestIsProtected_LeadingSlash(t *testing.T) {
	protect := []string{"/.env"}
	if !IsProtected(".env", protect) {
		t.Error("baştaki / kaldırıldıktan sonra eşleşmeli")
	}
}

func TestIsProtected_NotProtected(t *testing.T) {
	protect := []string{".env", "config/"}
	if IsProtected("index.php", protect) {
		t.Error("index.php korunmamalıydı")
	}
	if IsProtected("public/app.js", protect) {
		t.Error("public/app.js korunmamalıydı")
	}
}

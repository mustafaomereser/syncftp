package main

import (
	"os"
	"sort"
	"strings"
)

// localDirEntries yerel klasörün girdileri: .syncftp ve .git atılır (sync onları hiç taramaz),
// önce klasörler, sonra büyük/küçük harf duyarsız ad. CLI config ekranının yerel tarayıcısı ve
// panelin yerel ağaç seçicisi ortak (uzaktaki karşılığı remoteEntries).
func localDirEntries(dir string) ([]os.DirEntry, error) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	out := ents[:0]
	for _, e := range ents {
		if n := e.Name(); n == ".syncftp" || n == ".git" {
			continue
		}
		out = append(out, e)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].IsDir() != out[j].IsDir() {
			return out[i].IsDir()
		}
		return strings.ToLower(out[i].Name()) < strings.ToLower(out[j].Name())
	})
	return out, nil
}

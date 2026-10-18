package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"syncftp/internal/config"
	"syncftp/internal/lang"
	"syncftp/internal/projects"
)

// Projeyi syncFTP'den çıkarma — CLI `project delete` ve panel Proje ayarları → Tehlikeli bölge ortak.
// .syncftp klasörü (config, state, release yedekleri, loglar, freeze…) çöp kutusuna taşınır;
// çöp kutusu kullanılamazsa .syncftp.removed-<zaman> olarak yeniden adlandırılır. İkisinde de
// geri alınabilir ve klasör taraması projeyi artık görmez. Proje dosyalarına ve sunuculara
// dokunulmaz. Anahtarlıktaki şifreler SİLİNMEZ: kütüphane ve diğer projeler aynı kaydı
// (protokol://user@host:port) paylaşabilir.

type projectRemoval struct {
	Trashed bool   `json:"trashed"`
	Renamed string `json:"renamed,omitempty"` // çöp kutusu yoksa yeni ad
}

func removeProject(dir string) (projectRemoval, error) {
	var res projectRemoval
	syncDir := filepath.Join(dir, ".syncftp")
	if !projects.HasConfig(dir) {
		return res, fmt.Errorf(lang.L.ProjectNoConfigFmt, dir)
	}
	// Başka bir süreç (CLI watch, başka panel) bu projeyi izliyorsa klasör altından çekilmesin
	if cfg, err := config.Load(dir); err == nil {
		for _, srv := range cfg.Servers {
			if pid := watchLockOwner(dir, srv.Name); pid != 0 {
				return res, fmt.Errorf("%s: "+lang.L.WatchLockedFmt, srv.Name, pid)
			}
		}
	}
	if err := moveToTrash(syncDir); err == nil {
		res.Trashed = true
	} else {
		dst := filepath.Join(dir, ".syncftp.removed-"+time.Now().Format("20060102-150405"))
		if rerr := os.Rename(syncDir, dst); rerr != nil {
			return res, rerr
		}
		res.Renamed = dst
	}
	projects.Forget(dir)
	return res, nil
}

// projectRemovalMsg sonucu standart biçimde döner (CLI ve panel bildirimi).
func projectRemovalMsg(name string, r projectRemoval) string {
	if r.Trashed {
		return fmt.Sprintf(lang.L.ProjectRemovedTrashFmt, name)
	}
	return fmt.Sprintf(lang.L.ProjectRemovedRenamedFmt, name, r.Renamed)
}

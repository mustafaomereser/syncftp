package main

import (
	"path/filepath"
	"sort"
	"strings"
	"time"

	"syncftp/internal/config"
	"syncftp/internal/frozen"
	ftpclient "syncftp/internal/ftp"
	"syncftp/internal/milestone"
	"syncftp/internal/state"
)

// Sync filtre zincirinin TEK kaynağı. Sync (syncToServerOpts), CLI `status`, shell `status`,
// serve /api/status ve panelin genel bakış/durum ekranları bunu kullanır → "bekleyen değişiklik"
// sayısı her yerde sync'in gerçekten yükleyeceği listeyle aynıdır.
//
// Sıra: include → exclude → milestone (mtime) → frozen → block → protect

type syncFilters struct {
	include, exclude, protect, block []string
}

// effectiveFilters: include önceliği CLI > sunucu > global; exclude/protect/block toplanır.
func effectiveFilters(cfg *config.Config, srv config.Server, cliInclude, cliExclude []string) syncFilters {
	include := cfg.Sync.Include
	if len(srv.Include) > 0 {
		include = srv.Include
	}
	if len(cliInclude) > 0 {
		include = cliInclude
	}
	exclude := append(append(append([]string{}, cfg.Sync.Exclude...), srv.Exclude...), cliExclude...)
	protect := append(append([]string{}, cfg.Sync.Protect...), srv.Protect...)
	lp := srv.EffectiveLocalPath(cfg.Project)
	return syncFilters{
		include: normalizePatterns(include, lp),
		exclude: normalizePatterns(exclude, lp),
		protect: normalizePatterns(protect, lp),
		block:   append(append([]string{}, cfg.Sync.BlockExtensions...), srv.BlockExtensions...),
	}
}

// pick include + exclude uygular.
func (f syncFilters) pick(paths []string) []string {
	return filterByExclude(filterByInclude(paths, f.include), f.exclude)
}

type uploadPlan struct {
	upload    []string   // yüklenecekler (sıralı)
	frozen    []string   // freeze'li olduğu için atlananlar
	blocked   []string   // engelli uzantı
	protected []string   // protect listesi
	msDate    *time.Time // milestone filtresi uygulandıysa tarihi
	msDropped int        // milestone öncesi değiştiği için düşenler
}

// planUploads include/exclude'dan geçmiş adaylara kalan filtreleri uygular.
// withMilestone=false: milestone sync'i (since modu) ve retry-failed kalıcı filtreden muaftır.
func planUploads(configDir string, cfg *config.Config, srv config.Server, cands []string, f syncFilters, withMilestone bool) uploadPlan {
	var p uploadPlan
	if withMilestone {
		if ms, _ := milestone.Load(configDir, srv.Name); ms != nil {
			d := ms.Date
			p.msDate = &d
			cands, p.msDropped = milestoneFilterMtime(cands, filepath.Join(configDir, srv.EffectiveLocalPath(cfg.Project)), d)
		}
	}
	fl, _ := frozen.Load(configDir, srv.Name)
	blocks := make(map[string]bool, len(f.block))
	for _, e := range f.block {
		e = strings.ToLower(e)
		if !strings.HasPrefix(e, ".") {
			e = "." + e
		}
		blocks[e] = true
	}
	for _, rel := range cands {
		switch {
		case fl != nil && frozen.IsFrozen(fl, rel):
			p.frozen = append(p.frozen, rel)
		case blocks[strings.ToLower(filepath.Ext(rel))]:
			p.blocked = append(p.blocked, rel)
		case ftpclient.IsProtected(rel, f.protect):
			p.protected = append(p.protected, rel)
		default:
			p.upload = append(p.upload, rel)
		}
	}
	sort.Strings(p.upload)
	sort.Strings(p.frozen)
	sort.Strings(p.blocked)
	sort.Strings(p.protected)
	return p
}

// pendingInfo durum ekranlarının ortak sonucu: yeni/değişen = plan.upload'ın parçaları.
type pendingInfo struct {
	newL, changed, deleted []string
	plan                   uploadPlan
}

// pendingChanges state'e göre bekleyenler (normal sync akışı). Silinenler yalnızca raporlanır
// (sync uzaktan silmez) ama include/exclude dışındakiler gösterilmez.
func pendingChanges(configDir string, cfg *config.Config, srv config.Server, current map[string]string, st *state.State, cliInclude, cliExclude []string) pendingInfo {
	f := effectiveFilters(cfg, srv, cliInclude, cliExclude)
	diff := state.Diff(st, current)
	newSet := make(map[string]bool, len(diff.New))
	for _, p := range diff.New {
		newSet[p] = true
	}
	plan := planUploads(configDir, cfg, srv, f.pick(append(append([]string{}, diff.New...), diff.Changed...)), f, true)
	info := pendingInfo{plan: plan, deleted: f.pick(diff.Deleted)}
	for _, p := range plan.upload {
		if newSet[p] {
			info.newL = append(info.newL, p)
		} else {
			info.changed = append(info.changed, p)
		}
	}
	sort.Strings(info.deleted)
	return info
}

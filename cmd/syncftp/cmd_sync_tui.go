package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	ftpclient "syncftp/internal/ftp"
	"syncftp/internal/lang"
	"syncftp/internal/synclog"
)

// ── stiller ───────────────────────────────────────────────────────────────────

var (
	stSync       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5"))
	stOK         = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	stFail       = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	stDim        = lipgloss.NewStyle().Faint(true)
	stBold       = lipgloss.NewStyle().Bold(true)
	stBar        = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	stBarBg      = lipgloss.NewStyle().Faint(true)
	stCurrent    = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	stServerName = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5"))
	stDryRun     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("3"))
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// ── mesajlar ──────────────────────────────────────────────────────────────────

type syncResultMsg ftpclient.UploadResult
type syncAllDoneMsg struct{}
type syncTickMsg struct{}

// ── model ─────────────────────────────────────────────────────────────────────

type syncTUIModel struct {
	serverName string
	configDir  string // log kaydı için; boşsa s tuşu devre dışı
	total      int
	dryRun     bool
	dryFiles   []string

	done      int
	failed    int
	bytesDone int64
	startTime time.Time
	results   []ftpclient.UploadResult
	current   string
	finished  bool
	spinner   int
	width     int
	height    int // terminal yüksekliği
	statusMsg string

	scroll int // log kaydırma offseti; 999999 = en alta git

	resultCh <-chan ftpclient.UploadResult
}

func newSyncTUI(configDir, serverName string, total int, dryRun bool, dryFiles []string, ch <-chan ftpclient.UploadResult) syncTUIModel {
	return syncTUIModel{
		serverName: serverName,
		configDir:  configDir,
		total:      total,
		dryRun:     dryRun,
		dryFiles:   dryFiles,
		startTime:  time.Now(),
		resultCh:   ch,
	}
}

func (m syncTUIModel) Init() tea.Cmd {
	if m.dryRun || m.total == 0 {
		return nil
	}
	return tea.Batch(m.waitResult(), tickCmd())
}

func (m syncTUIModel) waitResult() tea.Cmd {
	return func() tea.Msg {
		r, ok := <-m.resultCh
		if !ok {
			return syncAllDoneMsg{}
		}
		return syncResultMsg(r)
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
		return syncTickMsg{}
	})
}

// buildLogLines tüm yükleme sonuçlarını stillendirilmiş metin satırları olarak döner.
// Sync bitmiş ve hata varsa, sonuna başarısız dosyalar özeti eklenir.
func (m syncTUIModel) buildLogLines() []string {
	var lines []string
	for _, r := range m.results {
		if r.Err != nil {
			suffix := ""
			if r.Attempts > 1 {
				suffix = stDim.Render(fmt.Sprintf(lang.L.SyncRetryFmt, r.Attempts))
			}
			lines = append(lines, fmt.Sprintf("  %s %s%s", stFail.Render("✗"), r.RelPath, suffix))
			for i, e := range r.RetryErrs {
				lines = append(lines, fmt.Sprintf("       %s",
					stDim.Render(fmt.Sprintf(lang.L.SyncAttemptDetailFmt, i+1, e))))
			}
			lines = append(lines, fmt.Sprintf("       %s",
				stFail.Render(fmt.Sprintf(lang.L.SyncAttemptDetailFmt, r.Attempts, r.Err))))
		} else {
			suffix := ""
			if r.Attempts > 1 {
				firstErr := ""
				if len(r.RetryErrs) > 0 {
					firstErr = stDim.Render(" ← " + r.RetryErrs[0].Error())
				}
				suffix = stDim.Render(fmt.Sprintf(lang.L.SyncRetryOkFmt, r.Attempts)) + firstErr
			}
			lines = append(lines, fmt.Sprintf("  %s %s%s", stOK.Render("✓"), r.RelPath, suffix))
		}
	}
	// Sync bitti ve hata var → sonuna özet bölümü ekle (G ile hızla ulaşılabilir)
	if m.finished && m.failed > 0 {
		lines = append(lines, "")
		lines = append(lines, stFail.Render("  ── "+lang.L.SyncFailedHeader+" ──"))
		for _, r := range m.results {
			if r.Err != nil {
				lines = append(lines, fmt.Sprintf("  %s  %s", stFail.Render("✗"), r.RelPath))
			}
		}
	}
	return lines
}

// visibleLogRows terminalin yüksekliğine göre kaç log satırı gösterilebileceğini hesaplar.
func (m syncTUIModel) visibleLogRows() int {
	h := m.height
	if h < 20 {
		h = 24
	}
	// overhead: blank(1)+serverName(1)+blank(1)+bar(1)+blank(1)+spinner/blank(1)+blank(1)+summary(1)+blank(1)+hint(1) ≈ 10
	rows := h - 10
	if rows < 5 {
		rows = 5
	}
	return rows
}

func (m syncTUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case syncTickMsg:
		m.spinner = (m.spinner + 1) % len(spinnerFrames)
		if !m.finished {
			return m, tickCmd()
		}

	case syncResultMsg:
		r := ftpclient.UploadResult(msg)
		m.results = append(m.results, r)
		if r.Err != nil {
			m.failed++
		} else {
			m.done++
			m.bytesDone += r.Size
			m.current = r.RelPath
		}
		return m, m.waitResult()

	case syncAllDoneMsg:
		m.finished = true
		m.scroll = 999999 // en alta otomatik kaydır; View'da maxScroll'a kısıtlanır

	case tea.KeyMsg:
		if m.finished || m.dryRun || m.total == 0 {
			// Gerçek upload bittiyse kaydırma tuşlarını yönet
			if m.finished && m.total > 0 && !m.dryRun {
				lines := m.buildLogLines()
				visH := m.visibleLogRows()
				maxScroll := len(lines) - visH
				if maxScroll < 0 {
					maxScroll = 0
				}
				sc := m.scroll
				if sc > maxScroll {
					sc = maxScroll
				}
				if sc < 0 {
					sc = 0
				}
				switch msg.String() {
				case "up", "k":
					if sc > 0 {
						m.scroll = sc - 1
					}
					return m, nil
				case "down", "j":
					if sc < maxScroll {
						m.scroll = sc + 1
					}
					return m, nil
				case "g":
					m.scroll = 0
					return m, nil
				case "G":
					m.scroll = maxScroll
					return m, nil
				case "pgup":
					m.scroll = sc - visH
					if m.scroll < 0 {
						m.scroll = 0
					}
					return m, nil
				case "pgdown":
					m.scroll = sc + visH
					if m.scroll > maxScroll {
						m.scroll = maxScroll
					}
					return m, nil
				case "s":
					if m.configDir != "" {
						if p, err := synclog.Save(m.configDir, m.serverName, m.plainLog()); err == nil {
							m.statusMsg = fmt.Sprintf(lang.L.SyncLogSavedFmt, p)
						} else {
							m.statusMsg = fmt.Sprintf("  ! %v\n", err)
						}
						return m, nil
					}
				}
			}
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m syncTUIModel) View() string {
	w := m.width
	if w < 40 {
		w = 80
	}
	barW := w - 20

	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(stServerName.Render("  ══ "+m.serverName+" ══") + "\n\n")

	if m.dryRun {
		b.WriteString(stDryRun.Render(lang.L.SyncDryRunTitle) + "\n\n")
		for _, f := range m.dryFiles {
			b.WriteString(stDim.Render("    → ") + f + "\n")
		}
		b.WriteString("\n")
		b.WriteString(stDim.Render(fmt.Sprintf(lang.L.SyncDryRunFmt, len(m.dryFiles))) + "\n")
		return b.String()
	}

	if m.total == 0 {
		b.WriteString(stOK.Render(lang.L.SyncNoChange) + "\n")
		b.WriteString("\n" + stDim.Render(lang.L.SyncAnyKey) + "\n")
		return b.String()
	}

	// İlerleme çubuğu
	total := m.total
	progress := m.done + m.failed
	pct := 0.0
	if total > 0 {
		pct = float64(progress) / float64(total)
	}
	filled := int(pct * float64(barW))
	if filled > barW {
		filled = barW
	}
	bar := stBar.Render(strings.Repeat("█", filled)) +
		stBarBg.Render(strings.Repeat("░", barW-filled))

	pctStr := fmt.Sprintf("%3.0f%%", pct*100)
	b.WriteString(fmt.Sprintf("  %s %s %s\n", bar, stBold.Render(pctStr),
		stDim.Render(fmt.Sprintf("%d/%d", progress, total))))

	// Spinner + mevcut dosya (sadece çalışırken)
	if !m.finished {
		spin := spinnerFrames[m.spinner]
		cur := m.current
		if cur == "" {
			cur = lang.L.SyncConnecting
		}
		maxLen := w - 10
		if len(cur) > maxLen {
			cur = "..." + cur[len(cur)-maxLen+3:]
		}
		b.WriteString(fmt.Sprintf("\n  %s %s\n", stSync.Render(spin), stCurrent.Render(cur)))
	} else {
		b.WriteString("\n") // spinner satırı yerine boşluk
	}

	// Kaydırılabilir log
	b.WriteString("\n")
	lines := m.buildLogLines()
	visH := m.visibleLogRows()

	maxScroll := len(lines) - visH
	if maxScroll < 0 {
		maxScroll = 0
	}
	sc := m.scroll
	if sc > maxScroll {
		sc = maxScroll
	}
	if sc < 0 {
		sc = 0
	}

	if !m.finished {
		// Sync sırasında: en yeni sonuçları göster (alta takip et)
		end := len(lines)
		start := end - visH
		if start < 0 {
			start = 0
		}
		for _, l := range lines[start:end] {
			b.WriteString(l + "\n")
		}
	} else {
		// Sync bitti: kullanıcı kaydırabilir
		end := sc + visH
		if end > len(lines) {
			end = len(lines)
		}
		for _, l := range lines[sc:end] {
			b.WriteString(l + "\n")
		}
		// Konum göstergesi
		if len(lines) > visH {
			pos := min(sc+visH, len(lines))
			b.WriteString(stDim.Render(fmt.Sprintf("  [%d/%d]", pos, len(lines))) + "\n")
		}
	}

	// Özet + hint (sync bittikten sonra)
	if m.finished {
		b.WriteString("\n")
		summary := fmt.Sprintf(lang.L.SyncDoneFmt, stOK.Render(fmt.Sprintf("%d", m.done)))
		if m.failed > 0 {
			summary += fmt.Sprintf(lang.L.SyncDoneFailFmt, stFail.Render(fmt.Sprintf("%d", m.failed)))
		}
		b.WriteString(stBold.Render(summary) + "\n")
		b.WriteString(stDim.Render(fmt.Sprintf(lang.L.SyncReportFmt,
			humanBytes(m.bytesDone), time.Since(m.startTime).Round(time.Second))) + "\n")
		if m.statusMsg != "" {
			b.WriteString(stOK.Render(m.statusMsg))
		}

		logHint := ""
		if m.configDir != "" {
			logHint = lang.L.SyncLogHint
		}
		if len(lines) > visH {
			b.WriteString("\n" + stDim.Render(lang.L.SyncScrollHint+logHint) + "\n")
		} else {
			b.WriteString("\n" + stDim.Render(lang.L.SyncAnyKey+logHint) + "\n")
		}
	}

	return b.String()
}

// plainLog TUI log satırlarını + özeti düz metin olarak döner (log dosyası için).
func (m syncTUIModel) plainLog() string {
	var b strings.Builder
	b.WriteString("══ " + m.serverName + " ══\n\n")
	for _, l := range m.buildLogLines() {
		b.WriteString(l + "\n")
	}
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf(lang.L.SyncDoneFmt, fmt.Sprintf("%d", m.done)))
	if m.failed > 0 {
		b.WriteString(fmt.Sprintf(lang.L.SyncDoneFailFmt, fmt.Sprintf("%d", m.failed)))
	}
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf(lang.L.SyncReportFmt,
		humanBytes(m.bytesDone), time.Since(m.startTime).Round(time.Second)) + "\n")
	return b.String()
}

// RunSyncTUI sync işlemini TUI ile gösterir ve tamamlanan yükleme sonuçlarını döner.
// dryRun=true ise yükleme yapmadan dosya listesini gösterir.
// ch=nil ise sadece "değişiklik yok" gösterir.
// configDir boş değilse sync bitince s tuşu logu .syncftp/logs/ altına kaydeder.
func RunSyncTUI(configDir, serverName string, total int, dryRun bool, dryFiles []string, ch <-chan ftpclient.UploadResult) ([]ftpclient.UploadResult, error) {
	m := newSyncTUI(configDir, serverName, total, dryRun, dryFiles, ch)
	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return nil, err
	}
	fm := final.(syncTUIModel)
	return fm.results, nil
}

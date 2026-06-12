package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	ftpclient "syncftp/internal/ftp"
	"syncftp/internal/lang"
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
	total      int
	dryRun     bool
	dryFiles   []string // dry-run dosya listesi

	done      int
	failed    int
	results   []ftpclient.UploadResult
	current   string
	finished  bool
	spinner   int
	width     int

	resultCh <-chan ftpclient.UploadResult
}

func newSyncTUI(serverName string, total int, dryRun bool, dryFiles []string, ch <-chan ftpclient.UploadResult) syncTUIModel {
	return syncTUIModel{
		serverName: serverName,
		total:      total,
		dryRun:     dryRun,
		dryFiles:   dryFiles,
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

func (m syncTUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width

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
			m.current = r.RelPath
		}
		return m, m.waitResult()

	case syncAllDoneMsg:
		m.finished = true

	case tea.KeyMsg:
		if m.finished || m.dryRun || m.total == 0 {
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

	// Progress bar
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

	// Spinner + mevcut dosya
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
	}

	// Son N sonuç (kaydırmalı log)
	b.WriteString("\n")
	maxLog := 12
	start := 0
	if len(m.results) > maxLog {
		start = len(m.results) - maxLog
	}
	for _, r := range m.results[start:] {
		if r.Err != nil {
			retry := ""
			if r.Attempts > 1 {
				retry = stDim.Render(fmt.Sprintf(lang.L.SyncRetryFmt, r.Attempts))
			}
			b.WriteString(fmt.Sprintf("  %s %s%s\n",
				stFail.Render("✗"), r.RelPath, retry))
		} else {
			retry := ""
			if r.Attempts > 1 {
				retry = stDim.Render(fmt.Sprintf(lang.L.SyncRetryOkFmt, r.Attempts))
			}
			b.WriteString(fmt.Sprintf("  %s %s%s\n",
				stOK.Render("✓"), r.RelPath, retry))
		}
	}

	// Özet
	if m.finished {
		b.WriteString("\n")
		summary := fmt.Sprintf(lang.L.SyncDoneFmt, stOK.Render(fmt.Sprintf("%d", m.done)))
		if m.failed > 0 {
			summary += fmt.Sprintf(lang.L.SyncDoneFailFmt, stFail.Render(fmt.Sprintf("%d", m.failed)))
		}
		b.WriteString(stBold.Render(summary) + "\n")
		b.WriteString("\n" + stDim.Render(lang.L.SyncAnyKey) + "\n")
	}

	return b.String()
}

// RunSyncTUI sync işlemini TUI ile gösterir ve tamamlanan yükleme sonuçlarını döner.
// dryRun=true ise yükleme yapmadan dosya listesini gösterir.
// ch=nil ise sadece "değişiklik yok" gösterir.
func RunSyncTUI(serverName string, total int, dryRun bool, dryFiles []string, ch <-chan ftpclient.UploadResult) ([]ftpclient.UploadResult, error) {
	m := newSyncTUI(serverName, total, dryRun, dryFiles, ch)
	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return nil, err
	}
	fm := final.(syncTUIModel)
	return fm.results, nil
}

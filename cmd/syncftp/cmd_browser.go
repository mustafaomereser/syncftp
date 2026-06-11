package main

import (
	"fmt"
	"path"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	goftp "github.com/jlaffaye/ftp"

	ftpclient "syncftp/internal/ftp"
)

// ── stiller ───────────────────────────────────────────────────────────────────

var (
	styleCursor   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	styleDir      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4"))
	styleFile     = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	styleHeader   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5"))
	styleHint     = lipgloss.NewStyle().Faint(true)
	styleSelected = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))
	styleErr      = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	styleMeta     = lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("3"))
)

// ── sonuç ─────────────────────────────────────────────────────────────────────

type BrowserResult struct {
	CWD      string
	Selected string
	Quit     bool
}

// ── sıralanmış giriş ──────────────────────────────────────────────────────────

type sortedEntry struct {
	entry *goftp.Entry
}

func sortEntries(entries []*goftp.Entry) []sortedEntry {
	var dirs, files []sortedEntry
	for _, e := range entries {
		if e.Name == "." || e.Name == ".." {
			continue
		}
		se := sortedEntry{entry: e}
		if e.Type == goftp.EntryTypeFolder {
			dirs = append(dirs, se)
		} else {
			files = append(files, se)
		}
	}
	sort.Slice(dirs, func(i, j int) bool {
		return strings.ToLower(dirs[i].entry.Name) < strings.ToLower(dirs[j].entry.Name)
	})
	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(files[i].entry.Name) < strings.ToLower(files[j].entry.Name)
	})
	return append(dirs, files...)
}

// ── mesajlar ──────────────────────────────────────────────────────────────────

type entriesLoadedMsg struct{ entries []*goftp.Entry }
type browseErrMsg struct{ err error }
type childCountsMsg map[string]int // klasör adı → sayı (-1 = çok boyutlu, -2 = hata)

// ── model ─────────────────────────────────────────────────────────────────────

type browserModel struct {
	client      *ftpclient.Client
	root        string
	cwd         string
	entries     []sortedEntry
	childCounts map[string]int
	cursor      int
	loading     bool
	err         error
	result      *BrowserResult
	width       int
	height      int
}

func newBrowserModel(client *ftpclient.Client, startPath, root string) browserModel {
	return browserModel{
		client:      client,
		root:        root,
		cwd:         startPath,
		loading:     true,
		childCounts: make(map[string]int),
	}
}

func (m browserModel) Init() tea.Cmd {
	return m.fetchEntries()
}

func (m browserModel) fetchEntries() tea.Cmd {
	cwd := m.cwd
	client := m.client
	return func() tea.Msg {
		entries, err := client.List(cwd)
		if err != nil {
			return browseErrMsg{err}
		}
		return entriesLoadedMsg{entries}
	}
}

// fetchChildCounts: listelenen klasörlerin içindeki öğe sayısını arka planda çeker.
func (m browserModel) fetchChildCounts(entries []sortedEntry) tea.Cmd {
	cwd := m.cwd
	client := m.client
	var folders []string
	for _, e := range entries {
		if e.entry.Type == goftp.EntryTypeFolder {
			folders = append(folders, e.entry.Name)
		}
	}
	if len(folders) == 0 {
		return nil
	}
	return func() tea.Msg {
		counts := make(map[string]int, len(folders))
		for _, name := range folders {
			children, err := client.List(path.Join(cwd, name))
			if err != nil {
				counts[name] = -2
				continue
			}
			n := 0
			for _, c := range children {
				if c.Name != "." && c.Name != ".." {
					n++
				}
			}
			if n > 999 {
				counts[name] = -1 // çok boyutlu
			} else {
				counts[name] = n
			}
		}
		return childCountsMsg(counts)
	}
}

func (m browserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case entriesLoadedMsg:
		m.loading = false
		m.err = nil
		m.entries = sortEntries(msg.entries)
		m.childCounts = make(map[string]int)
		if m.cursor >= len(m.entries) {
			m.cursor = max(0, len(m.entries)-1)
		}
		return m, m.fetchChildCounts(m.entries)

	case childCountsMsg:
		for k, v := range msg {
			m.childCounts[k] = v
		}

	case browseErrMsg:
		m.loading = false
		m.err = msg.err

	case tea.KeyMsg:
		if m.loading {
			break
		}
		switch msg.String() {
		case "ctrl+c":
			m.result = &BrowserResult{CWD: m.cwd, Quit: true}
			return m, tea.Quit

		case "esc":
			if m.cwd != m.root && m.cwd != "/" {
				return m.goUp()
			}
			m.result = &BrowserResult{CWD: m.cwd, Quit: true}
			return m, tea.Quit

		case "q", "Q":
			m.result = &BrowserResult{CWD: m.cwd, Quit: true}
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.entries)-1 {
				m.cursor++
			}

		case "left", "h":
			if m.cwd != m.root && m.cwd != "/" {
				return m.goUp()
			}

		case "right", "l", "enter":
			if len(m.entries) == 0 {
				break
			}
			e := m.entries[m.cursor]
			if e.entry.Type == goftp.EntryTypeFolder {
				m.cwd = path.Join(m.cwd, e.entry.Name)
				m.loading = true
				m.cursor = 0
				return m, m.fetchEntries()
			}
			m.result = &BrowserResult{
				CWD:      m.cwd,
				Selected: path.Join(m.cwd, e.entry.Name),
			}
			return m, tea.Quit

		case "g":
			m.cursor = 0

		case "G":
			if len(m.entries) > 0 {
				m.cursor = len(m.entries) - 1
			}
		}
	}
	return m, nil
}

func (m browserModel) goUp() (tea.Model, tea.Cmd) {
	parent := path.Dir(m.cwd)
	if parent == "." {
		parent = "/"
	}
	m.cwd = parent
	m.loading = true
	m.cursor = 0
	return m, m.fetchEntries()
}

func (m browserModel) View() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(styleHeader.Render("  📁 "+m.cwd) + "\n")
	b.WriteString(styleHint.Render("  ↑↓/jk  Gezin    →/Enter  Gir/Seç    ←/ESC  Üst dizin    g/G  İlk/Son    q  Çık") + "\n")
	b.WriteString(styleHint.Render("  "+strings.Repeat("─", 70)) + "\n\n")

	if m.loading {
		b.WriteString("  Yükleniyor...\n")
		return b.String()
	}

	if m.err != nil {
		b.WriteString(styleErr.Render(fmt.Sprintf("  Hata: %v", m.err)) + "\n")
		return b.String()
	}

	if len(m.entries) == 0 {
		b.WriteString(styleHint.Render("  (boş dizin)") + "\n")
		return b.String()
	}

	// Görünür alan
	visibleLines := m.height - 10
	if visibleLines < 5 {
		visibleLines = 5
	}
	start := 0
	if m.cursor >= visibleLines {
		start = m.cursor - visibleLines + 1
	}
	end := start + visibleLines
	if end > len(m.entries) {
		end = len(m.entries)
	}

	// Klasör/dosya sınırını bul (separator için)
	dirEnd := 0
	for _, e := range m.entries {
		if e.entry.Type != goftp.EntryTypeFolder {
			break
		}
		dirEnd++
	}

	for i := start; i < end; i++ {
		e := m.entries[i]

		// Klasör/dosya arası ayırıcı
		if i == dirEnd && dirEnd > 0 && dirEnd < len(m.entries) {
			b.WriteString(styleHint.Render("  "+strings.Repeat("·", 70)) + "\n")
		}

		var icon, nameStr, metaStr string

		if e.entry.Type == goftp.EntryTypeFolder {
			icon = "📁"
			nameStr = styleDir.Render(e.entry.Name + "/")

			// Klasör boyut bilgisi
			if count, ok := m.childCounts[e.entry.Name]; ok {
				switch count {
				case -1:
					metaStr = styleMeta.Render("çok boyutlu")
				case -2:
					metaStr = styleMeta.Render("?")
				default:
					metaStr = styleMeta.Render(fmt.Sprintf("%d öğe", count))
				}
			} else {
				metaStr = styleHint.Render("...")
			}
		} else {
			icon = "📄"
			nameStr = styleFile.Render(e.entry.Name)
			metaStr = styleMeta.Render(formatSize(e.entry.Size))
		}

		date := styleHint.Render(e.entry.Time.Format("2006-01-02"))

		if i == m.cursor {
			cursor := styleCursor.Render("▶")
			nameFmt := styleSelected.Render(e.entry.Name)
			if e.entry.Type == goftp.EntryTypeFolder {
				nameFmt = styleSelected.Render(e.entry.Name + "/")
			}
			_ = nameStr
			b.WriteString(fmt.Sprintf(" %s %s  %-42s  %-14s  %s\n",
				cursor, icon, nameFmt, metaStr, date))
		} else {
			b.WriteString(fmt.Sprintf("    %s  %-42s  %-14s  %s\n",
				icon, nameStr, metaStr, date))
		}
	}

	// Alt bilgi: klasör ve dosya sayısı
	nDirs, nFiles := 0, 0
	for _, e := range m.entries {
		if e.entry.Type == goftp.EntryTypeFolder {
			nDirs++
		} else {
			nFiles++
		}
	}
	b.WriteString("\n")
	b.WriteString(styleHint.Render(fmt.Sprintf("  %d klasör, %d dosya  ·  %d/%d",
		nDirs, nFiles, m.cursor+1, len(m.entries))) + "\n")

	return b.String()
}

// RunBrowser tam ekran interaktif FTP dosya tarayıcısını başlatır.
func RunBrowser(client *ftpclient.Client, startPath, root string) (*BrowserResult, error) {
	m := newBrowserModel(client, startPath, root)
	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return nil, err
	}
	fm := final.(browserModel)
	if fm.result == nil {
		return &BrowserResult{CWD: fm.cwd, Quit: true}, nil
	}
	return fm.result, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

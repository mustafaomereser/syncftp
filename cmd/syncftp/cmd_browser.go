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
	styleHeader   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5"))
	styleHint     = lipgloss.NewStyle().Faint(true)
	styleSelected = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))
	styleErr      = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	styleMeta     = lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("3"))
	styleMarked   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("3"))
	styleSearch   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
	stylePreview  = lipgloss.NewStyle().Faint(true)
)

// ── sonuç ─────────────────────────────────────────────────────────────────────

type BrowserResult struct {
	CWD      string
	Selected string   // tek dosya (Enter)
	Marked   []string // çoklu seçim (Space + Enter)
	Action   string   // "delete", "move", ""
	Quit     bool
}

// ── sıralanmış giriş ──────────────────────────────────────────────────────────

type sortedEntry struct {
	entry    *goftp.Entry
	dir      string // recursive aramada dolu: dosyanın bulunduğu dizin
	selectMe bool   // pickDirMode'da "bu dizini seç" sanal girişi
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
type childCountsMsg map[string]int
type previewLoadedMsg struct{ content string }
type recursiveSearchMsg struct {
	results []sortedEntry
	err     error
}

// ── model ─────────────────────────────────────────────────────────────────────

type browserModel struct {
	client      *ftpclient.Client
	root        string
	cwd         string
	entries     []sortedEntry
	filtered    []sortedEntry // arama sonucu
	childCounts map[string]int
	marked      map[string]bool

	cursor  int
	loading bool
	err     error
	result  *BrowserResult
	width   int
	height  int

	// Arama
	searching        bool
	searchText       string
	recursiveSearch  bool // global arama aktif mi
	recursiveLoading bool // FTP'den sonuçlar bekleniyor

	// Önizleme
	preview        string
	previewLoading bool
	previewFile    string

	// Tam ekran önizleme modu
	previewMode   bool
	previewLines  []string
	previewScroll int
	previewName   string
	previewMeta   string // "4.2 KB  2026-06-11 14:30"

	// Mod: klasör seçme (taşıma için)
	pickDirMode bool
}

func newBrowserModel(client *ftpclient.Client, startPath, root string) browserModel {
	return browserModel{
		client:      client,
		root:        root,
		cwd:         startPath,
		loading:     true,
		childCounts: make(map[string]int),
		marked:      make(map[string]bool),
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
				// Bağlantı protokol hatası ise devam etme — sonraki çağrılar da bozulur
				if isConnectionError(err) {
					break
				}
				continue
			}
			n := 0
			for _, c := range children {
				if c.Name != "." && c.Name != ".." {
					n++
				}
			}
			if n > 999 {
				counts[name] = -1
			} else {
				counts[name] = n
			}
		}
		return childCountsMsg(counts)
	}
}

func (m browserModel) fetchPreview(filePath string) tea.Cmd {
	client := m.client
	return func() tea.Msg {
		data, err := client.Preview(filePath, 256*1024)
		if err != nil {
			return previewLoadedMsg{"(önizleme alınamadı)"}
		}
		return previewLoadedMsg{string(data)}
	}
}

func (m browserModel) visible() []sortedEntry {
	var base []sortedEntry
	if (m.searching || m.recursiveSearch) && m.searchText != "" {
		base = m.filtered
	} else {
		base = m.entries
	}
	// pickDirMode'da en üste "bu dizini seç" sanal girişi ekle
	if m.pickDirMode {
		pseudo := sortedEntry{selectMe: true}
		return append([]sortedEntry{pseudo}, base...)
	}
	return base
}

// entryFullPath bir girdinin tam FTP yolunu döner.
func (m browserModel) entryFullPath(e sortedEntry) string {
	if e.dir != "" {
		return path.Join(e.dir, e.entry.Name)
	}
	return path.Join(m.cwd, e.entry.Name)
}

func (m *browserModel) applySearch() {
	q := strings.ToLower(m.searchText)
	m.filtered = nil
	for _, e := range m.entries {
		if strings.Contains(strings.ToLower(e.entry.Name), q) {
			m.filtered = append(m.filtered, e)
		}
	}
	m.cursor = 0
}

// startRecursiveSearch tüm alt dizinleri FTP üzerinden tarayarak sonuç döner.
func (m browserModel) startRecursiveSearch(query string) tea.Cmd {
	client := m.client
	root := m.root
	q := strings.ToLower(query)
	return func() tea.Msg {
		var results []sortedEntry
		var walk func(dir string)
		walk = func(dir string) {
			entries, err := client.List(dir)
			if err != nil {
				return
			}
			for _, e := range entries {
				if e.Name == "." || e.Name == ".." {
					continue
				}
				if e.Type == goftp.EntryTypeFolder {
					walk(path.Join(dir, e.Name))
				} else {
					if strings.Contains(strings.ToLower(e.Name), q) {
						results = append(results, sortedEntry{entry: e, dir: dir})
					}
				}
			}
		}
		walk(root)
		return recursiveSearchMsg{results: results}
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
		m.filtered = nil
		m.childCounts = make(map[string]int)
		if m.cursor >= len(m.entries) {
			m.cursor = max(0, len(m.entries)-1)
		}
		m.preview = ""
		m.previewFile = ""
		return m, m.fetchChildCounts(m.entries)

	case childCountsMsg:
		for k, v := range msg {
			m.childCounts[k] = v
		}

	case previewLoadedMsg:
		m.previewLoading = false
		m.preview = msg.content
		m.previewLines = strings.Split(msg.content, "\n")
		// split sondaki boş satırı temizle
		if len(m.previewLines) > 0 && m.previewLines[len(m.previewLines)-1] == "" {
			m.previewLines = m.previewLines[:len(m.previewLines)-1]
		}

	case recursiveSearchMsg:
		m.recursiveLoading = false
		m.recursiveSearch = true
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.filtered = msg.results
			m.cursor = 0
		}

	case browseErrMsg:
		m.loading = false
		m.err = msg.err

	case tea.KeyMsg:
		// Loading veya recursive arama sırasında bile çıkışa izin ver
		if m.loading || m.recursiveLoading {
			switch msg.String() {
			case "ctrl+c", "q", "Q":
				m.result = &BrowserResult{CWD: m.cwd, Quit: true}
				return m, tea.Quit
			}
			break
		}

		// Tam ekran önizleme modu — tüm tuşlar burada tüketilir
		if m.previewMode {
			previewRows := m.height - 6
			if previewRows < 2 {
				previewRows = 2
			}
			maxScroll := len(m.previewLines) - previewRows
			if maxScroll < 0 {
				maxScroll = 0
			}
			switch msg.String() {
			case "ctrl+c":
				m.result = &BrowserResult{CWD: m.cwd, Quit: true}
				return m, tea.Quit
			case "esc", "q", "Q", "left", "h":
				m.previewMode = false
			case "up", "k":
				if m.previewScroll > 0 {
					m.previewScroll--
				}
			case "down", "j":
				if m.previewScroll < maxScroll {
					m.previewScroll++
				}
			case "pgup":
				m.previewScroll -= previewRows - 2
				if m.previewScroll < 0 {
					m.previewScroll = 0
				}
			case "pgdown":
				m.previewScroll += previewRows - 2
				if m.previewScroll > maxScroll {
					m.previewScroll = maxScroll
				}
			case "g":
				m.previewScroll = 0
			case "G":
				m.previewScroll = maxScroll
			}
			return m, nil
		}

		// Arama modu — tüm tuşlar burada tüketilir, outer switch'e düşmez
		if m.searching {
			switch msg.String() {
			case "esc":
				m.searching = false
				m.searchText = ""
				m.filtered = nil
				m.recursiveSearch = false
				m.recursiveLoading = false
				m.cursor = 0
				m.preview = ""
				m.previewFile = ""
				m.previewLoading = false
			case "enter":
				if m.searchText != "" {
					m.searching = false
					m.recursiveLoading = true
					m.filtered = nil
					m.cursor = 0
					return m, m.startRecursiveSearch(m.searchText)
				}
				m.searching = false
			case "backspace":
				if len(m.searchText) > 0 {
					// Rune tabanlı kes — byte kes multi-byte char'ı bozar
					r := []rune(m.searchText)
					m.searchText = string(r[:len(r)-1])
					m.recursiveSearch = false
					m.applySearch()
				}
			default:
				s := msg.String()
				// Windows Terminal bazı tuşları "alt+alt+x" olarak raporluyor.
				// alt+ prefix'lerini soy, gerçek karakteri al.
				for strings.HasPrefix(s, "alt+") {
					s = strings.TrimPrefix(s, "alt+")
				}
				r := []rune(s)
				if len(r) == 1 && r[0] >= 0x20 && r[0] != 0x7F {
					m.searchText += s
					m.recursiveSearch = false
					m.applySearch()
				}
			}
			// Arama modunda outer switch'e düşme
			return m, nil
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

		case "/":
			m.searching = true
			m.searchText = ""
			return m, nil

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			// Önizleme cache'ini temizle — farklı dosyaya geçildi
			m.preview = ""
			m.previewFile = ""
			m.previewLines = nil
			m.previewLoading = false

		case "down", "j":
			if m.cursor < len(m.visible())-1 {
				m.cursor++
			}
			m.preview = ""
			m.previewFile = ""
			m.previewLines = nil
			m.previewLoading = false

		case "left", "h":
			if m.cwd != m.root && m.cwd != "/" {
				return m.goUp()
			}

		case "right", "l":
			vis := m.visible()
			if len(vis) == 0 {
				break
			}
			e := vis[m.cursor]
			if e.selectMe {
				break // sanal giriş — sağ ok'ta bir şey yapma
			}
			if e.entry.Type == goftp.EntryTypeFolder {
				dirPath := path.Join(m.cwd, e.entry.Name)
				if e.dir != "" {
					dirPath = path.Join(e.dir, e.entry.Name)
				}
				m.cwd = dirPath
				m.loading = true
				m.cursor = 0
				m.searching = false
				m.recursiveSearch = false
				m.recursiveLoading = false
				m.searchText = ""
				return m, m.fetchEntries()
			}
			// Dosya: tam ekran önizleme moduna geç
			fullPath := m.entryFullPath(e)
			m.previewMode = true
			m.previewScroll = 0
			m.previewName = e.entry.Name
			m.previewMeta = formatSize(e.entry.Size) + "   " + e.entry.Time.Format("2006-01-02 15:04")
			// İçerik zaten yüklüyse (aynı dosya + satırlar var) yeniden fetch etme
			if fullPath == m.previewFile && m.previewLines != nil {
				return m, nil
			}
			m.previewFile = fullPath
			m.previewLoading = true
			m.preview = ""
			m.previewLines = nil
			return m, m.fetchPreview(fullPath)

		case "enter":
			vis := m.visible()
			if len(vis) == 0 {
				break
			}
			e := vis[m.cursor]

			// "Bu dizini seç" sanal girişi
			if e.selectMe {
				m.result = &BrowserResult{CWD: m.cwd, Selected: m.cwd}
				return m, tea.Quit
			}

			if e.entry.Type == goftp.EntryTypeFolder {
				if m.pickDirMode {
					// Klasör seç modunda: Enter bu klasörü seçer
					selectedDir := path.Join(m.cwd, e.entry.Name)
					if e.dir != "" {
						selectedDir = path.Join(e.dir, e.entry.Name)
					}
					m.result = &BrowserResult{CWD: selectedDir, Selected: selectedDir}
					return m, tea.Quit
				}
				dirPath := path.Join(m.cwd, e.entry.Name)
				if e.dir != "" {
					dirPath = path.Join(e.dir, e.entry.Name)
				}
				m.cwd = dirPath
				m.loading = true
				m.cursor = 0
				m.searching = false
				m.recursiveSearch = false
				m.recursiveLoading = false
				m.searchText = ""
				return m, m.fetchEntries()
			}
			fullPath := m.entryFullPath(e)
			if len(m.marked) > 0 {
				m.result = &BrowserResult{CWD: m.cwd, Marked: m.markedList()}
				return m, tea.Quit
			}
			m.result = &BrowserResult{CWD: m.cwd, Selected: fullPath}
			return m, tea.Quit

		case "d", "D":
			if len(m.marked) == 0 {
				break
			}
			m.result = &BrowserResult{CWD: m.cwd, Marked: m.markedList(), Action: "delete"}
			return m, tea.Quit

		case "m", "M":
			if len(m.marked) == 0 {
				break
			}
			m.result = &BrowserResult{CWD: m.cwd, Marked: m.markedList(), Action: "move"}
			return m, tea.Quit

		case " ":
			vis := m.visible()
			if len(vis) == 0 {
				break
			}
			e := vis[m.cursor]
			if e.selectMe || e.entry == nil {
				break
			}
			fullPath := m.entryFullPath(e)
			if m.marked[fullPath] {
				delete(m.marked, fullPath)
			} else {
				m.marked[fullPath] = true
			}
			if m.cursor < len(vis)-1 {
				m.cursor++
			}
			m.preview = ""
			m.previewFile = ""
			m.previewLoading = false

		case "a", "A":
			// Arama/recursive arama sırasında toplu işaret devre dışı
			if m.searching || m.recursiveSearch {
				break
			}
			vis := m.visible()
			allMarked := true
			for _, e := range vis {
				if e.selectMe || e.entry == nil {
					continue
				}
				fp := m.entryFullPath(e)
				if !m.marked[fp] {
					allMarked = false
					break
				}
			}
			for _, e := range vis {
				if e.selectMe || e.entry == nil {
					continue
				}
				fp := m.entryFullPath(e)
				if allMarked {
					delete(m.marked, fp)
				} else {
					m.marked[fp] = true
				}
			}

		case "g":
			m.cursor = 0
			m.preview = ""
			m.previewFile = ""
			m.previewLoading = false

		case "G":
			vis := m.visible()
			if len(vis) > 0 {
				m.cursor = len(vis) - 1
			}
			m.preview = ""
			m.previewFile = ""
			m.previewLoading = false
		}
	}
	return m, nil
}


func (m browserModel) markedList() []string {
	var out []string
	for p := range m.marked {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

func (m browserModel) goUp() (tea.Model, tea.Cmd) {
	parent := path.Dir(m.cwd)
	if parent == "." {
		parent = "/"
	}
	m.cwd = parent
	m.loading = true
	m.cursor = 0
	m.searching = false
	m.searchText = ""
	return m, m.fetchEntries()
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m browserModel) View() string {
	w := m.width
	if w < 40 {
		w = 80
	}
	h := m.height
	if h < 8 {
		h = 24
	}

	// Tam ekran önizleme modu
	if m.previewMode {
		return m.viewPreview(w, h)
	}

	vis := m.visible()
	listW := w

	// ── kolon genişlikleri ──
	// Satır düzeni (görsel kolon sayısı):
	//   cur(2) mark(2) icon(3=emoji2+sp1) name(nameW) sp(2) meta(12) sp(2) date(10)
	//   toplam sabit = 2+2+3+2+12+2+10 = 33
	nameW := listW - 33
	if nameW < 12 {
		nameW = 12
	}
	colName := lipgloss.NewStyle().Width(nameW).MaxWidth(nameW)
	colMeta := lipgloss.NewStyle().Width(12).MaxWidth(12).Faint(true).Foreground(lipgloss.Color("3"))
	colDate := lipgloss.NewStyle().Width(10).MaxWidth(10).Faint(true)

	// ── başlık ──
	crumbs := buildBreadcrumb(m.cwd, m.root, listW-4)
	header := styleHeader.Render("  " + crumbs)

	// ── hint / arama satırı ──
	sep := styleHint.Render("  |  ")
	var hintLine string
	if m.searching {
		hintLine = styleSearch.Render(fmt.Sprintf("  /%s_", m.searchText)) +
			styleHint.Render("  Enter=Tümünü Ara  ESC=İptal")
	} else if m.recursiveLoading {
		hintLine = styleSearch.Render(fmt.Sprintf("  Aranıyor: %q ...", m.searchText))
	} else if m.recursiveSearch && m.searchText != "" {
		hintLine = styleSearch.Render(fmt.Sprintf("  /%s", m.searchText)) +
			styleHint.Render(fmt.Sprintf("  (tüm sunucu, %d sonuç)  ESC temizle", len(vis)))
	} else if m.pickDirMode {
		hintLine = styleCursor.Render("  HEDEF DİZİN SEÇ") +
			sep + styleHint.Render("Enter=Bu Klasörü Seç") +
			sep + styleHint.Render("→ Gir") +
			sep + styleHint.Render("← Üst") +
			sep + styleHint.Render("q İptal")
	} else if len(m.marked) > 0 {
		hintLine = styleMarked.Render(fmt.Sprintf("  ✓ %d işaretli", len(m.marked))) +
			sep + styleHint.Render("d=Sil") +
			sep + styleHint.Render("m=Taşı") +
			sep + styleHint.Render("a=Tümünü kaldır") +
			sep + styleHint.Render("ESC=İptal")
	} else if m.searchText != "" {
		hintLine = styleSearch.Render("  /"+m.searchText) + styleHint.Render("  ESC temizle")
	} else {
		hintLine = styleHint.Render("  ↑↓ Gezin") +
			sep + styleHint.Render("Enter Gir/Seç") +
			sep + styleHint.Render("→ Önizle") +
			sep + styleHint.Render("← Üst") +
			sep + styleHint.Render("Space İşaretle") +
			sep + styleHint.Render("/ Ara") +
			sep + styleHint.Render("q Çık")
	}

	divider := styleHint.Render(strings.Repeat("─", listW))

	if m.loading {
		return "\n" + header + "\n" + hintLine + "\n" + divider + "\n\n  Yükleniyor...\n"
	}
	if m.err != nil {
		return "\n" + header + "\n" + divider + "\n\n" + styleErr.Render("  Hata: "+m.err.Error()) + "\n"
	}

	// Görünür satır sayısı.
	// Ekranda sabit satırlar:
	//   \n(1) header(1) hint(1) divider(1) \n(1) [body] \n(1) footer(1) \n(1) = 8 satır
	visibleRows := h - 8
	if visibleRows < 2 {
		visibleRows = 2
	}

	// scroll penceresi
	start := 0
	if m.cursor >= visibleRows {
		start = m.cursor - visibleRows + 1
	}
	end := start + visibleRows
	if end > len(vis) {
		end = len(vis)
	}

	// klasör↔dosya sınırı
	dirEnd := 0
	for _, e := range vis {
		if e.selectMe {
			dirEnd++ // sanal giriş sayılmaz ama offset'i kaydırır
			continue
		}
		if e.entry.Type != goftp.EntryTypeFolder {
			break
		}
		dirEnd++
	}

	// ── satırları oluştur ──
	var listBuf strings.Builder
	for i := start; i < end; i++ {
		e := vis[i]

		// pickDirMode sanal "bu dizini seç" girişi
		if e.selectMe {
			if i == m.cursor {
				listBuf.WriteString(styleCursor.Render("▶ ") +
					styleSelected.Bold(true).Render("[ Bu dizini seç: "+m.cwd+" ]") + "\n")
			} else {
				listBuf.WriteString("  " + styleHint.Render("[ Bu dizini seç: "+m.cwd+" ]") + "\n")
			}
			continue
		}

		// ayırıcı çizgi (klasörler bitti, dosyalar başlıyor)
		if i == dirEnd && dirEnd > 0 && dirEnd < len(vis) {
			listBuf.WriteString(styleHint.Render(strings.Repeat("·", listW)) + "\n")
		}

		fullPath := path.Join(m.cwd, e.entry.Name)
		isMarked := m.marked[fullPath]
		isDir := e.entry.Type == goftp.EntryTypeFolder

		icon := "📄"
		rawName := e.entry.Name
		if isDir {
			icon = "📁"
			rawName += "/"
		}
		// Recursive aramada: ismin önüne kısa dizin yolu ekle
		if e.dir != "" {
			rel := strings.TrimPrefix(e.dir, m.root)
			if rel == "" {
				rel = "/"
			}
			rawName = rel + "/" + e.entry.Name
		}

		var metaStr string
		if isDir {
			if count, ok := m.childCounts[e.entry.Name]; ok {
				switch count {
				case -1:
					metaStr = "çok boyutlu"
				case -2:
					metaStr = "?"
				default:
					metaStr = fmt.Sprintf("%d öğe", count)
				}
			} else {
				metaStr = "..."
			}
		} else {
			metaStr = formatSize(e.entry.Size)
		}

		date := e.entry.Time.Format("2006-01-02")
		markChar := " "
		if isMarked {
			markChar = "✓"
		}

		// lipgloss ile her hücre sabit genişlikte render ediliyor
		metaCell := colMeta.Render(metaStr)
		dateCell := colDate.Render(date)

		if i == m.cursor {
			cur := styleCursor.Render("▶ ")
			mark := styleMarked.Render(markChar + " ")
			ico := styleCursor.Render(icon + " ")
			var nameCell string
			if isMarked {
				nameCell = colName.Bold(true).Foreground(lipgloss.Color("3")).Render(rawName)
			} else if isDir {
				nameCell = colName.Bold(true).Foreground(lipgloss.Color("6")).Render(rawName)
			} else {
				nameCell = colName.Bold(true).Foreground(lipgloss.Color("2")).Render(rawName)
			}
			listBuf.WriteString(cur + mark + ico + nameCell + "  " + metaCell + "  " + dateCell + "\n")
		} else {
			cur := "  "
			mark := styleHint.Render(markChar + " ")
			if isMarked {
				mark = styleMarked.Render(markChar + " ")
			}
			ico := styleHint.Render(icon + " ")
			var nameCell string
			if isMarked {
				nameCell = colName.Bold(true).Foreground(lipgloss.Color("3")).Render(rawName)
			} else if isDir {
				nameCell = colName.Foreground(lipgloss.Color("4")).Render(rawName)
			} else {
				nameCell = colName.Foreground(lipgloss.Color("7")).Render(rawName)
			}
			listBuf.WriteString(cur + mark + ico + nameCell + "  " + metaCell + "  " + dateCell + "\n")
		}
	}

	body := listBuf.String()

	// ── alt bilgi ──
	nDirs, nFiles := 0, 0
	for _, e := range m.entries {
		if e.entry.Type == goftp.EntryTypeFolder {
			nDirs++
		} else {
			nFiles++
		}
	}
	pos := m.cursor + 1
	total := len(vis)
	if total == 0 {
		pos = 0
	}
	footerText := fmt.Sprintf("  %d klasör, %d dosya  ·  %d/%d", nDirs, nFiles, pos, total)
	if m.searchText != "" {
		footerText += styleSearch.Render(fmt.Sprintf("  ara:%q %d eşleşme", m.searchText, len(vis)))
	}
	footer := styleHint.Render(footerText)

	return "\n" + header + "\n" + hintLine + "\n" + divider + "\n" + body + "\n" + footer + "\n"
}

// viewPreview tam ekran önizleme görünümünü döner.
func (m browserModel) viewPreview(w, h int) string {
	sep := styleHint.Render("  |  ")
	divider := styleHint.Render(strings.Repeat("─", w))

	header := styleHeader.Render("  📄 "+m.previewName) +
		"   " + styleMeta.Render(m.previewMeta)

	if m.previewLoading || m.previewLines == nil {
		hint := styleHint.Render("  ESC/q geri")
		return "\n" + header + "\n" + divider + "\n\n" +
			styleHint.Render("  Yükleniyor...") + "\n\n" +
			divider + "\n" + hint + "\n"
	}

	// Görünür satır sayısı: \n(1) header(1) divider(1) body hint(1) divider(1) hint(1) = 6 sabit
	previewRows := h - 6
	if previewRows < 2 {
		previewRows = 2
	}

	lines := m.previewLines
	total := len(lines)
	maxScroll := total - previewRows
	if maxScroll < 0 {
		maxScroll = 0
	}
	scroll := m.previewScroll
	if scroll > maxScroll {
		scroll = maxScroll
	}
	if scroll < 0 {
		scroll = 0
	}

	endLine := scroll + previewRows
	if endLine > total {
		endLine = total
	}

	var bodyBuf strings.Builder
	for _, line := range lines[scroll:endLine] {
		runes := []rune(line)
		if len(runes) > w-2 && w > 4 {
			runes = runes[:w-2]
		}
		bodyBuf.WriteString("  " + stylePreview.Render(string(runes)) + "\n")
	}

	var posText string
	if total == 0 {
		posText = "(boş dosya)"
	} else {
		posText = fmt.Sprintf("%d-%d / %d satır", scroll+1, endLine, total)
	}

	hint := styleHint.Render("  ESC/q geri") +
		sep + styleHint.Render("↑↓ kaydır") +
		sep + styleHint.Render("PgUp/PgDn") +
		sep + styleHint.Render("g başa  G sona") +
		"   " + styleMeta.Render(posText)

	return "\n" + header + "\n" + divider + "\n" +
		bodyBuf.String() + divider + "\n" + hint + "\n"
}

// isConnectionError bağlantı protokolünün bozulduğunu gösteren hataları tanır.
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "short response") ||
		strings.Contains(msg, "EOF") ||
		strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "use of closed")
}

// truncatePad saf metin s'yi w görsel genişliğine trun/pad eder (ANSI yok).
func truncatePad(s string, w int) string {
	runes := []rune(s)
	if len(runes) > w {
		if w > 1 {
			return string(runes[:w-1]) + "…"
		}
		return string(runes[:w])
	}
	return s + strings.Repeat(" ", w-len(runes))
}

// buildBreadcrumb yolu tıklanabilir parçalara böler.
func buildBreadcrumb(cwd, root string, maxW int) string {
	parts := strings.Split(strings.TrimPrefix(cwd, "/"), "/")
	var crumbs []string
	for _, p := range parts {
		if p != "" {
			crumbs = append(crumbs, styleHint.Render(p))
		}
	}
	result := "/" + strings.Join(crumbs, styleHint.Render("/"))
	if len(result) > maxW {
		result = ".../" + crumbs[len(crumbs)-1]
	}
	return result
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

// RunBrowserPickDir hedef klasör seçmek için browser açar.
// Enter tuşu klasöre girmek yerine onu seçer.
func RunBrowserPickDir(client *ftpclient.Client, startPath, root string) (string, error) {
	m := newBrowserModel(client, startPath, root)
	m.pickDirMode = true
	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return "", err
	}
	fm := final.(browserModel)
	if fm.result == nil || fm.result.Quit {
		return "", nil
	}
	return fm.result.Selected, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

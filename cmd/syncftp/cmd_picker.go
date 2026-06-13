package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"syncftp/internal/lang"
)

// ── stiller ───────────────────────────────────────────────────────────────────

var (
	stylePickerTitle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5")).Padding(0, 2)
	stylePickerSub      = lipgloss.NewStyle().Faint(true).Padding(0, 2)
	stylePickerCursor   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	stylePickerSelected = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))
	stylePickerNormal   = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	stylePickerDesc     = lipgloss.NewStyle().Faint(true)
	stylePickerDivider  = lipgloss.NewStyle().Faint(true)
	stylePickerHint     = lipgloss.NewStyle().Faint(true).Italic(true).Padding(0, 2)

	styleCheckOn  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))
	styleCheckOff = lipgloss.NewStyle().Faint(true)
)

// ── PickerItem ────────────────────────────────────────────────────────────────

type PickerItem struct {
	Icon  string // opsiyonel emoji/simge
	Label string // ana metin
	Desc  string // sağda veya altında gri metin
	Value string // döndürülecek değer
}

// ── tek seçim ─────────────────────────────────────────────────────────────────

type pickerModel struct {
	title    string
	subtitle string
	all      []PickerItem // tüm öğeler
	filtered []PickerItem // arama sonrası
	search   string
	cursor   int
	result   string
	quit     bool
	width    int
	height   int
}

func newPickerModel(title, subtitle string, items []PickerItem) pickerModel {
	return pickerModel{title: title, subtitle: subtitle, all: items, filtered: items}
}

func (m *pickerModel) applySearch() {
	if m.search == "" {
		m.filtered = m.all
		return
	}
	q := strings.ToLower(m.search)
	m.filtered = nil
	for _, item := range m.all {
		if strings.Contains(strings.ToLower(item.Label), q) ||
			strings.Contains(strings.ToLower(item.Desc), q) {
			m.filtered = append(m.filtered, item)
		}
	}
	m.cursor = 0
}

func (m pickerModel) Init() tea.Cmd { return nil }

func (m pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quit = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
		case "enter", "right", "l":
			if len(m.filtered) > 0 {
				m.result = m.filtered[m.cursor].Value
				return m, tea.Quit
			}
		case "backspace":
			if len(m.search) > 0 {
				m.search = m.search[:len(m.search)-1]
				m.applySearch()
			}
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			n := int(msg.String()[0]-'0') - 1
			if n >= 0 && n < len(m.filtered) {
				m.result = m.filtered[n].Value
				return m, tea.Quit
			}
		default:
			// Yazdırılabilir karakter → fuzzy search
			if len(msg.String()) == 1 && msg.String() != "q" {
				m.search += msg.String()
				m.applySearch()
			} else if msg.String() == "q" {
				if m.search != "" {
					m.search += "q"
					m.applySearch()
				} else {
					m.quit = true
					return m, tea.Quit
				}
			}
		}
	}
	return m, nil
}

func (m pickerModel) View() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(stylePickerTitle.Render(m.title) + "\n")
	if m.subtitle != "" {
		b.WriteString(stylePickerSub.Render(m.subtitle) + "\n")
	}
	w := 60
	if m.width > 10 {
		w = m.width - 4
	}
	b.WriteString(stylePickerDivider.Render("  "+strings.Repeat("─", w)) + "\n")

	// Arama kutusu
	if m.search != "" {
		b.WriteString(stylePickerSelected.Render(fmt.Sprintf("  🔍 %s", m.search)) +
			stylePickerDesc.Render(fmt.Sprintf(lang.L.PickerMatchFmt, len(m.filtered))) + "\n")
	} else {
		b.WriteString(stylePickerHint.Render(lang.L.PickerFilterHint) + "\n")
	}
	b.WriteString(stylePickerDivider.Render("  "+strings.Repeat("─", w)) + "\n\n")

	for i, item := range m.filtered {
		icon := item.Icon
		if icon == "" {
			icon = " "
		}
		num := stylePickerDesc.Render(fmt.Sprintf("[%d]", i+1))
		if i == m.cursor {
			cur := stylePickerCursor.Render("▶")
			lbl := stylePickerSelected.Render(item.Label)
			desc := stylePickerDesc.Render(item.Desc)
			b.WriteString(fmt.Sprintf(" %s %s %s  %-30s  %s\n", cur, num, icon, lbl, desc))
		} else {
			lbl := stylePickerNormal.Render(item.Label)
			desc := stylePickerDesc.Render(item.Desc)
			b.WriteString(fmt.Sprintf("    %s %s  %-30s  %s\n", num, icon, lbl, desc))
		}
	}

	if len(m.filtered) == 0 {
		b.WriteString(stylePickerHint.Render(lang.L.PickerNoMatch) + "\n")
	}

	b.WriteString("\n")
	b.WriteString(stylePickerHint.Render(lang.L.PickerNavHint) + "\n")
	return b.String()
}

// RunPicker tek seçimli interaktif menü açar. Seçilen item'ın Value'sunu döner.
// Kullanıcı q/ESC/Ctrl+C ile çıkarsa ("", nil) döner.
func RunPicker(title, subtitle string, items []PickerItem) (string, error) {
	m := newPickerModel(title, subtitle, items)
	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return "", err
	}
	fm := final.(pickerModel)
	if fm.quit {
		return "", nil
	}
	return fm.result, nil
}

// ── çoklu seçim ───────────────────────────────────────────────────────────────

type multiPickerModel struct {
	title    string
	subtitle string
	items    []PickerItem
	checked  []bool
	cursor   int
	result   []string
	quit     bool
	width    int
	height   int
}

func (m multiPickerModel) Init() tea.Cmd { return nil }

func (m multiPickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "Q", "esc":
			m.quit = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case " ": // boşluk ile işaretle
			m.checked[m.cursor] = !m.checked[m.cursor]
		case "a", "A": // hepsini seç/kaldır
			all := true
			for _, c := range m.checked {
				if !c {
					all = false
					break
				}
			}
			for i := range m.checked {
				m.checked[i] = !all
			}
		case "enter":
			for i, item := range m.items {
				if m.checked[i] {
					m.result = append(m.result, item.Value)
				}
			}
			if len(m.result) == 0 && len(m.items) > 0 {
				// Hiçbir şey seçilmemişse cursor'dakini seç
				m.result = []string{m.items[m.cursor].Value}
			}
			return m, tea.Quit
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			n := int(msg.String()[0]-'0') - 1
			if n >= 0 && n < len(m.items) {
				m.checked[n] = !m.checked[n]
			}
		}
	}
	return m, nil
}

func (m multiPickerModel) View() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(stylePickerTitle.Render(m.title) + "\n")
	if m.subtitle != "" {
		b.WriteString(stylePickerSub.Render(m.subtitle) + "\n")
	}
	w := 60
	if m.width > 10 {
		w = m.width - 4
	}
	b.WriteString(stylePickerDivider.Render("  "+strings.Repeat("─", w)) + "\n\n")

	for i, item := range m.items {
		icon := item.Icon
		if icon == "" {
			icon = " "
		}
		check := styleCheckOff.Render("[ ]")
		if m.checked[i] {
			check = styleCheckOn.Render("[✓]")
		}
		num := stylePickerDesc.Render(fmt.Sprintf("[%d]", i+1))
		desc := stylePickerDesc.Render(item.Desc)

		if i == m.cursor {
			cur := stylePickerCursor.Render("▶")
			lbl := stylePickerSelected.Render(item.Label)
			b.WriteString(fmt.Sprintf(" %s %s %s %s  %-28s  %s\n", cur, num, check, icon, lbl, desc))
		} else {
			lbl := stylePickerNormal.Render(item.Label)
			b.WriteString(fmt.Sprintf("    %s %s %s  %-28s  %s\n", num, check, icon, lbl, desc))
		}
	}

	b.WriteString("\n")
	b.WriteString(stylePickerHint.Render(lang.L.MultiPickerNavHint) + "\n")
	return b.String()
}

// runMultiPickerRaw verilen model ile multi-picker'ı çalıştırır.
func runMultiPickerRaw(m multiPickerModel) ([]string, error) {
	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return nil, err
	}
	fm := final.(multiPickerModel)
	if fm.quit {
		return nil, nil
	}
	return fm.result, nil
}

// RunMultiPicker çoklu seçimli interaktif menü açar. Seçilen Value'ların listesini döner.
func RunMultiPicker(title, subtitle string, items []PickerItem) ([]string, error) {
	return runMultiPickerRaw(newMultiPickerModel(title, subtitle, items))
}

func newMultiPickerModel(title, subtitle string, items []PickerItem) multiPickerModel {
	return multiPickerModel{
		title:    title,
		subtitle: subtitle,
		items:    items,
		checked:  make([]bool, len(items)),
		// Varsayılan: hiçbiri seçili değil — Enter ile sadece cursor seçilir
	}
}

// ── freeze seçici ─────────────────────────────────────────────────────────────

type freezePickerModel struct {
	title     string
	all       []PickerItem
	filtered  []PickerItem
	checked   map[string]bool // value → frozen (filter değişse de korunur)
	search    string
	cursor    int
	scroll    int
	confirmed bool
	quit      bool
	width     int
	height    int
}

func newFreezePickerModel(title string, items []PickerItem, preSelected map[string]bool) freezePickerModel {
	checked := make(map[string]bool, len(preSelected))
	for k, v := range preSelected {
		checked[k] = v
	}
	return freezePickerModel{
		title:    title,
		all:      items,
		filtered: items,
		checked:  checked,
	}
}

func (m *freezePickerModel) applySearch() {
	if m.search == "" {
		m.filtered = m.all
		m.cursor = 0
		return
	}
	q := strings.ToLower(m.search)
	m.filtered = nil
	for _, item := range m.all {
		if strings.Contains(strings.ToLower(item.Label), q) {
			m.filtered = append(m.filtered, item)
		}
	}
	m.cursor = 0
	m.scroll = 0
}

func (m freezePickerModel) Init() tea.Cmd { return nil }

func (m freezePickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q", "Q":
			m.quit = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.scroll {
					m.scroll = m.cursor
				}
			}
		case "down", "j":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
				rows := m.visibleRows()
				if m.cursor >= m.scroll+rows {
					m.scroll = m.cursor - rows + 1
				}
			}
		case " ":
			if len(m.filtered) > 0 {
				v := m.filtered[m.cursor].Value
				m.checked[v] = !m.checked[v]
			}
		case "a", "A":
			// tümünü seç veya kaldır (filtrelenmiş liste üzerinde)
			allChecked := true
			for _, item := range m.filtered {
				if !m.checked[item.Value] {
					allChecked = false
					break
				}
			}
			for _, item := range m.filtered {
				m.checked[item.Value] = !allChecked
			}
		case "enter":
			m.confirmed = true
			return m, tea.Quit
		case "backspace":
			r := []rune(m.search)
			if len(r) > 0 {
				m.search = string(r[:len(r)-1])
				m.applySearch()
			}
		default:
			s := msg.String()
			for strings.HasPrefix(s, "alt+") {
				s = strings.TrimPrefix(s, "alt+")
			}
			r := []rune(s)
			if len(r) == 1 && r[0] >= 0x20 && r[0] != 0x7F {
				m.search += s
				m.applySearch()
			}
		}
	}
	return m, nil
}

func (m freezePickerModel) visibleRows() int {
	h := m.height
	if h < 12 {
		h = 24
	}
	return h - 8
}

func (m freezePickerModel) View() string {
	var b strings.Builder
	w := 60
	if m.width > 10 {
		w = m.width - 4
	}

	b.WriteString("\n")
	b.WriteString(stylePickerTitle.Render("🧊 "+m.title) + "\n")

	// Arama
	if m.search != "" {
		b.WriteString(stylePickerSelected.Render(fmt.Sprintf("  🔍 %s", m.search)) +
			stylePickerDesc.Render(fmt.Sprintf("  (%d match)", len(m.filtered))) + "\n")
	} else {
		b.WriteString(stylePickerHint.Render("  Type to filter  |  Space = freeze/unfreeze  |  a = toggle all  |  Enter = save  |  q = cancel") + "\n")
	}
	b.WriteString(stylePickerDivider.Render("  "+strings.Repeat("─", w)) + "\n")

	rows := m.visibleRows()
	end := m.scroll + rows
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	for i := m.scroll; i < end; i++ {
		item := m.filtered[i]
		frozen := m.checked[item.Value]
		check := styleCheckOff.Render("[ ]")
		lbl := stylePickerNormal.Render(item.Label)
		if frozen {
			check = styleCheckOn.Render("[❄]")
			lbl = stylePickerSelected.Render(item.Label)
		}
		if i == m.cursor {
			cur := stylePickerCursor.Render("▶")
			b.WriteString(fmt.Sprintf(" %s %s  %s\n", cur, check, lbl))
		} else {
			b.WriteString(fmt.Sprintf("    %s  %s\n", check, lbl))
		}
	}

	if len(m.filtered) == 0 {
		b.WriteString(stylePickerHint.Render("  (no match)") + "\n")
	}

	// Footer
	frozenCount := 0
	for _, v := range m.checked {
		if v {
			frozenCount++
		}
	}
	b.WriteString(stylePickerDivider.Render("  "+strings.Repeat("─", w)) + "\n")
	b.WriteString(stylePickerHint.Render(fmt.Sprintf("  ❄ %d frozen  ·  %d/%d shown", frozenCount, len(m.filtered), len(m.all))) + "\n")
	return b.String()
}

// RunFreezeList dosya listesini gösterir, kullanıcı freeze/unfreeze işaretler.
// preSelected: daha önce freeze edilmiş dosyalar.
// Döner: seçilen (frozen) value listesi, nil = iptal.
func RunFreezeList(title string, items []PickerItem, preSelected map[string]bool) ([]string, error) {
	m := newFreezePickerModel(title, items, preSelected)
	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return nil, err
	}
	fm := final.(freezePickerModel)
	if fm.quit || !fm.confirmed {
		return nil, nil
	}
	result := []string{} // nil değil — hiç seçilmese de "tüm freeze'leri kaldır" olarak kaydedilir
	for _, item := range fm.all {
		if fm.checked[item.Value] {
			result = append(result, item.Value)
		}
	}
	return result, nil
}

// ── onay ekranı ───────────────────────────────────────────────────────────────

// RunConfirm evet/hayır seçimi sunar. true = onaylandı.
func RunConfirm(title, subtitle string) (bool, error) {
	items := []PickerItem{
		{Icon: "✓", Label: lang.L.ConfirmYes, Value: "yes"},
		{Icon: "✕", Label: lang.L.ConfirmNo, Value: "no"},
	}
	val, err := RunPicker(title, subtitle, items)
	if err != nil {
		return false, err
	}
	return val == "yes", nil
}

package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	items    []PickerItem
	cursor   int
	result   string
	quit     bool
	width    int
	height   int
}

func (m pickerModel) Init() tea.Cmd { return nil }

func (m pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		case "enter", "right", "l":
			m.result = m.items[m.cursor].Value
			return m, tea.Quit
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			n := int(msg.String()[0]-'0') - 1
			if n >= 0 && n < len(m.items) {
				m.result = m.items[n].Value
				return m, tea.Quit
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
	b.WriteString(stylePickerDivider.Render("  "+strings.Repeat("─", w)) + "\n\n")

	for i, item := range m.items {
		icon := item.Icon
		if icon == "" {
			icon = " "
		}
		if i == m.cursor {
			cur := stylePickerCursor.Render("▶")
			lbl := stylePickerSelected.Render(item.Label)
			desc := stylePickerDesc.Render(item.Desc)
			num := stylePickerDesc.Render(fmt.Sprintf("[%d]", i+1))
			b.WriteString(fmt.Sprintf(" %s %s %s  %-30s  %s\n", cur, num, icon, lbl, desc))
		} else {
			lbl := stylePickerNormal.Render(item.Label)
			desc := stylePickerDesc.Render(item.Desc)
			num := stylePickerDesc.Render(fmt.Sprintf("[%d]", i+1))
			b.WriteString(fmt.Sprintf("    %s %s  %-30s  %s\n", num, icon, lbl, desc))
		}
	}

	b.WriteString("\n")
	b.WriteString(stylePickerHint.Render("↑↓ gezin   Enter seç   [1-9] hızlı seç   q iptal") + "\n")
	return b.String()
}

// RunPicker tek seçimli interaktif menü açar. Seçilen item'ın Value'sunu döner.
// Kullanıcı q/ESC/Ctrl+C ile çıkarsa ("", nil) döner.
func RunPicker(title, subtitle string, items []PickerItem) (string, error) {
	m := pickerModel{title: title, subtitle: subtitle, items: items}
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
	b.WriteString(stylePickerHint.Render("↑↓ gezin   Space işaretle   a hepsini   Enter onayla   q iptal") + "\n")
	return b.String()
}

// RunMultiPicker çoklu seçimli interaktif menü açar. Seçilen Value'ların listesini döner.
func RunMultiPicker(title, subtitle string, items []PickerItem) ([]string, error) {
	m := multiPickerModel{
		title:    title,
		subtitle: subtitle,
		items:    items,
		checked:  make([]bool, len(items)),
	}
	// Varsayılan: hepsi seçili
	for i := range m.checked {
		m.checked[i] = true
	}
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

// ── onay ekranı ───────────────────────────────────────────────────────────────

// RunConfirm evet/hayır seçimi sunar. true = onaylandı.
func RunConfirm(title, subtitle string) (bool, error) {
	items := []PickerItem{
		{Icon: "✓", Label: "Evet", Value: "yes"},
		{Icon: "✕", Label: "Hayır, iptal", Value: "no"},
	}
	val, err := RunPicker(title, subtitle, items)
	if err != nil {
		return false, err
	}
	return val == "yes", nil
}

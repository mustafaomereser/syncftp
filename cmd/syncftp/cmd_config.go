package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"syncftp/internal/config"
)

func init() {
	rootCmd.AddCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Server ayarlarını yönet (ekle, düzenle, sil)",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := os.Getwd()
		return runServerMgr(dir)
	},
}

// ── yardımcı ─────────────────────────────────────────────────────────────────

func boolIcon(b bool) string {
	if b {
		return "✓"
	}
	return "○"
}

func listDisplay(l []string) string {
	if len(l) == 0 {
		return "(boş)"
	}
	return strings.Join(l, ", ")
}

func parseList(s string) []string {
	if strings.TrimSpace(s) == "" {
		return []string{}
	}
	var out []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// ── stiller ──────────────────────────────────────────────────────────────────

var (
	cfgTitle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5")).Padding(0, 2)
	cfgCursor   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	cfgEnabled  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	cfgDisabled = lipgloss.NewStyle().Faint(true)
	cfgHint     = lipgloss.NewStyle().Faint(true)
	cfgWarn     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("1"))
	cfgAdd      = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	cfgDiv      = lipgloss.NewStyle().Faint(true)
	cfgLabel    = lipgloss.NewStyle().Width(20).Foreground(lipgloss.Color("7"))
	cfgValue    = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	cfgEditing  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
	cfgBrowse   = lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("6"))
	cfgGlobal   = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	cfgSelected = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
)

// ══════════════════════════════════════════════════════════════════════════════
// Server Yönetici TUI
// ══════════════════════════════════════════════════════════════════════════════

// cursor: 0 = Global Ayarlar, 1..N = serverler, N+1 = + Ekle
type serverMgrModel struct {
	servers    []config.Server
	cursor     int
	confirming bool
	action     string // "edit-global","edit","add","delete","toggle"
	actionIdx  int
	quit       bool
	width      int
	height     int
}

func newServerMgrModel(servers []config.Server) serverMgrModel {
	cur := 0
	if len(servers) > 0 {
		cur = 1
	}
	return serverMgrModel{servers: servers, cursor: cur}
}

func (m serverMgrModel) Init() tea.Cmd { return nil }

func (m serverMgrModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		if m.confirming {
			switch msg.String() {
			case "y", "Y", "enter":
				m.action = "delete"
				m.actionIdx = m.cursor - 1
				return m, tea.Quit
			default:
				m.confirming = false
			}
			return m, nil
		}
		total := len(m.servers) + 2
		switch msg.String() {
		case "ctrl+c", "q", "Q", "esc":
			m.quit = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < total-1 {
				m.cursor++
			}
		case "e", "E", "enter":
			switch {
			case m.cursor == 0:
				m.action = "edit-global"
				return m, tea.Quit
			case m.cursor <= len(m.servers):
				m.action = "edit"
				m.actionIdx = m.cursor - 1
				return m, tea.Quit
			default:
				m.action = "add"
				return m, tea.Quit
			}
		case "n", "N":
			m.action = "add"
			return m, tea.Quit
		case "d", "D":
			if m.cursor >= 1 && m.cursor <= len(m.servers) {
				m.confirming = true
			}
		case " ":
			if m.cursor >= 1 && m.cursor <= len(m.servers) {
				m.action = "toggle"
				m.actionIdx = m.cursor - 1
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m serverMgrModel) View() string {
	w := m.width
	if w < 60 {
		w = 80
	}
	div := cfgDiv.Render("  " + strings.Repeat("─", w-4))
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(cfgTitle.Render("⚙  Server Ayarları") + "\n")
	b.WriteString(cfgHint.Render("  ↑↓ gezin  |  Enter/e = düzenle  |  Space = aç/kapat  |  d = sil  |  n = yeni  |  q = çık") + "\n")
	b.WriteString(div + "\n\n")

	if m.confirming {
		name := m.servers[m.cursor-1].Name
		b.WriteString(cfgWarn.Render(fmt.Sprintf("  [%s] silinsin mi?  y = evet, diğer = iptal", name)) + "\n")
		return b.String()
	}

	// Global Ayarlar
	if m.cursor == 0 {
		b.WriteString(cfgCursor.Render("▶ ") + cfgGlobal.Bold(true).Render("⚙ Global Ayarlar") +
			cfgHint.Render("  (protect, include, exclude, ignore_files)") + "\n")
	} else {
		b.WriteString("  " + cfgGlobal.Render("⚙ Global Ayarlar") +
			cfgHint.Render("  (protect, include, exclude, ignore_files)") + "\n")
	}

	// Serverler
	for i, srv := range m.servers {
		pos := i + 1
		icon := cfgEnabled.Render("✓")
		nameS := cfgEnabled.Render(srv.Name)
		if !srv.Enabled {
			icon = cfgDisabled.Render("○")
			nameS = cfgDisabled.Render(srv.Name + " (devre dışı)")
		}
		host := cfgDisabled.Render(fmt.Sprintf("%s:%d%s", srv.Host, srv.Port, srv.RemotePath))
		info := cfgDisabled.Render(fmt.Sprintf("conn:%d  retry:%d", srv.MaxConnections, srv.MaxRetries))
		if pos == m.cursor {
			b.WriteString(fmt.Sprintf(" %s%s  %-22s  %-38s  %s\n",
				cfgCursor.Render("▶ "), icon, nameS, host, info))
		} else {
			b.WriteString(fmt.Sprintf("    %s  %-22s  %-38s  %s\n", icon, nameS, host, info))
		}
	}

	// + Ekle
	addPos := len(m.servers) + 1
	if m.cursor == addPos {
		b.WriteString(cfgCursor.Render("▶ ") + cfgAdd.Render("+ Yeni server ekle") + "\n")
	} else {
		b.WriteString("    " + cfgAdd.Render("+ Yeni server ekle") + "\n")
	}

	b.WriteString("\n" + div + "\n")
	b.WriteString(cfgHint.Render(fmt.Sprintf("  %d server", len(m.servers))) + "\n")
	return b.String()
}

// ── Ana döngü ─────────────────────────────────────────────────────────────────

func runServerMgr(configDir string) error {
	for {
		cfg, err := config.Load(configDir)
		if err != nil {
			return err
		}
		projectDir := configDir

		p := tea.NewProgram(newServerMgrModel(cfg.Servers), tea.WithAltScreen())
		raw, err := p.Run()
		if err != nil {
			return err
		}
		fm := raw.(serverMgrModel)
		if fm.quit {
			return nil
		}

		switch fm.action {
		case "edit-global":
			if runGlobalEdit(&cfg.Project, &cfg.Sync, projectDir) {
				if e := config.Save(configDir, cfg); e != nil {
					fmt.Printf("Kayıt hatası: %v\n", e)
				}
			}

		case "edit":
			srv := cfg.Servers[fm.actionIdx]
			if runServerEdit(&srv, projectDir, false) {
				cfg.Servers[fm.actionIdx] = srv
				if e := config.Save(configDir, cfg); e != nil {
					fmt.Printf("Kayıt hatası: %v\n", e)
				} else {
					fmt.Printf("✓ [%s] güncellendi\n", srv.Name)
				}
			}

		case "add":
			srv := config.Server{Port: 21, Passive: true, Enabled: true, MaxConnections: 1, MaxRetries: 2}
			if runServerEdit(&srv, projectDir, true) {
				cfg.Servers = append(cfg.Servers, srv)
				if e := config.Save(configDir, cfg); e != nil {
					fmt.Printf("Kayıt hatası: %v\n", e)
				} else {
					fmt.Printf("✓ [%s] eklendi\n", srv.Name)
				}
			}

		case "delete":
			name := cfg.Servers[fm.actionIdx].Name
			cfg.Servers = append(cfg.Servers[:fm.actionIdx], cfg.Servers[fm.actionIdx+1:]...)
			if e := config.Save(configDir, cfg); e != nil {
				fmt.Printf("Kayıt hatası: %v\n", e)
			} else {
				fmt.Printf("✓ [%s] silindi\n", name)
			}

		case "toggle":
			cfg.Servers[fm.actionIdx].Enabled = !cfg.Servers[fm.actionIdx].Enabled
			s := "devre dışı"
			if cfg.Servers[fm.actionIdx].Enabled {
				s = "aktif"
			}
			if e := config.Save(configDir, cfg); e != nil {
				fmt.Printf("Kayıt hatası: %v\n", e)
			} else {
				fmt.Printf("✓ [%s] %s\n", cfg.Servers[fm.actionIdx].Name, s)
			}
		}
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// Server Düzenleme TUI (field navigasyonu)
// ══════════════════════════════════════════════════════════════════════════════

type srvField struct {
	label    string
	kind     string // "text","pass","int","bool","list"
	browseID string // list alanları için: "include" veya "exclude"
}

var serverFields = []srvField{
	{"Ad", "text", ""},
	{"Host", "text", ""},
	{"Port", "int", ""},
	{"Kullanıcı", "text", ""},
	{"Şifre", "pass", ""},
	{"Uzak dizin", "text", ""},
	{"Yerel dizin", "localdir", ""},
	{"Aktif", "bool", ""},
	{"Passive mode", "bool", ""},
	{"EPSV devre dışı", "bool", ""},
	{"NAT workaround", "bool", ""},
	{"Max bağlantı", "int", ""},
	{"Max retry", "int", ""},
	{"Include", "list", "include"},
	{"Exclude", "list", "exclude"},
	{"Protect", "list", "protect"},
}

type serverEditModel struct {
	srv        config.Server
	isNew      bool
	projectDir string
	cursor     int
	editing    bool
	editBuf    string
	action     string // "save","cancel","browse"
	browseFor  string
	width      int
	height     int
}

func newServerEditModel(srv config.Server, isNew bool, projectDir string, cursor int) serverEditModel {
	return serverEditModel{srv: srv, isNew: isNew, projectDir: projectDir, cursor: cursor}
}

func (m *serverEditModel) getDisplayValue(i int) string {
	switch i {
	case 0:
		return m.srv.Name
	case 1:
		return m.srv.Host
	case 2:
		return strconv.Itoa(m.srv.Port)
	case 3:
		return m.srv.User
	case 4:
		if m.srv.Password != "" {
			return "****"
		}
		return "(boş)"
	case 5:
		return m.srv.RemotePath
	case 6:
		if m.srv.LocalPath != "" {
			return m.srv.LocalPath
		}
		return "(project default)"
	case 7:
		return boolIcon(m.srv.Enabled)
	case 8:
		return boolIcon(m.srv.Passive)
	case 9:
		return boolIcon(m.srv.DisableEPSV)
	case 10:
		return boolIcon(m.srv.NATWorkaround)
	case 11:
		return strconv.Itoa(m.srv.MaxConnections)
	case 12:
		return strconv.Itoa(m.srv.MaxRetries)
	case 13:
		return listDisplay(m.srv.Include)
	case 14:
		return listDisplay(m.srv.Exclude)
	case 15:
		return listDisplay(m.srv.Protect)
	}
	return ""
}

func (m *serverEditModel) getEditableValue(i int) string {
	switch i {
	case 0:
		return m.srv.Name
	case 1:
		return m.srv.Host
	case 2:
		return strconv.Itoa(m.srv.Port)
	case 3:
		return m.srv.User
	case 4:
		return m.srv.Password
	case 5:
		return m.srv.RemotePath
	case 6:
		return m.srv.LocalPath
	case 11:
		return strconv.Itoa(m.srv.MaxConnections)
	case 12:
		return strconv.Itoa(m.srv.MaxRetries)
	case 13:
		return strings.Join(m.srv.Include, ", ")
	case 14:
		return strings.Join(m.srv.Exclude, ", ")
	case 15:
		return strings.Join(m.srv.Protect, ", ")
	}
	return ""
}

func (m *serverEditModel) applyValue(i int, val string) {
	switch i {
	case 0:
		m.srv.Name = val
	case 1:
		m.srv.Host = val
	case 2:
		if n, e := strconv.Atoi(val); e == nil {
			m.srv.Port = n
		}
	case 3:
		m.srv.User = val
	case 4:
		m.srv.Password = val
	case 5:
		m.srv.RemotePath = val
	case 6:
		m.srv.LocalPath = val
	case 11:
		if n, e := strconv.Atoi(val); e == nil {
			m.srv.MaxConnections = n
		}
	case 12:
		if n, e := strconv.Atoi(val); e == nil {
			m.srv.MaxRetries = n
		}
	case 13:
		m.srv.Include = parseList(val)
	case 14:
		m.srv.Exclude = parseList(val)
	case 15:
		m.srv.Protect = parseList(val)
	}
}

func (m *serverEditModel) toggleBool(i int) {
	switch i {
	case 7:
		m.srv.Enabled = !m.srv.Enabled
	case 8:
		m.srv.Passive = !m.srv.Passive
	case 9:
		m.srv.DisableEPSV = !m.srv.DisableEPSV
	case 10:
		m.srv.NATWorkaround = !m.srv.NATWorkaround
	}
}

func (m serverEditModel) Init() tea.Cmd { return nil }

func (m serverEditModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		if m.editing {
			return m.handleEditKey(msg)
		}
		return m.handleNavKey(msg)
	}
	return m, nil
}

func (m serverEditModel) handleEditKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.applyValue(m.cursor, m.editBuf)
		m.editing = false
	case "esc":
		m.editing = false
	case "backspace":
		r := []rune(m.editBuf)
		if len(r) > 0 {
			m.editBuf = string(r[:len(r)-1])
		}
	default:
		s := msg.String()
		for strings.HasPrefix(s, "alt+") {
			s = strings.TrimPrefix(s, "alt+")
		}
		r := []rune(s)
		if len(r) == 1 && r[0] >= 0x20 && r[0] != 0x7F {
			m.editBuf += s
		}
	}
	return m, nil
}

func (m serverEditModel) handleNavKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	f := serverFields[m.cursor]
	switch msg.String() {
	case "ctrl+c", "q", "Q", "esc":
		m.action = "cancel"
		return m, tea.Quit
	case "s", "S":
		m.action = "save"
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(serverFields)-1 {
			m.cursor++
		}
	case "enter":
		switch f.kind {
		case "bool":
			m.toggleBool(m.cursor)
		default:
			m.editBuf = m.getEditableValue(m.cursor)
			m.editing = true
		}
	case " ":
		if f.kind == "bool" {
			m.toggleBool(m.cursor)
		}
	case "b", "B":
		if f.kind == "localdir" && m.projectDir != "" {
			m.action = "browse-localpath"
			return m, tea.Quit
		}
		if f.kind == "list" && m.projectDir != "" {
			m.browseFor = f.browseID
			m.action = "browse"
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m serverEditModel) View() string {
	w := m.width
	if w < 60 {
		w = 80
	}
	div := cfgDiv.Render("  " + strings.Repeat("─", w-4))

	title := "Server Düzenle"
	if m.isNew {
		title = "Yeni Server"
	}

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(cfgTitle.Render("⚙  "+title) + "\n")
	b.WriteString(cfgHint.Render("  ↑↓ gezin  |  Enter = düzenle/toggle  |  Space = bool toggle  |  b = yerel dosya gözat  |  s = kaydet  |  q = iptal") + "\n")
	b.WriteString(div + "\n\n")

	for i, f := range serverFields {
		label := cfgLabel.Render(f.label)
		var valStr string

		if i == m.cursor && m.editing {
			valStr = cfgEditing.Render(m.editBuf + "█")
		} else {
			raw := m.getDisplayValue(i)
			switch f.kind {
			case "bool":
				if raw == "✓" {
					valStr = cfgEnabled.Render(raw)
				} else {
					valStr = cfgDisabled.Render(raw)
				}
			case "localdir":
				valStr = cfgValue.Render(raw)
				if m.projectDir != "" {
					valStr += "  " + cfgBrowse.Render("[b=gözat]")
				}
			case "list":
				valStr = cfgValue.Render(raw)
				if m.projectDir != "" {
					valStr += "  " + cfgBrowse.Render("[b=gözat]")
				}
			default:
				valStr = cfgValue.Render(raw)
			}
		}

		if i == m.cursor && !m.editing {
			b.WriteString(cfgCursor.Render("▶ ") + label + "  " + valStr + "\n")
		} else {
			b.WriteString("  " + label + "  " + valStr + "\n")
		}
	}

	b.WriteString("\n" + div + "\n")
	b.WriteString(cfgHint.Render("  s = kaydet  |  q/Esc = iptal") + "\n")
	return b.String()
}

// runServerEdit — döngü: field edit + browse arası geçiş
func runServerEdit(srv *config.Server, projectDir string, isNew bool) (saved bool) {
	cursor := 0
	for {
		m := newServerEditModel(*srv, isNew, projectDir, cursor)
		p := tea.NewProgram(m, tea.WithAltScreen())
		raw, err := p.Run()
		if err != nil {
			return false
		}
		fm := raw.(serverEditModel)
		cursor = fm.cursor

		switch fm.action {
		case "save":
			*srv = fm.srv
			return true
		case "cancel":
			return false
		case "browse":
			*srv = fm.srv // mevcut değişiklikleri sakla
			localRoot := projectDir
			if srv.LocalPath != "" {
				candidate := filepath.Join(projectDir, filepath.FromSlash(srv.LocalPath))
				if fi, err := os.Stat(candidate); err == nil && fi.IsDir() {
					localRoot = candidate
				}
			}
			var current []string
			switch fm.browseFor {
			case "include":
				current = srv.Include
				cursor = 13
			case "exclude":
				current = srv.Exclude
				cursor = 14
			case "protect":
				current = srv.Protect
				cursor = 15
			}
			selected, ok := runLocalBrowser(localRoot, current)
			if ok {
				switch fm.browseFor {
				case "include":
					srv.Include = selected
				case "exclude":
					srv.Exclude = selected
				case "protect":
					srv.Protect = selected
				}
			}
		case "browse-localpath":
			*srv = fm.srv
			cursor = 6
			selected, ok := runLocalDirPicker(projectDir)
			if ok {
				srv.LocalPath = selected
			}
		}
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// Global Ayarlar TUI (Sync — protect / include / exclude / ignore_files)
// ══════════════════════════════════════════════════════════════════════════════

type globalField struct {
	label    string
	kind     string // "list","textlist"
	browseID string
}

var globalFields = []globalField{
	{"Varsayılan dizin", "localdir", ""},
	{"Protect", "list", "protect"},
	{"Include (global)", "list", "include"},
	{"Exclude (global)", "list", "exclude"},
	{"Ignore files", "list", "ignore"},
}

type globalEditModel struct {
	project    config.Project
	sync       config.Sync
	projectDir string
	cursor     int
	editing    bool
	editBuf    string
	action     string // "save","cancel","browse","browse-localpath"
	browseFor  string
	width      int
	height     int
}

func newGlobalEditModel(proj config.Project, s config.Sync, projectDir string, cursor int) globalEditModel {
	return globalEditModel{project: proj, sync: s, projectDir: projectDir, cursor: cursor}
}

func (m *globalEditModel) getDisplayValue(i int) string {
	switch i {
	case 0:
		if m.project.DefaultPath != "" {
			return m.project.DefaultPath
		}
		return "(boş = çalışma dizini)"
	case 1:
		return listDisplay(m.sync.Protect)
	case 2:
		return listDisplay(m.sync.Include)
	case 3:
		return listDisplay(m.sync.Exclude)
	case 4:
		if m.sync.IgnoreFiles == nil {
			return "(varsayılan: .gitignore + syncftp.ignore)"
		}
		if len(m.sync.IgnoreFiles) == 0 {
			return "(hiçbiri — ignore kullanılmaz)"
		}
		return strings.Join(m.sync.IgnoreFiles, ", ")
	}
	return ""
}

func (m *globalEditModel) getEditableValue(i int) string {
	switch i {
	case 0:
		return m.project.DefaultPath
	case 1:
		return strings.Join(m.sync.Protect, ", ")
	case 2:
		return strings.Join(m.sync.Include, ", ")
	case 3:
		return strings.Join(m.sync.Exclude, ", ")
	case 4:
		return strings.Join(m.sync.IgnoreFiles, ", ")
	}
	return ""
}

func (m *globalEditModel) applyValue(i int, val string) {
	switch i {
	case 0:
		m.project.DefaultPath = val
	case 1:
		m.sync.Protect = parseList(val)
	case 2:
		m.sync.Include = parseList(val)
	case 3:
		m.sync.Exclude = parseList(val)
	case 4:
		m.sync.IgnoreFiles = parseList(val)
	}
}

func (m globalEditModel) Init() tea.Cmd { return nil }

func (m globalEditModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		if m.editing {
			switch msg.String() {
			case "enter":
				m.applyValue(m.cursor, m.editBuf)
				m.editing = false
			case "esc":
				m.editing = false
			case "backspace":
				r := []rune(m.editBuf)
				if len(r) > 0 {
					m.editBuf = string(r[:len(r)-1])
				}
			default:
				s := msg.String()
				for strings.HasPrefix(s, "alt+") {
					s = strings.TrimPrefix(s, "alt+")
				}
				r := []rune(s)
				if len(r) == 1 && r[0] >= 0x20 && r[0] != 0x7F {
					m.editBuf += s
				}
			}
			return m, nil
		}
		switch msg.String() {
		case "ctrl+c", "q", "Q", "esc":
			m.action = "cancel"
			return m, tea.Quit
		case "s", "S":
			m.action = "save"
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(globalFields)-1 {
				m.cursor++
			}
		case "enter":
			m.editBuf = m.getEditableValue(m.cursor)
			m.editing = true
		case "b", "B":
			f := globalFields[m.cursor]
			if f.kind == "localdir" && m.projectDir != "" {
				m.action = "browse-localpath"
				return m, tea.Quit
			}
			if f.browseID != "" {
				m.browseFor = f.browseID
				m.action = "browse"
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m globalEditModel) View() string {
	w := m.width
	if w < 60 {
		w = 80
	}
	div := cfgDiv.Render("  " + strings.Repeat("─", w-4))
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(cfgTitle.Render("⚙  Global Ayarlar") + "\n")
	b.WriteString(cfgHint.Render("  ↑↓ gezin  |  Enter = düzenle/aç  |  b = yerel dosya gözat  |  s = kaydet  |  q = iptal") + "\n")
	b.WriteString(div + "\n\n")

	for i, f := range globalFields {
		label := cfgLabel.Render(f.label)
		var valStr string
		if i == m.cursor && m.editing {
			valStr = cfgEditing.Render(m.editBuf + "█")
		} else {
			raw := m.getDisplayValue(i)
			valStr = cfgValue.Render(raw)
			if f.kind == "localdir" && m.projectDir != "" {
				valStr += "  " + cfgBrowse.Render("[b=gözat]")
			} else if f.browseID != "" && m.projectDir != "" {
				valStr += "  " + cfgBrowse.Render("[b=gözat]")
			}
		}
		if i == m.cursor && !m.editing {
			b.WriteString(cfgCursor.Render("▶ ") + label + "  " + valStr + "\n")
		} else {
			b.WriteString("  " + label + "  " + valStr + "\n")
		}
	}

	b.WriteString("\n" + div + "\n")
	b.WriteString(cfgHint.Render("  s = kaydet  |  q/Esc = iptal") + "\n")
	b.WriteString(cfgHint.Render("  ignore_files: b=gözat ile seç; hiçbiri seçmeden kaydet = ignore yok; alan boşsa = varsayılan (.gitignore + syncftp.ignore)") + "\n")
	return b.String()
}

func runGlobalEdit(proj *config.Project, s *config.Sync, projectDir string) bool {
	cursor := 0
	changed := false
	for {
		m := newGlobalEditModel(*proj, *s, projectDir, cursor)
		p := tea.NewProgram(m, tea.WithAltScreen())
		raw, err := p.Run()
		if err != nil {
			return changed
		}
		fm := raw.(globalEditModel)
		cursor = fm.cursor

		switch fm.action {
		case "save":
			*proj = fm.project
			*s = fm.sync
			return true
		case "cancel":
			return changed
		case "browse-localpath":
			*proj = fm.project
			*s = fm.sync
			selected, ok := runLocalDirPicker(projectDir)
			if ok {
				proj.DefaultPath = selected
				changed = true
			}
			cursor = 0
		case "browse":
			*s = fm.sync
			if fm.browseFor == "ignore" {
				cursor = 4
				result, ok := runIgnoreFilePicker(s.IgnoreFiles)
				if ok {
					s.IgnoreFiles = result
				}
				break
			}
			localRoot := projectDir
			if fm.project.DefaultPath != "" {
				candidate := filepath.Join(projectDir, filepath.FromSlash(fm.project.DefaultPath))
				if fi, err := os.Stat(candidate); err == nil && fi.IsDir() {
					localRoot = candidate
				}
			}
			var current []string
			switch fm.browseFor {
			case "protect":
				current = s.Protect
				cursor = 1
			case "include":
				current = s.Include
				cursor = 2
			case "exclude":
				current = s.Exclude
				cursor = 3
			}
			selected, ok := runLocalBrowser(localRoot, current)
			if ok {
				switch fm.browseFor {
				case "protect":
					s.Protect = selected
				case "include":
					s.Include = selected
				case "exclude":
					s.Exclude = selected
				}
			}
		}
	}
}

// runLocalDirPicker yerel dosya sisteminde tek bir dizin seçtirir.
// Döner: seçilen dizinin projectDir'e göreceli yolu.
func runLocalDirPicker(projectDir string) (string, bool) {
	m := newLocalBrowserModel(projectDir, nil)
	m.dirPickMode = true
	p := tea.NewProgram(m, tea.WithAltScreen())
	raw, err := p.Run()
	if err != nil {
		return "", false
	}
	fm := raw.(localBrowserModel)
	if fm.quit || !fm.done {
		return "", false
	}
	// Seçilen dizin = cwd (not relative)
	rel, err := filepath.Rel(projectDir, fm.cwd)
	if err != nil {
		return fm.cwd, true
	}
	return filepath.ToSlash(rel), true
}

// ══════════════════════════════════════════════════════════════════════════════
// Yerel Dosya Sistemi Browser
// ══════════════════════════════════════════════════════════════════════════════

type localEntry struct {
	name  string
	rel   string // projectRoot'a göreceli
	isDir bool
}

type localBrowserModel struct {
	root    string
	cwd     string
	entries []localEntry
	marked  map[string]bool // rel → işaretli
	// glob kalıpları gibi dosya sisteminde olmayan eski öğeler
	nonFsItems   []string
	cursor       int
	done         bool
	quit         bool
	dirPickMode  bool // true = sadece dizin seç (s/Enter ile cwd döner)
	addingCustom bool // n tuşuyla açılan özel yol giriş modu
	customBuf    string
	width        int
	height       int
}

func newLocalBrowserModel(root string, preSelected []string) localBrowserModel {
	m := localBrowserModel{
		root:   root,
		cwd:    root,
		marked: make(map[string]bool),
	}
	for _, item := range preSelected {
		fullPath := filepath.Join(root, filepath.FromSlash(item))
		if _, err := os.Stat(fullPath); err == nil {
			m.marked[filepath.ToSlash(item)] = true
		} else {
			// glob veya olmayan yol — dosya sisteminde yok, sakla
			m.nonFsItems = append(m.nonFsItems, item)
		}
	}
	m.entries = m.loadEntries()
	return m
}

func (m localBrowserModel) loadEntries() []localEntry {
	dirEntries, err := os.ReadDir(m.cwd)
	if err != nil {
		return nil
	}
	rel, _ := filepath.Rel(m.root, m.cwd)

	// .syncftp ve gizli klasörleri atla (. ile başlayanlar hariç seçilebilir dosyalar)
	var dirs, files []localEntry
	for _, e := range dirEntries {
		name := e.Name()
		if name == ".syncftp" {
			continue
		}
		var entryRel string
		if rel == "." {
			entryRel = filepath.ToSlash(name)
		} else {
			entryRel = filepath.ToSlash(filepath.Join(rel, name))
		}
		le := localEntry{name: name, rel: entryRel, isDir: e.IsDir()}
		if e.IsDir() {
			dirs = append(dirs, le)
		} else {
			files = append(files, le)
		}
	}
	sort.Slice(dirs, func(i, j int) bool { return strings.ToLower(dirs[i].name) < strings.ToLower(dirs[j].name) })
	sort.Slice(files, func(i, j int) bool { return strings.ToLower(files[i].name) < strings.ToLower(files[j].name) })
	return append(dirs, files...)
}

func (m localBrowserModel) relCwd() string {
	rel, _ := filepath.Rel(m.root, m.cwd)
	if rel == "." {
		return "/"
	}
	return "/" + filepath.ToSlash(rel)
}

func (m localBrowserModel) Init() tea.Cmd { return nil }

func (m localBrowserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		// Özel yol giriş modu aktifse tüm tuşları yakala
		if m.addingCustom {
			switch msg.String() {
			case "esc":
				m.addingCustom = false
				m.customBuf = ""
			case "enter":
				p := strings.TrimSpace(m.customBuf)
				if p != "" {
					found := false
					for _, it := range m.nonFsItems {
						if it == p {
							found = true
							break
						}
					}
					if !found {
						m.nonFsItems = append(m.nonFsItems, p)
					}
				}
				m.addingCustom = false
				m.customBuf = ""
			case "backspace":
				r := []rune(m.customBuf)
				if len(r) > 0 {
					m.customBuf = string(r[:len(r)-1])
				}
			default:
				s := msg.String()
				for strings.HasPrefix(s, "alt+") {
					s = strings.TrimPrefix(s, "alt+")
				}
				r := []rune(s)
				if len(r) == 1 && r[0] >= 0x20 && r[0] != 0x7F {
					m.customBuf += s
				}
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "q", "Q":
			m.quit = true
			return m, tea.Quit
		case "esc", "left", "h":
			if m.cwd != m.root {
				m.cwd = filepath.Dir(m.cwd)
				m.entries = m.loadEntries()
				m.cursor = 0
			} else if m.dirPickMode {
				m.quit = true
				return m, tea.Quit
			}
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.entries)-1 {
				m.cursor++
			}
		case "enter", "right", "l":
			if m.dirPickMode {
				// dirPickMode: Enter = bu dizini seç (içine girme); dizin üzerindeyse gir
				if len(m.entries) > 0 {
					e := m.entries[m.cursor]
					if e.isDir {
						m.cwd = filepath.Join(m.cwd, e.name)
						m.entries = m.loadEntries()
						m.cursor = 0
						break
					}
				}
				// Dosya veya boş = geçerli cwd'yi seç
				m.done = true
				return m, tea.Quit
			}
			if len(m.entries) == 0 {
				break
			}
			e := m.entries[m.cursor]
			if e.isDir {
				m.cwd = filepath.Join(m.cwd, e.name)
				m.entries = m.loadEntries()
				m.cursor = 0
			} else {
				m.marked[e.rel] = !m.marked[e.rel]
			}
		case " ":
			if m.dirPickMode {
				// dirPickMode'da space = bu dizini seç
				m.done = true
				return m, tea.Quit
			}
			if len(m.entries) > 0 {
				e := m.entries[m.cursor]
				m.marked[e.rel] = !m.marked[e.rel]
				if m.cursor < len(m.entries)-1 {
					m.cursor++
				}
			}
		case "a", "A":
			if !m.dirPickMode {
				allMarked := true
				for _, e := range m.entries {
					if !m.marked[e.rel] {
						allMarked = false
						break
					}
				}
				for _, e := range m.entries {
					m.marked[e.rel] = !allMarked
				}
			}
		case "n", "N":
			if !m.dirPickMode {
				m.addingCustom = true
				m.customBuf = ""
			}
		case "s", "S", "d", "D":
			m.done = true
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

func (m localBrowserModel) View() string {
	w := m.width
	if w < 60 {
		w = 80
	}
	h := m.height
	if h < 10 {
		h = 24
	}
	div := cfgDiv.Render(strings.Repeat("─", w))

	markedCount := 0
	for _, v := range m.marked {
		if v {
			markedCount++
		}
	}

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(cfgTitle.Render("  📁 " + m.relCwd()) + "\n")
	if m.dirPickMode {
		b.WriteString(cfgHint.Render("  ↑↓ gezin  |  Enter/→=klasöre gir  |  Space/s=bu dizini seç  |  ←/Esc=çık  |  q=iptal") + "\n")
	} else if m.addingCustom {
		b.WriteString(cfgHint.Render("  Dizin dışı dosya/klasör yolu girebilirsiniz (örn: ../shared/config.php)") + "\n")
	} else {
		b.WriteString(cfgHint.Render("  Space=işaretle  |  n=özel yol  |  Enter/→=gir  |  ←/Esc=çık  |  a=tümü  |  s=kaydet  |  q=iptal") + "\n")
	}
	b.WriteString(div + "\n")

	visRows := h - 7
	if visRows < 2 {
		visRows = 2
	}
	start := 0
	if m.cursor >= visRows {
		start = m.cursor - visRows + 1
	}
	end := start + visRows
	if end > len(m.entries) {
		end = len(m.entries)
	}

	// klasör/dosya ayırıcı
	dirEnd := 0
	for _, e := range m.entries {
		if !e.isDir {
			break
		}
		dirEnd++
	}

	for i := start; i < end; i++ {
		e := m.entries[i]
		if i == dirEnd && dirEnd > 0 && dirEnd < len(m.entries) {
			b.WriteString(cfgDiv.Render(strings.Repeat("·", w)) + "\n")
		}
		check := cfgDisabled.Render("[ ]")
		if m.marked[e.rel] {
			check = cfgEnabled.Render("[✓]")
		}
		icon := "📄"
		name := e.name
		if e.isDir {
			icon = "📁"
			name += "/"
		}
		if i == m.cursor {
			b.WriteString(fmt.Sprintf(" %s%s %s  %s\n", cfgCursor.Render("▶ "), check, icon, cfgSelected.Render(name)))
		} else {
			b.WriteString(fmt.Sprintf("    %s %s  %s\n", check, icon, name))
		}
	}

	if len(m.entries) == 0 {
		b.WriteString(cfgHint.Render("  (boş klasör)") + "\n")
	}

	b.WriteString(div + "\n")
	if m.dirPickMode {
		b.WriteString(cfgHint.Render("  Şu an: "+m.relCwd()+"  |  Space veya s = bu dizini seç") + "\n")
	} else if m.addingCustom {
		b.WriteString(cfgEditing.Render("  Yol: "+m.customBuf+"█") + "\n")
		b.WriteString(cfgHint.Render("  Enter=ekle  |  Esc=iptal") + "\n")
	} else {
		customCount := len(m.nonFsItems)
		if customCount > 0 {
			b.WriteString(cfgHint.Render(fmt.Sprintf("  ✓ %d işaretli  +%d özel yol  |  n=özel yol ekle", markedCount, customCount)) + "\n")
		} else {
			b.WriteString(cfgHint.Render(fmt.Sprintf("  ✓ %d işaretli  |  n=özel yol ekle", markedCount)) + "\n")
		}
	}
	return b.String()
}

// runLocalBrowser yerel dosya tarayıcısını açar.
// preSelected: mevcut liste — var olanlar ön-işaretli gösterilir.
// Döner: yeni liste (dosya sistemi dışı öğeler + browser seçimi), ok=true ise kaydet.
func runLocalBrowser(root string, preSelected []string) ([]string, bool) {
	m := newLocalBrowserModel(root, preSelected)
	p := tea.NewProgram(m, tea.WithAltScreen())
	raw, err := p.Run()
	if err != nil {
		return preSelected, false
	}
	fm := raw.(localBrowserModel)
	if fm.quit || !fm.done {
		return preSelected, false
	}

	// Seçilen yollar
	var selected []string
	for rel, v := range fm.marked {
		if v {
			selected = append(selected, rel)
		}
	}
	sort.Strings(selected)

	// Dosya sisteminde olmayan eski öğeleri koru
	result := append(fm.nonFsItems, selected...)
	return result, true
}

// ══════════════════════════════════════════════════════════════════════════════
// Ignore Dosyası Seçici
// ══════════════════════════════════════════════════════════════════════════════

var ignoreFileOptions = []string{".gitignore", "syncftp.ignore"}

type ignorePickerModel struct {
	checked [2]bool // 0=.gitignore, 1=syncftp.ignore
	cursor  int
	saved   bool
	quit    bool
}

func newIgnorePickerModel(current []string) ignorePickerModel {
	m := ignorePickerModel{}
	if current == nil {
		// nil = henüz ayarlanmamış → ikisi de seçili default
		m.checked[0] = true
		m.checked[1] = true
	} else {
		for _, f := range current {
			for i, opt := range ignoreFileOptions {
				if f == opt {
					m.checked[i] = true
				}
			}
		}
	}
	return m
}

func (m ignorePickerModel) Init() tea.Cmd { return nil }

func (m ignorePickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "Q", "esc":
			m.quit = true
			return m, tea.Quit
		case "s", "S":
			m.saved = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(ignoreFileOptions)-1 {
				m.cursor++
			}
		case " ", "enter":
			m.checked[m.cursor] = !m.checked[m.cursor]
		}
	}
	return m, nil
}

func (m ignorePickerModel) View() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(cfgTitle.Render("⚙  Ignore dosyaları") + "\n")
	b.WriteString(cfgHint.Render("  Space = seç/kaldır  |  s = kaydet  |  q = iptal") + "\n\n")

	for i, opt := range ignoreFileOptions {
		check := cfgDisabled.Render("[ ]")
		if m.checked[i] {
			check = cfgEnabled.Render("[✓]")
		}
		if i == m.cursor {
			b.WriteString(fmt.Sprintf(" %s%s  %s\n", cfgCursor.Render("▶ "), check, cfgValue.Render(opt)))
		} else {
			b.WriteString(fmt.Sprintf("    %s  %s\n", check, opt))
		}
	}

	b.WriteString("\n")
	b.WriteString(cfgHint.Render("  Hiçbirini seçmezsen ignore kullanılmaz.") + "\n")
	return b.String()
}

// runIgnoreFilePicker iki seçenekli ignore dosyası seçici açar.
// Döner: (seçilen liste, kaydedildi mi). nil = iptal.
func runIgnoreFilePicker(current []string) ([]string, bool) {
	m := newIgnorePickerModel(current)
	p := tea.NewProgram(m, tea.WithAltScreen())
	raw, err := p.Run()
	if err != nil {
		return nil, false
	}
	fm := raw.(ignorePickerModel)
	if fm.quit || !fm.saved {
		return nil, false
	}
	result := []string{}
	for i, opt := range ignoreFileOptions {
		if fm.checked[i] {
			result = append(result, opt)
		}
	}
	return result, true
}

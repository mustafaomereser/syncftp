package main

import (
	"encoding/json"
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
	"syncftp/internal/lang"
)

func init() {
	configCmd.Flags().Bool("export", false, "Config'i şifresiz JSON olarak stdout'a yaz")
	rootCmd.AddCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Server ayarlarını yönet (ekle, düzenle, sil)",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := os.Getwd()
		if exp, _ := cmd.Flags().GetBool("export"); exp {
			return runConfigExport(dir)
		}
		return runServerMgr(dir)
	},
}

func runConfigExport(dir string) error {
	cfg, err := config.Load(dir)
	if err != nil {
		return err
	}
	// Şifreleri maskele
	for i := range cfg.Connections {
		cfg.Connections[i].Password = "***"
	}
	for i := range cfg.Servers {
		cfg.Servers[i].Password = "***"
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(cfg)
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
		return lang.L.CfgValueEmpty
	}
	return strings.Join(l, ", ")
}

func padLabel(s string, w int) string {
	r := []rune(s)
	if len(r) >= w {
		return s
	}
	return s + strings.Repeat(" ", w-len(r))
}

func srvFieldLabel(i int) string {
	labels := []string{
		lang.L.CfgFldName, lang.L.CfgFldConnection, lang.L.CfgFldHost,
		lang.L.CfgFldPort, lang.L.CfgFldUser, lang.L.CfgFldPassword,
		lang.L.CfgFldRemotePath, lang.L.CfgFldLocalPath, lang.L.CfgFldEnabled,
		lang.L.CfgFldPassive, lang.L.CfgFldDisableEPSV, lang.L.CfgFldNAT,
		lang.L.CfgFldMaxConn, lang.L.CfgFldMaxRetry,
		lang.L.CfgFldInclude, lang.L.CfgFldExclude, lang.L.CfgFldProtect,
		lang.L.CfgFldWebhook, lang.L.CfgFldMinify, lang.L.CfgFldObfuscate,
		lang.L.CfgFldBlockExts,
	}
	if i >= 0 && i < len(labels) {
		return labels[i]
	}
	return ""
}

func globalFieldLabel(i int) string {
	labels := []string{
		lang.L.CfgGFldDefaultPath, lang.L.CfgGFldProtect,
		lang.L.CfgGFldInclude, lang.L.CfgGFldExclude, lang.L.CfgGFldIgnore,
		lang.L.CfgGFldWebhook, lang.L.CfgGFldBlockExts,
	}
	if i >= 0 && i < len(labels) {
		return labels[i]
	}
	return ""
}

func connFieldLabel(i int) string {
	labels := []string{
		lang.L.CfgFldName, lang.L.CfgFldHost, lang.L.CfgFldPort,
		lang.L.CfgFldUser, lang.L.CfgFldPassword,
		lang.L.CfgFldPassive, lang.L.CfgFldDisableEPSV, lang.L.CfgFldNAT,
	}
	if i >= 0 && i < len(labels) {
		return labels[i]
	}
	return ""
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
	cfgLabel    = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	cfgValue    = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	cfgEditing  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
	cfgBrowse   = lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("6"))
	cfgGlobal   = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	cfgSelected = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
)

// ══════════════════════════════════════════════════════════════════════════════
// Server Yönetici TUI
// ══════════════════════════════════════════════════════════════════════════════

// cursor: 0 = Global Ayarlar, 1 = Bağlantı Profilleri, 2..N+1 = serverler, N+2 = + Ekle
type serverMgrModel struct {
	servers     []config.Server
	connections []config.Connection
	cursor      int
	confirming  bool
	action      string // "edit-global","edit-connections","edit","add","delete","toggle"
	actionIdx   int
	quit        bool
	width       int
	height      int
}

func newServerMgrModel(servers []config.Server, connections []config.Connection) serverMgrModel {
	cur := 2
	if len(servers) == 0 {
		cur = 2 // "+ Ekle" konumu
	}
	return serverMgrModel{servers: servers, connections: connections, cursor: cur}
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
				m.actionIdx = m.cursor - 2
				return m, tea.Quit
			default:
				m.confirming = false
			}
			return m, nil
		}
		total := len(m.servers) + 3
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
			case m.cursor == 1:
				m.action = "edit-connections"
				return m, tea.Quit
			case m.cursor >= 2 && m.cursor <= len(m.servers)+1:
				m.action = "edit"
				m.actionIdx = m.cursor - 2
				return m, tea.Quit
			default:
				m.action = "add"
				return m, tea.Quit
			}
		case "n", "N":
			m.action = "add"
			return m, tea.Quit
		case "d", "D":
			if m.cursor >= 2 && m.cursor <= len(m.servers)+1 {
				m.confirming = true
			}
		case " ":
			if m.cursor >= 2 && m.cursor <= len(m.servers)+1 {
				m.action = "toggle"
				m.actionIdx = m.cursor - 2
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
	b.WriteString(cfgTitle.Render("⚙  "+lang.L.CfgSrvTitle) + "\n")
	b.WriteString(cfgHint.Render(lang.L.CfgSrvNavHint) + "\n")
	b.WriteString(div + "\n\n")

	if m.confirming {
		name := m.servers[m.cursor-2].Name
		b.WriteString(cfgWarn.Render(fmt.Sprintf("  "+lang.L.CfgDeleteConfirmFmt, name)) + "\n")
		return b.String()
	}

	// Global Ayarlar
	if m.cursor == 0 {
		b.WriteString(cfgCursor.Render("▶ ") + cfgGlobal.Bold(true).Render(lang.L.CfgGlobalLabel) +
			cfgHint.Render(lang.L.CfgGlobalHint) + "\n")
	} else {
		b.WriteString("  " + cfgGlobal.Render(lang.L.CfgGlobalLabel) +
			cfgHint.Render(lang.L.CfgGlobalHint) + "\n")
	}

	// Bağlantı Profilleri
	connCount := cfgDisabled.Render(fmt.Sprintf(lang.L.CfgConnCountFmt, len(m.connections)))
	if m.cursor == 1 {
		b.WriteString(cfgCursor.Render("▶ ") + cfgGlobal.Bold(true).Render(lang.L.CfgConnProfilesLabel) +
			"  " + connCount + "\n")
	} else {
		b.WriteString("  " + cfgGlobal.Render(lang.L.CfgConnProfilesLabel) +
			"  " + connCount + "\n")
	}

	// Serverler
	for i, srv := range m.servers {
		pos := i + 2
		icon := cfgEnabled.Render("✓")
		nameS := cfgEnabled.Render(srv.Name)
		if !srv.Enabled {
			icon = cfgDisabled.Render("○")
			nameS = cfgDisabled.Render(srv.Name + " (" + lang.L.CfgDisabledLabel + ")")
		}
		hostStr := srv.Host
		if srv.Connection != "" {
			hostStr = "→ " + srv.Connection
		}
		host := cfgDisabled.Render(fmt.Sprintf("%s:%d%s", hostStr, srv.Port, srv.RemotePath))
		info := cfgDisabled.Render(fmt.Sprintf("conn:%d  retry:%d", srv.MaxConnections, srv.MaxRetries))
		if pos == m.cursor {
			b.WriteString(fmt.Sprintf(" %s%s  %-22s  %-38s  %s\n",
				cfgCursor.Render("▶ "), icon, nameS, host, info))
		} else {
			b.WriteString(fmt.Sprintf("    %s  %-22s  %-38s  %s\n", icon, nameS, host, info))
		}
	}

	// + Ekle
	addPos := len(m.servers) + 2
	if m.cursor == addPos {
		b.WriteString(cfgCursor.Render("▶ ") + cfgAdd.Render(lang.L.CfgNewServerLabel) + "\n")
	} else {
		b.WriteString("    " + cfgAdd.Render(lang.L.CfgNewServerLabel) + "\n")
	}

	b.WriteString("\n" + div + "\n")
	b.WriteString(cfgHint.Render(fmt.Sprintf(lang.L.CfgSrvCountFmt, len(m.servers), len(m.connections))) + "\n")
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

		p := tea.NewProgram(newServerMgrModel(cfg.Servers, cfg.Connections), tea.WithAltScreen())
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
					fmt.Printf(lang.L.CfgSaveErr, e)
				}
			}

		case "edit-connections":
			runConnectionMgr(configDir, cfg)

		case "edit":
			srv := cfg.Servers[fm.actionIdx]
			if runServerEdit(&srv, cfg.Connections, projectDir, false) {
				cfg.Servers[fm.actionIdx] = srv
				if e := config.Save(configDir, cfg); e != nil {
					fmt.Printf(lang.L.CfgSaveErr, e)
				} else {
					fmt.Printf(lang.L.CfgUpdatedFmt, srv.Name)
				}
			}

		case "add":
			srv := config.Server{Port: 21, Passive: true, Enabled: true, MaxConnections: 1, MaxRetries: 2}
			if runServerEdit(&srv, cfg.Connections, projectDir, true) {
				cfg.Servers = append(cfg.Servers, srv)
				if e := config.Save(configDir, cfg); e != nil {
					fmt.Printf(lang.L.CfgSaveErr, e)
				} else {
					fmt.Printf(lang.L.CfgAddedFmt, srv.Name)
				}
			}

		case "delete":
			name := cfg.Servers[fm.actionIdx].Name
			cfg.Servers = append(cfg.Servers[:fm.actionIdx], cfg.Servers[fm.actionIdx+1:]...)
			// connection referansları kaldı — kasıtlı, kullanıcı bunları yönetir
			if e := config.Save(configDir, cfg); e != nil {
				fmt.Printf(lang.L.CfgSaveErr, e)
			} else {
				fmt.Printf(lang.L.CfgDeletedFmt, name)
			}

		case "toggle":
			cfg.Servers[fm.actionIdx].Enabled = !cfg.Servers[fm.actionIdx].Enabled
			s := lang.L.CfgDisabledLabel
			if cfg.Servers[fm.actionIdx].Enabled {
				s = lang.L.CfgEnabledLabel
			}
			if e := config.Save(configDir, cfg); e != nil {
				fmt.Printf(lang.L.CfgSaveErr, e)
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

// isCredentialField returns true for fields whose values come from a Connection profile.
// When srv.Connection != "", these fields are read-only in the editor.
//
//	2=Host 3=Port 4=User 5=Password 9=Passive 10=DisableEPSV 11=NATWorkaround
func isCredentialField(idx int) bool {
	return idx == 2 || idx == 3 || idx == 4 || idx == 5 || idx == 9 || idx == 10 || idx == 11
}

var serverFields = []srvField{
	{"Ad", "text", ""},
	{"Connection", "conn", ""},                // idx 1 — bağlantı profili; boş = manuel giriş
	{"Host", "text", ""},                      // idx 2
	{"Port", "int", ""},                       // idx 3
	{"Kullanıcı", "text", ""},                 // idx 4
	{"Şifre", "pass", ""},                     // idx 5
	{"Uzak dizin", "text", ""},                // idx 6
	{"Yerel dizin", "localdir", ""},           // idx 7
	{"Aktif", "bool", ""},                     // idx 8
	{"Passive mode", "bool", ""},              // idx 9
	{"EPSV devre dışı", "bool", ""},           // idx 10
	{"NAT workaround", "bool", ""},            // idx 11
	{"Max bağlantı", "int", ""},               // idx 12
	{"Max retry", "int", ""},                  // idx 13
	{"Include", "list", "include"},            // idx 14
	{"Exclude", "list", "exclude"},            // idx 15
	{"Protect", "list", "protect"},            // idx 16
	{"Webhook URL", "text", ""},               // idx 17
	{"Minify CSS/JS", "bool", ""},             // idx 18
	{"Obfuscate JS", "bool", ""},              // idx 19
	{"Block extensions", "list", "block_ext"}, // idx 20
}

type serverEditModel struct {
	srv         config.Server
	connections []config.Connection // mevcut bağlantı profilleri (Connection picker için)
	isNew       bool
	projectDir  string
	cursor      int
	editing     bool
	editBuf     string
	action      string // "save","cancel","browse","browse-localpath","pick-connection"
	browseFor   string
	width       int
	height      int
}

func newServerEditModel(srv config.Server, connections []config.Connection, isNew bool, projectDir string, cursor int) serverEditModel {
	return serverEditModel{srv: srv, connections: connections, isNew: isNew, projectDir: projectDir, cursor: cursor}
}

func (m *serverEditModel) getDisplayValue(i int) string {
	switch i {
	case 0:
		return m.srv.Name
	case 1: // Connection
		if m.srv.Connection != "" {
			return m.srv.Connection
		}
		return lang.L.CfgSrvManual
	case 2:
		return m.srv.Host
	case 3:
		return strconv.Itoa(m.srv.Port)
	case 4:
		return m.srv.User
	case 5:
		if m.srv.Password != "" {
			return "****"
		}
		return lang.L.CfgValueEmpty
	case 6:
		return m.srv.RemotePath
	case 7:
		if m.srv.LocalPath != "" {
			return m.srv.LocalPath
		}
		return lang.L.CfgSrvProjectDef
	case 8:
		return boolIcon(m.srv.Enabled)
	case 9:
		return boolIcon(m.srv.Passive)
	case 10:
		return boolIcon(m.srv.DisableEPSV)
	case 11:
		return boolIcon(m.srv.NATWorkaround)
	case 12:
		return strconv.Itoa(m.srv.MaxConnections)
	case 13:
		return strconv.Itoa(m.srv.MaxRetries)
	case 14:
		return listDisplay(m.srv.Include)
	case 15:
		return listDisplay(m.srv.Exclude)
	case 16:
		return listDisplay(m.srv.Protect)
	case 17:
		if m.srv.Webhook != "" {
			return m.srv.Webhook
		}
		return lang.L.CfgValueEmpty
	case 18:
		return boolIcon(m.srv.Minify)
	case 19:
		return boolIcon(m.srv.Obfuscate)
	case 20:
		return listDisplay(m.srv.BlockExtensions)
	}
	return ""
}

func (m *serverEditModel) getEditableValue(i int) string {
	switch i {
	case 0:
		return m.srv.Name
	case 2:
		return m.srv.Host
	case 3:
		return strconv.Itoa(m.srv.Port)
	case 4:
		return m.srv.User
	case 5:
		return m.srv.Password
	case 6:
		return m.srv.RemotePath
	case 7:
		return m.srv.LocalPath
	case 12:
		return strconv.Itoa(m.srv.MaxConnections)
	case 13:
		return strconv.Itoa(m.srv.MaxRetries)
	case 14:
		return strings.Join(m.srv.Include, ", ")
	case 15:
		return strings.Join(m.srv.Exclude, ", ")
	case 16:
		return strings.Join(m.srv.Protect, ", ")
	case 17:
		return m.srv.Webhook
	case 20:
		return strings.Join(m.srv.BlockExtensions, ", ")
	}
	return ""
}

func (m *serverEditModel) applyValue(i int, val string) {
	switch i {
	case 0:
		m.srv.Name = val
	case 2:
		m.srv.Host = val
	case 3:
		if n, e := strconv.Atoi(val); e == nil {
			m.srv.Port = n
		}
	case 4:
		m.srv.User = val
	case 5:
		m.srv.Password = val
	case 6:
		m.srv.RemotePath = val
	case 7:
		m.srv.LocalPath = val
	case 12:
		if n, e := strconv.Atoi(val); e == nil {
			m.srv.MaxConnections = n
		}
	case 13:
		if n, e := strconv.Atoi(val); e == nil {
			m.srv.MaxRetries = n
		}
	case 14:
		m.srv.Include = parseList(val)
	case 15:
		m.srv.Exclude = parseList(val)
	case 16:
		m.srv.Protect = parseList(val)
	case 17:
		m.srv.Webhook = val
	case 20:
		m.srv.BlockExtensions = parseList(val)
	}
}

func (m *serverEditModel) toggleBool(i int) {
	switch i {
	case 8:
		m.srv.Enabled = !m.srv.Enabled
	case 9:
		m.srv.Passive = !m.srv.Passive
	case 10:
		m.srv.DisableEPSV = !m.srv.DisableEPSV
	case 11:
		m.srv.NATWorkaround = !m.srv.NATWorkaround
	case 18:
		m.srv.Minify = !m.srv.Minify
	case 19:
		m.srv.Obfuscate = !m.srv.Obfuscate
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
	locked := m.srv.Connection != "" && isCredentialField(m.cursor)
	switch msg.String() {
	case "ctrl+c", "q", "Q", "esc":
		m.action = "cancel"
		return m, tea.Quit
	case "s", "S":
		m.action = "save"
		return m, tea.Quit
	case "up", "k":
		next := m.cursor - 1
		for next > 0 && m.srv.Connection != "" && isCredentialField(next) {
			next--
		}
		if next >= 0 {
			m.cursor = next
		}
	case "down", "j":
		next := m.cursor + 1
		for next < len(serverFields) && m.srv.Connection != "" && isCredentialField(next) {
			next++
		}
		if next < len(serverFields) {
			m.cursor = next
		}
	case "enter":
		if locked {
			break
		}
		switch f.kind {
		case "conn":
			m.action = "pick-connection"
			return m, tea.Quit
		case "bool":
			m.toggleBool(m.cursor)
		default:
			m.editBuf = m.getEditableValue(m.cursor)
			m.editing = true
		}
	case " ":
		if !locked && f.kind == "bool" {
			m.toggleBool(m.cursor)
		}
	case "b", "B":
		if f.kind == "conn" {
			m.action = "pick-connection"
			return m, tea.Quit
		}
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

	title := lang.L.CfgSrvEditTitle
	if m.isNew {
		title = lang.L.CfgSrvNewTitle
	}

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(cfgTitle.Render("⚙  "+title) + "\n")
	b.WriteString(cfgHint.Render(lang.L.CfgSrvEditNavHint) + "\n")
	b.WriteString(div + "\n\n")

	for i, f := range serverFields {
		locked := m.srv.Connection != "" && isCredentialField(i)
		label := cfgLabel.Render(padLabel(srvFieldLabel(i), 32))
		var valStr string

		if i == m.cursor && m.editing {
			valStr = cfgEditing.Render(m.editBuf + "█")
		} else {
			raw := m.getDisplayValue(i)
			switch {
			case locked:
				valStr = cfgDisabled.Render(raw) + cfgHint.Render("  ↳ "+m.srv.Connection)
			case f.kind == "conn":
				if m.srv.Connection != "" {
					valStr = cfgEnabled.Render(raw) + "  " + cfgBrowse.Render(lang.L.CfgSrvConnChange)
				} else {
					valStr = cfgDisabled.Render(raw) + "  " + cfgBrowse.Render(lang.L.CfgSrvConnSelect)
				}
			case f.kind == "bool":
				if raw == "✓" {
					valStr = cfgEnabled.Render(raw)
				} else {
					valStr = cfgDisabled.Render(raw)
				}
			case f.kind == "localdir":
				valStr = cfgValue.Render(raw)
				if m.projectDir != "" {
					valStr += "  " + cfgBrowse.Render(lang.L.CfgSrvBrowseHint)
				}
			case f.kind == "list":
				valStr = cfgValue.Render(raw)
				if m.projectDir != "" {
					valStr += "  " + cfgBrowse.Render(lang.L.CfgSrvBrowseHint)
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
	b.WriteString(cfgHint.Render(lang.L.CfgSrvSaveHint) + "\n")
	return b.String()
}

// runServerEdit — döngü: field edit + browse + connection pick arası geçiş
func runServerEdit(srv *config.Server, connections []config.Connection, projectDir string, isNew bool) (saved bool) {
	cursor := 0
	for {
		m := newServerEditModel(*srv, connections, isNew, projectDir, cursor)
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
		case "pick-connection":
			*srv = fm.srv
			cursor = 1
			connName := runConnectionPicker(connections, srv.Connection)
			if connName == "" {
				// kullanıcı "(manuel)" seçti → bağlantı profilini temizle
				srv.Connection = ""
			} else {
				srv.Connection = connName
				// Seçilen profili hemen çöz (UI'da credential alanları güncellenir)
				for _, c := range connections {
					if c.Name == connName {
						srv.Host = c.Host
						srv.Port = c.Port
						srv.User = c.User
						srv.Password = c.Password
						srv.Passive = c.Passive
						srv.DisableEPSV = c.DisableEPSV
						srv.NATWorkaround = c.NATWorkaround
						break
					}
				}
			}
		case "browse":
			*srv = fm.srv
			if fm.browseFor == "block_ext" {
				cursor = 20
				if result, ok := runBlockExtPicker(srv.BlockExtensions); ok {
					srv.BlockExtensions = result
				}
			} else {
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
					cursor = 14
				case "exclude":
					current = srv.Exclude
					cursor = 15
				case "protect":
					current = srv.Protect
					cursor = 16
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
			}
		case "browse-localpath":
			*srv = fm.srv
			cursor = 7
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
	{"Webhook URL (global)", "text", ""},
	{"Block extensions (global)", "list", "block_ext"},
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
		return lang.L.CfgGlobalDefaultEmpty
	case 1:
		return listDisplay(m.sync.Protect)
	case 2:
		return listDisplay(m.sync.Include)
	case 3:
		return listDisplay(m.sync.Exclude)
	case 4:
		if m.sync.IgnoreFiles == nil {
			return lang.L.CfgGlobalIgnoreDefVal
		}
		if len(m.sync.IgnoreFiles) == 0 {
			return lang.L.CfgGlobalIgnoreNone
		}
		return strings.Join(m.sync.IgnoreFiles, ", ")
	case 5:
		if m.sync.Webhook != "" {
			return m.sync.Webhook
		}
		return lang.L.CfgValueEmpty
	case 6:
		return listDisplay(m.sync.BlockExtensions)
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
	case 5:
		return m.sync.Webhook
	case 6:
		return strings.Join(m.sync.BlockExtensions, ", ")
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
	case 5:
		m.sync.Webhook = val
	case 6:
		m.sync.BlockExtensions = parseList(val)
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
	b.WriteString(cfgTitle.Render(lang.L.CfgGlobalEditTitle) + "\n")
	b.WriteString(cfgHint.Render(lang.L.CfgGlobalEditNavHint) + "\n")
	b.WriteString(div + "\n\n")

	for i, f := range globalFields {
		label := cfgLabel.Render(padLabel(globalFieldLabel(i), 32))
		var valStr string
		if i == m.cursor && m.editing {
			valStr = cfgEditing.Render(m.editBuf + "█")
		} else {
			raw := m.getDisplayValue(i)
			valStr = cfgValue.Render(raw)
			if f.kind == "localdir" && m.projectDir != "" {
				valStr += "  " + cfgBrowse.Render(lang.L.CfgSrvBrowseHint)
			} else if f.browseID != "" && m.projectDir != "" {
				valStr += "  " + cfgBrowse.Render(lang.L.CfgSrvBrowseHint)
			}
		}
		if i == m.cursor && !m.editing {
			b.WriteString(cfgCursor.Render("▶ ") + label + "  " + valStr + "\n")
		} else {
			b.WriteString("  " + label + "  " + valStr + "\n")
		}
	}

	b.WriteString("\n" + div + "\n")
	b.WriteString(cfgHint.Render(lang.L.CfgSrvSaveHint) + "\n")
	b.WriteString(cfgHint.Render(lang.L.CfgGlobalIgnoreHint) + "\n")
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
			if fm.browseFor == "block_ext" {
				cursor = 6
				if result, ok := runBlockExtPicker(s.BlockExtensions); ok {
					s.BlockExtensions = result
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

// runBlockExtPicker yaygın uzantıları multi-select olarak gösterir; mevcut olanlar önceden işaretli gelir.
// Elle eklenmiş (listede olmayan) uzantılar korunur.
func runBlockExtPicker(current []string) ([]string, bool) {
	common := []string{
		".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg", ".ico", ".bmp", ".tiff",
		".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
		".zip", ".rar", ".tar", ".gz", ".7z", ".bz2",
		".mp4", ".mp3", ".avi", ".mov", ".mkv", ".wav", ".ogg", ".flac",
		".woff", ".woff2", ".ttf", ".eot", ".otf",
		".exe", ".dll", ".so", ".dylib",
		".map",
	}
	commonSet := make(map[string]bool, len(common))
	for _, e := range common {
		commonSet[e] = true
	}

	currentSet := make(map[string]bool, len(current))
	for _, c := range current {
		ext := strings.ToLower(c)
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		currentSet[ext] = true
	}

	items := make([]PickerItem, len(common))
	checked := make([]bool, len(common))
	for i, ext := range common {
		items[i] = PickerItem{Label: ext, Value: ext}
		checked[i] = currentSet[ext]
	}

	// Listede olmayan, elle eklenmiş uzantıları koru
	var extra []string
	for _, c := range current {
		ext := strings.ToLower(c)
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		if !commonSet[ext] {
			extra = append(extra, ext)
		}
	}

	m := multiPickerModel{
		title:    lang.L.CfgFldBlockExts,
		subtitle: "Boşluk=işaretle  Enter=onayla  Esc=iptal  |  Elle eklemek için TUI'dan çık, Enter ile düzenle",
		items:    items,
		checked:  checked,
	}
	selected, err := runMultiPickerRaw(m)
	if err != nil || selected == nil {
		return nil, false
	}
	return append(selected, extra...), true
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
	b.WriteString(cfgTitle.Render("  📁 "+m.relCwd()) + "\n")
	if m.dirPickMode {
		b.WriteString(cfgHint.Render(lang.L.CfgLocalDirPickHint) + "\n")
	} else if m.addingCustom {
		b.WriteString(cfgHint.Render(lang.L.CfgLocalCustomHint) + "\n")
	} else {
		b.WriteString(cfgHint.Render(lang.L.CfgLocalNavHint) + "\n")
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
		b.WriteString(cfgHint.Render(lang.L.CfgLocalEmpty) + "\n")
	}

	b.WriteString(div + "\n")
	if m.dirPickMode {
		b.WriteString(cfgHint.Render(fmt.Sprintf(lang.L.CfgLocalSelectFmt, m.relCwd())) + "\n")
	} else if m.addingCustom {
		b.WriteString(cfgEditing.Render(lang.L.CfgLocalPathLabel+m.customBuf+"█") + "\n")
		b.WriteString(cfgHint.Render(lang.L.CfgLocalCustomSave) + "\n")
	} else {
		customCount := len(m.nonFsItems)
		if customCount > 0 {
			b.WriteString(cfgHint.Render(fmt.Sprintf(lang.L.CfgLocalMarkedFmt, markedCount, customCount)) + "\n")
		} else {
			b.WriteString(cfgHint.Render(fmt.Sprintf(lang.L.CfgLocalMarkedSimple, markedCount)) + "\n")
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
	b.WriteString(cfgTitle.Render(lang.L.CfgIgnoreTitle) + "\n")
	b.WriteString(cfgHint.Render(lang.L.CfgIgnoreNavHint) + "\n\n")

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
	b.WriteString(cfgHint.Render(lang.L.CfgIgnoreNote) + "\n")
	return b.String()
}

// ══════════════════════════════════════════════════════════════════════════════
// Bağlantı Profili Seçici
// ══════════════════════════════════════════════════════════════════════════════

// runConnectionPicker RunPicker ile bir bağlantı profili seçtirir.
// Döner: seçilen profil adı ("" = manuel / iptal).
func runConnectionPicker(connections []config.Connection, current string) string {
	items := []PickerItem{
		{Icon: "✏️ ", Label: lang.L.CfgConnPickerManual, Desc: lang.L.CfgConnPickerManualDesc, Value: ""},
	}
	for _, c := range connections {
		desc := fmt.Sprintf("%s:%d  %s", c.Host, c.Port, c.User)
		items = append(items, PickerItem{Icon: "🔗", Label: c.Name, Desc: desc, Value: c.Name})
	}
	subtitle := lang.L.CfgConnPickerSub
	if current != "" {
		subtitle = fmt.Sprintf(lang.L.CfgConnPickerCurrentFmt, current)
	}
	val, err := RunPicker(lang.L.CfgConnPickerTitle, subtitle, items)
	if err != nil {
		return current
	}
	return val
}

// ══════════════════════════════════════════════════════════════════════════════
// Bağlantı Profili Yönetici TUI
// ══════════════════════════════════════════════════════════════════════════════

type connMgrModel struct {
	connections []config.Connection
	cursor      int
	confirming  bool
	action      string // "edit","add","delete","quit"
	actionIdx   int
	width       int
	height      int
}

func newConnMgrModel(connections []config.Connection) connMgrModel {
	cur := 0
	if len(connections) > 0 {
		cur = 0
	}
	return connMgrModel{connections: connections, cursor: cur}
}

func (m connMgrModel) Init() tea.Cmd { return nil }

func (m connMgrModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		if m.confirming {
			switch msg.String() {
			case "y", "Y", "enter":
				m.action = "delete"
				m.actionIdx = m.cursor
				return m, tea.Quit
			default:
				m.confirming = false
			}
			return m, nil
		}
		total := len(m.connections) + 1 // +1 = "+ Ekle"
		switch msg.String() {
		case "ctrl+c", "q", "Q", "esc":
			m.action = "quit"
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
			if m.cursor < len(m.connections) {
				m.action = "edit"
				m.actionIdx = m.cursor
			} else {
				m.action = "add"
			}
			return m, tea.Quit
		case "n", "N":
			m.action = "add"
			return m, tea.Quit
		case "d", "D":
			if m.cursor < len(m.connections) {
				m.confirming = true
			}
		}
	}
	return m, nil
}

func (m connMgrModel) View() string {
	w := m.width
	if w < 60 {
		w = 80
	}
	div := cfgDiv.Render("  " + strings.Repeat("─", w-4))
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(cfgTitle.Render(lang.L.CfgConnMgrTitle) + "\n")
	b.WriteString(cfgHint.Render(lang.L.CfgConnMgrNavHint) + "\n")
	b.WriteString(div + "\n\n")

	if m.confirming {
		name := m.connections[m.cursor].Name
		b.WriteString(cfgWarn.Render(fmt.Sprintf("  "+lang.L.CfgDeleteConfirmFmt, name)) + "\n")
		return b.String()
	}

	for i, c := range m.connections {
		host := cfgDisabled.Render(fmt.Sprintf("%s:%d  %s", c.Host, c.Port, c.User))
		if i == m.cursor {
			b.WriteString(fmt.Sprintf(" %s🔗 %-22s  %s\n", cfgCursor.Render("▶ "), cfgValue.Render(c.Name), host))
		} else {
			b.WriteString(fmt.Sprintf("    🔗 %-22s  %s\n", c.Name, host))
		}
	}

	addPos := len(m.connections)
	if m.cursor == addPos {
		b.WriteString(cfgCursor.Render("▶ ") + cfgAdd.Render(lang.L.CfgNewConnLabel) + "\n")
	} else {
		b.WriteString("    " + cfgAdd.Render(lang.L.CfgNewConnLabel) + "\n")
	}

	b.WriteString("\n" + div + "\n")
	b.WriteString(cfgHint.Render(fmt.Sprintf(lang.L.CfgConnTotalFmt, len(m.connections))) + "\n")
	return b.String()
}

// ── Bağlantı Profili Düzenleme TUI ───────────────────────────────────────────

type connField struct {
	label string
	kind  string // "text","pass","int","bool"
}

var connFields = []connField{
	{"Ad", "text"},
	{"Host", "text"},
	{"Port", "int"},
	{"Kullanıcı", "text"},
	{"Şifre", "pass"},
	{"Passive mode", "bool"},
	{"EPSV devre dışı", "bool"},
	{"NAT workaround", "bool"},
}

type connEditModel struct {
	conn    config.Connection
	isNew   bool
	cursor  int
	editing bool
	editBuf string
	action  string // "save","cancel"
	width   int
	height  int
}

func newConnEditModel(conn config.Connection, isNew bool) connEditModel {
	return connEditModel{conn: conn, isNew: isNew}
}

func (m *connEditModel) getDisplayValue(i int) string {
	switch i {
	case 0:
		return m.conn.Name
	case 1:
		return m.conn.Host
	case 2:
		return strconv.Itoa(m.conn.Port)
	case 3:
		return m.conn.User
	case 4:
		if m.conn.Password != "" {
			return "****"
		}
		return lang.L.CfgValueEmpty
	case 5:
		return boolIcon(m.conn.Passive)
	case 6:
		return boolIcon(m.conn.DisableEPSV)
	case 7:
		return boolIcon(m.conn.NATWorkaround)
	}
	return ""
}

func (m *connEditModel) getEditableValue(i int) string {
	switch i {
	case 0:
		return m.conn.Name
	case 1:
		return m.conn.Host
	case 2:
		return strconv.Itoa(m.conn.Port)
	case 3:
		return m.conn.User
	case 4:
		return m.conn.Password
	}
	return ""
}

func (m *connEditModel) applyValue(i int, val string) {
	switch i {
	case 0:
		m.conn.Name = val
	case 1:
		m.conn.Host = val
	case 2:
		if n, e := strconv.Atoi(val); e == nil {
			m.conn.Port = n
		}
	case 3:
		m.conn.User = val
	case 4:
		m.conn.Password = val
	}
}

func (m *connEditModel) toggleBool(i int) {
	switch i {
	case 5:
		m.conn.Passive = !m.conn.Passive
	case 6:
		m.conn.DisableEPSV = !m.conn.DisableEPSV
	case 7:
		m.conn.NATWorkaround = !m.conn.NATWorkaround
	}
}

func (m connEditModel) Init() tea.Cmd { return nil }

func (m connEditModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.cursor < len(connFields)-1 {
				m.cursor++
			}
		case "enter":
			f := connFields[m.cursor]
			if f.kind == "bool" {
				m.toggleBool(m.cursor)
			} else {
				m.editBuf = m.getEditableValue(m.cursor)
				m.editing = true
			}
		case " ":
			if connFields[m.cursor].kind == "bool" {
				m.toggleBool(m.cursor)
			}
		}
	}
	return m, nil
}

func (m connEditModel) View() string {
	w := m.width
	if w < 60 {
		w = 80
	}
	div := cfgDiv.Render("  " + strings.Repeat("─", w-4))
	title := lang.L.CfgConnEditTitle
	if m.isNew {
		title = lang.L.CfgConnNewTitle
	}
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(cfgTitle.Render(title) + "\n")
	b.WriteString(cfgHint.Render(lang.L.CfgConnEditNavHint) + "\n")
	b.WriteString(div + "\n\n")

	for i, f := range connFields {
		label := cfgLabel.Render(padLabel(connFieldLabel(i), 32))
		var valStr string
		if i == m.cursor && m.editing {
			valStr = cfgEditing.Render(m.editBuf + "█")
		} else {
			raw := m.getDisplayValue(i)
			if f.kind == "bool" {
				if raw == "✓" {
					valStr = cfgEnabled.Render(raw)
				} else {
					valStr = cfgDisabled.Render(raw)
				}
			} else {
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
	b.WriteString(cfgHint.Render(lang.L.CfgSrvSaveHint) + "\n")
	return b.String()
}

// runConnectionEdit profil düzenleme döngüsünü çalıştırır.
func runConnectionEdit(conn *config.Connection, isNew bool) bool {
	m := newConnEditModel(*conn, isNew)
	p := tea.NewProgram(m, tea.WithAltScreen())
	raw, err := p.Run()
	if err != nil {
		return false
	}
	fm := raw.(connEditModel)
	if fm.action != "save" {
		return false
	}
	*conn = fm.conn
	return true
}

// runConnectionMgr bağlantı profilleri yönetim döngüsü.
func runConnectionMgr(configDir string, cfg *config.Config) {
	for {
		p := tea.NewProgram(newConnMgrModel(cfg.Connections), tea.WithAltScreen())
		raw, err := p.Run()
		if err != nil {
			return
		}
		fm := raw.(connMgrModel)

		switch fm.action {
		case "quit", "":
			return
		case "edit":
			conn := cfg.Connections[fm.actionIdx]
			if runConnectionEdit(&conn, false) {
				cfg.Connections[fm.actionIdx] = conn
				if e := config.Save(configDir, cfg); e != nil {
					fmt.Printf(lang.L.CfgSaveErr, e)
				} else {
					fmt.Printf(lang.L.CfgUpdatedFmt, conn.Name)
				}
			}
		case "add":
			conn := config.Connection{Port: 21, Passive: true}
			if runConnectionEdit(&conn, true) {
				cfg.Connections = append(cfg.Connections, conn)
				if e := config.Save(configDir, cfg); e != nil {
					fmt.Printf(lang.L.CfgSaveErr, e)
				} else {
					fmt.Printf(lang.L.CfgAddedFmt, conn.Name)
				}
			}
		case "delete":
			name := cfg.Connections[fm.actionIdx].Name
			cfg.Connections = append(cfg.Connections[:fm.actionIdx], cfg.Connections[fm.actionIdx+1:]...)
			if e := config.Save(configDir, cfg); e != nil {
				fmt.Printf(lang.L.CfgSaveErr, e)
			} else {
				fmt.Printf(lang.L.CfgDeletedFmt, name)
			}
		}
	}
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

package views

import (
	"fmt"
	"strings"

	"github.com/brunoomariano/ShotGum-Toolchain/internal/registry"
	"github.com/brunoomariano/ShotGum-Toolchain/internal/runner"
	"github.com/brunoomariano/ShotGum-Toolchain/internal/tui/styles"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type detailMode int

const (
	detailNone detailMode = iota
	detailCategory
	detailScript
)

// DetailHelpMsg carries the captured --help output for a specific script.
type DetailHelpMsg struct {
	ScriptName string
	Text       string
}

// LoadScriptHelpCmd runs the script with its help flag asynchronously and
// returns a DetailHelpMsg with the captured output.
// width is the right-panel content width, passed to gum via COLUMNS so it can
// render borders correctly without a real TTY.
func LoadScriptHelpCmd(entry registry.ScriptEntry, reg *registry.Registry, width int) tea.Cmd {
	return func() tea.Msg {
		text, _ := runner.CaptureRunForPreview(entry, []string{reg.ResolveHelpFlag(entry)}, reg, width)
		return DetailHelpMsg{
			ScriptName: entry.Name,
			Text:       strings.TrimSpace(text),
		}
	}
}

// DetailModel renders contextual information in the right panel.
// The help section is contained in a scrollable viewport.
type DetailModel struct {
	mode       detailMode
	catEntry   *registry.CategoryEntry
	scripts    []registry.ScriptEntry
	scrEntry   *registry.ScriptEntry
	reg        *registry.Registry
	helpText   string
	helpScript string
	vp         viewport.Model
	vpReady    bool
}

// SetSize must be called whenever the panel dimensions change.
// It sizes the help viewport to fill the space below the KV section.
func (m *DetailModel) SetSize(w, h int) {
	// Fixed section: title(1) + sep(1) + name(1) + desc(1) + blank(1) + kv(5) + blank(1) = 11
	// Help header:   helpTitle(1) + helpSep(1) = 2  →  total fixed = 13, leave 1 for margin
	vpH := h - 14
	if vpH < 2 {
		vpH = 2
	}
	if !m.vpReady {
		m.vp = viewport.New(w, vpH)
		m.vpReady = true
	} else {
		m.vp.Width = w
		m.vp.Height = vpH
	}
	if m.helpText != "" {
		m.vp.SetContent(m.helpText)
	}
}

// SetCategory updates the detail view to show category information.
func (m *DetailModel) SetCategory(entry *registry.CategoryEntry, scripts []registry.ScriptEntry, reg *registry.Registry) {
	m.mode = detailCategory
	m.catEntry = entry
	m.scripts = scripts
	m.scrEntry = nil
	m.reg = reg
	m.helpText = ""
	m.helpScript = ""
}

// SetScript updates the detail view to show script information and clears
// any previously cached help text.
func (m *DetailModel) SetScript(entry *registry.ScriptEntry, reg *registry.Registry) {
	m.mode = detailScript
	m.scrEntry = entry
	m.catEntry = nil
	m.reg = reg
	m.helpText = ""
	m.helpScript = ""
	m.vp.SetContent("")
	m.vp.GotoTop()
}

// HelpText returns the cached help text for the currently shown script.
func (m DetailModel) HelpText() string {
	return m.helpText
}

// SetHelpText stores the help output if it belongs to the currently shown script.
func (m *DetailModel) SetHelpText(scriptName, text string) {
	if m.scrEntry != nil && m.scrEntry.Name == scriptName {
		m.helpText = text
		m.helpScript = scriptName
		m.vp.SetContent(text)
		m.vp.GotoTop()
	}
}

// Update forwards scroll messages to the help viewport.
func (m DetailModel) Update(msg tea.Msg) (DetailModel, tea.Cmd) {
	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(msg)
	return m, cmd
}

// View renders the detail panel content (no border). w must match the last SetSize call.
func (m DetailModel) View(w int) string {
	switch m.mode {
	case detailCategory:
		return m.renderCategory(w)
	case detailScript:
		return m.renderScript(w)
	default:
		return styles.DescStyle.Render("Select an item to see details")
	}
}

func (m DetailModel) renderCategory(w int) string {
	if m.catEntry == nil {
		return ""
	}
	e := m.catEntry
	sep := styles.DescStyle.Render(strings.Repeat("─", w))
	title := styles.TitleStyle.Render("Category")
	name := styles.CategoryStyle.Render(e.Name)
	desc := styles.DescStyle.Render(e.Description)

	path := e.ScriptsPath
	if path == "" {
		path = "(none)"
	}

	kv := strings.Join([]string{
		"  " + styles.DescStyle.Render("Source") + "    " + styles.Badge(e.Source),
		"  " + styles.DescStyle.Render("Path") + "      " + styles.DescStyle.Render(path),
		"  " + styles.DescStyle.Render("Scripts") + "   " + styles.DescStyle.Render(fmt.Sprintf("%d", len(m.scripts))),
	}, "\n")

	scriptList := ""
	for _, s := range m.scripts {
		scriptList += "\n    " + styles.ScriptStyle.Render(s.Name)
	}

	return strings.Join([]string{title, sep, name, desc, "", kv, scriptList}, "\n")
}

func (m DetailModel) renderScript(w int) string {
	if m.scrEntry == nil {
		return ""
	}
	e := m.scrEntry
	sep := styles.DescStyle.Render(strings.Repeat("─", w))
	title := styles.TitleStyle.Render("Script")
	name := styles.ScriptStyle.Render(e.Name)
	desc := styles.DescStyle.Render(e.Description)

	path := ""
	executable := "/bin/sh"
	command := ""
	helpFlag := "--help"
	if m.reg != nil {
		path = m.reg.ResolveScriptPath(*e)
		executable = m.reg.ResolveExecutable(*e)
		command = executable + " " + path
		helpFlag = m.reg.ResolveHelpFlag(*e)
	}

	kv := strings.Join([]string{
		"  " + styles.DescStyle.Render("Category") + "    " + styles.CategoryStyle.Render(e.Category),
		"  " + styles.DescStyle.Render("Source") + "      " + styles.Badge(e.Source),
		"  " + styles.DescStyle.Render("Executable") + "  " + styles.DescStyle.Render(executable),
		"  " + styles.DescStyle.Render("Command") + "     " + styles.DescStyle.Render(command),
		"  " + styles.DescStyle.Render("Path") + "        " + styles.DescStyle.Render(path),
		"  " + styles.DescStyle.Render("Help flag") + "   " + styles.DescStyle.Render(helpFlag),
	}, "\n")

	top := strings.Join([]string{title, sep, name, desc, "", kv}, "\n")

	// Help viewport section
	var helpSection string
	if m.helpText != "" && m.helpScript == e.Name {
		helpTitle := styles.TitleStyle.Render("Help")
		helpSep := styles.DescStyle.Render(strings.Repeat("─", w))

		// Scroll position indicator
		scrollInfo := ""
		if m.vp.TotalLineCount() > m.vp.Height {
			pct := int(m.vp.ScrollPercent() * 100)
			scrollInfo = "  " + styles.StatusStyle.Render(fmt.Sprintf("%d%%", pct))
		}

		helpSection = "\n\n" + helpTitle + "\n" + helpSep + "\n" + m.vp.View() + scrollInfo
	} else if m.helpScript == "" {
		helpSection = "\n\n" + styles.DescStyle.Render("  loading help…")
	}

	return top + helpSection
}

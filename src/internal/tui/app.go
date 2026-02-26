// Package tui implements the ShotGum interactive terminal UI using BubbleTea.
// The root AppModel is a state machine with four states:
//
//	stateCategories → stateScripts → stateOutput
//
// A two-panel layout (category/script list on the left, detail on the right) is
// rendered by renderTwoPanel. An optional rolling execution-log footer can be
// toggled with the [l] key.
package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/brunoomariano/ShotGum-Toolchain/internal/registry"
	"github.com/brunoomariano/ShotGum-Toolchain/internal/tui/styles"
	"github.com/brunoomariano/ShotGum-Toolchain/internal/tui/views"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const maxExecutionLogs = 80
const executionLogWindowLines = 6

type state int

const (
	stateCategories state = iota
	stateScripts
	stateConfirm
	stateOutput
)

var (
	// Panel border styles — active panel uses the purple accent; inactive uses subtle.
	panelBorderActive   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(styles.Purple)
	panelBorderInactive = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(styles.Subtle)

	// logFooterStyle is the border style applied to the rolling execution log footer.
	logFooterStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(styles.Subtle).
			Padding(0, 1)
)

// AppModel is the root BubbleTea model (state machine).
type AppModel struct {
	state         state
	reg           *registry.Registry
	width         int
	height        int
	header        views.HeaderModel
	catList       list.Model
	scriptList    list.Model
	detail        views.DetailModel
	confirm       views.ConfirmModel
	output        views.OutputModel
	detailFocused bool // when true, ↑/↓ scroll the detail viewport instead of the list
	// currentCategory holds the active category when navigating into scripts/output
	currentCategory *registry.CategoryEntry
	// currentScript holds the active script when running or confirming
	currentScript *registry.ScriptEntry

	showExecutionLogs bool
	executionLogs     []string
}

// panelDims returns the content widths and panel height for the two-panel layout.
// It uses the rendered header/footer heights to avoid vertical overflow.
func (m AppModel) panelDims() (leftW, rightW, panelH int) {
	leftW = max((m.width-4)*2/5, 10)
	rightW = max((m.width-4)-leftW, 10)

	headerH := lipgloss.Height(m.header.View())
	footerH := 0
	if m.showExecutionLogs {
		footerH = lipgloss.Height(m.renderExecutionLogsFooter())
	}

	// Layout rows:
	// header + panel borders/content + help + optional footer.
	panelH = max(m.height-headerH-footerH-3, 5)
	return
}

// NewAppModel creates the root model loaded from the registry.
func NewAppModel(reg *registry.Registry) (AppModel, error) {
	cats := reg.GetCategories()
	catList := views.NewCategoryList(cats, 80, 20)

	m := AppModel{
		state:      stateCategories,
		reg:        reg,
		header:     views.NewHeaderModel(),
		catList:    catList,
		scriptList: views.NewScriptList("", []registry.ScriptEntry{}, 80, 20),
	}
	m, _ = m.syncDetail() // categories state: no async help cmd needed at startup
	return m, nil
}

// SetExecutionLogsVisible toggles the execution log footer visibility.
func (m *AppModel) SetExecutionLogsVisible(enabled bool) {
	m.showExecutionLogs = enabled
	if enabled {
		m.appendExecutionLog("INFO", "execution log footer enabled")
	}
}

func (m AppModel) Init() tea.Cmd {
	return m.header.Init()
}

// syncDetail refreshes the right-panel detail model to match the currently
// selected list item. For stateScripts it also returns a Cmd that fetches the
// script's --help output asynchronously.
func (m AppModel) syncDetail() (AppModel, tea.Cmd) {
	switch m.state {
	case stateCategories:
		if item, ok := m.catList.SelectedItem().(views.CategoryItem); ok {
			entry := item.Entry()
			m.detail.SetCategory(&entry, m.reg.GetScripts(entry.Name), m.reg)
		}
	case stateScripts:
		if item, ok := m.scriptList.SelectedItem().(views.ScriptItem); ok {
			entry := item.Entry()
			m.detail.SetScript(&entry, m.reg)
			_, rightW, _ := m.panelDims()
			return m, views.LoadScriptHelpCmd(entry, m.reg, rightW)
		}
	}
	return m, nil
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.header, _ = m.header.Update(msg)
		leftW, rightW, panelH := m.panelDims()
		m.catList.SetSize(leftW, panelH)
		m.scriptList.SetSize(leftW, panelH)
		m.detail.SetSize(rightW, panelH)
		if m.state == stateOutput {
			updated, cmd := m.output.Update(msg)
			m.output = updated
			return m, cmd
		}
		return m, nil

	case views.DetailHelpMsg:
		m.detail.SetHelpText(msg.ScriptName, msg.Text)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "l", "L":
			m.showExecutionLogs = !m.showExecutionLogs
			if m.showExecutionLogs {
				m.appendExecutionLog("INFO", "execution log footer enabled")
			}
			return m, nil
		}

		switch m.state {
		case stateCategories:
			return m.updateCategories(msg)
		case stateScripts:
			return m.updateScripts(msg)
		case stateOutput:
			return m.updateOutput(msg)
		}

	case views.ScriptDoneMsg:
		updated, cmd := m.output.Update(msg)
		m.output = updated
		if msg.Err != nil {
			m.appendExecutionLog("ERROR", fmt.Sprintf("script %q finished with error: %v", m.output.ScriptName(), msg.Err))
		} else {
			m.appendExecutionLog("INFO", fmt.Sprintf("script %q finished successfully", m.output.ScriptName()))
		}
		return m, cmd
	}

	// Delegate non-key messages to the active list
	switch m.state {
	case stateCategories:
		updated, cmd := m.catList.Update(msg)
		m.catList = updated
		return m, cmd
	case stateScripts:
		updated, cmd := m.scriptList.Update(msg)
		m.scriptList = updated
		return m, cmd
	case stateOutput:
		updated, cmd := m.output.Update(msg)
		m.output = updated
		return m, cmd
	}

	return m, nil
}

func (m AppModel) updateCategories(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "enter":
		item, ok := m.catList.SelectedItem().(views.CategoryItem)
		if !ok {
			break
		}
		entry := item.Entry()
		m.currentCategory = &entry
		m.appendExecutionLog("INFO", fmt.Sprintf("opened category %q", entry.Name))
		scripts := m.reg.GetScripts(entry.Name)
		leftW, _, panelH := m.panelDims()
		m.scriptList = views.NewScriptList(entry.Name, scripts, leftW, panelH)
		m.state = stateScripts
		var helpCmd tea.Cmd
		m, helpCmd = m.syncDetail()
		return m, helpCmd
	}

	updated, cmd := m.catList.Update(msg)
	m.catList = updated
	// Auto-skip section header items so the cursor always lands on a real entry.
	if _, ok := m.catList.SelectedItem().(views.SectionHeaderItem); ok {
		switch msg.String() {
		case "up", "k":
			m.catList.CursorUp()
		default:
			m.catList.CursorDown()
		}
	}
	m, _ = m.syncDetail() // categories: no help cmd
	return m, cmd
}

func (m AppModel) updateScripts(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// When the detail panel has focus, scroll keys go to its viewport.
	if m.detailFocused {
		switch msg.String() {
		case "tab", "esc":
			m.detailFocused = false
			return m, nil
		case "q", "ctrl+c":
			return m, tea.Quit
		}
		updated, cmd := m.detail.Update(msg)
		m.detail = updated
		return m, cmd
	}

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "tab":
		m.detailFocused = true
		return m, nil

	case "esc":
		m.state = stateCategories
		m.detailFocused = false
		m.appendExecutionLog("DEBUG", "returned to categories")
		m, _ = m.syncDetail() // back to categories: no help cmd
		return m, nil

	case "enter", "i":
		item, ok := m.scriptList.SelectedItem().(views.ScriptItem)
		if !ok {
			break
		}
		return m.runScript(item.Entry(), nil)

	case "?":
		item, ok := m.scriptList.SelectedItem().(views.ScriptItem)
		if !ok {
			break
		}
		entry := item.Entry()
		helpFlag := m.reg.ResolveHelpFlag(entry)
		return m.runScript(entry, []string{helpFlag})
	}

	updated, cmd := m.scriptList.Update(msg)
	m.scriptList = updated
	// Auto-skip section header items so the cursor always lands on a real entry.
	if _, ok := m.scriptList.SelectedItem().(views.SectionHeaderItem); ok {
		switch msg.String() {
		case "up", "k":
			m.scriptList.CursorUp()
		default:
			m.scriptList.CursorDown()
		}
	}
	var helpCmd tea.Cmd
	m, helpCmd = m.syncDetail()
	return m, tea.Batch(cmd, helpCmd)
}

func (m AppModel) runScript(entry registry.ScriptEntry, args []string) (tea.Model, tea.Cmd) {
	m.currentScript = &entry
	if len(args) == 0 {
		m.appendExecutionLog("INFO", fmt.Sprintf("starting script %q", entry.Name))
	} else {
		m.appendExecutionLog("INFO", fmt.Sprintf("starting script %q with args %q", entry.Name, strings.Join(args, " ")))
	}

	path := m.reg.ResolveScriptPath(entry)
	executable := m.reg.ResolveExecutable(entry)
	cmd := exec.Command(executable, append([]string{path}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	execCmd := tea.ExecProcess(cmd, func(err error) tea.Msg {
		_ = err
		return tea.QuitMsg{}
	})

	return m, execCmd
}

func (m AppModel) updateOutput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		if m.output.IsLoading() {
			return m, nil // don't navigate away while loading
		}
		m.state = stateScripts
		m.appendExecutionLog("DEBUG", "returned to scripts from output")
		return m, nil

	case "q", "ctrl+c":
		return m, tea.Quit
	}

	updated, cmd := m.output.Update(msg)
	m.output = updated
	return m, cmd
}

func (m *AppModel) appendExecutionLog(level, text string) {
	if strings.TrimSpace(text) == "" {
		return
	}
	ts := time.Now().Format("15:04:05.000")
	entry := fmt.Sprintf("%s [%s] %s", ts, strings.ToUpper(level), strings.TrimSpace(text))
	m.executionLogs = append(m.executionLogs, entry)
	if len(m.executionLogs) > maxExecutionLogs {
		m.executionLogs = m.executionLogs[len(m.executionLogs)-maxExecutionLogs:]
	}
}

func (m AppModel) renderExecutionLogsFooter() string {
	if !m.showExecutionLogs {
		return ""
	}

	lines := m.executionLogs
	if len(lines) == 0 {
		lines = []string{"no execution events yet"}
	}
	if len(lines) > executionLogWindowLines {
		lines = lines[len(lines)-executionLogWindowLines:]
	}
	rendered := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.Contains(line, "[ERROR]") {
			rendered = append(rendered, styles.ErrorStyle.Render(line))
			continue
		}
		rendered = append(rendered, styles.DescStyle.Render(line))
	}

	title := styles.StatusStyle.Render("execution logs (rolling)")
	return logFooterStyle.MaxWidth(max(m.width-2, 0)).Render(title + "\n" + strings.Join(rendered, "\n"))
}

// renderTwoPanel assembles the header + side-by-side panels + help bar.
func (m AppModel) renderTwoPanel(leftContent, helpText string) string {
	leftW, rightW, panelH := m.panelDims()

	leftBorder, rightBorder := panelBorderActive, panelBorderInactive
	if m.detailFocused {
		leftBorder, rightBorder = panelBorderInactive, panelBorderActive
	}

	leftPanel := leftBorder.Width(leftW).Height(panelH).Render(leftContent)
	rightPanel := rightBorder.Width(rightW).Height(panelH).Render(m.detail.View(rightW))

	panels := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
	help := styles.HelpStyle.Render(helpText)
	footer := m.renderExecutionLogsFooter()
	if footer == "" {
		return m.header.View() + "\n" + panels + "\n" + help
	}
	return m.header.View() + "\n" + panels + "\n" + help + "\n" + footer
}

func (m AppModel) View() string {
	switch m.state {
	case stateCategories:
		return m.renderTwoPanel(
			m.catList.View(),
			"enter: open  •  /: filter  •  l: logs  •  q: quit",
		)
	case stateScripts:
		helpText := "enter: run  •  ?: help  •  tab: scroll help  •  l: logs  •  esc: back  •  q: quit"
		if m.detailFocused {
			helpText = "↑/↓: scroll  •  tab/esc: back to list  •  l: logs"
		}
		return m.renderTwoPanel(m.scriptList.View(), helpText)
	case stateOutput:
		base := m.output.View()
		footer := m.renderExecutionLogsFooter()
		if footer == "" {
			return base
		}
		return base + "\n" + footer
	}
	return ""
}

// Start launches the full TUI application.
func Start(reg *registry.Registry) error {
	return StartWithLogs(reg, false)
}

// StartWithLogs launches the full TUI application with optional execution logs footer.
func StartWithLogs(reg *registry.Registry, showExecutionLogs bool) error {
	model, err := NewAppModel(reg)
	if err != nil {
		return fmt.Errorf("creating TUI: %w", err)
	}
	model.SetExecutionLogsVisible(showExecutionLogs)

	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

// StartAtCategory launches the TUI pre-navigated to a specific category.
func StartAtCategory(reg *registry.Registry, categoryName string) error {
	return StartAtCategoryWithLogs(reg, categoryName, false)
}

// StartAtCategoryWithLogs launches the TUI pre-navigated to a specific category.
func StartAtCategoryWithLogs(reg *registry.Registry, categoryName string, showExecutionLogs bool) error {
	cats := reg.GetCategories()
	var found *registry.CategoryEntry
	for _, c := range cats {
		if c.Name == categoryName {
			cp := c
			found = &cp
			break
		}
	}

	model, err := NewAppModel(reg)
	if err != nil {
		return err
	}
	model.SetExecutionLogsVisible(showExecutionLogs)

	if found != nil {
		scripts := reg.GetScripts(found.Name)
		model.currentCategory = found
		model.scriptList = views.NewScriptList(found.Name, scripts, 80, 20)
		model.state = stateScripts
		model, _ = model.syncDetail()
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

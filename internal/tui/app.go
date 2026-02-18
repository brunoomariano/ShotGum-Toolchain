package tui

import (
	"fmt"
	"os/exec"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shotgum/stg/internal/registry"
	"github.com/shotgum/stg/internal/tui/styles"
	"github.com/shotgum/stg/internal/tui/views"
	"github.com/shotgum/stg/internal/version"
)

const appIssues = "github.com/brunoomariano/ShotGum-Toolchain/issues"

type state int

const (
	stateCategories state = iota
	stateScripts
	stateConfirm
	stateOutput
)

// interactiveDoneMsg is sent when a tea.ExecProcess script finishes.
type interactiveDoneMsg struct{ err error }

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
}

// panelDims returns the content widths and panel height for the two-panel layout.
// Header occupies 3 rows (top border + content + bottom border).
// Help bar occupies 1 row. Panel border adds 2 rows each.
func (m AppModel) panelDims() (leftW, rightW, panelH int) {
	leftW = max((m.width-4)*2/5, 10)
	rightW = max((m.width-4)-leftW, 10)
	panelH = max(m.height-6, 5) // 3 header + 2 panel borders + 1 help bar
	return
}

// NewAppModel creates the root model loaded from the registry.
func NewAppModel(reg *registry.Registry) (AppModel, error) {
	cats := reg.GetCategories()
	catList := views.NewCategoryList(cats, 80, 20)

	m := AppModel{
		state:      stateCategories,
		reg:        reg,
		header:     views.NewHeaderModel(version.Version, appIssues),
		catList:    catList,
		scriptList: views.NewScriptList("", []registry.ScriptEntry{}, 80, 20),
	}
	m, _ = m.syncDetail() // categories state: no async help cmd needed at startup
	return m, nil
}

func (m AppModel) Init() tea.Cmd {
	return m.header.Init()
}

// syncDetail refreshes the right-panel detail. For script state it also
// returns a Cmd that loads the help text asynchronously.
func (m AppModel) syncDetail() (AppModel, tea.Cmd) {
	switch m.state {
	case stateCategories:
		item, ok := m.catList.SelectedItem().(views.CategoryItem)
		if ok {
			entry := item.Entry()
			scripts := m.reg.GetScripts(entry.Name)
			m.detail.SetCategory(&entry, scripts, m.reg)
		}
		return m, nil
	case stateScripts, stateConfirm:
		item, ok := m.scriptList.SelectedItem().(views.ScriptItem)
		if ok {
			entry := item.Entry()
			m.detail.SetScript(&entry, m.reg)
			_, rightW, _ := m.panelDims()
			return m, views.LoadScriptHelpCmd(entry, m.reg, rightW)
		}
		return m, nil
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
		m.confirm, _ = m.confirm.Update(msg)
		if m.state == stateOutput {
			updated, cmd := m.output.Update(msg)
			m.output = updated
			return m, cmd
		}
		return m, nil

	case interactiveDoneMsg:
		// Script ran interactively; TUI is already restored — go back to scripts.
		m.state = stateScripts
		return m, nil

	case views.DetailHelpMsg:
		m.detail.SetHelpText(msg.ScriptName, msg.Text)
		return m, nil

	case views.ConfirmRunMsg:
		if m.currentScript != nil {
			return m.runScript(*m.currentScript, msg.ExtraArgs)
		}
		return m, nil

	case views.ConfirmCancelMsg:
		m.state = stateScripts
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case stateCategories:
			return m.updateCategories(msg)
		case stateScripts:
			return m.updateScripts(msg)
		case stateConfirm:
			return m.updateConfirm(msg)
		case stateOutput:
			return m.updateOutput(msg)
		}

	case views.ScriptDoneMsg:
		updated, cmd := m.output.Update(msg)
		m.output = updated
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
		m, _ = m.syncDetail() // back to categories: no help cmd
		return m, nil

	case "enter":
		item, ok := m.scriptList.SelectedItem().(views.ScriptItem)
		if !ok {
			break
		}
		entry := item.Entry()
		m.currentScript = &entry
		m.confirm = views.NewConfirmModel(entry, m.reg, m.width, m.height)
		m.state = stateConfirm
		return m, nil

	case "i":
		item, ok := m.scriptList.SelectedItem().(views.ScriptItem)
		if !ok {
			break
		}
		entry := item.Entry()
		return m.runScriptInteractive(entry, nil)

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
	var helpCmd tea.Cmd
	m, helpCmd = m.syncDetail()
	return m, tea.Batch(cmd, helpCmd)
}

func (m AppModel) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	updated, cmd := m.confirm.Update(msg)
	m.confirm = updated
	return m, cmd
}

// runScriptInteractive suspends the TUI and hands the terminal directly to the
// script process, allowing full interactivity (gum prompts, spinners, etc.).
// The TUI is automatically restored when the process exits.
func (m AppModel) runScriptInteractive(entry registry.ScriptEntry, args []string) (tea.Model, tea.Cmd) {
	m.currentScript = &entry
	path := m.reg.ResolveScriptPath(entry)

	var cmd *exec.Cmd
	switch entry.Type {
	case "executable":
		cmd = exec.Command(path, args...)
	default:
		cmd = exec.Command("bash", append([]string{path}, args...)...)
	}

	return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
		return interactiveDoneMsg{err: err}
	})
}

func (m AppModel) runScript(entry registry.ScriptEntry, args []string) (tea.Model, tea.Cmd) {
	m.currentScript = &entry
	m.output = views.NewOutputModel(entry.Name, m.width, m.height)
	m.state = stateOutput
	runCmd := views.RunScriptCmd(entry, args, m.reg)
	initCmd := m.output.Init()
	return m, tea.Batch(initCmd, runCmd)
}

func (m AppModel) updateOutput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		if m.output.IsLoading() {
			return m, nil // don't navigate away while loading
		}
		m.state = stateScripts
		return m, nil

	case "ctrl+c":
		return m, tea.Quit
	}

	updated, cmd := m.output.Update(msg)
	m.output = updated
	return m, cmd
}

// renderTwoPanel assembles the header + side-by-side panels + help bar.
func (m AppModel) renderTwoPanel(leftContent, helpText string) string {
	leftW, rightW, panelH := m.panelDims()

	inactiveStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Subtle)
	activeStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Purple)

	var leftStyle, rightStyle lipgloss.Style
	if m.detailFocused {
		leftStyle = inactiveStyle
		rightStyle = activeStyle
	} else {
		leftStyle = activeStyle
		rightStyle = inactiveStyle
	}

	leftPanel := leftStyle.Width(leftW).Height(panelH).Render(leftContent)
	rightPanel := rightStyle.Width(rightW).Height(panelH).Render(m.detail.View(rightW, panelH))

	panels := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
	help := styles.HelpStyle.Render(helpText)

	return m.header.View() + "\n" + panels + "\n" + help
}

func (m AppModel) View() string {
	switch m.state {
	case stateCategories:
		return m.renderTwoPanel(
			m.catList.View(),
			"enter: open  •  /: filter  •  q: quit",
		)
	case stateScripts:
		helpText := "enter: run  •  i: interactive  •  ?: help  •  tab: scroll help  •  esc: back  •  q: quit"
		if m.detailFocused {
			helpText = "↑/↓: scroll  •  tab/esc: back to list"
		}
		return m.renderTwoPanel(m.scriptList.View(), helpText)
	case stateConfirm:
		return m.confirm.View()
	case stateOutput:
		return m.output.View()
	}
	return ""
}

// Start launches the full TUI application.
func Start(reg *registry.Registry) error {
	model, err := NewAppModel(reg)
	if err != nil {
		return fmt.Errorf("creating TUI: %w", err)
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

// StartAtCategory launches the TUI pre-navigated to a specific category.
func StartAtCategory(reg *registry.Registry, categoryName string) error {
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

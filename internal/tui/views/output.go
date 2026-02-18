package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shotgum/stg/internal/registry"
	"github.com/shotgum/stg/internal/runner"
	"github.com/shotgum/stg/internal/tui/styles"
)

// ScriptDoneMsg is sent when the script finishes running.
type ScriptDoneMsg struct {
	Output string
	Err    error
}

// RunScriptCmd returns a Cmd that runs a script and captures its output.
func RunScriptCmd(entry registry.ScriptEntry, args []string, reg *registry.Registry) tea.Cmd {
	return func() tea.Msg {
		out, err := runner.CaptureRun(entry, args, reg)
		return ScriptDoneMsg{Output: out, Err: err}
	}
}

// OutputModel is the scrollable script output view.
type OutputModel struct {
	viewport   viewport.Model
	spinner    spinner.Model
	loading    bool
	scriptName string
	output     string
	err        error
}

// NewOutputModel creates a new output view for a script.
func NewOutputModel(scriptName string, width, height int) OutputModel {
	vp := viewport.New(width, height-4)
	vp.SetContent("")

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = styles.TitleStyle

	return OutputModel{
		viewport:   vp,
		spinner:    sp,
		loading:    true,
		scriptName: scriptName,
	}
}

func (m OutputModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m OutputModel) Update(msg tea.Msg) (OutputModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case ScriptDoneMsg:
		m.loading = false
		m.output = msg.Output
		m.err = msg.Err
		content := m.output
		if m.err != nil {
			content += "\n" + styles.ErrorStyle.Render(fmt.Sprintf("[error] %v", m.err))
		}
		m.viewport.SetContent(content)
		m.viewport.GotoBottom()

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 6
	}

	var vpCmd tea.Cmd
	m.viewport, vpCmd = m.viewport.Update(msg)
	cmds = append(cmds, vpCmd)

	return m, tea.Batch(cmds...)
}

func (m OutputModel) View() string {
	title := styles.TitleStyle.Render("ShotGum") +
		styles.DescStyle.Render("  — ") +
		styles.CategoryStyle.Render(m.scriptName)

	var body string
	if m.loading {
		body = fmt.Sprintf("\n  %s Running %s...\n", m.spinner.View(), m.scriptName)
	} else {
		scrollPct := ""
		if m.viewport.TotalLineCount() > 0 {
			pct := int(m.viewport.ScrollPercent() * 100)
			scrollPct = styles.StatusStyle.Render(fmt.Sprintf("%d%%", pct))
		}
		body = m.viewport.View() + "\n" + scrollPct
	}

	help := styles.HelpStyle.Render("↑/↓: scroll  •  pgup/pgdn: page  •  esc/q: back")
	sep := strings.Repeat("─", m.viewport.Width)

	return styles.BorderBox.Render(
		title+"\n"+styles.DescStyle.Render(sep)+"\n"+body,
	) + "\n" + help
}

// IsLoading returns true while the script is still running.
func (m OutputModel) IsLoading() bool {
	return m.loading
}

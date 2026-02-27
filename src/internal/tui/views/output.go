// Package views contains the BubbleTea sub-models that render individual UI
// containers: Header Container, Navigation Container (categories/scripts),
// Info Container (detail/help), confirm dialog, and output.
package views

import (
	"fmt"
	"strings"

	"github.com/brunoomariano/ShotGum-Toolchain/internal/registry"
	"github.com/brunoomariano/ShotGum-Toolchain/internal/runner"
	"github.com/brunoomariano/ShotGum-Toolchain/internal/tui/styles"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
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
	viewport    viewport.Model
	spinner     spinner.Model
	loading     bool
	interactive bool
	scriptName  string
	output      string
	err         error
}

// NewOutputModel creates a new output view for a script.
func NewOutputModel(scriptName string, width, height int) OutputModel {
	vp := viewport.New(width, height-4)
	vp.SetContent("")

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = styles.TitleStyle

	return OutputModel{
		viewport:    vp,
		spinner:     sp,
		loading:     true,
		interactive: false,
		scriptName:  scriptName,
	}
}

// EnableInteractive marks this output as interactive streaming mode.
func (m *OutputModel) EnableInteractive() {
	m.interactive = true
}

// AppendChunk appends streamed output and keeps the viewport pinned to bottom.
func (m *OutputModel) AppendChunk(chunk string) {
	m.output += chunk
	m.viewport.SetContent(m.output)
	m.viewport.GotoBottom()
}

// Finish marks execution as completed and preserves rendered output/logs.
func (m *OutputModel) Finish(err error) {
	m.loading = false
	m.err = err
	content := m.output
	if m.err != nil {
		content += "\n" + styles.ErrorStyle.Render(fmt.Sprintf("[error] %v", m.err))
	}
	m.viewport.SetContent(content)
	m.viewport.GotoBottom()
}

func (m OutputModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m OutputModel) Update(msg tea.Msg) (OutputModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case ScriptDoneMsg:
		m.output = msg.Output
		m.Finish(msg.Err)

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

// scrollIndicator returns a styled scroll-percentage string, or "" if there is
// no content to scroll.
func (m OutputModel) scrollIndicator() string {
	if m.viewport.TotalLineCount() == 0 {
		return ""
	}
	return styles.StatusStyle.Render(fmt.Sprintf("%d%%", int(m.viewport.ScrollPercent()*100)))
}

func (m OutputModel) View() string {
	title := styles.TitleStyle.Render("ShotGum") +
		styles.DescStyle.Render("  — ") +
		styles.CategoryStyle.Render(m.scriptName)

	// Show the spinner while waiting for output to begin; once there is content
	// (or when execution is done) display the scrollable viewport instead.
	showSpinner := m.loading && (!m.interactive || strings.TrimSpace(m.output) == "")
	var body string
	if showSpinner {
		body = fmt.Sprintf("\n  %s Running %s...\n", m.spinner.View(), m.scriptName)
	} else {
		body = m.viewport.View() + "\n" + m.scrollIndicator()
	}

	helpText := "↑/↓: scroll  •  pgup/pgdn: page  •  esc: back  •  q: quit"
	if m.loading && m.interactive {
		helpText = "type to interact  •  enter/tab/arrows supported  •  esc: disabled while running  •  ctrl+c: interrupt"
	}

	sep := strings.Repeat("─", m.viewport.Width)
	return styles.BorderBox.Render(
		title+"\n"+styles.DescStyle.Render(sep)+"\n"+body,
	) + "\n" + styles.HelpStyle.Render(helpText)
}

// IsLoading returns true while the script is still running.
func (m OutputModel) IsLoading() bool {
	return m.loading
}

// ScriptName returns the script currently bound to this output view.
func (m OutputModel) ScriptName() string {
	return m.scriptName
}

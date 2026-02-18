package views

import (
	"fmt"
	"strings"

	"github.com/brunoomariano/ShotGum-Toolchain/internal/registry"
	"github.com/brunoomariano/ShotGum-Toolchain/internal/tui/styles"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfirmRunMsg is emitted when the user confirms script execution.
// ExtraArgs holds any additional arguments the user typed.
type ConfirmRunMsg struct {
	ExtraArgs []string
}

// ConfirmCancelMsg is emitted when the user cancels the dialog.
type ConfirmCancelMsg struct{}

// ConfirmModel is the confirmation dialog overlay.
type ConfirmModel struct {
	entry        registry.ScriptEntry
	resolvedPath string
	width        int
	height       int
	argsInput    textinput.Model
	inputActive  bool
	helpText     string
	helpVp       viewport.Model
}

// NewConfirmModel creates a new confirmation dialog for a script.
// helpText is the pre-loaded --help output for the script (may be empty).
func NewConfirmModel(entry registry.ScriptEntry, reg *registry.Registry, w, h int, helpText string) ConfirmModel {
	path := ""
	if reg != nil {
		path = reg.ResolveScriptPath(entry)
	}

	ti := textinput.New()
	ti.Placeholder = "--flag value  --other …"
	ti.CharLimit = 256
	ti.PromptStyle = lipgloss.NewStyle().Foreground(styles.Purple)
	ti.TextStyle = lipgloss.NewStyle().Foreground(styles.White)

	vp := viewport.New(0, 0)
	if helpText != "" {
		vp.SetContent(helpText)
	}

	return ConfirmModel{
		entry:        entry,
		resolvedPath: path,
		width:        w,
		height:       h,
		argsInput:    ti,
		helpText:     helpText,
		helpVp:       vp,
	}
}

// SetHelpText updates the cached help text (called when async load completes).
func (m *ConfirmModel) SetHelpText(text string) {
	m.helpText = text
	m.helpVp.SetContent(text)
	m.helpVp.GotoTop()
}

func (m ConfirmModel) Update(msg tea.Msg) (ConfirmModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		if m.inputActive {
			switch msg.String() {
			case "enter":
				args := strings.Fields(m.argsInput.Value())
				return m, func() tea.Msg { return ConfirmRunMsg{ExtraArgs: args} }
			case "esc":
				m.inputActive = false
				m.argsInput.Blur()
				return m, nil
			case "up", "down", "pgup", "pgdown":
				// Scroll the help viewport while typing
				var cmd tea.Cmd
				m.helpVp, cmd = m.helpVp.Update(msg)
				return m, cmd
			}
			var cmd tea.Cmd
			m.argsInput, cmd = m.argsInput.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "enter":
			return m, func() tea.Msg { return ConfirmRunMsg{} }
		case "a":
			m.inputActive = true
			return m, m.argsInput.Focus()
		case "esc", "q":
			return m, func() tea.Msg { return ConfirmCancelMsg{} }
		}
	}
	return m, nil
}

func (m ConfirmModel) View() string {
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Purple).
		Padding(1, 2)

	label := styles.TitleStyle.Render("Execute script?")
	scriptRef := styles.CategoryStyle.Render(m.entry.Category) +
		styles.DescStyle.Render(" › ") +
		styles.ScriptStyle.Render(m.entry.Name)
	path := styles.DescStyle.Render(m.resolvedPath)

	boxWidth := 50
	if m.width > 0 {
		candidate := m.width - 12
		if candidate > boxWidth {
			boxWidth = candidate
		}
		if boxWidth > 64 {
			boxWidth = 64
		}
	}

	// Args input section
	var argsSection string
	if m.inputActive {
		inputStyle := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(styles.Purple).
			Width(boxWidth - 2)
		m.argsInput.Width = boxWidth - 6
		argsSection = "\n" + styles.DescStyle.Render("Extra args:") + "\n" +
			inputStyle.Render(m.argsInput.View())
	} else {
		argsSection = "\n" + styles.HelpStyle.Render("[a] to add arguments before running")
	}

	// Key hints
	var hint string
	if m.inputActive {
		hint = styles.HelpStyle.Render("[Enter] Run with args  •  [Esc] Back  •  [↑/↓] Scroll help")
	} else {
		hint = styles.HelpStyle.Render("[Enter] Run  •  [a] Add args  •  [Esc] Cancel")
	}

	inner := fmt.Sprintf("%s\n\n%s\n%s%s\n\n%s", label, scriptRef, path, argsSection, hint)
	box := dialogStyle.Width(boxWidth).Render(inner)

	if m.width == 0 || m.height == 0 {
		return box
	}

	// When inputActive, render a help panel below the dialog.
	if m.inputActive {
		helpPanel := m.renderHelpPanel(boxWidth)
		combined := lipgloss.JoinVertical(lipgloss.Left, box, helpPanel)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, combined)
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

// renderHelpPanel builds the help container shown below the confirm dialog.
func (m ConfirmModel) renderHelpPanel(width int) string {
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Teal).
		Padding(0, 1)

	// Estimate dialog height: ~10 lines content + 2 border + 2 padding = 14
	const dialogH = 14
	helpH := m.height - dialogH - 6
	if helpH < 3 {
		helpH = 3
	}

	var content string
	if m.helpText == "" {
		content = styles.DescStyle.Render("  loading help…")
	} else {
		m.helpVp.Width = width - 2
		m.helpVp.Height = helpH

		scrollInfo := ""
		if m.helpVp.TotalLineCount() > m.helpVp.Height {
			pct := int(m.helpVp.ScrollPercent() * 100)
			scrollInfo = "  " + styles.StatusStyle.Render(fmt.Sprintf("%d%%", pct))
		}
		content = m.helpVp.View() + scrollInfo
	}

	title := styles.TitleStyle.Render("Script help")
	sep := styles.DescStyle.Render(strings.Repeat("─", width-4))
	inner := title + "\n" + sep + "\n" + content

	return panelStyle.Width(width).Render(inner)
}

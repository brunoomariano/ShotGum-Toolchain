package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shotgum/stg/internal/registry"
	"github.com/shotgum/stg/internal/tui/styles"
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
}

// NewConfirmModel creates a new confirmation dialog for a script.
func NewConfirmModel(entry registry.ScriptEntry, reg *registry.Registry, w, h int) ConfirmModel {
	path := ""
	if reg != nil {
		path = reg.ResolveScriptPath(entry)
	}

	ti := textinput.New()
	ti.Placeholder = "--flag value  --other …"
	ti.CharLimit = 256
	ti.PromptStyle = lipgloss.NewStyle().Foreground(styles.Purple)
	ti.TextStyle = lipgloss.NewStyle().Foreground(styles.White)

	return ConfirmModel{
		entry:        entry,
		resolvedPath: path,
		width:        w,
		height:       h,
		argsInput:    ti,
	}
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
		hint = styles.HelpStyle.Render("[Enter] Run with args  •  [Esc] Back")
	} else {
		hint = styles.HelpStyle.Render("[Enter] Run  •  [a] Add args  •  [Esc] Cancel")
	}

	inner := fmt.Sprintf("%s\n\n%s\n%s%s\n\n%s", label, scriptRef, path, argsSection, hint)
	box := dialogStyle.Width(boxWidth).Render(inner)

	if m.width == 0 || m.height == 0 {
		return box
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

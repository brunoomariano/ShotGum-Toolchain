package views

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shotgum/stg/internal/tui/styles"
)

// Bubble cluster: four circles of ascending size, suggesting floating soap/gum bubbles.
//
//	· ° o O
const bubbleIcon = "· ° o O"

// Shotgun: side-view barrel with stock and muzzle direction.
//
//	¬══════►
const shotgunIcon = "¬══════►"

// HeaderModel renders the top banner with project metadata.
type HeaderModel struct {
	width   int
	version string
	issues  string
}

// NewHeaderModel creates the header with project metadata.
func NewHeaderModel(version, issues string) HeaderModel {
	return HeaderModel{version: version, issues: issues}
}

func (m HeaderModel) Init() tea.Cmd { return nil }

func (m HeaderModel) Update(msg tea.Msg) (HeaderModel, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = msg.Width
	}
	return m, nil
}

func (m HeaderModel) View() string {
	bubble := styles.TitleStyle.Render(bubbleIcon)
	shotgun := styles.TitleStyle.Render(shotgunIcon)
	title := lipgloss.NewStyle().Bold(true).Foreground(styles.Purple).Render("SHOTGUM")

	left := bubble + "  " + title + "  " + shotgun

	version := styles.DescStyle.Render("v" + m.version)
	sep := styles.DescStyle.Render("  ·  ")
	issues := lipgloss.NewStyle().Foreground(styles.Teal).Underline(true).Render(m.issues)
	right := version + sep + issues

	innerW := m.width - 2
	if innerW < 20 {
		innerW = 20
	}

	gap := innerW - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}

	content := left + strings.Repeat(" ", gap) + right

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Purple).
		Width(innerW).
		Render(content)
}

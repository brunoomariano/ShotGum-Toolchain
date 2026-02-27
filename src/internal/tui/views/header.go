package views

import (
	"strings"

	"github.com/brunoomariano/ShotGum-Toolchain/internal/tui/styles"
	"github.com/brunoomariano/ShotGum-Toolchain/internal/version"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Bubble cluster: four circles of ascending size, suggesting floating soap/gum bubbles.
//
//	· ° o O
const bubbleIcon = "· ° o O"

// HeaderModel renders the Header Container (app title/subtitle/version).
type HeaderModel struct {
	width int
}

// NewHeaderModel creates the header with project metadata.
func NewHeaderModel() HeaderModel {
	return HeaderModel{}
}

func (m HeaderModel) Init() tea.Cmd { return nil }

func (m *HeaderModel) SetWidth(w int) {
	m.width = w
}

func (m HeaderModel) Update(msg tea.Msg) (HeaderModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
	}
	return m, nil
}

// SHOTGUM rendered as a 3-row box-drawing ASCII art banner.
// Each letter is 3 columns wide; letters are separated by a single space.
// Total width: 7 letters × 3 + 6 separators = 27 columns.
const (
	artRow1 = "╔═╗ ╦ ╦ ╔═╗ ╔╦╗ ╔═╗ ╦ ╦ ╔╦╗"
	artRow2 = "╚═╗ ╠═╣ ║ ║  ║  ║═╗ ║ ║ ║║║"
	artRow3 = "╚═╝ ╩ ╩ ╚═╝  ╩  ╚═╝ ╚═╝ ╩ ╩"
)

func (m HeaderModel) View() string {
	artStyle := styles.TitleStyle // bold + purple

	versionText := "v" + strings.TrimPrefix(version.Version, "v")
	line1 := artStyle.Render(artRow1)
	line2 := artStyle.Render(artRow2)
	line3 := artStyle.Render(artRow3)

	bubble := styles.TitleStyle.Render(bubbleIcon) // "· ° o O"  7 cols

	ver := styles.DescStyle.Render(versionText)

	subtitle := lipgloss.NewStyle().Bold(true).Foreground(styles.Teal).Render("Script Manager")
	subtitleLine := subtitle + styles.DescStyle.Render("  ·  ") + ver

	innerW := m.width - 2
	if innerW < 10 {
		innerW = 10
	}

	leftRow1 := bubble + "  " + line1
	pad := strings.Repeat(" ", lipgloss.Width(bubble)+2)

	var rows []string

	// Narrow terminals: avoid ASCII-art overflow and keep essential metadata visible.
	if innerW < lipgloss.Width(leftRow1) {
		title := styles.TitleStyle.Render("ShotGum")
		rows = append(rows, title, subtitleLine)
	} else {
		// Full header: keep banner; version sits next to the subtitle line.
		rows = append(rows, leftRow1)
		rows = append(rows, pad+line2, pad+line3, pad+subtitleLine)
	}
	content := strings.Join(rows, "\n")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Subtle).
		Width(innerW).
		Render(content)
}

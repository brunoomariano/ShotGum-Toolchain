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

// HeaderModel renders the top banner with project metadata.
type HeaderModel struct {
	width int
}

// NewHeaderModel creates the header with project metadata.
func NewHeaderModel() HeaderModel {
	return HeaderModel{}
}

func (m HeaderModel) Init() tea.Cmd { return nil }

func (m HeaderModel) Update(msg tea.Msg) (HeaderModel, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
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

	line1 := artStyle.Render(artRow1)
	line2 := artStyle.Render(artRow2)
	line3 := artStyle.Render(artRow3)

	bubble := styles.TitleStyle.Render(bubbleIcon) // "· ° o O"  7 cols

	ver := styles.DescStyle.Render("v" + strings.TrimPrefix(version.Version, "v"))
	sep := styles.DescStyle.Render("  ·  ")
	repoLink := lipgloss.NewStyle().Foreground(styles.Teal).Underline(true).Render(version.RepoURL)
	rightInfo := ver + sep + repoLink

	subtitle := lipgloss.NewStyle().Bold(true).Foreground(styles.Teal).Render("Script Manager")

	innerW := m.width - 2
	if innerW < 20 {
		innerW = 20
	}

	leftRow1 := bubble + "  " + line1
	pad := strings.Repeat(" ", lipgloss.Width(bubble)+2)

	var rows []string

	// Narrow terminals: avoid ASCII-art overflow and keep essential metadata visible.
	if innerW < lipgloss.Width(leftRow1) {
		title := styles.TitleStyle.Render("ShotGum")
		rows = append(rows, title, subtitle, ver)
	} else {
		// Full header: keep banner; include repo link only when there is room.
		gap1 := innerW - lipgloss.Width(leftRow1) - lipgloss.Width(rightInfo)
		if gap1 >= 1 {
			rows = append(rows, leftRow1+strings.Repeat(" ", gap1)+rightInfo)
		} else {
			rows = append(rows, leftRow1)
			rows = append(rows, pad+ver)
		}
		rows = append(rows, pad+line2, pad+line3, pad+subtitle)
	}
	content := strings.Join(rows, "\n")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Purple).
		Width(innerW).
		Render(content)
}

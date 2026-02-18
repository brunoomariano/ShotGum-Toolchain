package styles

import "github.com/charmbracelet/lipgloss"

var (
	// Palette
	Purple = lipgloss.Color("#C084FC")
	Teal   = lipgloss.Color("#5EEAD4")
	Gray   = lipgloss.Color("#6B7280")
	Subtle = lipgloss.Color("#374151")
	White  = lipgloss.Color("#F9FAFB")
	Red    = lipgloss.Color("#F87171")

	TitleStyle    = lipgloss.NewStyle().Bold(true).Foreground(Purple)
	CategoryStyle = lipgloss.NewStyle().Foreground(Teal).Bold(true)
	ScriptStyle   = lipgloss.NewStyle().Foreground(White)
	DescStyle     = lipgloss.NewStyle().Foreground(Gray)
	LocalBadge    = lipgloss.NewStyle().Foreground(Teal).Italic(true)
	GlobalBadge   = lipgloss.NewStyle().Foreground(Gray).Italic(true)
	TypeBadge     = lipgloss.NewStyle().Foreground(Purple)
	BorderBox     = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(Subtle).Padding(0, 1)
	StatusStyle   = lipgloss.NewStyle().Foreground(Gray).Padding(0, 1)
	ErrorStyle    = lipgloss.NewStyle().Foreground(Red).Bold(true)
	SelectedStyle = lipgloss.NewStyle().Foreground(Purple).Bold(true)

	HelpStyle = lipgloss.NewStyle().Foreground(Gray)
)

// Badge returns a styled source badge string.
func Badge(source string) string {
	switch source {
	case "local":
		return LocalBadge.Render("[local]")
	default:
		return GlobalBadge.Render("[global]")
	}
}

// TypeTag returns a styled type badge for scripts.
func TypeTag(t string) string {
	switch t {
	case "executable":
		return TypeBadge.Render("[bin]")
	default:
		return TypeBadge.Render("[sh]")
	}
}

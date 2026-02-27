// Package styles defines the ShotGum color palette and shared lipgloss styles
// used across all TUI views. Import this package instead of re-declaring colors
// or styles locally.
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

	TitleStyle         = lipgloss.NewStyle().Bold(true).Foreground(Purple)
	CategoryStyle      = lipgloss.NewStyle().Foreground(Teal).Bold(true)
	ScriptStyle        = lipgloss.NewStyle().Foreground(White)
	DescStyle          = lipgloss.NewStyle().Foreground(Gray)
	LocalBadge         = lipgloss.NewStyle().Foreground(Teal).Italic(true)
	UserBadge          = lipgloss.NewStyle().Foreground(Gray).Italic(true)
	SectionHeaderStyle = lipgloss.NewStyle().Foreground(Gray)
	BorderBox          = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(Subtle).Padding(0, 1)
	StatusStyle        = lipgloss.NewStyle().Foreground(Gray).Padding(0, 1)
	ErrorStyle         = lipgloss.NewStyle().Foreground(Red).Bold(true)
	SelectedStyle      = lipgloss.NewStyle().Foreground(Purple).Bold(true)

	HelpStyle = lipgloss.NewStyle().Foreground(Gray)
)

// Badge returns a styled source badge string.
func Badge(source string) string {
	switch source {
	case "local":
		return LocalBadge.Render("[local]")
	case "make":
		return LocalBadge.Render("[makefile]")
	default:
		return UserBadge.Render("[user]")
	}
}

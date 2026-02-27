package views

import (
	"fmt"
	"io"

	"github.com/brunoomariano/ShotGum-Toolchain/internal/tui/styles"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SettingItem represents a configurable option in the settings list.
type SettingItem struct {
	Key       string
	TitleText string
	DescText  string
	Value     string
}

func (i SettingItem) Title() string {
	return styles.TitleStyle.Render(i.TitleText)
}

func (i SettingItem) Description() string {
	return styles.DescStyle.Render(i.DescText)
}

func (i SettingItem) FilterValue() string { return i.TitleText }

// settingDelegate renders settings list items.
type settingDelegate struct{}

func (d settingDelegate) Height() int                             { return 2 }
func (d settingDelegate) Spacing() int                            { return 1 }
func (d settingDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d settingDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(SettingItem)
	if !ok {
		return
	}

	title := fmt.Sprintf("  %s", styles.TitleStyle.Render(i.TitleText))
	desc := fmt.Sprintf("  %s  %s", styles.DescStyle.Render(i.DescText), styles.StatusStyle.Render(i.Value))

	if index == m.Index() {
		title = fmt.Sprintf("  %s", lipgloss.NewStyle().Foreground(styles.Purple).Bold(true).Render("▶ "+i.TitleText))
	}

	fmt.Fprintf(w, "%s\n%s", title, desc)
}

// NewSettingsList creates a list model for the settings menu.
func NewSettingsList(items []list.Item, w, h int) list.Model {
	l := list.New(items, settingDelegate{}, w, h)
	l.Title = styles.TitleStyle.Render("Settings")
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false)
	l.Styles.Title = styles.TitleStyle
	l.Styles.HelpStyle = styles.HelpStyle
	return l
}

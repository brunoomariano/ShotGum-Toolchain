package views

import (
	"fmt"
	"io"

	"github.com/brunoomariano/ShotGum-Toolchain/internal/registry"
	"github.com/brunoomariano/ShotGum-Toolchain/internal/tui/styles"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ScriptItem wraps a ScriptEntry for bubbles/list.
type ScriptItem struct {
	entry registry.ScriptEntry
}

// Entry returns the underlying ScriptEntry.
func (i ScriptItem) Entry() registry.ScriptEntry { return i.entry }

func (i ScriptItem) Title() string {
	return styles.ScriptStyle.Render(i.entry.Name)
}

func (i ScriptItem) Description() string {
	return styles.DescStyle.Render(i.entry.Description)
}

func (i ScriptItem) FilterValue() string { return i.entry.Name }

// scriptDelegate renders script list items.
type scriptDelegate struct{}

func (d scriptDelegate) Height() int                             { return 2 }
func (d scriptDelegate) Spacing() int                            { return 1 }
func (d scriptDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d scriptDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	if h, ok := item.(SectionHeaderItem); ok {
		line := sectionHeaderLine(h.label, m.Width())
		fmt.Fprintf(w, "%s\n%s", line, "")
		return
	}

	i, ok := item.(ScriptItem)
	if !ok {
		return
	}

	title := fmt.Sprintf("  %s", styles.ScriptStyle.Render(i.entry.Name))
	desc := fmt.Sprintf("  %s", styles.DescStyle.Render(i.entry.Description))

	if index == m.Index() {
		title = fmt.Sprintf("  %s", lipgloss.NewStyle().Foreground(styles.Purple).Bold(true).Render("▶ "+i.entry.Name))
	}

	fmt.Fprintf(w, "%s\n%s", title, desc)
}

// NewScriptList creates a new script list.Model for a given category,
// grouping entries under "Local" and "User" section headers.
func NewScriptList(categoryName string, entries []registry.ScriptEntry, w, h int) list.Model {
	var items []list.Item
	var currentSource string

	for _, e := range entries {
		if e.Source != currentSource {
			currentSource = e.Source
			label := "Local"
			if e.Source == "user" {
				label = "User"
			}
			items = append(items, SectionHeaderItem{label: label})
		}
		items = append(items, ScriptItem{entry: e})
	}

	l := list.New(items, scriptDelegate{}, w, h)
	l.Title = styles.TitleStyle.Render("ShotGum") +
		styles.DescStyle.Render("  — ") +
		styles.CategoryStyle.Render(categoryName)
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = styles.TitleStyle
	l.Styles.HelpStyle = styles.HelpStyle

	// Start cursor on first real item, skipping any leading section header.
	if len(items) > 1 {
		if _, ok := items[0].(SectionHeaderItem); ok {
			l.Select(1)
		}
	}

	return l
}

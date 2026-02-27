package views

import (
	"fmt"
	"io"
	"strings"

	"github.com/brunoomariano/ShotGum-Toolchain/internal/registry"
	"github.com/brunoomariano/ShotGum-Toolchain/internal/tui/styles"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SectionHeaderItem is a non-selectable section divider used in both the
// category and script lists to separate "Local" from "User" entries.
type SectionHeaderItem struct {
	label string
}

func (h SectionHeaderItem) FilterValue() string { return "" }

// sectionHeaderLine renders a full-width section divider such as:
//
//	─ Local ─────────────────────────────────────
func sectionHeaderLine(label string, width int) string {
	if width <= 0 {
		width = 40
	}
	prefix := " ─ " + label + " "
	remaining := width - lipgloss.Width(prefix) - 2
	if remaining < 2 {
		remaining = 2
	}
	return styles.SectionHeaderStyle.Render(prefix + strings.Repeat("─", remaining))
}

// CategoryItem wraps a CategoryEntry for the Navigation Container list.
type CategoryItem struct {
	entry registry.CategoryEntry
}

// Entry returns the underlying CategoryEntry.
func (i CategoryItem) Entry() registry.CategoryEntry { return i.entry }

func (i CategoryItem) Title() string {
	return styles.CategoryStyle.Render(i.entry.Name)
}

func (i CategoryItem) Description() string {
	return styles.DescStyle.Render(i.entry.Description)
}

func (i CategoryItem) FilterValue() string { return i.entry.Name }

// categoryDelegate renders list items.
type categoryDelegate struct{}

func (d categoryDelegate) Height() int                             { return 2 }
func (d categoryDelegate) Spacing() int                            { return 1 }
func (d categoryDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d categoryDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	if h, ok := item.(SectionHeaderItem); ok {
		line := sectionHeaderLine(h.label, m.Width())
		fmt.Fprintf(w, "%s\n%s", line, "")
		return
	}

	i, ok := item.(CategoryItem)
	if !ok {
		return
	}

	title := styles.CategoryStyle.Render(i.entry.Name)
	desc := styles.DescStyle.Render(i.entry.Description)

	if index == m.Index() {
		title = lipgloss.NewStyle().Foreground(styles.Purple).Bold(true).Render(
			fmt.Sprintf("▶ %s", i.entry.Name),
		)
	}

	fmt.Fprintf(w, "  %s\n  %s", title, desc)
}

// NewCategoryList creates a new category list.Model from registry entries,
// grouping them under "Local" and "User" section headers.
func NewCategoryList(entries []registry.CategoryEntry, w, h int) list.Model {
	var items []list.Item
	var currentSource string

	for _, e := range entries {
		if e.Source != currentSource {
			currentSource = e.Source
			label := sourceLabel(e.Source)
			items = append(items, SectionHeaderItem{label: label})
		}
		items = append(items, CategoryItem{entry: e})
	}

	l := list.New(items, categoryDelegate{}, w, h)
	l.SetShowTitle(false)
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

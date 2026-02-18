package views

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shotgum/stg/internal/registry"
	"github.com/shotgum/stg/internal/tui/styles"
)

// CategoryItem wraps a CategoryEntry for bubbles/list.
type CategoryItem struct {
	entry registry.CategoryEntry
}

// Entry returns the underlying CategoryEntry.
func (i CategoryItem) Entry() registry.CategoryEntry { return i.entry }

func (i CategoryItem) Title() string {
	return fmt.Sprintf("%s %s", styles.Badge(i.entry.Source), styles.CategoryStyle.Render(i.entry.Name))
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
	i, ok := item.(CategoryItem)
	if !ok {
		return
	}

	title := fmt.Sprintf("%s %s", styles.Badge(i.entry.Source), styles.CategoryStyle.Render(i.entry.Name))
	desc := styles.DescStyle.Render(i.entry.Description)

	if index == m.Index() {
		title = lipgloss.NewStyle().Foreground(styles.Purple).Bold(true).Render(
			fmt.Sprintf("▶ %s %s", styles.Badge(i.entry.Source), i.entry.Name),
		)
	}

	fmt.Fprintf(w, "  %s\n  %s", title, desc)
}

// NewCategoryList creates a new category list.Model from registry entries.
func NewCategoryList(entries []registry.CategoryEntry, w, h int) list.Model {
	items := make([]list.Item, len(entries))
	for i, e := range entries {
		items[i] = CategoryItem{entry: e}
	}

	l := list.New(items, categoryDelegate{}, w, h)
	l.Title = styles.TitleStyle.Render("ShotGum") + styles.DescStyle.Render("  — script manager")
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = styles.TitleStyle
	l.Styles.HelpStyle = styles.HelpStyle

	return l
}

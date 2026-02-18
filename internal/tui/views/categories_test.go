package views

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shotgum/stg/internal/config"
	"github.com/shotgum/stg/internal/registry"
)

func makeCategoryEntry(name, desc, source string) registry.CategoryEntry {
	return registry.CategoryEntry{
		Category: config.Category{Name: name, Description: desc},
		Source:   source,
	}
}

func TestCategoryItem_Title(t *testing.T) {
	entry := makeCategoryEntry("tools", "My tools", "local")
	item := CategoryItem{entry: entry}
	title := item.Title()
	if !strings.Contains(title, "tools") {
		t.Errorf("Title() = %q, want it to contain 'tools'", title)
	}
	if !strings.Contains(title, "[local]") {
		t.Errorf("Title() = %q, want it to contain '[local]'", title)
	}
}

func TestCategoryItem_Title_Global(t *testing.T) {
	entry := makeCategoryEntry("dev", "Dev tools", "global")
	item := CategoryItem{entry: entry}
	title := item.Title()
	if !strings.Contains(title, "[global]") {
		t.Errorf("Title() = %q, want it to contain '[global]'", title)
	}
}

func TestCategoryItem_Description(t *testing.T) {
	entry := makeCategoryEntry("tools", "My tools", "global")
	item := CategoryItem{entry: entry}
	desc := item.Description()
	if !strings.Contains(desc, "My tools") {
		t.Errorf("Description() = %q, want it to contain 'My tools'", desc)
	}
}

func TestCategoryItem_FilterValue(t *testing.T) {
	entry := makeCategoryEntry("deploy", "Deploy scripts", "global")
	item := CategoryItem{entry: entry}
	if got := item.FilterValue(); got != "deploy" {
		t.Errorf("FilterValue() = %q, want 'deploy'", got)
	}
}

func TestCategoryItem_Entry(t *testing.T) {
	entry := makeCategoryEntry("ops", "Ops scripts", "local")
	item := CategoryItem{entry: entry}
	got := item.Entry()
	if got.Name != "ops" || got.Source != "local" {
		t.Errorf("Entry() = %+v, unexpected", got)
	}
}

func TestNewCategoryList_Empty(t *testing.T) {
	// Should not panic with zero items
	l := NewCategoryList(nil, 80, 24)
	_ = l
}

func TestNewCategoryList_WithItems(t *testing.T) {
	entries := []registry.CategoryEntry{
		makeCategoryEntry("dev", "Dev tools", "global"),
		makeCategoryEntry("ops", "Ops tools", "local"),
	}
	l := NewCategoryList(entries, 80, 24)
	if len(l.Items()) != 2 {
		t.Errorf("expected 2 items, got %d", len(l.Items()))
	}
}

func TestNewCategoryList_View_TriggersRender(t *testing.T) {
	entries := []registry.CategoryEntry{
		makeCategoryEntry("dev", "Dev tools", "global"),
		makeCategoryEntry("ops", "Ops tools", "local"),
	}
	l := NewCategoryList(entries, 80, 24)
	// View() calls the delegate's Render method internally
	view := l.View()
	_ = view // should not panic
}

func TestCategoryList_Update_TriggersDelegate(t *testing.T) {
	entries := []registry.CategoryEntry{
		makeCategoryEntry("dev", "Dev tools", "global"),
	}
	l := NewCategoryList(entries, 80, 24)
	// Update() calls the delegate's Update method internally
	var cmd interface{ Run() }
	l2, _ := l.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	_ = l2
	_ = cmd
}

// TestCategoryDelegate_Render_InvalidItem verifies that Render exits early and writes
// nothing when the list.Item cannot be type-asserted to CategoryItem.
func TestCategoryDelegate_Render_InvalidItem(t *testing.T) {
	d := categoryDelegate{}
	var buf strings.Builder
	l := list.New(nil, d, 80, 24)
	// Pass a ScriptItem, which is not a CategoryItem → the `if !ok { return }` branch
	d.Render(&buf, l, 0, ScriptItem{})
	if buf.Len() != 0 {
		t.Errorf("Render with invalid item type should write nothing, got %q", buf.String())
	}
}

package views

import (
	"strings"
	"testing"

	"github.com/brunoomariano/ShotGum-Toolchain/internal/config"
	"github.com/brunoomariano/ShotGum-Toolchain/internal/registry"
	"github.com/charmbracelet/bubbles/list"
)

func makeScriptEntry(name, desc, source string) registry.ScriptEntry {
	return registry.ScriptEntry{
		Script: config.Script{Name: name, Description: desc, Category: "testcat"},
		Source: source,
	}
}

func TestScriptItem_Title(t *testing.T) {
	entry := makeScriptEntry("build", "Build project", "local")
	item := ScriptItem{entry: entry}
	title := item.Title()
	if !strings.Contains(title, "build") {
		t.Errorf("Title() = %q, want it to contain 'build'", title)
	}
}

func TestScriptItem_Description(t *testing.T) {
	entry := makeScriptEntry("test", "Run tests", "user")
	item := ScriptItem{entry: entry}
	desc := item.Description()
	if !strings.Contains(desc, "Run tests") {
		t.Errorf("Description() = %q, want it to contain 'Run tests'", desc)
	}
}

func TestScriptItem_FilterValue(t *testing.T) {
	entry := makeScriptEntry("deploy", "Deploy app", "local")
	item := ScriptItem{entry: entry}
	if got := item.FilterValue(); got != "deploy" {
		t.Errorf("FilterValue() = %q, want 'deploy'", got)
	}
}

func TestScriptItem_Entry(t *testing.T) {
	entry := makeScriptEntry("lint", "Lint code", "user")
	item := ScriptItem{entry: entry}
	got := item.Entry()
	if got.Name != "lint" || got.Source != "user" {
		t.Errorf("Entry() = %+v, unexpected", got)
	}
}

func TestNewScriptList_WithItems(t *testing.T) {
	entries := []registry.ScriptEntry{
		makeScriptEntry("build", "Build", "user"),
		makeScriptEntry("test", "Test", "local"),
	}
	l := NewScriptList("mycat", entries, 80, 24)
	// 2 real entries + 2 section headers ("User" and "Local")
	if len(l.Items()) != 4 {
		t.Errorf("expected 4 items (2 entries + 2 section headers), got %d", len(l.Items()))
	}
}

func TestNewScriptList_Empty(t *testing.T) {
	// Should not panic with zero items
	l := NewScriptList("mycat", nil, 80, 24)
	_ = l
}

func TestNewScriptList_View_TriggersRender(t *testing.T) {
	entries := []registry.ScriptEntry{
		makeScriptEntry("build", "Build", "user"),
		makeScriptEntry("mybin", "A bin", "local"),
	}
	l := NewScriptList("mycat", entries, 80, 24)
	// View() calls the delegate's Render method internally
	view := l.View()
	_ = view // should not panic
}

// TestScriptDelegate_Update_ReturnsNil verifies that scriptDelegate.Update always
// returns nil (it does not handle any messages itself; bubbles/list handles them).
func TestScriptDelegate_Update_ReturnsNil(t *testing.T) {
	d := scriptDelegate{}
	cmd := d.Update(nil, nil)
	if cmd != nil {
		t.Error("scriptDelegate.Update should always return nil")
	}
}

// TestScriptDelegate_Render_InvalidItem verifies that Render exits early and writes
// nothing when the list.Item cannot be type-asserted to ScriptItem or SectionHeaderItem.
func TestScriptDelegate_Render_InvalidItem(t *testing.T) {
	d := scriptDelegate{}
	var buf strings.Builder
	l := list.New(nil, d, 80, 24)
	// Pass a CategoryItem, which is not a ScriptItem → the `if !ok { return }` branch
	d.Render(&buf, l, 0, CategoryItem{})
	if buf.Len() != 0 {
		t.Errorf("Render with invalid item type should write nothing, got %q", buf.String())
	}
}

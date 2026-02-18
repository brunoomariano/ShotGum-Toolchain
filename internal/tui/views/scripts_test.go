package views

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	"github.com/shotgum/stg/internal/config"
	"github.com/shotgum/stg/internal/registry"
)

func makeScriptEntry(name, desc, typ, source string) registry.ScriptEntry {
	return registry.ScriptEntry{
		Script: config.Script{Name: name, Description: desc, Type: typ, Category: "testcat"},
		Source: source,
	}
}

func TestScriptItem_Title_Script(t *testing.T) {
	entry := makeScriptEntry("build", "Build project", "script", "local")
	item := ScriptItem{entry: entry}
	title := item.Title()
	if !strings.Contains(title, "build") {
		t.Errorf("Title() = %q, want it to contain 'build'", title)
	}
	if !strings.Contains(title, "[local]") {
		t.Errorf("Title() = %q, want it to contain '[local]'", title)
	}
	if !strings.Contains(title, "[sh]") {
		t.Errorf("Title() = %q, want it to contain '[sh]'", title)
	}
}

func TestScriptItem_Title_Executable(t *testing.T) {
	entry := makeScriptEntry("mybin", "A binary", "executable", "global")
	item := ScriptItem{entry: entry}
	title := item.Title()
	if !strings.Contains(title, "[bin]") {
		t.Errorf("Title() = %q, want it to contain '[bin]'", title)
	}
	if !strings.Contains(title, "[global]") {
		t.Errorf("Title() = %q, want it to contain '[global]'", title)
	}
}

func TestScriptItem_Description(t *testing.T) {
	entry := makeScriptEntry("test", "Run tests", "script", "global")
	item := ScriptItem{entry: entry}
	desc := item.Description()
	if !strings.Contains(desc, "Run tests") {
		t.Errorf("Description() = %q, want it to contain 'Run tests'", desc)
	}
}

func TestScriptItem_FilterValue(t *testing.T) {
	entry := makeScriptEntry("deploy", "Deploy app", "script", "local")
	item := ScriptItem{entry: entry}
	if got := item.FilterValue(); got != "deploy" {
		t.Errorf("FilterValue() = %q, want 'deploy'", got)
	}
}

func TestScriptItem_Entry(t *testing.T) {
	entry := makeScriptEntry("lint", "Lint code", "script", "global")
	item := ScriptItem{entry: entry}
	got := item.Entry()
	if got.Name != "lint" || got.Source != "global" {
		t.Errorf("Entry() = %+v, unexpected", got)
	}
}

func TestNewScriptList_WithItems(t *testing.T) {
	entries := []registry.ScriptEntry{
		makeScriptEntry("build", "Build", "script", "global"),
		makeScriptEntry("test", "Test", "script", "local"),
	}
	l := NewScriptList("mycat", entries, 80, 24)
	if len(l.Items()) != 2 {
		t.Errorf("expected 2 items, got %d", len(l.Items()))
	}
}

func TestNewScriptList_Empty(t *testing.T) {
	// Should not panic with zero items
	l := NewScriptList("mycat", nil, 80, 24)
	_ = l
}

func TestNewScriptList_View_TriggersRender(t *testing.T) {
	entries := []registry.ScriptEntry{
		makeScriptEntry("build", "Build", "script", "global"),
		makeScriptEntry("mybin", "A bin", "executable", "local"),
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
// nothing when the list.Item cannot be type-asserted to ScriptItem.
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

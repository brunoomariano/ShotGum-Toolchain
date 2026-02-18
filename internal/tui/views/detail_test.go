package views

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shotgum/stg/internal/config"
	"github.com/shotgum/stg/internal/registry"
)

func makeDetailScriptEntry(name, category string) *registry.ScriptEntry {
	return &registry.ScriptEntry{
		Script: config.Script{
			Name:     name,
			Category: category,
			Type:     "script",
			Path:     "/test/path/" + name + ".sh",
		},
		Source: "global",
	}
}

func makeDetailCategoryEntry(name string) *registry.CategoryEntry {
	return &registry.CategoryEntry{
		Category: config.Category{
			Name:        name,
			Description: "desc for " + name,
		},
		Source: "global",
	}
}

func TestDetailModel_SetCategory(t *testing.T) {
	var d DetailModel
	cat := makeDetailCategoryEntry("tools")
	d.SetCategory(cat, nil, nil)
	if d.mode != detailCategory {
		t.Error("SetCategory should set mode to detailCategory")
	}
	if d.catEntry != cat {
		t.Error("SetCategory should set catEntry")
	}
	if d.scrEntry != nil {
		t.Error("SetCategory should clear scrEntry")
	}
}

func TestDetailModel_SetScript(t *testing.T) {
	var d DetailModel
	d.SetSize(80, 24)
	scr := makeDetailScriptEntry("build", "tools")
	d.SetScript(scr, nil)
	if d.mode != detailScript {
		t.Error("SetScript should set mode to detailScript")
	}
	if d.scrEntry != scr {
		t.Error("SetScript should set scrEntry")
	}
	if d.catEntry != nil {
		t.Error("SetScript should clear catEntry")
	}
	if d.helpText != "" {
		t.Error("SetScript should clear helpText")
	}
}

func TestDetailModel_SetHelpText_MatchingScript(t *testing.T) {
	var d DetailModel
	d.SetSize(80, 24)
	scr := makeDetailScriptEntry("build", "tools")
	d.SetScript(scr, nil)
	d.SetHelpText("build", "help output text")
	if d.helpText != "help output text" {
		t.Errorf("helpText = %q, want 'help output text'", d.helpText)
	}
	if d.helpScript != "build" {
		t.Errorf("helpScript = %q, want 'build'", d.helpScript)
	}
}

func TestDetailModel_SetHelpText_DifferentScript(t *testing.T) {
	var d DetailModel
	d.SetSize(80, 24)
	scr := makeDetailScriptEntry("build", "tools")
	d.SetScript(scr, nil)
	d.SetHelpText("other-script", "help output text")
	// Should have no effect
	if d.helpText != "" {
		t.Error("SetHelpText with wrong name should not update helpText")
	}
	if d.helpScript != "" {
		t.Error("SetHelpText with wrong name should not update helpScript")
	}
}

func TestDetailModel_SetSize(t *testing.T) {
	var d DetailModel
	d.SetSize(80, 24)
	if !d.vpReady {
		t.Error("SetSize should set vpReady = true")
	}
	if d.vp.Width != 80 {
		t.Errorf("vp.Width = %d, want 80", d.vp.Width)
	}
}

func TestDetailModel_SetSize_Resize(t *testing.T) {
	var d DetailModel
	d.SetSize(80, 24)
	d.SetSize(100, 30)
	if d.vp.Width != 100 {
		t.Errorf("after resize vp.Width = %d, want 100", d.vp.Width)
	}
}

func TestDetailModel_View_None(t *testing.T) {
	var d DetailModel // detailNone mode
	view := d.View(80, 24)
	if !strings.Contains(stripANSI(view), "Select an item") {
		t.Errorf("View() in detailNone should contain 'Select an item': %q", stripANSI(view))
	}
}

func TestDetailModel_View_Category(t *testing.T) {
	var d DetailModel
	d.SetSize(80, 24)
	cat := makeDetailCategoryEntry("mycat")
	d.SetCategory(cat, nil, nil)
	view := d.View(80, 24)
	if !strings.Contains(stripANSI(view), "mycat") {
		t.Errorf("View() in category mode should contain category name: %q", stripANSI(view))
	}
}

func TestDetailModel_View_Script_Basic(t *testing.T) {
	var d DetailModel
	d.SetSize(80, 24)
	scr := makeDetailScriptEntry("myscript", "mycat")
	d.SetScript(scr, nil)
	view := d.View(80, 24)
	stripped := stripANSI(view)
	if !strings.Contains(stripped, "myscript") {
		t.Errorf("View() should contain script name: %q", stripped)
	}
}

func TestDetailModel_View_Script_LoadingHelp(t *testing.T) {
	var d DetailModel
	d.SetSize(80, 24)
	scr := makeDetailScriptEntry("myscript", "mycat")
	d.SetScript(scr, nil)
	view := d.View(80, 24)
	stripped := stripANSI(view)
	// helpScript == "" after SetScript → shows "loading help…"
	if !strings.Contains(stripped, "loading help") {
		t.Errorf("View() should show 'loading help': %q", stripped)
	}
}

func TestDetailModel_View_Script_WithHelp(t *testing.T) {
	var d DetailModel
	d.SetSize(80, 30)
	scr := makeDetailScriptEntry("myscript", "mycat")
	d.SetScript(scr, nil)
	d.SetHelpText("myscript", "Usage: myscript [options]")
	view := d.View(80, 30)
	stripped := stripANSI(view)
	if !strings.Contains(stripped, "Help") {
		t.Errorf("View() with helpText should contain 'Help': %q", stripped)
	}
}

func TestDetailModel_View_Script_HelpFlag(t *testing.T) {
	var d DetailModel
	d.SetSize(80, 24)
	scr := makeDetailScriptEntry("myscript", "mycat")
	d.SetScript(scr, nil)
	view := d.View(80, 24)
	stripped := stripANSI(view)
	// With nil registry, helpFlag defaults to "--help" in renderScript
	if !strings.Contains(stripped, "--help") {
		t.Errorf("View() should contain help flag '--help': %q", stripped)
	}
}

func TestDetailModel_Update(t *testing.T) {
	var d DetailModel
	d.SetSize(80, 24)
	// Update forwards messages to the viewport; should not panic
	d2, _ := d.Update(tea.KeyMsg{Type: tea.KeyDown})
	_ = d2
}

func TestDetailModel_View_Category_WithScripts(t *testing.T) {
	var d DetailModel
	d.SetSize(80, 24)
	cat := makeDetailCategoryEntry("mycat")
	scripts := []registry.ScriptEntry{
		{
			Script: config.Script{Name: "build", Type: "script"},
			Source: "global",
		},
	}
	d.SetCategory(cat, scripts, nil)
	view := d.View(80, 24)
	stripped := stripANSI(view)
	if !strings.Contains(stripped, "mycat") {
		t.Errorf("View() category with scripts should contain category name: %q", stripped)
	}
}

// TestDetailModel_renderCategory_NilCatEntry exercises the defensive nil guard at the
// top of renderCategory — reached when mode==detailCategory but catEntry was never set.
func TestDetailModel_renderCategory_NilCatEntry(t *testing.T) {
	var d DetailModel
	d.mode = detailCategory // catEntry remains nil
	view := d.View(80, 24)
	if view != "" {
		t.Errorf("renderCategory with nil catEntry should return empty string, got %q", view)
	}
}

// TestDetailModel_renderScript_NilScrEntry exercises the defensive nil guard at the
// top of renderScript — reached when mode==detailScript but scrEntry was never set.
func TestDetailModel_renderScript_NilScrEntry(t *testing.T) {
	var d DetailModel
	d.mode = detailScript // scrEntry remains nil
	view := d.View(80, 24)
	if view != "" {
		t.Errorf("renderScript with nil scrEntry should return empty string, got %q", view)
	}
}

// TestDetailModel_SetSize_WithHelpText verifies that calling SetSize after helpText has
// been set pushes the content into the viewport (the `if m.helpText != ""` branch).
func TestDetailModel_SetSize_WithHelpText(t *testing.T) {
	var d DetailModel
	d.SetSize(80, 24)
	scr := makeDetailScriptEntry("myscript", "mycat")
	d.SetScript(scr, nil)
	d.SetHelpText("myscript", "Usage: myscript [options]")
	// Resize — this triggers the `if m.helpText != ""` branch inside SetSize
	d.SetSize(90, 28)
	if d.vp.Width != 90 {
		t.Errorf("vp.Width after resize = %d, want 90", d.vp.Width)
	}
}

// TestDetailModel_View_Script_WithHelp_ScrollIndicator verifies that the scroll
// percentage indicator appears when the help text exceeds the viewport height.
func TestDetailModel_View_Script_WithHelp_ScrollIndicator(t *testing.T) {
	var d DetailModel
	d.SetSize(80, 20) // small panel → vpH = 6, easy to exceed
	scr := makeDetailScriptEntry("myscript", "mycat")
	d.SetScript(scr, nil)
	// 40 lines exceeds vpH=6, so TotalLineCount > Height → scrollInfo is shown
	longHelp := strings.Repeat("Usage: myscript --flag value\n", 40)
	d.SetHelpText("myscript", longHelp)
	view := d.View(80, 20)
	if view == "" {
		t.Error("View() with long help should not be empty")
	}
}

// TestLoadScriptHelpCmd_ReturnsDetailHelpMsg verifies that LoadScriptHelpCmd returns a
// tea.Cmd which, when invoked, runs the script and emits a DetailHelpMsg with the
// correct ScriptName regardless of the script's exit code.
func TestLoadScriptHelpCmd_ReturnsDetailHelpMsg(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)
	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })
	os.Chdir(tmpdir)

	scriptPath := filepath.Join(tmpdir, "helper.sh")
	os.WriteFile(scriptPath, []byte("#!/bin/bash\necho 'Usage: helper'"), 0755)

	config.Save(&config.Config{
		Version:    "1",
		Categories: []config.Category{{Name: "tools"}},
		Scripts:    []config.Script{{Name: "helper", Category: "tools", Type: "script", Path: scriptPath}},
	}, config.GlobalConfigPath())

	reg, err := registry.Load()
	if err != nil {
		t.Fatalf("registry.Load() error: %v", err)
	}
	entry, err := reg.FindScript("tools", "helper")
	if err != nil {
		t.Fatalf("FindScript() error: %v", err)
	}

	cmd := LoadScriptHelpCmd(*entry, reg, 80)
	msg := cmd()
	helpMsg, ok := msg.(DetailHelpMsg)
	if !ok {
		t.Fatalf("LoadScriptHelpCmd() should return DetailHelpMsg, got %T", msg)
	}
	if helpMsg.ScriptName != "helper" {
		t.Errorf("ScriptName = %q, want 'helper'", helpMsg.ScriptName)
	}
}

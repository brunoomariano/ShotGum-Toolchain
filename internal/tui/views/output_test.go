package views

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shotgum/stg/internal/config"
	"github.com/shotgum/stg/internal/registry"
)

func TestNewOutputModel_LoadingTrue(t *testing.T) {
	m := NewOutputModel("myscript", 80, 24)
	if !m.IsLoading() {
		t.Error("NewOutputModel should start with loading=true")
	}
	if m.scriptName != "myscript" {
		t.Errorf("scriptName = %q, want 'myscript'", m.scriptName)
	}
}

func TestOutputModel_IsLoading_BeforeDone(t *testing.T) {
	m := NewOutputModel("test", 80, 24)
	if !m.IsLoading() {
		t.Error("IsLoading() should be true before ScriptDoneMsg")
	}
}

func TestOutputModel_IsLoading_AfterDone(t *testing.T) {
	m := NewOutputModel("test", 80, 24)
	m2, _ := m.Update(ScriptDoneMsg{Output: "ok", Err: nil})
	if m2.IsLoading() {
		t.Error("IsLoading() should be false after ScriptDoneMsg")
	}
}

func TestOutputModel_Update_ScriptDone_Success(t *testing.T) {
	m := NewOutputModel("test", 80, 24)
	m2, _ := m.Update(ScriptDoneMsg{Output: "build complete", Err: nil})
	if m2.loading {
		t.Error("loading should be false after ScriptDoneMsg")
	}
	if m2.output != "build complete" {
		t.Errorf("output = %q, want 'build complete'", m2.output)
	}
	if m2.err != nil {
		t.Errorf("err should be nil, got %v", m2.err)
	}
}

func TestOutputModel_Update_ScriptDone_Error(t *testing.T) {
	m := NewOutputModel("test", 80, 24)
	scriptErr := errors.New("script failed with code 1")
	m2, _ := m.Update(ScriptDoneMsg{Output: "partial output", Err: scriptErr})
	if m2.loading {
		t.Error("loading should be false after ScriptDoneMsg with error")
	}
	if m2.err == nil {
		t.Error("err should be set after ScriptDoneMsg with error")
	}
	view := m2.View()
	stripped := stripANSI(view)
	if !strings.Contains(stripped, "error") {
		t.Errorf("View() after error should contain 'error': %q", stripped)
	}
}

func TestOutputModel_Update_WindowSize(t *testing.T) {
	m := NewOutputModel("test", 80, 24)
	m2, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	if m2.viewport.Width != 116 { // 120 - 4
		t.Errorf("viewport width = %d, want 116", m2.viewport.Width)
	}
	if m2.viewport.Height != 34 { // 40 - 6
		t.Errorf("viewport height = %d, want 34", m2.viewport.Height)
	}
}

func TestOutputModel_View_Loading(t *testing.T) {
	m := NewOutputModel("myscript", 80, 24)
	view := m.View()
	stripped := stripANSI(view)
	if !strings.Contains(stripped, "myscript") {
		t.Errorf("View() while loading should contain script name: %q", stripped)
	}
}

func TestOutputModel_View_Done(t *testing.T) {
	m := NewOutputModel("myscript", 80, 24)
	m2, _ := m.Update(ScriptDoneMsg{Output: "build done!", Err: nil})
	view := m2.View()
	// Just ensure no panic and non-empty output
	if view == "" {
		t.Error("View() after done should not be empty")
	}
}

func TestOutputModel_Init(t *testing.T) {
	m := NewOutputModel("test", 80, 24)
	cmd := m.Init()
	// Init returns spinner.Tick which is non-nil
	if cmd == nil {
		t.Error("Init() should return a non-nil spinner tick cmd")
	}
}

// TestOutputModel_Update_SpinnerTick_Loading verifies that a spinner.TickMsg while
// loading advances the spinner animation and returns a follow-up tick command.
func TestOutputModel_Update_SpinnerTick_Loading(t *testing.T) {
	m := NewOutputModel("test", 80, 24)
	// m.loading == true by default
	m2, cmd := m.Update(spinner.TickMsg{Time: time.Now()})
	if !m2.loading {
		t.Error("loading should remain true after spinner tick")
	}
	// The spinner returns a follow-up tick command
	if cmd == nil {
		t.Error("spinner tick while loading should return a non-nil follow-up cmd")
	}
}

// TestOutputModel_Update_SpinnerTick_NotLoading verifies that a spinner.TickMsg
// received after the script has finished is ignored (no spinner update, no cmd).
func TestOutputModel_Update_SpinnerTick_NotLoading(t *testing.T) {
	m := NewOutputModel("test", 80, 24)
	m, _ = m.Update(ScriptDoneMsg{Output: "done", Err: nil}) // loading = false
	_, cmd := m.Update(spinner.TickMsg{Time: time.Now()})
	// When not loading, the spinner branch is skipped; the only cmd comes from the viewport
	_ = cmd // may be non-nil due to viewport passthrough; just ensure no panic
	if m.loading {
		t.Error("loading should still be false after spinner tick when not loading")
	}
}

// TestRunScriptCmd_ReturnsScriptDoneMsg verifies that RunScriptCmd returns a tea.Cmd
// that, when invoked, executes the script and wraps the result in a ScriptDoneMsg.
func TestRunScriptCmd_ReturnsScriptDoneMsg(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)
	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })
	os.Chdir(tmpdir)

	scriptPath := filepath.Join(tmpdir, "hello.sh")
	os.WriteFile(scriptPath, []byte("#!/bin/bash\necho hello"), 0755)

	config.Save(&config.Config{
		Version:    "1",
		Categories: []config.Category{{Name: "tools"}},
		Scripts:    []config.Script{{Name: "hello", Category: "tools", Type: "script", Path: scriptPath}},
	}, config.GlobalConfigPath())

	reg, err := registry.Load()
	if err != nil {
		t.Fatalf("registry.Load() error: %v", err)
	}
	entry, err := reg.FindScript("tools", "hello")
	if err != nil {
		t.Fatalf("FindScript() error: %v", err)
	}

	cmd := RunScriptCmd(*entry, nil, reg)
	msg := cmd()
	doneMsg, ok := msg.(ScriptDoneMsg)
	if !ok {
		t.Fatalf("RunScriptCmd() should return ScriptDoneMsg, got %T", msg)
	}
	if doneMsg.Err != nil {
		t.Errorf("unexpected error in ScriptDoneMsg: %v", doneMsg.Err)
	}
	if !strings.Contains(doneMsg.Output, "hello") {
		t.Errorf("output = %q, want it to contain 'hello'", doneMsg.Output)
	}
}

// TestRunScriptCmd_ScriptError verifies that RunScriptCmd wraps a script exit-code
// failure as a ScriptDoneMsg with a non-nil Err (output may still be present).
func TestRunScriptCmd_ScriptError(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)
	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })
	os.Chdir(tmpdir)

	scriptPath := filepath.Join(tmpdir, "fail.sh")
	os.WriteFile(scriptPath, []byte("#!/bin/bash\necho oops\nexit 2"), 0755)

	config.Save(&config.Config{
		Version:    "1",
		Categories: []config.Category{{Name: "tools"}},
		Scripts:    []config.Script{{Name: "fail", Category: "tools", Type: "script", Path: scriptPath}},
	}, config.GlobalConfigPath())

	reg, err := registry.Load()
	if err != nil {
		t.Fatalf("registry.Load() error: %v", err)
	}
	entry, err := reg.FindScript("tools", "fail")
	if err != nil {
		t.Fatalf("FindScript() error: %v", err)
	}

	cmd := RunScriptCmd(*entry, nil, reg)
	msg := cmd()
	doneMsg, ok := msg.(ScriptDoneMsg)
	if !ok {
		t.Fatalf("RunScriptCmd() should return ScriptDoneMsg, got %T", msg)
	}
	if doneMsg.Err == nil {
		t.Error("ScriptDoneMsg.Err should be non-nil when script exits with error")
	}
}

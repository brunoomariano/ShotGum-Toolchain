package views

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/brunoomariano/ShotGum-Toolchain/internal/config"
	"github.com/brunoomariano/ShotGum-Toolchain/internal/registry"
	tea "github.com/charmbracelet/bubbletea"
)

func makeConfirmEntry() registry.ScriptEntry {
	return registry.ScriptEntry{
		Script: config.Script{
			Name:     "deploy",
			Category: "ops",
			Path:     "/scripts/deploy.sh",
		},
		Source: "local",
	}
}

func collectMsg(cmd tea.Cmd) tea.Msg {
	if cmd == nil {
		return nil
	}
	return cmd()
}

func TestNewConfirmModel(t *testing.T) {
	entry := makeConfirmEntry()
	m := NewConfirmModel(entry, nil, 80, 24, "")
	if m.entry.Name != "deploy" {
		t.Errorf("entry.Name = %q, want 'deploy'", m.entry.Name)
	}
	if m.width != 80 || m.height != 24 {
		t.Errorf("width/height = %d/%d, want 80/24", m.width, m.height)
	}
	if m.inputActive {
		t.Error("inputActive should be false initially")
	}
}

func TestConfirmModel_Update_Enter_NoInput(t *testing.T) {
	m := NewConfirmModel(makeConfirmEntry(), nil, 80, 24, "")
	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_ = m2
	msg := collectMsg(cmd)
	if _, ok := msg.(ConfirmRunMsg); !ok {
		t.Errorf("Enter without input should emit ConfirmRunMsg, got %T", msg)
	}
}

func TestConfirmModel_Update_Esc_NoInput(t *testing.T) {
	m := NewConfirmModel(makeConfirmEntry(), nil, 80, 24, "")
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	msg := collectMsg(cmd)
	if _, ok := msg.(ConfirmCancelMsg); !ok {
		t.Errorf("Esc without input should emit ConfirmCancelMsg, got %T", msg)
	}
}

func TestConfirmModel_Update_Q_NoInput(t *testing.T) {
	m := NewConfirmModel(makeConfirmEntry(), nil, 80, 24, "")
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	msg := collectMsg(cmd)
	if _, ok := msg.(ConfirmCancelMsg); !ok {
		t.Errorf("q without input should emit ConfirmCancelMsg, got %T", msg)
	}
}

func TestConfirmModel_Update_A_ActivatesInput(t *testing.T) {
	m := NewConfirmModel(makeConfirmEntry(), nil, 80, 24, "")
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if !m2.inputActive {
		t.Error("pressing 'a' should activate inputActive")
	}
}

func TestConfirmModel_Update_Enter_WithInput(t *testing.T) {
	m := NewConfirmModel(makeConfirmEntry(), nil, 80, 24, "")
	// Activate input
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	// Press enter with input active
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	msg := collectMsg(cmd)
	runMsg, ok := msg.(ConfirmRunMsg)
	if !ok {
		t.Errorf("Enter with input active should emit ConfirmRunMsg, got %T", msg)
	}
	// ExtraArgs should be the (empty) args from the input field
	_ = runMsg
}

func TestConfirmModel_Update_Esc_WithInput_DeactivatesInput(t *testing.T) {
	m := NewConfirmModel(makeConfirmEntry(), nil, 80, 24, "")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if !m.inputActive {
		t.Fatal("precondition: inputActive should be true")
	}
	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	// Esc with input active should deactivate, not cancel
	if m2.inputActive {
		t.Error("Esc with input active should deactivate inputActive")
	}
	// Should NOT emit ConfirmCancelMsg
	msg := collectMsg(cmd)
	if _, ok := msg.(ConfirmCancelMsg); ok {
		t.Error("Esc with input active should not emit ConfirmCancelMsg")
	}
}

func TestConfirmModel_Update_WindowSize(t *testing.T) {
	m := NewConfirmModel(makeConfirmEntry(), nil, 0, 0, "")
	m2, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	if m2.width != 100 || m2.height != 50 {
		t.Errorf("width/height = %d/%d, want 100/50", m2.width, m2.height)
	}
}

func TestConfirmModel_View_ZeroSize(t *testing.T) {
	m := NewConfirmModel(makeConfirmEntry(), nil, 0, 0, "")
	view := m.View()
	stripped := stripANSI(view)
	if !strings.Contains(stripped, "Execute script?") {
		t.Errorf("View() should contain 'Execute script?': %q", stripped)
	}
	if !strings.Contains(stripped, "deploy") {
		t.Errorf("View() should contain script name 'deploy': %q", stripped)
	}
}

func TestConfirmModel_View_WithSize(t *testing.T) {
	m := NewConfirmModel(makeConfirmEntry(), nil, 80, 24, "")
	view := m.View()
	if view == "" {
		t.Error("View() with size should not be empty")
	}
}

// TestConfirmModel_View_WithInputActive verifies that View() renders the inline args
// text-input and the correct key-hint when inputActive is true.
func TestConfirmModel_View_WithInputActive(t *testing.T) {
	m := NewConfirmModel(makeConfirmEntry(), nil, 80, 24, "")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if !m.inputActive {
		t.Fatal("precondition: inputActive should be true after pressing 'a'")
	}
	view := m.View()
	stripped := stripANSI(view)
	if !strings.Contains(stripped, "Extra args:") {
		t.Errorf("View() with inputActive should contain 'Extra args:', got %q", stripped)
	}
	if !strings.Contains(stripped, "Run with args") {
		t.Errorf("View() with inputActive should show 'Run with args' hint, got %q", stripped)
	}
}

// TestConfirmModel_View_WithHelpText verifies that when inputActive and helpText is set,
// the help panel is rendered with the script help content.
func TestConfirmModel_View_WithHelpText(t *testing.T) {
	m := NewConfirmModel(makeConfirmEntry(), nil, 120, 40, "Usage: deploy [--env prod]")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if !m.inputActive {
		t.Fatal("precondition: inputActive should be true")
	}
	view := m.View()
	stripped := stripANSI(view)
	if !strings.Contains(stripped, "Script help") {
		t.Errorf("View() with inputActive+helpText should contain 'Script help', got %q", stripped)
	}
	if !strings.Contains(stripped, "Usage: deploy") {
		t.Errorf("View() with inputActive+helpText should contain help content, got %q", stripped)
	}
}

// TestConfirmModel_SetHelpText verifies that SetHelpText updates the model.
func TestConfirmModel_SetHelpText(t *testing.T) {
	m := NewConfirmModel(makeConfirmEntry(), nil, 120, 40, "")
	if m.helpText != "" {
		t.Error("helpText should be empty initially")
	}
	m.SetHelpText("Usage: deploy --env prod")
	if m.helpText != "Usage: deploy --env prod" {
		t.Errorf("SetHelpText: helpText = %q, want 'Usage: deploy --env prod'", m.helpText)
	}
}

// TestConfirmModel_View_LoadingHelp verifies that when inputActive and helpText is empty,
// the help panel shows a loading indicator.
func TestConfirmModel_View_LoadingHelp(t *testing.T) {
	m := NewConfirmModel(makeConfirmEntry(), nil, 120, 40, "")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	view := m.View()
	stripped := stripANSI(view)
	if !strings.Contains(stripped, "loading help") {
		t.Errorf("View() with inputActive and no helpText should show 'loading help', got %q", stripped)
	}
}

// TestConfirmModel_View_SmallWidth verifies the boxWidth clamping path where the
// candidate width (m.width - 12) does not exceed the default boxWidth of 50.
func TestConfirmModel_View_SmallWidth(t *testing.T) {
	// width=60 → candidate=48 < 50 → boxWidth stays at 50
	m := NewConfirmModel(makeConfirmEntry(), nil, 60, 24, "")
	view := m.View()
	if view == "" {
		t.Error("View() with small width should not be empty")
	}
}

// TestConfirmModel_Update_OtherKey_WithInput verifies that when inputActive is true and
// a non-special key is pressed, the event is forwarded to the text-input via
// argsInput.Update, returning the updated model and a potential cmd.
func TestConfirmModel_Update_OtherKey_WithInput(t *testing.T) {
	m := NewConfirmModel(makeConfirmEntry(), nil, 80, 24, "")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}) // activate input
	if !m.inputActive {
		t.Fatal("precondition: inputActive should be true")
	}
	// 'x' is not enter or esc, so it should be forwarded to argsInput.Update
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	_ = m2 // just ensure no panic
}

// TestNewConfirmModel_WithRegistry verifies that when a non-nil registry is provided,
// NewConfirmModel resolves and stores the script path via reg.ResolveScriptPath.
func TestNewConfirmModel_WithRegistry(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)
	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })
	os.Chdir(tmpdir)

	scriptPath := filepath.Join(tmpdir, "deploy.sh")
	os.WriteFile(scriptPath, []byte("#!/bin/bash\necho deploy"), 0755)
	config.Save(&config.Config{
		Version:    "1",
		Categories: []config.Category{{Name: "ops"}},
		Scripts:    []config.Script{{Name: "deploy", Category: "ops", Executable: "/bin/sh", Path: scriptPath}},
	}, config.GlobalConfigPath())

	reg, err := registry.Load()
	if err != nil {
		t.Fatalf("registry.Load() error: %v", err)
	}
	entry, err := reg.FindScript("ops", "deploy")
	if err != nil {
		t.Fatalf("FindScript() error: %v", err)
	}

	m := NewConfirmModel(*entry, reg, 80, 24, "")
	if m.resolvedPath != scriptPath {
		t.Errorf("resolvedPath = %q, want %q", m.resolvedPath, scriptPath)
	}
}

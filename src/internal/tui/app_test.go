package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/brunoomariano/ShotGum-Toolchain/internal/config"
	"github.com/brunoomariano/ShotGum-Toolchain/internal/registry"
	"github.com/brunoomariano/ShotGum-Toolchain/internal/tui/views"
	tea "github.com/charmbracelet/bubbletea"
)

// loadTestReg creates a temp HOME, writes cfg if provided, sets CWD to tmpdir,
// and returns a loaded registry. HOME and CWD are restored after the test.
func loadTestReg(t *testing.T, cfg *config.Config) *registry.Registry {
	t.Helper()
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)
	if cfg != nil {
		cfgPath := filepath.Join(tmpdir, ".config", "shotgum", "config.yaml")
		os.MkdirAll(filepath.Dir(cfgPath), 0755)
		config.Save(cfg, cfgPath)
	}
	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })
	os.Chdir(tmpdir)
	reg, err := registry.Load()
	if err != nil {
		t.Fatalf("registry.Load(): %v", err)
	}
	return reg
}

// appWithSize returns a fully initialized AppModel with the given terminal size.
func appWithSize(t *testing.T, reg *registry.Registry, w, h int) AppModel {
	t.Helper()
	m, err := NewAppModel(reg)
	if err != nil {
		t.Fatalf("NewAppModel: %v", err)
	}
	m2i, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	return m2i.(AppModel)
}

// ── panelDims ──────────────────────────────────────────────────────────────

func TestPanelDims_NormalWindow(t *testing.T) {
	m := AppModel{width: 100, height: 30}
	leftW, rightW, panelH := m.panelDims()
	if leftW < 10 {
		t.Errorf("leftW = %d, want >= 10", leftW)
	}
	if rightW < 10 {
		t.Errorf("rightW = %d, want >= 10", rightW)
	}
	if panelH < 5 {
		t.Errorf("panelH = %d, want >= 5", panelH)
	}
	// leftW + rightW should equal width
	if leftW+rightW != 100 {
		t.Errorf("leftW+rightW = %d, want %d", leftW+rightW, 100)
	}
}

func TestPanelDims_SmallWindow(t *testing.T) {
	m := AppModel{width: 5, height: 5}
	leftW, rightW, panelH := m.panelDims()
	if leftW < 10 {
		t.Errorf("leftW = %d, should clamp to at least 10", leftW)
	}
	if rightW < 10 {
		t.Errorf("rightW = %d, should clamp to at least 10", rightW)
	}
	if panelH < 5 {
		t.Errorf("panelH = %d, should clamp to at least 5", panelH)
	}
}

// ── NewAppModel ────────────────────────────────────────────────────────────

func TestNewAppModel_Empty(t *testing.T) {
	reg := loadTestReg(t, nil)
	m, err := NewAppModel(reg)
	if err != nil {
		t.Fatalf("NewAppModel() error: %v", err)
	}
	if m.state != stateCategories {
		t.Errorf("initial state should be stateCategories, got %d", m.state)
	}
	if m.reg == nil {
		t.Error("reg should not be nil")
	}
}

func TestNewAppModel_WithCategories(t *testing.T) {
	cfg := &config.Config{
		Version: "1",
		Categories: []config.Category{
			{Name: "tools", Description: "Dev tools"},
		},
	}
	reg := loadTestReg(t, cfg)
	m, err := NewAppModel(reg)
	if err != nil {
		t.Fatalf("NewAppModel() error: %v", err)
	}
	// 1 real category + 1 section header ("User") = 2 items
	if len(m.catList.Items()) != 2 {
		t.Errorf("catList should have 2 items (1 entry + 1 section header), got %d", len(m.catList.Items()))
	}
}

// ── Init ───────────────────────────────────────────────────────────────────

func TestAppModel_Init(t *testing.T) {
	reg := loadTestReg(t, nil)
	m, _ := NewAppModel(reg)
	_ = m.Init() // should not panic; returns header.Init() which is nil
}

// ── Settings ───────────────────────────────────────────────────────────────

func TestAppModel_Settings_ToggleMakefileImport(t *testing.T) {
	reg := loadTestReg(t, nil)
	m := appWithSize(t, reg, 120, 40)

	m2i, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m2 := m2i.(AppModel)
	if m2.state != stateSettings {
		t.Fatalf("state = %d, want stateSettings", m2.state)
	}

	m3i, _ := m2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m3 := m3i.(AppModel)
	if m3.state != stateCategories {
		t.Fatalf("state = %d, want stateCategories after toggle", m3.state)
	}

	cfg, err := config.LoadGlobal()
	if err != nil {
		t.Fatalf("LoadGlobal() error: %v", err)
	}
	if cfg == nil || cfg.MakefileImport == nil || *cfg.MakefileImport {
		t.Errorf("MakefileImport = %v, want false", cfg.MakefileImport)
	}
}

func TestAppModel_Settings_ToggleMakefileSource(t *testing.T) {
	reg := loadTestReg(t, nil)
	m := appWithSize(t, reg, 120, 40)

	m2i, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m2 := m2i.(AppModel)
	m2i, _ = m2.Update(tea.KeyMsg{Type: tea.KeyDown})
	m2 = m2i.(AppModel)

	m3i, _ := m2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_ = m3i.(AppModel)

	cfg, err := config.LoadGlobal()
	if err != nil {
		t.Fatalf("LoadGlobal() error: %v", err)
	}
	if cfg == nil || cfg.MakefileImportSource != "make_qp" {
		t.Errorf("MakefileImportSource = %q, want make_qp", cfg.MakefileImportSource)
	}
}

func TestAppModel_Settings_ToggleMakefileMode(t *testing.T) {
	reg := loadTestReg(t, nil)
	m := appWithSize(t, reg, 120, 40)

	m2i, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m2 := m2i.(AppModel)
	m2i, _ = m2.Update(tea.KeyMsg{Type: tea.KeyDown})
	m2 = m2i.(AppModel)
	m2i, _ = m2.Update(tea.KeyMsg{Type: tea.KeyDown})
	m2 = m2i.(AppModel)

	m3i, _ := m2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_ = m3i.(AppModel)

	cfg, err := config.LoadGlobal()
	if err != nil {
		t.Fatalf("LoadGlobal() error: %v", err)
	}
	if cfg == nil || cfg.MakefileImportMode != "includes_only" {
		t.Errorf("MakefileImportMode = %q, want includes_only", cfg.MakefileImportMode)
	}
}

func TestAppModel_Settings_RenderDetail(t *testing.T) {
	reg := loadTestReg(t, nil)
	m := appWithSize(t, reg, 120, 40)

	m2i, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m2 := m2i.(AppModel)
	if m2.state != stateSettings {
		t.Fatalf("state = %d, want stateSettings", m2.state)
	}

	detail := m2.renderSettingsDetail()
	if !strings.Contains(detail, "Auto-import Makefile targets") {
		t.Errorf("renderSettingsDetail() should include item title, got: %q", detail)
	}
}

func TestAppModel_RenderTwoPanelWithRight_NotEmpty(t *testing.T) {
	reg := loadTestReg(t, nil)
	m := appWithSize(t, reg, 120, 40)
	view := m.renderTwoPanelWithRight("left", "right", "help")
	if strings.TrimSpace(view) == "" {
		t.Error("renderTwoPanelWithRight() should not be empty")
	}
}

func TestAppModel_Settings_Esc_ReturnsCategories(t *testing.T) {
	reg := loadTestReg(t, nil)
	m := appWithSize(t, reg, 120, 40)

	m2i, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m2 := m2i.(AppModel)
	if m2.state != stateSettings {
		t.Fatalf("state = %d, want stateSettings", m2.state)
	}

	m3i, _ := m2.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m3 := m3i.(AppModel)
	if m3.state != stateCategories {
		t.Fatalf("state = %d, want stateCategories", m3.state)
	}
}

func TestAppModel_Settings_Q_QuitCmd(t *testing.T) {
	reg := loadTestReg(t, nil)
	m := appWithSize(t, reg, 120, 40)

	m2i, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m2 := m2i.(AppModel)
	if m2.state != stateSettings {
		t.Fatalf("state = %d, want stateSettings", m2.state)
	}

	_, cmd := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Error("q in settings should return quit cmd")
	}
}

func TestAppModel_RenderSettingsDetail_NoItem(t *testing.T) {
	reg := loadTestReg(t, nil)
	m := appWithSize(t, reg, 120, 40)
	m.settingsList = views.NewSettingsList(nil, 80, 20)

	detail := m.renderSettingsDetail()
	if !strings.Contains(detail, "Select a setting") {
		t.Errorf("renderSettingsDetail() should show fallback, got: %q", detail)
	}
}

func TestAppModel_ExecutionLogsFooter_RendersError(t *testing.T) {
	reg := loadTestReg(t, nil)
	m := appWithSize(t, reg, 80, 24)
	m.SetExecutionLogsVisible(true)
	m.appendExecutionLog("ERROR", "boom")
	view := m.renderExecutionLogsFooter()
	if !strings.Contains(view, "ERROR") {
		t.Errorf("footer should contain ERROR entry, got: %q", view)
	}
}

func TestAppModel_ExecutionLogsFooter_Trims(t *testing.T) {
	reg := loadTestReg(t, nil)
	m := appWithSize(t, reg, 80, 24)
	m.SetExecutionLogsVisible(true)
	for i := 0; i < executionLogWindowLines+5; i++ {
		m.appendExecutionLog("INFO", "line")
	}
	view := m.renderExecutionLogsFooter()
	if strings.Count(view, "\n") < executionLogWindowLines {
		t.Errorf("footer should render at least %d lines, got: %q", executionLogWindowLines, view)
	}
}

func TestAppModel_AppendExecutionLog_IgnoresEmpty(t *testing.T) {
	reg := loadTestReg(t, nil)
	m := appWithSize(t, reg, 80, 24)
	m.appendExecutionLog("INFO", " ")
	if len(m.executionLogs) != 0 {
		t.Errorf("expected no logs, got %d", len(m.executionLogs))
	}
}

func TestAppModel_Output_Esc_Loading_NoChange(t *testing.T) {
	reg := loadTestReg(t, nil)
	m := appWithSize(t, reg, 80, 24)
	m.state = stateOutput
	m.output = views.NewOutputModel("test", 80, 24) // loading=true
	m2i, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m2 := m2i.(AppModel)
	if m2.state != stateOutput {
		t.Errorf("Esc while loading should stay in stateOutput, got %d", m2.state)
	}
}

// ── Update: WindowSizeMsg ──────────────────────────────────────────────────

func TestAppModel_Update_WindowSize(t *testing.T) {
	reg := loadTestReg(t, nil)
	m, _ := NewAppModel(reg)
	m2i, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m2 := m2i.(AppModel)
	if m2.width != 120 || m2.height != 40 {
		t.Errorf("width/height = %d/%d, want 120/40", m2.width, m2.height)
	}
}

func TestAppModel_Update_WindowSize_InOutputState(t *testing.T) {
	reg := loadTestReg(t, nil)
	m := appWithSize(t, reg, 80, 24)
	m.state = stateOutput
	m.output = views.NewOutputModel("test", 80, 24)
	_, cmd := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	_ = cmd // should not panic
}

// ── Update: DetailHelpMsg ──────────────────────────────────────────────────

func TestAppModel_Update_DetailHelpMsg(t *testing.T) {
	reg := loadTestReg(t, nil)
	m, _ := NewAppModel(reg)
	m2i, _ := m.Update(views.DetailHelpMsg{ScriptName: "build", Text: "Usage: build"})
	_ = m2i // should not panic
}

// ── Update: ConfirmRunMsg ──────────────────────────────────────────────────

func TestAppModel_Update_ConfirmRun_NoScript(t *testing.T) {
	reg := loadTestReg(t, nil)
	m, _ := NewAppModel(reg)
	m.currentScript = nil
	m2i, _ := m.Update(views.ConfirmRunMsg{})
	_ = m2i // should not panic when currentScript is nil
}

func TestAppModel_Update_ConfirmRun_WithScript(t *testing.T) {
	cfg := &config.Config{
		Version:    "1",
		Categories: []config.Category{{Name: "tools"}},
		Scripts:    []config.Script{{Name: "build", Category: "tools", Executable: "/bin/sh", Path: "/dev/null"}},
	}
	reg := loadTestReg(t, cfg)
	m, _ := NewAppModel(reg)
	entry, err := reg.FindScript("tools", "build")
	if err != nil {
		t.Fatal(err)
	}
	m.currentScript = entry
	m2i, _ := m.Update(views.ConfirmRunMsg{ExtraArgs: []string{}})
	m2 := m2i.(AppModel)
	if m2.state != stateCategories {
		t.Errorf("ConfirmRunMsg should be ignored in current flow, got state %d", m2.state)
	}
}

// ── Update: ConfirmCancelMsg ───────────────────────────────────────────────

func TestAppModel_Update_ConfirmCancel(t *testing.T) {
	reg := loadTestReg(t, nil)
	m, _ := NewAppModel(reg)
	m.state = stateConfirm
	m2i, _ := m.Update(views.ConfirmCancelMsg{})
	m2 := m2i.(AppModel)
	if m2.state != stateConfirm {
		t.Errorf("ConfirmCancelMsg should be ignored in current flow, got state %d", m2.state)
	}
}

// ── Update: ScriptDoneMsg ──────────────────────────────────────────────────

func TestAppModel_Update_ScriptDone(t *testing.T) {
	reg := loadTestReg(t, nil)
	m, _ := NewAppModel(reg)
	m.state = stateOutput
	m.output = views.NewOutputModel("test", 80, 24)
	m2i, _ := m.Update(views.ScriptDoneMsg{Output: "done", Err: nil})
	m2 := m2i.(AppModel)
	if m2.output.IsLoading() {
		t.Error("output should not be loading after ScriptDoneMsg")
	}
}

// ── updateCategories ───────────────────────────────────────────────────────

func TestAppModel_Categories_Quit(t *testing.T) {
	reg := loadTestReg(t, nil)
	m, _ := NewAppModel(reg)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Error("q should return a non-nil cmd")
	}
}

func TestAppModel_Categories_CtrlC(t *testing.T) {
	reg := loadTestReg(t, nil)
	m, _ := NewAppModel(reg)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Error("ctrl+c should return a non-nil cmd")
	}
}

func TestAppModel_Categories_Enter_NoItem(t *testing.T) {
	reg := loadTestReg(t, nil) // empty registry
	m, _ := NewAppModel(reg)
	m2i, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := m2i.(AppModel)
	if m2.state != stateCategories {
		t.Error("Enter with no item should stay in stateCategories")
	}
}

func TestAppModel_Categories_Enter_WithItem(t *testing.T) {
	cfg := &config.Config{
		Version: "1",
		Categories: []config.Category{
			{Name: "tools", Description: "Dev tools"},
		},
	}
	reg := loadTestReg(t, cfg)
	m := appWithSize(t, reg, 80, 24)
	m2i, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := m2i.(AppModel)
	if m2.state != stateScripts {
		t.Errorf("Enter with category should navigate to stateScripts, got %d", m2.state)
	}
}

func TestAppModel_Categories_OtherKey(t *testing.T) {
	reg := loadTestReg(t, nil)
	m, _ := NewAppModel(reg)
	// Arrow down: should update the list but stay in categories
	m2i, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m2 := m2i.(AppModel)
	if m2.state != stateCategories {
		t.Error("arrow key should stay in stateCategories")
	}
}

func TestAppModel_Categories_SkipSectionHeader_Down(t *testing.T) {
	cfg := &config.Config{
		Version: "1",
		Categories: []config.Category{
			{Name: "tools"},
		},
	}
	reg := loadTestReg(t, cfg)
	m := appWithSize(t, reg, 80, 24)
	m.catList.Select(0) // force header selection

	m2i, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m2 := m2i.(AppModel)
	if _, ok := m2.catList.SelectedItem().(views.SectionHeaderItem); ok {
		t.Error("cursor should skip section header")
	}
}

// ── updateScripts ──────────────────────────────────────────────────────────

func TestAppModel_Scripts_Esc(t *testing.T) {
	reg := loadTestReg(t, nil)
	m, _ := NewAppModel(reg)
	m.state = stateScripts
	m2i, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m2 := m2i.(AppModel)
	if m2.state != stateCategories {
		t.Error("Esc in scripts should go to stateCategories")
	}
}

func TestAppModel_Scripts_Quit(t *testing.T) {
	reg := loadTestReg(t, nil)
	m, _ := NewAppModel(reg)
	m.state = stateScripts
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Error("q in scripts should return a non-nil cmd")
	}
}

func TestAppModel_Scripts_Tab_FocusesDetail(t *testing.T) {
	reg := loadTestReg(t, nil)
	m, _ := NewAppModel(reg)
	m.state = stateScripts
	m2i, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m2 := m2i.(AppModel)
	if !m2.detailFocused {
		t.Error("Tab in scripts should focus the detail panel")
	}
}

func TestAppModel_Scripts_DetailFocused_Esc(t *testing.T) {
	reg := loadTestReg(t, nil)
	m, _ := NewAppModel(reg)
	m.state = stateScripts
	m.detailFocused = true
	m2i, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m2 := m2i.(AppModel)
	if m2.detailFocused {
		t.Error("Esc with detail focused should unfocus")
	}
	// Should NOT navigate away from scripts
	if m2.state != stateScripts {
		t.Error("Esc when detail focused should stay in stateScripts")
	}
}

func TestAppModel_Scripts_DetailFocused_Tab(t *testing.T) {
	reg := loadTestReg(t, nil)
	m, _ := NewAppModel(reg)
	m.state = stateScripts
	m.detailFocused = true
	m2i, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m2 := m2i.(AppModel)
	if m2.detailFocused {
		t.Error("Tab when detail focused should unfocus")
	}
}

func TestAppModel_Scripts_DetailFocused_Quit(t *testing.T) {
	reg := loadTestReg(t, nil)
	m, _ := NewAppModel(reg)
	m.state = stateScripts
	m.detailFocused = true
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Error("q with detail focused should return a cmd")
	}
}

func TestAppModel_Scripts_Enter_NoItem(t *testing.T) {
	reg := loadTestReg(t, nil)
	m, _ := NewAppModel(reg)
	m.state = stateScripts
	m2i, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := m2i.(AppModel)
	// No item selected → should stay in stateScripts (break without transitioning)
	if m2.state != stateScripts {
		t.Error("Enter with no script item should stay in stateScripts")
	}
}

func TestAppModel_Scripts_Enter_WithItem(t *testing.T) {
	cfg := &config.Config{
		Version:    "1",
		Categories: []config.Category{{Name: "tools"}},
		Scripts:    []config.Script{{Name: "build", Category: "tools", Executable: "/bin/sh", Path: "/dev/null"}},
	}
	reg := loadTestReg(t, cfg)
	m := appWithSize(t, reg, 80, 24)
	m.state = stateScripts
	scripts := reg.GetScripts("tools")
	m.scriptList = views.NewScriptList("tools", scripts, 80, 20)

	m2i, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := m2i.(AppModel)
	if m2.state != stateScripts {
		t.Errorf("Enter with script should stay in stateScripts, got %d", m2.state)
	}
	if cmd == nil {
		t.Error("Enter with script should return non-nil run cmd")
	}
}

func TestAppModel_Scripts_Interactive_NoItem(t *testing.T) {
	reg := loadTestReg(t, nil)
	m, _ := NewAppModel(reg)
	m.state = stateScripts
	m2i, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	m2 := m2i.(AppModel)
	// No item → break; then fallthrough to scriptList.Update
	_ = m2
}

func TestAppModel_Scripts_Interactive_WithItem(t *testing.T) {
	cfg := &config.Config{
		Version:    "1",
		Categories: []config.Category{{Name: "tools"}},
		Scripts:    []config.Script{{Name: "build", Category: "tools", Executable: "/bin/sh", Path: "/bin/echo"}},
	}
	reg := loadTestReg(t, cfg)
	m := appWithSize(t, reg, 80, 24)
	m.state = stateScripts
	scripts := reg.GetScripts("tools")
	m.scriptList = views.NewScriptList("tools", scripts, 80, 20)

	m2i, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	m2 := m2i.(AppModel)
	if m2.state != stateScripts {
		t.Errorf("'i' with script should stay in stateScripts, got %d", m2.state)
	}
	if cmd == nil {
		t.Error("'i' with a script should return a non-nil run cmd")
	}
}

func TestAppModel_Scripts_Help_NoItem(t *testing.T) {
	reg := loadTestReg(t, nil)
	m, _ := NewAppModel(reg)
	m.state = stateScripts
	m2i, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	_ = m2i // should not panic
}

func TestAppModel_Scripts_Help_WithItem(t *testing.T) {
	cfg := &config.Config{
		Version:    "1",
		Categories: []config.Category{{Name: "tools"}},
		Scripts:    []config.Script{{Name: "build", Category: "tools", Executable: "/bin/sh", Path: "/dev/null"}},
	}
	reg := loadTestReg(t, cfg)
	m := appWithSize(t, reg, 80, 24)
	m.state = stateScripts
	scripts := reg.GetScripts("tools")
	m.scriptList = views.NewScriptList("tools", scripts, 80, 20)

	m2i, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m2 := m2i.(AppModel)
	if m2.state != stateScripts {
		t.Errorf("'?' with script should stay in stateScripts, got %d", m2.state)
	}
}

func TestAppModel_Scripts_Help_MakeTarget_NoCmd(t *testing.T) {
	reg := loadTestReg(t, nil)
	m := appWithSize(t, reg, 80, 24)
	m.state = stateScripts
	m.scriptList = views.NewScriptList("make", []registry.ScriptEntry{
		{Script: config.Script{Name: "build", Category: "make"}, Source: "make"},
	}, 80, 20)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if cmd != nil {
		t.Error("'?' on make target should not return a cmd")
	}
}

func TestAppModel_Scripts_OtherKey(t *testing.T) {
	reg := loadTestReg(t, nil)
	m, _ := NewAppModel(reg)
	m.state = stateScripts
	m2i, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m2 := m2i.(AppModel)
	if m2.state != stateScripts {
		t.Error("arrow key should stay in stateScripts")
	}
}

func TestAppModel_Scripts_SkipSectionHeader_Down(t *testing.T) {
	cfg := &config.Config{
		Version:    "1",
		Categories: []config.Category{{Name: "tools"}},
		Scripts:    []config.Script{{Name: "build", Category: "tools", Executable: "/bin/sh", Path: "/dev/null"}},
	}
	reg := loadTestReg(t, cfg)
	m := appWithSize(t, reg, 80, 24)
	m.state = stateScripts
	scripts := reg.GetScripts("tools")
	m.scriptList = views.NewScriptList("tools", scripts, 80, 20)
	m.scriptList.Select(0) // force header selection

	m2i, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m2 := m2i.(AppModel)
	if _, ok := m2.scriptList.SelectedItem().(views.SectionHeaderItem); ok {
		t.Error("cursor should skip section header")
	}
}

// ── updateConfirm ──────────────────────────────────────────────────────────

func TestAppModel_Confirm_Esc(t *testing.T) {
	reg := loadTestReg(t, nil)
	m, _ := NewAppModel(reg)
	m.state = stateConfirm
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd != nil {
		t.Error("Esc in stateConfirm should be ignored in current flow")
	}
}

// ── updateOutput ───────────────────────────────────────────────────────────

func TestAppModel_Output_Q_Loading_Quits(t *testing.T) {
	reg := loadTestReg(t, nil)
	m, _ := NewAppModel(reg)
	m.state = stateOutput
	m.output = views.NewOutputModel("test", 80, 24) // loading=true
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Error("q should return quit cmd in stateOutput, even while loading")
	}
}

func TestAppModel_Output_Q_NotLoading_Quits(t *testing.T) {
	reg := loadTestReg(t, nil)
	m, _ := NewAppModel(reg)
	m.state = stateOutput
	m.output = views.NewOutputModel("test", 80, 24)
	m.output, _ = m.output.Update(views.ScriptDoneMsg{Output: "done"})
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Error("q should return quit cmd in stateOutput")
	}
}

func TestAppModel_Output_Esc_NotLoading(t *testing.T) {
	reg := loadTestReg(t, nil)
	m, _ := NewAppModel(reg)
	m.state = stateOutput
	m.output = views.NewOutputModel("test", 80, 24)
	m.output, _ = m.output.Update(views.ScriptDoneMsg{Output: "done"})
	m2i, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m2 := m2i.(AppModel)
	if m2.state != stateScripts {
		t.Errorf("Esc when not loading should go to stateScripts, got %d", m2.state)
	}
}

func TestAppModel_Output_CtrlC(t *testing.T) {
	reg := loadTestReg(t, nil)
	m, _ := NewAppModel(reg)
	m.state = stateOutput
	m.output = views.NewOutputModel("test", 80, 24)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Error("ctrl+c should return quit cmd")
	}
}

func TestAppModel_Output_OtherKey(t *testing.T) {
	reg := loadTestReg(t, nil)
	m, _ := NewAppModel(reg)
	m.state = stateOutput
	m.output = views.NewOutputModel("test", 80, 24)
	m2i, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	_ = m2i // should not panic
}

func TestAppModel_Output_S_KeyIgnored(t *testing.T) {
	reg := loadTestReg(t, nil)
	m, _ := NewAppModel(reg)
	m.state = stateOutput
	m.output = views.NewOutputModel("test", 80, 24)
	m2i, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m2 := m2i.(AppModel)
	if m2.state != stateOutput {
		t.Errorf("state = %d, want stateOutput", m2.state)
	}
}

// ── View ───────────────────────────────────────────────────────────────────

func TestAppModel_View_Categories(t *testing.T) {
	reg := loadTestReg(t, nil)
	m := appWithSize(t, reg, 80, 24)
	view := m.View()
	if view == "" {
		t.Error("View() in stateCategories should not be empty")
	}
}

func TestAppModel_View_Scripts(t *testing.T) {
	reg := loadTestReg(t, nil)
	m := appWithSize(t, reg, 80, 24)
	m.state = stateScripts
	view := m.View()
	if view == "" {
		t.Error("View() in stateScripts should not be empty")
	}
}

func TestAppModel_View_Scripts_DetailFocused(t *testing.T) {
	reg := loadTestReg(t, nil)
	m := appWithSize(t, reg, 80, 24)
	m.state = stateScripts
	m.detailFocused = true
	view := m.View()
	_ = view // should not panic
}

func TestAppModel_View_Confirm(t *testing.T) {
	reg := loadTestReg(t, nil)
	m := appWithSize(t, reg, 80, 24)
	m.state = stateConfirm
	view := m.View()
	_ = view // should not panic (ConfirmModel might have zero entry)
}

func TestAppModel_View_Output(t *testing.T) {
	reg := loadTestReg(t, nil)
	m := appWithSize(t, reg, 80, 24)
	m.output = views.NewOutputModel("test", 80, 24)
	m.state = stateOutput
	view := m.View()
	if view == "" {
		t.Error("View() in stateOutput should not be empty")
	}
}

// ── renderTwoPanel ─────────────────────────────────────────────────────────

func TestAppModel_RenderTwoPanel(t *testing.T) {
	reg := loadTestReg(t, nil)
	m := appWithSize(t, reg, 80, 24)
	m.detail.SetSize(30, 18)
	view := m.renderTwoPanel("left content", "help text")
	if view == "" {
		t.Error("renderTwoPanel should not return empty string")
	}
}

func TestAppModel_RenderTwoPanel_DetailFocused(t *testing.T) {
	reg := loadTestReg(t, nil)
	m := appWithSize(t, reg, 80, 24)
	m.detailFocused = true
	view := m.renderTwoPanel("left", "help")
	_ = view // should not panic; border colors differ from non-focused
}

func TestAppModel_View_DoesNotOverflowWindowHeight(t *testing.T) {
	reg := loadTestReg(t, nil)
	m := appWithSize(t, reg, 80, 24)

	view := m.View()
	if lines := strings.Count(view, "\n") + 1; lines > m.height {
		t.Fatalf("view height overflow: got %d lines, window height %d", lines, m.height)
	}
}

func TestAppModel_ToggleExecutionLogs(t *testing.T) {
	reg := loadTestReg(t, nil)
	m := appWithSize(t, reg, 80, 24)

	m2i, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	m2 := m2i.(AppModel)
	if !m2.showExecutionLogs {
		t.Error("'l' should enable execution logs footer")
	}
	if len(m2.executionLogs) == 0 {
		t.Error("enabling logs should register an execution event")
	}

	m3i, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	m3 := m3i.(AppModel)
	if m3.showExecutionLogs {
		t.Error("second 'l' should disable execution logs footer")
	}
}

func TestAppModel_View_WithExecutionLogsFooter(t *testing.T) {
	reg := loadTestReg(t, nil)
	m := appWithSize(t, reg, 80, 24)
	m.SetExecutionLogsVisible(true)
	view := m.View()
	if !strings.Contains(view, "execution logs") {
		t.Error("view should include execution logs footer when enabled")
	}
}

func TestAppModel_View_ExecutionLogsFooter_DoesNotMutateEntries(t *testing.T) {
	reg := loadTestReg(t, nil)
	m := appWithSize(t, reg, 80, 24)
	m.SetExecutionLogsVisible(true)

	_ = m.View()
	_ = m.View()

	if len(m.executionLogs) != 1 {
		t.Fatalf("execution logs should keep a single entry after repeated renders, got %d", len(m.executionLogs))
	}
	if strings.Count(m.executionLogs[0], "[INFO]") != 1 {
		t.Fatalf("log entry should not be rewritten on render, got %q", m.executionLogs[0])
	}
}

// ── syncDetail ─────────────────────────────────────────────────────────────

func TestAppModel_SyncDetail_CategoriesState_NoItem(t *testing.T) {
	reg := loadTestReg(t, nil)
	m, _ := NewAppModel(reg)
	m.state = stateCategories
	m2, _ := m.syncDetail()
	_ = m2 // should not panic
}

func TestAppModel_SyncDetail_CategoriesState_WithItem(t *testing.T) {
	cfg := &config.Config{
		Version:    "1",
		Categories: []config.Category{{Name: "tools"}},
	}
	reg := loadTestReg(t, cfg)
	m, _ := NewAppModel(reg)
	m.state = stateCategories
	m2, _ := m.syncDetail()
	_ = m2 // should not panic
}

func TestAppModel_SyncDetail_ScriptsState_NoItem(t *testing.T) {
	reg := loadTestReg(t, nil)
	m, _ := NewAppModel(reg)
	m.state = stateScripts
	m2, _ := m.syncDetail()
	_ = m2 // should not panic
}

func TestAppModel_SyncDetail_ScriptsState_WithItem(t *testing.T) {
	cfg := &config.Config{
		Version:    "1",
		Categories: []config.Category{{Name: "tools"}},
		Scripts:    []config.Script{{Name: "build", Category: "tools", Executable: "/bin/sh", Path: "/dev/null"}},
	}
	reg := loadTestReg(t, cfg)
	m, _ := NewAppModel(reg)
	m.state = stateScripts
	scripts := reg.GetScripts("tools")
	m.scriptList = views.NewScriptList("tools", scripts, 80, 20)
	m2, _ := m.syncDetail()
	_ = m2 // should not panic
}

func TestAppModel_SyncDetail_OtherState(t *testing.T) {
	reg := loadTestReg(t, nil)
	m, _ := NewAppModel(reg)
	m.state = stateOutput
	m2, _ := m.syncDetail()
	_ = m2 // should return immediately
}

// ── non-key messages to catList/scriptList ─────────────────────────────────

func TestAppModel_Update_NonKey_CategoriesState(t *testing.T) {
	reg := loadTestReg(t, nil)
	m, _ := NewAppModel(reg)
	m.state = stateCategories
	// Send a non-key, non-window message (e.g., a spinner tick)
	// The model should delegate to catList.Update
	m2i, _ := m.Update(struct{}{})
	_ = m2i
}

func TestAppModel_Update_NonKey_ScriptsState(t *testing.T) {
	reg := loadTestReg(t, nil)
	m, _ := NewAppModel(reg)
	m.state = stateScripts
	m2i, _ := m.Update(struct{}{})
	_ = m2i
}

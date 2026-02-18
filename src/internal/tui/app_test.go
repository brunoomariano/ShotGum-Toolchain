package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/brunoomariano/ShotGum-Toolchain/internal/config"
	"github.com/brunoomariano/ShotGum-Toolchain/internal/registry"
	"github.com/brunoomariano/ShotGum-Toolchain/internal/runner"
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
	// leftW + rightW should equal width-4
	if leftW+rightW != 100-4 {
		t.Errorf("leftW+rightW = %d, want %d", leftW+rightW, 100-4)
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
	if m2.state != stateOutput {
		t.Errorf("Enter with script should go to stateOutput, got %d", m2.state)
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
	if m2.state != stateOutput {
		t.Errorf("'i' with script should go to stateOutput, got %d", m2.state)
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
	if m2.state != stateOutput {
		t.Errorf("'?' with script should go to stateOutput, got %d", m2.state)
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

func TestAppModel_Update_InteractiveChunkMsg_AppendsOutput(t *testing.T) {
	reg := loadTestReg(t, nil)
	m, _ := NewAppModel(reg)
	m.state = stateOutput
	m.output = views.NewOutputModel("test", 80, 24)
	m.output.EnableInteractive()

	m2i, _ := m.Update(interactiveChunkMsg{text: "hello from chunk"})
	m2 := m2i.(AppModel)

	view := m2.output.View()
	if !strings.Contains(view, "hello from chunk") {
		t.Fatalf("output view should contain streamed chunk, got: %q", view)
	}
}

func TestAppModel_Update_InteractiveDoneMsg_FinishesOutput(t *testing.T) {
	reg := loadTestReg(t, nil)
	m, _ := NewAppModel(reg)
	m.state = stateOutput
	m.output = views.NewOutputModel("test", 80, 24)
	m.output.EnableInteractive()

	m2i, _ := m.Update(interactiveDoneMsg{err: nil})
	m2 := m2i.(AppModel)
	if m2.output.IsLoading() {
		t.Fatal("output should stop loading after interactiveDoneMsg")
	}
}

func TestKeyMsgToInput_Mappings(t *testing.T) {
	tests := []struct {
		name string
		msg  tea.KeyMsg
		want string
		ok   bool
	}{
		{name: "runes", msg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}, want: "a", ok: true},
		{name: "enter", msg: tea.KeyMsg{Type: tea.KeyEnter}, want: "\n", ok: true},
		{name: "space", msg: tea.KeyMsg{Type: tea.KeySpace}, want: " ", ok: true},
		{name: "tab", msg: tea.KeyMsg{Type: tea.KeyTab}, want: "\t", ok: true},
		{name: "backspace", msg: tea.KeyMsg{Type: tea.KeyBackspace}, want: string([]byte{127}), ok: true},
		{name: "up", msg: tea.KeyMsg{Type: tea.KeyUp}, want: "\x1b[A", ok: true},
		{name: "down", msg: tea.KeyMsg{Type: tea.KeyDown}, want: "\x1b[B", ok: true},
		{name: "left", msg: tea.KeyMsg{Type: tea.KeyLeft}, want: "\x1b[D", ok: true},
		{name: "right", msg: tea.KeyMsg{Type: tea.KeyRight}, want: "\x1b[C", ok: true},
		{name: "ctrl+c", msg: tea.KeyMsg{Type: tea.KeyCtrlC}, want: string([]byte{3}), ok: true},
		{name: "unsupported", msg: tea.KeyMsg{Type: tea.KeyEsc}, want: "", ok: false},
	}

	for _, tc := range tests {
		got, ok := keyMsgToInput(tc.msg)
		if ok != tc.ok {
			t.Fatalf("%s: ok=%v want %v", tc.name, ok, tc.ok)
		}
		if string(got) != tc.want {
			t.Fatalf("%s: got %q want %q", tc.name, string(got), tc.want)
		}
	}
}

func TestSanitizeInteractiveChunk_RemovesOSC(t *testing.T) {
	in := "\x1b]11;rgb:1a1a/1b1b/2626\x1b\\  +---+"
	got := sanitizeInteractiveChunk(in)
	if strings.Contains(got, "]11;rgb:") {
		t.Fatalf("sanitizeInteractiveChunk should remove OSC payload, got: %q", got)
	}
	if !strings.Contains(got, "+---+") {
		t.Fatalf("sanitizeInteractiveChunk should preserve regular output, got: %q", got)
	}
}

func TestSanitizeInteractiveChunk_RemovesLeakedOSCText(t *testing.T) {
	in := "]11;rgb:1a1a/1b1b/2626\\  |   |"
	got := sanitizeInteractiveChunk(in)
	if strings.Contains(got, "]11;rgb:") {
		t.Fatalf("sanitizeInteractiveChunk should remove leaked OSC text, got: %q", got)
	}
	if !strings.Contains(got, "|   |") {
		t.Fatalf("sanitizeInteractiveChunk should preserve regular output, got: %q", got)
	}
}

func TestReadInteractiveChunkCmd_AndWaitDone(t *testing.T) {
	tmpdir := t.TempDir()
	script := filepath.Join(tmpdir, "hello.sh")
	if err := os.WriteFile(script, []byte("#!/usr/bin/env bash\necho hello-live"), 0755); err != nil {
		t.Fatalf("write script: %v", err)
	}

	cfg := &config.Config{
		Version:    "1",
		Categories: []config.Category{{Name: "tools"}},
		Scripts: []config.Script{{
			Name:       "hello",
			Category:   "tools",
			Executable: "/bin/bash",
			Path:       script,
		}},
	}
	reg := loadTestReg(t, cfg)
	entry, err := reg.FindScript("tools", "hello")
	if err != nil {
		t.Fatalf("FindScript: %v", err)
	}

	session, err := runner.StartInteractive(*entry, nil, reg)
	if err != nil {
		t.Fatalf("StartInteractive: %v", err)
	}

	msg := readInteractiveChunkCmd(session)()
	chunk, ok := msg.(interactiveChunkMsg)
	if !ok {
		t.Fatalf("readInteractiveChunkCmd returned %T", msg)
	}
	if strings.TrimSpace(chunk.text) == "" {
		t.Fatalf("expected non-empty interactive chunk")
	}

	doneMsg := waitInteractiveDoneCmd(session)()
	done, ok := doneMsg.(interactiveDoneMsg)
	if !ok {
		t.Fatalf("waitInteractiveDoneCmd returned %T", doneMsg)
	}
	if done.err != nil {
		t.Fatalf("unexpected done error: %v", done.err)
	}
}

func TestAppModel_UpdateOutput_ForwardsInteractiveInput(t *testing.T) {
	tmpdir := t.TempDir()
	script := filepath.Join(tmpdir, "read.sh")
	content := "#!/usr/bin/env bash\nread -r x\necho typed:$x"
	if err := os.WriteFile(script, []byte(content), 0755); err != nil {
		t.Fatalf("write script: %v", err)
	}

	cfg := &config.Config{
		Version:    "1",
		Categories: []config.Category{{Name: "tools"}},
		Scripts: []config.Script{{
			Name:       "reader",
			Category:   "tools",
			Executable: "/bin/bash",
			Path:       script,
		}},
	}
	reg := loadTestReg(t, cfg)
	entry, err := reg.FindScript("tools", "reader")
	if err != nil {
		t.Fatalf("FindScript: %v", err)
	}

	session, err := runner.StartInteractive(*entry, nil, reg)
	if err != nil {
		t.Fatalf("StartInteractive: %v", err)
	}

	m, err := NewAppModel(reg)
	if err != nil {
		t.Fatalf("NewAppModel: %v", err)
	}
	m.state = stateOutput
	m.output = views.NewOutputModel("reader", 80, 24)
	m.output.EnableInteractive()
	m.liveSession = session

	m2i, _ := m.updateOutput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	m2 := m2i.(AppModel)
	if !m2.output.IsLoading() {
		t.Fatalf("output should still be loading during interactive run")
	}

	_, _ = m2.updateOutput(tea.KeyMsg{Type: tea.KeyEnter})

	select {
	case <-session.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting interactive process after forwarded input")
	}
}

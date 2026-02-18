package tui

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shotgum/stg/internal/config"
	"github.com/shotgum/stg/internal/registry"
	"github.com/shotgum/stg/internal/tui/views"
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
	if len(m.catList.Items()) != 1 {
		t.Errorf("catList should have 1 item, got %d", len(m.catList.Items()))
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

// ── Update: interactiveDoneMsg ─────────────────────────────────────────────

func TestAppModel_Update_InteractiveDone(t *testing.T) {
	reg := loadTestReg(t, nil)
	m, _ := NewAppModel(reg)
	m2i, _ := m.Update(interactiveDoneMsg{err: nil})
	m2 := m2i.(AppModel)
	if m2.state != stateScripts {
		t.Errorf("after interactiveDone, state should be stateScripts, got %d", m2.state)
	}
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
		Scripts:    []config.Script{{Name: "build", Category: "tools", Type: "script", Path: "/dev/null"}},
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
	if m2.state != stateOutput {
		t.Errorf("after ConfirmRun, state should be stateOutput, got %d", m2.state)
	}
}

// ── Update: ConfirmCancelMsg ───────────────────────────────────────────────

func TestAppModel_Update_ConfirmCancel(t *testing.T) {
	reg := loadTestReg(t, nil)
	m, _ := NewAppModel(reg)
	m.state = stateConfirm
	m2i, _ := m.Update(views.ConfirmCancelMsg{})
	m2 := m2i.(AppModel)
	if m2.state != stateScripts {
		t.Errorf("after ConfirmCancel, state should be stateScripts, got %d", m2.state)
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
		Scripts:    []config.Script{{Name: "build", Category: "tools", Type: "script", Path: "/dev/null"}},
	}
	reg := loadTestReg(t, cfg)
	m := appWithSize(t, reg, 80, 24)
	m.state = stateScripts
	scripts := reg.GetScripts("tools")
	m.scriptList = views.NewScriptList("tools", scripts, 80, 20)

	m2i, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := m2i.(AppModel)
	if m2.state != stateConfirm {
		t.Errorf("Enter with script should navigate to stateConfirm, got %d", m2.state)
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
		Scripts:    []config.Script{{Name: "build", Category: "tools", Type: "script", Path: "/bin/echo"}},
	}
	reg := loadTestReg(t, cfg)
	m := appWithSize(t, reg, 80, 24)
	m.state = stateScripts
	scripts := reg.GetScripts("tools")
	m.scriptList = views.NewScriptList("tools", scripts, 80, 20)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	// cmd should be a tea.ExecProcess command (non-nil)
	if cmd == nil {
		t.Error("'i' with a script should return a non-nil ExecProcess cmd")
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
		Scripts:    []config.Script{{Name: "build", Category: "tools", Type: "script", Path: "/dev/null"}},
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
	// Esc in confirm → ConfirmCancelMsg is returned via cmd
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Error("Esc in stateConfirm should return a non-nil cmd")
	}
}

// ── updateOutput ───────────────────────────────────────────────────────────

func TestAppModel_Output_Q_Loading(t *testing.T) {
	reg := loadTestReg(t, nil)
	m, _ := NewAppModel(reg)
	m.state = stateOutput
	m.output = views.NewOutputModel("test", 80, 24) // loading=true
	m2i, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m2 := m2i.(AppModel)
	if m2.state != stateOutput {
		t.Error("q while loading should stay in stateOutput")
	}
}

func TestAppModel_Output_Q_NotLoading(t *testing.T) {
	reg := loadTestReg(t, nil)
	m, _ := NewAppModel(reg)
	m.state = stateOutput
	m.output = views.NewOutputModel("test", 80, 24)
	m.output, _ = m.output.Update(views.ScriptDoneMsg{Output: "done"})
	m2i, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m2 := m2i.(AppModel)
	if m2.state != stateScripts {
		t.Errorf("q when not loading should go to stateScripts, got %d", m2.state)
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
		Scripts:    []config.Script{{Name: "build", Category: "tools", Type: "script", Path: "/dev/null"}},
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

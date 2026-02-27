package tui

import (
	"fmt"

	"github.com/brunoomariano/ShotGum-Toolchain/internal/config"
	"github.com/brunoomariano/ShotGum-Toolchain/internal/registry"
	"github.com/brunoomariano/ShotGum-Toolchain/internal/tui/views"
	"github.com/charmbracelet/bubbles/list"
)

const (
	settingMakefileImport       = "makefile_import"
	settingMakefileImportMode   = "makefile_import_mode"
	settingMakefileImportSource = "makefile_import_source"
)

func buildSettingsItems(reg *registry.Registry) []list.Item {
	cfg := reg.GlobalConfig()
	enabled := config.EffectiveMakefileImport(cfg)
	mode := config.EffectiveMakefileImportMode(cfg)

	modeValue := "All (except #shotgum:ignore)"
	if mode == "includes_only" {
		modeValue = "Only #shotgum:includes"
	}
	if !enabled {
		modeValue = modeValue + " (disabled)"
	}
	source := config.EffectiveMakefileImportSource(cfg)
	sourceValue := "Simple parser"
	if source == "make_qp" {
		sourceValue = "make -qp -rR"
	}
	if !enabled {
		sourceValue = sourceValue + " (disabled)"
	}

	items := []list.Item{
		views.SettingItem{
			Key:       settingMakefileImport,
			TitleText: "Auto-import Makefile targets",
			DescText:  "Read targets from Makefile in current directory.",
			Value:     onOff(enabled),
		},
		views.SettingItem{
			Key:       settingMakefileImportSource,
			TitleText: "Makefile discovery engine",
			DescText:  "Choose how targets are detected.",
			Value:     sourceValue,
		},
		views.SettingItem{
			Key:       settingMakefileImportMode,
			TitleText: "Makefile import mode",
			DescText:  "Filter targets by comment directives.",
			Value:     modeValue,
		},
	}
	return items
}

func onOff(v bool) string {
	if v {
		return "ON"
	}
	return "OFF"
}

func saveMakefileSettings(enabled bool, mode string, source string) error {
	if err := config.EnsureDefault(); err != nil {
		return fmt.Errorf("ensuring default config: %w", err)
	}
	cfg, err := config.LoadGlobal()
	if err != nil {
		return fmt.Errorf("loading global config: %w", err)
	}
	if cfg == nil {
		cfg = &config.Config{Version: "1", HelpFlag: "--help"}
	}
	cfg.MakefileImport = &enabled
	cfg.MakefileImportMode = mode
	if source == "" {
		source = "parser"
	}
	cfg.MakefileImportSource = source
	return config.Save(cfg, config.GlobalConfigPath())
}

func (m *AppModel) reloadRegistry() {
	reg, err := registry.Load()
	if err != nil {
		m.appendExecutionLog("ERROR", fmt.Sprintf("reloading registry: %v", err))
		return
	}
	m.reg = reg
	leftW, _, panelH := m.panelDims()
	m.catList = views.NewCategoryList(reg.GetCategories(), leftW, panelH)
	m.scriptList = views.NewScriptList("", []registry.ScriptEntry{}, leftW, panelH)
	m.settingsList = views.NewSettingsList(buildSettingsItems(reg), leftW, panelH)
	m.state = stateCategories
	m.detailFocused = false
	m.currentCategory = nil
	m.currentScript = nil
	updated, _ := m.syncDetail()
	*m = updated
}

package registry

import (
	"testing"

	"github.com/brunoomariano/ShotGum-Toolchain/internal/config"
)

func makeRegistry(global, local *config.Config) *Registry {
	return &Registry{global: global, local: local}
}

func TestGetCategories_BothNil(t *testing.T) {
	reg := makeRegistry(nil, nil)
	cats := reg.GetCategories()
	if len(cats) != 0 {
		t.Errorf("expected 0 categories, got %d", len(cats))
	}
}

func TestGetCategories_GlobalOnly(t *testing.T) {
	global := &config.Config{
		Categories: []config.Category{
			{Name: "tools", Description: "tools desc"},
		},
	}
	reg := makeRegistry(global, nil)
	cats := reg.GetCategories()
	if len(cats) != 1 {
		t.Fatalf("expected 1 category, got %d", len(cats))
	}
	if cats[0].Name != "tools" || cats[0].Source != "user" {
		t.Errorf("unexpected category: %+v", cats[0])
	}
}

func TestGetCategories_LocalOnly(t *testing.T) {
	local := &config.Config{
		Categories: []config.Category{
			{Name: "local-cat", Description: "local desc"},
		},
	}
	reg := makeRegistry(nil, local)
	cats := reg.GetCategories()
	if len(cats) != 1 {
		t.Fatalf("expected 1 category, got %d", len(cats))
	}
	if cats[0].Name != "local-cat" || cats[0].Source != "local" {
		t.Errorf("unexpected category: %+v", cats[0])
	}
}

func TestGetCategories_MergeLocalOverride(t *testing.T) {
	global := &config.Config{
		Categories: []config.Category{
			{Name: "shared", Description: "global desc"},
			{Name: "global-only", Description: "global only"},
		},
	}
	local := &config.Config{
		Categories: []config.Category{
			{Name: "shared", Description: "local desc"},
		},
	}
	reg := makeRegistry(global, local)
	cats := reg.GetCategories()
	if len(cats) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(cats))
	}
	if cats[0].Name != "shared" || cats[0].Source != "local" {
		t.Errorf("first category should be local shared, got %+v", cats[0])
	}
	if cats[1].Name != "global-only" || cats[1].Source != "user" {
		t.Errorf("second category should be user-only, got %+v", cats[1])
	}
}

func TestGetScripts_NonexistentCategory(t *testing.T) {
	global := &config.Config{
		Scripts: []config.Script{
			{Name: "build", Category: "tools", Executable: "/bin/sh"},
		},
	}
	reg := makeRegistry(global, nil)
	scripts := reg.GetScripts("nonexistent")
	if len(scripts) != 0 {
		t.Errorf("expected 0 scripts, got %d", len(scripts))
	}
}

func TestGetScripts_MergeLocalOverride(t *testing.T) {
	global := &config.Config{
		Scripts: []config.Script{
			{Name: "build", Category: "tools", Executable: "/bin/sh", Description: "global build"},
			{Name: "test", Category: "tools", Executable: "/bin/sh"},
		},
	}
	local := &config.Config{
		Scripts: []config.Script{
			{Name: "build", Category: "tools", Executable: "/bin/sh", Description: "local build"},
		},
	}
	reg := makeRegistry(global, local)
	scripts := reg.GetScripts("tools")
	if len(scripts) != 2 {
		t.Fatalf("expected 2 scripts, got %d", len(scripts))
	}
	if scripts[0].Name != "build" || scripts[0].Source != "local" {
		t.Errorf("first script should be local build, got %+v", scripts[0])
	}
	if scripts[1].Name != "test" || scripts[1].Source != "user" {
		t.Errorf("second script should be user test, got %+v", scripts[1])
	}
}

func TestFindScript_Local(t *testing.T) {
	local := &config.Config{
		Scripts: []config.Script{
			{Name: "deploy", Category: "ops", Executable: "/bin/sh"},
		},
	}
	reg := makeRegistry(nil, local)
	entry, err := reg.FindScript("ops", "deploy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Source != "local" {
		t.Errorf("expected source=local, got %q", entry.Source)
	}
}

func TestFindScript_Global(t *testing.T) {
	global := &config.Config{
		Scripts: []config.Script{
			{Name: "lint", Category: "dev", Executable: "/bin/sh"},
		},
	}
	reg := makeRegistry(global, nil)
	entry, err := reg.FindScript("dev", "lint")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Source != "user" {
		t.Errorf("expected source=user, got %q", entry.Source)
	}
}

func TestFindScript_NotFound(t *testing.T) {
	reg := makeRegistry(nil, nil)
	_, err := reg.FindScript("cat", "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent script, got nil")
	}
}

func TestResolveScriptPath_AbsolutePath(t *testing.T) {
	reg := makeRegistry(nil, nil)
	entry := ScriptEntry{
		Script: config.Script{
			Name:     "test",
			Category: "dev",
			Path:     "/absolute/path/to/script.sh",
		},
		Source: "global",
	}
	got := reg.ResolveScriptPath(entry)
	if got != "/absolute/path/to/script.sh" {
		t.Errorf("expected absolute path, got %q", got)
	}
}

func TestResolveScriptPath_RelativeWithCategoryScriptsPath(t *testing.T) {
	global := &config.Config{
		Categories: []config.Category{
			{Name: "dev", ScriptsPath: "/scripts/dev"},
		},
	}
	reg := makeRegistry(global, nil)
	entry := ScriptEntry{
		Script: config.Script{
			Name:     "build",
			Category: "dev",
			Path:     "build.sh",
		},
		Source: "global",
	}
	got := reg.ResolveScriptPath(entry)
	if got != "/scripts/dev/build.sh" {
		t.Errorf("expected /scripts/dev/build.sh, got %q", got)
	}
}

func TestResolveScriptPath_RelativeWithScriptsHome(t *testing.T) {
	global := &config.Config{
		ScriptsHome: "/home/user/scripts",
	}
	reg := makeRegistry(global, nil)
	entry := ScriptEntry{
		Script: config.Script{
			Name:     "build",
			Category: "dev",
			Path:     "build.sh",
		},
		Source: "global",
	}
	got := reg.ResolveScriptPath(entry)
	if got != "/home/user/scripts/dev/build.sh" {
		t.Errorf("expected /home/user/scripts/dev/build.sh, got %q", got)
	}
}

func TestResolveHelpFlag_ScriptLevel(t *testing.T) {
	global := &config.Config{
		HelpFlag: "--global-help",
		Categories: []config.Category{
			{Name: "dev", HelpFlag: "--cat-help"},
		},
	}
	reg := makeRegistry(global, nil)
	entry := ScriptEntry{
		Script: config.Script{
			Name:     "build",
			Category: "dev",
			HelpFlag: "--script-help",
		},
		Source: "global",
	}
	if got := reg.ResolveHelpFlag(entry); got != "--script-help" {
		t.Errorf("expected --script-help, got %q", got)
	}
}

func TestResolveHelpFlag_CategoryLevel(t *testing.T) {
	global := &config.Config{
		HelpFlag: "--global-help",
		Categories: []config.Category{
			{Name: "dev", HelpFlag: "--cat-help"},
		},
	}
	reg := makeRegistry(global, nil)
	entry := ScriptEntry{
		Script: config.Script{
			Name:     "build",
			Category: "dev",
		},
		Source: "global",
	}
	if got := reg.ResolveHelpFlag(entry); got != "--cat-help" {
		t.Errorf("expected --cat-help, got %q", got)
	}
}

func TestResolveHelpFlag_ConfigLevel(t *testing.T) {
	global := &config.Config{
		HelpFlag: "--global-help",
	}
	reg := makeRegistry(global, nil)
	entry := ScriptEntry{
		Script: config.Script{
			Name:     "build",
			Category: "dev",
		},
		Source: "global",
	}
	if got := reg.ResolveHelpFlag(entry); got != "--global-help" {
		t.Errorf("expected --global-help, got %q", got)
	}
}

func TestResolveHelpFlag_DefaultHelp(t *testing.T) {
	reg := makeRegistry(nil, nil)
	entry := ScriptEntry{
		Script: config.Script{
			Name:     "build",
			Category: "dev",
		},
		Source: "global",
	}
	if got := reg.ResolveHelpFlag(entry); got != "--help" {
		t.Errorf("expected --help, got %q", got)
	}
}

func TestGlobalConfig(t *testing.T) {
	global := &config.Config{Version: "1"}
	reg := makeRegistry(global, nil)
	if reg.GlobalConfig() != global {
		t.Error("GlobalConfig() should return the global config pointer")
	}
}

func TestLocalConfig(t *testing.T) {
	local := &config.Config{Version: "1"}
	reg := makeRegistry(nil, local)
	if reg.LocalConfig() != local {
		t.Error("LocalConfig() should return the local config pointer")
	}
}

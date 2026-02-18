package registry

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shotgum/stg/internal/config"
)

func TestLoad_EmptyHome(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)
	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })
	os.Chdir(tmpdir)

	reg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if reg == nil {
		t.Fatal("Load() returned nil registry")
	}
	// Both configs should be nil (no files found)
	if reg.GlobalConfig() != nil {
		t.Error("GlobalConfig() should be nil with empty home")
	}
	if reg.LocalConfig() != nil {
		t.Error("LocalConfig() should be nil with empty dir")
	}
}

func TestLoad_WithGlobalConfig(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)
	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })
	os.Chdir(tmpdir)

	cfg := &config.Config{
		Version:    "1",
		Categories: []config.Category{{Name: "tools"}},
	}
	cfgPath := filepath.Join(tmpdir, ".config", "shotgum", "config.yaml")
	os.MkdirAll(filepath.Dir(cfgPath), 0755)
	config.Save(cfg, cfgPath)

	reg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if reg.GlobalConfig() == nil {
		t.Error("GlobalConfig() should not be nil after writing global config")
	}
	if len(reg.GetCategories()) != 1 {
		t.Errorf("expected 1 category, got %d", len(reg.GetCategories()))
	}
}

func TestScriptsHome_FallbackToGlobal(t *testing.T) {
	// When source="local" and local has no ScriptsHome, should fallback to global.ScriptsHome
	global := &config.Config{ScriptsHome: "/global/scripts"}
	local := &config.Config{} // no ScriptsHome
	reg := makeRegistry(global, local)

	// Local entry with relative path, local config has no ScriptsHome
	entry := ScriptEntry{
		Script: config.Script{
			Name:     "build",
			Category: "dev",
			Path:     "build.sh",
			Type:     "script",
		},
		Source: "local",
	}
	got := reg.ResolveScriptPath(entry)
	// Should fallback to global.ScriptsHome/category/path
	if got != "/global/scripts/dev/build.sh" {
		t.Errorf("ResolveScriptPath = %q, want /global/scripts/dev/build.sh", got)
	}
}

func TestConfigFor_Local(t *testing.T) {
	local := &config.Config{Version: "local"}
	reg := makeRegistry(nil, local)
	got := reg.configFor("local")
	if got != local {
		t.Error("configFor('local') should return local config")
	}
}

func TestConfigFor_Global(t *testing.T) {
	global := &config.Config{Version: "global"}
	reg := makeRegistry(global, nil)
	got := reg.configFor("global")
	if got != global {
		t.Error("configFor('global') should return global config")
	}
}

func TestScriptsHome_NoConfigs(t *testing.T) {
	reg := makeRegistry(nil, nil)
	// Should return "" when both configs are nil
	entry := ScriptEntry{
		Script: config.Script{Name: "build", Category: "dev", Path: "build.sh"},
		Source: "global",
	}
	got := reg.ResolveScriptPath(entry)
	// Falls back to scriptsHome="", so path is just "dev/build.sh"
	if got == "" {
		t.Error("ResolveScriptPath should return a non-empty path")
	}
}

// TestLoad_GlobalConfigError verifies that Load() propagates the error returned by
// LoadGlobal when the config path exists but is not a readable file (e.g. a directory).
func TestLoad_GlobalConfigError(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)
	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })
	os.Chdir(tmpdir)

	// Place a directory at the exact path where config.yaml should be.
	// os.ReadFile on a directory returns "is a directory", not os.IsNotExist.
	configDir := filepath.Join(tmpdir, ".config", "shotgum")
	os.MkdirAll(configDir, 0755)
	os.MkdirAll(filepath.Join(configDir, "config.yaml"), 0755)

	_, err := Load()
	if err == nil {
		t.Error("Load() should fail when global config path is a directory")
	}
}

// TestLoad_LocalConfigError verifies that Load() propagates the error returned by
// LoadLocal when .shotgum.yaml in CWD exists as a directory (unreadable as a file).
func TestLoad_LocalConfigError(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir) // no global config → LoadGlobal succeeds with nil
	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })
	os.Chdir(tmpdir)

	// Place a directory named .shotgum.yaml so ReadFile fails with "is a directory".
	os.MkdirAll(filepath.Join(tmpdir, ".shotgum.yaml"), 0755)

	_, err := Load()
	if err == nil {
		t.Error("Load() should fail when .shotgum.yaml is a directory")
	}
}

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSave_CreatesDirectoryAndFile(t *testing.T) {
	tmpdir := t.TempDir()
	path := filepath.Join(tmpdir, "subdir", "config.yaml")
	cfg := &Config{Version: "1", HelpFlag: "--help"}
	if err := Save(cfg, path); err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("config file not created: %v", err)
	}
}

func TestSave_InvalidPath(t *testing.T) {
	tmpdir := t.TempDir()
	// Create a file where we want a directory
	filePath := filepath.Join(tmpdir, "notadir")
	os.WriteFile(filePath, []byte("x"), 0644)
	// Try to save inside the file (impossible path)
	path := filepath.Join(filePath, "config.yaml")
	cfg := &Config{Version: "1"}
	err := Save(cfg, path)
	if err == nil {
		t.Error("Save() should return error for invalid path")
	}
}

func TestLoadGlobal_FileNotExist(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)
	cfg, err := LoadGlobal()
	if err != nil {
		t.Fatalf("LoadGlobal() unexpected error: %v", err)
	}
	if cfg != nil {
		t.Error("LoadGlobal() should return nil when file doesn't exist")
	}
}

func TestLoadGlobal_ValidYAML(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)
	cfgPath := globalConfigPath()
	os.MkdirAll(filepath.Dir(cfgPath), 0755)
	yamlContent := "version: \"1\"\nhelp_flag: \"--help\"\ncategories:\n  - name: tools\n    description: My tools\n"
	os.WriteFile(cfgPath, []byte(yamlContent), 0644)

	cfg, err := LoadGlobal()
	if err != nil {
		t.Fatalf("LoadGlobal() error: %v", err)
	}
	if cfg == nil {
		t.Fatal("LoadGlobal() returned nil for existing file")
	}
	if cfg.HelpFlag != "--help" {
		t.Errorf("HelpFlag = %q, want --help", cfg.HelpFlag)
	}
	if len(cfg.Categories) != 1 || cfg.Categories[0].Name != "tools" {
		t.Errorf("unexpected categories: %+v", cfg.Categories)
	}
	if cfg.Source != "global" {
		t.Errorf("Source = %q, want global", cfg.Source)
	}
}

func TestLoadGlobal_MalformedYAML(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)
	cfgPath := globalConfigPath()
	os.MkdirAll(filepath.Dir(cfgPath), 0755)
	os.WriteFile(cfgPath, []byte(":\n  bad: yaml:\n    :invalid"), 0644)

	_, err := LoadGlobal()
	if err == nil {
		t.Error("LoadGlobal() should return error for malformed YAML")
	}
}

// TestLoadGlobal_WithScripts verifies that LoadGlobal expands paths inside the
// scripts slice, exercising the `for i := range cfg.Scripts` loop body.
func TestLoadGlobal_WithScripts(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)
	t.Setenv("STGSCRIPTVAR", "/opt/scripts")
	cfgPath := globalConfigPath()
	os.MkdirAll(filepath.Dir(cfgPath), 0755)
	yaml := "version: \"1\"\nscripts:\n  - name: build\n    category: tools\n    path: \"$STGSCRIPTVAR/build.sh\"\n"
	os.WriteFile(cfgPath, []byte(yaml), 0644)

	cfg, err := LoadGlobal()
	if err != nil {
		t.Fatalf("LoadGlobal() error: %v", err)
	}
	if len(cfg.Scripts) == 0 {
		t.Fatal("expected at least one script")
	}
	if cfg.Scripts[0].Path != "/opt/scripts/build.sh" {
		t.Errorf("Scripts[0].Path = %q, want /opt/scripts/build.sh", cfg.Scripts[0].Path)
	}
}

// TestLoadGlobal_ReadError verifies that LoadGlobal returns a non-nil error when the
// config file path is unreadable for a reason other than non-existence (e.g. a directory
// placed where the file should be, causing EISDIR on ReadFile).
func TestLoadGlobal_ReadError(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)
	cfgDir := filepath.Join(tmpdir, ".config", "shotgum")
	os.MkdirAll(cfgDir, 0755)
	// Place a directory at the config file path so ReadFile fails with EISDIR.
	os.MkdirAll(filepath.Join(cfgDir, "config.yaml"), 0755)

	_, err := LoadGlobal()
	if err == nil {
		t.Error("LoadGlobal() should return error when config path is a directory")
	}
}

func TestLoadGlobal_ExpandsPath(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)
	t.Setenv("STGTESTVAR", "/custom/scripts")
	cfgPath := globalConfigPath()
	os.MkdirAll(filepath.Dir(cfgPath), 0755)
	yamlContent := "version: \"1\"\nscripts_home: \"~/my-scripts\"\ncategories:\n  - name: cat1\n    scripts_path: \"$STGTESTVAR/cat1\"\n"
	os.WriteFile(cfgPath, []byte(yamlContent), 0644)

	cfg, err := LoadGlobal()
	if err != nil {
		t.Fatalf("LoadGlobal() error: %v", err)
	}
	wantScriptsHome := filepath.Join(tmpdir, "my-scripts")
	if cfg.ScriptsHome != wantScriptsHome {
		t.Errorf("ScriptsHome = %q, want %q", cfg.ScriptsHome, wantScriptsHome)
	}
	if len(cfg.Categories) > 0 && cfg.Categories[0].ScriptsPath != "/custom/scripts/cat1" {
		t.Errorf("ScriptsPath = %q, want /custom/scripts/cat1", cfg.Categories[0].ScriptsPath)
	}
}

func TestLoadLocal_FileNotExist(t *testing.T) {
	tmpdir := t.TempDir()
	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })
	os.Chdir(tmpdir)

	cfg, err := LoadLocal()
	if err != nil {
		t.Fatalf("LoadLocal() unexpected error: %v", err)
	}
	if cfg != nil {
		t.Error("LoadLocal() should return nil when no .shotgum.yaml found")
	}
}

func TestLoadLocal_FileInCurrentDir(t *testing.T) {
	tmpdir := t.TempDir()
	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })
	os.Chdir(tmpdir)

	yamlContent := "version: \"1\"\nhelp_flag: \"--local-help\"\n"
	os.WriteFile(filepath.Join(tmpdir, ".shotgum.yaml"), []byte(yamlContent), 0644)

	cfg, err := LoadLocal()
	if err != nil {
		t.Fatalf("LoadLocal() error: %v", err)
	}
	if cfg == nil {
		t.Fatal("LoadLocal() returned nil for existing file")
	}
	if cfg.HelpFlag != "--local-help" {
		t.Errorf("HelpFlag = %q, want --local-help", cfg.HelpFlag)
	}
	if cfg.Source != "local" {
		t.Errorf("Source = %q, want local", cfg.Source)
	}
}

func TestLoadLocal_WalkUpToParent(t *testing.T) {
	tmpdir := t.TempDir()
	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })

	os.WriteFile(filepath.Join(tmpdir, ".shotgum.yaml"), []byte("version: \"1\"\nhelp_flag: \"--parent-help\"\n"), 0644)
	subdir := filepath.Join(tmpdir, "subdir")
	os.MkdirAll(subdir, 0755)
	os.Chdir(subdir)

	cfg, err := LoadLocal()
	if err != nil {
		t.Fatalf("LoadLocal() error: %v", err)
	}
	if cfg == nil {
		t.Fatal("LoadLocal() should find config in parent dir")
	}
	if cfg.HelpFlag != "--parent-help" {
		t.Errorf("HelpFlag = %q, want --parent-help", cfg.HelpFlag)
	}
}

// TestLoadLocal_WithCategoriesAndScripts verifies that LoadLocal expands paths inside
// both the categories and scripts slices, exercising their respective loop bodies.
func TestLoadLocal_WithCategoriesAndScripts(t *testing.T) {
	tmpdir := t.TempDir()
	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })
	os.Chdir(tmpdir)
	t.Setenv("STGLOCALVAR", "/local/path")

	yaml := "version: \"1\"\ncategories:\n  - name: mycat\n    scripts_path: \"$STGLOCALVAR/scripts\"\nscripts:\n  - name: run\n    category: mycat\n    path: \"$STGLOCALVAR/run.sh\"\n"
	os.WriteFile(filepath.Join(tmpdir, ".shotgum.yaml"), []byte(yaml), 0644)

	cfg, err := LoadLocal()
	if err != nil {
		t.Fatalf("LoadLocal() error: %v", err)
	}
	if len(cfg.Categories) == 0 || cfg.Categories[0].ScriptsPath != "/local/path/scripts" {
		t.Errorf("Categories[0].ScriptsPath = %q, want /local/path/scripts", cfg.Categories[0].ScriptsPath)
	}
	if len(cfg.Scripts) == 0 || cfg.Scripts[0].Path != "/local/path/run.sh" {
		t.Errorf("Scripts[0].Path = %q, want /local/path/run.sh", cfg.Scripts[0].Path)
	}
}

func TestLoadLocal_MalformedYAML(t *testing.T) {
	tmpdir := t.TempDir()
	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })
	os.Chdir(tmpdir)

	os.WriteFile(filepath.Join(tmpdir, ".shotgum.yaml"), []byte(":\n  bad:\n    :invalid"), 0644)

	_, err := LoadLocal()
	if err == nil {
		t.Error("LoadLocal() should return error for malformed YAML")
	}
}

func TestEnsureDefault_CreatesOnFirstRun(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)

	if err := EnsureDefault(); err != nil {
		t.Fatalf("EnsureDefault() error: %v", err)
	}
	cfgPath := globalConfigPath()
	if _, err := os.Stat(cfgPath); err != nil {
		t.Errorf("config file not created: %v", err)
	}
	scriptsHome := defaultScriptsHome()
	if _, err := os.Stat(scriptsHome); err != nil {
		t.Errorf("scripts dir not created: %v", err)
	}
}

func TestEnsureDefault_Idempotent(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)

	if err := EnsureDefault(); err != nil {
		t.Fatalf("first EnsureDefault() error: %v", err)
	}
	cfgPath := globalConfigPath()
	os.WriteFile(cfgPath, []byte("version: \"custom\""), 0644)

	if err := EnsureDefault(); err != nil {
		t.Fatalf("second EnsureDefault() error: %v", err)
	}
	data, _ := os.ReadFile(cfgPath)
	if string(data) != "version: \"custom\"" {
		t.Errorf("EnsureDefault() should be idempotent, but config was overwritten; got %q", string(data))
	}
}

func TestGlobalConfigPath(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)

	got := GlobalConfigPath()
	want := filepath.Join(tmpdir, ".config", "shotgum", "config.yaml")
	if got != want {
		t.Errorf("GlobalConfigPath() = %q, want %q", got, want)
	}
}

// TestSave_WriteFileFails verifies that Save() returns an error when the config
// directory is read-only, causing os.WriteFile to fail.
func TestSave_WriteFileFails(t *testing.T) {
	tmpdir := t.TempDir()
	cfgDir := filepath.Join(tmpdir, "ro")
	os.MkdirAll(cfgDir, 0755)
	os.Chmod(cfgDir, 0444) // read-only: WriteFile will fail
	t.Cleanup(func() { os.Chmod(cfgDir, 0755) })

	err := Save(&Config{Version: "1"}, filepath.Join(cfgDir, "config.yaml"))
	if err == nil {
		t.Error("Save() should return error when directory is read-only")
	}
}

// TestEnsureDefault_SaveFails verifies that EnsureDefault() propagates the error from
// Save() when HOME is 0555 (read+execute, no write). Stat of the (non-existent)
// config file returns ENOENT (traversal allowed), so the creation branch is entered;
// MkdirAll then fails with EACCES because HOME is not writable.
func TestEnsureDefault_SaveFails(t *testing.T) {
	tmpdir := t.TempDir()
	// 0555: allows stat/traverse but not write — MkdirAll inside Save will fail.
	os.Chmod(tmpdir, 0555)
	t.Cleanup(func() { os.Chmod(tmpdir, 0755) })
	t.Setenv("HOME", tmpdir)

	err := EnsureDefault()
	if err == nil {
		t.Error("EnsureDefault() should fail when HOME directory is not writable")
	}
}

// TestEnsureDefault_ScriptsDirFails verifies that EnsureDefault() propagates the error
// from os.MkdirAll when the scripts home path is blocked by a regular file.
func TestEnsureDefault_ScriptsDirFails(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)

	// Create the config file first so the "create config" branch is skipped.
	cfgPath := GlobalConfigPath()
	os.MkdirAll(filepath.Dir(cfgPath), 0755)
	os.WriteFile(cfgPath, []byte("version: \"1\"\n"), 0644)

	// Block ~/.shotgum/scripts creation by placing a regular file at ~/.shotgum.
	os.WriteFile(filepath.Join(tmpdir, ".shotgum"), []byte("block"), 0644)

	err := EnsureDefault()
	if err == nil {
		t.Error("EnsureDefault() should fail when scripts home is blocked by a file")
	}
}

func TestExpandPath(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)
	t.Setenv("STGTESTEXPANDVAR", "/custom/path")

	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"~/foo", filepath.Join(tmpdir, "foo")},
		{"$STGTESTEXPANDVAR/scripts", "/custom/path/scripts"},
		{"/absolute/path", "/absolute/path"},
	}
	for _, tt := range tests {
		got := expandPath(tt.input)
		if got != tt.want {
			t.Errorf("expandPath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

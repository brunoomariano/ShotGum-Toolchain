package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shotgum/stg/internal/config"
)

// TestAddCmd_SubcommandRegistration verifies addCmd creates parent with subcommands.
func TestAddCmd_SubcommandRegistration(t *testing.T) {
	cmd := addCmd()
	if cmd.Use != "add" {
		t.Errorf("addCmd().Use = %q, want 'add'", cmd.Use)
	}
	// Should have two subcommands: folder and script
	found := map[string]bool{}
	for _, sub := range cmd.Commands() {
		found[sub.Name()] = true
	}
	if !found["folder"] {
		t.Error("addCmd should have 'folder' subcommand")
	}
	if !found["script"] {
		t.Error("addCmd should have 'script' subcommand")
	}
}

// TestListCmd_CommandStructure verifies listCmd creates the right cobra command.
func TestListCmd_CommandStructure(t *testing.T) {
	cmd := listCmd()
	if cmd.Use != "list [category]" {
		t.Errorf("listCmd().Use = %q, want 'list [category]'", cmd.Use)
	}
}

// TestInitCmd_CommandStructure verifies initCmd creates the right cobra command.
func TestInitCmd_CommandStructure(t *testing.T) {
	cmd := initCmd()
	if cmd.Use != "init" {
		t.Errorf("initCmd().Use = %q, want 'init'", cmd.Use)
	}
}

// TestRoot_CommandStructure verifies Root creates the root command with subcommands.
func TestRoot_CommandStructure(t *testing.T) {
	root := Root()
	if root.Use != "stg [category] [script] [args...]" {
		t.Errorf("Root().Use = %q, unexpected", root.Use)
	}
	found := map[string]bool{}
	for _, sub := range root.Commands() {
		found[sub.Name()] = true
	}
	if !found["list"] {
		t.Error("root should have 'list' subcommand")
	}
	if !found["add"] {
		t.Error("root should have 'add' subcommand")
	}
	if !found["init"] {
		t.Error("root should have 'init' subcommand")
	}
}

// TestListCmd_Execute_NoCategories exercises the full listCmd path.
func TestListCmd_Execute_NoCategories(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)
	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })
	os.Chdir(tmpdir)

	config.Save(&config.Config{Version: "1"}, config.GlobalConfigPath())

	out := captureStdout(func() {
		cmd := listCmd()
		cmd.SetArgs([]string{})
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true
		cmd.Execute()
	})
	if !strings.Contains(out, "No categories") {
		t.Errorf("output = %q, want 'No categories'", out)
	}
}

// TestListCmd_Execute_WithCategory exercises listCmd with a category argument.
func TestListCmd_Execute_WithCategory(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)
	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })
	os.Chdir(tmpdir)

	config.Save(&config.Config{
		Version:    "1",
		Categories: []config.Category{{Name: "tools"}},
		Scripts:    []config.Script{{Name: "build", Category: "tools", Type: "script"}},
	}, config.GlobalConfigPath())

	out := captureStdout(func() {
		cmd := listCmd()
		cmd.SetArgs([]string{"tools"})
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true
		cmd.Execute()
	})
	stripped := stripCmdsANSI(out)
	if !strings.Contains(stripped, "build") {
		t.Errorf("output = %q, want 'build'", stripped)
	}
}

// TestRoot_DirectRun_ScriptNotFound exercises the root command's default branch
// (len(args) >= 2) when the requested script does not exist in the registry,
// verifying that FindScript's error propagates back through cobra.
func TestRoot_DirectRun_ScriptNotFound(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)
	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })
	os.Chdir(tmpdir)

	config.Save(&config.Config{
		Version:    "1",
		Categories: []config.Category{{Name: "tools"}},
	}, config.GlobalConfigPath())

	cmd := Root()
	cmd.SetArgs([]string{"tools", "nonexistent"})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	err := cmd.Execute()
	if err == nil {
		t.Error("Root with nonexistent script should return error")
	}
	if !strings.Contains(err.Error(), "nonexistent") {
		t.Errorf("error = %q, want it to mention the script name", err.Error())
	}
}

// TestRoot_DirectRun_Success exercises the root command's default branch (len(args)>=2)
// when the script exists, verifying end-to-end execution via runDirect.
func TestRoot_DirectRun_Success(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)
	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })
	os.Chdir(tmpdir)

	scriptPath := filepath.Join(tmpdir, "echo.sh")
	os.WriteFile(scriptPath, []byte("#!/bin/bash\necho direct-run-ok"), 0755)

	config.Save(&config.Config{
		Version:    "1",
		Categories: []config.Category{{Name: "tools"}},
		Scripts:    []config.Script{{Name: "echo", Category: "tools", Type: "script", Path: scriptPath}},
	}, config.GlobalConfigPath())

	out := captureStdout(func() {
		cmd := Root()
		cmd.SetArgs([]string{"tools", "echo"})
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true
		cmd.Execute()
	})
	if !strings.Contains(out, "direct-run-ok") {
		t.Errorf("output = %q, want 'direct-run-ok'", out)
	}
}

// TestAddScriptCmd_Duplicate tests that adding a duplicate script returns error.
func TestAddScriptCmd_Duplicate(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)
	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })
	os.Chdir(tmpdir)

	// Create script file
	scriptPath := tmpdir + "/build.sh"
	os.WriteFile(scriptPath, []byte("#!/bin/bash\necho build"), 0755)

	// Write config with existing script
	config.Save(&config.Config{
		Version:    "1",
		Categories: []config.Category{{Name: "tools"}},
		Scripts:    []config.Script{{Name: "build", Category: "tools", Type: "script", Path: scriptPath}},
	}, config.GlobalConfigPath())

	err := runAddScript([]string{scriptPath, "--category", "tools", "--name", "build"})
	if err == nil {
		t.Error("adding duplicate script should return error")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error = %q, want 'already exists'", err.Error())
	}
}

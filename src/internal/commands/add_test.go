package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/brunoomariano/ShotGum-Toolchain/internal/config"
	"github.com/brunoomariano/ShotGum-Toolchain/internal/registry"
)

// runAddFolder runs addFolderCmd with the given args and returns the error.
func runAddFolder(args []string) error {
	cmd := addFolderCmd()
	cmd.SetArgs(args)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	return cmd.Execute()
}

// runAddScript runs addScriptCmd with the given args and returns the error.
func runAddScript(args []string) error {
	cmd := addScriptCmd()
	cmd.SetArgs(args)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	return cmd.Execute()
}

func TestAddFolder_InvalidPath(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)
	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })
	os.Chdir(tmpdir)

	err := runAddFolder([]string{"/nonexistent/path/xyz"})
	if err == nil {
		t.Error("add folder with nonexistent path should return error")
	}
}

func TestAddFolder_PathIsFile(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)
	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })
	os.Chdir(tmpdir)

	filePath := filepath.Join(tmpdir, "notadir")
	os.WriteFile(filePath, []byte("x"), 0644)

	err := runAddFolder([]string{filePath})
	if err == nil {
		t.Error("add folder with file path should return error")
	}
	if !strings.Contains(err.Error(), "not a directory") {
		t.Errorf("error = %q, want 'not a directory'", err.Error())
	}
}

func TestAddFolder_Valid(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)
	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })
	os.Chdir(tmpdir)

	folderPath := filepath.Join(tmpdir, "myscripts")
	os.MkdirAll(folderPath, 0755)

	out := captureStdout(func() {
		runAddFolder([]string{folderPath, "--name", "mycat", "--desc", "My category"})
	})
	_ = out

	cfg, err := config.LoadGlobal()
	if err != nil {
		t.Fatalf("LoadGlobal() error: %v", err)
	}
	found := false
	for _, c := range cfg.Categories {
		if c.Name == "mycat" {
			found = true
			break
		}
	}
	if !found {
		t.Error("category 'mycat' should have been added to global config")
	}
}

func TestAddFolder_Duplicate(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)
	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })
	os.Chdir(tmpdir)

	initialCfg := &config.Config{
		Version: "1",
		Categories: []config.Category{
			{Name: "mycat", ScriptsPath: tmpdir},
		},
	}
	config.Save(initialCfg, config.GlobalConfigPath())

	folderPath := filepath.Join(tmpdir, "other")
	os.MkdirAll(folderPath, 0755)

	err := runAddFolder([]string{folderPath, "--name", "mycat"})
	if err == nil {
		t.Error("add folder with duplicate name should return error")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error = %q, want 'already exists'", err.Error())
	}
}

func TestAddScript_NoCategory_Flag(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)
	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })
	os.Chdir(tmpdir)

	scriptPath := filepath.Join(tmpdir, "myscript.sh")
	os.WriteFile(scriptPath, []byte("#!/bin/bash\necho hi"), 0755)

	// No --category flag
	err := runAddScript([]string{scriptPath})
	if err == nil {
		t.Error("add script without --category should return error")
	}
	if !strings.Contains(err.Error(), "--category") {
		t.Errorf("error = %q, want '--category'", err.Error())
	}
}

func TestAddScript_CategoryNotFound(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)
	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })
	os.Chdir(tmpdir)

	// Write empty global config (no categories)
	config.Save(&config.Config{Version: "1"}, config.GlobalConfigPath())

	scriptPath := filepath.Join(tmpdir, "myscript.sh")
	os.WriteFile(scriptPath, []byte("#!/bin/bash\necho hi"), 0755)

	err := runAddScript([]string{scriptPath, "--category", "nonexistent"})
	if err == nil {
		t.Error("add script with nonexistent category should return error")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want 'not found'", err.Error())
	}
}

func TestAddScript_PathNotExist(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)
	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })
	os.Chdir(tmpdir)

	config.Save(&config.Config{
		Version:    "1",
		Categories: []config.Category{{Name: "tools"}},
	}, config.GlobalConfigPath())

	err := runAddScript([]string{"/nonexistent/script.sh", "--category", "tools"})
	if err == nil {
		t.Error("add script with nonexistent path should return error")
	}
}

func TestAddScript_Valid(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)
	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })
	os.Chdir(tmpdir)

	config.Save(&config.Config{
		Version:    "1",
		Categories: []config.Category{{Name: "tools", ScriptsPath: tmpdir}},
	}, config.GlobalConfigPath())

	scriptPath := filepath.Join(tmpdir, "build.sh")
	os.WriteFile(scriptPath, []byte("#!/bin/bash\necho build"), 0755)

	out := captureStdout(func() {
		runAddScript([]string{scriptPath, "--category", "tools", "--name", "build", "--desc", "Build script"})
	})
	_ = out

	cfg, err := config.LoadGlobal()
	if err != nil {
		t.Fatalf("LoadGlobal() error: %v", err)
	}
	found := false
	for _, s := range cfg.Scripts {
		if s.Name == "build" && s.Category == "tools" {
			found = true
			break
		}
	}
	if !found {
		t.Error("script 'build' should have been added to global config")
	}
}

// TestAddFolder_FolderWithShFiles covers the "found N scripts" hint output branch.
func TestAddFolder_FolderWithShFiles(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)
	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })
	os.Chdir(tmpdir)

	folderPath := filepath.Join(tmpdir, "scripts")
	os.MkdirAll(folderPath, 0755)
	// Create .sh files in the folder so the hint branch is exercised
	os.WriteFile(filepath.Join(folderPath, "deploy.sh"), []byte("#!/bin/bash\necho deploy"), 0755)
	os.WriteFile(filepath.Join(folderPath, "test.sh"), []byte("#!/bin/bash\necho test"), 0755)

	out := captureStdout(func() {
		runAddFolder([]string{folderPath, "--name", "scripts"})
	})
	stripped := stripCmdsANSI(out)
	// The hint message contains the count of found scripts
	if !strings.Contains(stripped, "2") {
		t.Errorf("output = %q, expected hint about 2 script(s)", stripped)
	}
}

// TestRunDirect exercises the runDirect command path (calls runner.Run).
func TestRunDirect(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)
	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })
	os.Chdir(tmpdir)

	// Write config with the script
	scriptPath := filepath.Join(tmpdir, "echo.sh")
	os.WriteFile(scriptPath, []byte("#!/bin/bash\necho direct-run"), 0755)

	config.Save(&config.Config{
		Version:    "1",
		Categories: []config.Category{{Name: "tools"}},
		Scripts:    []config.Script{{Name: "echo", Category: "tools", Executable: "/bin/sh", Path: scriptPath}},
	}, config.GlobalConfigPath())

	reg, err := registry.Load()
	if err != nil {
		t.Fatalf("registry.Load() error: %v", err)
	}
	entry, err := reg.FindScript("tools", "echo")
	if err != nil {
		t.Fatalf("FindScript() error: %v", err)
	}

	captureStdout(func() {
		runDirect(*entry, nil, reg)
	})
	// Just ensure no panic; runDirect streams to os.Stdout
}

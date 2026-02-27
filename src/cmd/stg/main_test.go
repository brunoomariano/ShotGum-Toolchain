package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/brunoomariano/ShotGum-Toolchain/internal/config"
)

func TestIsDirectRun_True(t *testing.T) {
	if !isDirectRun([]string{"dev", "build"}) {
		t.Error("expected direct run for category + script")
	}
}

func TestIsDirectRun_False_ForShortArgs(t *testing.T) {
	if isDirectRun([]string{"dev"}) {
		t.Error("expected false for <2 args")
	}
}

func TestIsDirectRun_False_ForFlags(t *testing.T) {
	if isDirectRun([]string{"-h", "build"}) {
		t.Error("expected false when first arg is flag")
	}
	if isDirectRun([]string{"dev", "--help"}) {
		t.Error("expected false when second arg is flag")
	}
}

func TestIsDirectRun_False_ForSubcommand(t *testing.T) {
	if isDirectRun([]string{"list", "tools"}) {
		t.Error("expected false for known subcommand")
	}
}

func TestRun_Help_ReturnsZero(t *testing.T) {
	if code := run([]string{"--help"}); code != 0 {
		t.Errorf("run(--help) = %d, want 0", code)
	}
}

func TestRun_DirectRun_ScriptNotFound(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)
	cfgPath := filepath.Join(tmpdir, ".config", "shotgum", "config.yaml")
	os.MkdirAll(filepath.Dir(cfgPath), 0755)
	config.Save(&config.Config{Version: "1"}, cfgPath)

	if code := run([]string{"dev", "missing"}); code == 0 {
		t.Errorf("run() should return non-zero when script is missing")
	}
}

func TestRun_DirectRun_Success(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)

	scriptsDir := filepath.Join(tmpdir, "scripts")
	os.MkdirAll(scriptsDir, 0755)
	scriptPath := filepath.Join(scriptsDir, "run.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho ok\n"), 0755); err != nil {
		t.Fatalf("write script: %v", err)
	}

	cfg := &config.Config{
		Version: "1",
		Categories: []config.Category{
			{Name: "dev", ScriptsPath: scriptsDir},
		},
		Scripts: []config.Script{
			{Name: "hello", Category: "dev", Path: "run.sh", Executable: "/bin/sh"},
		},
	}
	cfgPath := filepath.Join(tmpdir, ".config", "shotgum", "config.yaml")
	os.MkdirAll(filepath.Dir(cfgPath), 0755)
	if err := config.Save(cfg, cfgPath); err != nil {
		t.Fatalf("save config: %v", err)
	}

	if code := run([]string{"dev", "hello"}); code != 0 {
		t.Errorf("run() = %d, want 0", code)
	}
}

func TestRun_DirectRun_ExitCode(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)

	scriptsDir := filepath.Join(tmpdir, "scripts")
	os.MkdirAll(scriptsDir, 0755)
	scriptPath := filepath.Join(scriptsDir, "fail.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\nexit 7\n"), 0755); err != nil {
		t.Fatalf("write script: %v", err)
	}

	cfg := &config.Config{
		Version: "1",
		Categories: []config.Category{
			{Name: "dev", ScriptsPath: scriptsDir},
		},
		Scripts: []config.Script{
			{Name: "fail", Category: "dev", Path: "fail.sh", Executable: "/bin/sh"},
		},
	}
	cfgPath := filepath.Join(tmpdir, ".config", "shotgum", "config.yaml")
	os.MkdirAll(filepath.Dir(cfgPath), 0755)
	if err := config.Save(cfg, cfgPath); err != nil {
		t.Fatalf("save config: %v", err)
	}

	if code := run([]string{"dev", "fail"}); code != 7 {
		t.Errorf("run() = %d, want 7", code)
	}
}

func TestRun_DirectRun_LoadError(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)
	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })
	os.Chdir(tmpdir)

	if err := os.MkdirAll(filepath.Join(tmpdir, ".shotgum.yaml"), 0755); err != nil {
		t.Fatalf("mkdir .shotgum.yaml: %v", err)
	}

	if code := run([]string{"dev", "build"}); code == 0 {
		t.Errorf("run() should return non-zero on registry load error")
	}
}

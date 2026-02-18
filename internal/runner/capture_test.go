package runner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shotgum/stg/internal/config"
	"github.com/shotgum/stg/internal/registry"
)

// makeReg creates a minimal registry via Load() with an empty home dir.
func makeReg(t *testing.T) *registry.Registry {
	t.Helper()
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)
	// Write minimal global config so LoadGlobal doesn't fail
	cfgPath := filepath.Join(tmpdir, ".config", "shotgum", "config.yaml")
	os.MkdirAll(filepath.Dir(cfgPath), 0755)
	os.WriteFile(cfgPath, []byte("version: \"1\"\n"), 0644)

	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })
	os.Chdir(tmpdir)

	reg, err := registry.Load()
	if err != nil {
		t.Fatalf("registry.Load(): %v", err)
	}
	return reg
}

func makeAbsEntry(path, typ string) registry.ScriptEntry {
	return registry.ScriptEntry{
		Script: config.Script{
			Name:     "test",
			Category: "dev",
			Type:     typ,
			Path:     path,
		},
		Source: "global",
	}
}

func TestCaptureRun_Script(t *testing.T) {
	reg := makeReg(t)
	tmpdir := t.TempDir()
	script := writeScript(t, tmpdir, "hello.sh", "#!/bin/bash\necho capturerun-ok")

	entry := makeAbsEntry(script, "script")
	out, err := CaptureRun(entry, nil, reg)
	if err != nil {
		t.Fatalf("CaptureRun() error: %v", err)
	}
	if !strings.Contains(out, "capturerun-ok") {
		t.Errorf("output = %q, want 'capturerun-ok'", out)
	}
}

func TestCaptureRun_Executable(t *testing.T) {
	reg := makeReg(t)
	entry := makeAbsEntry("/bin/echo", "executable")
	out, err := CaptureRun(entry, []string{"capturerun-exec"}, reg)
	if err != nil {
		t.Fatalf("CaptureRun() error: %v", err)
	}
	if !strings.Contains(out, "capturerun-exec") {
		t.Errorf("output = %q, want 'capturerun-exec'", out)
	}
}

func TestCaptureRun_ExitError(t *testing.T) {
	reg := makeReg(t)
	tmpdir := t.TempDir()
	script := writeScript(t, tmpdir, "fail.sh", "#!/bin/bash\necho partial\nexit 3")

	entry := makeAbsEntry(script, "script")
	out, err := CaptureRun(entry, nil, reg)
	if err == nil {
		t.Error("CaptureRun() should return error for non-zero exit")
	}
	_ = out
}

func TestCaptureRunForPreview_SetsColumns(t *testing.T) {
	reg := makeReg(t)
	tmpdir := t.TempDir()
	script := writeScript(t, tmpdir, "cols.sh", "#!/bin/bash\necho COLUMNS=$COLUMNS")

	entry := makeAbsEntry(script, "script")
	out, err := CaptureRunForPreview(entry, nil, reg, 77)
	if err != nil {
		t.Fatalf("CaptureRunForPreview() error: %v", err)
	}
	if !strings.Contains(out, "COLUMNS=77") {
		t.Errorf("output = %q, want 'COLUMNS=77'", out)
	}
}

func TestCaptureRunForPreview_SetsTERM(t *testing.T) {
	reg := makeReg(t)
	tmpdir := t.TempDir()
	script := writeScript(t, tmpdir, "term.sh", "#!/bin/bash\necho TERM=$TERM")

	entry := makeAbsEntry(script, "script")
	out, err := CaptureRunForPreview(entry, nil, reg, 80)
	if err != nil {
		t.Fatalf("CaptureRunForPreview() error: %v", err)
	}
	if !strings.Contains(out, "TERM=xterm-256color") {
		t.Errorf("output = %q, want TERM=xterm-256color", out)
	}
}

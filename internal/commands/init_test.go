package commands

import (
	"os"
	"strings"
	"testing"

	"github.com/shotgum/stg/internal/config"
)

func runInit() (string, error) {
	var out string
	var execErr error
	out = captureStdout(func() {
		cmd := initCmd()
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true
		execErr = cmd.Execute()
	})
	return out, execErr
}

func TestInit_CleanEnvironment(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)
	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })
	os.Chdir(tmpdir)

	out, err := runInit()
	if err != nil {
		t.Fatalf("stg init error: %v", err)
	}
	stripped := stripCmdsANSI(out)
	if !strings.Contains(stripped, "initialized") {
		t.Errorf("output = %q, want 'initialized'", stripped)
	}

	// Config file should have been created
	if _, statErr := os.Stat(config.GlobalConfigPath()); statErr != nil {
		t.Errorf("config file not created: %v", statErr)
	}
}

func TestInit_AlreadyInitialized(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)
	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })
	os.Chdir(tmpdir)

	// Run init first
	if _, err := runInit(); err != nil {
		t.Fatalf("first stg init error: %v", err)
	}

	// Run init again
	out, err := runInit()
	if err != nil {
		t.Fatalf("second stg init error: %v", err)
	}
	stripped := stripCmdsANSI(out)
	if !strings.Contains(stripped, "already initialized") {
		t.Errorf("output = %q, want 'already initialized'", stripped)
	}
}

package runner

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/brunoomariano/ShotGum-Toolchain/internal/config"
	"github.com/brunoomariano/ShotGum-Toolchain/internal/registry"
)

// captureTerminalOutput redirects os.Stdout and os.Stderr during f(), returning
// the combined output. This lets us test functions that write directly to os.Stdout.
func captureTerminalOutput(f func()) string {
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w
	f()
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	os.Stdout = oldStdout
	os.Stderr = oldStderr
	return buf.String()
}

func makeRunEntry(path, executable string) registry.ScriptEntry {
	return registry.ScriptEntry{
		Script: config.Script{
			Name:       "test",
			Category:   "dev",
			Executable: executable,
			Path:       path,
		},
		Source: "global",
	}
}

// ── run() ──────────────────────────────────────────────────────────────────

func TestRun_Script_Success(t *testing.T) {
	tmpdir := t.TempDir()
	script := writeScript(t, tmpdir, "run.sh", "#!/bin/bash\necho run-success")

	captureTerminalOutput(func() {
		err := run("/bin/sh", script, nil)
		if err != nil {
			t.Errorf("run() error: %v", err)
		}
	})
}

func TestRun_Executable_Success(t *testing.T) {
	tmpdir := t.TempDir()
	script := writeScript(t, tmpdir, "echo.sh", "#!/bin/bash\necho run-exec-ok")

	captureTerminalOutput(func() {
		err := run("/bin/sh", script, nil)
		if err != nil {
			t.Errorf("run() error: %v", err)
		}
	})
}

func TestRun_Script_ExitError(t *testing.T) {
	tmpdir := t.TempDir()
	script := writeScript(t, tmpdir, "fail.sh", "#!/bin/bash\nexit 5")

	var runErr *RunError
	captureTerminalOutput(func() {
		err := run("/bin/sh", script, nil)
		if err == nil {
			t.Error("run() should return error for non-zero exit")
			return
		}
		if !errors.As(err, &runErr) {
			t.Errorf("error should be *RunError, got %T: %v", err, err)
		}
	})
	if runErr != nil && runErr.ExitCode != 5 {
		t.Errorf("ExitCode = %d, want 5", runErr.ExitCode)
	}
}

func TestRun_Script_NotFound(t *testing.T) {
	captureTerminalOutput(func() {
		err := run("/bin/sh", "/nonexistent/script.sh", nil)
		if err == nil {
			t.Error("run() should return error for nonexistent script")
		}
	})
}

// ── Run() ──────────────────────────────────────────────────────────────────

func TestRun_ViaPublicFunction(t *testing.T) {
	reg := makeReg(t)
	tmpdir := t.TempDir()
	script := writeScript(t, tmpdir, "run2.sh", "#!/bin/bash\necho run-via-public")

	entry := makeRunEntry(script, "/bin/sh")
	captureTerminalOutput(func() {
		err := Run(entry, nil, reg)
		if err != nil {
			t.Errorf("Run() error: %v", err)
		}
	})
}

// ── RunHelp() ──────────────────────────────────────────────────────────────

func TestRunHelp(t *testing.T) {
	reg := makeReg(t)
	tmpdir := t.TempDir()
	// Script that prints a message when given --help
	script := writeScript(t, tmpdir, "help.sh", "#!/bin/bash\nif [ \"$1\" = \"--help\" ]; then echo 'help text'; fi")

	entry := makeRunEntry(script, "/bin/sh")
	captureTerminalOutput(func() {
		err := RunHelp(entry, reg)
		if err != nil {
			t.Errorf("RunHelp() error: %v", err)
		}
	})
}

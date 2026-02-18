package runner

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunError_Error(t *testing.T) {
	err := &RunError{ExitCode: 42, Err: errors.New("exit status 42")}
	got := err.Error()
	if !strings.Contains(got, "42") {
		t.Errorf("RunError.Error() = %q, should contain exit code 42", got)
	}
}

func writeScript(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		t.Fatalf("writing script %s: %v", name, err)
	}
	return path
}

func TestCapture_ScriptStdout(t *testing.T) {
	tmpdir := t.TempDir()
	script := writeScript(t, tmpdir, "hello.sh", "#!/bin/bash\necho hello")

	out, err := capture("script", script, nil, nil)
	if err != nil {
		t.Fatalf("capture() error: %v", err)
	}
	if !strings.Contains(out, "hello") {
		t.Errorf("capture() output = %q, want it to contain 'hello'", out)
	}
}

func TestCapture_ExecutableBinary(t *testing.T) {
	out, err := capture("executable", "/bin/echo", []string{"world"}, nil)
	if err != nil {
		t.Fatalf("capture() error: %v", err)
	}
	if !strings.Contains(out, "world") {
		t.Errorf("capture() output = %q, want it to contain 'world'", out)
	}
}

func TestCapture_ExitCodeError(t *testing.T) {
	tmpdir := t.TempDir()
	script := writeScript(t, tmpdir, "fail.sh", "#!/bin/bash\nexit 2")

	_, err := capture("script", script, nil, nil)
	if err == nil {
		t.Fatal("capture() should return error for non-zero exit")
	}
	var runErr *RunError
	if !errors.As(err, &runErr) {
		t.Errorf("error should be *RunError, got %T", err)
	}
	if runErr.ExitCode != 2 {
		t.Errorf("ExitCode = %d, want 2", runErr.ExitCode)
	}
}

func TestCapture_NonexistentScript(t *testing.T) {
	_, err := capture("script", "/nonexistent/path/script.sh", nil, nil)
	if err == nil {
		t.Error("capture() should return error for nonexistent script")
	}
}

func TestCapture_InjectsExtraEnv(t *testing.T) {
	tmpdir := t.TempDir()
	script := writeScript(t, tmpdir, "env.sh", "#!/bin/bash\necho COLUMNS=$COLUMNS")
	extraEnv := []string{"TERM=xterm-256color", "COLUMNS=42"}

	out, err := capture("script", script, nil, extraEnv)
	if err != nil {
		t.Fatalf("capture() error: %v", err)
	}
	if !strings.Contains(out, "COLUMNS=42") {
		t.Errorf("capture() output = %q, want it to contain 'COLUMNS=42'", out)
	}
}

func TestCapture_Width80(t *testing.T) {
	tmpdir := t.TempDir()
	script := writeScript(t, tmpdir, "cols.sh", "#!/bin/bash\necho COLUMNS=$COLUMNS")
	extraEnv := []string{"TERM=xterm-256color", "COLUMNS=80"}

	out, err := capture("script", script, nil, extraEnv)
	if err != nil {
		t.Fatalf("capture() error: %v", err)
	}
	if !strings.Contains(out, "COLUMNS=80") {
		t.Errorf("output = %q, want COLUMNS=80", out)
	}
}

func TestCapture_ScriptStderr(t *testing.T) {
	tmpdir := t.TempDir()
	script := writeScript(t, tmpdir, "stderr.sh", "#!/bin/bash\necho err >&2\nexit 1")

	out, err := capture("script", script, nil, nil)
	if err == nil {
		t.Error("expected error from non-zero exit")
	}
	// CombinedOutput captures both stdout and stderr
	if !strings.Contains(out, "err") {
		t.Errorf("output = %q, want stderr 'err' in combined output", out)
	}
}

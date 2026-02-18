package runner

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestShellQuote_Empty verifies shellQuote handles empty strings safely.
func TestShellQuote_Empty(t *testing.T) {
	if got := shellQuote(""); got != "''" {
		t.Fatalf("shellQuote(\"\") = %q, want %q", got, "''")
	}
}

// TestShellQuote_WithSingleQuote verifies shellQuote escapes single quotes.
func TestShellQuote_WithSingleQuote(t *testing.T) {
	got := shellQuote("a'b")
	if !strings.Contains(got, "'\"'\"'") {
		t.Fatalf("shellQuote should escape single quotes, got: %q", got)
	}
}

// TestShellJoin verifies shellJoin quotes and joins command parts.
func TestShellJoin(t *testing.T) {
	got := shellJoin([]string{"echo", "hello world"})
	if !strings.Contains(got, "'hello world'") {
		t.Fatalf("shellJoin should quote spaced args, got: %q", got)
	}
}

// TestStartInteractive_SendInputAndRead verifies interactive startup, input
// forwarding, output streaming, and done notification.
func TestStartInteractive_SendInputAndRead(t *testing.T) {
	reg := makeReg(t)
	tmpdir := t.TempDir()
	scriptPath := writeScript(t, tmpdir, "interactive.sh", "#!/usr/bin/env bash\nread -r name\necho \"hello:$name\"")

	entry := makeAbsEntry(filepath.Join(tmpdir, filepath.Base(scriptPath)), "/bin/bash")
	session, err := StartInteractive(entry, nil, reg)
	if err != nil {
		t.Fatalf("StartInteractive() error: %v", err)
	}

	if err := session.SendInput([]byte("shotgum\n")); err != nil {
		t.Fatalf("SendInput() error: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	var out strings.Builder
	for time.Now().Before(deadline) {
		buf := make([]byte, 512)
		n, readErr := session.ReadChunk(buf)
		if n > 0 {
			out.Write(buf[:n])
			if strings.Contains(out.String(), "hello:shotgum") {
				break
			}
		}
		if readErr != nil {
			break
		}
	}

	if !strings.Contains(out.String(), "hello:shotgum") {
		t.Fatalf("interactive output missing expected text, got: %q", out.String())
	}

	select {
	case doneErr := <-session.Done():
		if doneErr != nil {
			t.Fatalf("process finished with error: %v", doneErr)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for interactive process to finish")
	}
}

// TestStartInteractive_ExitError verifies non-zero exits are wrapped as RunError.
func TestStartInteractive_ExitError(t *testing.T) {
	reg := makeReg(t)
	tmpdir := t.TempDir()
	scriptPath := writeScript(t, tmpdir, "boom.sh", "#!/usr/bin/env bash\nexit 7")

	entry := makeAbsEntry(filepath.Join(tmpdir, filepath.Base(scriptPath)), "/bin/bash")
	session, err := StartInteractive(entry, nil, reg)
	if err != nil {
		t.Fatalf("StartInteractive() error: %v", err)
	}

	select {
	case doneErr := <-session.Done():
		if doneErr == nil {
			t.Fatal("expected non-nil error for exit code 7")
		}
		var runErr *RunError
		if !errors.As(doneErr, &runErr) {
			t.Fatalf("expected *RunError, got %T (%v)", doneErr, doneErr)
		}
		if runErr.ExitCode != 7 {
			t.Fatalf("ExitCode=%d want 7", runErr.ExitCode)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for interactive process error")
	}
}

// TestStartInteractive_WithArgs verifies args are forwarded to the process.
func TestStartInteractive_WithArgs(t *testing.T) {
	reg := makeReg(t)
	tmpdir := t.TempDir()
	scriptPath := writeScript(t, tmpdir, "args.sh", "#!/usr/bin/env bash\necho \"arg:$1\"")

	entry := makeAbsEntry(filepath.Join(tmpdir, filepath.Base(scriptPath)), "/bin/bash")
	session, err := StartInteractive(entry, []string{"--hello"}, reg)
	if err != nil {
		t.Fatalf("StartInteractive() error: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	var out strings.Builder
	for time.Now().Before(deadline) {
		buf := make([]byte, 512)
		n, readErr := session.ReadChunk(buf)
		if n > 0 {
			out.Write(buf[:n])
			if strings.Contains(out.String(), "arg:--hello") {
				break
			}
		}
		if readErr != nil {
			break
		}
	}

	if !strings.Contains(out.String(), "arg:--hello") {
		t.Fatalf("interactive output missing arg forwarding, got: %q", out.String())
	}

	select {
	case doneErr := <-session.Done():
		if doneErr != nil {
			t.Fatalf("process finished with error: %v", doneErr)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for interactive process finish")
	}
}

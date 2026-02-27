// Package runner executes ShotGum scripts in three modes:
//   - Run: streams stdout/stderr directly to the terminal.
//   - CaptureRun: captures combined output for display in the TUI output panel.
//   - StartInteractive: starts a PTY-wrapped process so tools like gum detect a
//     real terminal; falls back to direct execution when `script(1)` is absent.
package runner

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/brunoomariano/ShotGum-Toolchain/internal/registry"
)

// RunError wraps a script execution error with its exit code.
type RunError struct {
	ExitCode int
	Err      error
}

// InteractiveSession holds a running process with stdin/stdout connected so the
// TUI can stream output and forward keyboard input.
type InteractiveSession struct {
	stdin  io.WriteCloser
	stdout io.ReadCloser
	done   chan error
}

func (e *RunError) Error() string {
	return fmt.Sprintf("script exited with code %d: %v", e.ExitCode, e.Err)
}

// Run executes a script, streaming stdout/stderr to the terminal.
func Run(entry registry.ScriptEntry, args []string, reg *registry.Registry) error {
	executable, baseArgs := reg.ResolveInvocation(entry)
	return run(executable, baseArgs, args)
}

// RunHelp executes a script with its resolved help flag.
func RunHelp(entry registry.ScriptEntry, reg *registry.Registry) error {
	helpFlag := reg.ResolveHelpFlag(entry)
	if helpFlag == "" {
		return nil
	}
	executable, baseArgs := reg.ResolveInvocation(entry)
	return run(executable, baseArgs, []string{helpFlag})
}

func run(executable string, baseArgs []string, args []string) error {
	cmd := exec.Command(executable, append(baseArgs, args...)...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return &RunError{ExitCode: exitErr.ExitCode(), Err: err}
		}
		return fmt.Errorf("running script: %w", err)
	}
	return nil
}

// CaptureRun executes a script and returns its combined output as a string.
// Used by the TUI output view.
func CaptureRun(entry registry.ScriptEntry, args []string, reg *registry.Registry) (string, error) {
	executable, baseArgs := reg.ResolveInvocation(entry)
	return capture(executable, baseArgs, args, nil)
}

// CaptureRunForPreview executes a script with extra env vars so gum can render
// without a real TTY (sets TERM and COLUMNS for lipgloss/gum width detection).
func CaptureRunForPreview(entry registry.ScriptEntry, args []string, reg *registry.Registry, width int) (string, error) {
	executable, baseArgs := reg.ResolveInvocation(entry)
	extraEnv := []string{
		"TERM=xterm-256color",
		fmt.Sprintf("COLUMNS=%d", width),
	}
	return capture(executable, baseArgs, args, extraEnv)
}

func capture(executable string, baseArgs []string, args []string, extraEnv []string) (string, error) {
	cmd := exec.Command(executable, append(baseArgs, args...)...)

	if len(extraEnv) > 0 {
		cmd.Env = append(os.Environ(), extraEnv...)
	}

	out, err := cmd.CombinedOutput()
	output := string(out)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return output, &RunError{ExitCode: exitErr.ExitCode(), Err: err}
		}
		return output, fmt.Errorf("running script: %w", err)
	}
	return output, nil
}

// StartInteractive starts a process suitable for interactive usage inside the TUI.
// It prefers wrapping with `script` (PTY) when available so tools like gum detect
// a terminal; otherwise it falls back to direct execution.
func StartInteractive(entry registry.ScriptEntry, args []string, reg *registry.Registry) (*InteractiveSession, error) {
	executable, baseArgs := reg.ResolveInvocation(entry)

	var cmd *exec.Cmd
	if _, err := exec.LookPath("script"); err == nil {
		commandLine := shellJoin(append(append([]string{executable}, baseArgs...), args...))
		if runtime.GOOS == "darwin" {
			// BSD script: script -q /dev/null <command>
			cmd = exec.Command("script", "-q", "/dev/null", commandLine)
		} else {
			// util-linux script: script -qfec "<command>" /dev/null
			cmd = exec.Command("script", "-qfec", commandLine, "/dev/null")
		}
	} else {
		cmd = exec.Command(executable, append(baseArgs, args...)...)
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("opening stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("opening stdout pipe: %w", err)
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting interactive process: %w", err)
	}

	done := make(chan error, 1)
	go func() {
		err := cmd.Wait()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				done <- &RunError{ExitCode: exitErr.ExitCode(), Err: err}
				return
			}
			done <- fmt.Errorf("waiting interactive process: %w", err)
			return
		}
		done <- nil
	}()

	return &InteractiveSession{
		stdin:  stdin,
		stdout: stdout,
		done:   done,
	}, nil
}

// ReadChunk reads output from the interactive process.
func (s *InteractiveSession) ReadChunk(p []byte) (int, error) {
	return s.stdout.Read(p)
}

// SendInput forwards user input bytes to the interactive process.
func (s *InteractiveSession) SendInput(data []byte) error {
	_, err := s.stdin.Write(data)
	return err
}

// Done returns a channel that receives when the process exits.
func (s *InteractiveSession) Done() <-chan error {
	return s.done
}

func shellJoin(parts []string) string {
	quoted := make([]string, len(parts))
	for i, part := range parts {
		quoted[i] = shellQuote(part)
	}
	return strings.Join(quoted, " ")
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}

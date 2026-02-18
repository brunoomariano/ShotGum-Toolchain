package runner

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/shotgum/stg/internal/registry"
)

// RunError wraps a script execution error with its exit code.
type RunError struct {
	ExitCode int
	Err      error
}

func (e *RunError) Error() string {
	return fmt.Sprintf("script exited with code %d: %v", e.ExitCode, e.Err)
}

// Run executes a script, streaming stdout/stderr to the terminal.
func Run(entry registry.ScriptEntry, args []string, reg *registry.Registry) error {
	path := reg.ResolveScriptPath(entry)
	return run(entry.Type, path, args)
}

// RunHelp executes a script with its resolved help flag.
func RunHelp(entry registry.ScriptEntry, reg *registry.Registry) error {
	path := reg.ResolveScriptPath(entry)
	helpFlag := reg.ResolveHelpFlag(entry)
	return run(entry.Type, path, []string{helpFlag})
}

func run(scriptType, path string, args []string) error {
	var cmd *exec.Cmd
	switch scriptType {
	case "executable":
		cmd = exec.Command(path, args...)
	default: // "script" and default
		cmd = exec.Command("bash", append([]string{path}, args...)...)
	}

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
	path := reg.ResolveScriptPath(entry)
	return capture(entry.Type, path, args, nil)
}

// CaptureRunForPreview executes a script with extra env vars so gum can render
// without a real TTY (sets TERM and COLUMNS for lipgloss/gum width detection).
func CaptureRunForPreview(entry registry.ScriptEntry, args []string, reg *registry.Registry, width int) (string, error) {
	path := reg.ResolveScriptPath(entry)
	extraEnv := []string{
		"TERM=xterm-256color",
		fmt.Sprintf("COLUMNS=%d", width),
	}
	return capture(entry.Type, path, args, extraEnv)
}

func capture(scriptType, path string, args []string, extraEnv []string) (string, error) {
	var cmd *exec.Cmd
	switch scriptType {
	case "executable":
		cmd = exec.Command(path, args...)
	default:
		cmd = exec.Command("bash", append([]string{path}, args...)...)
	}

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

# Contributing to ShotGum Toolchain

## Prerequisites

- Go 1.22+
- `make` (GNU Make)
- `bash` (for script-based tests)

## Building

```bash
make build          # produces ./stg
make run            # build + run the TUI
make snapshot       # cross-build local binaries into ./dist
```

## Testing

```bash
make test           # tests with race detector
make cover          # tests + coverage summary
make ci             # fmt-check + vet + tests + coverage >= 85%
```

The minimum accepted total coverage is **85%**.

---

## Code Style

### Formatting

All code must pass `gofmt`. Run `make fmt` before committing.

### Language version

Go 1.22 — use builtin `max()` and `min()` when appropriate.

### Package structure

```text
cmd/stg/          — main entry point
internal/
  commands/       — Cobra commands (root, add, init, list, run)
  config/         — YAML schema + load/save + defaults
  registry/       — merged global/local view + resolution
  runner/         — Run/CaptureRun/StartInteractive
  tui/            — BubbleTea app (app.go)
  tui/styles/     — lipgloss palette + badges
  tui/views/      — categories, scripts, detail, output, confirm, header
  version/        — version string injected with ldflags
```

### Naming

- Exported types end with `Model` where applicable.
- Message types end with `Msg`.
- Constructors start with `New`.
- Cobra factory funcs end with `Cmd`.

### TUI architecture (current)

- Active state flow: `stateCategories -> stateScripts -> stateOutput`.
- `stateConfirm` exists in code but is not in the active navigation flow.
- `syncDetail()` returns `(AppModel, tea.Cmd)`; update call sites accordingly.
- `scriptList` must always be initialized with a valid delegate.
- Execution in TUI prioritizes interactive mode via `runner.StartInteractive()` with fallback to captured execution.
- `esc` is blocked in output while execution is still loading.

### Colors (lipgloss)

| Token | Hex |
|---|---|
| Purple | `#C084FC` |
| Teal | `#5EEAD4` |
| Gray | `#6B7280` |
| Red | `#F87171` |

If you introduce new visual tokens, update `internal/tui/styles/styles.go` and tests.

### Error handling

- Return errors instead of exiting inside packages.
- Wrap errors with context (`fmt.Errorf("context: %w", err)`).

---

## Writing Tests

### General rules

- Use stdlib `testing` package.
- Prefer table-driven tests for repeated scenarios.
- Avoid `t.Parallel()` in tests that mutate CWD or environment.

### Isolation for FS/HOME tests

```go
tmpdir := t.TempDir()
t.Setenv("HOME", tmpdir)

origDir, _ := os.Getwd()
t.Cleanup(func() { _ = os.Chdir(origDir) })
_ = os.Chdir(tmpdir)
```

### Testing Cobra commands

```go
cmd := addScriptCmd()
cmd.SetArgs([]string{scriptPath, "--category", "tools"})
cmd.SilenceUsage = true
cmd.SilenceErrors = true
err := cmd.Execute()
```

### Capturing stdout

```go
func captureStdout(f func()) string {
    old := os.Stdout
    r, w, _ := os.Pipe()
    os.Stdout = w
    f()
    _ = w.Close()

    var buf bytes.Buffer
    _, _ = io.Copy(&buf, r)
    os.Stdout = old
    return buf.String()
}
```

### BubbleTea model tests

```go
m := NewOutputModel("test", 80, 24)
m2, _ := m.Update(ScriptDoneMsg{Output: "ok", Err: nil})
if m2.IsLoading() { t.Error("expected not loading") }
```

Use ANSI stripping helpers in view assertions where needed.

---

## Commit & PR conventions

- Commit message in imperative mood.
- Keep one logical change per commit.
- PR must pass `make ci`.

## CI workflow

The workflow in `.github/workflows/ci.yml` runs `make ci` on push to `main` and on pull requests.

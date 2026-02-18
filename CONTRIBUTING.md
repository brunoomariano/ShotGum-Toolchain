# Contributing to ShotGum Toolchain

## Prerequisites

- Go 1.22+
- `make` (GNU Make)
- `bash` (for test scripts)

## Building

```bash
make build          # produces ./stg
make run            # build + run in testenv/
make snapshot       # GoReleaser snapshot (no publish)
```

## Testing

```bash
make test           # all tests with race detector
make cover          # tests + coverage summary
make ci             # fmt-check + vet + tests + coverage ≥ 85%
```

The minimum accepted total coverage is **85%** (currently ≥ 90%). `make ci` fails if this threshold is not met.

---

## Code Style

### Formatting

All code must pass `gofmt`. Run `make fmt` before committing. `make ci` enforces this.

### Language version

Go 1.22 — use the builtin `max()` and `min()` functions; do **not** import helpers for them.

### Package structure

```
cmd/stg/          — main entry point (thin, no logic)
internal/
  commands/       — Cobra subcommands (add, list, init, run, root)
  config/         — YAML config load/save + schema
  registry/       — merged global+local config view
  runner/         — script execution (Run, CaptureRun, CaptureRunForPreview)
  tui/            — BubbleTea application (app.go)
  tui/styles/     — lipgloss palette + Badge / TypeTag helpers
  tui/views/      — individual view models (categories, scripts, detail, confirm, output, header)
  version/        — version string injected at build time via ldflags
```

### Naming

- Exported types end in `Model` (e.g. `DetailModel`, `OutputModel`).
- Message types end in `Msg` (e.g. `ScriptDoneMsg`, `ConfirmRunMsg`).
- Constructor functions start with `New` (e.g. `NewOutputModel`, `NewConfirmModel`).
- Cobra command factories end in `Cmd` (e.g. `addFolderCmd`, `addScriptCmd`).

### TUI architecture (BubbleTea)

- State machine: `stateCategories → stateScripts → stateConfirm → stateOutput`.
- Two-panel layout: left panel (list) + right panel (`DetailModel`), header on top.
- Views return **raw content** (no border). The border is applied once inside `renderTwoPanel`.
- `syncDetail()` is a **value receiver** returning `AppModel`; always call as `m = m.syncDetail()`.
- `scriptList` must be initialized in `NewAppModel` — a zero-value `list.Model` has a nil delegate and panics on `SetSize`.

### Colors (lipgloss)

| Token | Hex |
|---|---|
| Purple | `#C084FC` |
| Teal | `#5EEAD4` |
| Gray | `#6B7280` |
| Red | `#F87171` |

Do not add new color constants without updating `internal/tui/styles/styles.go`.

### Error handling

- Return errors; do not `log.Fatal` or `os.Exit` inside packages.
- Format errors with `fmt.Errorf("context: %w", err)` for wrapping.
- Only validate at system boundaries (user input, file I/O, external commands).

---

## Writing Tests

### General rules

- Every test function **must have a doc comment** that explains what is being tested and which code path or branch it exercises. Example:

  ```go
  // TestOutputModel_Update_SpinnerTick_Loading verifies that a spinner.TickMsg while
  // loading advances the spinner animation and returns a follow-up tick command.
  func TestOutputModel_Update_SpinnerTick_Loading(t *testing.T) { … }
  ```

- Use the standard library only (`testing`, `os`, `path/filepath`, …). No third-party assertion libraries.
- Use table-driven tests (`[]struct{ … }`) when testing multiple similar inputs.
- Do **not** call `t.Parallel()` in tests that change the CWD — it causes races.

### Isolation

Tests that touch the filesystem or `HOME` must be fully isolated:

```go
tmpdir := t.TempDir()          // cleaned up automatically
t.Setenv("HOME", tmpdir)       // restored automatically by t.Cleanup

origDir, _ := os.Getwd()
t.Cleanup(func() { os.Chdir(origDir) })
os.Chdir(tmpdir)
```

### Testing Cobra commands

Execute subcommands in isolation via `cmd.Execute()`:

```go
cmd := addScriptCmd()
cmd.SetArgs([]string{scriptPath, "--category", "tools"})
cmd.SilenceUsage = true
cmd.SilenceErrors = true
err := cmd.Execute()
```

### Capturing stdout

Use `os.Pipe` to capture output from functions that write directly to `os.Stdout`:

```go
func captureStdout(f func()) string {
    old := os.Stdout
    r, w, _ := os.Pipe()
    os.Stdout = w
    f()
    w.Close()
    var buf bytes.Buffer
    io.Copy(&buf, r)
    os.Stdout = old
    return buf.String()
}
```

### BubbleTea model tests

Update models by calling their `Update` method directly, then inspect state:

```go
m := NewOutputModel("test", 80, 24)
m2, _ := m.Update(ScriptDoneMsg{Output: "ok", Err: nil})
if m2.loading { t.Error("…") }
```

To invoke a `tea.Cmd` and receive the message:

```go
func collectMsg(cmd tea.Cmd) tea.Msg {
    if cmd == nil { return nil }
    return cmd()
}
```

### Stripping ANSI codes

All view assertions should strip ANSI escape codes before using `strings.Contains`:

```go
// stripANSI is defined in internal/tui/views/header_test.go (package views)
stripped := stripANSI(view)
if !strings.Contains(stripped, "expected text") { … }
```

### Coverage target

All PRs must pass `make ci`, which enforces ≥ 85% total coverage.

---

## Commit & PR conventions

- Commit messages: imperative mood, present tense. Focus on *why*, not *what*.
- One logical change per commit.
- PRs must pass `make ci` (format + vet + tests + coverage).

## Release process

Releases are triggered by pushing a `v*` tag:

```bash
git tag v0.1.0
git push origin v0.1.0
```

The GitHub Actions workflow (`.github/workflows/release.yml`) builds cross-platform binaries and publishes a GitHub Release automatically.

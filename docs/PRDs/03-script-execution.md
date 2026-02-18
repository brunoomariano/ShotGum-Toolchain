# PRD 03 — Execução de Scripts

## Visão Geral

A execução é centralizada em `internal/runner/runner.go`.

## Resolução do Executável

Ordem de precedência (`Registry.ResolveExecutable`):

1. `script.executable`
2. `config.default_executable` do mesmo source
3. `global.default_executable`
4. `"/bin/sh"`

Se o valor não for absoluto, o sistema tenta `exec.LookPath()`.

## Modos de Execução

### 1. Execução direta (CLI sem TUI)

`stg <categoria> <script> [args...]`

- Usa `runner.Run()`
- Streaming: `Stdout/Stderr/Stdin` ligados ao terminal
- Retorna `RunError` com exit code quando aplicável

### 2. Execução na TUI (fluxo principal atual)

Ao pressionar `Enter`, `i` ou `?` em scripts:

- `app.runScript()` abre `stateOutput`
- chama `runner.StartInteractive()`
- se disponível, usa utilitário `script` (PTY) para melhor suporte a ferramentas interativas
- faz stream incremental para o viewport e encaminha teclas do usuário para o processo

### 3. Fallback capturado na TUI

Se `StartInteractive` falhar:

- usa `views.RunScriptCmd` -> `runner.CaptureRun()`
- output completo vem de `CombinedOutput()`

### 4. Preview de help no painel direito

- `views.LoadScriptHelpCmd()` chama `runner.CaptureRunForPreview()`
- injeta `TERM=xterm-256color` e `COLUMNS=<largura painel>`

## Erros e Exit Code

- `RunError{ExitCode, Err}` é usado em `Run`, `CaptureRun` e sessão interativa.
- No atalho de execução direta em `main.go`, o processo principal sai com o mesmo exit code do script.

## Regras de Negócio

| ID | Regra |
|---|---|
| R-RUN-01 | Todo script é executado como `<executable> <path> [args...]` |
| R-RUN-02 | `CaptureRun` usa `CombinedOutput()` e preserva output parcial em erro |
| R-RUN-03 | `Run` usa streaming real (`os.Stdout`, `os.Stderr`, `os.Stdin`) |
| R-RUN-04 | TUI prioriza execução interativa com PTY (`StartInteractive`) |
| R-RUN-05 | Fallback de `StartInteractive` é execução capturada |
| R-RUN-06 | Preview de help usa `TERM` e `COLUMNS` para renderização coerente |
| R-RUN-07 | `?` executa script com help flag resolvido via registry |

# PRD 05 — Comandos CLI

## Visão Geral

Há dois caminhos de roteamento:

1. Atalho em `cmd/stg/main.go` para `stg <categoria> <script> [args...]`
2. Cobra (`internal/commands/root.go`) para TUI e subcomandos

## Árvore de Comandos

```text
stg
├── (sem args)                -> TUI
├── <categoria>               -> TUI pré-posicionada na categoria
├── <categoria> <script> ...  -> execução direta
├── list [categoria]
├── add
│   ├── folder <path>
│   └── script <path>
└── init
```

## Atalho de Execução Direta (main.go)

Ativa quando:

- há pelo menos 2 args
- os 2 primeiros não começam com `-`
- o primeiro não é subcomando conhecido (`list`, `add`, `init`, etc.)

Objetivo: repassar flags extras ao script sem conflito de parsing do Cobra.

## Comando raiz (`Root()`)

Uso: `stg [category] [script] [args...]`

- `0 args` -> `tui.StartWithLogs(reg, --logs)`
- `1 arg` -> `tui.StartAtCategoryWithLogs(reg, category, --logs)`
- `>=2 args` -> `runDirect`

Flag persistente: `--logs` (exibe footer de logs no TUI).

## `stg list [category]`

- sem categoria: lista categorias
- com categoria: lista scripts da categoria

Origem é exibida como badge `[local]` ou `[user]`.

## `stg add folder <path>`

Flags:

- `--name` (default: basename do diretório)
- `--desc`

Comportamento:

- valida que o path existe e é diretório
- grava sempre na config global
- bloqueia categoria duplicada
- sugere comandos para registrar `*.sh` encontrados

## `stg add script <path>`

Flags:

- `--category` (obrigatória)
- `--name` (default: nome do arquivo sem extensão)
- `--desc`
- `--executable` (opcional)

Comportamento:

- valida que script existe
- exige categoria existente
- bloqueia script duplicado por categoria
- grava sempre na config global

## `stg init`

- se config global já existe: informa e sai
- se não existe: chama `config.EnsureDefault()` e mostra próximos passos

## Regras de Negócio

| ID | Regra |
|---|---|
| R-CLI-01 | `add` escreve apenas na config global |
| R-CLI-02 | `add folder` exige diretório válido |
| R-CLI-03 | `add script` exige `--category` e categoria existente |
| R-CLI-04 | Categoria única no global; script único por categoria |
| R-CLI-05 | `init` é idempotente |
| R-CLI-06 | Atalho do `main.go` preserva argumentos extras do script |
| R-CLI-07 | `stg <categoria>` inexistente mantém fallback para `stateCategories` |
| R-CLI-08 | `SilenceUsage: true` no comando raiz |
| R-CLI-09 | Campo suportado para interpretador é `--executable` |

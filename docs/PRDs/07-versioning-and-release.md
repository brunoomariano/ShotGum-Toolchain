# PRD 07 — Versionamento, Build, CI e Instalação

## Visão Geral

O projeto injeta versão em build via ldflags e possui pipeline de CI. Atualmente não há workflow de release automatizado no repositório.

## Versão em Build

Arquivo: `internal/version/version.go`

```go
var Version = "dev"
```

No `Makefile`:

- `VERSION := git describe --tags --always --dirty || dev`
- `LDFLAGS := -s -w -X 'github.com/shotgum/stg/internal/version.Version=$(VERSION)'`

## Targets Principais (`Makefile`)

- `make build`
- `make run`
- `make snapshot` (linux/darwin/windows, amd64/arm64)
- `make install` / `make uninstall`
- `make test` / `make cover` / `make ci`

Cobertura mínima no `ci`: `85%`.

## CI Atual

Workflow existente: `.github/workflows/ci.yml`

Trigger:

- push em `main`
- pull_request

Etapas:

1. checkout
2. setup go (cache)
3. `make ci`

## Instalador (`install.sh`)

O instalador atual baixa o código-fonte e compila localmente.

### Fluxo

1. Resolve ref de instalação:
   - `STG_REF`, fallback `STG_VERSION`, fallback `main`
2. Baixa tarball do GitHub (tag ou branch)
3. Executa `VERSION=<ref> make build`
4. Instala binário em `STG_INSTALL_DIR` (default `~/.local/bin`)
5. Opcionalmente:
   - copia `playground` para `~/playground`
   - instala scripts default (`defaults/scripts`) e tenta registrá-los via `stg add`

### Dependências exigidas

- `curl`
- `tar`
- `make`
- `go`

## Regras de Negócio

| ID | Regra |
|---|---|
| R-VER-01 | Fallback de versão em runtime é `dev` |
| R-VER-02 | Build padrão injeta versão via `-X` em `version.Version` |
| R-VER-03 | `make ci` exige cobertura mínima de 85% |
| R-VER-04 | CI atual executa apenas validação de qualidade (`make ci`) |
| R-VER-05 | Instalador atual compila do fonte (não baixa binário pronto) |
| R-VER-06 | Ref de instalação aceita branch ou tag via `STG_REF/STG_VERSION` |
| R-VER-07 | Instalação padrão é em `~/.local/bin` |
| R-VER-08 | Extras do instalador (playground/scripts) são opt-in interativos |

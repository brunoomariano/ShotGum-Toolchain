<p align="center">
  <img src="assets/logo.png" alt="ShotGum logo" width="320" />
</p>

# ShotGum (stg)

Gerenciador de scripts para terminal com CLI e TUI, construído com o ecossistema [Charmbracelet](https://charm.sh/).

O `stg` unifica scripts globais (usuário) e locais (projeto) em um catálogo navegável, com execução direta e execução interativa no terminal.

---

## Funcionalidades

- TUI com navegação por teclado e filtro
- Merge de configuração global (`~/.config/shotgum/config.yaml`) + local (`.shotgum.yaml`)
- Execução direta: `stg <categoria> <script> [args...]`
- Execução interativa na TUI (stream de output + input ao processo)
- Help preview assíncrono no Info Container
- Cascata de `help_flag`: script -> categoria -> config -> `--help`
- Resolução de executável com fallback (`script.executable`, `default_executable`, `/bin/sh`)
- Badges de origem: `[local]` e `[user]`

---

## Instalação

### Via script (recomendado)

```bash
curl -fsSL https://raw.githubusercontent.com/brunoomariano/ShotGum-Toolchain/main/install.sh | bash
```

O instalador baixa o código-fonte e compila localmente.

Variáveis opcionais:

```bash
STG_REF=v0.2.0              # tag ou branch (default: main)
STG_VERSION=v0.2.0          # alias compatível para STG_REF
STG_INSTALL_DIR=~/.local/bin
```

Pré-requisitos do instalador: `curl`, `tar`, `make`, `go`.

### A partir do código-fonte

```bash
git clone https://github.com/brunoomariano/ShotGum-Toolchain
cd ShotGum-Toolchain
make install
```

---

## Início rápido

```bash
# 1) Inicializa config global
stg init

# 2) Registra pasta como categoria
stg add folder ~/meus-scripts --name dev --desc "Scripts de desenvolvimento"

# 3) Registra script na categoria
stg add script ~/meus-scripts/build.sh --category dev --name build --executable bash

# 4) Lista
stg list

# 5) Abre TUI
stg
```

---

## Uso

### TUI

```bash
stg              # abre em categorias
stg <categoria>  # abre já na lista de scripts da categoria
stg --logs       # abre com footer de logs de execução
```

### Atalhos principais

| Tela | Tecla | Ação |
|---|---|---|
| Categorias | `Enter` | Entra na categoria |
| Categorias | `/` | Filtra |
| Categorias | `l` | Toggle logs |
| Scripts | `Enter` | Executa script |
| Scripts | `i` | Executa script (mesmo fluxo de `Enter`) |
| Scripts | `?` | Executa script com `help_flag` resolvido |
| Scripts | `tab` | Foca Info Container |
| Scripts | `esc` | Volta para categorias |
| Saída | `↑` `↓` `PgUp` `PgDn` | Scroll |
| Saída | `esc` | Volta para scripts (quando não estiver carregando) |

### Execução direta (sem TUI)

```bash
stg <categoria> <script>
stg <categoria> <script> [args...]

# args/flags são repassados ao script
stg git sync --branch main --push
```

### Listagem

```bash
stg list
stg list <categoria>
```

Exemplo:

```text
Categories
──────────────────────────────────────────────────
  [local] dev               Scripts do projeto
  [user]  infra             Scripts pessoais
```

### Adição de categoria/script

```bash
stg add folder <caminho> [--name <nome>] [--desc <descrição>]
stg add script <caminho> --category <categoria> [--name <nome>] [--desc <descrição>] [--executable <binário>]
```

`stg add` sempre grava na configuração global.

---

## Configuração

### Global: `~/.config/shotgum/config.yaml`

```yaml
version: "1"
scripts_home: "~/.shotgum/scripts"
help_flag: "--help"
default_executable: "/bin/sh"

categories:
  - name: docker
    description: "Docker scripts"
    scripts_path: "~/scripts/docker"
    help_flag: ""

scripts:
  - name: cleanup
    category: docker
    description: "Cleanup"
    executable: "bash"
    path: "cleanup.sh"
    help_flag: ""
```

### Local: `.shotgum.yaml`

```yaml
version: "1"
help_flag: "--help"
default_executable: "bash"

categories:
  - name: dev
    description: "Project scripts"
    scripts_path: "./scripts"

scripts:
  - name: build
    category: dev
    description: "Build"
    path: "./build.sh"
```

### Regras de resolução

- Merge: local sobrepõe global
- Path de script: absoluto -> `category.scripts_path` -> `scripts_home/<categoria>/`
- Help flag: `script.help_flag` -> `category.help_flag` -> `config.help_flag` -> `--help`
- Executável: `script.executable` -> `config.default_executable` (source) -> `global.default_executable` -> `/bin/sh`

---

## Estrutura do projeto

```text
cmd/stg/main.go
internal/
  commands/       # comandos Cobra
  config/         # schema + load/save + defaults
  registry/       # merge e resolução
  runner/         # execução direta/capturada/interativa
  tui/
    app.go        # state machine BubbleTea
    views/        # Header/Navigation/Info containers, output, confirm
    styles/       # estilos lipgloss
  version/        # version.Version (ldflags)
```

---

## Desenvolvimento

```bash
make build
make test
make ci
```

Workflow de CI: `.github/workflows/ci.yml` (push em `main` e pull requests).

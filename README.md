<p align="center">
  <img src="assets/logo.png" alt="ShotGum logo" width="320" />
</p>

# ShotGum (stg)

Um gerenciador de scripts TUI para o terminal, construído com o ecossistema [Charmbracelet](https://charm.sh/).

Desenvolvido para quem acumula scripts espalhados por diretórios e não tem uma interface unificada para descobri-los, documentá-los ou executá-los. O `stg` cria um ponto de entrada único — com visual limpo e navegação por teclado — para scripts pessoais globais e scripts de projeto via `.shotgum.yaml`.

---

## Funcionalidades

- **Interface TUI interativa** com navegação por teclado, filtragem e visual em pastéis suaves
- **Configuração em camadas** — config global (`~/.config/shotgum/config.yaml`) mesclada com config de projeto (`.shotgum.yaml`) em tempo de execução
- **Execução direta** sem abrir o TUI, com todos os argumentos e flags repassados ao script
- **Execução interativa** — `i` suspende o TUI e entrega o TTY completo ao script (gum prompts, spinners, etc.)
- **Dois tipos de script** — `script` (executado via `bash`) e `executable` (executado diretamente)
- **Cascata de help flag** — cada script/categoria/config pode ter seu próprio flag de ajuda
- **Preview de help** no painel direito, com scroll, ao navegar pelos scripts
- **Badges visuais** — `[local]` e `[global]` indicam a origem de cada categoria e script
- **Totalmente scriptável** — todos os comandos funcionam sem TUI para uso em automações

---

## Instalação

### Via script (recomendado)

```bash
curl -fsSL https://raw.githubusercontent.com/brunoomariano/ShotGum-Toolchain/main/install.sh | bash
```

Suporta Linux e macOS, x86_64 e arm64. Instala em `~/.local/bin` por padrão.

Variáveis de ambiente opcionais:

```bash
STG_VERSION=v0.2.0    # fixa uma versão específica
STG_INSTALL_DIR=/usr/local/bin  # sobrescreve o diretório de instalação
```

### A partir do código-fonte

Requer Go 1.22+.

```bash
git clone https://github.com/brunoomariano/ShotGum-Toolchain
cd stg
make install        # compila e instala em ~/.local/bin/stg
```

Ou, sem o Makefile:

```bash
go build -o ~/.local/bin/stg ./cmd/stg
```

---

## Início rápido

```bash
# 1. Inicializa a configuração global
stg init

# 2. Registra um diretório como categoria
stg add folder ~/meus-scripts --name dev --desc "Scripts de desenvolvimento"

# 3. Registra scripts individualmente
stg add script ~/meus-scripts/build.sh --category dev --desc "Faz o build do projeto"
stg add script ~/meus-scripts/deploy.sh --category dev --desc "Deploy para produção"

# 4. Lista tudo
stg list

# 5. Abre o TUI interativo
stg
```

---

## Uso

### Interface TUI

```
stg              # Abre o TUI com lista de categorias
stg <categoria>  # Abre o TUI já na lista de scripts da categoria
```

#### Atalhos de teclado

| Tela                  | Tecla           | Ação                                    |
|-----------------------|-----------------|-----------------------------------------|
| Categorias            | `Enter`         | Entra na categoria                      |
| Categorias            | `/`             | Filtra por nome                         |
| Categorias            | `q`             | Sai                                     |
| Scripts               | `Enter`         | Executa o script (saída capturada)      |
| Scripts               | `i`             | Executa de forma interativa (TTY completo) |
| Scripts               | `?`             | Executa com a help flag                 |
| Scripts               | `tab`           | Foca o painel de detalhes (scroll do help) |
| Scripts               | `/`             | Filtra por nome                         |
| Scripts               | `Esc`           | Volta para categorias                   |
| Scripts               | `q`             | Sai                                     |
| Detalhe (focado)      | `↑` `↓`         | Rola o painel de help                   |
| Detalhe (focado)      | `tab` / `Esc`   | Volta o foco para a lista               |
| Saída                 | `↑` `↓`         | Rola linha a linha                      |
| Saída                 | `PgUp` / `PgDn` | Rola página a página                    |
| Saída                 | `Esc` / `q`     | Volta para scripts                      |

### Execução direta (sem TUI)

```bash
stg <categoria> <script>
stg <categoria> <script> [argumentos...]

# Todos os argumentos e flags após o nome do script são repassados diretamente
stg docker cleanup --force
stg git sync --branch main --push
```

### Listagem

```bash
stg list               # Lista todas as categorias (global + local)
stg list <categoria>   # Lista os scripts de uma categoria
```

Exemplo de saída:

```
Categories
──────────────────────────────────────────────────
  [local]  dev      Project dev scripts
  [global] docker   Docker management scripts
  [global] git      Git workflow helpers
```

### Adicionar categorias e scripts

```bash
# Registra um diretório como categoria
stg add folder <caminho> [--name <nome>] [--desc <descrição>]

# Registra um script em uma categoria existente
stg add script <caminho> --category <categoria> [--name <nome>] [--desc <descrição>] [--type script|executable]
```

O `stg add folder` detecta automaticamente arquivos `.sh` no diretório e exibe os comandos sugeridos para registrá-los.

Se `--type` não for fornecido, o tipo é detectado automaticamente: arquivos `.sh` são registrados como `script`, demais arquivos como `executable`.

---

## Configuração

### Configuração global — `~/.config/shotgum/config.yaml`

```yaml
version: "1"
scripts_home: "~/.shotgum/scripts"
help_flag: "--help"

categories:
  - name: docker
    description: "Docker management scripts"
    scripts_path: "~/scripts/docker"
    help_flag: ""          # herda o global se vazio

  - name: git
    description: "Git workflow helpers"
    scripts_path: "/custom/path/git-scripts"
    help_flag: "-h"

scripts:
  - name: cleanup
    category: docker
    description: "Remove stopped containers and dangling images"
    type: script           # "script" | "executable"
    path: "cleanup.sh"     # relativo a scripts_path da categoria
    help_flag: ""
```

### Configuração de projeto — `.shotgum.yaml`

Coloque um arquivo `.shotgum.yaml` na raiz do projeto (pode ser commitado no git). O `stg` busca por ele percorrendo os diretórios acima do CWD, como o git faz com `.git`.

```yaml
version: "1"
help_flag: "--help"

categories:
  - name: dev
    description: "Project development scripts"
    scripts_path: "./scripts"

scripts:
  - name: build
    category: dev
    description: "Build the project"
    type: script
    path: "./scripts/build.sh"

  - name: test
    category: dev
    description: "Run test suite"
    type: script
    path: "./scripts/test.sh"
```

### Hierarquia de configuração

```
~/.config/shotgum/config.yaml   ← global (scripts pessoais)
.shotgum.yaml                   ← projeto (descoberto a partir do CWD)
```

Quando ambos existem, são **mesclados em tempo de execução**. Categorias e scripts locais têm precedência sobre os globais de mesmo nome. A origem de cada item é indicada no TUI e na saída do `stg list` com os badges `[local]` e `[global]`.

### Cascata de help flag

A flag de ajuda usada ao pressionar `?` no TUI (ou passada ao script via execução direta) é resolvida na seguinte ordem:

```
script.help_flag → category.help_flag → config.help_flag → "--help"
```

### Resolução de caminhos

- Caminhos absolutos (`/` ou `~`) são usados diretamente
- Caminhos relativos são resolvidos em relação a `category.scripts_path`
- Se `scripts_path` estiver vazio, usa `scripts_home/<nome-da-categoria>/`
- `~` e variáveis de ambiente (`$VAR`) são expandidas em todos os caminhos

---

## Estrutura do projeto

```
cmd/stg/main.go                 # Ponto de entrada
internal/
├── config/
│   ├── schema.go               # Structs Config, Category, Script
│   └── config.go               # LoadGlobal, LoadLocal, Save, EnsureDefault
├── registry/
│   └── registry.go             # Mescla configs, resolução de paths e flags
├── runner/
│   └── runner.go               # Execução de scripts (streaming e captura)
├── tui/
│   ├── app.go                  # Máquina de estados BubbleTea
│   ├── views/
│   │   ├── header.go           # Header animado com logo e versão
│   │   ├── categories.go       # Vista de lista de categorias
│   │   ├── scripts.go          # Vista de lista de scripts
│   │   ├── detail.go           # Painel direito: info + viewport de help
│   │   ├── confirm.go          # Vista de confirmação antes de executar
│   │   └── output.go           # Vista de saída com scroll e spinner
│   └── styles/
│       └── styles.go           # Paleta de cores e estilos Lipgloss
└── version/
    └── version.go              # Versão injetada via ldflags na build
```

---

## Dependências

| Biblioteca | Uso |
|---|---|
| [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) | Framework TUI (Elm Architecture) |
| [charmbracelet/bubbles](https://github.com/charmbracelet/bubbles) | Componentes: list, viewport, spinner |
| [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) | Estilos e layout terminal |
| [spf13/cobra](https://github.com/spf13/cobra) | CLI e parsing de argumentos |
| [gopkg.in/yaml.v3](https://pkg.go.dev/gopkg.in/yaml.v3) | Leitura e escrita de configs YAML |

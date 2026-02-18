# PRD 01 — Sistema de Configuração

## Visão Geral

O ShotGum usa arquivos YAML para definir categorias e scripts. Existem dois escopos de configuração: **global** (usuário) e **local** (projeto). Ambos são carregados e mesclados pelo Registry com precedência local.

---

## Schema de Dados

### Config (raiz do YAML)

```yaml
version: "1"
scripts_home: "~/.shotgum/scripts"   # diretório padrão para scripts globais
help_flag: "--help"                   # flag padrão de help para scripts

categories:
  - name: dev
    description: "Development scripts"
    scripts_path: "./scripts"         # relativo ao arquivo YAML
    help_flag: "--help"               # override de categoria (opcional)

scripts:
  - name: build
    category: dev
    description: "Build the project"
    type: script                      # "script" | "executable"
    path: "./scripts/build.sh"
    help_flag: "--help"               # override de script (opcional)
```

### Estruturas Go

```go
// internal/config/schema.go

type Config struct {
    Version     string     `yaml:"version"`
    ScriptsHome string     `yaml:"scripts_home"`
    HelpFlag    string     `yaml:"help_flag"`
    Categories  []Category `yaml:"categories"`
    Scripts     []Script   `yaml:"scripts"`
    Source      string     // "global" | "local" — injetado no load, não no YAML
}

type Category struct {
    Name        string `yaml:"name"`
    Description string `yaml:"description"`
    ScriptsPath string `yaml:"scripts_path"`
    HelpFlag    string `yaml:"help_flag"`
}

type Script struct {
    Name        string `yaml:"name"`
    Category    string `yaml:"category"`
    Description string `yaml:"description"`
    Type        string `yaml:"type"`   // "script" | "executable"
    Path        string `yaml:"path"`
    HelpFlag    string `yaml:"help_flag"`
}
```

---

## Hierarquia de Configuração

```mermaid
graph TD
    subgraph "Escopo Global"
        GF["~/.config/shotgum/config.yaml"]
    end

    subgraph "Escopo Local (projeto)"
        LF[".shotgum.yaml\n(descoberto upward a partir do CWD)"]
    end

    Registry["Registry\n(Merger com precedência local)"]

    GF -->|"Source = 'global'"| Registry
    LF -->|"Source = 'local'"| Registry

    Registry --> Categories["Categorias mescladas"]
    Registry --> Scripts["Scripts mesclados"]
```

---

## Fluxo de Loading

```mermaid
flowchart TD
    Start([Inicializa Registry]) --> LoadGlobal

    LoadGlobal["LoadGlobal()\n~/.config/shotgum/config.yaml"] --> GExists{Arquivo existe?}
    GExists -->|Não| GNil["global = nil"]
    GExists -->|Sim| GParse["Faz parse YAML\nDefine Source = 'global'"]
    GParse --> GExpand["Expande paths\n(~, $ENV_VAR)"]
    GExpand --> GLoaded["global = &Config{...}"]

    GNil --> LoadLocal
    GLoaded --> LoadLocal

    LoadLocal["LoadLocal()\nWalk CWD upward"] --> FindYAML{".shotgum.yaml\nencontrado?"}
    FindYAML -->|Não| LNil["local = nil"]
    FindYAML -->|Sim| LParse["Faz parse YAML\nDefine Source = 'local'"]
    LParse --> LExpand["Expande paths\n(~, $ENV_VAR)"]
    LExpand --> LLoaded["local = &Config{...}"]

    LNil --> Done
    LLoaded --> Done

    Done([Registry pronto])
```

---

## Descoberta do Arquivo Local

O arquivo `.shotgum.yaml` é descoberto caminhando do CWD para a raiz, semelhante ao comportamento do `git`:

```mermaid
flowchart LR
    CWD["/home/user/proj/src/cmd"] -->|"Não encontrou"| P1
    P1["/home/user/proj/src"] -->|"Não encontrou"| P2
    P2["/home/user/proj"] -->|"Encontrou!"| Found[".shotgum.yaml"]
    Found --> Stop([Para a busca])
```

**Regra:** A busca para na primeira ocorrência de `.shotgum.yaml` encontrada ao subir nos diretórios. Se chegar na raiz do sistema de arquivos sem encontrar, `local = nil`.

---

## Expansão de Paths

Todos os paths em configuração passam por `expandPath()` antes de uso:

| Entrada | Resultado |
|---|---|
| `~/scripts` | `/home/user/scripts` |
| `$HOME/scripts` | `/home/user/scripts` |
| `$PROJECT_DIR/scripts` | `[valor de $PROJECT_DIR]/scripts` |
| `./scripts` | Mantido relativo (resolvido pelo Registry later) |
| `/absolute/path` | Mantido como está |

---

## Inicialização (First Run)

Quando o usuário executa `stg init`, o sistema garante a existência da configuração global padrão:

```mermaid
flowchart TD
    Init["stg init"] --> CheckDir{"~/.config/shotgum/\nexiste?"}
    CheckDir -->|Não| CreateDir["MkdirAll ~/.config/shotgum/"]
    CheckDir -->|Sim| CheckFile{"config.yaml\nexiste?"}
    CreateDir --> CheckFile
    CheckFile -->|Não| WriteDefault["Escreve config.yaml\ncom defaults"]
    CheckFile -->|Sim| Skip["Nenhuma ação"]
    WriteDefault --> CreateScripts["MkdirAll ~/.shotgum/scripts/"]
    CreateScripts --> Done([Configuração pronta])
    Skip --> Done
```

**Config padrão gerada:**
```yaml
version: "1"
scripts_home: "~/.shotgum/scripts"
help_flag: "--help"
categories: []
scripts: []
```

---

## Regras de Negócio

| # | Regra |
|---|---|
| R-CFG-01 | `Source` nunca é escrito no YAML — é injetado em memória no momento do load |
| R-CFG-02 | Um arquivo YAML ausente resulta em `nil`, não em erro — o sistema continua sem ele |
| R-CFG-03 | A descoberta local segue o CWD no momento de início do processo `stg` |
| R-CFG-04 | Paths com `~` são expandidos usando `os.UserHomeDir()` |
| R-CFG-05 | Paths com `$VAR` são expandidos usando `os.ExpandEnv()` |
| R-CFG-06 | `version: "1"` é o único schema suportado atualmente |
| R-CFG-07 | `scripts_home` só é usado como base de resolução quando um path de script é relativo e não há `scripts_path` de categoria |
| R-CFG-08 | `stg init` é idempotente — nunca sobrescreve config existente |

# PRD 02 — TUI: State Machine e Fluxo de Navegação

## Visão Geral

A interface TUI do `stg` é construída com o framework BubbleTea (arquitetura Elm). O estado da aplicação é representado por um único struct `AppModel` que transita entre estados bem definidos conforme o usuário interage.

---

## Estados da Aplicação

```mermaid
stateDiagram-v2
    [*] --> stateCategories : NewAppModel()

    stateCategories --> stateScripts : Enter (seleciona categoria)
    stateScripts --> stateCategories : Esc / Backspace
    stateScripts --> stateConfirm : Enter (seleciona script)
    stateScripts --> stateOutput : i (run interativo)
    stateConfirm --> stateScripts : Esc (cancela)
    stateConfirm --> stateOutput : Enter (confirma execução)
    stateOutput --> stateScripts : Esc / q (fecha output)

    stateCategories --> [*] : Ctrl+C / q
    stateScripts --> [*] : Ctrl+C
    stateOutput --> [*] : Ctrl+C
```

---

## Mapa de Key Bindings por Estado

### stateCategories

| Tecla | Ação |
|---|---|
| `↑` / `k` | Move seleção para cima |
| `↓` / `j` | Move seleção para baixo |
| `Enter` | Entra na categoria selecionada → `stateScripts` |
| `/` | Ativa filtro de lista |
| `q` / `Ctrl+C` | Sai do programa |

### stateScripts

| Tecla | Ação |
|---|---|
| `↑` / `k` | Move seleção para cima |
| `↓` / `j` | Move seleção para baixo |
| `Enter` | Confirma script selecionado → `stateConfirm` |
| `i` | Executa script interativo (full TTY) → suspende TUI |
| `?` | Carrega help do script no painel direito |
| `Esc` / `Backspace` | Volta para `stateCategories` |
| `/` | Ativa filtro de lista |
| `Ctrl+C` | Sai do programa |

### stateConfirm

| Tecla | Ação |
|---|---|
| `Enter` | Executa script → `stateOutput` |
| `a` | Foca no campo de argumentos extras |
| `Esc` | Cancela → volta para `stateScripts` |

### stateOutput

| Tecla | Ação |
|---|---|
| `↑` / `k` | Scroll up no viewport |
| `↓` / `j` | Scroll down no viewport |
| `PgUp` / `b` | Page up |
| `PgDn` / `f` | Page down |
| `g` | Vai para o início |
| `G` | Vai para o final |
| `Esc` / `q` | Fecha output → volta para `stateScripts` |
| `Ctrl+C` | Sai do programa |

---

## Fluxo Completo de Navegação

```mermaid
flowchart TD
    Start(["stg (sem args)"]) --> Init["NewAppModel()\nCarrega Registry\nInicia header ticker"]

    Init --> CatView["stateCategories\nExibe lista de categorias"]

    CatView -->|"↑/↓ navega"| CatView
    CatView -->|"/ filtra"| CatView
    CatView -->|"Enter"| ScriptView

    ScriptView["stateScripts\nExibe lista de scripts\nda categoria selecionada"]

    ScriptView -->|"↑/↓ navega"| ScriptView
    ScriptView -->|"Esc/Backspace"| CatView
    ScriptView -->|"?"| LoadHelp["LoadScriptHelpCmd()\nasync via Cmd"]
    LoadHelp -->|"DetailHelpMsg"| ScriptView

    ScriptView -->|"i"| Interactive
    Interactive["tea.ExecProcess\nSuspende TUI\nDá TTY completo ao script"]
    Interactive -->|"Script termina"| ScriptView

    ScriptView -->|"Enter"| Confirm["stateConfirm\nDialog de confirmação\nCampo de args extras"]

    Confirm -->|"Esc"| ScriptView
    Confirm -->|"Enter"| RunScript

    RunScript["runner.CaptureRun()\nCaptura stdout+stderr"] --> Output

    Output["stateOutput\nViewport com output\nSpinner enquanto carrega"]
    Output -->|"↑/↓/PgUp/PgDn"| Output
    Output -->|"Esc/q"| ScriptView

    CatView -->|"q/Ctrl+C"| Exit([Sai])
```

---

## Fluxo de Inicialização (NewAppModel)

```mermaid
flowchart TD
    A["NewAppModel(registry)"] --> B["Carrega categorias do Registry"]
    B --> C["Cria catList\n(list.Model com CategoryItems)"]
    C --> D["Cria scriptList\n(list.Model com delegate)"]

    D -->|"IMPORTANTE: scriptList DEVE ser inicializado\naqui ou panic em SetSize()"| E

    E["Cria DetailModel\n(modo detailNone)"]
    E --> F["Inicia header ticker\n(HeaderTickMsg a cada 280ms)"]
    F --> G["state = stateCategories"]
    G --> Return([AppModel pronto])
```

> **Atenção:** `list.Model` com valor zero tem delegate `nil`, o que causa panic ao chamar `SetSize()`. O `scriptList` deve ser inicializado com um delegate válido em `NewAppModel`, mesmo que vazio.

---

## Fluxo de Redimensionamento (WindowSizeMsg)

```mermaid
flowchart TD
    Resize["tea.WindowSizeMsg\n(terminal redimensionou)"] --> Store["Salva width, height\nem AppModel"]
    Store --> Calc["panelDims()\nCalcula leftW, rightW, panelH"]
    Calc --> SetCat["catList.SetSize(leftW, panelH)"]
    SetCat --> SetScript["scriptList.SetSize(leftW, panelH)"]
    SetScript --> SetDetail["detail.SetSize(rightW, panelH)"]
    SetDetail --> Done([Layout atualizado])
```

### Fórmula de Dimensionamento

```
leftW  = max((w - 4) * 2 / 5, 10)
rightW = w - leftW - 4        (margens e borda)
panelH = max(h - 6, 5)        (header=3 + helpbar=1 + borders=2)
```

---

## Fluxo de Carregamento de Help Assíncrono

```mermaid
sequenceDiagram
    participant User
    participant App as app.go (AppModel)
    participant Detail as detail.go (DetailModel)
    participant Runner as runner.go

    User->>App: Pressiona "?"
    App->>App: state permanece stateScripts
    App->>Detail: syncDetail() com script selecionado
    Detail->>Detail: detailMode = detailScript
    Detail->>App: retorna LoadScriptHelpCmd() como tea.Cmd
    App->>Runner: CaptureRunForPreview(script --help)
    Note over Runner: Executa com TERM=xterm-256color\nCOLUMNS=rightW
    Runner-->>App: DetailHelpMsg{content}
    App->>Detail: detail.SetHelp(content)
    Detail->>Detail: Atualiza viewport com help text
```

---

## Componentes do Layout

```
┌─────────────────────────────────────────────────────────┐
│  HEADER (3 linhas)                                      │
│  · ° o O  SHOTGUM  ¬══════►  v1.2.3  [link]           │
├──────────────────────┬──────────────────────────────────┤
│                      │                                  │
│  LEFT PANEL          │  RIGHT PANEL (DetailModel)       │
│  (catList /          │                                  │
│   scriptList)        │  - Category mode:                │
│                      │    scripts, source, path         │
│  leftW = max(        │  - Script mode:                  │
│    (w-4)*2/5, 10)    │    path, help flag, metadata     │
│                      │  - Help viewport (scrollable)    │
│                      │                                  │
├──────────────────────┴──────────────────────────────────┤
│  HELP BAR (1 linha) — atalhos contextuais               │
└─────────────────────────────────────────────────────────┘
```

---

## Regras de Negócio

| # | Regra |
|---|---|
| R-TUI-01 | O estado inicial sempre é `stateCategories` |
| R-TUI-02 | `stateOutput` não volta para `stateCategories` — sempre para `stateScripts` |
| R-TUI-03 | `tea.ExecProcess` suspende completamente o TUI — o terminal pertence ao script durante execução interativa |
| R-TUI-04 | `syncDetail()` é value receiver — deve ser chamado como `m = m.syncDetail()` |
| R-TUI-05 | O help de script (`?`) é carregado de forma assíncrona via `tea.Cmd` — nunca bloqueia o event loop |
| R-TUI-06 | `panelDims()` garante mínimos: `leftW >= 10`, `panelH >= 5` para evitar layouts quebrados em terminais pequenos |
| R-TUI-07 | Bordas são aplicadas uma única vez em `renderTwoPanel()` — as views retornam conteúdo sem borda |
| R-TUI-08 | O header ticker envia `HeaderTickMsg` a cada 280ms para animação dos frames |

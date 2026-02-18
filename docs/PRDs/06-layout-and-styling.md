# PRD 06 — Layout e Estilos

## Visão Geral

Layout principal em `AppModel.View()`:

- header
- dois painéis (lista + detalhe)
- help bar
- footer opcional de logs

`stateOutput` usa tela de output própria.

## Paleta (`internal/tui/styles/styles.go`)

| Token | Hex | Uso |
|---|---|---|
| Purple | `#C084FC` | títulos, borda ativa, seleção |
| Teal | `#5EEAD4` | categoria, links, badge local |
| Gray | `#6B7280` | descrições, ajuda, status, badge user |
| Subtle | `#374151` | borda inativa |
| White | `#F9FAFB` | nome de script |
| Red | `#F87171` | erros |

Badge de origem:

- `local` -> `[local]`
- default (`user`) -> `[user]`

## Dimensões (`panelDims`)

```text
leftW  = max((w-4)*2/5, 10)
rightW = max((w-4)-leftW, 10)
panelH = max(h-6, 5)
```

## Componentes Visuais

### Header (`views/header.go`)

- título `SHOTGUM`
- versão (`v` + `version.Version`)
- link de issues

### Listas (`views/categories.go` e `views/scripts.go`)

- usam `bubbles/list`
- agrupamento por origem com `SectionHeaderItem` (`Local` / `User`)
- cursor inicia no primeiro item real (pula header de seção)

### Painel de detalhe (`views/detail.go`)

Modo categoria:

- source, path e contagem de scripts

Modo script:

- category, source, executable, command, path, help flag
- viewport scrollável para help

### Output (`views/output.go`)

- spinner durante loading
- modo interativo mostra stream incremental
- help text muda quando está recebendo input ao vivo

### Footer de logs (`app.go`)

- buffer circular (até 80 eventos)
- janela visível de 6 linhas
- toggle por `l` ou `--logs`

## Componente de confirmação

`views/confirm.go` existe e possui implementação completa de dialog, porém não está conectado ao fluxo principal atual da state machine.

## Regras de Negócio

| ID | Regra |
|---|---|
| R-UI-01 | Borda dos painéis é aplicada em `renderTwoPanel()` |
| R-UI-02 | `stateOutput` renderiza tela própria |
| R-UI-03 | Mínimos de layout evitam quebra em terminal pequeno |
| R-UI-04 | `detail.SetSize()` é aplicado em `WindowSizeMsg` |
| R-UI-05 | `tab` alterna foco visual entre lista e detalhe |
| R-UI-06 | Scroll percentual aparece quando aplicável no viewport |
| R-UI-07 | Footer de logs é opcional e não altera estado de navegação |
| R-UI-08 | `confirm.go` é componente disponível, mas não ativo no fluxo atual |

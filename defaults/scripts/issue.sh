#!/usr/bin/env bash
# issue.sh — 🐛 Guide to open a well-documented GitHub issue
set -euo pipefail

PURPLE="#C084FC"
TEAL="#5EEAD4"
RED="#F87171"
YELLOW="#FCD34D"
GRAY="#6B7280"
GREEN="#4ADE80"

ISSUES_URL="$(stg --repo-url)/issues/new"

# ─── Help ─────────────────────────────────────────────────────────────────────

show_help() {
  gum style \
    --border rounded --border-foreground "$PURPLE" \
    --padding "1 2" --margin "1 0" \
    "$(gum style --bold --foreground "$PURPLE" "issue — 🐛 Reportar um problema no GitHub")

$(gum style --foreground "$GRAY" "Uso:")
  stg shotgum issue
  stg shotgum issue --help

$(gum style --foreground "$TEAL" "O que faz:")
  Guia você pelos elementos de uma boa issue, coleta
  informações do ambiente automaticamente e abre o
  GitHub para criar o reporte."
}

if [[ "${1:-}" == "--help" ]]; then
  show_help
  exit 0
fi

IS_TTY=false
[ -t 1 ] && IS_TTY=true

# ─── Collect environment info ─────────────────────────────────────────────────

STG_VERSION=$(stg --version 2>/dev/null | head -1 || echo "(não disponível no PATH)")
OS_INFO="$(uname -s) $(uname -r) $(uname -m)"
SHELL_NAME="${SHELL##*/}"
SHELL_VERSION=$("$SHELL" --version 2>/dev/null | head -1 || echo "?")
GUM_VERSION=$(gum --version 2>/dev/null | head -1 || echo "(não encontrado)")
TERM_INFO="${TERM:-não definido}"
LOCAL_CFG="false"
[[ -f ".shotgum.yaml" ]] && LOCAL_CFG="true (.shotgum.yaml presente)"

# ─── Header ───────────────────────────────────────────────────────────────────

echo
gum style \
  --border double \
  --border-foreground "$RED" \
  --padding "1 4" \
  --align center \
  "$(gum style --bold --foreground "$RED" "🐛  Reportar um Problema")
$(gum style --foreground "$GRAY" "Issues bem documentadas são resolvidas mais rápido.")"

echo

# ─── Guide ────────────────────────────────────────────────────────────────────

gum style \
  --border rounded \
  --border-foreground "$PURPLE" \
  --padding "1 2" \
  "$(gum style --bold --foreground "$PURPLE" "  Como escrever uma boa issue")

$(gum style --bold --foreground "$TEAL" "  1. Título")
     $(gum style --foreground "$GRAY" "Uma frase curta e específica descrevendo o problema.")
     $(gum style --foreground "$GREEN" "  ✓ Bom:") 'stg congela ao rodar script com gum input sem TTY'
     $(gum style --foreground "$RED"   "  ✗ Ruim:") 'não funciona'

$(gum style --bold --foreground "$TEAL" "  2. Passos para Reproduzir")
     $(gum style --foreground "$GRAY" "Liste exatamente o que você fez para chegar no erro.")
     $(gum style --foreground "$GRAY" "  1. Execute 'stg' em um diretório com .shotgum.yaml")
     $(gum style --foreground "$GRAY" "  2. Selecione a categoria X e o script Y")
     $(gum style --foreground "$GRAY" "  3. Pressione Enter (não 'i')")
     $(gum style --foreground "$GRAY" "  4. Observe o comportamento inesperado")

$(gum style --bold --foreground "$TEAL" "  3. Comportamento Esperado vs Observado")
     $(gum style --foreground "$GREEN" "  Esperado:") O script executa e exibe a saída na viewport.
     $(gum style --foreground "$RED"   "  Observado:") O TUI trava sem exibir erro.

$(gum style --bold --foreground "$TEAL" "  4. Logs e Saída do Terminal")
     $(gum style --foreground "$GRAY" "Copie a saída completa do erro. Use blocos de código")
     $(gum style --foreground "$GRAY" "com \`\`\` no GitHub Markdown para formatar corretamente.")

$(gum style --bold --foreground "$TEAL" "  5. Contexto Adicional")
     $(gum style --foreground "$GRAY" "Screenshots, .shotgum.yaml relevante, configurações")
     $(gum style --foreground "$GRAY" "especiais, versão de ferramentas externas (gum, etc.).")"

echo

# ─── Environment info ─────────────────────────────────────────────────────────

gum style \
  --border rounded \
  --border-foreground "$TEAL" \
  --padding "1 2" \
  "$(gum style --bold --foreground "$TEAL" "  🖥️  Seu Ambiente")
  $(gum style --foreground "$GRAY" "Copie e cole este bloco na sua issue:")

  $(gum style --foreground "$GRAY" "stg version:    ") $STG_VERSION
  $(gum style --foreground "$GRAY" "OS:             ") $OS_INFO
  $(gum style --foreground "$GRAY" "Shell:          ") $SHELL_NAME — $SHELL_VERSION
  $(gum style --foreground "$GRAY" "TERM:           ") $TERM_INFO
  $(gum style --foreground "$GRAY" "gum version:    ") $GUM_VERSION
  $(gum style --foreground "$GRAY" "Config local:   ") $LOCAL_CFG"

echo

# ─── Quick template ───────────────────────────────────────────────────────────

gum style \
  --border rounded \
  --border-foreground "$YELLOW" \
  --padding "1 2" \
  "$(gum style --bold --foreground "$YELLOW" "  📋  Template rápido (copie para a issue)")

$(gum style --foreground "$GRAY" "
## Descrição
<!-- Explique o problema em 1-2 frases -->

## Passos para Reproduzir
1.
2.
3.

## Comportamento Esperado
<!-- O que deveria acontecer -->

## Comportamento Observado
<!-- O que aconteceu de fato -->

## Ambiente
\`\`\`
stg version:  $STG_VERSION
OS:           $OS_INFO
Shell:        $SHELL_NAME
TERM:         $TERM_INFO
gum:          $GUM_VERSION
\`\`\`

## Contexto adicional
<!-- Screenshots, logs, .shotgum.yaml, etc. -->")"

echo

# ─── Open browser ─────────────────────────────────────────────────────────────

if $IS_TTY; then
  if gum confirm \
    --affirmative "Abrir no GitHub" \
    --negative "Agora não" \
    "  Abrir o GitHub para criar a issue?"; then

    opened=false
    if command -v xdg-open &>/dev/null; then
      xdg-open "$ISSUES_URL" 2>/dev/null & opened=true
    elif command -v open &>/dev/null; then
      open "$ISSUES_URL" && opened=true
    fi

    echo
    if $opened; then
      gum style --foreground "$PURPLE" --bold "  Obrigado por reportar! 🙏"
      gum style --foreground "$GRAY" "  Lembre de incluir o bloco de ambiente acima."
    else
      gum style --foreground "$YELLOW" "  Não foi possível abrir o navegador automaticamente."
      gum style --foreground "$TEAL" "  Acesse: $ISSUES_URL"
    fi
  fi
else
  gum style --foreground "$TEAL" "  Para abrir a issue: $ISSUES_URL"
fi

echo

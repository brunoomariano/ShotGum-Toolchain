#!/usr/bin/env bash
# star.sh — ⭐ Guide to star ShotGum on GitHub
set -euo pipefail

PURPLE="#C084FC"
TEAL="#5EEAD4"
YELLOW="#FCD34D"
GRAY="#6B7280"
GREEN="#4ADE80"

REPO_URL=$(stg --repo-url)

# ─── Help ─────────────────────────────────────────────────────────────────────

show_help() {
  gum style \
    --border rounded --border-foreground "$PURPLE" \
    --padding "1 2" --margin "1 0" \
    "$(gum style --bold --foreground "$PURPLE" "star — ⭐ Dar estrela no ShotGum")

$(gum style --foreground "$GRAY" "Uso:")
  stg shotgum star
  stg shotgum star --help

$(gum style --foreground "$TEAL" "O que faz:")
  Apresenta o projeto ShotGum, mostra o link do GitHub e
  oferece abrir o navegador para você dar uma estrela."
}

if [[ "${1:-}" == "--help" ]]; then
  show_help
  exit 0
fi

IS_TTY=false
[ -t 1 ] && IS_TTY=true

# ─── Header ───────────────────────────────────────────────────────────────────

echo
gum style \
  --border double \
  --border-foreground "$YELLOW" \
  --padding "1 4" \
  --align center \
  "$(gum style --bold --foreground "$YELLOW" "⭐  Gostou do ShotGum?")
$(gum style --foreground "$GRAY" "Cada estrela ajuda o projeto a crescer.")"

echo

# ─── Project card ─────────────────────────────────────────────────────────────

STG_VERSION=$(stg --version 2>/dev/null | head -1 || echo "dev")

gum style \
  --border rounded \
  --border-foreground "$PURPLE" \
  --padding "1 3" \
  "$(gum style --bold --foreground "$PURPLE" "  ShotGum  (stg)")  $(gum style --foreground "$GRAY" "versão:") $(gum style --foreground "$TEAL" "$STG_VERSION")

  $(gum style --foreground "$GRAY" "Um gerenciador de scripts TUI para o terminal.")
  $(gum style --foreground "$GRAY" "Construído com o ecossistema Charmbracelet.")

  $(gum style --bold --foreground "$TEAL" "🔗  $REPO_URL")"

echo

# ─── What a star does ─────────────────────────────────────────────────────────

gum style \
  --border rounded \
  --border-foreground "$GREEN" \
  --padding "0 2" \
  "$(gum style --bold --foreground "$GREEN" "  Por que dar estrela?")

  $(gum style --foreground "$GRAY" "★") Ajuda outras pessoas a descobrir a ferramenta
  $(gum style --foreground "$GRAY" "★") Mostra ao autor que o projeto é útil
  $(gum style --foreground "$GRAY" "★") Motiva novas funcionalidades e manutenção"

echo

# ─── Open browser ─────────────────────────────────────────────────────────────

if $IS_TTY; then
  if gum confirm \
    --affirmative "Abrir no navegador" \
    --negative "Agora não" \
    "  Abrir o GitHub para dar uma estrela?"; then

    opened=false
    if command -v xdg-open &>/dev/null; then
      xdg-open "$REPO_URL" 2>/dev/null & opened=true
    elif command -v open &>/dev/null; then
      open "$REPO_URL" && opened=true
    fi

    echo
    if $opened; then
      gum style \
        --foreground "$GREEN" \
        --bold \
        "  🙏 Obrigado pelo apoio!"
    else
      gum style --foreground "$YELLOW" "  Não foi possível abrir o navegador automaticamente."
      gum style --foreground "$TEAL" "  Acesse: $REPO_URL"
    fi
  fi
else
  gum style --foreground "$TEAL" "  Para dar estrela, acesse: $REPO_URL"
fi

echo

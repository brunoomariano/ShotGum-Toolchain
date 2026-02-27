#!/usr/bin/env bash
# forca.sh — Jogo da Forca com Bash + gum
set -euo pipefail

PURPLE="#C084FC"
TEAL="#5EEAD4"
RED="#F87171"
YELLOW="#FCD34D"
GREEN="#4ADE80"
GRAY="#6B7280"

# ─── Ajuda ────────────────────────────────────────────────────────────────────

show_help() {
  gum style \
    --border rounded --border-foreground "$PURPLE" \
    --padding "1 2" --margin "1 0" \
    "$(gum style --bold --foreground "$PURPLE" "forca — Jogo da Forca (Bash + gum)")

$(gum style --foreground "$GRAY" "Uso:")
  forca.sh
  forca.sh --help

$(gum style --foreground "$TEAL" "Como jogar:")
  Uma palavra secreta é escolhida aleatoriamente.
  Digite uma letra por vez e pressione Enter.
  Você tem 6 erros antes de perder.
  A dica abaixo da forca ajuda a descobrir a palavra.

$(gum style --foreground "$TEAL" "Requisito:")
  Terminal interativo — use a tecla $(gum style --bold --foreground "$YELLOW" "'i'") no stg."
}

if [[ "${1:-}" == "--help" ]]; then
  show_help
  exit 0
fi

IS_TTY=false
[ -t 1 ] && IS_TTY=true

if ! $IS_TTY; then
  echo "forca: requer terminal interativo."
  echo "  No stg, pressione 'i' para rodar de forma interativa."
  exit 1
fi

# ─── Palavras (TERMO:dica) ────────────────────────────────────────────────────

declare -a PALAVRAS=(
  "PYTHON:linguagem de programação criada nos anos 90"
  "TERMINAL:onde digitamos comandos no computador"
  "ALGORITMO:sequência de passos para resolver um problema"
  "VARIAVEL:espaço na memória que guarda um valor"
  "COMPILADOR:transforma código-fonte em executável"
  "DEPURACAO:processo de encontrar e corrigir erros"
  "FRAMEWORK:estrutura base para desenvolver aplicações"
  "PROTOCOLO:conjunto de regras para comunicação em rede"
  "BINARIO:sistema numérico de base dois"
  "SERVIDOR:máquina que fornece serviços na rede"
  "RECURSIVIDADE:quando uma função chama a si mesma"
  "SINTAXE:conjunto de regras de escrita de uma linguagem"
  "REPOSITORIO:onde o código-fonte fica armazenado"
  "INTERFACE:ponto de contato entre dois sistemas"
)

# ─── Estágios da forca ───────────────────────────────────────────────────────

get_stage() {
  case "$1" in
    0) printf '  +---+\n  |   |\n      |\n      |\n      |\n      |\n=========' ;;
    1) printf '  +---+\n  |   |\n  O   |\n      |\n      |\n      |\n=========' ;;
    2) printf '  +---+\n  |   |\n  O   |\n  |   |\n      |\n      |\n=========' ;;
    3) printf '  +---+\n  |   |\n  O   |\n /|   |\n      |\n      |\n=========' ;;
    4) printf '  +---+\n  |   |\n  O   |\n /|\  |\n      |\n      |\n=========' ;;
    5) printf '  +---+\n  |   |\n  O   |\n /|\  |\n /    |\n      |\n=========' ;;
    6) printf '  +---+\n  |   |\n  O   |\n /|\  |\n / \  |\n      |\n=========' ;;
  esac
}

MAX_ERROS=6

# ─── Funções auxiliares ───────────────────────────────────────────────────────

contains() { [[ "$1" == *"$2"* ]]; }

# Verifica se todas as letras da palavra foram acertadas
word_done() {
  local i
  for (( i=0; i<${#PALAVRA}; i++ )); do
    local c="${PALAVRA:$i:1}"
    contains "$ACERTOS" "$c" || return 1
  done
  return 0
}

# Exibe a palavra com letras reveladas e underscores
show_word() {
  local display="" i
  for (( i=0; i<${#PALAVRA}; i++ )); do
    local c="${PALAVRA:$i:1}"
    if contains "$ACERTOS" "$c"; then
      display+="$(gum style --foreground "$GREEN" --bold "$c") "
    else
      display+="$(gum style --foreground "$GRAY" "_") "
    fi
  done
  echo "$display"
}

# ─── Renderização ─────────────────────────────────────────────────────────────

render() {
  clear

  # Header
  gum style \
    --border rounded \
    --border-foreground "$PURPLE" \
    --bold --foreground "$PURPLE" \
    --padding "0 4" \
    "🪤  JOGO DA FORCA"

  echo

  # Forca — vermelho no game over, amarelo nos outros estágios
  local forca_color="$YELLOW"
  [[ "$ERROS" -eq "$MAX_ERROS" ]] && forca_color="$RED"

  while IFS= read -r line; do
    gum style --foreground "$forca_color" "  $line"
  done < <(get_stage "$ERROS")

  echo

  # Palavra revelada
  echo "  $(show_word)"
  echo

  # Dica
  gum style --foreground "$GRAY" "  Dica: $DICA"
  echo

  # Tentativas restantes
  local rest=$(( MAX_ERROS - ERROS ))
  local rest_color="$GREEN"
  (( rest <= 3 )) && rest_color="$YELLOW"
  (( rest <= 1 )) && rest_color="$RED"

  echo "  $(gum style --foreground "$GRAY" "Erros:") \
$(gum style --foreground "$RED" "$ERROS")/$MAX_ERROS   \
$(gum style --foreground "$GRAY" "Restam:") \
$(gum style --foreground "$rest_color" "$rest") tentativa(s)"
  echo

  # Letras erradas
  if [[ -n "$ERRADAS" ]]; then
    echo "  $(gum style --foreground "$GRAY" "Erradas:") \
$(gum style --foreground "$RED" "$ERRADAS")"
  else
    gum style --foreground "$GRAY" "  (nenhuma letra errada ainda)"
  fi
  echo
}

# ─── Selecionar palavra ───────────────────────────────────────────────────────

nova_palavra() {
  local entry="${PALAVRAS[$RANDOM % ${#PALAVRAS[@]}]}"
  PALAVRA="${entry%%:*}"
  DICA="${entry##*:}"
}

# ─── Loop do jogo ─────────────────────────────────────────────────────────────

jogar() {
  ACERTOS=""
  ERRADAS=""
  ERROS=0
  nova_palavra

  while (( ERROS < MAX_ERROS )); do
    render

    if word_done; then
      break
    fi

    LETRA=$(
      gum input \
        --placeholder "letra..." \
        --prompt "  » " \
        --prompt.foreground "$PURPLE" \
        --char-limit 1 \
      | tr '[:lower:]' '[:upper:]'
    ) || { echo; echo "$(gum style --foreground "$RED" "  Jogo interrompido.")"; exit 0; }

    # Validação
    if [[ -z "$LETRA" ]] || ! [[ "$LETRA" =~ ^[A-Z]$ ]]; then
      gum style --foreground "$YELLOW" "  ⚠  Digite apenas uma letra."
      sleep 0.7
      continue
    fi

    if contains "$ACERTOS" "$LETRA" || contains "$ERRADAS" "$LETRA"; then
      gum style --foreground "$YELLOW" "  ⚠  Você já tentou essa letra!"
      sleep 0.7
      continue
    fi

    if contains "$PALAVRA" "$LETRA"; then
      ACERTOS+="$LETRA"
    else
      ERRADAS+="$LETRA "
      (( ERROS++ )) || true
    fi
  done

  # Tela final
  render

  if word_done; then
    gum style \
      --border rounded \
      --border-foreground "$GREEN" \
      --bold --foreground "$GREEN" \
      --padding "1 3" \
      "🎉 Parabéns! Você descobriu: $PALAVRA"
  else
    gum style \
      --border double \
      --border-foreground "$RED" \
      --bold --foreground "$RED" \
      --padding "1 3" \
      "💀 Game over! A palavra era: $PALAVRA"
  fi

  echo

  if gum confirm "  Jogar novamente?"; then
    jogar
  fi
}

# ─── Início ───────────────────────────────────────────────────────────────────

jogar

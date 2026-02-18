#!/usr/bin/env bash
set -euo pipefail

PURPLE="#C084FC"
TEAL="#5EEAD4"
GRAY="#6B7280"
YELLOW="#FCD34D"
WHITE="#F9FAFB"

if [[ "${1:-}" == "--show-me" ]]; then
  gum style \
    --border rounded --border-foreground "$PURPLE" \
    --padding "1 2" --margin "1 0" \
    "$(gum style --bold --foreground "$PURPLE" "env — Print project environment variables")

$(gum style --foreground "$GRAY" "Usage:") env [OPTIONS]

$(gum style --foreground "$TEAL" "--all")        Include all env vars, not just project ones
$(gum style --foreground "$TEAL" "--export")     Output as export VAR=value statements
$(gum style --foreground "$TEAL" "--show-me")    Show this message"
  exit 0
fi

IS_TTY=false
[ -t 1 ] && IS_TTY=true

ALL=false
EXPORT=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --all)    ALL=true; shift ;;
    --export) EXPORT=true; shift ;;
    *)        gum log --level error "Unknown option: $1"; exit 1 ;;
  esac
done

# Project-relevant vars with fallback labels
declare -A VARS
declare -a KEYS=(APP_ENV DATABASE_URL REDIS_URL LOG_LEVEL PORT SECRET_KEY DEBUG)

VARS[APP_ENV]="${APP_ENV:-}"
VARS[DATABASE_URL]="${DATABASE_URL:-}"
VARS[REDIS_URL]="${REDIS_URL:-}"
VARS[LOG_LEVEL]="${LOG_LEVEL:-}"
VARS[PORT]="${PORT:-}"
VARS[SECRET_KEY]="${SECRET_KEY:-}"
VARS[DEBUG]="${DEBUG:-}"

if $EXPORT; then
  for key in "${KEYS[@]}"; do
    val="${VARS[$key]}"
    [[ -z "$val" ]] && continue
    echo "export $key=\"$val\""
  done
  exit 0
fi

echo ""
gum style \
  --border rounded --border-foreground "$PURPLE" \
  --padding "1 2" \
  "$(gum style --bold --foreground "$PURPLE" "Project Environment")"
echo ""

row() {
  local key="$1" val="${VARS[$1]:-}"
  local val_rendered
  if [[ -z "$val" ]]; then
    val_rendered=$(gum style --foreground "$GRAY" --italic "unset")
  elif [[ "$key" == *SECRET* || "$key" == *PASSWORD* || "$key" == *TOKEN* ]]; then
    val_rendered=$(gum style --foreground "$YELLOW" "••••••••")
  else
    val_rendered=$(gum style --foreground "$WHITE" "$val")
  fi
  printf "  %s  %s\n" \
    "$(gum style --foreground "$TEAL" "$(printf "%-14s" "$key")")" \
    "$val_rendered"
}

for key in "${KEYS[@]}"; do
  row "$key"
done

# Interactive filter when ALL flag is set in a terminal
if $ALL && $IS_TTY; then
  echo ""
  gum style --foreground "$GRAY" --italic "  Filter all variables:"
  echo ""
  env | sort | gum filter --placeholder "search..." \
    --prompt "  › " --prompt.foreground "$PURPLE" \
    --match.foreground "$TEAL"
elif $ALL; then
  echo ""
  gum style --foreground "$GRAY" --italic "  All variables:"
  echo ""
  env | sort | while IFS= read -r line; do
    key="${line%%=*}"
    val="${line#*=}"
    printf "  %s  %s\n" \
      "$(gum style --foreground "$TEAL" "$(printf "%-20s" "$key")")" \
      "$(gum style --foreground "$GRAY" "$val")"
  done
fi

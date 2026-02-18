#!/usr/bin/env bash
set -euo pipefail

PURPLE="#C084FC"
TEAL="#5EEAD4"
RED="#F87171"
YELLOW="#FCD34D"
GRAY="#6B7280"
WHITE="#F9FAFB"

if [[ "${1:-}" == "--show-me" ]]; then
  gum style \
    --border rounded --border-foreground "$PURPLE" \
    --padding "1 2" --margin "1 0" \
    "$(gum style --bold --foreground "$PURPLE" "status — Service health dashboard")

$(gum style --foreground "$GRAY" "Usage:") status [OPTIONS]

$(gum style --foreground "$TEAL" -- "--json")        Output as JSON
$(gum style --foreground "$TEAL" -- "--watch")       Refresh every 3 seconds
$(gum style --foreground "$TEAL" -- "--show-me")     Show this message"
  exit 0
fi

IS_TTY=false
[ -t 1 ] && IS_TTY=true

JSON=false
WATCH=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --json)  JSON=true; shift ;;
    --watch) WATCH=true; shift ;;
    *)       gum log --level error "Unknown option: $1"; exit 1 ;;
  esac
done

# Simulated service data: name latency status
declare -a SERVICES=(
  "api:18ms:ok"
  "worker:—:ok"
  "postgres:4ms:ok"
  "redis:1ms:ok"
  "scheduler:—:warn"
  "mailer:—:down"
)

render_dashboard() {
  local ts
  ts=$(date "+%H:%M:%S")

  if $JSON; then
    echo "["
    for svc in "${SERVICES[@]}"; do
      IFS=':' read -r name latency state <<< "$svc"
      printf '  {"service": "%s", "latency": "%s", "status": "%s"},\n' "$name" "$latency" "$state"
    done
    echo "]"
    return
  fi

  local rows=""
  local up=0 warn_count=0 down=0

  for svc in "${SERVICES[@]}"; do
    IFS=':' read -r name latency state <<< "$svc"

    local badge icon lat_rendered
    case "$state" in
      ok)
        badge=$(gum style --foreground "$TEAL"   --bold "● ok  ")
        icon="✓"
        ((up++)) || true
        ;;
      warn)
        badge=$(gum style --foreground "$YELLOW" --bold "● warn")
        icon="⚠"
        ((warn_count++)) || true
        ;;
      down)
        badge=$(gum style --foreground "$RED"    --bold "● down")
        icon="✗"
        ((down++)) || true
        ;;
    esac

    local lat_col
    if [[ "$latency" == "—" ]]; then
      lat_col=$(gum style --foreground "$GRAY" "     —")
    else
      lat_col=$(gum style --foreground "$GRAY" "$(printf "%6s" "$latency")")
    fi

    rows+=$(printf "  %s  %-14s  %s\n" \
      "$badge" \
      "$(gum style --foreground "$WHITE" "$name")" \
      "$lat_col")
    rows+=$'\n'
  done

  local total=${#SERVICES[@]}
  local summary
  summary=$(gum join --horizontal \
    "$(gum style --foreground "$TEAL"   "  ✓ $up up")" \
    "$(gum style --foreground "$YELLOW" "  ⚠ $warn_count warn")" \
    "$(gum style --foreground "$RED"    "  ✗ $down down")")

  local header
  header=$(gum join --horizontal \
    "$(gum style --bold --foreground "$PURPLE" "Service Dashboard")" \
    "$(gum style --foreground "$GRAY" "  updated $ts")")

  gum style \
    --border rounded --border-foreground "$PURPLE" \
    --padding "1 2" \
    "$header

$(gum style --foreground "$GRAY" "  STATUS  SERVICE         LATENCY")
$(gum style --foreground "$GRAY" "  ──────────────────────────────")
${rows}
$summary"
}

spin() { local title="$1"; shift
  if $IS_TTY; then gum spin --spinner pulse --title "  $title" -- "$@"
  else "$@"
  fi
}

if $WATCH && $IS_TTY; then
  while true; do
    clear
    spin "Checking services..." sleep 0.4
    render_dashboard
    echo ""
    gum style --foreground "$GRAY" --italic "  Refreshing in 3s — Ctrl+C to stop"
    sleep 3
  done
else
  spin "Checking services..." sleep 0.5
  echo ""
  render_dashboard
fi

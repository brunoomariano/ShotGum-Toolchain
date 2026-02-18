#!/usr/bin/env bash
set -euo pipefail

PURPLE="#C084FC"
TEAL="#5EEAD4"
RED="#F87171"
YELLOW="#FCD34D"
GRAY="#6B7280"
WHITE="#F9FAFB"

show_help() {
  gum style \
    --border rounded --border-foreground "$PURPLE" \
    --padding "1 2" --margin "1 0" \
    "$(gum style --bold --foreground "$PURPLE" "logs — Tail recent application logs")

$(gum style --foreground "$GRAY" "Usage:") logs [OPTIONS]

$(gum style --foreground "$TEAL" -- "--lines N")     Number of entries to show (default: 20)
$(gum style --foreground "$TEAL" -- "--level LEVEL") Filter: debug | info | warn | error
$(gum style --foreground "$TEAL" -- "--follow")      Keep tailing (simulated)
$(gum style --foreground "$TEAL" -- "--help")        Show this message"
}

if [[ "${1:-}" == "--help" ]]; then
  show_help
  exit 0
fi

IS_TTY=false
[ -t 1 ] && IS_TTY=true

LINES=20
LEVEL="all"
FOLLOW=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --lines)  LINES="$2"; shift 2 ;;
    --level)  LEVEL="$2"; shift 2 ;;
    --follow) FOLLOW=true; shift ;;
    *)        gum log --level error "Unknown option: $1"; exit 1 ;;
  esac
done

if $IS_TTY && [[ "$LEVEL" == "all" ]]; then
  gum style --foreground "$PURPLE" --bold "  Application Logs"
  echo ""

  LEVEL=$(gum choose --header "  Filter by level:" \
    --cursor.foreground "$PURPLE" \
    --selected.foreground "$TEAL" \
    "all" "debug" "info" "warn" "error")
fi

echo ""
gum style --foreground "$PURPLE" --bold "  Logs $(gum style --foreground "$GRAY" "— last $LINES entries  [level: $LEVEL]")"
echo ""

log_line() {
  local ts="$1" lvl="$2" msg="$3"
  [[ "$LEVEL" != "all" && "$LEVEL" != "$lvl" ]] && return 0

  local color="$GRAY"
  case "$lvl" in
    debug) color="$GRAY" ;;
    info)  color="$TEAL" ;;
    warn)  color="$YELLOW" ;;
    error) color="$RED" ;;
  esac

  local badge
  badge=$(gum style --foreground "$color" --bold "$(printf "%-5s" "$lvl")")
  printf "  %s  %s  %s\n" \
    "$(gum style --foreground "$GRAY" "$ts")" \
    "$badge" \
    "$(gum style --foreground "$WHITE" "$msg")"
}

log_line "10:01:03" "info"  "server started on :8080"
log_line "10:01:11" "info"  "GET /api/health → 200 (1ms)"
log_line "10:01:44" "debug" "cache miss: registry.categories"
log_line "10:02:01" "info"  "GET /api/scripts → 200 (4ms)"
log_line "10:02:18" "debug" "cache hit: registry.categories"
log_line "10:02:45" "warn"  "slow query detected (312ms): SELECT * FROM scripts"
log_line "10:03:10" "info"  "POST /api/scripts/run → 202 (11ms)"
log_line "10:03:11" "debug" "runner: spawning bash /usr/local/scripts/build.sh"
log_line "10:03:14" "info"  "runner: script exited with code 0"
log_line "10:04:02" "info"  "GET /api/categories → 200 (2ms)"
log_line "10:04:55" "warn"  "config reload triggered — file changed"
log_line "10:05:17" "error" "connection timeout: db-primary:5432 (retry 1/3)"
log_line "10:05:19" "error" "connection timeout: db-primary:5432 (retry 2/3)"
log_line "10:05:22" "info"  "connection restored: db-primary:5432"
log_line "10:05:30" "info"  "GET /api/health → 200 (2ms)"
log_line "10:06:01" "debug" "GC pause: 1.2ms"
log_line "10:06:44" "info"  "POST /api/scripts/run → 202 (8ms)"
log_line "10:06:45" "debug" "runner: spawning bash /usr/local/scripts/deploy.sh"
log_line "10:07:51" "warn"  "deploy: canary at 50% — watching error rate"
log_line "10:08:12" "info"  "deploy: rollout complete (3/3 instances)"

if $FOLLOW; then
  echo ""
  gum style --foreground "$GRAY" --italic "  (following — Ctrl+C to stop)"
  sleep 1
  log_line "10:08:30" "info"  "GET /api/health → 200 (1ms)"
  sleep 1
  log_line "10:08:31" "debug" "heartbeat ok"
fi

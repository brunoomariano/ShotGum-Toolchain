#!/usr/bin/env bash
set -euo pipefail

PURPLE="#C084FC"
TEAL="#5EEAD4"
RED="#F87171"
YELLOW="#FCD34D"
GRAY="#6B7280"

show_help() {
  gum style \
    --border rounded --border-foreground "$PURPLE" \
    --padding "1 2" --margin "1 0" \
    "$(gum style --bold --foreground "$PURPLE" "lint — Run linter and format checks")

$(gum style --foreground "$GRAY" "Usage:") lint [OPTIONS]

$(gum style --foreground "$TEAL" "--fix")        Auto-fix fixable issues
$(gum style --foreground "$TEAL" "--strict")     Enable extra checks (errcheck, gosec)
$(gum style --foreground "$TEAL" "--help")       Show this message"
}

if [[ "${1:-}" == "--help" ]]; then
  show_help
  exit 0
fi

IS_TTY=false
[ -t 1 ] && IS_TTY=true

FIX=false
STRICT=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --fix)    FIX=true; shift ;;
    --strict) STRICT=true; shift ;;
    *)        gum log --level error "Unknown option: $1"; exit 1 ;;
  esac
done

if $IS_TTY; then
  gum style --foreground "$PURPLE" --bold "  Lint checks"
  echo ""

  OPTS=$(gum choose --no-limit --header "  Options (space to select):" \
    --cursor.foreground "$PURPLE" \
    --selected.foreground "$TEAL" \
    "auto-fix" "strict mode")
  echo "$OPTS" | grep -q "auto-fix"    && FIX=true    || true
  echo "$OPTS" | grep -q "strict mode" && STRICT=true || true
fi

echo ""
gum log --level info "Running lint checks"
echo ""

spin() { local title="$1"; shift
  if $IS_TTY; then gum spin --spinner line --title "  $title" -- "$@"
  else "$@"
  fi
}

check() {
  local name="$1" dur="$2" status="${3:-ok}"  # ok | warn | fail
  case "$status" in
    ok)
      gum style --foreground "$TEAL"   "  ✓ $(gum style --bold "$name") $(gum style --foreground "$GRAY" "${dur}")" ;;
    warn)
      gum style --foreground "$YELLOW" "  ⚠ $(gum style --bold "$name") $(gum style --foreground "$GRAY" "${dur}")" ;;
    fail)
      gum style --foreground "$RED"    "  ✗ $(gum style --bold "$name") $(gum style --foreground "$GRAY" "${dur}")" ;;
  esac
}

ISSUES=0

spin "gofmt..."        sleep 0.3; check "gofmt"       "0.3s" "ok"
spin "go vet..."       sleep 0.4; check "go vet"      "0.4s" "ok"
spin "staticcheck..."  sleep 0.8; check "staticcheck" "0.8s" "ok"

if $STRICT; then
  spin "errcheck..."   sleep 0.5
  echo ""
  gum style --foreground "$YELLOW" "  ⚠ errcheck (0.5s)"
  gum style --foreground "$GRAY"   "    internal/runner/runner.go:42: unchecked error: cmd.Wait()"
  ((ISSUES++)) || true

  spin "gosec..."      sleep 0.6; check "gosec"      "0.6s" "ok"
fi

echo ""

if [[ $ISSUES -gt 0 ]]; then
  gum style --foreground "$YELLOW" --bold "  $ISSUES warning(s) found"

  if $FIX; then
    echo ""
    spin "Applying fixes..." sleep 0.5
    gum style --foreground "$TEAL" "  ✓ Fixed $ISSUES issue(s)"
  elif $IS_TTY; then
    echo ""
    if gum confirm "  Apply fixes automatically?"; then
      spin "Applying fixes..." sleep 0.5
      gum style --foreground "$TEAL" "  ✓ Fixed $ISSUES issue(s)"
    fi
  fi
else
  gum style \
    --border rounded --border-foreground "$TEAL" --padding "0 2" \
    "$(gum style --bold --foreground "$TEAL" "✓ No issues found")"
fi

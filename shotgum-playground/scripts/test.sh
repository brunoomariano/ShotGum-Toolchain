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
    "$(gum style --bold --foreground "$PURPLE" "test — Run the test suite")

$(gum style --foreground "$GRAY" "Usage:") test [OPTIONS]

$(gum style --foreground "$TEAL" -- "--unit")           Run only unit tests
$(gum style --foreground "$TEAL" -- "--integration")    Run only integration tests
$(gum style --foreground "$TEAL" -- "--coverage")       Generate coverage report
$(gum style --foreground "$TEAL" -- "--verbose")        Show individual test output
$(gum style --foreground "$TEAL" -- "--help")           Show this message"
}

if [[ "${1:-}" == "--help" ]]; then
  show_help
  exit 0
fi

IS_TTY=false
[ -t 1 ] && IS_TTY=true

UNIT=false
INTEGRATION=false
COVERAGE=false
VERBOSE=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --unit)        UNIT=true; shift ;;
    --integration) INTEGRATION=true; shift ;;
    --coverage)    COVERAGE=true; shift ;;
    --verbose)     VERBOSE=true; shift ;;
    *)             gum log --level error "Unknown option: $1"; exit 1 ;;
  esac
done

if $IS_TTY && ! $UNIT && ! $INTEGRATION; then
  gum style --foreground "$PURPLE" --bold "  Run tests"
  echo ""

  SUITE=$(gum choose --header "  Which suite?" \
    --cursor.foreground "$PURPLE" \
    --selected.foreground "$TEAL" \
    "all" "unit" "integration")
  [[ "$SUITE" == "unit" ]]        && UNIT=true
  [[ "$SUITE" == "integration" ]] && INTEGRATION=true

  if gum confirm "  Generate coverage report?"; then
    COVERAGE=true
  fi
fi

SUITE_LABEL="all"
$UNIT        && SUITE_LABEL="unit"
$INTEGRATION && SUITE_LABEL="integration"

echo ""
gum log --level info "Running $SUITE_LABEL tests"
echo ""

spin() { local title="$1"; shift
  if $IS_TTY; then gum spin --spinner points --title "  $title" -- "$@"
  else "$@"
  fi
}

spin "Discovering tests..." sleep 0.4

declare -A TESTS
TESTS=(
  ["TestRegistryMerge"]="unit:12ms"
  ["TestConfigLoad"]="unit:3ms"
  ["TestResolveHelpFlag"]="unit:2ms"
  ["TestResolveScriptPath"]="unit:4ms"
  ["TestRunnerCapture"]="integration:85ms"
  ["TestTUIStateTransitions"]="integration:140ms"
  ["TestLocalConfigDiscovery"]="integration:67ms"
)

PASS=0; FAIL=0; SKIP=0

for name in "${!TESTS[@]}"; do
  info="${TESTS[$name]}"
  kind="${info%%:*}"
  dur="${info##*:}"

  # Filter by suite
  if $UNIT && [[ "$kind" != "unit" ]]; then continue; fi
  if $INTEGRATION && [[ "$kind" != "integration" ]]; then continue; fi

  spin "  $name" sleep 0.1

  # Simulate one flaky test
  if [[ "$name" == "TestRunnerCapture" ]] && $INTEGRATION; then
    gum style --foreground "$RED" "  ✗ $name $(gum style --foreground "$GRAY" "(${dur})")"
    gum style --foreground "$RED" --italic "    assertion failed: expected exit code 0, got 1"
    ((FAIL++)) || true
  else
    gum style --foreground "$TEAL" "  ✓ $name $(gum style --foreground "$GRAY" "(${dur})")"
    ((PASS++)) || true
  fi
done

echo ""

if $COVERAGE; then
  spin "Generating coverage report..." sleep 0.5
  echo ""
  gum style --foreground "$TEAL" "  Coverage:"
  gum style --foreground "$GRAY"  "    internal/config      91.2%"
  gum style --foreground "$GRAY"  "    internal/registry    87.6%"
  gum style --foreground "$GRAY"  "    internal/runner      78.4%"
  gum style --foreground "$YELLOW" "    internal/tui         62.1%"
  echo ""
fi

TOTAL=$((PASS + FAIL + SKIP))
if [[ $FAIL -gt 0 ]]; then
  gum style \
    --border rounded --border-foreground "$RED" --padding "0 2" \
    "$(gum style --bold --foreground "$RED" "✗ $FAIL/$TOTAL tests failed")"
  exit 1
else
  gum style \
    --border rounded --border-foreground "$TEAL" --padding "0 2" \
    "$(gum style --bold --foreground "$TEAL" "✓ $PASS/$TOTAL tests passed")"
fi

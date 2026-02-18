#!/usr/bin/env bash
set -euo pipefail

PURPLE="#C084FC"
TEAL="#5EEAD4"
RED="#F87171"
GRAY="#6B7280"

show_help() {
  gum style \
    --border rounded --border-foreground "$PURPLE" \
    --padding "1 2" --margin "1 0" \
    "$(gum style --bold --foreground "$PURPLE" "build — Build the project binary")

$(gum style --foreground "$GRAY" "Usage:") build [OPTIONS]

$(gum style --foreground "$TEAL" -- "--release")       Build in release mode (default: debug)
$(gum style --foreground "$TEAL" -- "--target DIR")    Output directory    (default: ./build)
$(gum style --foreground "$TEAL" -- "--help")          Show this message"
}

if [[ "${1:-}" == "--help" ]]; then
  show_help
  exit 0
fi

IS_TTY=false
[ -t 1 ] && IS_TTY=true

RELEASE=false
TARGET=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --release) RELEASE=true; shift ;;
    --target)  TARGET="$2"; shift 2 ;;
    *)         gum log --level error "Unknown option: $1"; exit 1 ;;
  esac
done

# Interactive prompts when running in a terminal with no flags
if $IS_TTY && [[ -z "$TARGET" ]] && [[ "$RELEASE" == "false" ]]; then
  gum style --foreground "$PURPLE" --bold "  Build the project"
  echo ""

  MODE=$(gum choose --header "  Build mode:" \
    --cursor.foreground "$PURPLE" \
    --selected.foreground "$TEAL" \
    "debug" "release")
  [[ "$MODE" == "release" ]] && RELEASE=true

  TARGET=$(gum input \
    --placeholder "./build" \
    --prompt "  Output dir › " \
    --prompt.foreground "$PURPLE")
fi

TARGET="${TARGET:-./build}"
MODE_LABEL="debug"
$RELEASE && MODE_LABEL="release"

echo ""
gum log --level info "Starting build"
gum log --level info "mode=$MODE_LABEL target=$TARGET"
echo ""

spin() { local title="$1"; shift
  if $IS_TTY; then gum spin --spinner dot --title "  $title" -- "$@"
  else "$@"
  fi
}

spin "Downloading modules..."       sleep 0.6
gum style --foreground "$TEAL" "  ✓ Modules ready"

spin "Compiling packages..."        sleep 1.0
gum style --foreground "$TEAL" "  ✓ Compiled 14 packages"

spin "Linking binary..."            sleep 0.4
gum style --foreground "$TEAL" "  ✓ Binary linked"

spin "Running post-build checks..."  sleep 0.3
gum style --foreground "$TEAL" "  ✓ Checks passed"

echo ""
gum style \
  --border rounded --border-foreground "$TEAL" \
  --padding "0 2" \
  "$(gum style --bold --foreground "$TEAL" "✓ Build complete") $(gum style --foreground "$GRAY" "→ $TARGET/app  [$MODE_LABEL]")"

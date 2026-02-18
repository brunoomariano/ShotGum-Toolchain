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
    "$(gum style --bold --foreground "$PURPLE" "clean — Remove build artifacts and cache")

$(gum style --foreground "$GRAY" "Usage:") clean [OPTIONS]

$(gum style --foreground "$TEAL" -- "--all")        Also clear module cache
$(gum style --foreground "$TEAL" -- "--dry-run")    Show what would be removed
$(gum style --foreground "$TEAL" -- "--help")       Show this message"
}

if [[ "${1:-}" == "--help" ]]; then
  show_help
  exit 0
fi

IS_TTY=false
[ -t 1 ] && IS_TTY=true

ALL=false
DRY_RUN=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --all)     ALL=true; shift ;;
    --dry-run) DRY_RUN=true; shift ;;
    *)         gum log --level error "Unknown option: $1"; exit 1 ;;
  esac
done

if $IS_TTY && ! $ALL; then
  gum style --foreground "$PURPLE" --bold "  Clean project"
  echo ""

  if gum confirm "  Remove module cache as well?"; then
    ALL=true
  fi
fi

echo ""
gum style --foreground "$PURPLE" --bold "  Files to remove:"
echo ""

FILES=(
  "./build/          $(gum style --foreground "$GRAY" "(build output)")"
  "./stg             $(gum style --foreground "$GRAY" "(binary)")"
  "/tmp/stg-*.log   $(gum style --foreground "$GRAY" "(temp logs)")"
)
$ALL && FILES+=(
  "~/.cache/go-build $(gum style --foreground "$YELLOW" "(module cache — slow to restore!)")"
  "./vendor/         $(gum style --foreground "$GRAY" "(vendor dir)")"
)

for f in "${FILES[@]}"; do
  gum style --foreground "$GRAY" "  · $f"
done

echo ""

if $DRY_RUN; then
  gum style \
    --border rounded --border-foreground "$YELLOW" --padding "0 2" \
    "$(gum style --foreground "$YELLOW" --bold "dry-run: no files were removed")"
  exit 0
fi

if $ALL && $IS_TTY; then
  echo ""
  gum style --foreground "$RED" --bold "  ⚠ This will clear the Go module cache."
  if ! gum confirm --affirmative "Yes, clear it" --negative "Cancel" "  Are you sure?"; then
    gum log --level warn "Aborted"
    exit 0
  fi
fi

echo ""
gum log --level info "Cleaning..."

spin() { local title="$1"; shift
  if $IS_TTY; then gum spin --spinner meter --title "  $title" -- "$@"
  else "$@"
  fi
}

spin "Removing ./build ..."       sleep 0.3
gum style --foreground "$TEAL" "  ✓ ./build"

spin "Removing binary ..."        sleep 0.2
gum style --foreground "$TEAL" "  ✓ ./stg"

spin "Removing temp files ..."    sleep 0.2
gum style --foreground "$TEAL" "  ✓ /tmp/stg-*.log"

if $ALL; then
  spin "Clearing module cache ..." sleep 1.2
  gum style --foreground "$TEAL" "  ✓ ~/.cache/go-build"

  spin "Removing vendor/ ..."      sleep 0.3
  gum style --foreground "$TEAL" "  ✓ ./vendor"
fi

echo ""
gum style \
  --border rounded --border-foreground "$TEAL" --padding "0 2" \
  "$(gum style --bold --foreground "$TEAL" "✓ Clean complete")"

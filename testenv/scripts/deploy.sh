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
    "$(gum style --bold --foreground "$PURPLE" "deploy — Deploy to an environment")

$(gum style --foreground "$GRAY" "Usage:") deploy [OPTIONS]

$(gum style --foreground "$TEAL" "--env ENV")     Target: staging | production (default: staging)
$(gum style --foreground "$TEAL" "--tag TAG")     Image tag to deploy            (default: latest)
$(gum style --foreground "$TEAL" "--dry-run")     Plan without executing
$(gum style --foreground "$TEAL" "--help")        Show this message"
}

if [[ "${1:-}" == "--help" ]]; then
  show_help
  exit 0
fi

IS_TTY=false
[ -t 1 ] && IS_TTY=true

ENV="staging"
TAG="latest"
DRY_RUN=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --env)     ENV="$2"; shift 2 ;;
    --tag)     TAG="$2"; shift 2 ;;
    --dry-run) DRY_RUN=true; shift ;;
    *)         gum log --level error "Unknown option: $1"; exit 1 ;;
  esac
done

if $IS_TTY; then
  gum style --foreground "$PURPLE" --bold "  Deploy"
  echo ""

  ENV=$(gum choose --header "  Target environment:" \
    --cursor.foreground "$PURPLE" \
    --selected.foreground "$TEAL" \
    "staging" "production")

  TAG=$(gum input \
    --placeholder "latest" \
    --prompt "  Image tag › " \
    --prompt.foreground "$PURPLE")
  TAG="${TAG:-latest}"
fi

echo ""
gum log --level info "env=$ENV tag=$TAG dry_run=$DRY_RUN"
echo ""

# Extra confirmation for production
if [[ "$ENV" == "production" ]]; then
  gum style \
    --border double --border-foreground "$RED" \
    --padding "1 3" \
    "$(gum style --bold --foreground "$RED" "⚠  PRODUCTION DEPLOYMENT")
$(gum style --foreground "$YELLOW" "  This will affect live users.")"
  echo ""

  if $IS_TTY; then
    CONFIRM=$(gum input \
      --placeholder "type \"deploy\" to confirm" \
      --prompt "  › " \
      --prompt.foreground "$RED")
    if [[ "$CONFIRM" != "deploy" ]]; then
      gum log --level warn "Aborted — confirmation did not match"
      exit 1
    fi
  fi
fi

spin() { local title="$1"; shift
  if $IS_TTY; then gum spin --spinner globe --title "  $title" -- "$@"
  else "$@"
  fi
}

step() {
  local icon="$1" msg="$2"
  gum style --foreground "$TEAL" "  $icon $msg"
}

DRY_PREFIX=""
$DRY_RUN && DRY_PREFIX="[dry-run] "

spin "${DRY_PREFIX}Pulling image app:$TAG ..."      sleep 0.7;  step "↓" "Image pulled: app:$TAG"
spin "${DRY_PREFIX}Running pre-deploy checks ..."   sleep 0.5;  step "✓" "Health checks passed"
spin "${DRY_PREFIX}Draining old instances ..."      sleep 0.6;  step "⇄" "Traffic shifted to canary"
spin "${DRY_PREFIX}Deploying new instances ..."     sleep 1.0;  step "↑" "3/3 instances updated"
spin "${DRY_PREFIX}Running smoke tests ..."         sleep 0.8;  step "✓" "Smoke tests passed"
spin "${DRY_PREFIX}Finalizing rollout ..."          sleep 0.4;  step "✓" "Rollout complete"

echo ""
ENV_COLOR="$TEAL"
[[ "$ENV" == "production" ]] && ENV_COLOR="$RED"

gum style \
  --border rounded \
  --border-foreground "$ENV_COLOR" \
  --padding "0 2" \
  "$(gum style --bold --foreground "$ENV_COLOR" "✓ Deployed") $(gum style --foreground "$GRAY" "app:$TAG → $ENV")$([ "$DRY_RUN" = true ] && echo " $(gum style --foreground "$YELLOW" "(dry-run)")" || echo "")"

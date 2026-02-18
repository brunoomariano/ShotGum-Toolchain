#!/usr/bin/env bash
# install.sh — Install or update stg (ShotGum script manager)
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/brunoomariano/ShotGum-Toolchain/main/install.sh | bash
#
# Environment overrides:
#   STG_VERSION      — pin a specific release, e.g. STG_VERSION=v0.2.0
#   STG_INSTALL_DIR  — install location  (default: ~/.local/bin)

set -euo pipefail

REPO="brunoomariano/ShotGum-Toolchain"
BINARY="stg"
INSTALL_DIR="${STG_INSTALL_DIR:-$HOME/.local/bin}"
GITHUB_API="https://api.github.com"
GITHUB_BASE="https://github.com"

# ── Output helpers ────────────────────────────────────────────────────────────

BOLD="\033[1m"; GREEN="\033[32m"; YELLOW="\033[33m"; RED="\033[31m"
GRAY="\033[90m"; RESET="\033[0m"

step() { printf "  ${BOLD}→${RESET}  %s\n"        "$*"; }
ok()   { printf "  ${GREEN}✓${RESET}  %s\n"       "$*"; }
warn() { printf "  ${YELLOW}⚠${RESET}  %s\n"      "$*"; }
die()  { printf "  ${RED}✗${RESET}  %s\n" "$*" >&2; exit 1; }
dim()  { printf "  ${GRAY}%s${RESET}\n"            "$*"; }

need() {
  command -v "$1" &>/dev/null || die "Required tool not found: $1"
}

# ── Platform detection ────────────────────────────────────────────────────────

detect_os() {
  case "$(uname -s)" in
    Linux)  echo "linux"  ;;
    Darwin) echo "darwin" ;;
    *)      die "Unsupported OS: $(uname -s). Install manually from $GITHUB_BASE/$REPO/releases" ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64)        echo "amd64" ;;
    aarch64|arm64) echo "arm64" ;;
    *)             die "Unsupported architecture: $(uname -m). Install manually from $GITHUB_BASE/$REPO/releases" ;;
  esac
}

# ── Version helpers ───────────────────────────────────────────────────────────

latest_version() {
  curl -fsSL "$GITHUB_API/repos/$REPO/releases/latest" \
    | grep '"tag_name"' \
    | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/'
}

current_version() {
  command -v "$BINARY" &>/dev/null \
    && "$BINARY" --version 2>/dev/null | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' | head -1 \
    || echo ""
}

# ── Install ───────────────────────────────────────────────────────────────────

main() {
  printf "\n${BOLD}  · ° o O  ShotGum (stg)  ¬══════►${RESET}\n"
  printf "  ${GRAY}$GITHUB_BASE/$REPO${RESET}\n\n"

  need curl
  need tar

  local os arch
  os=$(detect_os)
  arch=$(detect_arch)
  step "Platform: $os/$arch"

  # Resolve target version
  local version
  version="${STG_VERSION:-$(latest_version)}"
  [[ -z "$version" ]] && die "Could not fetch latest version. Check your internet connection."
  step "Version:  $version"

  # Skip if already up to date
  local current
  current=$(current_version)
  if [[ -n "$current" && "$current" == "$version" ]]; then
    ok "Already up to date ($version)"
    printf "\n"
    exit 0
  fi
  [[ -n "$current" ]] \
    && step "Updating: $current → $version" \
    || step "Installing $version..."

  # Download
  local archive="stg-${os}-${arch}.tar.gz"
  local url="$GITHUB_BASE/$REPO/releases/download/$version/$archive"
  local tmp
  tmp=$(mktemp -d)
  trap 'rm -rf "$tmp"' EXIT

  step "Downloading..."
  dim "$url"
  curl -fsSL --progress-bar "$url" -o "$tmp/$archive" \
    || die "Download failed. Does release $version exist? $GITHUB_BASE/$REPO/releases"

  # Verify checksum if sha256sum is available
  if command -v sha256sum &>/dev/null; then
    local checksum_url="$GITHUB_BASE/$REPO/releases/download/$version/checksums.txt"
    if curl -fsSL "$checksum_url" -o "$tmp/checksums.txt" 2>/dev/null; then
      (cd "$tmp" && grep "$archive" checksums.txt | sha256sum --check --status) \
        && ok "Checksum verified" \
        || { warn "Checksum mismatch — aborting"; exit 1; }
    fi
  fi

  step "Extracting..."
  tar -xzf "$tmp/$archive" -C "$tmp"

  # Install binary
  mkdir -p "$INSTALL_DIR"
  install -m755 "$tmp/$BINARY" "$INSTALL_DIR/$BINARY"
  ok "Installed → $INSTALL_DIR/$BINARY"

  # First-time setup hint
  if [[ -z "$current" ]]; then
    printf "\n"
    step "Run once to finish setup:"
    dim "  stg init"
  fi

  # PATH warning
  if ! echo ":${PATH}:" | grep -q ":${INSTALL_DIR}:"; then
    printf "\n"
    warn "$INSTALL_DIR is not in your PATH. Add to your shell profile:"
    printf "\n"
    printf "    ${GRAY}export PATH=\"\$HOME/.local/bin:\$PATH\"${RESET}\n"
    printf "\n"
    dim "Then reload: source ~/.bashrc  (or ~/.zshrc)"
  else
    printf "\n"
    ok "Done — run: stg"
  fi

  printf "\n"
}

main "$@"

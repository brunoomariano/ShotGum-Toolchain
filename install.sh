#!/usr/bin/env bash
# install.sh — Install or update stg (ShotGum script manager) from source
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/brunoomariano/ShotGum-Toolchain/main/install.sh | bash
#
# Environment overrides:
#   STG_REF          — source ref to install (default: main). Accepts tag or branch name.
#   STG_VERSION      — backward-compatible alias for STG_REF
#   STG_INSTALL_DIR  — install location (default: ~/.local/bin)

set -euo pipefail

REPO="brunoomariano/ShotGum-Toolchain"  # keep in sync with internal/version/version.go → Repo
BINARY="stg"
INSTALL_DIR="${STG_INSTALL_DIR:-$HOME/.local/bin}"
REF="${STG_REF:-${STG_VERSION:-main}}"
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

# ── Interactive prompt (works with curl | bash) ─────────────────────────────

ask_tty() {
  local prompt="$1"
  [[ ! -e /dev/tty ]] && return 1
  local answer=""
  printf "\n  %s [y/N]: " "$prompt" >/dev/tty
  read -r answer </dev/tty 2>/dev/null || return 1
  [[ "${answer,,}" == "y" ]]
}

# ── Source download/build helpers ─────────────────────────────────────────────

TMP_ROOT=""
SOURCE_DIR=""
RESOLVED_REF_KIND=""

cleanup() {
  [[ -n "${TMP_ROOT:-}" && -d "$TMP_ROOT" ]] && rm -rf "$TMP_ROOT"
}

download_source() {
  local ref="$1"
  local archive="$TMP_ROOT/repo.tar.gz"

  local url_tag="$GITHUB_BASE/$REPO/archive/refs/tags/${ref}.tar.gz"
  local url_branch="$GITHUB_BASE/$REPO/archive/refs/heads/${ref}.tar.gz"

  step "Downloading source ($ref)..."

  if curl -fsSL "$url_tag" -o "$archive" 2>/dev/null; then
    RESOLVED_REF_KIND="tag"
    ok "Using tag: $ref"
  elif curl -fsSL "$url_branch" -o "$archive" 2>/dev/null; then
    RESOLVED_REF_KIND="branch"
    ok "Using branch: $ref"
  else
    die "Could not download ref '$ref' as tag or branch from $GITHUB_BASE/$REPO"
  fi

  tar -xzf "$archive" -C "$TMP_ROOT"

  SOURCE_DIR="$(find "$TMP_ROOT" -mindepth 1 -maxdepth 1 -type d | head -1)"
  [[ -z "$SOURCE_DIR" || ! -d "$SOURCE_DIR" ]] && die "Could not locate extracted source directory"

  ok "Source ready → $SOURCE_DIR"
}

build_and_install() {
  local ref="$1"
  local build_version="$ref"

  step "Building from source with make build..."
  (
    cd "$SOURCE_DIR"
    VERSION="$build_version" make build
  )

  [[ -x "$SOURCE_DIR/$BINARY" ]] || die "Build finished but binary '$BINARY' was not produced"

  mkdir -p "$INSTALL_DIR"
  install -m755 "$SOURCE_DIR/$BINARY" "$INSTALL_DIR/$BINARY"
  ok "Installed → $INSTALL_DIR/$BINARY"
}

# ── Optional extras from downloaded source ────────────────────────────────────

install_playground_from_source() {
  local dest="$HOME/shotgum-playground"
  local src="$SOURCE_DIR/shotgum-playground"

  [[ -d "$src" ]] || { warn "shotgum-playground not found in downloaded source."; return 1; }

  if [[ -d "$dest" ]]; then
    warn "Directory already exists: $dest — skipping copy."
    return 0
  fi

  cp -R "$src" "$dest"

  ok "Playground ready → $dest"
  printf "\n"
  dim "  Explore it:"
  dim "    cd ~/shotgum-playground"
  dim "    stg"
}

install_defaults_from_source() {
  local scripts_dir="$HOME/.shotgum/scripts/shotgum"
  local src_dir="$SOURCE_DIR/defaults/scripts"
  local stg_bin="$INSTALL_DIR/$BINARY"

  step "Installing default project scripts..."

  [[ -d "$src_dir" ]] || { warn "defaults/scripts not found in downloaded source."; return 1; }

  if [[ ! -x "$stg_bin" ]]; then
    stg_bin="$(command -v "$BINARY" || true)"
  fi

  mkdir -p "$scripts_dir"

  local ok_count=0
  for name in star issue; do
    local src="$src_dir/${name}.sh"
    local dest="$scripts_dir/${name}.sh"
    if [[ -f "$src" ]]; then
      cp "$src" "$dest"
      chmod +x "$dest"
      ok "  Installed: $dest"
      ok_count=$(( ok_count + 1 ))
    else
      warn "  Missing ${name}.sh in source — skipping."
    fi
  done

  [[ "$ok_count" -eq 0 ]] && { warn "No scripts installed. Aborting registration."; return 1; }

  if [[ -z "${stg_bin:-}" || ! -x "$stg_bin" ]]; then
    warn "Could not find stg binary to register scripts automatically."
    warn "Run 'stg init' and add scripts from $scripts_dir manually."
    return 0
  fi

  "$stg_bin" init 2>/dev/null || true

  "$stg_bin" add folder "$scripts_dir" \
    --name shotgum \
    --desc "ShotGum community scripts (star, issue)" \
    2>/dev/null || true

  [[ -f "$scripts_dir/star.sh" ]] && \
    "$stg_bin" add script "$scripts_dir/star.sh" \
      --category shotgum \
      --name star \
      --desc "Star ShotGum on GitHub" \
      --executable bash \
      2>/dev/null || true

  [[ -f "$scripts_dir/issue.sh" ]] && \
    "$stg_bin" add script "$scripts_dir/issue.sh" \
      --category shotgum \
      --name issue \
      --desc "Open a well-documented issue on GitHub" \
      --executable bash \
      2>/dev/null || true

  ok "Default project scripts registered in global config"
  printf "\n"
  dim "  Open stg and look for the 'shotgum' category."
  dim "  Or run directly: stg shotgum star"
}

# ── Install ───────────────────────────────────────────────────────────────────

main() {
  printf "\n${BOLD}  · ° o O  ShotGum (stg)  ¬══════►${RESET}\n"
  printf "  ${GRAY}$GITHUB_BASE/$REPO${RESET}\n\n"

  need curl
  need tar
  need make
  need go

  step "Source ref: $REF"

  TMP_ROOT="$(mktemp -d)"
  trap cleanup EXIT

  download_source "$REF"
  build_and_install "$REF"
  step "Installed ref: ${REF} (${RESOLVED_REF_KIND})"

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
  printf "  ${BOLD}Optional extras${RESET}\n"

  if ask_tty "Copy shotgum-playground from source to ~/shotgum-playground and test stg?"; then
    printf "\n"
    install_playground_from_source
  fi

  if ask_tty "Install default project scripts in your user folder? (shotgum/star + shotgum/issue)"; then
    printf "\n"
    install_defaults_from_source
  fi

  printf "\n"
}

main "$@"

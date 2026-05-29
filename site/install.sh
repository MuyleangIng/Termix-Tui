#!/usr/bin/env bash
# Public URL:
# curl -fsSL https://muyleanging.github.io/Termix-Tui/install.sh | bash

set -euo pipefail

REPO_OWNER="MuyleangIng"
REPO_NAME="Termix-Tui"
VERSION="${TERMIX_VERSION:-latest}"
INSTALL_DIR="${TERMIX_INSTALL_DIR:-$HOME/.local/bin}"
TEMP_DIR=""
YES=0

cleanup() {
  if [ -n "${TEMP_DIR:-}" ] && [ -d "$TEMP_DIR" ]; then
    rm -rf "$TEMP_DIR"
  fi
}

trap cleanup EXIT

info() {
  printf '\033[36m[TERMIX]\033[0m %s\n' "$1"
}

success() {
  printf '\033[32m[SUCCESS]\033[0m %s\n' "$1"
}

warn() {
  printf '\033[33m[WARN]\033[0m %s\n' "$1"
}

error() {
  printf '\033[31m[ERROR]\033[0m %s\n' "$1"
}

usage() {
  cat <<EOF
Usage: install.sh [--version VERSION] [--install-dir DIR] [--yes]

Environment:
  TERMIX_VERSION      Release tag to install, default: latest
  TERMIX_INSTALL_DIR  Install directory, default: \$HOME/.local/bin
EOF
}

parse_args() {
  while [ "$#" -gt 0 ]; do
    case "$1" in
      --version|-v)
        if [ "$#" -lt 2 ]; then
          error "--version requires a value"
          exit 1
        fi
        VERSION="$2"
        shift 2
        ;;
      --install-dir|-d)
        if [ "$#" -lt 2 ]; then
          error "--install-dir requires a value"
          exit 1
        fi
        INSTALL_DIR="$2"
        shift 2
        ;;
      --yes|-y)
        YES=1
        shift
        ;;
      --help|-h)
        usage
        exit 0
        ;;
      *)
        error "Unknown option: $1"
        usage
        exit 1
        ;;
    esac
  done
}

confirm_existing_install_update() {
  local path="$1"
  local version_label="$2"
  local answer=""

  if [ "$YES" = "1" ]; then
    return 0
  fi

  warn "Termix is already installed at $path"

  if [ -r /dev/tty ]; then
    printf "Update Termix to %s? [Y/n] " "$version_label" > /dev/tty
    IFS= read -r answer < /dev/tty || answer=""
  elif [ -t 0 ]; then
    printf "Update Termix to %s? [Y/n] " "$version_label"
    IFS= read -r answer || answer=""
  else
    warn "No interactive terminal available for update confirmation; continuing. Use --yes to suppress this message."
    return 0
  fi

  case "$(printf '%s' "$answer" | tr '[:upper:]' '[:lower:]')" in
    n|no)
      info "Update cancelled. Existing Termix was not changed."
      exit 0
      ;;
  esac
}

detect_os() {
  case "$(uname -s)" in
    Linux*) echo "Linux" ;;
    Darwin*) echo "Darwin" ;;
    *) error "Unsupported OS: $(uname -s)"; exit 1 ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64) echo "x86_64" ;;
    arm64|aarch64) echo "arm64" ;;
    *) error "Unsupported architecture: $(uname -m)"; exit 1 ;;
  esac
}

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    error "Missing required command: $1"
    exit 1
  fi
}

download_file() {
  local url="$1"
  local output="$2"

  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url" -o "$output"
  elif command -v wget >/dev/null 2>&1; then
    wget -q "$url" -O "$output"
  else
    error "curl or wget is required."
    exit 1
  fi
}

get_release_json() {
  local url
  if [ "$VERSION" = "latest" ]; then
    url="https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases/latest"
  else
    url="https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases/tags/$VERSION"
  fi

  if command -v curl >/dev/null 2>&1; then
    curl -fsSL -H "User-Agent: TermixInstaller" "$url"
  else
    wget -q --header="User-Agent: TermixInstaller" -O - "$url"
  fi
}

extract_asset_url() {
  local asset_name="$1"
  local release_file="$2"

  python3 - "$asset_name" "$release_file" <<'PY'
import json
import sys

asset_name = sys.argv[1]
release_file = sys.argv[2]

with open(release_file, "r", encoding="utf-8") as handle:
    data = json.load(handle)

for asset in data.get("assets", []):
    if asset.get("name") == asset_name:
        print(asset.get("browser_download_url"))
        sys.exit(0)

sys.exit(1)
PY
}

extract_release_tag() {
  local release_file="$1"

  python3 - "$release_file" <<'PY'
import json
import sys

with open(sys.argv[1], "r", encoding="utf-8") as handle:
    data = json.load(handle)

print(data.get("tag_name", "latest"))
PY
}

install_termix() {
  need_cmd python3
  need_cmd tar

  local os
  os="$(detect_os)"

  local arch
  arch="$(detect_arch)"

  local asset_name="termix_${os}_${arch}.tar.gz"

  info "Selected platform: $os $arch"
  info "Looking for asset: $asset_name"

  TEMP_DIR="$(mktemp -d)"

  local release_json="$TEMP_DIR/release.json"
  get_release_json > "$release_json"

  local release_tag
  release_tag="$(extract_release_tag "$release_json")"
  info "Selected release: $release_tag"

  local asset_url
  asset_url="$(extract_asset_url "$asset_name" "$release_json" || true)"

  if [ -z "$asset_url" ]; then
    error "Could not find release asset: $asset_name"
    info "Check GitHub Releases:"
    info "https://github.com/$REPO_OWNER/$REPO_NAME/releases"
    exit 1
  fi

  local archive="$TEMP_DIR/$asset_name"
  local version_label="$release_tag"

  mkdir -p "$INSTALL_DIR"

  if [ -f "$INSTALL_DIR/termix" ]; then
    confirm_existing_install_update "$INSTALL_DIR/termix" "$version_label"
  fi

  info "Downloading Termix..."
  download_file "$asset_url" "$archive"

  info "Extracting..."
  tar -xzf "$archive" -C "$TEMP_DIR"

  local binary
  binary="$(find "$TEMP_DIR" -type f -name termix | head -n 1)"

  if [ -z "$binary" ]; then
    error "termix binary was not found in archive."
    exit 1
  fi

  if [ -f "$INSTALL_DIR/termix" ]; then
    cp "$INSTALL_DIR/termix" "$INSTALL_DIR/termix.bak-$(date +%Y%m%d-%H%M%S)"
    info "Backed up existing termix binary."
  fi

  cp "$binary" "$INSTALL_DIR/termix"
  cp "$binary" "$INSTALL_DIR/termix-tui"
  chmod +x "$INSTALL_DIR/termix"
  chmod +x "$INSTALL_DIR/termix-tui"

  success "Installed Termix to $INSTALL_DIR/termix"
  success "Installed TUI launcher to $INSTALL_DIR/termix-tui"

  if ! echo "$PATH" | tr ':' '\n' | grep -qx "$INSTALL_DIR"; then
    warn "$INSTALL_DIR is not in PATH."
    info "Add this to your shell config:"
    echo "export PATH=\"\$PATH:$INSTALL_DIR\""
  fi

  "$INSTALL_DIR/termix" --version || {
    error "Termix installed, but version check failed."
    exit 1
  }

  success "Termix installed successfully."

  info "Bootstrapping Oh My Posh, Meslo font, official themes, and terminal config..."
  bootstrap_step 25 "Oh My Posh" "$INSTALL_DIR/termix" install oh-my-posh || true
  bootstrap_step 50 "MesloLGM Nerd Font" "$INSTALL_DIR/termix" fonts install "MesloLGM Nerd Font" --yes || true
  bootstrap_step 75 "Official themes" "$INSTALL_DIR/termix" install themes || warn "Retry themes later with: termix install themes"
  bootstrap_step 100 "Terminal and VS Code font config" "$INSTALL_DIR/termix" fonts apply "MesloLGM Nerd Font" || true

  echo ""
  success "Done."
  info "Next steps:"
  echo "  1. Close and reopen your terminal so the Meslo font/profile settings load."
  echo "  2. Run: termix setup"
  echo "  3. Open the dashboard with: termix-tui"
}

bootstrap_step() {
  local percent="$1"
  local name="$2"
  shift 2
  info "[$percent%] $name..."
  if "$@"; then
    success "[$percent%] $name ready."
    return 0
  fi
  warn "[$percent%] $name failed. Continuing bootstrap."
  return 1
}

parse_args "$@"
install_termix

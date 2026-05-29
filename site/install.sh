#!/usr/bin/env bash
# Public URL:
# curl -fsSL https://muyleanging.github.io/Termix-Tui/install.sh | bash

set -euo pipefail

REPO_OWNER="MuyleangIng"
REPO_NAME="Termix-Tui"
VERSION="${TERMIX_VERSION:-latest}"
INSTALL_DIR="${TERMIX_INSTALL_DIR:-$HOME/.local/bin}"
TEMP_DIR=""

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
Usage: install.sh [--version VERSION] [--install-dir DIR]

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

  local asset_url
  asset_url="$(extract_asset_url "$asset_name" "$release_json" || true)"

  if [ -z "$asset_url" ]; then
    error "Could not find release asset: $asset_name"
    info "Check GitHub Releases:"
    info "https://github.com/$REPO_OWNER/$REPO_NAME/releases"
    exit 1
  fi

  local archive="$TEMP_DIR/$asset_name"

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

  mkdir -p "$INSTALL_DIR"

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

  info "Bootstrapping Oh My Posh, Meslo font, and official themes..."
  if "$INSTALL_DIR/termix" install; then
    success "Default tools, Meslo font, and official themes are ready."
  else
    warn "Bootstrap did not complete. You can retry later with: termix install"
  fi

  echo ""
  success "Done."
  info "Next steps:"
  echo "  1. Open a new terminal."
  echo "  2. Run: termix setup"
  echo "  3. Open the dashboard with: termix-tui"
}

parse_args "$@"
install_termix

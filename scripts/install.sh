#!/usr/bin/env bash
# Public URL:
# curl -fsSL https://muyleanging.github.io/Termix-Tui/install.sh | bash

set -euo pipefail

REPO_OWNER="muyleanging"
REPO_NAME="Termix-Tui"
VERSION="${TERMIX_VERSION:-latest}"
INSTALL_DIR="${TERMIX_INSTALL_DIR:-$HOME/.local/bin}"
NO_SETUP="${TERMIX_NO_SETUP:-0}"

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

  local temp_dir
  temp_dir="$(mktemp -d)"

  trap 'rm -rf "$temp_dir"' EXIT

  local release_json="$temp_dir/release.json"
  get_release_json > "$release_json"

  local asset_url
  asset_url="$(extract_asset_url "$asset_name" "$release_json" || true)"

  if [ -z "$asset_url" ]; then
    error "Could not find release asset: $asset_name"
    info "Check GitHub Releases:"
    info "https://github.com/$REPO_OWNER/$REPO_NAME/releases"
    exit 1
  fi

  local archive="$temp_dir/$asset_name"

  info "Downloading Termix..."
  download_file "$asset_url" "$archive"

  info "Extracting..."
  tar -xzf "$archive" -C "$temp_dir"

  local binary
  binary="$(find "$temp_dir" -type f -name termix | head -n 1)"

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
  chmod +x "$INSTALL_DIR/termix"

  success "Installed Termix to $INSTALL_DIR/termix"

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

  if [ "$NO_SETUP" != "1" ]; then
    printf "Run termix setup now? [Y/n] "
    read -r answer
    case "$answer" in
      ""|y|Y|yes|YES)
        "$INSTALL_DIR/termix" setup
        ;;
      *)
        info "You can run setup later with: termix setup"
        ;;
    esac
  fi
}

install_termix

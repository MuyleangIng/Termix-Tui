#!/usr/bin/env bash
# Public URL:
# curl -fsSL https://muyleanging.github.io/Termix-Tui/uninstall.sh | bash

set -euo pipefail

INSTALL_DIR="${TERMIX_INSTALL_DIR:-$HOME/.local/bin}"
CLEAN_DATA="${TERMIX_CLEAN_DATA:-0}"

info() {
  printf '\033[36m[TERMIX]\033[0m %s\n' "$1"
}

success() {
  printf '\033[32m[SUCCESS]\033[0m %s\n' "$1"
}

warn() {
  printf '\033[33m[WARN]\033[0m %s\n' "$1"
}

warn "This removes the Termix executable."
warn "To remove shell profile integration safely, run this before deleting Termix:"
echo "  termix uninstall"
echo ""

printf "Continue uninstalling Termix executable? [y/N] "
read -r answer

case "$answer" in
  y|Y|yes|YES)
    ;;
  *)
    info "Cancelled."
    exit 0
    ;;
esac

if [ -f "$INSTALL_DIR/termix" ]; then
  rm -f "$INSTALL_DIR/termix"
  success "Removed $INSTALL_DIR/termix"
else
  info "Termix binary not found at $INSTALL_DIR/termix"
fi

if [ -f "$INSTALL_DIR/termix-tui" ]; then
  rm -f "$INSTALL_DIR/termix-tui"
  success "Removed $INSTALL_DIR/termix-tui"
fi

DATA_DIR="$HOME/.termix"

if [ "$CLEAN_DATA" = "1" ]; then
  if [ -d "$DATA_DIR" ]; then
    printf "Remove Termix data folder %s ? This removes config/cache/themes. [y/N] " "$DATA_DIR"
    read -r data_answer
    case "$data_answer" in
      y|Y|yes|YES)
        rm -rf "$DATA_DIR"
        success "Removed Termix data folder."
        ;;
      *)
        info "Kept Termix data folder."
        ;;
    esac
  fi
else
  info "Termix data folder was kept: $DATA_DIR"
  info "Run with TERMIX_CLEAN_DATA=1 if you want to remove config/cache/themes."
fi

success "Uninstall complete."

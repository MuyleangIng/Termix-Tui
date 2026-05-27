#!/usr/bin/env sh
set -eu

ROOT="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
cd "$ROOT"

if ! command -v go >/dev/null 2>&1; then
  echo "Go is required. Install it from https://go.dev/dl/ or your package manager." >&2
  exit 1
fi

echo "[TERMIX] Resolving Go modules"
go mod tidy

echo "[TERMIX] Building termix"
mkdir -p "$ROOT/bin"
go build -o "$ROOT/bin/termix" .

echo "[TERMIX] Built $ROOT/bin/termix"

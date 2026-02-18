#!/usr/bin/env bash
set -euo pipefail

REPO_URL="${REPO_URL:-https://github.com/t4Linux/t4gfm.git}"
BRANCH="${BRANCH:-main}"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
BINARY_NAME="${BINARY_NAME:-gfm}"

if ! command -v git >/dev/null 2>&1; then
  printf "Error: git is required but not installed.\n" >&2
  exit 1
fi

if ! command -v go >/dev/null 2>&1; then
  printf "Error: Go is required but not installed.\n" >&2
  exit 1
fi

tmp_dir="$(mktemp -d)"
cleanup() {
  rm -rf "$tmp_dir"
}
trap cleanup EXIT

printf "Cloning %s (branch: %s)...\n" "$REPO_URL" "$BRANCH"
git clone --depth 1 --branch "$BRANCH" "$REPO_URL" "$tmp_dir/t4gfm"

printf "Building %s...\n" "$BINARY_NAME"
mkdir -p "$INSTALL_DIR"

(
  cd "$tmp_dir/t4gfm"
  go build -o "$INSTALL_DIR/$BINARY_NAME" .
)

printf "Installed: %s\n" "$INSTALL_DIR/$BINARY_NAME"

case ":$PATH:" in
  *":$INSTALL_DIR:"*)
    printf "You can run it now with: %s\n" "$BINARY_NAME"
    ;;
  *)
    printf "Note: %s is not in PATH. Run with full path or add it to your shell config.\n" "$INSTALL_DIR"
    ;;
esac

printf "\nRecommended first run:\n"
printf "%s --fix-hotkeys\n" "$INSTALL_DIR/$BINARY_NAME"

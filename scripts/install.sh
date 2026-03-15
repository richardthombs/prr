#!/usr/bin/env bash
# Install prr on macOS or Linux.
# Usage: curl -fsSL https://raw.githubusercontent.com/richardthombs/prr/main/scripts/install.sh | bash
set -euo pipefail

REPO="richardthombs/prr"
BINARY="prr"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

# Detect OS and architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH" >&2
    exit 1
    ;;
esac

case "$OS" in
  linux|darwin) ;;
  *)
    echo "Unsupported OS: $OS" >&2
    echo "For Windows, use: irm https://raw.githubusercontent.com/${REPO}/main/scripts/install.ps1 | iex" >&2
    exit 1
    ;;
esac

# Resolve latest version
VERSION="${VERSION:-}"
if [[ -z "$VERSION" ]]; then
  VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')"
fi

if [[ -z "$VERSION" ]]; then
  echo "Failed to resolve latest release version" >&2
  exit 1
fi

ARCHIVE="${BINARY}_${VERSION}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE}"

echo "Installing prr ${VERSION} (${OS}/${ARCH}) to ${INSTALL_DIR}..."

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

curl -fsSL "$URL" -o "$TMP/$ARCHIVE"
tar -xzf "$TMP/$ARCHIVE" -C "$TMP"

mkdir -p "$INSTALL_DIR"
install -m 755 "$TMP/$BINARY" "$INSTALL_DIR/$BINARY"

echo "Installed: ${INSTALL_DIR}/${BINARY}"

if ! command -v "$BINARY" >/dev/null 2>&1; then
  echo ""
  echo "NOTE: ${INSTALL_DIR} is not in your PATH."
  echo "Add it with: export PATH=\"\$PATH:${INSTALL_DIR}\""
fi

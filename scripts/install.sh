#!/usr/bin/env bash
set -euo pipefail

REPO="Jacknic/gradlex"
INSTALL_DIR="/usr/local/bin"
TMP_DIR="$(mktemp -d)"
cleanup() { rm -rf "$TMP_DIR"; }
trap cleanup EXIT

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$OS" in
  linux*) platform="linux" ;;
  darwin*) platform="darwin" ;;
  *) echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac
case "$ARCH" in
  x86_64|amd64) arch="amd64" ;;
  aarch64|arm64) arch="arm64" ;;
  *) echo "Unsupported arch: $ARCH" >&2; exit 1 ;;
esac

if [ "$platform" = "linux" ]; then
  pattern="linux-${arch}"
else
  pattern="darwin-${arch}"
fi

echo "Detected $platform/$arch — searching release asset matching '*${pattern}*'..."

url=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" \
  | grep "browser_download_url" | grep "${pattern}" | head -n1 | cut -d '"' -f4 || true)

if [ -z "$url" ]; then
  echo "No release asset found for pattern: ${pattern}" >&2
  exit 1
fi

echo "Found asset: $url"
filename="$TMP_DIR/$(basename "$url")"

curl -L "$url" -o "$filename"

case "$filename" in
  *.tar.gz|*.tgz)
    tar -xzf "$filename" -C "$TMP_DIR"
    ;;
  *.zip)
    if command -v unzip >/dev/null 2>&1; then
      unzip -q "$filename" -d "$TMP_DIR"
    else
      echo "unzip not installed" >&2; exit 1
    fi
    ;;
  *)
    echo "Unknown archive format: $filename" >&2
    ;;
esac

# find gradlex binary
if [ -f "$TMP_DIR/gradlex" ]; then
  BIN="$TMP_DIR/gradlex"
else
  BIN="$(find "$TMP_DIR" -maxdepth 3 -type f -name gradlex -print -quit 2>/dev/null || true)"
fi

if [ -z "$BIN" ]; then
  echo "gradlex binary not found in archive" >&2
  exit 1
fi

echo "Installing gradlex from $BIN to $INSTALL_DIR (may ask for sudo)..."

if [ -w "$INSTALL_DIR" ]; then
  cp "$BIN" "$INSTALL_DIR/gradlex"
  chmod +x "$INSTALL_DIR/gradlex"
else
  sudo cp "$BIN" "$INSTALL_DIR/gradlex"
  sudo chmod +x "$INSTALL_DIR/gradlex"
fi

echo "Installed gradlex to $INSTALL_DIR/gradlex"
exit 0

#!/bin/bash
set -e

REPO="zulfikawr/fm"
LATEST_URL="https://api.github.com/repos/$REPO/releases/latest"

echo "Detecting architecture..."
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
    Linux)     OS="linux" ;;
    Darwin)    OS="darwin" ;;
    *)         echo "Unsupported OS: $OS"; exit 1 ;;
esac

case "$ARCH" in
    x86_64)  ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    arm64)   ARCH="arm64" ;;
    *)       echo "Unsupported Architecture: $ARCH"; exit 1 ;;
esac

echo "Fetching latest release for $OS/$ARCH..."
DOWNLOAD_URL=$(curl -s $LATEST_URL | grep "browser_download_url" | grep "$OS-$ARCH" | cut -d '"' -f 4)

if [ -z "$DOWNLOAD_URL" ]; then
    echo "Error: Could not find a release for $OS/$ARCH"
    exit 1
fi

echo "Downloading $DOWNLOAD_URL..."
curl -L -o fm "$DOWNLOAD_URL"
chmod +x fm

echo "Installing to /usr/local/bin (requires sudo)..."
sudo mv fm /usr/local/bin/fm

echo "Successfully installed fm!"
fm --version

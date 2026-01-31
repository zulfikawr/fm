#!/bin/bash
set -e

VERSION=$1
ARCH=$2
BINARY_PATH=$3

if [[ -z "$VERSION" || -z "$ARCH" || -z "$BINARY_PATH" ]]; then
  echo "Usage: $0 <version> <arch> <binary_path>"
  exit 1
fi

# Sanitize version (remove v prefix)
VERSION_NUM=${VERSION#v}

# Map Go arch to Debian arch
DEB_ARCH=$ARCH
if [ "$ARCH" = "amd64" ]; then
  DEB_ARCH="amd64"
elif [ "$ARCH" = "arm64" ]; then
  DEB_ARCH="arm64"
else
  echo "Unsupported architecture for deb: $ARCH"
  exit 0
fi

WORKDIR="deb-build"
rm -rf "$WORKDIR"
mkdir -p "$WORKDIR/DEBIAN"
mkdir -p "$WORKDIR/usr/local/bin"

# Copy binary
cp "$BINARY_PATH" "$WORKDIR/usr/local/bin/fm"
chmod 755 "$WORKDIR/usr/local/bin/fm"

# Create control file
cat > "$WORKDIR/DEBIAN/control" <<EOF
Package: fm
Version: $VERSION_NUM
Section: utils
Priority: optional
Architecture: $DEB_ARCH
Maintainer: Zulfikar <zulfikawr@gmail.com>
Description: A fast, modular, and feature-rich TUI file manager.
 fm is a terminal file manager written in Go, featuring fuzzy search,
 remote access, git integration, and more.
EOF

# Build package
DEB_NAME="fm_${VERSION_NUM}_linux_${DEB_ARCH}.deb"
dpkg-deb --build "$WORKDIR" "$DEB_NAME"

echo "Created $DEB_NAME"

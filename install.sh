#!/bin/bash
# Looty Install Script for macOS/Linux
#
# Run with:
#   curl -sL https://raw.githubusercontent.com/YOUR_GITHUB_USER/BlipSync/main/install.sh | bash
#
# Or download and run manually:
#   chmod +x install.sh && ./install.sh

set -e

# Configuration - UPDATE THIS FOR YOUR REPO
REPO="YOUR_GITHUB_USER/BlipSync"
VERSION="latest"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="x86_64" ;;
    arm64|aarch64) ARCH="arm64" ;;
esac

# Map uname to asset name
case $OS in
    darwin) OS="macos" ;;
    linux)  OS="linux" ;;
esac

ASSET_NAME="looty-${OS}-${ARCH}"

echo "========================================"
echo "         LOOTY INSTALLER"
echo "========================================"
echo ""
echo "Detected: $OS ($ARCH)"
echo ""

# Get latest release info
if [ "$VERSION" = "latest" ]; then
    API_URL="https://api.github.com/repos/$REPO/releases/latest"
else
    API_URL="https://api.github.com/repos/$REPO/releases/tags/$VERSION"
fi

echo "Finding release..."

# Use GitHub API to get download URL
DOWNLOAD_URL=$(curl -sL "$API_URL" | grep -o "\"browser_download_url\": \"[^\"]*${ASSET_NAME}[^\"]*\"" | head -1 | cut -d'"' -f4)

if [ -z "$DOWNLOAD_URL" ]; then
    echo "Error: Could not find $ASSET_NAME in release"
    echo "Available assets:"
    curl -sL "$API_URL" | grep -o '"name": "[^"]*"' | head -20
    exit 1
fi

echo "Downloading $ASSET_NAME..."

# Create temp directory
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

# Download
curl -L -o "$TEMP_DIR/looty" "$DOWNLOAD_URL"

# Install to ~/.local/bin
INSTALL_DIR="$HOME/.local/bin"
mkdir -p "$INSTALL_DIR"
cp "$TEMP_DIR/looty" "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/looty"

# Add to PATH if needed
SHELL_RC="$HOME/.bashrc"
if [ "$OS" = "macos" ]; then
    SHELL_RC="$HOME/.zshrc"
fi

PATH_LINE="export PATH=\"\$PATH:$INSTALL_DIR\""
if ! grep -qF "$INSTALL_DIR" "$SHELL_RC" 2>/dev/null; then
    echo "" >> "$SHELL_RC"
    echo "# Looty" >> "$SHELL_RC"
    echo "$PATH_LINE" >> "$SHELL_RC"
    echo "Added to PATH in $SHELL_RC"
    echo "Restart your terminal or run: source $SHELL_RC"
else
    echo "Already in PATH"
fi

echo ""
echo "========================================"
echo "  SUCCESS! Looty installed!"
echo "========================================"
echo ""
echo "Installed to: $INSTALL_DIR"
echo ""
echo "Now open a NEW terminal and run:"
echo "  looty"
echo ""
echo "This will serve the CURRENT FOLDER on your network!"
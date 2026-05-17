#!/usr/bin/env bash
# Looty Latest Release Downloader
# Usage: curl -sL https://raw.githubusercontent.com/lirrensi/looty/main/scripts/get-latest.sh | sh

set -e

REPO="lirrensi/looty"
API_URL="https://api.github.com/repos/${REPO}/releases/latest"

# Detect platform
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$OS" in
    linux)
        PLATFORM="linux"
        ;;
    darwin)
        PLATFORM="macos"
        ;;
    msys*|mingw*|cygwin*|windows*)
        PLATFORM="windows"
        ;;
    *)
        echo "Unsupported OS: $OS"
        exit 1
        ;;
esac

case "$ARCH" in
    x86_64|amd64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

if [ "$PLATFORM" = "windows" ]; then
    SUFFIX="${PLATFORM}-${ARCH}"
    EXT="zip"
    BINARY="looty.exe"
else
    SUFFIX="${PLATFORM}-${ARCH}"
    EXT="tar.gz"
    BINARY="looty"
fi

ASSET_NAME="looty-${SUFFIX}.${EXT}"

echo "📦 Detected platform: ${PLATFORM}-${ARCH}"
echo "🐱 Fetching latest Looty release..."

# Get download URL from GitHub API
DOWNLOAD_URL=$(curl -sL "$API_URL" | grep -o '"browser_download_url": "[^"]*' | grep "$ASSET_NAME" | cut -d'"' -f4)

if [ -z "$DOWNLOAD_URL" ]; then
    echo "❌ Could not find asset: $ASSET_NAME"
    echo "Available assets:"
    curl -sL "$API_URL" | grep -o '"browser_download_url": "[^"]*' | cut -d'"' -f4 | grep looty- || true
    exit 1
fi

echo "⬇️  Downloading: $ASSET_NAME"
curl -sL -o "/tmp/${ASSET_NAME}" "$DOWNLOAD_URL"

# Extract
if [ "$EXT" = "zip" ]; then
    unzip -q "/tmp/${ASSET_NAME}" -d /tmp/
else
    tar -xzf "/tmp/${ASSET_NAME}" -C /tmp/
fi

# Install to /usr/local/bin or current directory
echo "🔧 Installing looty..."
if [ -w /usr/local/bin ]; then
    mv "/tmp/${BINARY}" /usr/local/bin/looty
    chmod +x /usr/local/bin/looty
    echo "✅ Installed to /usr/local/bin/looty"
else
    mv "/tmp/${BINARY}" ./looty
    chmod +x ./looty
    echo "✅ Installed to ./looty"
    echo "💡 Run: sudo mv looty /usr/local/bin/looty"
fi

# Cleanup
rm -f "/tmp/${ASSET_NAME}"

echo "🎉 Looty is ready! Run: looty --help"

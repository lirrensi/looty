#!/bin/sh
set -e

ARCH="$(uname -m | sed -e 's/x86_64/amd64/' -e 's/aarch64/arm64/')"
OS="$(uname | sed -e 's/Darwin/macos/' -e 's/Linux/linux/')"

curl -sL "https://github.com/lirrensi/looty/releases/latest/download/looty-${OS}-${ARCH}.tar.gz" | tar xz

mkdir -p ~/.local/bin
mv looty ~/.local/bin/

# Create ~/looty directory for easy access to looty.html
mkdir -p ~/looty

echo "Installed looty to ~/.local/bin"
echo "Run 'looty' to start - it will extract looty.html to:"
echo "  1. The current directory"
echo "  2. ~/looty/ (for easy phone transfer)"
echo ""
echo "Restart your terminal or run: export PATH=\"$HOME/.local/bin:\$PATH\""
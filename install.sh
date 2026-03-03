#!/bin/sh
set -e

ARCH="$(uname -m | sed -e 's/x86_64/amd64/' -e 's/aarch64/arm64/')"
OS="$(uname | sed -e 's/Darwin/macos/' -e 's/Linux/linux/')"

curl -sL "https://github.com/lirrensi/looty/releases/latest/download/looty-${OS}-${ARCH}.tar.gz" | tar xz

mkdir -p ~/.local/bin
mv looty ~/.local/bin/

echo "Installed looty to ~/.local/bin - restart your terminal or run: export PATH=\"$HOME/.local/bin:\$PATH\""
# Looty 🐱

A portable file sync & clipboard sharing tool between desktop and mobile on your local network. Zero config, auto-discovery, single executable.

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Downloads](https://img.shields.io/github/downloads/lirrensi/looty/latest/total?label=Downloads)](https://github.com/lirrensi/looty/releases/latest)

![Looty](assets/le_cat.jpg)

---

## ✨ Features

- 📱 **Zero-Config**: Auto-discovery on local network
- 🔄 **File Sync**: Drag & drop files between your desktop and phone
- 📋 **Clipboard Sync**: Copy text/pixels from desktop to mobile
- 💾 **Scratchpad**: Quick note-taking between devices
- 🌐 **Local Only**: No internet required, completely private
- 🚀 **Portable**: Single executable, runs anywhere

---

## 🚀 Quick Start

### Windows
```powershell
# Download the latest release
$arch = if ($env:PROCESSOR_ARCHITECTURE -eq "AMD64") { "amd64" } else { "arm64" }
curl -sL "https://github.com/lirrensi/looty/releases/latest/download/looty-windows-${arch}.zip" -o looty.zip
Expand-Archive looty.zip -Force
Remove-Item looty.zip

# Run looty in any folder
looty-windows-${arch}.exe
```

### macOS / Linux
```bash
# Download the latest release
ARCH="$(uname -m | sed -e 's/x86_64/amd64/' -e 's/aarch64/arm64/')"
OS="$(uname | tr '[:upper:]' '[:lower:]' | sed -e 's/darwin/macos/')"
curl -sL "https://github.com/lirrensi/looty/releases/latest/download/looty-${OS}-${ARCH}.tar.gz" | tar xz

# Run looty in any folder
./looty
```

---

## 📱 How It Works

1. **Run Looty** on your desktop in any folder
2. **Copy `looty.html`** from the looty.exe directory to your phone
3. **Open in browser** - your phone will auto-discover your desktop
4. **Done!** File sync and clipboard sharing work instantly

### Desktop URLs
After running, Looty will show you these URLs:
```
http://192.168.1.X:41111
http://localhost:41111
```

---

## 📦 Building from Source

### Prerequisites
- Go 1.25.5+
- Node.js 20+
- pnpm

### Build
```bash
# Clone the repo
git clone https://github.com/lirrensi/looty.git
cd looty

# Install frontend dependencies
cd web && npm install && cd ..

# Build
make build
```

The executable will be created in the root directory.

---

## 🛠️ Development

### Start the web dev server
```bash
cd web
pnpm install
pnpm dev
```

### Run tests
```bash
cd web
pnpm test
```

---

## 📁 Project Structure

```
looty/
├── cmd/blip/           # Main entry point
├── internal/
│   ├── clipboard/      # Clipboard sync logic
│   ├── files/          # File upload/download handlers
│   └── server/         # HTTP + WebSocket server
├── web/                # Frontend (Vite + Alpine + Tailwind)
├── embed/              # Embedded assets
└── assets/
    └── le_cat.jpg      # Looty mascot 🐱
```

---

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

---

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

---

## 🐱 Made with ❤️ for the local network

# Looty 🐱

A portable file sync & clipboard sharing tool for desktop, mobile, and remote servers. Zero config, auto-discovery, single executable.

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Downloads](https://img.shields.io/github/downloads/lirrensi/looty/latest/total?label=Downloads)](https://github.com/lirrensi/looty/releases/latest)

![Looty](assets/le_cat.jpg)

---

## 📱 How It Works

**Zero Bullshit — 3 Steps:**

1. **Run Looty** in your terminal
2. **Scan the QR code** with your phone
3. **Open mobile, grab files** — done!

### 🌍 Remote Server Mode

Looty can also run on a remote server, VM, or internal machine. When it binds to a non-loopback address, it automatically enables TLS with a self-signed certificate and prints a fingerprint for verification.

### ⚙️ Run Modes & Commands

Looty now has a few startup shapes:

- **Interactive / foreground**: `looty`
  - Stays attached to your terminal and prints the normal startup view
- **Background / daemon**: `looty -daemon`
  - Keeps running after launch and preserves startup details
- **Programmatic / machine-readable**: `looty -daemon -json`
  - Emits startup details as JSON for scripts, agents, or other processes
- **Save startup details to a file**: `looty -daemon -json-file startup.json`
  - Writes the same startup record for later pickup

For remote or non-loopback hosts, Looty auto-enables TLS and prints the certificate fingerprint/friend code so you can verify the server before connecting.

```
┌─────────────────────────────────────────────┐
│  🐱 Looty v0.1.0                            │
│                                             │
│  📱 Scan to connect:                        │
│                                             │
│       ████████████████                      │
│       ██          ██                        │
│       ██  ██  ██  ██                        │
│       ██    ██      ██                      │
│       ██  ██  ██  ██                        │
│       ██          ██                        │
│       ████████████████                      │
│                                             │
│  Or open: http://192.168.1.X:41111          │
│                                             │
│  🎯 Ready to sync!                          │
└─────────────────────────────────────────────┘
```

### Screenshots

| Desktop (Terminal) | Mobile (Files) | Mobile (Scratchpad) |
|--------------------|----------------|---------------------|
| ![CLI](assets/cli.png) | ![Mobile Files](assets/mobile_file.jpg) | ![Mobile Pad](assets/mobile_pad.jpg) |

**That's it.** Run Looty → Scan QR → Use your phone as a wireless drive.

---

## ✨ Features

- 📱 **Zero-Config**: Auto-discovery on local network
- 📷 **QR Connect**: Scan QR code with your phone to connect instantly
- 🔄 **File Sync**: Drag & drop files between your desktop and phone
- 📋 **Clipboard Sync**: Copy text/pixels from desktop to mobile
- 💾 **Scratchpad**: Quick note-taking between devices
- 🌐 **Local-First**: No internet required for localhost; remote servers use auto-TLS with self-signed certs
- 🔒 **Remote-Ready**: Works on non-loopback hosts with fingerprint verification
- 🚀 **Portable**: Single executable, runs anywhere

---

## 🚀 Quick Install

```bash
# macOS / Linux
curl -sL https://raw.githubusercontent.com/lirrensi/looty/main/install.sh | sh
```

```powershell
# Windows
irm https://raw.githubusercontent.com/lirrensi/looty/main/install.ps1 | iex
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

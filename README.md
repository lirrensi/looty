# Looty

Portable file sync & clipboard sharing between desktop and mobile on your local network. Zero config, auto-discovery, single executable.

## Quick Start

```bash
# Run the server (serves current directory)
./looty

# Or run from source (Go)
go run ./cmd/blip

# Or run the web dev server
cd web && pnpm dev
```

## Install (One-time)

### Windows (PowerShell)
```powershell
irm https://raw.githubusercontent.com/YOUR_GITHUB_USER/BlipSync/main/install.ps1 | iex
```

### macOS / Linux
```bash
curl -sL https://raw.githubusercontent.com/YOUR_GITHUB_USER/BlipSync/main/install.sh | bash
```

After install, run `looty` from ANY folder — it will serve that folder on your network!

## Project Structure

```
BlipSync/
├── cmd/blip/           # Main entry point
├── internal/           # Core server logic
│   ├── server/         # HTTP + WebSocket server
│   ├── clipboard/      # Clipboard sync
│   └── files/          # File operations
├── web/                # Frontend (Vite + Alpine + Tailwind)
├── embed/              # Embedded assets
├── install.ps1         # Windows install script
├── install.sh          # macOS/Linux install script
└── .github/workflows/  # CI/CD for releases
```

## Development

### Backend (Go)
```bash
go run ./cmd/blip
```

### Frontend (Web)
```bash
cd web
pnpm install
pnpm dev
```

### Build for release
```bash
make build
```

## Usage

1. Run `looty` in any folder on your desktop
2. Copy the generated `looty.html` to your phone
3. Open `looty.html` in your phone's browser
4. Auto-discovery connects your phone to your desktop instantly!

## License

MIT
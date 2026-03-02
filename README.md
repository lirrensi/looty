# Blip

Portable file sync & clipboard sharing between desktop and mobile on your local network. Zero config, auto-discovery, single executable.

## Quick Start

```bash
# Run the server
./blip.exe

# Or run from source (Go)
go run ./cmd/blip

# Or run the web dev server
cd web && pnpm dev
```

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
├── private/            # Private ideas & notes
└── docs/               # Documentation
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

## Usage

1. Run `blip.exe` on your desktop
2. Copy the generated `blip.html` to your phone
3. Open `blip.html` in your phone's browser
4. Auto-discovery connects your phone to your desktop instantly!

## License

MIT
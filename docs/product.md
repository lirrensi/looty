# Looty

Zero-config file access and clipboard sync between desktop and mobile on local network.

---

## What It Is

Drop `looty.exe` in any folder. Run it. Open `looty.html` on your phone. Done — you have instant access to that folder and a synced clipboard.

No configuration. No accounts. No cloud. Just works.

---

## Core Value

**"Copy the file once, stop thinking about networking."**

- Desktop: One executable, serves the folder it lives in
- Mobile: One HTML file, auto-discovers the server
- Clipboard: Type on phone, appears on desktop instantly
- Files: Browse, upload, download, preview — from anywhere on LAN

---

## How It Works

```
1. User copies looty.exe to a folder
2. User runs looty.exe
   - Server starts on port 41111
   - looty.html is extracted to the same folder (if not present)
3. User copies looty.html to phone (once, ever)
4. User opens looty.html on phone
   - Auto-scans local network for looty.exe server
   - Connects to first server found
5. User browses files, uploads, downloads, uses clipboard
```

### Run Modes

Looty supports three user-visible ways to run the same server:

1. **Foreground mode** — run in a terminal, see the access URL, QR code, and any TLS trust details, then stop serving when the process exits.
2. **Background mode** — start the same folder server as a long-running process without holding the terminal open, while still capturing the startup details needed to connect.
3. **Agent-managed mode** — start Looty in the background on behalf of a user, but return the same startup details programmatically so another tool or agent can relay them back to the user.

In all modes, the served resource is still "this folder for as long as this process lives". The difference is how the process is attached and how startup information is delivered.

---

## Features

### File Browser
- List files and folders in the served directory
- Navigate into subfolders
- **Breadcrumb navigation** with individual folder click
- **Up button** for quick navigation
- **Sort options**: Name (A→Z, Z→A), Date (Newest, Oldest)
- **File preview** for text files and images
- **Binary detection** (server-side) to avoid showing binary content
- **Download any file** to phone
- **Upload files** from phone to current folder (max 100MB)
- **Real-time updates** when files change on desktop
- Multiple devices can connect simultaneously

### Clipboard Sync (Scratchpad)
- **Real-time text sync** across all connected devices
- **Type on phone → appears on desktop instantly**
- **Type on desktop → appears on phone instantly**
- **History panel** showing last 50 items
- **Click history item** to restore it
- **Copy to clipboard** button with fallback for older browsers
- Debounced sync to avoid spamming

### Discovery
- **Cache first**: Checks localStorage for previously-found server IP (instant reconnect)
- **mDNS magic**: Server announces as `looty.local:41111` — browsers resolve automatically
- **Smart parallel scan**: Probes common subnets (192.168.0.x, 192.168.1.x, 10.0.0.x) with first 32 IPs each in parallel
- **Fallback expansion**: If not found, expands to full 254 IPs per subnet
- **Manual IP entry**: Final fallback if auto-discovery fails
- **Debug log**: Shows all discovery attempts
- **Success indicators**: Green dot for connected, yellow for searching

### Technical Features
- **Single executable** (~15MB)
- **Single HTML file** for mobile
- **Embedded assets** (icons, styles)
- **Build time injection** (optional timestamp in UI)
- **Auto-reconnect** WebSocket on connection loss
- **File watching** on desktop server

---

## User Interface

### Desktop (looty.exe)
- Command-line output with:
  - Served directory path
  - Available IP addresses
  - Access URLs (http://IP:41111)
  - Instructions for copying looty.html to phone
  - QR code in foreground mode
  - TLS fingerprint and friend code when self-signed TLS is active

### Background Startup Artifact
- Background-capable runs must still produce a retrievable startup record
- That record must include the connection URL and any trust material needed to connect safely
- In self-signed TLS mode, the startup record must include the certificate fingerprint and friend code
- The startup record is intended for users, service managers, and agents that need fire-and-forget startup without losing connection details

### Mobile (looty.html)
- **Dark theme** optimized for mobile
- **Discovery overlay** with status and debug log
- **Header**: Shows connection status (green/yellow dot)
- **Breadcrumb navigation**: Click any folder to navigate
- **Up button**: Quick return to parent folder
- **Sort bar**: Toggle between sort modes
- **File list**: Click files/folders, shows preview panel below
- **Preview panel**: Shows text content, images, or binary file message
- **Upload progress**: Visual progress bar
- **Download progress**: Visual progress bar
- **Bottom bar**: Upload button, Scratchpad button, Refresh button

### Scratchpad Panel
- **Full-screen overlay** with dark theme
- **Main textarea**: Large editing area for scratchpad content
- **History sidebar**: Shows last 50 items on the right
- **Connection status**: Shows connecting/connected/disconnected state
- **Auto-sync**: Every 300ms debounce while typing

---

## File Preview Support

### Text Files
- `.txt`
- `.md`
- `.json`
- `.log`
- `.js`
- `.css`
- `.html`
- `.py`
- `.rs`
- `.go`
- `.c`
- `.cpp`
- `.h`
- `.java`
- `.rb`
- `.php`
- `.sql`
- `.yaml`
- `.yml`
- `.toml`
- `.xml`
- `.csv`
- `.tsv`
- `.sh`
- `.bat`
- `.ps1`
- `.env`
- `.config`
- Any other text-based files

### Image Files
- `.jpg`
- `.jpeg`
- `.png`
- `.gif`
- `.webp`
- `.bmp`
- `.ico`

### Binary Files
- Detected by server-side binary check (first 8KB looking for null bytes)
- Shown with "Binary file" message
- Download button available

---

## API Endpoints

### HTTP API
- `GET /` - Serve index.html
- `GET /ping` - Health check for discovery (returns "pong")
- `GET /api/files?path={path}` - List files in directory
- `GET /api/download?path={path}` - Download file
- `POST /api/upload` - Upload file (max 100MB)
- `POST /api/scratchpad` - Update scratchpad content
- `GET /api/scratchpad` - Get current scratchpad content
- `OPTIONS /` and `OPTIONS /api/*` - CORS preflight support

### WebSocket
- `ws://host:41111/ws` - Real-time sync endpoint
- **Message types**:
  - `{"type":"clipboard","data":"text"}` - Clipboard sync
  - `{"type":"scratchpad","data":"text"}` - Scratchpad sync
  - `{"type":"refresh","data":""}` - File change notification

---

## Technical Stack

| Component | Technology |
|-----------|------------|
| Server | Go 1.25.5 |
| Frontend | Vite 7.3.1 + Alpine.js 3.15.8 + Tailwind CSS 4.2.1 |
| Frontend Output | Single HTML file (Vite plugin singlefile) |
| Real-time | WebSocket (gorilla/websocket 1.5.3) |
| File watching | fsnotify 1.9.0 |
| Build | Go ldflags for build time injection |

---

## Distribution

```
looty.exe          # ~15MB, self-contained Go binary
  ├── embeds: assets/index.html (single file UI)
  └── extracts: looty.html (on first run, for phone)

looty.html         # Copied to phone once, works forever
                   # Also saved to ~/looty/ (Unix) or %USERPROFILE%\looty\ (Windows)
                   # for easy discovery after install
```

---

## Security Model

### Current Implementation
- **Auto-TLS for remote access** — when binding to non-loopback addresses, Looty auto-generates a self-signed TLS certificate per run
- **Fingerprint verification** — the terminal displays the certificate SHA-256 fingerprint and a human-readable "friend code". Compare it in your browser to verify you're connecting to your server (SSH-style host key trust)
- **Plain HTTP for localhost** — when binding to `localhost` or `127.0.0.1`, serves plain HTTP (loopback is trusted)
- **Opt-out available** — use `-no-tls` flag for plain HTTP on any interface (legacy LAN mode)
- **No authentication** — anyone who can reach the address can access (protected by TLS when enabled)
- **Path traversal protection** — absolute path validation prevents directory traversal
- **Port 41111** — dedicated port, not a standard service
- **Serves ONLY the folder it's in** — cannot access parent directories

### TLS Modes

| Scenario | Command | Behavior |
|---|---|---|
| Localhost only | `looty -host 127.0.0.1` | Plain HTTP, no certificate warnings |
| Share across networks / untrusted LAN | `looty` or `looty -host 0.0.0.0` | Auto-TLS with fingerprint verification |
| Legacy LAN mode (trusted network) | `looty -no-tls` | Plain HTTP on all interfaces |
| Custom certificate | `looty -cert cert.pem -key key.pem` | Uses your own TLS certificate |
| Force TLS on localhost | `looty -tls -host 127.0.0.1` | Auto-TLS even on loopback |

### Process Modes

| Scenario | Expected behavior |
|---|---|
| Interactive terminal use | Looty prints human-friendly startup details and stays attached to the terminal |
| Persistent local serving | Looty can be started in a background/daemon style while keeping startup details available after launch |
| Agent/service launch | Looty can expose startup details in a machine-readable form so another process can return them to the user |

### Security Considerations
- On untrusted networks, always use TLS mode and verify the fingerprint
- Don't expose `-no-tls` mode to public internet
- Self-signed certificates are short-lived (24 hours) and generated per run
- Regularly update executable to get security patches
- Firewall rules can restrict access to specific devices
- Background or daemon launch must not suppress or discard the TLS trust details needed for safe first connection

---

## User Experience Goals

- **Time to first connection**: < 10 seconds (usually < 3 seconds)
- **Zero configuration required** — just run and go
- **Works on any local network** — home, office, coffee shop
- **Mobile-first interface** — touch-optimized UI
- **Graceful degradation** — manual IP entry if discovery fails
- **Visual feedback** — progress bars, status indicators, toast messages
- **Error handling** — clear error messages with debug information

---

## What Looty Is Not

- Not a cloud sync service (no Dropbox/S3/OneDrive)
- Not a file versioning system
- Not a multi-folder server
- Not a public internet tool (LAN only)
- Not a P2P network (server-client model)
- Not a file transfer protocol (FTP/SFTP)
- Not a backup solution

---

## Success

One executable. One HTML file. Open both. They find each other. You're done.

---

## Version History

### v1.0 (Current)
- Zero-config file browser with preview
- Real-time clipboard sync (scratchpad)
- Auto-discovery with subnet scanning
- File watching and real-time updates
- Breadcrumb navigation and sorting
- Upload/download with progress tracking
- Binary file detection
- Multi-device support

---

## Platform Support

### Desktop Server
- **Windows**: 64-bit (tested)
- **macOS**: 64-bit (compatible)
- **Linux**: 64-bit (compatible)

### Mobile Client
- **iOS**: Safari, Chrome, Firefox (any modern browser)
- **Android**: Chrome, Firefox, Safari (any modern browser)
- **Requirements**: Modern browser with JavaScript, WebSocket support

---

## Known Limitations

- No file deletion API (read-only in MVP)
- No folder creation API
- No file renaming API
- No search functionality
- No file metadata editing
- No folder permissions
- No upload queue management
- No download resume
- No thumbnail generation
- No text editing with server-side persistence
- No collaborative editing
- No offline mode

---

## Troubleshooting

### Server not found
1. Check server is running on desktop
2. Verify firewall allows port 41111
3. Check server output for IP addresses
4. Try manual IP entry in discovery screen

### Clipboard not syncing
1. Verify both devices are connected
2. Check WebSocket connection in Scratchpad panel
3. Ensure both devices are on same LAN
4. Try refreshing page

### Files not updating
1. Check file watcher is running (server logs will show "File changed")
2. Refresh file browser manually
3. Verify file is within served directory

### Upload fails
1. Check max file size (100MB limit)
2. Verify write permissions on server folder
3. Check disk space
4. Try uploading smaller file

---

## FAQ

**Q: Can I use it across different networks?**
A: Yes, if you can reach the host IP. Run `looty` (auto-TLS), share the `https://` link and fingerprint with your phone. Open the link directly — `looty.html` file:// discovery does not work with HTTPS.

**Q: Is my data secure?**
A: By default on non-loopback addresses, Looty uses auto-generated TLS with certificate fingerprint verification. Compare the fingerprint shown in your terminal with what your browser displays to confirm it's your server. For trusted LAN only, you can use `-no-tls`.

**Q: Can I run Looty in the background or under systemd?**
A: Yes — the intended behavior is that Looty supports persistent serving without holding the terminal open, while still producing a startup record with the URL and, when relevant, the TLS fingerprint and friend code.

**Q: Can multiple phones connect?**
A: Yes, multiple devices can connect simultaneously.

**Q: Does it work over VPN?**
A: Yes, as long as devices are on the same virtual network.

**Q: Can I delete files?**
A: Not yet in v1.0 (upload-only). Future versions may add file management.

**Q: Does it work with external drives?**
A: Yes, if the drive is mapped and accessible by the server's user.

**Q: Can I customize the UI?**
A: Yes, modify the source HTML and rebuild.

**Q: Can I use it for commercial projects?**
A: Yes, MIT license allows commercial use.

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) (if exists) or see README.md for contribution guidelines.

---

## License

MIT License - see LICENSE file for details.

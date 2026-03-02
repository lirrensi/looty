# Blip

Zero-config file access and clipboard sync between desktop and mobile on local network.

---

## What It Is

Drop `blip.exe` in any folder. Run it. Open `blip.html` on your phone. Done — you have instant access to that folder and a synced clipboard.

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
1. User copies blip.exe to a folder
2. User runs blip.exe
   - Server starts on port 8080
   - blip.html is extracted to the same folder (if not present)
3. User copies blip.html to phone (once, ever)
4. User opens blip.html on phone
   - Auto-scans local network for blip.exe server
   - Connects to first server found
5. User browses files, uploads, downloads, uses clipboard
```

---

## Features

### File Browser
- List files and folders in the served directory
- Navigate into subfolders
- Preview text files (.txt, .md, .json, .log, .js, .css, .html)
- Preview images (.jpg, .png, .gif, .webp)
- Download any file to phone
- Upload files from phone to current folder

### Clipboard Sync
- Text input on phone → instantly appears on desktop
- Text input on desktop → instantly appears on phone
- History of last 10 clipboard items
- One-click copy to system clipboard

### Real-time Updates
- File changes on desktop → all connected devices refresh
- Multiple devices can connect simultaneously
- Auto-reconnect on connection loss

---

## Technical Stack

| Component | Technology |
|-----------|------------|
| Server | Go (single binary) |
| Frontend | Vite + Alpine.js + Tailwind |
| Frontend Output | Single HTML file (inlined) |
| Real-time | WebSocket |
| File watching | fsnotify |

---

## Distribution

```
blip.exe          # ~15MB, self-contained Go binary
  ├── embeds: index.html (single file UI)
  └── extracts: blip.html (on first run, for phone)

blip.html         # Copied to phone once, works forever
```

---

## Security Model (MVP)

- No authentication — anyone on LAN can access
- HTTP only — no encryption (local network only)
- Serves ONLY the folder it's in — no parent directory traversal
- Port scanning limited to local subnet

*Future: optional password, HTTPS, read-only mode*

---

## User Experience Goals

- Time to first connection: < 10 seconds
- Zero configuration required
- Works on any local network (home, office, coffee shop)
- Mobile-first interface
- Graceful degradation (manual IP entry if discovery fails)

---

## What Blip Is Not

- Not a cloud sync service (no Dropbox/S3)
- Not a file versioning system
- Not a multi-folder server
- Not a public internet tool (LAN only)

---

## Success

One executable. One HTML file. Open both. They find each other. You're done.

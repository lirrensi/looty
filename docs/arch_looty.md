# Looty Architecture

High-level overview of Looty's technical design and component relationships.

---

## Overview

Looty is a client-server application with a Go backend serving a local network and a web-based mobile client. The architecture is intentionally simple and focused on zero configuration and ease of use.

---

## System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         Client                               │
│  ┌───────────────────────────────────────────────────────┐  │
│  │                     Mobile Browser                     │  │
│  │  ┌────────────┐  ┌──────────────┐  ┌──────────────┐  │  │
│  │  │ Discovery  │  │ File Browser │  │ Scratchpad   │  │  │
│  │  │ Component  │  │ Component    │  │ Component    │  │  │
│  │  └────────────┘  └──────────────┘  └──────────────┘  │  │
│  └───────────────────────────────────────────────────────┘  │
│         │         │         │                               │
│         └─────────┴─────────┴───────────────────────────────┘
└─────────────────────────────────────────────────────────────┘
                            ▲
                            │ WebSocket
                            │ HTTP
┌─────────────────────────────────────────────────────────────┐
│                     Server (Go)                              │
│  ┌───────────────────────────────────────────────────────┐  │
│  │                    HTTP Server                        │  │
│  │  ┌────────────┐ ┌─────────────┐ ┌─────────────┐      │  │
│  │  │ File API   │ │ Scratchpad  │ │ Discovery   │      │  │
│  │  │ Handlers   │ │ Handlers    │ │ Ping/Health │      │  │
│  │  └────────────┘ └─────────────┘ └─────────────┘      │  │
│  └───────────────────────────────────────────────────────┘  │
│  ┌───────────────────────────────────────────────────────┐  │
│  │                  WebSocket Server                      │  │
│  │  ┌────────────┐ ┌─────────────┐ ┌─────────────┐      │  │
│  │  │ Hub        │ │ Clients     │ │ Broadcast   │      │  │
│  │  │ Manager    │ │ Manager     │ │ Router      │      │  │
│  │  └────────────┘ └─────────────┘ └─────────────┘      │  │
│  └───────────────────────────────────────────────────────┘  │
│  ┌───────────────────────────────────────────────────────┐  │
│  │                  File Watcher                          │  │
│  │         (fsnotify) - detects file changes              │  │
│  └───────────────────────────────────────────────────────┘  │
│  ┌───────────────────────────────────────────────────────┐  │
│  │                  File System                           │  │
│  │  ┌────────────┐ ┌─────────────┐ ┌─────────────┐      │  │
│  │  │ List Files │ │ Download    │ │ Upload      │      │  │
│  │  │ Operations │ │ Operations  │ │ Operations  │      │  │
│  │  └────────────┘ └─────────────┘ └─────────────┘      │  │
│  └───────────────────────────────────────────────────────┘  │
│  ┌───────────────────────────────────────────────────────┐  │
│  │                   Scratchpad                           │  │
│  │           (in-memory string with mutex)                │  │
│  └───────────────────────────────────────────────────────┘  │
│  ┌───────────────────────────────────────────────────────┐  │
│  │                   Assets                               │  │
│  │            (embedded HTML with build time)             │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                            │
                            │ Serves directory
┌─────────────────────────────────────────────────────────────┐
│                  Local File System                          │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

---

## Component Breakdown

### 1. Main Entry Point (`cmd/blip/main.go`)

**Responsibilities:**
- Get current working directory (serves this folder)
- Get executable path for extracting looty.html
- Call `server.Start()` to begin serving
- Print access URLs and instructions to console

**Key Functions:**
- `getLocalIPs()` - Scans network interfaces for local IPs
- `main()` - Orchestrates startup

**Dependencies:**
- `github.com/user/looty/internal/server`

---

### 2. Server Package (`internal/server/`)

**Responsibilities:**
- HTTP server with CORS middleware
- WebSocket server with broadcast hub
- File watching integration
- Scratchpad state management
- Route registration

**Key Components:**

#### 2.1 HTTP Server (`server.go`)

**Routes:**
- `GET /` - Serve embedded index.html
- `GET /ping` - Health check for discovery
- `GET /api/files` - List directory contents
- `GET /api/download` - Download file
- `POST /api/upload` - Upload file
- `GET/POST /api/scratchpad` - Scratchpad operations
- `ws` - WebSocket upgrade endpoint

**Middleware:**
- CORS handler for all routes (Access-Control-Allow-Origin: *)

**Key Functions:**
- `Start(serveDir, port)` - Initialize and start HTTP server
- `GetHTML()` - Retrieve embedded HTML with build time injected
- `GetScratchpad()` / `SetScratchpad()` - Thread-safe scratchpad access
- `Broadcast(message)` - Send message to all WebSocket clients

#### 2.2 WebSocket Hub (`websocket.go`)

**Responsibilities:**
- Manage connected WebSocket clients
- Broadcast messages to all clients
- Handle client registration/disconnection

**Key Types:**
- `Hub` - Central broadcast hub with channels
- `Client` - Individual WebSocket connection

**Concurrency:**
- Uses buffered channels for messages (256 capacity)
- Mutex-protected client map
- Separate goroutines for reading/writing

**Message Flow:**
```
Client send → readPump → Parse → Broadcast hub → writePump → All clients
```

#### 2.3 File Watcher (`watcher.go`)

**Responsibilities:**
- Monitor served directory for file changes
- Broadcast "refresh" messages on changes

**Events Detected:**
- Create
- Write
- Remove

**Key Functions:**
- `StartWatcher(dir)` - Initialize fsnotify watcher

---

### 3. File Operations (`internal/files/`)

**Responsibilities:**
- List directory contents
- Download files
- Upload files
- Binary file detection

**Key Functions:**

#### 3.1 ListHandler

**Security:**
- Prevents path traversal (`..` in path)
- Validates absolute path is within serveDir
- Returns file metadata (name, path, size, modified, isBinary)

**Binary Detection:**
- Reads first 8KB of file
- Checks for null bytes
- Returns `isBinary: true` if null bytes found

#### 3.2 DownloadHandler

**Security:**
- Path traversal protection
- Absolute path validation
- Returns 404 for non-existent files
- Returns 400 for directories

**Streaming:**
- Uses `http.ServeFile` for efficient streaming

#### 3.3 UploadHandler

**Security:**
- Path traversal protection
- Validates file name and destination path
- Absolute path validation
- Max 100MB limit

**File Creation:**
- Creates file atomically
- Writes in chunks via `io.Copy`

---

### 4. Clipboard System (`internal/clipboard/`)

**Responsibilities:**
- Define message types
- Serialize/deserialize messages

**Message Types:**
- `TypeClipboard` - Generic clipboard sync
- `TypeRefresh` - File change notification
- `TypeScratchpad` - Scratchpad-specific messages

**Message Structure:**
```json
{
  "type": "clipboard",
  "data": "text content"
}
```

---

### 5. Frontend Architecture (`web/`)

**Tech Stack:**
- **Build Tool**: Vite 7.3.1
- **Framework**: Alpine.js 3.15.8 (reactive UI)
- **Styling**: Tailwind CSS 4.2.1
- **Plugin**: vite-plugin-singlefile (inlines everything into one HTML file)
- **Icons**: Custom icon SVG strings

**Output:**
- Single `index.html` file embedded in Go binary
- Extracts to `looty.html` on desktop

#### 5.1 Main Application (`src/main.js`)

**Responsibilities:**
- Initialize Alpine.js
- Manage application state (connected, serverIP, status, showClipboard)
- Handle server discovery
- Initialize sub-components

**State:**
- `connected` - WebSocket connection status
- `serverIP` - Discovered server IP
- `status` - 'searching', 'connected', 'failed'
- `showClipboard` - Scratchpad panel visibility
- `manualIP` - User-entered fallback IP

#### 5.2 Discovery Component (`src/components/discovery.js`)

**Discovery Strategies (in order of priority):**

1. **Current Host** - If running from server (file:// or same host)
   - Checks `window.location.hostname:41111`
   - Fastest option for local testing

2. **Localhost** - Always check first
   - `localhost:41111`
   - Instant for local development

3. **Subnet Scan** - Fallback for network discovery
   - Detects subnet by pinging common subnets (192.168.1.x, 192.168.0.x, 10.0.0.x, 192.168.2.x)
   - Pings IPs 1-254
   - Aborts on first success

**Key Functions:**
- `findServer()` - Orchestrates discovery process
- `detectSubnet()` - Tries common subnets to detect network
- `pingServer(ip)` - Tests connectivity with 500ms timeout
- `logMsg(msg)` - Adds to debug log

#### 5.3 File Browser Component (`src/components/fileBrowser.js`)

**Responsibilities:**
- File listing and navigation
- File preview
- Upload/download
- Sorting

**State:**
- `files` - Array of FileInfo objects
- `currentPath` - Current directory path
- `selectedFile` - Currently selected file
- `preview` - Preview content (text, image, binary, error)
- `loading` - Loading state
- `uploadProgress` / `downloadProgress` - Progress percentages
- `sortModes` - [name-asc, name-desc, date-desc, date-asc]
- `sortModeIndex` - Current sort mode

**Features:**
- **Breadcrumb Navigation**: Click any folder in path
- **Up Button**: Quick return to parent
- **Sorting**: Name (A→Z/Z→A), Date (Newest/Oldest)
- **File Preview**:
  - Text files: Read entire file
  - Images: Display in preview panel
  - Binary: Show message, provide download
- **Progress Tracking**: Real-time upload/download progress
- **Auto-refresh**: Manual refresh button

#### 5.4 Clipboard Panel Component (`src/components/clipboard.js`)

**Responsibilities:**
- Scratchpad editing
- WebSocket message handling
- History management
- Clipboard copy

**State:**
- `content` - Current scratchpad text
- `history` - Array of {text, time} objects (max 50)
- `ws` - WebSocket connection
- `wsConnected` - WebSocket status
- `wsError` - Connection error message
- `syncTimeout` - Debounce timer

**Features:**
- **Real-time Sync**: Debounced (300ms) sync on input
- **History**: Last 50 items, clickable to restore
- **Copy to Clipboard**: Uses `navigator.clipboard` with fallback
- **Auto-reconnect**: Reconnects every 3 seconds if disconnected
- **Connection Status**: Shows connecting/connected/disconnected state

**Message Types Handled:**
- `scratchpad`: Update content from other clients
- `clipboard`: Add to history
- `refresh`: Dispatch custom event to refresh file browser

---

## Data Flow

### File Browser Flow
```
User navigates → loadFiles(path) → fetch /api/files → update files array
User clicks file → selectFile() → fetch /api/download → render preview
User uploads file → uploadFile() → POST /api/upload → update files array
Server changes file → fsnotify event → Broadcast refresh → WebSocket → Client refreshes
```

### Clipboard Sync Flow
```
User types → onInput() (debounce) → syncToServer() → POST /api/scratchpad
WebSocket receives message → onmessage() → update content + history
```

### Discovery Flow
```
App starts → discoverServer() → findServer()
  → Try current host → pingServer() → success?
  → Try localhost → pingServer() → success?
  → Detect subnet → scan 1-254 IPs → pingServer() → success?
  → Manual entry or fail
```

---

## State Management

### Server-Side State

#### Scratchpad
- **Location**: In-memory string
- **Protection**: `sync.RWMutex` for thread safety
- **Access**: `GetScratchpad()` (read), `SetScratchpad()` (write + broadcast)

#### WebSocket Clients
- **Location**: In-memory map `map[*Client]bool`
- **Protection**: `sync.RWMutex`
- **Broadcast**: Buffered channel with 256 capacity
- **Lifecycle**: Registered via channel, unregistered on disconnect

### Client-Side State

#### Alpine.js Components
- **File Browser**: `x-data="fileBrowser()"` with reactive methods
- **Clipboard Panel**: `x-data="clipboardPanel()"` with reactive methods
- **Main App**: `x-data="app()"` with global state
- **State Sharing**: Root app's `showClipboard` shared with clipboard panel via `$data`

#### Reactive Updates
- Alpine's reactivity system automatically updates UI when state changes
- Custom events dispatched for cross-component communication
- `x-init` hooks for initialization

---

## Error Handling

### Server-Side
- Path traversal attempts: Return 400 "Invalid path"
- File not found: Return 404 "File not found"
- Directory download: Return 400 "Cannot download directory"
- Invalid JSON: Return 400 "Invalid JSON"
- WebSocket upgrade failure: Log error, return early
- File watcher errors: Log error, continue running
- Binary detection: Returns `isBinary: true` (not an error)

### Client-Side
- Network errors: Display error message in UI
- WebSocket connection failure: Show "Connecting..." status, retry every 3 seconds
- File loading errors: Display error message in preview
- Upload/download failures: Clear progress, show error
- Discovery failures: Show manual IP entry option, debug log

---

## Performance Considerations

### Server-Side
- **File watching**: Single goroutine per directory, lightweight
- **Binary detection**: Only reads first 8KB of files
- **WebSocket broadcasting**: Buffered channel prevents blocking
- **Upload limit**: 100MB max prevents DoS

### Client-Side
- **Debounced sync**: 300ms debounce prevents spamming
- **History limit**: Max 50 items prevents memory bloat
- **Binary file check**: Server-side, client trusts it
- **Progress tracking**: Real-time updates via XHR progress events
- **File preview**: Lazy load on selection

---

## Security Model

### Current Protections
- **Path Traversal**: Absolute path validation prevents directory traversal
- **Directory Traversal**: Checks for `..` in paths
- **Binary Detection**: Prevents showing potentially malicious files
- **Port Isolation**: Uses non-standard port (41111)
- **CORS**: Allows all origins (LAN use case)

### Limitations
- **No Authentication**: Anyone on LAN can connect
- **No Encryption**: HTTP only, data transmitted in plaintext
- **No Rate Limiting**: No protection against DoS
- **No File Deletion**: Upload-only prevents accidental deletion

### Future Enhancements
- Optional password protection
- HTTPS/TLS encryption
- Rate limiting and authentication
- File access logging
- IP-based access control

---

## Build Process

### Frontend Build
```
web/index.html
  ↓ (Vite + vite-plugin-singlefile)
web/dist/index.html (single file, all assets inlined)
  ↓ (Copy to)
internal/server/assets/index.html
  ↓ (Go embed)
looty.exe binary
```

### Backend Build
```
go build -ldflags "-X github.com/user/looty/internal/server.BuildTime=$(date)" -o looty.exe ./cmd/blip
```

- `BuildTime` string injected into embedded HTML
- `assets/index.html` embedded via `//go:embed` directive
- Extracts `looty.html` on first run

---

## Testing Strategy

### Unit Tests (Not Implemented)
- File binary detection logic
- Path traversal prevention
- Message serialization/deserialization
- Scratchpad thread safety

### Integration Tests (Not Implemented)
- File upload/download flow
- WebSocket message broadcast
- Discovery with multiple scenarios

### Manual Testing Checklist
- [ ] Server starts on different ports
- [ ] Auto-discovery works on different networks
- [ ] Manual IP entry works
- [ ] File upload with progress tracking
- [ ] File download with progress tracking
- [ ] Clipboard sync across devices
- [ ] History panel shows last 50 items
- [ ] Binary file detection works
- [ ] Preview shows text/images correctly
- [ ] Breadcrumb navigation works
- [ ] Sorting works correctly
- [ ] WebSocket reconnects on disconnect
- [ ] File watching triggers refresh
- [ ] Multiple devices can connect simultaneously

---

## Deployment

### Desktop
1. Copy `looty.exe` to target folder
2. Run `looty.exe`
3. Copy generated `looty.html` to phone
4. Open `looty.html` in phone browser

### Production Considerations
- **Firewall**: Allow port 41111 through firewall
- **Port Forwarding**: If accessing from outside LAN, forward port 41111
- **IP Restrictions**: Use firewall rules to limit access to specific devices
- **Updates**: Regularly update executable to get bug fixes
- **Monitoring**: Monitor server logs for errors

---

## Future Enhancements (Not Implemented)

### Short Term
- File deletion API
- Folder creation API
- File renaming API
- Search functionality
- Thumbnail generation
- Download resume

### Medium Term
- User authentication
- HTTPS/TLS support
- Read-only mode
- File access logging
- Rate limiting
- Upload queue management

### Long Term
- Collaborative editing
- Offline mode
- Multi-folder support
- File versioning
- Sync conflicts resolution
- Cross-platform mobile apps (native)

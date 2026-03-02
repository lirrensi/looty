# Plan: Blip MVP Implementation
_A single Go binary with embedded web UI that provides zero-config file access and clipboard sync over LAN._

---

# Checklist
- [x] Step 1: Initialize Go module and create directory structure
- [x] Step 2: Create basic HTTP server with static file embedding
- [x] Step 3: Implement file listing API endpoint
- [x] Step 4: Implement file download endpoint
- [x] Step 5: Implement file upload endpoint
- [x] Step 6: Implement WebSocket server for real-time communication
- [x] Step 7: Implement clipboard sync over WebSocket
- [x] Step 8: Implement file watching with fsnotify
- [x] Step 9: Implement blip.html extraction on first run
- [x] Step 10: Initialize frontend with Vite, Alpine.js, and Tailwind
- [x] Step 11: Create file browser UI component
- [x] Step 12: Create clipboard UI component
- [x] Step 13: Implement auto-discovery client-side logic
- [x] Step 14: Build frontend to single HTML file
- [x] Step 15: Integrate frontend build into Go binary
- [x] Step 16: End-to-end testing and verification

---

## Context

Fresh repository at `C:\Users\rx\001_Code\105_DeadProjects\BlipSync`. Contains only `docs/product.md` and `idea.md`. No existing code.

Target architecture:
```
BlipSync/
├── cmd/
│   └── blip/
│       └── main.go           # Entry point
├── internal/
│   ├── server/
│   │   └── server.go         # HTTP + WebSocket server
│   ├── files/
│   │   └── files.go          # File operations (list, upload, download)
│   ├── clipboard/
│   │   └── clipboard.go      # Clipboard sync logic
│   └── discovery/
│       └── discovery.go      # Network discovery helpers
├── web/
│   ├── index.html            # Entry template
│   ├── src/
│   │   ├── main.js           # Alpine.js app
│   │   ├── components/
│   │   │   ├── fileBrowser.js
│   │   │   └── clipboard.js
│   │   └── style.css         # Tailwind directives
│   ├── package.json
│   ├── vite.config.js
│   └── tailwind.config.js
├── embed/
│   └── index.html            # Built frontend (checked in)
├── go.mod
└── go.sum
```

## Prerequisites

- Go 1.21+ installed and available in PATH
- Node.js 18+ installed and available in PATH
- npm installed and available in PATH
- Working directory: `C:\Users\rx\001_Code\105_DeadProjects\BlipSync`

## Scope Boundaries

- No authentication/authorization system
- No HTTPS/TLS
- No database
- No cloud sync
- No file versioning
- No settings UI

---

## Steps

### Step 1: Initialize Go module and create directory structure

Execute the following commands from `C:\Users\rx\001_Code\105_DeadProjects\BlipSync`:

```bash
go mod init github.com/user/blip
mkdir cmd\blip
mkdir internal\server
mkdir internal\files
mkdir internal\clipboard
mkdir internal\discovery
mkdir web\src\components
mkdir embed
```

Create `cmd/blip/main.go` with the following content:

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/user/blip/internal/server"
)

func main() {
	// Get the directory where blip.exe is located
	execPath, err := os.Executable()
	if err != nil {
		log.Fatal("Failed to get executable path:", err)
	}
	serveDir := filepath.Dir(execPath)
	
	fmt.Printf("Blip serving: %s\n", serveDir)
	fmt.Printf("Open http://localhost:8080 in your browser\n")
	
	if err := server.Start(serveDir, 8080); err != nil {
		log.Fatal("Server error:", err)
	}
}
```

✅ Success: All directories exist, `go.mod` file exists at project root, `cmd/blip/main.go` exists and contains the entry point code.
❌ If failed: If `go mod init` returns error, ensure Go is installed and in PATH. If directory creation fails, check permissions.

---

### Step 2: Create basic HTTP server with static file embedding

Install required Go dependency:
```bash
go get github.com/gorilla/websocket
```

Create `internal/server/server.go` with the following content:

```go
package server

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
)

//go:embed ../../embed/index.html
var staticFiles embed.FS

type Server struct {
	serveDir string
	port     int
}

func Start(serveDir string, port int) error {
	s := &Server{
		serveDir: serveDir,
		port:     port,
	}
	
	mux := http.NewServeMux()
	
	// Serve embedded index.html at root
	mux.HandleFunc("/", s.serveIndex)
	
	// Health check for discovery
	mux.HandleFunc("/ping", s.handlePing)
	
	addr := fmt.Sprintf(":%d", port)
	return http.ListenAndServe(addr, mux)
}

func (s *Server) serveIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	
	content, err := staticFiles.ReadFile("embed/index.html")
	if err != nil {
		http.Error(w, "Failed to load index", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "text/html")
	w.Write(content)
}

func (s *Server) handlePing(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}
```

✅ Success: `go build ./cmd/blip` compiles without errors. Running the binary prints "Blip serving: [dir]" and "Open http://localhost:8080". Accessing `http://localhost:8080/ping` returns "pong".
❌ If failed: If compile fails with "embed: pattern mismatch", create an empty `embed/index.html` file first (placeholder). If port is busy, return error message with "Port 8080 already in use".

---

### Step 3: Implement file listing API endpoint

Create `internal/files/files.go` with the following content:

```go
package files

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type FileInfo struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	IsDir   bool   `json:"isDir"`
	Size    int64  `json:"size"`
	ModTime string `json:"modTime"`
}

func ListHandler(serveDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get relative path from query, default to root
		relPath := r.URL.Query().Get("path")
		if relPath == "" {
			relPath = "."
		}
		
		// Security: prevent path traversal
		if strings.Contains(relPath, "..") {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}
		
		fullPath := filepath.Join(serveDir, relPath)
		
		// Verify path is within serveDir
		absServeDir, _ := filepath.Abs(serveDir)
		absFullPath, _ := filepath.Abs(fullPath)
		if !strings.HasPrefix(absFullPath, absServeDir) {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}
		
		entries, err := os.ReadDir(fullPath)
		if err != nil {
			http.Error(w, "Failed to read directory", http.StatusInternalServerError)
			return
		}
		
		files := make([]FileInfo, 0)
		for _, entry := range entries {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			
			files = append(files, FileInfo{
				Name:    entry.Name(),
				Path:    filepath.Join(relPath, entry.Name()),
				IsDir:   entry.IsDir(),
				Size:    info.Size(),
				ModTime: info.ModTime().Format("2006-01-02 15:04:05"),
			})
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"path":  relPath,
			"files": files,
		})
	}
}
```

Update `internal/server/server.go` to register the file listing endpoint. Add to the `Start` function before `http.ListenAndServe`:

```go
import "github.com/user/blip/internal/files"

// In Start function, add:
mux.HandleFunc("/api/files", files.ListHandler(s.serveDir))
```

✅ Success: Running server and accessing `http://localhost:8080/api/files` returns JSON array of files in the serve directory. `http://localhost:8080/api/files?path=subdir` returns files in subdirectory.
❌ If failed: If JSON is empty, verify serve directory is not empty. If 403 error, check path traversal logic.

---

### Step 4: Implement file download endpoint

Add to `internal/files/files.go`:

```go
func DownloadHandler(serveDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		relPath := r.URL.Query().Get("path")
		if relPath == "" {
			http.Error(w, "Path required", http.StatusBadRequest)
			return
		}
		
		// Security: prevent path traversal
		if strings.Contains(relPath, "..") {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}
		
		fullPath := filepath.Join(serveDir, relPath)
		
		// Verify path is within serveDir
		absServeDir, _ := filepath.Abs(serveDir)
		absFullPath, _ := filepath.Abs(fullPath)
		if !strings.HasPrefix(absFullPath, absServeDir) {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}
		
		// Check if file exists and is not a directory
		info, err := os.Stat(fullPath)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		if info.IsDir() {
			http.Error(w, "Cannot download directory", http.StatusBadRequest)
			return
		}
		
		http.ServeFile(w, r, fullPath)
	}
}
```

Update `internal/server/server.go` to register the download endpoint:

```go
mux.HandleFunc("/api/download", files.DownloadHandler(s.serveDir))
```

✅ Success: Accessing `http://localhost:8080/api/download?path=test.txt` downloads the file. Browser receives correct file with original filename.
❌ If failed: If 404, verify file exists in serve directory. If 403, check path traversal logic.

---

### Step 5: Implement file upload endpoint

Add to `internal/files/files.go`:

```go
import (
	"io"
	"path/filepath"
)

func UploadHandler(serveDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		// Max 100MB upload
		r.ParseMultipartForm(100 << 20)
		
		file, handler, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "No file provided", http.StatusBadRequest)
			return
		}
		defer file.Close()
		
		// Get destination path
		destPath := r.FormValue("path")
		if destPath == "" {
			destPath = "."
		}
		
		// Security: prevent path traversal
		if strings.Contains(destPath, "..") || strings.Contains(handler.Filename, "..") {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}
		
		fullPath := filepath.Join(serveDir, destPath, handler.Filename)
		
		// Verify path is within serveDir
		absServeDir, _ := filepath.Abs(serveDir)
		absFullPath, _ := filepath.Abs(fullPath)
		if !strings.HasPrefix(absFullPath, absServeDir) {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}
		
		// Create destination file
		dst, err := os.Create(fullPath)
		if err != nil {
			http.Error(w, "Failed to create file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()
		
		// Copy file content
		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, "Failed to save file", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"path":    filepath.Join(destPath, handler.Filename),
		})
	}
}
```

Update `internal/server/server.go` to register the upload endpoint:

```go
mux.HandleFunc("/api/upload", files.UploadHandler(s.serveDir))
```

✅ Success: POST request to `http://localhost:8080/api/upload` with multipart form file uploads successfully. File appears in serve directory.
❌ If failed: If 400 error, verify request is multipart/form-data with "file" field. If 500 error, check directory write permissions.

---

### Step 6: Implement WebSocket server for real-time communication

Create `internal/server/websocket.go`:

```go
package server

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for LAN use
	},
}

type Client struct {
	conn *websocket.Conn
	send chan []byte
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("Client connected. Total: %d", len(h.clients))
			
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("Client disconnected. Total: %d", len(h.clients))
			
		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (c *Client) writePump() {
	defer c.conn.Close()
	for message := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			break
		}
	}
}

func (c *Client) readPump(h *Hub, onMessage func([]byte)) {
	defer func() {
		h.unregister <- c
		c.conn.Close()
	}()
	
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		onMessage(message)
	}
}
```

Update `internal/server/server.go` to create hub and register WebSocket endpoint:

```go
var hub *Hub

func Start(serveDir string, port int) error {
	hub = NewHub()
	go hub.Run()
	
	// ... existing code ...
	
	mux.HandleFunc("/ws", s.handleWebSocket)
	
	// ... rest of function
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	
	client := &Client{
		conn: conn,
		send: make(chan []byte, 256),
	}
	hub.register <- client
	
	go client.writePump()
	go client.readPump(hub, func(msg []byte) {
		// Message handling will be added in Step 7
	})
}

func Broadcast(message []byte) {
	if hub != nil {
		hub.broadcast <- message
	}
}
```

✅ Success: WebSocket endpoint at `ws://localhost:8080/ws` accepts connections. Multiple clients can connect. Server logs "Client connected" messages.
❌ If failed: If connection fails, check browser console for WebSocket errors. Ensure gorilla/websocket is installed.

---

### Step 7: Implement clipboard sync over WebSocket

Create `internal/clipboard/clipboard.go`:

```go
package clipboard

import (
	"encoding/json"
)

type MessageType string

const (
	TypeClipboard MessageType = "clipboard"
	TypeRefresh   MessageType = "refresh"
)

type Message struct {
	Type MessageType `json:"type"`
	Data string      `json:"data"`
}

func ParseMessage(data []byte) (*Message, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func NewClipboardMessage(text string) []byte {
	msg := Message{
		Type: TypeClipboard,
		Data: text,
	}
	data, _ := json.Marshal(msg)
	return data
}

func NewRefreshMessage() []byte {
	msg := Message{
		Type: TypeRefresh,
		Data: "",
	}
	data, _ := json.Marshal(msg)
	return data
}
```

Update `internal/server/server.go` WebSocket handler to process clipboard messages:

```go
import "github.com/user/blip/internal/clipboard"

// In handleWebSocket, replace the onMessage callback:
go client.readPump(hub, func(msg []byte) {
	// Parse and broadcast clipboard messages
	message, err := clipboard.ParseMessage(msg)
	if err != nil {
		return
	}
	
	if message.Type == clipboard.TypeClipboard {
		// Broadcast to all other clients
		hub.broadcast <- msg
	}
})
```

✅ Success: Sending `{"type":"clipboard","data":"hello"}` via WebSocket from one client broadcasts to all connected clients. All clients receive the same JSON message.
❌ If failed: If messages not received, check hub is running and clients are registered. Use browser DevTools WebSocket frames to debug.

---

### Step 8: Implement file watching with fsnotify

Install fsnotify:
```bash
go get github.com/fsnotify/fsnotify
```

Create `internal/server/watcher.go`:

```go
package server

import (
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/user/blip/internal/clipboard"
)

func StartWatcher(dir string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("Failed to create watcher: %v", err)
		return
	}
	
	err = watcher.Add(dir)
	if err != nil {
		log.Printf("Failed to watch directory: %v", err)
		return
	}
	
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Create == fsnotify.Create ||
					event.Op&fsnotify.Write == fsnotify.Write ||
					event.Op&fsnotify.Remove == fsnotify.Remove {
					log.Printf("File changed: %s", event.Name)
					Broadcast(clipboard.NewRefreshMessage())
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("Watcher error: %v", err)
			}
		}
	}()
}
```

Update `internal/server/server.go` to start watcher:

```go
func Start(serveDir string, port int) error {
	// ... existing initialization ...
	
	StartWatcher(serveDir)
	
	// ... rest of function
}
```

✅ Success: Creating, modifying, or deleting a file in serve directory triggers broadcast of `{"type":"refresh"}` to all WebSocket clients. Server logs "File changed: [filename]".
❌ If failed: If watcher fails to start, check directory permissions. If no refresh messages sent, verify fsnotify is installed and watcher.Add succeeded.

---

### Step 9: Implement blip.html extraction on first run

Update `cmd/blip/main.go`:

```go
package main

import (
	"embed"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/user/blip/internal/server"
)

//go:embed ../../embed/index.html
var staticFiles embed.FS

func main() {
	execPath, err := os.Executable()
	if err != nil {
		log.Fatal("Failed to get executable path:", err)
	}
	serveDir := filepath.Dir(execPath)
	
	// Extract blip.html if it doesn't exist
	blipHTMLPath := filepath.Join(serveDir, "blip.html")
	if _, err := os.Stat(blipHTMLPath); os.IsNotExist(err) {
		src, err := staticFiles.Open("embed/index.html")
		if err != nil {
			log.Printf("Warning: Could not open embedded index.html: %v", err)
		} else {
			defer src.Close()
			dst, err := os.Create(blipHTMLPath)
			if err != nil {
				log.Printf("Warning: Could not create blip.html: %v", err)
			} else {
				defer dst.Close()
				io.Copy(dst, src)
				fmt.Println("Extracted blip.html")
			}
		}
	}
	
	fmt.Printf("Blip serving: %s\n", serveDir)
	fmt.Printf("Open http://localhost:8080 in your browser\n")
	fmt.Printf("Or copy blip.html to your phone\n")
	
	if err := server.Start(serveDir, 8080); err != nil {
		log.Fatal("Server error:", err)
	}
}
```

✅ Success: Running blip.exe creates `blip.html` in the same directory if it doesn't exist. Subsequent runs do not overwrite. File content matches `embed/index.html`.
❌ If failed: If extraction fails, check embed directive path. If permission denied, check directory write permissions.

---

### Step 10: Initialize frontend with Vite, Alpine.js, and Tailwind

Execute from `C:\Users\rx\001_Code\105_DeadProjects\BlipSync\web`:

```bash
cd web
npm init -y
npm install alpinejs
npm install -D vite vite-plugin-singlefile tailwindcss postcss autoprefixer
npx tailwindcss init -p
```

Create `web/vite.config.js`:

```javascript
import { defineConfig } from 'vite'
import viteSingleFile from 'vite-plugin-singlefile'

export default defineConfig({
  plugins: [viteSingleFile()],
  build: {
    outDir: '../embed',
    emptyOutDir: false,
    assetsInlineLimit: 100000000,
  },
  server: {
    port: 3000,
    proxy: {
      '/api': 'http://localhost:8080',
      '/ws': {
        target: 'ws://localhost:8080',
        ws: true,
      },
    },
  },
})
```

Create `web/tailwind.config.js`:

```javascript
/** @type {import('tailwindcss').Config} */
export default {
  content: [
    './index.html',
    './src/**/*.{js,ts,jsx,tsx}',
  ],
  theme: {
    extend: {},
  },
  plugins: [],
}
```

Create `web/postcss.config.js`:

```javascript
export default {
  plugins: {
    tailwindcss: {},
    autoprefixer: {},
  },
}
```

Create `web/src/style.css`:

```css
@tailwind base;
@tailwind components;
@tailwind utilities;

/* Custom styles */
html, body {
  @apply h-full;
}
```

Create `web/index.html`:

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Blip</title>
</head>
<body class="bg-gray-900 text-gray-100 h-full" x-data="app()" x-init="init()">
  <div id="app"></div>
  <script type="module" src="/src/main.js"></script>
</body>
</html>
```

Create `web/src/main.js`:

```javascript
import Alpine from 'alpinejs'
import './style.css'

Alpine.data('app', () => ({
  connected: false,
  serverIP: '',
  status: 'searching',
  
  init() {
    this.discoverServer()
  },
  
  async discoverServer() {
    // Will be implemented in Step 13
  },
}))

Alpine.start()
```

Update `web/package.json` to add scripts:

```json
{
  "scripts": {
    "dev": "vite",
    "build": "vite build"
  }
}
```

✅ Success: Running `npm run dev` from `web` directory starts Vite dev server on port 3000. Browser shows "Blip" title. No console errors.
❌ If failed: If npm install fails, check Node.js version (18+). If Vite fails to start, check port 3000 is not in use.

---

### Step 11: Create file browser UI component

Create `web/src/components/fileBrowser.js`:

```javascript
export function fileBrowser() {
  return {
    files: [],
    currentPath: '.',
    selectedFile: null,
    preview: null,
    loading: false,
    
    async loadFiles(path = '.') {
      this.loading = true
      this.currentPath = path
      try {
        const res = await fetch(`/api/files?path=${encodeURIComponent(path)}`)
        const data = await res.json()
        this.files = data.files
      } catch (err) {
        console.error('Failed to load files:', err)
      }
      this.loading = false
    },
    
    navigateTo(path) {
      this.loadFiles(path)
      this.selectedFile = null
      this.preview = null
    },
    
    navigateUp() {
      const parts = this.currentPath.split('/')
      parts.pop()
      const newPath = parts.join('/') || '.'
      this.navigateTo(newPath)
    },
    
    async selectFile(file) {
      if (file.isDir) {
        this.navigateTo(file.path)
        return
      }
      this.selectedFile = file
      
      // Check if previewable
      const ext = file.name.split('.').pop().toLowerCase()
      const textExts = ['txt', 'md', 'json', 'log', 'js', 'ts', 'css', 'html', 'xml', 'yaml', 'yml']
      const imageExts = ['jpg', 'jpeg', 'png', 'gif', 'webp', 'svg']
      
      if (textExts.includes(ext)) {
        this.preview = { type: 'text', loading: true }
        const res = await fetch(`/api/download?path=${encodeURIComponent(file.path)}`)
        this.preview.content = await res.text()
        this.preview.loading = false
      } else if (imageExts.includes(ext)) {
        this.preview = { type: 'image', url: `/api/download?path=${encodeURIComponent(file.path)}` }
      } else {
        this.preview = { type: 'binary' }
      }
    },
    
    downloadFile() {
      if (!this.selectedFile) return
      window.open(`/api/download?path=${encodeURIComponent(this.selectedFile.path)}`)
    },
    
    async uploadFile(event) {
      const file = event.target.files[0]
      if (!file) return
      
      const formData = new FormData()
      formData.append('file', file)
      formData.append('path', this.currentPath)
      
      try {
        await fetch('/api/upload', {
          method: 'POST',
          body: formData,
        })
        this.loadFiles(this.currentPath)
      } catch (err) {
        console.error('Upload failed:', err)
      }
      
      event.target.value = ''
    },
    
    formatSize(bytes) {
      if (bytes < 1024) return bytes + ' B'
      if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
      return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
    },
  }
}
```

Update `web/src/main.js` to import and use the component:

```javascript
import Alpine from 'alpinejs'
import './style.css'
import { fileBrowser } from './components/fileBrowser.js'

Alpine.data('app', () => ({
  connected: false,
  serverIP: '',
  status: 'searching',
  showClipboard: false,
  
  init() {
    this.discoverServer()
  },
  
  async discoverServer() {
    // Step 13
  },
}))

Alpine.data('fileBrowser', fileBrowser)

Alpine.start()
```

Update `web/index.html` with file browser UI:

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Blip</title>
</head>
<body class="bg-gray-900 text-gray-100 h-full flex flex-col" x-data="app()" x-init="init()">
  
  <!-- Header -->
  <header class="bg-gray-800 px-4 py-3 flex items-center justify-between">
    <h1 class="text-xl font-bold">Blip</h1>
    <div class="flex items-center gap-2">
      <span x-show="connected" class="text-green-400">● Connected</span>
      <span x-show="!connected" class="text-yellow-400">● Searching...</span>
    </div>
  </header>
  
  <!-- Main Content -->
  <main class="flex-1 overflow-hidden flex flex-col" x-data="fileBrowser()" x-init="loadFiles()">
    
    <!-- Path breadcrumb -->
    <div class="bg-gray-800 px-4 py-2 flex items-center gap-2 text-sm">
      <button @click="navigateUp()" class="text-blue-400 hover:text-blue-300">..</button>
      <span x-text="currentPath" class="text-gray-400"></span>
    </div>
    
    <!-- File list -->
    <div class="flex-1 overflow-y-auto">
      <template x-for="file in files" :key="file.path">
        <div 
          @click="selectFile(file)"
          class="px-4 py-3 flex items-center gap-3 hover:bg-gray-800 cursor-pointer border-b border-gray-700"
          :class="{ 'bg-gray-800': selectedFile?.path === file.path }"
        >
          <span x-show="file.isDir" class="text-2xl">📁</span>
          <span x-show="!file.isDir" class="text-2xl">📄</span>
          <div class="flex-1">
            <div x-text="file.name" class="font-medium"></div>
            <div x-show="!file.isDir" x-text="formatSize(file.size)" class="text-xs text-gray-500"></div>
          </div>
        </div>
      </template>
    </div>
    
    <!-- Preview panel -->
    <div x-show="preview" class="bg-gray-800 border-t border-gray-700 p-4 max-h-64 overflow-y-auto">
      <template x-if="preview?.type === 'text'">
        <pre x-text="preview.content" class="text-sm whitespace-pre-wrap"></pre>
      </template>
      <template x-if="preview?.type === 'image'">
        <img :src="preview.url" class="max-w-full max-h-48 mx-auto">
      </template>
      <template x-if="preview?.type === 'binary'">
        <div class="text-center text-gray-400">
          <p>Binary file</p>
          <button @click="downloadFile()" class="mt-2 bg-blue-600 px-4 py-2 rounded">Download</button>
        </div>
      </template>
    </div>
    
    <!-- Bottom actions -->
    <div class="bg-gray-800 p-4 flex gap-2">
      <label class="flex-1 bg-blue-600 text-center py-3 rounded cursor-pointer">
        Upload
        <input type="file" @change="uploadFile($event)" class="hidden">
      </label>
      <button @click="$dispatch('toggle-clipboard')" class="px-4 py-3 bg-gray-700 rounded">
        Clipboard
      </button>
    </div>
  </main>
  
  <!-- Clipboard Panel (Step 12) -->
  <div x-show="showClipboard" class="fixed inset-0 bg-gray-900 flex flex-col">
    <!-- Will be added in Step 12 -->
  </div>
  
  <script type="module" src="/src/main.js"></script>
</body>
</html>
```

✅ Success: Running `npm run dev` shows file list. Clicking folders navigates into them. Clicking files shows preview or download button. Upload button opens file picker. All interactions work without console errors.
❌ If failed: If files don't load, check `/api/files` endpoint is accessible. If UI doesn't render, check Alpine.js is loaded correctly.

---

### Step 12: Create clipboard UI component

Create `web/src/components/clipboard.js`:

```javascript
export function clipboardPanel() {
  return {
    text: '',
    history: [],
    ws: null,
    
    init() {
      this.connectWebSocket()
    },
    
    connectWebSocket() {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const wsUrl = `${protocol}//${window.location.host}/ws`
      
      this.ws = new WebSocket(wsUrl)
      
      this.ws.onopen = () => {
        console.log('WebSocket connected')
      }
      
      this.ws.onmessage = (event) => {
        try {
          const msg = JSON.parse(event.data)
          if (msg.type === 'clipboard') {
            this.addToHistory(msg.data)
          } else if (msg.type === 'refresh') {
            this.$dispatch('files-refresh')
          }
        } catch (err) {
          console.error('Failed to parse message:', err)
        }
      }
      
      this.ws.onclose = () => {
        console.log('WebSocket disconnected, reconnecting...')
        setTimeout(() => this.connectWebSocket(), 3000)
      }
      
      this.ws.onerror = (err) => {
        console.error('WebSocket error:', err)
      }
    },
    
    send() {
      if (!this.text.trim()) return
      
      const msg = JSON.stringify({
        type: 'clipboard',
        data: this.text,
      })
      
      this.ws.send(msg)
      this.addToHistory(this.text)
      this.text = ''
    },
    
    addToHistory(text) {
      // Add to beginning
      this.history.unshift({
        text: text,
        time: new Date().toLocaleTimeString(),
      })
      
      // Keep only last 10
      if (this.history.length > 10) {
        this.history.pop()
      }
    },
    
    copyToClipboard(text) {
      navigator.clipboard.writeText(text)
    },
    
    useHistoryItem(item) {
      this.text = item.text
    },
  }
}
```

Update `web/src/main.js` to import clipboard:

```javascript
import Alpine from 'alpinejs'
import './style.css'
import { fileBrowser } from './components/fileBrowser.js'
import { clipboardPanel } from './components/clipboard.js'

// ... existing code ...

Alpine.data('fileBrowser', fileBrowser)
Alpine.data('clipboardPanel', clipboardPanel)

Alpine.start()
```

Update `web/index.html` to add clipboard panel:

```html
<!-- Replace the clipboard panel placeholder in index.html -->

<!-- Clipboard Panel -->
<div 
  x-show="showClipboard" 
  x-transition
  class="fixed inset-0 bg-gray-900 flex flex-col z-50"
  x-data="clipboardPanel()"
  x-init="init()"
>
  <header class="bg-gray-800 px-4 py-3 flex items-center justify-between">
    <h2 class="text-xl font-bold">Clipboard</h2>
    <button @click="$dispatch('toggle-clipboard')" class="text-2xl">&times;</button>
  </header>
  
  <div class="flex-1 overflow-y-auto p-4">
    <!-- Text input -->
    <textarea 
      x-model="text" 
      placeholder="Type or paste text..."
      class="w-full h-32 bg-gray-800 rounded p-3 text-white resize-none"
    ></textarea>
    
    <button 
      @click="send()"
      class="w-full mt-3 bg-blue-600 py-3 rounded font-medium"
    >
      Send to Desktop
    </button>
    
    <!-- History -->
    <h3 class="mt-6 mb-2 text-gray-400 text-sm">History</h3>
    <template x-for="(item, i) in history" :key="i">
      <div class="bg-gray-800 rounded p-3 mb-2">
        <div x-text="item.text" class="text-sm truncate"></div>
        <div class="flex justify-between items-center mt-2">
          <span x-text="item.time" class="text-xs text-gray-500"></span>
          <div class="flex gap-2">
            <button @click="useHistoryItem(item)" class="text-xs text-blue-400">Use</button>
            <button @click="copyToClipboard(item.text)" class="text-xs text-green-400">Copy</button>
          </div>
        </div>
      </div>
    </template>
  </div>
</div>
```

Update `web/src/main.js` app data to handle clipboard toggle:

```javascript
Alpine.data('app', () => ({
  connected: false,
  serverIP: '',
  status: 'searching',
  showClipboard: false,
  
  init() {
    this.discoverServer()
    
    // Listen for clipboard toggle
    window.addEventListener('toggle-clipboard', () => {
      this.showClipboard = !this.showClipboard
    })
    
    // Listen for file refresh
    window.addEventListener('files-refresh', () => {
      // Dispatch to fileBrowser
      window.dispatchEvent(new CustomEvent('refresh-files'))
    })
  },
  
  async discoverServer() {
    // Step 13
  },
}))
```

Update file browser to listen for refresh:

```javascript
// In fileBrowser function, add to init or after:
init() {
  this.loadFiles()
  
  window.addEventListener('refresh-files', () => {
    this.loadFiles(this.currentPath)
  })
},
```

✅ Success: Clicking "Clipboard" button opens clipboard panel. Typing text and clicking "Send" broadcasts to all connected clients. History shows last 10 items. Copy button copies to device clipboard.
❌ If failed: If WebSocket fails to connect, check `/ws` endpoint. If messages not received, check server-side hub broadcast logic.

---

### Step 13: Implement auto-discovery client-side logic

Update `web/src/components/discovery.js`:

```javascript
export function discovery() {
  return {
    status: 'searching', // searching, connected, failed
    serverIP: '',
    
    async findServer() {
      // Get current IP to determine subnet
      const subnet = await this.detectSubnet()
      
      if (!subnet) {
        this.status = 'failed'
        return null
      }
      
      // Scan subnet
      const promises = []
      for (let i = 1; i <= 254; i++) {
        promises.push(this.pingServer(`${subnet}.${i}`))
      }
      
      const results = await Promise.allSettled(promises)
      const found = results.find(r => r.status === 'fulfilled' && r.value)
      
      if (found) {
        this.serverIP = found.value
        this.status = 'connected'
        return found.value
      }
      
      this.status = 'failed'
      return null
    },
    
    async detectSubnet() {
      // Try to detect by pinging common servers or using WebRTC
      // For simplicity, we'll try common subnets
      const commonSubnets = ['192.168.1', '192.168.0', '10.0.0', '192.168.2']
      
      for (const subnet of commonSubnets) {
        // Try a few IPs quickly
        for (let i = 1; i <= 5; i++) {
          const found = await this.pingServer(`${subnet}.${i}`)
          if (found) {
            return subnet
          }
        }
      }
      
      // Default to most common
      return '192.168.1'
    },
    
    async pingServer(ip) {
      const controller = new AbortController()
      const timeout = setTimeout(() => controller.abort(), 500)
      
      try {
        const response = await fetch(`http://${ip}:8080/ping`, {
          method: 'GET',
          signal: controller.signal,
        })
        clearTimeout(timeout)
        
        if (response.ok) {
          return ip
        }
        return null
      } catch {
        clearTimeout(timeout)
        return null
      }
    },
    
    useManualIP(ip) {
      this.serverIP = ip
      this.status = 'connected'
    },
  }
}
```

Update `web/src/main.js`:

```javascript
import Alpine from 'alpinejs'
import './style.css'
import { fileBrowser } from './components/fileBrowser.js'
import { clipboardPanel } from './components/clipboard.js'
import { discovery } from './components/discovery.js'

Alpine.data('app', () => ({
  connected: false,
  serverIP: '',
  status: 'searching',
  showClipboard: false,
  manualIP: '',
  
  init() {
    this.discoverServer()
    
    window.addEventListener('toggle-clipboard', () => {
      this.showClipboard = !this.showClipboard
    })
    
    window.addEventListener('files-refresh', () => {
      window.dispatchEvent(new CustomEvent('refresh-files'))
    })
  },
  
  async discoverServer() {
    const disco = discovery()
    const ip = await disco.findServer()
    
    if (ip) {
      this.serverIP = ip
      this.connected = true
      this.status = 'connected'
      // Redirect or set API base URL
      window.API_BASE = `http://${ip}:8080`
    } else {
      this.status = 'failed'
    }
  },
  
  async connectManual() {
    if (!this.manualIP) return
    
    const disco = discovery()
    const found = await disco.pingServer(this.manualIP)
    
    if (found) {
      this.serverIP = this.manualIP
      this.connected = true
      this.status = 'connected'
      window.API_BASE = `http://${this.manualIP}:8080`
    }
  },
}))

Alpine.data('fileBrowser', fileBrowser)
Alpine.data('clipboardPanel', clipboardPanel)

Alpine.start()
```

Update `web/index.html` to show discovery status and manual entry:

```html
<!-- Add after header, before main content -->

<!-- Discovery overlay -->
<div 
  x-show="status === 'searching' || status === 'failed'" 
  class="fixed inset-0 bg-gray-900 flex items-center justify-center z-50"
>
  <div class="text-center p-8">
    <template x-if="status === 'searching'">
      <div>
        <div class="animate-spin text-4xl mb-4">⟳</div>
        <p class="text-xl">Searching for server...</p>
        <p class="text-gray-400 mt-2">Make sure blip.exe is running on your computer</p>
      </div>
    </template>
    
    <template x-if="status === 'failed'">
      <div>
        <div class="text-4xl mb-4">⚠️</div>
        <p class="text-xl mb-4">Server not found</p>
        <div class="bg-gray-800 rounded p-4">
          <p class="text-sm text-gray-400 mb-2">Enter server IP manually:</p>
          <div class="flex gap-2">
            <input 
              type="text" 
              x-model="manualIP" 
              placeholder="192.168.1.42"
              class="flex-1 bg-gray-700 rounded px-3 py-2 text-white"
            >
            <button 
              @click="connectManual()"
              class="bg-blue-600 px-4 py-2 rounded"
            >
              Connect
            </button>
          </div>
        </div>
      </div>
    </template>
  </div>
</div>
```

✅ Success: Opening blip.html on phone shows "Searching for server..." then connects when found. If server not found within ~30 seconds, shows manual IP entry. Manual IP entry works and connects to server.
❌ If failed: If discovery always fails, check phone and computer are on same WiFi. Check server is running and accessible. Check browser doesn't block mixed content.

---

### Step 14: Build frontend to single HTML file

Execute from `web` directory:

```bash
cd web
npm run build
```

Verify `embed/index.html` exists and contains all CSS/JS inlined.

✅ Success: `embed/index.html` exists, is a single file (no external CSS/JS links), and contains all app code. File size is reasonable (under 200KB).
❌ If failed: If build fails, check vite-plugin-singlefile is installed. If file has external links, check Vite config.

---

### Step 15: Integrate frontend build into Go binary

Update `cmd/blip/main.go` to ensure embed path is correct:

```go
// The embed directive should be at package level
//go:embed ../../embed/index.html
var staticFiles embed.FS
```

Update `internal/server/server.go` embed directive:

```go
//go:embed ../../embed/index.html
var staticFiles embed.FS
```

Build the Go binary:

```bash
go build -o blip.exe ./cmd/blip
```

✅ Success: `blip.exe` is created in project root. Running it starts server. Accessing `http://localhost:8080` shows the full UI. `blip.html` is extracted to the same directory.
❌ If failed: If build fails with embed error, verify `embed/index.html` exists. If UI doesn't load, check serveIndex function serves correct file.

---

### Step 16: End-to-end testing and verification

Test the complete flow:

1. Run `blip.exe` in a test folder with some files
2. Verify server starts and prints IP address
3. Open `http://localhost:8080` in desktop browser
4. Verify file list shows test files
5. Open `blip.html` on mobile device (same WiFi)
6. Verify auto-discovery finds the server
7. Browse files on mobile
8. Upload a file from mobile
9. Verify file appears on desktop
10. Test clipboard sync: type on mobile, verify appears on desktop
11. Test clipboard sync: type on desktop, verify appears on mobile
12. Edit a file in the folder, verify mobile refreshes

✅ Success: All 12 test steps pass. No errors in browser console or server logs. File operations work bidirectionally. Clipboard sync is instant (< 500ms).
❌ If failed: Document which step failed and what error was observed.

---

## Verification

Run the following commands and verify outputs:

```bash
# 1. Build succeeds
go build -o blip.exe ./cmd/blip
# Expected: No errors, blip.exe created

# 2. Server starts
./blip.exe
# Expected: "Blip serving: [path]", "Open http://localhost:8080"

# 3. API endpoints work
curl http://localhost:8080/ping
# Expected: "pong"

curl http://localhost:8080/api/files
# Expected: JSON array of files

# 4. WebSocket connects (use wscat or browser)
# Expected: Connection established, no errors

# 5. Frontend build succeeds
cd web && npm run build
# Expected: dist/index.html created

# 6. Binary size reasonable
ls -lh blip.exe
# Expected: Under 20MB
```

---

## Rollback

If a critical step fails and cannot be recovered:

```bash
# Remove generated files
rm -rf web/node_modules
rm -rf embed
rm blip.exe

# Revert to clean state
git checkout -- .

# Re-run from Step 1
```

---

## Notes

- The frontend development server (`npm run dev`) proxies API requests to the Go backend, enabling hot-reload during development
- The `embed/index.html` is checked into git so Go builds have no npm dependency
- Discovery uses parallel HTTP requests with 500ms timeout each, completing in 2-3 seconds worst case
- WebSocket reconnection is automatic with 3-second delay

---

**Plan complete. Handing off to Executor.**

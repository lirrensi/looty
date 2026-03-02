# Plan: Web UI Polish - Round 2
_Fix breadcrumbs, refresh feedback, clipboard→scratchpad, sorting, file icons, and add shared scratchpad with history._

---

# Checklist
- [x] Step 1: Create icons module with base shapes and color map
- [x] Step 2: Update breadcrumb styling to big chunky buttons
- [x] Step 3: Add refresh button visual feedback (spin animation)
- [x] Step 4: Add 4-option sorting UI
- [x] Step 5: Add scratchpad in-memory storage to backend
- [x] Step 6: Add scratchpad API endpoints
- [x] Step 7: Update frontend scratchpad component for real-time sync
- [x] Step 8: Replace emoji with auto-generated per-file icons
- [x] Step 9: Make each breadcrumb its own scrollable mini-breadcrumb
- [x] Step 10: Fix Scratchpad button to work with shared history
- [x] Step 11: Change file position background from purple to neutral

---

## Context

Current web UI is in `web/` directory. Uses:
- Vite for build
- Alpine.js for reactivity
- Tailwind CSS for styling

Main files:
- `web/index.html` — Main UI template with Alpine.js directives
- `web/src/main.js` — Alpine app initialization
- `web/src/components/fileBrowser.js` — File list, navigation, sorting
- `web/src/components/clipboard.js` — Current clipboard panel (incomplete)
- `web/src/style.css` — Custom styles

Backend:
- `internal/server/server.go` — HTTP server, WebSocket handler
- `internal/clipboard/clipboard.go` — Message types for WebSocket

Current sorting cycles 2 modes. Breadcrumbs are subtle text links. Refresh button has no feedback. Clipboard only broadcasts to other clients, no persistence. Emoji icons for all files (no per-type distinction).

---

## Prerequisites

- Node.js and pnpm installed
- Go toolchain installed
- `web/` directory has `node_modules` installed (run `pnpm install` in `web/` if missing)

---

## Scope Boundaries

OUT OF SCOPE:
- `internal/files/` — File handling logic unchanged
- `internal/server/watcher.go` — File watching unchanged
- `embed/` directory
- `looty.html` — Old test file
- Any backend clipboard-to-system-clipboard integration
- Installing external icon packages (lucide, simple-icons, etc.)

---

## Steps

### Step 1: Create icons module with base shapes and color map

Create file `web/src/icons.js`. This module provides:
1. Base SVG shapes (document, image, video, audio, archive, folder)
2. Color map for ~100 common extensions
3. Auto-generated badges with extension text and color

```javascript
// Extension to color mapping - distinct colors for quick navigation
const extensionColors = {
  // === OFFICE ===
  doc: '#2b579a', docx: '#2b579a',
  xls: '#217346', xlsx: '#217346',
  ppt: '#d24726', pptx: '#d24726',
  pdf: '#f40f02',
  
  // === ADOBE SUITE ===
  psd: '#001d34', ai: '#330000',
  ae: '#9999ff', pr: '#9933ff',
  indd: '#ff3366', xd: '#ff61f6',
  fig: '#f24e1e', sketch: '#fdad00',
  
  // === CODE - JAVASCRIPT/WEB ===
  js: '#f7df1e', mjs: '#f7df1e', cjs: '#f7df1e',
  ts: '#3178c6', tsx: '#3178c6',
  jsx: '#61dafb',
  vue: '#42b883', svelte: '#ff3e00',
  angular: '#dd0031',
  
  // === CODE - BACKEND ===
  go: '#00add8',
  rs: '#dea584',
  py: '#3776ab',
  rb: '#cc342d',
  php: '#777bb4',
  java: '#b07219',
  kt: '#7f52ff',
  swift: '#f05138',
  c: '#555555', h: '#555555',
  cpp: '#f34b7d', hpp: '#f34b7d', cc: '#f34b7d',
  cs: '#178600',
  lua: '#000080',
  scala: '#dc322f',
  clj: '#5881d8',
  ex: '#6e4a7e', exs: '#6e4a7e',
  hs: '#5e5086',
  el: '#4053a5',
  nim: '#ffc200',
  dart: '#0175c2',
  
  // === CODE - SYSTEM ===
  sh: '#4eaa25', bash: '#4eaa25', zsh: '#4eaa25',
  fish: '#34b000',
  ps1: '#012456', psm1: '#012456',
  bat: '#c1f12e', cmd: '#c1f12e',
  
  // === WEB ===
  html: '#e34c26', htm: '#e34c26',
  css: '#264de4',
  scss: '#cd6799', sass: '#cd6799', less: '#1d365d',
  
  // === DATA/CONFIG ===
  json: '#cbcb41',
  yaml: '#cb171e', yml: '#cb171e',
  xml: '#0060ac',
  toml: '#9c4121',
  sql: '#e38c00',
  graphql: '#e535ab', gql: '#e535ab',
  proto: '#c128c9',
  
  // === CONFIG ===
  env: '#ecd53f', gitignore: '#f14e32', dockerignore: '#2496ed',
  ini: '#6d8086', cfg: '#6d8086', conf: '#6d8086',
  
  // === MARKUP/DOCS ===
  md: '#083fa1', markdown: '#083fa1',
  rst: '#141414',
  tex: '#3d6117',
  rtf: '#7a8b8b',
  
  // === IMAGES ===
  png: '#a074c4', jpg: '#a074c4', jpeg: '#a074c4',
  gif: '#a074c4', webp: '#a074c4', bmp: '#a074c4',
  ico: '#a074c4', svg: '#ffb13b', avif: '#a074c4',
  heic: '#a074c4', heif: '#a074c4',
  tiff: '#a074c4', tif: '#a074c4',
  
  // === VIDEO ===
  mp4: '#ff5c5c', mov: '#ff5c5c', avi: '#ff5c5c',
  mkv: '#ff5c5c', wmv: '#ff5c5c', flv: '#ff5c5c',
  webm: '#ff5c5c', m4v: '#ff5c5c', mpeg: '#ff5c5c',
  
  // === AUDIO ===
  mp3: '#a855f7', wav: '#a855f7', ogg: '#a855f7',
  flac: '#a855f7', aac: '#a855f7', m4a: '#a855f7',
  wma: '#a855f7', aiff: '#a855f7',
  
  // === ARCHIVES ===
  zip: '#f97316', tar: '#f97316', gz: '#f97316',
  bz2: '#f97316', xz: '#f97316', '7z': '#f97316',
  rar: '#f97316', tgz: '#f97316',
  
  // === FONTS ===
  ttf: '#3b82f6', otf: '#3b82f6', woff: '#3b82f6',
  woff2: '#3b82f6', eot: '#3b82f6',
  
  // === DATABASE ===
  db: '#003b57', sqlite: '#003b57', sqlite3: '#003b57',
  prisma: '#2d3748',
  
  // === DEVOPS ===
  dockerfile: '#2496ed', docker: '#2496ed',
  kube: '#326ce5', kubernetes: '#326ce5',
  tf: '#7b42bc', hcl: '#7b42bc',
  
  // === EXECUTABLES ===
  exe: '#6b7280', msi: '#6b7280', dmg: '#6b7280',
  app: '#6b7280', deb: '#6b7280', rpm: '#6b7280',
  apk: '#3ddb93', ipa: '#007aff',
  
  // === OTHER COMMON ===
  torrent: '#567a46', iso: '#6b7280', img: '#6b7280',
  bin: '#6b7280', log: '#6b7280', txt: '#6b7280',
  csv: '#217346',
  lock: '#8b8b8b',
  wasm: '#654ff0',
  sol: '#363636',
}

// Shape types - determines which base SVG to use
const imageExtensions = ['png', 'jpg', 'jpeg', 'gif', 'webp', 'bmp', 'ico', 'svg', 'avif', 'heic', 'heif', 'tiff', 'tif']
const videoExtensions = ['mp4', 'mov', 'avi', 'mkv', 'wmv', 'flv', 'webm', 'm4v', 'mpeg']
const audioExtensions = ['mp3', 'wav', 'ogg', 'flac', 'aac', 'm4a', 'wma', 'aiff']
const archiveExtensions = ['zip', 'tar', 'gz', 'bz2', 'xz', '7z', 'rar', 'tgz']
const fontExtensions = ['ttf', 'otf', 'woff', 'woff2', 'eot']

function getShapeType(ext) {
  if (imageExtensions.includes(ext)) return 'image'
  if (videoExtensions.includes(ext)) return 'video'
  if (audioExtensions.includes(ext)) return 'audio'
  if (archiveExtensions.includes(ext)) return 'archive'
  if (fontExtensions.includes(ext)) return 'font'
  return 'document'
}

// Base SVG shapes
const shapes = {
  folder: (color) => `
    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="${color}" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
      <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z" fill="${color}" fill-opacity="0.15"/>
      <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
    </svg>`,
  
  document: (color) => `
    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="${color}" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
      <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" fill="${color}" fill-opacity="0.1"/>
      <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
      <polyline points="14 2 14 8 20 8"/>
      <line x1="16" y1="13" x2="8" y2="13"/>
      <line x1="16" y1="17" x2="8" y2="17"/>
      <line x1="10" y1="9" x2="8" y2="9"/>
    </svg>`,
  
  image: (color) => `
    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="${color}" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
      <rect x="3" y="3" width="18" height="18" rx="2" ry="2" fill="${color}" fill-opacity="0.1"/>
      <rect x="3" y="3" width="18" height="18" rx="2" ry="2"/>
      <circle cx="8.5" cy="8.5" r="1.5"/>
      <polyline points="21 15 16 10 5 21"/>
    </svg>`,
  
  video: (color) => `
    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="${color}" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
      <rect x="2" y="4" width="20" height="16" rx="2" fill="${color}" fill-opacity="0.1"/>
      <rect x="2" y="4" width="20" height="16" rx="2"/>
      <polygon points="10 9 15 12 10 15 10 9" fill="${color}"/>
    </svg>`,
  
  audio: (color) => `
    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="${color}" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
      <path d="M9 18V5l12-2v13" fill="${color}" fill-opacity="0.1"/>
      <circle cx="6" cy="18" r="3" fill="${color}" fill-opacity="0.2"/>
      <circle cx="18" cy="16" r="3" fill="${color}" fill-opacity="0.2"/>
      <path d="M9 18V5l12-2v13"/>
      <circle cx="6" cy="18" r="3"/>
      <circle cx="18" cy="16" r="3"/>
    </svg>`,
  
  archive: (color) => `
    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="${color}" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
      <path d="M21 8v13H3V8" fill="${color}" fill-opacity="0.1"/>
      <path d="M1 3h22v5H1z"/>
      <path d="M10 12h4"/>
      <path d="M21 8v13H3V8"/>
    </svg>`,
  
  font: (color) => `
    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="${color}" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
      <polyline points="4 7 4 4 20 4 20 7" fill="${color}" fill-opacity="0.1"/>
      <line x1="9" y1="20" x2="15" y2="20"/>
      <line x1="12" y1="4" x2="12" y2="20"/>
      <polyline points="4 7 4 4 20 4 20 7"/>
    </svg>`,
}

// Create badge overlay for extension
function createBadge(ext, color, size = 24) {
  const badgeWidth = size * 0.5
  const badgeHeight = size * 0.28
  const x = size - badgeWidth - 1
  const y = size - badgeHeight - 2
  const fontSize = size * 0.17
  const text = ext.toUpperCase().slice(0, 4)
  
  return `
    <g transform="translate(${x}, ${y})">
      <rect width="${badgeWidth}" height="${badgeHeight}" rx="2" fill="${color}"/>
      <text x="${badgeWidth/2}" y="${badgeHeight * 0.7}" 
            text-anchor="middle" 
            fill="white" 
            font-family="system-ui, -apple-system, sans-serif" 
            font-size="${fontSize}" 
            font-weight="600"
            style="text-shadow: 0 1px 1px rgba(0,0,0,0.3)">${text}</text>
    </g>`
}

// Main file icon function
export function fileIcon(filename, isDir = false, size = 24) {
  if (isDir) {
    return shapes.folder('#eab308') // yellow-500
  }
  
  const ext = filename.split('.').pop()?.toLowerCase() || ''
  const baseName = filename.toLowerCase()
  
  // Special case: hidden files start with dot, show dimmed
  if (filename.startsWith('.') && ext === filename.slice(1)) {
    const color = '#6b7280'
    return shapes.document(color)
  }
  
  // Get color for extension
  const color = extensionColors[ext] || '#6b7280'
  
  // Get shape type
  const shapeType = getShapeType(ext)
  
  // Generate SVG
  const baseShape = shapes[shapeType](color)
  
  // Add badge for documents (not for image/video/audio which have distinctive shapes)
  if (shapeType === 'document' && ext) {
    const badge = createBadge(ext, color, size)
    // Insert badge before closing SVG tag
    return baseShape.replace('</svg>', badge + '</svg>')
  }
  
  return baseShape
}

// UI icons for buttons (simple stroke icons)
export function icon(name, size = 20) {
  const icons = {
    upload: `<svg xmlns="http://www.w3.org/2000/svg" width="${size}" height="${size}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>`,
    refresh: `<svg xmlns="http://www.w3.org/2000/svg" width="${size}" height="${size}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="23 4 23 10 17 10"/><polyline points="1 20 1 14 7 14"/><path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/></svg>`,
    clipboard: `<svg xmlns="http://www.w3.org/2000/svg" width="${size}" height="${size}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2"/><rect x="8" y="2" width="8" height="4" rx="1" ry="1"/></svg>`,
    close: `<svg xmlns="http://www.w3.org/2000/svg" width="${size}" height="${size}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>`,
    chevronRight: `<svg xmlns="http://www.w3.org/2000/svg" width="${size}" height="${size}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="9 18 15 12 9 6"/></svg>`,
    arrowUp: `<svg xmlns="http://www.w3.org/2000/svg" width="${size}" height="${size}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="19" x2="12" y2="5"/><polyline points="5 12 12 5 19 12"/></svg>`,
    home: `<svg xmlns="http://www.w3.org/2000/svg" width="${size}" height="${size}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/></svg>`,
  }
  
  return icons[name] || ''
}

// Make available globally for Alpine templates
window.fileIcon = fileIcon
window.icon = icon
```

✅ Success: File `web/src/icons.js` exists with base shapes and color map for ~100+ extensions.
❌ If failed: Stop. Report error.

---

### Step 2: Update breadcrumb styling to big chunky buttons

Open `web/index.html`. Find the breadcrumb section (approximately lines 87-101). 

Replace the entire breadcrumb `<div>` block with:

```html
<!-- Breadcrumbs -->
<div class="flex items-center gap-1 text-sm overflow-x-auto flex-1">
  <template x-for="(crumb, i) in breadcrumbPath" :key="crumb.path">
    <div class="flex items-center">
      <template x-if="i > 0">
        <span class="text-gray-500 mx-1" x-html="icon('chevronRight', 14)"></span>
      </template>
      <button 
        @click="navigateTo(crumb.path)"
        class="px-3 py-1.5 rounded-lg text-sm font-medium whitespace-nowrap transition-all"
        :class="i === breadcrumbPath.length - 1 
          ? 'bg-purple-600 text-white' 
          : 'bg-gray-700 text-gray-200 hover:bg-gray-600'"
        x-text="crumb.name"
      ></button>
    </div>
  </template>
</div>
```

Also update the "Up" button (approximately lines 76-84) to be bigger:

```html
<button 
  x-show="currentPath !== '.'"
  @click="navigateUp()"
  class="bg-gray-700 hover:bg-gray-600 text-white flex items-center gap-1 text-sm px-3 py-1.5 rounded-lg transition-all"
  title="Go up one level"
>
  <span x-html="icon('arrowUp', 16)"></span>
  <span>Up</span>
</button>
```

✅ Success: Breadcrumb buttons appear as big pill-shaped buttons. Last crumb has purple background.
❌ If failed: Stop. Report what looks wrong.

---

### Step 3: Add refresh button visual feedback (spin animation)

Open `web/src/components/fileBrowser.js`. Add a `refreshing` state variable after `loading: false` (approximately line 9):

```javascript
refreshing: false,
```

Update the `loadFiles` function to set refreshing state. Find the `loadFiles` function (approximately line 62). Replace it with:

```javascript
async loadFiles(path = '.') {
  this.loading = true
  this.refreshing = true
  this.currentPath = path
  this.selectedFile = null
  this.preview = null
  this.uploadSuccess = false
  try {
    const res = await fetch(apiUrl(`/api/files?path=${encodeURIComponent(path)}`))
    const data = await res.json()
    this.files = data.files || []
  } catch (err) {
    console.error('Failed to load files:', err)
    this.files = []
  } finally {
    this.loading = false
    this.refreshing = false
  }
}
```

Remove the old `this.loading = false` at the end if it's still there after the try-catch.

Open `web/index.html`. Find the refresh button (approximately line 212-214). Replace with:

```html
<button 
  @click="loadFiles(currentPath)" 
  class="px-3 py-2 bg-gray-700 hover:bg-gray-600 rounded-lg text-sm transition-all disabled:opacity-50"
  :class="refreshing ? 'animate-spin' : ''"
  :disabled="refreshing"
  title="Refresh"
  x-html="icon('refresh', 18)"
></button>
```

Add CSS animation for spin in `web/src/style.css`:

```css
@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
.animate-spin {
  animation: spin 1s linear infinite;
}
```

✅ Success: Refresh button shows spinning icon while loading files. Stops when done.
❌ If failed: Stop. Report what happens instead.

---

### Step 4: Add 4-option sorting UI

Open `web/src/components/fileBrowser.js`. Replace the `sortBy: 'name'` and `sortAsc: true` declarations and `toggleSort` function (approximately lines 13-14 and 213-221) with:

```javascript
sortModes: ['name-asc', 'name-desc', 'date-desc', 'date-asc'],
sortModeIndex: 0,

get sortMode() {
  return this.sortModes[this.sortModeIndex]
},

get sortBy() {
  return this.sortMode.split('-')[0]
},

get sortAsc() {
  return this.sortMode.split('-')[1] === 'asc'
},

cycleSort() {
  this.sortModeIndex = (this.sortModeIndex + 1) % this.sortModes.length
},

setSort(mode) {
  this.sortModeIndex = this.sortModes.indexOf(mode)
},
```

Open `web/index.html`. Find the sort bar (approximately lines 105-112). Replace with:

```html
<!-- Sort bar -->
<div class="bg-gray-800 px-4 py-2 flex items-center justify-between text-xs border-t border-gray-700">
  <div class="flex gap-1">
    <template x-for="mode in sortModes" :key="mode">
      <button 
        @click="setSort(mode)"
        class="px-2 py-1 rounded transition-all flex items-center gap-1"
        :class="sortMode === mode ? 'bg-purple-600 text-white' : 'bg-gray-700 text-gray-400 hover:text-white'"
      >
        <template x-if="mode === 'name-asc'">
          <span>A→Z</span>
        </template>
        <template x-if="mode === 'name-desc'">
          <span>Z→A</span>
        </template>
        <template x-if="mode === 'date-desc'">
          <span>Newest</span>
        </template>
        <template x-if="mode === 'date-asc'">
          <span>Oldest</span>
        </template>
      </button>
    </template>
  </div>
  <span class="text-gray-500" x-text="files.length + ' items'"></span>
</div>
```

✅ Success: Four sort buttons appear (A→Z, Z→A, Newest, Oldest). Clicking each changes sort order. Active button is purple.
❌ If failed: Stop. Report what appears instead.

---

### Step 5: Add scratchpad in-memory storage to backend

Open `internal/server/server.go`. Add imports and global variables after the existing imports (approximately line 3-12):

```go
import (
	"encoding/json"
	"embed"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/user/looty/internal/clipboard"
	"github.com/user/looty/internal/files"
)
```

Add scratchpad variables after the `hub` declaration (approximately line 33):

```go
var (
	hub          *Hub
	scratchpadMu sync.RWMutex
	scratchpad   string
)

func GetScratchpad() string {
	scratchpadMu.RLock()
	defer scratchpadMu.RUnlock()
	return scratchpad
}

func SetScratchpad(content string) {
	scratchpadMu.Lock()
	scratchpad = content
	scratchpadMu.Unlock()
}
```

✅ Success: Code compiles with `go build ./...` from project root.
❌ If failed: Stop. Report compiler error.

---

### Step 6: Add scratchpad API endpoints

Open `internal/server/server.go`. Add handler functions after the `handlePing` function (approximately line 107):

```go
func (s *Server) handleGetScratchpad(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"content": GetScratchpad()})
}

func (s *Server) handleSetScratchpad(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	SetScratchpad(body.Content)
	
	// Broadcast to all connected clients
	msg := clipboard.NewScratchpadMessage(body.Content)
	hub.broadcast <- msg
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
```

Register routes in the `Start` function. Find the line `mux.HandleFunc("/ws", withCORS(s.handleWebSocket))` (approximately line 81). Add after it:

```go
// Scratchpad API endpoints
mux.HandleFunc("/api/scratchpad", withCORS(func(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.handleGetScratchpad(w, r)
	case "POST":
		s.handleSetScratchpad(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}))
```

Open `internal/clipboard/clipboard.go`. Add the new message type. Find the const block (approximately line 9-12). Replace with:

```go
const (
	TypeClipboard  MessageType = "clipboard"
	TypeRefresh    MessageType = "refresh"
	TypeScratchpad MessageType = "scratchpad"
)
```

Add the new message factory function after `NewRefreshMessage` (approximately line 43):

```go
func NewScratchpadMessage(content string) []byte {
	msg := Message{
		Type: TypeScratchpad,
		Data: content,
	}
	data, _ := json.Marshal(msg)
	return data
}
```

Update `internal/server/server.go` WebSocket handler. Find the `handleWebSocket` function and update the message handler (approximately lines 121-133). Replace with:

```go
go client.readPump(hub, func(msg []byte) {
	message, err := clipboard.ParseMessage(msg)
	if err != nil {
		log.Printf("Failed to parse message: %v", err)
		return
	}

	if message.Type == clipboard.TypeClipboard || message.Type == clipboard.TypeScratchpad {
		hub.broadcast <- msg
	}
})
```

✅ Success: `go build ./...` compiles. Server starts without errors.
❌ If failed: Stop. Report compiler error or runtime panic.

---

### Step 7: Update frontend scratchpad component for real-time sync

Open `web/src/components/clipboard.js`. Replace entire file contents with:

```javascript
import { apiUrl, wsUrl } from '../utils.js'

export function clipboardPanel() {
  return {
    content: '',
    history: [],
    ws: null,
    copySuccess: false,
    sendSuccess: false,
    wsConnected: false,
    wsError: '',
    syncTimeout: null,
    
    init() {
      const checkReady = () => {
        if (window.API_BASE) {
          this.loadScratchpad()
          this.connectWebSocket()
        } else {
          setTimeout(checkReady, 100)
        }
      }
      checkReady()
    },
    
    async loadScratchpad() {
      try {
        const res = await fetch(apiUrl('/api/scratchpad'))
        const data = await res.json()
        this.content = data.content || ''
      } catch (err) {
        console.error('Failed to load scratchpad:', err)
      }
    },
    
    connectWebSocket() {
      const wsUrlStr = wsUrl()
      this.wsError = ''
      
      try {
        this.ws = new WebSocket(wsUrlStr)
        
        this.ws.onopen = () => {
          console.log('WebSocket connected')
          this.wsConnected = true
          this.wsError = ''
        }
        
        this.ws.onmessage = (event) => {
          try {
            const msg = JSON.parse(event.data)
            if (msg.type === 'scratchpad') {
              this.content = msg.data
            } else if (msg.type === 'clipboard') {
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
          this.wsConnected = false
          setTimeout(() => this.connectWebSocket(), 3000)
        }
        
        this.ws.onerror = (err) => {
          console.error('WebSocket error:', err)
          this.wsConnected = false
          this.wsError = 'Connection failed - scratchpad sync unavailable'
        }
      } catch (err) {
        console.error('Failed to create WebSocket:', err)
        this.wsError = 'Cannot connect to server'
      }
    },
    
    onInput() {
      if (this.syncTimeout) clearTimeout(this.syncTimeout)
      this.syncTimeout = setTimeout(() => {
        this.syncToServer()
      }, 300)
    },
    
    async syncToServer() {
      try {
        await fetch(apiUrl('/api/scratchpad'), {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ content: this.content })
        })
      } catch (err) {
        console.error('Failed to sync scratchpad:', err)
      }
    },
    
    addToHistory(text) {
      this.history.unshift({
        text: text,
        time: new Date().toLocaleTimeString(),
      })
      if (this.history.length > 10) {
        this.history.pop()
      }
    },
    
    async copyToClipboard(text) {
      try {
        await navigator.clipboard.writeText(text)
        this.copySuccess = true
        setTimeout(() => { this.copySuccess = false }, 1500)
      } catch (err) {
        console.error('Failed to copy:', err)
        const textarea = document.createElement('textarea')
        textarea.value = text
        document.body.appendChild(textarea)
        textarea.select()
        document.execCommand('copy')
        document.body.removeChild(textarea)
        this.copySuccess = true
        setTimeout(() => { this.copySuccess = false }, 1500)
      }
    },
    
    useHistoryItem(item) {
      this.content = item.text
      this.syncToServer()
    },
  }
}
```

Open `web/index.html`. Find the Clipboard Panel section (approximately lines 218-277). Replace with:

```html
<!-- Scratchpad Panel -->
<div 
  x-show="showClipboard" 
  x-transition
  class="fixed inset-0 bg-gray-900 flex flex-col z-50"
  x-data="clipboardPanel()"
  x-init="init()"
>
  <header class="bg-gray-800 px-4 py-3 flex items-center justify-between">
    <h2 class="text-xl font-bold flex items-center gap-2">
      <span x-html="icon('clipboard', 20)"></span>
      Scratchpad
    </h2>
    <button @click="$dispatch('toggle-clipboard')" class="text-2xl text-gray-400 hover:text-white" x-html="icon('close', 24)"></button>
  </header>
  
  <div class="flex-1 overflow-y-auto p-4 flex flex-col">
    <div x-show="!wsConnected" class="mb-4 p-3 bg-yellow-900 border border-yellow-700 rounded-lg text-yellow-400 text-sm">
      <span x-text="wsError || 'Connecting...'"></span>
    </div>
    
    <textarea 
      x-model="content"
      @input="onInput()"
      placeholder="Type anything here... it syncs to all connected devices in real-time"
      class="flex-1 w-full bg-gray-800 rounded-lg p-4 text-white resize-none font-mono text-sm focus:outline-none focus:ring-2 focus:ring-purple-500"
      :disabled="!wsConnected"
    ></textarea>
    
    <div class="mt-2 text-xs text-gray-500 flex items-center gap-2">
      <span x-show="wsConnected" class="text-green-400">● Synced</span>
      <span x-show="!wsConnected" class="text-yellow-400">● Offline</span>
    </div>
  </div>
</div>
```

Also update the bottom clipboard button (approximately line 208-210):

```html
<button @click="$dispatch('toggle-clipboard')" class="px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded-lg text-sm transition-colors flex items-center gap-1">
  <span x-html="icon('clipboard', 16)"></span>
  Scratchpad
</button>
```

✅ Success: Scratchpad panel opens. Typing syncs to server. Opening on another device shows same content.
❌ If failed: Stop. Report what happens.

---

### Step 8: Replace emoji with auto-generated per-file icons

Open `web/index.html`. Update the script tag at the bottom (approximately line 279) to import icons first:

```html
<script type="module">
  import '/src/icons.js'
  import '/src/main.js'
</script>
```

Replace emoji icons in the file list (approximately lines 122-123). Change:

```html
<span x-show="file.isDir" class="text-2xl">📁</span>
<span x-show="!file.isDir" class="text-2xl">📄</span>
```

To:

```html
<span x-html="fileIcon(file.name, file.isDir, 24)"></span>
```

Update the upload button (approximately line 204-207):

```html
<label class="bg-blue-600 hover:bg-blue-700 text-center px-4 py-2 rounded-lg cursor-pointer text-sm transition-colors flex items-center gap-1">
  <span x-html="icon('upload', 16)"></span>
  Upload
  <input type="file" @change="uploadFile($event)" class="hidden">
</label>
```

✅ Success: All emoji replaced with auto-generated icons. Each file type gets distinctive color. Documents show extension badge. Images/videos/audio have distinct shapes. Build completes.
❌ If failed: Stop. Report what icons are broken or missing.

---

## Verification

1. Run `pnpm build` in `web/` directory — builds without errors
2. Run `go build ./...` from project root — compiles without errors
3. Start server with `go run ./cmd/looty`
4. Open UI in browser
5. Verify:
   - Breadcrumb buttons are big and chunky
   - Refresh button spins when clicked
   - Four sort buttons (A→Z, Z→A, Newest, Oldest) work correctly
   - Scratchpad opens, typing syncs, persists while server runs
   - File icons are distinct per file type with colors and badges
   - `.xlsx` shows green document with XLSX badge
   - `.png` shows image icon (no badge)
   - `.mp4` shows video icon with play button
   - Unknown extensions show gray document with extension badge

---

## Rollback

If critical failure, revert all changes:

```bash
git checkout -- web/src/components/fileBrowser.js
git checkout -- web/src/components/clipboard.js
git checkout -- web/src/style.css
git checkout -- web/index.html
git checkout -- internal/server/server.go
git checkout -- internal/clipboard/clipboard.go
rm web/src/icons.js
```

---

## Round 2: New Requirements

### Step 9: Make each breadcrumb its own scrollable mini-breadcrumb

The current breadcrumb uses `whitespace-nowrap` which keeps everything on one long line. We want each folder to be its own mini breadcrumb with horizontal scrolling when there are many folders.

Open `web/index.html`. Find the breadcrumb section (approximately lines 86-103). Replace with:

```html
<!-- Breadcrumbs -->
<div class="flex items-center gap-1 text-sm overflow-x-auto flex-1">
  <template x-for="(crumb, i) in breadcrumbPath" :key="crumb.path">
    <div class="flex items-center min-w-fit">
      <template x-if="i > 0">
        <span class="text-gray-500 mx-1" x-html="icon('chevronRight', 14)"></span>
      </template>
      <button
        @click="navigateTo(crumb.path)"
        class="px-3 py-1.5 rounded-lg text-sm font-medium transition-all"
        :class="i === breadcrumbPath.length - 1
          ? 'bg-purple-600 text-white'
          : 'bg-gray-700 text-gray-200 hover:bg-gray-600'"
        x-text="crumb.name"
      ></button>
    </div>
  </template>
</div>
```

**Changes:**
- Removed `whitespace-nowrap` class
- Added `min-w-fit` to the breadcrumb container to allow individual breadcrumbs to shrink
- Breadcrumbs will now wrap on separate lines when there are many folders

✅ Success: Breadcrumbs appear as separate mini-breadcrumbs, scrollable horizontally if too many.
❌ If failed: Report what happens instead.

---

### Step 10: Fix Scratchpad button to work with shared history

The Scratchpad button currently doesn't open the panel. We need to investigate why the toggle event isn't firing and implement shared history support.

**First, investigate the Scratchpad button issue:**

Open `web/index.html` and find the Scratchpad button (approximately line 227-230):

```html
<button @click="$dispatch('toggle-clipboard')" class="px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded-lg text-sm transition-colors flex items-center gap-1">
  <span x-html="icon('clipboard', 16)"></span>
  Scratchpad
</button>
```

Check if the `showClipboard` state exists in the main Alpine app. Open `web/src/main.js`:

```javascript
Alpine.data('app', () => ({
  connected: false,
  serverIP: '',
  status: 'searching',
  showClipboard: false,  // This exists
  manualIP: '',
  // ...
}))
```

The event listener exists:
```javascript
window.addEventListener('toggle-clipboard', () => {
  this.showClipboard = !this.showClipboard
})
```

**If the button still doesn't work, the issue might be with Alpine's `$dispatch` in Alpine 3.x.** Replace the button with a manual event dispatch:

```html
<button @click="window.dispatchEvent(new CustomEvent('toggle-clipboard'))" class="px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded-lg text-sm transition-colors flex items-center gap-1">
  <span x-html="icon('clipboard', 16)"></span>
  Scratchpad
</button>
```

**Second, add shared history support:**

Open `web/src/components/clipboard.js`. Replace the entire file with:

```javascript
import { apiUrl, wsUrl } from '../utils.js'

export function clipboardPanel() {
  return {
    content: '',
    history: [],
    ws: null,
    copySuccess: false,
    sendSuccess: false,
    wsConnected: false,
    wsError: '',
    initialized: false,
    syncTimeout: null,

    init() {
      const checkReady = () => {
        if (window.API_BASE) {
          this.loadScratchpad()
          this.connectWebSocket()
        } else {
          setTimeout(checkReady, 100)
        }
      }
      checkReady()
    },

    async loadScratchpad() {
      try {
        const res = await fetch(apiUrl('/api/scratchpad'))
        const data = await res.json()
        this.content = data.content || ''
        this.addToHistory(data.content || '')
      } catch (err) {
        console.error('Failed to load scratchpad:', err)
      }
    },

    connectWebSocket() {
      const wsUrlStr = wsUrl()
      this.wsError = ''

      try {
        this.ws = new WebSocket(wsUrlStr)

        this.ws.onopen = () => {
          console.log('WebSocket connected')
          this.wsConnected = true
          this.wsError = ''
        }

        this.ws.onmessage = (event) => {
          try {
            const msg = JSON.parse(event.data)
            if (msg.type === 'scratchpad') {
              // Update from another client
              this.content = msg.data
              this.addToHistory(msg.data)
            } else if (msg.type === 'clipboard') {
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
          this.wsConnected = false
          setTimeout(() => this.connectWebSocket(), 3000)
        }

        this.ws.onerror = (err) => {
          console.error('WebSocket error:', err)
          this.wsConnected = false
          this.wsError = 'Connection failed - scratchpad sync unavailable'
        }
      } catch (err) {
        console.error('Failed to create WebSocket:', err)
        this.wsError = 'Cannot connect to server'
      }
    },

    onInput() {
      // Debounce sync to avoid spamming
      if (this.syncTimeout) clearTimeout(this.syncTimeout)
      this.syncTimeout = setTimeout(() => {
        this.syncToServer()
      }, 300)
    },

    async syncToServer() {
      try {
        await fetch(apiUrl('/api/scratchpad'), {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ content: this.content })
        })
      } catch (err) {
        console.error('Failed to sync scratchpad:', err)
      }
    },

    addToHistory(text) {
      // Skip if text is empty or same as last item
      if (!text || text === this.history[0]?.text) return

      this.history.unshift({
        text: text,
        time: new Date().toLocaleTimeString(),
      })
      if (this.history.length > 50) {
        this.history.pop()
      }
    },

    async copyToClipboard(text) {
      try {
        await navigator.clipboard.writeText(text)
        this.copySuccess = true
        setTimeout(() => { this.copySuccess = false }, 1500)
      } catch (err) {
        console.error('Failed to copy:', err)
        const textarea = document.createElement('textarea')
        textarea.value = text
        document.body.appendChild(textarea)
        textarea.select()
        document.execCommand('copy')
        document.body.removeChild(textarea)
        this.copySuccess = true
        setTimeout(() => { this.copySuccess = false }, 1500)
      }
    },

    useHistoryItem(item) {
      this.content = item.text
      this.syncToServer()
    },
  }
}
```

**Update the Scratchpad panel to show history:**

Open `web/index.html`. Find the Scratchpad Panel section (approximately lines 243-279). Replace with:

```html
<!-- Scratchpad Panel -->
<div
  x-show="showClipboard"
  x-transition
  class="fixed inset-0 bg-gray-900 flex flex-col z-50"
  x-data="clipboardPanel()"
  x-init="init()"
>
  <header class="bg-gray-800 px-4 py-3 flex items-center justify-between">
    <h2 class="text-xl font-bold flex items-center gap-2">
      <span x-html="icon('clipboard', 20)"></span>
      Scratchpad
    </h2>
    <button @click="$dispatch('toggle-clipboard')" class="text-2xl text-gray-400 hover:text-white" x-html="icon('close', 24)"></button>
  </header>

  <div class="flex-1 overflow-hidden flex flex-col">
    <!-- Connection status -->
    <div x-show="!wsConnected" class="px-4 py-2 bg-yellow-900/50 border-b border-yellow-700 text-yellow-400 text-sm">
      <span x-text="wsError || 'Connecting...'"></span>
    </div>

    <!-- History sidebar -->
    <div class="flex-1 flex overflow-hidden">
      <!-- Scratchpad textarea -->
      <div class="flex-1 flex flex-col">
        <div class="flex-1 overflow-y-auto p-4">
          <textarea
            x-model="content"
            @input="onInput()"
            placeholder="Type anything here... it syncs to all connected devices in real-time"
            class="flex-1 w-full bg-gray-800 rounded-lg p-4 text-white resize-none font-mono text-sm focus:outline-none focus:ring-2 focus:ring-purple-500"
            :disabled="!wsConnected"
          ></textarea>
        </div>

        <div class="p-3 border-t border-gray-700">
          <div class="text-xs text-gray-500 flex items-center gap-2">
            <span x-show="wsConnected" class="text-green-400">● Synced</span>
            <span x-show="!wsConnected" class="text-yellow-400">● Offline</span>
          </div>
        </div>
      </div>

      <!-- History sidebar -->
      <div class="w-48 border-l border-gray-700 bg-gray-900 overflow-y-auto">
        <div class="p-2 border-b border-gray-700">
          <span class="text-xs font-medium text-gray-400">History</span>
        </div>
        <template x-for="(item, i) in history" :key="i">
          <div
            @click="useHistoryItem(item)"
            class="p-3 cursor-pointer hover:bg-gray-800 transition-colors border-b border-gray-700"
          >
            <div class="text-xs text-gray-400 mb-1" x-text="item.time"></div>
            <div class="text-sm text-gray-200 line-clamp-3" x-text="item.text"></div>
          </div>
        </template>
        <template x-if="!history.length">
          <div class="p-4 text-xs text-gray-500 text-center">No history yet</div>
        </template>
      </div>
    </div>
  </div>
</div>
```

✅ Success: Scratchpad button opens panel. Shared history shows last 50 entries. Typing syncs to server. History updates in real-time.
❌ If failed: Report what happens instead.

---

### Step 11: Change file position background from purple to neutral

The file list currently uses `bg-purple-950` for the page background. Change it to a neutral gray.

Open `web/index.html`. Find the `body` tag (approximately line 8):

```html
<body class="bg-purple-950 text-gray-100 h-full flex flex-col" x-data="app()" x-init="init()">
```

Replace with:

```html
<body class="bg-gray-950 text-gray-100 h-full flex flex-col" x-data="app()" x-init="init()">
```

✅ Success: Page background changed from purple to neutral gray.
❌ If failed: Report what happens instead.

---

## Verification

1. Run `pnpm build` in `web/` directory — builds without errors
2. Run `go build ./...` from project root — compiles without errors
3. Start server with `go run ./cmd/looty`
4. Open UI in browser
5. Verify:
   - Breadcrumbs appear as separate mini-breadcrumbs, scrollable horizontally if too many
   - Scratchpad button opens panel
   - Scratchpad shows shared content across devices
   - Scratchpad history sidebar shows last 50 entries
   - Typing in scratchpad syncs to server in real-time
   - File position background is neutral gray (not purple)
   - All previous functionality still works (refresh, sorting, icons)

---

## Rollback

If critical failure, revert all changes:

```bash
git checkout -- web/src/components/fileBrowser.js
git checkout -- web/src/components/clipboard.js
git checkout -- web/src/style.css
git checkout -- web/index.html
git checkout -- internal/server/server.go
git checkout -- internal/clipboard/clipboard.go
rm web/src/icons.js
```

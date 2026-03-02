package server

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/user/looty/internal/clipboard"
	"github.com/user/looty/internal/files"
)

// BuildTime is set via ldflags at build time
var BuildTime string

//go:embed assets/index.html
var staticFiles embed.FS

// GetHTML returns the embedded index.html content with build time injected
func GetHTML() ([]byte, error) {
	content, err := staticFiles.ReadFile("assets/index.html")
	if err != nil {
		return nil, err
	}
	// Inject build time
	if BuildTime != "" {
		return []byte(strings.Replace(string(content), "__BUILD_TIME__", BuildTime, 1)), nil
	}
	return content, nil
}

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

type Server struct {
	serveDir string
	port     int
}

func Start(serveDir string, port int) error {
	hub = NewHub()
	go hub.Run()

	// Start file watcher
	StartWatcher(serveDir)

	s := &Server{
		serveDir: serveDir,
		port:     port,
	}

	mux := http.NewServeMux()

	// CORS middleware - wrap all handlers
	withCORS := func(handler http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			handler(w, r)
		}
	}

	// Serve embedded index.html at root
	mux.HandleFunc("/", withCORS(s.serveIndex))

	// Health check for discovery
	mux.HandleFunc("/ping", withCORS(s.handlePing))

	// File API endpoints
	mux.HandleFunc("/api/files", withCORS(files.ListHandler(s.serveDir)))
	mux.HandleFunc("/api/download", withCORS(files.DownloadHandler(s.serveDir)))
	mux.HandleFunc("/api/upload", withCORS(files.UploadHandler(s.serveDir)))

	// WebSocket endpoint
	mux.HandleFunc("/ws", withCORS(s.handleWebSocket))

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

	addr := fmt.Sprintf(":%d", port)
	return http.ListenAndServe(addr, mux)
}

func (s *Server) serveIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	content, err := GetHTML()
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
		// Parse and broadcast clipboard messages
		message, err := clipboard.ParseMessage(msg)
		if err != nil {
			log.Printf("Failed to parse clipboard message: %v", err)
			return
		}

		if message.Type == clipboard.TypeClipboard || message.Type == clipboard.TypeScratchpad {
			// Broadcast to all other clients
			hub.broadcast <- msg
		}
	})
}

func Broadcast(message []byte) {
	if hub != nil {
		hub.broadcast <- message
	}
}

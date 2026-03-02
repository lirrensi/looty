package server

import (
	"embed"
	"fmt"
	"log"
	"net/http"

	"github.com/user/looty/internal/clipboard"
	"github.com/user/looty/internal/files"
)

//go:embed assets/index.html
var staticFiles embed.FS

var hub *Hub

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

	addr := fmt.Sprintf(":%d", port)
	return http.ListenAndServe(addr, mux)
}

func (s *Server) serveIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	content, err := staticFiles.ReadFile("assets/index.html")
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

		if message.Type == clipboard.TypeClipboard {
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

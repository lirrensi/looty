// FILE: internal/server/server.go
// PURPOSE: HTTP/HTTPS server with WebSocket, file API, scratchpad, and CORS.
// OWNS: Server lifecycle, route registration, TLS-aware listener construction.
// EXPORTS: Start, BuildHandler, CreateListener, ServeListener, Config, GetHTML, GetScratchpad, SetScratchpad, Broadcast
// DOCS: agent_chat/plan_daemon-mode_2026-05-17.md, agent_chat/plan_tls-paradigm_2026-05-17.md

package server

import (
	"crypto/tls"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/lirrensi/looty/internal/clipboard"
	"github.com/lirrensi/looty/internal/files"
)

// Version is set via ldflags at build time (e.g., -X ...Version=1.0.0)
var Version string

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

// Config holds server startup configuration.
type Config struct {
	ServeDir string
	Host     string
	Port     int
	UseTLS   bool
	Cert     tls.Certificate // zero value if not TLS
}

func Start(cfg Config) error {
	handler := BuildHandler(cfg)
	listener, err := CreateListener(cfg)
	if err != nil {
		return err
	}
	return ServeListener(cfg, listener, handler)
}

func BuildHandler(cfg Config) http.Handler {
	s := &Server{
		serveDir: cfg.ServeDir,
		port:     cfg.Port,
	}

	mux := http.NewServeMux()

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

	mux.HandleFunc("/", withCORS(s.serveIndex))
	mux.HandleFunc("/ping", withCORS(s.handlePing))
	mux.HandleFunc("/api/files", withCORS(files.ListHandler(s.serveDir)))
	mux.HandleFunc("/api/download", withCORS(files.DownloadHandler(s.serveDir)))
	mux.HandleFunc("/api/upload", withCORS(files.UploadHandler(s.serveDir)))
	mux.HandleFunc("/ws", withCORS(s.handleWebSocket))
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

	return mux
}

func CreateListener(cfg Config) (net.Listener, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	if cfg.UseTLS {
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cfg.Cert},
		}
		listener, err := tls.Listen("tcp", addr, tlsConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create TLS listener: %w", err)
		}
		return listener, nil
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}
	return listener, nil
}

func ServeListener(cfg Config, listener net.Listener, handler http.Handler) error {
	hub = NewHub()
	go hub.Run()

	// Start file watcher
	StartWatcher(cfg.ServeDir)

	// Start mDNS announcement
	if err := StartMDNS(cfg.Port); err != nil {
		log.Printf("Warning: mDNS failed: %v", err)
	}

	return http.Serve(listener, handler)
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

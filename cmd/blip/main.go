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

//go:embed index.html
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
		src, err := staticFiles.Open("index.html")
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

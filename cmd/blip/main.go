package main

import (
	"embed"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/user/blip/internal/server"
)

//go:embed index.html
var staticFiles embed.FS

func getLocalIPs() []string {
	var ips []string
	ifaces, err := net.Interfaces()
	if err != nil {
		return ips
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip != nil && ip.To4() != nil {
				ips = append(ips, ip.String())
			}
		}
	}
	return ips
}

func main() {
	execPath, err := os.Executable()
	if err != nil {
		log.Fatal("Failed to get executable path:", err)
	}
	serveDir := filepath.Dir(execPath)

	// Always extract fresh blip.html
	blipHTMLPath := filepath.Join(serveDir, "blip.html")
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

	fmt.Printf("\nBlip serving: %s\n", serveDir)
	fmt.Println("Access URLs:")
	ips := getLocalIPs()
	for _, ip := range ips {
		fmt.Printf("  http://%s:41111\n", ip)
	}
	fmt.Printf("  http://localhost:41111\n")
	fmt.Printf("\nOr copy blip.html to your phone\n\n")

	if err := server.Start(serveDir, 41111); err != nil {
		log.Fatal("Server error:", err)
	}
}

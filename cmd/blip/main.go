package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/user/looty/internal/server"
)

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
	serveDir, err := os.Getwd()
	if err != nil {
		log.Fatal("Failed to get working directory:", err)
	}

	// Extract looty.html to exe's directory (so it can be copied to phone)
	execPath, err := os.Executable()
	if err != nil {
		log.Printf("Warning: Could not get executable path: %v", err)
	}
	exeDir := filepath.Dir(execPath)
	lootyHTMLPath := filepath.Join(exeDir, "looty.html")
	html, err := server.GetHTML()
	if err != nil {
		log.Printf("Warning: Could not get embedded HTML: %v", err)
	} else {
		err = os.WriteFile(lootyHTMLPath, html, 0644)
		if err != nil {
			log.Printf("Warning: Could not create looty.html: %v", err)
		} else {
			fmt.Println("Extracted looty.html")
		}
	}

	fmt.Printf("\nLOOTY serving: %s\n", serveDir)
	fmt.Println("Access URLs:")
	ips := getLocalIPs()
	for _, ip := range ips {
		fmt.Printf("  http://%s:41111\n", ip)
	}
	fmt.Printf("  http://localhost:41111\n")
	fmt.Printf("\nOr copy looty.html to your phone\n\n")

	if err := server.Start(serveDir, 41111); err != nil {
		log.Fatal("Server error:", err)
	}
}

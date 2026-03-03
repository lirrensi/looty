package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/lirrensi/looty/internal/server"
	"github.com/mdp/qrterminal/v3"
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

// getPrimaryIP returns the IP of the interface with the default gateway
func getPrimaryIP() string {
	// Get all network interfaces
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	for _, iface := range ifaces {
		// Skip interfaces that are down or loopback
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// Skip common virtual adapter names
		name := iface.Name
		if isVirtualAdapter(name) {
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

			if ip == nil || ip.To4() == nil {
				continue
			}

			// Skip link-local addresses (169.254.x.x)
			if ip.IsLinkLocalUnicast() {
				continue
			}

			return ip.String()
		}
	}
	return ""
}

func isVirtualAdapter(name string) bool {
	// Common virtual adapter patterns
	virtuals := []string{"VMware", "VirtualBox", "Hyper-V", "WSL", "Docker", "vEthernet", "Loopback", "Tunnel", "TAP", "WireGuard"}
	for _, v := range virtuals {
		if containsIgnoreCase(name, v) {
			return true
		}
	}
	return false
}

func containsIgnoreCase(s, substr string) bool {
	sLower := make([]byte, len(s))
	substrLower := make([]byte, len(substr))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		sLower[i] = c
	}
	for i := 0; i < len(substr); i++ {
		c := substr[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		substrLower[i] = c
	}
	return contains(string(sLower), string(substrLower))
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
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
		// Write to exe directory
		err = os.WriteFile(lootyHTMLPath, html, 0644)
		if err != nil {
			log.Printf("Warning: Could not create looty.html: %v", err)
		} else {
			fmt.Println("Extracted looty.html to exe directory")
		}

		// Also write to home folder for easy access
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Printf("Warning: Could not get home directory: %v", err)
		} else {
			homeLootyDir := filepath.Join(homeDir, "looty")
			os.MkdirAll(homeLootyDir, 0755)
			homeLootyPath := filepath.Join(homeLootyDir, "looty.html")
			err = os.WriteFile(homeLootyPath, html, 0644)
			if err != nil {
				log.Printf("Warning: Could not create looty.html in home: %v", err)
			} else {
				fmt.Printf("Also saved to: %s\n", homeLootyPath)
			}
		}
	}

	fmt.Printf("\nLOOTY serving: %s\n\n", serveDir)

	ips := getLocalIPs()
	primaryIP := getPrimaryIP()

	if primaryIP != "" {
		url := fmt.Sprintf("http://%s:41111", primaryIP)
		fmt.Printf("Scan QR code or open: %s\n\n", url)

		// Print QR code (half-block mode for smaller size)
		qrterminal.GenerateWithConfig(url, qrterminal.Config{
			Level:      qrterminal.M,
			Writer:     os.Stdout,
			HalfBlocks: true,
		})
	} else if len(ips) > 0 {
		// Fallback to first IP if no primary detected
		url := fmt.Sprintf("http://%s:41111", ips[0])
		fmt.Printf("Scan QR code or open: %s\n\n", url)

		qrterminal.GenerateWithConfig(url, qrterminal.Config{
			Level:      qrterminal.M,
			Writer:     os.Stdout,
			HalfBlocks: true,
		})
	}

	fmt.Printf("\nAll addresses:\n")
	for _, ip := range ips {
		fmt.Printf("  http://%s:41111\n", ip)
	}
	fmt.Printf("  http://localhost:41111\n")
	fmt.Printf("\nOr open ~/looty/looty.html on your phone\n\n")

	if err := server.Start(serveDir, 41111); err != nil {
		log.Fatal("Server error:", err)
	}
}

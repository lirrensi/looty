// FILE: cmd/blip/main.go
// PURPOSE: CLI entry point for Looty. Parses flags, decides TLS mode, prints URLs/QR, and starts the server.
// OWNS: Flag parsing, TLS decision logic, startup orchestration, console output.
// EXPORTS: main
// DOCS: agent_chat/plan_tls-paradigm_2026-05-17.md

package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/lirrensi/looty/internal/certgen"
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
		if strings.Contains(strings.ToLower(name), strings.ToLower(v)) {
			return true
		}
	}
	return false
}

func isLoopback(host string) bool {
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return true
	}
	ip := net.ParseIP(host)
	if ip != nil && ip.IsLoopback() {
		return true
	}
	return false
}

func isAllInterfaces(host string) bool {
	return host == "" || host == "0.0.0.0" || host == "::"
}

func main() {
	// CLI flags
	hostFlag := flag.String("host", "", "Host to bind to (default: all interfaces)")
	portFlag := flag.Int("port", 41111, "Port to listen on")
	useTLSFlag := flag.Bool("tls", false, "Force TLS with auto-generated certificate")
	noTLSFlag := flag.Bool("no-tls", false, "Force plain HTTP (opt out of auto-TLS)")
	certPath := flag.String("cert", "", "Path to TLS certificate file")
	keyPath := flag.String("key", "", "Path to TLS private key file")
	flag.Parse()

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

	// TLS decision logic
	var useTLS bool
	var cert tls.Certificate
	var fingerprint, friendCode string

	if *certPath != "" && *keyPath != "" {
		useTLS = true
		loadedCert, err := tls.LoadX509KeyPair(*certPath, *keyPath)
		if err != nil {
			log.Fatalf("Failed to load TLS certificate: %v", err)
		}
		cert = loadedCert
	} else if *noTLSFlag {
		useTLS = false
	} else if *useTLSFlag {
		useTLS = true
	} else {
		// Default based on bind address
		if isLoopback(*hostFlag) {
			useTLS = false
		} else {
			// Empty string (all interfaces), 0.0.0.0, ::, or any non-loopback IP
			useTLS = true
		}
	}

	if useTLS {
		if cert.Certificate == nil {
			// Auto-generate certificate
			genCert, fp, fc, err := certgen.GenerateSelfSigned()
			if err != nil {
				log.Fatalf("Failed to generate TLS certificate: %v", err)
			}
			cert = *genCert
			fingerprint = fp
			friendCode = fc
		}
	}

	// Determine protocol and addresses to print
	protocol := "http"
	if useTLS {
		protocol = "https"
	}

	var displayAddrs []string
	if isAllInterfaces(*hostFlag) {
		ips := getLocalIPs()
		primaryIP := getPrimaryIP()

		if primaryIP != "" {
			displayAddrs = append(displayAddrs, primaryIP)
		} else if len(ips) > 0 {
			displayAddrs = append(displayAddrs, ips[0])
		}
		for _, ip := range ips {
			// Avoid duplicates
			found := false
			for _, da := range displayAddrs {
				if da == ip {
					found = true
					break
				}
			}
			if !found {
				displayAddrs = append(displayAddrs, ip)
			}
		}
		displayAddrs = append(displayAddrs, "localhost")
	} else if isLoopback(*hostFlag) {
		displayAddrs = []string{"localhost", "127.0.0.1"}
	} else {
		displayAddrs = []string{*hostFlag}
	}

	// Print primary URL and QR code
	if len(displayAddrs) > 0 {
		primaryURL := fmt.Sprintf("%s://%s:%d", protocol, displayAddrs[0], *portFlag)
		if useTLS {
			fmt.Printf("🔒 %s\n", primaryURL)
			if fingerprint != "" {
				fmt.Printf("   Fingerprint: %s\n", fingerprint)
			}
			if friendCode != "" {
				fmt.Printf("   Friend code: %s\n", friendCode)
			}
			fmt.Println()
		} else {
			fmt.Printf("Scan QR code or open: %s\n\n", primaryURL)
		}

		qrterminal.GenerateWithConfig(primaryURL, qrterminal.Config{
			Level:      qrterminal.M,
			Writer:     os.Stdout,
			HalfBlocks: true,
		})
	}

	// Print all addresses
	fmt.Printf("\nAll addresses:\n")
	for _, addr := range displayAddrs {
		fmt.Printf("  %s://%s:%d\n", protocol, addr, *portFlag)
	}

	if useTLS {
		fmt.Println()
		fmt.Println("Verify the certificate fingerprint matches what was shared with you.")
		fmt.Println("⚠️  looty.html discovery does not work with HTTPS. Share the direct link above.")
	} else {
		fmt.Printf("\nOr open ~/looty/looty.html on your phone\n")
	}
	fmt.Println()

	cfg := server.Config{
		ServeDir: serveDir,
		Host:     *hostFlag,
		Port:     *portFlag,
		UseTLS:   useTLS,
		Cert:     cert,
	}

	if err := server.Start(cfg); err != nil {
		log.Fatal("Server error:", err)
	}
}

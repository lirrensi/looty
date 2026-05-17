// FILE: internal/server/startup.go
// PURPOSE: Build, render, and persist structured startup metadata for foreground and background Looty launches.
// OWNS: Startup record shape, address and URL derivation, JSON serialization, human output rendering, QR rendering, atomic record writes.
// EXPORTS: StartupRecord, ProtocolForTLS, DeriveAddresses, DeriveURLs, NewStartupRecord, RenderStartupRecord, RenderStartupQRCode, WriteStartupRecordFile
// DOCS: agent_chat/plan_daemon-mode_2026-05-17.md, docs/spec.md, docs/arch.md

package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mdp/qrterminal/v3"
)

type StartupRecord struct {
	Mode        string    `json:"mode"`
	ServeDir    string    `json:"serveDir"`
	Protocol    string    `json:"protocol"`
	Host        string    `json:"host"`
	Port        int       `json:"port"`
	PrimaryURL  string    `json:"primaryUrl"`
	Addresses   []string  `json:"addresses"`
	AllURLs     []string  `json:"allUrls"`
	Fingerprint string    `json:"fingerprint,omitempty"`
	FriendCode  string    `json:"friendCode,omitempty"`
	PID         int       `json:"pid,omitempty"`
	StartedAt   time.Time `json:"startedAt,omitempty"`
	HTMLPaths   []string  `json:"htmlPaths,omitempty"`
}

func ProtocolForTLS(useTLS bool) string {
	if useTLS {
		return "https"
	}
	return "http"
}

func DeriveAddresses(host string) []string {
	if isAllInterfaces(host) {
		ips := getLocalIPs()
		primaryIP := getPrimaryIP()
		addresses := make([]string, 0, len(ips)+2)
		seen := map[string]struct{}{}

		appendAddress := func(addr string) {
			addr = strings.TrimSpace(addr)
			if addr == "" {
				return
			}
			if _, ok := seen[addr]; ok {
				return
			}
			seen[addr] = struct{}{}
			addresses = append(addresses, addr)
		}

		appendAddress(primaryIP)
		for _, ip := range ips {
			appendAddress(ip)
		}
		appendAddress("localhost")

		return addresses
	}

	if isLoopback(host) {
		return []string{"localhost", "127.0.0.1"}
	}

	return []string{host}
}

func DeriveURLs(protocol string, addresses []string, port int) (string, []string) {
	urls := make([]string, 0, len(addresses))
	for _, address := range addresses {
		urls = append(urls, fmt.Sprintf("%s://%s:%d", protocol, address, port))
	}
	if len(urls) == 0 {
		return "", nil
	}
	return urls[0], urls
}

func NewStartupRecord(mode, serveDir, host string, port int, useTLS bool, fingerprint, friendCode string, pid int, startedAt time.Time, htmlPaths []string) StartupRecord {
	protocol := ProtocolForTLS(useTLS)
	addresses := DeriveAddresses(host)
	primaryURL, allURLs := DeriveURLs(protocol, addresses, port)

	return StartupRecord{
		Mode:        mode,
		ServeDir:    serveDir,
		Protocol:    protocol,
		Host:        host,
		Port:        port,
		PrimaryURL:  primaryURL,
		Addresses:   addresses,
		AllURLs:     allURLs,
		Fingerprint: fingerprint,
		FriendCode:  friendCode,
		PID:         pid,
		StartedAt:   startedAt.UTC(),
		HTMLPaths:   append([]string(nil), htmlPaths...),
	}
}

func (record StartupRecord) JSON() ([]byte, error) {
	return json.MarshalIndent(record, "", "  ")
}

func RenderStartupRecord(record StartupRecord) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("\nLOOTY serving: %s\n\n", record.ServeDir))

	if record.PrimaryURL != "" {
		if record.Protocol == "https" {
			builder.WriteString(fmt.Sprintf("🔒 %s\n", record.PrimaryURL))
			if record.Fingerprint != "" {
				builder.WriteString(fmt.Sprintf("   Fingerprint: %s\n", record.Fingerprint))
			}
			if record.FriendCode != "" {
				builder.WriteString(fmt.Sprintf("   Friend code: %s\n", record.FriendCode))
			}
			builder.WriteString("\n")
		} else {
			builder.WriteString(fmt.Sprintf("Scan QR code or open: %s\n\n", record.PrimaryURL))
		}
	}

	builder.WriteString("All addresses:\n")
	for _, url := range record.AllURLs {
		builder.WriteString(fmt.Sprintf("  %s\n", url))
	}

	if record.Protocol == "https" {
		builder.WriteString("\nVerify the certificate fingerprint matches what was shared with you.\n")
		builder.WriteString("⚠️  looty.html discovery does not work with HTTPS. Share the direct link above.\n")
	} else if preferredHTMLPath(record.HTMLPaths) != "" {
		builder.WriteString(fmt.Sprintf("\nOr open %s on your phone\n", preferredHTMLPath(record.HTMLPaths)))
	}

	builder.WriteString("\n")
	return builder.String()
}

func RenderStartupQRCode(record StartupRecord) string {
	if record.PrimaryURL == "" {
		return ""
	}

	var buffer bytes.Buffer
	qrterminal.GenerateWithConfig(record.PrimaryURL, qrterminal.Config{
		Level:      qrterminal.M,
		Writer:     &buffer,
		HalfBlocks: true,
	})
	return buffer.String()
}

func WriteStartupRecordFile(path string, record StartupRecord) error {
	data, err := record.JSON()
	if err != nil {
		return err
	}
	return writeAtomically(path, data)
}

func writeAtomically(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create startup record directory: %w", err)
	}

	tempFile, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return fmt.Errorf("create startup temp file: %w", err)
	}
	tempPath := tempFile.Name()

	defer func() {
		_ = os.Remove(tempPath)
	}()

	if _, err := tempFile.Write(data); err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("write startup temp file: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("close startup temp file: %w", err)
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("replace existing startup file: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("rename startup temp file: %w", err)
	}
	return nil
}

func preferredHTMLPath(paths []string) string {
	if len(paths) == 0 {
		return ""
	}
	for _, path := range paths {
		if strings.Contains(strings.ToLower(path), string(filepath.Separator)+"looty"+string(filepath.Separator)+"looty.html") {
			return path
		}
	}
	return paths[len(paths)-1]
}

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
			switch value := addr.(type) {
			case *net.IPNet:
				ip = value.IP
			case *net.IPAddr:
				ip = value.IP
			}
			if ip != nil && ip.To4() != nil {
				ips = append(ips, ip.String())
			}
		}
	}
	return ips
}

func getPrimaryIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if isVirtualAdapter(iface.Name) {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch value := addr.(type) {
			case *net.IPNet:
				ip = value.IP
			case *net.IPAddr:
				ip = value.IP
			}
			if ip == nil || ip.To4() == nil || ip.IsLinkLocalUnicast() {
				continue
			}
			return ip.String()
		}
	}

	return ""
}

func isVirtualAdapter(name string) bool {
	virtuals := []string{"VMware", "VirtualBox", "Hyper-V", "WSL", "Docker", "vEthernet", "Loopback", "Tunnel", "TAP", "WireGuard"}
	for _, virtual := range virtuals {
		if strings.Contains(strings.ToLower(name), strings.ToLower(virtual)) {
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
	return ip != nil && ip.IsLoopback()
}

func isAllInterfaces(host string) bool {
	return host == "" || host == "0.0.0.0" || host == "::"
}

// FILE: cmd/blip/main.go
// PURPOSE: Orchestrate Looty startup modes, startup record emission, and the server serve loop.
// OWNS: CLI flag parsing, TLS selection, daemon parent-child launch, startup record generation, startup output routing.
// EXPORTS: main
// DOCS: agent_chat/plan_qr-port-artifact_2026-05-17.md, docs/spec.md, docs/arch.md

package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/lirrensi/looty/internal/certgen"
	"github.com/lirrensi/looty/internal/server"
)

const daemonStartupTimeout = 10 * time.Second

type cliOptions struct {
	host          string
	port          int
	serveDir      string
	useTLSFlag    bool
	noTLSFlag     bool
	certPath      string
	keyPath       string
	daemon        bool
	json          bool
	jsonFile      string
	daemonChild   bool
	startupFile   string
	parseError    error
	originalFlags *flag.FlagSet
}

func main() {
	options := parseCLI(os.Args[1:])
	if options.parseError != nil {
		log.Fatal(options.parseError)
	}

	if err := run(options); err != nil {
		log.Fatal(err)
	}
}

func run(options cliOptions) error {
	serveDir := options.serveDir
	if serveDir == "" {
		var err error
		serveDir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	if options.daemon && !options.daemonChild {
		return runDaemonParent(options, serveDir)
	}

	htmlPaths, err := extractHTMLPaths(options.port)
	if err != nil {
		log.Printf("Warning: %v", err)
	}

	useTLS, cert, fingerprint, friendCode, err := prepareTLS(options)
	if err != nil {
		return err
	}

	cfg := server.Config{
		ServeDir: serveDir,
		Host:     options.host,
		Port:     options.port,
		UseTLS:   useTLS,
		Cert:     cert,
	}

	handler := server.BuildHandler(cfg)
	listener, err := server.CreateListener(cfg)
	if err != nil {
		return err
	}

	// Warn about firewall when binding to non-loopback (remote / VPS usage).
	if !isLoopbackHost(options.host) {
		log.Printf("⚠️  Port %d — make sure your firewall allows it", options.port)
	}

	mode := determineMode(options)
	record := server.NewStartupRecord(
		mode,
		serveDir,
		options.host,
		options.port,
		useTLS,
		fingerprint,
		friendCode,
		os.Getpid(),
		time.Now(),
		htmlPaths,
	)

	record, err = persistStartupOutputs(options, record)
	if err != nil {
		_ = listener.Close()
		return err
	}

	if !options.daemonChild {
		emitStartupOutput(options, record)
	}

	return server.ServeListener(cfg, listener, handler)
}

func parseCLI(args []string) cliOptions {
	flags := flag.NewFlagSet("looty", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)

	options := cliOptions{originalFlags: flags}
	flags.StringVar(&options.host, "host", "", "Host to bind to (default: all interfaces)")
	flags.IntVar(&options.port, "port", 41111, "Port to listen on")
	flags.StringVar(&options.serveDir, "serve-dir", "", "Directory to serve (default: current working directory)")
	flags.BoolVar(&options.useTLSFlag, "tls", false, "Force TLS with auto-generated certificate")
	flags.BoolVar(&options.noTLSFlag, "no-tls", false, "Force plain HTTP (opt out of auto-TLS)")
	flags.StringVar(&options.certPath, "cert", "", "Path to TLS certificate file")
	flags.StringVar(&options.keyPath, "key", "", "Path to TLS private key file")
	flags.BoolVar(&options.daemon, "daemon", false, "Start Looty in background-capable mode")
	flags.BoolVar(&options.json, "json", false, "Emit the startup record as JSON")
	flags.StringVar(&options.jsonFile, "json-file", "", "Write the startup record JSON to a file")
	flags.BoolVar(&options.daemonChild, "daemon-child", false, "(internal) run as daemon child")
	flags.StringVar(&options.startupFile, "startup-file", "", "(internal) startup record handoff file")

	options.parseError = flags.Parse(args)
	return options
}

func runDaemonParent(options cliOptions, serveDir string) error {
	startupFile, err := newStartupHandoffPath()
	if err != nil {
		return err
	}
	defer os.Remove(startupFile)

	childArgs := buildDaemonChildArgs(options, startupFile)
	execPath, err := currentExecutablePath()
	if err != nil {
		return fmt.Errorf("resolve executable path: %w", err)
	}

	devNull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("open null device: %w", err)
	}
	defer devNull.Close()

	proc, err := os.StartProcess(execPath, append([]string{execPath}, childArgs...), &os.ProcAttr{
		Dir:   serveDir,
		Files: []*os.File{devNull, devNull, devNull},
		Sys:   detachedProcessAttr(),
	})
	if err != nil {
		return fmt.Errorf("start daemon child: %w", err)
	}

	childExit := make(chan error, 1)
	go func() {
		_, waitErr := proc.Wait()
		childExit <- waitErr
	}()

	record, err := waitForStartupRecord(startupFile, childExit, daemonStartupTimeout)
	if err != nil {
		return err
	}

	if options.jsonFile != "" {
		record, err = writeUserFacingStartupArtifacts(options.jsonFile, record)
		if err != nil {
			return fmt.Errorf("write requested startup record file: %w", err)
		}
	}

	emitStartupOutput(options, record)
	return nil
}

func prepareTLS(options cliOptions) (bool, tls.Certificate, string, string, error) {
	var cert tls.Certificate

	if options.certPath != "" || options.keyPath != "" {
		if options.certPath == "" || options.keyPath == "" {
			return false, cert, "", "", errors.New("both -cert and -key are required together")
		}
		loadedCert, err := tls.LoadX509KeyPair(options.certPath, options.keyPath)
		if err != nil {
			return false, cert, "", "", fmt.Errorf("failed to load TLS certificate: %w", err)
		}
		return true, loadedCert, "", "", nil
	}

	useTLS := resolveTLSMode(options.host, options.useTLSFlag, options.noTLSFlag)
	if !useTLS {
		return false, cert, "", "", nil
	}

	generatedCert, fp, fc, err := certgen.GenerateSelfSigned()
	if err != nil {
		return false, cert, "", "", fmt.Errorf("failed to generate TLS certificate: %w", err)
	}

	return true, *generatedCert, fp, fc, nil
}

func resolveTLSMode(host string, forceTLS, noTLS bool) bool {
	if noTLS {
		return false
	}
	if forceTLS {
		return true
	}
	return !isLoopbackHost(host)
}

func determineMode(options cliOptions) string {
	if options.daemon || options.daemonChild {
		if options.json || options.jsonFile != "" {
			return "agent-managed"
		}
		return "background"
	}
	return "foreground"
}

func persistStartupOutputs(options cliOptions, record server.StartupRecord) (server.StartupRecord, error) {
	if options.startupFile != "" {
		if err := server.WriteStartupRecordFile(options.startupFile, record); err != nil {
			return record, fmt.Errorf("write daemon startup handoff: %w", err)
		}
	}
	if options.jsonFile != "" && options.daemonChild {
		return record, nil
	}
	if options.jsonFile != "" {
		updatedRecord, err := writeUserFacingStartupArtifacts(options.jsonFile, record)
		if err != nil {
			return record, fmt.Errorf("write startup record file: %w", err)
		}
		record = updatedRecord
	}
	return record, nil
}

func emitStartupOutput(options cliOptions, record server.StartupRecord) {
	if options.json {
		data, err := record.JSON()
		if err != nil {
			log.Printf("Warning: failed to serialize startup record: %v", err)
			return
		}
		fmt.Println(string(data))
		return
	}

	fmt.Print(server.RenderStartupRecord(record))
	if qr := server.RenderStartupQRCode(record); qr != "" {
		fmt.Print(qr)
	}
	if record.Protocol == "https" {
		return
	}
}

func extractHTMLPaths(port int) ([]string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("could not get executable path: %w", err)
	}

	html, err := server.GetHTML()
	if err != nil {
		return nil, fmt.Errorf("could not get embedded HTML: %w", err)
	}

	html = applyExtractedHTMLPort(html, port)

	paths := make([]string, 0, 2)

	exeDir := filepath.Dir(execPath)
	exePath := filepath.Join(exeDir, "looty.html")
	if err := os.WriteFile(exePath, html, 0o644); err != nil {
		log.Printf("Warning: Could not create looty.html: %v", err)
	} else {
		paths = append(paths, exePath)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Warning: Could not get home directory: %v", err)
		return paths, nil
	}

	homeLootyDir := filepath.Join(homeDir, "looty")
	if err := os.MkdirAll(homeLootyDir, 0o755); err != nil {
		log.Printf("Warning: Could not create looty home directory: %v", err)
		return paths, nil
	}

	homePath := filepath.Join(homeLootyDir, "looty.html")
	if err := os.WriteFile(homePath, html, 0o644); err != nil {
		log.Printf("Warning: Could not create looty.html in home: %v", err)
	} else {
		paths = append(paths, homePath)
	}

	return paths, nil
}

func writeUserFacingStartupArtifacts(jsonPath string, record server.StartupRecord) (server.StartupRecord, error) {
	qrPath := server.DeriveStartupQRFilePath(jsonPath)
	if err := server.WriteStartupQRCodeSVGFile(qrPath, record.PrimaryURL); err != nil {
		return record, fmt.Errorf("write startup QR artifact: %w", err)
	}
	record.QRImagePath = qrPath
	if err := server.WriteStartupRecordFile(jsonPath, record); err != nil {
		return record, err
	}
	return record, nil
}

func applyExtractedHTMLPort(html []byte, port int) []byte {
	const marker = "<head>"
	injection := fmt.Sprintf("<head>\n  <script>window.__LOOTY_PORT__ = %d;</script>", port)
	return []byte(strings.Replace(string(html), marker, injection, 1))
}

func buildDaemonChildArgs(options cliOptions, startupFile string) []string {
	args := []string{"-daemon-child", "-startup-file", startupFile, "-port", strconv.Itoa(options.port)}
	if options.host != "" {
		args = append(args, "-host", options.host)
	}
	if options.serveDir != "" {
		args = append(args, "-serve-dir", options.serveDir)
	}
	if options.useTLSFlag {
		args = append(args, "-tls")
	}
	if options.noTLSFlag {
		args = append(args, "-no-tls")
	}
	if options.certPath != "" {
		args = append(args, "-cert", options.certPath)
	}
	if options.keyPath != "" {
		args = append(args, "-key", options.keyPath)
	}
	if options.json {
		args = append(args, "-json")
	}
	if options.jsonFile != "" {
		args = append(args, "-json-file", options.jsonFile)
	}
	return args
}

func newStartupHandoffPath() (string, error) {
	baseDir := filepath.Join(os.TempDir(), "looty")
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return "", fmt.Errorf("create startup handoff directory: %w", err)
	}
	return filepath.Join(baseDir, fmt.Sprintf("startup-%d.json", time.Now().UnixNano())), nil
}

func currentExecutablePath() (string, error) {
	if os.Args[0] != "" {
		candidate := os.Args[0]
		if filepath.IsAbs(candidate) {
			if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
				return candidate, nil
			}
		}
		if strings.ContainsRune(candidate, filepath.Separator) || strings.Contains(candidate, "/") {
			if absolute, err := filepath.Abs(candidate); err == nil {
				if info, statErr := os.Stat(absolute); statErr == nil && !info.IsDir() {
					return absolute, nil
				}
			}
		}
		if resolved, err := exec.LookPath(candidate); err == nil {
			return resolved, nil
		}
	}

	return os.Executable()
}

func waitForStartupRecord(path string, childExit <-chan error, timeout time.Duration) (server.StartupRecord, error) {
	deadline := time.Now().Add(timeout)
	for {
		select {
		case err := <-childExit:
			if err != nil {
				return server.StartupRecord{}, fmt.Errorf("daemon child exited before startup handoff: %w", err)
			}
			return server.StartupRecord{}, errors.New("daemon child exited before startup handoff")
		default:
		}

		data, err := os.ReadFile(path)
		if err == nil {
			var record server.StartupRecord
			if unmarshalErr := json.Unmarshal(data, &record); unmarshalErr != nil {
				return server.StartupRecord{}, fmt.Errorf("decode startup handoff: %w", unmarshalErr)
			}
			return record, nil
		}

		if !errors.Is(err, os.ErrNotExist) {
			return server.StartupRecord{}, fmt.Errorf("read startup handoff: %w", err)
		}
		if time.Now().After(deadline) {
			return server.StartupRecord{}, fmt.Errorf("timed out waiting for startup handoff file: %s", path)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func isLoopbackHost(host string) bool {
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

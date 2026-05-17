package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/lirrensi/looty/internal/server"
)

func TestPersistStartupOutputsWritesHandoffForDaemonChild(t *testing.T) {
	dir := t.TempDir()
	startupFile := filepath.Join(dir, "handoff.json")
	jsonFile := filepath.Join(dir, "record.json")
	record := server.NewStartupRecord("agent-managed", dir, "127.0.0.1", 41111, false, "", "", 1234, time.Unix(0, 0), nil)

	if err := persistStartupOutputs(cliOptions{startupFile: startupFile, jsonFile: jsonFile, daemonChild: true}, record); err != nil {
		t.Fatalf("persistStartupOutputs: %v", err)
	}
	if _, err := os.Stat(startupFile); err != nil {
		t.Fatalf("handoff file missing: %v", err)
	}
	if _, err := os.Stat(jsonFile); !os.IsNotExist(err) {
		t.Fatalf("json file should not be written for daemon child, got err=%v", err)
	}
}

func TestDaemonChildArgsIncludeNoTLSAndJSONHandoff(t *testing.T) {
	args := buildDaemonChildArgs(cliOptions{
		host:      "0.0.0.0",
		noTLSFlag: true,
		jsonFile:  "startup.json",
	}, "handoff.json")
	joined := strings.Join(args, " ")
	for _, expected := range []string{"-daemon-child", "-startup-file handoff.json", "-no-tls", "-json-file startup.json"} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("missing %q in %q", expected, joined)
		}
	}
}

func TestResolveTLSModePrecedence(t *testing.T) {
	if got := resolveTLSMode("127.0.0.1", true, false); !got {
		t.Fatalf("force TLS on loopback = %v, want true", got)
	}
	if got := resolveTLSMode("0.0.0.0", true, true); got {
		t.Fatalf("-no-tls should win over -tls, got %v", got)
	}
	if got := resolveTLSMode("127.0.0.1", false, false); got {
		t.Fatalf("loopback should default to HTTP, got %v", got)
	}
	if got := resolveTLSMode("", false, false); !got {
		t.Fatalf("all interfaces should default to TLS, got %v", got)
	}
}

func TestPrepareTLSRequiresBothCertificateAndKey(t *testing.T) {
	_, _, _, _, err := prepareTLS(cliOptions{certPath: "cert.pem"})
	if err == nil || !strings.Contains(err.Error(), "both -cert and -key are required together") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPrepareTLSReportsCertificateLoadErrors(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")
	if err := os.WriteFile(certPath, []byte("not a cert"), 0o644); err != nil {
		t.Fatalf("write cert: %v", err)
	}
	if err := os.WriteFile(keyPath, []byte("not a key"), 0o644); err != nil {
		t.Fatalf("write key: %v", err)
	}

	_, _, _, _, err := prepareTLS(cliOptions{certPath: certPath, keyPath: keyPath})
	if err == nil || !strings.Contains(err.Error(), "failed to load TLS certificate") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWaitForStartupRecordReportsMalformedJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "startup.json")
	if err := os.WriteFile(path, []byte("{not valid json"), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	_, err := waitForStartupRecord(path, nil, time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "decode startup handoff") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWaitForStartupRecordReportsChildExit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "startup.json")
	childExit := make(chan error, 1)
	childExit <- errors.New("boom")

	_, err := waitForStartupRecord(path, childExit, time.Second)
	if err == nil || !strings.Contains(err.Error(), "daemon child exited before startup handoff") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWaitForStartupRecordTimesOutWhenFileNeverAppears(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "startup.json")

	_, err := waitForStartupRecord(path, nil, time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "timed out waiting for startup handoff file") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPersistStartupOutputsSkipsJSONFileForDaemonChild(t *testing.T) {
	dir := t.TempDir()
	startupFile := filepath.Join(dir, "handoff.json")
	jsonFile := filepath.Join(dir, "record.json")
	record := server.NewStartupRecord("agent-managed", dir, "127.0.0.1", 41111, false, "", "", 1234, time.Unix(0, 0), nil)

	if err := persistStartupOutputs(cliOptions{startupFile: startupFile, jsonFile: jsonFile, daemonChild: true}, record); err != nil {
		t.Fatalf("persistStartupOutputs: %v", err)
	}
	if _, err := os.Stat(startupFile); err != nil {
		t.Fatalf("startup handoff missing: %v", err)
	}
	if _, err := os.Stat(jsonFile); !os.IsNotExist(err) {
		t.Fatalf("json file should not be written for daemon child, got err=%v", err)
	}
}

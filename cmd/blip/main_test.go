package main

import (
	"strings"
	"testing"
)

func TestBuildDaemonChildArgsAddsInternalFlagOnceAndPreservesFlags(t *testing.T) {
	t.Parallel()

	args := buildDaemonChildArgs(cliOptions{
		host:       "127.0.0.1",
		port:       41112,
		serveDir:   "/data/shared",
		useTLSFlag: true,
		certPath:   "cert.pem",
		keyPath:    "key.pem",
		json:       true,
		jsonFile:   "startup.json",
	}, "handoff.json")

	joined := strings.Join(args, " ")
	if strings.Count(joined, "-daemon-child") != 1 {
		t.Fatalf("expected exactly one -daemon-child, got %q", joined)
	}
	for _, expected := range []string{"-startup-file handoff.json", "-host 127.0.0.1", "-port 41112", "-serve-dir /data/shared", "-tls", "-cert cert.pem", "-key key.pem", "-json", "-json-file startup.json"} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("missing %q in args %q", expected, joined)
		}
	}
	if strings.Contains(joined, "-daemon ") {
		t.Fatalf("daemon flag should not be forwarded to child: %q", joined)
	}
}

func TestBuildDaemonChildArgsOmitsServeDirWhenEmpty(t *testing.T) {
	t.Parallel()

	args := buildDaemonChildArgs(cliOptions{
		port: 41111,
	}, "handoff.json")

	joined := strings.Join(args, " ")
	if strings.Contains(joined, "-serve-dir") {
		t.Fatalf("serve-dir should not appear when empty, got %q", joined)
	}
}

func TestDetermineMode(t *testing.T) {
	t.Parallel()

	if got := determineMode(cliOptions{}); got != "foreground" {
		t.Fatalf("foreground mode = %q", got)
	}
	if got := determineMode(cliOptions{daemon: true}); got != "background" {
		t.Fatalf("background mode = %q", got)
	}
	if got := determineMode(cliOptions{daemon: true, jsonFile: "startup.json"}); got != "agent-managed" {
		t.Fatalf("agent-managed mode = %q", got)
	}
}

func TestApplyExtractedHTMLPortInjectsConfiguredPort(t *testing.T) {
	t.Parallel()

	html := []byte("<html><head><title>Looty</title></head><body></body></html>")
	updated := string(applyExtractedHTMLPort(html, 41112))

	if !strings.Contains(updated, "window.__LOOTY_PORT__ = 41112;") {
		t.Fatalf("configured port missing from HTML: %s", updated)
	}
	if strings.Count(updated, "window.__LOOTY_PORT__") != 1 {
		t.Fatalf("configured port injected unexpected number of times: %s", updated)
	}
}

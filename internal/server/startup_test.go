package server

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestProtocolAndURLDerivation(t *testing.T) {
	t.Parallel()

	if got := ProtocolForTLS(false); got != "http" {
		t.Fatalf("ProtocolForTLS(false) = %q", got)
	}
	if got := ProtocolForTLS(true); got != "https" {
		t.Fatalf("ProtocolForTLS(true) = %q", got)
	}

	primary, urls := DeriveURLs("https", []string{"localhost", "127.0.0.1"}, 41111)
	if primary != "https://localhost:41111" {
		t.Fatalf("primary URL = %q", primary)
	}
	if len(urls) != 2 || urls[1] != "https://127.0.0.1:41111" {
		t.Fatalf("unexpected URLs: %#v", urls)
	}

	addresses := DeriveAddresses("127.0.0.1")
	if len(addresses) != 2 || addresses[0] != "localhost" || addresses[1] != "127.0.0.1" {
		t.Fatalf("unexpected loopback addresses: %#v", addresses)
	}
}

func TestStartupRecordJSONSerialization(t *testing.T) {
	t.Parallel()

	withTLS := NewStartupRecord(
		"agent-managed",
		"C:/loot",
		"127.0.0.1",
		41111,
		true,
		"AA:BB",
		"looty-brave-dolphin-4217",
		1234,
		time.Date(2026, time.May, 17, 12, 0, 0, 0, time.UTC),
		[]string{"C:/Users/rx/looty/looty.html"},
	)

	data, err := withTLS.JSON()
	if err != nil {
		t.Fatalf("JSON() error: %v", err)
	}

	jsonText := string(data)
	if !strings.Contains(jsonText, `"fingerprint": "AA:BB"`) {
		t.Fatalf("fingerprint missing from JSON: %s", jsonText)
	}
	if !strings.Contains(jsonText, `"friendCode": "looty-brave-dolphin-4217"`) {
		t.Fatalf("friend code missing from JSON: %s", jsonText)
	}

	withoutTLS := NewStartupRecord("foreground", "C:/loot", "127.0.0.1", 41111, false, "", "", 999, time.Now(), nil)
	data, err = withoutTLS.JSON()
	if err != nil {
		t.Fatalf("JSON() error: %v", err)
	}
	jsonText = string(data)
	if strings.Contains(jsonText, `"fingerprint"`) || strings.Contains(jsonText, `"friendCode"`) {
		t.Fatalf("unexpected TLS fields in JSON: %s", jsonText)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}
	if decoded["primaryUrl"] != "http://localhost:41111" {
		t.Fatalf("primaryUrl = %#v", decoded["primaryUrl"])
	}
}

func TestWriteStartupRecordFileIsAtomicReplacement(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "startup.json")
	record := NewStartupRecord("foreground", dir, "127.0.0.1", 41111, false, "", "", 42, time.Unix(0, 0), nil)

	if err := os.WriteFile(path, []byte("stale"), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}
	if err := WriteStartupRecordFile(path, record); err != nil {
		t.Fatalf("WriteStartupRecordFile error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	if strings.Contains(string(data), "stale") {
		t.Fatalf("startup record file still contains stale content: %s", string(data))
	}

	matches, err := filepath.Glob(filepath.Join(dir, "startup.json.*.tmp"))
	if err != nil {
		t.Fatalf("Glob error: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("temporary files left behind: %#v", matches)
	}
}

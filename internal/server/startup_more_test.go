package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCreateListenerRejectsInvalidAddress(t *testing.T) {
	_, err := CreateListener(Config{Host: "256.256.256.256", Port: 41111})
	if err == nil || !strings.Contains(err.Error(), "failed to create listener") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildHandlerCORSPreflight(t *testing.T) {
	handler := BuildHandler(Config{ServeDir: t.TempDir(), Port: 41111})
	req := httptest.NewRequest(http.MethodOptions, "/api/files", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("OPTIONS status = %d, want %d", rr.Code, http.StatusOK)
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("CORS origin = %q", got)
	}
	if got := rr.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(got, "OPTIONS") {
		t.Fatalf("CORS methods = %q", got)
	}
}

func TestRenderStartupRecordShowsHTTPSTrustDetails(t *testing.T) {
	record := StartupRecord{
		ServeDir:    t.TempDir(),
		Protocol:    "https",
		PrimaryURL:  "https://10.0.0.5:41111",
		Fingerprint: "AA:BB:CC",
		FriendCode:  "looty-brave-dolphin-4217",
		AllURLs:     []string{"https://10.0.0.5:41111"},
	}

	text := RenderStartupRecord(record)
	for _, expected := range []string{"Fingerprint: AA:BB:CC", "Friend code: looty-brave-dolphin-4217", "Verify the certificate fingerprint", "Share the direct link above"} {
		if !strings.Contains(text, expected) {
			t.Fatalf("missing %q in output:\n%s", expected, text)
		}
	}
}

func TestHandleSetScratchpadRejectsInvalidJSON(t *testing.T) {
	oldScratchpad := GetScratchpad()
	t.Cleanup(func() { SetScratchpad(oldScratchpad) })

	req := httptest.NewRequest(http.MethodPost, "/api/scratchpad", strings.NewReader("not json"))
	rr := httptest.NewRecorder()

	(&Server{}).handleSetScratchpad(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rr.Body.String(), "Invalid JSON") {
		t.Fatalf("response = %q", rr.Body.String())
	}
}

func TestHandleSetScratchpadBroadcastsUpdatedContent(t *testing.T) {
	oldHub := hub
	oldScratchpad := GetScratchpad()
	hub = &Hub{broadcast: make(chan []byte, 1)}
	t.Cleanup(func() {
		hub = oldHub
		SetScratchpad(oldScratchpad)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/scratchpad", strings.NewReader(`{"content":"hello"}`))
	rr := httptest.NewRecorder()

	(&Server{}).handleSetScratchpad(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if got := GetScratchpad(); got != "hello" {
		t.Fatalf("scratchpad = %q", got)
	}
	select {
	case msg := <-hub.broadcast:
		if string(msg) != `{"type":"scratchpad","data":"hello"}` {
			t.Fatalf("broadcast = %s", string(msg))
		}
	default:
		t.Fatal("expected scratchpad broadcast")
	}
}

func TestHandleGetScratchpadReturnsCurrentContent(t *testing.T) {
	oldScratchpad := GetScratchpad()
	SetScratchpad("saved value")
	t.Cleanup(func() { SetScratchpad(oldScratchpad) })

	req := httptest.NewRequest(http.MethodGet, "/api/scratchpad", nil)
	rr := httptest.NewRecorder()

	(&Server{}).handleGetScratchpad(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("content type = %q", got)
	}
	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["content"] != "saved value" {
		t.Fatalf("response = %#v", body)
	}
}

func TestWriteStartupRecordFileReportsDirectoryErrors(t *testing.T) {
	dir := t.TempDir()
	blocked := filepath.Join(dir, "blocked")
	if err := os.WriteFile(blocked, []byte("blocker"), 0o644); err != nil {
		t.Fatalf("create blocker: %v", err)
	}

	record := NewStartupRecord("foreground", dir, "127.0.0.1", 41111, false, "", "", 42, time.Unix(0, 0), nil)
	err := WriteStartupRecordFile(filepath.Join(blocked, "startup.json"), record)
	if err == nil || !strings.Contains(err.Error(), "create startup record directory") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWriteStartupQRCodeSVGFileRejectsEmptyPayload(t *testing.T) {
	err := WriteStartupQRCodeSVGFile(filepath.Join(t.TempDir(), "startup-qr.svg"), "")
	if err == nil || !strings.Contains(err.Error(), "empty payload") {
		t.Fatalf("unexpected error: %v", err)
	}
}

package files

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newUploadRequest(t *testing.T, destPath, filename string, content []byte) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("write file content: %v", err)
	}
	if err := writer.WriteField("path", destPath); err != nil {
		t.Fatalf("WriteField: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func TestListHandlerListsFilesAndDetectsBinaryContent(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatalf("write text file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "blob.bin"), []byte{0x01, 0x00, 0x02}, 0o644); err != nil {
		t.Fatalf("write binary file: %v", err)
	}
	if err := os.Mkdir(filepath.Join(dir, "folder"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/files?path=.", nil)
	rr := httptest.NewRecorder()

	ListHandler(dir).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	var body struct {
		Path  string     `json:"path"`
		Files []FileInfo `json:"files"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Path != "." {
		t.Fatalf("path = %q", body.Path)
	}

	filesByName := map[string]FileInfo{}
	for _, fi := range body.Files {
		filesByName[fi.Name] = fi
	}
	if filesByName["notes.txt"].IsBinary {
		t.Fatalf("notes.txt should be text: %#v", filesByName["notes.txt"])
	}
	if !filesByName["blob.bin"].IsBinary {
		t.Fatalf("blob.bin should be binary: %#v", filesByName["blob.bin"])
	}
	if !filesByName["folder"].IsDir {
		t.Fatalf("folder should be a directory: %#v", filesByName["folder"])
	}
}

func TestListHandlerRejectsTraversal(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/files?path=..", nil)
	rr := httptest.NewRecorder()

	ListHandler(t.TempDir()).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rr.Body.String(), "Invalid path") {
		t.Fatalf("response = %q", rr.Body.String())
	}
}

func TestListHandlerReportsMissingDirectory(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/files?path=missing", nil)
	rr := httptest.NewRecorder()

	ListHandler(t.TempDir()).ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
	if !strings.Contains(rr.Body.String(), "Failed to read directory") {
		t.Fatalf("response = %q", rr.Body.String())
	}
}

func TestDownloadHandlerRejectsMissingPath(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/download", nil)
	rr := httptest.NewRecorder()

	DownloadHandler(t.TempDir()).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rr.Body.String(), "Path required") {
		t.Fatalf("response = %q", rr.Body.String())
	}
}

func TestDownloadHandlerReturnsNotFoundAndDirectoryErrors(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "folder"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	t.Run("missing file", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/download?path=missing.txt", nil)
		rr := httptest.NewRecorder()

		DownloadHandler(dir).ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusNotFound)
		}
	})

	t.Run("directory", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/download?path=folder", nil)
		rr := httptest.NewRecorder()

		DownloadHandler(dir).ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
		if !strings.Contains(rr.Body.String(), "Cannot download directory") {
			t.Fatalf("response = %q", rr.Body.String())
		}
	})
}

func TestDownloadHandlerServesFileContents(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "notes.txt")
	if err := os.WriteFile(path, []byte("hello world"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/download?path=notes.txt", nil)
	rr := httptest.NewRecorder()

	DownloadHandler(dir).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if got := rr.Body.String(); got != "hello world" {
		t.Fatalf("body = %q", got)
	}
}

func TestUploadHandlerRejectsMissingFileAndTraversal(t *testing.T) {
	dir := t.TempDir()

	t.Run("missing file", func(t *testing.T) {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		if err := writer.WriteField("path", "."); err != nil {
			t.Fatalf("WriteField: %v", err)
		}
		if err := writer.Close(); err != nil {
			t.Fatalf("close writer: %v", err)
		}
		req := httptest.NewRequest(http.MethodPost, "/api/upload", &body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rr := httptest.NewRecorder()

		UploadHandler(dir).ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
		if !strings.Contains(rr.Body.String(), "No file provided") {
			t.Fatalf("response = %q", rr.Body.String())
		}
	})

	t.Run("traversal", func(t *testing.T) {
		req := newUploadRequest(t, "..", "evil.txt", []byte("bad"))
		rr := httptest.NewRecorder()

		UploadHandler(dir).ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
		if !strings.Contains(rr.Body.String(), "Invalid path") {
			t.Fatalf("response = %q", rr.Body.String())
		}
	})
}

func TestUploadHandlerWritesFileAndReturnsJSON(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "nested"), 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}
	req := newUploadRequest(t, "nested", "hello.txt", []byte("uploaded content"))
	rr := httptest.NewRecorder()

	UploadHandler(dir).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	var body map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["success"] != true {
		t.Fatalf("response = %#v", body)
	}
	if body["path"] != filepath.Join("nested", "hello.txt") {
		t.Fatalf("path = %#v", body["path"])
	}
	written, err := os.ReadFile(filepath.Join(dir, "nested", "hello.txt"))
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}
	if string(written) != "uploaded content" {
		t.Fatalf("written file = %q", string(written))
	}
}

package files

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type FileInfo struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	IsDir    bool   `json:"isDir"`
	Size     int64  `json:"size"`
	ModTime  string `json:"modTime"`
	IsBinary bool   `json:"isBinary"`
}

// isBinaryFile checks if a file is binary by reading the first 8KB and looking for null bytes
func isBinaryFile(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	buf := make([]byte, 8192)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return false
	}
	if n == 0 {
		return false // empty file = not binary
	}

	// Check for null bytes - definitive sign of binary
	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true
		}
	}
	return false
}

func ListHandler(serveDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get relative path from query, default to root
		relPath := r.URL.Query().Get("path")
		if relPath == "" {
			relPath = "."
		}

		// Security: prevent path traversal
		if strings.Contains(relPath, "..") {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		fullPath := filepath.Join(serveDir, relPath)

		// Verify path is within serveDir
		absServeDir, _ := filepath.Abs(serveDir)
		absFullPath, _ := filepath.Abs(fullPath)
		if !strings.HasPrefix(absFullPath, absServeDir) {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		entries, err := os.ReadDir(fullPath)
		if err != nil {
			http.Error(w, "Failed to read directory", http.StatusInternalServerError)
			return
		}

		files := make([]FileInfo, 0)
		for _, entry := range entries {
			info, err := entry.Info()
			if err != nil {
				continue
			}

			isBinary := false
			if !entry.IsDir() {
				isBinary = isBinaryFile(filepath.Join(fullPath, entry.Name()))
			}

			files = append(files, FileInfo{
				Name:     entry.Name(),
				Path:     filepath.Join(relPath, entry.Name()),
				IsDir:    entry.IsDir(),
				Size:     info.Size(),
				ModTime:  info.ModTime().Format("2006-01-02 15:04:05"),
				IsBinary: isBinary,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"path":  relPath,
			"files": files,
		})
	}
}

func DownloadHandler(serveDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		relPath := r.URL.Query().Get("path")
		if relPath == "" {
			http.Error(w, "Path required", http.StatusBadRequest)
			return
		}

		// Security: prevent path traversal
		if strings.Contains(relPath, "..") {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		fullPath := filepath.Join(serveDir, relPath)

		// Verify path is within serveDir
		absServeDir, _ := filepath.Abs(serveDir)
		absFullPath, _ := filepath.Abs(fullPath)
		if !strings.HasPrefix(absFullPath, absServeDir) {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		// Check if file exists and is not a directory
		info, err := os.Stat(fullPath)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		if info.IsDir() {
			http.Error(w, "Cannot download directory", http.StatusBadRequest)
			return
		}

		http.ServeFile(w, r, fullPath)
	}
}

func UploadHandler(serveDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Max 100MB upload
		r.ParseMultipartForm(100 << 20)

		file, handler, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "No file provided", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Get destination path
		destPath := r.FormValue("path")
		if destPath == "" {
			destPath = "."
		}

		// Security: prevent path traversal
		if strings.Contains(destPath, "..") || strings.Contains(handler.Filename, "..") {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		fullPath := filepath.Join(serveDir, destPath, handler.Filename)

		// Verify path is within serveDir
		absServeDir, _ := filepath.Abs(serveDir)
		absFullPath, _ := filepath.Abs(fullPath)
		if !strings.HasPrefix(absFullPath, absServeDir) {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		// Create destination file
		dst, err := os.Create(fullPath)
		if err != nil {
			http.Error(w, "Failed to create file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		// Copy file content
		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, "Failed to save file", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"path":    filepath.Join(destPath, handler.Filename),
		})
	}
}

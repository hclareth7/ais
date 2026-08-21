package fileserver

import (
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

var allowedImageExts = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
	".svg":  true,
	".webp": true,
	".bmp":  true,
	".ico":  true,
}

const maxFileSize = 10 * 1024 * 1024 // 10MB

// Handler serves local image files from the project root directory.
// It validates paths to prevent traversal attacks and restricts served
// files to known image extensions only.
type Handler struct {
	getRootPath func() string
}

// NewHandler creates a new file server handler. The getRootPath function
// is called on each request to obtain the current root directory path.
func NewHandler(getRootPath func() string) *Handler {
	return &Handler{getRootPath: getRootPath}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rawPath := strings.TrimPrefix(r.URL.Path, "/local/")
	if rawPath == "" || rawPath == r.URL.Path {
		http.NotFound(w, r)
		return
	}

	decoded, err := url.PathUnescape(rawPath)
	if err != nil {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	root := h.getRootPath()
	absPath := filepath.Join(root, decoded)
	resolved, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if !strings.HasPrefix(resolved, root+string(os.PathSeparator)) && resolved != root {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	ext := strings.ToLower(filepath.Ext(resolved))
	if !allowedImageExts[ext] {
		http.Error(w, "forbidden: not an image", http.StatusForbidden)
		return
	}

	info, err := os.Stat(resolved)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if info.IsDir() {
		http.NotFound(w, r)
		return
	}
	if info.Size() > maxFileSize {
		http.Error(w, fmt.Sprintf("file too large: %d bytes", info.Size()), http.StatusRequestEntityTooLarge)
		return
	}

	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if ext == ".svg" {
		w.Header().Set("Content-Security-Policy", "script-src 'none'; style-src 'unsafe-inline'")
	}

	http.ServeFile(w, r, resolved)
}

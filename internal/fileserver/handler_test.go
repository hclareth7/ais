package fileserver

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func setupTestRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	// Create a valid PNG file (minimal 1x1 PNG)
	pngData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, // IHDR chunk
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
		0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41,
		0x54, 0x08, 0xD7, 0x63, 0xF8, 0xCF, 0xC0, 0x00,
		0x00, 0x00, 0x02, 0x00, 0x01, 0xE2, 0x21, 0xBC,
		0x33, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E,
		0x44, 0xAE, 0x42, 0x60, 0x82,
	}
	if err := os.WriteFile(filepath.Join(root, "image.png"), pngData, 0o644); err != nil {
		t.Fatal(err)
	}

	// Create an SVG file
	svgData := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="1" height="1"></svg>`)
	if err := os.WriteFile(filepath.Join(root, "icon.svg"), svgData, 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a non-image file
	if err := os.WriteFile(filepath.Join(root, "data.json"), []byte(`{"key":"value"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a subdirectory
	subdir := filepath.Join(root, "images")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a nested image
	if err := os.WriteFile(filepath.Join(subdir, "nested.png"), pngData, 0o644); err != nil {
		t.Fatal(err)
	}

	return root
}

func TestPathTraversalDotDot(t *testing.T) {
	root := setupTestRoot(t)
	h := NewHandler(func() string { return root })

	req := httptest.NewRequest(http.MethodGet, "/local/../etc/passwd", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden && w.Code != http.StatusNotFound {
		t.Errorf("expected 403 or 404 for path traversal, got %d", w.Code)
	}
}

func TestPathTraversalEncoded(t *testing.T) {
	root := setupTestRoot(t)
	h := NewHandler(func() string { return root })

	req := httptest.NewRequest(http.MethodGet, "/local/%2e%2e%2fetc%2fpasswd", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden && w.Code != http.StatusNotFound {
		t.Errorf("expected 403 or 404 for encoded traversal, got %d", w.Code)
	}
}

func TestValidPNG(t *testing.T) {
	root := setupTestRoot(t)
	h := NewHandler(func() string { return root })

	req := httptest.NewRequest(http.MethodGet, "/local/image.png", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for valid PNG, got %d", w.Code)
	}
	ct := w.Header().Get("Content-Type")
	if ct != "image/png" {
		t.Errorf("expected Content-Type image/png, got %s", ct)
	}
}

func TestValidSVGHasCSPHeader(t *testing.T) {
	root := setupTestRoot(t)
	h := NewHandler(func() string { return root })

	req := httptest.NewRequest(http.MethodGet, "/local/icon.svg", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for valid SVG, got %d", w.Code)
	}
	csp := w.Header().Get("Content-Security-Policy")
	if csp == "" {
		t.Error("expected Content-Security-Policy header for SVG")
	}
}

func TestNonImageFileForbidden(t *testing.T) {
	root := setupTestRoot(t)
	h := NewHandler(func() string { return root })

	req := httptest.NewRequest(http.MethodGet, "/local/data.json", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for non-image file, got %d", w.Code)
	}
}

func TestMissingFile(t *testing.T) {
	root := setupTestRoot(t)
	h := NewHandler(func() string { return root })

	req := httptest.NewRequest(http.MethodGet, "/local/nonexistent.png", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for missing file, got %d", w.Code)
	}
}

func TestDirectoryNotFound(t *testing.T) {
	root := setupTestRoot(t)
	h := NewHandler(func() string { return root })

	req := httptest.NewRequest(http.MethodGet, "/local/images", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	// Directory should be rejected — either not found or forbidden (no image ext)
	if w.Code != http.StatusNotFound && w.Code != http.StatusForbidden {
		t.Errorf("expected 404 or 403 for directory, got %d", w.Code)
	}
}

func TestNosniffHeaderPresent(t *testing.T) {
	root := setupTestRoot(t)
	h := NewHandler(func() string { return root })

	tests := []struct {
		name string
		path string
	}{
		{"png", "/local/image.png"},
		{"svg", "/local/icon.svg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)

			nosniff := w.Header().Get("X-Content-Type-Options")
			if nosniff != "nosniff" {
				t.Errorf("expected X-Content-Type-Options: nosniff, got %q", nosniff)
			}
		})
	}
}

func TestMethodNotAllowed(t *testing.T) {
	root := setupTestRoot(t)
	h := NewHandler(func() string { return root })

	req := httptest.NewRequest(http.MethodPost, "/local/image.png", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 for POST, got %d", w.Code)
	}
}

func TestEmptyPath(t *testing.T) {
	root := setupTestRoot(t)
	h := NewHandler(func() string { return root })

	req := httptest.NewRequest(http.MethodGet, "/local/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for empty path, got %d", w.Code)
	}
}

func TestNestedPath(t *testing.T) {
	root := setupTestRoot(t)
	h := NewHandler(func() string { return root })

	req := httptest.NewRequest(http.MethodGet, "/local/images/nested.png", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for nested image, got %d", w.Code)
	}
}

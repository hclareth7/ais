package search

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Create directory structure:
	// root/
	//   README.md         "Hello world, this is a test document"
	//   docs/
	//     guide.md        "Guide for beginners\nAnother hello line"
	//     deep/
	//       nested.md     "Deeply nested content with hello"
	//   .git/
	//     HEAD            (skipped dir)
	//   node_modules/
	//     pkg.md          (skipped dir)
	//   src/
	//     main.go         (not markdown, ignored)

	os.MkdirAll(filepath.Join(dir, "docs", "deep"), 0o755)
	os.MkdirAll(filepath.Join(dir, ".git"), 0o755)
	os.MkdirAll(filepath.Join(dir, "node_modules"), 0o755)
	os.MkdirAll(filepath.Join(dir, "src"), 0o755)

	os.WriteFile(filepath.Join(dir, "README.md"), []byte("Hello world, this is a test document"), 0o644)
	os.WriteFile(filepath.Join(dir, "docs", "guide.md"), []byte("Guide for beginners\nAnother hello line"), 0o644)
	os.WriteFile(filepath.Join(dir, "docs", "deep", "nested.md"), []byte("Deeply nested content with hello"), 0o644)
	os.WriteFile(filepath.Join(dir, ".git", "HEAD"), []byte("ref: refs/heads/main"), 0o644)
	os.WriteFile(filepath.Join(dir, "node_modules", "pkg.md"), []byte("Package with hello"), 0o644)
	os.WriteFile(filepath.Join(dir, "src", "main.go"), []byte("package main // hello"), 0o644)

	return dir
}

var defaultSkipDirs = []string{".git", "node_modules", "vendor", ".svn", "__pycache__", ".venv"}

func TestSearchFiles(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T) string
		query      string
		skipDirs   []string
		wantCount  int
		wantErr    bool
		validate   func(t *testing.T, results []SearchResult)
	}{
		{
			name:      "basic match",
			setup:     setupTestDir,
			query:     "test document",
			skipDirs:  defaultSkipDirs,
			wantCount: 1,
			validate: func(t *testing.T, results []SearchResult) {
				t.Helper()
				r := results[0]
				if r.FilePath != "README.md" {
					t.Errorf("expected FilePath README.md, got %s", r.FilePath)
				}
				if r.LineNumber != 1 {
					t.Errorf("expected LineNumber 1, got %d", r.LineNumber)
				}
				if !strings.Contains(r.Context, "test document") {
					t.Errorf("expected context to contain 'test document', got %q", r.Context)
				}
			},
		},
		{
			name: "multi-match across files",
			setup: setupTestDir,
			query: "hello",
			skipDirs: defaultSkipDirs,
			wantCount: 3, // README.md, guide.md, nested.md (not .git, node_modules, src)
			validate: func(t *testing.T, results []SearchResult) {
				t.Helper()
				paths := make(map[string]bool)
				for _, r := range results {
					paths[r.FilePath] = true
				}
				for _, expected := range []string{"README.md", filepath.Join("docs", "guide.md"), filepath.Join("docs", "deep", "nested.md")} {
					if !paths[expected] {
						t.Errorf("expected result from %s", expected)
					}
				}
			},
		},
		{
			name:      "case-insensitive",
			setup:     setupTestDir,
			query:     "HELLO",
			skipDirs:  defaultSkipDirs,
			wantCount: 3,
		},
		{
			name:      "no match",
			setup:     setupTestDir,
			query:     "nonexistent_xyz_query",
			skipDirs:  defaultSkipDirs,
			wantCount: 0,
		},
		{
			name: "skipped dirs",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				os.MkdirAll(filepath.Join(dir, "vendor"), 0o755)
				os.WriteFile(filepath.Join(dir, "vendor", "lib.md"), []byte("vendor hello"), 0o644)
				os.WriteFile(filepath.Join(dir, "top.md"), []byte("top hello"), 0o644)
				return dir
			},
			query:     "hello",
			skipDirs:  []string{"vendor"},
			wantCount: 1,
			validate: func(t *testing.T, results []SearchResult) {
				t.Helper()
				if results[0].FilePath != "top.md" {
					t.Errorf("expected top.md, got %s", results[0].FilePath)
				}
			},
		},
		{
			name:      "nested dirs",
			setup:     setupTestDir,
			query:     "nested content",
			skipDirs:  defaultSkipDirs,
			wantCount: 1,
			validate: func(t *testing.T, results []SearchResult) {
				t.Helper()
				expected := filepath.Join("docs", "deep", "nested.md")
				if results[0].FilePath != expected {
					t.Errorf("expected %s, got %s", expected, results[0].FilePath)
				}
			},
		},
		{
			name: "special chars treated as literal",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				os.WriteFile(filepath.Join(dir, "special.md"), []byte("version 1.2.3 is here"), 0o644)
				return dir
			},
			query:     "1.2.3",
			skipDirs:  defaultSkipDirs,
			wantCount: 1,
			validate: func(t *testing.T, results []SearchResult) {
				t.Helper()
				if !strings.Contains(results[0].Context, "1.2.3") {
					t.Errorf("expected context to contain '1.2.3', got %q", results[0].Context)
				}
			},
		},
		{
			name: "max results capped at 50",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				// Create a file with 60 occurrences of "match"
				lines := make([]string, 60)
				for i := range lines {
					lines[i] = "this is a match here"
				}
				content := strings.Join(lines, "\n")
				os.WriteFile(filepath.Join(dir, "many.md"), []byte(content), 0o644)
				return dir
			},
			query:     "match",
			skipDirs:  defaultSkipDirs,
			wantCount: 50,
		},
		{
			name:      "empty query",
			setup:     setupTestDir,
			query:     "",
			skipDirs:  defaultSkipDirs,
			wantCount: 0,
		},
		{
			name: "line number tracking",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				os.WriteFile(filepath.Join(dir, "lines.md"), []byte("line one\nline two\ntarget word\nline four"), 0o644)
				return dir
			},
			query:     "target",
			skipDirs:  defaultSkipDirs,
			wantCount: 1,
			validate: func(t *testing.T, results []SearchResult) {
				t.Helper()
				if results[0].LineNumber != 3 {
					t.Errorf("expected line 3, got %d", results[0].LineNumber)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup(t)
			results, err := SearchFiles(dir, tt.query, tt.skipDirs)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(results) != tt.wantCount {
				t.Errorf("expected %d results, got %d", tt.wantCount, len(results))
				for i, r := range results {
					t.Logf("  result[%d]: %s:%d %q", i, r.FilePath, r.LineNumber, r.Context)
				}
			}

			if tt.validate != nil && len(results) == tt.wantCount {
				tt.validate(t, results)
			}
		})
	}
}

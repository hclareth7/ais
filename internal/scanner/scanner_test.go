package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Create directory structure:
	// root/
	//   README.md
	//   notes.markdown
	//   docs/
	//     guide.md
	//     api.md
	//   .git/
	//     config
	//   node_modules/
	//     pkg.md
	//   src/
	//     main.go  (not markdown)

	os.MkdirAll(filepath.Join(dir, "docs"), 0o755)
	os.MkdirAll(filepath.Join(dir, ".git"), 0o755)
	os.MkdirAll(filepath.Join(dir, "node_modules"), 0o755)
	os.MkdirAll(filepath.Join(dir, "src"), 0o755)

	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Hello"), 0o644)
	os.WriteFile(filepath.Join(dir, "notes.markdown"), []byte("# Notes"), 0o644)
	os.WriteFile(filepath.Join(dir, "docs", "guide.md"), []byte("# Guide"), 0o644)
	os.WriteFile(filepath.Join(dir, "docs", "api.md"), []byte("# API"), 0o644)
	os.WriteFile(filepath.Join(dir, ".git", "config"), []byte("[core]"), 0o644)
	os.WriteFile(filepath.Join(dir, "node_modules", "pkg.md"), []byte("# Pkg"), 0o644)
	os.WriteFile(filepath.Join(dir, "src", "main.go"), []byte("package main"), 0o644)

	return dir
}

func TestScanDirectory(t *testing.T) {
	dir := setupTestDir(t)

	tree, err := ScanDirectory(dir)
	if err != nil {
		t.Fatalf("ScanDirectory failed: %v", err)
	}

	if tree == nil {
		t.Fatal("tree is nil")
	}
	if !tree.IsDir {
		t.Error("root should be a directory")
	}

	// Expect: docs/ folder, then README.md, notes.markdown
	// .git, node_modules, src (no .md files) should be excluded
	if len(tree.Children) != 3 {
		t.Errorf("expected 3 children, got %d", len(tree.Children))
		for _, c := range tree.Children {
			t.Logf("  child: %s (isDir=%v)", c.Name, c.IsDir)
		}
		return
	}

	// Folders first
	if tree.Children[0].Name != "docs" || !tree.Children[0].IsDir {
		t.Errorf("expected first child to be docs/, got %s", tree.Children[0].Name)
	}

	// docs should have 2 files sorted alphabetically
	docs := tree.Children[0]
	if len(docs.Children) != 2 {
		t.Errorf("expected 2 files in docs/, got %d", len(docs.Children))
	} else {
		if docs.Children[0].Name != "api.md" {
			t.Errorf("expected api.md first in docs/, got %s", docs.Children[0].Name)
		}
		if docs.Children[1].Name != "guide.md" {
			t.Errorf("expected guide.md second in docs/, got %s", docs.Children[1].Name)
		}
	}

	// Then files sorted alphabetically
	if tree.Children[1].Name != "notes.markdown" {
		t.Errorf("expected notes.markdown, got %s", tree.Children[1].Name)
	}
	if tree.Children[2].Name != "README.md" {
		t.Errorf("expected README.md, got %s", tree.Children[2].Name)
	}
}

func TestScanDirectoryNotFound(t *testing.T) {
	_, err := ScanDirectory("/nonexistent/path/xyz")
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}

func TestReadFileContent(t *testing.T) {
	dir := setupTestDir(t)

	content, err := ReadFileContent(filepath.Join(dir, "README.md"))
	if err != nil {
		t.Fatalf("ReadFileContent failed: %v", err)
	}
	if content != "# Hello" {
		t.Errorf("expected '# Hello', got %q", content)
	}
}

func TestReadFileContentNotFound(t *testing.T) {
	_, err := ReadFileContent("/nonexistent/file.md")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestReadFileContentDirectory(t *testing.T) {
	dir := setupTestDir(t)

	_, err := ReadFileContent(dir)
	if err == nil {
		t.Error("expected error when reading a directory")
	}
}

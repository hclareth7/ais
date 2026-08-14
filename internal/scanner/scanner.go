package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hclareth7/ais/internal/types"
)

const maxFileSize = 10 * 1024 * 1024 // 10MB

var skipDirs = map[string]bool{
	".git":        true,
	"node_modules": true,
	".svn":        true,
	"__pycache__": true,
	"vendor":      true,
	".venv":       true,
}

func isMarkdown(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasSuffix(lower, ".md") || strings.HasSuffix(lower, ".markdown")
}

func isHidden(name string) bool {
	return len(name) > 0 && name[0] == '.'
}

func ScanDirectory(root string) (*types.FileNode, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("cannot access path: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", root)
	}

	node := &types.FileNode{
		Name:  filepath.Base(root),
		Path:  "",
		IsDir: true,
	}
	node.Children, err = scanChildren(root, "")
	if err != nil {
		return nil, err
	}
	return node, nil
}

func scanChildren(absDir, relDir string) ([]*types.FileNode, error) {
	entries, err := os.ReadDir(absDir)
	if err != nil {
		return nil, fmt.Errorf("cannot read directory: %w", err)
	}

	var dirs, files []*types.FileNode

	for _, entry := range entries {
		name := entry.Name()

		if skipDirs[name] {
			continue
		}

		relPath := filepath.Join(relDir, name)

		if entry.IsDir() {
			children, err := scanChildren(filepath.Join(absDir, name), relPath)
			if err != nil {
				continue
			}
			if len(children) > 0 {
				dirs = append(dirs, &types.FileNode{
					Name:     name,
					Path:     relPath,
					IsDir:    true,
					Children: children,
				})
			}
		} else if isMarkdown(name) {
			files = append(files, &types.FileNode{
				Name:  name,
				Path:  relPath,
				IsDir: false,
			})
		}
	}

	sort.Slice(dirs, func(i, j int) bool {
		return strings.ToLower(dirs[i].Name) < strings.ToLower(dirs[j].Name)
	})
	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})

	result := make([]*types.FileNode, 0, len(dirs)+len(files))
	result = append(result, dirs...)
	result = append(result, files...)
	return result, nil
}

func ReadFileContent(absPath string) (string, error) {
	info, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("file not found: %w", err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("path is a directory: %s", absPath)
	}
	if info.Size() > maxFileSize {
		return "", fmt.Errorf("file too large: %d bytes (max %d)", info.Size(), maxFileSize)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("cannot read file: %w", err)
	}
	return string(data), nil
}

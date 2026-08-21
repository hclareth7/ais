package search

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	maxFileSize  = 10 * 1024 * 1024 // 10MB
	maxResults   = 50
	contextChars = 30
)

// SearchResult represents a single search match within a markdown file.
type SearchResult struct {
	FilePath    string `json:"filePath"`
	LineNumber  int    `json:"lineNumber"`
	MatchOffset int    `json:"matchOffset"`
	Context     string `json:"context"`
}

// SearchFiles searches all markdown files under root for the given query.
// Matching is case-insensitive. Returns up to 50 results with context snippets.
// Directories in skipDirs are excluded from the walk.
func SearchFiles(root, query string, skipDirs []string) ([]SearchResult, error) {
	if query == "" {
		return []SearchResult{}, nil
	}

	skip := make(map[string]bool, len(skipDirs))
	for _, d := range skipDirs {
		skip[d] = true
	}

	lowerQuery := strings.ToLower(query)
	var results []SearchResult

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip inaccessible entries
		}

		if d.IsDir() {
			if skip[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		if !isMarkdown(d.Name()) {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}
		if info.Size() > maxFileSize {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		content := string(data)
		lowerContent := strings.ToLower(content)

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}

		offset := 0
		for {
			if len(results) >= maxResults {
				return filepath.SkipAll
			}

			idx := strings.Index(lowerContent[offset:], lowerQuery)
			if idx < 0 {
				break
			}

			matchOffset := offset + idx
			lineNumber := countLines(content, matchOffset)
			context := buildContext(content, matchOffset, len(query))

			results = append(results, SearchResult{
				FilePath:    relPath,
				LineNumber:  lineNumber,
				MatchOffset: matchOffset,
				Context:     context,
			})

			offset = matchOffset + len(query)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("search files: %w", err)
	}

	if results == nil {
		results = []SearchResult{}
	}

	return results, nil
}

// countLines returns the 1-based line number for a byte offset in content.
func countLines(content string, offset int) int {
	return strings.Count(content[:offset], "\n") + 1
}

// buildContext extracts a snippet around the match, approximately contextChars
// before and after, trimmed to word boundaries when possible.
func buildContext(content string, matchOffset, matchLen int) string {
	start := max(0, matchOffset-contextChars)
	end := min(len(content), matchOffset+matchLen+contextChars)

	// Trim to word boundaries if we're not at the edges
	if start > 0 {
		// Find the next space after start to trim to a word boundary
		spaceIdx := strings.IndexByte(content[start:matchOffset], ' ')
		if spaceIdx >= 0 {
			start = start + spaceIdx + 1
		}
	}

	if end < len(content) {
		// Find the last space before end to trim to a word boundary
		segment := content[matchOffset+matchLen : end]
		spaceIdx := strings.LastIndexByte(segment, ' ')
		if spaceIdx >= 0 {
			end = matchOffset + matchLen + spaceIdx
		}
	}

	snippet := content[start:end]

	// Replace newlines with spaces for a clean single-line context
	snippet = strings.ReplaceAll(snippet, "\n", " ")
	snippet = strings.ReplaceAll(snippet, "\r", "")

	// Add ellipsis indicators
	prefix := ""
	suffix := ""
	if start > 0 {
		prefix = "..."
	}
	if end < len(content) {
		suffix = "..."
	}

	return prefix + snippet + suffix
}

func isMarkdown(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasSuffix(lower, ".md") || strings.HasSuffix(lower, ".markdown")
}

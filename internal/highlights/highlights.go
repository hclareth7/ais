package highlights

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Highlight represents a single text highlight within a markdown file.
type Highlight struct {
	ID            string    `json:"id"`
	FilePath      string    `json:"filePath"`
	AnchorText    string    `json:"anchorText"`
	PrefixContext string    `json:"prefixContext"`
	SuffixContext string    `json:"suffixContext"`
	Color         string    `json:"color"`
	CreatedAt     string    `json:"createdAt"`
}

// Store manages highlight persistence for markdown files. Highlights are
// stored as JSON files that mirror the source directory structure under
// a base directory. For example, docs/arch.md highlights are stored at
// <baseDir>/docs/arch.md.json.
type Store struct {
	mu      sync.RWMutex
	baseDir string
}

// NewStore creates a highlight store rooted at the given base directory.
func NewStore(baseDir string) *Store {
	return &Store{baseDir: baseDir}
}

// SetRoot updates the base directory for highlight storage. This is
// called when the user opens a different project root.
func (s *Store) SetRoot(baseDir string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.baseDir = baseDir
}

// storagePath returns the JSON file path for a given markdown file path.
func (s *Store) storagePath(filePath string) string {
	return filepath.Join(s.baseDir, filePath+".json")
}

// Load reads all highlights for a given file path. Returns an empty
// slice (not an error) if the storage file does not exist.
func (s *Store) Load(filePath string) ([]Highlight, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := s.storagePath(filePath)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Highlight{}, nil
		}
		return nil, fmt.Errorf("read highlights: %w", err)
	}

	var highlights []Highlight
	if err := json.Unmarshal(data, &highlights); err != nil {
		return nil, fmt.Errorf("parse highlights: %w", err)
	}
	return highlights, nil
}

// Save writes the highlight list to disk, creating subdirectories as needed.
func (s *Store) Save(filePath string, highlights []Highlight) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := s.storagePath(filePath)

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create highlight dir: %w", err)
	}

	data, err := json.MarshalIndent(highlights, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal highlights: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write highlights: %w", err)
	}
	return nil
}

// Add appends a highlight to the given file's highlight list and persists it.
func (s *Store) Add(h Highlight) error {
	// Use Load (which acquires RLock) then Save (which acquires Lock).
	// We must not hold a lock across both calls.
	existing, err := s.Load(h.FilePath)
	if err != nil {
		return fmt.Errorf("add highlight: %w", err)
	}
	existing = append(existing, h)
	return s.Save(h.FilePath, existing)
}

// Remove deletes a highlight by ID from the given file's highlight list.
func (s *Store) Remove(filePath, highlightID string) error {
	existing, err := s.Load(filePath)
	if err != nil {
		return fmt.Errorf("remove highlight: %w", err)
	}

	filtered := make([]Highlight, 0, len(existing))
	for _, h := range existing {
		if h.ID != highlightID {
			filtered = append(filtered, h)
		}
	}

	return s.Save(filePath, filtered)
}

// Clear removes all highlights for a given file by deleting the storage file.
func (s *Store) Clear(filePath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := s.storagePath(filePath)
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("clear highlights: %w", err)
	}
	return nil
}

package highlights

import (
	"path/filepath"
	"testing"
	"time"
)

func nowISO() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func newTestStore(t *testing.T) *Store {
	t.Helper()
	return NewStore(t.TempDir())
}

func TestEmptyLoadReturnsEmptySlice(t *testing.T) {
	store := newTestStore(t)

	highlights, err := store.Load("nonexistent.md")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(highlights) != 0 {
		t.Errorf("expected empty slice, got %d items", len(highlights))
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	store := newTestStore(t)

	input := []Highlight{
		{
			ID:            "h1",
			FilePath:      "readme.md",
			AnchorText:    "important text",
			PrefixContext: "some ",
			SuffixContext: " here",
			Color:         "yellow",
			CreatedAt:     nowISO(),
		},
		{
			ID:            "h2",
			FilePath:      "readme.md",
			AnchorText:    "another highlight",
			PrefixContext: "with ",
			SuffixContext: " context",
			Color:         "blue",
			CreatedAt:     nowISO(),
		},
	}

	if err := store.Save("readme.md", input); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := store.Load("readme.md")
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if len(loaded) != len(input) {
		t.Fatalf("expected %d highlights, got %d", len(input), len(loaded))
	}

	for i, h := range loaded {
		if h.ID != input[i].ID {
			t.Errorf("highlight %d: expected ID %q, got %q", i, input[i].ID, h.ID)
		}
		if h.AnchorText != input[i].AnchorText {
			t.Errorf("highlight %d: expected AnchorText %q, got %q", i, input[i].AnchorText, h.AnchorText)
		}
		if h.Color != input[i].Color {
			t.Errorf("highlight %d: expected Color %q, got %q", i, input[i].Color, h.Color)
		}
	}
}

func TestAddAppendsHighlight(t *testing.T) {
	store := newTestStore(t)

	h1 := Highlight{ID: "h1", FilePath: "test.md", AnchorText: "first", Color: "yellow"}
	h2 := Highlight{ID: "h2", FilePath: "test.md", AnchorText: "second", Color: "blue"}

	if err := store.Add(h1); err != nil {
		t.Fatalf("add h1 failed: %v", err)
	}
	if err := store.Add(h2); err != nil {
		t.Fatalf("add h2 failed: %v", err)
	}

	loaded, err := store.Load("test.md")
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(loaded) != 2 {
		t.Fatalf("expected 2 highlights, got %d", len(loaded))
	}
	if loaded[0].ID != "h1" || loaded[1].ID != "h2" {
		t.Errorf("expected IDs h1, h2; got %s, %s", loaded[0].ID, loaded[1].ID)
	}
}

func TestRemoveByID(t *testing.T) {
	store := newTestStore(t)

	h1 := Highlight{ID: "h1", FilePath: "test.md", AnchorText: "keep", Color: "yellow"}
	h2 := Highlight{ID: "h2", FilePath: "test.md", AnchorText: "remove", Color: "red"}
	h3 := Highlight{ID: "h3", FilePath: "test.md", AnchorText: "keep too", Color: "blue"}

	for _, h := range []Highlight{h1, h2, h3} {
		if err := store.Add(h); err != nil {
			t.Fatalf("add failed: %v", err)
		}
	}

	if err := store.Remove("test.md", "h2"); err != nil {
		t.Fatalf("remove failed: %v", err)
	}

	loaded, err := store.Load("test.md")
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(loaded) != 2 {
		t.Fatalf("expected 2 highlights after remove, got %d", len(loaded))
	}
	for _, h := range loaded {
		if h.ID == "h2" {
			t.Error("highlight h2 should have been removed")
		}
	}
}

func TestClearDeletesFile(t *testing.T) {
	store := newTestStore(t)

	h := Highlight{ID: "h1", FilePath: "test.md", AnchorText: "text", Color: "yellow"}
	if err := store.Add(h); err != nil {
		t.Fatalf("add failed: %v", err)
	}

	if err := store.Clear("test.md"); err != nil {
		t.Fatalf("clear failed: %v", err)
	}

	loaded, err := store.Load("test.md")
	if err != nil {
		t.Fatalf("load after clear failed: %v", err)
	}
	if len(loaded) != 0 {
		t.Errorf("expected 0 highlights after clear, got %d", len(loaded))
	}
}

func TestClearNonexistentFileNoError(t *testing.T) {
	store := newTestStore(t)

	if err := store.Clear("nonexistent.md"); err != nil {
		t.Errorf("clear of nonexistent file should not error, got: %v", err)
	}
}

func TestNestedPathsCreateSubdirs(t *testing.T) {
	store := newTestStore(t)

	h := Highlight{
		ID:         "h1",
		FilePath:   filepath.Join("docs", "nested", "deep.md"),
		AnchorText: "deep text",
		Color:      "green",
	}

	if err := store.Add(h); err != nil {
		t.Fatalf("add nested highlight failed: %v", err)
	}

	loaded, err := store.Load(filepath.Join("docs", "nested", "deep.md"))
	if err != nil {
		t.Fatalf("load nested failed: %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("expected 1 highlight, got %d", len(loaded))
	}
	if loaded[0].AnchorText != "deep text" {
		t.Errorf("expected anchor 'deep text', got %q", loaded[0].AnchorText)
	}
}

func TestSetRootUpdatesBase(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()
	store := NewStore(dir1)

	h := Highlight{ID: "h1", FilePath: "test.md", AnchorText: "text", Color: "yellow"}
	if err := store.Add(h); err != nil {
		t.Fatalf("add in dir1 failed: %v", err)
	}

	store.SetRoot(dir2)

	// After SetRoot, the old data should not be visible
	loaded, err := store.Load("test.md")
	if err != nil {
		t.Fatalf("load in dir2 failed: %v", err)
	}
	if len(loaded) != 0 {
		t.Errorf("expected 0 highlights in new root, got %d", len(loaded))
	}

	// Add in new root
	h2 := Highlight{ID: "h2", FilePath: "test.md", AnchorText: "new root", Color: "blue"}
	if err := store.Add(h2); err != nil {
		t.Fatalf("add in dir2 failed: %v", err)
	}

	loaded, err = store.Load("test.md")
	if err != nil {
		t.Fatalf("load in dir2 after add failed: %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("expected 1 highlight in new root, got %d", len(loaded))
	}
}

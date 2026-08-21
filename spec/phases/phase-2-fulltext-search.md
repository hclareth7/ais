# Phase 2: Full-text Search

**Status:** COMPLETE (v1.3.0, PR #20)
**Branch:** `feat/fulltext-search`

## Feature

Search text across all markdown files in the current path. Two modes:

- **Ctrl+F** — In-file search: inline bar at top of viewer, highlights all matches, prev/next navigation, current match visually distinct. Highlights persist until Escape or clearing the query.
- **Ctrl+Shift+F** — Cross-file search: opens Command Palette in Search tab. Results show filename, line number, and context snippet. Selecting a result opens the file and scrolls to the match with all occurrences highlighted.

## Tasks (all complete)

1. Go backend: `internal/search/` package — case-insensitive matching, context snippets, 50-result cap
2. Go backend: `SearchFiles` Wails binding in `app.go`
3. Wails bindings: `App.js`, `App.d.ts`, `models.ts`
4. Frontend: `InFileSearch.svelte` — inline bar with input, X/Y counter, prev/next/close
5. Frontend: In-file search logic in `MarkdownViewer.svelte` — TreeWalker DOM search, match cycling
6. Frontend: Command Palette Search tab in `CommandPalette.svelte` — debounced cross-file search
7. Frontend: `search.ts` store (`searchScrollTarget`, `inFileSearchOpen`)
8. Frontend: `commandPaletteCategory` store in `ui.ts` for Ctrl+Shift+F routing
9. CSS: `--search-match-current-bg` token + `.search-match-current` rule
10. CI fix: Build tags for Unix-only pipe code + `fail-fast: false` in release workflow

## Files changed

- `internal/search/search.go` (new)
- `internal/search/search_test.go` (new)
- `internal/input/pipe.go` (build tag)
- `internal/input/pipe_test.go` (build tag)
- `internal/input/pipe_windows.go` (new, stub)
- `app.go`
- `frontend/src/lib/components/InFileSearch.svelte` (new)
- `frontend/src/lib/components/MarkdownViewer.svelte`
- `frontend/src/lib/components/CommandPalette.svelte`
- `frontend/src/lib/stores/search.ts` (new)
- `frontend/src/lib/stores/ui.ts`
- `frontend/src/App.svelte`
- `frontend/src/style.css`
- `.github/workflows/release.yml`

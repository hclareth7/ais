# Phase 4: Open MD File with Ctrl+O

**Status:** PENDING
**Branch:** `feat/open-file-ctrl-o`
**Release:** v1.5.0

## Feature

When the user presses Ctrl+O, a native file picker dialog opens. If the user selects a file, the application sets the root path to the parent directory of the selected file, loading all `.md` files from that folder into the sidebar file tree.

This differs from `OpenFolder` (which opens a directory picker). Ctrl+O opens a file picker filtered to markdown files, then derives the folder from the selected file's parent directory.

## UX Flow

1. User presses `Ctrl+O`
2. Native file picker dialog opens (filtered to `.md` files)
3. User selects a markdown file (or cancels)
4. If cancelled: no-op
5. If file selected: root path is set to the file's parent directory, sidebar refreshes, selected file opens as a tab

---

## Tasks

### Task 4.1: Add `OpenFile` method to Go backend

**Files:** `app.go`

Add a new Wails-bound method:
```go
func (a *App) OpenFile() (string, error)
```

- Uses `wailsRuntime.OpenFileDialog` with:
  - Title: `"Open File"`
  - Filters: `[]wailsRuntime.FileFilter{{DisplayName: "Markdown Files", Pattern: "*.md;*.markdown"}}`
- If user cancels (empty path), return `("", nil)`
- Extract parent directory via `filepath.Dir(selectedPath)`
- Call `a.SetRootPath(parentDir)` to update root, watcher, config
- Return the relative path of the selected file within the new root (for tab opening)

**Acceptance:**
- [ ] Ctrl+O triggers native file picker dialog
- [ ] File picker is filtered to markdown files
- [ ] Cancel returns empty string, no error
- [ ] Selecting a file sets root path to parent directory
- [ ] Returns relative path of the selected file

---

### Task 4.2: Add Wails binding for `OpenFile`

**Files:** `frontend/wailsjs/go/main/App.js`, `frontend/wailsjs/go/main/App.d.ts`

Add the `OpenFile` binding manually:
- `App.js`: `export function OpenFile() { return window['go']['main']['App']['OpenFile'](); }`
- `App.d.ts`: `export function OpenFile(): Promise<string>;`

**Acceptance:**
- [ ] `OpenFile` is callable from frontend
- [ ] Returns the relative file path as a string

---

### Task 4.3: Add Ctrl+O keyboard handler in App.svelte

**Files:** `frontend/src/App.svelte`

Add a `Ctrl+O` handler in the `handleKeydown` function (before the command palette guard, alongside `Ctrl+K`):

```typescript
if (e.ctrlKey && (e.key === 'o' || e.key === 'O') && !e.shiftKey) {
  e.preventDefault();
  // Call OpenFile, reload file tree, open selected file as tab
}
```

- Call `App.OpenFile()` via dynamic import
- If result is non-empty:
  - Call `loadFileTree()` to refresh sidebar
  - Open the returned file path as a tab via `openTab(filePath, filename)`
- Prevent default to avoid browser "Open File" behavior

Note: `Ctrl+Shift+O` is already bound to toggle TOC panel. This uses `Ctrl+O` (no shift).

**Acceptance:**
- [ ] Ctrl+O opens native file picker
- [ ] Selected file opens as a tab
- [ ] Sidebar refreshes with new folder contents
- [ ] Ctrl+Shift+O (TOC toggle) still works
- [ ] No-op on cancel
- [ ] Does not fire when command palette is open

---

## File Summary

| File | Change |
|------|--------|
| `app.go` | Add `OpenFile()` Wails-bound method |
| `frontend/wailsjs/go/main/App.js` | Add `OpenFile` binding |
| `frontend/wailsjs/go/main/App.d.ts` | Add `OpenFile` type |
| `frontend/src/App.svelte` | Add `Ctrl+O` keyboard handler |

## Verification

1. Press `Ctrl+O` -> file picker opens filtered to `.md` files
2. Select a markdown file -> sidebar loads parent directory, file opens in tab
3. Cancel the dialog -> nothing happens
4. Press `Ctrl+Shift+O` -> TOC panel toggles (no regression)
5. Open command palette (`Ctrl+K`), press `Ctrl+O` -> nothing happens (blocked by palette guard)

# Phase 7: Edit Mode

**Status:** PENDING
**Branch:** `feat/edit-mode`
**Release:** v1.6.0

## Feature

Add an edit mode toggle that allows users to make small text changes directly in the rendered markdown output. In edit mode, triple-clicking a text block (paragraph, heading, list item) makes it contenteditable. Changes auto-save back to the source file by finding and replacing the corresponding text in the raw markdown.

This is NOT a raw markdown editor. Users edit the rendered output — bold text stays bold, links stay clickable. The system maps changes back to the raw markdown source.

## UX Flow

1. User presses `Ctrl+E` (or clicks an edit button in ControlStrip) to toggle edit mode
2. A subtle visual indicator shows edit mode is active (e.g., border glow or label)
3. User triple-clicks a text block -> that block becomes contenteditable
4. User edits the text inline
5. On blur (click away) or Enter, changes are saved back to the file
6. Pressing Escape while editing a block reverts the change and exits contenteditable
7. `Ctrl+E` again exits edit mode, removing all contenteditable attributes

## Constraints

- Only works on file tabs (not stream tabs)
- Only text blocks are editable: paragraphs (`p`), headings (`h1-h6`), list items (`li`), blockquote content
- Code blocks (`pre`, `code`) are NOT editable (too complex to map back to markdown)
- Images, tables, and horizontal rules are NOT editable
- Changes map back by finding the original text in the raw markdown and replacing it
- If the text appears multiple times in the markdown, use context (surrounding text) to disambiguate

## Security Considerations

- **File write validation:** The file path must be validated against the root path (same as `ReadFile`) to prevent path traversal
- **Content sanitization:** The edited text comes from contenteditable (user's own content), but we strip any injected HTML tags before writing back to markdown
- **File size limit:** Same 10MB limit as reading applies to writing

---

## Tasks

### Task 7.1: Add `WriteFile` method to Go backend

**Files:** `app.go`

Add a new Wails-bound method:
```go
func (a *App) WriteFile(relativePath string, content string) error
```

- Resolve absolute path from `rootPath + relativePath`
- **Security: validate path starts with `rootPath + separator`** (same check as `ReadFile`)
- **Security: validate file size** (reject if `len(content) > maxFileSize` — 10MB)
- **Security: validate file has `.md` or `.markdown` extension** (only allow markdown writes)
- Write content to file using `os.WriteFile` with mode `0644`
- The watcher will automatically detect the change and emit `file:changed`

**Acceptance:**
- [ ] Writes content to a markdown file within root path
- [ ] Rejects paths outside root (path traversal prevention)
- [ ] Rejects non-markdown file extensions
- [ ] Rejects content exceeding 10MB
- [ ] File watcher detects the change

---

### Task 7.2: Add `WriteFile` Wails binding

**Files:** `frontend/wailsjs/go/main/App.js`, `frontend/wailsjs/go/main/App.d.ts`

Add the `WriteFile` binding:
- `App.js`: `export function WriteFile(relativePath, content) { return window['go']['main']['App']['WriteFile'](relativePath, content); }`
- `App.d.ts`: `export function WriteFile(relativePath: string, content: string): Promise<void>;`

**Acceptance:**
- [ ] `WriteFile` is callable from frontend
- [ ] Types match Go method signature

---

### Task 7.3: Add edit mode state to UI store

**Files:** `frontend/src/lib/stores/ui.ts`

Add to the UI store:
- `editMode`: writable store (`boolean`, default `false`)
- `toggleEditMode()`: function to toggle the store value
- Export both

**Acceptance:**
- [ ] `editMode` store is reactive
- [ ] `toggleEditMode()` flips the boolean
- [ ] Default value is `false`

---

### Task 7.4: Add `Ctrl+E` keyboard handler

**Files:** `frontend/src/App.svelte`

Add `Ctrl+E` handler in `handleKeydown`:
- Import `editMode, toggleEditMode` from UI store
- On `Ctrl+E`: call `toggleEditMode()`, prevent default
- Only toggle if the active tab is a file tab (not stream)

**Acceptance:**
- [ ] `Ctrl+E` toggles edit mode
- [ ] Does not toggle on stream tabs
- [ ] Does not fire when command palette is open

---

### Task 7.5: Add edit mode indicator to MarkdownViewer

**Files:** `frontend/src/lib/components/MarkdownViewer.svelte`

When `editMode` is true:
- Add a CSS class `.edit-mode` to the `.doc` container
- Show a subtle indicator: a thin `1px` accent border at the top of the `.doc` container and a floating "Editing" label (same style as "Stopped" label for streams)
- When edit mode is deactivated, remove all `contenteditable` attributes from child elements

**Acceptance:**
- [ ] Visual indicator when edit mode is active
- [ ] `.edit-mode` class applied to `.doc`
- [ ] Indicator disappears when edit mode is off
- [ ] All contenteditable attributes cleaned up on exit

---

### Task 7.6: Implement triple-click contenteditable activation

**Files:** `frontend/src/lib/components/MarkdownViewer.svelte`

Add a click handler on `.doc-inner` that detects triple-clicks when edit mode is active:
- On `click` with `detail === 3` (triple-click):
  - Find the closest editable element: `p`, `h1-h6`, `li`, `blockquote > p`
  - Skip if element is inside `pre`, `code`, `table`, or `.stream-error`
  - Store the original `textContent` for revert capability
  - Set `contenteditable="true"` on the element
  - Add `.editing` CSS class for visual feedback (subtle background highlight)
  - Focus the element
- Only one element should be editable at a time (blur previous before activating new)

**Acceptance:**
- [ ] Triple-click on a paragraph makes it contenteditable
- [ ] Triple-click on a heading makes it contenteditable
- [ ] Triple-click on a list item makes it contenteditable
- [ ] Code blocks are not editable
- [ ] Tables are not editable
- [ ] Only one element editable at a time
- [ ] Visual feedback on the editable element

---

### Task 7.7: Implement save-on-blur and keyboard controls

**Files:** `frontend/src/lib/components/MarkdownViewer.svelte`

Add `blur`, `keydown` handlers on the contenteditable element:
- **On blur:** Save the change (call the save function from Task 7.8), remove `contenteditable`, remove `.editing` class
- **On Escape:** Revert to original text (stored in Task 7.6), remove `contenteditable`, remove `.editing` class
- **On Enter (in paragraphs):** Save and exit editing (prevent newline insertion). In headings, same behavior. In list items, allow Enter for new items (or save on Enter — simpler).

For simplicity in v1, Enter always saves and exits editing.

**Acceptance:**
- [ ] Clicking away (blur) saves changes
- [ ] Escape reverts to original text
- [ ] Enter saves and exits editing
- [ ] `contenteditable` and `.editing` class removed after save/revert

---

### Task 7.8: Implement markdown text replacement and file save

**Files:** `frontend/src/lib/components/MarkdownViewer.svelte`

Add a `saveEdit` function:
- Takes `originalText` (from Task 7.6) and `newText` (current textContent)
- If they are identical, skip (no-op)
- Read the current raw markdown content from the active tab's content
- Find `originalText` in the raw markdown (plain text search, ignore markdown formatting)
- If found, replace the first occurrence with `newText`
- Call `WriteFile(activeTab.path, updatedMarkdown)` to persist
- The file watcher will emit `file:changed`, which updates the tab content and triggers re-render

**Matching strategy:**
- Strip inline markdown formatting from the raw source to find the text match
- Match against the text content (what the user sees), not the raw markdown
- Use surrounding context (previous/next lines) to disambiguate duplicate matches
- If no match is found, log a warning and skip (do not corrupt the file)

**Acceptance:**
- [ ] Edited text is saved back to the markdown file
- [ ] Simple text changes (typo fixes, word changes) save correctly
- [ ] No data corruption if match fails
- [ ] File content reloads after save (via watcher)

---

### Task 7.9: Add edit mode button to ControlStrip

**Files:** `frontend/src/lib/components/ControlStrip.svelte`

Add an edit toggle button to the ControlStrip:
- Pencil icon (SVG)
- Visually indicates active/inactive state
- Calls `toggleEditMode()`
- Tooltip: "Edit mode (Ctrl+E)"
- Only shown for file tabs

**Acceptance:**
- [ ] Edit button appears in ControlStrip for file tabs
- [ ] Clicking toggles edit mode
- [ ] Active state visually distinct (accent color)
- [ ] Hidden for stream tabs

---

### Task 7.10: Add edit mode CSS

**Files:** `frontend/src/lib/components/MarkdownViewer.svelte`

Add CSS for edit mode:

```css
.doc.edit-mode {
  border-top: 1px solid var(--accent-dim);
}

.doc-inner :global(.editing) {
  background: var(--accent-dim);
  border-radius: 4px;
  outline: none;
  padding: 2px 4px;
  margin: -2px -4px;
}

.edit-label {
  position: absolute;
  top: 8px;
  right: 12px;
  font-size: 11px;
  color: var(--accent-text);
  opacity: 0.6;
  pointer-events: none;
}
```

**Acceptance:**
- [ ] Edit mode has subtle visual indicator
- [ ] Editing elements have background highlight
- [ ] Styling follows design system (CSS custom properties)
- [ ] No visual regression in non-edit mode

---

## File Summary

| File | Change |
|------|--------|
| `app.go` | Add `WriteFile()` Wails-bound method with path/extension/size validation |
| `frontend/wailsjs/go/main/App.js` | Add `WriteFile` binding |
| `frontend/wailsjs/go/main/App.d.ts` | Add `WriteFile` type |
| `frontend/src/lib/stores/ui.ts` | Add `editMode` store and `toggleEditMode()` |
| `frontend/src/App.svelte` | Add `Ctrl+E` keyboard handler |
| `frontend/src/lib/components/MarkdownViewer.svelte` | Triple-click activation, save-on-blur, markdown replacement, edit mode indicator, CSS |
| `frontend/src/lib/components/ControlStrip.svelte` | Add edit mode toggle button |

## Security Notes

- `WriteFile` validates path traversal (same pattern as `ReadFile`)
- `WriteFile` only allows `.md`/`.markdown` extensions
- `WriteFile` enforces 10MB size limit
- Edited content is the user's own text (no external input) — minimal XSS risk since markdown-it renders with `html: false`
- Contenteditable content is read via `textContent` (not `innerHTML`) to avoid injected HTML

## Verification

1. Press `Ctrl+E` -> edit mode indicator appears
2. Triple-click a paragraph -> it becomes editable with visual highlight
3. Change some text, click away -> file is saved, content re-renders
4. Triple-click a heading, press Escape -> text reverts to original
5. Try to triple-click a code block -> nothing happens (not editable)
6. Press `Ctrl+E` again -> edit mode off, no contenteditable elements remain
7. Edit button in ControlStrip -> toggles edit mode
8. Stream tab -> edit mode toggle hidden/disabled
9. Path traversal attempt -> rejected by backend

# ais — Complete Task Breakdown

All 8 features decomposed into atomic tasks. Each task is self-contained, specifies exact file paths, and targets under 30 minutes of implementation.

**References:** `spec/feature-plans.md`, `spec/feature-ux.md`, `spec/architecture-review.md`

**Architecture decisions incorporated** (from `spec/architecture-review.md`):
- **DR-3:** Highlight storage mirrors directory structure, not SHA-256 filenames
- **DR-4:** Watcher bug root cause is symlinks in rootPath; fix via `filepath.EvalSymlinks`
- **DR-5:** Use field-specific `SaveUISettings` method, not full `UpdateConfig` replacement
- **DR-6:** rootPath data race must be fixed before Features 2 and 3 (add `rootMu sync.RWMutex`)
- **DR-7:** Use `GetInitialFile()` binding, not events, for initial file open (avoids race)
- **S-6:** File server must allowlist MIME types (images only)

---

## Feature Index

| # | Feature | Tasks | Phases |
|---|---------|-------|--------|
| P | Prerequisites (rootPath race, symlinks) | 2 | P1 |
| 1 | External Link Handling | 7 | P2-P3 |
| 2 | Image Support | 9 | P2-P5 |
| 3 | Multi-Color Text Highlighting | 13 | P2-P5 |
| 4 | File Watcher Bug Fix | 1 | P2 |
| 5 | Default Opacity 100% | 2 | P0 |
| 6 | Window Maximized by Default | 1 | P0 |
| 7 | Persistent UI Config | 5 | P2-P3 |
| 8 | CLI Open Specific File | 3 | P2-P3 |
| **Total** | | **43** | |

---

## Phase 0: Quick Fixes (no dependencies, all parallel)

### Task 5.1: Change Default Opacity to 100%

**Feature:** F5
**File(s):** `frontend/src/lib/stores/ui.ts`
**Change:** Line 10 — change `export const opacity = writable(75);` to `export const opacity = writable(100);`
**Acceptance:**
- Default opacity is 100 on fresh launch
- Opacity slider in SettingsPanel shows 100%
- Keyboard shortcuts (Ctrl+Shift++/-) still work

---

### Task 5.2: Update CSS Default Surface Opacity

**Feature:** F5
**File(s):** `frontend/src/style.css`
**Change:** Line 14 in `:root` block — change `--surface-opacity: 0.75;` to `--surface-opacity: 1;`. Must match Task 5.1.
**Acceptance:**
- Reader surface fully opaque by default
- Dynamic opacity via JS still works (`applyOpacity` sets `--surface-opacity`)

---

### Task 6.1: Start Window Maximized by Default

**Feature:** F6
**File(s):** `main.go`
**Change:** Add `WindowStartState: options.Maximised,` to the `options.App` struct after `MinHeight` (line 81 area). Import already covers `options`.
**Acceptance:**
- Window opens maximized on Linux, macOS, Windows
- User can restore via titlebar button
- Width/Height (1200x800) remain as restored-window dimensions

---

## Phase 1: Prerequisites (sequential, before all feature work)

### Task P.1: Fix rootPath Data Race (DR-6)

**Feature:** Cross-cutting (prerequisite for F2, F3)
**File(s):** `app.go`
**Change:**
1. Add `rootMu sync.RWMutex` field to `App` struct
2. Add method:
   ```go
   func (a *App) getRootPath() string {
       a.rootMu.RLock()
       defer a.rootMu.RUnlock()
       return a.rootPath
   }
   ```
3. In `SetRootPath`: wrap `a.rootPath = absPath` with `a.rootMu.Lock()`/`Unlock()`
4. Replace all direct `a.rootPath` reads with `a.getRootPath()` in: `GetFileTree`, `ReadFile`, `GetRootPath` (the public one delegates to `getRootPath`), `StartPipe`

The file server handler (F2) will call `getRootPath()` from a separate HTTP goroutine, making this race real.
**Acceptance:**
- `go test -race ./...` passes
- `getRootPath()` safe to call from any goroutine
- `SetRootPath` holds write lock during mutation
- All existing functionality unaffected

---

### Task P.2: Apply EvalSymlinks to rootPath (DR-4)

**Feature:** F4 (prerequisite)
**Depends on:** Task P.1
**File(s):** `main.go`, `app.go`
**Change:**
1. `main.go` after line 58 (`filepath.Abs`): add `absPath, err = filepath.EvalSymlinks(absPath)` with error handling
2. `app.go` `SetRootPath` after `filepath.Abs` (line 121): add `absPath, err = filepath.EvalSymlinks(absPath)` with error handling

Fixes watcher bug: fsnotify reports resolved symlink paths while `filepath.Rel` uses unresolved root, producing incorrect relative paths.
**Acceptance:**
- Symlinked directories resolve to real path
- Watcher events produce paths matching scanner output
- Non-symlinked paths unaffected (EvalSymlinks is no-op for real paths)

---

## Phase 2: Backend Foundation (parallel within phase, after Phase 1)

### Task 1.1: Add Go Method for External URL Opening

**Feature:** F1
**File(s):** `app.go`
**Change:** Add `OpenExternalURL(url string) error`. Parse with `net/url.Parse()`. Whitelist `http` and `https` schemes only. Call `wailsRuntime.BrowserOpenURL(a.ctx, url)`. Add `neturl "net/url"` import.
**Acceptance:**
- http/https URLs open in default browser
- javascript:, data:, file:, ftp: rejected
- Malformed/empty URLs rejected
- Wails auto-generates binding

---

### Task 1.7: Unit Tests for URL Validation

**Feature:** F1
**File(s):** New file `app_test.go`
**Change:** Table-driven tests for `OpenExternalURL`: valid http/https, invalid schemes (javascript, data, file, ftp), empty string, malformed URLs, relative paths.
**Acceptance:**
- All edge cases covered
- Tests pass with `go test ./...`

---

### Task 2.1: Create Local File Server Handler

**Feature:** F2
**Depends on:** Task P.1
**File(s):** New file `internal/fileserver/handler.go`
**Change:** Create `http.Handler` per DR-1. Constructor: `NewHandler(getRootPath func() string) *Handler`.

Handler logic:
1. Strip `/local/` prefix, URL-decode path
2. Resolve via `filepath.Join(root, rel)` then `filepath.EvalSymlinks`
3. Validate resolved path prefix (same as `ReadFile`)
4. Check file exists, not directory, under 10MB
5. **MIME type allowlist (S-6):** Only serve image types: `.png`, `.jpg`, `.jpeg`, `.gif`, `.svg`, `.webp`, `.bmp`, `.ico`. Reject all other extensions with 403.
6. Set `Content-Type`, `X-Content-Type-Options: nosniff` on all responses
7. For SVG: add `Content-Security-Policy: script-src 'none'; style-src 'unsafe-inline'` (DR-2)
8. Return 404/403/413 for errors
**Acceptance:**
- Valid image paths: correct content + MIME
- Non-image files (.env, .json, .go): 403
- Path traversal (../): 403
- Missing files: 404
- Over 10MB: 413
- SVG responses have CSP header

---

### Task 2.5: File Server Security Tests

**Feature:** F2
**Depends on:** Task 2.1
**File(s):** New file `internal/fileserver/handler_test.go`
**Change:** Tests using `httptest` and `t.TempDir()`:
- Path traversal via `../` rejected (403)
- Path traversal via `%2e%2e%2f` rejected (403)
- Valid PNG: 200 + correct MIME
- Valid SVG: 200 + CSP header
- Non-image file (.json): 403
- Missing file: 404
- Directory: 404
- Oversized file: 413
- nosniff header present
**Acceptance:**
- All security vectors tested
- `go test ./internal/fileserver/...` passes

---

### Task 2.6: Integrate File Server into Wails AssetServer

**Feature:** F2
**Depends on:** Task 2.1, Task P.1
**File(s):** `main.go`
**Change:** Import `internal/fileserver`. Create handler: `handler := fileserver.NewHandler(app.getRootPath)`. Wire into `AssetServer.Handler`:
```go
AssetServer: &assetserver.Options{
    Assets:  assets,
    Handler: handler,
},
```
Note: `app.getRootPath` (lowercase, unexported) is only available within the `main` package. If `app` is in package `main`, this works. Otherwise, use `app.GetRootPath`.
**Acceptance:**
- `/local/path/to/image.png` serves workspace files
- After `SetRootPath`, images resolve from new root
- Frontend assets unaffected

---

### Task 3.1: Create Highlight Storage Package

**Feature:** F3
**Depends on:** Task P.1
**File(s):** New file `internal/highlights/highlights.go`
**Change:** Per DR-3 — mirror directory structure (no SHA-256):
- `Highlight` struct: ID, FilePath, AnchorText, PrefixContext, SuffixContext, Color, CreatedAt
- `Store` struct: `baseDir string`, `mu sync.RWMutex`
- `NewStore(baseDir string) *Store`
- `SetRoot(baseDir string)` — for workspace switches
- `Load(filePath string) ([]Highlight, error)` — loads `{baseDir}/{filePath}.json` (e.g., `docs/arch.md` -> `.ais/highlights/docs/arch.md.json`). Returns `[]Highlight{}` if missing.
- `Save(filePath string, highlights []Highlight) error` — creates subdirs, writes JSON
- `Add(h Highlight) error` — load, append, save
- `Remove(filePath, highlightID string) error` — load, filter, save
- `Clear(filePath string) error` — delete file

Storage: `{rootPath}/.ais/highlights/`. Watcher already skips `.ais` (hidden dir rule at watcher.go line 66). Scanner already skips `.json` files.
**Acceptance:**
- Storage at `{rootPath}/.ais/highlights/`
- File structure mirrors document paths
- Empty load returns `[]Highlight{}`, not error
- Subdirectories created automatically
- File permissions 0644

---

### Task 3.8: Highlight Storage Tests

**Feature:** F3
**Depends on:** Task 3.1
**File(s):** New file `internal/highlights/highlights_test.go`
**Change:** Tests: empty load, save/load round-trip, add appends, remove by ID, remove non-existent returns error, clear deletes, nested paths create subdirs, SetRoot updates base.
**Acceptance:**
- `go test ./internal/highlights/...` passes
- Uses `t.TempDir()`

---

### Task 3.2: Add Wails Bindings for Highlight CRUD

**Feature:** F3
**Depends on:** Task 3.1, Task P.1
**File(s):** `app.go`
**Change:**
1. Add `highlightStore *highlights.Store` to App struct
2. Initialize in `startup()`: `a.highlightStore = highlights.NewStore(filepath.Join(a.getRootPath(), ".ais", "highlights"))`
3. Update `SetRootPath` to call `a.highlightStore.SetRoot(...)`
4. Add four bound methods: `GetHighlights`, `AddHighlight`, `RemoveHighlight`, `ClearHighlights`
5. Path validation: same prefix check as `ReadFile`
**Acceptance:**
- Four methods auto-bound by Wails
- Path validation prevents traversal
- Store re-initialized on workspace switch

---

### Task 7.1: Add UI Settings Fields to Go Config Struct

**Feature:** F7
**File(s):** `internal/config/config.go`, `internal/config/defaults.go`
**Change:** Add five fields to `Config` struct:
```go
ZoomLevel      int    `json:"zoomLevel"`
Opacity        int    `json:"opacity"`
ReadingWidth   int    `json:"readingWidth"`
ReaderRadius   int    `json:"readerRadius"`
BackgroundMode string `json:"backgroundMode"`
```
In `defaults.go`: `ZoomLevel: 100`, `Opacity: 100`, `ReadingWidth: 1000`, `ReaderRadius: 20`, `BackgroundMode: "gradient"`.

In `Load()`: zero-value guards (same pattern as `VertexRegion`).
**Acceptance:**
- Fields serialize/deserialize in config.json
- Defaults match frontend defaults
- Old configs load without error

---

### Task 7.1b: Add SaveUISettings Method (DR-5)

**Feature:** F7
**Depends on:** Task 7.1
**File(s):** `app.go`
**Change:** Per DR-5, add field-specific save method (NOT full `UpdateConfig` replacement):
```go
type UISettings struct {
    ZoomLevel      int    `json:"zoomLevel"`
    Opacity        int    `json:"opacity"`
    ReadingWidth   int    `json:"readingWidth"`
    ReaderRadius   int    `json:"readerRadius"`
    BackgroundMode string `json:"backgroundMode"`
}

func (a *App) SaveUISettings(s UISettings) error {
    a.cfgMgr.Update(func(c *config.Config) {
        c.ZoomLevel = s.ZoomLevel
        c.Opacity = s.Opacity
        c.ReadingWidth = s.ReadingWidth
        c.ReaderRadius = s.ReaderRadius
        c.BackgroundMode = s.BackgroundMode
    })
    return a.cfgMgr.Save()
}
```
**Acceptance:**
- Only UI fields modified
- Other config fields preserved
- Wails auto-generates binding
- Concurrent AI settings saves not clobbered

---

### Task 8.1: Modify CLI to Accept File Paths

**Feature:** F8
**File(s):** `main.go`
**Change:** Lines 64-72: if `!info.IsDir()`:
1. Check markdown extension (`.md` or `.markdown`)
2. Set `rootPath` to `filepath.Dir(absPath)`
3. Store `initialFile = filepath.Base(absPath)`
4. Re-resolve `absPath` to directory for downstream validation
5. Non-markdown files: print error, exit

Declare `initialFile := ""` before the arg loop.
**Acceptance:**
- `ais README.md` opens with parent dir as root
- `ais /abs/path/file.md` works
- `ais somedir/` unchanged
- Non-markdown files rejected
- Non-existent paths error

---

### Task 8.2: Add Initial File to App Struct (DR-7)

**Feature:** F8
**Depends on:** Task 8.1
**File(s):** `app.go`, `main.go`
**Change:** Per DR-7, use binding (not events) to avoid race:
1. Add `initialFile string` to App struct
2. Update `NewApp(rootPath, initialFile string)`
3. Add bound method: `func (a *App) GetInitialFile() string { return a.initialFile }`
4. Update `main.go`: `app := NewApp(absPath, initialFile)`
**Acceptance:**
- `NewApp` accepts two parameters
- `GetInitialFile()` returns filename or empty string
- Wails auto-generates binding
- No race condition

---

### Task 4.1: Add Watcher Path Alignment Test

**Feature:** F4
**File(s):** `internal/watcher/watcher_test.go`
**Change:** Test that:
1. Creates temp dir with nested markdown files
2. Starts Watcher, writes to file
3. Asserts callback path matches `scanner.ScanDirectory` output
4. Tests with EvalSymlinks applied (after P.2)
**Acceptance:**
- Watcher paths match scanner paths
- Root-level and nested files covered
- `go test ./internal/watcher/...` passes

---

## Phase 3: Frontend Foundation (depends on Phase 2)

### Task 1.2: Intercept All Link Clicks in MarkdownViewer

**Feature:** F1
**File(s):** `frontend/src/lib/components/MarkdownViewer.svelte`
**Change:** In `handleContentClick`, after code-copy-btn check (line 125), before heading check (line 132):
```typescript
const link = target.closest('a') as HTMLAnchorElement | null;
if (link) {
    e.preventDefault();
    const href = link.getAttribute('href');
    if (!href) return;
    if (href.startsWith('http://') || href.startsWith('https://')) {
        handleExternalLink(href);
    } else if (href.startsWith('#')) {
        handleAnchorLink(href);
    } else if (/\.(?:md|markdown)(?:#|$)/i.test(href)) {
        handleLocalLink(href);
    }
    return;
}
```
Define stub handler functions.
**Acceptance:**
- All `<a>` clicks intercepted
- WebView never navigates away
- Child elements inside `<a>` detected via `closest()`
- Non-link clicks (headings, copy) unaffected

---

### Task 1.3: Implement External Link Handler

**Feature:** F1
**Depends on:** Task 1.1, Task 1.2
**File(s):** `frontend/src/lib/components/MarkdownViewer.svelte`
**Change:** Implement `handleExternalLink(href)`: import `OpenExternalURL` from Wails binding, call it, catch errors to console.
**Acceptance:**
- External links open in default browser
- Errors logged, not shown to user

---

### Task 1.4: Implement Local Markdown Link Handler

**Feature:** F1
**Depends on:** Task 1.2
**File(s):** `frontend/src/lib/components/MarkdownViewer.svelte`
**Change:** Implement `handleLocalLink(href)`:
1. Strip anchor fragment (split on `#`)
2. Get current dir from `$activeTab.path`
3. Join and normalize (resolve `../`)
4. Call `openTab(resolvedPath, filename)`
**Acceptance:**
- `[link](other.md)` opens tab
- `../README.md` from `docs/arch.md` resolves to `README.md`
- Already-open files focused

---

### Task 1.5: Implement Anchor Link Handler + Heading IDs

**Feature:** F1
**Depends on:** Task 1.2
**File(s):** `frontend/src/lib/markdown/renderer.ts`, `frontend/src/lib/components/MarkdownViewer.svelte`
**Change:**
1. `renderer.ts`: add markdown-it core ruler `heading_ids` that sets `id` from heading text (slugify: lowercase, spaces to hyphens, strip special chars)
2. `MarkdownViewer.svelte`: implement `handleAnchorLink(href)` using `viewerEl.querySelector` + `scrollIntoView({ behavior: 'smooth' })`
**Acceptance:**
- Headings have `id` attributes
- `#section` links smooth-scroll
- Missing anchors do nothing
- `html: false` invariant preserved

---

### Task 1.6: Add Visual Indicator for External Links

**Feature:** F1
**File(s):** `frontend/src/lib/markdown/renderer.ts`, `frontend/src/lib/components/MarkdownViewer.svelte`
**Change:**
1. `renderer.ts`: add `link_open` rule adding `data-external="true"` for http/https links
2. MarkdownViewer CSS: `a[data-external]::after { content: '\2197'; font-size: 0.7em; opacity: 0.35; color: var(--text-tertiary); }`
**Acceptance:**
- Subtle arrow on external links
- Uses --text-tertiary token
- No indicator on local/anchor links

---

### Task 2.2: Add Image URL Rewriting to Renderer

**Feature:** F2
**Depends on:** Task 2.6
**File(s):** `frontend/src/lib/markdown/renderer.ts`, `frontend/src/lib/components/MarkdownViewer.svelte`
**Change:**
1. `renderer.ts`: change to `renderMarkdown(source, basePath?)`. Add `image` rule: http/https/data: pass through; relative resolved against basePath, rewritten to `/local/`; absolute rewritten to `/local/`.
2. MarkdownViewer: derive basePath from `$activeTab.path`, pass to renderMarkdown. Stream tabs pass undefined.
**Acceptance:**
- Local images render via `/local/`
- External/data URIs unchanged
- Relative paths resolved correctly
- Stream tabs work without basePath

---

### Task 2.3: Add Image Error Fallback Placeholder

**Feature:** F2
**File(s):** `frontend/src/lib/components/MarkdownViewer.svelte`
**Change:** Add `$effect()` after render. Attach `onerror` to `<img>` elements. Replace broken images with styled placeholder div (broken-image SVG + alt text). CSS uses `--hover-bg`, `--border`, `--text-tertiary`.
**Acceptance:**
- Broken images show styled placeholder
- Alt text displayed if available
- Re-attached on re-render
- Successfully loaded images unaffected

---

### Task 2.4: Add Image Design Tokens and Styling

**Feature:** F2
**File(s):** `frontend/src/style.css`, `frontend/src/lib/components/MarkdownViewer.svelte`
**Change:**
1. `style.css`: add image tokens from `spec/feature-ux.md` (both `:root` and `.light`)
2. MarkdownViewer: update `.doc-inner :global(img)` from `border-radius: 6px` to 12px, add border, shadow, hover state, `cursor: zoom-in`
3. Add `.md-image-error` and `.md-image-placeholder` classes
**Acceptance:**
- 12px radius, border, shadow
- Hover brightens border
- Both themes work
- No hardcoded colors

---

### Task 2.7: Create Image Lightbox Component

**Feature:** F2
**File(s):** New file `frontend/src/lib/components/ImageLightbox.svelte`
**Change:** Svelte 5 component. Props: `src`, `alt`, `open`, `onclose`. Per `spec/feature-ux.md`: overlay z-index 300, image max 90vw/85vh, 16px radius, 120ms animation, close on overlay/image/Escape/scroll. ARIA dialog.
**Acceptance:**
- Centered image on overlay
- All four close triggers
- Escape priority: lightbox > command palette > stream > nav
- Reduced motion: instant
- Keyboard accessible

---

### Task 2.8: Wire Lightbox into MarkdownViewer

**Feature:** F2
**Depends on:** Task 2.7
**File(s):** `frontend/src/lib/components/MarkdownViewer.svelte`
**Change:** Add lightbox state. In `handleContentClick`, detect `<img>` clicks (before link handler). Exclude error placeholders. Import and render `<ImageLightbox>`. Add Escape priority.
**Acceptance:**
- Image clicks open lightbox
- Broken images do not trigger
- Escape closes lightbox

---

### Task 3.3: Define Highlight Color CSS Custom Properties

**Feature:** F3
**File(s):** `frontend/src/style.css`
**Change:** Add from `spec/feature-ux.md`: 6 colors x 3 variants (background, border, dot) for both themes. Add `<mark>` base styling with `data-highlight` attribute.
**Acceptance:**
- 6 colors for both themes
- Subtle opacity (0.14-0.28)
- WCAG AA readable

---

### Task 3.4: Create Text Selection Capture Utility

**Feature:** F3
**File(s):** New file `frontend/src/lib/highlights/selection.ts`
**Change:** Export `captureSelection(viewerEl): SelectionAnchor | null`. Uses `window.getSelection()`. Returns `{ anchorText, prefixContext(30), suffixContext(30) }`. Returns null if: no selection, outside viewer, whitespace-only, inside `<pre>`/`<code>`.
**Acceptance:**
- Cross-element selections handled
- Code blocks return null
- Pure function

---

### Task 3.5: Create Highlight DOM Renderer

**Feature:** F3
**Depends on:** Task 3.3
**File(s):** New file `frontend/src/lib/highlights/renderer.ts`
**Change:** Export `applyHighlights(viewerEl, highlights[])` and `clearHighlightMarks(viewerEl)`. Removes existing marks, collects text nodes, finds anchors by context, wraps in `<mark>`. Skips code blocks. Idempotent.
**Acceptance:**
- Correct color on marks
- Orphaned skipped
- Idempotent
- Existing handlers work

---

### Task 3.7: Create Highlight Svelte Store

**Feature:** F3
**Depends on:** Task 3.2
**File(s):** New file `frontend/src/lib/stores/highlights.ts`
**Change:** Cache-backed store with `loadHighlightsForFile`, `addHighlight`, `removeHighlight`, `getHighlightsForFile`. Backend calls on cache miss. Invalidate on mutations.
**Acceptance:**
- Cache avoids redundant Go calls
- Add/remove update backend and cache

---

### Task 7.2: Load UI Settings from Go Config on Startup

**Feature:** F7
**Depends on:** Task 7.1
**File(s):** `frontend/src/lib/stores/settings.ts`
**Change:** Extend `loadSettings()`: after loading theme, call `GetConfig()` and apply UI fields (opacity, readingWidth, readerRadius, backgroundMode, zoomLevel) to ui.ts stores.
**Acceptance:**
- UI settings restored on startup
- Old configs use defaults
- Theme loading unaffected

---

### Task 7.3: Save UI Settings on Change (Debounced, DR-5)

**Feature:** F7
**Depends on:** Task 7.1b
**File(s):** `frontend/src/lib/stores/ui.ts`
**Change:** Add debounced `persistUISettings()` that calls `SaveUISettings` binding (NOT `UpdateConfig`). 500ms debounce. Call from: `changeOpacity`, `setOpacity`, `setReaderRadius`, `setBackgroundMode`, `zoomIn`, `zoomOut`, `resetZoom`. Subscribe to `readingWidth` changes for slider persistence.
**Acceptance:**
- All UI changes persist
- 500ms debounce
- Uses `SaveUISettings` (field-specific, per DR-5)
- AI settings not clobbered
- Survives restart

---

### Task 8.3: Handle Initial File Open in Frontend (DR-7)

**Feature:** F8
**Depends on:** Task 8.2
**File(s):** `frontend/src/App.svelte`
**Change:** Per DR-7, use binding call (no events). In onMount after `loadFileTree()`:
```typescript
const { GetInitialFile } = await import('../wailsjs/go/main/App');
const initialFile = await GetInitialFile();
if (initialFile) {
    openTab(initialFile, initialFile.split('/').pop() ?? initialFile);
}
```
**Acceptance:**
- `ais README.md` opens with file in tab
- `ais somedir/` opens with no auto-tab
- No race condition

---

## Phase 4: Integration (depends on Phase 3)

### Task 3.6a: Wire Highlight Loading into MarkdownViewer

**Feature:** F3
**Depends on:** Task 3.5, Task 3.7
**File(s):** `frontend/src/lib/components/MarkdownViewer.svelte`
**Change:** Import store + renderer. Add two effects: (1) load highlights on tab change for file tabs, (2) apply highlights after content render via `requestAnimationFrame`.
**Acceptance:**
- Highlights load on tab switch
- Render after content updates
- Stream tabs show none
- Live reload re-applies

---

### Task 3.6b: Wire Highlight Creation via Keyboard Shortcut

**Feature:** F3
**Depends on:** Task 3.4, Task 3.6a
**File(s):** `frontend/src/App.svelte`, `frontend/src/lib/components/MarkdownViewer.svelte`
**Change:** Ctrl+Shift+H in App.svelte dispatches `highlight:toggle` event. MarkdownViewer listens: capture selection, toggle (remove if exists, add yellow if not), persist, re-apply.
**Acceptance:**
- Ctrl+Shift+H creates/removes highlights
- No selection or code block: no-op
- Persists across restarts

---

### Task 3.6c: Create Quick Action Bar Component

**Feature:** F3
**File(s):** New file `frontend/src/lib/components/QuickActionBar.svelte`
**Change:** Floating toolbar on text selection. Props: position, visible, inCodeBlock, isStreaming, onhighlight. 6 color dots per feature-ux.md. Single click: last-used color. Long press/second click: expand all. Glass surface. Disabled in code blocks and during streaming.
**Acceptance:**
- Appears above selection
- Disappears on deselection
- Disabled states work
- ARIA labels

---

### Task 3.6d: Wire Quick Action Bar into MarkdownViewer

**Feature:** F3
**Depends on:** Task 3.6c, Task 3.6a
**File(s):** `frontend/src/lib/components/MarkdownViewer.svelte`
**Change:** Listen `selectionchange`. Position via `Range.getBoundingClientRect()`. Detect code block context. On highlight: capture, add, re-apply.
**Acceptance:**
- Full workflow: select, click color, mark appears
- No interference with existing handlers

---

### Task 3.6e: Add Right-Click Context Menu for Highlights

**Feature:** F3
**File(s):** `frontend/src/lib/components/MarkdownViewer.svelte`
**Change:** `oncontextmenu` handler. With selection: "Highlight" + 6 dots. On existing mark: "Remove highlight". Prevent default.
**Acceptance:**
- Context menu with color options
- Remove option on highlighted text
- Closes on outside click or Escape

---

### Task 3.10: Add Highlight Search to Command Palette

**Feature:** F3
**File(s):** `frontend/src/lib/components/CommandPalette.svelte`
**Change:** When query contains "highlight", add current file's highlights as results. Each: color dot + first 60 chars. Action: scroll to highlight.
**Acceptance:**
- "highlight" query shows highlights
- Click scrolls to highlight

---

## Phase 5: Polish and Accessibility (depends on Phase 4)

### Task 3.9: Add Highlight Accessibility

**Feature:** F3
**File(s):** `frontend/src/lib/highlights/renderer.ts`, `frontend/src/style.css`
**Change:**
1. Renderer: add `<span class="sr-only">{Color} highlight: </span>` inside marks
2. CSS: `mark[data-highlight]:focus-visible` with `var(--border-focus)` outline
3. Announcements via `#srAnnounce`
**Acceptance:**
- Screen readers announce color
- Marks keyboard-navigable
- Focus indicator visible

---

### Task 2.9: Add Image Accessibility

**Feature:** F2
**File(s):** `frontend/src/lib/components/MarkdownViewer.svelte`, `frontend/src/lib/components/ImageLightbox.svelte`
**Change:** Post-render: `tabindex="0"` on images, Enter opens lightbox. Lightbox ARIA. Announcements. Error placeholder ARIA. Focus return.
**Acceptance:**
- Images focusable via Tab
- Enter opens lightbox
- Lightbox announced
- Focus management correct

---

## Dependency Graph

```
Phase 0 (parallel):
    5.1, 5.2, 6.1

Phase 1 (sequential prerequisites):
    P.1 (rootPath mutex) -> P.2 (EvalSymlinks)

Phase 2 (parallel, after Phase 1):
    1.1, 1.7                     (F1 backend)
    2.1 -> 2.5, 2.6             (F2 backend)
    3.1 -> 3.8, 3.2             (F3 backend)
    7.1 -> 7.1b                  (F7 backend)
    8.1 -> 8.2                   (F8 backend)
    4.1                          (F4 test)

Phase 3 (parallel, after Phase 2):
    1.2 -> 1.3, 1.4, 1.5        (1.3-1.5 parallel)
    1.6                          (independent)
    2.2 -> 2.3                   (depends on 2.6)
    2.4                          (independent CSS)
    2.7 -> 2.8                   (lightbox)
    3.3, 3.4                     (parallel)
    3.5                          (depends on 3.3)
    3.7                          (depends on 3.2)
    7.2 -> 7.3                   (depends on 7.1b)
    8.3                          (depends on 8.2)

Phase 4 (after Phase 3):
    3.6a -> 3.6b                 (highlight wiring)
    3.6c -> 3.6d                 (Quick Action bar)
    3.6e                         (context menu)
    3.10                         (Command Palette)

Phase 5 (after Phase 4):
    3.9, 2.9                     (accessibility, parallel)
```

## Code Invariants Checklist

Every task must preserve these:

- **Path traversal:** All file paths validated against rootPath prefix
- **XSS prevention:** markdown-it `html: false` never changed; renderer uses rule system
- **File size limit:** 10MB max for all file reads/serves
- **Concurrent config:** Config access through Manager.Get()/Update() only
- **rootPath safety:** Access via `getRootPath()` with mutex (after P.1)
- **API key isolation:** Keys never in config.json, never logged, never in error messages
- **Skip directories:** .git, node_modules, vendor, .svn, __pycache__, .venv skipped; .ais skipped by hidden-dir rule
- **CSS tokens only:** No hardcoded colors; use `var(--token)`
- **Svelte 5 runes:** `$state()`, `$derived()`, `$effect()`, `$props()` only
- **ARIA:** All interactive elements have appropriate attributes

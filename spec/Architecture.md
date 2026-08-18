# Architecture — ais

## System Overview

ais is a desktop application that visualizes markdown files.

The application is a single binary built with Wails v2. The backend is written in Go. The frontend is written in Svelte 5 with Tailwind CSS 4.

The architecture follows a strict separation:

- Go handles filesystem operations, configuration, and platform integration
- Svelte handles rendering, state management, and user interaction
- Wails provides the bridge between the two layers via auto-generated RPC bindings

The application ships as a single executable with the frontend assets embedded at compile time.

---

## Backend Architecture

### Package Structure

```
main.go                     Entry point: CLI argument parsing, Wails app bootstrap
app.go                      App struct: all Wails-bound methods, lifecycle hooks
internal/
  config/
    config.go               Config struct, Manager with sync.RWMutex, JSON persistence
    defaults.go             DefaultConfig factory
  scanner/
    scanner.go              Recursive directory scanning, file content reading
    scanner_test.go         Table-driven tests
  types/
    types.go                Shared domain types (FileNode)
  watcher/
    watcher.go              fsnotify-based file watcher with debounce
```

### Entry Point — main.go

Responsibilities:

- Parse CLI arguments (`--version`, `--mcp`, positional path)
- Resolve and validate the root directory path
- Configure Wails application options (window size, background color, asset server, frameless mode)
- Embed the compiled frontend via `//go:embed all:frontend/dist`
- Start the Wails runtime

Window configuration:

- `Frameless: true` — removes OS window decorations; the frontend renders custom window controls
- `BackgroundColour: RGBA{11, 13, 16, 1}` — matches `--bg: #0B0D10`
- Build tags: `webkit2_41` for Fedora 43+ compatibility (webkit2gtk-4.1)

The `version` variable is set at build time via `-ldflags`.

### Application Core — app.go

The `App` struct holds all runtime state:

```go
type App struct {
    ctx      context.Context    // Wails runtime context
    rootPath string             // Current workspace root
    cfgMgr   *config.Manager   // Thread-safe configuration
    watcher  *watcher.Watcher  // Filesystem change watcher
}
```

Lifecycle hooks:

```
startup(ctx)    Initialize config manager, load config, start watcher
shutdown(ctx)   Stop watcher, save config
```

### Config Package — internal/config

The `Config` struct represents all persistent user preferences:

```go
type Config struct {
    Theme          string   // "system", "light", "dark"
    SSHKeyPaths    []string // Reserved for future use
    IgnoreDirs     []string // Directories to skip during scanning
    LastOpenedPath string   // Last opened workspace path
    FontSize       int      // Editor font size in pixels
    SidebarWidth   int      // Sidebar width in pixels
    RecentPaths    []string // Recently opened workspace paths (max 10)
}
```

Storage location:

```
~/.config/ais/config.json
```

The `Manager` wraps the config with a `sync.RWMutex` to guarantee thread-safe access. All reads go through `Get()`, all mutations go through `Update(fn)`, both of which acquire the appropriate lock.

Default values:

| Field | Default |
|-------|---------|
| Theme | `"system"` |
| FontSize | `16` |
| SidebarWidth | `260` |
| IgnoreDirs | `.git`, `node_modules`, `.svn`, `__pycache__`, `vendor`, `.venv` |

If the config file does not exist on first load, defaults are used silently. No error is raised.

### Scanner Package — internal/scanner

Two exported functions:

```
ScanDirectory(root string) → (*FileNode, error)
ReadFileContent(absPath string) → (string, error)
```

`ScanDirectory` walks the filesystem recursively and builds a tree of `FileNode` values. It applies the following rules:

- Directories in the `skipDirs` set are excluded entirely
- Only files with `.md` or `.markdown` extensions are included
- Empty directories (containing no markdown files at any depth) are pruned
- Results are sorted alphabetically, directories before files, case-insensitive

`ReadFileContent` reads a file into memory with two guards:

- The path must resolve to a file, not a directory
- The file must not exceed `maxFileSize` (10 MB)

### Types Package — internal/types

A single shared type:

```go
type FileNode struct {
    Name     string      `json:"name"`
    Path     string      `json:"path"`
    IsDir    bool        `json:"isDir"`
    Children []*FileNode `json:"children,omitempty"`
}
```

`Path` is always relative to the workspace root. The root node itself has `Path: ""`.

### Watcher Package — internal/watcher

Built on `fsnotify`. Watches the workspace root and all subdirectories recursively.

Behavior:

```
Filesystem event (Write/Create/Remove/Rename)
    ↓
Is it a markdown file?
    ├─ Yes → Debounce 100ms → Invoke callback with relative path
    └─ No  → Is it a new directory?
              ├─ Yes → Add to watch list recursively
              └─ No  → Ignore
```

Directories in the `skipDirs` set and hidden directories (prefixed with `.`) are never watched.

The debounce prevents rapid successive events (common with editors that write atomically via temp files) from triggering multiple callbacks.

Shutdown is idempotent via `sync.Once`.

---

## Frontend Architecture

### Component Hierarchy

```
App.svelte                          Shell: background, floating reader, keyboard handlers
  ├─ Sidebar.svelte                 Left nav panel (hover-reveal), file explorer, search
  │     ├─ FileTree.svelte          Tree container
  │     │     └─ FileTreeNode.svelte   Recursive node (expand/collapse, filtering)
  │     └─ ThemeToggle.svelte       Theme cycle button (system → light → dark)
  ├─ TabBar.svelte                  Floating tab strip (hidden when single tab)
  ├─ MarkdownViewer.svelte          Renders active tab content, heading collapse, scroll restore
  │     └─ WelcomeScreen.svelte     Shown when no tabs are open
  ├─ TocPanel.svelte                Right nav panel (hover-reveal), document outline
  ├─ ControlStrip.svelte            Bottom floating control bar (zoom, width, theme, settings)
  ├─ SettingsPanel.svelte           Appearance settings (theme, opacity, radius, background)
  └─ CommandPalette.svelte          Ctrl+K command/document search overlay
```

### Layout

The shell uses a floating reader surface centered on the viewport. No CSS Grid.

```
Background layer (fixed, full viewport)
  ├─ Sky gradient
  ├─ Mountains
  ├─ Fog
  └─ Stars

Reader surface (fixed, centered, 82vw × 90vh)
  ├─ Titlebar (logo, breadcrumb, window controls)
  ├─ TabBar (hidden when ≤1 tab)
  └─ Content area (flex: 1, relative)
        ├─ Nav trigger zone (left 24px, hover-reveal)
        ├─ Nav panel (Sidebar — absolute positioned)
        ├─ Document viewer (flex: 1, scrollable)
        ├─ TOC trigger zone (right 24px, hover-reveal)
        ├─ TOC panel (absolute positioned)
        ├─ Edge indicators (left/right, fade on hover)
        ├─ Bottom trigger zone (24px, hover-reveal)
        ├─ Control strip (absolute, bottom center)
        └─ Settings panel (absolute, above control strip)
```

The reader surface is a frameless window — no OS window decorations. The `border-radius: var(--reader-radius)` defines the visual window boundary. Custom window controls (minimize, maximize, close) are rendered in the titlebar.

Window dragging is enabled via `--wails-draggable: drag` on the titlebar element.

### Svelte 5 Runes

All components use Svelte 5 runes mode:

- `$state()` for local reactive state
- `$derived()` for computed values
- `$effect()` for side effects
- `$props()` for component inputs

Legacy `let` reactivity is not used.

### Stores

Four store modules manage global state:

**files.ts** — File tree and content access

```
fileTree: Writable<FileNode | null>     Current workspace tree
rootPath: Writable<string>              Current workspace root path
loadFileTree(): Promise<void>           Fetch tree from Go backend
readFile(path): Promise<string>         Fetch file content from Go backend
```

**tabs.ts** — Tab management

```
tabs: Writable<Tab[]>                   Open tabs (max 20)
activeTabId: Writable<string | null>    Currently active tab ID
activeTab: Derived<Tab | null>          Currently active tab object
openTab(path, name): Promise<void>      Open or focus a tab
closeTab(id): void                      Close a tab, activate neighbor
nextTab() / prevTab(): void             Cycle through tabs
updateTabContent(path, content): void   Update tab content (live reload)
saveScrollPos(id, pos): void            Persist scroll position per tab
```

Tab IDs are the file path. Maximum 20 open tabs; the oldest is evicted when the limit is reached.

**settings.ts** — Theme management

```
theme: Writable<ThemeMode>              "light" | "dark" | "system"
loadSettings(): Promise<void>           Load theme from Go backend
setTheme(mode): Promise<void>           Apply and persist theme
applyTheme(mode): void                  Toggle .light class on <html> (dark is default)
```

System theme preference changes are tracked via `matchMedia` listener.

**ui.ts** — UI state and controls

```
zoomLevel: Writable<number>             Current zoom (50–200%, step via ZOOM_LEVELS)
readingWidth: Writable<number>          Document max-width in pixels (500–1000)
focusMode: Writable<boolean>            Focus mode active (hides titlebar, tabs, shows only content)
commandPaletteOpen: Writable<boolean>   Command palette visibility
opacity: Writable<number>              Reader surface opacity (40–100%)
readerRadius: Writable<number>         Window corner radius (20, 28, 36, 48 px)
backgroundMode: Writable<string>       Background style ("gradient", "solid", "frost")

zoomIn() / zoomOut() / resetZoom()     Step through ZOOM_LEVELS array
toggleFocusMode()                       Toggle focus mode
toggleCommandPalette()                  Toggle command palette
changeOpacity(delta)                    Adjust opacity by delta, clamp 40–100
setReaderRadius(px)                     Set --reader-radius CSS variable
setBackgroundMode(mode)                 Toggle background layer visibility
```

ZOOM_LEVELS: `[50, 75, 90, 100, 110, 125, 150, 175, 200]`

### Markdown Rendering

The `renderer.ts` module configures markdown-it with:

```
html: false          Raw HTML is disabled (XSS prevention)
linkify: true        URLs are auto-linked
typographer: true    Smart quotes and dashes
```

Syntax highlighting uses highlight.js with 15 registered languages:

```
Go, TypeScript, JavaScript, Python, Bash/Shell, JSON, YAML/YML,
CSS, HTML/XML, SQL, Markdown, Dockerfile, Rust, Java, Makefile
```

The highlight function:

- Detects language from the fenced code block info string
- Applies syntax highlighting via `hljs.highlight()`
- Escapes all output via `md.utils.escapeHtml()`
- Wraps result in `<pre class="code-block">` with a language label

If language detection fails, content renders as escaped plain text.

### Styling

All colors are defined as CSS custom properties in `style.css`. Components reference tokens like `var(--text-primary)`, `var(--code-bg)`, etc.

Dark mode is the default (`:root`). Light mode is activated by toggling the `.light` class on `<html>`. Both token sets are defined in the same stylesheet.

Hardcoded color values are not permitted in components.

Layout variables:

- `--reader-radius` — window corner radius (default `28px`, configurable via settings)
- `--reading-max-width` — document max-width (default `720px`)
- `--surface-opacity` — reader surface opacity (default `0.75`)

---

## Wails Bridge

### Binding Model

Wails v2 generates JavaScript bindings for all exported methods on Go structs passed to `options.App.Bind`. The generated files live in:

```
frontend/wailsjs/go/main/App.js      Function stubs
frontend/wailsjs/go/main/App.d.ts    TypeScript declarations
```

Each Go method becomes an async JavaScript function. The call is an IPC round-trip: the frontend sends a JSON-encoded request, the Go runtime dispatches it to the bound method, and the result returns as a JSON-encoded response.

### Bound Methods

| Method | Signature | Purpose |
|--------|-----------|---------|
| `GetFileTree` | `() → FileNode` | Returns the workspace file tree |
| `ReadFile` | `(relativePath) → string` | Returns file content (path-validated) |
| `GetRootPath` | `() → string` | Returns current workspace root |
| `SetRootPath` | `(path) → error` | Changes workspace root, restarts watcher |
| `OpenFolder` | `() → string` | Opens native directory picker dialog |
| `GetConfig` | `() → Config` | Returns full configuration object |
| `UpdateConfig` | `(cfg) → error` | Replaces configuration, saves to disk |
| `GetTheme` | `() → string` | Returns current theme preference |
| `SetTheme` | `(theme) → error` | Sets theme preference, saves to disk |

### Event System

Go-to-frontend events use the Wails runtime event bus:

```
Event: "file:changed"
Payload: string (relative path of changed file)
Source: watcher callback → wailsRuntime.EventsEmit
Consumer: App.svelte → EventsOn listener
```

When a `file:changed` event arrives, the frontend checks all open tabs. If a tab's path matches the changed file, its content is re-fetched from the backend and updated in the store.

There are no frontend-to-backend events. All frontend-to-backend communication uses bound method calls.

---

## Data Flow

### Application Startup

```
User launches binary (optionally with path argument)
    ↓
main.go parses CLI args, resolves absolute path
    ↓
App.startup() runs:
    ├─ Load config from ~/.config/ais/config.json
    ├─ Store rootPath in config.LastOpenedPath
    └─ Start filesystem watcher on rootPath
    ↓
Wails renders frontend from embedded assets
    ↓
App.svelte onMount():
    ├─ loadSettings() → GetTheme() → apply CSS class
    ├─ loadFileTree() → GetFileTree() → populate store
    └─ Register EventsOn("file:changed") listener
```

### Document Reading

```
User clicks file in sidebar
    ↓
handleFileClick(node) → openTab(path, name)
    ↓
tabs store checks for existing tab
    ├─ Exists → activate it
    └─ New    → readFile(path) → ReadFile(relativePath)
                    ↓
               app.go validates path (no traversal)
                    ↓
               scanner.ReadFileContent(absPath)
                    ↓
               Content returned to frontend
                    ↓
               New Tab created, added to store
                    ↓
               MarkdownViewer renders via markdown-it
```

### Live Reload

```
External editor modifies a .md file
    ↓
fsnotify detects Write/Create event
    ↓
Watcher debounces (100ms)
    ↓
Callback emits "file:changed" event with relative path
    ↓
App.svelte listener receives event
    ↓
For each open tab matching the path:
    └─ readFile(path) → updateTabContent(path, content)
        ↓
       MarkdownViewer re-renders
```

### Workspace Switching

```
User invokes OpenFolder()
    ↓
Native directory picker dialog
    ↓
SetRootPath(newPath):
    ├─ Stop existing watcher
    ├─ Update rootPath
    ├─ Update config (LastOpenedPath, RecentPaths)
    └─ Start new watcher on new path
    ↓
Frontend calls loadFileTree() to refresh sidebar
```

---

## Dependency Graph

### Go Dependencies

```
main.go
  ├─ github.com/wailsapp/wails/v2           Wails framework
  └─ app.go
       ├─ internal/config                    Config management
       ├─ internal/scanner                   Directory scanning, file reading
       ├─ internal/types                     FileNode type
       ├─ internal/watcher                   Filesystem watching
       │    └─ github.com/fsnotify/fsnotify  OS-level file notifications
       └─ github.com/wailsapp/wails/v2/pkg/runtime   Events, dialogs
```

### Frontend Dependencies

```
App.svelte
  ├─ svelte (5.x)                           Component framework
  ├─ @sveltejs/vite-plugin-svelte           Build tooling
  ├─ vite                                    Dev server and bundler
  ├─ tailwindcss (4.x)                       Utility CSS framework
  ├─ stores/
  │    ├─ files.ts    → wailsjs bindings
  │    ├─ tabs.ts     → files store
  │    ├─ settings.ts → wailsjs bindings
  │    └─ ui.ts       → pure frontend state (zoom, focus, opacity, radius, background)
  └─ markdown/
       ├─ markdown-it                        Markdown parsing
       └─ highlight.js (core + 15 languages) Syntax highlighting
```

### Internal Package Dependencies

```
internal/types      (no dependencies — leaf package)
internal/config     (stdlib only: encoding/json, os, path/filepath, sync)
internal/scanner    (depends on: internal/types)
internal/watcher    (depends on: fsnotify — no internal dependencies)
```

No circular dependencies exist. `types` is the leaf. `scanner` and `watcher` are siblings that do not reference each other.

---

## Build and Deployment

### Development

```bash
wails dev
```

Starts the Go backend and the Vite dev server simultaneously. Frontend changes hot-reload. Go changes trigger a rebuild.

### Production Build

```bash
wails build
```

Produces a single binary at `build/bin/ais`.

Build process:

```
1. Vite compiles Svelte → frontend/dist/
2. Go embeds frontend/dist/ via //go:embed
3. Go compiles everything into a single binary
4. The binary contains the full application: backend + frontend + assets
```

The output binary has zero runtime dependencies. No Node.js, no npm, no separate asset files.

### Frontend-Only Build

```bash
cd frontend && npm run build     # Compile frontend
cd frontend && npm run check     # Svelte type checking
```

### Testing

```bash
go test ./internal/...           # Run all Go tests
```

Test patterns:

- `t.TempDir()` for filesystem tests (automatic cleanup)
- `t.Helper()` on test utility functions
- Table-driven subtests for combinatorial coverage

Frontend has no test framework. Type checking via `npm run check` is the current verification.

---

## Design Constraints

### Domain Logic Isolation

All domain logic lives in `internal/` packages. The `main` package contains only the entry point and the Wails-bound `App` struct. The `App` struct delegates all work to internal packages.

### Standard Library Preference

Go code prefers stdlib over third-party dependencies. The only external Go dependencies are:

- `wails/v2` — required for the desktop runtime
- `fsnotify` — required for cross-platform filesystem events

Everything else (JSON, file I/O, path manipulation, concurrency) uses the Go standard library.

### Concurrent Configuration Access

The `Config` struct is accessed from multiple goroutines (main thread, watcher callbacks, Wails RPC handlers). All access is mediated by `Manager` with `sync.RWMutex`. Direct field access on `Config` is never safe outside of `Get()` and `Update()`.

### Path Traversal Prevention

`ReadFile` in `app.go` resolves the requested path to an absolute path and validates that it begins with `rootPath + os.PathSeparator`. This prevents directory traversal attacks via relative paths like `../../etc/passwd`.

### XSS Prevention

markdown-it is configured with `html: false`. Raw HTML in markdown source is not rendered. All code block content is escaped via `md.utils.escapeHtml()` before insertion into the DOM.

### File Size Limit

`scanner.ReadFileContent` rejects files exceeding 10 MB. This prevents memory exhaustion from accidentally opening large binary or log files.

### Directory Filtering

Both `scanner` and `watcher` maintain identical `skipDirs` sets. The following directories are always excluded:

```
.git  node_modules  .svn  __pycache__  vendor  .venv
```

The watcher additionally skips all hidden directories (prefixed with `.`) except the root itself.

---

## LLM Streaming — Architecture Addition

### Overview

The LLM streaming feature introduces the first network-connected subsystem in ais. It connects to the Anthropic API to stream Claude responses and render them as markdown in real-time within the existing reading surface.

The architecture adds one new internal package (`internal/llm`), five new Wails bindings, three new backend-to-frontend events, one new frontend store, and one new npm dependency.

---

### New Package: internal/llm

```
internal/
  llm/
    client.go           HTTP streaming client using anthropic-sdk-go
    types.go            StreamRequest, StreamChunk, StreamError
    keystore.go         API key resolution chain (env > keychain > file)
    client_test.go      Tests with httptest SSE mock
    keystore_test.go    Keystore resolution and permissions tests
```

**client.go** — Streaming Client

Responsibilities:

- Create an Anthropic API client from an API key
- Execute a streaming request with `context.Context` for cancellation
- Read SSE events from the response stream
- Batch events at 50ms intervals to reduce frontend event rate
- Emit `StreamChunk` values via a callback function
- Map API errors to typed `StreamError` values

The client is stateless per stream. A new stream creates a new HTTP connection. Cancellation is achieved by cancelling the context, which closes the underlying HTTP connection.

Batching strategy:

```
SSE event arrives
    ↓
Append text to batch buffer
    ↓
50ms timer running?
    ├─ Yes → Continue accumulating
    └─ No  → Start 50ms timer
                ↓
            Timer fires → Emit batch as StreamChunk, reset buffer
```

On stream completion, the buffer is flushed immediately regardless of the timer.

**keystore.go** — API Key Storage

Resolution chain (checked in order, first match wins):

```
1. Environment variable: ANTHROPIC_API_KEY
2. OS keychain: go-keyring (service: "ais", user: "api-key")
3. Credentials file: ~/.config/ais/credentials.json (permissions: 0600)
```

The credentials file is a separate file from `config.json`. The config file has `0644` permissions (world-readable). The credentials file has `0600` permissions (owner-only). This separation is a security boundary — the API key must never appear in `config.json`.

**types.go** — Data Types

```go
type StreamRequest struct {
    Prompt string `json:"prompt"`
    Model  string `json:"model"`
}

type StreamChunk struct {
    Text        string `json:"text"`
    Done        bool   `json:"done"`
    TotalTokens int    `json:"totalTokens,omitempty"`
}

type StreamError struct {
    Code    string `json:"code"`    // "network", "auth", "rate_limit", "cancelled", "api"
    Message string `json:"message"`
}
```

---

### New Dependencies

| Dependency | Purpose | Justification |
|------------|---------|---------------|
| `github.com/anthropics/anthropic-sdk-go` | Claude API client with SSE streaming | Official SDK; handles auth, retries, SSE parsing |
| `github.com/zalando/go-keyring` | Cross-platform OS keychain access | macOS Keychain, Linux Secret Service (D-Bus), Windows Credential Manager |
| `morphdom` (npm) | Incremental DOM diffing | Minimal DOM mutations during streaming; avoids full re-render per chunk |

---

### Extended App Struct

```go
type App struct {
    ctx          context.Context
    rootPath     string
    cfgMgr       *config.Manager
    watcher      *watcher.Watcher
    llmClient    *llm.Client           // Lazy-initialized on first stream
    streamCancel context.CancelFunc    // nil when no stream active
    streamMu     sync.Mutex            // Guards streamCancel
}
```

`llmClient` is created lazily on the first `StartStream` call. It is invalidated when the API key changes via `SetAPIKey`. The `streamMu` mutex protects `streamCancel` from concurrent access (e.g., `CancelStream` called while `StartStream` goroutine is setting up).

---

### New Data Flow: LLM Streaming

```
User invokes "Ask AI" from Command Palette
    ↓
Frontend calls StartStream(prompt) via Wails binding
    ↓
App.StartStream():
    ├─ Acquire streamMu
    ├─ Check no active stream (streamCancel == nil)
    ├─ Resolve API key via keystore.GetAPIKey()
    ├─ Create llmClient if nil
    ├─ Create context.WithCancel
    ├─ Store cancel func in streamCancel
    ├─ Release streamMu
    └─ Launch goroutine:
         ├─ Call llmClient.Stream(ctx, req, emitFunc)
         │     ↓
         │   emitFunc(chunk):
         │     └─ EventsEmit("llm:chunk", chunk)
         ├─ On success: EventsEmit("llm:done", {totalTokens})
         ├─ On error:   EventsEmit("llm:error", {code, message})
         └─ Clear streamCancel (acquire streamMu)
```

```
User clicks Stop in ControlStrip
    ↓
Frontend calls CancelStream() via Wails binding
    ↓
App.CancelStream():
    ├─ Acquire streamMu
    ├─ Call streamCancel() if non-nil
    ├─ Set streamCancel = nil
    └─ Release streamMu
    ↓
Context cancellation propagates to HTTP client
    ↓
Stream goroutine receives context.Canceled
    ↓
Emits llm:error with code "cancelled" (or silently completes)
```

---

### New Event System

| Event | Direction | Payload | Rate |
|-------|-----------|---------|------|
| `llm:chunk` | Backend → Frontend | `StreamChunk{text, done: false}` | <= 20/sec (50ms batching) |
| `llm:done` | Backend → Frontend | `StreamChunk{done: true, totalTokens}` | Once per stream |
| `llm:error` | Backend → Frontend | `StreamError{code, message}` | Once per stream |

These events follow the same pattern as the existing `file:changed` event — emitted via `wailsRuntime.EventsEmit`, consumed via `EventsOn` in the frontend.

---

### Frontend Architecture Changes

**New store: `stores/stream.ts`**

Manages stream lifecycle state. Listens for `llm:chunk`, `llm:done`, `llm:error` events. Coordinates with the tabs store to update stream tab content.

**Extended Tab interface:**

The `Tab` type gains two fields: `type: 'file' | 'stream'` and `streamActive?: boolean`. Existing file tabs default to `type: 'file'`. Stream tabs are created with a generated ID (UUID or timestamp-based) rather than a file path.

**MarkdownViewer changes:**

When rendering a stream tab with `streamActive === true`, the viewer uses `morphdom` to apply incremental DOM diffs instead of replacing `innerHTML`. This produces minimal DOM mutations — only the new content is inserted. Once streaming completes, the viewer reverts to standard rendering.

Scroll behavior during streaming: if the user is at the bottom of the document (within 50px of scroll end), new content auto-scrolls into view. If the user has scrolled up to read earlier content, the scroll position is locked.

---

### Updated Dependency Graph

```
main.go
  ├─ github.com/wailsapp/wails/v2
  └─ app.go
       ├─ internal/config
       ├─ internal/scanner
       ├─ internal/types
       ├─ internal/watcher
       │    └─ github.com/fsnotify/fsnotify
       ├─ internal/llm                              ← NEW
       │    ├─ github.com/anthropics/anthropic-sdk-go  ← NEW
       │    └─ github.com/zalando/go-keyring           ← NEW
       └─ github.com/wailsapp/wails/v2/pkg/runtime
```

```
Frontend
  ├─ stores/
  │    ├─ files.ts
  │    ├─ tabs.ts      (extended Tab interface)
  │    ├─ settings.ts
  │    ├─ ui.ts
  │    └─ stream.ts    ← NEW
  └─ markdown/
       ├─ markdown-it
       ├─ highlight.js
       └─ morphdom     ← NEW
```

### Internal Package Dependencies (Updated)

```
internal/types      (no dependencies — leaf package)
internal/config     (stdlib only)
internal/scanner    (depends on: internal/types)
internal/watcher    (depends on: fsnotify)
internal/llm        (depends on: anthropic-sdk-go, go-keyring — no internal dependencies)
```

No circular dependencies. The `llm` package is a new sibling to `scanner` and `watcher`. It has no dependencies on other internal packages.

---

### ADR-1: morphdom for Incremental DOM Diffing

**Status:** Accepted

**Context:** During LLM streaming, markdown content grows with each chunk. The viewer must re-render the accumulated content without visible flicker or layout shift. Full `innerHTML` replacement causes the entire DOM subtree to be destroyed and recreated, losing scroll position, selection state, and causing visible flicker.

**Decision:** Use `morphdom` for DOM diffing during active streams. morphdom compares two DOM trees and applies the minimal set of mutations to transform one into the other. It is 4KB gzipped, has zero dependencies, and is purpose-built for this exact problem.

**Alternatives considered:**

- **Virtual DOM (Svelte built-in):** Svelte 5 reactivity operates at the component level. The markdown HTML is rendered outside Svelte's reactive system (via `{@html}`), so Svelte's diffing does not apply here.
- **lit-html / uhtml:** Template-literal-based rendering. Adds a templating paradigm that conflicts with the existing markdown-it pipeline.
- **Manual DOM diffing:** Error-prone and harder to maintain than a well-tested library.
- **Full innerHTML on every chunk:** Causes flicker and scroll reset. Fails NFR-1 performance targets at scale.

**Consequences:**
- New npm dependency: `morphdom` (~4KB gzipped)
- morphdom is only active during streaming; static content continues to use `innerHTML`
- The viewer must manage two rendering paths (streaming vs. static)

---

### ADR-2: anthropic-sdk-go for API Client

**Status:** Accepted

**Context:** The LLM streaming feature requires an HTTP client that can authenticate with the Anthropic API and parse SSE (Server-Sent Events) streams. The SSE protocol has edge cases around reconnection, event types, and buffering.

**Decision:** Use `anthropic-sdk-go`, the official Anthropic Go SDK. It handles authentication, SSE parsing, request construction, and error mapping.

**Alternatives considered:**

- **Hand-rolled HTTP + SSE:** More control but requires maintaining SSE parsing logic, auth header management, and error mapping. High surface area for bugs.
- **go-sse library + manual HTTP:** Separates SSE parsing from API semantics. Requires glue code for auth, models, and error types.

**Consequences:**
- New Go dependency: `github.com/anthropics/anthropic-sdk-go`
- Transitive dependencies from the SDK (HTTP, JSON utilities)
- SDK version must be tracked for API changes

---

### ADR-3: go-keyring for API Key Storage

**Status:** Accepted

**Context:** The API key is a secret credential. The existing `config.json` file has `0644` permissions (world-readable). Storing the API key there would expose it to other users on shared systems. A secure storage mechanism is needed.

**Decision:** Use `go-keyring` for cross-platform OS keychain access. Fall back to a dedicated credentials file (`~/.config/ais/credentials.json`) with `0600` permissions when the keychain is unavailable (e.g., headless Linux without D-Bus).

**Alternatives considered:**

- **Encrypted file only:** Requires a master password or key derivation, adding UX complexity.
- **Keychain only (no fallback):** Fails on headless Linux or systems without a secrets service.
- **Store in config.json:** Violates security requirements. Config file is `0644`.

**Consequences:**
- New Go dependency: `github.com/zalando/go-keyring`
- Requires D-Bus secret service on Linux (GNOME Keyring, KDE Wallet) for keychain path
- Fallback file path is separate from config.json to maintain the security boundary
- The resolution chain (env > keychain > file) is documented and tested

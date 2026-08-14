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
- Configure Wails application options (window size, background color, asset server)
- Embed the compiled frontend via `//go:embed all:frontend/dist`
- Start the Wails runtime

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
App.svelte                          Shell layout, keyboard handlers, event listener
  ├─ Sidebar.svelte                 File explorer panel, search, theme toggle
  │     ├─ FileTree.svelte          Tree container
  │     │     └─ FileTreeNode.svelte   Recursive node (expand/collapse, filtering)
  │     └─ ThemeToggle.svelte       Theme cycle button (system → light → dark)
  ├─ TabBar.svelte                  Tab strip (click to switch, middle-click to close)
  ├─ MarkdownViewer.svelte          Renders active tab content, heading collapse, scroll restore
  └─ WelcomeScreen.svelte           Shown when no tabs are open
```

### Layout

The shell uses CSS Grid with named areas:

```
grid-template-areas:
  "sidebar tabs"
  "sidebar viewer"
```

Grid dimensions:

```
Columns: var(--sidebar-width) 1fr
Rows:    var(--tab-height) 1fr
```

When the sidebar is hidden (`Ctrl+B`), the first column collapses to `0` with a 200ms transition.

### Svelte 5 Runes

All components use Svelte 5 runes mode:

- `$state()` for local reactive state
- `$derived()` for computed values
- `$effect()` for side effects
- `$props()` for component inputs

Legacy `let` reactivity is not used.

### Stores

Three store modules manage global state:

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
applyTheme(mode): void                  Toggle CSS class on <html>
```

System theme preference changes are tracked via `matchMedia` listener.

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

All colors are defined as CSS custom properties in `style.css`. Components reference tokens like `var(--text-primary)`, `var(--bg-code)`, etc.

Theme switching toggles the `.dark` class on `<html>`. Light and dark token sets are defined in the same stylesheet.

Hardcoded color values are not permitted in components.

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
  │    └─ settings.ts → wailsjs bindings
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

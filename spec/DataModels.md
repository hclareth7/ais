# Data Models — ais

This document defines every data type, store, and state contract in the ais application. It serves as the authoritative reference for data shape, persistence, serialization, and cross-boundary contracts between the Go backend and Svelte frontend.

---

## Backend Models (Go)

### FileNode

Recursive tree node representing a file or directory in the scanned workspace.

**Package:** `internal/types`

| Field | Type | JSON Key | Description |
|-------|------|----------|-------------|
| Name | `string` | `name` | Base name of the file or directory |
| Path | `string` | `path` | Relative path from the workspace root (empty string for root node itself) |
| IsDir | `bool` | `isDir` | `true` for directories, `false` for files |
| Children | `[]*FileNode` | `children,omitempty` | Child nodes; omitted from JSON when `nil` or empty |

```go
type FileNode struct {
    Name     string      `json:"name"`
    Path     string      `json:"path"`
    IsDir    bool        `json:"isDir"`
    Children []*FileNode `json:"children,omitempty"`
}
```

**Constraints:**

- Path is always relative to `rootPath`. The root node itself has `Path: ""`.
- Only markdown files (`.md`, `.markdown`, case-insensitive) appear as file nodes.
- Directories appear only if they contain at least one markdown file (directly or transitively).
- Children are sorted: directories first, then files, both alphabetically (case-insensitive).
- Skipped directories: `.git`, `node_modules`, `.svn`, `__pycache__`, `vendor`, `.venv`.
- Hidden directories (name starts with `.`) are skipped during scan, except the root itself.

**Produced by:** `scanner.ScanDirectory(root string) (*FileNode, error)`

**Consumed by:** Frontend via Wails binding `GetFileTree()`, serialized to JSON automatically by Wails.

---

### Config

User preferences persisted to disk.

**Package:** `internal/config`

| Field | Type | JSON Key | Default | Description |
|-------|------|----------|---------|-------------|
| Theme | `string` | `theme` | `"system"` | Theme preference: `"light"`, `"dark"`, or `"system"` |
| SSHKeyPaths | `[]string` | `sshKeyPaths` | `[]` | Reserved for future SSH key integration |
| IgnoreDirs | `[]string` | `ignoreDirs` | See below | Directory names excluded from scanning |
| LastOpenedPath | `string` | `lastOpenedPath` | `""` | Absolute path of last opened workspace |
| FontSize | `int` | `fontSize` | `16` | Base font size in pixels |
| SidebarWidth | `int` | `sidebarWidth` | `260` | Sidebar width in pixels |
| RecentPaths | `[]string` | `recentPaths` | `[]` | Recently opened workspace paths (max 10, most recent first) |

```go
type Config struct {
    Theme          string   `json:"theme"`
    SSHKeyPaths    []string `json:"sshKeyPaths"`
    IgnoreDirs     []string `json:"ignoreDirs"`
    LastOpenedPath string   `json:"lastOpenedPath"`
    FontSize       int      `json:"fontSize"`
    SidebarWidth   int      `json:"sidebarWidth"`
    RecentPaths    []string `json:"recentPaths"`
}
```

**Default IgnoreDirs:**

```json
[".git", "node_modules", ".svn", "__pycache__", "vendor", ".venv"]
```

**Persistence:**

- File: `~/.config/ais/config.json`
- Format: JSON with 2-space indentation
- Permissions: `0644` for the file, `0755` for the directory
- On missing file: defaults are used silently (no error)
- On parse error: error is returned, defaults remain

---

### Config Manager

Concurrent-safe wrapper around Config providing thread-safe read/write access.

**Package:** `internal/config`

```go
type Manager struct {
    mu       sync.RWMutex
    cfg      Config
    filePath string
}
```

**Methods:**

| Method | Lock Type | Description |
|--------|-----------|-------------|
| `NewManager() *Manager` | None | Creates manager with defaults and path `~/.config/ais/config.json` |
| `Load() error` | Write lock | Reads and parses config from disk; no-op if file missing |
| `Save() error` | Write lock | Marshals and writes config to disk; creates directory if needed |
| `Get() Config` | Read lock | Returns a copy of the current config |
| `Update(fn func(*Config)) error` | Write lock | Applies mutation function to config in place |

**Invariants:**

- All config field access must go through `Get()` and `Update()`. Direct field access is never safe.
- `Update()` does not auto-save. Callers must call `Save()` explicitly when persistence is needed.
- `Load()` replaces the entire config struct. Partial updates from disk are not supported.

---

### Watcher

Filesystem event monitor for live reload of markdown content.

**Package:** `internal/watcher`

```go
type Watcher struct {
    fsWatcher *fsnotify.Watcher
    root      string
    callback  func(string)
    done      chan struct{}
    once      sync.Once
    stopped   bool
    stoppedMu sync.RWMutex
}
```

**Behavior:**

- Recursively watches all non-skipped directories under `root`.
- Only fires callback for markdown file events (`.md`, `.markdown`).
- Events are debounced at 100ms per file path.
- Callback receives the relative path from root.
- New directories created at runtime are automatically added to the watch list.
- Same `skipDirs` set as the scanner.

---

## Frontend Models (TypeScript)

### FileNode

Mirrors the Go `FileNode` struct. Deserialized from Wails binding JSON response.

**File:** `frontend/src/lib/stores/files.ts`

```typescript
export interface FileNode {
  name: string;
  path: string;
  isDir: boolean;
  children?: FileNode[];
}
```

**Contract:** Field names and types match the Go struct's JSON serialization exactly. The `children` field is `undefined` (not present) for leaf file nodes and empty directories that were pruned.

---

### Tab

Represents an open document in the tab bar.

**File:** `frontend/src/lib/stores/tabs.ts`

```typescript
export interface Tab {
  id: string;
  name: string;
  path: string;
  content: string;
  scrollPos: number;
}
```

| Field | Type | Description |
|-------|------|-------------|
| id | `string` | Unique identifier; set to the file's relative path |
| name | `string` | Display name; the file's base name |
| path | `string` | Relative path from workspace root; matches `FileNode.path` |
| content | `string` | Raw markdown source loaded from backend |
| scrollPos | `number` | Saved vertical scroll position (pixels from top) |

**Constraints:**

- `id === path` — the relative path serves as both identifier and path.
- Maximum 20 tabs open simultaneously. When exceeded, the oldest tab (index 0) is evicted (FIFO).
- Duplicate paths are prevented: opening an already-open file activates its existing tab.
- Content is loaded once on tab creation and updated on `file:changed` events from the watcher.

---

### ThemeMode

Union type for theme preference values.

**File:** `frontend/src/lib/stores/settings.ts`

```typescript
export type ThemeMode = 'light' | 'dark' | 'system';
```

**Behavior:**

- `'system'` defers to OS preference via `prefers-color-scheme` media query.
- Changes to OS preference are detected in real time and applied when theme is `'system'`.
- Theme is applied by toggling the `dark` class on `<html>`.

---

## Stores (Svelte)

All stores use Svelte's `writable` or `derived` from `svelte/store`.

### File Stores

**File:** `frontend/src/lib/stores/files.ts`

| Store | Type | Initial Value | Description |
|-------|------|---------------|-------------|
| `fileTree` | `writable<FileNode \| null>` | `null` | Root node of the scanned directory tree |
| `rootPath` | `writable<string>` | `''` | Absolute path of the current workspace root |

**Functions:**

| Function | Return | Description |
|----------|--------|-------------|
| `loadFileTree()` | `Promise<void>` | Calls `GetFileTree()` and `GetRootPath()` Wails bindings; populates both stores |
| `readFile(path)` | `Promise<string>` | Calls `ReadFile(path)` Wails binding; returns markdown content or error string |

---

### Tab Stores

**File:** `frontend/src/lib/stores/tabs.ts`

| Store | Type | Initial Value | Description |
|-------|------|---------------|-------------|
| `tabs` | `writable<Tab[]>` | `[]` | Ordered list of open tabs |
| `activeTabId` | `writable<string \| null>` | `null` | ID of the currently active tab |
| `activeTab` | `derived<Tab \| null>` | `null` | Computed: the Tab object matching `activeTabId`, or `null` |

**Functions:**

| Function | Signature | Description |
|----------|-----------|-------------|
| `openTab` | `(path: string, name: string) => Promise<void>` | Opens or activates a tab; loads content from backend; enforces 20-tab limit |
| `closeTab` | `(id: string) => void` | Removes tab; if active, activates adjacent tab (prefer next, then previous) |
| `nextTab` | `() => void` | Activates next tab (wraps around) |
| `prevTab` | `() => void` | Activates previous tab (wraps around) |
| `updateTabContent` | `(path: string, content: string) => void` | Updates content for a tab by path (used by file watcher) |
| `saveScrollPos` | `(id: string, pos: number) => void` | Persists scroll position for a tab |

**Tab Close Behavior:**

When closing the active tab:
1. If other tabs remain and closed tab was not the last: activate the tab at the same index.
2. If closed tab was the last in the list: activate the new last tab.
3. If no tabs remain: set `activeTabId` to `null`.

---

### Settings Store

**File:** `frontend/src/lib/stores/settings.ts`

| Store | Type | Initial Value | Description |
|-------|------|---------------|-------------|
| `theme` | `writable<ThemeMode>` | `'system'` | Current theme preference |

**Functions:**

| Function | Signature | Description |
|----------|-----------|-------------|
| `loadSettings` | `() => Promise<void>` | Loads theme from backend via `GetTheme()`; applies to DOM |
| `setTheme` | `(mode: ThemeMode) => Promise<void>` | Updates store, applies to DOM, persists via `SetTheme()` |
| `applyTheme` | `(mode: ThemeMode) => void` | Toggles `dark` class on `<html>` based on mode and OS preference |

---

## State Persistence Summary

| State | Persistence | Location | Mechanism |
|-------|-------------|----------|-----------|
| Config (all fields) | Disk | `~/.config/ais/config.json` | Go `config.Manager.Save()` |
| Theme | Disk (via Config) | `~/.config/ais/config.json` | Synced Go <-> Frontend via `GetTheme`/`SetTheme` |
| File tree | In-memory | Frontend store | Rebuilt on load and directory change |
| Open tabs | In-memory | Frontend store | Lost on application close |
| Active tab | In-memory | Frontend store | Lost on application close |
| Scroll positions | In-memory | Frontend store (per tab) | Lost on application close |
| Recent paths | Disk (via Config) | `~/.config/ais/config.json` | Updated on `SetRootPath`, max 10 entries |
| Last opened path | Disk (via Config) | `~/.config/ais/config.json` | Updated on startup and `SetRootPath` |

---

## Cross-Boundary Contract

Data crosses the Go-Frontend boundary via Wails bindings. Serialization is automatic JSON.

### Go to Frontend (JSON serialization)

| Go Type | JSON | TypeScript Type |
|---------|------|-----------------|
| `string` | `"value"` | `string` |
| `int` | `16` | `number` |
| `bool` | `true` | `boolean` |
| `[]string` | `["a","b"]` | `string[]` |
| `*FileNode` | `{...}` | `FileNode` |
| `Config` | `{...}` | Untyped (consumed as-is) |

### Frontend to Go (JSON deserialization)

| TypeScript | JSON | Go Type |
|------------|------|---------|
| `string` | `"value"` | `string` |
| `ThemeMode` | `"dark"` | `string` |
| `Config object` | `{...}` | `config.Config` |

**Key contract rules:**

1. JSON field names use `camelCase` (matching Go struct tags).
2. `Children` is omitted from JSON when nil/empty (`omitempty`).
3. `Config` is passed as a full object on `UpdateConfig` — partial updates are not supported.
4. All paths in `FileNode` are relative to the workspace root.
5. All paths in `Config` (LastOpenedPath, RecentPaths) are absolute.

---

## Constants

| Constant | Value | Location | Description |
|----------|-------|----------|-------------|
| `maxFileSize` | `10 * 1024 * 1024` (10MB) | `internal/scanner/scanner.go` | Maximum file size `ReadFileContent` will read |
| Max tabs | `20` | `frontend/src/lib/stores/tabs.ts` | FIFO eviction when exceeded |
| Max recent paths | `10` | `app.go` | Oldest entries dropped beyond this limit |
| Debounce interval | `100ms` | `internal/watcher/watcher.go` | Per-file debounce for filesystem events |
| Default font size | `16` | `internal/config/defaults.go` | Pixels |
| Default sidebar width | `260` | `internal/config/defaults.go` | Pixels |

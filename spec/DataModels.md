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
- Dark mode is the default (`:root` CSS). Light mode is applied by toggling the `.light` class on `<html>`.

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
| `applyTheme` | `(mode: ThemeMode) => void` | Toggles `.light` class on `<html>` based on mode and OS preference |

---

### UI Store

**File:** `frontend/src/lib/stores/ui.ts`

| Store | Type | Initial Value | Description |
|-------|------|---------------|-------------|
| `zoomLevel` | `writable<number>` | `100` | Document zoom percentage |
| `readingWidth` | `writable<number>` | `720` | Document max-width in pixels |
| `focusMode` | `writable<boolean>` | `false` | Focus mode state (hides chrome) |
| `commandPaletteOpen` | `writable<boolean>` | `false` | Command palette visibility |
| `opacity` | `writable<number>` | `75` | Reader surface opacity (40–100%) |
| `readerRadius` | `writable<number>` | `28` | Window corner radius in pixels |
| `backgroundMode` | `writable<string>` | `'gradient'` | Background mode: `"gradient"`, `"solid"`, `"frost"` |

**Functions:**

| Function | Signature | Description |
|----------|-----------|-------------|
| `zoomIn` | `() => void` | Step up through `ZOOM_LEVELS` array |
| `zoomOut` | `() => void` | Step down through `ZOOM_LEVELS` array |
| `resetZoom` | `() => void` | Set zoom to 100% |
| `toggleFocusMode` | `() => void` | Toggle `focusMode` |
| `toggleCommandPalette` | `() => void` | Toggle `commandPaletteOpen` |
| `changeOpacity` | `(delta: number) => void` | Adjust opacity by delta, clamp 40–100 |
| `setReaderRadius` | `(px: number) => void` | Set `--reader-radius` CSS variable |
| `setBackgroundMode` | `(mode: string) => void` | Toggle background layer visibility |

**Zoom Levels:**

```typescript
const ZOOM_LEVELS = [50, 75, 90, 100, 110, 125, 150, 175, 200];
```

Zoom steps through discrete levels rather than arbitrary values.

**Opacity:**

Keyboard shortcuts: `Ctrl+Shift+Plus` (increase), `Ctrl+Shift+Minus` (decrease), step size: 5%.

Contrast warning shown when opacity drops below 55%.

**Background Modes:**

| Mode | Sky | Mountains | Fog | Stars |
|------|-----|-----------|-----|-------|
| `gradient` | Visible | Visible | Visible | Visible (dark only) |
| `solid` | Hidden | Hidden | Hidden | Hidden |
| `frost` | 30% opacity | Hidden | Visible | Hidden |

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
| Zoom level | In-memory | Frontend store (`ui.ts`) | Reset to 100% on application close |
| Reading width | In-memory | Frontend store (`ui.ts`) | Reset to 720px on application close |
| Focus mode | In-memory | Frontend store (`ui.ts`) | Always starts as `false` |
| Opacity | In-memory | Frontend store (`ui.ts`) | Reset to 75% on application close |
| Reader radius | In-memory | Frontend store (`ui.ts`) | Reset to 28px on application close |
| Background mode | In-memory | Frontend store (`ui.ts`) | Reset to `gradient` on application close |

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
| Zoom levels | `[50,75,90,100,110,125,150,175,200]` | `frontend/src/lib/stores/ui.ts` | Discrete zoom steps |
| Default zoom | `100` | `frontend/src/lib/stores/ui.ts` | Percentage |
| Default reading width | `720` | `frontend/src/lib/stores/ui.ts` | Pixels |
| Min/max reading width | `500–1000` | `frontend/src/lib/stores/ui.ts` | Pixels |
| Default opacity | `75` | `frontend/src/lib/stores/ui.ts` | Percentage |
| Min/max opacity | `40–100` | `frontend/src/lib/stores/ui.ts` | Percentage |
| Opacity step | `5` | `frontend/src/lib/stores/ui.ts` | Keyboard shortcut increment |
| Contrast warning threshold | `55` | `frontend/src/lib/stores/ui.ts` | Show warning below this opacity |
| Default reader radius | `28` | `frontend/src/lib/stores/ui.ts` | Pixels |
| Radius options | `[20, 28, 36, 48]` | `frontend/src/lib/stores/ui.ts` | Pixels |
| Default background mode | `gradient` | `frontend/src/lib/stores/ui.ts` | String |
| Stream batch interval | `50ms` | `internal/llm/client.go` | Chunk batching period |
| Max tokens per stream | `4096` | `internal/llm/client.go` | Anthropic API max_tokens parameter |
| Default model | `claude-sonnet-5` | `internal/llm/client.go` | Quality-first default per Design.md |
| `--stream-glow` | CSS custom property | `style.css` | Streaming border pulse color (per Design.md) |
| `--stream-caret` | CSS custom property | `style.css` | Streaming caret color (per Design.md) |
| `--stream-status-*` | CSS custom properties | `style.css` | Status indicator colors (per Design.md) |
| `--stream-stop-*` | CSS custom properties | `style.css` | Stop button colors (per Design.md) |
| Stream tab label format | `"AI: {30 chars}..."` | `frontend/src/lib/stores/tabs.ts` | Per Design.md |
| Stream border pulse | `2s ease-in-out, 0.15-0.55` | `style.css` | Animation spec (per Design.md) |
| Streaming caret blink | `1s step-end` | `style.css` | Animation spec (per Design.md) |
| Caret fade-out | `200ms` | `style.css` | On stream completion (per Design.md) |
| Code block accent fade | `120ms` | `style.css` | When fence closes (per Design.md) |
| Error fade-in | `80ms` | `style.css` | Inline error appearance (per Design.md) |
| Tab indicator fade-out | `200ms` | `style.css` | Streaming tab on completion (per Design.md) |

---

## LLM Streaming — Data Model Additions

### Backend Models (Go)

#### StreamRequest

Request payload for initiating an LLM stream.

**Package:** `internal/llm`

| Field | Type | JSON Key | Description |
|-------|------|----------|-------------|
| Prompt | `string` | `prompt` | Free-form text prompt from the user |
| Model | `string` | `model` | Claude model identifier (e.g., `"claude-haiku-4-5"`) |

```go
type StreamRequest struct {
    Prompt string `json:"prompt"`
    Model  string `json:"model"`
}
```

**Constraints:**
- Prompt must be non-empty
- Model must be one of the supported models; if empty, defaults to `claude-sonnet-5`

**Produced by:** `App.StartStream()` in `app.go`

**Consumed by:** `llm.Client.Stream()` in `internal/llm/client.go`

---

#### StreamChunk

A batch of streamed content. This is the payload for `llm:chunk` and `llm:done` events.

**Package:** `internal/llm`

| Field | Type | JSON Key | Description |
|-------|------|----------|-------------|
| Text | `string` | `text` | Accumulated text since the last chunk (empty on `done`) |
| Done | `bool` | `done` | `true` when the stream has completed |
| TotalTokens | `int` | `totalTokens,omitempty` | Total tokens consumed; present only when `Done` is `true` |

```go
type StreamChunk struct {
    Text        string `json:"text"`
    Done        bool   `json:"done"`
    TotalTokens int    `json:"totalTokens,omitempty"`
}
```

**Constraints:**
- When `Done` is `false`: `Text` contains the new text since the last chunk, `TotalTokens` is zero/omitted
- When `Done` is `true`: `Text` is empty, `TotalTokens` contains the final count
- Text is batched at 50ms intervals — multiple SSE events may be concatenated into a single chunk

**Produced by:** `llm.Client.Stream()` emit callback

**Consumed by:** Frontend via `EventsOn("llm:chunk")` and `EventsOn("llm:done")`

---

#### StreamError

Error payload for stream failures. This is the payload for `llm:error` events.

**Package:** `internal/llm`

| Field | Type | JSON Key | Description |
|-------|------|----------|-------------|
| Code | `string` | `code` | Error category identifier |
| Message | `string` | `message` | Human-readable error description |

```go
type StreamError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}
```

**Error codes:**

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `network` | N/A | DNS failure, timeout, connection refused |
| `auth` | 401 | Invalid or expired API key |
| `rate_limit` | 429 | Too many requests; may include retry-after |
| `cancelled` | N/A | User cancelled via `CancelStream()` |
| `api` | 4xx/5xx | Other Anthropic API error |

**Produced by:** `llm.Client.Stream()` on error

**Consumed by:** Frontend via `EventsOn("llm:error")`

---

#### Credentials

API key storage format for the fallback credentials file.

**Package:** `internal/llm`

| Field | Type | JSON Key | Description |
|-------|------|----------|-------------|
| AnthropicAPIKey | `string` | `anthropic_api_key` | The Anthropic API key |

```go
type Credentials struct {
    AnthropicAPIKey string `json:"anthropic_api_key"`
}
```

**Persistence:**
- File: `~/.config/ais/credentials.json`
- Format: JSON
- Permissions: `0600` (owner-only read/write)
- This file is separate from `config.json` (`0644`) to maintain the security boundary

---

### Frontend Models (TypeScript)

#### Extended Tab

The Tab interface gains two fields for stream support.

**File:** `frontend/src/lib/stores/tabs.ts`

```typescript
export interface Tab {
  id: string;
  name: string;
  path: string;
  content: string;
  scrollPos: number;
  type: 'file' | 'stream';       // NEW — defaults to 'file'
  streamActive?: boolean;         // NEW — true during active streaming
}
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| type | `'file' \| 'stream'` | `'file'` | Distinguishes file tabs from stream tabs |
| streamActive | `boolean \| undefined` | `undefined` | `true` while content is being streamed; `false` or `undefined` otherwise |

**Constraints:**
- File tabs: `type === 'file'`, `streamActive` is always `undefined`
- Stream tabs: `type === 'stream'`, `id` is a generated value (not a file path), `path` is empty
- Stream tab `name` uses format "AI: {first 30 characters}..." (per Design.md)
- When `streamActive` transitions from `true` to `false`, the tab behaves identically to a file tab
- Closing a stream tab with `streamActive === true` triggers `CancelStream()`

---

#### StreamChunk (TypeScript)

Mirrors the Go `StreamChunk` struct.

**File:** `frontend/src/lib/stores/stream.ts`

```typescript
export interface StreamChunk {
  text: string;
  done: boolean;
  totalTokens?: number;
}
```

---

#### StreamError (TypeScript)

Mirrors the Go `StreamError` struct.

**File:** `frontend/src/lib/stores/stream.ts`

```typescript
export interface StreamError {
  code: 'network' | 'auth' | 'rate_limit' | 'cancelled' | 'api';
  message: string;
}
```

---

### StreamState (TypeScript)

State enum for the stream lifecycle. Matches the 6-state machine defined in `spec/ux.md`.

**File:** `frontend/src/lib/stores/stream.ts`

```typescript
export type StreamState = 'idle' | 'prompting' | 'streaming' | 'complete' | 'cancelled' | 'error';
```

| State | Meaning | Visual Treatment |
|-------|---------|-----------------|
| `idle` | No stream activity | Default UI |
| `prompting` | Command Palette is in prompt input mode | Palette open with textarea |
| `streaming` | Active stream, chunks arriving | Border pulse, caret blink, auto-scroll |
| `complete` | Stream finished normally | Border returns to default, caret fades out |
| `cancelled` | User cancelled the stream | "Stopped" label below content |
| `error` | Stream failed | Inline error with colored dot |

**State transitions:**

```
idle → prompting       (user opens "Ask AI" in Command Palette)
prompting → streaming  (user submits prompt)
prompting → idle       (user cancels prompt / closes palette)
streaming → complete   (llm:done event)
streaming → cancelled  (user calls CancelStream)
streaming → error      (llm:error event)
complete → idle        (new stream initiated)
cancelled → idle       (new stream initiated)
error → idle           (new stream initiated)
```

---

### New Store: Stream Store

**File:** `frontend/src/lib/stores/stream.ts`

| Store | Type | Initial Value | Description |
|-------|------|---------------|-------------|
| `streamContent` | `writable<string>` | `''` | Accumulated markdown content from the active stream |
| `streamState` | `writable<StreamState>` | `'idle'` | Current state in the stream lifecycle |
| `streamError` | `writable<StreamError \| null>` | `null` | Error from the last failed stream, or null |

**Derived stores:**

| Store | Type | Derivation | Description |
|-------|------|------------|-------------|
| `streamActive` | `derived<boolean>` | `$streamState === 'streaming'` | Convenience: is a stream currently active? |

**Functions:**

| Function | Signature | Description |
|----------|-----------|-------------|
| `startStream` | `(prompt: string) => Promise<void>` | Calls `StartStream` binding, opens stream tab, registers event listeners |
| `cancelStream` | `() => Promise<void>` | Calls `CancelStream` binding, transitions to `'cancelled'` |

**Event listener lifecycle:**

```
startStream(prompt) called
    ↓
Register listeners:
    EventsOn("llm:chunk", handler)  → append chunk.text to streamContent
    EventsOn("llm:done", handler)   → set streamState = 'complete', remove listeners
    EventsOn("llm:error", handler)  → set streamError, set streamState = 'error', remove listeners
    ↓
Set streamContent = ''
Set streamState = 'streaming'
Set streamError = null
    ↓
Call StartStream(prompt) via Wails binding
```

Listeners are removed on terminal states (`complete`, `cancelled`, `error`) to prevent leaks across multiple stream sessions.

---

### Config Extension

**File:** `internal/config/config.go`

The Config struct gains one field for model selection:

| Field | Type | JSON Key | Default | Description |
|-------|------|----------|---------|-------------|
| SelectedModel | `string` | `selectedModel` | `""` | Selected Claude model; empty means default (`claude-sonnet-5`) |

```go
type Config struct {
    // ... existing fields ...
    SelectedModel string `json:"selectedModel"`
}
```

This field is stored in `config.json` alongside other preferences. It does NOT contain any credentials.

---

### State Persistence — LLM Streaming Additions

| State | Persistence | Location | Mechanism |
|-------|-------------|----------|-----------|
| API key | Disk | OS keychain or `~/.config/ais/credentials.json` | `keystore.SetAPIKey()` / `keystore.GetAPIKey()` |
| Selected model | Disk (via Config) | `~/.config/ais/config.json` | Synced via `GetConfig`/`UpdateConfig` |
| Stream content | In-memory | Frontend store (`stream.ts`) | Lost on application close |
| Stream active state | In-memory | Frontend store (`stream.ts`) | Always starts as `false` |
| Stream tabs | In-memory | Frontend store (`tabs.ts`) | Lost on application close |

---

### Cross-Boundary Contract — LLM Streaming Additions

#### Go to Frontend (event payloads)

| Go Type | Event | JSON | TypeScript Type |
|---------|-------|------|-----------------|
| `StreamChunk` | `llm:chunk` | `{"text":"...","done":false}` | `StreamChunk` |
| `StreamChunk` | `llm:done` | `{"text":"","done":true,"totalTokens":N}` | `StreamChunk` |
| `StreamError` | `llm:error` | `{"code":"auth","message":"..."}` | `StreamError` |

#### Frontend to Go (method calls)

| TypeScript | Method | JSON | Go Type |
|------------|--------|------|---------|
| `string` | `StartStream` | `"prompt text"` | `string` |
| `string` | `SetAPIKey` | `"sk-ant-..."` | `string` |
| (none) | `CancelStream` | — | — |
| (none) | `HasAPIKey` | — | `bool` return |
| (none) | `GetAvailableModels` | — | `[]string` return |

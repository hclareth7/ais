# API Contract — ais

## Overview

ais exposes a set of Go methods to the Svelte frontend via Wails v2 bindings. The Wails framework auto-generates JavaScript stubs in `frontend/wailsjs/go/main/App.js`. All calls are asynchronous RPC — the frontend awaits a Promise that resolves with the return value or rejects with an error string.

Communication is bidirectional:

- **Frontend → Backend:** method calls via auto-generated bindings
- **Backend → Frontend:** events via `wailsRuntime.EventsEmit`, consumed with `EventsOn`

All methods are bound to a single `App` struct instance. There is no authentication layer — the frontend and backend share a trust boundary within the same OS process.

---

## Data Types

### FileNode

Recursive tree structure representing the scanned directory.

```typescript
interface FileNode {
  name: string;       // Base filename or directory name
  path: string;       // Relative path from root ("" for root node)
  isDir: boolean;     // true for directories, false for files
  children?: FileNode[]; // Present only when isDir is true and directory contains markdown
}
```

JSON serialization:

```json
{
  "name": "docs",
  "path": "",
  "isDir": true,
  "children": [
    {
      "name": "guide",
      "path": "guide",
      "isDir": true,
      "children": [
        { "name": "setup.md", "path": "guide/setup.md", "isDir": false }
      ]
    },
    { "name": "README.md", "path": "README.md", "isDir": false }
  ]
}
```

Invariants:

- `children` is omitted (not null, not empty array) when the node is a file
- Directories with no markdown descendants are pruned from the tree
- Children are sorted: directories first, then files, both case-insensitive alphabetical
- Hidden directories (prefixed with `.`) are excluded except the root itself
- Paths use the OS path separator

### Config

Application configuration. Persisted to `~/.config/ais/config.json`.

```typescript
interface Config {
  theme: string;          // "system" | "light" | "dark"
  sshKeyPaths: string[];  // Reserved for future use
  ignoreDirs: string[];   // Directories to skip during scan and watch
  lastOpenedPath: string; // Last root directory opened
  fontSize: number;       // Font size in pixels
  sidebarWidth: number;   // Sidebar width in pixels
  recentPaths: string[];  // Last 10 opened root directories
}
```

Default values:

| Field | Default |
|-------|---------|
| `theme` | `"system"` |
| `sshKeyPaths` | `[]` |
| `ignoreDirs` | `[".git", "node_modules", ".svn", "__pycache__", "vendor", ".venv"]` |
| `lastOpenedPath` | `""` |
| `fontSize` | `16` |
| `sidebarWidth` | `260` |
| `recentPaths` | `[]` |

Concurrency: all access is guarded by `sync.RWMutex` via `config.Manager`. Direct field access outside the Manager is forbidden.

---

## Methods

### GetFileTree

Returns the recursive file tree rooted at the current root path.

```
GetFileTree() → FileNode | error
```

| Aspect | Detail |
|--------|--------|
| Parameters | None |
| Returns | `FileNode` — root node of the scanned tree |
| Error conditions | Root path does not exist; root path is not a directory; permission denied on directory read |
| Side effects | None — read-only scan |
| Security | Scans only within root path; skips directories listed in `skipDirs` |

Behavior:

- Only `*.md` and `*.markdown` files appear in the tree (case-insensitive)
- Directories containing no markdown descendants are excluded
- Directories in the skip list (`.git`, `node_modules`, `.svn`, `__pycache__`, `vendor`, `.venv`) are excluded
- Individual directory read errors are silently skipped — the scan continues

---

### ReadFile

Reads and returns the content of a markdown file.

```
ReadFile(relativePath: string) → string | error
```

| Aspect | Detail |
|--------|--------|
| Parameters | `relativePath` — path relative to the root directory |
| Returns | File content as a UTF-8 string |
| Error conditions | Path resolves outside root (path traversal); file not found; path is a directory; file exceeds 10MB |
| Side effects | None — read-only |
| Security | **Path traversal protection** — resolves to absolute path and validates prefix against `rootPath + os.PathSeparator` |

Security detail:

```
absPath = filepath.Abs(filepath.Join(rootPath, relativePath))

REJECT if absPath does not start with rootPath + separator
REJECT if absPath equals rootPath (it's the directory itself)
```

This prevents directory traversal attacks via `../` sequences or symlink resolution.

File size enforcement:

```
maxFileSize = 10 * 1024 * 1024  // 10MB
REJECT if file.Size() > maxFileSize
```

---

### GetRootPath

Returns the current root directory path.

```
GetRootPath() → string
```

| Aspect | Detail |
|--------|--------|
| Parameters | None |
| Returns | Absolute path to the current root directory |
| Error conditions | None |
| Side effects | None |
| Security | Exposes local filesystem path — acceptable within same-process trust boundary |

---

### SetRootPath

Changes the root directory, restarts the file watcher, and updates configuration.

```
SetRootPath(path: string) → error
```

| Aspect | Detail |
|--------|--------|
| Parameters | `path` — absolute or relative path to a directory |
| Returns | `nil` on success |
| Error conditions | Path cannot be resolved; path does not exist; path is not a directory; watcher creation fails |
| Side effects | Stops existing watcher; starts new watcher; updates `lastOpenedPath` in config; prepends to `recentPaths` (capped at 10); persists config to disk |
| Security | Validates path exists and is a directory before accepting |

Sequence:

```
1. Resolve path to absolute
2. Validate: exists and is directory
3. Stop current watcher (if running)
4. Update rootPath
5. Update config (lastOpenedPath, recentPaths)
6. Create and start new watcher
```

The `recentPaths` list deduplicates — if the path already appears, it moves to the front. Maximum 10 entries.

---

### OpenFolder

Opens a native OS directory picker dialog and sets the selected path as root.

```
OpenFolder() → string | error
```

| Aspect | Detail |
|--------|--------|
| Parameters | None |
| Returns | Selected directory path, or empty string if cancelled |
| Error conditions | Dialog system error; all errors from `SetRootPath` |
| Side effects | All side effects of `SetRootPath`; displays native OS dialog |
| Security | User-initiated action via OS dialog — no path injection risk |

Behavior:

- If the user cancels the dialog, returns `""` with `nil` error
- If a directory is selected, calls `SetRootPath` internally and returns the path
- Dialog title: `"Open Folder"`

---

### GetConfig

Returns the full application configuration.

```
GetConfig() → Config
```

| Aspect | Detail |
|--------|--------|
| Parameters | None |
| Returns | `Config` — snapshot of current configuration |
| Error conditions | None |
| Side effects | None — acquires read lock, copies, releases |
| Security | Exposes filesystem paths (lastOpenedPath, recentPaths) — acceptable within trust boundary |

The returned Config is a copy. Modifications to the returned object do not affect the stored configuration.

---

### UpdateConfig

Replaces the entire configuration and persists to disk.

```
UpdateConfig(cfg: Config) → error
```

| Aspect | Detail |
|--------|--------|
| Parameters | `cfg` — complete Config object |
| Returns | `nil` on success |
| Error conditions | Failed to create config directory; failed to marshal JSON; failed to write file |
| Side effects | Overwrites in-memory config; writes to `~/.config/ais/config.json` |
| Security | No validation on field values — the caller is trusted (same-process frontend) |

This is a full replacement, not a merge. All fields in the provided Config overwrite the stored values. Omitted or zero-valued fields will overwrite with their zero value.

Config file path: `~/.config/ais/config.json`

File permissions: `0644`

Directory permissions: `0755`

---

### GetTheme

Returns the current theme mode.

```
GetTheme() → string
```

| Aspect | Detail |
|--------|--------|
| Parameters | None |
| Returns | `"system"`, `"light"`, or `"dark"` |
| Error conditions | None |
| Side effects | None |
| Security | None |

Convenience method — equivalent to `GetConfig().theme`.

---

### SetTheme

Updates the theme preference and persists to disk.

```
SetTheme(theme: string) → error
```

| Aspect | Detail |
|--------|--------|
| Parameters | `theme` — `"system"`, `"light"`, or `"dark"` |
| Returns | `nil` on success |
| Error conditions | Config save failure (directory creation, file write) |
| Side effects | Updates theme in config; persists config to disk |
| Security | No validation on theme value — the frontend is responsible for passing valid values |

Convenience method — updates only the `theme` field without affecting other config values.

---

## Wails Runtime Methods (Frontend Direct Calls)

In addition to the bound Go methods above, the frontend uses Wails runtime methods directly. These are not auto-generated — they are provided by the Wails runtime library at `frontend/wailsjs/runtime/runtime.js`.

### Window Control

Used by the custom titlebar for frameless window management:

| Method | Description |
|--------|-------------|
| `WindowMinimise()` | Minimizes the window |
| `WindowToggleMaximise()` | Toggles between maximized and restored |
| `Quit()` | Closes the application |

These methods are called from the custom window control buttons in the titlebar (`tb-right`). The window is frameless (`Frameless: true` in Go options), so no OS-provided window controls exist.

### Window Drag

The titlebar element uses the CSS property `--wails-draggable: drag` to enable native window dragging in frameless mode.

### Other Runtime Methods Used

| Method | Where Used | Purpose |
|--------|-----------|---------|
| `EventsOn(name, callback)` | `App.svelte` | Listen for backend events (`file:changed`) |

---

## Events

### file:changed

Emitted by the Go backend when a markdown file within the watched root directory changes.

```
Event: "file:changed"
Payload: string (relative path from root)
Direction: Backend → Frontend
```

| Aspect | Detail |
|--------|--------|
| Trigger | File write, create, remove, or rename of a `*.md` or `*.markdown` file |
| Payload | Relative path from root directory (e.g., `"docs/guide.md"`) |
| Debounce | 100ms per file path — rapid successive changes to the same file emit only once |
| Scope | Only markdown files; non-markdown file changes are ignored |

Frontend consumption:

```typescript
import { EventsOn } from '../wailsjs/runtime/runtime';

EventsOn('file:changed', async (changedPath: string) => {
  // changedPath is relative to root
  // Reload content for any open tab matching this path
});
```

Watcher behavior:

- Watches all directories recursively under the root
- Skips directories in the skip list (`.git`, `node_modules`, `.svn`, `__pycache__`, `vendor`, `.venv`)
- Skips hidden directories (prefixed with `.`) except the root itself
- Automatically watches newly created directories
- Stops emitting after `watcher.Stop()` is called, even if debounced callbacks are pending

---

## Error Handling

All methods that return errors produce Go-formatted error strings. The frontend receives these as rejected Promise values.

Error format:

```
"context: underlying error"
```

Examples:

```
"invalid path: stat /nonexistent: no such file or directory"
"path outside root: ../../etc/passwd"
"file too large: 15728640 bytes (max 10485760)"
"path is not a directory: /home/user/file.txt"
```

The frontend should handle errors gracefully — display a user-friendly message derived from the error string, never expose raw error text in the UI.

---

## Binding Generation

Wails auto-generates TypeScript-compatible JavaScript bindings at build time:

```
frontend/wailsjs/go/main/App.js     — method stubs
frontend/wailsjs/go/main/App.d.ts   — type declarations
frontend/wailsjs/runtime/runtime.js — EventsOn, EventsEmit, etc.
```

These files are generated by `wails dev` and `wails build`. They must not be edited manually. The generated types mirror the Go struct JSON tags.

---

## Method Summary

| Method | Parameters | Returns | Mutates State | Writes Disk |
|--------|-----------|---------|---------------|-------------|
| `GetFileTree` | — | `FileNode` | No | No |
| `ReadFile` | `relativePath` | `string` | No | No |
| `GetRootPath` | — | `string` | No | No |
| `SetRootPath` | `path` | — | Yes | Yes |
| `OpenFolder` | — | `string` | Yes | Yes |
| `GetConfig` | — | `Config` | No | No |
| `UpdateConfig` | `cfg` | — | Yes | Yes |
| `GetTheme` | — | `string` | No | No |
| `SetTheme` | `theme` | — | Yes | Yes |
| `StartStream` | `prompt` | — | Yes | No |
| `CancelStream` | — | — | Yes | No |
| `SetAPIKey` | `key` | — | No | Yes |
| `HasAPIKey` | — | `bool` | No | No |
| `GetAvailableModels` | — | `[]string` | No | No |

---

## LLM Streaming — API Additions

### StartStream

Begins streaming a Claude API response. Opens an HTTP connection to the Anthropic API and emits chunk events as content arrives.

```
StartStream(prompt: string) → error
```

| Aspect | Detail |
|--------|--------|
| Parameters | `prompt` — free-form text prompt for the AI model |
| Returns | `nil` on success (streaming begins asynchronously) |
| Error conditions | No API key configured; stream already active; API key resolution failure |
| Side effects | Creates HTTP connection to Anthropic API; starts goroutine; emits `llm:chunk`, `llm:done`, `llm:error` events |
| Security | API key is resolved server-side; prompt is sent over HTTPS to Anthropic API |

Behavior:

```
1. Acquire streamMu
2. Reject if streamCancel is non-nil (stream already active)
3. Resolve API key via keystore chain
4. Create LLM client if not cached
5. Create cancellable context
6. Store cancel function in streamCancel
7. Release streamMu
8. Launch goroutine:
   a. Call client.Stream(ctx, request, emitCallback)
   b. emitCallback emits "llm:chunk" events
   c. On completion: emit "llm:done"
   d. On error: emit "llm:error"
   e. Clear streamCancel
```

The method returns immediately after launching the goroutine. All streaming events are asynchronous.

The model used is determined by the `selectedModel` field in Config. If not set, defaults to `claude-sonnet-5`.

---

### CancelStream

Cancels the currently active stream by cancelling the context.

```
CancelStream() → error
```

| Aspect | Detail |
|--------|--------|
| Parameters | None |
| Returns | `nil` on success |
| Error conditions | No active stream |
| Side effects | Cancels context; closes HTTP connection; stream goroutine emits final event |
| Security | None — only affects local state |

Behavior:

```
1. Acquire streamMu
2. If streamCancel is nil, return error "no active stream"
3. Call streamCancel()
4. Set streamCancel = nil
5. Release streamMu
```

After cancellation, the stream goroutine detects `context.Canceled` and emits either `llm:error` with code `"cancelled"` or completes silently. Partial content received before cancellation remains in the frontend store.

---

### SetAPIKey

Stores the Anthropic API key in secure storage.

```
SetAPIKey(key: string) → error
```

| Aspect | Detail |
|--------|--------|
| Parameters | `key` — Anthropic API key string (e.g., `"sk-ant-..."`) |
| Returns | `nil` on success |
| Error conditions | Keychain write failure; credentials file write failure |
| Side effects | Writes to OS keychain (preferred) or `~/.config/ais/credentials.json` (`0600`); invalidates cached LLM client |
| Security | Key is stored in OS keychain when available. Fallback file has `0600` permissions. Key is NEVER stored in `config.json`. Key is NEVER returned to the frontend |

Storage priority (write):

```
1. OS keychain via go-keyring (service: "ais", user: "api-key")
2. Fallback: ~/.config/ais/credentials.json (permissions: 0600)
```

The fallback is used when the OS keychain is unavailable (e.g., headless Linux without D-Bus secret service).

After storing, the cached `llmClient` is set to `nil` so the next `StartStream` call creates a new client with the updated key.

---

### HasAPIKey

Checks whether an API key is configured in any storage location.

```
HasAPIKey() → bool
```

| Aspect | Detail |
|--------|--------|
| Parameters | None |
| Returns | `true` if an API key is available from any source in the resolution chain |
| Error conditions | None — returns `false` on any error |
| Side effects | None |
| Security | Never exposes the key value. Only returns a boolean |

Resolution chain checked:

```
1. ANTHROPIC_API_KEY environment variable
2. OS keychain (go-keyring)
3. ~/.config/ais/credentials.json
```

Returns `true` if any source provides a non-empty key.

---

### GetAvailableModels

Returns the list of supported Claude models.

```
GetAvailableModels() → []string
```

| Aspect | Detail |
|--------|--------|
| Parameters | None |
| Returns | Ordered list of model identifiers |
| Error conditions | None |
| Side effects | None |
| Security | None — static data |

Returns:

```json
["claude-haiku-4-5", "claude-sonnet-5", "claude-opus-5"]
```

The order reflects cost (cheapest first). The default model is `claude-sonnet-5` (second entry) — quality-first default per Design.md.

---

## LLM Streaming — Events

### llm:chunk

Emitted by the Go backend during active streaming. Contains batched text content.

```
Event: "llm:chunk"
Payload: StreamChunk (JSON)
Direction: Backend → Frontend
```

| Aspect | Detail |
|--------|--------|
| Trigger | Anthropic API SSE event received and batch timer fires (50ms intervals) |
| Payload | `{ "text": "...", "done": false }` |
| Rate | Maximum ~20 events/second (50ms batch interval) |
| Scope | Only emitted during an active stream |

The text field contains the accumulated text since the last chunk event, not the full document. The frontend appends it to the stream content store.

---

### llm:done

Emitted when the stream completes successfully.

```
Event: "llm:done"
Payload: StreamChunk (JSON)
Direction: Backend → Frontend
```

| Aspect | Detail |
|--------|--------|
| Trigger | Anthropic API stream ends normally |
| Payload | `{ "text": "", "done": true, "totalTokens": N }` |
| Rate | Once per stream |
| Scope | Signals stream completion |

After this event, no more `llm:chunk` events will fire for this stream.

---

### llm:error

Emitted when the stream encounters an error.

```
Event: "llm:error"
Payload: StreamError (JSON)
Direction: Backend → Frontend
```

| Aspect | Detail |
|--------|--------|
| Trigger | API error, network failure, or cancellation |
| Payload | `{ "code": "auth", "message": "Invalid API key" }` |
| Rate | Once per stream (terminal event) |
| Scope | Signals stream failure |

Error codes:

| Code | Meaning |
|------|---------|
| `network` | DNS failure, timeout, connection refused |
| `auth` | Invalid or expired API key (HTTP 401) |
| `rate_limit` | Too many requests (HTTP 429) |
| `cancelled` | Stream cancelled by user via `CancelStream()` |
| `api` | Other API error (HTTP 4xx/5xx) |

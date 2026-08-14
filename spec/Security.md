# Security — ais

## Threat Model

ais is a desktop application that reads and renders markdown files from the local filesystem.

It is not a server.

It is not a web application.

It does not connect to the internet.

It does not sync to the cloud.

It does not collect telemetry.

The attack surface is narrow by design:

- File content rendering (markdown to HTML in a WebView)
- Path handling (user-supplied directory and file paths)
- Configuration persistence (JSON on disk)

---

## Trust Boundaries

```
Untrusted                          Trusted
─────────────────────────────────────────────────────────
User-selected root directory  →    App.SetRootPath()
Markdown file content         →    renderer.ts → WebView DOM
File paths from directory     →    App.ReadFile() validation
OS filesystem events          →    watcher.go filtering
Config file on disk           →    config.Manager.Load()
```

The application trusts its own code and the Go runtime.

It does not trust:

- The content of any file it reads
- The structure of any directory it scans
- The paths constructed from user input

---

# Path Traversal Prevention

## The Invariant

Every file read request must resolve to a path that is a child of the current root directory.

No file outside the root directory may be read through the application interface.

---

## Implementation

### ReadFile Validation

Location: `app.go:68-76`

```go
absPath, err := filepath.Abs(filepath.Join(a.rootPath, relativePath))
if !strings.HasPrefix(absPath, a.rootPath+string(os.PathSeparator)) && absPath != a.rootPath {
    return "", fmt.Errorf("path outside root: %s", relativePath)
}
```

The defense:

1. Join the relative path with the root path
2. Resolve to an absolute path (eliminates `..`, `.`, symlink ambiguity)
3. Verify the resolved path starts with `rootPath + separator`
4. Reject if the prefix check fails

The separator suffix prevents partial directory name matches:

```
rootPath = "/home/user/docs"
absPath  = "/home/user/docs-secret/file.md"
```

Without the separator check, `docs-secret` would pass the prefix test.

---

### Directory Scanning

Location: `internal/scanner/scanner.go:33-52`

`ScanDirectory` only returns `FileNode` entries that are children of the root.

Relative paths are constructed by appending entry names to the current relative directory.

The scanner never follows symlinks — `os.ReadDir` returns directory entries, not resolved targets.

---

### Directory Filtering

Location: `internal/scanner/scanner.go:14-21`

```go
var skipDirs = map[string]bool{
    ".git":         true,
    "node_modules": true,
    ".svn":         true,
    "__pycache__":  true,
    "vendor":       true,
    ".venv":        true,
}
```

These directories are excluded from scanning.

This is a defense-in-depth measure, not a primary security control.

The primary defense is the path prefix check in `ReadFile`.

---

### Watcher Path Validation

Location: `internal/watcher/watcher.go:108-110`

```go
rel, err := filepath.Rel(w.root, name)
```

The watcher converts absolute filesystem event paths back to relative paths before emitting.

Events from outside the root directory are not possible — fsnotify only watches directories that were explicitly added via `addRecursive`, which walks from root.

---

## What Breaks If Violated

If the path traversal check is removed or weakened:

- Any frontend code (or injected script) could read arbitrary files on the system
- Sensitive files (`~/.ssh/id_rsa`, `/etc/passwd`, `~/.aws/credentials`) become accessible
- The Wails binding exposes `ReadFile` to the WebView JavaScript context

This is the most critical security invariant in the application.

---

# XSS Prevention

## The Invariant

No user-supplied content may execute as JavaScript in the WebView.

Markdown content is data, never code.

---

## Implementation

### HTML Rendering Disabled

Location: `frontend/src/lib/markdown/renderer.ts:38-39`

```typescript
const md = new MarkdownIt({
  html: false,
```

With `html: false`, markdown-it treats raw HTML tags in markdown as text.

A markdown file containing `<script>alert('xss')</script>` renders as visible text, not executable code.

---

### Code Block Escaping

Location: `frontend/src/lib/markdown/renderer.ts:42-52`

```typescript
highlight: (str: string, lang: string): string => {
    const escapedLang = md.utils.escapeHtml(lang);
    return `<pre class="code-block"><div class="code-lang">${escapedLang}</div><code class="hljs language-${escapedLang}">${result.value}</code></pre>`;
```

Two escape points:

1. The language label is escaped with `md.utils.escapeHtml()` before insertion into the DOM
2. The code content is highlighted by highlight.js, which produces safe HTML tokens
3. The fallback path escapes content directly: `md.utils.escapeHtml(str)`

A code block with language `` ```<img onerror=alert(1)> `` renders the tag as text.

---

### No Raw innerHTML From User Content

All markdown content flows through one path:

```
file content → renderMarkdown() → markdown-it → sanitized HTML → DOM
```

There is no secondary path that inserts user content directly into the DOM.

---

## What Breaks If Violated

If `html: true` is set on markdown-it:

- Any markdown file can execute JavaScript in the WebView
- The WebView has access to all Wails bindings (file read, config write, directory open)
- A malicious markdown file could read other files and exfiltrate data via the config system

If code block escaping is removed:

- A crafted language tag or code content could inject HTML/JavaScript into the rendered output

---

# Filesystem Safety

## File Size Limit

Location: `internal/scanner/scanner.go:13`

```go
const maxFileSize = 10 * 1024 * 1024 // 10MB
```

Location: `internal/scanner/scanner.go:113-115`

```go
if info.Size() > maxFileSize {
    return "", fmt.Errorf("file too large: %d bytes (max %d)", info.Size(), maxFileSize)
}
```

The size check runs before `os.ReadFile`.

This prevents memory exhaustion from:

- Accidentally opening a large binary file renamed to `.md`
- A directory containing auto-generated markdown exceeding reasonable size
- Denial of service through crafted large files

---

## Directory Skip List

Location: `internal/scanner/scanner.go:14-21` and `internal/watcher/watcher.go:13-20`

Both the scanner and watcher maintain identical skip lists.

Skipped directories:

| Directory | Reason |
|-----------|--------|
| `.git` | Contains pack files, potentially gigabytes of history |
| `node_modules` | Can contain thousands of nested directories |
| `.svn` | Version control metadata |
| `__pycache__` | Python bytecode, not useful as markdown |
| `vendor` | Vendored dependencies, potentially large |
| `.venv` | Python virtual environments |

The skip lists must remain synchronized between scanner and watcher.

If they diverge, the watcher could emit events for files the scanner never indexed.

---

## Event Debouncing

Location: `internal/watcher/watcher.go:101`

```go
debounce[event.Name] = time.AfterFunc(100*time.Millisecond, func() {
```

File change events are debounced at 100ms.

This prevents:

- CPU exhaustion from rapid file system events (e.g., `git checkout` modifying many files)
- Redundant re-reads when an editor writes multiple times during save
- UI thrashing from rapid content updates

---

# Concurrent Access Safety

## Config Manager

Location: `internal/config/config.go:21-25`

```go
type Manager struct {
    mu       sync.RWMutex
    cfg      Config
    filePath string
}
```

All config reads go through `Get()`, which holds `RLock`.

All config writes go through `Update()`, which holds full `Lock`.

The Wails runtime calls Go methods from the WebView thread.

Multiple concurrent calls to `GetConfig`, `SetTheme`, or `UpdateConfig` are safe.

Direct field access on `Manager.cfg` without holding the lock is a data race.

---

## Watcher Lifecycle

Location: `internal/watcher/watcher.go:131-139`

```go
func (w *Watcher) Stop() {
    w.once.Do(func() {
```

`sync.Once` ensures:

- The fsnotify watcher is closed exactly once
- The done channel is closed exactly once
- Double-calling `Stop()` (e.g., from `shutdown` + `SetRootPath`) does not panic

The `stopped` flag with its own `RWMutex` prevents debounced callbacks from firing after `Stop()`.

---

## What Breaks If Violated

If config is accessed without the Manager:

- Data races between the Wails WebView thread and Go goroutines
- Corrupted config state
- Non-deterministic behavior on save

If watcher Stop() lacks sync.Once:

- Panic from closing an already-closed channel
- Panic from closing an already-closed fsnotify watcher

---

# Configuration Storage

## Location

```
~/.config/ais/config.json
```

## Permissions

Location: `internal/config/config.go:64`

```go
if err := os.MkdirAll(dir, 0o755); err != nil {
```

Location: `internal/config/config.go:72`

```go
if err := os.WriteFile(m.filePath, data, 0o644); err != nil {
```

Directory: `0755` (owner: rwx, group: rx, others: rx)

File: `0644` (owner: rw, group: r, others: r)

---

## Content

The config file contains:

```json
{
  "theme": "system",
  "sshKeyPaths": [],
  "ignoreDirs": [".git", "node_modules", ...],
  "lastOpenedPath": "/home/user/docs",
  "fontSize": 16,
  "sidebarWidth": 260,
  "recentPaths": ["/home/user/docs"]
}
```

No secrets are stored.

`sshKeyPaths` contains filesystem paths to SSH keys, not key material.

`lastOpenedPath` and `recentPaths` reveal directory names but no file content.

---

## Risks

The config file is world-readable (`0644`).

On a shared system, other users can see:

- Which directories the user has opened
- Theme preferences
- SSH key file locations

This is low risk for a desktop application on a personal machine.

For shared systems, config permissions could be tightened to `0600`.

---

# Security Invariants

These invariants must never be violated. Each is load-bearing.

---

### 1. ReadFile Path Prefix Check

```
ReadFile MUST validate that the resolved absolute path starts with rootPath + os.PathSeparator
```

Location: `app.go:73`

Threat: Arbitrary file read via path traversal

---

### 2. Markdown HTML Rendering Disabled

```
markdown-it MUST be configured with html: false
```

Location: `frontend/src/lib/markdown/renderer.ts:39`

Threat: JavaScript execution via malicious markdown content

---

### 3. Code Block Content Escaping

```
Code block language tags and fallback content MUST be escaped with md.utils.escapeHtml()
```

Location: `frontend/src/lib/markdown/renderer.ts:47,50`

Threat: HTML injection via crafted code fence language tags

---

### 4. Config Access Through Manager Only

```
Config fields MUST only be read via Manager.Get() and written via Manager.Update()
```

Location: `internal/config/config.go:78-89`

Threat: Data race between WebView and Go threads

---

### 5. File Size Check Before Read

```
File size MUST be validated against maxFileSize before reading content into memory
```

Location: `internal/scanner/scanner.go:113`

Threat: Memory exhaustion from large files

---

### 6. Skip Directory Synchronization

```
Scanner skipDirs and Watcher skipDirs MUST contain identical entries
```

Location: `internal/scanner/scanner.go:14-21` and `internal/watcher/watcher.go:13-20`

Threat: Watcher emitting events for files invisible to the scanner, causing errors or inconsistent state

---

# Known Limitations

These are acknowledged gaps, not bugs.

---

### No Content Security Policy

The Wails WebView does not enforce a Content Security Policy (CSP).

If an XSS vulnerability is introduced (e.g., enabling `html: true` in markdown-it), there is no CSP to limit the damage.

Mitigation: The `html: false` setting is the primary defense. CSP would be defense-in-depth.

---

### No Theme Value Validation

Location: `app.go:153-157`

`SetTheme` accepts any string and stores it in config.

The frontend checks for `'light' | 'dark' | 'system'` at the TypeScript level, but the Go backend does not validate.

Risk: Low. An invalid theme value results in the system default being applied. No security impact.

---

### SSHKeyPaths Field Unused

Location: `internal/config/config.go:13`

The `SSHKeyPaths` field exists in the config struct but is not used by any feature.

It stores filesystem paths, not key material.

Risk: Low. The field reveals SSH key locations if the config file is read by another user.

Recommendation: Remove the field if no feature requires it, or document its intended use.

---

### Symlink Following

Neither the scanner nor the watcher explicitly checks for symbolic links.

`os.ReadDir` returns the link entry, not the target.

`filepath.Abs` in `ReadFile` resolves symlinks via the OS.

A symlink inside the root directory pointing outside could bypass the path prefix check if `filepath.Abs` resolves through the symlink.

Mitigation: Use `filepath.EvalSymlinks` before the prefix check. Not currently implemented.

Risk: Medium on systems where untrusted users can create symlinks inside the scanned directory.

---

### No File Type Validation

The scanner filters by `.md` and `.markdown` extensions only.

It does not validate that the file content is actually text.

A binary file renamed to `.md` would be read (up to 10MB) and passed to markdown-it.

Risk: Low. markdown-it treats binary content as text. No execution risk. Possible garbled rendering.

---

# Review Checklist

When reviewing changes to ais, verify:

- [ ] `html: false` remains set on markdown-it
- [ ] All user content inserted into DOM passes through escaping
- [ ] Path prefix check in `ReadFile` is intact
- [ ] Config access uses `Manager.Get()` and `Manager.Update()`
- [ ] Scanner and watcher skipDirs remain synchronized
- [ ] No new `os.ReadFile` calls bypass the size check
- [ ] No secrets are stored in config.json
- [ ] No new network connections are introduced without explicit design

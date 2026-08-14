# CLAUDE.md — ais

## Project Overview

ais is a Go desktop application that visualizes markdown files. Built with Wails v2 (Go backend) + Svelte 5 (frontend) + Tailwind CSS 4, it scans a directory for `*.md` files, renders them with syntax highlighting and collapsible sections, and watches for changes via fsnotify. The design philosophy — Ambient Intuition Design — is defined in `spec/Design.md`: the content is the protagonist, zero presence, cognitive calm.

## Orchestration

**@duit is the primary entry point for multi-domain tasks.** Do not manually chain agents — invoke `@duit` and it will select the right agents and skills automatically.

### Relevant Flows

- **Feature Development:** `planner -> steve -> software-architect -> developer -> /code-review -> /security-review`
- **UI-First Development:** `steve -> /ux-design -> /ui-to-code -> developer -> /code-review`
- **Full PR Review:** `/code-review -> security -> /security-review`
- **New Project Bootstrap:** `/project-bootstrap -> planner + steve -> software-architect -> security -> developer`

For single-domain tasks (pure code, pure architecture, pure UI), invoke the specific agent directly: `@developer`, `@software-architect`, `@steve`, etc.

## Architecture

### Backend (Go)

```
main.go              CLI arg parsing, Wails app bootstrap, embeds frontend/dist
app.go               App struct with 9 Wails-bound methods (GetFileTree, ReadFile, SetRootPath, etc.)
internal/
  config/
    config.go        Config struct + Manager with sync.RWMutex, loads/saves ~/.config/ais/config.json
    defaults.go      DefaultConfig: theme "system", fontSize 16, sidebarWidth 260
  scanner/
    scanner.go       ScanDirectory (recursive FileNode tree), ReadFileContent (10MB limit)
    scanner_test.go  Table-driven tests for scan and read
  types/
    types.go         FileNode{Name, Path, IsDir, Children}
  watcher/
    watcher.go       fsnotify watcher with 100ms debounce, recursive dir watching
```

### Frontend (Svelte 5 + Tailwind CSS 4)

```
frontend/src/
  App.svelte                    Shell layout (grid), keyboard handlers, Wails event listener
  style.css                     CSS custom properties (light/dark tokens), Tailwind import
  lib/
    components/
      Sidebar.svelte            File explorer panel with search input, theme toggle
      FileTree.svelte           Tree container, passes nodes to FileTreeNode
      FileTreeNode.svelte       Recursive node with expand/collapse, search filtering
      TabBar.svelte             Tab strip with click, middle-click close
      MarkdownViewer.svelte     Renders active tab, heading collapse, scroll restore
      ThemeToggle.svelte        Cycles system/light/dark, calls Go SetTheme
      WelcomeScreen.svelte      Shown when no tabs open
    markdown/
      renderer.ts               markdown-it (html:false), 15 highlight.js languages
    stores/
      files.ts                  fileTree/rootPath stores, loadFileTree/readFile via Wails bindings
      tabs.ts                   Tab management: open, close, next, prev, updateContent
      settings.ts               Theme store, system preference listener
```

### Spec & Design

```
spec/
  Design.md                     Product soul, design principles, visual language, interaction specs
  mockups/
    asi-ui-dark.png             Dark mode reference mockup
    asi-ui-light.png            Light mode reference mockup
    asi-ui-general-witgets.png  Component inventory: all states, panels, command palette, focus mode
```

### Wails Bindings (Go <-> Frontend)

Go methods on `App` are exposed via auto-generated `frontend/wailsjs/go/main/App.js`:

- `GetFileTree()` — returns FileNode tree
- `ReadFile(relativePath)` — returns file content (path-traversal-safe)
- `GetRootPath()` / `SetRootPath(path)` — manage root directory
- `OpenFolder()` — native directory picker dialog
- `GetConfig()` / `UpdateConfig(cfg)` — full config read/write
- `GetTheme()` / `SetTheme(mode)` — theme preference

Events: `file:changed` emitted from Go watcher, consumed in `App.svelte` via `EventsOn`.

## Build Commands

```bash
wails dev                    # Dev mode with hot reload
wails build                  # Production build -> build/bin/ais
go test ./internal/...       # Run Go tests
cd frontend && npm run build # Build frontend only
cd frontend && npm run check # Svelte type checking
```

## Code Conventions

### Go

- All domain logic lives in `internal/` packages — never in `main`
- Prefer stdlib over third-party dependencies
- Config access is concurrent-safe via `sync.RWMutex` — always use `Manager.Get()` and `Manager.Update()`
- Error handling: wrap with `fmt.Errorf("context: %w", err)`, never panic
- File paths: always resolve to absolute, validate prefix against root for path traversal

### Svelte

- **Svelte 5 runes mode** — use `$state()`, `$derived()`, `$effect()`, `$props()`. No legacy `let` reactivity
- **CSS custom properties only** — never hardcode colors. Use `var(--text-primary)`, `var(--bg-code)`, etc. All tokens defined in `style.css`
- Component props use `$props()` with TypeScript interface destructuring
- Stores use Svelte `writable`/`derived` from `svelte/store`
- Grid areas for layout: `sidebar`, `tabs`, `viewer`
- ARIA attributes on all interactive elements

### Naming

- Go: `PascalCase` exported, `camelCase` unexported
- Svelte components: `PascalCase.svelte`
- Stores/utilities: `camelCase.ts`
- CSS variables: `--kebab-case`
- Wails bindings: match Go method names exactly

## Key Files

| File | Why it matters |
|------|---------------|
| `spec/Design.md` | UI authority — all visual decisions reference this |
| `spec/mockups/` | Visual reference — dark, light, and component inventory mockups |
| `app.go` | All Wails-bound methods, the Go-frontend bridge |
| `style.css` | All CSS custom properties, theme definitions, layout tokens |
| `renderer.ts` | Markdown rendering config, syntax highlighting setup |
| `internal/scanner/scanner.go` | File tree scanning, directory filtering, file reading |
| `internal/watcher/watcher.go` | Live reload mechanism |

## Security Invariants

- **Path traversal:** `ReadFile` in `app.go` validates that resolved absolute path starts with `rootPath + separator`. Always maintain this check
- **XSS prevention:** markdown-it is configured with `html: false`. Code blocks use `md.utils.escapeHtml()`. Never enable raw HTML rendering
- **File size limit:** scanner enforces 10MB max via `maxFileSize` constant
- **Concurrent config:** Config Manager uses `sync.RWMutex`. Never access config fields directly — use `Get()` and `Update()`
- **Skip directories:** Both scanner and watcher skip `.git`, `node_modules`, `vendor`, `.svn`, `__pycache__`, `.venv`

## Testing

```bash
go test ./internal/...
```

Tests exist for `internal/scanner/`. When adding new internal packages, add tests. Patterns used:
- `t.TempDir()` for filesystem tests
- Helper functions with `t.Helper()`
- Table-driven subtests where appropriate

Frontend has no test framework yet. Type checking via `npm run check`.

## spec/Design.md Reference

**Always consult `spec/Design.md` and `spec/mockups/` before making UI decisions.** The design defines:

- The product soul: **Ambient Intuition Design** — "an adaptive reading surface for technical knowledge"
- 9 design principles: the content is the protagonist, zero presence, ambient discovery, floating knowledge, instant navigation, AI native, transparent accessibility, natural input, cognitive calm
- Theme system: exact color tokens for dark (#0B0D10) and light (#F4F5F7) modes
- Glass surfaces: backdrop-filter blur, translucent panels, rounded corners (28px default)
- Animation specs: navigation reveal 120ms, document switch 0ms, zoom instant
- Accessibility: WCAG AA baseline, semantic HTML, keyboard-first, screen reader announcements
- Typography: body 16px (floor 14px), code 13px, reading width 720-1000px
- Interaction model: hover-reveal edge zones (24px), command palette (Ctrl+K), focus mode (F11)

If a UI change conflicts with `spec/Design.md`, `spec/Design.md` wins.

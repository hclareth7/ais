# ais

A desktop markdown visualizer. Ambient Intuition Design.

Point it at a folder, read your docs.

## The Idea

```
+------------------------------------------------------------+
|  ais   < / architecture / system-design.md    — □ ✕        |
+--------+--------------------------------------+------------+
|        |                                      |  OUTLINE   |
| Library|  System Design                       |            |
|  ★ Fav |  High level architecture and key     | 1. Over... |
|  ◷ Rec |  decisions                           | 2. Prin... |
|  ☰ All |                                      | 3. Arch... |
|        |  1. Overview                          |  3.1 Front |
| SPACES |                                      |  3.2 Back  |
| [arch] |  The platform is built as a set      | 4. Secu... |
|  ops   |  of loosely coupled services...       | 5. Road... |
|  prod  |                                      |            |
|  research                        5 min read   |            |
|        |                                      |            |
| ⚙ Set  |                                      |            |
+--------+--------------------------------------+------------+
  ← hover                reader              hover →
  left edge                                  right edge
```

## Features

- Recursive file tree with search — scans for `*.md` files, skips `.git`, `node_modules`, `vendor`
- Syntax highlighting for 15 languages — Go, TypeScript, Python, Rust, Java, Bash, and more
- Collapsible heading sections — click any heading to fold its content
- Tabs — open multiple documents, switch with keyboard, middle-click to close
- Dark / Light / System theme — follows OS preference or manual toggle
- Live reload — edits to markdown files appear instantly via filesystem watcher
- Safe — path traversal protection, HTML disabled in markdown rendering, 10MB file size limit

## Quick Start

### Prerequisites

- Go 1.25+
- Node.js 18+
- Wails CLI v2: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- Linux: `gtk3-devel` and `webkit2gtk4.1-devel`

### Build and Run

```bash
git clone https://github.com/hclareth7/ais.git
cd ais

cd frontend && npm install --legacy-peer-deps && cd ..

# Development mode (hot reload)
wails dev

# Production build
wails build
./build/bin/ais
```

## Usage

```bash
ais .               # Open current directory
ais /path/to/docs   # Open a specific path
ais --version       # Print version
```

The `--mcp` flag is reserved for future MCP server mode.

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl+B` | Toggle navigation panel |
| `Ctrl+K` | Command palette |
| `Ctrl+Shift+O` | Toggle outline / table of contents |
| `Ctrl+Tab` | Next tab |
| `Ctrl+Shift+Tab` | Previous tab |
| `Ctrl+W` | Close current tab |
| `Ctrl+Shift+T` | Reopen closed tab |
| `F11` | Focus mode (document only) |
| `Escape` | Close any open panel |
| Click heading | Collapse/expand section |

## Configuration

Settings stored at `~/.config/ais/config.json`:

| Setting | Default | Description |
|---------|---------|-------------|
| `theme` | `"system"` | `"light"`, `"dark"`, or `"system"` |
| `fontSize` | `16` | Base font size in pixels |
| `sidebarWidth` | `260` | Sidebar width in pixels |
| `ignoreDirs` | `.git`, `node_modules`, ... | Directories to skip when scanning |
| `recentPaths` | `[]` | Last 10 opened folders |

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.25, Wails v2.14 |
| Frontend | Svelte 5, Tailwind CSS 4, Vite |
| Rendering | markdown-it, highlight.js |
| File watching | fsnotify |

## Development

```bash
wails dev                        # Dev mode with hot reload
go test ./internal/...           # Run Go tests
cd frontend && npm run build     # Build frontend only
cd frontend && npm run check     # Svelte type checking
wails build                      # Production build
```

After modifying Go methods on `App`, Wails regenerates bindings in `frontend/wailsjs/go/main/App.js`.

## Design

The product vision lives in `spec/Design.md` with reference mockups in `spec/mockups/`. The design philosophy is **Ambient Intuition Design** — the interface appears when needed and disappears when it stops helping.

## Roadmap

- [ ] MCP server mode — expose markdown reading as an MCP resource for LLMs
- [ ] Hover-reveal navigation — left edge reveals file tree, right edge reveals outline
- [ ] Command palette (`Ctrl+K`) — search, navigate, settings
- [ ] Outline panel — document structure via right-edge hover
- [ ] Workspaces / Spaces — group documents by context
- [ ] Focus mode (`F11`) — document only, all chrome hidden
- [ ] Glass surfaces — translucent panels with backdrop blur
- [ ] Settings overlay — appearance, font, line height
- [ ] Reading time indicator
- [ ] Quick actions — hover toolbar on text selection
- [ ] AI streaming — progressive paragraph-by-paragraph rendering
- [ ] Side-by-side comparison (`Ctrl+\`)
- [ ] Code block folding — auto-collapse blocks >25 lines
- [ ] SSH key support — clone and browse remote repositories
- [ ] Git clone integration — open repos by URL

## License

MIT

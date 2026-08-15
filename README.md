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

## Install

Download the latest binary from [Releases](https://github.com/hclareth7/ais/releases/latest):

| Platform | File |
|----------|------|
| Linux | `ais-linux-amd64.tar.gz` |
| macOS | `ais-macos-universal.zip` |
| Windows | `ais-windows-amd64.zip` |

```bash
# Linux example
tar xzf ais-linux-amd64.tar.gz
sudo mv ais /usr/local/bin/
```

See [docs/install.md](docs/install.md) for platform-specific instructions.

## Usage

```bash
cd ~/my-project
ais .               # Open current directory
ais /path/to/docs   # Open a specific path
ais --version       # Print version
```

See [docs/usage.md](docs/usage.md) for the full guide — keyboard shortcuts, hover zones, focus mode, configuration.

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
- [ ] Workspaces / Spaces — group documents by context
- [ ] Reading time indicator
- [ ] Quick actions — hover toolbar on text selection
- [ ] AI streaming — progressive paragraph-by-paragraph rendering
- [ ] Side-by-side comparison (`Ctrl+\`)
- [ ] Code block folding — auto-collapse blocks >25 lines
- [ ] Git clone integration — open repos by URL

## License

MIT

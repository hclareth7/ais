# Usage

## Open a folder

```bash
cd ~/projects/my-docs
ais .
```

ais scans the directory recursively for `*.md` files and displays them in a file tree. Directories like `.git`, `node_modules`, `vendor`, and `__pycache__` are skipped automatically.

You can also pass an explicit path:

```bash
ais /path/to/any/folder
```

Or launch without arguments to open the current directory:

```bash
ais
```

## CLI

```
ais [path]       Open a directory (default: current directory)
ais --version    Print version
ais -v           Print version
```

---

## Interface

ais uses a frameless window with hover-reveal panels. The content is always the focus — controls appear only when you need them.

### Navigation (left edge)

Move your mouse to the **left edge** of the window to reveal the file tree.

- Click a file to open it in a new tab
- The tree shows only `*.md` files
- Use the search field at the top to filter files by name

You can also toggle the navigation with `Ctrl+B`.

### Outline (right edge)

Move your mouse to the **right edge** to reveal the table of contents.

- Shows all headings from the current document
- Click a heading to scroll to it

Toggle with `Ctrl+Shift+O`.

### Tabs

Open multiple documents as tabs. Tabs appear at the top of the reader area.

- Click a tab to switch
- Middle-click a tab to close it
- `Ctrl+Tab` / `Ctrl+Shift+Tab` to cycle

### Controls (bottom edge)

Move your mouse to the **bottom edge** to reveal the control strip:

- **Zoom** — increase/decrease text size
- **Width** — adjust reading column width (600–1000px)
- **Focus mode** — hide all chrome, full immersion
- **Theme** — cycle between dark, light, and system
- **Settings** — appearance options (opacity, border radius, background)

### Focus mode

Press `F11` to enter focus mode. Everything disappears except the document. Move to the bottom edge to reveal controls. Press `Escape` or `F11` again to exit.

### Collapsible sections

Click any heading to collapse everything under it. Click again to expand. A ▶ indicator shows collapsed sections.

### Live reload

Edit a markdown file externally and ais updates the content instantly — no manual refresh needed.

---

## Keyboard shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl+B` | Toggle navigation panel |
| `Ctrl+K` | Command palette |
| `Ctrl+Shift+O` | Toggle outline |
| `Ctrl+Tab` | Next tab |
| `Ctrl+Shift+Tab` | Previous tab |
| `Ctrl+W` | Close current tab |
| `F11` | Toggle focus mode |
| `Ctrl+=` | Zoom in |
| `Ctrl+-` | Zoom out |
| `Ctrl+0` | Reset zoom |
| `Escape` | Exit focus mode / close panels |

---

## Configuration

Settings are stored at `~/.config/ais/config.json` and persist across sessions.

| Setting | Default | Description |
|---------|---------|-------------|
| `theme` | `"system"` | `"light"`, `"dark"`, or `"system"` |
| `fontSize` | `16` | Base font size in pixels |
| `sidebarWidth` | `260` | Sidebar width in pixels |
| `ignoreDirs` | `.git`, `node_modules`, ... | Directories to skip |
| `recentPaths` | `[]` | Last 10 opened folders |

---

## Supported languages (syntax highlighting)

Code blocks with language tags get syntax highlighting for: Go, TypeScript, JavaScript, Python, Rust, Java, C, C++, Bash, YAML, JSON, HTML, CSS, SQL, and Dockerfile.

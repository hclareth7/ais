# Installation

Download the latest release from [GitHub Releases](https://github.com/hclareth7/ais/releases/latest).

| Platform | File | Architecture |
|----------|------|-------------|
| Linux | `ais-linux-amd64.tar.gz` | x86_64 |
| macOS | `ais-macos-universal.zip` | Intel + Apple Silicon |
| Windows | `ais-windows-amd64.zip` | x86_64 |

---

## Linux

```bash
# Download and extract
tar xzf ais-linux-amd64.tar.gz

# Move to PATH
sudo mv ais /usr/local/bin/

# Verify
ais --version
```

### Desktop entry (optional)

To see ais in your application launcher:

```bash
cat > ~/.local/share/applications/ais.desktop << 'EOF'
[Desktop Entry]
Name=ais
Comment=Ambient Intuition — markdown reading surface
Exec=ais %f
Icon=ais
Type=Application
Categories=Utility;TextEditor;
StartupWMClass=ais
Terminal=false
EOF
```

### Dependencies

The binary is self-contained. On some minimal Linux installations you may need:

```bash
# Fedora
sudo dnf install gtk3 webkit2gtk4.1

# Ubuntu / Debian
sudo apt install libgtk-3-0 libwebkit2gtk-4.1-0
```

---

## macOS

```bash
# Extract
unzip ais-macos-universal.zip

# Move to Applications
mv ais.app /Applications/
```

To also use ais from the terminal:

```bash
sudo ln -s /Applications/ais.app/Contents/MacOS/ais /usr/local/bin/ais
```

On first launch, macOS may block the app. Go to **System Settings > Privacy & Security** and click **Open Anyway**.

---

## Windows

1. Extract `ais-windows-amd64.zip`
2. Move `ais.exe` to a permanent location (e.g. `C:\Program Files\ais\`)
3. Add that folder to your PATH:
   - Open **Settings > System > About > Advanced system settings**
   - Click **Environment Variables**
   - Edit **Path** under User variables, add `C:\Program Files\ais\`
4. Open a new terminal and verify:

```powershell
ais --version
```

### WebView2

ais requires Microsoft Edge WebView2 Runtime. It comes pre-installed on Windows 10 (2004+) and Windows 11. If missing, download it from [developer.microsoft.com/en-us/microsoft-edge/webview2](https://developer.microsoft.com/en-us/microsoft-edge/webview2/).

---

## Build from source

Requirements: Go 1.25+, Node.js 18+, [Wails CLI v2](https://wails.io/docs/gettingstarted/installation)

```bash
git clone https://github.com/hclareth7/ais.git
cd ais
wails build
sudo mv build/bin/ais /usr/local/bin/
```

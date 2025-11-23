### emoji-board 

GTK emoji picker for XWayland built with Wails v2 (Go + Web frontend)

## Features

- 🔍 **Fast emoji search** - Find emojis by name
- ⌨️ **Keyboard navigation** - Arrow keys, Enter, and ESC
- 🎨 **Dark theme** - Easy on the eyes
- 🚀 **Lightweight** - Small binary with web technologies
- 🖱️ **Mouse support** - Click to select emojis
- 📦 **Wayland native** - Works on KDE Plasma with Wayland

## Dependencies

Runtime dependencies:
- `kdotool` - Window management
- `ydotool` - Keyboard input simulation
- `wl-clipboard` - Clipboard operations
- `noto-fonts-emoji` - Emoji font
- `webkit2gtk` - Web view (webkit2gtk-4.1 on modern systems)

Build dependencies:
- `go` >= 1.25
- `gtk3` - GTK3 development files
- `webkit2gtk` - WebKit2GTK development files
- `gcc` or `clang` - C compiler (for CGO)
- `pkg-config` - For finding libraries

**Note:** This project requires CGO (C bindings) for webkit2gtk integration. Ensure your Go installation supports CGO.

**Arch Linux users:** All dependencies are automatically handled by the PKGBUILD. Simply run `makepkg -si` and pacman will install everything needed.

## Building

### Quick Start (Easiest)

```bash
# Using Make (recommended)
make build

# Or using the build script
./build.sh
```

### Using Wails directly

```bash
# Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Add wails to PATH
export PATH=$PATH:$(go env GOPATH)/bin

# Build
wails build
```

### Using Go directly (without Wails CLI)

```bash
# Build all Go files in the project
go build .
# Or specify output name
go build -o emoji-keyboard .
```

**Note**: Do NOT use `go build main.go` as it will only compile `main.go` and miss `app.go`, causing "undefined: NewApp" error.

### Using PKGBUILD (Arch Linux)

```bash
makepkg -si
```

**Note:** The PKGBUILD uses `go build` instead of `wails build` to avoid webkit2gtk version compatibility issues. All dependencies (kdotool, ydotool, wl-clipboard, webkit2gtk) are automatically installed by pacman.

## Installation

After building with `make build` or `go build`:

```bash
sudo install -Dm755 emoji-keyboard /usr/bin/emoji-keyboard
sudo install -Dm644 emoji-keyboard.desktop /usr/share/applications/emoji-keyboard.desktop
sudo install -Dm644 icon.png /usr/share/pixmaps/emoji-keyboard.png
```

After building with `wails build`:

```bash
sudo install -Dm755 build/bin/emoji-keyboard /usr/bin/emoji-keyboard
sudo install -Dm644 emoji-keyboard.desktop /usr/share/applications/emoji-keyboard.desktop
sudo install -Dm644 icon.png /usr/share/pixmaps/emoji-keyboard.png
```

## Usage

Run `emoji-keyboard` from your application launcher or terminal. The window that was focused before launching will receive the selected emoji.

**Keyboard shortcuts:**
- `↑ ↓ ← →` - Navigate emoji grid
- `Enter` or `Space` - Select emoji
- `ESC` - Return to search / Quit
- Type to search emojis

## Credits

Initially inspired by [emoji-picker](https://github.com/Quoteme/emoji-board)

## Technology Stack

- **Backend**: Go with Wails v2 framework
- **Frontend**: Vanilla JavaScript, HTML5, CSS3
- **UI**: Custom responsive grid layout
- **System Integration**: kdotool, ydotool, wl-clipboard

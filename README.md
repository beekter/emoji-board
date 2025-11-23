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
- `webkit2gtk` - Web view

Build dependencies:
- `go` >= 1.25
- `gtk3` - GTK3 development files
- `webkit2gtk` - WebKit2GTK development files

## Building

### Using the build script

```bash
./build.sh
```

### Using Wails directly

```bash
# Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Build
wails build
```

### Using PKGBUILD (Arch Linux)

```bash
makepkg -si
```

## Installation

After building:

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

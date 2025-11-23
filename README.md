### emoji-board 

GTK emoji picker for XWayland built with Wails v2 (Go + Web frontend)

## Features

- 🔍 **Fast emoji search** - Find emojis by name
- ⌨️ **Keyboard navigation** - Arrow keys, Enter, and ESC
- 🎨 **Dark theme** - Easy on the eyes
- 🚀 **Lightweight** - Small binary with web technologies
- 🖱️ **Mouse support** - Click to select emojis
- 📦 **Wayland native** - Works on KDE Plasma with Wayland

## Building (Arch Linux)

```bash
makepkg -si
```

Всё! PKGBUILD автоматически:
- Очищает все старые кеши и артефакты
- Устанавливает все зависимости (kdotool, ydotool, wl-clipboard, webkit2gtk, go, gtk3)
- Собирает приложение с правильными build tags
- Устанавливает в систему

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

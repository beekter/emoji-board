### emoji-board 

Fyne emoji picker for Wayland/X11 with ydotool integration

## Features

- Fyne-based UI with text emoji rendering (uses system fonts)
- Uses ydotool for typing emojis
- Fuzzy search by emoji name
- Categorized emoji display (smileys, people, animals, food, etc.)
- Fast compilation (< 2 seconds on modern hardware)

## Dependencies

Runtime:
- ydotool

Build:
- go
- libxxf86vm-dev, libxcursor-dev, libxrandr-dev, libxinerama-dev, libxi-dev, libgl1-mesa-dev (X11 libraries)

## Building

```bash
go build -o emoji-keyboard .
```

Build time: ~1-2 seconds (much faster than GTK4-based alternatives)

## Installation (Arch Linux)

```bash
makepkg -si
```

## Usage

1. Make sure ydotoold daemon is running:
   ```bash
   ydotoold &
   ```

2. Run the emoji picker:
   ```bash
   emoji-keyboard
   ```

3. Search for an emoji and click it to type into the currently focused window


### Credits

Initially inspired by [emoji-picker](https://github.com/Quoteme/emoji-board)

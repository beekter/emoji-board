### emoji-keyboard 

Emoji picker built with Wails v2 (Go + Web frontend)

## Features

- **Multi-language support**: Emoji names and keywords are loaded from CLDR (Common Locale Data Repository) based on system locales
- **Smart search**: Search works across all keywords in all detected languages
- **System locale detection**: Automatically detects available system locales at build time
- **CLDR integration**: Uses up-to-date emoji data from unicode-org/cldr repository

## Building (Arch Linux)

```bash
make install
```

The build process will:
1. Detect system locales (e.g., from `locale -a` or environment variables)
2. Download emoji annotations from CLDR for detected languages
3. Generate embedded emoji database with multi-language support
4. Build the application

## Usage

Run `emoji-keyboard` from your application launcher or terminal. The window that was focused before launching will receive the selected emoji.

You can search for emojis using keywords in any of your system languages. For example, if you have Russian locale installed, you can search for emojis using Russian keywords.

## Credits

Initially inspired by [emoji-picker](https://github.com/Quoteme/emoji-board)

Emoji data from [CLDR](https://github.com/unicode-org/cldr) (Unicode Common Locale Data Repository)

## Technology Stack

- **Backend**: Go with Wails v2 framework
- **Frontend**: Vanilla JavaScript, HTML5, CSS3
- **System Integration**: kdotool, ydotool, wl-clipboard
- **Emoji Data**: CLDR (Unicode Common Locale Data Repository)

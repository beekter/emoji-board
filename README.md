### emoji-keyboard 

Emoji picker built with Wails v2 (Go + Web frontend)

## Features

- **Multi-language support**: English emoji names and keywords are included in the repository, with additional languages loaded from CLDR at build time
- **Smart search**: Search works across all keywords in all detected languages
- **Automatic locale detection**: Detects system locales from `locale -a` at build time
- **CLDR integration**: Uses up-to-date emoji data from unicode-org/cldr repository
- **Fallback**: If additional languages fail to download, continues with English

## Building (Arch Linux)

```bash
make install
```

The build process will:
1. Load English emoji data from the repository (always available)
2. Detect system locales using `locale -a` command
3. Try to download emoji annotations from CLDR for detected languages (e.g., `kk_KZ` → Kazakh, `sah_RU` → Russian)
4. If downloads fail, continue with English only
5. Generate embedded emoji database with multi-language support
6. Build the application

## Usage

Run `emoji-keyboard` from your application launcher or terminal. The window that was focused before launching will receive the selected emoji.

You can search for emojis using keywords in any of your system languages. For example, if you have `kk_KZ` or `sah_RU` in your locales, you can search for emojis using Kazakh or Russian keywords.

## Credits

Initially inspired by [emoji-picker](https://github.com/Quoteme/emoji-board)

Emoji data from [CLDR](https://github.com/unicode-org/cldr) (Unicode Common Locale Data Repository)

## Technology Stack

- **Backend**: Go with Wails v2 framework
- **Frontend**: Vanilla JavaScript, HTML5, CSS3
- **System Integration**: kdotool, ydotool, wl-clipboard
- **Emoji Data**: CLDR (Unicode Common Locale Data Repository)

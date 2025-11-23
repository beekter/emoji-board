#!/bin/bash
# Simple build script for emoji-keyboard

set -e

echo "Building emoji-keyboard with Wails..."

# Check if wails is installed
if ! command -v wails &> /dev/null; then
    echo "Wails CLI not found. Installing..."
    go install github.com/wailsapp/wails/v2/cmd/wails@latest
    export PATH=$PATH:$(go env GOPATH)/bin
fi

# Build the application
wails build -clean

echo ""
echo "Build complete! Binary is located at: build/bin/emoji-keyboard"
echo ""
echo "To install system-wide, run:"
echo "  sudo install -Dm755 build/bin/emoji-keyboard /usr/bin/emoji-keyboard"
echo "  sudo install -Dm644 emoji-keyboard.desktop /usr/share/applications/emoji-keyboard.desktop"
echo "  sudo install -Dm644 icon.png /usr/share/pixmaps/emoji-keyboard.png"

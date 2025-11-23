pkgname=emoji-board
pkgver=1.0
pkgrel=1
pkgdesc="Wails emoji picker for Wayland (KDE/KWin)"
arch=('x86_64')
url="https://github.com/beekter/emoji-board"
license=('BSD-3-Clause')
depends=('kdotool' 'ydotool' 'wl-clipboard' 'noto-fonts-emoji' 'webkit2gtk')
makedepends=('go' 'gtk3' 'webkit2gtk')

build() {
    cd "$startdir"
    export CGO_ENABLED=1
    export GOFLAGS="-buildmode=pie -trimpath -modcacherw"
    # Install wails if not present
    if ! command -v wails &> /dev/null; then
        go install github.com/wailsapp/wails/v2/cmd/wails@latest
        # Add go bin to PATH
        export PATH="$PATH:$(go env GOPATH)/bin"
    fi
    # Build with wails
    wails build -clean
}

package() {
    cd "$startdir"
    install -Dm755 build/bin/emoji-keyboard "$pkgdir/usr/bin/emoji-keyboard"
    install -Dm644 emoji-keyboard.desktop "$pkgdir/usr/share/applications/emoji-keyboard.desktop"
    install -Dm644 icon.png "$pkgdir/usr/share/pixmaps/emoji-keyboard.png"
}




pkgname=emoji-board
pkgver=1.0
pkgrel=1
pkgdesc="Wails emoji picker for Wayland (KDE/KWin)"
arch=('x86_64')
url="https://github.com/beekter/emoji-board"
license=('BSD-3-Clause')
depends=('kdotool' 'ydotool' 'wl-clipboard' 'noto-fonts-emoji' 'webkit2gtk')
makedepends=('go' 'gtk3' 'webkit2gtk')

prepare() {
    cd "$startdir"
    # Clean all previous build artifacts and caches
    msg2 "Cleaning previous build artifacts..."
    rm -rf pkg/ vendor/ build/ emoji-keyboard emoji-board
    rm -rf frontend/wailsjs/
    # Clean Go build cache for this module
    go clean -cache -modcache -i -r 2>/dev/null || true
    msg2 "Clean complete"
}

build() {
    cd "$startdir"
    export CGO_ENABLED=1
    export GOFLAGS="-buildmode=pie -trimpath -modcacherw"
    # Build with standard Go (no Wails CLI needed)
    # Wails requires -tags desktop,production for proper compilation
    go build -tags desktop,production -o emoji-keyboard .
}

package() {
    cd "$startdir"
    install -Dm755 emoji-keyboard "$pkgdir/usr/bin/emoji-keyboard"
    install -Dm644 emoji-keyboard.desktop "$pkgdir/usr/share/applications/emoji-keyboard.desktop"
    install -Dm644 icon.png "$pkgdir/usr/share/pixmaps/emoji-keyboard.png"
}
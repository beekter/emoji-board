pkgname=emoji-board
pkgver=1.0
pkgrel=1
pkgdesc="Fyne emoji picker for Wayland/X11 with ydotool integration and Noto Color Emoji graphics"
arch=('x86_64')
url="https://github.com/beekter/emoji-board"
license=('BSD-3-Clause' 'Apache-2.0')
depends=('ydotool')
makedepends=('go')

build() {
    cd "$startdir"
    export CGO_ENABLED=1
    export GOFLAGS="-buildmode=pie -trimpath -modcacherw"
    go build -o emoji-keyboard .
}

package() {
    cd "$startdir"
    install -Dm755 emoji-keyboard "$pkgdir/usr/bin/emoji-keyboard"
    install -Dm644 emoji-keyboard.desktop "$pkgdir/usr/share/applications/emoji-keyboard.desktop"
}

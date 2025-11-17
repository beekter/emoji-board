# Maintainer: Luca Leon Happel <lucahappel99@gmx.de>
pkgname=emoji-board
pkgver=1.0
pkgrel=1
pkgdesc="A super simple, blazing fast, and lightweight emoji picker for Wayland"
arch=('x86_64')
url="https://github.com/Quoteme/emoji-board"
license=('BSD')
depends=('gmp' 'libffi' 'gtk3' 'xdotool' 'wl-clipboard' 'ydotool')
makedepends=('ghc' 'cabal-install' 'git')
source=()
sha256sums=()

build() {
    cd "$srcdir/.."
    
    # Update cabal package list
    cabal update
    
    # Build the project
    cabal build --builddir="$srcdir/dist"
}

package() {
    cd "$srcdir/.."
    
    # Find the built executable in the cabal dist directory
    local exe=$(find "$srcdir/dist" -name emoji-keyboard -type f -executable)
    
    # Install the binary
    install -Dm755 "$exe" "$pkgdir/usr/bin/emoji-board"
    
    # Install the license
    install -Dm644 README.md "$pkgdir/usr/share/doc/$pkgname/README.md"
}

# Cleanup function is handled automatically by makepkg
# It removes the src/ and pkg/ directories after building and packaging

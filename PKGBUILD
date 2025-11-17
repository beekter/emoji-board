# Maintainer: Luca Leon Happel <lucahappel99@gmx.de>
pkgname=emoji-board
pkgver=1.0
pkgrel=1
pkgdesc="A super simple, blazing fast, and lightweight emoji picker for Wayland"
arch=('x86_64')
url="https://github.com/beekter/emoji-board"
license=('BSD')
depends=('gmp' 'libffi' 'gtk3' 'xdotool' 'wl-clipboard' 'ydotool')
makedepends=('ghc' 'cabal-install')
source=("git+$url.git")
sha256sums=('SKIP')

build() {
    cd "$srcdir/$pkgname"
    
    # Update cabal package list
    cabal update
    
    # Build the project
    cabal build
}

package() {
    cd "$srcdir/$pkgname"
    
    # Find the built executable in the cabal dist directory
    local exe=$(find dist-newstyle -name emoji-keyboard -type f -executable | head -n1)
    
    # Install the binary
    install -Dm755 "$exe" "$pkgdir/usr/bin/emoji-board"
    
    # Install the README as documentation
    install -Dm644 README.md "$pkgdir/usr/share/doc/$pkgname/README.md"
}

# Note: makepkg automatically cleans up the src/ and pkg/ directories 
# after building and packaging, removing all build artifacts


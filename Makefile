.PHONY: all build clean install wails-build

# Default target
all: build

# Build with standard Go (no Wails CLI needed)
build:
	@echo "Building with Go..."
	CGO_ENABLED=1 go build -o emoji-keyboard .
	@echo "Build complete: ./emoji-keyboard"

# Build with Wails CLI (for production builds)
wails-build:
	@echo "Building with Wails..."
	@if ! command -v wails >/dev/null 2>&1; then \
		echo "Installing Wails CLI..."; \
		go install github.com/wailsapp/wails/v2/cmd/wails@latest; \
	fi
	@PATH="$$PATH:$$(go env GOPATH)/bin" wails build
	@echo "Build complete: ./build/bin/emoji-keyboard"

# Clean build artifacts
clean:
	rm -f emoji-keyboard
	rm -rf build/

# Install to system (requires sudo)
install: wails-build
	install -Dm755 build/bin/emoji-keyboard /usr/bin/emoji-keyboard
	install -Dm644 emoji-keyboard.desktop /usr/share/applications/emoji-keyboard.desktop
	install -Dm644 icon.png /usr/share/pixmaps/emoji-keyboard.png
	@echo "Installed to /usr/bin/emoji-keyboard"

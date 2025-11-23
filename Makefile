.PHONY: install

install:
	@echo "Installing dependencies and building..."
	@pacman -Qi kdotool >/dev/null 2>&1 || yay -S --needed kdotool
	@pacman -Qi ydotool >/dev/null 2>&1 || yay -S --needed ydotool
	yay -Bi .

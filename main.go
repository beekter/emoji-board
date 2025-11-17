package main

import (
	"image/color"
	"os/exec"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/enescakir/emoji"
)

// EmojiData represents an emoji with its key name
type EmojiData struct {
	Emoji string
	Key   string
}

// fixedButton - кнопка с фиксированным размером
type fixedButton struct {
	widget.Button
}

func newFixedButton(label string, tapped func()) *fixedButton {
	btn := &fixedButton{}
	btn.Text = label
	btn.OnTapped = tapped
	btn.Importance = widget.LowImportance
	btn.ExtendBaseWidget(btn)
	return btn
}

func (b *fixedButton) MinSize() fyne.Size {
	return fyne.NewSize(35, 35)
}

// Custom dark gray theme
type grayTheme struct{}

func (grayTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 40, G: 40, B: 40, A: 255}
	case theme.ColorNameButton:
		return color.NRGBA{R: 60, G: 60, B: 60, A: 255}
	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 50, G: 50, B: 50, A: 255}
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 50, G: 50, B: 50, A: 255}
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (grayTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (grayTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (grayTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

// getAllEmojis returns all available emojis from the library
func getAllEmojis() []EmojiData {
	var result []EmojiData
	for key, emojiStr := range emoji.Map() {
		result = append(result, EmojiData{
			Emoji: emojiStr,
			Key:   key,
		})
	}
	return result
}

// fuzzySearch performs simple search on emojis by key name
func fuzzySearch(query string, maxResults int) []EmojiData {
	allEmojis := getAllEmojis()

	if query == "" {
		if len(allEmojis) > maxResults {
			return allEmojis[:maxResults]
		}
		return allEmojis
	}

	query = strings.ToLower(query)
	var results []EmojiData

	for _, e := range allEmojis {
		if strings.Contains(strings.ToLower(e.Key), query) {
			results = append(results, e)
			if len(results) >= maxResults {
				break
			}
		}
	}

	return results
}

// typeEmoji copies emoji to clipboard and pastes it using xdotool
func typeEmoji(windowID, emojiStr string) error {
	// Copy emoji to clipboard using wl-copy
	cmd := exec.Command("wl-copy")
	cmd.Stdin = strings.NewReader(emojiStr)
	if err := cmd.Run(); err != nil {
		return err
	}

	// Focus target window
	if err := exec.Command("xdotool", "windowactivate", windowID).Run(); err != nil {
		return err
	}

	// Paste using Shift+Insert
	if err := exec.Command("xdotool", "key", "shift+Insert").Run(); err != nil {
		return err
	}

	return nil
}

// getActiveWindow returns the currently focused window ID
func getActiveWindow() (string, error) {
	out, err := exec.Command("xdotool", "getactivewindow").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func main() {
	// Get the currently focused window before creating GUI
	focusedWindowID, err := getActiveWindow()
	if err != nil {
		panic("Failed to get active window: " + err.Error())
	}

	// Create Fyne app with custom theme
	myApp := app.New()
	myApp.Settings().SetTheme(&grayTheme{})

	myWindow := myApp.NewWindow("Emoji Keyboard")
	myWindow.Resize(fyne.NewSize(180, 300))
	myWindow.CenterOnScreen()

	// Search entry
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search emoji...")

	// Grid for emojis
	var emojiButtons []*widget.Button
	var emojiCallbacks []func()
	var currentGrid *fyne.Container
	var selectedIndex int

	// Function to update emoji grid
	updateEmojis := func(query string) {
		results := fuzzySearch(query, 100)
		emojiButtons = make([]*widget.Button, 0)
		emojiCallbacks = make([]func(), 0)
		selectedIndex = -1

		// Create grid container (5 columns for narrow window)
		gridItems := []fyne.CanvasObject{}

		for _, e := range results {
			e := e // capture variable
			callback := func() {
				// Type emoji into focused window
				if err := typeEmoji(focusedWindowID, e.Emoji); err != nil {
					return
				}
				// Close app completely
				myApp.Quit()
			}
			emojiCallbacks = append(emojiCallbacks, callback)

			btn := newFixedButton(e.Emoji, callback)
			emojiButtons = append(emojiButtons, &btn.Button)
			gridItems = append(gridItems, btn)
		}

		// Create grid with 5 columns
		currentGrid = container.NewGridWithColumns(5, gridItems...)

		// Update scroll content
		scroll := myWindow.Content().(*fyne.Container).Objects[0].(*container.Scroll)
		scroll.Content = currentGrid
		scroll.Refresh()
	}

	// Handle Enter from search to select first emoji
	searchEntry.OnSubmitted = func(text string) {
		if len(emojiCallbacks) > 0 {
			selectedIndex = 0
			// Don't insert - just set selection
		}
	}

	// Initial grid
	scroll := container.NewVScroll(container.NewGridWithColumns(5))

	// Search handler
	searchEntry.OnChanged = func(text string) {
		updateEmojis(text)
		selectedIndex = -1 // Reset selection when search changes
	}

	// Keyboard navigation
	myWindow.Canvas().SetOnTypedKey(func(ev *fyne.KeyEvent) {
		// Escape always closes
		if ev.Name == fyne.KeyEscape {
			myApp.Quit()
			return
		}

		// Navigation only works when we have emojis
		if len(emojiCallbacks) == 0 {
			return
		}

		switch ev.Name {
		case fyne.KeyDown:
			if selectedIndex == -1 {
				selectedIndex = 0
			} else if selectedIndex+5 < len(emojiCallbacks) {
				selectedIndex += 5
			}
		case fyne.KeyUp:
			if selectedIndex == -1 {
				selectedIndex = 0
			} else if selectedIndex >= 5 {
				selectedIndex -= 5
			}
		case fyne.KeyLeft:
			if selectedIndex == -1 {
				selectedIndex = 0
			} else if selectedIndex > 0 {
				selectedIndex--
			}
		case fyne.KeyRight:
			if selectedIndex == -1 {
				selectedIndex = 0
			} else if selectedIndex < len(emojiCallbacks)-1 {
				selectedIndex++
			}
		case fyne.KeyEnter, fyne.KeyReturn:
			// Only insert if something is selected
			if selectedIndex >= 0 && selectedIndex < len(emojiCallbacks) {
				emojiCallbacks[selectedIndex]()
			}
		}
	})

	// Main layout
	content := container.NewBorder(
		searchEntry, // top
		nil,         // bottom
		nil,         // left
		nil,         // right
		scroll,      // center
	)

	myWindow.SetContent(content)

	// Focus search entry on start
	myWindow.Canvas().Focus(searchEntry)

	// Initial load
	updateEmojis("")

	myWindow.ShowAndRun()
}

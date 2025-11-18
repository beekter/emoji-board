package main

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	"image/color"
	_ "image/png" // Register PNG format
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/enescakir/emoji"
)

//go:embed emojis/*.png
var emojiAssets embed.FS

// EmojiData represents an emoji with its key name
type EmojiData struct {
	Emoji    string
	Key      string
	Filename string
}

// emojiGrid is a custom widget that displays emojis in a grid with keyboard navigation
type emojiGrid struct {
	widget.BaseWidget
	emojis           []EmojiData
	onSelected       func(string)
	onEscape         func()
	onReturnToSearch func()
	selectedIndex    int
	columns          int
	cellSize         float32
	scroll           *container.Scroll
	hasFocus         bool
}

func newEmojiGrid(columns int, onSelected func(string), onEscape func(), onReturnToSearch func()) *emojiGrid {
	g := &emojiGrid{
		columns:          columns,
		cellSize:         40, // Increased from 35 to 40 for more spacing
		selectedIndex:    -1,
		onSelected:       onSelected,
		onEscape:         onEscape,
		onReturnToSearch: onReturnToSearch,
	}
	g.ExtendBaseWidget(g)
	return g
}

func (g *emojiGrid) SetEmojis(emojis []EmojiData) {
	g.emojis = emojis
	g.selectedIndex = -1
	g.Refresh()
}

func (g *emojiGrid) MinSize() fyne.Size {
	rows := (len(g.emojis) + g.columns - 1) / g.columns
	if rows == 0 {
		rows = 1
	}
	return fyne.NewSize(g.cellSize*float32(g.columns), g.cellSize*float32(rows))
}

func (g *emojiGrid) CreateRenderer() fyne.WidgetRenderer {
	return &emojiGridRenderer{grid: g}
}

// Focusable Make widget focusable
func (g *emojiGrid) Focusable() bool {
	return true
}

func (g *emojiGrid) FocusGained() {
	// Initialize selection when gaining focus
	if g.selectedIndex == -1 && len(g.emojis) > 0 {
		g.selectedIndex = 0
	}
	g.hasFocus = true
	g.Refresh()
}

func (g *emojiGrid) FocusLost() {
	// Clear hasFocus flag when losing focus
	// This hides the selection square
	g.hasFocus = false
	g.Refresh()
}

func (g *emojiGrid) TypedRune(_ rune) {
	// Ignore typed runes
}

func (g *emojiGrid) TypedKey(key *fyne.KeyEvent) {
	if key.Name == fyne.KeyEscape {
		// First ESC returns to search, second ESC quits
		if g.onReturnToSearch != nil {
			g.onReturnToSearch()
			return
		}
		if g.onEscape != nil {
			g.onEscape()
		}
		return
	}

	if len(g.emojis) == 0 {
		return
	}

	// Initialize selection if needed
	if g.selectedIndex == -1 {
		g.selectedIndex = 0
	}

	oldIndex := g.selectedIndex

	switch key.Name {
	case fyne.KeyDown:
		if g.selectedIndex+g.columns < len(g.emojis) {
			g.selectedIndex += g.columns
		}
	case fyne.KeyUp:
		if g.selectedIndex >= g.columns {
			g.selectedIndex -= g.columns
		} else if g.onReturnToSearch != nil {
			// If in first row, return to search
			g.onReturnToSearch()
			return
		}
	case fyne.KeyLeft:
		if g.selectedIndex > 0 {
			g.selectedIndex--
		}
	case fyne.KeyRight:
		if g.selectedIndex < len(g.emojis)-1 {
			g.selectedIndex++
		}
	case fyne.KeyReturn, fyne.KeyEnter, fyne.KeySpace:
		if g.selectedIndex >= 0 && g.selectedIndex < len(g.emojis) {
			if g.onSelected != nil {
				g.onSelected(g.emojis[g.selectedIndex].Emoji)
			}
		}
		return
	}

	// Scroll to selected emoji if selection changed
	if oldIndex != g.selectedIndex && g.scroll != nil {
		g.scrollToSelected()
	}

	g.Refresh()
}

// scrollToSelected ensures the selected emoji is visible in the scroll container
func (g *emojiGrid) scrollToSelected() {
	if g.scroll == nil || g.selectedIndex < 0 || g.selectedIndex >= len(g.emojis) {
		return
	}

	row := g.selectedIndex / g.columns
	y := float32(row) * g.cellSize

	// Get scroll container size
	scrollSize := g.scroll.Size()
	contentHeight := g.MinSize().Height

	// Calculate visible range
	currentOffset := g.scroll.Offset.Y
	visibleTop := currentOffset
	visibleBottom := currentOffset + scrollSize.Height

	// Check if selected emoji is outside visible area
	emojiTop := y
	emojiBottom := y + g.cellSize

	if emojiTop < visibleTop {
		// Scroll up to show emoji
		g.scroll.Offset.Y = emojiTop
	} else if emojiBottom > visibleBottom {
		// Scroll down to show emoji
		g.scroll.Offset.Y = emojiBottom - scrollSize.Height
	}

	// Clamp offset
	if g.scroll.Offset.Y < 0 {
		g.scroll.Offset.Y = 0
	}
	maxOffset := contentHeight - scrollSize.Height
	if maxOffset > 0 && g.scroll.Offset.Y > maxOffset {
		g.scroll.Offset.Y = maxOffset
	}

	g.scroll.Refresh()
}

// Tapped Handle mouse/touch
func (g *emojiGrid) Tapped(ev *fyne.PointEvent) {
	col := int(ev.Position.X / g.cellSize)
	row := int(ev.Position.Y / g.cellSize)
	index := row*g.columns + col

	if index >= 0 && index < len(g.emojis) {
		g.selectedIndex = index
		if g.onSelected != nil {
			g.onSelected(g.emojis[index].Emoji)
		}
	}
}

// MouseIn Implement desktop.Hoverable for hover support
func (g *emojiGrid) MouseIn(_ *desktop.MouseEvent)    {}
func (g *emojiGrid) MouseOut()                        {}
func (g *emojiGrid) MouseMoved(_ *desktop.MouseEvent) {}

type emojiGridRenderer struct {
	grid   *emojiGrid
	images []*canvas.Image
	bg     *canvas.Rectangle
}

func (r *emojiGridRenderer) Layout(_ fyne.Size) {
	// Layout is handled in Refresh
}

func (r *emojiGridRenderer) MinSize() fyne.Size {
	return r.grid.MinSize()
}

func (r *emojiGridRenderer) Refresh() {
	// Recreate images for emojis
	r.images = make([]*canvas.Image, len(r.grid.emojis))

	for i, e := range r.grid.emojis {
		// Load emoji image from embedded assets
		img := loadEmojiImage(e.Filename)
		if img == nil {
			// Fallback to empty image if loading fails
			img = canvas.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 1, 1)))
		}

		img.FillMode = canvas.ImageFillContain
		img.SetMinSize(fyne.NewSize(28, 28)) // Emoji display size

		col := i % r.grid.columns
		row := i / r.grid.columns

		x := float32(col) * r.grid.cellSize
		y := float32(row) * r.grid.cellSize

		// Center emoji in cell
		img.Move(fyne.NewPos(x+6, y+6))
		img.Resize(fyne.NewSize(28, 28))
		r.images[i] = img
	}

	canvas.Refresh(r.grid)
}

func (r *emojiGridRenderer) BackgroundColor() color.Color {
	return color.Transparent
}

func (r *emojiGridRenderer) Objects() []fyne.CanvasObject {
	objects := make([]fyne.CanvasObject, 0, len(r.images)+10)

	// Add selection highlight with rounded corners only if grid has focus
	if r.grid.hasFocus && r.grid.selectedIndex >= 0 && r.grid.selectedIndex < len(r.grid.emojis) {
		col := r.grid.selectedIndex % r.grid.columns
		row := r.grid.selectedIndex / r.grid.columns

		x := float32(col) * r.grid.cellSize
		y := float32(row) * r.grid.cellSize

		highlightColor := color.NRGBA{R: 255, G: 255, B: 255, A: 100}
		cornerRadius := float32(6) // Moderate rounding

		// Main rectangles
		mainRect := &canvas.Rectangle{
			FillColor:               highlightColor,
			Aspect:                  1,
			TopRightCornerRadius:    cornerRadius,
			TopLeftCornerRadius:     cornerRadius,
			BottomRightCornerRadius: cornerRadius,
			BottomLeftCornerRadius:  cornerRadius,
		}
		mainRect.Move(fyne.NewPos(x, y))
		mainRect.Resize(fyne.NewSize(r.grid.cellSize, r.grid.cellSize))
		objects = append(objects, mainRect)
	}

	for _, img := range r.images {
		objects = append(objects, img)
	}

	return objects
}

func (r *emojiGridRenderer) Destroy() {}

// customEntry - entry that allows ESC to propagate and down arrow to move to grid
type customEntry struct {
	widget.Entry
	onEscape     func()
	onMoveToGrid func()
}

func newCustomEntry() *customEntry {
	entry := &customEntry{}
	entry.ExtendBaseWidget(entry)
	return entry
}

func (e *customEntry) TypedKey(key *fyne.KeyEvent) {
	// Allow ESC to propagate by calling the callback
	if key.Name == fyne.KeyEscape && e.onEscape != nil {
		e.onEscape()
		return
	}
	// Allow Down arrow to move to grid
	if key.Name == fyne.KeyDown && e.onMoveToGrid != nil {
		e.onMoveToGrid()
		return
	}
	e.Entry.TypedKey(key)
}

// Custom dark gray theme
type grayTheme struct{}

func (grayTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 0x33, G: 0x4d, B: 0x66, A: 255} // #334d66
	case theme.ColorNameButton:
		return color.NRGBA{R: 0x72, G: 0x9b, B: 0xa7, A: 255} // Slightly lighter
	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 0x6a, G: 0x93, B: 0x9f, A: 255}
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 0x7a, G: 0xa3, B: 0xaf, A: 255} // Lighter for focus visibility
	case theme.ColorNameForeground:
		return color.NRGBA{R: 220, G: 220, B: 220, A: 255}
	case theme.ColorNameHover:
		return color.NRGBA{R: 0x82, G: 0xab, B: 0xb7, A: 255}
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 140, G: 160, B: 170, A: 255}
	case theme.ColorNamePressed:
		return color.NRGBA{R: 0x8a, G: 0xb3, B: 0xbf, A: 255}
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 0x92, G: 0xbb, B: 0xc7, A: 255}
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 0x82, G: 0xab, B: 0xb7, A: 255}
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 100}
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

// getEmojiCategory returns a category order for sorting
// Lower numbers come first (faces and people first, then other categories)
func getEmojiCategory(emojiStr string) int {
	if len(emojiStr) == 0 {
		return 999
	}

	// Get first rune to determine category based on Unicode ranges
	firstRune := []rune(emojiStr)[0]
	codePoint := int(firstRune)

	// Emoticons & Smileys (U+1F600-U+1F64F) - faces and emotions
	if codePoint >= 0x1F600 && codePoint <= 0x1F64F {
		return 0
	}
	// People & Body (U+1F466-U+1F4FF and others)
	if (codePoint >= 0x1F466 && codePoint <= 0x1F487) ||
		(codePoint >= 0x1F574 && codePoint <= 0x1F5FF) ||
		(codePoint >= 0x1F926 && codePoint <= 0x1F937) ||
		(codePoint >= 0x1F9D0 && codePoint <= 0x1F9FF) {
		return 1
	}
	// Animals & Nature (U+1F400-U+1F43F, U+1F980-U+1F9CF)
	if (codePoint >= 0x1F400 && codePoint <= 0x1F43F) ||
		(codePoint >= 0x1F980 && codePoint <= 0x1F9CF) {
		return 2
	}
	// Food & Drink (U+1F32D-U+1F37F, U+1F950-U+1F96F)
	if (codePoint >= 0x1F32D && codePoint <= 0x1F37F) ||
		(codePoint >= 0x1F950 && codePoint <= 0x1F96F) {
		return 3
	}
	// Activities & Sports (U+1F3A0-U+1F3F0, U+1F93A-U+1F94F)
	if (codePoint >= 0x1F3A0 && codePoint <= 0x1F3F0) ||
		(codePoint >= 0x1F93A && codePoint <= 0x1F94F) {
		return 4
	}
	// Travel & Places (U+1F680-U+1F6FF, U+1F6E0-U+1F6EC)
	if codePoint >= 0x1F680 && codePoint <= 0x1F6FF {
		return 5
	}
	// Objects (U+1F4A0-U+1F4FF, U+1F50A-U+1F53D, U+1F56F-U+1F570)
	if (codePoint >= 0x1F4A0 && codePoint <= 0x1F4FF) ||
		(codePoint >= 0x1F50A && codePoint <= 0x1F53D) ||
		(codePoint >= 0x1F56F && codePoint <= 0x1F570) {
		return 6
	}
	// Symbols (U+1F300-U+1F320, hearts, arrows, etc.)
	if (codePoint >= 0x1F300 && codePoint <= 0x1F320) ||
		(codePoint >= 0x2600 && codePoint <= 0x26FF) ||
		(codePoint >= 0x2700 && codePoint <= 0x27BF) {
		return 7
	}
	// Flags (U+1F1E6-U+1F1FF)
	if codePoint >= 0x1F1E6 && codePoint <= 0x1F1FF {
		return 8
	}

	// Everything else
	return 10
}

// unicodeToFilename converts a Unicode emoji string to Noto Emoji filename format
// e.g. "😀" -> "emoji_u1f600.png", "👨🏻" -> "emoji_u1f468_1f3fb.png"
func unicodeToFilename(emojiStr string) string {
	runes := []rune(emojiStr)
	var hexParts []string

	for _, r := range runes {
		// Skip variation selectors
		if r == 0xFE0E || r == 0xFE0F {
			continue
		}
		hexParts = append(hexParts, fmt.Sprintf("%x", r))
	}

	return "emoji_u" + strings.Join(hexParts, "_") + ".png"
}

// loadEmojiImage loads an emoji image from embedded assets
func loadEmojiImage(filename string) *canvas.Image {
	path := "emojis/" + filename
	data, err := emojiAssets.ReadFile(path)
	if err != nil {
		return nil
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil
	}

	return canvas.NewImageFromImage(img)
}

// getAllEmojis returns all available emojis sorted by category (faces first, then others)
func getAllEmojis() []EmojiData {
	var result []EmojiData
	for key, emojiStr := range emoji.Map() {
		result = append(result, EmojiData{
			Emoji:    emojiStr,
			Key:      key,
			Filename: unicodeToFilename(emojiStr),
		})
	}

	// Sort by category first, then by Unicode code point within category
	// This ensures smileys start with 😀 (grinning face) not alphabetically
	sort.Slice(result, func(i, j int) bool {
		catI := getEmojiCategory(result[i].Emoji)
		catJ := getEmojiCategory(result[j].Emoji)

		if catI != catJ {
			return catI < catJ
		}

		// Within same category, sort by first Unicode code point (natural emoji order)
		// This makes smileys start with U+1F600 (grinning face) instead of alphabetical
		runesI := []rune(result[i].Emoji)
		runesJ := []rune(result[j].Emoji)

		if len(runesI) > 0 && len(runesJ) > 0 {
			return runesI[0] < runesJ[0]
		}

		// Fallback to key name
		return result[i].Key < result[j].Key
	})

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

	// Set window icon
	iconPath := "icon.png"
	// Try to find icon.png in the same directory as the executable
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		potentialPath := filepath.Join(exeDir, "icon.png")
		if _, err := os.Stat(potentialPath); err == nil {
			iconPath = potentialPath
		}
	}
	// Also try current directory
	if _, err := os.Stat("icon.png"); err == nil {
		iconPath = "icon.png"
	}

	if iconResource, err := fyne.LoadResourceFromPath(iconPath); err == nil {
		myWindow.SetIcon(iconResource)
	}

	// Callback for emoji selection
	onEmojiSelected := func(emojiStr string) {
		// Type emoji into focused window
		if err := typeEmoji(focusedWindowID, emojiStr); err != nil {
			return
		}
		// Close app completely
		myApp.Quit()
	}

	// Callback for escape
	onEscape := func() {
		myApp.Quit()
	}

	// Create custom entry that handles ESC and down arrow
	searchEntry := newCustomEntry()
	searchEntry.SetPlaceHolder("Search emoji...")
	searchEntry.onEscape = onEscape

	// Create emoji grid (will set onReturnToSearch and scroll later)
	grid := newEmojiGrid(5, onEmojiSelected, onEscape, nil)

	// Wrap grid in scroll container
	scroll := container.NewVScroll(grid)
	grid.scroll = scroll // Set scroll reference for auto-scrolling

	// Set up bidirectional navigation
	searchEntry.onMoveToGrid = func() {
		if len(grid.emojis) > 0 {
			grid.selectedIndex = 0
			myWindow.Canvas().Focus(grid)
			grid.Refresh()
		}
	}

	grid.onReturnToSearch = func() {
		// Scroll to top when returning to search
		if grid.scroll != nil {
			grid.scroll.Offset.Y = 0
			grid.scroll.Refresh()
		}
		myWindow.Canvas().Focus(searchEntry)
	}

	// Update function
	updateEmojis := func(query string) {
		results := fuzzySearch(query, 100)
		grid.SetEmojis(results)
	}

	// Handle Enter from search - move focus to grid
	searchEntry.OnSubmitted = func(text string) {
		if len(grid.emojis) > 0 {
			grid.selectedIndex = 0
			myWindow.Canvas().Focus(grid)
			grid.Refresh()
		}
	}

	// Search handler
	searchEntry.OnChanged = func(text string) {
		updateEmojis(text)
		// Keep focus on search while typing
		myWindow.Canvas().Focus(searchEntry)
	}

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

package main

import (
	"image/color"
	"os/exec"
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

// EmojiData represents an emoji with its key name
type EmojiData struct {
	Emoji string
	Key   string
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

// Make widget focusable
func (g *emojiGrid) Focusable() bool {
	return true
}

func (g *emojiGrid) FocusGained() {
	// Initialize selection when gaining focus
	if g.selectedIndex == -1 && len(g.emojis) > 0 {
		g.selectedIndex = 0
	}
	g.Refresh()
}

func (g *emojiGrid) FocusLost() {
	// Optionally clear selection when losing focus
	// Keeping selection visible for now
	g.Refresh()
}

func (g *emojiGrid) TypedRune(r rune) {
	// Ignore typed runes
}

func (g *emojiGrid) TypedKey(key *fyne.KeyEvent) {
	if key.Name == fyne.KeyEscape {
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

// Handle mouse/touch
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

// Implement desktop.Hoverable for hover support
func (g *emojiGrid) MouseIn(ev *desktop.MouseEvent)    {}
func (g *emojiGrid) MouseOut()                         {}
func (g *emojiGrid) MouseMoved(ev *desktop.MouseEvent) {}

type emojiGridRenderer struct {
	grid   *emojiGrid
	labels []*canvas.Text
	bg     *canvas.Rectangle
}

func (r *emojiGridRenderer) Layout(size fyne.Size) {
	// Layout is handled in Refresh
}

func (r *emojiGridRenderer) MinSize() fyne.Size {
	return r.grid.MinSize()
}

func (r *emojiGridRenderer) Refresh() {
	// Recreate labels for emojis
	r.labels = make([]*canvas.Text, len(r.grid.emojis))

	for i, e := range r.grid.emojis {
		text := canvas.NewText(e.Emoji, color.White)
		text.TextSize = 20 // Reduced from 24 to 20 for smaller emojis

		col := i % r.grid.columns
		row := i / r.grid.columns

		x := float32(col) * r.grid.cellSize
		y := float32(row) * r.grid.cellSize

		// Center emoji in cell with more spacing
		text.Move(fyne.NewPos(x+10, y+10))
		r.labels[i] = text
	}

	canvas.Refresh(r.grid)
}

func (r *emojiGridRenderer) BackgroundColor() color.Color {
	return color.Transparent
}

func (r *emojiGridRenderer) Objects() []fyne.CanvasObject {
	objects := make([]fyne.CanvasObject, 0, len(r.labels)+1)

	// Add selection highlight if something is selected
	if r.grid.selectedIndex >= 0 && r.grid.selectedIndex < len(r.grid.emojis) {
		col := r.grid.selectedIndex % r.grid.columns
		row := r.grid.selectedIndex / r.grid.columns

		x := float32(col) * r.grid.cellSize
		y := float32(row) * r.grid.cellSize

		highlight := canvas.NewRectangle(color.NRGBA{R: 100, G: 100, B: 100, A: 255}) // Lighter highlight (was 80,80,80)
		highlight.Move(fyne.NewPos(x, y))
		highlight.Resize(fyne.NewSize(r.grid.cellSize, r.grid.cellSize))
		objects = append(objects, highlight)
	}

	for _, label := range r.labels {
		objects = append(objects, label)
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
		return color.NRGBA{R: 60, G: 60, B: 60, A: 255} // Lighter background (was 40,40,40)
	case theme.ColorNameButton:
		return color.NRGBA{R: 80, G: 80, B: 80, A: 255} // Lighter buttons (was 60,60,60)
	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 70, G: 70, B: 70, A: 255} // Lighter (was 50,50,50)
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 70, G: 70, B: 70, A: 255} // Lighter (was 50,50,50)
	case theme.ColorNameForeground:
		return color.NRGBA{R: 220, G: 220, B: 220, A: 255}
	case theme.ColorNameHover:
		return color.NRGBA{R: 90, G: 90, B: 90, A: 255} // Lighter (was 70,70,70)
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 140, G: 140, B: 140, A: 255} // Lighter (was 120,120,120)
	case theme.ColorNamePressed:
		return color.NRGBA{R: 100, G: 100, B: 100, A: 255} // Lighter (was 80,80,80)
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 110, G: 110, B: 110, A: 255} // Lighter (was 90,90,90)
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 80, G: 80, B: 80, A: 255} // Lighter (was 60,60,60)
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

// getAllEmojis returns all available emojis from the library, sorted by key
func getAllEmojis() []EmojiData {
	var result []EmojiData
	for key, emojiStr := range emoji.Map() {
		result = append(result, EmojiData{
			Emoji: emojiStr,
			Key:   key,
		})
	}
	// Sort emojis by their key name for consistent ordering
	sort.Slice(result, func(i, j int) bool {
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

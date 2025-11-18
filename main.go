package main

import (
"fmt"
"image/color"
"os"
"os/exec"
"path/filepath"
"sort"
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

// getAllEmojis returns all available emojis sorted by category
func getAllEmojis() []EmojiData {
var result []EmojiData
for key, emojiStr := range emoji.Map() {
result = append(result, EmojiData{
Emoji: emojiStr,
Key:   key,
})
}

// Sort by category first, then by Unicode code point
sort.Slice(result, func(i, j int) bool {
catI := getEmojiCategory(result[i].Emoji)
catJ := getEmojiCategory(result[j].Emoji)

if catI != catJ {
return catI < catJ
}

runesI := []rune(result[i].Emoji)
runesJ := []rune(result[j].Emoji)

if len(runesI) > 0 && len(runesJ) > 0 {
return runesI[0] < runesJ[0]
}

return result[i].Key < result[j].Key
})

return result
}

// getEmojiCategory returns a category order for sorting
func getEmojiCategory(emojiStr string) int {
if len(emojiStr) == 0 {
return 999
}

firstRune := []rune(emojiStr)[0]
codePoint := int(firstRune)

if codePoint >= 0x1F600 && codePoint <= 0x1F64F {
return 0 // Smileys
}
if (codePoint >= 0x1F466 && codePoint <= 0x1F487) ||
(codePoint >= 0x1F574 && codePoint <= 0x1F5FF) ||
(codePoint >= 0x1F926 && codePoint <= 0x1F937) ||
(codePoint >= 0x1F9D0 && codePoint <= 0x1F9FF) {
return 1 // People
}
if (codePoint >= 0x1F400 && codePoint <= 0x1F43F) ||
(codePoint >= 0x1F980 && codePoint <= 0x1F9CF) {
return 2 // Animals
}
if (codePoint >= 0x1F32D && codePoint <= 0x1F37F) ||
(codePoint >= 0x1F950 && codePoint <= 0x1F96F) {
return 3 // Food
}
if (codePoint >= 0x1F3A0 && codePoint <= 0x1F3F0) ||
(codePoint >= 0x1F93A && codePoint <= 0x1F94F) {
return 4 // Activities
}
if codePoint >= 0x1F680 && codePoint <= 0x1F6FF {
return 5 // Travel
}
if (codePoint >= 0x1F4A0 && codePoint <= 0x1F4FF) ||
(codePoint >= 0x1F50A && codePoint <= 0x1F53D) ||
(codePoint >= 0x1F56F && codePoint <= 0x1F570) {
return 6 // Objects
}
if (codePoint >= 0x1F300 && codePoint <= 0x1F320) ||
(codePoint >= 0x2600 && codePoint <= 0x26FF) ||
(codePoint >= 0x2700 && codePoint <= 0x27BF) {
return 7 // Symbols
}
if codePoint >= 0x1F1E6 && codePoint <= 0x1F1FF {
return 8 // Flags
}

return 10
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

// typeEmoji types emoji using ydotool
func typeEmoji(emojiStr string) error {
cmd := exec.Command("ydotool", "type", emojiStr)
return cmd.Run()
}

// Custom theme
type grayTheme struct{}

func (grayTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
switch name {
case theme.ColorNameBackground:
return color.NRGBA{R: 0x33, G: 0x4d, B: 0x66, A: 0xff}
case theme.ColorNameButton:
return color.NRGBA{R: 0x72, G: 0x9b, B: 0xa7, A: 0xff}
case theme.ColorNameInputBackground:
return color.NRGBA{R: 0x7a, G: 0xa3, B: 0xaf, A: 0xff}
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

func main() {
// Create Fyne app with custom theme
myApp := app.New()
myApp.Settings().SetTheme(&grayTheme{})

myWindow := myApp.NewWindow("Emoji Keyboard")
myWindow.Resize(fyne.NewSize(400, 500))
myWindow.CenterOnScreen()

// Set window icon
iconPath := "icon.png"
if exePath, err := os.Executable(); err == nil {
exeDir := filepath.Dir(exePath)
potentialPath := filepath.Join(exeDir, "icon.png")
if _, err := os.Stat(potentialPath); err == nil {
iconPath = potentialPath
}
}
if _, err := os.Stat("icon.png"); err == nil {
iconPath = "icon.png"
}

if iconResource, err := fyne.LoadResourceFromPath(iconPath); err == nil {
myWindow.SetIcon(iconResource)
}

// Callback for emoji selection
onEmojiSelected := func(emojiStr string) {
if err := typeEmoji(emojiStr); err != nil {
fmt.Fprintf(os.Stderr, "Error typing emoji: %v\n", err)
}
myApp.Quit()
}

// Search entry
searchEntry := widget.NewEntry()
searchEntry.SetPlaceHolder("Search emoji...")

// Scrollable container for emojis
var scrollContainer *container.Scroll

// Update function
updateEmojis := func(query string) {
results := fuzzySearch(query, 100)

buttons := make([]fyne.CanvasObject, 0, len(results))
for _, e := range results {
emojiStr := e.Emoji
btn := widget.NewButton(emojiStr, func() {
onEmojiSelected(emojiStr)
})
buttons = append(buttons, btn)
}

emojiContainer := container.NewGridWrap(fyne.NewSize(50, 50), buttons...)
scrollContainer.Content = emojiContainer
scrollContainer.Refresh()
}

// Initial setup
scrollContainer = container.NewVScroll(container.New(nil))
updateEmojis("")

// Search handler
searchEntry.OnChanged = func(text string) {
updateEmojis(text)
}

// Main layout
content := container.NewBorder(
searchEntry,     // top
nil,            // bottom
nil,            // left
nil,            // right
scrollContainer, // center
)

myWindow.SetContent(content)
myWindow.ShowAndRun()
}

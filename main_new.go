package main

import (
"bytes"
"embed"
"fmt"
"image"
_ "image/png"
"os"
"os/exec"
"path/filepath"
"sort"
"strings"

"fyne.io/fyne/v2"
"fyne.io/fyne/v2/app"
"fyne.io/fyne/v2/canvas"
"fyne.io/fyne/v2/container"
"fyne.io/fyne/v2/widget"
"github.com/enescakir/emoji"
)

//go:embed emojis/*.png
var emojiAssets embed.FS

type EmojiData struct {
Emoji    string
Key      string
Filename string
Category string
}

type Category struct {
Name     string
Emojis   []EmojiData
Expanded bool
}

var categories = []string{
"Smileys & Emotion",
"People & Body",
"Animals & Nature",
"Food & Drink",
"Travel & Places",
"Activities",
"Objects",
"Symbols",
"Flags",
}

func unicodeToNotoFilename(emoji string) string {
runes := []rune(emoji)
parts := make([]string, len(runes))
for i, r := range runes {
parts[i] = fmt.Sprintf("%x", r)
}
return fmt.Sprintf("emoji_u%s.png", strings.Join(parts, "_"))
}

func loadEmojiImage(filename string) *canvas.Image {
data, err := emojiAssets.ReadFile("emojis/" + filename)
if err != nil {
return nil
}
img, _, err := image.Decode(bytes.NewReader(data))
if err != nil {
return nil
}
return canvas.NewImageFromImage(img)
}

func getEmojiCategory(e string) string {
runes := []rune(e)
if len(runes) == 0 {
return "Symbols"
}

code := int(runes[0])

// Smileys & Emotion
if (code >= 0x1F600 && code <= 0x1F64F) || (code >= 0x1F910 && code <= 0x1F9FF) {
return "Smileys & Emotion"
}
// People & Body  
if (code >= 0x1F466 && code <= 0x1F487) || (code >= 0x1F574 && code <= 0x1F5FF) ||
(code >= 0x1F645 && code <= 0x1F6C5) || (code >= 0x1F926 && code <= 0x1F937) ||
(code >= 0x1F9D0 && code <= 0x1F9FF) {
return "People & Body"
}
// Animals & Nature
if (code >= 0x1F400 && code <= 0x1F43F) || (code >= 0x1F980 && code <= 0x1F9AE) ||
(code >= 0x1F330 && code <= 0x1F335) || (code >= 0x1F337 && code <= 0x1F34C) {
return "Animals & Nature"
}
// Food & Drink
if (code >= 0x1F32D && code <= 0x1F37F) || (code >= 0x1F950 && code <= 0x1F96F) {
return "Food & Drink"
}
// Travel & Places
if (code >= 0x1F680 && code <= 0x1F6C5) || (code >= 0x1F3E0 && code <= 0x1F3F0) ||
(code >= 0x1F6E0 && code <= 0x1F6FF) || (code >= 0x1F3F3 && code <= 0x1F3FA) {
return "Travel & Places"
}
// Activities
if (code >= 0x1F3A0 && code <= 0x1F3DF) {
return "Activities"
}
// Objects
if (code >= 0x1F4A0 && code <= 0x1F4FF) || (code >= 0x1F507 && code <= 0x1F573) {
return "Objects"
}
// Flags
if (code >= 0x1F1E6 && code <= 0x1F1FF) {
return "Flags"
}

return "Symbols"
}

func getAllEmojis() []EmojiData {
var emojis []EmojiData

for key, e := range emoji.Map() {
filename := unicodeToNotoFilename(e)
category := getEmojiCategory(e)
emojis = append(emojis, EmojiData{
Emoji:    e,
Key:      key,
Filename: filename,
Category: category,
})
}

// Sort by category order, then by Unicode
sort.Slice(emojis, func(i, j int) bool {
catI, catJ := -1, -1
for idx, cat := range categories {
if emojis[i].Category == cat {
catI = idx
}
if emojis[j].Category == cat {
catJ = idx
}
}
if catI != catJ {
return catI < catJ
}
// Within category, sort by Unicode codepoint
return emojis[i].Emoji < emojis[j].Emoji
})

return emojis
}

func groupByCategory(emojis []EmojiData) []Category {
categoryMap := make(map[string][]EmojiData)
for _, e := range emojis {
categoryMap[e.Category] = append(categoryMap[e.Category], e)
}

var cats []Category
for _, catName := range categories {
if emojis, ok := categoryMap[catName]; ok {
cats = append(cats, Category{
Name:     catName,
Emojis:   emojis,
Expanded: false,
})
}
}

return cats
}

func insertEmoji(emoji string) {
// Get active window
cmd := exec.Command("xdotool", "getactivewindow")
out, _ := cmd.Output()
windowID := strings.TrimSpace(string(out))

// Copy to clipboard
clipCmd := exec.Command("wl-copy", emoji)
clipCmd.Run()

// Paste
exec.Command("xdotool", "windowactivate", windowID).Run()
exec.Command("xdotool", "key", "shift+Insert").Run()
}

func main() {
myApp := app.New()
myWindow := myApp.NewWindow("Emoji Keyboard")

// Load icon
if iconPath := findIcon(); iconPath != "" {
if iconRes, err := fyne.LoadResourceFromPath(iconPath); err == nil {
myWindow.SetIcon(iconRes)
}
}

allEmojis := getAllEmojis()
categorizedEmojis := groupByCategory(allEmojis)

searchEntry := widget.NewEntry()
searchEntry.SetPlaceHolder("Search emoji...")

var categoryContainers []*fyne.Container
var categoryButtons []*widget.Button
var emojiContainers []*fyne.Container

filtering := false
currentOpenCategory := 0

updateDisplay := func() {
for i := range categorizedEmojis {
if filtering {
categoryContainers[i].Show()
} else {
if categorizedEmojis[i].Expanded {
emojiContainers[i].Show()
categoryButtons[i].SetText("▼ " + categorizedEmojis[i].Name)
} else {
emojiContainers[i].Hide()
categoryButtons[i].SetText("▶ " + categorizedEmojis[i].Name)
}
}
}
}

toggleCategory := func(idx int) {
if !filtering {
// Accordion: close all others
for i := range categorizedEmojis {
categorizedEmojis[i].Expanded = (i == idx)
}
currentOpenCategory = idx
updateDisplay()
}
}

// Create UI for each category
for i, cat := range categorizedEmojis {
catIdx := i

// Category header button
btn := widget.NewButton("▶ "+cat.Name, func() {
toggleCategory(catIdx)
})
categoryButtons = append(categoryButtons, btn)

// Emoji grid for this category
emojiGrid := container.NewGridWrap(fyne.NewSize(40, 40))
for _, e := range cat.Emojis {
emojiStr := e.Emoji
emojiBtn := widget.NewButton(emojiStr, func() {
insertEmoji(emojiStr)
myWindow.Close()
})
emojiGrid.Add(emojiBtn)
}

emojiCont := container.NewVBox(emojiGrid)
emojiCont.Hide()
emojiContainers = append(emojiContainers, emojiCont)

categoryContainers = append(categoryContainers, container.NewVBox(btn, emojiCont))
}

// Open first category by default
categorizedEmojis[0].Expanded = true
updateDisplay()

// Search functionality
searchEntry.OnChanged = func(query string) {
query = strings.ToLower(strings.TrimSpace(query))

if query == "" {
filtering = false
for i := range categorizedEmojis {
categorizedEmojis[i].Expanded = (i == currentOpenCategory)
}
} else {
filtering = true
for i := range categorizedEmojis {
emojiGrid := emojiContainers[i].Objects[0].(*fyne.Container)
emojiGrid.Objects = nil

hasMatches := false
for _, e := range categorizedEmojis[i].Emojis {
if strings.Contains(strings.ToLower(e.Key), query) {
emojiStr := e.Emoji
btn := widget.NewButton(emojiStr, func() {
insertEmoji(emojiStr)
myWindow.Close()
})
emojiGrid.Add(btn)
hasMatches = true
}
}

if hasMatches {
emojiContainers[i].Show()
} else {
emojiContainers[i].Hide()
}
}
}
updateDisplay()
}

mainContent := container.NewVBox(categoryContainers...)
scrollContent := container.NewVScroll(mainContent)
scrollContent.SetMinSize(fyne.NewSize(400, 500))

content := container.NewBorder(
searchEntry,
nil, nil, nil,
scrollContent,
)

myWindow.SetContent(content)
myWindow.Resize(fyne.NewSize(450, 600))
myWindow.CenterOnScreen()
myWindow.ShowAndRun()
}

func findIcon() string {
paths := []string{
"icon.png",
filepath.Join(filepath.Dir(os.Args[0]), "icon.png"),
}
for _, p := range paths {
if _, err := os.Stat(p); err == nil {
return p
}
}
return ""
}

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
parts = make([]string, len(runes))
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

if (code >= 0x1F600 && code <= 0x1F64F) || (code >= 0x1F910 && code <= 0x1F9FF) {
return "Smileys & Emotion"
}
if (code >= 0x1F466 && code <= 0x1F487) || (code >= 0x1F574 && code <= 0x1F5FF) ||
(code >= 0x1F645 && code <= 0x1F6C5) || (code >= 0x1F926 && code <= 0x1F937) ||
(code >= 0x1F9D0 && code <= 0x1F9FF) {
return "People & Body"
}
if (code >= 0x1F400 && code <= 0x1F43F) || (code >= 0x1F980 && code <= 0x1F9AE) ||
(code >= 0x1F330 && code <= 0x1F335) || (code >= 0x1F337 && code <= 0x1F34C) {
return "Animals & Nature"
}
if (code >= 0x1F32D && code <= 0x1F37F) || (code >= 0x1F950 && code <= 0x1F96F) {
return "Food & Drink"
}
if (code >= 0x1F680 && code <= 0x1F6C5) || (code >= 0x1F3E0 && code <= 0x1F3F0) ||
(code >= 0x1F6E0 && code <= 0x1F6FF) || (code >= 0x1F3F3 && code <= 0x1F3FA) {
return "Travel & Places"
}
if (code >= 0x1F3A0 && code <= 0x1F3DF) {
return "Activities"
}
if (code >= 0x1F4A0 && code <= 0x1F4FF) || (code >= 0x1F507 && code <= 0x1F573) {
return "Objects"
}
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
return emojis[i].Emoji < emojis[j].Emoji
})

return emojis
}

func insertEmoji(emoji string) {
cmd := exec.Command("xdotool", "getactivewindow")
out, _ := cmd.Output()
windowID := strings.TrimSpace(string(out))

clipCmd := exec.Command("wl-copy", emoji)
clipCmd.Run()

exec.Command("xdotool", "windowactivate", windowID).Run()
exec.Command("xdotool", "key", "shift+Insert").Run()
}

func main() {
myApp := app.New()
myWindow := myApp.NewWindow("Emoji Keyboard")

if iconPath := findIcon(); iconPath != "" {
if iconRes, err := fyne.LoadResourceFromPath(iconPath); err == nil {
myWindow.SetIcon(iconRes)
}
}

allEmojis := getAllEmojis()

searchEntry := widget.NewEntry()
searchEntry.SetPlaceHolder("Search emoji...")

mainGrid := container.NewGridWrap(fyne.NewSize(40, 40))

for _, e := range allEmojis {
emojiStr := e.Emoji
btn := widget.NewButton(emojiStr, func() {
insertEmoji(emojiStr)
myWindow.Close()
})
mainGrid.Add(btn)
}

searchEntry.OnChanged = func(query string) {
mainGrid.Objects = nil
query = strings.ToLower(strings.TrimSpace(query))

for _, e := range allEmojis {
if query == "" || strings.Contains(strings.ToLower(e.Key), query) {
emojiStr := e.Emoji
btn := widget.NewButton(emojiStr, func() {
insertEmoji(emojiStr)
myWindow.Close()
})
mainGrid.Add(btn)
}
}
mainGrid.Refresh()
}

scrollContent := container.NewVScroll(mainGrid)
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

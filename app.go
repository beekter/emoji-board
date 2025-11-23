package main

import (
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/enescakir/emoji"
)

// Linux input event key codes for ydotool
const (
	keyLeftShift = 42  // KEY_LEFTSHIFT
	keyInsert    = 110 // KEY_INSERT
)

// App struct
type App struct {
	ctx             context.Context
	focusedWindowID string
}

// EmojiData represents an emoji with its key name
type EmojiData struct {
	Emoji string `json:"emoji"`
	Key   string `json:"key"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// Get the currently focused window before creating GUI
	focusedWindowID, err := getActiveWindow()
	if err != nil {
		// If we can't get the active window, we'll still continue
		// but typing emojis won't work
		focusedWindowID = ""
	}
	a.focusedWindowID = focusedWindowID
}

// GetAllEmojis returns all available emojis sorted by category
func (a *App) GetAllEmojis() []EmojiData {
	return getAllEmojis()
}

// SearchEmojis performs search on emojis by key name
func (a *App) SearchEmojis(query string, maxResults int) []EmojiData {
	return fuzzySearch(query, maxResults)
}

// TypeEmoji types the selected emoji into the previously focused window
func (a *App) TypeEmoji(emojiStr string) error {
	if a.focusedWindowID == "" {
		return fmt.Errorf("no window to type into")
	}
	return typeEmoji(a.focusedWindowID, emojiStr)
}

// getEmojiCategory returns a category order for sorting
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
	// People & Body
	if (codePoint >= 0x1F466 && codePoint <= 0x1F487) ||
		(codePoint >= 0x1F574 && codePoint <= 0x1F5FF) ||
		(codePoint >= 0x1F926 && codePoint <= 0x1F937) ||
		(codePoint >= 0x1F9D0 && codePoint <= 0x1F9FF) {
		return 1
	}
	// Animals & Nature
	if (codePoint >= 0x1F400 && codePoint <= 0x1F43F) ||
		(codePoint >= 0x1F980 && codePoint <= 0x1F9CF) {
		return 2
	}
	// Food & Drink
	if (codePoint >= 0x1F32D && codePoint <= 0x1F37F) ||
		(codePoint >= 0x1F950 && codePoint <= 0x1F96F) {
		return 3
	}
	// Activities & Sports
	if (codePoint >= 0x1F3A0 && codePoint <= 0x1F3F0) ||
		(codePoint >= 0x1F93A && codePoint <= 0x1F94F) {
		return 4
	}
	// Travel & Places
	if codePoint >= 0x1F680 && codePoint <= 0x1F6FF {
		return 5
	}
	// Objects
	if (codePoint >= 0x1F4A0 && codePoint <= 0x1F4FF) ||
		(codePoint >= 0x1F50A && codePoint <= 0x1F53D) ||
		(codePoint >= 0x1F56F && codePoint <= 0x1F570) {
		return 6
	}
	// Symbols
	if (codePoint >= 0x1F300 && codePoint <= 0x1F320) ||
		(codePoint >= 0x2600 && codePoint <= 0x26FF) ||
		(codePoint >= 0x2700 && codePoint <= 0x27BF) {
		return 7
	}
	// Flags
	if codePoint >= 0x1F1E6 && codePoint <= 0x1F1FF {
		return 8
	}

	// Everything else
	return 10
}

// getAllEmojis returns all available emojis sorted by category
func getAllEmojis() []EmojiData {
	// Use a map to deduplicate emojis (many keys map to same emoji)
	uniqueEmojis := make(map[string]string)
	
	for key, emojiStr := range emoji.Map() {
		// Keep the first (or shortest) key for each unique emoji
		if existingKey, exists := uniqueEmojis[emojiStr]; !exists || len(key) < len(existingKey) {
			uniqueEmojis[emojiStr] = key
		}
	}
	
	// Add missing emojis not in the library
	uniqueEmojis["\U0001F979"] = ":face_holding_back_tears:" // 🥹 face holding back tears
	
	// Convert map to slice
	var result []EmojiData
	for emojiStr, key := range uniqueEmojis {
		result = append(result, EmojiData{
			Emoji: emojiStr,
			Key:   key,
		})
	}

	// Sort by category first, then by Unicode code point within category
	sort.Slice(result, func(i, j int) bool {
		catI := getEmojiCategory(result[i].Emoji)
		catJ := getEmojiCategory(result[j].Emoji)

		if catI != catJ {
			return catI < catJ
		}

		// Within same category, sort by first Unicode code point
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

// typeEmoji copies emoji to clipboard and pastes it using kdotool + ydotool
func typeEmoji(windowID, emojiStr string) error {
	// Copy emoji to clipboard using wl-copy
	cmd := exec.Command("wl-copy")
	cmd.Stdin = strings.NewReader(emojiStr)
	if err := cmd.Run(); err != nil {
		return err
	}

	// Focus target window using kdotool
	if err := exec.Command("kdotool", "windowactivate", windowID).Run(); err != nil {
		return err
	}

	// Paste using Shift+Insert via ydotool
	if err := exec.Command("ydotool", "key",
		fmt.Sprintf("%d:1", keyLeftShift),
		fmt.Sprintf("%d:1", keyInsert),
		fmt.Sprintf("%d:0", keyInsert),
		fmt.Sprintf("%d:0", keyLeftShift),
	).Run(); err != nil {
		return err
	}

	return nil
}

// getActiveWindow returns the currently focused window ID using kdotool
func getActiveWindow() (string, error) {
	out, err := exec.Command("kdotool", "getactivewindow").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

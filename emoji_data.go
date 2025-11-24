package main

//go:generate sh -c "go run tools/cldr-fetch/main.go > emoji_data_generated.go"

import (
	"sort"
	"strings"
)

// EmojiInfo stores all information about an emoji
type EmojiInfo struct {
	Emoji    string
	Name     string   // The main name (from type="tts")
	Keywords []string // All keywords including names from all languages
}

var (
	emojiDatabase map[string]*EmojiInfo // Map of emoji -> EmojiInfo
)

// initEmojiDatabase initializes the emoji database with embedded data
func initEmojiDatabase() error {
	// The data will be populated by the generated file
	initEmbeddedEmojiData()
	return nil
}

// getAllEmojisFromDatabase returns all emojis sorted by category
func getAllEmojisFromDatabase() []EmojiData {
	var result []EmojiData

	for emoji, info := range emojiDatabase {
		displayName := info.Name
		if displayName == "" && len(info.Keywords) > 0 {
			displayName = info.Keywords[0]
		}
		if displayName == "" {
			displayName = emoji
		}

		result = append(result, EmojiData{
			Emoji: emoji,
			Key:   displayName,
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

// fuzzySearchInDatabase performs search on emojis by keywords across all languages
func fuzzySearchInDatabase(query string, maxResults int) []EmojiData {
	if query == "" {
		allEmojis := getAllEmojisFromDatabase()
		if len(allEmojis) > maxResults {
			return allEmojis[:maxResults]
		}
		return allEmojis
	}

	query = strings.ToLower(query)
	var results []EmojiData
	
	// Track which emojis we've already added to avoid duplicates
	added := make(map[string]bool)

	for emoji, info := range emojiDatabase {
		// Search in all keywords
		found := false
		for _, keyword := range info.Keywords {
			if strings.Contains(strings.ToLower(keyword), query) {
				found = true
				break
			}
		}

		if found && !added[emoji] {
			displayName := info.Name
			if displayName == "" && len(info.Keywords) > 0 {
				displayName = info.Keywords[0]
			}
			if displayName == "" {
				displayName = emoji
			}

			results = append(results, EmojiData{
				Emoji: emoji,
				Key:   displayName,
			})
			added[emoji] = true

			if len(results) >= maxResults {
				break
			}
		}
	}

	return results
}

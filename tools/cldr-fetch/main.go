package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

// EmojiAnnotation represents emoji annotation data from CLDR
type EmojiAnnotation struct {
	CP       string `xml:"cp,attr"`
	Type     string `xml:"type,attr"` // "tts" for name, empty for keywords
	Keywords string `xml:",chardata"`
}

// LDML represents the CLDR XML structure
type LDML struct {
	Annotations struct {
		Annotations []EmojiAnnotation `xml:"annotation"`
	} `xml:"annotations"`
}

// EmojiInfo stores all information about an emoji
type EmojiInfo struct {
	Emoji    string
	Name     string
	Keywords []string
}

func main() {
	// Detect system locales
	locales := detectSystemLocales()
	
	// Always include English
	languages := []string{"en"}
	
	// Add other detected languages
	for _, locale := range locales {
		langCode := strings.Split(locale, "_")[0]
		if langCode != "en" && langCode != "" && langCode != "C" && langCode != "POSIX" {
			languages = append(languages, langCode)
		}
	}
	
	// Deduplicate
	seen := make(map[string]bool)
	var uniqueLangs []string
	for _, lang := range languages {
		if !seen[lang] {
			seen[lang] = true
			uniqueLangs = append(uniqueLangs, lang)
		}
	}
	
	// Output header to stdout
	fmt.Printf("// Auto-generated file. Do not edit.\n")
	fmt.Printf("// Generated for locales: %s\n\n", strings.Join(uniqueLangs, ", "))
	fmt.Printf("package main\n\n")
	
	// Fetch and process emoji data
	emojiDatabase := make(map[string]*EmojiInfo)
	
	for _, lang := range uniqueLangs {
		if err := loadAnnotations(lang, emojiDatabase); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load %s: %v\n", lang, err)
		} else {
			fmt.Fprintf(os.Stderr, "Loaded annotations for: %s\n", lang)
		}
	}
	
	// Generate Go code to stdout
	fmt.Printf("func initEmbeddedEmojiData() {\n")
	fmt.Printf("\temojiDatabase = make(map[string]*EmojiInfo)\n\n")
	
	// Sort emojis for consistent output
	var emojis []string
	for emoji := range emojiDatabase {
		emojis = append(emojis, emoji)
	}
	sort.Strings(emojis)
	
	for _, emoji := range emojis {
		info := emojiDatabase[emoji]
		
		// Deduplicate keywords
		kwMap := make(map[string]bool)
		for _, kw := range info.Keywords {
			kwMap[kw] = true
		}
		var uniqueKws []string
		for kw := range kwMap {
			uniqueKws = append(uniqueKws, kw)
		}
		sort.Strings(uniqueKws)
		
		fmt.Printf("\temojiDatabase[%q] = &EmojiInfo{\n", emoji)
		fmt.Printf("\t\tEmoji: %q,\n", emoji)
		fmt.Printf("\t\tName: %q,\n", info.Name)
		fmt.Printf("\t\tKeywords: []string{")
		for i, kw := range uniqueKws {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%q", kw)
		}
		fmt.Printf("},\n")
		fmt.Printf("\t}\n")
	}
	
	fmt.Printf("}\n")
}

// detectSystemLocales tries to detect available system locales
func detectSystemLocales() []string {
	var locales []string
	
	// Try locale -a command
	cmd := exec.Command("locale", "-a")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && line != "C" && line != "POSIX" {
				// Remove encoding (e.g., .utf8, .UTF-8)
				locale := strings.Split(line, ".")[0]
				locales = append(locales, locale)
			}
		}
	}
	
	// Also check environment variables
	if lang := os.Getenv("LANG"); lang != "" {
		locale := strings.Split(lang, ".")[0]
		locales = append(locales, locale)
	}
	
	if language := os.Getenv("LANGUAGE"); language != "" {
		langs := strings.Split(language, ":")
		for _, lang := range langs {
			locale := strings.Split(lang, ".")[0]
			locales = append(locales, locale)
		}
	}
	
	// Deduplicate
	seen := make(map[string]bool)
	var unique []string
	for _, locale := range locales {
		if !seen[locale] && locale != "" {
			seen[locale] = true
			unique = append(unique, locale)
		}
	}
	
	return unique
}

// loadAnnotations loads emoji annotations for a specific language from CLDR
func loadAnnotations(langCode string, emojiDatabase map[string]*EmojiInfo) error {
	// Validate langCode to prevent injection attacks
	// Language codes should only contain lowercase letters, underscores, and hyphens
	for _, ch := range langCode {
		if !((ch >= 'a' && ch <= 'z') || ch == '_' || ch == '-') {
			return fmt.Errorf("invalid language code: %s", langCode)
		}
	}
	
	// URL to CLDR annotation file
	url := "https://raw.githubusercontent.com/unicode-org/cldr/main/common/annotations/" + langCode + ".xml"
	
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	// Download the file
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	
	// Parse XML with size limit to prevent resource exhaustion
	limitedReader := io.LimitReader(resp.Body, 10*1024*1024) // 10MB limit
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return err
	}
	
	var ldml LDML
	if err := xml.Unmarshal(data, &ldml); err != nil {
		return err
	}
	
	// Process annotations
	for _, ann := range ldml.Annotations.Annotations {
		emoji := ann.CP
		if emoji == "" {
			continue
		}
		
		// Initialize if not exists
		if _, exists := emojiDatabase[emoji]; !exists {
			emojiDatabase[emoji] = &EmojiInfo{
				Emoji:    emoji,
				Keywords: []string{},
			}
		}
		
		info := emojiDatabase[emoji]
		
		if ann.Type == "tts" {
			// This is the name
			if info.Name == "" || langCode == "en" {
				// Prefer English name, or use first available
				info.Name = ann.Keywords
			}
			// Also add name as a keyword
			info.Keywords = append(info.Keywords, ann.Keywords)
		} else {
			// These are keywords
			keywords := strings.Split(ann.Keywords, "|")
			for _, kw := range keywords {
				kw = strings.TrimSpace(kw)
				if kw != "" {
					info.Keywords = append(info.Keywords, kw)
				}
			}
		}
	}
	
	return nil
}

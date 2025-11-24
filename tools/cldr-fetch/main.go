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
	// Detect system locales from locale -a command
	locales := detectSystemLocalesFromCommand()
	
	// Always include English (loaded from repository)
	languages := []string{"en"}
	
	// Parse locales and extract language codes
	// For example: kk_KZ -> add "kk"
	//             sah_RU -> add "sah" and "ru" (region part)
	seen := make(map[string]bool)
	seen["en"] = true
	
	for _, locale := range locales {
		// Parse locale format: language_REGION or language
		parts := strings.Split(locale, "_")
		
		if len(parts) >= 1 {
			lang := parts[0]
			if lang != "" && lang != "C" && lang != "POSIX" && !seen[lang] {
				languages = append(languages, lang)
				seen[lang] = true
			}
		}
		
		if len(parts) >= 2 {
			// Also try to load region as language code (e.g., RU from sah_RU)
			region := strings.ToLower(parts[1])
			if region != "" && !seen[region] {
				languages = append(languages, region)
				seen[region] = true
			}
		}
	}
	
	// Output header to stdout
	fmt.Printf("// Auto-generated file. Do not edit.\n")
	fmt.Printf("// Generated for locales: %s\n\n", strings.Join(languages, ", "))
	fmt.Printf("package main\n\n")
	
	// Load and process emoji data
	emojiDatabase := make(map[string]*EmojiInfo)
	
	// Load English from repository (always available)
	if err := loadAnnotationsFromFile("cldr_data/en.xml", emojiDatabase); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load English data from repository: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Loaded annotations for: en (from repository)\n")
	
	// Try to download additional languages from CLDR
	for _, lang := range languages[1:] { // Skip "en" as it's already loaded
		if err := loadAnnotationsFromURL(lang, emojiDatabase); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load %s from CLDR: %v (continuing with available data)\n", lang, err)
		} else {
			fmt.Fprintf(os.Stderr, "Loaded annotations for: %s (from CLDR)\n", lang)
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

// detectSystemLocalesFromCommand detects locales using locale -a command only
func detectSystemLocalesFromCommand() []string {
	var locales []string
	
	// Use locale -a command to get system locales
	cmd := exec.Command("locale", "-a")
	output, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to run 'locale -a': %v\n", err)
		return locales
	}
	
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && line != "C" && line != "POSIX" {
			// Remove encoding (e.g., .utf8, .UTF-8)
			locale := strings.Split(line, ".")[0]
			locales = append(locales, locale)
		}
	}
	
	return locales
}

// loadAnnotationsFromFile loads emoji annotations from a local file
func loadAnnotationsFromFile(filePath string, emojiDatabase map[string]*EmojiInfo) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	
	return parseAnnotationsXML(data, emojiDatabase, "en")
}

// loadAnnotationsFromURL loads emoji annotations from CLDR URL
func loadAnnotationsFromURL(langCode string, emojiDatabase map[string]*EmojiInfo) error {
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
	
	return parseAnnotationsXML(data, emojiDatabase, langCode)
}

// parseAnnotationsXML parses CLDR annotation XML data
func parseAnnotationsXML(data []byte, emojiDatabase map[string]*EmojiInfo, langCode string) error {
	
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

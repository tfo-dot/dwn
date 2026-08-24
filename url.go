package main

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/disgoorg/disgo/discord"
)

// urlRegex matches http and https URLs in text
var rawURLRegex = regexp.MustCompile(`(?i)\bhttps?://[^\s<>]+`)

// ExtractURL parses and cleans the first valid HTTP/HTTPS URL found in a string.
func ExtractURL(text string) (string, bool) {
	matches := rawURLRegex.FindAllString(text, -1)
	for _, raw := range matches {
		cleaned := CleanURL(raw)
		if isValidURL(cleaned) {
			return cleaned, true
		}
	}
	return "", false
}

// CleanURL trims unwanted punctuation and wrapping characters around URLs.
func CleanURL(raw string) string {
	raw = strings.TrimSpace(raw)
	// Remove angle brackets <url> used by Discord to suppress embeds
	raw = strings.TrimPrefix(raw, "<")
	raw = strings.TrimSuffix(raw, ">")
	raw = strings.Trim(raw, `"'`)

	// Trim trailing punctuation that often attaches to URLs at the end of sentences
	for len(raw) > 0 {
		last := raw[len(raw)-1]
		if last == '.' || last == ',' || last == '!' || last == '?' || last == ';' || last == ':' || last == ')' || last == ']' {
			// Check if paren has a matching open paren in the URL (like Wikipedia URLs: /wiki/Foo_(bar))
			if (last == ')' && strings.Count(raw, "(") >= strings.Count(raw, ")")) ||
				(last == ']' && strings.Count(raw, "[") >= strings.Count(raw, "]")) {
				break
			}
			raw = raw[:len(raw)-1]
		} else {
			break
		}
	}

	return raw
}

func isValidURL(raw string) bool {
	if raw == "" {
		return false
	}
	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	if u.Host == "" {
		return false
	}
	return true
}

// ExtractURLFromMessage extracts a media URL from message content, embeds, or attachments.
func ExtractURLFromMessage(msg discord.Message) (string, bool) {
	// 1. Try content first
	if u, found := ExtractURL(msg.Content); found {
		return u, true
	}

	// 2. Try embeds
	for _, embed := range msg.Embeds {
		if embed.URL != "" && isValidURL(embed.URL) {
			return embed.URL, true
		}
		if embed.Video != nil && embed.Video.URL != "" && isValidURL(embed.Video.URL) {
			return embed.Video.URL, true
		}
	}

	// 3. Try attachments
	for _, att := range msg.Attachments {
		if att.URL != "" && isValidURL(att.URL) {
			return att.URL, true
		}
	}

	return "", false
}

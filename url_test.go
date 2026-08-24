package main

import (
	"testing"

	"github.com/disgoorg/disgo/discord"
)

func TestExtractURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		found    bool
	}{
		{
			input:    "Check this out https://youtu.be/dQw4w9WgXcQ",
			expected: "https://youtu.be/dQw4w9WgXcQ",
			found:    true,
		},
		{
			input:    "Here is a link <https://example.com/video.mp4>",
			expected: "https://example.com/video.mp4",
			found:    true,
		},
		{
			input:    "Visit https://tiktok.com/@user/video/123456789.",
			expected: "https://tiktok.com/@user/video/123456789",
			found:    true,
		},
		{
			input:    "Look: https://x.com/user/status/123! Awesome!",
			expected: "https://x.com/user/status/123",
			found:    true,
		},
		{
			input:    "No links here at all",
			expected: "",
			found:    false,
		},
		{
			input:    "Domain without slash https://x.com and https://youtube.com",
			expected: "https://x.com",
			found:    true,
		},
		{
			input:    "Wikipedia style URL: https://en.wikipedia.org/wiki/Go_(programming_language)",
			expected: "https://en.wikipedia.org/wiki/Go_(programming_language)",
			found:    true,
		},
		{
			input:    "Modern TLD https://media.download.technology/stream?id=42",
			expected: "https://media.download.technology/stream?id=42",
			found:    true,
		},
	}

	for _, tc := range tests {
		result, ok := ExtractURL(tc.input)
		if ok != tc.found {
			t.Errorf("ExtractURL(%q): got found=%v, want %v", tc.input, ok, tc.found)
		}
		if result != tc.expected {
			t.Errorf("ExtractURL(%q): got %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func TestExtractURLFromMessage(t *testing.T) {
	msgContent := discord.Message{
		Content: "https://youtu.be/12345",
	}
	url, ok := ExtractURLFromMessage(msgContent)
	if !ok || url != "https://youtu.be/12345" {
		t.Errorf("expected content URL, got %q (ok=%v)", url, ok)
	}

	msgEmbed := discord.Message{
		Content: "no url here",
		Embeds: []discord.Embed{
			{URL: "https://vimeo.com/98765"},
		},
	}
	url, ok = ExtractURLFromMessage(msgEmbed)
	if !ok || url != "https://vimeo.com/98765" {
		t.Errorf("expected embed URL, got %q (ok=%v)", url, ok)
	}
}

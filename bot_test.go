package main

import (
	"testing"

	"github.com/disgoorg/disgo/discord"
)

func TestApplicationCommands(t *testing.T) {
	if len(ApplicationCommands) != 2 {
		t.Fatalf("expected 2 application commands, got %d", len(ApplicationCommands))
	}

	hasMessageCmd := false
	hasSlashCmd := false

	for _, cmd := range ApplicationCommands {
		switch c := cmd.(type) {
		case discord.MessageCommandCreate:
			if c.CommandName() == "Download media" {
				hasMessageCmd = true
			}
		case discord.SlashCommandCreate:
			if c.CommandName() == "download" {
				hasSlashCmd = true
				if len(c.Options) == 0 || c.Options[0].OptionName() != "url" {
					t.Errorf("expected slash command option 'url', got %+v", c.Options)
				}
			}
		}
	}

	if !hasMessageCmd {
		t.Error("missing 'Download media' message command")
	}
	if !hasSlashCmd {
		t.Error("missing 'download' slash command")
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"hello", 10, "hello"},
		{"hello world", 8, "hello..."},
		{"abc", 2, "ab"},
		{"", 5, ""},
	}

	for _, tc := range tests {
		got := truncateString(tc.input, tc.maxLen)
		if got != tc.expected {
			t.Errorf("truncateString(%q, %d) = %q; want %q", tc.input, tc.maxLen, got, tc.expected)
		}
	}
}

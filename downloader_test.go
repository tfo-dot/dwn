package main

import (
	"testing"
)

func TestExtractUserFacingError(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "WARNING: [generic] unable to download video\nERROR: [youtube] 12345: Video unavailable\n",
			expected: "[youtube] 12345: Video unavailable",
		},
		{
			input:    "ERROR: First error\nERROR: Second error\n",
			expected: "First error; Second error",
		},
		{
			input:    "Just some generic output\nLast line failure",
			expected: "Last line failure",
		},
		{
			input:    "",
			expected: "",
		},
	}

	for _, tc := range tests {
		got := extractUserFacingError(tc.input)
		if got != tc.expected {
			t.Errorf("extractUserFacingError(%q) = %q; want %q", tc.input, got, tc.expected)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1024 * 1024, "1.0 MB"},
		{20 * 1024 * 1024, "20.0 MB"},
		{25 * 1024 * 1024, "25.0 MB"},
	}

	for _, tc := range tests {
		got := formatBytes(tc.bytes)
		if got != tc.expected {
			t.Errorf("formatBytes(%d) = %q; want %q", tc.bytes, got, tc.expected)
		}
	}
}

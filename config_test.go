package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	content := `{
		"DISCORD_BOT_TOKEN": "test_token_123",
		"YTDLP_PATH": "/usr/bin/yt-dlp",
		"MAX_FILE_SIZE_MB": 20,
		"MAX_CONCURRENT": 2,
		"DOWNLOAD_TIMEOUT_S": 120
	}`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := ReadConfigFile(configPath)
	if err != nil {
		t.Fatalf("unexpected error reading config: %v", err)
	}

	if cfg.DiscordBotToken != "test_token_123" {
		t.Errorf("expected token 'test_token_123', got %q", cfg.DiscordBotToken)
	}
	if cfg.YtDlpPath != "/usr/bin/yt-dlp" {
		t.Errorf("expected yt-dlp path '/usr/bin/yt-dlp', got %q", cfg.YtDlpPath)
	}
	if cfg.MaxFileSizeMB != 20 {
		t.Errorf("expected max file size 20, got %d", cfg.MaxFileSizeMB)
	}
	if cfg.MaxConcurrent != 2 {
		t.Errorf("expected max concurrent 2, got %d", cfg.MaxConcurrent)
	}
	if cfg.DownloadTimeout != 120 {
		t.Errorf("expected timeout 120, got %d", cfg.DownloadTimeout)
	}
}

func TestLoadConfig_DefaultsAndValidation(t *testing.T) {
	// Missing token should fail
	os.Unsetenv("DISCORD_BOT_TOKEN")
	_, err := LoadConfig([]string{})
	if err == nil {
		t.Error("expected error when token is missing, got nil")
	}

	// With token in env
	t.Setenv("DISCORD_BOT_TOKEN", "env_token_abc")
	t.Setenv("MAX_FILE_SIZE_MB", "15")

	// If yt-dlp is not installed / not found, it should report yt-dlp missing or find it
	cfg, err := LoadConfig([]string{})
	if err == nil {
		if cfg.DiscordBotToken != "env_token_abc" {
			t.Errorf("expected env_token_abc, got %q", cfg.DiscordBotToken)
		}
		if cfg.MaxFileSizeMB != 15 {
			t.Errorf("expected max file size 15, got %d", cfg.MaxFileSizeMB)
		}
		if cfg.MaxConcurrent != DefaultMaxConcurrent {
			t.Errorf("expected default max concurrent %d, got %d", DefaultMaxConcurrent, cfg.MaxConcurrent)
		}
	}
}

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

const (
	DefaultMaxFileSizeMB   = 20
	DefaultMaxConcurrent   = 4
	DefaultDownloadTimeout = 180 // seconds
)

type Config struct {
	DiscordBotToken string `json:"DISCORD_BOT_TOKEN"`
	YtDlpPath       string `json:"YTDLP_PATH"`
	MaxFileSizeMB   int    `json:"MAX_FILE_SIZE_MB"`
	MaxConcurrent   int    `json:"MAX_CONCURRENT"`
	DownloadTimeout int    `json:"DOWNLOAD_TIMEOUT_S"`
}

// LoadConfig loads configuration from CLI arguments, JSON config file, or environment variables.
func LoadConfig(args []string) (*Config, error) {
	fs := flag.NewFlagSet("dwn", flag.ContinueOnError)
	configPath := fs.String("config", "", "Path to JSON configuration file")
	
	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	// Also check positional argument if -config was not passed
	targetPath := *configPath
	if targetPath == "" && fs.NArg() > 0 {
		targetPath = fs.Arg(0)
	}

	cfg := &Config{
		MaxFileSizeMB:   DefaultMaxFileSizeMB,
		MaxConcurrent:   DefaultMaxConcurrent,
		DownloadTimeout: DefaultDownloadTimeout,
	}

	if targetPath != "" {
		fileCfg, err := ReadConfigFile(targetPath)
		if err != nil {
			return nil, fmt.Errorf("loading config file: %w", err)
		}
		if fileCfg.DiscordBotToken != "" {
			cfg.DiscordBotToken = fileCfg.DiscordBotToken
		}
		if fileCfg.YtDlpPath != "" {
			cfg.YtDlpPath = fileCfg.YtDlpPath
		}
		if fileCfg.MaxFileSizeMB > 0 {
			cfg.MaxFileSizeMB = fileCfg.MaxFileSizeMB
		}
		if fileCfg.MaxConcurrent > 0 {
			cfg.MaxConcurrent = fileCfg.MaxConcurrent
		}
		if fileCfg.DownloadTimeout > 0 {
			cfg.DownloadTimeout = fileCfg.DownloadTimeout
		}
	}

	// Environment variables take precedence or fill in missing values
	if token := os.Getenv("DISCORD_BOT_TOKEN"); token != "" {
		cfg.DiscordBotToken = token
	}
	if ytPath := os.Getenv("YTDLP_PATH"); ytPath != "" {
		cfg.YtDlpPath = ytPath
	}
	if maxMB := os.Getenv("MAX_FILE_SIZE_MB"); maxMB != "" {
		if val, err := strconv.Atoi(maxMB); err == nil && val > 0 {
			cfg.MaxFileSizeMB = val
		}
	}
	if maxConc := os.Getenv("MAX_CONCURRENT"); maxConc != "" {
		if val, err := strconv.Atoi(maxConc); err == nil && val > 0 {
			cfg.MaxConcurrent = val
		}
	}
	if timeout := os.Getenv("DOWNLOAD_TIMEOUT_S"); timeout != "" {
		if val, err := strconv.Atoi(timeout); err == nil && val > 0 {
			cfg.DownloadTimeout = val
		}
	}

	if cfg.DiscordBotToken == "" {
		return nil, errors.New("DISCORD_BOT_TOKEN is required (via config file or environment variable)")
	}

	// Resolve yt-dlp path if not specified
	if cfg.YtDlpPath == "" {
		resolvedPath, err := exec.LookPath("yt-dlp")
		if err != nil {
			// Fallback check for youtube-dl just in case
			resolvedPath, err = exec.LookPath("youtube-dl")
			if err != nil {
				return nil, errors.New("yt-dlp executable not found in PATH; please install it or set YTDLP_PATH")
			}
		}
		cfg.YtDlpPath = resolvedPath
	} else {
		// Verify specified path exists
		resolvedPath, err := exec.LookPath(cfg.YtDlpPath)
		if err != nil {
			return nil, fmt.Errorf("configured YTDLP_PATH %q not found or not executable: %w", cfg.YtDlpPath, err)
		}
		cfg.YtDlpPath = resolvedPath
	}

	return cfg, nil
}

// ReadConfigFile reads and parses a JSON config file.
func ReadConfigFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", path, err)
	}

	var config Config
	if err = json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing json config: %w", err)
	}

	return &config, nil
}

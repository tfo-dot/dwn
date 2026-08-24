package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrNoMediaFound    = errors.New("no downloadable media found at the provided URL")
	ErrFileTooLarge    = errors.New("downloaded media exceeds the 20 MB size limit")
	ErrDownloadTimeout = errors.New("media download timed out")
	ErrEmptyMedia      = errors.New("downloaded media file is empty")
)

type DownloadedMedia struct {
	FilePath string
	Filename string
	Size     int64
	TempDir  string
}

func (m *DownloadedMedia) Cleanup() error {
	if m.TempDir != "" {
		return os.RemoveAll(m.TempDir)
	}
	return nil
}

type Downloader struct {
	ytDlpPath      string
	maxSizeBytes   int64
	timeoutSeconds int
}

func NewDownloader(cfg *Config) *Downloader {
	maxBytes := int64(cfg.MaxFileSizeMB) * 1024 * 1024
	if maxBytes <= 0 {
		maxBytes = int64(DefaultMaxFileSizeMB) * 1024 * 1024
	}

	timeout := cfg.DownloadTimeout
	if timeout <= 0 {
		timeout = DefaultDownloadTimeout
	}

	return &Downloader{
		ytDlpPath:      cfg.YtDlpPath,
		maxSizeBytes:   maxBytes,
		timeoutSeconds: timeout,
	}
}

// Download downloads media from a URL to a temporary directory and returns file details.
func (d *Downloader) Download(parentCtx context.Context, url string) (*DownloadedMedia, error) {
	tempDir, err := os.MkdirTemp("", "dwn-media-*")
	if err != nil {
		return nil, fmt.Errorf("creating temp dir: %w", err)
	}

	// In case of error before returning, ensure tempDir is cleaned up
	cleanupOnErr := true
	defer func() {
		if cleanupOnErr {
			_ = os.RemoveAll(tempDir)
		}
	}()

	ctx, cancel := context.WithTimeout(parentCtx, time.Duration(d.timeoutSeconds)*time.Second)
	defer cancel()

	outputTemplate := filepath.Join(tempDir, "%(title).80B [%(id)s].%(ext)s")
	maxSizeArg := fmt.Sprintf("%d", d.maxSizeBytes)

	// Format selection:
	// 1. Prefer formats that fit within size limit (e.g. 720p or lower, filesize soft limit ~18M)
	// 2. Limit max-filesize hard limit
	// 3. Prevent playlist expansion
	// 4. Force windows/safe filenames to prevent path traversal or invalid Discord attachment names
	args := []string{
		"--no-playlist",
		"--max-filesize", maxSizeArg,
		"-S", "res:720,filesize~18M,fps",
		"--windows-filenames",
		"-o", outputTemplate,
		url,
	}

	var stderrBuf bytes.Buffer
	cmd := exec.CommandContext(ctx, d.ytDlpPath, args...)
	cmd.Stderr = &stderrBuf

	slog.Debug("executing yt-dlp", slog.String("path", d.ytDlpPath), slog.String("url", url))

	if err = cmd.Run(); err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, ErrDownloadTimeout
		}

		errOutput := strings.TrimSpace(stderrBuf.String())
		if strings.Contains(errOutput, "File is larger than max-filesize") ||
			strings.Contains(errOutput, "larger than maximum file size") {
			return nil, ErrFileTooLarge
		}

		cleanErr := extractUserFacingError(errOutput)
		if cleanErr != "" {
			return nil, fmt.Errorf("yt-dlp error: %s", cleanErr)
		}
		return nil, fmt.Errorf("download failed: %w", err)
	}

	// Locate downloaded file in tempDir
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		return nil, fmt.Errorf("reading temp download dir: %w", err)
	}

	if len(entries) == 0 {
		return nil, ErrNoMediaFound
	}

	// Pick the downloaded media file (ignore temporary part files if any)
	var downloadedPath string
	var downloadedName string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".part") || strings.HasSuffix(name, ".ytdl") {
			continue
		}
		downloadedPath = filepath.Join(tempDir, name)
		downloadedName = name
		break
	}

	if downloadedPath == "" {
		return nil, ErrNoMediaFound
	}

	info, err := os.Stat(downloadedPath)
	if err != nil {
		return nil, fmt.Errorf("checking file info: %w", err)
	}

	if info.Size() == 0 {
		return nil, ErrEmptyMedia
	}

	if info.Size() > d.maxSizeBytes {
		return nil, fmt.Errorf("%w (%s > %s)", ErrFileTooLarge, formatBytes(info.Size()), formatBytes(d.maxSizeBytes))
	}

	cleanupOnErr = false
	return &DownloadedMedia{
		FilePath: downloadedPath,
		Filename: downloadedName,
		Size:     info.Size(),
		TempDir:  tempDir,
	}, nil
}

// extractUserFacingError cleans yt-dlp stderr output to show concise relevant error lines.
func extractUserFacingError(stderr string) string {
	if stderr == "" {
		return ""
	}
	lines := strings.Split(stderr, "\n")
	var errLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "ERROR:") {
			errLines = append(errLines, strings.TrimSpace(strings.TrimPrefix(trimmed, "ERROR:")))
		}
	}
	if len(errLines) > 0 {
		return strings.TrimSpace(strings.Join(errLines, "; "))
	}
	// If no line starts with ERROR:, return last non-empty line
	for i := len(lines) - 1; i >= 0; i-- {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Initialize slog handler
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(handler))

	cfg, err := LoadConfig(os.Args[1:])
	if err != nil {
		slog.Error("configuration error", slog.Any("err", err))
		os.Exit(1)
	}

	slog.Info("starting dwn discord bot",
		slog.String("ytdlp_path", cfg.YtDlpPath),
		slog.Int("max_file_size_mb", cfg.MaxFileSizeMB),
		slog.Int("max_concurrent", cfg.MaxConcurrent),
		slog.Int("download_timeout_s", cfg.DownloadTimeout),
	)

	b, err := NewBot(cfg)
	if err != nil {
		slog.Error("failed to create bot instance", slog.Any("err", err))
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer stop()

	if err := b.Start(ctx); err != nil {
		slog.Error("failed to start bot", slog.Any("err", err))
		os.Exit(1)
	}

	slog.Info("Discord bot is running. Press CTRL-C to exit.")
	<-ctx.Done()

	slog.Info("shutting down bot...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	b.Close(shutdownCtx)
	slog.Info("shutdown complete")
}

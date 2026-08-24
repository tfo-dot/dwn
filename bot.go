package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

var ApplicationCommands = []discord.ApplicationCommandCreate{
	discord.MessageCommandCreate{
		Name: "Download media",
		Contexts: []discord.InteractionContextType{
			discord.InteractionContextTypeGuild,
			discord.InteractionContextTypeBotDM,
			discord.InteractionContextTypePrivateChannel,
		},
		IntegrationTypes: []discord.ApplicationIntegrationType{
			discord.ApplicationIntegrationTypeGuildInstall,
			discord.ApplicationIntegrationTypeUserInstall,
		},
	},
	discord.SlashCommandCreate{
		Name:        "download",
		Description: "Download video or audio from a URL (up to 20 MB)",
		Contexts: []discord.InteractionContextType{
			discord.InteractionContextTypeGuild,
			discord.InteractionContextTypeBotDM,
			discord.InteractionContextTypePrivateChannel,
		},
		IntegrationTypes: []discord.ApplicationIntegrationType{
			discord.ApplicationIntegrationTypeGuildInstall,
			discord.ApplicationIntegrationTypeUserInstall,
		},
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name:        "url",
				Description: "The URL of the media to download",
				Required:    true,
			},
		},
	},
}

type Bot struct {
	client     *bot.Client
	downloader *Downloader
	semaphore  chan struct{}
	cfg        *Config
}

func NewBot(cfg *Config) (*Bot, error) {
	dl := NewDownloader(cfg)
	b := &Bot{
		downloader: dl,
		semaphore:  make(chan struct{}, cfg.MaxConcurrent),
		cfg:        cfg,
	}

	client, err := disgo.New(cfg.DiscordBotToken,
		bot.WithDefaultGateway(),
		bot.WithEventListenerFunc(b.onApplicationCommand),
	)
	if err != nil {
		return nil, fmt.Errorf("building disgo client: %w", err)
	}

	b.client = client
	return b, nil
}

func (b *Bot) Start(ctx context.Context) error {
	// Register global commands
	if _, err := b.client.Rest.SetGlobalCommands(b.client.ApplicationID, ApplicationCommands); err != nil {
		slog.Error("error registering global commands", slog.Any("err", err))
	} else {
		slog.Info("registered application commands successfully")
	}

	if err := b.client.OpenGateway(ctx); err != nil {
		return fmt.Errorf("opening gateway: %w", err)
	}

	return nil
}

func (b *Bot) Close(ctx context.Context) {
	b.client.Close(ctx)
}

func (b *Bot) onApplicationCommand(event *events.ApplicationCommandInteractionCreate) {
	switch event.Data.Type() {
	case discord.ApplicationCommandTypeMessage:
		b.handleMessageCommand(event)
	case discord.ApplicationCommandTypeSlash:
		b.handleSlashCommand(event)
	default:
		slog.Debug("unhandled command type", slog.Any("type", event.Data.Type()))
	}
}

func (b *Bot) handleMessageCommand(event *events.ApplicationCommandInteractionCreate) {
	data := event.MessageCommandInteractionData()
	if data.CommandName() != "Download media" {
		return
	}

	targetMsg := data.TargetMessage()
	rawURL, found := ExtractURLFromMessage(targetMsg)
	if !found {
		_ = event.CreateMessage(discord.MessageCreate{
			Content: "❌ No media URL found in the selected message.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return
	}

	b.processDownload(event, rawURL)
}

func (b *Bot) handleSlashCommand(event *events.ApplicationCommandInteractionCreate) {
	data := event.SlashCommandInteractionData()
	if data.CommandName() != "download" {
		return
	}

	rawInput := data.String("url")
	cleanedURL, found := ExtractURL(rawInput)
	if !found {
		_ = event.CreateMessage(discord.MessageCreate{
			Content: "❌ Please provide a valid HTTP/HTTPS media URL.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return
	}

	b.processDownload(event, cleanedURL)
}

func (b *Bot) processDownload(event *events.ApplicationCommandInteractionCreate, targetURL string) {
	// Defer response so Discord doesn't time out the interaction
	if err := event.DeferCreateMessage(false); err != nil {
		slog.Error("error deferring interaction", slog.Any("err", err))
		return
	}

	// Acquire concurrency slot
	select {
	case b.semaphore <- struct{}{}:
		defer func() { <-b.semaphore }()
	default:
		// Queue busy notification if desired, or wait
		b.semaphore <- struct{}{}
		defer func() { <-b.semaphore }()
	}

	ctx := context.Background()
	media, err := b.downloader.Download(ctx, targetURL)
	if err != nil {
		slog.Warn("download failed", slog.String("url", targetURL), slog.Any("err", err))
		b.sendFollowup(event, discord.MessageCreate{
			Content: fmt.Sprintf("❌ %s", truncateString(err.Error(), 1800)),
		})
		return
	}
	defer media.Cleanup()

	file, err := os.Open(media.FilePath)
	if err != nil {
		slog.Error("error opening downloaded media file", slog.Any("err", err))
		b.sendFollowup(event, discord.MessageCreate{
			Content: "❌ Error reading downloaded media file for upload.",
		})
		return
	}
	defer file.Close()

	slog.Info("uploading media to Discord",
		slog.String("filename", media.Filename),
		slog.String("size", formatBytes(media.Size)),
	)

	_, err = event.Client().Rest.CreateFollowupMessage(event.ApplicationID(), event.Token(),
		discord.MessageCreate{
			Files: []*discord.File{
				{
					Name:   media.Filename,
					Reader: file,
				},
			},
		},
	)
	if err != nil {
		slog.Error("error uploading file to Discord", slog.Any("err", err))
		b.sendFollowup(event, discord.MessageCreate{
			Content: fmt.Sprintf("❌ Failed to upload media attachment to Discord: %s", truncateString(err.Error(), 1500)),
		})
	}
}

func (b *Bot) sendFollowup(event *events.ApplicationCommandInteractionCreate, message discord.MessageCreate) {
	if _, err := event.Client().Rest.CreateFollowupMessage(event.ApplicationID(), event.Token(), message); err != nil {
		slog.Error("error sending followup message", slog.Any("err", err))
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

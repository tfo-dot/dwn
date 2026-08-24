# dwn

Discord bot for downloading media from URLs (using `yt-dlp`) and uploading them directly into Discord channels.

## Features

- **Context Menu Command:** Right-click any message → `Apps` → `Download media`.
- **Slash Command:** Use `/download url:<media_url>` in any channel or DM.
- **Discord Attachment Limit (20 MB):** Enforces a strict 20 MB size limit (configurable) and formats media to stay within limit.
- **Auto-Discovery:** Automatically locates `yt-dlp` in `$PATH` if `YTDLP_PATH` is not explicitly set.
- **Concurrency Control:** Limits simultaneous downloads via a configurable worker semaphore to avoid CPU/network exhaustion.
- **Safe Temp File Handling:** Downloads directly to temporary directories and reliably cleans them up after upload.
- **Clean Error Handling:** Catches and sanitizes yt-dlp errors for user-friendly Discord feedback.

## Requirements

- Go 1.26+
- [`yt-dlp`](https://github.com/yt-dlp/yt-dlp) installed and in your `PATH` (or specified via config)
- `ffmpeg` (recommended for yt-dlp audio/video muxing)

## Configuration

You can configure `dwn` via a JSON file or environment variables.

### Configuration File (`example_config.json`)

```json
{
  "DISCORD_BOT_TOKEN": "YOUR_DISCORD_BOT_TOKEN",
  "YTDLP_PATH": "/usr/bin/yt-dlp",
  "MAX_FILE_SIZE_MB": 20,
  "MAX_CONCURRENT": 4,
  "DOWNLOAD_TIMEOUT_S": 180
}
```

### Environment Variables

| Variable | Description | Default |
| --- | --- | --- |
| `DISCORD_BOT_TOKEN` | Discord Bot Token (**required**) | — |
| `YTDLP_PATH` | Path to `yt-dlp` binary | Auto-detected in `$PATH` |
| `MAX_FILE_SIZE_MB` | Maximum attachment file size limit in MB | `20` |
| `MAX_CONCURRENT` | Maximum concurrent downloads | `4` |
| `DOWNLOAD_TIMEOUT_S` | Download timeout per item in seconds | `180` |

## Running

### From Environment Variables:
```bash
export DISCORD_BOT_TOKEN="your_token_here"
go run .
```

### From Configuration File:
```bash
go run . config.json
# or
go run . -config config.json
```

### Running Tests:
```bash
go test -v ./...
```

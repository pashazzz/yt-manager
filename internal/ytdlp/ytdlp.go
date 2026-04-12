package ytdlp

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	gytdlp "github.com/lrstanley/go-ytdlp"
)

// PlaylistEntry — одно видео из плейлиста, как его возвращает yt-dlp.
type PlaylistEntry struct {
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	Duration float64 `json:"duration"`
	URL      string  `json:"url"`
}

// PlaylistInfo — результат парсинга --dump-single-json для плейлиста.
type PlaylistInfo struct {
	Title   string          `json:"title"`
	Entries []PlaylistEntry `json:"entries"`
}

// Client — обёртка над yt-dlp, которая отвечает за получение метаданных.
type Client struct {
	// бинарь yt-dlp (по умолчанию ищется в PATH)
	binary string
}

// NewClient создаёт клиент и при необходимости загружает yt-dlp через go-ytdlp.
func NewClient(ctx context.Context) (*Client, error) {
	// Пытаемся найти yt-dlp в PATH.
	path, err := exec.LookPath("yt-dlp")
	if err != nil {
		// Если не найден — просим go-ytdlp установить его.
		if _, installErr := gytdlp.Install(ctx, nil); installErr != nil {
			return nil, fmt.Errorf("yt-dlp not found in PATH and auto-install failed: %w", installErr)
		}
		path = "yt-dlp"
	}
	return &Client{binary: path}, nil
}

// FetchPlaylist вызывает:
//
//	yt-dlp --flat-playlist --dump-single-json <url>
//
// и возвращает распарсенные метаданные плейлиста.
func (c *Client) FetchPlaylist(ctx context.Context, url string) (*PlaylistInfo, error) {
	cmd := exec.CommandContext(ctx, c.binary,
		"--flat-playlist",
		"--dump-single-json",
		"--no-warnings",
		url,
	)

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("yt-dlp execution failed: %w", err)
	}

	var info PlaylistInfo
	if err := json.Unmarshal(out, &info); err != nil {
		return nil, fmt.Errorf("failed to parse yt-dlp output: %w", err)
	}

	return &info, nil
}

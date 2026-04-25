package providers

import (
	"context"
	"strings"

	"github.com/pavlo/yt-manager/internal/ytdlp"
)

// ProviderYouTube — канонический идентификатор YouTube-провайдера.
const ProviderYouTube = "youtube"

// YouTube — провайдер для youtube.com / youtu.be. Делегирует загрузку в yt-dlp.
type YouTube struct {
	yt *ytdlp.Client
}

// NewYouTube создаёт YouTube-провайдера.
func NewYouTube(yt *ytdlp.Client) *YouTube { return &YouTube{yt: yt} }

// Name возвращает канонический идентификатор.
func (y *YouTube) Name() string { return ProviderYouTube }

// Matches распознаёт URL YouTube по хосту.
func (y *YouTube) Matches(url string) bool {
	u := strings.ToLower(url)
	return strings.Contains(u, "youtube.com/") ||
		strings.Contains(u, "youtu.be/") ||
		strings.Contains(u, "youtube-nocookie.com/")
}

// Fetch получает метаданные через yt-dlp.
func (y *YouTube) Fetch(ctx context.Context, url string) (*PlaylistInfo, error) {
	info, err := y.yt.FetchPlaylist(ctx, url)
	if err != nil {
		return nil, err
	}
	return &PlaylistInfo{Title: info.Title, Entries: info.Entries}, nil
}

// VideoURL строит URL одиночного видео.
func (y *YouTube) VideoURL(videoID string) string {
	return "https://www.youtube.com/watch?v=" + videoID
}

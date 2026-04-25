package providers

import (
	"context"
	"strings"

	"github.com/pavlo/yt-manager/internal/ytdlp"
)

// ProviderRutube — канонический идентификатор Rutube-провайдера.
const ProviderRutube = "rutube"

// Rutube — провайдер для rutube.ru. yt-dlp поддерживает Rutube нативно,
// поэтому реализация делегирует загрузку туда же.
type Rutube struct {
	yt *ytdlp.Client
}

// NewRutube создаёт Rutube-провайдера.
func NewRutube(yt *ytdlp.Client) *Rutube { return &Rutube{yt: yt} }

// Name возвращает канонический идентификатор.
func (r *Rutube) Name() string { return ProviderRutube }

// Matches распознаёт URL Rutube по хосту.
func (r *Rutube) Matches(url string) bool {
	u := strings.ToLower(url)
	return strings.Contains(u, "rutube.ru/") || strings.Contains(u, "rutube.com/")
}

// Fetch получает метаданные через yt-dlp.
func (r *Rutube) Fetch(ctx context.Context, url string) (*PlaylistInfo, error) {
	info, err := r.yt.FetchPlaylist(ctx, url)
	if err != nil {
		return nil, err
	}
	return &PlaylistInfo{Title: info.Title, Entries: info.Entries}, nil
}

// VideoURL строит URL одиночного видео.
func (r *Rutube) VideoURL(videoID string) string {
	return "https://rutube.ru/video/" + videoID + "/"
}

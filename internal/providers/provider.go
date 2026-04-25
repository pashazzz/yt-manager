// Package providers описывает источники видео (YouTube, Rutube, …) и предоставляет
// реестр для автоматического выбора по URL. Новые источники добавляются реализацией
// интерфейса Provider и регистрацией в DefaultRegistry.
package providers

import (
	"context"
	"fmt"

	"github.com/pavlo/yt-manager/internal/ytdlp"
)

// PlaylistInfo — результат получения метаданных плейлиста или одиночного видео.
type PlaylistInfo struct {
	Title   string                `json:"title"`
	Entries []ytdlp.PlaylistEntry `json:"entries"`
}

// Provider — источник видео (YouTube, Rutube, …).
type Provider interface {
	// Name — канонический идентификатор провайдера, сохраняется в БД у эпизодов.
	Name() string
	// Matches возвращает true, если URL относится к данному провайдеру.
	Matches(url string) bool
	// Fetch получает метаданные плейлиста или одиночного видео.
	Fetch(ctx context.Context, url string) (*PlaylistInfo, error)
	// VideoURL строит каноничный URL одиночного видео по его videoId.
	// Используется бэкфилом превью для старых эпизодов (когда мы не сохраняли URL).
	VideoURL(videoID string) string
}

// Registry — упорядоченный список провайдеров для автодетекта по URL.
type Registry struct {
	providers []Provider
}

// NewRegistry создаёт пустой реестр. Порядок регистрации определяет порядок проверки.
func NewRegistry() *Registry { return &Registry{} }

// NewDefaultRegistry регистрирует встроенные провайдеры (YouTube, Rutube).
func NewDefaultRegistry(yt *ytdlp.Client) *Registry {
	r := NewRegistry()
	r.Register(NewYouTube(yt))
	r.Register(NewRutube(yt))
	return r
}

// Register добавляет провайдера в реестр.
func (r *Registry) Register(p Provider) { r.providers = append(r.providers, p) }

// ByName ищет зарегистрированного провайдера по каноническому имени.
// Возвращает nil, если такого нет (например, эпизоды с устаревшим именем).
func (r *Registry) ByName(name string) Provider {
	for _, p := range r.providers {
		if p.Name() == name {
			return p
		}
	}
	return nil
}

// Detect возвращает провайдера, который поддерживает данный URL.
// Ошибка возвращается, если ни один провайдер не распознал URL.
func (r *Registry) Detect(url string) (Provider, error) {
	for _, p := range r.providers {
		if p.Matches(url) {
			return p, nil
		}
	}
	return nil, fmt.Errorf("unsupported video provider for url: %s", url)
}

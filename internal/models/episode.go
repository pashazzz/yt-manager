package models

// Episode представляет отдельное видео внутри плейлиста (шоу).
type Episode struct {
	ID          string  `json:"id"`
	ShowID      string  `json:"showId"`
	VideoID     string  `json:"videoId"`    // YouTube video ID
	Title       string  `json:"title"`
	Duration    float64 `json:"duration"`   // длительность в секундах
	CurrentTime float64 `json:"currentTime"` // текущая позиция в секундах
	IsWatched   bool     `json:"isWatched"`
	OrderIndex  int      `json:"orderIndex"` // порядок в плейлисте
	TagIDs      []string `json:"tagIds"`     // теги для одиночных видео
}

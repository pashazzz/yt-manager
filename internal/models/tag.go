package models

import "time"

// Tag — тег для группировки шоу и видео одного профиля.
type Tag struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	OwnerID    string    `json:"ownerId"`
	IsDefault  bool      `json:"isDefault"` // тег «Default» создаётся автоматически
	OrderIndex int       `json:"orderIndex"`
	UseThumb   bool      `json:"useThumb"`
	CreatedAt  time.Time `json:"createdAt"`
}

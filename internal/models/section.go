package models

import "time"

// Section — раздел (папка) для группировки шоу одного профиля.
type Section struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	OwnerID   string    `json:"ownerId"`
	IsDefault bool      `json:"isDefault"` // раздел «Default» создаётся автоматически
	CreatedAt time.Time `json:"createdAt"`
}

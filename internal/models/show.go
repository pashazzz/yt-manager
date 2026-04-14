package models

import "time"

// Show представляет YouTube-плейлист как сериал.
type Show struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	PlaylistURL  string    `json:"playlistUrl"`
	OwnerID      string    `json:"ownerId"`
	TagIDs       []string  `json:"tagIds"` // список тегов
	ReverseOrder bool      `json:"reverseOrder"`
	IsSingles    bool      `json:"isSingles"`
	OrderIndex   int       `json:"orderIndex"`
	CreatedAt    time.Time `json:"createdAt"`
}

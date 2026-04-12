package models

import "time"

// Show представляет YouTube-плейлист как сериал.
type Show struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	PlaylistURL string    `json:"playlistUrl"`
	OwnerID     string    `json:"ownerId"` // Profile.ID
	CreatedAt   time.Time `json:"createdAt"`
}

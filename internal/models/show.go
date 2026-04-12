package models

import "time"

// Show представляет YouTube-плейлист как сериал.
type Show struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	PlaylistURL  string    `json:"playlistUrl"`
	OwnerID      string    `json:"ownerId"`
	SectionID    string    `json:"sectionId"` // к какому разделу принадлежит
	ReverseOrder bool      `json:"reverseOrder"`
	IsSingles    bool      `json:"isSingles"`
	CreatedAt    time.Time `json:"createdAt"`
}

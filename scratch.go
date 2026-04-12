package main

import (
	"context"
	"fmt"
	"github.com/pavlo/yt-manager/internal/ytdlp"
)

func main() {
	c, _ := ytdlp.NewClient(context.Background())
	res, err := c.FetchPlaylist(context.Background(), "https://www.youtube.com/watch?v=dQw4w9WgXcQ")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Title: %s, Entries: %d\n", res.Title, len(res.Entries))
		for _, e := range res.Entries {
			fmt.Printf("  - %s (%s) %vs\n", e.Title, e.ID, e.Duration)
		}
		// print all json representation
		fmt.Printf("Title: %s, duration: %v, id: %s\n", res.Title, res.Entries[0].Duration, res.Entries[0].ID)
	}
}

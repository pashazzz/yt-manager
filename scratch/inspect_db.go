package main

import (
	"fmt"
	"log"

	clover "github.com/ostafen/clover/v2"
	"github.com/ostafen/clover/v2/query"
)

func main() {
	db, err := clover.Open("./data")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("--- Shows (Singles) ---")
	docs, _ := db.FindAll(query.NewQuery("shows").Where(query.Field("isSingles").Eq(true)))
	for _, d := range docs {
		fmt.Printf("Show ID: %s, Title: %s, Owner: %v, Tags: %v\n", d.ObjectId(), d.Get("title"), d.Get("ownerId"), d.Get("tagIds"))
	}

	fmt.Println("\n--- Episodes with empty TagIDs ---")
	// Попробуем разные варианты поиска пустых тегов
	q := query.NewQuery("episodes").Where(query.Field("showId").IsNotNull())
    // Мы отфильтруем в коде, чтобы быть уверенными
	eps, _ := db.FindAll(q)
	count := 0
	for _, d := range eps {
		tagIds := d.Get("tagIds")
        showId := d.Get("showId")
		if tagIds == nil {
			fmt.Printf("Episode: %s, ShowID: %s, TagIDs: nil (null)\n", d.ObjectId(), showId)
            count++
		} else if slice, ok := tagIds.([]any); ok && len(slice) == 0 {
			fmt.Printf("Episode: %s, ShowID: %s, TagIDs: [] (empty)\n", d.ObjectId(), showId)
            count++
		}
	}
    fmt.Printf("Total orphaned episodes found: %d\n", count)
}

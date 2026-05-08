package repository

import (
	"fmt"
	"log"

	clover "github.com/ostafen/clover/v2"
	"github.com/ostafen/clover/v2/query"

	"github.com/pavlo/yt-manager/internal/db"
)

// MigrateSectionIDs переносит устаревшее поле sectionId в массив tagIds для
// коллекций shows и episodes. Это одноразовая миграция: после её прогона
// репозитории больше не должны опираться на sectionId.
//
// Стратегия: документы, у которых уже есть непустой tagIds, не трогаем.
// Поле sectionId оставляем нетронутым на случай отката.
func MigrateSectionIDs(database *clover.DB) error {
	totalShows, err := migrateCollection(database, db.CollectionShows)
	if err != nil {
		return fmt.Errorf("migrate shows: %w", err)
	}
	totalEpisodes, err := migrateCollection(database, db.CollectionEpisodes)
	if err != nil {
		return fmt.Errorf("migrate episodes: %w", err)
	}
	if totalShows > 0 || totalEpisodes > 0 {
		log.Printf("migrations: sectionId→tagIds — shows: %d, episodes: %d", totalShows, totalEpisodes)
	}
	return nil
}

func migrateCollection(database *clover.DB, collection string) (int, error) {
	docs, err := database.FindAll(query.NewQuery(collection))
	if err != nil {
		return 0, err
	}
	updated := 0
	for _, d := range docs {
		// Если уже есть непустой tagIds — пропускаем.
		if hasNonEmptyStringSlice(d, "tagIds") {
			continue
		}
		sid, ok := d.Get("sectionId").(string)
		if !ok || sid == "" {
			continue
		}
		q := query.NewQuery(collection).Where(query.Field("_id").Eq(d.ObjectId()))
		if err := database.Update(q, map[string]any{"tagIds": []string{sid}}); err != nil {
			return updated, err
		}
		updated++
	}
	return updated, nil
}

func hasNonEmptyStringSlice(d interface{ Get(string) any }, key string) bool {
	v := d.Get(key)
	switch s := v.(type) {
	case []string:
		return len(s) > 0
	case []any:
		return len(s) > 0
	default:
		return false
	}
}

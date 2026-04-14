package db

import (
	clover "github.com/ostafen/clover/v2"
)

const (
	CollectionProfiles = "profiles"
	CollectionTags     = "sections" // Название в БД оставляем для совместимости
	CollectionShows    = "shows"
	CollectionEpisodes = "episodes"
)

// Open открывает (или создаёт) базу данных CloverDB по указанному пути
// и гарантирует существование всех необходимых коллекций.
func Open(dataDir string) (*clover.DB, error) {
	db, err := clover.Open(dataDir)
	if err != nil {
		return nil, err
	}

	for _, col := range []string{CollectionProfiles, CollectionTags, CollectionShows, CollectionEpisodes} {
		exists, err := db.HasCollection(col)
		if err != nil {
			return nil, err
		}
		if !exists {
			if err := db.CreateCollection(col); err != nil {
				return nil, err
			}
		}
	}

	return db, nil
}

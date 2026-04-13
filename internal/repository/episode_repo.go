package repository

import (
	"errors"

	"github.com/google/uuid"
	clover "github.com/ostafen/clover/v2"
	"github.com/ostafen/clover/v2/document"
	"github.com/ostafen/clover/v2/query"

	"github.com/pavlo/yt-manager/internal/db"
	"github.com/pavlo/yt-manager/internal/models"
)

// EpisodeRepo обеспечивает хранение и получение эпизодов.
type EpisodeRepo struct {
	db *clover.DB
}

func NewEpisodeRepo(database *clover.DB) *EpisodeRepo {
	return &EpisodeRepo{db: database}
}

// BulkCreate вставляет список эпизодов одним вызовом (при создании шоу).
func (r *EpisodeRepo) BulkCreate(episodes []*models.Episode) error {
	docs := make([]*document.Document, 0, len(episodes))
	for _, ep := range episodes {
		ep.ID = uuid.NewString()
		docs = append(docs, episodeToDoc(ep))
	}
	return r.db.Insert(db.CollectionEpisodes, docs...)
}

// FindByShow возвращает все эпизоды шоу, отсортированные по OrderIndex.
func (r *EpisodeRepo) FindByShow(showID string) ([]*models.Episode, error) {
	q := query.NewQuery(db.CollectionEpisodes).
		Where(query.Field("showId").Eq(showID)).
		Sort(query.SortOption{Field: "orderIndex", Direction: 1})

	docs, err := r.db.FindAll(q)
	if err != nil {
		return nil, err
	}

	episodes := make([]*models.Episode, 0, len(docs))
	for _, d := range docs {
		episodes = append(episodes, docToEpisode(d))
	}
	return episodes, nil
}

// FindByID возвращает эпизод по ID.
func (r *EpisodeRepo) FindByID(id string) (*models.Episode, error) {
	q := query.NewQuery(db.CollectionEpisodes).Where(query.Field("_id").Eq(id))
	doc, err := r.db.FindFirst(q)
	if err != nil {
		if errors.Is(err, clover.ErrDocumentNotExist) {
			return nil, nil
		}
		return nil, err
	}
	if doc == nil {
		return nil, nil
	}
	return docToEpisode(doc), nil
}

// GetMaxOrderIndex возвращает максимальный orderIndex для данного шоу.
// Если эпизодов нет, возвращает -1.
func (r *EpisodeRepo) GetMaxOrderIndex(showID string) (int, error) {
	q := query.NewQuery(db.CollectionEpisodes).
		Where(query.Field("showId").Eq(showID)).
		Sort(query.SortOption{Field: "orderIndex", Direction: -1})

	doc, err := r.db.FindFirst(q)
	if err != nil {
		if errors.Is(err, clover.ErrDocumentNotExist) {
			return -1, nil
		}
		return 0, err
	}
	if doc == nil {
		return -1, nil
	}
	return intField(doc, "orderIndex"), nil
}

// UpdateProgress сохраняет текущую позицию воспроизведения и флаг просмотра.
func (r *EpisodeRepo) UpdateProgress(id string, currentTime float64, isWatched bool) error {
	q := query.NewQuery(db.CollectionEpisodes).Where(query.Field("_id").Eq(id))
	return r.db.Update(q, map[string]any{
		"currentTime": currentTime,
		"isWatched":   isWatched,
	})
}

// DeleteByShow удаляет все эпизоды указанного шоу.
func (r *EpisodeRepo) DeleteByShow(showID string) error {
	return r.db.Delete(
		query.NewQuery(db.CollectionEpisodes).Where(query.Field("showId").Eq(showID)),
	)
}

// UpdateOrder обновляет orderIndex для списка эпизодов (используется для drag & drop в пользовательских плейлистах).
func (r *EpisodeRepo) UpdateOrder(showID string, orderedIDs []string) error {
	for i, id := range orderedIDs {
		q := query.NewQuery(db.CollectionEpisodes).Where(
			query.Field("_id").Eq(id).And(query.Field("showId").Eq(showID)),
		)
		if err := r.db.Update(q, map[string]any{"orderIndex": i}); err != nil {
			return err
		}
	}
	return nil
}

// MoveEpisode обновляет ID шоу и orderIndex у выбранного эпизода
func (r *EpisodeRepo) MoveEpisode(id string, targetShowID string, newOrderIndex int) error {
	q := query.NewQuery(db.CollectionEpisodes).Where(query.Field("_id").Eq(id))
	return r.db.Update(q, map[string]any{
		"showId":     targetShowID,
		"orderIndex": newOrderIndex,
	})
}

// --- helpers ---

func episodeToDoc(ep *models.Episode) *document.Document {
	return document.NewDocumentOf(map[string]any{
		"_id":         ep.ID,
		"showId":      ep.ShowID,
		"videoId":     ep.VideoID,
		"title":       ep.Title,
		"duration":    ep.Duration,
		"currentTime": ep.CurrentTime,
		"isWatched":   ep.IsWatched,
		"orderIndex":  ep.OrderIndex,
	})
}

func docToEpisode(d *document.Document) *models.Episode {
	return &models.Episode{
		ID:          d.ObjectId(),
		ShowID:      stringField(d, "showId"),
		VideoID:     stringField(d, "videoId"),
		Title:       stringField(d, "title"),
		Duration:    floatField(d, "duration"),
		CurrentTime: floatField(d, "currentTime"),
		IsWatched:   boolField(d, "isWatched"),
		OrderIndex:  intField(d, "orderIndex"),
	}
}

func floatField(d *document.Document, key string) float64 {
	switch v := d.Get(key).(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	}
	return 0
}

func boolField(d *document.Document, key string) bool {
	v, _ := d.Get(key).(bool)
	return v
}

func intField(d *document.Document, key string) int {
	switch v := d.Get(key).(type) {
	case int:
		return v
	case float64:
		return int(v)
	}
	return 0
}

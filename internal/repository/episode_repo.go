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

// FindByOwnerID возвращает все эпизоды, принадлежащие владельцу (из всех шоу).
func (r *EpisodeRepo) FindByOwnerID(ownerID string) ([]*models.Episode, error) {
	// 1. Находим все шоу владельца
	qShows := query.NewQuery(db.CollectionShows).
		Where(query.Field("ownerId").Eq(ownerID))
	showsDocs, err := r.db.FindAll(qShows)
	if err != nil {
		return nil, err
	}
	if len(showsDocs) == 0 {
		return nil, nil
	}

	showIDs := make([]any, 0, len(showsDocs))
	for _, d := range showsDocs {
		showIDs = append(showIDs, d.ObjectId())
	}

	// 2. Ищем эпизоды в этих шоу
	qEps := query.NewQuery(db.CollectionEpisodes).
		Where(query.Field("showId").In(showIDs...)).
		Sort(query.SortOption{Field: "orderIndex", Direction: 1})
	docs, err := r.db.FindAll(qEps)
	if err != nil {
		return nil, err
	}

	eps := make([]*models.Episode, 0, len(docs))
	for _, d := range docs {
		eps = append(eps, docToEpisode(d))
	}
	return eps, nil
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

// FindSinglesByTag возвращает все одиночные видео (эпизоды из скрытых шоу),
// которые отмечены данным тегом. Сначала находит все служебные шоу пользователя,
// затем выбирает из них эпизоды с нужным тегом.
func (r *EpisodeRepo) FindSinglesByTag(ownerID, tagID string, isDefault bool) ([]*models.Episode, error) {
	// 1. Получаем все служебные шоу владельца
	qShows := query.NewQuery(db.CollectionShows).
		Where(query.Field("ownerId").Eq(ownerID).And(query.Field("isSingles").Eq(true)))
	showsDocs, err := r.db.FindAll(qShows)
	if err != nil {
		return nil, err
	}
	if len(showsDocs) == 0 {
		return nil, nil
	}

	showIDs := make([]any, 0, len(showsDocs))
	for _, d := range showsDocs {
		showIDs = append(showIDs, d.ObjectId())
	}

	// 2. Ищем ВСЕ эпизоды в этих шоу (без фильтрации по тегам на уровне БД, чтобы сработала миграция в docToEpisode)
	qEps := query.NewQuery(db.CollectionEpisodes).
		Where(query.Field("showId").In(showIDs...)).
		Sort(query.SortOption{Field: "orderIndex", Direction: 1})
	docs, err := r.db.FindAll(qEps)
	if err != nil {
		return nil, err
	}

	episodes := make([]*models.Episode, 0, len(docs))
	for _, d := range docs {
		ep := docToEpisode(d)
		
		isMatch := false
		// Проверяем, есть ли нужный тег в (уже смигрированном) списке
		for _, tid := range ep.TagIDs {
			if tid == tagID {
				isMatch = true
				break
			}
		}
		
		// Если это дефолтный тег — показываем также те, у которых нет тегов
		if !isMatch && isDefault && len(ep.TagIDs) == 0 {
			isMatch = true
		}

		if isMatch {
			episodes = append(episodes, ep)
		}
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

// Delete удаляет один эпизод по ID.
func (r *EpisodeRepo) Delete(id string) error {
	return r.db.Delete(
		query.NewQuery(db.CollectionEpisodes).Where(query.Field("_id").Eq(id)),
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
	provider := ep.Provider
	if provider == "" {
		provider = "youtube"
	}
	return document.NewDocumentOf(map[string]any{
		"_id":          ep.ID,
		"showId":       ep.ShowID,
		"provider":     provider,
		"videoId":      ep.VideoID,
		"title":        ep.Title,
		"duration":     ep.Duration,
		"currentTime":  ep.CurrentTime,
		"isWatched":    ep.IsWatched,
		"orderIndex":   ep.OrderIndex,
		"tagIds":       ep.TagIDs,
		"thumbnailUrl": ep.ThumbnailURL,
	})
}

func docToEpisode(d *document.Document) *models.Episode {
	tagIDs := stringSliceField(d, "tagIds")

	// Миграция: если нет tagIds, но есть sectionId, используем его как единственный тег
	if len(tagIDs) == 0 && d.Has("sectionId") {
		if sid, ok := d.Get("sectionId").(string); ok && sid != "" {
			tagIDs = []string{sid}
		}
	}

	// Обратная совместимость: старые записи без provider считаем YouTube.
	provider := stringField(d, "provider")
	if provider == "" {
		provider = "youtube"
	}

	return &models.Episode{
		ID:           d.ObjectId(),
		ShowID:       stringField(d, "showId"),
		Provider:     provider,
		VideoID:      stringField(d, "videoId"),
		Title:        stringField(d, "title"),
		Duration:     floatField(d, "duration"),
		CurrentTime:  floatField(d, "currentTime"),
		IsWatched:    boolField(d, "isWatched"),
		OrderIndex:   intField(d, "orderIndex"),
		TagIDs:       tagIDs,
		ThumbnailURL: stringField(d, "thumbnailUrl"),
	}
}

func (r *EpisodeRepo) UpdateTags(id string, tagIDs []string) error {
	q := query.NewQuery(db.CollectionEpisodes).Where(query.Field("_id").Eq(id))
	return r.db.Update(q, map[string]any{"tagIds": tagIDs})
}

// UpdateThumbnail обновляет URL превью для эпизода. Используется бэкфилом.
func (r *EpisodeRepo) UpdateThumbnail(id, thumbnailURL string) error {
	q := query.NewQuery(db.CollectionEpisodes).Where(query.Field("_id").Eq(id))
	return r.db.Update(q, map[string]any{"thumbnailUrl": thumbnailURL})
}

// FindMissingThumbnails возвращает все эпизоды, у которых не заполнен thumbnailUrl.
// Используется на старте сервера для ленивой миграции.
func (r *EpisodeRepo) FindMissingThumbnails() ([]*models.Episode, error) {
	docs, err := r.db.FindAll(query.NewQuery(db.CollectionEpisodes))
	if err != nil {
		return nil, err
	}
	out := make([]*models.Episode, 0)
	for _, d := range docs {
		ep := docToEpisode(d)
		if ep.ThumbnailURL == "" {
			out = append(out, ep)
		}
	}
	return out, nil
}

func stringSliceField(d *document.Document, key string) []string {
	result := []string{}
	if !d.Has(key) {
		return result
	}
	if raw, ok := d.Get(key).([]any); ok {
		for _, v := range raw {
			if s, ok := v.(string); ok {
				result = append(result, s)
			}
		}
	} else if raw, ok := d.Get(key).([]string); ok {
		result = raw
	}
	return result
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

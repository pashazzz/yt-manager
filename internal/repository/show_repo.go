package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
	clover "github.com/ostafen/clover/v2"
	"github.com/ostafen/clover/v2/document"
	"github.com/ostafen/clover/v2/query"

	"github.com/pavlo/yt-manager/internal/db"
	"github.com/pavlo/yt-manager/internal/models"
)

// ShowRepo обеспечивает хранение и получение шоу (плейлистов).
type ShowRepo struct {
	db *clover.DB
}

func NewShowRepo(database *clover.DB) *ShowRepo {
	return &ShowRepo{db: database}
}

// Create сохраняет новое шоу с заполненным ID. Если OrderIndex не задан явно,
// ставим max(orderIndex)+1 среди шоу владельца — так новое шоу всегда оказывается
// в конце списка, в том числе после ручных drag-and-drop переупорядочиваний.
func (r *ShowRepo) Create(s *models.Show) error {
	s.ID = uuid.NewString()
	s.CreatedAt = time.Now().UTC()

	if s.OrderIndex == 0 {
		maxIdx, err := r.getMaxOrderIndex(s.OwnerID)
		if err != nil {
			return err
		}
		s.OrderIndex = maxIdx + 1
	}

	doc := document.NewDocumentOf(map[string]any{
		"_id":         s.ID,
		"title":       s.Title,
		"playlistUrl": s.PlaylistURL,
		"ownerId":      s.OwnerID,
		"tagIds":       s.TagIDs,
		"reverseOrder": s.ReverseOrder,
		"isSingles":    s.IsSingles,
		"orderIndex":   s.OrderIndex,
		"createdAt":    s.CreatedAt,
	})
	return r.db.Insert(db.CollectionShows, doc)
}

// getMaxOrderIndex возвращает максимальный orderIndex среди шоу владельца.
// Если шоу ещё нет, возвращает 0.
func (r *ShowRepo) getMaxOrderIndex(ownerID string) (int, error) {
	q := query.NewQuery(db.CollectionShows).
		Where(query.Field("ownerId").Eq(ownerID)).
		Sort(query.SortOption{Field: "orderIndex", Direction: -1})
	doc, err := r.db.FindFirst(q)
	if err != nil {
		if errors.Is(err, clover.ErrDocumentNotExist) {
			return 0, nil
		}
		return 0, err
	}
	if doc == nil {
		return 0, nil
	}
	return intField(doc, "orderIndex"), nil
}

// FindByOwner возвращает все шоу профиля (для обратной совместимости).
func (r *ShowRepo) FindByOwner(ownerID string) ([]*models.Show, error) {
	q := query.NewQuery(db.CollectionShows).
		Where(query.Field("ownerId").Eq(ownerID)).
		Sort(query.SortOption{Field: "orderIndex", Direction: 1}, query.SortOption{Field: "createdAt", Direction: 1})

	docs, err := r.db.FindAll(q)
	if err != nil {
		return nil, err
	}
	shows := make([]*models.Show, 0, len(docs))
	for _, d := range docs {
		shows = append(shows, docToShow(d))
	}
	return shows, nil
}

// FindByTag возвращает шоу, отмеченные конкретным тегом (исключая служебные синглы).
// Если includeUncategorized=true (для Default-тега), также возвращает шоу без тегов.
func (r *ShowRepo) FindByTag(ownerID, tagID string, includeUncategorized bool) ([]*models.Show, error) {
	cond := query.Field("ownerId").Eq(ownerID).
		And(query.Field("isSingles").Neq(true)).
		And(query.Field("tagIds").Contains(tagID))

	q := query.NewQuery(db.CollectionShows).
		Where(cond).
		Sort(query.SortOption{Field: "orderIndex", Direction: 1}, query.SortOption{Field: "createdAt", Direction: 1})

	docs, err := r.db.FindAll(q)
	if err != nil {
		return nil, err
	}

	result := make([]*models.Show, 0, len(docs))
	for _, d := range docs {
		result = append(result, docToShow(d))
	}

	if !includeUncategorized {
		return result, nil
	}

	// Для Default-тега добавляем шоу без тегов вообще (legacy / новые без явного тега).
	qAll := query.NewQuery(db.CollectionShows).
		Where(query.Field("ownerId").Eq(ownerID).And(query.Field("isSingles").Neq(true))).
		Sort(query.SortOption{Field: "orderIndex", Direction: 1}, query.SortOption{Field: "createdAt", Direction: 1})
	allDocs, err := r.db.FindAll(qAll)
	if err != nil {
		return result, nil
	}
	seen := make(map[string]struct{}, len(result))
	for _, s := range result {
		seen[s.ID] = struct{}{}
	}
	for _, d := range allDocs {
		s := docToShow(d)
		if _, ok := seen[s.ID]; ok {
			continue
		}
		if len(s.TagIDs) == 0 {
			result = append(result, s)
		}
	}
	return result, nil
}

// FindByID возвращает шоу по ID или nil.
func (r *ShowRepo) FindByID(id string) (*models.Show, error) {
	q := query.NewQuery(db.CollectionShows).Where(query.Field("_id").Eq(id))
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
	return docToShow(doc), nil
}

// UpdateTags обновляет теги шоу.
func (r *ShowRepo) UpdateTags(id string, tagIDs []string) error {
	q := query.NewQuery(db.CollectionShows).Where(query.Field("_id").Eq(id))
	return r.db.Update(q, map[string]any{"tagIds": tagIDs})
}

// UpdateReverseOrder меняет порядок сортировки эпизодов.
func (r *ShowRepo) UpdateReverseOrder(id string, reversed bool) error {
	q := query.NewQuery(db.CollectionShows).Where(query.Field("_id").Eq(id))
	return r.db.Update(q, map[string]any{"reverseOrder": reversed})
}

// Delete удаляет шоу по ID.
func (r *ShowRepo) Delete(id string) error {
	return r.db.Delete(
		query.NewQuery(db.CollectionShows).Where(query.Field("_id").Eq(id)),
	)
}

// UpdateOrder обновляет orderIndex для списка шоу внутри тега.
func (r *ShowRepo) UpdateOrder(ownerID, tagID string, orderedIDs []string) error {
	for i, id := range orderedIDs {
		// Мы проверяем и ownerId для безопасности.
		q := query.NewQuery(db.CollectionShows).Where(
			query.Field("_id").Eq(id).
				And(query.Field("ownerId").Eq(ownerID)),
		)
		if err := r.db.Update(q, map[string]any{"orderIndex": i}); err != nil {
			return err
		}
	}
	return nil
}

// EnsureSinglesShow возвращает специальное скрытое шоу для отдельных видео тега,
// или создаёт его, если оно ещё не существует.
func (r *ShowRepo) EnsureSinglesShow(ownerID, tagID string) (*models.Show, error) {
	q := query.NewQuery(db.CollectionShows).
		Where(query.Field("ownerId").Eq(ownerID).
			And(query.Field("tagIds").Contains(tagID)).
			And(query.Field("isSingles").Eq(true)))

	doc, err := r.db.FindFirst(q)
	if err != nil && !errors.Is(err, clover.ErrDocumentNotExist) {
		return nil, err
	}
	if doc != nil {
		return docToShow(doc), nil
	}

	s := &models.Show{
		Title:       "Отдельные видео",
		OwnerID:     ownerID,
		TagIDs:      []string{tagID},
		IsSingles:   true,
	}
	if err := r.Create(s); err != nil {
		return nil, err
	}
	return s, nil
}

// --- helpers ---

func docToShow(d *document.Document) *models.Show {
	createdAt, _ := d.Get("createdAt").(time.Time)
	tagIDs := stringSliceField(d, "tagIds")

	return &models.Show{
		ID:           d.ObjectId(),
		Title:        stringField(d, "title"),
		PlaylistURL:  stringField(d, "playlistUrl"),
		OwnerID:      stringField(d, "ownerId"),
		TagIDs:       tagIDs,
		ReverseOrder: boolField(d, "reverseOrder"),
		IsSingles:    boolField(d, "isSingles"),
		OrderIndex:   intField(d, "orderIndex"),
		CreatedAt:    createdAt,
	}
}

func stringField(d *document.Document, key string) string {
	v, _ := d.Get(key).(string)
	return v
}

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
	return showIntField(doc, "orderIndex"), nil
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

// FindByTag возвращает шоу, принадлежащие конкретному тегу.
func (r *ShowRepo) FindByTag(ownerID, tagID string, includeUncategorized bool) ([]*models.Show, error) {
	all, err := r.FindByOwner(ownerID)
	if err != nil {
		return nil, err
	}

	result := make([]*models.Show, 0, len(all))
	for _, s := range all {
		if s.IsSingles {
			continue // Не возвращаем служебное шоу в списке обычных шоу
		}
		
		isTagged := false
		for _, tid := range s.TagIDs {
			if tid == tagID {
				isTagged = true
				break
			}
		}

		if isTagged {
			result = append(result, s)
		} else if includeUncategorized && len(s.TagIDs) == 0 {
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
	
	// Миграция: если нет tagIds, но есть sectionId, используем его как единственный тег
	tagIDs := []string{}
	if d.Has("tagIds") {
		if raw, ok := d.Get("tagIds").([]any); ok {
			for _, v := range raw {
				if s, ok := v.(string); ok {
					tagIDs = append(tagIDs, s)
				}
			}
		} else if raw, ok := d.Get("tagIds").([]string); ok {
			tagIDs = raw
		}
	}
	
	if len(tagIDs) == 0 && d.Has("sectionId") {
		if sid, ok := d.Get("sectionId").(string); ok && sid != "" {
			tagIDs = []string{sid}
		}
	}

	return &models.Show{
		ID:           d.ObjectId(),
		Title:        stringField(d, "title"),
		PlaylistURL:  stringField(d, "playlistUrl"),
		OwnerID:      stringField(d, "ownerId"),
		TagIDs:       tagIDs,
		ReverseOrder: showBoolField(d, "reverseOrder"),
		IsSingles:    showBoolField(d, "isSingles"),
		OrderIndex:   showIntField(d, "orderIndex"),
		CreatedAt:    createdAt,
	}
}

func stringField(d *document.Document, key string) string {
	v, _ := d.Get(key).(string)
	return v
}

func showBoolField(d *document.Document, key string) bool {
	v, _ := d.Get(key).(bool)
	return v
}

func showIntField(d *document.Document, key string) int {
	switch v := d.Get(key).(type) {
	case int:
		return v
	case float64:
		return int(v)
	case float32:
		return int(v)
	default:
		return 0
	}
}

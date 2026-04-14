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

// TagRepo обеспечивает хранение и получение тегов.
type TagRepo struct {
	db *clover.DB
}

func NewTagRepo(database *clover.DB) *TagRepo {
	return &TagRepo{db: database}
}

// Create сохраняет новый тег.
func (r *TagRepo) Create(s *models.Tag) error {
	s.ID = uuid.NewString()
	s.CreatedAt = time.Now().UTC()

	doc := document.NewDocumentOf(map[string]any{
		"_id":        s.ID,
		"name":       s.Name,
		"ownerId":    s.OwnerID,
		"isDefault":  s.IsDefault,
		"orderIndex": s.OrderIndex,
		"useThumb":   s.UseThumb,
		"createdAt":  s.CreatedAt,
	})
	return r.db.Insert(db.CollectionTags, doc)
}

// FindByOwner возвращает все теги профиля (сначала default, потом по дате).
func (r *TagRepo) FindByOwner(ownerID string) ([]*models.Tag, error) {
	q := query.NewQuery(db.CollectionTags).
		Where(query.Field("ownerId").Eq(ownerID)).
		Sort(query.SortOption{Field: "orderIndex", Direction: 1}, query.SortOption{Field: "createdAt", Direction: 1})

	docs, err := r.db.FindAll(q)
	if err != nil {
		return nil, err
	}
	tags := make([]*models.Tag, 0, len(docs))
	for _, d := range docs {
		tags = append(tags, docToTag(d))
	}
	return tags, nil
}

// FindByID возвращает тег по ID или nil.
func (r *TagRepo) FindByID(id string) (*models.Tag, error) {
	q := query.NewQuery(db.CollectionTags).Where(query.Field("_id").Eq(id))
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
	return docToTag(doc), nil
}

// EnsureDefault возвращает тег Default текущего профиля,
// создавая его при первом обращении.
func (r *TagRepo) EnsureDefault(ownerID string) (*models.Tag, error) {
	q := query.NewQuery(db.CollectionTags).Where(
		query.Field("ownerId").Eq(ownerID).And(query.Field("isDefault").Eq(true)),
	)
	doc, err := r.db.FindFirst(q)
	if err != nil && !errors.Is(err, clover.ErrDocumentNotExist) {
		return nil, err
	}
	if doc != nil {
		return docToTag(doc), nil
	}

	s := &models.Tag{Name: "Default", OwnerID: ownerID, IsDefault: true}
	if err := r.Create(s); err != nil {
		return nil, err
	}
	return s, nil
}

// Delete удаляет тег по ID.
func (r *TagRepo) Delete(id string) error {
	return r.db.Delete(
		query.NewQuery(db.CollectionTags).Where(query.Field("_id").Eq(id)),
	)
}

// UpdateOrder обновляет orderIndex для списка тегов.
func (r *TagRepo) UpdateOrder(ownerID string, orderedIDs []string) error {
	for i, id := range orderedIDs {
		q := query.NewQuery(db.CollectionTags).Where(
			query.Field("_id").Eq(id).And(query.Field("ownerId").Eq(ownerID)),
		)
		if err := r.db.Update(q, map[string]any{"orderIndex": i}); err != nil {
			return err
		}
	}
	return nil
}

// UpdateSettings обновляет настройки отображения тега.
func (r *TagRepo) UpdateSettings(id string, useThumb bool) error {
	q := query.NewQuery(db.CollectionTags).Where(query.Field("_id").Eq(id))
	return r.db.Update(q, map[string]any{"useThumb": useThumb})
}

func docToTag(d *document.Document) *models.Tag {
	createdAt, _ := d.Get("createdAt").(time.Time)
	return &models.Tag{
		ID:         d.ObjectId(),
		Name:       stringField(d, "name"),
		OwnerID:    stringField(d, "ownerId"),
		IsDefault:  boolField(d, "isDefault"),
		OrderIndex: tagIntField(d, "orderIndex"),
		UseThumb:   boolField(d, "useThumb"),
		CreatedAt:  createdAt,
	}
}

func tagIntField(d *document.Document, key string) int {
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

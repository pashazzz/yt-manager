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

// SectionRepo обеспечивает хранение и получение разделов.
type SectionRepo struct {
	db *clover.DB
}

func NewSectionRepo(database *clover.DB) *SectionRepo {
	return &SectionRepo{db: database}
}

// Create сохраняет новый раздел.
func (r *SectionRepo) Create(s *models.Section) error {
	s.ID = uuid.NewString()
	s.CreatedAt = time.Now().UTC()

	doc := document.NewDocumentOf(map[string]any{
		"_id":        s.ID,
		"name":       s.Name,
		"ownerId":    s.OwnerID,
		"isDefault":  s.IsDefault,
		"orderIndex": s.OrderIndex,
		"createdAt":  s.CreatedAt,
	})
	return r.db.Insert(db.CollectionSections, doc)
}

// FindByOwner возвращает все разделы профиля (сначала default, потом по дате).
func (r *SectionRepo) FindByOwner(ownerID string) ([]*models.Section, error) {
	q := query.NewQuery(db.CollectionSections).
		Where(query.Field("ownerId").Eq(ownerID)).
		Sort(query.SortOption{Field: "orderIndex", Direction: 1}, query.SortOption{Field: "createdAt", Direction: 1})

	docs, err := r.db.FindAll(q)
	if err != nil {
		return nil, err
	}
	sections := make([]*models.Section, 0, len(docs))
	for _, d := range docs {
		sections = append(sections, docToSection(d))
	}
	return sections, nil
}

// FindByID возвращает раздел по ID или nil.
func (r *SectionRepo) FindByID(id string) (*models.Section, error) {
	q := query.NewQuery(db.CollectionSections).Where(query.Field("_id").Eq(id))
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
	return docToSection(doc), nil
}

// EnsureDefault возвращает раздел Default текущего профиля,
// создавая его при первом обращении.
func (r *SectionRepo) EnsureDefault(ownerID string) (*models.Section, error) {
	q := query.NewQuery(db.CollectionSections).Where(
		query.Field("ownerId").Eq(ownerID).And(query.Field("isDefault").Eq(true)),
	)
	doc, err := r.db.FindFirst(q)
	if err != nil && !errors.Is(err, clover.ErrDocumentNotExist) {
		return nil, err
	}
	if doc != nil {
		return docToSection(doc), nil
	}

	s := &models.Section{Name: "Default", OwnerID: ownerID, IsDefault: true}
	if err := r.Create(s); err != nil {
		return nil, err
	}
	return s, nil
}

// Delete удаляет раздел по ID.
func (r *SectionRepo) Delete(id string) error {
	return r.db.Delete(
		query.NewQuery(db.CollectionSections).Where(query.Field("_id").Eq(id)),
	)
}

// UpdateOrder обновляет orderIndex для списка разделов.
func (r *SectionRepo) UpdateOrder(ownerID string, orderedIDs []string) error {
	for i, id := range orderedIDs {
		q := query.NewQuery(db.CollectionSections).Where(
			query.Field("_id").Eq(id).And(query.Field("ownerId").Eq(ownerID)),
		)
		if err := r.db.Update(q, map[string]any{"orderIndex": i}); err != nil {
			return err
		}
	}
	return nil
}

func docToSection(d *document.Document) *models.Section {
	createdAt, _ := d.Get("createdAt").(time.Time)
	return &models.Section{
		ID:         d.ObjectId(),
		Name:       stringField(d, "name"),
		OwnerID:    stringField(d, "ownerId"),
		IsDefault:  boolField(d, "isDefault"),
		OrderIndex: sectionIntField(d, "orderIndex"),
		CreatedAt:  createdAt,
	}
}

func sectionIntField(d *document.Document, key string) int {
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

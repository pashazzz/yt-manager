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

// Create сохраняет новое шоу с заполненным ID.
func (r *ShowRepo) Create(s *models.Show) error {
	s.ID = uuid.NewString()
	s.CreatedAt = time.Now().UTC()

	doc := document.NewDocumentOf(map[string]any{
		"_id":         s.ID,
		"title":       s.Title,
		"playlistUrl": s.PlaylistURL,
		"ownerId":      s.OwnerID,
		"sectionId":    s.SectionID,
		"reverseOrder": s.ReverseOrder,
		"createdAt":    s.CreatedAt,
	})
	return r.db.Insert(db.CollectionShows, doc)
}

// FindByOwner возвращает все шоу профиля (для обратной совместимости).
func (r *ShowRepo) FindByOwner(ownerID string) ([]*models.Show, error) {
	q := query.NewQuery(db.CollectionShows).
		Where(query.Field("ownerId").Eq(ownerID)).
		Sort(query.SortOption{Field: "createdAt", Direction: -1})

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

// FindBySection возвращает шоу, принадлежащие конкретному разделу.
// Если includeUncategorized=true, также возвращает шоу без sectionId
// (созданные до введения разделов) — используется для раздела Default.
func (r *ShowRepo) FindBySection(ownerID, sectionID string, includeUncategorized bool) ([]*models.Show, error) {
	// Загружаем все шоу профиля, фильтруем в Go — безопасно для pet-project
	all, err := r.FindByOwner(ownerID)
	if err != nil {
		return nil, err
	}

	result := make([]*models.Show, 0, len(all))
	for _, s := range all {
		if s.SectionID == sectionID {
			result = append(result, s)
		} else if includeUncategorized && s.SectionID == "" {
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

// UpdateSection перемещает шоу в другой раздел.
func (r *ShowRepo) UpdateSection(id, sectionID string) error {
	q := query.NewQuery(db.CollectionShows).Where(query.Field("_id").Eq(id))
	return r.db.Update(q, map[string]any{"sectionId": sectionID})
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

// --- helpers ---

func docToShow(d *document.Document) *models.Show {
	createdAt, _ := d.Get("createdAt").(time.Time)
	return &models.Show{
		ID:           d.ObjectId(),
		Title:        stringField(d, "title"),
		PlaylistURL:  stringField(d, "playlistUrl"),
		OwnerID:      stringField(d, "ownerId"),
		SectionID:    stringField(d, "sectionId"),
		ReverseOrder: showBoolField(d, "reverseOrder"),
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

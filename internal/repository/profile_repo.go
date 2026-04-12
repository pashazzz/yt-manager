package repository

import (
	"errors"

	clover "github.com/ostafen/clover/v2"
	"github.com/ostafen/clover/v2/document"
	"github.com/ostafen/clover/v2/query"

	"github.com/pavlo/yt-manager/internal/db"
	"github.com/pavlo/yt-manager/internal/models"
)

// ProfileRepo обеспечивает хранение и получение профилей.
type ProfileRepo struct {
	db *clover.DB
}

func NewProfileRepo(database *clover.DB) *ProfileRepo {
	return &ProfileRepo{db: database}
}

// Upsert создаёт профиль, если его ещё нет, иначе обновляет имя.
// Profile.ID (email из Tailscale) хранится в поле "login", а не "_id",
// чтобы не нарушать валидацию UUID-формата CloverDB.
func (r *ProfileRepo) Upsert(p *models.Profile) error {
	q := query.NewQuery(db.CollectionProfiles).Where(query.Field("login").Eq(p.ID))

	existing, err := r.db.FindFirst(q)
	if err != nil && !errors.Is(err, clover.ErrDocumentNotExist) {
		return err
	}

	if existing == nil {
		// Не задаём "_id" — CloverDB сам сгенерирует валидный UUID.
		doc := document.NewDocumentOf(map[string]any{
			"login": p.ID,
			"name":  p.Name,
		})
		return r.db.Insert(db.CollectionProfiles, doc)
	}

	return r.db.Update(q, map[string]any{"name": p.Name})
}

// FindByID возвращает профиль по Tailscale-логину (email) или nil.
func (r *ProfileRepo) FindByID(id string) (*models.Profile, error) {
	q := query.NewQuery(db.CollectionProfiles).Where(query.Field("login").Eq(id))

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

	login, _ := doc.Get("login").(string)
	name, _ := doc.Get("name").(string)
	return &models.Profile{
		ID:   login,
		Name: name,
	}, nil
}

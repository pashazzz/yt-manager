package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/pavlo/yt-manager/internal/middleware"
	"github.com/pavlo/yt-manager/internal/models"
	"github.com/pavlo/yt-manager/internal/repository"
)

// SectionHandler держит зависимости для хендлеров разделов.
type SectionHandler struct {
	sections *repository.SectionRepo
	shows    *repository.ShowRepo
}

func NewSectionHandler(sections *repository.SectionRepo, shows *repository.ShowRepo) *SectionHandler {
	return &SectionHandler{sections: sections, shows: shows}
}

// ListSections godoc
// GET /sections — возвращает все разделы, гарантируя существование Default.
func (h *SectionHandler) ListSections(c *gin.Context) {
	profile := middleware.GetProfile(c)

	// Гарантируем существование дефолтного раздела
	if _, err := h.sections.EnsureDefault(profile.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	sections, err := h.sections.FindByOwner(profile.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sections)
}

// CreateSection godoc
// POST /sections
// Body: { "name": "Аниме" }
func (h *SectionHandler) CreateSection(c *gin.Context) {
	profile := middleware.GetProfile(c)

	var body struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s := &models.Section{Name: body.Name, OwnerID: profile.ID}
	if err := h.sections.Create(s); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, s)
}

// DeleteSection godoc
// DELETE /sections/:id — нельзя удалить Default. Шоу переезжают в Default.
func (h *SectionHandler) DeleteSection(c *gin.Context) {
	profile := middleware.GetProfile(c)
	sectionID := c.Param("id")

	section, err := h.sections.FindByID(sectionID)
	if err != nil || section == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "section not found"})
		return
	}
	if section.OwnerID != profile.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	if section.IsDefault {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete default section"})
		return
	}

	// Переносим все шоу раздела в Default
	defaultSec, err := h.sections.EnsureDefault(profile.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	showsInSection, err := h.shows.FindBySection(profile.ID, sectionID, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, s := range showsInSection {
		_ = h.shows.UpdateSection(s.ID, defaultSec.ID)
	}

	if err := h.sections.Delete(sectionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ReorderSections godoc
// PATCH /sections/reorder
// Body: { "orderedIds": ["id1", "id2"] }
func (h *SectionHandler) ReorderSections(c *gin.Context) {
	profile := middleware.GetProfile(c)

	var body struct {
		OrderedIDs []string `json:"orderedIds" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.sections.UpdateOrder(profile.ID, body.OrderedIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// ListShowsBySection godoc
// GET /sections/:id/shows — шоу конкретного раздела.
func (h *SectionHandler) ListShowsBySection(c *gin.Context) {
	profile := middleware.GetProfile(c)
	sectionID := c.Param("id")

	section, err := h.sections.FindByID(sectionID)
	if err != nil || section == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "section not found"})
		return
	}
	if section.OwnerID != profile.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Default-раздел подбирает и «бесхозные» шоу (созданные до введения разделов)
	shows, err := h.shows.FindBySection(profile.ID, sectionID, section.IsDefault)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"section": section, "shows": shows})
}

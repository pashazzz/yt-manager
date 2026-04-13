package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/pavlo/yt-manager/internal/middleware"
	"github.com/pavlo/yt-manager/internal/models"
	"github.com/pavlo/yt-manager/internal/repository"
	"github.com/pavlo/yt-manager/internal/ytdlp"
)

// SectionHandler держит зависимости для хендлеров разделов.
type SectionHandler struct {
	sections *repository.SectionRepo
	shows    *repository.ShowRepo
	episodes *repository.EpisodeRepo
	ytClient *ytdlp.Client
}

func NewSectionHandler(
	sections *repository.SectionRepo,
	shows *repository.ShowRepo,
	episodes *repository.EpisodeRepo,
	ytClient *ytdlp.Client,
) *SectionHandler {
	return &SectionHandler{sections: sections, shows: shows, episodes: episodes, ytClient: ytClient}
}

type SectionInfo struct {
	models.Section
	ShowCount    int    `json:"showCount"`
	EpisodeCount int    `json:"episodeCount"`
	FirstVideoID string `json:"firstVideoId"`
}

// ListSections godoc
// GET /sections — возвращает все разделы со статистикой.
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

	// Собираем статистику для каждого раздела
	result := make([]SectionInfo, 0, len(sections))
	for _, s := range sections {
		shows, _ := h.shows.FindBySection(profile.ID, s.ID, s.IsDefault)
		
		singlesShow, _ := h.shows.EnsureSinglesShow(profile.ID, s.ID)
		singlesEpisodes, _ := h.episodes.FindByShow(singlesShow.ID)

		firstVideoID := ""
		if len(shows) > 0 {
			// Берем первый эпизод первого шоу
			eps, _ := h.episodes.FindByShow(shows[0].ID)
			if len(eps) > 0 {
				firstVideoID = eps[0].VideoID
			}
		} else if len(singlesEpisodes) > 0 {
			firstVideoID = singlesEpisodes[0].VideoID
		}

		result = append(result, SectionInfo{
			Section:      *s,
			ShowCount:    len(shows),
			EpisodeCount: len(singlesEpisodes),
			FirstVideoID: firstVideoID,
		})
	}

	c.JSON(http.StatusOK, result)
}

// UpdateSectionSettings godoc
// POST /sections/:id/settings
// Body: { "useThumb": true }
func (h *SectionHandler) UpdateSectionSettings(c *gin.Context) {
	profile := middleware.GetProfile(c)
	sectionID := c.Param("id")

	var body struct {
		UseThumb bool `json:"useThumb"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	section, err := h.sections.FindByID(sectionID)
	if err != nil || section == nil || section.OwnerID != profile.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if err := h.sections.UpdateSettings(sectionID, body.UseThumb); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
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

	// Загружаем скрытое шоу для одиночных видео
	singlesShow, err := h.shows.EnsureSinglesShow(profile.ID, sectionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// И его эпизоды
	singlesEpisodes, err := h.episodes.FindByShow(singlesShow.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"section":         section,
		"shows":           shows,
		"singlesShow":     singlesShow,
		"singlesEpisodes": singlesEpisodes,
	})
}

// AddSingleVideo godoc
// POST /sections/:id/episodes
// Body: { "url": "..." }
func (h *SectionHandler) AddSingleVideo(c *gin.Context) {
	profile := middleware.GetProfile(c)
	sectionID := c.Param("id")

	var body struct {
		URL string `json:"url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	section, err := h.sections.FindByID(sectionID)
	if err != nil || section == nil || section.OwnerID != profile.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	singlesShow, err := h.shows.EnsureSinglesShow(profile.ID, sectionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to ensure singles show: " + err.Error()})
		return
	}

	info, err := h.ytClient.FetchPlaylist(c.Request.Context(), body.URL)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch video info: " + err.Error()})
		return
	}

	maxOrderIndex, err := h.episodes.GetMaxOrderIndex(singlesShow.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get order index: " + err.Error()})
		return
	}

	episodes := make([]*models.Episode, 0, len(info.Entries))
	for i, entry := range info.Entries {
		episodes = append(episodes, &models.Episode{
			ShowID:     singlesShow.ID,
			VideoID:    entry.ID,
			Title:      entry.Title,
			Duration:   entry.Duration,
			OrderIndex: maxOrderIndex + 1 + i,
		})
	}

	if len(episodes) > 0 {
		if err := h.episodes.BulkCreate(episodes); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusCreated, gin.H{"episodes": episodes})
}

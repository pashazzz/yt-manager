package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/pavlo/yt-manager/internal/middleware"
	"github.com/pavlo/yt-manager/internal/models"
	"github.com/pavlo/yt-manager/internal/repository"
	"github.com/pavlo/yt-manager/internal/ytdlp"
)

// ShowHandler держит зависимости для хендлеров шоу.
type ShowHandler struct {
	shows    *repository.ShowRepo
	episodes *repository.EpisodeRepo
	sections *repository.SectionRepo
	ytClient *ytdlp.Client
}

func NewShowHandler(
	shows *repository.ShowRepo,
	episodes *repository.EpisodeRepo,
	sections *repository.SectionRepo,
	ytClient *ytdlp.Client,
) *ShowHandler {
	return &ShowHandler{shows: shows, episodes: episodes, sections: sections, ytClient: ytClient}
}

// CreateShow godoc
// POST /shows
// Body: { "playlistUrl": "...", "sectionId": "uuid" (опционально) }
func (h *ShowHandler) CreateShow(c *gin.Context) {
	profile := middleware.GetProfile(c)

	var body struct {
		PlaylistURL string `json:"playlistUrl" binding:"required"`
		SectionID   string `json:"sectionId"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Если раздел не указан — используем Default
	if body.SectionID == "" {
		defaultSec, err := h.sections.EnsureDefault(profile.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		body.SectionID = defaultSec.ID
	} else {
		// Проверяем, что раздел принадлежит профилю
		sec, err := h.sections.FindByID(body.SectionID)
		if err != nil || sec == nil || sec.OwnerID != profile.ID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sectionId"})
			return
		}
	}

	info, err := h.ytClient.FetchPlaylist(c.Request.Context(), body.PlaylistURL)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch playlist: " + err.Error()})
		return
	}

	show := &models.Show{
		Title:       info.Title,
		PlaylistURL: body.PlaylistURL,
		OwnerID:     profile.ID,
		SectionID:   body.SectionID,
	}
	if err := h.shows.Create(show); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	episodes := make([]*models.Episode, 0, len(info.Entries))
	for i, entry := range info.Entries {
		episodes = append(episodes, &models.Episode{
			ShowID:     show.ID,
			VideoID:    entry.ID,
			Title:      entry.Title,
			Duration:   entry.Duration,
			OrderIndex: i,
		})
	}
	if len(episodes) > 0 {
		if err := h.episodes.BulkCreate(episodes); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{"show": show, "episodeCount": len(episodes)})
}

// ListShows godoc — GET /shows (обратная совместимость, возвращает все шоу профиля)
func (h *ShowHandler) ListShows(c *gin.Context) {
	profile := middleware.GetProfile(c)
	shows, err := h.shows.FindByOwner(profile.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, shows)
}

// GetShow godoc — GET /shows/:id (шоу + его эпизоды)
func (h *ShowHandler) GetShow(c *gin.Context) {
	profile := middleware.GetProfile(c)
	showID := c.Param("id")

	show, err := h.shows.FindByID(showID)
	if err != nil || show == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "show not found"})
		return
	}
	if show.OwnerID != profile.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	episodes, err := h.episodes.FindByShow(showID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"show": show, "episodes": episodes})
}

// DeleteShow godoc — DELETE /shows/:id
func (h *ShowHandler) DeleteShow(c *gin.Context) {
	profile := middleware.GetProfile(c)
	showID := c.Param("id")

	show, err := h.shows.FindByID(showID)
	if err != nil || show == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "show not found"})
		return
	}
	if show.OwnerID != profile.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	_ = h.episodes.DeleteByShow(showID)
	if err := h.shows.Delete(showID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// MoveShow godoc
// PATCH /shows/:id/section
// Body: { "sectionId": "uuid" }
func (h *ShowHandler) MoveShow(c *gin.Context) {
	profile := middleware.GetProfile(c)
	showID := c.Param("id")

	var body struct {
		SectionID string `json:"sectionId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	show, err := h.shows.FindByID(showID)
	if err != nil || show == nil || show.OwnerID != profile.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Проверяем что целевой раздел принадлежит профилю
	sec, err := h.sections.FindByID(body.SectionID)
	if err != nil || sec == nil || sec.OwnerID != profile.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sectionId"})
		return
	}

	if err := h.shows.UpdateSection(showID, body.SectionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": showID, "sectionId": body.SectionID})
}

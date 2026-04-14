package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/pavlo/yt-manager/internal/middleware"
	"github.com/pavlo/yt-manager/internal/models"
	"github.com/pavlo/yt-manager/internal/repository"
)

// EpisodeHandler держит зависимости для хендлеров эпизодов.
type EpisodeHandler struct {
	episodes *repository.EpisodeRepo
	shows    *repository.ShowRepo
	tags     *repository.TagRepo
}

func NewEpisodeHandler(
	episodes *repository.EpisodeRepo,
	shows *repository.ShowRepo,
	tags *repository.TagRepo,
) *EpisodeHandler {
	return &EpisodeHandler{episodes: episodes, shows: shows, tags: tags}
}

// ListEpisodes godoc
// GET /episodes
func (h *EpisodeHandler) ListEpisodes(c *gin.Context) {
	profile := middleware.GetProfile(c)

	var eps []*models.Episode
	var err error

	// Все эпизоды владельца (пока без пагинации)
	eps, err = h.episodes.FindByOwnerID(profile.ID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, eps)
}

// SaveProgress godoc
// POST /episodes/:id/progress
// Body: { "currentTime": 123.4, "isWatched": false }
func (h *EpisodeHandler) SaveProgress(c *gin.Context) {
	profile := middleware.GetProfile(c)
	episodeID := c.Param("id")

	var body struct {
		CurrentTime float64 `json:"currentTime"`
		IsWatched   bool    `json:"isWatched"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем, что эпизод существует и принадлежит шоу текущего профиля.
	ep, err := h.episodes.FindByID(episodeID)
	if err != nil || ep == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "episode not found"})
		return
	}

	show, err := h.shows.FindByID(ep.ShowID)
	if err != nil || show == nil || show.OwnerID != profile.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if err := h.episodes.UpdateProgress(episodeID, body.CurrentTime, body.IsWatched); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          episodeID,
		"currentTime": body.CurrentTime,
		"isWatched":   body.IsWatched,
	})
}

// UpdateEpisodeTags godoc
// PATCH /episodes/:id/tags
// Body: { "tagIds": ["uuid1", "uuid2"] }
func (h *EpisodeHandler) UpdateEpisodeTags(c *gin.Context) {
	profile := middleware.GetProfile(c)
	episodeID := c.Param("id")

	var body struct {
		TagIDs []string `json:"tagIds" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ep, err := h.episodes.FindByID(episodeID)
	if err != nil || ep == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "episode not found"})
		return
	}

	show, err := h.shows.FindByID(ep.ShowID)
	if err != nil || show == nil || show.OwnerID != profile.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Проверяем теги
	for _, tid := range body.TagIDs {
		t, err := h.tags.FindByID(tid)
		if err != nil || t == nil || t.OwnerID != profile.ID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tagId: " + tid})
			return
		}
	}

	if err := h.episodes.UpdateTags(episodeID, body.TagIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": episodeID, "tagIds": body.TagIDs})
}

// DeleteEpisode godoc
// POST /episodes/:id/delete
func (h *EpisodeHandler) DeleteEpisode(c *gin.Context) {
	profile := middleware.GetProfile(c)
	episodeID := c.Param("id")

	ep, err := h.episodes.FindByID(episodeID)
	if err != nil || ep == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "episode not found"})
		return
	}

	show, err := h.shows.FindByID(ep.ShowID)
	if err != nil || show == nil || show.OwnerID != profile.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if err := h.episodes.Delete(episodeID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": episodeID, "message": "deleted"})
}

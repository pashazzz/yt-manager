package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/pavlo/yt-manager/internal/middleware"
	"github.com/pavlo/yt-manager/internal/repository"
)

// EpisodeHandler держит зависимости для хендлеров эпизодов.
type EpisodeHandler struct {
	episodes *repository.EpisodeRepo
	shows    *repository.ShowRepo
}

func NewEpisodeHandler(
	episodes *repository.EpisodeRepo,
	shows *repository.ShowRepo,
) *EpisodeHandler {
	return &EpisodeHandler{episodes: episodes, shows: shows}
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

// MoveEpisode godoc
// POST /episodes/:id/move
// Body: { "sectionId": "newSectionId" }
func (h *EpisodeHandler) MoveEpisode(c *gin.Context) {
	profile := middleware.GetProfile(c)
	episodeID := c.Param("id")

	var body struct {
		SectionID string `json:"sectionId"`
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

	targetShow, err := h.shows.EnsureSinglesShow(profile.ID, body.SectionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to ensure target singles show"})
		return
	}

	maxOrder, _ := h.episodes.GetMaxOrderIndex(targetShow.ID)

	if err := h.episodes.MoveEpisode(ep.ID, targetShow.ID, maxOrder+1); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

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

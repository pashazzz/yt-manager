package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/pavlo/yt-manager/internal/middleware"
	"github.com/pavlo/yt-manager/internal/models"
	"github.com/pavlo/yt-manager/internal/repository"
	"github.com/pavlo/yt-manager/internal/ytdlp"
)

// TagHandler держит зависимости для хендлеров тегов.
type TagHandler struct {
	tags     *repository.TagRepo
	shows    *repository.ShowRepo
	episodes *repository.EpisodeRepo
	ytClient *ytdlp.Client
}

func NewTagHandler(
	tags *repository.TagRepo,
	shows *repository.ShowRepo,
	episodes *repository.EpisodeRepo,
	ytClient *ytdlp.Client,
) *TagHandler {
	return &TagHandler{tags: tags, shows: shows, episodes: episodes, ytClient: ytClient}
}

type TagInfo struct {
	models.Tag
	ShowCount    int    `json:"showCount"`
	EpisodeCount int    `json:"episodeCount"`
	FirstVideoID string `json:"firstVideoId"`
}

// ListTags godoc
// GET /tags — возвращает все теги со статистикой.
func (h *TagHandler) ListTags(c *gin.Context) {
	profile := middleware.GetProfile(c)

	// Гарантируем существование дефолтного тега
	if _, err := h.tags.EnsureDefault(profile.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tags, err := h.tags.FindByOwner(profile.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Собираем статистику для каждого тега
	result := make([]TagInfo, 0, len(tags))
	for _, t := range tags {
		shows, _ := h.shows.FindByTag(profile.ID, t.ID, t.IsDefault)
		
		singlesEpisodes, _ := h.episodes.FindSinglesByTag(profile.ID, t.ID, t.IsDefault)
		
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

		result = append(result, TagInfo{
			Tag:          *t,
			ShowCount:    len(shows),
			EpisodeCount: len(singlesEpisodes),
			FirstVideoID: firstVideoID,
		})
	}

	c.JSON(http.StatusOK, result)
}

// UpdateTagSettings godoc
// POST /tags/:id/settings
// Body: { "useThumb": true }
func (h *TagHandler) UpdateTagSettings(c *gin.Context) {
	profile := middleware.GetProfile(c)
	tagID := c.Param("id")

	var body struct {
		UseThumb bool `json:"useThumb"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tag, err := h.tags.FindByID(tagID)
	if err != nil || tag == nil || tag.OwnerID != profile.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if err := h.tags.UpdateSettings(tagID, body.UseThumb); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

// CreateTag godoc
// POST /tags
// Body: { "name": "Аниме" }
func (h *TagHandler) CreateTag(c *gin.Context) {
	profile := middleware.GetProfile(c)

	var body struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	t := &models.Tag{Name: body.Name, OwnerID: profile.ID}
	if err := h.tags.Create(t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, t)
}

// DeleteTag godoc
// DELETE /tags/:id — нельзя удалить Default. Теги просто снимаются с элементов.
func (h *TagHandler) DeleteTag(c *gin.Context) {
	profile := middleware.GetProfile(c)
	tagID := c.Param("id")

	tag, err := h.tags.FindByID(tagID)
	if err != nil || tag == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tag not found"})
		return
	}
	if tag.OwnerID != profile.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	if tag.IsDefault {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete default tag"})
		return
	}

	// При удалении тега мы не удаляем шоу, просто тег исчезает из их списков.
	// Это произойдет автоматически, так как элементы просто перестанут находиться по этому тегу.
	if err := h.tags.Delete(tagID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ReorderTags godoc
// PATCH /tags/reorder
// Body: { "orderedIds": ["id1", "id2"] }
func (h *TagHandler) ReorderTags(c *gin.Context) {
	profile := middleware.GetProfile(c)

	var body struct {
		OrderedIDs []string `json:"orderedIds" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.tags.UpdateOrder(profile.ID, body.OrderedIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// ListItemsByTag godoc
// GET /tags/:id/items — контент для конкретного тега.
func (h *TagHandler) ListItemsByTag(c *gin.Context) {
	profile := middleware.GetProfile(c)
	tagID := c.Param("id")

	tag, err := h.tags.FindByID(tagID)
	if err != nil || tag == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tag not found"})
		return
	}
	if tag.OwnerID != profile.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	shows, err := h.shows.FindByTag(profile.ID, tagID, tag.IsDefault)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Загружаем скрытое шоу для одиночных видео этого тега
	singlesShow, err := h.shows.EnsureSinglesShow(profile.ID, tagID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// И его эпизоды (все синглы, отмеченные этим тегом, либо без тегов если тег дефолтный)
	singlesEpisodes, err := h.episodes.FindSinglesByTag(profile.ID, tagID, tag.IsDefault)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tag":             tag,
		"shows":           shows,
		"singlesShow":     singlesShow,
		"singlesEpisodes": singlesEpisodes,
	})
}

// AddSingleVideoToTag godoc
// POST /tags/:id/episodes
// Body: { "url": "..." }
func (h *TagHandler) AddSingleVideoToTag(c *gin.Context) {
	profile := middleware.GetProfile(c)
	tagID := c.Param("id")

	var body struct {
		URL string `json:"url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tag, err := h.tags.FindByID(tagID)
	if err != nil || tag == nil || tag.OwnerID != profile.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	singlesShow, err := h.shows.EnsureSinglesShow(profile.ID, tagID)
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
			TagIDs:     []string{tagID},
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

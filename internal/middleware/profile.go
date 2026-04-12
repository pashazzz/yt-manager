package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pavlo/yt-manager/internal/models"
	"github.com/pavlo/yt-manager/internal/repository"
)

const ProfileKey = "profile"

// Profile извлекает данные пользователя из Tailscale-заголовков,
// делает upsert профиля в БД и кладёт его в контекст запроса.
//
// Tailscale добавляет эти заголовки автоматически при serve/funnel:
//   Tailscale-User-Login  — e-mail (используется как ID)
//   Tailscale-User-Name   — отображаемое имя
func Profile(profileRepo *repository.ProfileRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		loginEmail := c.GetHeader("Tailscale-User-Login")
		if loginEmail == "" {
			// В dev-режиме подставляем заглушку, чтобы не ломать разработку
			// вне Tailscale. В продакшне здесь можно вернуть 401.
			loginEmail = "dev@local"
		}

		displayName := c.GetHeader("Tailscale-User-Name")
		if displayName == "" {
			displayName = loginEmail
		}

		profile := &models.Profile{
			ID:   loginEmail,
			Name: displayName,
		}

		if err := profileRepo.Upsert(profile); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "failed to upsert profile: " + err.Error(),
			})
			return
		}

		c.Set(ProfileKey, profile)
		c.Next()
	}
}

// GetProfile достаёт профиль из gin.Context (helper для хендлеров).
func GetProfile(c *gin.Context) *models.Profile {
	v, _ := c.Get(ProfileKey)
	p, _ := v.(*models.Profile)
	return p
}

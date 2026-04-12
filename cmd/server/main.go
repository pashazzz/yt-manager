package main

import (
	"context"
	iofs "io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gin-gonic/gin"

	"github.com/pavlo/yt-manager/internal/db"
	"github.com/pavlo/yt-manager/internal/handlers"
	"github.com/pavlo/yt-manager/internal/middleware"
	"github.com/pavlo/yt-manager/internal/repository"
	web "github.com/pavlo/yt-manager/internal/web"
	"github.com/pavlo/yt-manager/internal/ytdlp"
)

func main() {
	// --- конфигурация ---
	dataDir := envOr("DATA_DIR", "./data")
	addr := envOr("ADDR", ":8090")

	// --- база данных ---
	database, err := db.Open(dataDir)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	// --- репозитории ---
	profileRepo := repository.NewProfileRepo(database)
	showRepo := repository.NewShowRepo(database)
	episodeRepo := repository.NewEpisodeRepo(database)

	// --- yt-dlp клиент ---
	ctx := context.Background()
	ytClient, err := ytdlp.NewClient(ctx)
	if err != nil {
		log.Fatalf("failed to initialise yt-dlp client: %v", err)
	}

	// --- HTTP-роутер ---
	r := gin.Default()

	// CORS для dev-режима (Vite на :5173)
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Tailscale-User-Login, Tailscale-User-Name")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Middleware: профиль из Tailscale-заголовков
	r.Use(middleware.Profile(profileRepo))

	// --- API маршруты ---
	showHandler := handlers.NewShowHandler(showRepo, episodeRepo, ytClient)
	episodeHandler := handlers.NewEpisodeHandler(episodeRepo, showRepo)

	api := r.Group("/api/v1")
	{
		api.POST("/shows", showHandler.CreateShow)
		api.GET("/shows", showHandler.ListShows)
		api.GET("/shows/:id", showHandler.GetShow)
		api.DELETE("/shows/:id", showHandler.DeleteShow)
		api.POST("/episodes/:id/progress", episodeHandler.SaveProgress)
	}

	// --- Раздача фронтенда (SPA) ---
	// Embedded FS из internal/web/dist/ (собирается командой `make frontend`)
	webFS := web.FS()
	fileServer := http.FileServer(http.FS(webFS))

	r.NoRoute(func(c *gin.Context) {
		urlPath := c.Request.URL.Path

		// /api/... без совпадения с роутами → 404 JSON
		if strings.HasPrefix(urlPath, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "api route not found"})
			return
		}

		// Обрезаем ведущий слэш для fs.Stat
		fsPath := strings.TrimPrefix(urlPath, "/")

		// Если реальный файл существует (assets, favicon, …) — отдаём его
		if _, err := iofs.Stat(webFS, fsPath); err == nil {
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}

		// SPA fallback: любой «красивый» URL → index.html
		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
	})

	// --- Graceful shutdown ---
	srv := startServer(r, addr)
	log.Printf("yt-manager listening on %s (data: %s)", addr, dataDir)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

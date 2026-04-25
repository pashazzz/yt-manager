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
	"github.com/pavlo/yt-manager/internal/providers"
	"github.com/pavlo/yt-manager/internal/repository"
	web "github.com/pavlo/yt-manager/internal/web"
	"github.com/pavlo/yt-manager/internal/ytdlp"
)

func main() {
	dataDir := envOr("DATA_DIR", "./data")
	addr := envOr("ADDR", ":8090")

	// --- БД ---
	database, err := db.Open(dataDir)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	// --- Репозитории ---
	profileRepo := repository.NewProfileRepo(database)
	tagRepo := repository.NewTagRepo(database)
	showRepo := repository.NewShowRepo(database)
	episodeRepo := repository.NewEpisodeRepo(database)

	// --- yt-dlp ---
	ctx := context.Background()
	ytClient, err := ytdlp.NewClient(ctx)
	if err != nil {
		log.Fatalf("failed to initialise yt-dlp client: %v", err)
	}

	// --- Провайдеры (YouTube, Rutube, …) ---
	providerRegistry := providers.NewDefaultRegistry(ytClient)

	// Фоновая миграция: подтягиваем превью для эпизодов, добавленных
	// до появления поля thumbnailUrl (важно для не-YouTube провайдеров).
	go handlers.BackfillThumbnails(episodeRepo, providerRegistry)

	// --- Роутер ---
	r := gin.Default()

	// CORS для dev-окружения (Vite на :5173)
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Tailscale-User-Login, Tailscale-User-Name")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.Use(middleware.Profile(profileRepo))

	// --- Хендлеры ---
	showHandler := handlers.NewShowHandler(showRepo, episodeRepo, tagRepo, providerRegistry)
	episodeHandler := handlers.NewEpisodeHandler(episodeRepo, showRepo, tagRepo)
	tagHandler := handlers.NewTagHandler(tagRepo, showRepo, episodeRepo, providerRegistry)

	// --- Маршруты ---
	api := r.Group("/api/v1")
	{
		// Теги (бывш. Разделы)
		api.GET("/tags", tagHandler.ListTags)
		api.POST("/tags", tagHandler.CreateTag)
		api.POST("/tags/:id/delete", tagHandler.DeleteTag)
		api.GET("/tags/:id/items", tagHandler.ListItemsByTag)
		api.POST("/tags/reorder", tagHandler.ReorderTags)
		api.POST("/tags/:id/settings", tagHandler.UpdateTagSettings)
		api.POST("/tags/:id/episodes", tagHandler.AddSingleVideoToTag)

		// Шоу
		api.POST("/shows", showHandler.CreateShow)
		api.POST("/shows/reorder", showHandler.ReorderShows)
		api.GET("/shows", showHandler.ListShows)
		api.GET("/shows/:id", showHandler.GetShow)
		api.POST("/shows/:id/delete", showHandler.DeleteShow)
		api.POST("/shows/:id/tags", showHandler.UpdateShowTags)
		api.POST("/shows/:id/reverse", showHandler.ReverseShow)
		api.POST("/shows/:id/episodes", showHandler.AddEpisode)
		api.POST("/shows/:id/episodes/reorder", showHandler.ReorderEpisodes)

		// Эпизоды
		api.POST("/episodes/:id/progress", episodeHandler.SaveProgress)
		api.POST("/episodes/:id/tags", episodeHandler.UpdateEpisodeTags)
		api.POST("/episodes/:id/delete", episodeHandler.DeleteEpisode)
		api.GET("/episodes", episodeHandler.ListEpisodes)
	}

	// --- SPA: раздача фронтенда ---
	webFS := web.FS()
	fileServer := http.FileServer(http.FS(webFS))

	r.NoRoute(func(c *gin.Context) {
		urlPath := c.Request.URL.Path
		if strings.HasPrefix(urlPath, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "api route not found"})
			return
		}
		fsPath := strings.TrimPrefix(urlPath, "/")
		if _, err := iofs.Stat(webFS, fsPath); err == nil {
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}
		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
	})

	// --- Запуск ---
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

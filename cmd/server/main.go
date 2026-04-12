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
	sectionRepo := repository.NewSectionRepo(database)
	showRepo := repository.NewShowRepo(database)
	episodeRepo := repository.NewEpisodeRepo(database)

	// --- yt-dlp ---
	ctx := context.Background()
	ytClient, err := ytdlp.NewClient(ctx)
	if err != nil {
		log.Fatalf("failed to initialise yt-dlp client: %v", err)
	}

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
	showHandler := handlers.NewShowHandler(showRepo, episodeRepo, sectionRepo, ytClient)
	episodeHandler := handlers.NewEpisodeHandler(episodeRepo, showRepo)
	sectionHandler := handlers.NewSectionHandler(sectionRepo, showRepo, episodeRepo, ytClient)

	// --- Маршруты ---
	api := r.Group("/api/v1")
	{
		// Разделы
		api.GET("/sections", sectionHandler.ListSections)
		api.POST("/sections", sectionHandler.CreateSection)
		api.POST("/sections/:id/delete", sectionHandler.DeleteSection)
		api.GET("/sections/:id/shows", sectionHandler.ListShowsBySection)
		api.POST("/sections/reorder", sectionHandler.ReorderSections)
		api.POST("/sections/:id/episodes", sectionHandler.AddSingleVideo)

		// Шоу
		api.POST("/shows", showHandler.CreateShow)
		api.GET("/shows", showHandler.ListShows)
		api.GET("/shows/:id", showHandler.GetShow)
		api.POST("/shows/:id/delete", showHandler.DeleteShow)
		api.POST("/shows/:id/section", showHandler.MoveShow)
		api.POST("/shows/:id/reverse", showHandler.ReverseShow)
		api.POST("/shows/:id/episodes", showHandler.AddEpisode)
		api.POST("/shows/:id/episodes/reorder", showHandler.ReorderEpisodes)

		// Эпизоды
		api.POST("/episodes/:id/progress", episodeHandler.SaveProgress)
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

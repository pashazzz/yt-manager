package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	"github.com/pavlo/yt-manager/internal/db"
	"github.com/pavlo/yt-manager/internal/handlers"
	"github.com/pavlo/yt-manager/internal/middleware"
	"github.com/pavlo/yt-manager/internal/repository"
	"github.com/pavlo/yt-manager/internal/ytdlp"
)

func main() {
	// --- конфигурация через ENV ---
	dataDir := envOr("DATA_DIR", "./data")
	addr := envOr("ADDR", ":8080")

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

	// Глобальный CORS для dev-окружения
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Tailscale-User-Login, Tailscale-User-Name")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Middleware: извлечь/создать профиль из Tailscale-заголовков
	r.Use(middleware.Profile(profileRepo))

	// --- хендлеры ---
	showHandler := handlers.NewShowHandler(showRepo, episodeRepo, ytClient)
	episodeHandler := handlers.NewEpisodeHandler(episodeRepo, showRepo)

	// --- маршруты ---
	api := r.Group("/api/v1")
	{
		api.POST("/shows", showHandler.CreateShow)
		api.GET("/shows", showHandler.ListShows)
		api.GET("/shows/:id", showHandler.GetShow)
		api.DELETE("/shows/:id", showHandler.DeleteShow)

		api.POST("/episodes/:id/progress", episodeHandler.SaveProgress)
	}

	// --- запуск сервера с graceful shutdown ---
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

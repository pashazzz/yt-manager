package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// startServer запускает HTTP-сервер в фоне и возвращает *http.Server
// для последующего graceful shutdown.
func startServer(r *gin.Engine, addr string) *http.Server {
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic("server error: " + err.Error())
		}
	}()
	return srv
}

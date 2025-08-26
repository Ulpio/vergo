// cmd/api/main.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	// importa o router do projeto
	"github.com/Ulpio/vergo/internal/http/router"
)

func main() {
	port := getEnvInt("APP_PORT", 8080)
	env := getEnv("APP_ENV", "dev")
	version := getEnv("APP_VERSION", "0.1.0")

	if env != "dev" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())

	// Health endpoints
	startedAt := time.Now()
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"uptime":  time.Since(startedAt).String(),
			"version": version,
			"env":     env,
		})
	})
	r.GET("/readyz", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Grupo da API v1
	api := r.Group("/v1")
	{
		// rota simples de ping
		api.GET("/_ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"pong": true})
		})

		// registra todos os endpoints stubs
		router.Register(api)
	}

	// servidor HTTP com shutdown gracioso
	srv := &http.Server{
		Addr:         ":" + strconv.Itoa(port),
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("vergo listening on :%d (env=%s, version=%s)", port, env, version)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced to shutdown: %v", err)
	}
	log.Println("bye!")
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

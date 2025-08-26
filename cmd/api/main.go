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
)

func main() {
	// Configs básicas via env
	port := getEnvInt("APP_PORT", 8080)
	env := getEnv("APP_ENV", "dev")
	version := getEnv("APP_VERSION", "0.1.0")

	// Gin em modo release quando não for dev
	if env != "dev" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())

	// Logging simples (você pode trocar por slog/zap depois)
	router.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		lat := time.Since(start)
		log.Printf("%s %s -> %d (%s)",
			c.Request.Method, c.Request.URL.Path, c.Writer.Status(), lat)
	})

	// Endpoints básicos
	startedAt := time.Now()

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"uptime":  time.Since(startedAt).String(),
			"version": version,
			"env":     env,
		})
	})
	router.GET("/readyz", func(c *gin.Context) {
		// aqui dá pra checar DB, fila, etc. Por enquanto: pronto.
		c.Status(http.StatusOK)
	})

	// Namespace da API (v1). Adicione suas rotas aqui depois.
	api := router.Group("/v1")
	{
		api.GET("/_ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"pong": true})
		})
	}

	// Servidor + shutdown gracioso
	srv := &http.Server{
		Addr:         ":" + strconv.Itoa(port),
		Handler:      router,
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

	// Espera por SIGINT/SIGTERM para desligar com graça
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

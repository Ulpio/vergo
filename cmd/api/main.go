// cmd/api/main.go
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/Ulpio/vergo/internal/http/middleware"
	"github.com/Ulpio/vergo/internal/http/router"
	"github.com/Ulpio/vergo/internal/pkg/config"
	"github.com/Ulpio/vergo/internal/pkg/db"
	"github.com/Ulpio/vergo/internal/pkg/logging"
	"github.com/Ulpio/vergo/internal/pkg/telemetry"
)

func init() {
	if err := godotenv.Load(); err != nil {
		_ = err
	}
}

func main() {
	cfg := config.Load()
	port := cfg.AppPort
	env := cfg.AppEnv
	version := cfg.AppVersion

	// Initialize structured logger
	logger := logging.New(env)
	slog.SetDefault(logger)

	// Initialize OpenTelemetry (TracerProvider + MeterProvider)
	ctx := context.Background()
	shutdownTelemetry, err := telemetry.Init(ctx, telemetry.Config{
		ServiceName:    "vergo",
		ServiceVersion: version,
	})
	if err != nil {
		slog.Error("telemetry init failed", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := shutdownTelemetry(context.Background()); err != nil {
			slog.Error("telemetry shutdown failed", "error", err)
		}
	}()

	// Connect to database
	slog.Info("connecting to database")
	database, err := db.Open(cfg)
	if err != nil {
		slog.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	if err := database.Ping(); err != nil {
		slog.Error("database ping failed", "error", err)
		os.Exit(1)
	}
	slog.Info("database connection established")

	// Run migrations
	slog.Info("running migrations")
	if err := db.RunMigrations(database); err != nil {
		slog.Error("migrations failed", "error", err)
		os.Exit(1)
	}

	if env != "dev" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(middleware.Recover())
	r.Use(otelgin.Middleware("vergo"))
	r.Use(middleware.Logging())

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

	// API v1
	api := r.Group("/v1")
	{
		api.GET("/_ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"pong": true})
		})

		router.Register(api)
	}

	// HTTP server with graceful shutdown
	srv := &http.Server{
		Addr:         ":" + strconv.Itoa(port),
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server started", "port", port, "env", env, "version", version)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("forced shutdown", "error", err)
		os.Exit(1)
	}
	slog.Info("server stopped")
}

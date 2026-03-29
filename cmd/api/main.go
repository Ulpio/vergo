// cmd/api/main.go
package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/Ulpio/vergo/internal/http/middleware"
	"github.com/Ulpio/vergo/internal/http/router"
	"github.com/Ulpio/vergo/internal/pkg/config"
	"github.com/Ulpio/vergo/internal/pkg/db"
	"github.com/Ulpio/vergo/internal/pkg/logging"
	"github.com/Ulpio/vergo/internal/pkg/ratelimit"
	"github.com/Ulpio/vergo/internal/pkg/telemetry"
)

const shutdownTimeout = 30 * time.Second

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
	otelResult, err := telemetry.Init(context.Background(), telemetry.Config{
		ServiceName:       "vergo",
		ServiceVersion:    version,
		OTLPEndpoint:      cfg.OTLPEndpoint,
		OTLPInsecure:      cfg.OTLPInsecure,
		PrometheusEnabled: cfg.MetricsPort > 0,
	})
	if err != nil {
		slog.Error("telemetry init failed", "error", err)
		os.Exit(1)
	}
	if cfg.OTLPEndpoint != "" {
		slog.Info("otlp exporter enabled", "endpoint", cfg.OTLPEndpoint)
	}

	// Connect to database
	slog.Info("connecting to database")
	database, err := db.Open(cfg)
	if err != nil {
		slog.Error("database connection failed", "error", err)
		os.Exit(1)
	}

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

	// Rate limiter (in-memory token bucket)
	limiter := ratelimit.New(float64(cfg.RateLimitRPS), cfg.RateLimitBurst)
	defer limiter.Stop()

	r := gin.New()
	r.Use(middleware.Recover())
	r.Use(otelgin.Middleware("vergo"))
	r.Use(middleware.RateLimit(limiter))
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

	// Prometheus metrics server (separate port for scraping)
	if cfg.MetricsPort > 0 && otelResult.PrometheusHandler != nil {
		metricsMux := http.NewServeMux()
		metricsMux.Handle("/metrics", promhttp.Handler())
		metricsSrv := &http.Server{
			Addr:         ":" + strconv.Itoa(cfg.MetricsPort),
			Handler:      metricsMux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
		}
		go func() {
			slog.Info("metrics server started", "port", cfg.MetricsPort)
			if err := metricsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("metrics server error", "error", err)
			}
		}()
	}

	// HTTP server
	srv := &http.Server{
		Addr:         ":" + strconv.Itoa(port),
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server started", "port", port, "env", env, "version", version)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	slog.Info("shutdown signal received", "signal", sig.String())

	// Graceful shutdown: HTTP → telemetry → DB (ordered, not deferred)
	gracefulShutdown(srv, otelResult.Shutdown, database)
}

// gracefulShutdown drains in-flight requests, flushes telemetry, and closes
// the database connection pool in a deterministic order.
func gracefulShutdown(srv *http.Server, shutdownTelemetry func(context.Context) error, database *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// 1. Drain HTTP connections (in-flight requests complete up to timeout)
	slog.Info("draining http connections", "timeout", shutdownTimeout.String())
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("http shutdown error", "error", err)
	} else {
		slog.Info("http server stopped")
	}

	// 2. Flush telemetry spans and metrics
	slog.Info("flushing telemetry")
	if err := shutdownTelemetry(ctx); err != nil {
		slog.Error("telemetry flush error", "error", err)
	} else {
		slog.Info("telemetry flushed")
	}

	// 3. Close database connection pool
	slog.Info("closing database connections")
	if err := database.Close(); err != nil {
		slog.Error("database close error", "error", err)
	} else {
		slog.Info("database connections closed")
	}

	slog.Info("shutdown complete")
}

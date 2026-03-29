package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// Logging returns a middleware that logs each HTTP request with structured
// fields: method, path, status, latency, and client_ip.
// Log level is Info for 2xx/3xx, Warn for 4xx, Error for 5xx.
// trace_id and span_id are injected automatically by the otelHandler via context.
func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		status := c.Writer.Status()
		latency := time.Since(start)
		ctx := c.Request.Context()

		attrs := []slog.Attr{
			slog.String("method", method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("client_ip", c.ClientIP()),
		}

		level := slog.LevelInfo
		if status >= 500 {
			level = slog.LevelError
		} else if status >= 400 {
			level = slog.LevelWarn
		}

		slog.LogAttrs(ctx, level, "http request", attrs...)
	}
}

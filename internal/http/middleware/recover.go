package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// Recover returns a middleware that recovers from panics, logs a structured
// error with stack trace, and returns a generic 500 JSON response.
// It replaces gin.Recovery() with structured logging via slog.
func Recover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := string(debug.Stack())
				ctx := c.Request.Context()

				slog.ErrorContext(ctx, "panic recovered",
					slog.String("error", fmt.Sprint(err)),
					slog.String("method", c.Request.Method),
					slog.String("path", c.Request.URL.Path),
					slog.String("stack", stack),
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "internal_server_error",
				})
			}
		}()
		c.Next()
	}
}

package middleware

import (
	"net/http"
	"strings"

	"github.com/Ulpio/vergo/internal/auth"
	"github.com/Ulpio/vergo/internal/pkg/config"
	"github.com/gin-gonic/gin"
)

const ctxUserID = "user_id"

func Auth(cfg config.Config) gin.HandlerFunc {
	const prefix = "bearer "
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(strings.ToLower(h), prefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing_bearer"})
			return
		}
		token := strings.TrimSpace(h[len(prefix):])
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing_bearer"})
			return
		}
		claims, err := auth.Parse(token, cfg.JWTAccessSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
			return
		}
		c.Set(ctxUserID, claims.UserID)
		c.Next()
	}
}

func UserID(c *gin.Context) (string, bool) {
	v, ok := c.Get(ctxUserID)
	if !ok {
		return "", false
	}
	id, _ := v.(string)
	return id, id != ""
}

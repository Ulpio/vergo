package middleware

import (
	"net/http"
	"strings"

	"github.com/Ulpio/vergo/internal/auth"
	"github.com/Ulpio/vergo/internal/domain/apikey"
	"github.com/Ulpio/vergo/internal/pkg/config"
	"github.com/gin-gonic/gin"
)

const ctxUserID = "user_id"
const ctxAPIKeyAuth = "api_key_auth"

func Auth(cfg config.Config) gin.HandlerFunc {
	return AuthWithAPIKeys(cfg, nil)
}

func AuthWithAPIKeys(cfg config.Config, keySvc apikey.Service) gin.HandlerFunc {
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

		// API Key authentication (sk_...)
		if strings.HasPrefix(token, "sk_") && keySvc != nil {
			result, err := keySvc.Validate(token)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "key_validation_failed"})
				return
			}
			if result == nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_api_key"})
				return
			}
			c.Set(ctxUserID, result.KeyID)
			c.Set(ctxOrgID, result.OrgID)
			c.Set(ctxAPIKeyAuth, true)
			c.Next()
			return
		}

		// JWT authentication
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

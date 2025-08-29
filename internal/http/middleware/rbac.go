package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const ctxRole = "role"

var roleOrder = map[string]int{
	"member": 1,
	"admin":  2,
	"owner":  3,
}

func RequireRole(minRole string) gin.HandlerFunc {
	min := roleOrder[minRole]
	return func(c *gin.Context) {
		v, ok := c.Get(ctxRole)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		role, _ := v.(string)
		if roleOrder[role] < min {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient_role"})
			return
		}
		c.Next()
	}
}

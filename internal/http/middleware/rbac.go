package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const ctxRole = "role"

func RequireRole(minRole string) gin.HandlerFunc {
	order := map[string]int{"member": 1, "admin": 2, "owner": 3}

	return func(c *gin.Context) {
		roleVal, ok := c.Get(ctxRole)
		if !ok {
			c.Next()
			return
		}

		role, _ := roleVal.(string)
		if !ok {
			c.Next()
			return
		}
		if order[role] < order[minRole] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}

package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const ctxOrgID = "org_id"

func Tenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		orgID := strings.TrimSpace(c.GetHeader("X-Org-ID"))
		if orgID == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing_org_id"})
			return
		}
		c.Set(ctxOrgID, orgID)
		c.Next()
	}
}

func OrgID(c *gin.Context) (string, bool) {
	v, ok := c.Get(ctxOrgID)
	if !ok {
		return "", false
	}
	id, _ := v.(string)
	return id, id != ""
}

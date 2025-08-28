package middleware

import (
	"net/http"
	"strings"

	"github.com/Ulpio/vergo/internal/domain/org"
	"github.com/gin-gonic/gin"
)

const ctxOrgID = "org_id"

func Tenant(orgSvc org.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		orgID := strings.TrimSpace(c.GetHeader("X-Org-ID"))
		if orgID == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing_org_id"})
			return
		}
		uid, ok := UserID(c)
		if !ok || uid == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing_user"})
			return
		}
		okMember, role, err := orgSvc.IsMember(orgID, uid)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "tenant_check_failed"})
			return
		}
		if !okMember {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "not_a_member"})
			return
		}
		c.Set(ctxOrgID, orgID)
		if role != "" {
			c.Set(ctxRole, role) // para o RBAC opcional
		}
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

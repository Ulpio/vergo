package middleware

import (
	"net/http"
	"strings"

	"github.com/Ulpio/vergo/internal/domain/org"
	"github.com/Ulpio/vergo/internal/domain/userctx"
	"github.com/gin-gonic/gin"
)

const ctxOrgID = "org_id"

func Tenant(orgSvc org.Service, ctxSvc userctx.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, ok := UserID(c)
		if !ok || uid == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing_user"})
			return
		}
		// 1) tenta header
		orgID := strings.TrimSpace(c.GetHeader("X-Org-ID"))
		// 2) fallback para contexto persistido
		if orgID == "" && ctxSvc != nil {
			if id, ok2, _ := ctxSvc.GetActiveOrg(uid); ok2 {
				orgID = id
			}
		}
		if orgID == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing_org_id"})
			return
		}
		okM, role, err := orgSvc.IsMember(orgID, uid)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "tenant_check_failed"})
			return
		}
		if !okM {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "not_a_member"})
			return
		}
		c.Set(ctxOrgID, orgID)
		if role != "" {
			c.Set(ctxRole, role)
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

package handlers

import (
	"net/http"

	"github.com/Ulpio/vergo/internal/domain/org"
	"github.com/Ulpio/vergo/internal/domain/userctx"
	"github.com/Ulpio/vergo/internal/http/middleware"
	"github.com/gin-gonic/gin"
)

type ContextHandler struct {
	cs userctx.Service
	os org.Service
}

func NewContextHandler(cs userctx.Service, os org.Service) *ContextHandler {
	return &ContextHandler{cs: cs, os: os}
}

func (h *ContextHandler) Get(c *gin.Context) {
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing_user"})
	}
	orgID, ok, err := h.cs.GetActiveOrg(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "get_failed", "details": err.Error()})
		return
	}
	if !ok {
		c.Status(http.StatusNoContent)
		return
	}
	role := ""
	if okM, r, _ := h.os.IsMember(orgID, uid); okM {
		role = r
	}
	c.JSON(http.StatusOK, gin.H{"org_id": orgID, "role": role})
}

type setContectIn struct {
	OrgID string `json:"org_id" binding:"required"`
}

func (h *ContextHandler) Set(c *gin.Context) {
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing_user"})
		return
	}
	var in setContectIn
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid_payload"})
		return
	}
	okM, _, err := h.os.IsMember(in.OrgID, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "checkfailed", "detail": err.Error()})
		return
	}
	if !okM {
		c.JSON(http.StatusForbidden, gin.H{"error": "not_a_member"})
		return
	}
	if err := h.cs.SetActiveOrg(uid, in.OrgID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "set_failed", "detail": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

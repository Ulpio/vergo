package handlers

import (
	"net/http"

	"github.com/Ulpio/vergo/internal/domain/org"
	"github.com/Ulpio/vergo/internal/domain/user"
	"github.com/Ulpio/vergo/internal/http/middleware"
	"github.com/gin-gonic/gin"
)

type MeHandler struct {
	us user.Service
	os org.Service
}

func NewMeHandler(us user.Service, os org.Service) *MeHandler { return &MeHandler{us: us, os: os} }

// Get returns the authenticated user's profile.
// @Summary Get current user
// @Tags User
// @Security BearerAuth
// @Produce json
// @Param X-Org-ID header string false "Organization ID (optional, adds role)"
// @Success 200 {object} MeResponse
// @Failure 401 {object} ErrorResponse
// @Router /me [get]
func (h *MeHandler) Get(c *gin.Context) {
	uid, ok := middleware.UserID(c)
	if !ok || uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing_user"})
		return
	}
	u, err := h.us.GetByID(uid)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_user"})
		return
	}
	// Opcional: se vier X-Org-ID, retornamos também a role do membership
	orgID := c.GetHeader("X-Org-ID")
	out := gin.H{
		"id":    u.ID,
		"email": u.Email,
	}
	if orgID != "" && h.os != nil {
		if ok, role, _ := h.os.IsMember(orgID, uid); ok {
			out["org_id"] = orgID
			out["role"] = role
		}
	}
	c.JSON(http.StatusOK, gin.H{"user": out})
}

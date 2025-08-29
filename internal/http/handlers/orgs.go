package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Ulpio/vergo/internal/domain/org"
	"github.com/Ulpio/vergo/internal/http/middleware"
)

type OrgsHandler struct {
	os org.Service
}

func NewOrgsHandler(os org.Service) *OrgsHandler { return &OrgsHandler{os: os} }

type createOrgIn struct {
	Name string `json:"name" binding:"required"`
}

func (h *OrgsHandler) Create(c *gin.Context) {
	uid, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing_user"})
		return
	}
	var in createOrgIn
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid_payload"})
		return
	}
	o, err := h.os.Create(in.Name, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create_failed"})
		return
	}
	c.JSON(http.StatusCreated, o)
}

func (h *OrgsHandler) Get(c *gin.Context) {
	id := c.Param("id")
	o, err := h.os.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	c.JSON(http.StatusOK, o)
}

type memberIn struct {
	UserID string `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required"` // owner|admin|member
}

func (h *OrgsHandler) AddMember(c *gin.Context) {
	orgID := c.Param("id")
	var in memberIn
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid_payload"})
		return
	}
	if err := h.os.AddMember(orgID, in.UserID, in.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "add_failed"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *OrgsHandler) UpdateMember(c *gin.Context) {
	orgID := c.Param("id")
	userID := c.Param("userID")
	var in struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid_payload"})
		return
	}
	if err := h.os.UpdateMember(orgID, userID, in.Role); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "updated_failed", "detai": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *OrgsHandler) RemoveMember(c *gin.Context) {
	orgID := c.Param("id")
	userID := c.Param("userID")

	if err := h.os.RemoveMember(orgID, userID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "failed_remove_member", "detail": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *OrgsHandler) Delete(c *gin.Context) {
	orgID := c.Param("id")
	if err := h.os.Delete(orgID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "failed_delete_organization", "detail": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

package handlers

import (
	"net/http"
	"time"

	"github.com/Ulpio/vergo/internal/domain/audit"
	"github.com/Ulpio/vergo/internal/domain/project"
	"github.com/Ulpio/vergo/internal/http/middleware"
	"github.com/gin-gonic/gin"
)

type ProjectsHandler struct {
	ps project.Service
	as audit.Service
}

func NewProjectsHandler(ps project.Service, as audit.Service) *ProjectsHandler {
	return &ProjectsHandler{ps: ps, as: as}
}

func (h *ProjectsHandler) List(c *gin.Context) {
	orgID, ok := middleware.OrgID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_org_id"})
		return
	}
	items, err := h.ps.List(orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list_failed", "detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

type ProjectIn struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

func (h *ProjectsHandler) Create(c *gin.Context) {
	orgID, ok := middleware.OrgID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_org_id"})
		return
	}
	userID, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing_user"})
		return
	}
	var in ProjectIn
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid_payload"})
		return
	}
	p, err := h.ps.Create(orgID, in.Name, in.Description, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create_fail", "detail": err.Error()})
		return
	}
	_ = h.as.Record(audit.Event{
		OrgID:     orgID,
		ActorID:   userID,
		Action:    "project.created",
		Entity:    "project",
		EntityID:  p.ID,
		Timestamp: time.Now(),
	})
	c.JSON(http.StatusCreated, p)
}

func (h *ProjectsHandler) Get(c *gin.Context) {
	orgID, ok := middleware.OrgID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_org_id"})
		return
	}
	id := c.Param("id")
	p, err := h.ps.Get(orgID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *ProjectsHandler) Update(c *gin.Context) {
	orgID, ok := middleware.OrgID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_org_id"})
		return
	}
	id := c.Param("id")
	var in ProjectIn
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid_payload"})
		return
	}
	p, err := h.ps.Update(orgID, id, in.Name, in.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "not_found", "detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *ProjectsHandler) Delete(c *gin.Context) {
	orgID, ok := middleware.OrgID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_org_id"})
		return
	}
	userID, _ := middleware.UserID(c) // só pra auditoria
	id := c.Param("id")
	if err := h.ps.Delete(orgID, id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	_ = h.as.Record(audit.Event{
		OrgID:     orgID,
		ActorID:   userID,
		Action:    "project.deleted",
		Entity:    "project",
		EntityID:  id,
		Timestamp: time.Now(),
	})
	c.Status(http.StatusNoContent)
}

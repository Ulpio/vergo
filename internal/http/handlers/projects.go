package handlers

import (
	"encoding/json"
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

// List returns all projects in the organization.
// @Summary List projects
// @Tags Projects
// @Security BearerAuth
// @Produce json
// @Param X-Org-ID header string true "Organization ID"
// @Success 200 {object} map[string][]project.Project "items"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorDetailResponse
// @Router /projects [get]
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

// Create creates a new project.
// @Summary Create project
// @Tags Projects
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param X-Org-ID header string true "Organization ID"
// @Param body body ProjectIn true "Project data"
// @Success 201 {object} project.Project
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Failure 500 {object} ErrorDetailResponse
// @Router /projects [post]
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
	after, _ := json.Marshal(p)
	_ = h.as.Record(audit.Event{
		OrgID: orgID, ActorID: userID, Action: "project.created",
		Entity: "project", EntityID: p.ID, Timestamp: time.Now(),
		Metadata: audit.Metadata{After: after},
	})
	c.JSON(http.StatusCreated, p)
}

// Get returns a project by ID.
// @Summary Get project
// @Tags Projects
// @Security BearerAuth
// @Produce json
// @Param X-Org-ID header string true "Organization ID"
// @Param id path string true "Project ID"
// @Success 200 {object} project.Project
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /projects/{id} [get]
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

// Update modifies a project.
// @Summary Update project
// @Tags Projects
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param X-Org-ID header string true "Organization ID"
// @Param id path string true "Project ID"
// @Param body body ProjectIn true "Updated project data"
// @Success 200 {object} project.Project
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Failure 500 {object} ErrorDetailResponse
// @Router /projects/{id} [patch]
func (h *ProjectsHandler) Update(c *gin.Context) {
	orgID, ok := middleware.OrgID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_org_id"})
		return
	}
	userID, _ := middleware.UserID(c)
	id := c.Param("id")

	// Capture before state
	old, _ := h.ps.Get(orgID, id)
	before, _ := json.Marshal(old)

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

	after, _ := json.Marshal(p)
	_ = h.as.Record(audit.Event{
		OrgID: orgID, ActorID: userID, Action: "project.updated",
		Entity: "project", EntityID: id, Timestamp: time.Now(),
		Metadata: audit.Metadata{Before: before, After: after},
	})

	c.JSON(http.StatusOK, p)
}

// Delete removes a project.
// @Summary Delete project
// @Tags Projects
// @Security BearerAuth
// @Produce json
// @Param X-Org-ID header string true "Organization ID"
// @Param id path string true "Project ID"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /projects/{id} [delete]
func (h *ProjectsHandler) Delete(c *gin.Context) {
	orgID, ok := middleware.OrgID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_org_id"})
		return
	}
	userID, _ := middleware.UserID(c)
	id := c.Param("id")

	// Capture before state
	old, _ := h.ps.Get(orgID, id)
	before, _ := json.Marshal(old)

	if err := h.ps.Delete(orgID, id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}

	_ = h.as.Record(audit.Event{
		OrgID: orgID, ActorID: userID, Action: "project.deleted",
		Entity: "project", EntityID: id, Timestamp: time.Now(),
		Metadata: audit.Metadata{Before: before},
	})
	c.Status(http.StatusNoContent)
}

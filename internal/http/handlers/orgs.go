package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Ulpio/vergo/internal/domain/audit"
	"github.com/Ulpio/vergo/internal/domain/org"
	"github.com/Ulpio/vergo/internal/http/middleware"
)

type OrgsHandler struct {
	os org.Service
	as audit.Service
}

func NewOrgsHandler(os org.Service, as audit.Service) *OrgsHandler {
	return &OrgsHandler{os: os, as: as}
}

type createOrgIn struct {
	Name string `json:"name" binding:"required"`
}

// Create creates a new organization.
// @Summary Create organization
// @Tags Organizations
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body createOrgIn true "Organization name"
// @Success 201 {object} org.Organization
// @Failure 401 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /orgs [post]
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

	after, _ := json.Marshal(o)
	_ = h.as.Record(audit.Event{
		OrgID: o.ID, ActorID: uid, Action: "org.created",
		Entity: "org", EntityID: o.ID, Timestamp: time.Now(),
		Metadata: audit.Metadata{After: after},
	})

	c.JSON(http.StatusCreated, o)
}

// Get returns an organization by ID.
// @Summary Get organization
// @Tags Organizations
// @Security BearerAuth
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {object} org.Organization
// @Failure 404 {object} ErrorResponse
// @Router /orgs/{id} [get]
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

// AddMember adds a user to an organization.
// @Summary Add member
// @Tags Organizations
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param X-Org-ID header string true "Organization ID"
// @Param id path string true "Organization ID"
// @Param body body memberIn true "Member details"
// @Success 204 "No Content"
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /orgs/{id}/members [post]
func (h *OrgsHandler) AddMember(c *gin.Context) {
	orgID := c.Param("id")
	actorID, _ := middleware.UserID(c)
	var in memberIn
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid_payload"})
		return
	}
	if err := h.os.AddMember(orgID, in.UserID, in.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "add_failed"})
		return
	}

	after, _ := json.Marshal(gin.H{"user_id": in.UserID, "role": in.Role})
	_ = h.as.Record(audit.Event{
		OrgID: orgID, ActorID: actorID, Action: "member.added",
		Entity: "membership", EntityID: in.UserID, Timestamp: time.Now(),
		Metadata: audit.Metadata{After: after},
	})

	c.Status(http.StatusNoContent)
}

// UpdateMember changes a member's role.
// @Summary Update member role
// @Tags Organizations
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param X-Org-ID header string true "Organization ID"
// @Param id path string true "Organization ID"
// @Param userId path string true "User ID"
// @Param body body memberIn true "New role"
// @Success 204 "No Content"
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorDetailResponse
// @Failure 422 {object} ErrorResponse
// @Router /orgs/{id}/members/{userId} [patch]
func (h *OrgsHandler) UpdateMember(c *gin.Context) {
	orgID := c.Param("id")
	userID := c.Param("userID")
	actorID, _ := middleware.UserID(c)
	var in struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid_payload"})
		return
	}
	if err := h.os.UpdateMember(orgID, userID, in.Role); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "updated_failed", "detail": err.Error()})
		return
	}

	after, _ := json.Marshal(gin.H{"user_id": userID, "role": in.Role})
	_ = h.as.Record(audit.Event{
		OrgID: orgID, ActorID: actorID, Action: "member.updated",
		Entity: "membership", EntityID: userID, Timestamp: time.Now(),
		Metadata: audit.Metadata{After: after},
	})

	c.Status(http.StatusNoContent)
}

// RemoveMember removes a user from an organization.
// @Summary Remove member
// @Tags Organizations
// @Security BearerAuth
// @Produce json
// @Param X-Org-ID header string true "Organization ID"
// @Param id path string true "Organization ID"
// @Param userId path string true "User ID"
// @Success 204 "No Content"
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorDetailResponse
// @Router /orgs/{id}/members/{userId} [delete]
func (h *OrgsHandler) RemoveMember(c *gin.Context) {
	orgID := c.Param("id")
	userID := c.Param("userID")
	actorID, _ := middleware.UserID(c)

	if err := h.os.RemoveMember(orgID, userID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "failed_remove_member", "detail": err.Error()})
		return
	}

	before, _ := json.Marshal(gin.H{"user_id": userID})
	_ = h.as.Record(audit.Event{
		OrgID: orgID, ActorID: actorID, Action: "member.removed",
		Entity: "membership", EntityID: userID, Timestamp: time.Now(),
		Metadata: audit.Metadata{Before: before},
	})

	c.Status(http.StatusNoContent)
}

// Delete removes an organization (owner only).
// @Summary Delete organization
// @Tags Organizations
// @Security BearerAuth
// @Produce json
// @Param X-Org-ID header string true "Organization ID"
// @Param id path string true "Organization ID"
// @Success 204 "No Content"
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorDetailResponse
// @Router /orgs/{id} [delete]
func (h *OrgsHandler) Delete(c *gin.Context) {
	orgID := c.Param("id")
	actorID, _ := middleware.UserID(c)

	if err := h.os.Delete(orgID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "failed_delete_organization", "detail": err.Error()})
		return
	}

	before, _ := json.Marshal(gin.H{"org_id": orgID})
	_ = h.as.Record(audit.Event{
		OrgID: orgID, ActorID: actorID, Action: "org.deleted",
		Entity: "org", EntityID: orgID, Timestamp: time.Now(),
		Metadata: audit.Metadata{Before: before},
	})

	c.Status(http.StatusNoContent)
}

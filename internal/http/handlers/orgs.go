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

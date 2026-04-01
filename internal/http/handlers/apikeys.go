package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"encoding/json"

	"github.com/Ulpio/vergo/internal/domain/apikey"
	"github.com/Ulpio/vergo/internal/domain/audit"
	"github.com/Ulpio/vergo/internal/http/middleware"
)

type APIKeysHandler struct {
	ks apikey.Service
	as audit.Service
}

func NewAPIKeysHandler(ks apikey.Service, as audit.Service) *APIKeysHandler {
	return &APIKeysHandler{ks: ks, as: as}
}

type createKeyIn struct {
	Name      string     `json:"name" binding:"required"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// Create creates a new API key.
// @Summary Create API key
// @Tags API Keys
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param X-Org-ID header string true "Organization ID"
// @Param body body createKeyIn true "Key name and optional expiration"
// @Success 201 {object} apikey.CreateResult
// @Failure 401 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api-keys [post]
func (h *APIKeysHandler) Create(c *gin.Context) {
	uid, _ := middleware.UserID(c)
	orgID, _ := middleware.OrgID(c)

	var in createKeyIn
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid_payload"})
		return
	}

	result, err := h.ks.Create(orgID, uid, in.Name, in.ExpiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create_failed"})
		return
	}

	_ = h.as.Record(audit.Event{
		OrgID:    orgID,
		ActorID:  uid,
		Action:   "api_key.created",
		Entity:   "api_key",
		EntityID: result.ID,
		Metadata: toAuditMeta(map[string]string{"name": in.Name}),
	})

	c.JSON(http.StatusCreated, result)
}

// List lists all active API keys for the organization.
// @Summary List API keys
// @Tags API Keys
// @Security BearerAuth
// @Produce json
// @Param X-Org-ID header string true "Organization ID"
// @Success 200 {array} apikey.APIKey
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api-keys [get]
func (h *APIKeysHandler) List(c *gin.Context) {
	orgID, _ := middleware.OrgID(c)

	keys, err := h.ks.List(orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list_failed"})
		return
	}
	c.JSON(http.StatusOK, keys)
}

// Revoke revokes an API key.
// @Summary Revoke API key
// @Tags API Keys
// @Security BearerAuth
// @Param X-Org-ID header string true "Organization ID"
// @Param id path string true "API Key ID"
// @Success 204
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api-keys/{id} [delete]
func (h *APIKeysHandler) Revoke(c *gin.Context) {
	uid, _ := middleware.UserID(c)
	orgID, _ := middleware.OrgID(c)
	keyID := c.Param("id")

	if err := h.ks.Revoke(orgID, keyID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "revoke_failed"})
		return
	}

	_ = h.as.Record(audit.Event{
		OrgID:    orgID,
		ActorID:  uid,
		Action:   "api_key.revoked",
		Entity:   "api_key",
		EntityID: keyID,
	})

	c.Status(http.StatusNoContent)
}

func toAuditMeta(data any) audit.Metadata {
	b, _ := json.Marshal(data)
	return audit.Metadata{After: b}
}

package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Ulpio/vergo/internal/domain/webhook"
	"github.com/Ulpio/vergo/internal/http/middleware"
)

type WebhooksHandler struct {
	ws webhook.Service
}

func NewWebhooksHandler(ws webhook.Service) *WebhooksHandler {
	return &WebhooksHandler{ws: ws}
}

type createEndpointIn struct {
	URL    string   `json:"url" binding:"required"`
	Events []string `json:"events" binding:"required"`
}

// CreateEndpoint creates a new webhook endpoint.
// @Summary Create webhook endpoint
// @Tags Webhooks
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param X-Org-ID header string true "Organization ID"
// @Param body body createEndpointIn true "Endpoint URL and events"
// @Success 201 {object} webhook.Endpoint
// @Failure 401 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /webhooks/endpoints [post]
func (h *WebhooksHandler) CreateEndpoint(c *gin.Context) {
	orgID, _ := middleware.OrgID(c)

	var in createEndpointIn
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid_payload"})
		return
	}

	ep, err := h.ws.CreateEndpoint(orgID, in.URL, in.Events)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create_failed"})
		return
	}
	c.JSON(http.StatusCreated, ep)
}

// ListEndpoints lists all webhook endpoints for the organization.
// @Summary List webhook endpoints
// @Tags Webhooks
// @Security BearerAuth
// @Produce json
// @Param X-Org-ID header string true "Organization ID"
// @Success 200 {array} webhook.Endpoint
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /webhooks/endpoints [get]
func (h *WebhooksHandler) ListEndpoints(c *gin.Context) {
	orgID, _ := middleware.OrgID(c)

	eps, err := h.ws.ListEndpoints(orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list_failed"})
		return
	}
	c.JSON(http.StatusOK, eps)
}

type updateEndpointIn struct {
	URL    string   `json:"url" binding:"required"`
	Events []string `json:"events" binding:"required"`
	Active bool     `json:"active"`
}

// UpdateEndpoint updates a webhook endpoint.
// @Summary Update webhook endpoint
// @Tags Webhooks
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param X-Org-ID header string true "Organization ID"
// @Param id path string true "Endpoint ID"
// @Param body body updateEndpointIn true "Updated endpoint config"
// @Success 204
// @Failure 401 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /webhooks/endpoints/{id} [patch]
func (h *WebhooksHandler) UpdateEndpoint(c *gin.Context) {
	orgID, _ := middleware.OrgID(c)
	id := c.Param("id")

	var in updateEndpointIn
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid_payload"})
		return
	}

	if err := h.ws.UpdateEndpoint(orgID, id, in.URL, in.Events, in.Active); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update_failed"})
		return
	}
	c.Status(http.StatusNoContent)
}

// Test sends a test webhook to an endpoint.
// @Summary Test webhook endpoint
// @Tags Webhooks
// @Security BearerAuth
// @Produce json
// @Param X-Org-ID header string true "Organization ID"
// @Param endpoint_id query string true "Endpoint ID to test"
// @Success 200 {object} map[string]string
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /webhooks/test [post]
func (h *WebhooksHandler) Test(c *gin.Context) {
	orgID, _ := middleware.OrgID(c)
	epID := c.Query("endpoint_id")
	if epID == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "missing_endpoint_id"})
		return
	}

	if err := h.ws.TestEndpoint(orgID, epID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "test_failed", "detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "delivered"})
}

package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Ulpio/vergo/internal/domain/audit"
	"github.com/Ulpio/vergo/internal/http/middleware"
	"github.com/gin-gonic/gin"
)

type AuditHandler struct {
	as audit.Service
}

func NewAuditHandler(as audit.Service) *AuditHandler { return &AuditHandler{as: as} }

// List returns audit log events for the organization.
// @Summary List audit events
// @Tags Audit
// @Security BearerAuth
// @Produce json
// @Param X-Org-ID header string true "Organization ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Items per page (max 100)" default(20)
// @Param actor_id query string false "Filter by actor"
// @Param action query string false "Filter by action"
// @Param entity query string false "Filter by entity type"
// @Param since query string false "Filter from date (RFC3339)"
// @Param until query string false "Filter to date (RFC3339)"
// @Success 200 {object} map[string]interface{} "items, page, page_size, next_offset"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /audit [get]
func (h *AuditHandler) List(c *gin.Context) {
	orgID, ok := middleware.OrgID(c)
	if !ok || orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_org_id"})
		return
	}

	pageSize := parseInt(c.Query("page_size"), 20)
	page := parseInt(c.Query("page"), 1)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	params := audit.ListParams{
		OrgID:  orgID,
		Limit:  pageSize,
		Offset: offset,
	}

	// filtros opcionais
	if v := c.Query("actor_id"); v != "" {
		params.ActorID = &v
	}
	if v := c.Query("action"); v != "" {
		params.Action = &v
	}
	if v := c.Query("entity"); v != "" {
		params.Entity = &v
	}
	if v := c.Query("since"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			params.Since = &t
		}
	}
	if v := c.Query("until"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			params.Until = &t
		}
	}

	items, err := h.as.List(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list_failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":       items,
		"page":        page,
		"page_size":   pageSize,
		"next_offset": offset + len(items),
	})
}

func parseInt(s string, def int) int {
	if s == "" {
		return def
	}
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return def
}

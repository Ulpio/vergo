package handlers

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v82"
	stripeWebhook "github.com/stripe/stripe-go/v82/webhook"

	"github.com/Ulpio/vergo/internal/domain/billing"
	"github.com/Ulpio/vergo/internal/http/middleware"
)

type BillingHandler struct {
	bs            billing.Service
	webhookSecret string
}

func NewBillingHandler(bs billing.Service, webhookSecret string) *BillingHandler {
	return &BillingHandler{bs: bs, webhookSecret: webhookSecret}
}

type checkoutIn struct {
	SuccessURL string `json:"success_url" binding:"required"`
	CancelURL  string `json:"cancel_url" binding:"required"`
	PriceID    string `json:"price_id" binding:"required"`
}

// CreateCheckoutSession creates a Stripe checkout session.
// @Summary Create Stripe checkout session
// @Tags Billing
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param X-Org-ID header string true "Organization ID"
// @Param body body checkoutIn true "Checkout parameters"
// @Success 200 {object} map[string]string
// @Failure 401 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /billing/checkout-session [post]
func (h *BillingHandler) CreateCheckoutSession(c *gin.Context) {
	orgID, _ := middleware.OrgID(c)

	var in checkoutIn
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid_payload"})
		return
	}

	url, err := h.bs.CreateCheckoutSession(orgID, orgID, in.SuccessURL, in.CancelURL, in.PriceID)
	if err != nil {
		slog.Error("billing: checkout", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "checkout_failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": url})
}

// GetSubscription returns the current subscription for the organization.
// @Summary Get subscription
// @Tags Billing
// @Security BearerAuth
// @Produce json
// @Param X-Org-ID header string true "Organization ID"
// @Success 200 {object} billing.Subscription
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /billing/subscription [get]
func (h *BillingHandler) GetSubscription(c *gin.Context) {
	orgID, _ := middleware.OrgID(c)

	sub, err := h.bs.GetSubscription(orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "fetch_failed"})
		return
	}
	c.JSON(http.StatusOK, sub)
}

// GetUsage returns current usage vs plan limits.
// @Summary Get usage vs plan limits
// @Tags Billing
// @Security BearerAuth
// @Produce json
// @Param X-Org-ID header string true "Organization ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /billing/usage [get]
func (h *BillingHandler) GetUsage(c *gin.Context) {
	orgID, _ := middleware.OrgID(c)

	sub, err := h.bs.GetSubscription(orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "fetch_failed"})
		return
	}

	limits := billing.GetLimits(sub.Plan)
	c.JSON(http.StatusOK, gin.H{
		"plan":   sub.Plan,
		"status": sub.Status,
		"limits": limits,
	})
}

// Webhook handles Stripe webhook events. This endpoint is public (no auth).
// @Summary Stripe webhook
// @Tags Billing
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Router /billing/webhook [post]
func (h *BillingHandler) Webhook(c *gin.Context) {
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, 65536))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "read_body"})
		return
	}

	event, err := stripeWebhook.ConstructEvent(body, c.GetHeader("Stripe-Signature"), h.webhookSecret)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_signature"})
		return
	}

	switch event.Type {
	case "checkout.session.completed":
		h.handleCheckoutCompleted(event)
	case "invoice.paid":
		// subscription renewed — same as update
		h.handleSubscriptionUpdated(event)
	case "customer.subscription.updated":
		h.handleSubscriptionUpdated(event)
	case "customer.subscription.deleted":
		h.handleSubscriptionDeleted(event)
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
}

func (h *BillingHandler) handleCheckoutCompleted(event stripe.Event) {
	var sess stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &sess); err != nil {
		slog.Error("billing: unmarshal checkout", "error", err)
		return
	}

	orgID := sess.ClientReferenceID
	if orgID == "" {
		slog.Warn("billing: checkout missing client_reference_id")
		return
	}

	_ = h.bs.HandleCheckoutCompleted(
		sess.Customer.ID,
		sess.Subscription.ID,
		"active",
		"pro",
		time.Time{}, // will be updated by subscription.updated event
		orgID,
	)
}

func (h *BillingHandler) handleSubscriptionUpdated(event stripe.Event) {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		slog.Error("billing: unmarshal subscription", "error", err)
		return
	}

	plan := "free"
	if sub.Status == stripe.SubscriptionStatusActive || sub.Status == stripe.SubscriptionStatusTrialing {
		plan = "pro"
	}

	var periodEnd time.Time
	if sub.Items != nil && len(sub.Items.Data) > 0 {
		periodEnd = time.Unix(sub.Items.Data[0].CurrentPeriodEnd, 0)
	}
	_ = h.bs.HandleSubscriptionUpdated(sub.ID, string(sub.Status), plan, periodEnd)
}

func (h *BillingHandler) handleSubscriptionDeleted(event stripe.Event) {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		slog.Error("billing: unmarshal subscription delete", "error", err)
		return
	}
	_ = h.bs.HandleSubscriptionDeleted(sub.ID)
}

package billing

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/customer"

	"github.com/Ulpio/vergo/internal/repo"
)

type Subscription struct {
	ID                   string     `json:"id"`
	OrgID                string     `json:"org_id"`
	Status               string     `json:"status"`
	Plan                 string     `json:"plan"`
	CurrentPeriodEnd     *time.Time `json:"current_period_end,omitempty"`
	StripeSubscriptionID string     `json:"stripe_subscription_id,omitempty"`
}

type Service interface {
	CreateCheckoutSession(orgID, orgName, successURL, cancelURL, priceID string) (string, error)
	GetSubscription(orgID string) (*Subscription, error)
	HandleCheckoutCompleted(stripeCustomerID, stripeSubID, status, plan string, periodEnd time.Time, orgID string) error
	HandleSubscriptionUpdated(stripeSubID, status, plan string, periodEnd time.Time) error
	HandleSubscriptionDeleted(stripeSubID string) error
}

type service struct {
	q *repo.Queries
}

func NewService(q *repo.Queries, stripeKey string) Service {
	stripe.Key = stripeKey
	return &service{q: q}
}

func (s *service) CreateCheckoutSession(orgID, orgName, successURL, cancelURL, priceID string) (string, error) {
	// Check for existing subscription to get or create Stripe customer
	var customerID string
	sub, err := s.q.GetSubscriptionByOrg(context.Background(), orgID)
	if err == nil {
		customerID = sub.StripeCustomerID
	} else if errors.Is(err, sql.ErrNoRows) {
		// Create new Stripe customer
		c, err := customer.New(&stripe.CustomerParams{
			Name: stripe.String(orgName),
			Params: stripe.Params{
				Metadata: map[string]string{"org_id": orgID},
			},
		})
		if err != nil {
			return "", fmt.Errorf("stripe customer: %w", err)
		}
		customerID = c.ID

		// Save subscription record
		_ = s.q.UpsertSubscription(context.Background(), repo.UpsertSubscriptionParams{
			OrgID:            orgID,
			StripeCustomerID: customerID,
			Status:           "incomplete",
			Plan:             "free",
		})
	} else {
		return "", err
	}

	params := &stripe.CheckoutSessionParams{
		Customer:           stripe.String(customerID),
		Mode:               stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL:         stripe.String(successURL),
		CancelURL:          stripe.String(cancelURL),
		ClientReferenceID:  stripe.String(orgID),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
	}

	sess, err := session.New(params)
	if err != nil {
		return "", fmt.Errorf("stripe checkout: %w", err)
	}
	return sess.URL, nil
}

func (s *service) GetSubscription(orgID string) (*Subscription, error) {
	row, err := s.q.GetSubscriptionByOrg(context.Background(), orgID)
	if errors.Is(err, sql.ErrNoRows) {
		return &Subscription{OrgID: orgID, Status: "none", Plan: "free"}, nil
	}
	if err != nil {
		return nil, err
	}

	sub := &Subscription{
		ID:     row.ID,
		OrgID:  row.OrgID,
		Status: row.Status,
		Plan:   row.Plan,
	}
	if row.StripeSubscriptionID.Valid {
		sub.StripeSubscriptionID = row.StripeSubscriptionID.String
	}
	if row.CurrentPeriodEnd.Valid {
		sub.CurrentPeriodEnd = &row.CurrentPeriodEnd.Time
	}
	return sub, nil
}

func (s *service) HandleCheckoutCompleted(stripeCustomerID, stripeSubID, status, plan string, periodEnd time.Time, orgID string) error {
	return s.q.UpsertSubscription(context.Background(), repo.UpsertSubscriptionParams{
		OrgID:                  orgID,
		StripeCustomerID:       stripeCustomerID,
		StripeSubscriptionID:   sql.NullString{String: stripeSubID, Valid: stripeSubID != ""},
		Status:                 status,
		Plan:                   plan,
		CurrentPeriodEnd:       sql.NullTime{Time: periodEnd, Valid: !periodEnd.IsZero()},
	})
}

func (s *service) HandleSubscriptionUpdated(stripeSubID, status, plan string, periodEnd time.Time) error {
	return s.q.UpdateSubscriptionStatus(context.Background(), repo.UpdateSubscriptionStatusParams{
		StripeSubscriptionID: sql.NullString{String: stripeSubID, Valid: true},
		Status:               status,
		Plan:                 plan,
		CurrentPeriodEnd:     sql.NullTime{Time: periodEnd, Valid: !periodEnd.IsZero()},
	})
}

func (s *service) HandleSubscriptionDeleted(stripeSubID string) error {
	return s.q.CancelSubscription(context.Background(), sql.NullString{String: stripeSubID, Valid: true})
}

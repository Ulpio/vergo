package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Ulpio/vergo/internal/repo"
	"github.com/lib/pq"
)

type Endpoint struct {
	ID        string    `json:"id"`
	OrgID     string    `json:"org_id"`
	URL       string    `json:"url"`
	Events    []string  `json:"events"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Service interface {
	CreateEndpoint(orgID, url string, events []string) (Endpoint, error)
	ListEndpoints(orgID string) ([]Endpoint, error)
	UpdateEndpoint(orgID, id, url string, events []string, active bool) error
	Dispatch(orgID, event string, payload json.RawMessage) error
	TestEndpoint(orgID, endpointID string) error
}

type service struct {
	q      *repo.Queries
	client *http.Client
}

func NewService(q *repo.Queries) Service {
	return &service{
		q:      q,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *service) CreateEndpoint(orgID, url string, events []string) (Endpoint, error) {
	secret, err := generateSecret()
	if err != nil {
		return Endpoint{}, fmt.Errorf("generate secret: %w", err)
	}

	row, err := s.q.CreateWebhookEndpoint(context.Background(), repo.CreateWebhookEndpointParams{
		OrgID:  orgID,
		Url:    url,
		Secret: secret,
		Events: events,
	})
	if err != nil {
		return Endpoint{}, err
	}

	return Endpoint{
		ID:        row.ID,
		OrgID:     row.OrgID,
		URL:       row.Url,
		Events:    row.Events,
		Active:    row.Active,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (s *service) ListEndpoints(orgID string) ([]Endpoint, error) {
	rows, err := s.q.ListWebhookEndpoints(context.Background(), orgID)
	if err != nil {
		return nil, err
	}
	out := make([]Endpoint, len(rows))
	for i, r := range rows {
		out[i] = Endpoint{
			ID:        r.ID,
			OrgID:     r.OrgID,
			URL:       r.Url,
			Events:    r.Events,
			Active:    r.Active,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		}
	}
	return out, nil
}

func (s *service) UpdateEndpoint(orgID, id, url string, events []string, active bool) error {
	return s.q.UpdateWebhookEndpoint(context.Background(), repo.UpdateWebhookEndpointParams{
		ID:     id,
		OrgID:  orgID,
		Url:    url,
		Events: events,
		Active: active,
	})
}

func (s *service) Dispatch(orgID, event string, payload json.RawMessage) error {
	endpoints, err := s.q.GetActiveEndpointsForEvent(context.Background(), repo.GetActiveEndpointsForEventParams{
		OrgID:  orgID,
		Events: []string{event},
	})
	if err != nil {
		return err
	}

	for _, ep := range endpoints {
		_, err := s.q.InsertDelivery(context.Background(), repo.InsertDeliveryParams{
			EndpointID: ep.ID,
			Event:      event,
			Payload:    payload,
		})
		if err != nil {
			slog.Error("webhook: insert delivery", "endpoint", ep.ID, "error", err)
		}
	}
	return nil
}

func (s *service) TestEndpoint(orgID, endpointID string) error {
	ep, err := s.q.GetWebhookEndpoint(context.Background(), repo.GetWebhookEndpointParams{
		ID:    endpointID,
		OrgID: orgID,
	})
	if err != nil {
		return err
	}

	payload := json.RawMessage(`{"type":"webhook.test","org_id":"` + orgID + `"}`)
	return deliver(s.client, ep.Url, ep.Secret, "webhook.test", payload)
}

func deliver(client *http.Client, url, secret, event string, payload json.RawMessage) error {
	sig := sign(payload, secret)

	req, err := http.NewRequest("POST", url, strings.NewReader(string(payload)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Event", event)
	req.Header.Set("X-Webhook-Signature", "sha256="+sig)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	return fmt.Errorf("webhook returned %d", resp.StatusCode)
}

func sign(payload json.RawMessage, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func generateSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "whsec_" + hex.EncodeToString(b), nil
}

// Ensure lib/pq is used (imported by sqlc generated code).
var _ = pq.Array

package apikey

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/Ulpio/vergo/internal/repo"
)

type APIKey struct {
	ID         string     `json:"id"`
	OrgID      string     `json:"org_id"`
	Name       string     `json:"name"`
	KeyPrefix  string     `json:"key_prefix"`
	CreatedBy  string     `json:"created_by"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

type CreateResult struct {
	APIKey
	PlaintextKey string `json:"key"`
}

type LookupResult struct {
	KeyID string
	OrgID string
}

type Service interface {
	Create(orgID, userID, name string, expiresAt *time.Time) (CreateResult, error)
	List(orgID string) ([]APIKey, error)
	Revoke(orgID, keyID string) error
	Validate(key string) (*LookupResult, error)
}

type service struct {
	q *repo.Queries
}

func NewService(q *repo.Queries) Service {
	return &service{q: q}
}

func (s *service) Create(orgID, userID, name string, expiresAt *time.Time) (CreateResult, error) {
	plaintext, err := generateKey()
	if err != nil {
		return CreateResult{}, fmt.Errorf("generate key: %w", err)
	}

	hash := hashKey(plaintext)
	prefix := plaintext[:12] // "sk_" + first 9 chars

	var expSQL sql.NullTime
	if expiresAt != nil {
		expSQL = sql.NullTime{Time: *expiresAt, Valid: true}
	}

	row, err := s.q.CreateAPIKey(context.Background(), repo.CreateAPIKeyParams{
		OrgID:     orgID,
		Name:      name,
		KeyPrefix: prefix,
		KeyHash:   hash,
		CreatedBy: userID,
		ExpiresAt: expSQL,
	})
	if err != nil {
		return CreateResult{}, fmt.Errorf("insert api key: %w", err)
	}

	ak := APIKey{
		ID:        row.ID,
		OrgID:     row.OrgID,
		Name:      row.Name,
		KeyPrefix: row.KeyPrefix,
		CreatedBy: row.CreatedBy,
		CreatedAt: row.CreatedAt,
	}
	if row.ExpiresAt.Valid {
		ak.ExpiresAt = &row.ExpiresAt.Time
	}

	return CreateResult{APIKey: ak, PlaintextKey: plaintext}, nil
}

func (s *service) List(orgID string) ([]APIKey, error) {
	rows, err := s.q.ListAPIKeysByOrg(context.Background(), orgID)
	if err != nil {
		return nil, err
	}

	out := make([]APIKey, len(rows))
	for i, r := range rows {
		out[i] = APIKey{
			ID:        r.ID,
			OrgID:     r.OrgID,
			Name:      r.Name,
			KeyPrefix: r.KeyPrefix,
			CreatedBy: r.CreatedBy,
			CreatedAt: r.CreatedAt,
		}
		if r.ExpiresAt.Valid {
			out[i].ExpiresAt = &r.ExpiresAt.Time
		}
		if r.LastUsedAt.Valid {
			out[i].LastUsedAt = &r.LastUsedAt.Time
		}
	}
	return out, nil
}

func (s *service) Revoke(orgID, keyID string) error {
	return s.q.RevokeAPIKey(context.Background(), repo.RevokeAPIKeyParams{
		ID:    keyID,
		OrgID: orgID,
	})
}

func (s *service) Validate(key string) (*LookupResult, error) {
	hash := hashKey(key)
	row, err := s.q.GetAPIKeyByHash(context.Background(), hash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if row.ExpiresAt.Valid && row.ExpiresAt.Time.Before(time.Now()) {
		return nil, nil
	}

	// fire-and-forget last_used update
	go func() {
		_ = s.q.TouchAPIKeyLastUsed(context.Background(), row.ID)
	}()

	return &LookupResult{KeyID: row.ID, OrgID: row.OrgID}, nil
}

func generateKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "sk_" + hex.EncodeToString(b), nil
}

func hashKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

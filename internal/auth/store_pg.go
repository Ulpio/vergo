package auth

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"github.com/Ulpio/vergo/internal/repo"
)

type RefreshStore interface {
	SaveRefresh(ctx context.Context, jti, userID, token string, expiresAt time.Time, rotatedFrom *string) error
	IsValid(ctx context.Context, jti, token string) (bool, string, time.Time, error)
	Revoke(ctx context.Context, jti string) error
	RevokeAllForUser(ctx context.Context, userID string) error
}

type pgStore struct {
	db *sql.DB
	q  *repo.Queries
}

func NewRefreshStore(db *sql.DB, q *repo.Queries) RefreshStore {
	return &pgStore{db: db, q: q}
}

func hash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (s *pgStore) SaveRefresh(ctx context.Context, jti, userID, token string, expiresAt time.Time, rotatedFrom *string) error {
	params := repo.InsertRefreshTokenParams{
		ID:        jti,
		UserID:    userID,
		TokenHash: hash(token),
		ExpiresAt: expiresAt,
	}
	if rotatedFrom != nil {
		params.RotatedFrom = sql.NullString{String: *rotatedFrom, Valid: true}
	}
	return s.q.InsertRefreshToken(ctx, params)
}

func (s *pgStore) IsValid(ctx context.Context, jti, token string) (bool, string, time.Time, error) {
	row, err := s.q.GetRefreshToken(ctx, jti)
	if errors.Is(err, sql.ErrNoRows) {
		return false, "", time.Time{}, nil
	}
	if err != nil {
		return false, "", time.Time{}, err
	}
	if row.RevokedAt.Valid {
		return false, "", time.Time{}, nil
	}
	if time.Now().After(row.ExpiresAt) {
		return false, "", time.Time{}, nil
	}
	if row.TokenHash != hash(token) {
		return false, "", time.Time{}, nil
	}
	return true, row.UserID, row.ExpiresAt, nil
}

func (s *pgStore) Revoke(ctx context.Context, jti string) error {
	return s.q.RevokeToken(ctx, jti)
}

func (s *pgStore) RevokeAllForUser(ctx context.Context, userID string) error {
	return s.q.RevokeAllUserTokens(ctx, userID)
}

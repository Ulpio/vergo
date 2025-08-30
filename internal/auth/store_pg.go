package auth

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"
)

type RefreshStore interface {
	SaveRefresh(ctx context.Context, jti, userID, token string, expiresAt time.Time, rotatedFrom *string) error
	IsValid(ctx context.Context, jti, token string) (bool, string, time.Time, error)
	Revoke(ctx context.Context, jti string) error
	RevokeAllForUser(ctx context.Context, userID string) error
}

type pgStore struct {
	db *sql.DB
}

func NewRefreshStore(db *sql.DB) RefreshStore { return &pgStore{db: db} }

func hash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (s *pgStore) SaveRefresh(ctx context.Context, jti, userID, token string, expiresAt time.Time, rotatedFrom *string) error {
	const q = `
INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, rotated_from)
VALUES ($1,$2,$3,$4,$5)`
	_, err := s.db.ExecContext(ctx, q, jti, userID, hash(token), expiresAt, rotatedFrom)
	return err
}

func (s *pgStore) IsValid(ctx context.Context, jti, token string) (bool, string, time.Time, error) {
	const q = `
SELECT user_id, token_hash, expires_at, revoked_at
FROM refresh_tokens
WHERE id = $1`
	var userID, tokenHash string
	var exp time.Time
	var revokedAt sql.NullTime

	err := s.db.QueryRowContext(ctx, q, jti).Scan(&userID, &tokenHash, &exp, &revokedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return false, "", time.Time{}, nil
	}
	if err != nil {
		return false, "", time.Time{}, err
	}
	if revokedAt.Valid {
		return false, "", time.Time{}, nil
	}
	if time.Now().After(exp) {
		return false, "", time.Time{}, nil
	}
	if tokenHash != hash(token) {
		return false, "", time.Time{}, nil
	}
	return true, userID, exp, nil
}

func (s *pgStore) Revoke(ctx context.Context, jti string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE refresh_tokens SET revoked_at = NOW() WHERE id = $1 AND revoked_at IS NULL`, jti)
	return err
}

func (s *pgStore) RevokeAllForUser(ctx context.Context, userID string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE refresh_tokens SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL`, userID)
	return err
}

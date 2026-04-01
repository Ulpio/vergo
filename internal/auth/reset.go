package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"github.com/Ulpio/vergo/internal/repo"
)

var (
	ErrResetTokenInvalid = errors.New("invalid or expired reset token")
)

type ResetStore interface {
	CreateResetToken(userID string) (plainToken string, err error)
	ValidateAndConsume(token string) (userID string, err error)
}

type resetStore struct {
	q *repo.Queries
}

func NewResetStore(q *repo.Queries) ResetStore {
	return &resetStore{q: q}
}

func (s *resetStore) CreateResetToken(userID string) (string, error) {
	plain, err := generateResetToken()
	if err != nil {
		return "", err
	}

	hash := hashResetToken(plain)
	expiresAt := time.Now().Add(1 * time.Hour)

	err = s.q.CreatePasswordResetToken(context.Background(), repo.CreatePasswordResetTokenParams{
		UserID:    userID,
		TokenHash: hash,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return "", err
	}
	return plain, nil
}

func (s *resetStore) ValidateAndConsume(token string) (string, error) {
	hash := hashResetToken(token)

	row, err := s.q.GetPasswordResetByHash(context.Background(), hash)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrResetTokenInvalid
	}
	if err != nil {
		return "", err
	}

	if time.Now().After(row.ExpiresAt) {
		return "", ErrResetTokenInvalid
	}

	_ = s.q.MarkPasswordResetUsed(context.Background(), row.ID)

	return row.UserID, nil
}

func generateResetToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashResetToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

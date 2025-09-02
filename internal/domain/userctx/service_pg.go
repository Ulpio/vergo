package userctx

import (
	"context"
	"database/sql"
)

type Service interface {
	GetActiveOrg(userID string) (string, bool, error)
	SetActiveOrg(userID, orgID string) error
}

type pgService struct{ db *sql.DB }

func NewPostgresService(db *sql.DB) Service { return &pgService{db: db} }

func (s *pgService) GetActiveOrg(userID string) (string, bool, error) {
	const q = `SELECT org_id FROM user_contexts WHERE user_id = $1`
	var orgID string
	err := s.db.QueryRowContext(context.Background(), q, userID).Scan(&orgID)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	return orgID, err == nil, err
}

func (s *pgService) SetActiveOrg(userID, orgID string) error {
	const q = `
INSERT INTO user_contexts (user_id, org_id, updated_at)
VALUES ($1, $2, NOW())
ON CONFLICT (user_id) DO UPDATE
SET org_id = $2,
    updated_at = NOW()`
	_, err := s.db.ExecContext(context.Background(), q, userID, orgID)
	return err
}

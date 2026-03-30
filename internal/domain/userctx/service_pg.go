package userctx

import (
	"context"
	"database/sql"

	"github.com/Ulpio/vergo/internal/repo"
)

type Service interface {
	GetActiveOrg(userID string) (string, bool, error)
	SetActiveOrg(userID, orgID string) error
}

type pgService struct {
	db *sql.DB
	q  *repo.Queries
}

func NewPostgresService(db *sql.DB, q *repo.Queries) Service {
	return &pgService{db: db, q: q}
}

func (s *pgService) GetActiveOrg(userID string) (string, bool, error) {
	orgID, err := s.q.GetActiveOrg(context.Background(), userID)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	return orgID, err == nil, err
}

func (s *pgService) SetActiveOrg(userID, orgID string) error {
	return s.q.UpsertActiveOrg(context.Background(), repo.UpsertActiveOrgParams{
		UserID: userID,
		OrgID:  orgID,
	})
}

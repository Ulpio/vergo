package audit

import (
	"context"
	"database/sql"
	"time"
)

type pgService struct{ db *sql.DB }

func NewPostgresService(db *sql.DB) Service { return &pgService{db: db} }

func (s *pgService) Record(e Event) error {
	const q = `
INSERT INTO audit_logs (org_id, actor_id, action, entity, entity_id, created_at)
VALUES ($1,$2,$3,$4,$5,$6)`
	_, err := s.db.ExecContext(context.Background(), q,
		e.OrgID, e.ActorID, e.Action, e.Entity, e.EntityID, time.Now())
	return err
}

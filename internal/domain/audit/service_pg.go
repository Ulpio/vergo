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

func (s *pgService) List(p ListParams) ([]Event, error) {
	const q = `
SELECT org_id, actor_id, action, entity, entity_id, created_at
FROM audit_logs
WHERE org_id = $1
  AND ($2::text IS NULL OR actor_id = $2)
  AND ($3::text IS NULL OR action  = $3)
  AND ($4::text IS NULL OR entity  = $4)
  AND ($5::timestamptz IS NULL OR created_at >= $5)
  AND ($6::timestamptz IS NULL OR created_at <= $6)
ORDER BY created_at DESC
LIMIT $7 OFFSET $8`
	// sane defaults
	limit := p.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := p.Offset
	if offset < 0 {
		offset = 0
	}

	rows, err := s.db.QueryContext(context.Background(), q,
		p.OrgID, p.ActorID, p.Action, p.Entity, p.Since, p.Until, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.OrgID, &e.ActorID, &e.Action, &e.Entity, &e.EntityID, &e.Timestamp); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

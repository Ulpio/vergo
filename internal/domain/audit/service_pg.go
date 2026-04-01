package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/sqlc-dev/pqtype"

	"github.com/Ulpio/vergo/internal/repo"
)

type pgService struct {
	db *sql.DB
	q  *repo.Queries
}

func NewPostgresService(db *sql.DB, q *repo.Queries) Service {
	return &pgService{db: db, q: q}
}

func (s *pgService) Record(e Event) error {
	metaJSON, err := json.Marshal(e.Metadata)
	if err != nil {
		metaJSON = []byte("{}")
	}

	return s.q.InsertAuditLog(context.Background(), repo.InsertAuditLogParams{
		OrgID:     e.OrgID,
		ActorID:   e.ActorID,
		Action:    e.Action,
		Entity:    e.Entity,
		EntityID:  e.EntityID,
		Metadata:  pqtype.NullRawMessage{RawMessage: metaJSON, Valid: true},
		CreatedAt: time.Now(),
	})
}

func (s *pgService) List(p ListParams) ([]Event, error) {
	limit := p.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := p.Offset
	if offset < 0 {
		offset = 0
	}

	rows, err := s.q.ListAuditLogs(context.Background(), repo.ListAuditLogsParams{
		OrgID:         p.OrgID,
		FilterActorID: toNullString(p.ActorID),
		FilterAction:  toNullString(p.Action),
		FilterEntity:  toNullString(p.Entity),
		FilterSince:   toNullTime(p.Since),
		FilterUntil:   toNullTime(p.Until),
		QueryLimit:    int32(limit),
		QueryOffset:   int32(offset),
	})
	if err != nil {
		return nil, err
	}

	out := make([]Event, 0, len(rows))
	for _, r := range rows {
		var meta Metadata
		if len(r.Metadata) > 0 {
			_ = json.Unmarshal(r.Metadata, &meta)
		}
		out = append(out, Event{
			OrgID:     r.OrgID,
			ActorID:   r.ActorID,
			Action:    r.Action,
			Entity:    r.Entity,
			EntityID:  r.EntityID,
			Timestamp: r.CreatedAt,
			Metadata:  meta,
		})
	}
	return out, nil
}

func toNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

func toNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

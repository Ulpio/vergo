package file

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("file not found")

type ListParams struct {
	OrgID  string
	Limit  int
	Offset int
}

type Service interface {
	List(p ListParams) ([]File, error)
	Create(orgID, userID, bucket, key string, size *int64, contentType string, metadata any) (File, error)
	Get(orgID, id string) (File, error)
	Delete(orgID, id string) error
}

type pgService struct{ db *sql.DB }

func NewPostgresService(db *sql.DB) Service { return &pgService{db: db} }

func (s *pgService) List(p ListParams) ([]File, error) {
	if p.Limit <= 0 || p.Limit > 100 {
		p.Limit = 20
	}
	if p.Offset < 0 {
		p.Offset = 0
	}

	const q = `
SELECT id, org_id, uploaded_by, bucket, object_key, size_bytes, content_type, created_at, metadata
FROM files
WHERE org_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3`
	rows, err := s.db.QueryContext(context.Background(), q, p.OrgID, p.Limit, p.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []File
	for rows.Next() {
		var f File
		if err := rows.Scan(&f.ID, &f.OrgID, &f.UploadedBy, &f.Bucket, &f.ObjectKey, &f.SizeBytes, &f.ContentType, &f.CreatedAt, &f.Metadata); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func (s *pgService) Create(orgID, userID, bucket, key string, size *int64, contentType string, metadata any) (File, error) {
	id := uuid.NewString()
	const q = `
INSERT INTO files (id, org_id, uploaded_by, bucket, object_key, size_bytes, content_type, created_at, metadata)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
RETURNING id, org_id, uploaded_by, bucket, object_key, size_bytes, content_type, created_at, metadata`
	var f File
	err := s.db.QueryRowContext(context.Background(), q,
		id, orgID, userID, bucket, key, size, contentType, time.Now(), metadata,
	).Scan(&f.ID, &f.OrgID, &f.UploadedBy, &f.Bucket, &f.ObjectKey, &f.SizeBytes, &f.ContentType, &f.CreatedAt, &f.Metadata)
	return f, err
}

func (s *pgService) Get(orgID, id string) (File, error) {
	const q = `
SELECT id, org_id, uploaded_by, bucket, object_key, size_bytes, content_type, created_at, metadata
FROM files
WHERE id = $1 AND org_id = $2`
	var f File
	err := s.db.QueryRowContext(context.Background(), q, id, orgID).
		Scan(&f.ID, &f.OrgID, &f.UploadedBy, &f.Bucket, &f.ObjectKey, &f.SizeBytes, &f.ContentType, &f.CreatedAt, &f.Metadata)
	if errors.Is(err, sql.ErrNoRows) {
		return File{}, ErrNotFound
	}
	return f, err
}

func (s *pgService) Delete(orgID, id string) error {
	res, err := s.db.ExecContext(context.Background(), `DELETE FROM files WHERE id=$1 AND org_id=$2`, id, orgID)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return ErrNotFound
	}
	return nil
}

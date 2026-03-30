package file

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"

	"github.com/Ulpio/vergo/internal/repo"
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

type pgService struct {
	db *sql.DB
	q  *repo.Queries
}

func NewPostgresService(db *sql.DB, q *repo.Queries) Service {
	return &pgService{db: db, q: q}
}

func repoToFile(r repo.File) File {
	f := File{
		ID:          r.ID,
		OrgID:       r.OrgID,
		UploadedBy:  r.UploadedBy,
		Bucket:      r.Bucket,
		ObjectKey:   r.ObjectKey,
		ContentType: r.ContentType.String,
		CreatedAt:   r.CreatedAt,
	}
	if r.SizeBytes.Valid {
		v := r.SizeBytes.Int64
		f.SizeBytes = &v
	}
	if r.Metadata.Valid && len(r.Metadata.RawMessage) > 0 {
		var m map[string]any
		if json.Unmarshal(r.Metadata.RawMessage, &m) == nil {
			f.Metadata = m
		}
	}
	return f
}

func (s *pgService) List(p ListParams) ([]File, error) {
	if p.Limit <= 0 || p.Limit > 100 {
		p.Limit = 20
	}
	if p.Offset < 0 {
		p.Offset = 0
	}

	rows, err := s.q.ListFiles(context.Background(), repo.ListFilesParams{
		OrgID:  p.OrgID,
		Limit:  int32(p.Limit),
		Offset: int32(p.Offset),
	})
	if err != nil {
		return nil, err
	}

	out := make([]File, len(rows))
	for i, r := range rows {
		out[i] = repoToFile(r)
	}
	return out, nil
}

func (s *pgService) Create(orgID, userID, bucket, key string, size *int64, contentType string, metadata any) (File, error) {
	id := uuid.NewString()

	sizeBytes := sql.NullInt64{}
	if size != nil {
		sizeBytes = sql.NullInt64{Int64: *size, Valid: true}
	}

	ct := sql.NullString{String: contentType, Valid: contentType != ""}

	metaNRM := pqtype.NullRawMessage{}
	if metadata != nil {
		b, err := json.Marshal(metadata)
		if err == nil {
			metaNRM = pqtype.NullRawMessage{RawMessage: b, Valid: true}
		}
	}

	r, err := s.q.InsertFile(context.Background(), repo.InsertFileParams{
		ID:          id,
		OrgID:       orgID,
		UploadedBy:  userID,
		Bucket:      bucket,
		ObjectKey:   key,
		SizeBytes:   sizeBytes,
		ContentType: ct,
		CreatedAt:   time.Now(),
		Metadata:    metaNRM,
	})
	if err != nil {
		return File{}, err
	}
	return repoToFile(r), nil
}

func (s *pgService) Get(orgID, id string) (File, error) {
	r, err := s.q.GetFile(context.Background(), repo.GetFileParams{
		ID:    id,
		OrgID: orgID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return File{}, ErrNotFound
	}
	if err != nil {
		return File{}, err
	}
	return repoToFile(r), nil
}

func (s *pgService) Delete(orgID, id string) error {
	res, err := s.q.DeleteFile(context.Background(), repo.DeleteFileParams{
		ID:    id,
		OrgID: orgID,
	})
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return ErrNotFound
	}
	return nil
}

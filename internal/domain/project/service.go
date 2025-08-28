package project

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrNotFound = errors.New("project not found")
)

type Service interface {
	List(orgID string) ([]Project, error)
	Create(orgID, name, description, userID string) (Project, error)
	Get(orgID, id string) (Project, error)
	Update(orgID, id, name, description string) (Project, error)
	Delete(orgID, id string) error
}

type pgService struct {
	db *sql.DB
}

func NewPostgresService(db *sql.DB) Service {
	return &pgService{db: db}
}
func (s *pgService) List(orgID string) ([]Project, error) {
	const q = `
		SELECT id, org_id, name, description, created_by, created_at, updated_at
		FROM projects
		WHERE org_id = $1
		ORDER BY created_at ASC`
	rows, err := s.db.QueryContext(context.Background(), q, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.OrgID, &p.Name, &p.Description, &p.CreatedAt, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *pgService) Create(orgID, name, description, userID string) (Project, error) {
	const q = `
INSERT INTO projects (id, org_id, name, description, created_by, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5, NOW(), NOW())
RETURNING id, org_id, name, description, created_by, created_at, updated_at`
	id := uuid.NewString()
	var p Project
	err := s.db.QueryRowContext(context.Background(), q, id, orgID, name, description, userID).
		Scan(&p.ID, &p.OrgID, &p.Name, &p.Description, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

func (s *pgService) Get(orgID, id string) (Project, error) {
	const q = `
SELECT id, org_id, name, description, created_by, created_at, updated_at
FROM projects
WHERE id = $1 AND org_id = $2`
	var p Project
	err := s.db.QueryRowContext(context.Background(), q, id, orgID).
		Scan(&p.ID, &p.OrgID, &p.Name, &p.Description, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Project{}, ErrNotFound
	}
	return p, err
}

func (s *pgService) Update(orgID, id, name, description string) (Project, error) {
	const q = `
UPDATE projects
SET name = COALESCE(NULLIF($3,''), name),
    description = $4,
    updated_at = NOW()
WHERE id = $1 AND org_id = $2
RETURNING id, org_id, name, description, created_by, created_at, updated_at`
	var p Project
	err := s.db.QueryRowContext(context.Background(), q, id, orgID, name, description).
		Scan(&p.ID, &p.OrgID, &p.Name, &p.Description, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Project{}, ErrNotFound
	}
	return p, err
}

func (s *pgService) Delete(orgID, id string) error {
	const q = `DELETE FROM projects WHERE id = $1 AND org_id = $2`
	res, err := s.db.ExecContext(context.Background(), q, id, orgID)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return ErrNotFound
	}
	return nil
}

package project

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"

	"github.com/Ulpio/vergo/internal/repo"
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
	q  *repo.Queries
}

func NewPostgresService(db *sql.DB, q *repo.Queries) Service {
	return &pgService{db: db, q: q}
}

func repoToProject(r repo.Project) Project {
	return Project{
		ID:          r.ID,
		OrgID:       r.OrgID,
		Name:        r.Name,
		Description: r.Description.String,
		CreatedBy:   r.CreatedBy,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

func (s *pgService) List(orgID string) ([]Project, error) {
	rows, err := s.q.ListProjects(context.Background(), orgID)
	if err != nil {
		return nil, err
	}
	out := make([]Project, len(rows))
	for i, r := range rows {
		out[i] = repoToProject(r)
	}
	return out, nil
}

func (s *pgService) Create(orgID, name, description, userID string) (Project, error) {
	id := uuid.NewString()
	r, err := s.q.InsertProject(context.Background(), repo.InsertProjectParams{
		ID:          id,
		OrgID:       orgID,
		Name:        name,
		Description: sql.NullString{String: description, Valid: description != ""},
		CreatedBy:   userID,
	})
	if err != nil {
		return Project{}, err
	}
	return repoToProject(r), nil
}

func (s *pgService) Get(orgID, id string) (Project, error) {
	r, err := s.q.GetProject(context.Background(), repo.GetProjectParams{
		ID:    id,
		OrgID: orgID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return Project{}, ErrNotFound
	}
	if err != nil {
		return Project{}, err
	}
	return repoToProject(r), nil
}

func (s *pgService) Update(orgID, id, name, description string) (Project, error) {
	r, err := s.q.UpdateProject(context.Background(), repo.UpdateProjectParams{
		ID:          id,
		OrgID:       orgID,
		Column3:     name,
		Description: sql.NullString{String: description, Valid: description != ""},
	})
	if errors.Is(err, sql.ErrNoRows) {
		return Project{}, ErrNotFound
	}
	if err != nil {
		return Project{}, err
	}
	return repoToProject(r), nil
}

func (s *pgService) Delete(orgID, id string) error {
	res, err := s.q.DeleteProject(context.Background(), repo.DeleteProjectParams{
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

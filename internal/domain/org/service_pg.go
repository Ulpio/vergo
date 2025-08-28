package org

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound = errors.New("org not found")
)

type Service interface {
	Create(name, ownerUserID string) (Organization, error)
	Get(id string) (Organization, error)

	AddMember(orgID, userID, role string) error
	UpdateMember(orgID, userID, role string) error
	RemoveMember(orgID, userID string) error
	IsMember(orgID, userID string) (bool, string, error) // (ok, role)
}

type pgService struct{ db *sql.DB }

func NewPostgresService(db *sql.DB) Service { return &pgService{db: db} }

func (s *pgService) Create(name, ownerUserID string) (Organization, error) {
	id := uuid.NewString()
	now := time.Now()
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return Organization{}, err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(context.Background(),
		`INSERT INTO organizations (id,name,owner_user_id,created_at) VALUES ($1,$2,$3,$4)`,
		id, name, ownerUserID, now)
	if err != nil {
		return Organization{}, err
	}

	_, err = tx.ExecContext(context.Background(),
		`INSERT INTO memberships (org_id,user_id,role) VALUES ($1,$2,'owner')`,
		id, ownerUserID)
	if err != nil {
		return Organization{}, err
	}

	if err := tx.Commit(); err != nil {
		return Organization{}, err
	}

	return Organization{ID: id, Name: name, OwnerUser: ownerUserID, CreatedAt: now}, nil
}

func (s *pgService) Get(id string) (Organization, error) {
	var o Organization
	err := s.db.QueryRowContext(context.Background(),
		`SELECT id,name,owner_user_id,created_at FROM organizations WHERE id = $1`, id).
		Scan(&o.ID, &o.Name, &o.OwnerUser, &o.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Organization{}, ErrNotFound
	}
	return o, err
}

func (s *pgService) AddMember(orgID, userID, role string) error {
	_, err := s.db.ExecContext(context.Background(),
		`INSERT INTO memberships (org_id,user_id,role) VALUES ($1,$2,$3)
		 ON CONFLICT (org_id,user_id) DO UPDATE SET role = EXCLUDED.role`,
		orgID, userID, role)
	return err
}

func (s *pgService) UpdateMember(orgID, userID, role string) error {
	res, err := s.db.ExecContext(context.Background(),
		`UPDATE memberships SET role=$3 WHERE org_id=$1 AND user_id=$2`, orgID, userID, role)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *pgService) RemoveMember(orgID, userID string) error {
	_, err := s.db.ExecContext(context.Background(),
		`DELETE FROM memberships WHERE org_id=$1 AND user_id=$2`, orgID, userID)
	return err
}

func (s *pgService) IsMember(orgID, userID string) (bool, string, error) {
	var role string
	err := s.db.QueryRowContext(context.Background(),
		`SELECT role FROM memberships WHERE org_id=$1 AND user_id=$2`, orgID, userID).
		Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return false, "", nil
	}
	return err == nil, role, err
}

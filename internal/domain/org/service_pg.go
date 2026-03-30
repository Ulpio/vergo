package org

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/Ulpio/vergo/internal/repo"
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

	Delete(orgID string) error
}

type pgService struct {
	db *sql.DB
	q  *repo.Queries
}

func NewPostgresService(db *sql.DB, q *repo.Queries) Service {
	return &pgService{db: db, q: q}
}

func (s *pgService) Create(name, ownerUserID string) (Organization, error) {
	id := uuid.NewString()
	now := time.Now()
	ctx := context.Background()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Organization{}, err
	}
	defer tx.Rollback()

	qtx := s.q.WithTx(tx)

	err = qtx.InsertOrg(ctx, repo.InsertOrgParams{
		ID:          id,
		Name:        name,
		OwnerUserID: ownerUserID,
		CreatedAt:   now,
	})
	if err != nil {
		return Organization{}, err
	}

	err = qtx.UpsertMember(ctx, repo.UpsertMemberParams{
		OrgID:  id,
		UserID: ownerUserID,
		Role:   "owner",
	})
	if err != nil {
		return Organization{}, err
	}

	if err := tx.Commit(); err != nil {
		return Organization{}, err
	}

	return Organization{ID: id, Name: name, OwnerUser: ownerUserID, CreatedAt: now}, nil
}

func (s *pgService) Get(id string) (Organization, error) {
	r, err := s.q.GetOrg(context.Background(), id)
	if errors.Is(err, sql.ErrNoRows) {
		return Organization{}, ErrNotFound
	}
	if err != nil {
		return Organization{}, err
	}
	return Organization{
		ID:        r.ID,
		Name:      r.Name,
		OwnerUser: r.OwnerUserID,
		CreatedAt: r.CreatedAt,
	}, nil
}

func (s *pgService) AddMember(orgID, userID, role string) error {
	return s.q.UpsertMember(context.Background(), repo.UpsertMemberParams{
		OrgID:  orgID,
		UserID: userID,
		Role:   role,
	})
}

func (s *pgService) UpdateMember(orgID, userID, role string) error {
	res, err := s.q.UpdateMemberRole(context.Background(), repo.UpdateMemberRoleParams{
		OrgID:  orgID,
		UserID: userID,
		Role:   role,
	})
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
	return s.q.DeleteMember(context.Background(), repo.DeleteMemberParams{
		OrgID:  orgID,
		UserID: userID,
	})
}

func (s *pgService) IsMember(orgID, userID string) (bool, string, error) {
	role, err := s.q.GetMemberRole(context.Background(), repo.GetMemberRoleParams{
		OrgID:  orgID,
		UserID: userID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return false, "", nil
	}
	return err == nil, role, err
}

func (s *pgService) Delete(id string) error {
	return s.q.DeleteOrg(context.Background(), id)
}

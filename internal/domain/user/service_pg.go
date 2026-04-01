package user

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/Ulpio/vergo/internal/repo"
)

type pgService struct {
	db *sql.DB
	q  *repo.Queries
}

func NewPostgresService(db *sql.DB, q *repo.Queries) Service {
	return &pgService{db: db, q: q}
}

func (s *pgService) Signup(email, password string) (User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}
	id := uuid.NewString()
	err = s.q.InsertUser(context.Background(), repo.InsertUserParams{
		ID:           id,
		Email:        email,
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
	})
	if err != nil {
		return User{}, err
	}
	return User{ID: id, Email: email, PasswordHash: string(hash)}, nil
}

func (s *pgService) Login(email, password string) (User, error) {
	row, err := s.q.GetUserByEmail(context.Background(), email)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrInvalidLogin
	}
	if err != nil {
		return User{}, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(row.PasswordHash), []byte(password)); err != nil {
		return User{}, ErrInvalidLogin
	}
	return User{ID: row.ID, Email: row.Email, PasswordHash: row.PasswordHash}, nil
}

func (s *pgService) GetByID(id string) (User, error) {
	row, err := s.q.GetUserByID(context.Background(), id)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, err
	}
	return User{ID: row.ID, Email: row.Email, PasswordHash: row.PasswordHash}, nil
}

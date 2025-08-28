package user

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type pgService struct{ db *sql.DB }

func NewPostgresService(db *sql.DB) Service { return &pgService{db: db} }

func (s *pgService) Signup(email, password string) (User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}
	id := uuid.NewString()
	const q = `INSERT INTO users (id,email,password_hash,created_at) VALUES ($1,$2,$3,$4)`
	_, err = s.db.ExecContext(context.Background(), q, id, email, string(hash), time.Now())
	if err != nil {
		return User{}, err
	}
	return User{ID: id, Email: email, PasswordHash: string(hash)}, nil
}

func (s *pgService) Login(email, password string) (User, error) {
	const q = `SELECT id, email, password_hash FROM users WHERE email = $1`
	var u User
	err := s.db.QueryRowContext(context.Background(), q, email).Scan(&u.ID, &u.Email, &u.PasswordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrInvalidLogin
	}
	if err != nil {
		return User{}, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return User{}, ErrInvalidLogin
	}
	return u, nil
}

func (s *pgService) GetByID(id string) (User, error) {
	const q = `SELECT id, email, password_hash FROM users WHERE id = $1`
	var u User
	err := s.db.QueryRowContext(context.Background(), q, id).Scan(&u.ID, &u.Email, &u.PasswordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	return u, err
}

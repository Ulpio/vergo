package user

import (
	"errors"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailInUse   = errors.New("email already in use")
	ErrInvalidLogin = errors.New("invalid email or password")
	ErrNotFound     = errors.New("user not found")
)

type Service interface {
	Signup(email, password string) (User, error)
	Login(email, password string) (User, error)
	GetByID(id string) (User, error)
}

type memoryRepo struct {
	mu      sync.RWMutex
	byID    map[string]User
	byEmail map[string]string // email -> id
}

func NewMemoryService() Service {
	return &memoryRepo{
		byID:    make(map[string]User),
		byEmail: make(map[string]string),
	}
}

func (m *memoryRepo) Signup(email, password string) (User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.byEmail[email]; ok {
		return User{}, ErrEmailInUse
	}
	id := uuid.NewString()
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	u := User{ID: id, Email: email, PasswordHash: string(hash)}
	m.byID[id] = u
	m.byEmail[email] = id
	return u, nil
}

func (m *memoryRepo) Login(email, password string) (User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	id, ok := m.byEmail[email]
	if !ok {
		return User{}, ErrInvalidLogin
	}
	u := m.byID[id]
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return User{}, ErrInvalidLogin
	}
	return u, nil
}

func (m *memoryRepo) GetByID(id string) (User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	u, ok := m.byID[id]
	if !ok {
		return User{}, ErrNotFound
	}
	return u, nil
}

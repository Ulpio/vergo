package project

import (
	"errors"
	"sort"
	"sync"
	"time"

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

type memoryRepo struct {
	mu   sync.RWMutex
	data map[string]Project
}

func NewMemoryService() Service {
	return &memoryRepo{data: make(map[string]Project)}
}

func (m *memoryRepo) List(orgID string) ([]Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []Project
	for _, p := range m.data {
		if p.OrgID == orgID {
			out = append(out, p)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}

func (m *memoryRepo) Create(orgID, name, description, userID string) (Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	now := time.Now()
	p := Project{
		ID:          uuid.NewString(),
		OrgID:       orgID,
		Name:        name,
		Description: description,
		CreatedBy:   userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	m.data[p.ID] = p
	return p, nil
}

func (m *memoryRepo) Get(orgID, id string) (Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.data[id]
	if !ok || p.OrgID != orgID {
		return Project{}, ErrNotFound
	}
	return p, nil
}

func (m *memoryRepo) Update(orgID, id, name, description string) (Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.data[id]
	if !ok || p.OrgID != orgID {
		return Project{}, ErrNotFound
	}
	if name != "" {
		p.Name = name
	}
	p.Description = description
	p.UpdatedAt = time.Now()
	m.data[id] = p
	return p, nil
}

func (m *memoryRepo) Delete(orgID, id string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.data[id]
	if !ok || p.OrgID != orgID {
		return ErrNotFound
	}
	delete(m.data, id)
	return nil
}

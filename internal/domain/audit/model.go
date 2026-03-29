package audit

import (
	"encoding/json"
	"time"
)

type Event struct {
	OrgID     string    `json:"org_id"`
	ActorID   string    `json:"actor_id"`
	Action    string    `json:"action"`
	Entity    string    `json:"entity"`
	EntityID  string    `json:"entity_id"`
	Timestamp time.Time `json:"timestamp"`
	Metadata  Metadata  `json:"metadata,omitempty"`
}

// Metadata holds optional before/after state for mutations.
type Metadata struct {
	Before json.RawMessage `json:"before,omitempty"`
	After  json.RawMessage `json:"after,omitempty"`
}

type ListParams struct {
	OrgID   string
	ActorID *string
	Action  *string
	Entity  *string
	Since   *time.Time
	Until   *time.Time
	Limit   int
	Offset  int
}

type Service interface {
	Record(e Event) error
	List(p ListParams) ([]Event, error)
}

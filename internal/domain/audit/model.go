package audit

import "time"

type Event struct {
	OrgID     string    `json:"org_id"`
	ActorID   string    `json:"actor_id"`
	Action    string    `json:"action"`
	Entity    string    `json:"entity"`
	EntityID  string    `json:"entity_id"`
	Timestamp time.Time `json:"timestamp"`
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

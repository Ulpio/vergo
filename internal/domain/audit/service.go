package audit

import (
	"log"
	"time"
)

type Event struct {
	OrgID     string
	ActorID   string
	Action    string
	Entity    string
	EntityID  string
	Timestamp time.Time
}

type Service interface {
	Record(e Event) error
}

type memoryAudit struct{}

func NewMemoryService() Service { return &memoryAudit{} }

func (m *memoryAudit) Record(e Event) error {
	log.Printf("[AUDIT] org=%s actor=%s action=%s entity=%s id=%s at=%s",
		e.OrgID, e.ActorID, e.Action, e.Entity, e.EntityID, e.Timestamp.Format(time.RFC3339))
	return nil
}

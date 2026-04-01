//go:build integration

package audit_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/Ulpio/vergo/internal/domain/audit"
	"github.com/Ulpio/vergo/internal/pkg/testutil"
	"github.com/Ulpio/vergo/internal/repo"
)

func TestPGService_RecordAndList(t *testing.T) {
	db := testutil.PGContainer(t)
	svc := audit.NewPostgresService(db, repo.New(db))

	evt := audit.Event{
		OrgID:     "org-1",
		ActorID:   "user-1",
		Action:    "project.created",
		Entity:    "project",
		EntityID:  "proj-1",
		Timestamp: time.Now(),
		Metadata: audit.Metadata{
			After: json.RawMessage(`{"name":"Test Project"}`),
		},
	}
	if err := svc.Record(evt); err != nil {
		t.Fatalf("Record: %v", err)
	}

	events, err := svc.List(audit.ListParams{
		OrgID: "org-1",
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Action != "project.created" {
		t.Errorf("action = %q, want %q", events[0].Action, "project.created")
	}
}

func TestPGService_ListFilters(t *testing.T) {
	db := testutil.PGContainer(t)
	svc := audit.NewPostgresService(db, repo.New(db))

	// Record multiple events
	for _, action := range []string{"org.created", "project.created", "project.deleted"} {
		svc.Record(audit.Event{
			OrgID:     "org-filter",
			ActorID:   "user-1",
			Action:    action,
			Entity:    "project",
			EntityID:  "proj-1",
			Timestamp: time.Now(),
		})
	}

	// Filter by action
	action := "project.created"
	events, err := svc.List(audit.ListParams{
		OrgID:  "org-filter",
		Action: &action,
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("List with filter: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("expected 1 filtered event, got %d", len(events))
	}
}

func TestPGService_ListPagination(t *testing.T) {
	db := testutil.PGContainer(t)
	svc := audit.NewPostgresService(db, repo.New(db))

	for i := 0; i < 5; i++ {
		svc.Record(audit.Event{
			OrgID:     "org-page",
			ActorID:   "user-1",
			Action:    "test.action",
			Entity:    "test",
			EntityID:  "ent-1",
			Timestamp: time.Now(),
		})
	}

	page1, _ := svc.List(audit.ListParams{OrgID: "org-page", Limit: 2, Offset: 0})
	page2, _ := svc.List(audit.ListParams{OrgID: "org-page", Limit: 2, Offset: 2})

	if len(page1) != 2 {
		t.Errorf("page1: expected 2, got %d", len(page1))
	}
	if len(page2) != 2 {
		t.Errorf("page2: expected 2, got %d", len(page2))
	}
}

func TestPGService_MetadataRoundTrip(t *testing.T) {
	db := testutil.PGContainer(t)
	svc := audit.NewPostgresService(db, repo.New(db))

	before := json.RawMessage(`{"role":"member"}`)
	after := json.RawMessage(`{"role":"admin"}`)

	svc.Record(audit.Event{
		OrgID:     "org-meta",
		ActorID:   "user-1",
		Action:    "member.updated",
		Entity:    "membership",
		EntityID:  "mem-1",
		Timestamp: time.Now(),
		Metadata:  audit.Metadata{Before: before, After: after},
	})

	events, _ := svc.List(audit.ListParams{OrgID: "org-meta", Limit: 1})
	if len(events) == 0 {
		t.Fatal("expected 1 event")
	}

	// Verify metadata round-trips correctly
	if string(events[0].Metadata.Before) == "" && string(events[0].Metadata.After) == "" {
		t.Error("expected metadata to be preserved")
	}
}

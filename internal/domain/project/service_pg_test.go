//go:build integration

package project_test

import (
	"testing"

	"github.com/Ulpio/vergo/internal/domain/org"
	"github.com/Ulpio/vergo/internal/domain/project"
	"github.com/Ulpio/vergo/internal/domain/user"
	"github.com/Ulpio/vergo/internal/pkg/testutil"
	"github.com/Ulpio/vergo/internal/repo"
)

func setup(t *testing.T) (project.Service, string, string) {
	t.Helper()
	db := testutil.PGContainer(t)
	userSvc := user.NewPostgresService(db, repo.New(db))
	orgSvc := org.NewPostgresService(db)

	u, _ := userSvc.Signup("projuser@test.com", "pass")
	o, _ := orgSvc.Create("ProjOrg", u.ID)

	return project.NewPostgresService(db), o.ID, u.ID
}

func TestPGService_CreateProject(t *testing.T) {
	svc, orgID, userID := setup(t)

	p, err := svc.Create(orgID, "My Project", "A test project", userID)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if p.Name != "My Project" {
		t.Errorf("name = %q, want %q", p.Name, "My Project")
	}
	if p.OrgID != orgID {
		t.Errorf("orgID = %q, want %q", p.OrgID, orgID)
	}
}

func TestPGService_ListProjects(t *testing.T) {
	svc, orgID, userID := setup(t)

	svc.Create(orgID, "Proj A", "desc", userID)
	svc.Create(orgID, "Proj B", "desc", userID)

	list, err := svc.List(orgID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 projects, got %d", len(list))
	}
}

func TestPGService_GetProject(t *testing.T) {
	svc, orgID, userID := setup(t)
	created, _ := svc.Create(orgID, "GetMe", "desc", userID)

	found, err := svc.Get(orgID, created.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if found.Name != "GetMe" {
		t.Errorf("name = %q, want %q", found.Name, "GetMe")
	}
}

func TestPGService_UpdateProject(t *testing.T) {
	svc, orgID, userID := setup(t)
	created, _ := svc.Create(orgID, "Old Name", "old desc", userID)

	updated, err := svc.Update(orgID, created.ID, "New Name", "new desc")
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Name != "New Name" {
		t.Errorf("name = %q, want %q", updated.Name, "New Name")
	}
	if updated.Description != "new desc" {
		t.Errorf("desc = %q, want %q", updated.Description, "new desc")
	}
}

func TestPGService_DeleteProject(t *testing.T) {
	svc, orgID, userID := setup(t)
	created, _ := svc.Create(orgID, "DeleteMe", "desc", userID)

	if err := svc.Delete(orgID, created.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := svc.Get(orgID, created.ID)
	if err == nil {
		t.Error("expected error after delete")
	}
}

//go:build integration

package org_test

import (
	"testing"

	"github.com/Ulpio/vergo/internal/domain/org"
	"github.com/Ulpio/vergo/internal/domain/user"
	"github.com/Ulpio/vergo/internal/pkg/testutil"
)

func setupOrg(t *testing.T) (org.Service, user.User) {
	t.Helper()
	db := testutil.PGContainer(t)
	userSvc := user.NewPostgresService(db)
	u, err := userSvc.Signup("orgowner@test.com", "pass123")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	return org.NewPostgresService(db), u
}

func TestPGService_CreateOrg(t *testing.T) {
	svc, owner := setupOrg(t)

	o, err := svc.Create("Acme Corp", owner.ID)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if o.Name != "Acme Corp" {
		t.Errorf("name = %q, want %q", o.Name, "Acme Corp")
	}
	if o.OwnerUser != owner.ID {
		t.Errorf("owner = %q, want %q", o.OwnerUser, owner.ID)
	}
}

func TestPGService_GetOrg(t *testing.T) {
	svc, owner := setupOrg(t)
	created, _ := svc.Create("Get Corp", owner.ID)

	found, err := svc.Get(created.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if found.Name != "Get Corp" {
		t.Errorf("name = %q, want %q", found.Name, "Get Corp")
	}
}

func TestPGService_GetOrg_NotFound(t *testing.T) {
	svc, _ := setupOrg(t)
	_, err := svc.Get("nonexistent")
	if err != org.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestPGService_Membership(t *testing.T) {
	db := testutil.PGContainer(t)
	userSvc := user.NewPostgresService(db)
	orgSvc := org.NewPostgresService(db)

	owner, _ := userSvc.Signup("owner@test.com", "pass")
	member, _ := userSvc.Signup("member@test.com", "pass")

	o, _ := orgSvc.Create("Team", owner.ID)

	// Owner should be a member with role "owner"
	ok, role, err := orgSvc.IsMember(o.ID, owner.ID)
	if err != nil || !ok || role != "owner" {
		t.Errorf("owner membership: ok=%v, role=%q, err=%v", ok, role, err)
	}

	// Add member
	if err := orgSvc.AddMember(o.ID, member.ID, "member"); err != nil {
		t.Fatalf("AddMember: %v", err)
	}
	ok, role, _ = orgSvc.IsMember(o.ID, member.ID)
	if !ok || role != "member" {
		t.Errorf("member: ok=%v, role=%q", ok, role)
	}

	// Update role
	if err := orgSvc.UpdateMember(o.ID, member.ID, "admin"); err != nil {
		t.Fatalf("UpdateMember: %v", err)
	}
	_, role, _ = orgSvc.IsMember(o.ID, member.ID)
	if role != "admin" {
		t.Errorf("updated role = %q, want admin", role)
	}

	// Remove member
	if err := orgSvc.RemoveMember(o.ID, member.ID); err != nil {
		t.Fatalf("RemoveMember: %v", err)
	}
	ok, _, _ = orgSvc.IsMember(o.ID, member.ID)
	if ok {
		t.Error("expected member to be removed")
	}
}

func TestPGService_DeleteOrg(t *testing.T) {
	svc, owner := setupOrg(t)
	o, _ := svc.Create("DeleteMe", owner.ID)

	if err := svc.Delete(o.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := svc.Get(o.ID)
	if err != org.ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

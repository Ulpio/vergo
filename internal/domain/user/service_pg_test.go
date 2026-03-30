//go:build integration

package user_test

import (
	"testing"

	"github.com/Ulpio/vergo/internal/domain/user"
	"github.com/Ulpio/vergo/internal/pkg/testutil"
	"github.com/Ulpio/vergo/internal/repo"
)

func TestPGService_Signup(t *testing.T) {
	db := testutil.PGContainer(t)
	svc := user.NewPostgresService(db, repo.New(db))

	u, err := svc.Signup("alice@example.com", "password123")
	if err != nil {
		t.Fatalf("Signup failed: %v", err)
	}
	if u.ID == "" {
		t.Error("expected non-empty user ID")
	}
	if u.Email != "alice@example.com" {
		t.Errorf("email = %q, want %q", u.Email, "alice@example.com")
	}
}

func TestPGService_SignupDuplicate(t *testing.T) {
	db := testutil.PGContainer(t)
	svc := user.NewPostgresService(db, repo.New(db))

	_, err := svc.Signup("dup@example.com", "pass123")
	if err != nil {
		t.Fatalf("first signup: %v", err)
	}
	_, err = svc.Signup("dup@example.com", "pass456")
	if err == nil {
		t.Error("expected error on duplicate signup")
	}
}

func TestPGService_Login(t *testing.T) {
	db := testutil.PGContainer(t)
	svc := user.NewPostgresService(db, repo.New(db))

	_, err := svc.Signup("bob@example.com", "secretpass")
	if err != nil {
		t.Fatalf("signup: %v", err)
	}

	u, err := svc.Login("bob@example.com", "secretpass")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if u.Email != "bob@example.com" {
		t.Errorf("email = %q, want %q", u.Email, "bob@example.com")
	}
}

func TestPGService_LoginWrongPassword(t *testing.T) {
	db := testutil.PGContainer(t)
	svc := user.NewPostgresService(db, repo.New(db))

	_, _ = svc.Signup("charlie@example.com", "correct")
	_, err := svc.Login("charlie@example.com", "wrong")
	if err == nil {
		t.Error("expected error on wrong password")
	}
}

func TestPGService_GetByID(t *testing.T) {
	db := testutil.PGContainer(t)
	svc := user.NewPostgresService(db, repo.New(db))

	created, _ := svc.Signup("dave@example.com", "pass")
	found, err := svc.GetByID(created.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if found.Email != "dave@example.com" {
		t.Errorf("email = %q, want %q", found.Email, "dave@example.com")
	}
}

func TestPGService_GetByID_NotFound(t *testing.T) {
	db := testutil.PGContainer(t)
	svc := user.NewPostgresService(db, repo.New(db))

	_, err := svc.GetByID("nonexistent-id")
	if err == nil {
		t.Error("expected error for nonexistent user")
	}
}

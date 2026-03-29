// Package testutil provides shared test helpers for integration tests.
package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/Ulpio/vergo/internal/pkg/db"
)

// PGContainer starts a PostgreSQL container and returns a connected *sql.DB
// with all migrations applied. The container is automatically terminated when
// the test finishes.
func PGContainer(t *testing.T) *sql.DB {
	t.Helper()
	ctx := context.Background()

	pgC, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("vergo_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	t.Cleanup(func() {
		if err := pgC.Terminate(ctx); err != nil {
			t.Logf("terminate postgres container: %v", err)
		}
	})

	host, err := pgC.Host(ctx)
	if err != nil {
		t.Fatalf("get container host: %v", err)
	}
	port, err := pgC.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("get container port: %v", err)
	}

	dsn := fmt.Sprintf("host=%s port=%s user=test password=test dbname=vergo_test sslmode=disable",
		host, port.Port())

	database, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	// Wait for connection
	for i := 0; i < 10; i++ {
		if err := database.Ping(); err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Run migrations
	if err := db.RunMigrations(database); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	return database
}

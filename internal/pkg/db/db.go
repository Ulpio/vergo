package db

import (
	"database/sql"
	"embed"
	"fmt"
	"log"

	_ "github.com/lib/pq"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/Ulpio/vergo/internal/pkg/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.up.sql
var migrationsFS embed.FS

func Open(cfg config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName, cfg.DBSSL)
	return sql.Open("pgx", dsn)
}

func RunMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("erro ao criar driver de migrations: %w", err)
	}

	src, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("erro ao carregar migrations do embed: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "postgres", driver)
	if err != nil {
		return fmt.Errorf("erro ao criar instância de migrate: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("erro ao executar migrations: %w", err)
	}

	if err == migrate.ErrNoChange {
		log.Println("✅ migrations já estão atualizadas")
	} else {
		log.Println("✅ migrations aplicadas com sucesso")
	}
	return nil
}

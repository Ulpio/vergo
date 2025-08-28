package db

import (
	"database/sql"
	"fmt"

	"github.com/Ulpio/vergo/internal/pkg/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func Open(cfg config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName, cfg.DBSSL)
	return sql.Open("pgx", dsn)
}

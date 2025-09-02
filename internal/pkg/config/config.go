package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	// App
	AppPort    int
	AppEnv     string
	AppVersion string

	// JWT
	JWTAccessTTLMinutes int
	JWTRefreshTTLDays   int
	JWTAccessSecret     string
	JWTRefreshSecret    string

	// Database (Postgres)
	DBHost string
	DBPort int
	DBUser string
	DBPass string
	DBName string
	DBSSL  string

	// Storage (S3)
	// Se você não configurar nada no .env, continuam válidos e inofensivos.
	StorageAllowedTypes []string // ex.: image/png,image/jpeg,application/pdf
	StorageMaxMB        int      // limite por arquivo em MB (0 = sem limite)
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getint(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func Load() Config {
	return Config{
		// App
		AppPort:    getint("APP_PORT", 8080),
		AppEnv:     getenv("APP_ENV", "dev"),
		AppVersion: getenv("APP_VERSION", "0.1.0"),

		// JWT
		JWTAccessTTLMinutes: getint("JWT_ACCESS_TTL_MINUTES", 15),
		JWTRefreshTTLDays:   getint("JWT_REFRESH_TTL_DAYS", 14),
		JWTAccessSecret:     getenv("JWT_ACCESS_SECRET", "dev-access"),
		JWTRefreshSecret:    getenv("JWT_REFRESH_SECRET", "dev-refresh"),

		// DB
		DBHost: getenv("DB_HOST", "localhost"),
		DBPort: getint("DB_PORT", 5432),
		DBUser: getenv("DB_USER", "app"),
		DBPass: getenv("DB_PASSWORD", "app"),
		DBName: getenv("DB_NAME", "vergo"),
		DBSSL:  getenv("DB_SSLMODE", "disable"),

		// Storage
		StorageAllowedTypes: splitCSV(getenv("STORAGE_ALLOWED_TYPES", "")),
		StorageMaxMB:        getint("STORAGE_MAX_MB", 25),
	}
}

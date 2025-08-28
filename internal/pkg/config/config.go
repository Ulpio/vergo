package config

import (
	"os"
	"strconv"
)

type Config struct {
	AppPort    int
	AppEnv     string
	AppVersion string

	JWTAccessTTLMinutes int
	JWTRefreshTTLDays   int
	JWTAccessSecret     string
	JWTRefreshSecret    string

	DBHost string
	DBPort int
	DBUser string
	DBPass string
	DBName string
	DBSSL  string
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
func getInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func Load() Config {
	return Config{
		AppPort:    getInt("APP_PORT", 8080),
		AppEnv:     getEnv("APP_ENV", "dev"),
		AppVersion: getEnv("APP_Version", "0.1.0"),

		JWTAccessTTLMinutes: getInt("JWT_ACCESS_TTL_MINUTES", 15),
		JWTRefreshTTLDays:   getInt("JWT_REFRESH_TTL_DAYS", 14),
		JWTAccessSecret:     getEnv("JWT_ACCESS_SECRET", "dev-access-secret"),   // troque em prod
		JWTRefreshSecret:    getEnv("JWT_REFRESH_SECRET", "dev-refresh-secret"), // troque em prod

		DBHost: getEnv("DB_HOST", "localhost"),
		DBPort: getInt("DB_PORT", 5432),
		DBUser: getEnv("DB_USER", "app"),
		DBPass: getEnv("DB_PASSWORD", "app"),
		DBName: getEnv("DB_NAME", "vergo"),
		DBSSL:  getEnv("DB_SSLMODE", "disable"),
	}
}

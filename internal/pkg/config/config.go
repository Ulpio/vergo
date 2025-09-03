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

	// S3 (carregado do .env)
	S3Region          string
	S3Bucket          string
	S3Endpoint        string
	S3ForcePathStyle  bool
	S3AccessKeyID     string
	S3SecretAccessKey string
	AWSSessionToken   string // opcional (para credenciais temporárias)

	// Storage policy (opcional)
	StorageAllowedTypes []string
	StorageMaxMB        int
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
func getbool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
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

		// S3
		S3Region:          getenv("S3_REGION", "us-east-1"),
		S3Bucket:          getenv("S3_BUCKET", ""),
		S3Endpoint:        getenv("S3_ENDPOINT", ""),
		S3ForcePathStyle:  getbool("S3_FORCE_PATH_STYLE", false),
		S3AccessKeyID:     getenv("S3_ACCESS_KEY_ID", ""),
		S3SecretAccessKey: getenv("S3_SECRET_ACCESS_KEY", ""),
		AWSSessionToken:   getenv("AWS_SESSION_TOKEN", ""),

		// Storage policy
		StorageAllowedTypes: splitCSV(getenv("STORAGE_ALLOWED_TYPES", "")),
		StorageMaxMB:        getint("STORAGE_MAX_MB", 25),
	}
}

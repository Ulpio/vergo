package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupLoggingTest(handler gin.HandlerFunc) (*httptest.ResponseRecorder, *bytes.Buffer) {
	var buf bytes.Buffer
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))

	r := gin.New()
	r.Use(Logging())
	r.GET("/test", handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	return w, &buf
}

func TestLogging_OK(t *testing.T) {
	w, buf := setupLoggingTest(func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSON log: %v\nraw: %s", err, buf.String())
	}

	if got := entry["method"]; got != "GET" {
		t.Errorf("method = %v, want GET", got)
	}
	if got := entry["path"]; got != "/test" {
		t.Errorf("path = %v, want /test", got)
	}
	if got, ok := entry["status"].(float64); !ok || int(got) != 200 {
		t.Errorf("status = %v, want 200", entry["status"])
	}
	if _, ok := entry["latency"]; !ok {
		t.Error("latency field missing")
	}
	if got := entry["level"]; got != "INFO" {
		t.Errorf("level = %v, want INFO", got)
	}
}

func TestLogging_NotFound(t *testing.T) {
	var buf bytes.Buffer
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))

	r := gin.New()
	r.Use(Logging())
	// No route registered — Gin returns 404

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSON log: %v\nraw: %s", err, buf.String())
	}

	if got := entry["level"]; got != "WARN" {
		t.Errorf("level = %v, want WARN for 404", got)
	}
}

func TestLogging_ServerError(t *testing.T) {
	_, buf := setupLoggingTest(func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "fail"})
	})

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSON log: %v\nraw: %s", err, buf.String())
	}

	if got := entry["level"]; got != "ERROR" {
		t.Errorf("level = %v, want ERROR for 500", got)
	}
}

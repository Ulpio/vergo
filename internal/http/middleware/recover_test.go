package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRecover_NoPanic(t *testing.T) {
	r := gin.New()
	r.Use(Recover())
	r.GET("/ok", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestRecover_WithPanic(t *testing.T) {
	var buf bytes.Buffer
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))

	r := gin.New()
	r.Use(Recover())
	r.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	r.ServeHTTP(w, req)

	// Verify 500 response
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", w.Code)
	}

	// Verify response body
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON body: %v", err)
	}
	if got := body["error"]; got != "internal_server_error" {
		t.Errorf("error = %v, want internal_server_error", got)
	}

	// Verify structured log
	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSON log: %v\nraw: %s", err, buf.String())
	}
	if got := entry["msg"]; got != "panic recovered" {
		t.Errorf("msg = %v, want panic recovered", got)
	}
	if got, ok := entry["error"].(string); !ok || got != "boom" {
		t.Errorf("error = %v, want boom", entry["error"])
	}
	if got, ok := entry["stack"].(string); !ok || got == "" {
		t.Error("stack trace missing or empty")
	}
}

func TestRecover_PanicDoesNotLeak(t *testing.T) {
	slog.SetDefault(slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil)))

	r := gin.New()
	r.Use(Recover())
	r.GET("/secret-panic", func(c *gin.Context) {
		panic("sensitive internal error details")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/secret-panic", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	if strings.Contains(body, "sensitive internal error details") {
		t.Error("panic message leaked in response body")
	}
}

package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/Ulpio/vergo/internal/pkg/ratelimit"
)

func TestRateLimit_AllowsWithinBurst(t *testing.T) {
	limiter := ratelimit.New(10, 3)
	defer limiter.Stop()

	r := gin.New()
	r.Use(RateLimit(limiter))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	for i := range 3 {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want 200", i+1, w.Code)
		}
		if w.Header().Get("X-RateLimit-Limit") != "3" {
			t.Errorf("X-RateLimit-Limit = %s, want 3", w.Header().Get("X-RateLimit-Limit"))
		}
	}
}

func TestRateLimit_RejectsOverBurst(t *testing.T) {
	limiter := ratelimit.New(10, 2)
	defer limiter.Stop()

	r := gin.New()
	r.Use(RateLimit(limiter))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// Exhaust burst
	for range 2 {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		r.ServeHTTP(w, req)
	}

	// 3rd should be 429
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", w.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if body["error"] != "rate_limit_exceeded" {
		t.Errorf("error = %v, want rate_limit_exceeded", body["error"])
	}
	if w.Header().Get("Retry-After") != "1" {
		t.Errorf("Retry-After = %s, want 1", w.Header().Get("Retry-After"))
	}
}

func TestRateLimit_HeadersPresent(t *testing.T) {
	limiter := ratelimit.New(10, 5)
	defer limiter.Stop()

	r := gin.New()
	r.Use(RateLimit(limiter))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	if w.Header().Get("X-RateLimit-Limit") == "" {
		t.Error("X-RateLimit-Limit header missing")
	}
	if w.Header().Get("X-RateLimit-Remaining") == "" {
		t.Error("X-RateLimit-Remaining header missing")
	}
}

func TestRateLimit_UsesTenantKeyWhenAvailable(t *testing.T) {
	limiter := ratelimit.New(10, 1)
	defer limiter.Stop()

	r := gin.New()
	// Simulate tenant middleware by setting org_id
	r.Use(func(c *gin.Context) {
		c.Set("org_id", "org-123")
		c.Next()
	})
	r.Use(RateLimit(limiter))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// First request allowed (tenant:org-123 bucket)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	// Second request rejected (same tenant bucket)
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429 (same tenant)", w2.Code)
	}
}

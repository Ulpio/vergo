package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Ulpio/vergo/internal/pkg/config"
	"github.com/gin-gonic/gin"
)

// FuzzAuthHeader feeds arbitrary Authorization header values into the Auth
// middleware to ensure it never panics and always returns a safe HTTP status.
func FuzzAuthHeader(f *testing.F) {
	// Seed corpus: common patterns and edge cases
	f.Add("")
	f.Add("Bearer ")
	f.Add("Bearer valid-looking-token")
	f.Add("bearer lowercase")
	f.Add("Basic dXNlcjpwYXNz")
	f.Add("BearerNoSpace")
	f.Add("Bearer\t\ttabs")
	f.Add("\x00\xff\xfe")
	f.Add("Bearer " + string(make([]byte, 10000)))

	f.Fuzz(func(t *testing.T, header string) {
		w := httptest.NewRecorder()
		_, r := gin.CreateTestContext(w)

		cfg := config.Config{JWTAccessSecret: "fuzz-secret"}
		r.GET("/test", Auth(cfg), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		if header != "" {
			req.Header.Set("Authorization", header)
		}
		r.ServeHTTP(w, req)

		// Must be a valid HTTP status, never panic
		status := w.Code
		if status < 200 || status >= 600 {
			t.Errorf("invalid status code: %d", status)
		}
	})
}

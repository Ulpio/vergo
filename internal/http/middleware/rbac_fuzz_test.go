package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// FuzzRequireRole feeds arbitrary role strings into the RBAC middleware to
// ensure it never panics and always enforces access control correctly.
func FuzzRequireRole(f *testing.F) {
	// Seed: valid roles, empty, unknown roles, special chars
	f.Add("owner", "owner")
	f.Add("admin", "member")
	f.Add("member", "admin")
	f.Add("", "")
	f.Add("superadmin", "owner")
	f.Add("\x00", "member")
	f.Add("owner", "\xff\xfe")
	f.Add("OWNER", "owner")

	f.Fuzz(func(t *testing.T, minRole, actualRole string) {
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)

		// Register handler with RBAC middleware
		r.GET("/test", func(c *gin.Context) {
			c.Set(ctxRole, actualRole)
			c.Next()
		}, RequireRole(minRole), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
		r.ServeHTTP(w, c.Request)

		// Must always return a valid HTTP status (never panic)
		status := w.Code
		if status != http.StatusOK && status != http.StatusForbidden {
			t.Errorf("unexpected status %d for minRole=%q actualRole=%q", status, minRole, actualRole)
		}

		// Known valid roles: verify correctness
		minLevel := roleOrder[minRole]
		actualLevel := roleOrder[actualRole]
		if minLevel > 0 && actualLevel > 0 {
			if actualLevel >= minLevel && status != http.StatusOK {
				t.Errorf("should allow: actualRole=%q (level %d) >= minRole=%q (level %d), got %d",
					actualRole, actualLevel, minRole, minLevel, status)
			}
			if actualLevel < minLevel && status != http.StatusForbidden {
				t.Errorf("should deny: actualRole=%q (level %d) < minRole=%q (level %d), got %d",
					actualRole, actualLevel, minRole, minLevel, status)
			}
		}
	})
}

// FuzzRoleOrder verifies that unknown roles always get level 0 (no access)
// and known roles maintain their expected hierarchy.
func FuzzRoleOrder(f *testing.F) {
	f.Add("member")
	f.Add("admin")
	f.Add("owner")
	f.Add("")
	f.Add("root")
	f.Add("ADMIN")

	f.Fuzz(func(t *testing.T, role string) {
		level := roleOrder[role]
		switch role {
		case "member":
			if level != 1 {
				t.Errorf("member should be level 1, got %d", level)
			}
		case "admin":
			if level != 2 {
				t.Errorf("admin should be level 2, got %d", level)
			}
		case "owner":
			if level != 3 {
				t.Errorf("owner should be level 3, got %d", level)
			}
		default:
			if level != 0 {
				t.Errorf("unknown role %q should be level 0, got %d", role, level)
			}
		}
	})
}

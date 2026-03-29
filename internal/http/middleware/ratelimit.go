package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/Ulpio/vergo/internal/pkg/ratelimit"
)

// RateLimit returns a middleware that limits requests using a token bucket
// algorithm. It applies rate limiting by client IP, and additionally by
// tenant (org_id) if available in the context.
//
// Response headers:
//   - X-RateLimit-Limit:     max requests (burst)
//   - X-RateLimit-Remaining: tokens left
//   - Retry-After:           seconds to wait (only on 429)
func RateLimit(limiter *ratelimit.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := rateLimitKey(c)

		if !limiter.Allow(key) {
			remaining := limiter.Remaining(key)
			c.Header("X-RateLimit-Limit", strconv.Itoa(limiter.Burst()))
			c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
			c.Header("Retry-After", "1")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate_limit_exceeded",
			})
			return
		}

		remaining := limiter.Remaining(key)
		c.Header("X-RateLimit-Limit", strconv.Itoa(limiter.Burst()))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))

		c.Next()
	}
}

// rateLimitKey builds the rate limit key. If an org_id is present in the
// context (set by Tenant middleware), it uses "tenant:<org_id>".
// Otherwise, it falls back to "ip:<client_ip>".
func rateLimitKey(c *gin.Context) string {
	if orgID, ok := OrgID(c); ok {
		return "tenant:" + orgID
	}
	return "ip:" + c.ClientIP()
}

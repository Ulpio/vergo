package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Ulpio/vergo/internal/domain/billing"
)

// RequirePlan checks that the org has an active subscription with the given plan or higher.
func RequirePlan(billSvc billing.Service, minPlan string) gin.HandlerFunc {
	order := map[string]int{"free": 0, "pro": 1, "enterprise": 2}
	minLevel := order[minPlan]

	return func(c *gin.Context) {
		orgID, ok := OrgID(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing_org_id"})
			return
		}

		sub, err := billSvc.GetSubscription(orgID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "plan_check_failed"})
			return
		}

		level := order[sub.Plan]
		if level < minLevel {
			c.AbortWithStatusJSON(http.StatusPaymentRequired, gin.H{
				"error":        "plan_required",
				"required_plan": minPlan,
				"current_plan":  sub.Plan,
			})
			return
		}
		c.Next()
	}
}

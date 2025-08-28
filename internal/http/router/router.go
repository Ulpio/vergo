package router

import (
	"net/http"

	"github.com/Ulpio/vergo/internal/domain/user"
	"github.com/Ulpio/vergo/internal/http/handlers"
	"github.com/Ulpio/vergo/internal/pkg/config"
	"github.com/gin-gonic/gin"
)

// Register registra todas as rotas v1.
func Register(v1 *gin.RouterGroup) {
	cfg := config.Load()
	usr := user.NewMemoryService()
	ah := handlers.NewAuthHandler(cfg, usr)

	auth := v1.Group("/auth")
	{
		auth.POST("/signup", ah.Signup)
		auth.POST("/login", ah.Login)
		auth.POST("/refresh", ah.Refresh)
		auth.POST("/forgot-password", notImplemented("auth.forgot_password"))
		auth.POST("/reset-password", notImplemented("auth.reset_password"))
	}
	// Organização e Membros
	orgs := v1.Group("/orgs")
	{
		orgs.POST("", notImplemented("orgs.create"))
		orgs.GET("/:id", notImplemented("orgs.get"))
		orgs.PATCH("/:id", notImplemented("orgs.update"))

		orgs.POST("/:id/invite", notImplemented("orgs.invite"))
		orgs.POST("/:id/members", notImplemented("orgs.members.add"))
		orgs.PATCH("/:id/members/:userId", notImplemented("orgs.members.update"))
		orgs.DELETE("/:id/members/:userId", notImplemented("orgs.members.remove"))
	}

	v1.GET("/context", notImplemented("context.get"))
	v1.POST("/context", notImplemented("context.set"))

	// Projects (multi-tenant)
	projects := v1.Group("/projects")
	{
		projects.GET("", notImplemented("projects.list"))
		projects.POST("", notImplemented("projects.create"))
		projects.GET("/:id", notImplemented("projects.get"))
		projects.PATCH("/:id", notImplemented("projects.update"))
		projects.DELETE("/:id", notImplemented("projects.delete"))
	}

	//  Auditoria
	v1.GET("/audit", notImplemented("audit.list"))

	//  API Keys
	keys := v1.Group("/api-keys")
	{
		keys.POST("", notImplemented("apikeys.create"))
		keys.GET("", notImplemented("apikeys.list"))
		keys.DELETE("/:id", notImplemented("apikeys.delete"))
	}

	//  Webhooks (saída)
	wh := v1.Group("/webhooks")
	{
		wh.POST("/endpoints", notImplemented("webhooks.endpoints.create"))
		wh.GET("/endpoints", notImplemented("webhooks.endpoints.list"))
		wh.PATCH("/endpoints/:id", notImplemented("webhooks.endpoints.update"))
		wh.POST("/test", notImplemented("webhooks.test"))
	}

	//  Billing (Stripe)
	billing := v1.Group("/billing")
	{
		billing.POST("/checkout-session", notImplemented("billing.checkout_session"))
		billing.GET("/subscription", notImplemented("billing.subscription.get"))
		billing.POST("/webhook", notImplemented("billing.webhook"))
	}

	//  Storage (S3 presign)
	storage := v1.Group("/storage")
	{
		storage.POST("/presign", notImplemented("storage.presign"))
	}
}

func notImplemented(code string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusNotImplemented, gin.H{
			"error": gin.H{
				"code":    code,
				"message": "not_implemented",
			},
		})
	}
}

package router

import (
	"net/http"

	"github.com/Ulpio/vergo/internal/domain/audit"
	"github.com/Ulpio/vergo/internal/domain/project"
	"github.com/Ulpio/vergo/internal/domain/user"
	"github.com/Ulpio/vergo/internal/http/handlers"
	"github.com/Ulpio/vergo/internal/http/middleware"
	"github.com/Ulpio/vergo/internal/pkg/config"
	"github.com/Ulpio/vergo/internal/pkg/db"
	"github.com/gin-gonic/gin"
)

// Register registra todas as rotas v1.
func Register(v1 *gin.RouterGroup) {
	cfg := config.Load()

	sqlDB, err := db.Open(cfg)
	if err != nil {
		panic(err)
	}

	//user services
	usr := user.NewMemoryService()
	ah := handlers.NewAuthHandler(cfg, usr)

	auditSvc := audit.NewMemoryService()
	projectSvc := project.NewPostgresService(sqlDB)
	ph := handlers.NewProjectsHandler(projectSvc, auditSvc)

	auth := v1.Group("/auth")
	{
		auth.POST("/signup", ah.Signup)
		auth.POST("/login", ah.Login)
		auth.POST("/refresh", ah.Refresh)
		auth.POST("/forgot-password", notImplemented("auth.forgot_password"))
		auth.POST("/reset-password", notImplemented("auth.reset_password"))
	}

	// ── Rotas protegidas (exigem Bearer + X-Org-ID) ───────────────────
	protected := v1.Group("/")
	protected.Use(
		middleware.Auth(cfg),
		middleware.Tenant(),
	)
	{
		// Organizações & membros (exemplo; refine RBAC depois)
		orgs := protected.Group("/orgs")
		{
			orgs.POST("", notImplemented("orgs.create"))
			orgs.GET("/:id", notImplemented("orgs.get"))
			orgs.PATCH("/:id", notImplemented("orgs.update"))
			orgs.POST("/:id/invite", notImplemented("orgs.invite"))
			orgs.POST("/:id/members", notImplemented("orgs.members.add"))
			orgs.PATCH("/:id/members/:userId", notImplemented("orgs.members.update"))
			orgs.DELETE("/:id/members/:userId", notImplemented("orgs.members.remove"))
		}

		// Contexto de organização
		protected.GET("/context", notImplemented("context.get"))
		protected.POST("/context", notImplemented("context.set"))

		// Projects
		projects := protected.Group("/projects")
		{
			projects.GET("", ph.List)
			projects.POST("", ph.Create)
			projects.GET("/:id", ph.Get)
			projects.PATCH("/:id", ph.Update)
			projects.DELETE("/:id", ph.Delete)
		}

		// Auditoria
		protected.GET("/audit", notImplemented("audit.list"))

		// API Keys
		keys := protected.Group("/api-keys")
		{
			keys.POST("", notImplemented("apikeys.create"))
			keys.GET("", notImplemented("apikeys.list"))
			keys.DELETE("/:id", notImplemented("apikeys.delete"))
		}

		// Webhooks
		wh := protected.Group("/webhooks")
		{
			wh.POST("/endpoints", notImplemented("webhooks.endpoints.create"))
			wh.GET("/endpoints", notImplemented("webhooks.endpoints.list"))
			wh.PATCH("/endpoints/:id", notImplemented("webhooks.endpoints.update"))
			wh.POST("/test", notImplemented("webhooks.test"))
		}

		// Billing
		billing := protected.Group("/billing")
		{
			billing.POST("/checkout-session", notImplemented("billing.checkout_session"))
			billing.GET("/subscription", notImplemented("billing.subscription.get"))
			billing.POST("/webhook", notImplemented("billing.webhook"))
		}

		// Storage
		storage := protected.Group("/storage")
		{
			storage.POST("/presign", notImplemented("storage.presign"))
		}
	}

	_ = http.StatusNotImplemented
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

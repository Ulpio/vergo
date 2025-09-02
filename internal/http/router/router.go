package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Ulpio/vergo/internal/auth"
	"github.com/Ulpio/vergo/internal/domain/audit"
	"github.com/Ulpio/vergo/internal/domain/file"
	"github.com/Ulpio/vergo/internal/domain/org"
	"github.com/Ulpio/vergo/internal/domain/project"
	"github.com/Ulpio/vergo/internal/domain/user"
	"github.com/Ulpio/vergo/internal/domain/userctx"
	"github.com/Ulpio/vergo/internal/http/handlers"
	"github.com/Ulpio/vergo/internal/http/middleware"
	"github.com/Ulpio/vergo/internal/pkg/config"
	"github.com/Ulpio/vergo/internal/pkg/db"
	s3store "github.com/Ulpio/vergo/internal/storage/s3"
)

// Register registra todas as rotas v1.
func Register(v1 *gin.RouterGroup) {
	cfg := config.Load()

	// DB
	sqlDB, err := db.Open(cfg)
	if err != nil {
		panic(err)
	}

	// Services
	userSvc := user.NewPostgresService(sqlDB)
	orgSvc := org.NewPostgresService(sqlDB)
	projSvc := project.NewPostgresService(sqlDB)
	auditSvc := audit.NewPostgresService(sqlDB)
	rfStore := auth.NewRefreshStore(sqlDB)
	ctxSvc := userctx.NewPostgresService(sqlDB)
	fileSvc := file.NewPostgresService(sqlDB)

	// Handler
	authH := handlers.NewAuthHandler(cfg, userSvc, rfStore)
	orgH := handlers.NewOrgsHandler(orgSvc)
	projH := handlers.NewProjectsHandler(projSvc, auditSvc)
	meH := handlers.NewMeHandler(userSvc, orgSvc)
	auditH := handlers.NewAuditHandler(auditSvc)
	ctxH := handlers.NewContextHandler(ctxSvc, orgSvc)

	s3c, err := s3store.NewFromEnv()
	if err != nil {
		panic(err)
	}
	storH := handlers.NewStorageHandler(s3c, fileSvc)

	// ── Público (sem token) ───────────────────────────────────────────
	auth := v1.Group("/auth")
	{
		auth.POST("/signup", authH.Signup)
		auth.POST("/login", authH.Login)
		auth.POST("/refresh", authH.Refresh)
		auth.POST("/logout", authH.Logout) // revoga um refresh específico
		auth.POST("/forgot-password", notImplemented("auth.forgot_password"))
		auth.POST("/reset-password", notImplemented("auth.reset_password"))
	}

	// ── Apenas autenticado (NÃO exige X-Org-ID) ───────────────────────
	authOnly := v1.Group("/")
	authOnly.Use(middleware.Auth(cfg))
	{
		authOnly.GET("/me", meH.Get)
		// logout de todos os devices do usuário logado
		authOnly.POST("/auth/logout-all", authH.LogoutAll)
		authOnly.GET("/context", ctxH.Get)
		authOnly.POST("/context", ctxH.Set)

		orgs := authOnly.Group("/orgs")
		{
			orgs.POST("", orgH.Create) // criar org não exige tenant
			orgs.GET("/:id", orgH.Get)
		}
	}

	// ── Autenticado + Tenant (exige X-Org-ID e membership) ────────────
	protected := v1.Group("/")
	protected.Use(middleware.Auth(cfg), middleware.Tenant(orgSvc, ctxSvc))
	{
		// Orgs (rotas sensíveis com RBAC)
		orgs := protected.Group("/orgs")
		{
			// gestão de membros: admin ou owner
			orgs.POST("/:id/members", middleware.RequireRole("admin"), orgH.AddMember)
			orgs.PATCH("/:id/members/:userId", middleware.RequireRole("admin"), orgH.UpdateMember)
			orgs.DELETE("/:id/members/:userId", middleware.RequireRole("admin"), orgH.RemoveMember)

			// excluir org: somente owner
			orgs.DELETE("/:id", middleware.RequireRole("owner"), orgH.Delete)
		}

		// Projects (qualquer member+ pode criar/editar)
		projects := protected.Group("/projects", middleware.RequireRole("member"))
		{
			projects.GET("", projH.List)
			projects.POST("", projH.Create)
			projects.GET("/:id", projH.Get)
			projects.PATCH("/:id", projH.Update)
			projects.DELETE("/:id", projH.Delete)
		}

		// Auditoria
		protected.GET("/audit", middleware.RequireRole("admin"), auditH.List)

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
		storage := protected.Group("/storage", middleware.RequireRole("member"))
		{
			storage.POST("/presign", storH.PresignPut)          // PUT upload
			storage.POST("/presign-download", storH.PresignGet) // GET download

			storage.GET("/files", storH.ListFiles)
			storage.POST("/files", storH.CreateFile) // registra metadados após upload
			storage.GET("/files/:id", storH.GetFile)
			storage.DELETE("/files/:id", storH.DeleteFile)
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

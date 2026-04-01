package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Ulpio/vergo/internal/auth"
	"github.com/Ulpio/vergo/internal/domain/apikey"
	"github.com/Ulpio/vergo/internal/domain/audit"
	"github.com/Ulpio/vergo/internal/domain/billing"
	"github.com/Ulpio/vergo/internal/domain/webhook"
	"github.com/Ulpio/vergo/internal/domain/file"
	"github.com/Ulpio/vergo/internal/domain/org"
	"github.com/Ulpio/vergo/internal/domain/project"
	"github.com/Ulpio/vergo/internal/domain/user"
	"github.com/Ulpio/vergo/internal/domain/userctx"
	"github.com/Ulpio/vergo/internal/http/handlers"
	"github.com/Ulpio/vergo/internal/http/middleware"
	"github.com/Ulpio/vergo/internal/pkg/config"
	"github.com/Ulpio/vergo/internal/pkg/db"
	"github.com/Ulpio/vergo/internal/repo"
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

	// Repo (sqlc generated queries)
	queries := repo.New(sqlDB)

	// Services
	userSvc := user.NewPostgresService(sqlDB, queries)
	orgSvc := org.NewPostgresService(sqlDB, queries)
	projSvc := project.NewPostgresService(sqlDB, queries)
	auditSvc := audit.NewPostgresService(sqlDB, queries)
	rfStore := auth.NewRefreshStore(sqlDB, queries)
	resetStore := auth.NewResetStore(queries)
	ctxSvc := userctx.NewPostgresService(sqlDB, queries)
	fileSvc := file.NewPostgresService(sqlDB, queries)
	keySvc := apikey.NewService(queries)
	whSvc := webhook.NewService(queries)
	billSvc := billing.NewService(queries, cfg.StripeSecretKey)

	// Handler
	authH := handlers.NewAuthHandler(cfg, userSvc, rfStore, resetStore)
	orgH := handlers.NewOrgsHandler(orgSvc, auditSvc)
	projH := handlers.NewProjectsHandler(projSvc, auditSvc)
	meH := handlers.NewMeHandler(userSvc, orgSvc)
	auditH := handlers.NewAuditHandler(auditSvc)
	ctxH := handlers.NewContextHandler(ctxSvc, orgSvc)
	keyH := handlers.NewAPIKeysHandler(keySvc, auditSvc)
	whH := handlers.NewWebhooksHandler(whSvc)
	billH := handlers.NewBillingHandler(billSvc, cfg.StripeWebhookSecret)

	s3c, err := s3store.NewFromConfig(cfg)
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
		auth.POST("/forgot-password", authH.ForgotPassword)
		auth.POST("/reset-password", authH.ResetPassword)
	}

	// Stripe webhook (público, sem auth — verifica assinatura Stripe)
	v1.POST("/billing/webhook", billH.Webhook)

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
	protected.Use(middleware.AuthWithAPIKeys(cfg, keySvc), middleware.Tenant(orgSvc, ctxSvc))
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
			keys.POST("", keyH.Create)
			keys.GET("", keyH.List)
			keys.DELETE("/:id", keyH.Revoke)
		}

		// Webhooks
		wh := protected.Group("/webhooks")
		{
			wh.POST("/endpoints", whH.CreateEndpoint)
			wh.GET("/endpoints", whH.ListEndpoints)
			wh.PATCH("/endpoints/:id", whH.UpdateEndpoint)
			wh.POST("/test", whH.Test)
		}

		// Billing (webhook is registered as public above)
		billingG := protected.Group("/billing")
		{
			billingG.POST("/checkout-session", billH.CreateCheckoutSession)
			billingG.GET("/subscription", billH.GetSubscription)
			billingG.GET("/usage", billH.GetUsage)
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

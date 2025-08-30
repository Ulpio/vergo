package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Ulpio/vergo/internal/auth"
	"github.com/Ulpio/vergo/internal/domain/user"
	"github.com/Ulpio/vergo/internal/pkg/config"
)

type AuthHandler struct {
	cfg config.Config
	us  user.Service
	rs  auth.RefreshStore
}

func NewAuthHandler(cfg config.Config, us user.Service, rs auth.RefreshStore) *AuthHandler {
	return &AuthHandler{cfg: cfg, us: us, rs: rs}
}

type creds struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

func (h *AuthHandler) Signup(c *gin.Context) {
	var in creds
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid_payload"})
		return
	}
	u, err := h.us.Signup(in.Email, in.Password)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	at, err := auth.NewAccessToken(u.ID, h.cfg.JWTAccessSecret, h.cfg.JWTAccessTTLMinutes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token_error"})
		return
	}
	rt, jti, exp, err := auth.NewRefreshToken(u.ID, h.cfg.JWTRefreshSecret, h.cfg.JWTRefreshTTLDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token_error"})
		return
	}

	if err := h.rs.SaveRefresh(context.Background(), jti, u.ID, rt, exp, nil); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "store_error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user":          gin.H{"id": u.ID, "email": u.Email},
		"access_token":  at,
		"refresh_token": rt,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var in creds
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid_payload"})
		return
	}
	u, err := h.us.Login(in.Email, in.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials"})
		return
	}

	at, err := auth.NewAccessToken(u.ID, h.cfg.JWTAccessSecret, h.cfg.JWTAccessTTLMinutes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token_error"})
		return
	}
	rt, jti, exp, err := auth.NewRefreshToken(u.ID, h.cfg.JWTRefreshSecret, h.cfg.JWTRefreshTTLDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token_error"})
		return
	}

	if err := h.rs.SaveRefresh(context.Background(), jti, u.ID, rt, exp, nil); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "store_error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":          gin.H{"id": u.ID, "email": u.Email},
		"access_token":  at,
		"refresh_token": rt,
	})
}

type refreshReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var in refreshReq
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid_payload"})
		return
	}
	claims, err := auth.Parse(in.RefreshToken, h.cfg.JWTRefreshSecret)
	if err != nil || claims.TokenType != "refresh" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_refresh"})
		return
	}

	ok, userID, _, err := h.rs.IsValid(context.Background(), claims.ID, in.RefreshToken)
	if err != nil || !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_refresh"})
		return
	}

	// rotate: revoga o antigo, emite novo par
	_ = h.rs.Revoke(context.Background(), claims.ID)

	at, err := auth.NewAccessToken(userID, h.cfg.JWTAccessSecret, h.cfg.JWTAccessTTLMinutes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token_error"})
		return
	}
	newRT, newJTI, exp, err := auth.NewRefreshToken(userID, h.cfg.JWTRefreshSecret, h.cfg.JWTRefreshTTLDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token_error"})
		return
	}

	rotFrom := claims.ID
	if err := h.rs.SaveRefresh(context.Background(), newJTI, userID, newRT, exp, &rotFrom); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "store_error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  at,
		"refresh_token": newRT,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var in refreshReq
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid_payload"})
		return
	}
	claims, err := auth.Parse(in.RefreshToken, h.cfg.JWTRefreshSecret)
	if err != nil || claims.TokenType != "refresh" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_refresh"})
		return
	}
	_ = h.rs.Revoke(context.Background(), claims.ID)
	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) LogoutAll(c *gin.Context) {
	uid, ok := middlewareUserID(c) // helper local para evitar import cycle
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing_user"})
		return
	}
	_ = h.rs.RevokeAllForUser(context.Background(), uid)
	c.Status(http.StatusNoContent)
}

// pequena função para ler user_id sem importar o middleware (evita dependência cruzada)
func middlewareUserID(c *gin.Context) (string, bool) {
	v, ok := c.Get("user_id")
	if !ok {
		return "", false
	}
	id, _ := v.(string)
	return id, id != ""
}

package handlers

import (
	"net/http"

	"github.com/Ulpio/vergo/internal/auth"
	"github.com/Ulpio/vergo/internal/domain/user"
	"github.com/Ulpio/vergo/internal/pkg/config"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	cfg config.Config
	us  user.Service
}

func NewAuthHandler(cfg config.Config, us user.Service) *AuthHandler {
	return &AuthHandler{cfg: cfg, us: us}
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
		code := http.StatusBadRequest
		if err == user.ErrEmailInUse {
			code = http.StatusConflict
		}
		c.JSON(code, gin.H{"error": err.Error()})
		return
	}

	at, err := auth.NewAccessToken(u.ID, h.cfg.JWTAccessSecret, h.cfg.JWTAccessTTLMinutes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token_error"})
		return
	}
	rt, err := auth.NewRefreshToken(u.ID, h.cfg.JWTRefreshSecret, h.cfg.JWTRefreshTTLDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token_error"})
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
	rt, err := auth.NewRefreshToken(u.ID, h.cfg.JWTRefreshSecret, h.cfg.JWTRefreshTTLDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token_error"})
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
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_refresh"})
		return
	}
	if _, err := h.us.GetByID(claims.UserID); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_user"})
		return
	}
	at, err := auth.NewAccessToken(claims.UserID, h.cfg.JWTAccessSecret, h.cfg.JWTAccessTTLMinutes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token_error"})
		return
	}
	rt, err := auth.NewRefreshToken(claims.UserID, h.cfg.JWTRefreshSecret, h.cfg.JWTRefreshTTLDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token_error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  at,
		"refresh_token": rt,
	})
}

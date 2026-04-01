package handlers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/Ulpio/vergo/internal/auth"
	"github.com/Ulpio/vergo/internal/domain/user"
	"github.com/Ulpio/vergo/internal/pkg/config"
)

type AuthHandler struct {
	cfg    config.Config
	us     user.Service
	rs     auth.RefreshStore
	resets auth.ResetStore
}

func NewAuthHandler(cfg config.Config, us user.Service, rs auth.RefreshStore, resets auth.ResetStore) *AuthHandler {
	return &AuthHandler{cfg: cfg, us: us, rs: rs, resets: resets}
}

type creds struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// Signup registers a new user.
// @Summary Register a new user
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body creds true "User credentials"
// @Success 201 {object} AuthResponse
// @Failure 409 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/signup [post]
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

// Login authenticates a user and returns tokens.
// @Summary Authenticate user
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body creds true "User credentials"
// @Success 200 {object} AuthResponse
// @Failure 401 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/login [post]
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

// Refresh rotates the token pair.
// @Summary Refresh access token
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body refreshReq true "Refresh token"
// @Success 200 {object} TokenResponse
// @Failure 401 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/refresh [post]
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

// Logout revokes a specific refresh token.
// @Summary Revoke refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body refreshReq true "Refresh token to revoke"
// @Success 204 "No Content"
// @Failure 401 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Router /auth/logout [post]
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

// LogoutAll revokes all refresh tokens for the authenticated user.
// @Summary Revoke all refresh tokens
// @Tags Auth
// @Security BearerAuth
// @Success 204 "No Content"
// @Failure 401 {object} ErrorResponse
// @Router /auth/logout-all [post]
func (h *AuthHandler) LogoutAll(c *gin.Context) {
	uid, ok := middlewareUserID(c) // helper local para evitar import cycle
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing_user"})
		return
	}
	_ = h.rs.RevokeAllForUser(context.Background(), uid)
	c.Status(http.StatusNoContent)
}

type forgotIn struct {
	Email string `json:"email" binding:"required,email"`
}

// ForgotPassword generates a password reset token.
// @Summary Request password reset
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body forgotIn true "User email"
// @Success 200 {object} map[string]string
// @Failure 422 {object} ErrorResponse
// @Router /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var in forgotIn
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid_payload"})
		return
	}

	// Always return success to prevent email enumeration
	u, err := h.us.GetByEmail(in.Email)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "if the email exists, a reset link was sent"})
		return
	}

	token, err := h.resets.CreateResetToken(u.ID)
	if err != nil {
		slog.Error("forgot-password: create token", "error", err)
		c.JSON(http.StatusOK, gin.H{"message": "if the email exists, a reset link was sent"})
		return
	}

	// In production, send via email. For now, log to console.
	slog.Info("password reset token generated",
		"email", in.Email,
		"token", token,
		"reset_url", "/auth/reset-password?token="+token,
	)

	c.JSON(http.StatusOK, gin.H{"message": "if the email exists, a reset link was sent"})
}

type resetIn struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// ResetPassword resets the user's password using a valid token.
// @Summary Reset password with token
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body resetIn true "Reset token and new password"
// @Success 200 {object} map[string]string
// @Failure 401 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var in resetIn
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid_payload"})
		return
	}

	userID, err := h.resets.ValidateAndConsume(in.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_or_expired_token"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "hash_error"})
		return
	}

	if err := h.us.UpdatePassword(userID, string(hash)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update_failed"})
		return
	}

	// Revoke all refresh tokens for security
	_ = h.rs.RevokeAllForUser(context.Background(), userID)

	c.JSON(http.StatusOK, gin.H{"message": "password_reset_successful"})
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

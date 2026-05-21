package handler

import (
	"net/http"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/config"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/service"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	cfg         *config.Config
	authService *service.AuthService
}

func NewAuthHandler(cfg *config.Config, as *service.AuthService) *AuthHandler {
	return &AuthHandler{cfg, as}
}

func (h *AuthHandler) GithubLogin(c *gin.Context) {
	url := "https://github.com/login/oauth/authorize?client_id=" + h.cfg.GithubClientID + "&scope=user:email"
	c.Redirect(http.StatusFound, url)
}

func (h *AuthHandler) GithubCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "MISSING_CODE", "message": "authorization code is required"}})
		return
	}
	user, err := h.authService.ExchangeGithubCode(code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "OAUTH_FAILED", "message": err.Error()}})
		return
	}
	token, err := h.authService.GenerateJWT(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "TOKEN_ERROR", "message": "failed to generate token"}})
		return
	}
	c.Redirect(http.StatusFound, "/callback?token="+token)
}

func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	url := "https://accounts.google.com/o/oauth2/v2/auth?client_id=" + h.cfg.GoogleClientID +
		"&redirect_uri=http://localhost:8080/api/v1/auth/google/callback" +
		"&response_type=code&scope=email+profile"
	c.Redirect(http.StatusFound, url)
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "MISSING_CODE", "message": "authorization code is required"}})
		return
	}
	user, err := h.authService.ExchangeGoogleCode(code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "OAUTH_FAILED", "message": err.Error()}})
		return
	}
	token, err := h.authService.GenerateJWT(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "TOKEN_ERROR", "message": "failed to generate token"}})
		return
	}
	c.Redirect(http.StatusFound, "/callback?token="+token)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	userID := c.GetString("user_id")
	token, err := h.authService.GenerateJWT(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "TOKEN_ERROR", "message": "failed to refresh token"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.GetString("user_id")
	c.JSON(http.StatusOK, gin.H{"user_id": userID})
}

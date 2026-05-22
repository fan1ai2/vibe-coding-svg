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

// GithubLogin godoc
// @Summary      GitHub OAuth login
// @Description  Redirect to GitHub OAuth authorization page
// @Tags         auth
// @Success      302
// @Router       /auth/github/login [get]
func (h *AuthHandler) GithubLogin(c *gin.Context) {
	url := "https://github.com/login/oauth/authorize?client_id=" + h.cfg.GithubClientID + "&scope=user:email"
	c.Redirect(http.StatusFound, url)
}

// GithubCallback godoc
// @Summary      GitHub OAuth callback
// @Description  Exchange OAuth code for JWT token, redirects to frontend with token
// @Tags         auth
// @Param        code  query     string  true  "OAuth authorization code"
// @Success      302
// @Failure      400  {object}  object{error=object{code=string,message=string}}
// @Failure      401  {object}  object{error=object{code=string,message=string}}
// @Router       /auth/github/callback [get]
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
	c.Redirect(http.StatusFound, h.cfg.FrontendURL+"/callback?token="+token)
}

// Refresh godoc
// @Summary      Refresh JWT token
// @Tags         auth
// @Security     BearerAuth
// @Success      200  {object}  object{token=string}
// @Router       /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	userID := c.GetString("user_id")
	token, err := h.authService.GenerateJWT(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "TOKEN_ERROR", "message": "failed to refresh token"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

// Me godoc
// @Summary      Get current user
// @Tags         auth
// @Security     BearerAuth
// @Success      200  {object}  object{user_id=string}
// @Router       /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.GetString("user_id")
	c.JSON(http.StatusOK, gin.H{"user_id": userID})
}

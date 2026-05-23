package handler

import (
	"net/http"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/config"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/service"
	"github.com/gin-gonic/gin"
)

// AuthHandler 认证相关接口处理器
type AuthHandler struct {
	cfg         *config.Config
	authService *service.AuthService
}

// NewAuthHandler 创建认证处理器实例
func NewAuthHandler(cfg *config.Config, as *service.AuthService) *AuthHandler {
	return &AuthHandler{cfg, as}
}

// GithubLogin godoc
// @Summary      GitHub OAuth 登录
// @Description  重定向到 GitHub OAuth 授权页面
// @Tags         auth
// @Success      302
// @Router       /auth/github/login [get]
func (h *AuthHandler) GithubLogin(c *gin.Context) {
	url := "https://github.com/login/oauth/authorize?client_id=" + h.cfg.GithubClientID + "&scope=user:email"
	c.Redirect(http.StatusFound, url)
}

// GithubCallback godoc
// @Summary      GitHub OAuth 回调
// @Description  用 OAuth code 换取 JWT token，并重定向到前端页面
// @Tags         auth
// @Param        code  query     string  true  "OAuth 授权码"
// @Success      302
// @Failure      400  {object}  object{error=object{code=string,message=string}}
// @Failure      401  {object}  object{error=object{code=string,message=string}}
// @Router       /auth/github/callback [get]
func (h *AuthHandler) GithubCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "MISSING_CODE", "message": "缺少授权码"}})
		return
	}

	// 用 GitHub 授权码换取用户信息
	user, err := h.authService.ExchangeGithubCode(code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "OAUTH_FAILED", "message": err.Error()}})
		return
	}

	// 生成 JWT token 并重定向到前端
	token, err := h.authService.GenerateJWT(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "TOKEN_ERROR", "message": "生成 token 失败"}})
		return
	}
	c.Redirect(http.StatusFound, h.cfg.FrontendURL+"/callback?token="+token)
}

// Refresh godoc
// @Summary      刷新 JWT token
// @Tags         auth
// @Security     BearerAuth
// @Success      200  {object}  object{token=string}
// @Router       /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	userID := c.GetString("user_id")
	token, err := h.authService.GenerateJWT(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "TOKEN_ERROR", "message": "刷新 token 失败"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

// Me godoc
// @Summary      获取当前用户信息
// @Tags         auth
// @Security     BearerAuth
// @Success      200  {object}  object{user_id=string}
// @Router       /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.GetString("user_id")
	c.JSON(http.StatusOK, gin.H{"user_id": userID})
}

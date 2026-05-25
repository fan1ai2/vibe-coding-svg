package handler

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"
	"strings"
	"time"

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
	state, err := generateState()
	if err != nil {
		log.Printf("[ERROR] generate oauth state: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "登录失败，请重试"}})
		return
	}

	secure := strings.HasPrefix(h.cfg.FrontendURL, "https://")
	c.SetCookie("oauth_state", state, int(10*time.Minute.Seconds()), "/api/v1/auth/github", "", secure, true)

	url := "https://github.com/login/oauth/authorize?client_id=" + h.cfg.GithubClientID + "&scope=user:email&state=" + state
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
	state := c.Query("state")
	if code == "" || state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "MISSING_PARAMS", "message": "缺少授权参数"}})
		return
	}

	cookieState, err := c.Cookie("oauth_state")
	if err != nil || cookieState != state {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_STATE", "message": "授权验证失败，请重新登录"}})
		return
	}

	// 清除 state cookie
	c.SetCookie("oauth_state", "", -1, "/api/v1/auth/github", "", false, true)

	// 用 GitHub 授权码换取用户信息
	user, err := h.authService.ExchangeGithubCode(code)
	if err != nil {
		log.Printf("[ERROR] github oauth exchange: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "OAUTH_FAILED", "message": "GitHub 授权失败，请重试"}})
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
// @Success      200  {object}  object{id=string,name=string,email=string,avatar_url=string,provider=string,created_at=string}
// @Router       /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.GetString("user_id")
	user, err := h.authService.FindByID(userID)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "USER_NOT_FOUND", "message": "用户不存在"}})
		return
	}
	c.JSON(http.StatusOK, user)
}

// GuestLogin handles guest user creation/login.
func (h *AuthHandler) GuestLogin(c *gin.Context) {
	guestID, _ := c.Cookie("guest_id")

	user, token, newGuestID, err := h.authService.GuestLogin(guestID)
	if err != nil {
		log.Printf("[ERROR] guest login: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "登录失败，请重试"}})
		return
	}

	secure := strings.HasPrefix(h.cfg.FrontendURL, "https://")
	c.SetCookie("guest_id", newGuestID, int(365*24*time.Hour.Seconds()), "/", "", secure, true)

	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

// EmailSendCode sends verification code to the given email.
func (h *AuthHandler) EmailSendCode(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_PARAMS", "message": "请输入邮箱地址"}})
		return
	}

	if err := h.authService.EmailSendCode(req.Email); err != nil {
		log.Printf("[ERROR] email send code: %v", err)
		c.JSON(http.StatusTooManyRequests, gin.H{"error": gin.H{"code": "RATE_LIMITED", "message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// EmailVerify validates the code and returns a JWT.
func (h *AuthHandler) EmailVerify(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" || req.Code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_PARAMS", "message": "请输入邮箱和验证码"}})
		return
	}

	user, token, err := h.authService.EmailVerify(req.Email, req.Code)
	if err != nil {
		log.Printf("[ERROR] email verify: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_CODE", "message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

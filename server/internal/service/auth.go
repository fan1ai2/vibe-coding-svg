package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/config"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/model"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/repo"
	"github.com/golang-jwt/jwt/v5"
)

// AuthService 认证服务，处理 JWT 签发和 GitHub OAuth 流程
type AuthService struct {
	cfg      *config.Config
	userRepo *repo.UserRepo
}

// NewAuthService 创建认证服务实例
func NewAuthService(cfg *config.Config, ur *repo.UserRepo) *AuthService {
	return &AuthService{cfg, ur}
}

// GenerateJWT 为用户生成 JWT token，有效期 7 天
func (s *AuthService) GenerateJWT(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(7 * 24 * time.Hour).Unix(),
	})
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

// GithubUser GitHub API 返回的用户信息
type GithubUser struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

// ExchangeGithubCode 用 GitHub OAuth 授权码换取用户信息
// 如果是新用户则自动创建账号，已存在则直接返回
func (s *AuthService) ExchangeGithubCode(code string) (*model.User, error) {
	// 第一步：用授权码换取 access_token
	accessToken, err := s.getGithubAccessToken(code)
	if err != nil {
		return nil, err
	}

	// 第二步：用 access_token 获取 GitHub 用户信息
	ghUser, err := s.getGithubUser(accessToken)
	if err != nil {
		return nil, err
	}

	// 第三步：查找或创建本地用户
	providerID := fmt.Sprintf("%d", ghUser.ID)
	user, err := s.userRepo.FindByProvider("github", providerID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		// 新用户，创建账号
		user = &model.User{
			Email:      ghUser.Email,
			Name:       firstNonEmpty(ghUser.Name, ghUser.Login),
			AvatarURL:  ghUser.AvatarURL,
			Provider:   "github",
			ProviderID: providerID,
		}
		if err := s.userRepo.Create(user); err != nil {
			return nil, err
		}
	}
	return user, nil
}

// getGithubAccessToken 用授权码向 GitHub 换取 access_token
func (s *AuthService) getGithubAccessToken(code string) (string, error) {
	url := fmt.Sprintf("https://github.com/login/oauth/access_token?client_id=%s&client_secret=%s&code=%s",
		s.cfg.GithubClientID, s.cfg.GithubSecret, code)
	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error_description"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	if result.Error != "" {
		return "", errors.New(result.Error)
	}
	return result.AccessToken, nil
}

// getGithubUser 用 access_token 获取 GitHub 用户信息
func (s *AuthService) getGithubUser(token string) (*GithubUser, error) {
	req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user GithubUser
	return &user, json.NewDecoder(resp.Body).Decode(&user)
}

// firstNonEmpty 返回第一个非空字符串
func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

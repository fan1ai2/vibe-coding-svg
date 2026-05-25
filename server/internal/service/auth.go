package service

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/config"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/model"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/repo"
	"github.com/golang-jwt/jwt/v5"
)

var httpClient = &http.Client{Timeout: 15 * time.Second}

// AuthService 认证服务，处理 JWT 签发、GitHub OAuth 和邮箱验证流程
type AuthService struct {
	cfg      *config.Config
	userRepo *repo.UserRepo
	emailSvc *EmailService
}

// NewAuthService 创建认证服务实例
func NewAuthService(cfg *config.Config, ur *repo.UserRepo, es *EmailService) *AuthService {
	return &AuthService{cfg, ur, es}
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

	// 第三步：创建或更新本地用户
	providerID := fmt.Sprintf("%d", ghUser.ID)
	user := &model.User{
		Email:      ghUser.Email,
		Name:       firstNonEmpty(ghUser.Name, ghUser.Login),
		AvatarURL:  ghUser.AvatarURL,
		Provider:   "github",
		ProviderID: providerID,
	}
	if err := s.userRepo.UpsertByProvider(user); err != nil {
		return nil, err
	}
	return user, nil
}

// getGithubAccessToken 用授权码向 GitHub 换取 access_token
func (s *AuthService) getGithubAccessToken(code string) (string, error) {
	url := fmt.Sprintf("https://github.com/login/oauth/access_token?client_id=%s&client_secret=%s&code=%s",
		s.cfg.GithubClientID, s.cfg.GithubSecret, code)
	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("Accept", "application/json")
	resp, err := httpClient.Do(req)
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
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user GithubUser
	return &user, json.NewDecoder(resp.Body).Decode(&user)
}

// FindByID 根据用户 ID 查找用户
func (s *AuthService) FindByID(userID string) (*model.User, error) {
	return s.userRepo.FindByID(userID)
}

// firstNonEmpty 返回第一个非空字符串
func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// GuestLogin creates or restores a guest user and returns a JWT.
func (s *AuthService) GuestLogin(guestID string) (*model.User, string, string, error) {
	var user *model.User
	var newGuestID string

	if guestID != "" {
		u, err := s.userRepo.FindByGuestID(guestID)
		if err == nil && u != nil {
			user = u
		}
	}

	if user == nil {
		u, err := s.userRepo.CreateGuest()
		if err != nil {
			return nil, "", "", fmt.Errorf("create guest: %w", err)
		}
		user = u
		newGuestID = u.ProviderID
	} else {
		newGuestID = guestID
	}

	token, err := s.GenerateJWT(user.ID)
	if err != nil {
		return nil, "", "", fmt.Errorf("generate jwt: %w", err)
	}
	return user, token, newGuestID, nil
}

// EmailSendCode generates a 6-digit code and sends it via SMTP.
func (s *AuthService) EmailSendCode(email string) error {
	lastSent, err := s.userRepo.LastCodeSentAt(email)
	if err != nil {
		return fmt.Errorf("check rate limit: %w", err)
	}
	if time.Since(lastSent) < 60*time.Second {
		return fmt.Errorf("请 60 秒后再试")
	}

	code, err := generateCode()
	if err != nil {
		return fmt.Errorf("generate code: %w", err)
	}

	if err := s.userRepo.SaveVerificationCode(email, code); err != nil {
		return fmt.Errorf("save code: %w", err)
	}

	if err := s.emailSvc.SendVerificationCode(email, code); err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	return nil
}

// EmailVerify checks the verification code and logs in / registers the user.
func (s *AuthService) EmailVerify(email, code string) (*model.User, string, error) {
	valid, err := s.userRepo.VerifyCode(email, code)
	if err != nil {
		return nil, "", fmt.Errorf("verify code: %w", err)
	}
	if !valid {
		return nil, "", fmt.Errorf("验证码错误或已过期")
	}

	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, "", fmt.Errorf("find user: %w", err)
	}
	if user == nil {
		user = &model.User{
			Email:      email,
			Provider:   "email",
			ProviderID: email,
		}
		if idx := strings.Index(email, "@"); idx > 0 {
			user.Name = email[:idx]
		}
		if err := s.userRepo.Create(user); err != nil {
			return nil, "", fmt.Errorf("create user: %w", err)
		}
	}

	token, err := s.GenerateJWT(user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("generate jwt: %w", err)
	}
	return user, token, nil
}

func generateCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

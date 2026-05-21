package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/config"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/model"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/repo"
	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	cfg      *config.Config
	userRepo *repo.UserRepo
}

func NewAuthService(cfg *config.Config, ur *repo.UserRepo) *AuthService {
	return &AuthService{cfg, ur}
}

func (s *AuthService) GenerateJWT(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(7 * 24 * time.Hour).Unix(),
	})
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

type GithubUser struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

type GoogleUser struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func (s *AuthService) ExchangeGithubCode(code string) (*model.User, error) {
	accessToken, err := s.getGithubAccessToken(code)
	if err != nil {
		return nil, err
	}
	ghUser, err := s.getGithubUser(accessToken)
	if err != nil {
		return nil, err
	}
	providerID := fmt.Sprintf("%d", ghUser.ID)
	user, err := s.userRepo.FindByProvider("github", providerID)
	if err != nil {
		return nil, err
	}
	if user == nil {
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

func (s *AuthService) ExchangeGoogleCode(code string) (*model.User, error) {
	accessToken, err := s.getGoogleAccessToken(code)
	if err != nil {
		return nil, err
	}
	gUser, err := s.getGoogleUser(accessToken)
	if err != nil {
		return nil, err
	}
	user, err := s.userRepo.FindByProvider("google", gUser.ID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		user = &model.User{
			Email:      gUser.Email,
			Name:       gUser.Name,
			AvatarURL:  gUser.Picture,
			Provider:   "google",
			ProviderID: gUser.ID,
		}
		if err := s.userRepo.Create(user); err != nil {
			return nil, err
		}
	}
	return user, nil
}

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

func (s *AuthService) getGoogleAccessToken(code string) (string, error) {
	url := "https://oauth2.googleapis.com/token"
	body := fmt.Sprintf("client_id=%s&client_secret=%s&code=%s&grant_type=authorization_code&redirect_uri=http://localhost:8080/api/v1/auth/google/callback",
		s.cfg.GoogleClientID, s.cfg.GoogleSecret, code)
	resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(body))
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

func (s *AuthService) getGoogleUser(token string) (*GoogleUser, error) {
	req, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var user GoogleUser
	return &user, json.NewDecoder(resp.Body).Decode(&user)
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

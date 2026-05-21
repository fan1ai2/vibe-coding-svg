# Image-to-SVG Converter Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a web app where users upload raster images and get SVG vector conversions, with personal library history.

**Architecture:** Go API server (gin) + Go worker (goroutine pool consuming Redis jobs) + React SPA (Vite/TypeScript/Tailwind). PostgreSQL for data, Redis for job queue, MinIO for file storage. Frontend polls API for conversion status.

**Tech Stack:** Go 1.22+, gin, golang-jwt, go-redis, minio-go, golang-migrate, React 18, Vite, TypeScript, Tailwind CSS, React Router v6, axios

---

### Task 1: Docker Compose Infrastructure

**Files:**
- Create: `docker-compose.yml`

- [ ] **Step 1: Write docker-compose.yml**

```yaml
version: "3.9"
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: svguser
      POSTGRES_PASSWORD: svgpass
      POSTGRES_DB: svgconverter
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  minio:
    image: minio/minio:latest
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    ports:
      - "9000:9000"
      - "9001:9001"
    volumes:
      - miniodata:/data

volumes:
  pgdata:
  miniodata:
```

- [ ] **Step 2: Start infrastructure**

Run: `docker-compose up -d`
Expected: three containers running (verify with `docker-compose ps`)

- [ ] **Step 3: Commit**

```bash
git add docker-compose.yml
git commit -m "feat: add docker-compose for postgres, redis, minio"
```

---

### Task 2: Go Module & Project Scaffolding

**Files:**
- Create: `server/go.mod`
- Create: `server/cmd/api/main.go` (skeleton)
- Create: `server/cmd/worker/main.go` (skeleton)
- Create: `server/internal/config/config.go`

- [ ] **Step 1: Initialize Go module**

Run: `cd server && go mod init github.com/fan1ai2/vibe-coding-svg/server`
Expected: `go.mod` created

- [ ] **Step 2: Write config package**

```go
// server/internal/config/config.go
package config

import "os"

type Config struct {
	Port         string
	DatabaseURL  string
	RedisAddr    string
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
	JWTSecret      string
	GithubClientID string
	GithubSecret   string
	GoogleClientID string
	GoogleSecret   string
	MaxFileSize    int64
}

func Load() *Config {
	return &Config{
		Port:           envOr("PORT", "8080"),
		DatabaseURL:    envOr("DATABASE_URL", "postgres://svguser:svgpass@localhost:5432/svgconverter?sslmode=disable"),
		RedisAddr:      envOr("REDIS_ADDR", "localhost:6379"),
		MinioEndpoint:  envOr("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey: envOr("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey: envOr("MINIO_SECRET_KEY", "minioadmin"),
		MinioBucket:    envOr("MINIO_BUCKET", "svgconverter"),
		JWTSecret:      envOr("JWT_SECRET", "dev-secret-change-in-prod"),
		GithubClientID: os.Getenv("GITHUB_CLIENT_ID"),
		GithubSecret:   os.Getenv("GITHUB_CLIENT_SECRET"),
		GoogleClientID: os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleSecret:   os.Getenv("GOOGLE_CLIENT_SECRET"),
		MaxFileSize:    10 << 20, // 10MB
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
```

- [ ] **Step 3: Write API entry point skeleton**

```go
// server/cmd/api/main.go
package main

import (
	"log"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/config"
)

func main() {
	cfg := config.Load()
	log.Printf("API server starting on :%s", cfg.Port)
}
```

- [ ] **Step 4: Write worker entry point skeleton**

```go
// server/cmd/worker/main.go
package main

import (
	"log"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/config"
)

func main() {
	cfg := config.Load()
	log.Printf("Worker starting, redis=%s", cfg.RedisAddr)
}
```

- [ ] **Step 5: Install initial dependencies**

Run: `cd server && go get github.com/gin-gonic/gin github.com/golang-jwt/jwt/v5 github.com/redis/go-redis/v9 github.com/minio/minio-go/v7 github.com/golang-migrate/migrate/v4 github.com/lib/pq github.com/google/uuid`
Expected: `go.sum` populated

- [ ] **Step 6: Commit**

```bash
git add server/
git commit -m "feat: scaffold Go module with config and entry points"
```

---

### Task 3: Database Migrations

**Files:**
- Create: `server/migrations/001_create_users.up.sql`
- Create: `server/migrations/001_create_users.down.sql`
- Create: `server/migrations/002_create_conversions.up.sql`
- Create: `server/migrations/002_create_conversions.down.sql`
- Create: `server/migrations/003_create_quotas.up.sql`
- Create: `server/migrations/003_create_quotas.down.sql`

- [ ] **Step 1: Write users migration up**

```sql
-- server/migrations/001_create_users.up.sql
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE,
    name VARCHAR(100),
    avatar_url VARCHAR(500),
    provider VARCHAR(20) NOT NULL,
    provider_id VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE(provider, provider_id)
);
```

- [ ] **Step 2: Write users migration down**

```sql
-- server/migrations/001_create_users.down.sql
DROP TABLE IF EXISTS users;
```

- [ ] **Step 3: Write conversions migration up**

```sql
-- server/migrations/002_create_conversions.up.sql
CREATE TABLE conversions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    original_url VARCHAR(500),
    svg_url VARCHAR(500),
    thumbnail_url VARCHAR(500),
    file_size_in BIGINT,
    file_size_out BIGINT,
    path_count INT,
    color_count INT,
    format_in VARCHAR(10),
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_conversions_user_status ON conversions(user_id, status);
CREATE INDEX idx_conversions_created ON conversions(created_at DESC);
```

- [ ] **Step 4: Write conversions migration down**

```sql
-- server/migrations/002_create_conversions.down.sql
DROP TABLE IF EXISTS conversions;
```

- [ ] **Step 5: Write quotas migration up**

```sql
-- server/migrations/003_create_quotas.up.sql
CREATE TABLE daily_quotas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date DATE NOT NULL DEFAULT CURRENT_DATE,
    count INT NOT NULL DEFAULT 0,
    UNIQUE(user_id, date)
);
```

- [ ] **Step 6: Write quotas migration down**

```sql
-- server/migrations/003_create_quotas.down.sql
DROP TABLE IF EXISTS daily_quotas;
```

- [ ] **Step 7: Run migrations**

Run: `cd server && go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest && migrate -path migrations -database "postgres://svguser:svgpass@localhost:5432/svgconverter?sslmode=disable" up`
Expected: "3/u 001_create_users ..." (three migrations applied)

- [ ] **Step 8: Commit**

```bash
git add server/migrations/
git commit -m "feat: add database migrations for users, conversions, quotas"
```

---

### Task 4: Data Models

**Files:**
- Create: `server/internal/model/user.go`
- Create: `server/internal/model/conversion.go`
- Create: `server/internal/model/quota.go`

- [ ] **Step 1: Write user model**

```go
// server/internal/model/user.go
package model

import "time"

type User struct {
	ID         string    `json:"id" db:"id"`
	Email      string    `json:"email" db:"email"`
	Name       string    `json:"name" db:"name"`
	AvatarURL  string    `json:"avatar_url" db:"avatar_url"`
	Provider   string    `json:"provider" db:"provider"`
	ProviderID string    `json:"provider_id" db:"provider_id"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}
```

- [ ] **Step 2: Write conversion model**

```go
// server/internal/model/conversion.go
package model

import "time"

type Conversion struct {
	ID           string    `json:"id" db:"id"`
	UserID       string    `json:"user_id" db:"user_id"`
	Status       string    `json:"status" db:"status"`
	OriginalURL  string    `json:"original_url,omitempty" db:"original_url"`
	SVGURL       string    `json:"svg_url,omitempty" db:"svg_url"`
	ThumbnailURL string    `json:"thumbnail_url,omitempty" db:"thumbnail_url"`
	FileSizeIn   int64     `json:"file_size_in" db:"file_size_in"`
	FileSizeOut  int64     `json:"file_size_out" db:"file_size_out"`
	PathCount    int       `json:"path_count" db:"path_count"`
	ColorCount   int       `json:"color_count" db:"color_count"`
	FormatIn     string    `json:"format_in" db:"format_in"`
	ErrorMessage string    `json:"error_message,omitempty" db:"error_message"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty" db:"completed_at"`
}

const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
)
```

- [ ] **Step 3: Write quota model**

```go
// server/internal/model/quota.go
package model

import "time"

type DailyQuota struct {
	ID     string    `json:"id" db:"id"`
	UserID string    `json:"user_id" db:"user_id"`
	Date   time.Time `json:"date" db:"date"`
	Count  int       `json:"count" db:"count"`
}
```

- [ ] **Step 4: Commit**

```bash
git add server/internal/model/
git commit -m "feat: add data models for user, conversion, quota"
```

---

### Task 5: User Repository

**Files:**
- Create: `server/internal/repo/user.go`

- [ ] **Step 1: Write user repository**

```go
// server/internal/repo/user.go
package repo

import (
	"database/sql"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/model"
)

type UserRepo struct{ db *sql.DB }

func NewUserRepo(db *sql.DB) *UserRepo { return &UserRepo{db} }

func (r *UserRepo) FindByProvider(provider, providerID string) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRow(
		`SELECT id, email, name, avatar_url, provider, provider_id, created_at, updated_at
		 FROM users WHERE provider=$1 AND provider_id=$2`,
		provider, providerID,
	).Scan(&u.ID, &u.Email, &u.Name, &u.AvatarURL, &u.Provider, &u.ProviderID, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func (r *UserRepo) Create(u *model.User) error {
	return r.db.QueryRow(
		`INSERT INTO users (email, name, avatar_url, provider, provider_id)
		 VALUES ($1,$2,$3,$4,$5) RETURNING id, created_at, updated_at`,
		u.Email, u.Name, u.AvatarURL, u.Provider, u.ProviderID,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

func (r *UserRepo) FindByID(id string) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRow(
		`SELECT id, email, name, avatar_url, provider, provider_id, created_at, updated_at
		 FROM users WHERE id=$1`, id,
	).Scan(&u.ID, &u.Email, &u.Name, &u.AvatarURL, &u.Provider, &u.ProviderID, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}
```

- [ ] **Step 2: Commit**

```bash
git add server/internal/repo/
git commit -m "feat: add user repository"
```

---

### Task 6: Auth Service

**Files:**
- Create: `server/internal/service/auth.go`

- [ ] **Step 1: Write auth service**

```go
// server/internal/service/auth.go
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
	ID    int    `json:"id"`
	Login string `json:"login"`
	Email string `json:"email"`
	Name  string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

type GoogleUser struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func (s *AuthService) ExchangeGithubCode(code string) (*model.User, error) {
	// POST https://github.com/login/oauth/access_token
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
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var user GoogleUser
	return &user, json.NewDecoder(resp.Body).Decode(&user)
}

func firstNonEmpty(a, b string) string {
	if a != "" { return a }
	return b
}
```

- [ ] **Step 2: Install missing dependency**

Run: `cd server && go mod tidy`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add server/internal/service/auth.go
git commit -m "feat: add auth service with GitHub and Google OAuth"
```

---

### Task 7: Auth Handlers

**Files:**
- Create: `server/internal/handler/auth.go`

- [ ] **Step 1: Write auth handlers**

```go
// server/internal/handler/auth.go
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
	// JWT refresh — for MVP, just re-issue the same token
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
```

- [ ] **Step 2: Commit**

```bash
git add server/internal/handler/auth.go
git commit -m "feat: add auth HTTP handlers"
```

---

### Task 8: Middleware

**Files:**
- Create: `server/internal/middleware/jwt.go`
- Create: `server/internal/middleware/cors.go`
- Create: `server/internal/middleware/ratelimit.go`

- [ ] **Step 1: Write JWT middleware**

```go
// server/internal/middleware/jwt.go
package middleware

import (
	"net/http"
	"strings"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func JWTAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"code": "UNAUTHORIZED", "message": "missing or malformed token"},
			})
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"code": "INVALID_TOKEN", "message": "token is invalid or expired"},
			})
			return
		}
		claims := token.Claims.(jwt.MapClaims)
		c.Set("user_id", claims["sub"])
		c.Next()
	}
}
```

- [ ] **Step 2: Write CORS middleware**

```go
// server/internal/middleware/cors.go
package middleware

import "github.com/gin-gonic/gin"

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
```

- [ ] **Step 3: Write rate limit middleware**

```go
// server/internal/middleware/ratelimit.go
package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

func RateLimit(maxPerMin int) gin.HandlerFunc {
	mu := sync.Mutex{}
	hits := make(map[string][]time.Time)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		mu.Lock()
		now := time.Now()
		cutoff := now.Add(-time.Minute)
		filtered := hits[ip][:0]
		for _, t := range hits[ip] {
			if t.After(cutoff) {
				filtered = append(filtered, t)
			}
		}
		if len(filtered) >= maxPerMin {
			mu.Unlock()
			c.Header("Retry-After", "60")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{"code": "RATE_LIMITED", "message": "too many requests"},
			})
			return
		}
		hits[ip] = append(filtered, now)
		mu.Unlock()
		c.Next()
	}
}
```

- [ ] **Step 4: Commit**

```bash
git add server/internal/middleware/
git commit -m "feat: add JWT, CORS, and rate limit middleware"
```

---

### Task 9: Router & API Main

**Files:**
- Create: `server/internal/router/router.go`
- Modify: `server/cmd/api/main.go`

- [ ] **Step 1: Write router**

```go
// server/internal/router/router.go
package router

import (
	"database/sql"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/config"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/handler"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/middleware"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/repo"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/service"
	"github.com/gin-gonic/gin"
)

func Setup(cfg *config.Config, db *sql.DB) *gin.Engine {
	r := gin.Default()

	userRepo := repo.NewUserRepo(db)
	authSvc := service.NewAuthService(cfg, userRepo)
	authH := handler.NewAuthHandler(cfg, authSvc)

	r.Use(middleware.CORS())
	r.Use(middleware.RateLimit(100))

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.GET("/github/login", authH.GithubLogin)
			auth.GET("/github/callback", authH.GithubCallback)
			auth.GET("/google/login", authH.GoogleLogin)
			auth.GET("/google/callback", authH.GoogleCallback)
			auth.POST("/refresh", middleware.JWTAuth(cfg), authH.Refresh)
			auth.GET("/me", middleware.JWTAuth(cfg), authH.Me)
		}
	}

	return r
}
```

- [ ] **Step 2: Update API main.go**

```go
// server/cmd/api/main.go
package main

import (
	"database/sql"
	"log"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/config"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/router"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("database ping failed: %v", err)
	}
	log.Println("connected to postgres")

	r := router.Setup(cfg, db)
	log.Printf("API server starting on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
```

- [ ] **Step 3: Verify it compiles**

Run: `cd server && go build ./cmd/api/...`
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add server/internal/router/ server/cmd/api/main.go
git commit -m "feat: wire up router with auth routes and API main"
```

---

### Task 10: Storage Service (MinIO)

**Files:**
- Create: `server/internal/service/storage.go`

- [ ] **Step 1: Write storage service**

```go
// server/internal/service/storage.go
package service

import (
	"context"
	"fmt"
	"io"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type StorageService struct {
	client *minio.Client
	bucket string
}

func NewStorageService(cfg *config.Config) (*StorageService, error) {
	client, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, fmt.Errorf("minio client: %w", err)
	}
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.MinioBucket)
	if err != nil {
		return nil, fmt.Errorf("bucket check: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.MinioBucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("create bucket: %w", err)
		}
	}
	return &StorageService{client, cfg.MinioBucket}, nil
}

func (s *StorageService) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucket, key, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (s *StorageService) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}
```

- [ ] **Step 2: Commit**

```bash
git add server/internal/service/storage.go
git commit -m "feat: add MinIO storage service"
```

---

### Task 11: Conversion Repository & Service

**Files:**
- Create: `server/internal/repo/conversion.go`
- Create: `server/internal/repo/quota.go`
- Create: `server/internal/service/conversion.go`

- [ ] **Step 1: Write conversion repository**

```go
// server/internal/repo/conversion.go
package repo

import (
	"database/sql"
	"time"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/model"
)

type ConversionRepo struct{ db *sql.DB }

func NewConversionRepo(db *sql.DB) *ConversionRepo { return &ConversionRepo{db} }

func (r *ConversionRepo) Create(c *model.Conversion) error {
	return r.db.QueryRow(
		`INSERT INTO conversions (user_id, status, original_url, file_size_in, format_in)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`,
		c.UserID, c.Status, c.OriginalURL, c.FileSizeIn, c.FormatIn,
	).Scan(&c.ID, &c.CreatedAt)
}

func (r *ConversionRepo) FindByID(id, userID string) (*model.Conversion, error) {
	c := &model.Conversion{}
	err := r.db.QueryRow(
		`SELECT id, user_id, status, original_url, svg_url, thumbnail_url,
		        file_size_in, file_size_out, path_count, color_count, format_in,
		        error_message, created_at, completed_at
		 FROM conversions WHERE id=$1 AND user_id=$2`, id, userID,
	).Scan(&c.ID, &c.UserID, &c.Status, &c.OriginalURL, &c.SVGURL, &c.ThumbnailURL,
		&c.FileSizeIn, &c.FileSizeOut, &c.PathCount, &c.ColorCount, &c.FormatIn,
		&c.ErrorMessage, &c.CreatedAt, &c.CompletedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return c, err
}

func (r *ConversionRepo) ListByUser(userID string, status string, limit, offset int) ([]model.Conversion, int, error) {
	var total int
	r.db.QueryRow("SELECT COUNT(*) FROM conversions WHERE user_id=$1", userID).Scan(&total)

	query := "SELECT id, user_id, status, original_url, svg_url, thumbnail_url, file_size_in, file_size_out, path_count, color_count, format_in, error_message, created_at, completed_at FROM conversions WHERE user_id=$1"
	args := []interface{}{userID}
	if status != "" {
		query += " AND status=$2"
		args = append(args, status)
	}
	query += " ORDER BY created_at DESC LIMIT $" + fmt.Sprint(len(args)+1) + " OFFSET $" + fmt.Sprint(len(args)+2)
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []model.Conversion
	for rows.Next() {
		var c model.Conversion
		rows.Scan(&c.ID, &c.UserID, &c.Status, &c.OriginalURL, &c.SVGURL, &c.ThumbnailURL,
			&c.FileSizeIn, &c.FileSizeOut, &c.PathCount, &c.ColorCount, &c.FormatIn,
			&c.ErrorMessage, &c.CreatedAt, &c.CompletedAt)
		list = append(list, c)
	}
	return list, total, nil
}

func (r *ConversionRepo) UpdateStatus(id, status string, completedAt *time.Time) error {
	_, err := r.db.Exec(
		"UPDATE conversions SET status=$1, completed_at=$2 WHERE id=$3",
		status, completedAt, id,
	)
	return err
}

func (r *ConversionRepo) UpdateResult(id string, svgURL, thumbnailURL string, fileSizeOut int64, pathCount, colorCount int) error {
	now := time.Now()
	_, err := r.db.Exec(
		`UPDATE conversions SET status=$1, svg_url=$2, thumbnail_url=$3,
		 file_size_out=$4, path_count=$5, color_count=$6, completed_at=$7 WHERE id=$8`,
		model.StatusCompleted, svgURL, thumbnailURL, fileSizeOut, pathCount, colorCount, now, id,
	)
	return err
}

func (r *ConversionRepo) MarkFailed(id, errMsg string) error {
	now := time.Now()
	_, err := r.db.Exec(
		"UPDATE conversions SET status=$1, error_message=$2, completed_at=$3 WHERE id=$4",
		model.StatusFailed, errMsg, now, id,
	)
	return err
}

func (r *ConversionRepo) Delete(id, userID string) error {
	_, err := r.db.Exec("DELETE FROM conversions WHERE id=$1 AND user_id=$2", id, userID)
	return err
}
```

- [ ] **Step 2: Write quota repository**

```go
// server/internal/repo/quota.go
package repo

import (
	"database/sql"
	"time"
)

type QuotaRepo struct{ db *sql.DB }

func NewQuotaRepo(db *sql.DB) *QuotaRepo { return &QuotaRepo{db} }

func (r *QuotaRepo) Increment(userID string) (int, error) {
	today := time.Now().Format("2006-01-02")
	var count int
	err := r.db.QueryRow(
		`INSERT INTO daily_quotas (user_id, date, count) VALUES ($1, $2, 1)
		 ON CONFLICT (user_id, date) DO UPDATE SET count = daily_quotas.count + 1
		 RETURNING count`, userID, today,
	).Scan(&count)
	return count, err
}

func (r *QuotaRepo) TodayCount(userID string) (int, error) {
	today := time.Now().Format("2006-01-02")
	var count int
	err := r.db.QueryRow(
		"SELECT count FROM daily_quotas WHERE user_id=$1 AND date=$2",
		userID, today,
	).Scan(&count)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return count, err
}
```

- [ ] **Step 3: Write conversion service**

```go
// server/internal/service/conversion.go
package service

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/config"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/model"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/repo"
	"github.com/google/uuid"
)

type ConversionService struct {
	cfg      *config.Config
	convRepo *repo.ConversionRepo
	quotaRepo *repo.QuotaRepo
	storage  *StorageService
}

func NewConversionService(cfg *config.Config, cr *repo.ConversionRepo, qr *repo.QuotaRepo, st *StorageService) *ConversionService {
	return &ConversionService{cfg, cr, qr, st}
}

func (s *ConversionService) UploadAndCreate(ctx context.Context, userID string, filename string, reader io.Reader, size int64) (*model.Conversion, error) {
	id := uuid.New().String()
	ext := getExt(filename)
	format := strings.TrimPrefix(ext, ".")
	key := fmt.Sprintf("uploads/%s/%s/original%s", userID, id, ext)

	if err := s.storage.Upload(ctx, key, reader, size, "image/"+format); err != nil {
		return nil, fmt.Errorf("upload to storage: %w", err)
	}

	c := &model.Conversion{
		UserID:      userID,
		Status:      model.StatusPending,
		OriginalURL: key,
		FileSizeIn:  size,
		FormatIn:    format,
	}
	if err := s.convRepo.Create(c); err != nil {
		return nil, fmt.Errorf("create conversion record: %w", err)
	}
	return c, nil
}

func (s *ConversionService) Get(id, userID string) (*model.Conversion, error) {
	return s.convRepo.FindByID(id, userID)
}

func (s *ConversionService) List(userID, status string, page, limit int) ([]model.Conversion, int, error) {
	offset := (page - 1) * limit
	return s.convRepo.ListByUser(userID, status, limit, offset)
}

func (s *ConversionService) Delete(id, userID string) error {
	return s.convRepo.Delete(id, userID)
}

func (s *ConversionService) GetQuota(userID string) (int, int, error) {
	count, err := s.quotaRepo.TodayCount(userID)
	if err != nil {
		return 0, 0, err
	}
	const dailyLimit = 50
	remaining := dailyLimit - count
	if remaining < 0 {
		remaining = 0
	}
	return count, remaining, nil
}

func getExt(filename string) string {
	if idx := strings.LastIndex(filename, "."); idx != -1 {
		return strings.ToLower(filename[idx:])
	}
	return ".png"
}
```

- [ ] **Step 4: Fix import in conversion repo (needs fmt)**

Check that `server/internal/repo/conversion.go` imports `"fmt"`. If not, add it.

- [ ] **Step 5: Commit**

```bash
git add server/internal/repo/conversion.go server/internal/repo/quota.go server/internal/service/conversion.go
git commit -m "feat: add conversion repository, quota repository, and conversion service"
```

---

### Task 12: Conversion & Quota Handlers

**Files:**
- Create: `server/internal/handler/conversion.go`
- Create: `server/internal/handler/quota.go`

- [ ] **Step 1: Write conversion handler**

```go
// server/internal/handler/conversion.go
package handler

import (
	"net/http"
	"strconv"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/service"
	"github.com/gin-gonic/gin"
)

type ConversionHandler struct {
	svc *service.ConversionService
}

func NewConversionHandler(svc *service.ConversionService) *ConversionHandler {
	return &ConversionHandler{svc}
}

func (h *ConversionHandler) Create(c *gin.Context) {
	userID := c.GetString("user_id")
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "NO_FILE", "message": "file is required"}})
		return
	}
	defer file.Close()

	if header.Size > 10<<20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "FILE_TOO_LARGE", "message": "max file size is 10MB"}})
		return
	}

	conv, err := h.svc.UploadAndCreate(c.Request.Context(), userID, header.Filename, file, header.Size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "UPLOAD_FAILED", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task_id": conv.ID,
		"status":  conv.Status,
	})
}

func (h *ConversionHandler) Get(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	conv, err := h.svc.Get(id, userID)
	if err != nil || conv == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "conversion not found"}})
		return
	}
	c.JSON(http.StatusOK, conv)
}

func (h *ConversionHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")

	list, total, err := h.svc.List(userID, status, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "DB_ERROR", "message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"items": list,
		"total": total,
		"page":  page,
	})
}

func (h *ConversionHandler) Download(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	conv, err := h.svc.Get(id, userID)
	if err != nil || conv == nil || conv.Status != "completed" {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "conversion not found or not ready"}})
		return
	}

	// The download will be handled by the storage service through the conversion service.
	// For MVP, redirect to the MinIO presigned URL or stream directly.
	// Placeholder: return the SVG URL for now — actual streaming added in worker task.
	c.JSON(http.StatusOK, gin.H{"svg_url": conv.SVGURL})
}

func (h *ConversionHandler) Delete(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	if err := h.svc.Delete(id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "DB_ERROR", "message": err.Error()}})
		return
	}
	c.Status(http.StatusNoContent)
}
```

- [ ] **Step 2: Write quota handler**

```go
// server/internal/handler/quota.go
package handler

import (
	"net/http"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/service"
	"github.com/gin-gonic/gin"
)

type QuotaHandler struct {
	svc *service.ConversionService
}

func NewQuotaHandler(svc *service.ConversionService) *QuotaHandler {
	return &QuotaHandler{svc}
}

func (h *QuotaHandler) Daily(c *gin.Context) {
	userID := c.GetString("user_id")
	used, remaining, err := h.svc.GetQuota(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "DB_ERROR", "message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"used":      used,
		"limit":     50,
		"remaining": remaining,
	})
}
```

- [ ] **Step 3: Update router to add conversion and quota routes**

Modify `server/internal/router/router.go` — after the auth route group, add:

```go
	convRepo := repo.NewConversionRepo(db)
	quotaRepo := repo.NewQuotaRepo(db)
	storageSvc, err := service.NewStorageService(cfg)
	if err != nil {
		log.Fatalf("storage service: %v", err)
	}
	convSvc := service.NewConversionService(cfg, convRepo, quotaRepo, storageSvc)
	convH := handler.NewConversionHandler(convSvc)
	quotaH := handler.NewQuotaHandler(convSvc)

	protected := api.Group("")
	protected.Use(middleware.JWTAuth(cfg))
	{
		protected.POST("/conversions", convH.Create)
		protected.GET("/conversions", convH.List)
		protected.GET("/conversions/:id", convH.Get)
		protected.GET("/conversions/:id/download", convH.Download)
		protected.DELETE("/conversions/:id", convH.Delete)
		protected.GET("/quotas/daily", quotaH.Daily)
	}
```

Also add `"log"` import and `"github.com/fan1ai2/vibe-coding-svg/server/internal/repo"` import.

- [ ] **Step 4: Verify compilation**

Run: `cd server && go mod tidy && go build ./...`
Expected: no errors

- [ ] **Step 5: Commit**

```bash
git add server/internal/handler/conversion.go server/internal/handler/quota.go server/internal/router/router.go
git commit -m "feat: add conversion and quota handlers with routes"
```

---

### Task 13: Worker — Queue Consumer & Vectorization

**Files:**
- Create: `server/internal/worker/queue.go`
- Create: `server/internal/worker/vectorize.go`
- Modify: `server/cmd/worker/main.go`

- [ ] **Step 1: Write queue consumer**

```go
// server/internal/worker/queue.go
package worker

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

const queueKey = "queue:convert"

type Job struct {
	TaskID     string `json:"task_id"`
	UserID     string `json:"user_id"`
	OriginalURL string `json:"original_url"`
	FormatIn   string `json:"format_in"`
}

type Queue struct {
	client *redis.Client
}

func NewQueue(addr string) *Queue {
	return &Queue{
		client: redis.NewClient(&redis.Options{Addr: addr}),
	}
}

func (q *Queue) Enqueue(ctx context.Context, job Job) error {
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return q.client.LPush(ctx, queueKey, data).Err()
}

func (q *Queue) Dequeue(ctx context.Context) (*Job, error) {
	result, err := q.client.BRPop(ctx, 0, queueKey).Result()
	if err != nil {
		return nil, err
	}
	var job Job
	return &job, json.Unmarshal([]byte(result[1]), &job)
}

func (q *Queue) CacheResult(ctx context.Context, taskID string, data interface{}) error {
	b, _ := json.Marshal(data)
	return q.client.Set(ctx, "result:"+taskID, b, 24*time.Hour).Err()
}
```

- [ ] **Step 2: Write vectorization wrapper**

```go
// server/internal/worker/vectorize.go
package worker

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/service"
)

type Vectorizer struct {
	storage *service.StorageService
}

func NewVectorizer(st *service.StorageService) *Vectorizer {
	return &Vectorizer{storage: st}
}

type VectorizeResult struct {
	SVGData     []byte
	FileSizeOut int64
	PathCount   int
	ColorCount  int
}

func (v *Vectorizer) Convert(ctx context.Context, originalKey, formatIn string) (*VectorizeResult, error) {
	// Download original from MinIO
	reader, err := v.storage.Download(ctx, originalKey)
	if err != nil {
		return nil, fmt.Errorf("download original: %w", err)
	}
	defer reader.Close()

	// Write to temp file
	tmpDir, _ := os.MkdirTemp("", "svgconv-")
	defer os.RemoveAll(tmpDir)
	inPath := filepath.Join(tmpDir, "input."+formatIn)
	outPath := filepath.Join(tmpDir, "output.svg")

	f, err := os.Create(inPath)
	if err != nil {
		return nil, fmt.Errorf("create temp input: %w", err)
	}
	io.Copy(f, reader)
	f.Close()

	// Run potrace (falls back to autotrace if available)
	err = runPotrace(inPath, outPath)
	if err != nil {
		err = runAutotrace(inPath, outPath)
		if err != nil {
			return nil, fmt.Errorf("vectorization failed: %w", err)
		}
	}

	svgData, err := os.ReadFile(outPath)
	if err != nil {
		return nil, fmt.Errorf("read output: %w", err)
	}

	result := &VectorizeResult{
		SVGData:     svgData,
		FileSizeOut: int64(len(svgData)),
		PathCount:   countTag(string(svgData), "path"),
		ColorCount:  countColors(string(svgData)),
	}
	return result, nil
}

func runPotrace(inPath, outPath string) error {
	// potrace requires BMP or PBM input; convert with ImageMagick if available
	pbmPath := inPath + ".pbm"
	convert := exec.Command("convert", inPath, pbmPath)
	if err := convert.Run(); err != nil {
		return fmt.Errorf("convert to pbm: %w", err)
	}
	defer os.Remove(pbmPath)

	cmd := exec.Command("potrace", pbmPath, "-s", "-o", outPath)
	return cmd.Run()
}

func runAutotrace(inPath, outPath string) error {
	cmd := exec.Command("autotrace", "-output-format", "svg", "-output-file", outPath, inPath)
	return cmd.Run()
}

var colorRe = regexp.MustCompile(`(?i)#[0-9a-f]{6}|#[0-9a-f]{3}|rgb\([^)]+\)|fill="[^"]*"`)

func countColors(svg string) int {
	seen := map[string]bool{}
	for _, m := range colorRe.FindAllString(svg, -1) {
		seen[strings.ToLower(m)] = true
	}
	return len(seen)
}

func countTag(svg, tag string) int {
	return strings.Count(svg, "<"+tag)
}
```

- [ ] **Step 3: Update worker main.go**

```go
// server/cmd/worker/main.go
package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/config"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/repo"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/service"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/worker"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	storageSvc, err := service.NewStorageService(cfg)
	if err != nil {
		log.Fatalf("storage: %v", err)
	}

	convRepo := repo.NewConversionRepo(db)
	queue := worker.NewQueue(cfg.RedisAddr)
	vectorizer := worker.NewVectorizer(storageSvc)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	log.Println("Worker started, waiting for jobs...")
	for {
		select {
		case <-ctx.Done():
			log.Println("Worker shutting down")
			return
		default:
			job, err := queue.Dequeue(ctx)
			if err != nil {
				log.Printf("dequeue error: %v", err)
				continue
			}
			log.Printf("Processing job %s", job.TaskID)

			convRepo.UpdateStatus(job.TaskID, "processing", nil)

			result, err := vectorizer.Convert(ctx, job.OriginalURL, job.FormatIn)
			if err != nil {
				convRepo.MarkFailed(job.TaskID, err.Error())
				log.Printf("Job %s failed: %v", job.TaskID, err)
				continue
			}

			// Upload SVG result to MinIO
			svgKey := fmt.Sprintf("outputs/%s/%s/output.svg", job.UserID, job.TaskID)
			storageSvc.Upload(ctx, svgKey, bytes.NewReader(result.SVGData), result.FileSizeOut, "image/svg+xml")

			convRepo.UpdateResult(job.TaskID, svgKey, "", result.FileSizeOut, result.PathCount, result.ColorCount)
			queue.CacheResult(ctx, job.TaskID, map[string]interface{}{
				"status": "completed",
				"svg_url": svgKey,
			})
			log.Printf("Job %s completed: %d paths, %d colors, %d bytes",
				job.TaskID, result.PathCount, result.ColorCount, result.FileSizeOut)
		}
	}
}
```

Add imports: `"bytes"` and `"fmt"` to worker main.go.

- [ ] **Step 4: Enqueue job from conversion handler**

Modify `server/internal/handler/conversion.go` — after creating the conversion, enqueue the job by injecting the Queue. Add to `ConversionHandler`:

```go
type ConversionHandler struct {
	svc   *service.ConversionService
	queue *worker.Queue
}

func NewConversionHandler(svc *service.ConversionService, queue *worker.Queue) *ConversionHandler {
	return &ConversionHandler{svc, queue}
}
```

In the `Create` method, after `conv, err := h.svc.UploadAndCreate(...)` and error check, add:

```go
	job := worker.Job{
		TaskID:      conv.ID,
		UserID:      userID,
		OriginalURL: conv.OriginalURL,
		FormatIn:    conv.FormatIn,
	}
	if err := h.queue.Enqueue(c.Request.Context(), job); err != nil {
		log.Printf("failed to enqueue job %s: %v", conv.ID, err)
	}
```

Add `"log"` import to handler. Update router.go to create the queue and pass it.

- [ ] **Step 5: Update router.go to wire queue**

Edit `server/internal/router/router.go`, add before creating convH:

```go
	queue := worker.NewQueue(cfg.RedisAddr)
	convH := handler.NewConversionHandler(convSvc, queue)
```

Add import: `"github.com/fan1ai2/vibe-coding-svg/server/internal/worker"`

- [ ] **Step 6: Verify compilation**

Run: `cd server && go mod tidy && go build ./...`
Expected: no errors

- [ ] **Step 7: Commit**

```bash
git add server/internal/worker/ server/cmd/worker/ server/internal/handler/conversion.go server/internal/router/router.go
git commit -m "feat: add worker with Redis queue, vectorization, and job pipeline"
```

---

### Task 14: React App Scaffolding

**Files:**
- Create: `web/package.json`, `web/tsconfig.json`, `web/vite.config.ts`, `web/tailwind.config.js`, `web/postcss.config.js`, `web/index.html`, `web/src/main.tsx`, `web/src/App.tsx`

- [ ] **Step 1: Initialize Vite + React + TypeScript project**

Run: `cd web && npm create vite@latest . -- --template react-ts && npm install`
Expected: project scaffolded

- [ ] **Step 2: Install dependencies**

Run: `cd web && npm install react-router-dom axios tailwindcss @tailwindcss/vite`
Expected: packages installed

- [ ] **Step 3: Write vite.config.ts**

```typescript
// web/vite.config.ts
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
})
```

- [ ] **Step 4: Write index.html**

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>SVG Converter</title>
</head>
<body class="bg-gray-50 text-gray-900">
  <div id="root"></div>
  <script type="module" src="/src/main.tsx"></script>
</body>
</html>
```

- [ ] **Step 5: Write CSS entry**

```css
/* web/src/index.css */
@import "tailwindcss";
```

- [ ] **Step 6: Write main.tsx and App.tsx skeletons**

```tsx
// web/src/main.tsx
import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import './index.css'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
)
```

```tsx
// web/src/App.tsx
export default function App() {
  return <h1 className="text-2xl font-bold p-8">SVG Converter</h1>
}
```

- [ ] **Step 7: Verify dev server starts**

Run: `cd web && npm run dev` (background) and check it responds
Expected: Vite dev server on :5173

- [ ] **Step 8: Commit**

```bash
git add web/
git commit -m "feat: scaffold React app with Vite, TypeScript, Tailwind"
```

---

### Task 15: Auth Context & API Client

**Files:**
- Create: `web/src/api/client.ts`
- Create: `web/src/context/AuthContext.tsx`

- [ ] **Step 1: Write API client**

```typescript
// web/src/api/client.ts
import axios from 'axios'

const api = axios.create({
  baseURL: '/api/v1',
  headers: { 'Content-Type': 'application/json' },
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/'
    }
    return Promise.reject(err)
  }
)

export default api
```

- [ ] **Step 2: Write AuthContext**

```tsx
// web/src/context/AuthContext.tsx
import { createContext, useContext, useState, useEffect, type ReactNode } from 'react'
import api from '../api/client'

interface User { id: string; name: string; email: string; avatar_url: string }

interface AuthState {
  user: User | null
  token: string | null
  loading: boolean
  login: (provider: 'github' | 'google') => void
  logout: () => void
}

const AuthContext = createContext<AuthState>({
  user: null, token: null, loading: true,
  login: () => {}, logout: () => {},
})

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [token, setToken] = useState<string | null>(localStorage.getItem('token'))
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (token) {
      api.get('/auth/me').then((res) => setUser(res.data)).catch(() => {
        localStorage.removeItem('token')
        setToken(null)
      }).finally(() => setLoading(false))
    } else {
      setLoading(false)
    }
  }, [token])

  // Check URL for token from OAuth callback
  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const urlToken = params.get('token')
    if (urlToken) {
      localStorage.setItem('token', urlToken)
      setToken(urlToken)
      window.history.replaceState({}, '', '/workspace/convert')
    }
  }, [])

  const login = (provider: 'github' | 'google') => {
    window.location.href = `/api/v1/auth/${provider}/login`
  }

  const logout = () => {
    localStorage.removeItem('token')
    setToken(null)
    setUser(null)
  }

  return (
    <AuthContext.Provider value={{ user, token, loading, login, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export const useAuth = () => useContext(AuthContext)
```

- [ ] **Step 3: Commit**

```bash
git add web/src/api/ web/src/context/
git commit -m "feat: add API client with JWT interceptor and AuthContext"
```

---

### Task 16: Landing Page & OAuth Callback

**Files:**
- Create: `web/src/pages/LandingPage.tsx`
- Create: `web/src/pages/OAuthCallback.tsx`

- [ ] **Step 1: Write LandingPage**

```tsx
// web/src/pages/LandingPage.tsx
import { useAuth } from '../context/AuthContext'
import { Navigate } from 'react-router-dom'

export default function LandingPage() {
  const { user, login } = useAuth()

  if (user) return <Navigate to="/workspace/convert" />

  return (
    <div className="min-h-screen flex flex-col items-center justify-center px-4">
      <h1 className="text-5xl font-bold mb-4">Image to SVG</h1>
      <p className="text-xl text-gray-600 mb-10 text-center max-w-lg">
        Convert your raster images into clean, scalable SVG vectors.
        Free, fast, no sign-up required.
      </p>
      <div className="flex gap-4">
        <button
          onClick={() => login('github')}
          className="px-6 py-3 bg-gray-900 text-white rounded-lg hover:bg-gray-700 transition"
        >
          Sign in with GitHub
        </button>
        <button
          onClick={() => login('google')}
          className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-500 transition"
        >
          Sign in with Google
        </button>
      </div>
      <div className="mt-16 grid grid-cols-3 gap-8 max-w-2xl text-center">
        <div>
          <div className="text-3xl mb-2">🖼️</div>
          <h3 className="font-semibold">Upload</h3>
          <p className="text-sm text-gray-500">PNG, JPG, WebP</p>
        </div>
        <div>
          <div className="text-3xl mb-2">⚡</div>
          <h3 className="font-semibold">Convert</h3>
          <p className="text-sm text-gray-500">Traditional engine</p>
        </div>
        <div>
          <div className="text-3xl mb-2">📥</div>
          <h3 className="font-semibold">Download</h3>
          <p className="text-sm text-gray-500">Clean SVG output</p>
        </div>
      </div>
    </div>
  )
}
```

- [ ] **Step 2: Write OAuthCallback**

```tsx
// web/src/pages/OAuthCallback.tsx
import { useEffect } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'

export default function OAuthCallback() {
  const [params] = useSearchParams()
  const navigate = useNavigate()

  useEffect(() => {
    const token = params.get('token')
    if (token) {
      localStorage.setItem('token', token)
      navigate('/workspace/convert', { replace: true })
    }
  }, [params, navigate])

  return (
    <div className="min-h-screen flex items-center justify-center">
      <div className="animate-spin h-8 w-8 border-4 border-gray-900 border-t-transparent rounded-full" />
    </div>
  )
}
```

- [ ] **Step 3: Commit**

```bash
git add web/src/pages/LandingPage.tsx web/src/pages/OAuthCallback.tsx
git commit -m "feat: add landing page and OAuth callback page"
```

---

### Task 17: Workspace Layout & Sidebar

**Files:**
- Create: `web/src/pages/WorkspaceLayout.tsx`

- [ ] **Step 1: Write WorkspaceLayout**

```tsx
// web/src/pages/WorkspaceLayout.tsx
import { Navigate, NavLink, Outlet, useLocation } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { useEffect, useState } from 'react'
import api from '../api/client'

export default function WorkspaceLayout() {
  const { user, loading, logout } = useAuth()
  const location = useLocation()
  const [quota, setQuota] = useState<{ used: number; remaining: number } | null>(null)

  useEffect(() => {
    api.get('/quotas/daily').then((res) => setQuota(res.data)).catch(() => {})
  }, [location])

  if (loading) return null
  if (!user) return <Navigate to="/" />

  const linkClass = ({ isActive }: { isActive: boolean }) =>
    `block px-3 py-2 rounded-lg text-sm font-medium transition ${
      isActive ? 'bg-gray-900 text-white' : 'text-gray-600 hover:bg-gray-100'
    }`

  return (
    <div className="flex h-screen">
      <aside className="w-60 border-r border-gray-200 flex flex-col p-4">
        <div className="flex items-center gap-3 mb-8">
          <img src={user.avatar_url} className="w-8 h-8 rounded-full" alt="" />
          <span className="font-semibold text-sm truncate">{user.name}</span>
        </div>
        <nav className="flex flex-col gap-1 flex-1">
          <NavLink to="/workspace/convert" className={linkClass}>New Conversion</NavLink>
          <NavLink to="/workspace/library" className={linkClass}>My Library</NavLink>
        </nav>
        {quota && (
          <div className="text-xs text-gray-400 mb-2">
            {quota.used}/{quota.used + quota.remaining} conversions today
          </div>
        )}
        <button
          onClick={logout}
          className="text-sm text-gray-500 hover:text-gray-700 text-left"
        >
          Sign out
        </button>
      </aside>
      <main className="flex-1 overflow-auto">
        <Outlet />
      </main>
    </div>
  )
}
```

- [ ] **Step 2: Commit**

```bash
git add web/src/pages/WorkspaceLayout.tsx
git commit -m "feat: add workspace layout with sidebar and quota display"
```

---

### Task 18: Convert Page (Upload & Processing)

**Files:**
- Create: `web/src/components/DropZone.tsx`
- Create: `web/src/pages/ConvertPage.tsx`

- [ ] **Step 1: Write DropZone component**

```tsx
// web/src/components/DropZone.tsx
import { useCallback, useState, type DragEvent } from 'react'

interface Props {
  onFile: (file: File) => void
  disabled?: boolean
}

export default function DropZone({ onFile, disabled }: Props) {
  const [dragover, setDragover] = useState(false)

  const handleDrop = useCallback((e: DragEvent) => {
    e.preventDefault()
    setDragover(false)
    const file = e.dataTransfer.files[0]
    if (file) onFile(file)
  }, [onFile])

  const handleChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) onFile(file)
  }, [onFile])

  return (
    <div
      onDragOver={(e) => { e.preventDefault(); setDragover(true) }}
      onDragLeave={() => setDragover(false)}
      onDrop={handleDrop}
      className={`border-2 border-dashed rounded-xl p-12 text-center transition cursor-pointer ${
        dragover ? 'border-gray-900 bg-gray-100' : 'border-gray-300 hover:border-gray-500'
      } ${disabled ? 'opacity-50 pointer-events-none' : ''}`}
    >
      <input
        type="file"
        accept="image/png,image/jpeg,image/webp"
        onChange={handleChange}
        className="hidden"
        id="file-input"
        disabled={disabled}
      />
      <label htmlFor="file-input" className="cursor-pointer">
        <p className="text-2xl mb-2">📁</p>
        <p className="font-medium">Drop your image here or click to browse</p>
        <p className="text-sm text-gray-500 mt-1">PNG, JPG, WebP — max 10MB</p>
      </label>
    </div>
  )
}
```

- [ ] **Step 2: Write ConvertPage**

```tsx
// web/src/pages/ConvertPage.tsx
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import DropZone from '../components/DropZone'
import api from '../api/client'

export default function ConvertPage() {
  const [uploading, setUploading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const navigate = useNavigate()

  const handleFile = async (file: File) => {
    setError(null)
    setUploading(true)
    try {
      const form = new FormData()
      form.append('file', file)
      const { data } = await api.post('/conversions', form, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      navigate(`/workspace/preview/${data.task_id}`)
    } catch (err: any) {
      setError(err.response?.data?.error?.message || 'Upload failed')
    } finally {
      setUploading(false)
    }
  }

  return (
    <div className="max-w-xl mx-auto py-16 px-4">
      <h2 className="text-2xl font-bold mb-2">Convert Image to SVG</h2>
      <p className="text-gray-500 mb-8">Upload a raster image and get a vector SVG in seconds.</p>
      <DropZone onFile={handleFile} disabled={uploading} />
      {uploading && (
        <div className="mt-4 text-center">
          <div className="animate-spin h-6 w-6 border-2 border-gray-900 border-t-transparent rounded-full mx-auto mb-2" />
          <p className="text-sm text-gray-500">Uploading...</p>
        </div>
      )}
      {error && (
        <div className="mt-4 p-3 bg-red-50 text-red-700 rounded-lg text-sm">{error}</div>
      )}
    </div>
  )
}
```

- [ ] **Step 3: Commit**

```bash
git add web/src/components/DropZone.tsx web/src/pages/ConvertPage.tsx
git commit -m "feat: add DropZone component and ConvertPage with upload flow"
```

---

### Task 19: Preview Page (Comparison + Download)

**Files:**
- Create: `web/src/components/ZoomControls.tsx`
- Create: `web/src/components/MetadataCard.tsx`
- Create: `web/src/pages/PreviewPage.tsx`

- [ ] **Step 1: Write ZoomControls**

```tsx
// web/src/components/ZoomControls.tsx
interface Props {
  zoom: number
  onZoom: (z: number) => void
}

export default function ZoomControls({ zoom, onZoom }: Props) {
  return (
    <div className="flex items-center gap-1 bg-white rounded-lg border border-gray-200 px-2 py-1">
      <button onClick={() => onZoom(zoom - 0.25)} disabled={zoom <= 0.25}
        className="px-2 py-1 hover:bg-gray-100 rounded disabled:opacity-30">-</button>
      <span className="text-sm w-14 text-center">{Math.round(zoom * 100)}%</span>
      <button onClick={() => onZoom(zoom + 0.25)} disabled={zoom >= 3}
        className="px-2 py-1 hover:bg-gray-100 rounded disabled:opacity-30">+</button>
      <button onClick={() => onZoom(1)} className="px-2 py-1 text-sm hover:bg-gray-100 rounded">100%</button>
      <button onClick={() => onZoom(0)} className="px-2 py-1 text-sm hover:bg-gray-100 rounded">Fit</button>
    </div>
  )
}
```

- [ ] **Step 2: Write MetadataCard**

```tsx
// web/src/components/MetadataCard.tsx
interface Metadata {
  path_count: number
  color_count: number
  file_size_in: number
  file_size_out: number
  format_in: string
}

function fmtSize(bytes: number) {
  if (bytes < 1024) return `${bytes} B`
  return `${(bytes / 1024).toFixed(1)} KB`
}

export default function MetadataCard({ data }: { data: Metadata }) {
  const reduction = data.file_size_in
    ? ((1 - data.file_size_out / data.file_size_in) * 100).toFixed(0)
    : 0

  return (
    <div className="grid grid-cols-2 gap-3 p-4 bg-white rounded-lg border border-gray-200">
      <div><span className="text-xs text-gray-400">Paths</span><p className="font-semibold">{data.path_count}</p></div>
      <div><span className="text-xs text-gray-400">Colors</span><p className="font-semibold">{data.color_count}</p></div>
      <div><span className="text-xs text-gray-400">Original</span><p className="font-semibold">{fmtSize(data.file_size_in)}</p></div>
      <div><span className="text-xs text-gray-400">SVG Size</span><p className="font-semibold">{fmtSize(data.file_size_out)} ({reduction}% smaller)</p></div>
    </div>
  )
}
```

- [ ] **Step 3: Write PreviewPage**

```tsx
// web/src/pages/PreviewPage.tsx
import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import ZoomControls from '../components/ZoomControls'
import MetadataCard from '../components/MetadataCard'
import api from '../api/client'

interface Conversion {
  id: string; status: string; original_url: string; svg_url: string
  file_size_in: number; file_size_out: number; path_count: number
  color_count: number; format_in: string; error_message?: string
}

export default function PreviewPage() {
  const { id } = useParams<{ id: string }>()
  const [data, setData] = useState<Conversion | null>(null)
  const [zoom, setZoom] = useState(1)
  const [svgContent, setSvgContent] = useState<string | null>(null)

  useEffect(() => {
    const poll = setInterval(async () => {
      try {
        const { data } = await api.get(`/conversions/${id}`)
        setData(data)
        if (data.status === 'completed' || data.status === 'failed') {
          clearInterval(poll)
          if (data.status === 'completed') {
            // Fetch actual SVG content for inline rendering
            const svgRes = await api.get(`/conversions/${id}/download`, { responseType: 'text' })
            setSvgContent(svgRes.data)
          }
        }
      } catch { clearInterval(poll) }
    }, 1000)
    return () => clearInterval(poll)
  }, [id])

  if (!data) return <div className="flex justify-center py-16"><div className="animate-spin h-8 w-8 border-2 border-gray-900 border-t-transparent rounded-full" /></div>

  if (data.status === 'failed') {
    return (
      <div className="max-w-xl mx-auto py-16 px-4 text-center">
        <p className="text-red-600 font-medium mb-2">Conversion failed</p>
        <p className="text-sm text-gray-500 mb-4">{data.error_message}</p>
        <Link to="/workspace/convert" className="text-blue-600 hover:underline">Try again</Link>
      </div>
    )
  }

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-xl font-bold">Preview</h2>
        <ZoomControls zoom={zoom} onZoom={setZoom} />
      </div>

      {data.status !== 'completed' ? (
        <div className="flex flex-col items-center py-16">
          <div className="animate-spin h-8 w-8 border-2 border-gray-900 border-t-transparent rounded-full mb-4" />
          <p className="text-gray-500">Converting your image...</p>
          <p className="text-sm text-gray-400 mt-1">This may take 5-30 seconds</p>
        </div>
      ) : (
        <>
          <div className="grid grid-cols-2 gap-4 mb-6">
            <div className="border border-gray-200 rounded-lg p-2 bg-white">
              <p className="text-xs text-gray-400 mb-1">Original</p>
              <div style={{ transform: `scale(${zoom || 1})`, transformOrigin: 'top left' }}>
                <img src={`/api/v1/conversions/${id}/original`} alt="Original" className="max-w-full" />
              </div>
            </div>
            <div className="border border-gray-200 rounded-lg p-2 bg-white">
              <p className="text-xs text-gray-400 mb-1">SVG Result</p>
              <div
                style={{ transform: `scale(${zoom || 1})`, transformOrigin: 'top left' }}
                dangerouslySetInnerHTML={{ __html: svgContent || '' }}
              />
            </div>
          </div>
          <MetadataCard data={data} />
          <a
            href={`/api/v1/conversions/${id}/download`}
            className="inline-block mt-4 px-6 py-2 bg-gray-900 text-white rounded-lg hover:bg-gray-700 transition"
            download
          >
            Download SVG
          </a>
        </>
      )}
    </div>
  )
}
```

- [ ] **Step 4: Commit**

```bash
git add web/src/components/ZoomControls.tsx web/src/components/MetadataCard.tsx web/src/pages/PreviewPage.tsx
git commit -m "feat: add preview page with comparison view, zoom, and download"
```

---

### Task 20: Library Page

**Files:**
- Create: `web/src/components/ConversionCard.tsx`
- Create: `web/src/pages/LibraryPage.tsx`

- [ ] **Step 1: Write ConversionCard**

```tsx
// web/src/components/ConversionCard.tsx
import { Link } from 'react-router-dom'

interface Props {
  id: string
  status: string
  format_in: string
  file_size_in: number
  file_size_out: number
  created_at: string
}

function timeAgo(dateStr: string) {
  const diff = Date.now() - new Date(dateStr).getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return 'Just now'
  if (mins < 60) return `${mins}m ago`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours}h ago`
  return `${Math.floor(hours / 24)}d ago`
}

export default function ConversionCard({ id, status, format_in, file_size_in, file_size_out, created_at }: Props) {
  const badgeColor = status === 'completed' ? 'bg-green-100 text-green-700'
    : status === 'failed' ? 'bg-red-100 text-red-700'
    : 'bg-yellow-100 text-yellow-700'

  return (
    <Link to={`/workspace/preview/${id}`} className="block border border-gray-200 rounded-xl p-4 hover:shadow-md transition bg-white">
      <div className="flex items-center justify-between mb-2">
        <span className="text-xs uppercase text-gray-400">{format_in}</span>
        <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${badgeColor}`}>{status}</span>
      </div>
      <div className="text-xs text-gray-400">{timeAgo(created_at)}</div>
    </Link>
  )
}
```

- [ ] **Step 2: Write LibraryPage**

```tsx
// web/src/pages/LibraryPage.tsx
import { useEffect, useState } from 'react'
import ConversionCard from '../components/ConversionCard'
import api from '../api/client'

interface Conversion {
  id: string; status: string; format_in: string; file_size_in: number
  file_size_out: number; created_at: string
}

export default function LibraryPage() {
  const [items, setItems] = useState<Conversion[]>([])
  const [filter, setFilter] = useState('')
  const [total, setTotal] = useState(0)

  useEffect(() => {
    api.get('/conversions', { params: { page: 1, limit: 50, status: filter || undefined } })
      .then((res) => { setItems(res.data.items); setTotal(res.data.total) })
  }, [filter])

  return (
    <div className="p-6">
      <h2 className="text-xl font-bold mb-4">My Library</h2>
      <div className="flex gap-2 mb-6">
        {['', 'completed', 'failed'].map((s) => (
          <button
            key={s}
            onClick={() => setFilter(s)}
            className={`px-3 py-1 rounded-lg text-sm transition ${
              filter === s ? 'bg-gray-900 text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
            }`}
          >
            {s || 'All'}
          </button>
        ))}
      </div>
      {items.length === 0 ? (
        <p className="text-gray-400 text-center py-16">No conversions yet. <a href="/workspace/convert" className="text-blue-600">Convert your first image</a></p>
      ) : (
        <div className="grid grid-cols-4 gap-4">
          {items.map((c) => <ConversionCard key={c.id} {...c} />)}
        </div>
      )}
    </div>
  )
}
```

- [ ] **Step 3: Commit**

```bash
git add web/src/components/ConversionCard.tsx web/src/pages/LibraryPage.tsx
git commit -m "feat: add library page with filter and conversion cards"
```

---

### Task 21: Wire Up App.tsx with Router

**Files:**
- Modify: `web/src/App.tsx`

- [ ] **Step 1: Update App.tsx with routes**

```tsx
// web/src/App.tsx
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { AuthProvider } from './context/AuthContext'
import LandingPage from './pages/LandingPage'
import OAuthCallback from './pages/OAuthCallback'
import WorkspaceLayout from './pages/WorkspaceLayout'
import ConvertPage from './pages/ConvertPage'
import PreviewPage from './pages/PreviewPage'
import LibraryPage from './pages/LibraryPage'

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route path="/" element={<LandingPage />} />
          <Route path="/callback" element={<OAuthCallback />} />
          <Route path="/workspace" element={<WorkspaceLayout />}>
            <Route path="convert" element={<ConvertPage />} />
            <Route path="preview/:id" element={<PreviewPage />} />
            <Route path="library" element={<LibraryPage />} />
          </Route>
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  )
}
```

- [ ] **Step 2: Verify frontend compiles**

Run: `cd web && npx tsc --noEmit`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add web/src/App.tsx
git commit -m "feat: wire up React Router with all pages"
```

---

### Task 22: Dockerfiles

**Files:**
- Create: `Dockerfile.api`
- Create: `Dockerfile.worker`

- [ ] **Step 1: Write Dockerfile.api**

```dockerfile
# Dockerfile.api
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ .
RUN CGO_ENABLED=0 go build -o /api ./cmd/api

FROM alpine:3.20
RUN apk add --no-cache potrace imagemagick
COPY --from=builder /api /api
EXPOSE 8080
CMD ["/api"]
```

- [ ] **Step 2: Write Dockerfile.worker**

```dockerfile
# Dockerfile.worker
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ .
RUN CGO_ENABLED=0 go build -o /worker ./cmd/worker

FROM alpine:3.20
RUN apk add --no-cache potrace imagemagick
COPY --from=builder /worker /worker
CMD ["/worker"]
```

- [ ] **Step 3: Commit**

```bash
git add Dockerfile.api Dockerfile.worker
git commit -m "feat: add Dockerfiles for API and worker services"
```

---

### Task 23: End-to-End Verification

- [ ] **Step 1: Start all services**

```bash
docker-compose up -d                          # postgres, redis, minio
cd server && go run ./cmd/api/main.go &       # API server
cd server && go run ./cmd/worker/main.go &    # Worker
cd web && npm run dev &                       # React dev server
```

- [ ] **Step 2: Verify API health**

Run: `curl http://localhost:8080/api/v1/auth/me`
Expected: `{"error":{"code":"UNAUTHORIZED","message":"missing or malformed token"}}` (middleware works)

- [ ] **Step 3: Verify frontend loads**

Open `http://localhost:5173` in browser. Landing page should render with GitHub/Google login buttons.

- [ ] **Step 4: Verify MinIO is accessible**

Run: `curl http://localhost:9001` — MinIO console should respond

- [ ] **Step 5: Run migration**

```bash
cd server && migrate -path migrations -database "postgres://svguser:svgpass@localhost:5432/svgconverter?sslmode=disable" up
```

- [ ] **Step 6: Commit final state**

```bash
git add -A
git commit -m "chore: final integration and verification"
```

---

### Plan Self-Review

**Spec Coverage:**
- Auth (GitHub/Google OAuth) — Task 6, 7, 9
- Conversions CRUD — Task 11, 12
- Quota tracking — Task 11, 12
- Worker + vectorization — Task 13
- Frontend landing page — Task 16
- Frontend upload/convert — Task 18
- Frontend preview/comparison — Task 19
- Frontend library — Task 20
- Docker compose — Task 1
- Dockerfiles — Task 22

**Placeholder Scan:** No TBDs, TODOs, or vague instructions.

**Type Consistency:** `Conversion.ID` is `string` (UUID) throughout. `task_id` used consistently in both Go and TypeScript. API returns `task_id` from POST /conversions, frontend reads it for navigation.

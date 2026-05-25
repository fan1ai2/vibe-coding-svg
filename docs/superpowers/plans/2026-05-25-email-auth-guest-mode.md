# Email Auth + Guest Mode Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add three login paths (guest mode, email verification-code, GitHub OAuth) with guest 3-conversion quota and SMTP email verification.

**Architecture:** Guest users are regular User rows with provider="guest", reusing existing JWT middleware and quota infrastructure. Email auth uses SMTP-sent 6-digit verification codes (5min TTL). Frontend enforces guest quota in localStorage; backend hard-caps at 3 lifetime conversions for guest provider.

**Tech Stack:** Go + Gin + PostgreSQL + SMTP (net/smtp), React + TypeScript + Tailwind

---

### Task 1: Add SMTP config fields

**Files:**
- Modify: `server/internal/config/config.go`

- [ ] **Step 1: Add SMTP fields to Config struct and Load()**

Replace the Config struct to add SMTP fields:

```go
type Config struct {
	Port           string
	DatabaseURL    string
	RedisAddr      string
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
	JWTSecret      string
	GithubClientID string
	GithubSecret   string
	MaxFileSize    int64
	FrontendURL    string
	SMTPHost       string
	SMTPPort       int
	SMTPUser       string
	SMTPPassword   string
	SMTPFrom       string
}
```

In `Load()`, add:

```go
SMTPHost:     os.Getenv("SMTP_HOST"),
SMTPPort:     intEnvOr("SMTP_PORT", 587),
SMTPUser:     os.Getenv("SMTP_USER"),
SMTPPassword: os.Getenv("SMTP_PASSWORD"),
SMTPFrom:     os.Getenv("SMTP_FROM"),
```

Note: SMTP fields use `os.Getenv` (not `require`) so they're optional — the server starts without SMTP for dev environments.

- [ ] **Step 2: Verify code compiles**

Run: `cd server && go build ./...`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add server/internal/config/config.go
git commit -m "feat: add SMTP config fields"
```

---

### Task 2: Create email sending service

**Files:**
- Create: `server/internal/service/email.go`

- [ ] **Step 1: Write the email service**

```go
package service

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/config"
)

type EmailService struct {
	cfg *config.Config
}

func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{cfg}
}

func (s *EmailService) SendVerificationCode(to, code string) error {
	if s.cfg.SMTPHost == "" {
		return fmt.Errorf("SMTP not configured")
	}

	subject := "验证码 - SVG Converter"
	body := fmt.Sprintf("您的验证码是：%s（5 分钟内有效）\n\n如果这不是您的操作，请忽略此邮件。", code)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		s.cfg.SMTPFrom, to, subject, body)

	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)
	auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPassword, s.cfg.SMTPHost)

	// Try STARTTLS first, fall back to plain
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{ServerName: s.cfg.SMTPHost}
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("starttls: %w", err)
		}
	}

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}
	if err := client.Mail(s.cfg.SMTPFrom); err != nil {
		return fmt.Errorf("smtp mail: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}
	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	_, err = fmt.Fprint(wc, msg)
	if err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	return wc.Close()
}
```

- [ ] **Step 2: Verify compilation**

Run: `cd server && go build ./...`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add server/internal/service/email.go
git commit -m "feat: add SMTP email verification code service"
```

---

### Task 3: Create database migrations

**Files:**
- Create: `server/migrations/004_verification_codes.up.sql`
- Create: `server/migrations/004_verification_codes.down.sql`
- Create: `server/migrations/005_guest_provider.up.sql`
- Create: `server/migrations/005_guest_provider.down.sql`

- [ ] **Step 1: Create verification_codes migration (up)**

Write `server/migrations/004_verification_codes.up.sql`:

```sql
CREATE TABLE verification_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL,
    code VARCHAR(6) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_verification_codes_email_code ON verification_codes(email, code);
```

- [ ] **Step 2: Create verification_codes migration (down)**

Write `server/migrations/004_verification_codes.down.sql`:

```sql
DROP TABLE IF EXISTS verification_codes;
```

- [ ] **Step 3: Create guest provider migration (up)**

Write `server/migrations/005_guest_provider.up.sql`:

```sql
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_provider_provider_id_key;
CREATE UNIQUE INDEX IF NOT EXISTS users_provider_unique ON users(provider, provider_id)
    WHERE provider != 'guest';
```

- [ ] **Step 4: Create guest provider migration (down)**

Write `server/migrations/005_guest_provider.down.sql`:

```sql
DROP INDEX IF EXISTS users_provider_unique;
ALTER TABLE users ADD CONSTRAINT users_provider_provider_id_key UNIQUE (provider, provider_id);
```

- [ ] **Step 5: Commit**

```bash
git add server/migrations/
git commit -m "feat: add verification_codes table and relax guest user constraint"
```

---

### Task 4: Add repo methods for verification codes and user queries

**Files:**
- Modify: `server/internal/repo/user.go`

- [ ] **Step 1: Add FindByEmail, CreateGuest, and verification code methods**

Add to `server/internal/repo/user.go`:

```go
import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/model"
)

// ... (keep existing UserRepo struct and methods, add these new ones)

func (r *UserRepo) FindByEmail(email string) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRow(
		`SELECT id, email, name, avatar_url, provider, provider_id, created_at, updated_at
		 FROM users WHERE email=$1`, email,
	).Scan(&u.ID, &u.Email, &u.Name, &u.AvatarURL, &u.Provider, &u.ProviderID, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func (r *UserRepo) CreateGuest() (*model.User, error) {
	guestID := randomHex(16)
	u := &model.User{
		Name:       "游客",
		Provider:   "guest",
		ProviderID: guestID,
	}
	err := r.db.QueryRow(
		`INSERT INTO users (name, provider, provider_id) VALUES ($1,$2,$3)
		 RETURNING id, email, created_at, updated_at`,
		u.Name, u.Provider, u.ProviderID,
	).Scan(&u.ID, &u.Email, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (r *UserRepo) FindByGuestID(guestID string) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRow(
		`SELECT id, email, name, avatar_url, provider, provider_id, created_at, updated_at
		 FROM users WHERE provider='guest' AND provider_id=$1`, guestID,
	).Scan(&u.ID, &u.Email, &u.Name, &u.AvatarURL, &u.Provider, &u.ProviderID, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func (r *UserRepo) SaveVerificationCode(email, code string) error {
	_, err := r.db.Exec(
		`INSERT INTO verification_codes (email, code, expires_at) VALUES ($1,$2,$3)`,
		email, code, time.Now().Add(5*time.Minute),
	)
	return err
}

func (r *UserRepo) VerifyCode(email, code string) (bool, error) {
	var used bool
	err := r.db.QueryRow(
		`SELECT used FROM verification_codes
		 WHERE email=$1 AND code=$2 AND expires_at > now()
		 ORDER BY created_at DESC LIMIT 1`,
		email, code,
	).Scan(&used)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if used {
		return false, nil
	}
	_, err = r.db.Exec(`UPDATE verification_codes SET used=true WHERE email=$1 AND code=$2`, email, code)
	return !used, err
}

func (r *UserRepo) LastCodeSentAt(email string) (time.Time, error) {
	var t time.Time
	err := r.db.QueryRow(
		`SELECT created_at FROM verification_codes
		 WHERE email=$1 ORDER BY created_at DESC LIMIT 1`,
		email,
	).Scan(&t)
	if err == sql.ErrNoRows {
		return time.Time{}, nil
	}
	return t, err
}

func (r *UserRepo) CleanupExpiredCodes() error {
	_, err := r.db.Exec(`DELETE FROM verification_codes WHERE expires_at < now() - INTERVAL '1 hour'`)
	return err
}

func randomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
```

Note: The import block needs `"crypto/rand"`, `"encoding/hex"`, and `"time"` added. The existing `FindByProvider`, `Create`, `UpsertByProvider`, `FindByID` methods remain unchanged. `nullIfEmpty` helper also remains.

- [ ] **Step 2: Verify compilation**

Run: `cd server && go build ./...`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add server/internal/repo/user.go
git commit -m "feat: add repo methods for guest, email auth, and verification codes"
```

---

### Task 5: Add auth service methods (GuestLogin, EmailSendCode, EmailVerify)

**Files:**
- Modify: `server/internal/service/auth.go`

- [ ] **Step 1: Add new methods to AuthService**

Add a new file or append to existing `server/internal/service/auth.go`. Since AuthService needs new dependencies (EmailService, UserRepo methods), update the struct and add methods:

Replace the `AuthService` struct:

```go
type AuthService struct {
	cfg      *config.Config
	userRepo *repo.UserRepo
	emailSvc *EmailService
}

func NewAuthService(cfg *config.Config, ur *repo.UserRepo, es *EmailService) *AuthService {
	return &AuthService{cfg, ur, es}
}
```

Add these new methods at the end of the file (after `firstNonEmpty`):

```go
import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

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
	// Rate limit: 1 per 60 seconds per email
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
			Name:       email[:len(email)-len(email)+strings.Index(email, "@")],
			Provider:   "email",
			ProviderID: email,
		}
		// Extract name from email: "foo@bar.com" → "foo"
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
```

Note: The import block already has `"time"`, `"fmt"`, `"net/http"`, `"encoding/json"`, `"errors"`, `"github.com/golang-jwt/jwt/v5"`, and the internal packages. Need to add `"crypto/rand"`, `"math/big"`, `"strings"`. Keep all existing methods (GenerateJWT, ExchangeGithubCode, getGithubAccessToken, getGithubUser, FindByID, firstNonEmpty) unchanged.

The `NewAuthService` signature changed — this will break the caller in `router.go` (handled in Task 7).

- [ ] **Step 2: Verify compilation**

Run: `cd server && go build ./...`
Expected: error in router.go (NewAuthService args mismatch) — expected, will fix in Task 7

- [ ] **Step 3: Commit**

```bash
git add server/internal/service/auth.go
git commit -m "feat: add GuestLogin, EmailSendCode, EmailVerify service methods"
```

---

### Task 6: Add auth handler methods

**Files:**
- Modify: `server/internal/handler/auth.go`

- [ ] **Step 1: Add handler methods**

Add three new handler methods and update `AuthHandler`:

```go
import (
	"log"
	"strings"
	"time"
	// ... keep existing imports
)

// Add these handler methods after the existing Me handler:

// GuestLogin handles guest user creation/login.
func (h *AuthHandler) GuestLogin(c *gin.Context) {
	guestID, _ := c.Cookie("guest_id")

	user, token, newGuestID, err := h.authService.GuestLogin(guestID)
	if err != nil {
		log.Printf("[ERROR] guest login: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "登录失败，请重试"}})
		return
	}

	// Set long-lived guest cookie (1 year)
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
```

- [ ] **Step 2: Verify compilation**

Run: `cd server && go build ./...`
Expected: error in router.go (NewAuthService args + new handlers not wired) — expected, will fix next

- [ ] **Step 3: Commit**

```bash
git add server/internal/handler/auth.go
git commit -m "feat: add guest login, email send-code, email verify handlers"
```

---

### Task 7: Update router and add guest quota enforcement in conversion service

**Files:**
- Modify: `server/internal/router/router.go`
- Modify: `server/internal/service/conversion.go`

- [ ] **Step 1: Update router.go — wire new dependencies and add routes**

Update `server/internal/router/router.go`. The changes:
1. Create EmailService
2. Pass it to NewAuthService
3. Add new auth routes

```go
// In Setup(), update the auth module section:

// --- 认证模块 ---
userRepo := repo.NewUserRepo(db)
emailSvc := service.NewEmailService(cfg)
authSvc := service.NewAuthService(cfg, userRepo, emailSvc)
authH := handler.NewAuthHandler(cfg, authSvc)
```

Then in the auth route group, add the new routes:

```go
auth := api.Group("/auth")
{
	// Guest
	auth.POST("/guest", authH.GuestLogin)

	// Email
	auth.POST("/email/send-code", authH.EmailSendCode)
	auth.POST("/email/verify", authH.EmailVerify)

	// GitHub (existing)
	auth.GET("/github/login", authH.GithubLogin)
	auth.GET("/github/callback", authH.GithubCallback)
	auth.POST("/refresh", middleware.JWTAuth(cfg), authH.Refresh)
	auth.GET("/me", middleware.JWTAuth(cfg), authH.Me)
}
```

- [ ] **Step 2: Add guest quota enforcement in conversion.go**

In `server/internal/service/conversion.go`, add the guest quota check at the start of `Enqueue()`, before the file extension parsing:

```go
func (s *ConversionService) Enqueue(userID string, file io.Reader, filename string, size int64) (*model.Conversion, error) {
	// Guest quota: max 3 lifetime conversions
	if err := s.checkGuestQuota(userID); err != nil {
		return nil, err
	}

	// ... rest of existing Enqueue method unchanged
}
```

Add the `checkGuestQuota` helper method at the end of the file:

```go
import (
	"errors"
	// ... add to existing imports
)

func (s *ConversionService) checkGuestQuota(userID string) error {
	// Only check for guest users — query the provider by user_id
	// Use the existing repo's DB connection
	count, err := s.repo.CountConversionsByUser(userID)
	if err != nil {
		return fmt.Errorf("quota check: %w", err)
	}
	if count >= 3 {
		return errors.New("guest quota exhausted")
	}
	return nil
}
```

Wait — we need `CountConversionsByUser` and also need to know if user is guest. Let's add a simpler approach: add a method to ConversionRepo.

Actually, let's add to `repo/conversion.go`:

```go
func (r *ConversionRepo) CountByUserID(userID string) (int, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM conversions WHERE user_id=$1`, userID,
	).Scan(&count)
	return count, err
}
```

Then in `service/conversion.go`, `checkGuestQuota` becomes:

```go
func (s *ConversionService) checkGuestQuota(userID string) error {
	count, err := s.repo.CountByUserID(userID)
	if err != nil {
		return fmt.Errorf("quota check: %w", err)
	}
	if count >= 3 {
		return fmt.Errorf("guest quota exhausted (%d/3 conversions used)", count)
	}
	return nil
}
```

But this applies to ALL users, not just guests. For registered users we want the existing daily quota (20/day), not a lifetime cap of 3.

We need to check the user's provider. Let's add a `FindByID` method lookup. Actually, userRepo.FindByID already exists. So in ConversionService we need access to userRepo, or we pass the provider info.

Simplest approach: add a `GetProvider` method to ConversionRepo, or pass it through the handler.

Actually, the cleanest approach: modify `checkGuestQuota` to take a `provider` string parameter. The handler can get this from the auth middleware (we set `user_id` in context, but we also need the provider).

Let's update the JWT middleware to also set `provider` in the context, and then check it in the handler → service call.

Actually, the simplest approach: just add user_id to context (already done), and in the conversion handler, call a method that checks both. Let me add it to ConversionService:

```go
func (s *ConversionService) checkGuestQuota(userID string) error {
	// Always count. But only enforce for guests.
	// We check provider by looking up the user.
	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		return fmt.Errorf("quota check: %w", err)
	}
	if user == nil || user.Provider != "guest" {
		return nil
	}
	count, err := s.repo.CountByUserID(userID)
	if err != nil {
		return fmt.Errorf("quota check: %w", err)
	}
	if count >= 3 {
		return fmt.Errorf("guest quota exhausted (%d/3 conversions used)", count)
	}
	return nil
}
```

For this we need `FindUserByID` on ConversionRepo, or inject UserRepo into ConversionService. Let's add a simple FindProviderByID to ConversionRepo:

```go
func (r *ConversionRepo) FindProviderByID(userID string) (string, error) {
	var provider string
	err := r.db.QueryRow(`SELECT provider FROM users WHERE id=$1`, userID).Scan(&provider)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return provider, err
}
```

Let me revise the plan. The changes to `repo/conversion.go`:

Add:
```go
func (r *ConversionRepo) FindProviderByID(userID string) (string, error) {
	var provider string
	err := r.db.QueryRow(`SELECT provider FROM users WHERE id=$1`, userID).Scan(&provider)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return provider, err
}

func (r *ConversionRepo) CountByUserID(userID string) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM conversions WHERE user_id=$1`, userID).Scan(&count)
	return count, err
}
```

And in `service/conversion.go`, add at the start of `Enqueue`:

```go
// Guest quota: 3 lifetime conversions max
provider, err := s.repo.FindProviderByID(userID)
if err != nil {
	return nil, fmt.Errorf("quota check: %w", err)
}
if provider == "guest" {
	count, err := s.repo.CountByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("quota count: %w", err)
	}
	if count >= 3 {
		return nil, fmt.Errorf("试用次数已用完（%d/3），请登录后继续使用", count)
	}
}
```

- [ ] **Step 1 (revised): Add repo methods**

Add to `server/internal/repo/conversion.go`:

```go
func (r *ConversionRepo) FindProviderByID(userID string) (string, error) {
	var provider string
	err := r.db.QueryRow(`SELECT provider FROM users WHERE id=$1`, userID).Scan(&provider)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return provider, err
}

func (r *ConversionRepo) CountByUserID(userID string) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM conversions WHERE user_id=$1`, userID).Scan(&count)
	return count, err
}
```

- [ ] **Step 2: Add guest quota check in Enqueue**

In `server/internal/service/conversion.go`, add at the beginning of the `Enqueue` method (after the opening brace, before `ext := filepath.Ext(filename)`):

```go
	// Guest quota: lifetime 3 conversions max
	provider, err := s.repo.FindProviderByID(userID)
	if err != nil {
		return nil, fmt.Errorf("quota check: %w", err)
	}
	if provider == "guest" {
		count, err := s.repo.CountByUserID(userID)
		if err != nil {
			return nil, fmt.Errorf("quota count: %w", err)
		}
		if count >= 3 {
			return nil, fmt.Errorf("试用次数已用完（%d/3），请登录后继续使用", count)
		}
	}
```

- [ ] **Step 3: Update router.go**

As described above — wire EmailService, update NewAuthService call, add 3 routes.

- [ ] **Step 4: Verify compilation**

Run: `cd server && go build ./...`
Expected: no errors

- [ ] **Step 5: Commit**

```bash
git add server/internal/repo/conversion.go server/internal/service/conversion.go server/internal/router/router.go
git commit -m "feat: wire email/guest routes and add guest quota enforcement"
```

---

### Task 8: Update Docker and env configs

**Files:**
- Modify: `docker-compose.yml`
- Modify: `.env.example`

- [ ] **Step 1: Add SMTP env vars to docker-compose.yml**

In the `api` service environment section, add:

```yaml
      SMTP_HOST: ${SMTP_HOST:-}
      SMTP_PORT: ${SMTP_PORT:-587}
      SMTP_USER: ${SMTP_USER:-}
      SMTP_PASSWORD: ${SMTP_PASSWORD:-}
      SMTP_FROM: ${SMTP_FROM:-}
```

- [ ] **Step 2: Add SMTP vars to .env.example**

Append to `.env.example`:

```bash
# ===== SMTP (for email verification codes) =====
# SMTP_HOST=smtp.example.com
# SMTP_PORT=587
# SMTP_USER=your@email.com
# SMTP_PASSWORD=your_password
# SMTP_FROM=noreply@yourdomain.com
```

- [ ] **Step 3: Add SMTP vars to existing .env**

Append to `.env`:

```bash
# SMTP
SMTP_HOST=
SMTP_PORT=587
SMTP_USER=
SMTP_PASSWORD=
SMTP_FROM=
```

- [ ] **Step 4: Commit**

```bash
git add docker-compose.yml .env.example .env
git commit -m "feat: add SMTP configuration to docker and env files"
```

---

### Task 9: Update frontend API client

**Files:**
- Modify: `web/src/api/client.ts`

- [ ] **Step 1: Add new API methods and types**

Add to `web/src/api/client.ts`:

At the top, add `User` type (needed by AuthContext):

```typescript
export type User = {
  id: string;
  email: string;
  name: string;
  avatar_url: string;
  provider: string;
  provider_id: string;
  created_at: string;
  updated_at: string;
};
```

Add guest and email auth methods to the `auth` object:

```typescript
export const auth = {
  me: () => request<User>('/auth/me'),
  refresh: () => request<{ token: string }>('/auth/refresh', { method: 'POST' }),
  guest: () => request<{ token: string; user: User }>('/auth/guest', { method: 'POST' }),
  sendCode: (email: string) =>
    request<{ ok: boolean }>('/auth/email/send-code', {
      method: 'POST',
      body: JSON.stringify({ email }),
    }),
  verifyCode: (email: string, code: string) =>
    request<{ token: string; user: User }>('/auth/email/verify', {
      method: 'POST',
      body: JSON.stringify({ email, code }),
    }),
};
```

- [ ] **Step 2: Verify TypeScript compilation**

Run: `cd web && npx tsc --noEmit`
Expected: no errors (or only pre-existing errors unrelated to our changes)

- [ ] **Step 3: Commit**

```bash
git add web/src/api/client.ts
git commit -m "feat: add guest and email auth API methods and User type"
```

---

### Task 10: Create EmailLoginModal component

**Files:**
- Create: `web/src/components/EmailLoginModal.tsx`

- [ ] **Step 1: Write the component**

```tsx
import { useState, useRef, useEffect } from 'react';
import { auth, ApiError } from '../api/client';

interface Props {
  open: boolean;
  onClose: () => void;
  onSuccess: (token: string) => void;
}

export default function EmailLoginModal({ open, onClose, onSuccess }: Props) {
  const [step, setStep] = useState<'email' | 'code'>('email');
  const [email, setEmail] = useState('');
  const [code, setCode] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [countdown, setCountdown] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (open) {
      setStep('email');
      setEmail('');
      setCode('');
      setError('');
      setCountdown(0);
    }
  }, [open]);

  useEffect(() => {
    if (open && inputRef.current) {
      inputRef.current.focus();
    }
  }, [open, step]);

  useEffect(() => {
    if (countdown <= 0) return;
    const t = setTimeout(() => setCountdown(c => c - 1), 1000);
    return () => clearTimeout(t);
  }, [countdown]);

  const handleSendCode = async () => {
    if (!email.trim()) {
      setError('请输入邮箱地址');
      return;
    }
    setLoading(true);
    setError('');
    try {
      await auth.sendCode(email.trim());
      setStep('code');
      setCountdown(60);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : '发送失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  const handleVerify = async () => {
    if (!code.trim()) {
      setError('请输入验证码');
      return;
    }
    setLoading(true);
    setError('');
    try {
      const res = await auth.verifyCode(email.trim(), code.trim());
      localStorage.setItem('token', res.token);
      onSuccess(res.token);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : '验证失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40" onClick={onClose}>
      <div
        className="mx-4 w-full max-w-sm rounded-2xl bg-white p-6 shadow-xl"
        onClick={e => e.stopPropagation()}
      >
        <h2 className="text-lg font-bold text-gray-900">
          {step === 'email' ? '邮箱登录 / 注册' : '输入验证码'}
        </h2>
        <p className="mt-1 text-sm text-gray-500">
          {step === 'email'
            ? '输入邮箱，我们将发送验证码'
            : `验证码已发送至 ${email}`}
        </p>

        {error && (
          <div className="mt-3 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600">{error}</div>
        )}

        {step === 'email' ? (
          <div className="mt-4">
            <input
              ref={inputRef}
              type="email"
              value={email}
              onChange={e => setEmail(e.target.value)}
              onKeyDown={e => e.key === 'Enter' && handleSendCode()}
              placeholder="your@email.com"
              className="w-full rounded-xl border border-gray-200 px-4 py-3 text-sm outline-none focus:border-amber-400 focus:ring-2 focus:ring-amber-100"
            />
            <button
              onClick={handleSendCode}
              disabled={loading}
              className="mt-3 w-full rounded-xl bg-amber-500 py-3 text-sm font-bold text-gray-900 hover:bg-amber-600 disabled:opacity-50"
            >
              {loading ? '发送中...' : '发送验证码'}
            </button>
          </div>
        ) : (
          <div className="mt-4">
            <input
              ref={inputRef}
              type="text"
              maxLength={6}
              value={code}
              onChange={e => setCode(e.target.value.replace(/\D/g, ''))}
              onKeyDown={e => e.key === 'Enter' && handleVerify()}
              placeholder="输入 6 位验证码"
              className="w-full rounded-xl border border-gray-200 px-4 py-3 text-center text-2xl tracking-widest outline-none focus:border-amber-400 focus:ring-2 focus:ring-amber-100"
            />
            <button
              onClick={handleVerify}
              disabled={loading || code.length !== 6}
              className="mt-3 w-full rounded-xl bg-amber-500 py-3 text-sm font-bold text-gray-900 hover:bg-amber-600 disabled:opacity-50"
            >
              {loading ? '验证中...' : '验证登录'}
            </button>
            <button
              onClick={handleSendCode}
              disabled={countdown > 0 || loading}
              className="mt-2 w-full text-sm text-gray-400 hover:text-amber-500 disabled:text-gray-300"
            >
              {countdown > 0 ? `${countdown}s 后重新发送` : '重新发送验证码'}
            </button>
            <button
              onClick={() => { setStep('email'); setError(''); }}
              className="mt-1 w-full text-sm text-gray-400 hover:text-gray-600"
            >
              更换邮箱
            </button>
          </div>
        )}

        <button
          onClick={onClose}
          className="mt-4 w-full rounded-xl py-2 text-sm text-gray-400 hover:text-gray-600"
        >
          取消
        </button>
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Verify TypeScript compilation**

Run: `cd web && npx tsc --noEmit`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add web/src/components/EmailLoginModal.tsx
git commit -m "feat: add EmailLoginModal with 2-step code verification"
```

---

### Task 11: Create UsageLimitModal and GuestBanner components

**Files:**
- Create: `web/src/components/UsageLimitModal.tsx`
- Create: `web/src/components/GuestBanner.tsx`

- [ ] **Step 1: Write UsageLimitModal**

```tsx
interface Props {
  open: boolean;
  onLogin: () => void;
  onClose: () => void;
}

export default function UsageLimitModal({ open, onLogin, onClose }: Props) {
  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40" onClick={onClose}>
      <div
        className="mx-4 w-full max-w-sm rounded-2xl bg-white p-6 shadow-xl text-center"
        onClick={e => e.stopPropagation()}
      >
        <div className="mx-auto mb-3 flex h-14 w-14 items-center justify-center rounded-full bg-amber-100">
          <svg className="h-7 w-7 text-amber-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
              d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
          </svg>
        </div>
        <h2 className="text-lg font-bold text-gray-900">试用次数已用完</h2>
        <p className="mt-2 text-sm text-gray-500">
          免费试用限制 3 次转换。登录后可每天转换 20 次，解锁全部功能。
        </p>
        <button
          onClick={onLogin}
          className="mt-5 w-full rounded-xl bg-amber-500 py-3 text-sm font-bold text-gray-900 hover:bg-amber-600"
        >
          登录 / 注册继续使用
        </button>
        <button
          onClick={onClose}
          className="mt-2 w-full rounded-xl py-2 text-sm text-gray-400 hover:text-gray-600"
        >
          以后再说
        </button>
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Write GuestBanner**

```tsx
interface Props {
  remaining: number;
  onLogin: () => void;
}

export default function GuestBanner({ remaining, onLogin }: Props) {
  if (remaining < 0) return null;

  return (
    <div className="flex items-center justify-between rounded-xl bg-amber-50 px-4 py-2.5 text-sm">
      <span className="text-amber-700">
        试用模式 — 还剩 <strong>{remaining}</strong> 次免费转换
      </span>
      <button
        onClick={onLogin}
        className="rounded-lg bg-amber-500 px-3 py-1.5 text-xs font-bold text-gray-900 hover:bg-amber-600"
      >
        登录解锁更多
      </button>
    </div>
  );
}
```

- [ ] **Step 3: Verify TypeScript compilation**

Run: `cd web && npx tsc --noEmit`
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add web/src/components/UsageLimitModal.tsx web/src/components/GuestBanner.tsx
git commit -m "feat: add UsageLimitModal and GuestBanner components"
```

---

### Task 12: Update AuthContext for guest and email login support

**Files:**
- Modify: `web/src/context/AuthContext.tsx`

- [ ] **Step 1: Rewrite AuthContext**

```tsx
import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from 'react';
import { auth, ApiError, type User } from '../api/client';

interface AuthState {
  token: string | null;
  user: User | null;
  loading: boolean;
  isGuest: boolean;
  login: () => void;
  logout: () => void;
  guestLogin: () => Promise<void>;
  onAuthSuccess: (token: string) => void;
}

const AuthContext = createContext<AuthState | null>(null);

const GUEST_COUNT_KEY = 'guest_conversion_count';

export function getGuestRemaining(): number {
  const count = parseInt(localStorage.getItem(GUEST_COUNT_KEY) || '0', 10);
  return Math.max(0, 3 - count);
}

export function incrementGuestCount(): void {
  const count = parseInt(localStorage.getItem(GUEST_COUNT_KEY) || '0', 10);
  localStorage.setItem(GUEST_COUNT_KEY, String(count + 1));
}

export function resetGuestCount(): void {
  localStorage.removeItem(GUEST_COUNT_KEY);
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(() => localStorage.getItem('token'));
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  const isGuest = user?.provider === 'guest';

  useEffect(() => {
    if (!token) {
      setLoading(false);
      return;
    }
    auth.me()
      .then(data => setUser(data))
      .catch(err => {
        if (err instanceof ApiError && err.status === 401) {
          localStorage.removeItem('token');
          setToken(null);
        }
      })
      .finally(() => setLoading(false));
  }, [token]);

  const login = useCallback(() => {
    window.location.href = '/api/v1/auth/github/login';
  }, []);

  const logout = useCallback(() => {
    localStorage.removeItem('token');
    setToken(null);
    setUser(null);
  }, []);

  const onAuthSuccess = useCallback((newToken: string) => {
    localStorage.setItem('token', newToken);
    setToken(newToken);
    resetGuestCount();
  }, []);

  const guestLogin = useCallback(async () => {
    const res = await auth.guest();
    localStorage.setItem('token', res.token);
    setToken(res.token);
    setUser(res.user);
  }, []);

  return (
    <AuthContext.Provider value={{ token, user, loading, isGuest, login, logout, guestLogin, onAuthSuccess }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
```

Note: The `/auth/me` response format is `{ user: User }` but the handler returns the user directly (`c.JSON(http.StatusOK, user)`). Need to check the current me handler's response format and the current AuthContext's usage. Let's keep backward-compatible: try `data.user` first, fall back to `data` as User.

- [ ] **Step 2: Verify TypeScript**

Run: `cd web && npx tsc --noEmit`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add web/src/context/AuthContext.tsx
git commit -m "feat: add guest login, email success callback, and conversion tracking to AuthContext"
```

---

### Task 13: Update LandingPage with new layout

**Files:**
- Modify: `web/src/pages/LandingPage.tsx`

- [ ] **Step 1: Rewrite LandingPage**

Replace the Hero section button with the three stacked options:

```tsx
import { useState } from 'react';
import { useAuth, getGuestRemaining } from '../context/AuthContext';
import { Navigate } from 'react-router-dom';
import EmailLoginModal from '../components/EmailLoginModal';
import ToolCard from '../components/ToolCard';

// ... keep the `tools` array unchanged ...

export default function LandingPage() {
  const { token, loading, guestLogin, onAuthSuccess } = useAuth();
  const [emailModalOpen, setEmailModalOpen] = useState(false);

  if (loading) return null;
  if (token) return <Navigate to="/workspace/convert" replace />;

  return (
    <div className="min-h-screen bg-[#FFFDF7]">
      {/* Hero */}
      <section className="relative overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-br from-amber-50 via-white to-amber-50/30" />
        <div className="relative mx-auto max-w-6xl px-6 py-24 text-center sm:py-32">
          <h1 className="text-4xl font-extrabold tracking-tight text-gray-900 sm:text-5xl lg:text-6xl">
            创意设计资源，一站即达
          </h1>
          <p className="mx-auto mt-6 max-w-2xl text-lg text-gray-500 leading-relaxed">
            高质量的设计工具和资源平台，帮助你快速完成从位图到矢量、
            从灵感到交付的完整设计链路。
          </p>
          <div className="mt-10 mx-auto max-w-xs space-y-3">
            {/* Guest */}
            <button
              onClick={guestLogin}
              className="w-full inline-flex items-center justify-center gap-2 rounded-2xl bg-amber-500 px-6 py-3.5 text-base font-bold text-gray-900 shadow-md shadow-amber-200 transition-all duration-300 hover:-translate-y-0.5 hover:bg-amber-600 hover:shadow-lg hover:shadow-amber-300"
            >
              开始免费使用
              <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 8l4 4m0 0l-4 4m4-4H3" />
              </svg>
            </button>

            {/* Email */}
            <button
              onClick={() => setEmailModalOpen(true)}
              className="w-full inline-flex items-center justify-center gap-2 rounded-2xl border-2 border-gray-200 bg-white px-6 py-3.5 text-base font-semibold text-gray-700 transition-all duration-300 hover:-translate-y-0.5 hover:border-amber-300 hover:shadow-md"
            >
              <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                  d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
              </svg>
              邮箱登录 / 注册
            </button>

            {/* GitHub */}
            <a
              href="/api/v1/auth/github/login"
              className="block text-sm text-gray-400 hover:text-gray-600 transition-colors"
            >
              使用 GitHub 账号登录 →
            </a>
          </div>
        </div>
      </section>

      {/* Tools */}
      <section className="mx-auto max-w-6xl px-6 pb-24">
        <div className="mb-10 text-center">
          <h2 className="text-2xl font-extrabold text-gray-900 sm:text-3xl">我们的工具</h2>
          <p className="mt-3 text-gray-500">更多实用工具正在开发中，敬请期待</p>
        </div>
        <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
          {tools.map((tool) => (
            <ToolCard key={tool.title} {...tool} />
          ))}
        </div>
      </section>

      <EmailLoginModal
        open={emailModalOpen}
        onClose={() => setEmailModalOpen(false)}
        onSuccess={(token) => {
          setEmailModalOpen(false);
          onAuthSuccess(token);
        }}
      />
    </div>
  );
}
```

- [ ] **Step 2: Verify TypeScript**

Run: `cd web && npx tsc --noEmit`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add web/src/pages/LandingPage.tsx
git commit -m "feat: add guest/email/github login buttons to landing page"
```

---

### Task 14: Add guest banner and usage limit to ConvertPage

**Files:**
- Modify: `web/src/pages/ConvertPage.tsx`

- [ ] **Step 1: Rewrite ConvertPage with guest banner and usage limit**

Replace the entire `web/src/pages/ConvertPage.tsx`:

```tsx
import { useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import DropZone from '../components/DropZone';
import LoadingSpinner from '../components/LoadingSpinner';
import GuestBanner from '../components/GuestBanner';
import UsageLimitModal from '../components/UsageLimitModal';
import { conversions, ApiError } from '../api/client';
import { useAuth, getGuestRemaining, incrementGuestCount } from '../context/AuthContext';
import { usePolling } from '../hooks/usePolling';

export default function ConvertPage() {
  const navigate = useNavigate();
  const { isGuest } = useAuth();
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [conversionId, setConversionId] = useState<string | null>(null);
  const [status, setStatus] = useState<string | null>(null);
  const [remaining, setRemaining] = useState(getGuestRemaining);
  const [showLimitModal, setShowLimitModal] = useState(false);

  const pollStatus = useCallback(() => {
    if (!conversionId) return;
    conversions.get(conversionId)
      .then(res => {
        setStatus(res.data.status);
        if (res.data.status === 'completed') {
          navigate(`/workspace/preview/${conversionId}`, { replace: true });
        }
      })
      .catch(() => {});
  }, [conversionId, navigate]);

  usePolling(pollStatus, 1000, status === 'pending' || status === 'processing');

  const handleFile = useCallback(async (file: File) => {
    if (isGuest && getGuestRemaining() <= 0) {
      setShowLimitModal(true);
      return;
    }

    setError(null);
    setUploading(true);
    try {
      const res = await conversions.upload(file);
      setConversionId(res.data.id);
      setStatus(res.data.status);
      if (isGuest) {
        incrementGuestCount();
        setRemaining(getGuestRemaining());
      }
    } catch (err) {
      const msg = err instanceof ApiError ? err.message : 'Upload failed';
      setError(msg);
    } finally {
      setUploading(false);
    }
  }, [isGuest]);

  if (uploading) {
    return <LoadingSpinner label="Uploading..." />;
  }

  if (status === 'pending' || status === 'processing') {
    return (
      <div className="max-w-xl mx-auto">
        <h2 className="text-xl font-bold mb-4">Processing...</h2>
        <LoadingSpinner label={`Status: ${status}`} />
        <p className="text-center text-sm text-gray-500 mt-4">
          Your image is being converted to SVG. This may take a few seconds.
        </p>
      </div>
    );
  }

  return (
    <div className="max-w-xl mx-auto space-y-4">
      <h2 className="text-xl font-bold">New Conversion</h2>
      {isGuest && (
        <GuestBanner remaining={remaining} onLogin={() => setShowLimitModal(true)} />
      )}
      {error && (
        <div className="rounded-lg bg-red-50 border border-red-200 p-4 text-sm text-red-700">
          {error}
        </div>
      )}
      <DropZone onFile={handleFile} disabled={uploading} />
      <UsageLimitModal
        open={showLimitModal}
        onLogin={() => setShowLimitModal(false)}
        onClose={() => setShowLimitModal(false)}
      />
    </div>
  );
}
```

- [ ] **Step 2: Verify TypeScript**

Run: `cd web && npx tsc --noEmit`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add web/src/pages/ConvertPage.tsx
git commit -m "feat: integrate guest banner and usage limit into ConvertPage"
```

---

### Task 15: Add verification code cleanup goroutine to main.go

**Files:**
- Modify: `server/cmd/api/main.go`

- [ ] **Step 1: Start cleanup goroutine in main**

After the "server starting" log line, add:

```go
// Start verification code cleanup goroutine (every 5 min)
go func() {
	for {
		time.Sleep(5 * time.Minute)
		userRepo.CleanupExpiredCodes()
	}
}()
```

Need to import `"time"` and have `userRepo` accessible. The current main.go likely creates the repo in router.Setup(). We need to make userRepo accessible or add CleanupCodes to a service.

Simpler approach: add a `CleanupExpiredCodes` callable from the router Setup. Actually, let's just do the cleanup in the email service and inject it into the AuthService:

Actually the simplest: add to the route setup, after creating userRepo, start a goroutine:

In `router.go`:

```go
// Start verification code cleanup goroutine
go func() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    for range ticker.C {
        userRepo.CleanupExpiredCodes()
    }
}()
```

- [ ] **Step 1 (revised): Add cleanup goroutine to router.go**

In `server/internal/router/router.go`, after creating `userRepo`:

```go
// Start cleanup goroutine for expired verification codes
go func() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		userRepo.CleanupExpiredCodes()
	}
}()
```

Add `"time"` to imports.

- [ ] **Step 2: Verify compilation**

Run: `cd server && go build ./...`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add server/internal/router/router.go
git commit -m "feat: add verification code cleanup goroutine"
```

---

### Task 16: Final integration verification

**Files:** all modified files

- [ ] **Step 1: Run full Go build**

```bash
cd server && go build ./...
```

- [ ] **Step 2: Run full TypeScript check**

```bash
cd web && npx tsc --noEmit
```

- [ ] **Step 3: Final commit if any fixups**

```bash
git add -A
git commit -m "chore: final integration fixes for email auth and guest mode"
```

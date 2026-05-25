# Email Auth + Guest Mode Design

**Date**: 2026-05-25
**Status**: approved

## Summary

Add three login paths on the landing page: guest mode (3 free conversions then lock), email verification-code login/register via SMTP, and existing GitHub OAuth. Guest users go through the same JWT auth pipeline as regular users, with provider="guest".

## User Flow

```
LandingPage
  ├── "开始免费使用" → POST /api/v1/auth/guest
  │     → Create guest user (provider="guest"), issue JWT (24h)
  │     → Set guest_id cookie + fingerprint header
  │     → 3 conversions allowed, then 429 → UsageLimitModal
  │
  ├── "邮箱登录/注册" → EmailLoginModal
  │     → Step 1: enter email → POST /api/v1/auth/email/send-code
  │     → SMTP sends 6-digit code (5min TTL)
  │     → Step 2: enter code → POST /api/v1/auth/email/verify
  │     → New email auto-registers, issue JWT (7d)
  │
  └── "GitHub 登录 →" → existing GitHub OAuth flow
```

## Database Changes

### New table: `verification_codes`

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

### Migration: `004_email_auth.up.sql`

### Migration: `003_create_quotas.up.sql` — no changes needed (guest users have user_id, quota works as-is)

### Migration: `005_guest_provider.up.sql` — relax UNIQUE(provider, provider_id)

```sql
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_provider_provider_id_key;
CREATE UNIQUE INDEX users_provider_unique ON users(provider, provider_id)
    WHERE provider != 'guest';
```

## API Endpoints

### POST `/api/v1/auth/guest`

Create or restore a guest user. Accepts optional `guest_id` cookie and `X-Fingerprint` header for session continuity.

- Request: empty body (optional cookie/fingerprint)
- Response: `{ "token": "jwt...", "user": { ... } }`
- Sets: `guest_id` cookie (httpOnly, long-lived)

Logic:
1. If valid `guest_id` cookie present → find existing guest user, return JWT
2. If `X-Fingerprint` header present → try to find guest by fingerprint
3. Otherwise → create new guest user with random provider_id

### POST `/api/v1/auth/email/send-code`

Send 6-digit verification code to email via SMTP.

- Request: `{ "email": "user@example.com" }`
- Response: `{ "ok": true }`
- Rate limit: 1 per email per 60 seconds
- Code TTL: 5 minutes

### POST `/api/v1/auth/email/verify`

Verify the code and login/register.

- Request: `{ "email": "user@example.com", "code": "123456" }`
- Response: `{ "token": "jwt...", "user": { ... } }`
- If email is new → auto-create user (provider="email")
- If email exists → login as existing user

## Config Changes (server/internal/config/config.go)

New SMTP fields:

```go
SMTPHost     string
SMTPPort     int
SMTPUser     string
SMTPPassword string
SMTPFrom     string
```

New env vars: `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASSWORD`, `SMTP_FROM`

## Backend File Changes

| File | Change |
|------|--------|
| `server/internal/config/config.go` | Add SMTP fields |
| `server/internal/model/user.go` | No change (guest uses existing User struct) |
| `server/internal/repo/user.go` | Add `FindByGuestID`, `FindByEmail`, `CreateGuest` methods |
| `server/internal/service/auth.go` | Add `EmailSendCode`, `EmailVerify`, `GuestLogin` methods |
| `server/internal/handler/auth.go` | Add `EmailSendCode`, `EmailVerify`, `GuestLogin` handlers |
| `server/internal/router/router.go` | Add new routes |
| `server/internal/service/email.go` | **New**: SMTP email sending service |
| `server/migrations/004_verification_codes.up.sql` | **New**: verification_codes table |
| `server/migrations/005_guest_provider.up.sql` | **New**: relax unique constraint |

## Frontend File Changes

| File | Change |
|------|--------|
| `web/src/pages/LandingPage.tsx` | New layout: 3 buttons stacked (guest / email / GitHub) |
| `web/src/components/EmailLoginModal.tsx` | **New**: 2-step modal (email input → code input) |
| `web/src/components/UsageLimitModal.tsx` | **New**: quota exhausted popup |
| `web/src/components/GuestBanner.tsx` | **New**: remaining free conversions indicator |
| `web/src/context/AuthContext.tsx` | Support guest token + remaining quota tracking |
| `web/src/api/client.ts` | Add email auth + guest API methods |

## Guest Quota Enforcement

Two-layer enforcement:

**Frontend (primary UX):** Track count in localStorage (`guest_conversion_count`). Increment after each upload. When count >= 3, show UsageLimitModal and block further uploads.

**Backend (hard cap):** In `ConversionService.Enqueue`, when user.provider == "guest", query `SELECT COUNT(*) FROM conversions WHERE user_id=$1` — if total >= 3, return error (code: `QUOTA_EXHAUSTED`). This prevents circumvention via curl/API direct calls.

Registered users (email/GitHub) use the existing daily quota (20/day), unaffected by guest limits.

## Verification Code Cleanup

Worker or cron that deletes expired verification codes. For simplicity, a goroutine in the API server runs every 5 minutes to clean up codes older than 1 hour.

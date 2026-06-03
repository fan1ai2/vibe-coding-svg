# 邮箱认证 + 游客模式 — 设计文档

**日期**：2026-05-25
**状态**：已批准

## 概述

在首页添加三种登录方式：游客模式（3 次免费转换后锁定）、通过 SMTP 的邮箱验证码登录/注册、以及现有的 GitHub OAuth。游客用户与注册用户使用相同的 JWT 认证管线，provider="guest"。

## 用户流程

```
LandingPage
  ├── "开始免费使用" → POST /api/v1/auth/guest
  │     → 创建游客用户 (provider="guest")，签发 JWT（24h）
  │     → 设置 guest_id cookie + fingerprint 头
  │     → 允许 3 次转换，超出后 429 → UsageLimitModal
  │
  ├── "邮箱登录/注册" → EmailLoginModal
  │     → 步骤 1：输入邮箱 → POST /api/v1/auth/email/send-code
  │     → SMTP 发送 6 位验证码（5 分钟有效）
  │     → 步骤 2：输入验证码 → POST /api/v1/auth/email/verify
  │     → 新邮箱自动注册，签发 JWT（7 天）
  │
  └── "GitHub 登录 →" → 现有 GitHub OAuth 流程
```

## 数据库变更

### 新表：`verification_codes`

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

### 迁移文件：`004_email_auth.up.sql`

### 迁移文件：`003_create_quotas.up.sql` —— 无需修改（游客用户有 user_id，配额机制同样适用）

### 迁移文件：`005_guest_provider.up.sql` —— 放宽 UNIQUE(provider, provider_id) 约束

```sql
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_provider_provider_id_key;
CREATE UNIQUE INDEX users_provider_unique ON users(provider, provider_id)
    WHERE provider != 'guest';
```

## API 接口

### POST `/api/v1/auth/guest`

创建或恢复游客用户。接受可选的 `guest_id` cookie 和 `X-Fingerprint` 请求头以保持会话连续性。

- 请求：空 body（可选 cookie/fingerprint）
- 响应：`{ "token": "jwt...", "user": { ... } }`
- 设置：`guest_id` cookie（httpOnly，长期有效）

逻辑：
1. 如果存在有效的 `guest_id` cookie → 找到已有游客用户，返回 JWT
2. 如果存在 `X-Fingerprint` 请求头 → 尝试通过指纹查找游客
3. 否则 → 创建新的游客用户，使用随机 provider_id

### POST `/api/v1/auth/email/send-code`

通过 SMTP 向邮箱发送 6 位验证码。

- 请求：`{ "email": "user@example.com" }`
- 响应：`{ "ok": true }`
- 限流：每个邮箱每 60 秒 1 次
- 验证码有效期：5 分钟

### POST `/api/v1/auth/email/verify`

验证验证码并登录/注册。

- 请求：`{ "email": "user@example.com", "code": "123456" }`
- 响应：`{ "token": "jwt...", "user": { ... } }`
- 如果邮箱是新的 → 自动创建用户（provider="email"）
- 如果邮箱已存在 → 作为已有用户登录

## 配置变更（server/internal/config/config.go）

新增 SMTP 字段：

```go
SMTPHost     string
SMTPPort     int
SMTPUser     string
SMTPPassword string
SMTPFrom     string
```

新增环境变量：`SMTP_HOST`、`SMTP_PORT`、`SMTP_USER`、`SMTP_PASSWORD`、`SMTP_FROM`

## 后端文件变更

| 文件 | 变更 |
|------|--------|
| `server/internal/config/config.go` | 添加 SMTP 字段 |
| `server/internal/model/user.go` | 无变化（游客复用现有 User 结构体） |
| `server/internal/repo/user.go` | 添加 `FindByGuestID`、`FindByEmail`、`CreateGuest` 方法 |
| `server/internal/service/auth.go` | 添加 `EmailSendCode`、`EmailVerify`、`GuestLogin` 方法 |
| `server/internal/handler/auth.go` | 添加 `EmailSendCode`、`EmailVerify`、`GuestLogin` 处理器 |
| `server/internal/router/router.go` | 添加新路由 |
| `server/internal/service/email.go` | **新建**：SMTP 邮件发送服务 |
| `server/migrations/004_verification_codes.up.sql` | **新建**：verification_codes 表 |
| `server/migrations/005_guest_provider.up.sql` | **新建**：放宽唯一约束 |

## 前端文件变更

| 文件 | 变更 |
|------|--------|
| `web/src/pages/LandingPage.tsx` | 全新布局：3 个按钮纵向排列（游客 / 邮箱 / GitHub） |
| `web/src/components/EmailLoginModal.tsx` | **新建**：两步式弹窗（输入邮箱 → 输入验证码） |
| `web/src/components/UsageLimitModal.tsx` | **新建**：配额耗尽弹窗 |
| `web/src/components/GuestBanner.tsx` | **新建**：剩余免费转换次数提示 |
| `web/src/context/AuthContext.tsx` | 支持游客 token + 剩余配额跟踪 |
| `web/src/api/client.ts` | 添加邮箱认证 + 游客 API 方法 |

## 游客配额控制

双层控制：

**前端（主要 UX）：** 在 localStorage 中跟踪计数（`guest_conversion_count`）。每次上传后递增。当 count >= 3 时，显示 UsageLimitModal 并阻止进一步上传。

**后端（硬上限）：** 在 `ConversionService.Enqueue` 中，当 user.provider == "guest" 时，查询 `SELECT COUNT(*) FROM conversions WHERE user_id=$1` —— 如果总数 >= 3，返回错误（code: `QUOTA_EXHAUSTED`）。这防止通过 curl/API 直接调用绕过限制。

注册用户（邮箱/GitHub）使用现有的每日配额（20 次/天），不受游客限制影响。

## 验证码清理

Worker 或定时任务删除过期的验证码。为简单起见，API 服务器中的一个 goroutine 每 5 分钟运行一次，清理超过 1 小时的验证码。

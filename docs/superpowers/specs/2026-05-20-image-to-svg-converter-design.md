# 位图转 SVG 转换器 — 设计规格

**日期：** 2026-05-20
**状态：** 已批准

## 产品概述

基于 Web 的图片转 SVG 转换工具。用户上传位图（PNG/JPG/WebP），系统使用传统矢量化引擎将其转换为 SVG 矢量图形，并将结果存储到个人 Library 中。免费工具，MVP 阶段无交易市场。

## 技术栈

- **后端：** Go（Gin 路由、JWT 认证、基于 goroutine 的 Worker）
- **前端：** React SPA（Vite、TypeScript、Tailwind CSS）
- **数据库：** PostgreSQL
- **缓存/队列：** Redis
- **存储：** MinIO（兼容 S3）
- **矢量化引擎：** potrace / autotrace（在 Worker 中以 CLI 方式调用）

## 架构

前后端分离。Go API 服务器（:8080）提供 RESTful JSON 接口。生产环境 React SPA 通过 Nginx/CDN 部署，开发环境使用 Vite 开发服务器。Go Worker 作为独立进程运行，从 Redis 队列消费任务。

```
浏览器 → React SPA → Go API (:8080) → PostgreSQL
                                     → Redis（任务队列）
                                     → MinIO（文件存储）
                Go Worker → Redis（消费）→ MinIO → 矢量化处理 → MinIO
```

### 核心原则

- 无状态 API —— 所有状态存储在 PostgreSQL/Redis 中，API 可水平扩展
- 异步优先 —— 上传立即返回 task_id，Worker 后台处理
- 外部文件存储 —— 所有文件存储在 MinIO 中，服务器磁盘为临时存储

## 核心工作流

1. **上传** —— POST multipart/form-data 到 `/api/v1/conversions`，返回 `{task_id, status: "pending"}`
2. **处理中** —— Worker 从 Redis 队列获取任务，从 MinIO 下载原始文件，执行矢量化，上传 SVG + 缩略图到 MinIO，更新数据库状态为 `completed`
3. **预览** —— 前端每秒轮询 `GET /api/v1/conversions/:id`。完成后显示原图与 SVG 的并排对比（支持缩放/平移）、元数据（路径数、颜色数、体积缩减）
4. **下载** —— `GET /api/v1/conversions/:id/download` 流式返回 SVG 文件

## 数据库 Schema

### users
| 列名 | 类型 | 备注 |
|--------|------|-------|
| id | UUID PK | gen_random_uuid() |
| email | VARCHAR(255) UNIQUE | 来自 OAuth |
| name | VARCHAR(100) | 显示名称 |
| avatar_url | VARCHAR(500) | 来自 OAuth |
| provider | VARCHAR(20) NOT NULL | 'github' 或 'google' |
| provider_id | VARCHAR(100) NOT NULL | OAuth 唯一标识 |
| created_at | TIMESTAMPTZ | DEFAULT now() |

UNIQUE(provider, provider_id)

### conversions
| 列名 | 类型 | 备注 |
|--------|------|-------|
| id | UUID PK | |
| user_id | UUID FK → users | |
| status | VARCHAR(20) | pending/processing/completed/failed |
| original_url | VARCHAR(500) | MinIO 路径 |
| svg_url | VARCHAR(500) | MinIO 路径 |
| thumbnail_url | VARCHAR(500) | MinIO 路径 |
| file_size_in | BIGINT | 字节 |
| file_size_out | BIGINT | 字节 |
| path_count | INT | SVG path 元素数量 |
| color_count | INT | 唯一颜色数 |
| format_in | VARCHAR(10) | png/jpg/webp |
| error_message | TEXT | 失败原因 |
| created_at | TIMESTAMPTZ | DEFAULT now() |
| completed_at | TIMESTAMPTZ | |

INDEX (user_id, status), INDEX (created_at DESC)

### daily_quotas
| 列名 | 类型 | 备注 |
|--------|------|-------|
| id | UUID PK | |
| user_id | UUID FK → users | |
| date | DATE | DEFAULT CURRENT_DATE |
| count | INT | DEFAULT 0 |

UNIQUE(user_id, date)

## API 接口

### 认证
- `GET /api/v1/auth/github/login` —— 重定向至 GitHub OAuth
- `GET /api/v1/auth/github/callback` —— 处理回调，设置 JWT cookie
- `GET /api/v1/auth/google/login` —— 重定向至 Google OAuth
- `GET /api/v1/auth/google/callback` —— 处理回调，设置 JWT cookie
- `POST /api/v1/auth/refresh` —— 刷新 JWT
- `GET /api/v1/auth/me` —— 当前用户信息

### 转换接口（需 JWT）
- `POST /api/v1/conversions` —— 上传图片，开始转换
- `GET /api/v1/conversions` —— 分页查询用户转换列表（?page=&limit=&status=）
- `GET /api/v1/conversions/:id` —— 获取单条转换状态/元数据
- `GET /api/v1/conversions/:id/download` —— 下载 SVG 文件
- `DELETE /api/v1/conversions/:id` —— 软删除

### 配额接口（需 JWT）
- `GET /api/v1/quotas/daily` —— 当日用量及限额

### 中间件链
请求 → Logger → CORS → RateLimit（100 次/分钟） → JWT 认证 → Handler

## 前端路由

| 路由 | 页面 | 认证 |
|-------|------|------|
| `/` | LandingPage（Hero 区、功能展示、CTA） | 否 |
| `/callback` | OAuthCallback（处理重定向） | 否 |
| `/workspace/convert` | ConvertPage（拖拽区、上传、处理中） | 是 |
| `/workspace/preview/:id` | PreviewPage（对比查看、下载） | 是 |
| `/workspace/library` | LibraryPage（历史记录网格、筛选） | 是 |

### 核心组件
- **DropZone** —— 拖拽 + 点击上传，文件类型/大小校验
- **ComparisonView** —— 原图与 SVG 渲染并排对比
- **ZoomControls** —— 放大 / 缩小 / 100% / 适应屏幕
- **MetadataCard** —— 路径数、颜色数、体积缩减百分比
- **ConversionCard** —— 缩略图、状态徽章、日期、点击预览

## 项目结构

```
vibe-coding-svg/
├── server/
│   ├── cmd/api/main.go
│   ├── cmd/worker/main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── handler/        # auth.go, conversion.go, quota.go
│   │   ├── middleware/      # jwt, cors, ratelimit, logger
│   │   ├── model/           # user, conversion, quota
│   │   ├── repo/            # PostgreSQL 数据访问
│   │   ├── service/         # 业务逻辑, 存储 (MinIO)
│   │   ├── worker/          # queue.go, vectorize.go
│   │   └── router/
│   └── migrations/
├── web/
│   ├── src/
│   │   ├── api/             # axios 封装
│   │   ├── components/      # 共享组件
│   │   ├── pages/           # 页面组件
│   │   ├── hooks/           # useAuth, useConversion, usePolling
│   │   ├── context/         # AuthContext
│   │   └── App.tsx
│   ├── vite.config.ts
│   └── tailwind.config.js
├── docker-compose.yml       # postgres + redis + minio
├── Dockerfile.api
└── Dockerfile.worker
```

## 错误处理

- 上传失败：文件过大（>10MB）、格式无效 → 400，附带用户友好提示
- 转换失败：引擎崩溃、超时 → status=failed，记录 error_message，用户可重试
- 认证失败：token 过期 → 401，重定向到登录页
- 限流：429，附带 Retry-After 头
- 所有错误返回 `{error: {code, message}}` JSON 格式

## 测试策略

- Go：service/repo 层单元测试，API handler 集成测试
- React：组件测试（Vitest），API Mock（MSW）
- E2E：可选，可对关键路径（上传 → 预览 → 下载）添加 Playwright 测试

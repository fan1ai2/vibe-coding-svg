# SVG 资源工坊

位图转 SVG 矢量图在线转换平台。上传 PNG/JPEG，一键生成 SVG。

## 系统架构

```
浏览器 ──▶ API (:8080) ──▶ PostgreSQL    (用户、转换记录、配额)
                 │                   
                 │  入队任务     
                 ▼                   
              Redis ──────▶ Worker ──▶ vtracer CLI
                                  │
                                  ▼
                               MinIO    (原始文件 + SVG 结果)
```

五个服务：**PostgreSQL**（数据存储）、**Redis**（任务队列）、**MinIO**（对象存储）、**API**（HTTP 服务）、**Worker**（转换处理）。API 和 Worker 是独立进程，通过 Redis 通信。

## 快速启动

```bash
git clone https://github.com/fan1ai2/vibe-coding-svg.git
cd vibe-coding-svg

# 配置 OAuth 密钥（开发环境可选）
export GITHUB_CLIENT_ID=xxx
export GITHUB_CLIENT_SECRET=xxx
export GOOGLE_CLIENT_ID=xxx
export GOOGLE_CLIENT_SECRET=xxx

docker-compose up -d
```

启动后数据库迁移自动执行。访问 `http://localhost:8080/health` 验证服务状态。

## API 接口

### 认证

| 方法 | 路径 | 认证 | 说明 |
|--------|------|------|-------------|
| GET | `/api/v1/auth/github/login` | - | 跳转 GitHub OAuth 授权 |
| GET | `/api/v1/auth/github/callback` | - | GitHub OAuth 回调 |
| GET | `/api/v1/auth/google/login` | - | 跳转 Google OAuth 授权 |
| GET | `/api/v1/auth/google/callback` | - | Google OAuth 回调 |
| POST | `/api/v1/auth/refresh` | JWT | 刷新令牌 |
| GET | `/api/v1/auth/me` | JWT | 获取当前用户完整信息 |

### 转换

| 方法 | 路径 | 认证 | 说明 |
|--------|------|------|-------------|
| POST | `/api/v1/conversions` | JWT | 上传图片（multipart 表单） |
| GET | `/api/v1/conversions` | JWT | 分页查询转换列表 |
| GET | `/api/v1/conversions/:id` | JWT | 查询单条转换状态 |
| GET | `/api/v1/conversions/:id/download` | JWT | 下载 SVG 结果文件 |

### 健康检查

| 方法 | 路径 | 说明 |
|--------|------|-------------|
| GET | `/health` | 数据库/Redis/MinIO 连通性检查 |

### Swagger 文档

| 方法 | 路径 | 说明 |
|--------|------|-------------|
| GET | `/docs/swagger.json` | OpenAPI 规范文档（可导入 Apifox/Postman） |

### 调用示例

```bash
# 浏览器登录
open http://localhost:8080/api/v1/auth/github/login

# 登录后从回调地址获取 JWT token
TOKEN="eyJ..."

# 上传图片
curl -X POST http://localhost:8080/api/v1/conversions \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@photo.png"

# 查询状态
curl http://localhost:8080/api/v1/conversions/<id> \
  -H "Authorization: Bearer $TOKEN"

# 下载结果
curl http://localhost:8080/api/v1/conversions/<id>/download \
  -H "Authorization: Bearer $TOKEN" \
  -o result.svg
```

转换状态流转：`pending` → `processing` → `completed`（或 `failed`）。

## 配置项

所有配置通过环境变量注入：

| 变量 | 默认值 | 说明 |
|----------|---------|-------------|
| `PORT` | `8080` | API 服务端口 |
| `DATABASE_URL` | - | PostgreSQL 连接串 |
| `REDIS_ADDR` | - | Redis 地址 |
| `MINIO_ENDPOINT` | `localhost:9000` | MinIO 端点 |
| `MINIO_ACCESS_KEY` | - | MinIO 访问密钥 |
| `MINIO_SECRET_KEY` | - | MinIO 密钥 |
| `JWT_SECRET` | - | JWT 签名密钥 |
| `GITHUB_CLIENT_ID` | - | GitHub OAuth 客户端 ID |
| `GITHUB_CLIENT_SECRET` | - | GitHub OAuth 密钥 |
| `MAX_FILE_SIZE` | `10485760` | 上传文件大小上限（字节，默认 10MB） |
| `FRONTEND_URL` | - | 前端地址（OAuth 回调跳转 + CORS 白名单） |

## 本地开发

```bash
# 启动基础设施
docker-compose up -d postgres redis minio

# 启动 API
cd server
go run ./cmd/api/

# 启动 Worker（另开终端）
go run ./cmd/worker/
```

Worker 依赖 **vtracer** 已安装到 `$PATH`：

```bash
# macOS
brew install vtracer

# Linux
cargo install vtracer
```

## 项目结构

```
server/
├── cmd/                              # 可执行程序入口
│   ├── api/main.go                   # API 服务：加载配置 → 连接数据库 → 执行迁移 → 注册路由 → 启动 HTTP 服务
│   └── worker/main.go                # Worker 服务：连接数据库和 MinIO → 注册 asynq 任务处理器 → 消费转换任务
├── internal/                         # 内部实现（不对外暴露）
│   ├── config/config.go              # 环境变量配置加载：数据库、Redis、MinIO、JWT、OAuth 等所有配置项
│   ├── model/                        # 数据结构定义（对应数据库表）
│   │   ├── user.go                   # 用户模型：ID、邮箱、昵称、头像、OAuth 提供商信息
│   │   ├── conversion.go             # 转换任务模型：状态、原始文件、SVG 结果、文件大小、路径数、错误信息
│   │   └── quota.go                  # 每日配额模型：用户 ID + 日期 + 当日已用次数
│   ├── repo/                         # 数据访问层（纯 SQL 操作，不包含业务逻辑）
│   │   ├── user.go                   # 用户查询：按 provider 查找 / 创建新用户 / 按 ID 查找
│   │   └── conversion.go             # 转换记录增删改查：创建 / 按 ID 查 / 按用户分页查 / 更新状态 / 写入结果 / 配额增减
│   ├── service/                      # 业务逻辑层（编排 repo + 外部服务）
│   │   ├── auth.go                   # 认证服务：JWT 签发（7 天有效期）、GitHub OAuth 授权码换 token → 获取用户信息 → 自动注册/登录、按 ID 查询用户
│   │   ├── storage.go                # MinIO 对象存储封装：创建桶、上传文件、下载文件、生成预签名 URL
│   │   └── conversion.go             # 转换服务：上传原始文件到 MinIO → 创建任务记录 → 推送 asynq 队列 → 每日配额检查（20 次/天）
│   ├── handler/                      # HTTP 请求处理器（参数校验 + 调用 service + 返回响应）
│   │   ├── auth.go                   # 认证接口：GitHub 登录跳转、OAuth 回调（含 state 防 CSRF）、刷新 token、获取当前用户（查数据库返回完整信息）
│   │   ├── conversion.go             # 转换接口：上传文件（校验类型/大小）→ 入队、分页列表、状态查询、SVG 下载
│   │   ├── health.go                 # 健康检查：Ping 数据库/Redis/MinIO，返回 ok 或 unhealthy
│   │   └── file.go                   # 文件代理：通过 UUID key 提供文件下载，URL 不可猜测因此无需鉴权
│   ├── middleware/                    # Gin 中间件（请求前置处理）
│   │   ├── jwt.go                    # JWT 鉴权：从 Authorization 头提取 Bearer token → 解析验证 → 注入 user_id 到上下文
│   │   ├── cors.go                   # 跨域配置：允许前端域名、Authorization 头、Cookie 携带
│   │   ├── logging.go                # 请求日志：记录每个请求的方法、路径、状态码、耗时
│   │   └── ratelimit.go              # 全局限流：基于 IP + Redis 滑动窗口，每分钟 100 次，Redis 不可用时自动放行
│   ├── router/router.go              # 路由注册 + 依赖注入：组装所有 handler/service/repo，定义公开/鉴权路由分组，注册 Swagger 文档路由
│   ├── worker/                       # 后台任务处理（asynq 消费者）
│   │   ├── converter.go              # vtracer CLI 封装：调用 vtracer 命令行工具将位图转为 SVG，统计 SVG 路径数
│   │   └── worker.go                 # 任务处理器：从 MinIO 下载原始文件 → 写临时文件 → 调用 vtracer → 上传 SVG 结果 → 更新数据库
│   └── migrate/migrate.go            # 数据库自动迁移：读取 migrations 目录下的 .up.sql 文件，按序执行，记录已执行版本
├── migrations/                       # SQL 迁移脚本（按编号顺序执行）
│   ├── 001_create_users.up.sql       # 创建 users 表：id、email、name、avatar_url、provider、provider_id
│   ├── 002_create_conversions.up.sql # 创建 conversions 表：转换状态、文件路径、大小、格式、错误信息
│   └── 003_create_quotas.up.sql      # 创建 daily_quotas 表：user_id + date 唯一约束，记录每日转换次数
├── Dockerfile.api                    # API 服务 Docker 镜像构建（含 Swagger 文档生成与打包）
├── Dockerfile.worker                 # Worker 服务 Docker 镜像构建
├── go.mod                            # Go 模块依赖声明
└── go.sum                            # Go 依赖校验和

web/                                 # React 前端（Vite + TypeScript）
├── src/
│   ├── api/client.ts                 # API 客户端：封装 fetch 请求，自动附加 JWT token，统一错误处理
│   ├── context/AuthContext.tsx        # 认证上下文：token 管理、完整用户信息存储、登录跳转、登出清除、启动时验证 token 有效性
│   ├── hooks/usePolling.ts           # 轮询 Hook：按固定间隔执行回调，用于实时刷新转换状态
│   ├── components/                   # 可复用 UI 组件
│   │   ├── Navbar.tsx                # 顶部导航栏：Logo + 搜索框占位 + 已登录时显示导航链接和用户头像菜单 + 未登录时显示 GitHub 登录按钮
│   │   ├── Footer.tsx                # 页脚：版权信息 + 快速链接
│   │   ├── WorkspaceShell.tsx        # 工作区布局：Navbar + 内容区（React Router Outlet）
│   │   ├── ToolCard.tsx              # 工具卡片：可用状态（可点击链接）和占位状态（即将推出）两种模式
│   │   ├── ConversionCard.tsx        # 转换卡片：展示单条转换记录的状态、缩略图、文件大小、路径数
│   │   ├── DropZone.tsx              # 拖拽上传区：支持点击选择或拖拽文件，预览图片后触发上传
│   │   ├── ErrorBoundary.tsx         # 错误边界：捕获子组件渲染异常，显示友好错误提示
│   │   └── LoadingSpinner.tsx        # 加载动画组件
│   ├── pages/                        # 页面组件
│   │   ├── LandingPage.tsx           # 平台首页：Hero 区 + 工具卡片网格（SVG 转换 可用 + 3 个占位卡片）
│   │   ├── CallbackPage.tsx          # OAuth 回调页：从 URL 提取 token → 存 localStorage → 跳转工作区
│   │   ├── ConvertPage.tsx           # 转换页：文件上传 + 转换参数
│   │   ├── PreviewPage.tsx           # 预览页：查看单条转换结果的 SVG 详情
│   │   └── LibraryPage.tsx           # 转换库：历史记录列表，支持分页浏览
│   ├── App.tsx                       # 根组件：路由定义（首页 + 回调 + 工作区子路由）+ 受保护路由（未登录重定向到首页）
│   ├── main.tsx                      # 应用入口：挂载 React 到 DOM
│   └── index.css                     # 全局样式
├── index.html                        # HTML 入口：Google Fonts（Nunito）+ 中文页面设置
├── tailwind.config.js                # Tailwind 配置：Nunito 字体扩展
├── Dockerfile.web                    # 前端 Docker 镜像构建
└── package.json                      # 前端依赖

Caddyfile                             # Caddy 反向代理配置（将 :80/:443 转发到 api:8080 和 web:5173）
docker-compose.yml                    # Docker Compose 编排：PostgreSQL + Redis + MinIO + API + Worker + Web + Caddy
```

## 前端开发

```bash
cd web
npm install
npm run dev          # 启动 Vite 开发服务器，监听 0.0.0.0:5173
```

前端 Vite 配置了 `/api` 代理到 `http://127.0.0.1:8080`，本地开发时无需额外配置跨域。

## 技术栈

| 层 | 技术 |
|------|------|
| 后端语言 | Go 1.25 |
| HTTP 框架 | Gin |
| 任务队列 | Asynq (Redis) |
| 对象存储 | MinIO (S3 兼容) |
| 数据库 | PostgreSQL 16 |
| 转换引擎 | vtracer (Rust) |
| 前端框架 | React 19 + TypeScript 5 |
| 构建工具 | Vite 6 |
| CSS 框架 | Tailwind CSS 3 |
| 路由 | React Router 7 |
| 反向代理 | Caddy 2 |
| 容器化 | Docker Compose |

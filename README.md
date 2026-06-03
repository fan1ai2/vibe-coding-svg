# SVG 资源工坊

位图转 SVG 矢量图、AI 图标生成、SVG 在线编辑、图标库管理与图谱搜索的一站式平台。

## 功能模块

| 模块 | 说明 |
|------|------|
| **位图转 SVG** | 上传 PNG/JPEG，通过 vtracer 引擎一键转为 SVG 矢量图 |
| **AI 图标生成** | 通过 OpenAI 兼容 API（支持 DeepSeek 等）用自然语言生成 SVG 图标 |
| **SVG 编辑器** | 在线编辑 SVG 颜色，支持粘贴/拖拽导入，元素选择和主题色替换 |
| **图标库** | Neo4j 图谱驱动的图标搜索，支持 CRUD、标签分类、相似推荐 |
| **认证系统** | 邮箱验证码登录、GitHub OAuth、访客一键登录、JWT 鉴权 |

## 系统架构

```
浏览器 ──▶ Caddy (:80/:443) ──▶ API (:8080) ──▶ PostgreSQL    (用户、转换、图标、配额)
                    │                    │
                    │                    ├──▶ Redis          (任务队列、限流、AI 配额)
                    │                    │
                    │                    ├──▶ MinIO           (原始文件 + SVG 结果)
                    │                    │
                    │                    ├──▶ Neo4j           (图标图谱搜索)
                    │                    │
                    │                    └──▶ AI API          (OpenAI 兼容端点)
                    │
                    └──▶ Web (Vite 开发服务器 / Caddy 静态文件)
```

六个核心服务：**PostgreSQL**（数据）、**Redis**（队列/缓存）、**MinIO**（对象存储）、**Neo4j**（图数据库）、**API**（HTTP 服务）、**Worker**（异步转换）。API 和 Worker 通过 Redis 通信。

## 快速启动

```bash
git clone https://github.com/fan1ai2/vibe-coding-svg.git
cd vibe-coding-svg

# 复制并编辑环境变量（至少修改 JWT_SECRET 和 OAuth 密钥）
cp .env.example .env

docker-compose up -d
```

启动后数据库迁移自动执行。访问 `http://localhost:8080/health` 验证服务状态。

## API 接口

### 认证

| 方法 | 路径 | 认证 | 说明 |
|------|------|------|------|
| POST | `/api/v1/auth/guest` | - | 访客一键登录，返回 JWT |
| POST | `/api/v1/auth/email/send-code` | - | 发送邮箱验证码 |
| POST | `/api/v1/auth/email/verify` | - | 验证码登录/注册，返回 JWT |
| GET | `/api/v1/auth/github/login` | - | 跳转 GitHub OAuth 授权 |
| GET | `/api/v1/auth/github/callback` | - | GitHub OAuth 回调 |
| POST | `/api/v1/auth/refresh` | JWT | 刷新令牌 |
| GET | `/api/v1/auth/me` | JWT | 获取当前用户信息 |

### 转换

| 方法 | 路径 | 认证 | 说明 |
|------|------|------|------|
| POST | `/api/v1/conversions` | JWT | 上传图片（multipart 表单） |
| GET | `/api/v1/conversions` | JWT | 分页查询转换列表 |
| GET | `/api/v1/conversions/:id` | JWT | 查询单条转换状态 |
| GET | `/api/v1/conversions/:id/download` | JWT | 下载 SVG 结果文件 |

### SVG 编辑器

| 方法 | 路径 | 认证 | 说明 |
|------|------|------|------|
| POST | `/api/v1/svgs` | JWT | 保存编辑后的 SVG |
| GET | `/api/v1/svgs` | JWT | 列出已保存的 SVG |
| GET | `/api/v1/svgs/:id` | JWT | 获取单个 SVG 详情 |
| GET | `/api/v1/svgs/:id/download` | JWT | 下载 SVG 文件 |
| DELETE | `/api/v1/svgs/:id` | JWT | 删除已保存的 SVG |

### 图标库

| 方法 | 路径 | 认证 | 说明 |
|------|------|------|------|
| GET | `/api/v1/icons/search` | 可选 | Neo4j 图谱搜索图标 |
| GET | `/api/v1/icons` | 可选 | 分页浏览图标列表 |
| GET | `/api/v1/icons/:id` | 可选 | 图标详情 |
| GET | `/api/v1/icons/:id/recommend` | 可选 | 基于图谱的相似推荐 |
| POST | `/api/v1/icons` | JWT | 上传单个图标 |
| POST | `/api/v1/icons/batch` | JWT | 批量上传图标 |
| DELETE | `/api/v1/icons/:id` | JWT | 删除图标 |
| GET | `/api/v1/tags` | - | 列出所有标签 |

### AI 图标生成

| 方法 | 路径 | 认证 | 说明 |
|------|------|------|------|
| POST | `/api/v1/ai/generate` | JWT | 自然语言生成 SVG 图标 |
| GET | `/api/v1/ai/quota` | JWT | 查询当日剩余生成配额 |

### 其他

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/health` | 数据库/Redis/MinIO 连通性检查 |
| GET | `/api/v1/files/:bucket/*key` | 文件下载代理（通过 UUID key 访问，无需鉴权） |
| GET | `/docs/swagger.json` | OpenAPI 规范文档（可导入 Apifox/Postman） |

### 调用示例

```bash
# --- 邮箱登录 ---
curl -X POST http://localhost:8080/api/v1/auth/email/send-code \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com"}'

curl -X POST http://localhost:8080/api/v1/auth/email/verify \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","code":"123456"}'
# 返回: {"token":"eyJ..."}

TOKEN="eyJ..."

# --- 上传图片转换 ---
curl -X POST http://localhost:8080/api/v1/conversions \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@photo.png"

# --- AI 生成图标 ---
curl -X POST http://localhost:8080/api/v1/ai/generate \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"prompt":"一个蓝色的心形图标"}'

# --- 保存编辑后的 SVG ---
curl -X POST http://localhost:8080/api/v1/svgs \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"我的图标","svg":"<svg>...</svg>"}'
```

转换状态流转：`pending` → `processing` → `completed`（或 `failed`）。

## 配置项

所有配置通过环境变量注入（`.env` 文件或直接 export）：

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `PORT` | `8080` | API 服务端口 |
| `DATABASE_URL` | - | PostgreSQL 连接串 |
| `REDIS_ADDR` | - | Redis 地址 |
| `MINIO_ENDPOINT` | `localhost:9000` | MinIO 端点 |
| `MINIO_ACCESS_KEY` | - | MinIO 访问密钥 |
| `MINIO_SECRET_KEY` | - | MinIO 密钥 |
| `JWT_SECRET` | - | JWT 签名密钥（必填，用 `openssl rand -hex 32` 生成） |
| `GITHUB_CLIENT_ID` | - | GitHub OAuth 客户端 ID |
| `GITHUB_CLIENT_SECRET` | - | GitHub OAuth 密钥 |
| `FRONTEND_URL` | - | 前端地址（OAuth 回调 + CORS） |
| `DOMAIN` | - | 部署域名（Caddy HTTPS 自动签发） |
| `MAX_FILE_SIZE` | `10485760` | 上传文件大小上限（字节，默认 10MB） |
| `NEO4J_PASSWORD` | - | Neo4j 数据库密码 |
| `AI_BASE_URL` | `https://api.openai.com/v1` | AI 服务地址（兼容 OpenAI 协议，支持 DeepSeek 等） |
| `AI_API_KEY` | - | AI 服务 API 密钥 |
| `AI_MODEL` | `gpt-4o` | AI 模型名称 |
| `SMTP_HOST` | - | SMTP 服务器（发送邮箱验证码） |
| `SMTP_PORT` | - | SMTP 端口 |
| `SMTP_USER` | - | SMTP 用户名 |
| `SMTP_PASSWORD` | - | SMTP 密码 |
| `SMTP_FROM` | - | 发件人地址 |

## 本地开发

```bash
# 1. 启动基础设施
docker-compose up -d postgres redis minio neo4j

# 2. 复制并编辑环境变量
cp .env.example .env
# 至少填写 JWT_SECRET，其他可按需配置

# 3. 启动 API
cd server
go run ./cmd/api/

# 4. 启动 Worker（另开终端）
go run ./cmd/worker/

# 5. 启动前端（另开终端）
cd web
npm install
npm run dev
```

Worker 依赖 **vtracer** 已安装到 `$PATH`：

```bash
# macOS
brew install vtracer

# Linux
cargo install vtracer
```

前端 Vite 配置了 `/api` 代理到 `http://127.0.0.1:8080`，本地开发时无需额外配置跨域。

## 项目结构

```
server/
├── cmd/                              # 可执行程序入口
│   ├── api/main.go                   # API 服务
│   └── worker/main.go                # Worker 服务（asynq 消费者）
├── internal/
│   ├── ai/                           # AI 图标生成
│   │   ├── client.go                 # OpenAI 兼容客户端
│   │   ├── prompt.go                 # SVG 生成提示词构建
│   │   └── quota.go                  # Redis 每日配额（20 次/天）
│   ├── config/config.go              # 环境变量配置加载
│   ├── handler/                      # HTTP 请求处理器
│   │   ├── ai.go                     # AI 生成接口
│   │   ├── auth.go                   # 认证接口（邮箱/GitHub/访客）
│   │   ├── conversion.go             # 转换接口（上传/列表/状态/下载）
│   │   ├── file.go                   # 文件代理下载
│   │   ├── health.go                 # 健康检查
│   │   ├── icon.go                   # 图标 CRUD + 搜索
│   │   ├── saved_svg.go              # 编辑器 SVG 保存
│   │   └── tag.go                    # 标签列表
│   ├── middleware/
│   │   ├── jwt.go                    # JWT 鉴权 + 可选鉴权
│   │   ├── cors.go                   # 跨域配置
│   │   ├── logging.go                # 请求日志
│   │   └── ratelimit.go              # 全局限流（Redis 滑动窗口）
│   ├── model/                        # 数据结构定义
│   │   ├── user.go                   # 用户
│   │   ├── conversion.go             # 转换任务
│   │   ├── quota.go                  # 每日配额
│   │   ├── icon.go                   # 图标
│   │   ├── icon_tag.go               # 图标-标签关联
│   │   ├── tag.go                    # 标签
│   │   └── saved_svg.go              # 已保存 SVG
│   ├── neo4j/                        # Neo4j 图数据库
│   │   ├── neo4j.go                  # 连接初始化
│   │   ├── client.go                 # 图谱操作客户端
│   │   └── sync.go                   # PostgreSQL → Neo4j 同步
│   ├── repo/                         # 数据访问层
│   │   ├── user.go                   # 用户 CRUD + 验证码管理
│   │   ├── conversion.go             # 转换记录 CRUD
│   │   ├── icon.go                   # 图标 CRUD + 搜索过滤
│   │   ├── tag.go                    # 标签查询
│   │   └── saved_svg.go              # SVG 存取
│   ├── router/router.go              # 路由注册 + 依赖注入
│   ├── service/                      # 业务逻辑层
│   │   ├── ai.go                     # AI 生成编排
│   │   ├── auth.go                   # JWT + OAuth + 邮箱验证码
│   │   ├── conversion.go             # 转换流程（上传→入队→配额检查）
│   │   ├── email.go                  # SMTP 邮件发送
│   │   ├── icon.go                   # 图标业务（含 Neo4j 同步）
│   │   └── storage.go                # MinIO 对象存储封装
│   ├── worker/                       # 后台任务处理
│   │   ├── converter.go              # vtracer CLI 封装
│   │   └── worker.go                 # 任务处理器
│   └── migrate/migrate.go            # 数据库自动迁移
├── migrations/                       # SQL 迁移脚本
│   ├── 001_create_users.up.sql
│   ├── 002_create_conversions.up.sql
│   ├── 003_create_quotas.up.sql
│   ├── 004_create_tags.up.sql
│   ├── 005_create_icons.up.sql
│   ├── 006_create_icon_tags.up.sql
│   ├── 007_create_verification_codes.up.sql
│   └── 008_create_saved_svgs.up.sql
├── Dockerfile.api
├── Dockerfile.worker
├── go.mod
└── go.sum

web/                                 # React 前端（Vite + TypeScript）
├── src/
│   ├── api/client.ts                 # API 客户端（自动附加 JWT、统一错误处理）
│   ├── context/AuthContext.tsx        # 认证上下文（token/用户/登出）
│   ├── hooks/usePolling.ts           # 轮询 Hook
│   ├── components/                   # 可复用 UI 组件
│   │   ├── Navbar.tsx                # 顶部导航栏
│   │   ├── Footer.tsx                # 页脚
│   │   ├── WorkspaceShell.tsx        # 工作区布局（含侧边导航）
│   │   ├── ToolCard.tsx              # 工具卡片
│   │   ├── ConversionCard.tsx         # 转换记录卡片
│   │   ├── DropZone.tsx              # 拖拽上传
│   │   ├── ErrorBoundary.tsx         # 错误边界
│   │   ├── LoadingSpinner.tsx        # 加载动画
│   │   ├── SearchBar.tsx             # 搜索栏（图标库）
│   │   ├── IconGrid.tsx              # 图标网格
│   │   ├── IconCard.tsx              # 图标卡片
│   │   ├── TagFilter.tsx             # 标签筛选
│   │   └── PublishDialog.tsx         # 图标发布弹窗
│   ├── features/svg-editor/          # SVG 编辑器模块
│   │   ├── components/
│   │   │   ├── SvgCanvas.tsx         # SVG 画布（粘贴/拖拽/选择）
│   │   │   ├── EditorToolbar.tsx     # 编辑工具栏
│   │   │   ├── SidePanel.tsx         # 侧边面板
│   │   │   ├── ColorPicker.tsx       # 颜色选择器
│   │   │   ├── HueSlider.tsx         # 色相滑块
│   │   │   ├── AlphaSlider.tsx       # 透明度滑块
│   │   │   ├── ColorPreview.tsx      # 颜色预览
│   │   │   ├── PresetColors.tsx      # 预设颜色
│   │   │   ├── FillStrokeTabs.tsx    # 填充/描边切换
│   │   │   ├── SBPanel.tsx           # 描边面板
│   │   │   ├── ElementInspector.tsx  # 元素检查器
│   │   │   └── ThemeReplacer.tsx     # 主题色替换
│   │   └── __tests__/                # 编辑器测试
│   ├── pages/                        # 页面组件
│   │   ├── LandingPage.tsx           # 首页（Hero + 工具卡片）
│   │   ├── CallbackPage.tsx          # OAuth 回调处理
│   │   ├── ConvertPage.tsx           # 位图转 SVG
│   │   ├── PreviewPage.tsx           # 转换结果预览
│   │   ├── LibraryPage.tsx           # 转换历史
│   │   ├── EditorPage.tsx            # SVG 颜色编辑器
│   │   ├── AiGeneratePage.tsx        # AI 图标生成
│   │   ├── IconLibraryPage.tsx       # 图标库浏览
│   │   └── IconDetailPage.tsx        # 图标详情（含推荐）
│   ├── App.tsx                       # 根组件 + 路由定义
│   ├── main.tsx                      # 应用入口
│   └── index.css                     # 全局样式 + 设计 tokens
├── index.html
├── tailwind.config.js
├── Dockerfile.web
└── package.json

Caddyfile                             # Caddy 反向代理（自动 HTTPS）
docker-compose.yml                    # 编排所有服务
docker-compose.prod.yml               # 生产环境部署配置
```

## 生产部署

```bash
# 1. 配置环境变量（填写 DOMAIN、FRONTEND_URL、OAuth、SMTP 等）
cp .env.example .env
vim .env

# 2. 构建并推送镜像到私有仓库（或使用 Docker Hub）
# 在 .env 中设置 REGISTRY=your-registry/namespace/
docker compose -f docker-compose.prod.yml build
docker compose -f docker-compose.prod.yml push

# 3. 在服务器上拉取并启动
docker compose -f docker-compose.prod.yml up -d
```

Caddy 会自动为 `DOMAIN` 配置的域名申请 Let's Encrypt HTTPS 证书。

## 技术栈

| 层 | 技术 |
|------|------|
| 后端语言 | Go 1.25 |
| HTTP 框架 | Gin |
| 任务队列 | Asynq (Redis) |
| 对象存储 | MinIO (S3 兼容) |
| 关系数据库 | PostgreSQL 16 |
| 图数据库 | Neo4j |
| 转换引擎 | vtracer (Rust) |
| AI 生成 | OpenAI 兼容 API（支持 DeepSeek 等） |
| 前端框架 | React 19 + TypeScript 5 |
| 构建工具 | Vite 6 |
| CSS 框架 | Tailwind CSS 3 |
| 路由 | React Router 7 |
| 反向代理 | Caddy 2 |
| 容器化 | Docker Compose |

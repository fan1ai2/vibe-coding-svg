## 1. Config & Dependencies

- [x] 1.1 `.env.example` / `.env` 添加 `AI_BASE_URL`、`AI_API_KEY`、`AI_MODEL` 配置项，默认指向 OpenAI
- [x] 1.2 `server/internal/config/config.go` 添加 AiBaseUrl、AiApiKey、AiModel 字段并读取环境变量
- [x] 1.3 `server/go.mod` 添加 `github.com/sashabaranov/go-openai` 依赖

## 2. Backend — AI Provider Layer

- [x] 2.1 创建 `server/internal/ai/provider.go`：定义 `Provider` 接口 `Generate(prompt, style string) ([]IconCandidate, error)`
- [x] 2.2 创建 `server/internal/ai/openai.go`：实现 OpenAI 兼容 API 调用，4 层 System Prompt 模板和 JSON 响应解析（容错 markdown fences）
- [x] 2.3 创建 `server/internal/ai/prompt.go`：Prompt 构建器，查询库内 Top 10 标签（usage_count DESC）和 5 色板注入 system prompt 层级 3
- [x] 2.4 创建 `server/internal/ai/validate.go`：SVG 校验（XML 解析、根元素检查、尺寸限制 50KB）

## 3. Backend — Quota Service

- [x] 3.1 创建 `server/internal/ai/quota.go`：Redis 计数器 `ai:quota:<user_id>:<date>`，INCR + TTL 管理
- [x] 3.2 实现 `Check(userID) (remaining int, ok bool)` 和 `Consume(userID) error`

## 4. Backend — API Layer

- [x] 4.1 创建 `server/internal/handler/ai.go`：`POST /api/v1/ai/generate` 处理器（参数校验、调用 service、返回候选）
- [x] 4.2 创建 `GET /api/v1/ai/quota` 处理器：返回剩余配额
- [x] 4.3 创建 `server/internal/service/ai.go`：编排 provider + prompt builder + validate + quota
- [x] 4.4 `server/internal/router/router.go` 注册 `/api/v1/ai` 路由组，需 JWT 认证

## 5. Frontend — API Client

- [x] 5.1 `web/src/api/client.ts` 新增 `ai` 模块：`generate(prompt, style)` 和 `quota()`
- [x] 5.2 新增 TypeScript 类型：`AiGenerateRequest`、`AiGenerateResponse`、`IconCandidate`、`QuotaInfo`

## 6. Frontend — AI Generate Page

- [x] 6.1 创建 `web/src/pages/AiGeneratePage.tsx`：3 阶段状态机（input → generating → candidates）
- [x] 6.2 Phase 1: 文本输入框 + 风格选择 (line/filled) + 剩余配额显示 + 生成按钮
- [x] 6.3 Phase 2: 生成动画（CSS pulse + "AI 正在为你设计图标..."）
- [x] 6.4 Phase 3: 2×2 候选网格，每格 SVG 内联预览 + 名称 + 推测标签
- [x] 6.5 每个候选卡片两个操作按钮：「在编辑器中打开」和「保存到图标库」
- [x] 6.6 错误状态：超时提示、配额用完提示、全部失败提示，均含重试按钮

## 7. Frontend — Route & Navigation

- [x] 7.1 `web/src/App.tsx` 注册路由 `/workspace/ai-generate`（需登录）
- [x] 7.2 `web/src/components/Navbar.tsx` 已登录状态添加「AI 生成」导航链接
- [x] 7.3 `web/src/pages/LandingPage.tsx` AI 生成图标卡片改为 `available: true`, `href: '/workspace/ai-generate'`

## 8. Integration & Verification

- [ ] 8.1 验证 AI 生成 → EditorPage 管线（sessionStorage 桥接，复用已有机制）
- [ ] 8.2 验证 AI 生成 → 直接入库管线（PublishDialog 复用）
- [ ] 8.3 验证配额计数和跨天重置
- [ ] 8.4 验证不合规 SVG 被拦截
- [ ] 8.5 手动全链路测试：首页入口 → AI 生成 → 候选选中 → 编辑器调色 → 入库

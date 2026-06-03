## Context

当前系统已有图标库（PG + Neo4j）和 SVG 编辑器，但图标创建依赖手动输入。引入 AI 文本→SVG 生成后，用户在 `AiGeneratePage` 输入描述，后端调用 AI API 返回 4 组候选 SVG，选中后通过 sessionStorage 传入 EditorPage 调色，再发布到图标库。整个链路复用已有的管线基础设施。

技术栈已有：Go 1.25 + Gin + React 19 + TypeScript。AI SDK 是纯新增依赖。

## Goals / Non-Goals

**Goals:**
- 支持自然语言描述生成 24×24 viewBox 的标准 UI 图标 SVG
- 每次生成返回 4 组候选（名称 + SVG + 推测标签 + 提取颜色）
- 生成时注入图标库现有标签和色板作为风格约束
- 每用户每日 20 次配额，Redis 计数
- 前端 3 步交互：输入描述 → 候选网格展示 → 选中进编辑器/入库
- 生成的 SVG 通过现有管线可编辑、可入库、可下载

**Non-Goals:**
- 不做图标风格迁移（已有图标 → 另一风格）
- 不做批量生成
- 不做 SVG 动画生成
- 不做本地模型推理，仅走远程 API
- 不限定单一 AI 厂商 —— 支持任意 OpenAI 兼容端点

## Decisions

### 1. AI Provider: OpenAI 兼容 SDK + 可配置端点

选择 `github.com/sashabaranov/go-openai`（Go OpenAI SDK），通过 `.env` 配置端点，而非绑死单一厂商。

```env
AI_BASE_URL=https://api.openai.com/v1     # 或任意兼容端点
AI_API_KEY=sk-your-key-here
AI_MODEL=gpt-4o
```

**支持的端点（只需改 `.env`，无需改代码）**：

| 服务商 | AI_BASE_URL | AI_MODEL 示例 |
|--------|-------------|---------------|
| OpenAI | `https://api.openai.com/v1` | `gpt-4o` |
| DeepSeek | `https://api.deepseek.com/v1` | `deepseek-chat` |
| 自定义代理 | `https://coding.bagusgo.com/v1` | 用户自行填写 |

**理由**：
- OpenAI Chat Completions API 已是行业事实标准，所有主流厂商和代理服务均兼容
- SDK 原生支持 `BaseURL` 配置，一行代码切换端点
- DeepSeek、通义千问、智谱等国产模型全部兼容此格式
- 不需为不同厂商维护多套 SDK

**备选方案**: 绑死单个 SDK（Anthropic / OpenAI 原生）—— 灵活度差，换厂商等同于重写。当前方案让用户完全掌控模型选择。

### 2. Prompt 工程：4 层 System Prompt

#### 层级 1 — 角色定义

```
You are a professional UI icon designer specialized in creating minimal,
pixel-perfect SVG icons for modern web and mobile applications.
```

#### 层级 2 — 格式硬约束（最关键）

```
Generate SVG icons with these EXACT specifications:
- <svg> must have viewBox="0 0 24 24" and NO width/height attributes
- Line style icons: fill="none", stroke="CURRENT_COLOR", stroke-width="2",
  stroke-linecap="round", stroke-linejoin="round"
- Filled style icons: use solid fill colors, no stroke elements
- Output pure SVG only — no markdown, no code fences, no HTML wrapper
- No external references (no <use>, no url() to external resources)
- Keep paths simple and clean — good icons use as few elements as possible
```

`CURRENT_COLOR` 在发送前替换为样式对应的默认颜色（line→`#3B82F6`, filled→实际色值）。

风格选择是 Prompt 硬约束。返回后执行轻量校验：检测 fill 属性分布判断风格匹配度。不匹配时在前端候选卡片上标记警告（如"可能为 filled 风格"），但不丢弃候选。用户自行判断是否使用。

#### 层级 3 — 库内风格注入（动态查询）

服务端在每次请求前查询 PG，注入到 system prompt：

```
Reference these common tags for icon style consistency: {home, user,
search, arrow, heart, star, lock, mail, settings, bell}
Preferred color palette: #3B82F6, #10B981, #EF4444, #F59E0B, #6B7280
```

取图标库 Top 10 高频标签（按 `tags.usage_count DESC`）和 5 色板。标签和颜色为空时使用内置默认值。

#### 层级 4 — 输出格式

```
Return exactly 4 icons as a raw JSON array (no markdown fences, no extra text):
[{"name":"icon-name-line","svg_content":"<svg viewBox=\"0 0 24 24\" ...>...</svg>","tags":["tag1","tag2"]}]
The "tags" field should suggest 2-3 appropriate tags for this icon.
```

User prompt 直接使用用户输入的自然语言描述。

### 3. SVG 质量保证

**校验流程**：API 返回后逐张检查：
1. `encoding/xml` 解析，确认是合法 XML
2. 检查根元素为 `<svg>`，且含 `viewBox` 属性
3. 尺寸检查：内容不超过 50KB
4. 颜色有效性检查：复用 `neo4j.ExtractColors()`

**自动重试策略**：若全部候选校验失败，自动重试一次（调整 temperature 从 0.7 → 1.0 增加多样性）。两次均失败才返回 500。重试对前端透明，仅增加 3-5s 延迟。

**部分失败处理**：部分候选通过、部分失败时，仅返回通过校验的候选（最少 1 个即可返回）。

### 4. API 配额控制

Redis 计数器 `ai:quota:<user_id>:<YYYY-MM-DD>`，TTL=24h。每次生成前 INCR 检查，超过 20 返回 429。计数器由请求本身 INCR，不依赖外部 cron。

### 5. 前端架构

```tsx
// AiGeneratePage 状态机
type Phase = 'input' | 'generating' | 'candidates' | 'error'

// Phase 1: input — 文本输入框，placeholder="描述你想要的图标，如：一个带屋顶的家图标"
//    下方显示剩余配额 + 风格选择器 + 生成按钮
//    保持简洁，不展示示例提示词
// Phase 2: generating — 加载动画，后端返回 4 个候选
// Phase 3: candidates — 2×2 网格展示，hover 高亮
// Phase 4: error — 错误提示 + 重试
```

候选选中后：
- 「在编辑器中打开」→ sessionStorage → EditorPage（复用 Task 2 的 bridge 模式）
- 「直接入库」→ 弹出 PublishDialog，名称预填、标签预填 AI 推测的 tags（类型默认 usage）、用户可编辑/增删，确认后调用 `POST /api/v1/icons` → 跳转图标详情。标签可自由修改，不锁死。

**会话恢复**：最近一次生成结果保存到 sessionStorage（key=`ai:lastCandidates`）。用户从编辑器返回、或意外刷新页面时，自动恢复候选展示。关闭标签页后清除。不消耗额外配额。

路由：`/workspace/ai-generate`（需登录）。

### 6. 数据流

```
用户输入 "a home icon with a roof"
  │
  ▼
AiGeneratePage → POST /api/v1/ai/generate {prompt, style: "line"}
  │
  ▼
Go handler → Service → PromptBuilder(查询标签+色板) → OpenAI 兼容 API
  │
  ▼
API 返回 JSON → Validation(解析/校验) → ExtractColors
  │
  ▼
返回 4 个候选 → 前端 2×2 网格
  │
  ▼
用户选中 → sessionStorage → EditorPage(调色) → 发布/下载
```

## Risks / Trade-offs

- [生成 SVG 质量波动] → System prompt 硬约束 viewBox 和元素结构，服务端校验拦截不合格输出，不合格的丢弃
- [API 调用延迟 3-5s] → 前端显示生成动画，不阻塞 UI；Go 端设置 30s 超时
- [API 费用] → 每日 20 次配额控制成本，用户自行配置 API Key 自负费用
- [生成 SVG 与库内风格不一致] → System prompt 层级 3 注入标签和色板缓解，后续可加 few-shot 示例
- [不同模型输出格式差异] → System prompt 层级 4 严格约束 JSON 格式；解析时做容错处理（去除 markdown fences）
- [配额绕过] → Redis key 绑定 user_id + date，JWT 认证强制要求

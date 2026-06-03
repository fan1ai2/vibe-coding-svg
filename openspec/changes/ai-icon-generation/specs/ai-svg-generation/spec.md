## ADDED Requirements

### Requirement: AI generate SVG icons from text prompt
系统 SHALL 接收自然语言描述，调用 AI 服务生成 4 组 SVG 图标候选，每组包含名称、SVG 内容和推测标签。

#### Scenario: Successful generation
- **WHEN** 已认证用户提交 `{"prompt": "a home icon with a roof", "style": "line"}`
- **THEN** AI 返回 4 组图标候选，每组含 name、svg_content、tags 数组，SVG 使用 24×24 viewBox 且格式合法，HTTP 200

#### Scenario: Generate without auth
- **WHEN** 未认证用户提交生成请求
- **THEN** 返回 401 未授权错误

#### Scenario: Prompt too short
- **WHEN** 用户提交 prompt 长度小于 2 个字符
- **THEN** 返回 400 参数错误

#### Scenario: Prompt too long
- **WHEN** 用户提交 prompt 长度超过 500 个字符
- **THEN** 返回 400 参数错误

### Requirement: AI generation daily quota
系统 SHALL 限制每用户每日最多 20 次 AI 生成请求。

#### Scenario: Within quota
- **WHEN** 用户当日生成次数未达 20 次
- **THEN** 正常处理生成请求，计数器加 1

#### Scenario: Quota exceeded
- **WHEN** 用户当日生成次数已达 20 次
- **THEN** 返回 429 配额超限错误，提示 "每日生成次数已用完，请明天再试"

#### Scenario: Quota resets daily
- **WHEN** 新的一天开始
- **THEN** 用户生成计数器归零，可重新生成

### Requirement: AI prompt injection of library context
系统 SHALL 在调用 AI 服务时注入图标库中的高频标签和色板作为风格约束。

#### Scenario: Style constraints in prompt
- **WHEN** 图标库中有标签 [home, user, arrow] 和颜色 [#3B82F6, #10B981]
- **THEN** AI system prompt 包含这些标签和颜色信息作为风格参考

#### Scenario: Empty library injection
- **WHEN** 图标库中无标签和颜色数据
- **THEN** 使用默认色板 [#3B82F6, #6B7280, #10B981, #EF4444, #F59E0B] 注入 prompt

### Requirement: SVG quality validation
系统 SHALL 对 AI 返回的每张 SVG 进行合法性校验，不合格的候选被丢弃。

#### Scenario: Valid SVG candidates
- **WHEN** AI 返回 4 张 SVG，全部通过 XML 解析和结构校验
- **THEN** 返回全部 4 个候选

#### Scenario: Partially valid SVG candidates
- **WHEN** AI 返回 4 张 SVG，其中 1 张无法解析为合法 XML
- **THEN** 丢弃无效候选，返回剩余 3 个

#### Scenario: All candidates invalid
- **WHEN** AI 返回的 SVG 全部校验失败
- **THEN** 返回 500 服务端错误

### Requirement: Quota check endpoint
系统 SHALL 提供接口查询用户当日剩余生成次数。

#### Scenario: Query remaining quota
- **WHEN** 已认证用户请求 `GET /api/v1/ai/quota`
- **THEN** 返回 `{"remaining": 15, "total": 20}`，表示当日剩余 15 次

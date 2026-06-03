## ADDED Requirements

### Requirement: AI generate page
系统 SHALL 在 `/workspace/ai-generate` 提供 AI 图标生成页面，包含文本输入、生成状态展示和候选选择三个阶段的交互。

#### Scenario: Page loads with input form
- **WHEN** 已认证用户访问 `/workspace/ai-generate`
- **THEN** 页面显示文本输入框、可选风格选择器 (line / filled) 和生成按钮，配额信息显示剩余次数

#### Scenario: Generating state
- **WHEN** 用户点击生成按钮
- **THEN** 显示生成动画和提示文字 "AI 正在为你设计图标..."，输入框禁用

#### Scenario: Candidates grid display
- **WHEN** AI 返回 4 组候选图标
- **THEN** 页面以 2×2 网格展示，每格内联渲染 SVG 缩略图，下方显示名称和推测标签

### Requirement: Select candidate and open in editor
系统 SHALL 支持选中候选图标后在 SVG 编辑器中打开。

#### Scenario: Open in editor
- **WHEN** 用户点击某个候选图标的「在编辑器中打开」按钮
- **THEN** 通过 sessionStorage 传递 SVG 内容，导航到 `/workspace/editor`，编辑器加载该 SVG

### Requirement: Select candidate and save to library
系统 SHALL 支持选中候选图标后直接保存到图标库。

#### Scenario: Save to icon library
- **WHEN** 用户点击某个候选图标的「保存到图标库」按钮
- **THEN** 弹出 PublishDialog（复用现有组件），填写名称和标签后调用 POST /icons，入库成功显示 toast + 跳转图标详情

### Requirement: Navbar and homepage entry
系统 SHALL 在导航栏和首页提供 AI 生成入口。

#### Scenario: Navbar shows AI generate link
- **WHEN** 用户已登录
- **THEN** 导航栏显示「AI 生成」链接，href 为 `/workspace/ai-generate`

#### Scenario: Homepage AI card is active
- **WHEN** 用户访问首页
- **THEN** AI 生成图标卡片显示为可用状态，点击跳转到 `/workspace/ai-generate`

### Requirement: Error handling
系统 SHALL 妥善处理 AI 生成过程中的错误状态。

#### Scenario: AI service timeout
- **WHEN** AI 调用超过 30 秒未返回
- **THEN** 显示 "生成超时，请重试" 错误提示，输入框重新可用

#### Scenario: Quota exceeded on page
- **WHEN** 用户配额已用完时访问页面
- **THEN** 显示 "今日生成次数已用完" 提示，输入框禁用

#### Scenario: All candidates validation failed
- **WHEN** AI 返回的候选全部校验不通过
- **THEN** 显示 "生成失败，请修改描述后重试" 错误提示

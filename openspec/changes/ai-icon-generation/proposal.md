## Why

当前项目具备图标库和 SVG 编辑器能力，但图标创建依赖人工上传或手动编辑。引入 AI 文本生成图标能力后，用户用自然语言描述即可获得多组候选 SVG，选中后直接进入编辑器调色、入库、下载。这是打通「从灵感到交付」完整工作流的最后一块拼图。

## What Changes

- 新增 AI 图标生成 API：接收文本描述，返回 4 组 SVG 候选图标
- 前端新增 AI 生成页面：输入框 + 候选展示 + 选中进编辑器
- 首页「AI 生成图标」卡片从即将推出变为可用
- Prompt 工程确保生成的 SVG 风格统一、尺寸规范、可直接渲染
- 生成时参考图标库中现有标签和颜色，在 Prompt 中注入风格约束
- API 速率限制防止滥用，每次生成消耗配额

## Capabilities

### New Capabilities
- `ai-svg-generation`: AI 驱动的 SVG 图标生成，接收自然语言描述返回多组 SVG 候选，包含 Prompt 工程、质量校验和配额控制
- `ai-generation-ui`: AI 图标生成的前端交互——输入描述、查看候选图标、选中进入编辑器或直接入库

### Modified Capabilities
<!-- 无已有规格需要修改 —— AI 生成是纯新增能力，不影响现有 API 和行为 -->

## Impact

- **Dependencies**: Go 服务新增 AI SDK 依赖（Claude API / OpenAI SDK）
- **API**: 新增 `POST /api/v1/ai/generate` 和 `GET /api/v1/ai/quota`
- **Config**: `.env` 新增 `AI_PROVIDER`、`AI_API_KEY`、`AI_MODEL` 配置项
- **Frontend**: 新增 AiGeneratePage，Navbar 和首页添加入口，与 EditorPage 串联
- **Rate Limit**: 每用户每日最多 20 次生成请求
- **Infrastructure**: 无需新增容器，AI 调用走外部 API

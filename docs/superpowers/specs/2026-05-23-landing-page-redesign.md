# 首页重构设计 — 平台型 Landing Page

## 概述

将首页从单功能登录页重构为"设计资源平台"型首页。SVG 转换作为当前唯一可用的功能板块，其余板块以占位卡片形式预留，后续按需填充。视觉风格参考 Flaticon（暖色系、大圆角、卡片布局）。

## 技术栈

- React 19 + TypeScript 5
- React Router 7
- Tailwind CSS 3
- 字体：Nunito（Google Fonts）

## 路由设计

```
/                       LandingPage（Hero + 工具卡片网格）
/callback               CallbackPage（OAuth 回调，不变）
/workspace/convert      ConvertPage（上传转换，不变）
/workspace/preview/:id  PreviewPage（预览下载，不变）
/workspace/library      LibraryPage（历史记录，不变）
*                       404 页面
```

登录态变化：
- 未登录：顶部导航显示"登录 GitHub"按钮
- 已登录：顶部导航显示"转换"、"库"链接 + 用户头像/退出

## 组件结构

```
App
├── Navbar（新增，全局顶部导航）
│   ├── Logo
│   ├── 导航链接（已登录时显示）
│   └── 用户区（登录按钮 / 头像+退出）
├── Routes
│   ├── LandingPage（重写）
│   │   ├── Hero 区
│   │   └── 工具卡片网格
│   │       ├── ToolCard × 1（SVG 转换，可用）
│   │       └── ToolCard × 3（占位，即将推出）
│   ├── CallbackPage（不变）
│   ├── ConvertPage（微调：无侧边栏）
│   ├── PreviewPage（微调：无侧边栏）
│   └── LibraryPage（微调：无侧边栏）
└── Footer（新增）
```

## 文件变更清单

| 操作 | 文件 | 说明 |
|------|------|------|
| 新建 | `src/components/Navbar.tsx` | 顶部导航栏 |
| 新建 | `src/components/ToolCard.tsx` | 工具卡片组件 |
| 新建 | `src/components/Footer.tsx` | 页脚 |
| 废弃 | `src/components/Layout.tsx` | 被 Navbar 替代，删除 |
| 重写 | `src/pages/LandingPage.tsx` | Hero + 工具卡片 |
| 修改 | `src/App.tsx` | 路由结构调整，Navbar 包裹 |
| 修改 | `tailwind.config.js` | 扩展 amber 色系 + 字体配置 |
| 修改 | `index.html` | 引入 Nunito Google Font |
| 修改 | `src/index.css` | 全局背景暖白 |

## 视觉规范

### 配色

| 用途 | Tailwind Class | 色值 |
|------|---------------|------|
| 主色 | amber-500 | #F6A623 |
| 主色 hover | amber-600 | #D4921A |
| 浅暖底色 | amber-50 | #FFF8EB |
| 页面背景 | warm-white | #FFFDF7 |
| 文字主色 | gray-900 | #1A1A1A |
| 文字辅色 | gray-500 | #6B7280 |

### 顶部导航栏

- 高度 64px（h-16）
- 背景：`bg-white/80 backdrop-blur-md`
- 滚动到非首页时变为纯白底 + 底部 1px 边框
- 左：Logo（SVG 图标 + "SVG 资源工坊"文字）
- 中：搜索框占位（后续实现，当前可用 opacity-0 占位）
- 右：已登录 → "转换"、"库"链接 + 用户菜单；未登录 → "登录 GitHub"按钮

### 工具卡片

- 大圆角 `rounded-2xl`
- 浅暖底色 `bg-amber-50`
- 居中大图标（48px，amber-500）
- hover：`scale-[1.02]` + `shadow-lg`
- 过渡动画 `transition-all duration-300`
- 占位卡片：`opacity-50` + 锁 icon + "即将推出"标签
- 网格：1 列（移动）→ 2 列（sm）→ 4 列（lg）

### 字体

- 主字体：Nunito（Google Fonts）
- Tailwind 配置：`fontFamily: { sans: ['Nunito', 'system-ui', 'sans-serif'] }`

## 不在范围内

- 搜索功能实现（仅 UI 占位）
- 用户头像图片（先用文字首字母替代）
- 移动端响应式优化（后续单独处理）
- 暗色模式

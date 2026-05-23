# 首页重构实现计划

> **面向自动化工作者：** 必需子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 按任务逐步实现此计划。步骤使用复选框 (`- [ ]`) 语法进行跟踪。

**目标：** 将首页从单功能登录页重构为平台型 Landing Page，参考 Flaticon 视觉风格（暖色系、大圆角、卡片布局）。

**架构：** 新增 Navbar（顶部导航栏）替代 Layout（侧边栏），新增 ToolCard 和 Footer 组件，重写 LandingPage 为 Hero + 工具卡片网格结构。

**技术栈：** React 19、TypeScript 5、React Router 7、Tailwind CSS 3、Nunito 字体

**文件清单：**
- 创建：`web/src/components/Navbar.tsx` — 顶部导航栏
- 创建：`web/src/components/ToolCard.tsx` — 工具卡片
- 创建：`web/src/components/Footer.tsx` — 页脚
- 创建：`web/src/components/WorkspaceShell.tsx` — 工作区页面壳
- 重写：`web/src/pages/LandingPage.tsx` — Hero + 工具卡片网格
- 修改：`web/src/App.tsx` — 路由结构调整
- 修改：`web/tailwind.config.js` — amber 色系 + 字体
- 修改：`web/index.html` — 字体 CDN + 中文 lang
- 修改：`web/src/index.css` — 全局背景
- 删除：`web/src/components/Layout.tsx` — 被 Navbar 替代

---

### 任务 1：Tailwind 配置 + 全局样式

**涉及文件：**
- 修改：`web/tailwind.config.js`
- 修改：`web/index.html`
- 修改：`web/src/index.css`

- [ ] **步骤 1：扩展 Tailwind 配置**

```js
/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ['Nunito', 'system-ui', 'sans-serif'],
      },
    },
  },
  plugins: [],
}
```

- [ ] **步骤 2：更新 index.html**

```html
<!DOCTYPE html>
<html lang="zh-CN">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <link rel="preconnect" href="https://fonts.googleapis.com" />
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin />
    <link href="https://fonts.googleapis.com/css2?family=Nunito:wght@400;600;700;800&display=swap" rel="stylesheet" />
    <title>SVG 资源工坊</title>
  </head>
  <body class="bg-[#FFFDF7] text-gray-900 min-h-screen">
    <div id="root"></div>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>
```

- [ ] **步骤 3：更新全局 CSS**

```css
@tailwind base;
@tailwind components;
@tailwind utilities;
```

- [ ] **步骤 4：提交**

```bash
cd /svg-project && git add web/tailwind.config.js web/index.html web/src/index.css
git commit -m "chore: amber warm theme, Nunito font, Chinese lang"
```

---

### 任务 2：ToolCard 组件

**涉及文件：**
- 创建：`web/src/components/ToolCard.tsx`

- [ ] **步骤 1：创建 ToolCard.tsx**

```tsx
import type { ReactNode } from 'react';
import { Link } from 'react-router-dom';

interface ToolCardProps {
  icon: ReactNode;
  title: string;
  description: string;
  href?: string;
  available?: boolean;
}

export default function ToolCard({ icon, title, description, href, available = false }: ToolCardProps) {
  if (available && href) {
    return (
      <Link
        to={href}
        className="group block rounded-2xl bg-amber-50 p-8 transition-all duration-300 hover:-translate-y-1 hover:shadow-lg"
      >
        <div className="flex h-12 w-12 items-center justify-center text-amber-500">
          {icon}
        </div>
        <h3 className="mt-4 text-lg font-bold text-gray-900">{title}</h3>
        <p className="mt-2 text-sm text-gray-500">{description}</p>
      </Link>
    );
  }

  return (
    <div className="cursor-default rounded-2xl bg-gray-50 p-8 opacity-50 transition-all duration-300">
      <div className="flex items-center gap-3">
        <div className="flex h-12 w-12 items-center justify-center text-gray-400">
          {icon}
        </div>
        <span className="rounded-full border border-gray-300 px-3 py-0.5 text-xs font-medium text-gray-400">
          即将推出
        </span>
      </div>
      <h3 className="mt-4 text-lg font-bold text-gray-900">{title}</h3>
      <p className="mt-2 text-sm text-gray-500">{description}</p>
    </div>
  );
}
```

- [ ] **步骤 2：验证类型检查**

```bash
cd /svg-project/web && npx tsc --noEmit
```

预期结果：无错误。

- [ ] **步骤 3：提交**

```bash
cd /svg-project && git add web/src/components/ToolCard.tsx
git commit -m "feat: add ToolCard component with available/placeholder states"
```

---

### 任务 3：Footer 组件

**涉及文件：**
- 创建：`web/src/components/Footer.tsx`

- [ ] **步骤 1：创建 Footer.tsx**

```tsx
import { Link } from 'react-router-dom';

export default function Footer() {
  return (
    <footer className="border-t border-gray-100 bg-white">
      <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-8">
        <div className="flex items-center gap-2 text-sm text-gray-500">
          <svg className="h-5 w-5 text-amber-500" fill="currentColor" viewBox="0 0 24 24">
            <path d="M4 4h16v2H4V4zm0 6h16v2H4v-2zm0 6h10v2H4v-2z" />
          </svg>
          <span>SVG 资源工坊</span>
        </div>
        <div className="flex gap-6 text-sm text-gray-400">
          <Link to="/" className="hover:text-gray-600 transition-colors">首页</Link>
          <a href="https://github.com/fan1ai2/vibe-coding-svg" target="_blank" rel="noopener noreferrer" className="hover:text-gray-600 transition-colors">GitHub</a>
        </div>
      </div>
    </footer>
  );
}
```

- [ ] **步骤 2：验证类型检查**

```bash
cd /svg-project/web && npx tsc --noEmit
```

预期结果：无错误。

- [ ] **步骤 3：提交**

```bash
cd /svg-project && git add web/src/components/Footer.tsx
git commit -m "feat: add Footer component"
```

---

### 任务 4：Navbar 组件

**涉及文件：**
- 创建：`web/src/components/Navbar.tsx`

- [ ] **步骤 1：创建 Navbar.tsx**

```tsx
import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export default function Navbar() {
  const { token, userId, login, logout } = useAuth();
  const navigate = useNavigate();
  const [menuOpen, setMenuOpen] = useState(false);

  const handleLogout = () => {
    logout();
    navigate('/');
    setMenuOpen(false);
  };

  return (
    <nav className="sticky top-0 z-50 bg-white/80 backdrop-blur-md border-b border-gray-100">
      <div className="mx-auto flex max-w-6xl items-center justify-between px-6 h-16">
        {/* 左侧 Logo */}
        <Link to="/" className="flex items-center gap-2 text-lg font-extrabold text-gray-900">
          <svg className="h-7 w-7 text-amber-500" fill="currentColor" viewBox="0 0 24 24">
            <path fillRule="evenodd" d="M10.5 3.75a6.75 6.75 0 100 13.5 6.75 6.75 0 000-13.5zM2.25 10.5a8.25 8.25 0 1114.59 5.28l4.69 4.69a.75.75 0 11-1.06 1.06l-4.69-4.69A8.25 8.25 0 012.25 10.5z" clipRule="evenodd" />
          </svg>
          <span>SVG 资源工坊</span>
        </Link>

        {/* 搜索框占位 */}
        <div className="hidden sm:block flex-1 max-w-md mx-8">
          <div className="relative">
            <svg className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z" />
            </svg>
            <input
              type="text"
              placeholder="搜索工具..."
              className="w-full rounded-xl border border-gray-200 bg-gray-50 py-2 pl-10 pr-4 text-sm text-gray-500 placeholder:text-gray-400 focus:outline-none focus:ring-2 focus:ring-amber-200"
            />
          </div>
        </div>

        {/* 右侧 */}
        <div className="flex items-center gap-3">
          {token ? (
            <>
              <Link
                to="/workspace/convert"
                className="rounded-lg px-3 py-2 text-sm font-semibold text-gray-600 hover:bg-gray-100 transition-colors"
              >
                转换
              </Link>
              <Link
                to="/workspace/library"
                className="rounded-lg px-3 py-2 text-sm font-semibold text-gray-600 hover:bg-gray-100 transition-colors"
              >
                库
              </Link>
              <div className="relative">
                <button
                  onClick={() => setMenuOpen(!menuOpen)}
                  className="flex h-8 w-8 items-center justify-center rounded-full bg-amber-100 text-sm font-bold text-amber-600 hover:bg-amber-200 transition-colors"
                >
                  {userId?.charAt(0).toUpperCase() ?? '?'}
                </button>
                {menuOpen && (
                  <>
                    <div className="fixed inset-0 z-10" onClick={() => setMenuOpen(false)} />
                    <div className="absolute right-0 top-full mt-2 w-48 rounded-xl border border-gray-100 bg-white shadow-lg z-20 py-1">
                      <div className="px-4 py-2 text-xs text-gray-400 truncate">{userId}</div>
                      <div className="border-t border-gray-50" />
                      <button
                        onClick={handleLogout}
                        className="w-full text-left px-4 py-2 text-sm text-gray-600 hover:bg-gray-50 transition-colors"
                      >
                        退出登录
                      </button>
                    </div>
                  </>
                )}
              </div>
            </>
          ) : (
            <button
              onClick={login}
              className="inline-flex items-center gap-2 rounded-xl border border-gray-300 px-4 py-2 text-sm font-semibold text-gray-700 hover:bg-gray-50 transition-colors"
            >
              <svg className="h-4 w-4" fill="currentColor" viewBox="0 0 24 24">
                <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
              </svg>
              登录 GitHub
            </button>
          )}
        </div>
      </div>
    </nav>
  );
}
```

- [ ] **步骤 2：验证类型检查**

```bash
cd /svg-project/web && npx tsc --noEmit
```

预期结果：无错误。

- [ ] **步骤 3：提交**

```bash
cd /svg-project && git add web/src/components/Navbar.tsx
git commit -m "feat: add Navbar with top navigation and auth state"
```

---

### 任务 5：重写 LandingPage

**涉及文件：**
- 重写：`web/src/pages/LandingPage.tsx`

- [ ] **步骤 1：重写 LandingPage.tsx**

```tsx
import { useAuth } from '../context/AuthContext';
import { Navigate } from 'react-router-dom';
import ToolCard from '../components/ToolCard';

const tools = [
  {
    icon: (
      <svg fill="none" viewBox="0 0 24 24" stroke="currentColor" className="h-12 w-12">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5}
          d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
      </svg>
    ),
    title: 'SVG 转换',
    description: '将位图图片快速转换为高质量矢量 SVG 文件，支持多种格式',
    href: '/workspace/convert',
    available: true,
  },
  {
    icon: (
      <svg fill="none" viewBox="0 0 24 24" stroke="currentColor" className="h-12 w-12">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5}
          d="M3.75 6A2.25 2.25 0 016 3.75h2.25A2.25 2.25 0 0110.5 6v2.25a2.25 2.25 0 01-2.25 2.25H6a2.25 2.25 0 01-2.25-2.25V6zM13.5 6a2.25 2.25 0 012.25-2.25H18A2.25 2.25 0 0120.25 6v2.25A2.25 2.25 0 0118 10.5h-2.25a2.25 2.25 0 01-2.25-2.25V6z" />
      </svg>
    ),
    title: '图标库',
    description: '海量高质量 SVG 图标资源，支持在线编辑和自定义导出',
    href: undefined,
    available: false,
  },
  {
    icon: (
      <svg fill="none" viewBox="0 0 24 24" stroke="currentColor" className="h-12 w-12">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5}
          d="M15.042 21.672L13.684 16.6m0 0l-2.51 2.225.569-9.47 5.227 7.917-3.286-.672zm-7.518-.267A8.25 8.25 0 1120.25 10.5M8.288 14.212A5.25 5.25 0 1117.25 10.5" />
      </svg>
    ),
    title: '调色板',
    description: '智能生成配色方案，支持渐变色提取和色彩对比度检测',
    href: undefined,
    available: false,
  },
  {
    icon: (
      <svg fill="none" viewBox="0 0 24 24" stroke="currentColor" className="h-12 w-12">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5}
          d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z" />
      </svg>
    ),
    title: '格式工厂',
    description: '支持 SVG、PNG、WebP、PDF 等多种格式的批量互转',
    href: undefined,
    available: false,
  },
];

export default function LandingPage() {
  const { token, loading } = useAuth();

  if (loading) return null;
  if (token) return <Navigate to="/workspace/convert" replace />;

  return (
    <div className="min-h-screen bg-[#FFFDF7]">
      {/* Hero 区 */}
      <section className="relative overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-br from-amber-50 via-white to-amber-50/30" />
        <div className="relative mx-auto max-w-6xl px-6 py-24 text-center sm:py-32">
          <h1 className="text-4xl font-extrabold tracking-tight text-gray-900 sm:text-5xl lg:text-6xl">
            创意设计资源，一站即达
          </h1>
          <p className="mx-auto mt-6 max-w-2xl text-lg text-gray-500 leading-relaxed">
            高质量的设计工具和资源平台，帮助你快速完成从位图到矢量、
            从灵感到交付的完整设计链路。
          </p>
          <div className="mt-10">
            <button
              onClick={() => {
                window.location.href = '/api/v1/auth/github/login';
              }}
              className="inline-flex items-center gap-2 rounded-2xl bg-amber-500 px-8 py-3.5 text-base font-bold text-gray-900 shadow-md shadow-amber-200 transition-all duration-300 hover:-translate-y-0.5 hover:bg-amber-600 hover:shadow-lg hover:shadow-amber-300"
            >
              开始免费使用
              <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 8l4 4m0 0l-4 4m4-4H3" />
              </svg>
            </button>
          </div>
        </div>
      </section>

      {/* 工具卡片区 */}
      <section className="mx-auto max-w-6xl px-6 pb-24">
        <div className="mb-10 text-center">
          <h2 className="text-2xl font-extrabold text-gray-900 sm:text-3xl">我们的工具</h2>
          <p className="mt-3 text-gray-500">更多实用工具正在开发中，敬请期待</p>
        </div>
        <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
          {tools.map((tool) => (
            <ToolCard key={tool.title} {...tool} />
          ))}
        </div>
      </section>
    </div>
  );
}
```

- [ ] **步骤 2：验证类型检查**

```bash
cd /svg-project/web && npx tsc --noEmit
```

预期结果：无错误。

- [ ] **步骤 3：提交**

```bash
cd /svg-project && git add web/src/pages/LandingPage.tsx
git commit -m "feat: rewrite LandingPage with Hero + tool grid"
```

---

### 任务 6：新增 WorkspaceShell + 更新 App.tsx 路由结构

**涉及文件：**
- 创建：`web/src/components/WorkspaceShell.tsx`
- 修改：`web/src/App.tsx`

- [ ] **步骤 1：创建 WorkspaceShell.tsx**

```tsx
import { Outlet } from 'react-router-dom';
import Navbar from './Navbar';

export default function WorkspaceShell() {
  return (
    <div className="min-h-screen bg-[#FFFDF7]">
      <Navbar />
      <div className="mx-auto max-w-6xl px-6 py-8">
        <Outlet />
      </div>
    </div>
  );
}
```

- [ ] **步骤 2：重写 App.tsx**

```tsx
import { Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './context/AuthContext';
import ErrorBoundary from './components/ErrorBoundary';
import Navbar from './components/Navbar';
import Footer from './components/Footer';
import WorkspaceShell from './components/WorkspaceShell';
import LoadingSpinner from './components/LoadingSpinner';
import LandingPage from './pages/LandingPage';
import CallbackPage from './pages/CallbackPage';
import ConvertPage from './pages/ConvertPage';
import PreviewPage from './pages/PreviewPage';
import LibraryPage from './pages/LibraryPage';

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { token, loading } = useAuth();
  if (loading) return <LoadingSpinner label="正在检查认证状态..." />;
  if (!token) return <Navigate to="/" replace />;
  return <>{children}</>;
}

export default function App() {
  return (
    <AuthProvider>
      <Routes>
        {/* 公开页面：首页 */}
        <Route path="/" element={<><Navbar /><LandingPage /><Footer /></>} />

        {/* 公开页面：OAuth 回调 */}
        <Route path="/callback" element={<CallbackPage />} />

        {/* 工作区（需登录） */}
        <Route
          path="/workspace"
          element={
            <ProtectedRoute>
              <ErrorBoundary>
                <WorkspaceShell />
              </ErrorBoundary>
            </ProtectedRoute>
          }
        >
          <Route index element={<Navigate to="convert" replace />} />
          <Route path="convert" element={<ConvertPage />} />
          <Route path="preview/:id" element={<PreviewPage />} />
          <Route path="library" element={<LibraryPage />} />
        </Route>

        {/* 404 */}
        <Route path="*" element={<div className="p-8 text-center text-gray-500">404 — 页面未找到</div>} />
      </Routes>
    </AuthProvider>
  );
}
```

- [ ] **步骤 3：验证类型检查**

```bash
cd /svg-project/web && npx tsc --noEmit
```

预期结果：无错误。

- [ ] **步骤 4：提交**

```bash
cd /svg-project && git add web/src/components/WorkspaceShell.tsx web/src/App.tsx
git commit -m "refactor: replace sidebar Layout with top Navbar and WorkspaceShell"
```

---

### 任务 7：删除 Layout.tsx

**涉及文件：**
- 删除：`web/src/components/Layout.tsx`

- [ ] **步骤 1：删除文件并验证**

```bash
cd /svg-project && git rm web/src/components/Layout.tsx
cd /svg-project/web && npx tsc --noEmit
```

预期结果：无错误（Layout 的引用已从 App.tsx 中移除）。

- [ ] **步骤 2：提交**

```bash
cd /svg-project && git add web/src/components/Layout.tsx
git commit -m "refactor: remove deprecated Layout component"
```

---

### 任务 8：完整验证

- [ ] **步骤 1：TypeScript 全量类型检查**

```bash
cd /svg-project/web && npx tsc --noEmit
```

预期结果：零类型错误。

- [ ] **步骤 2：确认文件结构正确**

```bash
cd /svg-project && ls web/src/components/
```

预期结果：Navbar.tsx、ToolCard.tsx、Footer.tsx 存在，Layout.tsx 不存在。

- [ ] **步骤 3：启动开发服务器冒烟测试**

```bash
cd /svg-project/web && npx vite build --emptyOutDir 2>&1 | tail -5
```

预期结果：构建成功，无报错。

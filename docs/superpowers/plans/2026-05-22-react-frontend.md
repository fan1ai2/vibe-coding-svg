# React Frontend Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a complete React SPA frontend in `web/` that covers upload → processing → preview → download, with OAuth login and conversion history.

**Architecture:** Vite dev server proxies `/api` to Go backend on :8080. OAuth callback flow redirects through backend, which sends the browser back to frontend's `/callback?token=...` route. JWT stored in localStorage. Protected routes behind AuthContext gate.

**Tech Stack:** Vite 6, React 19, React Router 7, Tailwind CSS 3, TypeScript 5

**File Structure:**
```
web/
├── index.html
├── package.json
├── tsconfig.json
├── tsconfig.app.json
├── vite.config.ts
├── tailwind.config.js
├── postcss.config.js
├── src/
│   ├── main.tsx              # Entry point
│   ├── App.tsx               # Router + AuthProvider
│   ├── index.css             # Tailwind directives
│   ├── api/
│   │   └── client.ts         # Fetch wrapper, JWT injection
│   ├── context/
│   │   └── AuthContext.tsx    # Token + user state
│   ├── components/
│   │   ├── Layout.tsx         # Sidebar + header + Outlet
│   │   ├── DropZone.tsx       # Drag/drop file upload
│   │   ├── ConversionCard.tsx # Library grid card
│   │   └── LoadingSpinner.tsx # Reusable spinner
│   ├── pages/
│   │   ├── LandingPage.tsx    # Hero + login CTAs
│   │   ├── CallbackPage.tsx   # Extract token from URL
│   │   ├── ConvertPage.tsx    # Upload + processing polling
│   │   ├── PreviewPage.tsx    # SVG + original comparison
│   │   └── LibraryPage.tsx    # Conversion history grid
│   └── hooks/
│       ├── usePolling.ts      # Generic polling hook
│       └── useConversions.ts  # Conversion data hook
server/internal/
    ├── config/config.go       # MODIFY: add FrontendURL
    └── handler/auth.go        # MODIFY: redirect to FRONTEND_URL
```

**Existing API Endpoints:**
| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| GET | /api/v1/auth/github/login | No | GitHub OAuth redirect |
| GET | /api/v1/auth/github/callback | No | GitHub callback → JWT |
| GET | /api/v1/auth/google/login | No | Google OAuth redirect |
| GET | /api/v1/auth/google/callback | No | Google callback → JWT |
| POST | /api/v1/auth/refresh | JWT | Refresh token |
| GET | /api/v1/auth/me | JWT | Current user info |
| POST | /api/v1/conversions | JWT | Upload file → conversion |
| GET | /api/v1/conversions | JWT | List user's conversions |
| GET | /api/v1/conversions/:id | JWT | Get conversion status |
| GET | /api/v1/conversions/:id/download | JWT | Download SVG file |

**Backend Response Types (mirrors Go model):**
```typescript
type Conversion = {
  id: string;
  user_id: string;
  status: "pending" | "processing" | "completed" | "failed";
  original_url: string;
  svg_url: string | null;
  thumbnail_url: string | null;
  file_size_in: number;
  file_size_out: number;
  path_count: number;
  color_count: number;
  format_in: string;
  error_message: string;
  created_at: string;
  completed_at: string | null;
};
```

---

### Task 1: Project scaffolding

**Files:**
- Create: `web/package.json`
- Create: `web/index.html`
- Create: `web/tsconfig.json`
- Create: `web/tsconfig.app.json`
- Create: `web/vite.config.ts`
- Create: `web/tailwind.config.js`
- Create: `web/postcss.config.js`
- Create: `web/src/main.tsx`
- Create: `web/src/index.css`
- Create: `web/src/App.tsx`

- [ ] **Step 1: Create web/package.json**

```json
{
  "name": "svg-converter-web",
  "private": true,
  "version": "0.1.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "tsc -b && vite build",
    "preview": "vite preview"
  },
  "dependencies": {
    "react": "^19.1.0",
    "react-dom": "^19.1.0",
    "react-router-dom": "^7.6.0"
  },
  "devDependencies": {
    "@types/react": "^19.1.0",
    "@types/react-dom": "^19.1.0",
    "@vitejs/plugin-react": "^4.4.1",
    "autoprefixer": "^10.4.21",
    "postcss": "^8.5.3",
    "tailwindcss": "^3.4.17",
    "typescript": "~5.8.3",
    "vite": "^6.3.0"
  }
}
```

- [ ] **Step 2: Create web/index.html**

```html
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>SVG Converter</title>
  </head>
  <body class="bg-gray-50 text-gray-900 min-h-screen">
    <div id="root"></div>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>
```

- [ ] **Step 3: Create web/tsconfig.json**

```json
{
  "files": [],
  "references": [
    { "path": "./tsconfig.app.json" }
  ]
}
```

- [ ] **Step 4: Create web/tsconfig.app.json**

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForClassFields": true,
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "isolatedModules": true,
    "moduleDetection": "force",
    "noEmit": true,
    "jsx": "react-jsx",
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true,
    "noUncheckedSideEffectImports": true
  },
  "include": ["src"]
}
```

- [ ] **Step 5: Create web/vite.config.ts**

```typescript
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
```

- [ ] **Step 6: Create web/tailwind.config.js**

```javascript
/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {},
  },
  plugins: [],
}
```

- [ ] **Step 7: Create web/postcss.config.js**

```javascript
export default {
  plugins: {
    tailwindcss: {},
    autoprefixer: {},
  },
}
```

- [ ] **Step 8: Create web/src/index.css**

```css
@tailwind base;
@tailwind components;
@tailwind utilities;
```

- [ ] **Step 9: Create web/src/main.tsx**

```tsx
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import App from './App'
import './index.css'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <BrowserRouter>
      <App />
    </BrowserRouter>
  </StrictMode>,
)
```

- [ ] **Step 10: Create web/src/App.tsx (placeholder)**

```tsx
export default function App() {
  return <div className="p-8 text-2xl font-bold">SVG Converter</div>
}
```

- [ ] **Step 11: Install dependencies and verify**

```bash
cd /svg-project/web && npm install
```

Expected: installs without errors.

- [ ] **Step 12: Verify dev server starts**

```bash
cd /svg-project/web && npx vite --host 0.0.0.0 &
sleep 3
curl -s http://localhost:5173 | head -5
kill %1
```

Expected: HTML response with "SVG Converter" title.

- [ ] **Step 13: Commit**

```bash
git add web/
git commit -m "feat: scaffold React frontend with Vite, Tailwind, and React Router"
```

---

### Task 2: API client module

**Files:**
- Create: `web/src/api/client.ts`

- [ ] **Step 1: Create web/src/api/client.ts**

```typescript
const BASE = '/api/v1';

function token(): string | null {
  return localStorage.getItem('token');
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers: Record<string, string> = {
    ...(options.headers as Record<string, string> ?? {}),
  };
  const tok = token();
  if (tok) {
    headers['Authorization'] = `Bearer ${tok}`;
  }
  // Don't set Content-Type for FormData (browser sets it with boundary)
  if (!(options.body instanceof FormData)) {
    headers['Content-Type'] = 'application/json';
  }

  const res = await fetch(`${BASE}${path}`, { ...options, headers });
  const data = await res.json();
  if (!res.ok) {
    throw new ApiError(res.status, data?.error?.code ?? 'UNKNOWN', data?.error?.message ?? 'Request failed');
  }
  return data as T;
}

export class ApiError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string,
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

// Auth
export const auth = {
  me: () => request<{ user_id: string }>('/auth/me'),
  refresh: () => request<{ token: string }>('/auth/refresh', { method: 'POST' }),
};

// Conversions
export type Conversion = {
  id: string;
  user_id: string;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  original_url: string;
  svg_url: string | null;
  thumbnail_url: string | null;
  file_size_in: number;
  file_size_out: number;
  path_count: number;
  color_count: number;
  format_in: string;
  error_message: string;
  created_at: string;
  completed_at: string | null;
};

export type ConversionListResponse = {
  data: Conversion[];
};

export type ConversionSingleResponse = {
  data: Conversion;
};

export const conversions = {
  upload: (file: File) => {
    const form = new FormData();
    form.append('file', file);
    return request<ConversionSingleResponse>('/conversions', { method: 'POST', body: form });
  },
  list: (limit = 20, offset = 0) =>
    request<ConversionListResponse>(`/conversions?limit=${limit}&offset=${offset}`),
  get: (id: string) =>
    request<ConversionSingleResponse>(`/conversions/${id}`),
  downloadUrl: (id: string) => `${BASE}/conversions/${id}/download`,
};
```

Run: `cd /svg-project/web && npx tsc --noEmit`
Expected: no type errors.

- [ ] **Step 2: Commit**

```bash
git add web/src/api/client.ts
git commit -m "feat: add API client module with typed endpoints"
```

---

### Task 3: AuthContext — JWT authentication state

**Files:**
- Create: `web/src/context/AuthContext.tsx`

- [ ] **Step 1: Create web/src/context/AuthContext.tsx**

```tsx
import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from 'react';
import { auth, ApiError } from '../api/client';

interface AuthState {
  token: string | null;
  userId: string | null;
  loading: boolean;
  login: () => void;
  logout: () => void;
}

const AuthContext = createContext<AuthState | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(() => localStorage.getItem('token'));
  const [userId, setUserId] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!token) {
      setLoading(false);
      return;
    }
    auth.me()
      .then(data => setUserId(data.user_id))
      .catch(err => {
        if (err instanceof ApiError && err.status === 401) {
          localStorage.removeItem('token');
          setToken(null);
        }
      })
      .finally(() => setLoading(false));
  }, [token]);

  const login = useCallback(() => {
    window.location.href = '/api/v1/auth/github/login';
  }, []);

  const logout = useCallback(() => {
    localStorage.removeItem('token');
    setToken(null);
    setUserId(null);
  }, []);

  return (
    <AuthContext.Provider value={{ token, userId, loading, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
```

- [ ] **Step 2: Verify no type errors**

```bash
cd /svg-project/web && npx tsc --noEmit
```
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add web/src/context/AuthContext.tsx
git commit -m "feat: add AuthContext with JWT token management"
```

---

### Task 4: LoadingSpinner component

**Files:**
- Create: `web/src/components/LoadingSpinner.tsx`

- [ ] **Step 1: Create web/src/components/LoadingSpinner.tsx**

```tsx
export default function LoadingSpinner({ label = 'Loading...' }: { label?: string }) {
  return (
    <div className="flex flex-col items-center justify-center gap-3 py-12">
      <div className="h-8 w-8 animate-spin rounded-full border-4 border-gray-300 border-t-indigo-600" />
      <span className="text-sm text-gray-500">{label}</span>
    </div>
  );
}
```

- [ ] **Step 2: Verify type check**

```bash
cd /svg-project/web && npx tsc --noEmit
```
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add web/src/components/LoadingSpinner.tsx
git commit -m "feat: add LoadingSpinner component"
```

---

### Task 5: App.tsx — Router with auth gating

**Files:**
- Modify: `web/src/App.tsx`

- [ ] **Step 1: Rewrite web/src/App.tsx**

```tsx
import { Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './context/AuthContext';
import Layout from './components/Layout';
import LoadingSpinner from './components/LoadingSpinner';
import LandingPage from './pages/LandingPage';
import CallbackPage from './pages/CallbackPage';
import ConvertPage from './pages/ConvertPage';
import PreviewPage from './pages/PreviewPage';
import LibraryPage from './pages/LibraryPage';

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { token, loading } = useAuth();
  if (loading) return <LoadingSpinner label="Checking authentication..." />;
  if (!token) return <Navigate to="/" replace />;
  return <>{children}</>;
}

export default function App() {
  return (
    <AuthProvider>
      <Routes>
        <Route path="/" element={<LandingPage />} />
        <Route path="/callback" element={<CallbackPage />} />
        <Route
          path="/workspace"
          element={
            <ProtectedRoute>
              <Layout />
            </ProtectedRoute>
          }
        >
          <Route index element={<Navigate to="convert" replace />} />
          <Route path="convert" element={<ConvertPage />} />
          <Route path="preview/:id" element={<PreviewPage />} />
          <Route path="library" element={<LibraryPage />} />
        </Route>
        <Route path="*" element={<div className="p-8 text-center text-gray-500">404 — Page not found</div>} />
      </Routes>
    </AuthProvider>
  );
}
```

- [ ] **Step 2: Verify type check**

```bash
cd /svg-project/web && npx tsc --noEmit
```
Expected: errors about missing page/component modules — those are created in next tasks. Acceptable at this point.

- [ ] **Step 3: Commit**

```bash
git add web/src/App.tsx
git commit -m "feat: add router with protected workspace routes"
```

---

### Task 6: LandingPage

**Files:**
- Create: `web/src/pages/LandingPage.tsx`

- [ ] **Step 1: Create web/src/pages/LandingPage.tsx**

```tsx
import { useAuth } from '../context/AuthContext';
import { Navigate } from 'react-router-dom';

export default function LandingPage() {
  const { token, loading, login } = useAuth();

  if (loading) return null;
  if (token) return <Navigate to="/workspace" replace />;

  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-gradient-to-br from-indigo-50 to-white px-4">
      <div className="max-w-lg text-center">
        <h1 className="text-4xl font-bold tracking-tight text-gray-900 sm:text-5xl">
          Image to SVG Converter
        </h1>
        <p className="mt-4 text-lg text-gray-600">
          Convert raster images to clean vector SVG files. Drag, drop, download.
          Free up to {20} conversions per day.
        </p>
        <div className="mt-10 flex flex-col sm:flex-row gap-4 justify-center">
          <button
            onClick={login}
            className="inline-flex items-center justify-center gap-2 rounded-lg bg-gray-900 px-6 py-3 text-sm font-semibold text-white hover:bg-gray-700 transition-colors"
          >
            <svg className="h-5 w-5" fill="currentColor" viewBox="0 0 24 24">
              <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
            </svg>
            Login with GitHub
          </button>
        </div>
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Verify type check**

```bash
cd /svg-project/web && npx tsc --noEmit
```
Expected: no new errors beyond missing sibling modules.

- [ ] **Step 3: Commit**

```bash
git add web/src/pages/LandingPage.tsx
git commit -m "feat: add LandingPage with GitHub OAuth login"
```

---

### Task 7: CallbackPage — capture OAuth JWT token

**Files:**
- Create: `web/src/pages/CallbackPage.tsx`

- [ ] **Step 1: Create web/src/pages/CallbackPage.tsx**

```tsx
import { useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import LoadingSpinner from '../components/LoadingSpinner';

export default function CallbackPage() {
  const [params] = useSearchParams();
  const navigate = useNavigate();

  useEffect(() => {
    const jwt = params.get('token');
    if (jwt) {
      localStorage.setItem('token', jwt);
      navigate('/workspace', { replace: true });
    } else {
      navigate('/', { replace: true });
    }
  }, [params, navigate]);

  return <LoadingSpinner label="Signing you in..." />;
}
```

- [ ] **Step 2: Verify type check**

```bash
cd /svg-project/web && npx tsc --noEmit
```
Expected: no new errors.

- [ ] **Step 3: Commit**

```bash
git add web/src/pages/CallbackPage.tsx
git commit -m "feat: add OAuth callback page to capture JWT token"
```

---

### Task 8: Layout component — sidebar + header

**Files:**
- Create: `web/src/components/Layout.tsx`

- [ ] **Step 1: Create web/src/components/Layout.tsx**

```tsx
import { NavLink, Outlet, useLocation } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export default function Layout() {
  const { userId, logout } = useAuth();
  const location = useLocation();

  const linkClass = (path: string) =>
    `block px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
      location.pathname === path
        ? 'bg-indigo-100 text-indigo-700'
        : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'
    }`;

  return (
    <div className="flex min-h-screen">
      {/* Sidebar */}
      <aside className="w-64 bg-white border-r border-gray-200 flex flex-col">
        <div className="p-6 border-b border-gray-100">
          <h1 className="text-lg font-bold text-gray-900">SVG Converter</h1>
          <p className="text-xs text-gray-500 mt-1 truncate">{userId}</p>
        </div>

        <nav className="flex-1 p-4 space-y-1">
          <NavLink to="/workspace/convert" className={linkClass('/workspace/convert')}>
            + New Conversion
          </NavLink>
          <NavLink to="/workspace/library" className={linkClass('/workspace/library')}>
            My Library
          </NavLink>
        </nav>

        <div className="p-4 border-t border-gray-100">
          <button
            onClick={logout}
            className="w-full text-left px-4 py-2 rounded-lg text-sm text-gray-500 hover:bg-gray-100 transition-colors"
          >
            Log out
          </button>
        </div>
      </aside>

      {/* Main content */}
      <main className="flex-1 p-8">
        <Outlet />
      </main>
    </div>
  );
}
```

- [ ] **Step 2: Verify type check**

```bash
cd /svg-project/web && npx tsc --noEmit
```
Expected: no new errors.

- [ ] **Step 3: Commit**

```bash
git add web/src/components/Layout.tsx
git commit -m "feat: add Layout with sidebar navigation"
```

---

### Task 9: DropZone component

**Files:**
- Create: `web/src/components/DropZone.tsx`

- [ ] **Step 1: Create web/src/components/DropZone.tsx**

```tsx
import { useState, useRef, type DragEvent, type ChangeEvent } from 'react';

interface DropZoneProps {
  onFile: (file: File) => void;
  disabled?: boolean;
}

export default function DropZone({ onFile, disabled = false }: DropZoneProps) {
  const [dragging, setDragging] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  const handleDragOver = (e: DragEvent) => {
    e.preventDefault();
    if (!disabled) setDragging(true);
  };

  const handleDragLeave = (e: DragEvent) => {
    e.preventDefault();
    setDragging(false);
  };

  const handleDrop = (e: DragEvent) => {
    e.preventDefault();
    setDragging(false);
    if (disabled) return;
    const file = e.dataTransfer.files[0];
    if (file && file.type.startsWith('image/')) {
      onFile(file);
    }
  };

  const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) onFile(file);
  };

  return (
    <div
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
      onClick={() => inputRef.current?.click()}
      className={`
        relative cursor-pointer rounded-2xl border-2 border-dashed p-12 text-center transition-colors
        ${dragging ? 'border-indigo-500 bg-indigo-50' : 'border-gray-300 hover:border-gray-400'}
        ${disabled ? 'opacity-50 pointer-events-none' : ''}
      `}
    >
      <input
        ref={inputRef}
        type="file"
        accept="image/*"
        onChange={handleChange}
        className="hidden"
      />
      <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5}
          d="M3 16.5v2.25A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75V16.5m-13.5-9L12 3m0 0l4.5 4.5M12 3v13.5" />
      </svg>
      <p className="mt-4 text-sm text-gray-600">
        <span className="font-semibold text-indigo-600">Click to upload</span> or drag and drop
      </p>
      <p className="mt-1 text-xs text-gray-400">PNG, JPG, GIF, BMP, WEBP up to 10 MB</p>
    </div>
  );
}
```

- [ ] **Step 2: Verify type check**

```bash
cd /svg-project/web && npx tsc --noEmit
```
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add web/src/components/DropZone.tsx
git commit -m "feat: add DropZone drag-and-drop upload component"
```

---

### Task 10: usePolling hook

**Files:**
- Create: `web/src/hooks/usePolling.ts`

- [ ] **Step 1: Create web/src/hooks/usePolling.ts**

```typescript
import { useEffect, useRef } from 'react';

export function usePolling(
  callback: () => void,
  intervalMs: number,
  enabled: boolean,
) {
  const savedCallback = useRef(callback);
  savedCallback.current = callback;

  useEffect(() => {
    if (!enabled) return;
    savedCallback.current();
    const id = setInterval(() => savedCallback.current(), intervalMs);
    return () => clearInterval(id);
  }, [intervalMs, enabled]);
}
```

- [ ] **Step 2: Verify type check**

```bash
cd /svg-project/web && npx tsc --noEmit
```
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add web/src/hooks/usePolling.ts
git commit -m "feat: add usePolling generic hook"
```

---

### Task 11: ConvertPage — upload + processing status

**Files:**
- Create: `web/src/pages/ConvertPage.tsx`

- [ ] **Step 1: Create web/src/pages/ConvertPage.tsx**

```tsx
import { useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import DropZone from '../components/DropZone';
import LoadingSpinner from '../components/LoadingSpinner';
import { conversions, ApiError, type Conversion } from '../api/client';
import { usePolling } from '../hooks/usePolling';

export default function ConvertPage() {
  const navigate = useNavigate();
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [conversionId, setConversionId] = useState<string | null>(null);
  const [status, setStatus] = useState<string | null>(null);

  const pollStatus = useCallback(() => {
    if (!conversionId) return;
    conversions.get(conversionId)
      .then(res => {
        setStatus(res.data.status);
        if (res.data.status === 'completed') {
          navigate(`/workspace/preview/${conversionId}`, { replace: true });
        }
      })
      .catch(() => {});
  }, [conversionId, navigate]);

  usePolling(pollStatus, 1000, status === 'pending' || status === 'processing');

  const handleFile = useCallback(async (file: File) => {
    setError(null);
    setUploading(true);
    try {
      const res = await conversions.upload(file);
      setConversionId(res.data.id);
      setStatus(res.data.status);
    } catch (err) {
      const msg = err instanceof ApiError ? err.message : 'Upload failed';
      setError(msg);
    } finally {
      setUploading(false);
    }
  }, []);

  if (uploading) {
    return <LoadingSpinner label="Uploading..." />;
  }

  if (status === 'pending' || status === 'processing') {
    return (
      <div className="max-w-xl mx-auto">
        <h2 className="text-xl font-bold mb-4">Processing...</h2>
        <LoadingSpinner label={`Status: ${status}`} />
        <p className="text-center text-sm text-gray-500 mt-4">
          Your image is being converted to SVG. This may take a few seconds.
        </p>
      </div>
    );
  }

  return (
    <div className="max-w-xl mx-auto">
      <h2 className="text-xl font-bold mb-6">New Conversion</h2>
      {error && (
        <div className="mb-4 rounded-lg bg-red-50 border border-red-200 p-4 text-sm text-red-700">
          {error}
        </div>
      )}
      <DropZone onFile={handleFile} disabled={uploading} />
    </div>
  );
}
```

- [ ] **Step 2: Verify type check**

```bash
cd /svg-project/web && npx tsc --noEmit
```
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add web/src/pages/ConvertPage.tsx
git commit -m "feat: add ConvertPage with upload and status polling"
```

---

### Task 12: PreviewPage — SVG vs original comparison

**Files:**
- Create: `web/src/pages/PreviewPage.tsx`

- [ ] **Step 1: Create web/src/pages/PreviewPage.tsx**

```tsx
import { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { conversions, ApiError, type Conversion } from '../api/client';
import LoadingSpinner from '../components/LoadingSpinner';

export default function PreviewPage() {
  const { id } = useParams<{ id: string }>();
  const [conv, setConv] = useState<Conversion | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [svgContent, setSvgContent] = useState<string | null>(null);

  useEffect(() => {
    if (!id) return;
    conversions.get(id)
      .then(res => {
        setConv(res.data);
        if (res.data.status === 'completed') {
          return fetch(res.data.svg_url!).then(r => r.text());
        }
        return null;
      })
      .then(svg => {
        if (svg) setSvgContent(svg);
      })
      .catch(err => {
        setError(err instanceof ApiError ? err.message : 'Failed to load');
      })
      .finally(() => setLoading(false));
  }, [id]);

  if (loading) return <LoadingSpinner label="Loading preview..." />;
  if (error) return <div className="text-red-600 text-center py-12">{error}</div>;
  if (!conv) return <div className="text-gray-500 text-center py-12">Not found</div>;

  const sizeReduction = conv.file_size_in > 0
    ? Math.round((1 - (conv.file_size_out || 0) / conv.file_size_in) * 100)
    : 0;

  return (
    <div className="max-w-5xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-bold">Preview</h2>
        <div className="flex gap-3">
          {conv.status === 'completed' && (
            <a
              href={conversions.downloadUrl(conv.id)}
              className="inline-flex items-center gap-2 rounded-lg bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-500 transition-colors"
            >
              <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                  d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
              Download SVG
            </a>
          )}
          <Link
            to="/workspace/library"
            className="inline-flex items-center gap-2 rounded-lg border border-gray-300 px-4 py-2 text-sm font-semibold text-gray-700 hover:bg-gray-50 transition-colors"
          >
            Back to Library
          </Link>
        </div>
      </div>

      {conv.status === 'failed' && (
        <div className="mb-6 rounded-lg bg-red-50 border border-red-200 p-4 text-sm text-red-700">
          Conversion failed: {conv.error_message || 'Unknown error'}
        </div>
      )}

      {/* Comparison view */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="rounded-xl border border-gray-200 bg-white p-4">
          <h3 className="text-xs font-semibold text-gray-400 uppercase mb-3">Original ({conv.format_in})</h3>
          <div className="aspect-square flex items-center justify-center bg-gray-50 rounded-lg overflow-hidden">
            <img
              src={conv.original_url}
              alt="Original"
              className="max-w-full max-h-full object-contain"
            />
          </div>
        </div>

        <div className="rounded-xl border border-gray-200 bg-white p-4">
          <h3 className="text-xs font-semibold text-gray-400 uppercase mb-3">SVG Result</h3>
          <div className="aspect-square flex items-center justify-center bg-gray-50 rounded-lg overflow-hidden">
            {svgContent ? (
              <div
                className="max-w-full max-h-full"
                dangerouslySetInnerHTML={{ __html: svgContent }}
              />
            ) : conv.status === 'pending' || conv.status === 'processing' ? (
              <LoadingSpinner label="Processing..." />
            ) : (
              <span className="text-gray-400 text-sm">Not available</span>
            )}
          </div>
        </div>
      </div>

      {/* Metadata */}
      {conv.status === 'completed' && (
        <div className="mt-6 grid grid-cols-2 sm:grid-cols-4 gap-4">
          <MetaItem label="Input Size" value={formatBytes(conv.file_size_in)} />
          <MetaItem label="Output Size" value={formatBytes(conv.file_size_out)} />
          <MetaItem label="Reduction" value={`${sizeReduction}%`} />
          <MetaItem label="Paths" value={String(conv.path_count)} />
          <MetaItem label="Colors" value={String(conv.color_count)} />
          <MetaItem label="Format" value={conv.format_in.toUpperCase()} />
        </div>
      )}
    </div>
  );
}

function MetaItem({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-gray-200 bg-white p-3">
      <div className="text-xs text-gray-400">{label}</div>
      <div className="text-sm font-semibold text-gray-900">{value}</div>
    </div>
  );
}

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}
```

- [ ] **Step 2: Verify type check**

```bash
cd /svg-project/web && npx tsc --noEmit
```
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add web/src/pages/PreviewPage.tsx
git commit -m "feat: add PreviewPage with SVG comparison and metadata"
```

---

### Task 13: ConversionCard component

**Files:**
- Create: `web/src/components/ConversionCard.tsx`

- [ ] **Step 1: Create web/src/components/ConversionCard.tsx**

```tsx
import { Link } from 'react-router-dom';
import type { Conversion } from '../api/client';

function statusBadge(s: Conversion['status']) {
  const map = {
    pending: 'bg-yellow-100 text-yellow-800',
    processing: 'bg-blue-100 text-blue-800',
    completed: 'bg-green-100 text-green-800',
    failed: 'bg-red-100 text-red-800',
  };
  return map[s] ?? 'bg-gray-100 text-gray-800';
}

function timeAgo(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return 'just now';
  if (mins < 60) return `${mins}m ago`;
  const hrs = Math.floor(mins / 60);
  if (hrs < 24) return `${hrs}h ago`;
  return `${Math.floor(hrs / 24)}d ago`;
}

export default function ConversionCard({ conv }: { conv: Conversion }) {
  return (
    <Link
      to={`/workspace/preview/${conv.id}`}
      className="block rounded-xl border border-gray-200 bg-white hover:shadow-md transition-shadow overflow-hidden"
    >
      <div className="aspect-video bg-gray-100 flex items-center justify-center">
        {conv.thumbnail_url ? (
          <img src={conv.thumbnail_url} alt="" className="w-full h-full object-cover" />
        ) : (
          <svg className="h-8 w-8 text-gray-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5}
              d="M2.25 15.75l5.159-5.159a2.25 2.25 0 013.182 0l5.159 5.159m-1.5-1.5l1.409-1.409a2.25 2.25 0 013.182 0l2.909 2.909M3.75 21h16.5A2.25 2.25 0 0022.5 18.75V5.25A2.25 2.25 0 0020.25 3H3.75A2.25 2.25 0 001.5 5.25v13.5A2.25 2.25 0 003.75 21z" />
          </svg>
        )}
      </div>
      <div className="p-3 flex items-center justify-between">
        <span className="text-xs text-gray-500">{timeAgo(conv.created_at)}</span>
        <span className={`inline-block px-2 py-0.5 rounded-full text-xs font-medium ${statusBadge(conv.status)}`}>
          {conv.status}
        </span>
      </div>
    </Link>
  );
}
```

- [ ] **Step 2: Verify type check**

```bash
cd /svg-project/web && npx tsc --noEmit
```
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add web/src/components/ConversionCard.tsx
git commit -m "feat: add ConversionCard component"
```

---

### Task 14: LibraryPage — conversion history grid

**Files:**
- Create: `web/src/pages/LibraryPage.tsx`

- [ ] **Step 1: Create web/src/pages/LibraryPage.tsx**

```tsx
import { useState, useEffect } from 'react';
import { conversions, ApiError, type Conversion } from '../api/client';
import ConversionCard from '../components/ConversionCard';
import LoadingSpinner from '../components/LoadingSpinner';

export default function LibraryPage() {
  const [items, setItems] = useState<Conversion[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    conversions.list(50, 0)
      .then(res => setItems(res.data))
      .catch(err => setError(err instanceof ApiError ? err.message : 'Load failed'))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <LoadingSpinner label="Loading library..." />;

  return (
    <div className="max-w-5xl mx-auto">
      <h2 className="text-xl font-bold mb-6">My Library</h2>
      {error && (
        <div className="mb-4 rounded-lg bg-red-50 border border-red-200 p-4 text-sm text-red-700">{error}</div>
      )}
      {items.length === 0 ? (
        <div className="text-center py-16 text-gray-400">
          <p className="text-lg">No conversions yet</p>
          <p className="text-sm mt-1">Upload an image to get started.</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {items.map(c => (
            <ConversionCard key={c.id} conv={c} />
          ))}
        </div>
      )}
    </div>
  );
}
```

- [ ] **Step 2: Verify type check**

```bash
cd /svg-project/web && npx tsc --noEmit
```
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add web/src/pages/LibraryPage.tsx
git commit -m "feat: add LibraryPage with conversion history grid"
```

---

### Task 15: Go backend — redirect OAuth callbacks to frontend

**Files:**
- Modify: `server/internal/config/config.go`
- Modify: `server/internal/handler/auth.go`

- [ ] **Step 1: Add FrontendURL to config**

In `server/internal/config/config.go`, add `FrontendURL` to the `Config` struct:

```go
FrontendURL    string
```

In `Load()`, add the field:

```go
FrontendURL:    envOr("FRONTEND_URL", "http://localhost:5173"),
```

- [ ] **Step 2: Fix GithubCallback redirect**

In `server/internal/handler/auth.go`, `GithubCallback` function, change:

```go
c.Redirect(http.StatusFound, "/callback?token="+token)
```

to:

```go
c.Redirect(http.StatusFound, h.cfg.FrontendURL+"/callback?token="+token)
```

- [ ] **Step 3: Fix GoogleCallback redirect**

In `server/internal/handler/auth.go`, `GoogleCallback` function, change:

```go
c.Redirect(http.StatusFound, "/callback?token="+token)
```

to:

```go
c.Redirect(http.StatusFound, h.cfg.FrontendURL+"/callback?token="+token)
```

- [ ] **Step 4: Rebuild Go API binary**

```bash
cd /svg-project/server && go build -o api ./cmd/api/
cd /svg-project && docker compose build api
```

Expected: no errors.

- [ ] **Step 5: Commit**

```bash
git add server/internal/config/config.go server/internal/handler/auth.go
git commit -m "fix: redirect OAuth callbacks to frontend URL"
```

---

### Task 16: Full type check and end-to-end smoke test

**Files:**
- None (verification only)

- [ ] **Step 1: TypeScript check — all pages exist and type correctly**

```bash
cd /svg-project/web && npx tsc --noEmit
```
Expected: zero type errors (all modules now exist).

- [ ] **Step 2: Start API backend**

```bash
cd /svg-project && docker compose up -d
```

- [ ] **Step 3: Start Vite dev server**

```bash
cd /svg-project/web && npx vite --host 0.0.0.0 &
```

- [ ] **Step 4: Verify landing page**

```bash
curl -s http://localhost:5173 | grep -o "SVG Converter"
```
Expected: "SVG Converter"

- [ ] **Step 5: Verify API proxy works**

```bash
curl -s http://localhost:5173/api/v1/auth/me | head -1
```
Expected: `{"error":{"code":"UNAUTHORIZED",...}}` (no token sent, but proves proxy works)

- [ ] **Step 6: Commit**

```bash
echo "Frontend implementation complete. Start with: cd web && npm run dev"
```

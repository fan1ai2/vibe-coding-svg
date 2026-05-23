import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from 'react';
import { auth, ApiError, type User } from '../api/client';

// 认证状态：token + 完整用户信息 + 加载/登录/登出方法
interface AuthState {
  token: string | null;
  user: User | null;
  loading: boolean;
  login: () => void;
  logout: () => void;
}

const AuthContext = createContext<AuthState | null>(null);

// 认证上下文提供者：管理 token 和用户信息、启动时验证 token 有效性、暴露 login/logout 方法
export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(() => localStorage.getItem('token'));
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!token) {
      setLoading(false);
      return;
    }
    // 有 token 时调用 /auth/me 获取完整用户信息（头像、昵称等）
    auth.me()
      .then(data => setUser(data))
      .catch(err => {
        // token 过期或无效时清除本地状态，强制重新登录
        if (err instanceof ApiError && err.status === 401) {
          localStorage.removeItem('token');
          setToken(null);
        }
      })
      .finally(() => setLoading(false));
  }, [token]);

  // 跳转 GitHub OAuth 登录
  const login = useCallback(() => {
    window.location.href = '/api/v1/auth/github/login';
  }, []);

  // 清除 token 和用户信息
  const logout = useCallback(() => {
    localStorage.removeItem('token');
    setToken(null);
    setUser(null);
  }, []);

  return (
    <AuthContext.Provider value={{ token, user, loading, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}

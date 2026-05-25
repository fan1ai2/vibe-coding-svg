import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from 'react';
import { auth, ApiError, type User } from '../api/client';

interface AuthState {
  token: string | null;
  user: User | null;
  loading: boolean;
  isGuest: boolean;
  login: () => void;
  logout: () => void;
  guestLogin: () => Promise<void>;
  onAuthSuccess: (token: string) => void;
}

const AuthContext = createContext<AuthState | null>(null);

const GUEST_COUNT_KEY = 'guest_conversion_count';

export function getGuestRemaining(): number {
  const count = parseInt(localStorage.getItem(GUEST_COUNT_KEY) || '0', 10);
  return Math.max(0, 3 - count);
}

export function incrementGuestCount(): void {
  const count = parseInt(localStorage.getItem(GUEST_COUNT_KEY) || '0', 10);
  localStorage.setItem(GUEST_COUNT_KEY, String(count + 1));
}

export function resetGuestCount(): void {
  localStorage.removeItem(GUEST_COUNT_KEY);
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(() => localStorage.getItem('token'));
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  const isGuest = user?.provider === 'guest';

  useEffect(() => {
    if (!token) {
      setLoading(false);
      return;
    }
    auth.me()
      .then(data => setUser(data))
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
    setUser(null);
  }, []);

  const onAuthSuccess = useCallback((newToken: string) => {
    localStorage.setItem('token', newToken);
    setToken(newToken);
    resetGuestCount();
  }, []);

  const guestLogin = useCallback(async () => {
    const res = await auth.guest();
    localStorage.setItem('token', res.token);
    setToken(res.token);
    setUser(res.user);
  }, []);

  return (
    <AuthContext.Provider value={{ token, user, loading, isGuest, login, logout, guestLogin, onAuthSuccess }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}

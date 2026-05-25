import { useState, useRef, useEffect } from 'react';
import { auth, ApiError } from '../api/client';

interface Props {
  open: boolean;
  onClose: () => void;
  onSuccess: (token: string) => void;
}

export default function EmailLoginModal({ open, onClose, onSuccess }: Props) {
  const [step, setStep] = useState<'email' | 'code'>('email');
  const [email, setEmail] = useState('');
  const [code, setCode] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [countdown, setCountdown] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (open) {
      setStep('email');
      setEmail('');
      setCode('');
      setError('');
      setCountdown(0);
    }
  }, [open]);

  useEffect(() => {
    if (open && inputRef.current) {
      inputRef.current.focus();
    }
  }, [open, step]);

  useEffect(() => {
    if (countdown <= 0) return;
    const t = setTimeout(() => setCountdown(c => c - 1), 1000);
    return () => clearTimeout(t);
  }, [countdown]);

  const handleSendCode = async () => {
    if (!email.trim()) {
      setError('请输入邮箱地址');
      return;
    }
    setLoading(true);
    setError('');
    try {
      await auth.sendCode(email.trim());
      setStep('code');
      setCountdown(60);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : '发送失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  const handleVerify = async () => {
    if (!code.trim()) {
      setError('请输入验证码');
      return;
    }
    setLoading(true);
    setError('');
    try {
      const res = await auth.verifyCode(email.trim(), code.trim());
      localStorage.setItem('token', res.token);
      onSuccess(res.token);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : '验证失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40" onClick={onClose}>
      <div
        className="mx-4 w-full max-w-sm rounded-2xl bg-white p-6 shadow-xl"
        onClick={e => e.stopPropagation()}
      >
        <h2 className="text-lg font-bold text-gray-900">
          {step === 'email' ? '邮箱登录 / 注册' : '输入验证码'}
        </h2>
        <p className="mt-1 text-sm text-gray-500">
          {step === 'email'
            ? '输入邮箱，我们将发送验证码'
            : `验证码已发送至 ${email}`}
        </p>

        {error && (
          <div className="mt-3 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600">{error}</div>
        )}

        {step === 'email' ? (
          <div className="mt-4">
            <input
              ref={inputRef}
              type="email"
              value={email}
              onChange={e => setEmail(e.target.value)}
              onKeyDown={e => e.key === 'Enter' && handleSendCode()}
              placeholder="your@email.com"
              className="w-full rounded-xl border border-gray-200 px-4 py-3 text-sm outline-none focus:border-amber-400 focus:ring-2 focus:ring-amber-100"
            />
            <button
              onClick={handleSendCode}
              disabled={loading}
              className="mt-3 w-full rounded-xl bg-amber-500 py-3 text-sm font-bold text-gray-900 hover:bg-amber-600 disabled:opacity-50"
            >
              {loading ? '发送中...' : '发送验证码'}
            </button>
          </div>
        ) : (
          <div className="mt-4">
            <input
              ref={inputRef}
              type="text"
              maxLength={6}
              value={code}
              onChange={e => setCode(e.target.value.replace(/\D/g, ''))}
              onKeyDown={e => e.key === 'Enter' && handleVerify()}
              placeholder="输入 6 位验证码"
              className="w-full rounded-xl border border-gray-200 px-4 py-3 text-center text-2xl tracking-widest outline-none focus:border-amber-400 focus:ring-2 focus:ring-amber-100"
            />
            <button
              onClick={handleVerify}
              disabled={loading || code.length !== 6}
              className="mt-3 w-full rounded-xl bg-amber-500 py-3 text-sm font-bold text-gray-900 hover:bg-amber-600 disabled:opacity-50"
            >
              {loading ? '验证中...' : '验证登录'}
            </button>
            <button
              onClick={handleSendCode}
              disabled={countdown > 0 || loading}
              className="mt-2 w-full text-sm text-gray-400 hover:text-amber-500 disabled:text-gray-300"
            >
              {countdown > 0 ? `${countdown}s 后重新发送` : '重新发送验证码'}
            </button>
            <button
              onClick={() => { setStep('email'); setError(''); }}
              className="mt-1 w-full text-sm text-gray-400 hover:text-gray-600"
            >
              更换邮箱
            </button>
          </div>
        )}

        <button
          onClick={onClose}
          className="mt-4 w-full rounded-xl py-2 text-sm text-gray-400 hover:text-gray-600"
        >
          取消
        </button>
      </div>
    </div>
  );
}

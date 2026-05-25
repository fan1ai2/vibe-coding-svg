interface Props {
  remaining: number;
  onLogin: () => void;
}

export default function GuestBanner({ remaining, onLogin }: Props) {
  if (remaining < 0) return null;

  return (
    <div className="flex items-center justify-between rounded-xl bg-amber-50 px-4 py-2.5 text-sm">
      <span className="text-amber-700">
        试用模式 — 还剩 <strong>{remaining}</strong> 次免费转换
      </span>
      <button
        onClick={onLogin}
        className="rounded-lg bg-amber-500 px-3 py-1.5 text-xs font-bold text-gray-900 hover:bg-amber-600"
      >
        登录解锁更多
      </button>
    </div>
  );
}

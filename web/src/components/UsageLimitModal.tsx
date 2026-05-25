interface Props {
  open: boolean;
  onLogin: () => void;
  onClose: () => void;
}

export default function UsageLimitModal({ open, onLogin, onClose }: Props) {
  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40" onClick={onClose}>
      <div
        className="mx-4 w-full max-w-sm rounded-2xl bg-white p-6 shadow-xl text-center"
        onClick={e => e.stopPropagation()}
      >
        <div className="mx-auto mb-3 flex h-14 w-14 items-center justify-center rounded-full bg-amber-100">
          <svg className="h-7 w-7 text-amber-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
              d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
          </svg>
        </div>
        <h2 className="text-lg font-bold text-gray-900">试用次数已用完</h2>
        <p className="mt-2 text-sm text-gray-500">
          免费试用限制 3 次转换。登录后可每天转换 20 次，解锁全部功能。
        </p>
        <button
          onClick={onLogin}
          className="mt-5 w-full rounded-xl bg-amber-500 py-3 text-sm font-bold text-gray-900 hover:bg-amber-600"
        >
          登录 / 注册继续使用
        </button>
        <button
          onClick={onClose}
          className="mt-2 w-full rounded-xl py-2 text-sm text-gray-400 hover:text-gray-600"
        >
          以后再说
        </button>
      </div>
    </div>
  );
}

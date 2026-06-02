import { useState } from 'react';
import { useAuth } from '../context/AuthContext';
import { Link } from 'react-router-dom';
import EmailLoginModal from '../components/EmailLoginModal';

const workflows = [
  {
    icon: (
      <svg fill="none" viewBox="0 0 24 24" stroke="currentColor" className="h-8 w-8">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5}
          d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
      </svg>
    ),
    title: '图片转 SVG',
    description: '将位图图片转换为高质量矢量 SVG，支持多种格式',
    action: '开始转换',
    href: '/workspace/convert',
    available: true,
  },
  {
    icon: (
      <svg fill="none" viewBox="0 0 24 24" stroke="currentColor" className="h-8 w-8">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5}
          d="M15.042 21.672L13.684 16.6m0 0l-2.51 2.225.569-9.47 5.227 7.917-3.286-.672zm-7.518-.267A8.25 8.25 0 1120.25 10.5M8.288 14.212A5.25 5.25 0 1117.25 10.5" />
      </svg>
    ),
    title: 'SVG 编辑器',
    description: '在线编辑 SVG 颜色，支持填充/描边、色相饱和度和透明度调整',
    action: '打开编辑器',
    href: '/workspace/editor',
    available: true,
  },
  {
    icon: (
      <svg fill="none" viewBox="0 0 24 24" stroke="currentColor" className="h-8 w-8">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5}
          d="M3.75 6A2.25 2.25 0 016 3.75h2.25A2.25 2.25 0 0110.5 6v2.25a2.25 2.25 0 01-2.25 2.25H6a2.25 2.25 0 01-2.25-2.25V6zM13.5 6a2.25 2.25 0 012.25-2.25H18A2.25 2.25 0 0120.25 6v2.25A2.25 2.25 0 0118 10.5h-2.25a2.25 2.25 0 01-2.25-2.25V6z" />
      </svg>
    ),
    title: '图标库',
    description: '浏览和搜索图标资源，支持标签、颜色、主题多维检索和关联推荐',
    action: '浏览图标库',
    href: '/icons',
    available: true,
  },
  {
    icon: (
      <svg fill="none" viewBox="0 0 24 24" stroke="currentColor" className="h-8 w-8">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5}
          d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09zM18.259 8.715L18 9.75l-.259-1.035a3.375 3.375 0 00-2.455-2.456L14.25 6l1.036-.259a3.375 3.375 0 002.455-2.456L18 2.25l.259 1.035a3.375 3.375 0 002.455 2.456L21.75 6l-1.036.259a3.375 3.375 0 00-2.455 2.456zM16.894 20.567L16.5 21.75l-.394-1.183a2.25 2.25 0 00-1.423-1.423L13.5 18.75l1.183-.394a2.25 2.25 0 001.423-1.423l.394-1.183.394 1.183a2.25 2.25 0 001.423 1.423l1.183.394-1.183.394a2.25 2.25 0 00-1.423 1.423z" />
      </svg>
    ),
    title: 'AI 生成图标',
    description: '用自然语言描述你想要的图标，AI 自动生成多组 SVG 候选',
    action: '开始生成',
    href: '/workspace/ai-generate',
    available: true,
  },
];

export default function LandingPage() {
  const { token, loading, guestLogin, onAuthSuccess } = useAuth();
  const [emailModalOpen, setEmailModalOpen] = useState(false);

  if (loading) return null;

  return (
    <div className="min-h-screen bg-[#FFFDF7]">
      {/* Hero */}
      <section className="relative overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-br from-amber-50 via-white to-amber-50/30" />
        <div className="relative mx-auto max-w-6xl px-6 py-20 text-center sm:py-28">
          <h1 className="text-4xl font-extrabold tracking-tight text-gray-900 sm:text-5xl lg:text-6xl">
            SVG 资源工坊
          </h1>
          <p className="mx-auto mt-6 max-w-2xl text-lg text-gray-500 leading-relaxed">
            图片转矢量 &middot; 在线调色 &middot; 图标库检索 &middot; AI 辅助生成
            <br />
            从灵感创意到 SVG 交付的完整工作流
          </p>
          {!token && (
            <div className="mt-10 mx-auto max-w-xs space-y-3">
              <button
                onClick={guestLogin}
                className="w-full inline-flex items-center justify-center gap-2 rounded-2xl bg-amber-500 px-6 py-3.5 text-base font-bold text-gray-900 shadow-md shadow-amber-200 transition-all duration-300 hover:-translate-y-0.5 hover:bg-amber-600 hover:shadow-lg hover:shadow-amber-300"
              >
                开始免费使用
                <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 8l4 4m0 0l-4 4m4-4H3" />
                </svg>
              </button>
              <button
                onClick={() => setEmailModalOpen(true)}
                className="w-full inline-flex items-center justify-center gap-2 rounded-2xl border-2 border-gray-200 bg-white px-6 py-3.5 text-base font-semibold text-gray-700 transition-all duration-300 hover:-translate-y-0.5 hover:border-amber-300 hover:shadow-md"
              >
                <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                    d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                </svg>
                邮箱登录 / 注册
              </button>
              <a
                href="/api/v1/auth/github/login"
                className="block text-sm text-gray-400 hover:text-gray-600 transition-colors"
              >
                使用 GitHub 账号登录 →
              </a>
            </div>
          )}
        </div>
      </section>

      {/* Workflows */}
      <section className="mx-auto max-w-3xl px-6 pb-24">
        <div className="mb-8 text-center">
          <h2 className="text-xl font-extrabold text-gray-900">选择工作流</h2>
        </div>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          {workflows.map(wf => (
            wf.available && wf.href ? (
              <Link
                key={wf.title}
                to={wf.href}
                className="group flex items-start gap-4 rounded-2xl border border-gray-200 bg-white p-5 transition-shadow hover:shadow-md"
              >
                <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-amber-50 text-amber-500 group-hover:bg-amber-100 transition-colors">
                  {wf.icon}
                </div>
                <div className="min-w-0">
                  <h3 className="text-base font-bold text-gray-900">{wf.title}</h3>
                  <p className="mt-1 text-sm text-gray-500">{wf.description}</p>
                  <span className="inline-block mt-2 text-sm font-semibold text-amber-600 group-hover:text-amber-700">
                    {wf.action} →
                  </span>
                </div>
              </Link>
            ) : (
              <div
                key={wf.title}
                className="flex items-start gap-4 rounded-2xl border border-gray-100 bg-gray-50 p-5 opacity-60"
              >
                <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-gray-100 text-gray-400">
                  {wf.icon}
                </div>
                <div className="min-w-0">
                  <h3 className="text-base font-bold text-gray-900">{wf.title}</h3>
                  <p className="mt-1 text-sm text-gray-500">{wf.description}</p>
                  <span className="inline-block mt-2 text-sm text-gray-400">{wf.action}</span>
                </div>
              </div>
            )
          ))}
        </div>
      </section>

      <EmailLoginModal
        open={emailModalOpen}
        onClose={() => setEmailModalOpen(false)}
        onSuccess={(token) => {
          setEmailModalOpen(false);
          onAuthSuccess(token);
        }}
      />
    </div>
  );
}

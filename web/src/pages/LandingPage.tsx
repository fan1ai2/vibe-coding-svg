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

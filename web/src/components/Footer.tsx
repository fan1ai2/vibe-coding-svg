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

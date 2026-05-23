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

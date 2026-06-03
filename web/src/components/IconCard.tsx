import { useNavigate } from 'react-router-dom';
import type { Icon } from '../api/client';

export default function IconCard({ icon }: { icon: Icon }) {
  const navigate = useNavigate();

  return (
    <div
      className="rounded-xl border border-gray-200 bg-white hover:shadow-md transition-shadow overflow-hidden cursor-pointer"
      onClick={() => navigate(`/icons/${icon.id}`)}
    >
      <div
        className="aspect-video bg-gray-50 flex items-center justify-center p-2"
        dangerouslySetInnerHTML={{ __html: icon.svg_content ?? '' }}
      />
      <div className="p-2 flex items-center justify-between">
        <span className="text-xs text-gray-600 truncate flex-1 mr-2">{icon.name}</span>
        {icon.tags && icon.tags.length > 0 && (
          <span className="text-xs px-1.5 py-0.5 rounded bg-gray-100 text-gray-500 whitespace-nowrap">
            {icon.tags[0].name}
          </span>
        )}
      </div>
    </div>
  );
}

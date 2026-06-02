import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { icons, type Icon } from '../api/client';
import IconCard from '../components/IconCard';
import LoadingSpinner from '../components/LoadingSpinner';

export default function IconDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [icon, setIcon] = useState<Icon | null>(null);
  const [recommended, setRecommended] = useState<Icon[]>([]);
  const [loading, setLoading] = useState(true);
  const [notFound, setNotFound] = useState(false);

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    icons.get(id)
      .then(res => {
        setIcon(res.data);
        setNotFound(false);
        return icons.recommend(id, 10);
      })
      .then(res => {
        if (res) setRecommended(res.data);
      })
      .catch(() => setNotFound(true))
      .finally(() => setLoading(false));
  }, [id]);

  if (loading) return <LoadingSpinner label="加载图标..." />;

  if (notFound || !icon) {
    return (
      <div className="max-w-3xl mx-auto text-center py-16">
        <p className="text-lg text-gray-400 mb-4">图标不存在</p>
        <button
          onClick={() => navigate('/icons')}
          className="text-amber-600 hover:text-amber-700 text-sm"
        >
          返回图标库
        </button>
      </div>
    );
  }

  return (
    <div>
      <div className="flex gap-6 mb-8">
        <div className="w-64 h-64 bg-gray-50 rounded-xl border border-gray-200 flex items-center justify-center p-4 shrink-0"
          dangerouslySetInnerHTML={{ __html: icon.svg_content ?? '' }}
        />
        <div className="flex-1 min-w-0">
          <h2 className="text-xl font-bold mb-3">{icon.name}</h2>
          {icon.tags && icon.tags.length > 0 && (
            <div className="mb-3">
              {(['usage', 'style', 'category'] as const).map(type => {
                const typeTags = icon.tags!.filter(t => t.type === type);
                if (typeTags.length === 0) return null;
                return (
                  <div key={type} className="mb-1.5">
                    <span className="text-xs text-gray-400 mr-2">{type}:</span>
                    {typeTags.map(t => (
                      <span key={t.slug} className="inline-block px-2 py-0.5 rounded text-xs bg-amber-50 text-amber-600 mr-1 mb-1">
                        {t.name}
                      </span>
                    ))}
                  </div>
                );
              })}
            </div>
          )}
          {icon.colors && icon.colors.length > 0 && (
            <div className="flex items-center gap-1 mb-3">
              <span className="text-xs text-gray-400 mr-2">颜色:</span>
              {icon.colors.map(c => (
                <span key={c} className="w-5 h-5 rounded-full border border-gray-300" style={{ backgroundColor: c }} title={c} />
              ))}
            </div>
          )}
          {icon.theme && (
            <p className="text-sm text-gray-500 mb-3">主题: {icon.theme}</p>
          )}
          <button
            onClick={() => {
              sessionStorage.setItem('editor:pendingSvg', icon.svg_content ?? '')
              navigate('/workspace/editor')
            }}
            className="inline-flex items-center gap-2 rounded-lg bg-amber-500 px-4 py-2 text-sm font-semibold text-white hover:bg-amber-600 transition-colors"
          >
            <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
            </svg>
            在编辑器中打开
          </button>
          <p className="text-xs text-gray-400">下载: {icon.download_count} 次</p>
        </div>
      </div>

      {recommended.length > 0 && (
        <>
          <h3 className="text-lg font-semibold mb-3">相关图标</h3>
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3">
            {recommended.map(r => (
              <IconCard key={r.id} icon={r} />
            ))}
          </div>
        </>
      )}
    </div>
  );
}

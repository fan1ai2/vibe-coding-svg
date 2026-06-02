import { useState, useEffect, useCallback } from 'react';
import { useSearchParams } from 'react-router-dom';
import { icons, type Icon } from '../api/client';
import IconCard from '../components/IconCard';
import SearchBar from '../components/SearchBar';
import LoadingSpinner from '../components/LoadingSpinner';

const PAGE_SIZE = 20;

export default function IconLibraryPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [iconList, setIconList] = useState<Icon[]>([]);
  const [loading, setLoading] = useState(true);
  const [hasMore, setHasMore] = useState(false);

  const fetchIcons = useCallback(async (query: string, tagStr: string, offset: number) => {
    setLoading(true);
    try {
      const res = await icons.search({
        q: query || undefined,
        tags: tagStr || undefined,
        sort: 'newest',
        limit: PAGE_SIZE,
        offset,
      });
      if (offset === 0) {
        setIconList(res.data);
      } else {
        setIconList(prev => [...prev, ...res.data]);
      }
      setHasMore(res.data.length === PAGE_SIZE);
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchIcons(searchParams.get('q') || '', searchParams.get('tags') || '', 0);
  }, [searchParams, fetchIcons]);

  const handleSearch = (query: string, tags: string[]) => {
    const params = new URLSearchParams();
    if (query) params.set('q', query);
    if (tags.length > 0) params.set('tags', tags.join(','));
    setSearchParams(params);
  };

  const loadMore = () => {
    fetchIcons(searchParams.get('q') || '', searchParams.get('tags') || '', iconList.length);
  };

  if (loading && iconList.length === 0) return <LoadingSpinner label="加载图标库..." />;

  return (
    <div>
      <h2 className="text-xl font-bold mb-4">图标库</h2>
      <SearchBar
        initialQuery={searchParams.get('q') || ''}
        onSearch={handleSearch}
      />
      {iconList.length === 0 ? (
        <p className="text-center py-16 text-gray-400">暂无图标</p>
      ) : (
        <>
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
            {iconList.map(icon => (
              <IconCard key={icon.id} icon={icon} />
            ))}
          </div>
          {hasMore && (
            <div className="mt-6 text-center">
              <button
                onClick={loadMore}
                disabled={loading}
                className="inline-flex items-center gap-2 rounded-lg border border-gray-300 px-6 py-3 text-sm font-semibold text-gray-700 hover:bg-gray-50 disabled:opacity-50 transition-colors"
              >
                {loading ? '加载中...' : '加载更多'}
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
}
